package accounts

import (
	"context"
	"errors"

	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/nodecfg"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/protocol/identity"
)

func NewSettingsAPI(messenger **protocol.Messenger, db *accounts.Database, config *params.NodeConfig) *SettingsAPI {
	return &SettingsAPI{messenger, db, config}
}

// SettingsAPI is class with methods available over RPC.
type SettingsAPI struct {
	messenger **protocol.Messenger
	db        *accounts.Database
	config    *params.NodeConfig
}

func (api *SettingsAPI) SaveSetting(ctx context.Context, typ string, val interface{}) error {
	// NOTE(Ferossgp): v0.62.0 Backward compatibility, skip this for older clients instead of returning error
	if typ == "waku-enabled" {
		return nil
	}

	err := api.db.SaveSetting(typ, val)
	if err != nil {
		return err
	}

	return nil
}

func (api *SettingsAPI) GetSettings(ctx context.Context) (settings.Settings, error) {
	return api.db.GetSettings()
}

// NodeConfig returns the currently used node configuration
func (api *SettingsAPI) NodeConfig(ctx context.Context) (*params.NodeConfig, error) {
	return api.config, nil
}

// Saves the nodeconfig in the database. The node must be restarted for the changes to be applied
func (api *SettingsAPI) SaveNodeConfig(ctx context.Context, n *params.NodeConfig) error {
	return nodecfg.SaveNodeConfig(api.db.DB(), n)
}

// Notifications Settings
func (api *SettingsAPI) NotificationsGetAllowNotifications() (bool, error) {
	return api.db.GetAllowNotifications()
}

func (api *SettingsAPI) NotificationsSetAllowNotifications(value bool) error {
	return api.db.SetAllowNotifications(value)
}

func (api *SettingsAPI) NotificationsGetOneToOneChats() (string, error) {
	return api.db.GetOneToOneChats()
}

func (api *SettingsAPI) NotificationsSetOneToOneChats(value string) error {
	return api.db.SetOneToOneChats(value)
}

func (api *SettingsAPI) NotificationsGetGroupChats() (string, error) {
	return api.db.GetGroupChats()
}

func (api *SettingsAPI) NotificationsSetGroupChats(value string) error {
	return api.db.SetGroupChats(value)
}

func (api *SettingsAPI) NotificationsGetPersonalMentions() (string, error) {
	return api.db.GetPersonalMentions()
}

func (api *SettingsAPI) NotificationsSetPersonalMentions(value string) error {
	return api.db.SetPersonalMentions(value)
}

func (api *SettingsAPI) NotificationsGetGlobalMentions() (string, error) {
	return api.db.GetGlobalMentions()
}

func (api *SettingsAPI) NotificationsSetGlobalMentions(value string) error {
	return api.db.SetGlobalMentions(value)
}

func (api *SettingsAPI) NotificationsGetAllMessages() (string, error) {
	return api.db.GetAllMessages()
}

func (api *SettingsAPI) NotificationsSetAllMessages(value string) error {
	return api.db.SetAllMessages(value)
}

func (api *SettingsAPI) NotificationsGetContactRequests() (string, error) {
	return api.db.GetContactRequests()
}

func (api *SettingsAPI) NotificationsSetContactRequests(value string) error {
	return api.db.SetContactRequests(value)
}

func (api *SettingsAPI) NotificationsGetIdentityVerificationRequests() (string, error) {
	return api.db.GetIdentityVerificationRequests()
}

func (api *SettingsAPI) NotificationsSetIdentityVerificationRequests(value string) error {
	return api.db.SetIdentityVerificationRequests(value)
}

func (api *SettingsAPI) NotificationsGetSoundEnabled() (bool, error) {
	return api.db.GetSoundEnabled()
}

func (api *SettingsAPI) NotificationsSetSoundEnabled(value bool) error {
	return api.db.SetSoundEnabled(value)
}

func (api *SettingsAPI) NotificationsGetVolume() (int, error) {
	return api.db.GetVolume()
}

func (api *SettingsAPI) NotificationsSetVolume(value int) error {
	return api.db.SetVolume(value)
}

func (api *SettingsAPI) NotificationsGetMessagePreview() (int, error) {
	return api.db.GetMessagePreview()
}

func (api *SettingsAPI) NotificationsSetMessagePreview(value int) error {
	return api.db.SetMessagePreview(value)
}

// Notifications Settings - Exemption settings
func (api *SettingsAPI) NotificationsGetExMuteAllMessages(id string) (bool, error) {
	return api.db.GetExMuteAllMessages(id)
}

func (api *SettingsAPI) NotificationsGetExPersonalMentions(id string) (string, error) {
	return api.db.GetExPersonalMentions(id)
}

func (api *SettingsAPI) NotificationsGetExGlobalMentions(id string) (string, error) {
	return api.db.GetExGlobalMentions(id)
}

func (api *SettingsAPI) NotificationsGetExOtherMessages(id string) (string, error) {
	return api.db.GetExOtherMessages(id)
}

func (api *SettingsAPI) NotificationsSetExemptions(id string, muteAllMessages bool, personalMentions string,
	globalMentions string, otherMessages string) error {
	return api.db.SetExemptions(id, muteAllMessages, personalMentions, globalMentions, otherMessages)
}

func (api *SettingsAPI) DeleteExemptions(id string) error {
	return api.db.DeleteExemptions(id)
}

// Deprecated: Use api.go/SetBio instead
func (api *SettingsAPI) SetBio(bio string) error {
	return (*api.messenger).SetBio(bio)
}

// Deprecated: use social links from ProfileShowcasePreferences
func (api *SettingsAPI) GetSocialLinks() (identity.SocialLinks, error) {
	return api.db.GetSocialLinks()
}

// Deprecated: use social links from ProfileShowcasePreferences
func (api *SettingsAPI) AddOrReplaceSocialLinks(links identity.SocialLinks) error {
	for _, link := range links {
		if len(link.Text) == 0 {
			return errors.New("`Text` field of a social link must be set")
		}
		if len(link.URL) == 0 {
			return errors.New("`URL` field of a social link must be set")
		}
	}

	return (*api.messenger).AddOrReplaceSocialLinks(links)
}

func (api *SettingsAPI) MnemonicWasShown() error {
	return api.db.MnemonicWasShown()
}
