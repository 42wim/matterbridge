package settings

import (
	"encoding/json"
	"reflect"

	accountJson "github.com/status-im/status-go/account/json"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

type ValueHandler func(interface{}) (interface{}, error)
type ValueCastHandler func(interface{}) (interface{}, error)
type SyncSettingProtobufFactoryInterface func(interface{}, uint64, string) (*common.RawMessage, *protobuf.SyncSetting, error)
type SyncSettingProtobufFactoryStruct func(Settings, uint64, string) (*common.RawMessage, *protobuf.SyncSetting, error)
type SyncSettingProtobufToValue func(setting *protobuf.SyncSetting) interface{}

// SyncProtobufFactory represents a collection of functionality to generate and parse *protobuf.SyncSetting
type SyncProtobufFactory struct {
	inactive          bool
	fromInterface     SyncSettingProtobufFactoryInterface
	fromStruct        SyncSettingProtobufFactoryStruct
	valueFromProtobuf SyncSettingProtobufToValue
	protobufType      protobuf.SyncSetting_Type
}

func (spf *SyncProtobufFactory) Inactive() bool {
	return spf.inactive
}

func (spf *SyncProtobufFactory) FromInterface() SyncSettingProtobufFactoryInterface {
	return spf.fromInterface
}

func (spf *SyncProtobufFactory) FromStruct() SyncSettingProtobufFactoryStruct {
	return spf.fromStruct
}

func (spf *SyncProtobufFactory) ExtractValueFromProtobuf() SyncSettingProtobufToValue {
	return spf.valueFromProtobuf
}

func (spf *SyncProtobufFactory) SyncSettingProtobufType() protobuf.SyncSetting_Type {
	return spf.protobufType
}

// SyncSettingField represents a binding between a Value and a SettingField
type SyncSettingField struct {
	SettingField
	Value interface{}
}

func (s SyncSettingField) MarshalJSON() ([]byte, error) {
	alias := struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	}{
		s.reactFieldName,
		s.Value,
	}

	return json.Marshal(alias)
}

// SettingField represents an individual setting in the database, it contains context dependant names and optional
// pre-store value parsing, along with optional *SyncProtobufFactory
type SettingField struct {
	reactFieldName      string
	dBColumnName        string
	valueHandler        ValueHandler
	syncProtobufFactory *SyncProtobufFactory
	valueCastHandler    ValueCastHandler
}

func (s SettingField) GetReactName() string {
	return s.reactFieldName
}

func (s SettingField) GetDBName() string {
	return s.dBColumnName
}

func (s SettingField) ValueHandler() ValueHandler {
	return s.valueHandler
}

func (s SettingField) ValueCastHandler() ValueCastHandler {
	return s.valueCastHandler
}

func (s SettingField) SyncProtobufFactory() *SyncProtobufFactory {
	return s.syncProtobufFactory
}

// CanSync checks if a SettingField has functions supporting the syncing of
func (s SettingField) CanSync(source SyncSource) bool {
	spf := s.syncProtobufFactory

	if spf == nil {
		return false
	}

	if spf.inactive {
		return false
	}

	switch source {
	case FromInterface:
		return spf.fromInterface != nil
	case FromStruct:
		return spf.fromStruct != nil
	default:
		return false
	}
}

func (s SettingField) Equals(other SettingField) bool {
	return s.reactFieldName == other.reactFieldName
}

// Settings represents the entire setting row stored in the application db
type Settings struct {
	// required
	Address                   types.Address    `json:"address"`
	AnonMetricsShouldSend     bool             `json:"anon-metrics/should-send?,omitempty"`
	ChaosMode                 bool             `json:"chaos-mode?,omitempty"`
	Currency                  string           `json:"currency,omitempty"`
	CurrentNetwork            string           `json:"networks/current-network"`
	CustomBootnodes           *json.RawMessage `json:"custom-bootnodes,omitempty"`
	CustomBootnodesEnabled    *json.RawMessage `json:"custom-bootnodes-enabled?,omitempty"`
	DappsAddress              types.Address    `json:"dapps-address"`
	DeviceName                string           `json:"device-name"`
	DisplayName               string           `json:"display-name"`
	Bio                       string           `json:"bio,omitempty"`
	EIP1581Address            types.Address    `json:"eip1581-address"`
	Fleet                     *string          `json:"fleet,omitempty"`
	HideHomeTooltip           bool             `json:"hide-home-tooltip?,omitempty"`
	InstallationID            string           `json:"installation-id"`
	KeyUID                    string           `json:"key-uid"`
	KeycardInstanceUID        string           `json:"keycard-instance-uid,omitempty"`
	KeycardPairedOn           int64            `json:"keycard-paired-on,omitempty"`
	KeycardPairing            string           `json:"keycard-pairing,omitempty"`
	LastUpdated               *int64           `json:"last-updated,omitempty"`
	LatestDerivedPath         uint             `json:"latest-derived-path"`
	LinkPreviewRequestEnabled bool             `json:"link-preview-request-enabled,omitempty"`
	LinkPreviewsEnabledSites  *json.RawMessage `json:"link-previews-enabled-sites,omitempty"`
	LogLevel                  *string          `json:"log-level,omitempty"`
	MessagesFromContactsOnly  bool             `json:"messages-from-contacts-only"`
	Mnemonic                  *string          `json:"mnemonic,omitempty"`
	// NOTE(rasom): negation here because it safer/simpler to have false by default
	MnemonicWasNotShown      bool             `json:"mnemonic-was-not-shown?,omitempty"`
	MnemonicRemoved          bool             `json:"mnemonic-removed?,omitempty"`
	OmitTransfersHistoryScan bool             `json:"omit-transfers-history-scan?,omitempty"`
	MutualContactEnabled     bool             `json:"mutual-contact-enabled?"`
	Name                     string           `json:"name,omitempty"`
	Networks                 *json.RawMessage `json:"networks/networks"`
	// NotificationsEnabled indicates whether local notifications should be enabled (android only)
	NotificationsEnabled bool             `json:"notifications-enabled?,omitempty"`
	PhotoPath            string           `json:"photo-path"`
	PinnedMailserver     *json.RawMessage `json:"pinned-mailservers,omitempty"`
	// PreferredName represents the user's preferred Ethereum Name Service (ENS) name.
	// If a user has multiple ENS names, they can select one as the PreferredName.
	// When PreferredName is set, it takes precedence over the DisplayName for displaying the user's name.
	// If PreferredName is empty or doesn't match any of the user's ENS names, the DisplayName is used instead.
	//
	// There is a race condition between updating DisplayName and PreferredName, where the account.Name field
	// could be incorrectly updated based on the order in which the backup messages (BackedUpProfile/BackedUpSettings) arrive.
	// To handle this race condition, the code checks the LastSynced clock value for both DisplayName and PreferredName,
	// and updates account.Name with the value that has the latest clock
	PreferredName  *string `json:"preferred-name,omitempty"`
	PreviewPrivacy bool    `json:"preview-privacy?"`
	PublicKey      string  `json:"public-key"`
	// PushNotificationsServerEnabled indicates whether we should be running a push notification server
	PushNotificationsServerEnabled bool `json:"push-notifications-server-enabled?,omitempty"`
	// PushNotificationsFromContactsOnly indicates whether we should only receive push notifications from contacts
	PushNotificationsFromContactsOnly bool `json:"push-notifications-from-contacts-only?,omitempty"`
	// PushNotificationsBlockMentions indicates whether we should receive notifications for mentions
	PushNotificationsBlockMentions bool `json:"push-notifications-block-mentions?,omitempty"`
	RememberSyncingChoice          bool `json:"remember-syncing-choice?,omitempty"`
	// RemotePushNotificationsEnabled indicates whether we should be using remote notifications (ios only for now)
	RemotePushNotificationsEnabled bool             `json:"remote-push-notifications-enabled?,omitempty"`
	SigningPhrase                  string           `json:"signing-phrase"`
	StickerPacksInstalled          *json.RawMessage `json:"stickers/packs-installed,omitempty"`
	StickerPacksPending            *json.RawMessage `json:"stickers/packs-pending,omitempty"`
	StickersRecentStickers         *json.RawMessage `json:"stickers/recent-stickers,omitempty"`
	SyncingOnMobileNetwork         bool             `json:"syncing-on-mobile-network?,omitempty"`
	// DefaultSyncPeriod is how far back in seconds we should pull messages from a mailserver
	DefaultSyncPeriod uint `json:"default-sync-period"`
	// SendPushNotifications indicates whether we should send push notifications for other clients
	SendPushNotifications bool `json:"send-push-notifications?,omitempty"`
	Appearance            uint `json:"appearance"`
	// ProfilePicturesShowTo indicates to whom the user shows their profile picture to (contacts, everyone)
	ProfilePicturesShowTo ProfilePicturesShowToType `json:"profile-pictures-show-to"`
	// ProfilePicturesVisibility indicates who we want to see profile pictures of (contacts, everyone or none)
	ProfilePicturesVisibility           ProfilePicturesVisibilityType `json:"profile-pictures-visibility"`
	UseMailservers                      bool                          `json:"use-mailservers?"`
	Usernames                           *json.RawMessage              `json:"usernames,omitempty"`
	WalletRootAddress                   types.Address                 `json:"wallet-root-address,omitempty"`
	WalletSetUpPassed                   bool                          `json:"wallet-set-up-passed?,omitempty"`
	WalletVisibleTokens                 *json.RawMessage              `json:"wallet/visible-tokens,omitempty"`
	WakuBloomFilterMode                 bool                          `json:"waku-bloom-filter-mode,omitempty"`
	WebViewAllowPermissionRequests      bool                          `json:"webview-allow-permission-requests?,omitempty"`
	SendStatusUpdates                   bool                          `json:"send-status-updates?,omitempty"`
	CurrentUserStatus                   *json.RawMessage              `json:"current-user-status"`
	GifRecents                          *json.RawMessage              `json:"gifs/recent-gifs"`
	GifFavorites                        *json.RawMessage              `json:"gifs/favorite-gifs"`
	OpenseaEnabled                      bool                          `json:"opensea-enabled?,omitempty"`
	TelemetryServerURL                  string                        `json:"telemetry-server-url,omitempty"`
	LastBackup                          uint64                        `json:"last-backup,omitempty"`
	BackupEnabled                       bool                          `json:"backup-enabled?,omitempty"`
	AutoMessageEnabled                  bool                          `json:"auto-message-enabled?,omitempty"`
	GifAPIKey                           string                        `json:"gifs/api-key"`
	TestNetworksEnabled                 bool                          `json:"test-networks-enabled?,omitempty"`
	ProfileMigrationNeeded              bool                          `json:"profile-migration-needed,omitempty"`
	IsGoerliEnabled                     bool                          `json:"is-goerli-enabled?,omitempty"`
	TokenGroupByCommunity               bool                          `json:"token-group-by-community?,omitempty"`
	ShowCommunityAssetWhenSendingTokens bool                          `json:"show-community-asset-when-sending-tokens?,omitempty"`
	DisplayAssetsBelowBalance           bool                          `json:"display-assets-below-balance?,omitempty"`
	DisplayAssetsBelowBalanceThreshold  int64                         `json:"display-assets-below-balance-threshold,omitempty"`
	CollectibleGroupByCollection        bool                          `json:"collectible-group-by-collection?,omitempty"`
	CollectibleGroupByCommunity         bool                          `json:"collectible-group-by-community?,omitempty"`
	URLUnfurlingMode                    URLUnfurlingModeType          `json:"url-unfurling-mode,omitempty"`
	PeerSyncingEnabled                  bool                          `json:"peer-syncing-enabled?,omitempty"`
}

func (s Settings) MarshalJSON() ([]byte, error) {
	// We need this typedef in order to overcome stack overflow
	// when marshaling JSON
	type Alias Settings

	ext, err := accountJson.ExtendStructWithPubKeyData(s.PublicKey, Alias(s))
	if err != nil {
		return nil, err
	}
	return json.Marshal(ext)
}

func (s Settings) IsEmpty() bool {
	empty := reflect.Zero(reflect.TypeOf(s)).Interface()
	return reflect.DeepEqual(s, empty)
}

func (s Settings) GetFleet() string {
	if s.Fleet == nil {
		return params.FleetUndefined
	}
	return *s.Fleet
}
