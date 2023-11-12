package notificationssettings

import (
	"database/sql"
	"fmt"
)

type settingType int

const (
	stringType settingType = iota
	integerType
	boolType
)

// Columns' names
type column string

const (
	columnTextValue          column = "text_value"
	columnIntValue           column = "int_value"
	columnBoolValue          column = "bool_value"
	columnExMuteAllMessages  column = "ex_mute_all_messages"
	columnExPersonalMentions column = "ex_personal_mentions"
	columnExGlobalMentions   column = "ex_global_mentions"
	columnExOtherMessages    column = "ex_other_messages"
)

// Static ids we use for notifications settings
const (
	keyAllowNotifications           = "AllowNotifications"
	keyOneToOneChats                = "OneToOneChats"
	keyGroupChats                   = "GroupChats"
	keyPersonalMentions             = "PersonalMentions"
	keyGlobalMentions               = "GlobalMentions"
	keyAllMessages                  = "AllMessages"
	keyContactRequests              = "ContactRequests"
	keyIdentityVerificationRequests = "IdentityVerificationRequests"
	keySoundEnabled                 = "SoundEnabled"
	keyVolume                       = "Volume"
	keyMessagePreview               = "MessagePreview"
)

// Possible values
const (
	valueSendAlerts = "SendAlerts"
	//valueDeliverQuietly = "DeliverQuietly"  // currently unused
	valueTurnOff = "TurnOff"
)

// Default values
const (
	defaultAllowNotificationsValue           = true
	defaultOneToOneChatsValue                = valueSendAlerts
	defaultGroupChatsValue                   = valueSendAlerts
	defaultPersonalMentionsValue             = valueSendAlerts
	defaultGlobalMentionsValue               = valueSendAlerts
	defaultAllMessagesValue                  = valueTurnOff
	defaultContactRequestsValue              = valueSendAlerts
	defaultIdentityVerificationRequestsValue = valueSendAlerts
	defaultSoundEnabledValue                 = true
	defaultVolumeValue                       = 50
	defaultMessagePreviewValue               = 2
	defaultExMuteAllMessagesValue            = false
	defaultExPersonalMentionsValue           = valueSendAlerts
	defaultExGlobalMentionsValue             = valueSendAlerts
	defaultExOtherMessagesValue              = valueTurnOff
)

type NotificationsSettings struct {
	db *sql.DB
}

func NewNotificationsSettings(db *sql.DB) *NotificationsSettings {
	return &NotificationsSettings{
		db: db,
	}
}

func (ns *NotificationsSettings) buildSelectQuery(column column) string {
	return fmt.Sprintf("SELECT %s FROM notifications_settings WHERE id = ?", column)
}

func (ns *NotificationsSettings) buildInsertOrUpdateQuery(isExemption bool, settingType settingType) string {
	var values string
	if isExemption {
		values = "VALUES(?, 1, NULL, NULL, NULL, ?, ?, ?, ?)"
	} else {
		switch settingType {
		case stringType:
			values = "VALUES(?, 0, ?, NULL, NULL, NULL, NULL, NULL, NULL)"
		case integerType:
			values = "VALUES(?, 0, NULL, ?, NULL, NULL, NULL, NULL, NULL)"
		case boolType:
			values = "VALUES(?, 0, NULL, NULL, ?, NULL, NULL, NULL, NULL)"
		}
	}

	return fmt.Sprintf(`INSERT INTO notifications_settings (
		id, 
		exemption, 
		text_value, 
		int_value, 
		bool_value, 
		ex_mute_all_messages, 
		ex_personal_mentions,
		ex_global_mentions,
		ex_other_messages) 
		%s;`, values)
}

// Non exemption settings
func (ns *NotificationsSettings) GetAllowNotifications() (result bool, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnBoolValue), keyAllowNotifications).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultAllowNotificationsValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetAllowNotifications(value bool) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, boolType), keyAllowNotifications, value, value)
	return err
}

func (ns *NotificationsSettings) GetOneToOneChats() (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnTextValue), keyOneToOneChats).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultOneToOneChatsValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetOneToOneChats(value string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, stringType), keyOneToOneChats, value, value)
	return err
}

func (ns *NotificationsSettings) GetGroupChats() (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnTextValue), keyGroupChats).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultGroupChatsValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetGroupChats(value string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, stringType), keyGroupChats, value, value)
	return err
}

func (ns *NotificationsSettings) GetPersonalMentions() (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnTextValue), keyPersonalMentions).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultPersonalMentionsValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetPersonalMentions(value string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, stringType), keyPersonalMentions, value, value)
	return err
}

func (ns *NotificationsSettings) GetGlobalMentions() (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnTextValue), keyGlobalMentions).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultGlobalMentionsValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetGlobalMentions(value string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, stringType), keyGlobalMentions, value, value)
	return err
}

func (ns *NotificationsSettings) GetAllMessages() (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnTextValue), keyAllMessages).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultAllMessagesValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetAllMessages(value string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, stringType), keyAllMessages, value, value)
	return err
}

func (ns *NotificationsSettings) GetContactRequests() (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnTextValue), keyContactRequests).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultContactRequestsValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetContactRequests(value string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, stringType), keyContactRequests, value, value)
	return err
}

func (ns *NotificationsSettings) GetIdentityVerificationRequests() (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnTextValue), keyIdentityVerificationRequests).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultIdentityVerificationRequestsValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetIdentityVerificationRequests(value string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, stringType), keyIdentityVerificationRequests, value, value)
	return err
}

func (ns *NotificationsSettings) GetSoundEnabled() (result bool, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnBoolValue), keySoundEnabled).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultSoundEnabledValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetSoundEnabled(value bool) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, boolType), keySoundEnabled, value, value)
	return err
}

func (ns *NotificationsSettings) GetVolume() (result int, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnIntValue), keyVolume).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultVolumeValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetVolume(value int) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, integerType), keyVolume, value, value)
	return err
}

func (ns *NotificationsSettings) GetMessagePreview() (result int, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnIntValue), keyMessagePreview).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultMessagePreviewValue, err
	}
	return result, err
}

func (ns *NotificationsSettings) SetMessagePreview(value int) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(false, integerType), keyMessagePreview, value, value)
	return err
}

// Exemption settings
func (ns *NotificationsSettings) GetExMuteAllMessages(id string) (result bool, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnExMuteAllMessages), id).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultExMuteAllMessagesValue, nil
	}
	return result, err
}

func (ns *NotificationsSettings) GetExPersonalMentions(id string) (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnExPersonalMentions), id).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultExPersonalMentionsValue, nil
	}
	return result, err
}

func (ns *NotificationsSettings) GetExGlobalMentions(id string) (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnExGlobalMentions), id).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultExGlobalMentionsValue, nil
	}
	return result, err
}

func (ns *NotificationsSettings) GetExOtherMessages(id string) (result string, err error) {
	err = ns.db.QueryRow(ns.buildSelectQuery(columnExOtherMessages), id).Scan(&result)
	if err != nil && err == sql.ErrNoRows {
		return defaultExOtherMessagesValue, nil
	}
	return result, err
}

func (ns *NotificationsSettings) SetExemptions(id string, muteAllMessages bool, personalMentions string,
	globalMentions string, otherMessages string) error {
	_, err := ns.db.Exec(ns.buildInsertOrUpdateQuery(true, stringType), id, muteAllMessages, personalMentions, globalMentions,
		otherMessages)
	return err
}

func (ns *NotificationsSettings) DeleteExemptions(id string) error {
	switch id {
	case
		keyAllowNotifications,
		keyOneToOneChats,
		keyGroupChats,
		keyPersonalMentions,
		keyGlobalMentions,
		keyAllMessages,
		keyContactRequests,
		keyIdentityVerificationRequests,
		keySoundEnabled,
		keyVolume,
		keyMessagePreview:
		return fmt.Errorf("%s, static notifications settings cannot be deleted", id)
	}

	_, err := ns.db.Exec("DELETE FROM notifications_settings WHERE id = ?", id)
	return err
}
