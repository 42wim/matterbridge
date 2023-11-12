package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/common/dbsetup"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/errors"
	"github.com/status-im/status-go/nodecfg"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/sqlite"
)

type Notifier func(SettingField, interface{})

var (
	// dbInstances holds a map of singleton instances of Database
	dbInstances map[string]*Database

	// mutex guards the instantiation of the dbInstances values, to prevent any concurrent instantiations
	mutex sync.Mutex
)

// Database sql wrapper for operations with browser objects.
type Database struct {
	db                   *sql.DB
	SyncQueue            chan SyncSettingField
	changesSubscriptions []chan *SyncSettingField
	notifier             Notifier
}

// MakeNewDB ensures that a singleton instance of Database is returned per sqlite db file
func MakeNewDB(db *sql.DB) (*Database, error) {
	filename, err := dbsetup.GetDBFilename(db)
	if err != nil {
		return nil, err
	}

	d := &Database{
		db:        db,
		SyncQueue: make(chan SyncSettingField, 100),
	}

	// An empty filename means that the sqlite database is held in memory
	// In this case we don't want to restrict the instantiation
	if filename == "" {
		return d, nil
	}

	// Lock to protect the map from concurrent access
	mutex.Lock()
	defer mutex.Unlock()

	// init dbInstances if it hasn't been already
	if dbInstances == nil {
		dbInstances = map[string]*Database{}
	}

	// If we haven't seen this database file before make an instance
	if _, ok := dbInstances[filename]; !ok {
		dbInstances[filename] = d
	}

	// Check if the current dbInstance is closed, if closed assign new Database
	if err := dbInstances[filename].db.Ping(); err != nil {
		dbInstances[filename] = d
	}

	return dbInstances[filename], nil
}

func (db *Database) GetDB() *sql.DB {
	return db.db
}

func (db *Database) GetSyncQueue() chan SyncSettingField {
	return db.SyncQueue
}

func (db *Database) GetChangesSubscriptions() []chan *SyncSettingField {
	return db.changesSubscriptions
}

func (db *Database) GetNotifier() Notifier {
	return db.notifier
}

func (db *Database) SetSettingsNotifier(n Notifier) {
	db.notifier = n
}

// TODO remove photoPath from settings
func (db *Database) CreateSettings(s Settings, n params.NodeConfig) error {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`
INSERT INTO settings (
  address,
  currency,
  current_network,
  dapps_address,
  device_name,
  preferred_name,
  display_name,
  bio,
  eip1581_address,
  installation_id,
  key_uid,
  keycard_instance_uid,
  keycard_paired_on,
  keycard_pairing,
  latest_derived_path,
  mnemonic,
  name,
  networks,
  photo_path,
  preview_privacy,
  public_key,
  signing_phrase,
  wallet_root_address,
  synthetic_id,
  current_user_status,
  profile_pictures_show_to,
  profile_pictures_visibility,
  url_unfurling_mode,
  omit_transfers_history_scan,
  mnemonic_was_not_shown,
  wallet_token_preferences_group_by_community,
  wallet_show_community_asset_when_sending_tokens,
  wallet_display_assets_below_balance,
  wallet_display_assets_below_balance_threshold,
  wallet_collectible_preferences_group_by_collection,
  wallet_collectible_preferences_group_by_community
) VALUES (
?,?,?,?,?,?,?,?,?,?,?,?,?,?,
?,?,?,?,?,?,?,?,?,'id',?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.Address,
		s.Currency,
		s.CurrentNetwork,
		s.DappsAddress,
		s.DeviceName,
		s.PreferredName,
		s.DisplayName,
		s.Bio,
		s.EIP1581Address,
		s.InstallationID,
		s.KeyUID,
		s.KeycardInstanceUID,
		s.KeycardPairedOn,
		s.KeycardPairing,
		s.LatestDerivedPath,
		s.Mnemonic,
		s.Name,
		s.Networks,
		s.PhotoPath,
		s.PreviewPrivacy,
		s.PublicKey,
		s.SigningPhrase,
		s.WalletRootAddress,
		s.CurrentUserStatus,
		s.ProfilePicturesShowTo,
		s.ProfilePicturesVisibility,
		s.URLUnfurlingMode,
		s.OmitTransfersHistoryScan,
		s.MnemonicWasNotShown,
		s.TokenGroupByCommunity,
		s.ShowCommunityAssetWhenSendingTokens,
		s.DisplayAssetsBelowBalance,
		s.DisplayAssetsBelowBalanceThreshold,
		s.CollectibleGroupByCollection,
		s.CollectibleGroupByCommunity,
	)
	if err != nil {
		return err
	}

	if s.DisplayName != "" {
		now := time.Now().Unix()
		query := db.buildUpdateSyncClockQueryForField(DisplayName)
		_, err := tx.Exec(query, uint64(now), uint64(now))
		if err != nil {
			return err
		}
	}

	return nodecfg.SaveConfigWithTx(tx, &n)
}

func (db *Database) getSettingFieldFromReactName(reactName string) (SettingField, error) {
	for _, s := range SettingFieldRegister {
		if s.GetReactName() == reactName {
			return s, nil
		}
	}
	return SettingField{}, errors.ErrInvalidConfig
}

func (db *Database) makeSelectRow(setting SettingField) *sql.Row {
	query := "SELECT %s FROM settings WHERE synthetic_id = 'id'"
	query = fmt.Sprintf(query, setting.GetDBName())
	return db.db.QueryRow(query)
}

func (db *Database) makeSelectString(setting SettingField) (string, error) {
	var result sql.NullString
	err := db.makeSelectRow(setting).Scan(&result)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if result.Valid {
		return result.String, nil
	}
	return "", err
}

func (db *Database) saveSetting(setting SettingField, value interface{}) error {
	query := "UPDATE settings SET %s = ? WHERE synthetic_id = 'id'"
	query = fmt.Sprintf(query, setting.GetDBName())

	update, err := db.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = update.Exec(value)

	if err != nil {
		return err
	}

	if db.notifier != nil {
		db.notifier(setting, value)
	}

	return nil
}

func (db *Database) parseSaveAndSyncSetting(sf SettingField, value interface{}) (err error) {
	if sf.ValueHandler() != nil {
		value, err = sf.ValueHandler()(value)
		if err != nil {
			return err
		}
	}

	// TODO(samyoul) this is ugly as hell need a more elegant solution
	if NodeConfig.GetReactName() == sf.GetReactName() {
		if err = nodecfg.SaveNodeConfig(db.db, value.(*params.NodeConfig)); err != nil {
			return err
		}
		value = nil
	}

	err = db.saveSetting(sf, value)
	if err != nil {
		return err
	}

	if sf.GetDBName() == DBColumnMnemonic {
		mnemonicRemoved := value == nil || value.(string) == ""
		err = db.saveSetting(MnemonicRemoved, mnemonicRemoved)
		if err != nil {
			return err
		}
		sf = MnemonicRemoved
		value = mnemonicRemoved
	}

	if sf.CanSync(FromInterface) {
		db.SyncQueue <- SyncSettingField{sf, value}
	}

	db.postChangesToSubscribers(&SyncSettingField{sf, value})

	return nil
}

// SaveSetting stores data from any non-sync source
// If the field requires syncing the field data is pushed on to the SyncQueue
func (db *Database) SaveSetting(setting string, value interface{}) error {
	sf, err := db.getSettingFieldFromReactName(setting)
	if err != nil {
		return err
	}

	return db.parseSaveAndSyncSetting(sf, value)
}

// SaveSettingField is identical in functionality to SaveSetting, except the setting parameter is a SettingField and
// doesn't require any SettingFieldRegister lookup.
// This func is useful if you already know the SettingField to save
func (db *Database) SaveSettingField(sf SettingField, value interface{}) error {
	return db.parseSaveAndSyncSetting(sf, value)
}

func (db *Database) DeleteMnemonic() error {
	return db.saveSetting(Mnemonic, nil)
}

// SaveSyncSetting stores setting data from a sync protobuf source, note it does not call SettingField.ValueHandler()
// nor does this function attempt to write to the Database.SyncQueue,
// yet it still writes to Database.changesSubscriptions.
func (db *Database) SaveSyncSetting(setting SettingField, value interface{}, clock uint64) error {
	ls, err := db.GetSettingLastSynced(setting)
	if err != nil {
		return err
	}
	if clock <= ls {
		return errors.ErrNewClockOlderThanCurrent
	}

	err = db.SetSettingLastSynced(setting, clock)
	if err != nil {
		return err
	}

	err = db.saveSetting(setting, value)
	if err != nil {
		return err
	}

	db.postChangesToSubscribers(&SyncSettingField{setting, value})
	return nil
}

func (db *Database) GetSettingLastSynced(setting SettingField) (result uint64, err error) {
	query := "SELECT %s FROM settings_sync_clock WHERE synthetic_id = 'id'"
	query = fmt.Sprintf(query, setting.GetDBName())

	err = db.db.QueryRow(query).Scan(&result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (db *Database) buildUpdateSyncClockQueryForField(setting SettingField) string {
	query := "UPDATE settings_sync_clock SET %s = ? WHERE synthetic_id = 'id' AND %s < ?"
	return fmt.Sprintf(query, setting.GetDBName(), setting.GetDBName())
}

func (db *Database) SetSettingLastSynced(setting SettingField, clock uint64) error {
	query := db.buildUpdateSyncClockQueryForField(setting)

	_, err := db.db.Exec(query, clock, clock)
	return err
}

func (db *Database) GetSettings() (Settings, error) {
	var s Settings
	err := db.db.QueryRow(`
	SELECT
		address, anon_metrics_should_send, chaos_mode, currency, current_network,
		custom_bootnodes, custom_bootnodes_enabled, dapps_address, display_name, bio, eip1581_address, fleet,
		hide_home_tooltip, installation_id, key_uid, keycard_instance_uid, keycard_paired_on, keycard_pairing,
		last_updated, latest_derived_path, link_preview_request_enabled, link_previews_enabled_sites, log_level,
		mnemonic, mnemonic_removed, name, networks, notifications_enabled, push_notifications_server_enabled,
		push_notifications_from_contacts_only, remote_push_notifications_enabled, send_push_notifications,
		push_notifications_block_mentions, photo_path, pinned_mailservers, preferred_name, preview_privacy, public_key,
		remember_syncing_choice, signing_phrase, stickers_packs_installed, stickers_packs_pending, stickers_recent_stickers,
		syncing_on_mobile_network, default_sync_period, use_mailservers, messages_from_contacts_only, usernames, appearance,
		profile_pictures_show_to, profile_pictures_visibility, wallet_root_address, wallet_set_up_passed, wallet_visible_tokens,
		waku_bloom_filter_mode, webview_allow_permission_requests, current_user_status, send_status_updates, gif_recents,
		gif_favorites, opensea_enabled, last_backup, backup_enabled, telemetry_server_url, auto_message_enabled, gif_api_key,
		test_networks_enabled, mutual_contact_enabled, profile_migration_needed, is_sepolia_enabled, wallet_token_preferences_group_by_community, url_unfurling_mode,
		omit_transfers_history_scan, mnemonic_was_not_shown, wallet_show_community_asset_when_sending_tokens, wallet_display_assets_below_balance,
		wallet_display_assets_below_balance_threshold, wallet_collectible_preferences_group_by_collection, wallet_collectible_preferences_group_by_community
	FROM
		settings
	WHERE
		synthetic_id = 'id'`).Scan(
		&s.Address,
		&s.AnonMetricsShouldSend,
		&s.ChaosMode,
		&s.Currency,
		&s.CurrentNetwork,
		&s.CustomBootnodes,
		&s.CustomBootnodesEnabled,
		&s.DappsAddress,
		&s.DisplayName,
		&s.Bio,
		&s.EIP1581Address,
		&s.Fleet,
		&s.HideHomeTooltip,
		&s.InstallationID,
		&s.KeyUID,
		&s.KeycardInstanceUID,
		&s.KeycardPairedOn,
		&s.KeycardPairing,
		&s.LastUpdated,
		&s.LatestDerivedPath,
		&s.LinkPreviewRequestEnabled,
		&s.LinkPreviewsEnabledSites,
		&s.LogLevel,
		&s.Mnemonic,
		&s.MnemonicRemoved,
		&s.Name,
		&s.Networks,
		&s.NotificationsEnabled,
		&s.PushNotificationsServerEnabled,
		&s.PushNotificationsFromContactsOnly,
		&s.RemotePushNotificationsEnabled,
		&s.SendPushNotifications,
		&s.PushNotificationsBlockMentions,
		&s.PhotoPath,
		&s.PinnedMailserver,
		&s.PreferredName,
		&s.PreviewPrivacy,
		&s.PublicKey,
		&s.RememberSyncingChoice,
		&s.SigningPhrase,
		&s.StickerPacksInstalled,
		&s.StickerPacksPending,
		&s.StickersRecentStickers,
		&s.SyncingOnMobileNetwork,
		&s.DefaultSyncPeriod,
		&s.UseMailservers,
		&s.MessagesFromContactsOnly,
		&s.Usernames,
		&s.Appearance,
		&s.ProfilePicturesShowTo,
		&s.ProfilePicturesVisibility,
		&s.WalletRootAddress,
		&s.WalletSetUpPassed,
		&s.WalletVisibleTokens,
		&s.WakuBloomFilterMode,
		&s.WebViewAllowPermissionRequests,
		&sqlite.JSONBlob{Data: &s.CurrentUserStatus},
		&s.SendStatusUpdates,
		&sqlite.JSONBlob{Data: &s.GifRecents},
		&sqlite.JSONBlob{Data: &s.GifFavorites},
		&s.OpenseaEnabled,
		&s.LastBackup,
		&s.BackupEnabled,
		&s.TelemetryServerURL,
		&s.AutoMessageEnabled,
		&s.GifAPIKey,
		&s.TestNetworksEnabled,
		&s.MutualContactEnabled,
		&s.ProfileMigrationNeeded,
		&s.IsSepoliaEnabled,
		&s.TokenGroupByCommunity,
		&s.URLUnfurlingMode,
		&s.OmitTransfersHistoryScan,
		&s.MnemonicWasNotShown,
		&s.ShowCommunityAssetWhenSendingTokens,
		&s.DisplayAssetsBelowBalance,
		&s.DisplayAssetsBelowBalanceThreshold,
		&s.CollectibleGroupByCollection,
		&s.CollectibleGroupByCommunity,
	)

	return s, err
}

// We should remove this and realated things once mobile team starts usign `settings_notifications` package
func (db *Database) GetNotificationsEnabled() (result bool, err error) {
	err = db.makeSelectRow(NotificationsEnabled).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetProfilePicturesVisibility() (result int, err error) {
	err = db.makeSelectRow(ProfilePicturesVisibility).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetPublicKey() (string, error) {
	return db.makeSelectString(PublicKey)
}

func (db *Database) GetFleet() (string, error) {
	return db.makeSelectString(Fleet)
}

func (db *Database) GetDappsAddress() (rst types.Address, err error) {
	err = db.makeSelectRow(DappsAddress).Scan(&rst)
	if err == sql.ErrNoRows {
		return rst, nil
	}
	return
}

func (db *Database) GetPinnedMailservers() (rst map[string]string, err error) {
	rst = make(map[string]string)
	var pinnedMailservers string
	err = db.db.QueryRow("SELECT COALESCE(pinned_mailservers, '') FROM settings WHERE synthetic_id = 'id'").Scan(&pinnedMailservers)
	if err == sql.ErrNoRows || pinnedMailservers == "" {
		return rst, nil
	}

	err = json.Unmarshal([]byte(pinnedMailservers), &rst)
	if err != nil {
		return nil, err
	}
	return
}

func (db *Database) CanUseMailservers() (result bool, err error) {
	err = db.makeSelectRow(UseMailservers).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) CanSyncOnMobileNetwork() (result bool, err error) {
	err = db.makeSelectRow(SyncingOnMobileNetwork).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetDefaultSyncPeriod() (result uint32, err error) {
	err = db.makeSelectRow(DefaultSyncPeriod).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetMessagesFromContactsOnly() (result bool, err error) {
	err = db.makeSelectRow(MessagesFromContactsOnly).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetProfilePicturesShowTo() (result int64, err error) {
	err = db.makeSelectRow(ProfilePicturesShowTo).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetLatestDerivedPath() (result uint, err error) {
	err = db.makeSelectRow(LatestDerivedPath).Scan(&result)
	return
}

func (db *Database) GetCurrentStatus(status interface{}) error {
	err := db.makeSelectRow(CurrentUserStatus).Scan(&sqlite.JSONBlob{Data: &status})
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (db *Database) ShouldBroadcastUserStatus() (result bool, err error) {
	err = db.makeSelectRow(SendStatusUpdates).Scan(&result)
	// If the `send_status_updates` value is nil the sql.ErrNoRows will be returned
	// because this feature is opt out, `true` should be returned in the case where no value is found
	if err == sql.ErrNoRows {
		return true, nil
	}
	return result, err
}

func (db *Database) BackupEnabled() (result bool, err error) {
	err = db.makeSelectRow(BackupEnabled).Scan(&result)
	if err == sql.ErrNoRows {
		return true, nil
	}
	return result, err
}

func (db *Database) AutoMessageEnabled() (result bool, err error) {
	err = db.makeSelectRow(AutoMessageEnabled).Scan(&result)
	if err == sql.ErrNoRows {
		return true, nil
	}
	return result, err
}

func (db *Database) LastBackup() (result uint64, err error) {
	err = db.makeSelectRow(LastBackup).Scan(&result)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return result, err
}

func (db *Database) SetLastBackup(time uint64) error {
	return db.SaveSettingField(LastBackup, time)
}

func (db *Database) SetBackupFetched(fetched bool) error {
	return db.SaveSettingField(BackupFetched, fetched)
}

func (db *Database) BackupFetched() (result bool, err error) {
	err = db.makeSelectRow(BackupFetched).Scan(&result)
	if err == sql.ErrNoRows {
		return true, nil
	}
	return result, err
}

func (db *Database) ENSName() (string, error) {
	return db.makeSelectString(PreferredName)
}

func (db *Database) DeviceName() (string, error) {
	return db.makeSelectString(DeviceName)
}

func (db *Database) DisplayName() (string, error) {
	return db.makeSelectString(DisplayName)
}

func (db *Database) Bio() (string, error) {
	return db.makeSelectString(Bio)
}

func (db *Database) Mnemonic() (string, error) {
	return db.makeSelectString(Mnemonic)
}

func (db *Database) MnemonicRemoved() (result bool, err error) {
	err = db.makeSelectRow(MnemonicRemoved).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetMnemonicWasNotShown() (result bool, err error) {
	err = db.makeSelectRow(MnemonicWasNotShown).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GifAPIKey() (string, error) {
	return db.makeSelectString(GifAPIKey)
}

func (db *Database) MutualContactEnabled() (result bool, err error) {
	err = db.makeSelectRow(MutualContactEnabled).Scan(&result)
	return result, err
}

func (db *Database) GifRecents() (recents json.RawMessage, err error) {
	err = db.makeSelectRow(GifRecents).Scan(&sqlite.JSONBlob{Data: &recents})
	if err == sql.ErrNoRows {
		return nil, err
	}
	return recents, nil
}

func (db *Database) GifFavorites() (favorites json.RawMessage, err error) {
	err = db.makeSelectRow(GifFavourites).Scan(&sqlite.JSONBlob{Data: &favorites})
	if err == sql.ErrNoRows {
		return nil, err
	}
	return favorites, nil
}

func (db *Database) GetPreferredUsername() (string, error) {
	return db.makeSelectString(PreferredName)
}

func (db *Database) GetCurrency() (string, error) {
	return db.makeSelectString(Currency)
}

func (db *Database) GetInstalledStickerPacks() (rst *json.RawMessage, err error) {
	err = db.makeSelectRow(StickersPacksInstalled).Scan(&rst)
	return
}

func (db *Database) GetPendingStickerPacks() (rst *json.RawMessage, err error) {
	err = db.makeSelectRow(StickersPacksPending).Scan(&rst)
	return
}

func (db *Database) GetRecentStickers() (rst *json.RawMessage, err error) {
	err = db.makeSelectRow(StickersRecentStickers).Scan(&rst)
	return
}

func (db *Database) SetPinnedMailservers(mailservers map[string]string) error {
	return db.SaveSettingField(PinnedMailservers, mailservers)
}

func (db *Database) SetUseMailservers(value bool) error {
	return db.SaveSettingField(UseMailservers, value)
}

func (db *Database) GetWalletRootAddress() (rst types.Address, err error) {
	err = db.makeSelectRow(WalletRootAddress).Scan(&rst)
	if err == sql.ErrNoRows {
		return rst, nil
	}
	return
}

func (db *Database) GetEIP1581Address() (rst types.Address, err error) {
	err = db.makeSelectRow(EIP1581Address).Scan(&rst)
	if err == sql.ErrNoRows {
		return rst, nil
	}
	return
}

func (db *Database) GetMasterAddress() (rst types.Address, err error) {
	err = db.makeSelectRow(MasterAddress).Scan(&rst)
	if err == sql.ErrNoRows {
		return rst, nil
	}
	return
}

func (db *Database) GetTestNetworksEnabled() (result bool, err error) {
	err = db.makeSelectRow(TestNetworksEnabled).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetIsSepoliaEnabled() (result bool, err error) {
	err = db.makeSelectRow(IsSepoliaEnabled).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) GetTokenGroupByCommunity() (result bool, err error) {
	err = db.makeSelectRow(TokenGroupByCommunity).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) SetTokenGroupByCommunity(value bool) error {
	return db.SaveSettingField(TokenGroupByCommunity, value)
}

func (db *Database) GetCollectibleGroupByCollection() (result bool, err error) {
	err = db.makeSelectRow(CollectibleGroupByCollection).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) SetCollectibleGroupByCollection(value bool) error {
	return db.SaveSettingField(CollectibleGroupByCollection, value)
}

func (db *Database) GetCollectibleGroupByCommunity() (result bool, err error) {
	err = db.makeSelectRow(CollectibleGroupByCommunity).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) SetCollectibleGroupByCommunity(value bool) error {
	return db.SaveSettingField(CollectibleGroupByCommunity, value)
}

func (db *Database) GetTelemetryServerURL() (string, error) {
	return db.makeSelectString(TelemetryServerURL)
}

func (db *Database) ProfileMigrationNeeded() (result bool, err error) {
	err = db.makeSelectRow(ProfileMigrationNeeded).Scan(&result)
	return result, err
}

func (db *Database) URLUnfurlingMode() (result int64, err error) {
	err = db.makeSelectRow(URLUnfurlingMode).Scan(&result)
	if err == sql.ErrNoRows {
		return result, nil
	}
	return result, err
}

func (db *Database) SubscribeToChanges() chan *SyncSettingField {
	s := make(chan *SyncSettingField, 100)
	db.changesSubscriptions = append(db.changesSubscriptions, s)
	return s
}

func (db *Database) postChangesToSubscribers(change *SyncSettingField) {
	// Publish on channels, drop if buffer is full
	for _, s := range db.changesSubscriptions {
		select {
		case s <- change:
		default:
			log.Warn("settings changes subscription channel full, dropping message")
		}
	}
}

func (db *Database) MnemonicWasShown() error {
	return db.SaveSettingField(MnemonicWasNotShown, false)
}
