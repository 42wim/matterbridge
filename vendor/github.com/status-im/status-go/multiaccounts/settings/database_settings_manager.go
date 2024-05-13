package settings

import (
	"database/sql"
	"encoding/json"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
)

type DatabaseSettingsManager interface {
	GetDB() *sql.DB
	GetSyncQueue() chan SyncSettingField
	GetChangesSubscriptions() []chan *SyncSettingField
	GetNotifier() Notifier
	GetSettingLastSynced(setting SettingField) (result uint64, err error)
	GetSettings() (Settings, error)
	GetNotificationsEnabled() (result bool, err error)
	GetProfilePicturesVisibility() (result int, err error)
	GetPublicKey() (string, error)
	GetFleet() (string, error)
	GetDappsAddress() (rst types.Address, err error)
	GetPinnedMailservers() (rst map[string]string, err error)
	GetDefaultSyncPeriod() (result uint32, err error)
	GetMessagesFromContactsOnly() (result bool, err error)
	GetProfilePicturesShowTo() (result int64, err error)
	GetLatestDerivedPath() (result uint, err error)
	GetCurrentStatus(status interface{}) error
	GetMnemonicWasNotShown() (result bool, err error)
	GetPreferredUsername() (string, error)
	GetCurrency() (string, error)
	GetInstalledStickerPacks() (rst *json.RawMessage, err error)
	GetPendingStickerPacks() (rst *json.RawMessage, err error)
	GetRecentStickers() (rst *json.RawMessage, err error)
	GetWalletRootAddress() (rst types.Address, err error)
	GetEIP1581Address() (rst types.Address, err error)
	GetMasterAddress() (rst types.Address, err error)
	GetTestNetworksEnabled() (result bool, err error)
	GetIsGoerliEnabled() (result bool, err error)
	GetTokenGroupByCommunity() (result bool, err error)
	GetCollectibleGroupByCommunity() (result bool, err error)
	GetCollectibleGroupByCollection() (result bool, err error)
	GetTelemetryServerURL() (string, error)

	SetSettingsNotifier(n Notifier)
	SetSettingLastSynced(setting SettingField, clock uint64) error
	SetLastBackup(time uint64) error
	SetBackupFetched(fetched bool) error
	SetPinnedMailservers(mailservers map[string]string) error
	SetUseMailservers(value bool) error
	SetTokenGroupByCommunity(value bool) error
	SetPeerSyncingEnabled(value bool) error

	CreateSettings(s Settings, n params.NodeConfig) error
	SaveSetting(setting string, value interface{}) error
	SaveSettingField(sf SettingField, value interface{}) error
	DeleteMnemonic() error
	SaveSyncSetting(setting SettingField, value interface{}, clock uint64) error
	CanUseMailservers() (result bool, err error)
	CanSyncOnMobileNetwork() (result bool, err error)
	ShouldBroadcastUserStatus() (result bool, err error)
	BackupEnabled() (result bool, err error)
	AutoMessageEnabled() (result bool, err error)
	LastBackup() (result uint64, err error)
	BackupFetched() (result bool, err error)
	ENSName() (string, error)
	DeviceName() (string, error)
	DisplayName() (string, error)
	Bio() (string, error)
	Mnemonic() (string, error)
	MnemonicRemoved() (result bool, err error)
	GifAPIKey() (string, error)
	MutualContactEnabled() (result bool, err error)
	GifRecents() (recents json.RawMessage, err error)
	GifFavorites() (favorites json.RawMessage, err error)
	ProfileMigrationNeeded() (result bool, err error)
	URLUnfurlingMode() (result int64, err error)
	SubscribeToChanges() chan *SyncSettingField
	MnemonicWasShown() error
	GetPeerSyncingEnabled() (result bool, err error)
}
