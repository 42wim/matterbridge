package settings

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/sqlite"
)

var (
	ErrTypeAssertionFailed = errors.New("type assertion of interface value failed")
)

func buildRawSyncSettingMessage(msg *protobuf.SyncSetting, chatID string) (*common.RawMessage, error) {
	encodedMessage, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &common.RawMessage{
		LocalChatID:         chatID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_SETTING,
		ResendAutomatically: true,
	}, nil
}

// Currency

func buildRawCurrencySyncMessage(v string, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_CURRENCY,
		Value: &protobuf.SyncSetting_ValueString{ValueString: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func currencyProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertString(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawCurrencySyncMessage(v, clock, chatID)
}

func currencyProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawCurrencySyncMessage(s.Currency, clock, chatID)
}

// GifFavorites

func buildRawGifFavoritesSyncMessage(v []byte, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_GIF_FAVOURITES,
		Value: &protobuf.SyncSetting_ValueBytes{ValueBytes: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func gifFavouritesProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBytes(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawGifFavoritesSyncMessage(v, clock, chatID)
}

func gifFavouritesProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	gf := extractJSONRawMessage(s.GifFavorites)
	return buildRawGifFavoritesSyncMessage(gf, clock, chatID)
}

// GifRecents

func buildRawGifRecentsSyncMessage(v []byte, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_GIF_RECENTS,
		Value: &protobuf.SyncSetting_ValueBytes{ValueBytes: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func gifRecentsProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBytes(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawGifRecentsSyncMessage(v, clock, chatID)
}

func gifRecentsProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	gr := extractJSONRawMessage(s.GifRecents)
	return buildRawGifRecentsSyncMessage(gr, clock, chatID)
}

// MessagesFromContactsOnly

func buildRawMessagesFromContactsOnlySyncMessage(v bool, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_MESSAGES_FROM_CONTACTS_ONLY,
		Value: &protobuf.SyncSetting_ValueBool{ValueBool: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func messagesFromContactsOnlyProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBool(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawMessagesFromContactsOnlySyncMessage(v, clock, chatID)
}

func messagesFromContactsOnlyProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawMessagesFromContactsOnlySyncMessage(s.MessagesFromContactsOnly, clock, chatID)
}

// PreferredName

func buildRawPreferredNameSyncMessage(v string, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_PREFERRED_NAME,
		Value: &protobuf.SyncSetting_ValueString{ValueString: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func preferredNameProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertString(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawPreferredNameSyncMessage(v, clock, chatID)
}

func preferredNameProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	var pn string
	if s.PreferredName != nil {
		pn = *s.PreferredName
	}

	return buildRawPreferredNameSyncMessage(pn, clock, chatID)
}

// PreviewPrivacy

func buildRawPreviewPrivacySyncMessage(v bool, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_PREVIEW_PRIVACY,
		Value: &protobuf.SyncSetting_ValueBool{ValueBool: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func previewPrivacyProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBool(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawPreviewPrivacySyncMessage(v, clock, chatID)
}

func previewPrivacyProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawPreviewPrivacySyncMessage(s.PreviewPrivacy, clock, chatID)
}

// ProfilePicturesShowTo

func buildRawProfilePicturesShowToSyncMessage(v int64, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_PROFILE_PICTURES_SHOW_TO,
		Value: &protobuf.SyncSetting_ValueInt64{ValueInt64: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func profilePicturesShowToProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := parseNumberToInt64(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawProfilePicturesShowToSyncMessage(v, clock, chatID)
}

func profilePicturesShowToProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawProfilePicturesShowToSyncMessage(int64(s.ProfilePicturesShowTo), clock, chatID)
}

// ProfilePicturesVisibility

func buildRawProfilePicturesVisibilitySyncMessage(v int64, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_PROFILE_PICTURES_VISIBILITY,
		Value: &protobuf.SyncSetting_ValueInt64{ValueInt64: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func profilePicturesVisibilityProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := parseNumberToInt64(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawProfilePicturesVisibilitySyncMessage(v, clock, chatID)
}

func profilePicturesVisibilityProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawProfilePicturesVisibilitySyncMessage(int64(s.ProfilePicturesVisibility), clock, chatID)
}

// SendStatusUpdates

func buildRawSendStatusUpdatesSyncMessage(v bool, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_SEND_STATUS_UPDATES,
		Value: &protobuf.SyncSetting_ValueBool{ValueBool: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func sendStatusUpdatesProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBool(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawSendStatusUpdatesSyncMessage(v, clock, chatID)
}

func sendStatusUpdatesProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawSendStatusUpdatesSyncMessage(s.SendStatusUpdates, clock, chatID)
}

// StickerPacksInstalled

func buildRawStickerPacksInstalledSyncMessage(v []byte, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_STICKERS_PACKS_INSTALLED,
		Value: &protobuf.SyncSetting_ValueBytes{ValueBytes: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func stickersPacksInstalledProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := parseJSONBlobData(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawStickerPacksInstalledSyncMessage(v, clock, chatID)
}

func stickersPacksInstalledProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	spi := extractJSONRawMessage(s.StickerPacksInstalled)
	return buildRawStickerPacksInstalledSyncMessage(spi, clock, chatID)
}

// StickerPacksPending

func buildRawStickerPacksPendingSyncMessage(v []byte, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_STICKERS_PACKS_PENDING,
		Value: &protobuf.SyncSetting_ValueBytes{ValueBytes: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func stickersPacksPendingProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := parseJSONBlobData(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawStickerPacksPendingSyncMessage(v, clock, chatID)
}

func stickersPacksPendingProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	spp := extractJSONRawMessage(s.StickerPacksPending)
	return buildRawStickerPacksPendingSyncMessage(spp, clock, chatID)
}

// StickersRecentStickers

func buildRawStickersRecentStickersSyncMessage(v []byte, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_STICKERS_RECENT_STICKERS,
		Value: &protobuf.SyncSetting_ValueBytes{ValueBytes: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func stickersRecentStickersProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := parseJSONBlobData(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawStickersRecentStickersSyncMessage(v, clock, chatID)
}

func stickersRecentStickersProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	srs := extractJSONRawMessage(s.StickersRecentStickers)
	return buildRawStickersRecentStickersSyncMessage(srs, clock, chatID)
}

// Helpers

func assertBytes(value interface{}) ([]byte, error) {
	v, ok := value.([]byte)
	if !ok {
		return nil, errors.Wrapf(ErrTypeAssertionFailed, "expected '[]byte', received %T", value)
	}
	return v, nil
}

func assertBool(value interface{}) (bool, error) {
	v, ok := value.(bool)
	if !ok {
		return false, errors.Wrapf(ErrTypeAssertionFailed, "expected 'bool', received %T", value)
	}
	return v, nil
}

func assertString(value interface{}) (string, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		value = *value.(*string)
	}
	v, ok := value.(string)
	if !ok {
		return "", errors.Wrapf(ErrTypeAssertionFailed, "expected 'string', received %T", value)
	}
	return v, nil
}

func parseJSONBlobData(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case *sqlite.JSONBlob:
		return extractJSONBlob(v)
	default:
		return nil, errors.Wrapf(ErrTypeAssertionFailed, "expected []byte or *sqlite.JSONBlob, received %T", value)
	}
}

func parseNumberToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case ProfilePicturesShowToType:
		return int64(v), nil
	case ProfilePicturesVisibilityType:
		return int64(v), nil
	case URLUnfurlingModeType:
		return int64(v), nil
	default:
		return 0, errors.Wrapf(ErrTypeAssertionFailed, "expected a numeric type, received %T", value)
	}
}

func extractJSONBlob(jb *sqlite.JSONBlob) ([]byte, error) {
	value, err := jb.Value()
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	return value.([]byte), nil
}

func extractJSONRawMessage(jrm *json.RawMessage) []byte {
	if jrm == nil {
		return nil
	}
	out, _ := jrm.MarshalJSON() // Don't need to parse error because it is always nil
	if len(out) == 0 || bytes.Equal(out, []byte("null")) {
		return nil
	}
	return out
}

// DisplayName

func buildRawDisplayNameSyncMessage(v string, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_DISPLAY_NAME,
		Value: &protobuf.SyncSetting_ValueString{ValueString: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func displayNameProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertString(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawDisplayNameSyncMessage(v, clock, chatID)
}

func displayNameProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {

	return buildRawDisplayNameSyncMessage(s.DisplayName, clock, chatID)
}

// Bio

func buildRawBioSyncMessage(v string, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_BIO,
		Value: &protobuf.SyncSetting_ValueString{ValueString: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func bioProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertString(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawBioSyncMessage(v, clock, chatID)
}

func bioProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawBioSyncMessage(s.Bio, clock, chatID)
}

// MnemonicRemoved

func buildRawMnemonicRemovedSyncMessage(v bool, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_MNEMONIC_REMOVED,
		Value: &protobuf.SyncSetting_ValueBool{ValueBool: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func mnemonicRemovedProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBool(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawMnemonicRemovedSyncMessage(v, clock, chatID)
}

func mnemonicRemovedProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawMnemonicRemovedSyncMessage(s.MnemonicRemoved, clock, chatID)
}

// UrlUnfurlingMode

func buildRawURLUnfurlingModeSyncMessage(v int64, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_URL_UNFURLING_MODE,
		Value: &protobuf.SyncSetting_ValueInt64{ValueInt64: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func urlUnfurlingModeProtobufFactory(value any, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := parseNumberToInt64(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawURLUnfurlingModeSyncMessage(v, clock, chatID)
}

func urlUnfurlingModeProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawURLUnfurlingModeSyncMessage(int64(s.URLUnfurlingMode), clock, chatID)
}

// ShowCommunityAssetWhenSendingTokens

func buildRawShowCommunityAssetWhenSendingTokensSyncMessage(v bool, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_SHOW_COMMUNITY_ASSET_WHEN_SENDING_TOKENS,
		Value: &protobuf.SyncSetting_ValueBool{ValueBool: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func showCommunityAssetWhenSendingTokensProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBool(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawShowCommunityAssetWhenSendingTokensSyncMessage(v, clock, chatID)
}

func showCommunityAssetWhenSendingTokensProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawShowCommunityAssetWhenSendingTokensSyncMessage(s.ShowCommunityAssetWhenSendingTokens, clock, chatID)
}

// DisplayAssetsBelowBalance

func buildRawDisplayAssetsBelowBalanceSyncMessage(v bool, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_DISPLAY_ASSETS_BELOW_BALANCE,
		Value: &protobuf.SyncSetting_ValueBool{ValueBool: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func displayAssetsBelowBalanceProtobufFactory(value interface{}, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := assertBool(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawDisplayAssetsBelowBalanceSyncMessage(v, clock, chatID)
}

func displayAssetsBelowBalanceProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawDisplayAssetsBelowBalanceSyncMessage(s.DisplayAssetsBelowBalance, clock, chatID)
}

// DisplayAssetsBelowBalanceThreshold

func buildRawDisplayAssetsBelowBalanceThresholdSyncMessage(v int64, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	pb := &protobuf.SyncSetting{
		Type:  protobuf.SyncSetting_DISPLAY_ASSETS_BELOW_BALANCE_THRESHOLD,
		Value: &protobuf.SyncSetting_ValueInt64{ValueInt64: v},
		Clock: clock,
	}
	rm, err := buildRawSyncSettingMessage(pb, chatID)
	return rm, pb, err
}

func displayAssetsBelowBalanceThresholdProtobufFactory(value any, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	v, err := parseNumberToInt64(value)
	if err != nil {
		return nil, nil, err
	}

	return buildRawDisplayAssetsBelowBalanceThresholdSyncMessage(v, clock, chatID)
}

func displayAssetsBelowBalanceThresholdProtobufFactoryStruct(s Settings, clock uint64, chatID string) (*common.RawMessage, *protobuf.SyncSetting, error) {
	return buildRawDisplayAssetsBelowBalanceThresholdSyncMessage(s.DisplayAssetsBelowBalanceThreshold, clock, chatID)
}
