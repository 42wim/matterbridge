// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const (
	CHANNEL_NOTIFY_DEFAULT              = "default"
	CHANNEL_NOTIFY_ALL                  = "all"
	CHANNEL_NOTIFY_MENTION              = "mention"
	CHANNEL_NOTIFY_NONE                 = "none"
	CHANNEL_MARK_UNREAD_ALL             = "all"
	CHANNEL_MARK_UNREAD_MENTION         = "mention"
	IGNORE_CHANNEL_MENTIONS_DEFAULT     = "default"
	IGNORE_CHANNEL_MENTIONS_OFF         = "off"
	IGNORE_CHANNEL_MENTIONS_ON          = "on"
	IGNORE_CHANNEL_MENTIONS_NOTIFY_PROP = "ignore_channel_mentions"
)

type ChannelUnread struct {
	TeamId           string    `json:"team_id"`
	ChannelId        string    `json:"channel_id"`
	MsgCount         int64     `json:"msg_count"`
	MentionCount     int64     `json:"mention_count"`
	MentionCountRoot int64     `json:"mention_count_root"`
	MsgCountRoot     int64     `json:"msg_count_root"`
	NotifyProps      StringMap `json:"-"`
}

type ChannelUnreadAt struct {
	TeamId           string    `json:"team_id"`
	UserId           string    `json:"user_id"`
	ChannelId        string    `json:"channel_id"`
	MsgCount         int64     `json:"msg_count"`
	MentionCount     int64     `json:"mention_count"`
	MentionCountRoot int64     `json:"mention_count_root"`
	MsgCountRoot     int64     `json:"msg_count_root"`
	LastViewedAt     int64     `json:"last_viewed_at"`
	NotifyProps      StringMap `json:"-"`
}

type ChannelMember struct {
	ChannelId        string    `json:"channel_id"`
	UserId           string    `json:"user_id"`
	Roles            string    `json:"roles"`
	LastViewedAt     int64     `json:"last_viewed_at"`
	MsgCount         int64     `json:"msg_count"`
	MentionCount     int64     `json:"mention_count"`
	MentionCountRoot int64     `json:"mention_count_root"`
	MsgCountRoot     int64     `json:"msg_count_root"`
	NotifyProps      StringMap `json:"notify_props"`
	LastUpdateAt     int64     `json:"last_update_at"`
	SchemeGuest      bool      `json:"scheme_guest"`
	SchemeUser       bool      `json:"scheme_user"`
	SchemeAdmin      bool      `json:"scheme_admin"`
	ExplicitRoles    string    `json:"explicit_roles"`
}

type ChannelMembers []ChannelMember

type ChannelMemberForExport struct {
	ChannelMember
	ChannelName string
	Username    string
}

func (o *ChannelMembers) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func (o *ChannelUnread) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *ChannelUnreadAt) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func ChannelMembersFromJson(data io.Reader) *ChannelMembers {
	var o *ChannelMembers
	json.NewDecoder(data).Decode(&o)
	return o
}

func ChannelUnreadFromJson(data io.Reader) *ChannelUnread {
	var o *ChannelUnread
	json.NewDecoder(data).Decode(&o)
	return o
}

func ChannelUnreadAtFromJson(data io.Reader) *ChannelUnreadAt {
	var o *ChannelUnreadAt
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *ChannelMember) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func ChannelMemberFromJson(data io.Reader) *ChannelMember {
	var o *ChannelMember
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *ChannelMember) IsValid() *AppError {

	if !IsValidId(o.ChannelId) {
		return NewAppError("ChannelMember.IsValid", "model.channel_member.is_valid.channel_id.app_error", nil, "", http.StatusBadRequest)
	}

	if !IsValidId(o.UserId) {
		return NewAppError("ChannelMember.IsValid", "model.channel_member.is_valid.user_id.app_error", nil, "", http.StatusBadRequest)
	}

	notifyLevel := o.NotifyProps[DESKTOP_NOTIFY_PROP]
	if len(notifyLevel) > 20 || !IsChannelNotifyLevelValid(notifyLevel) {
		return NewAppError("ChannelMember.IsValid", "model.channel_member.is_valid.notify_level.app_error", nil, "notify_level="+notifyLevel, http.StatusBadRequest)
	}

	markUnreadLevel := o.NotifyProps[MARK_UNREAD_NOTIFY_PROP]
	if len(markUnreadLevel) > 20 || !IsChannelMarkUnreadLevelValid(markUnreadLevel) {
		return NewAppError("ChannelMember.IsValid", "model.channel_member.is_valid.unread_level.app_error", nil, "mark_unread_level="+markUnreadLevel, http.StatusBadRequest)
	}

	if pushLevel, ok := o.NotifyProps[PUSH_NOTIFY_PROP]; ok {
		if len(pushLevel) > 20 || !IsChannelNotifyLevelValid(pushLevel) {
			return NewAppError("ChannelMember.IsValid", "model.channel_member.is_valid.push_level.app_error", nil, "push_notification_level="+pushLevel, http.StatusBadRequest)
		}
	}

	if sendEmail, ok := o.NotifyProps[EMAIL_NOTIFY_PROP]; ok {
		if len(sendEmail) > 20 || !IsSendEmailValid(sendEmail) {
			return NewAppError("ChannelMember.IsValid", "model.channel_member.is_valid.email_value.app_error", nil, "push_notification_level="+sendEmail, http.StatusBadRequest)
		}
	}

	if ignoreChannelMentions, ok := o.NotifyProps[IGNORE_CHANNEL_MENTIONS_NOTIFY_PROP]; ok {
		if len(ignoreChannelMentions) > 40 || !IsIgnoreChannelMentionsValid(ignoreChannelMentions) {
			return NewAppError("ChannelMember.IsValid", "model.channel_member.is_valid.ignore_channel_mentions_value.app_error", nil, "ignore_channel_mentions="+ignoreChannelMentions, http.StatusBadRequest)
		}
	}

	return nil
}

func (o *ChannelMember) PreSave() {
	o.LastUpdateAt = GetMillis()
}

func (o *ChannelMember) PreUpdate() {
	o.LastUpdateAt = GetMillis()
}

func (o *ChannelMember) GetRoles() []string {
	return strings.Fields(o.Roles)
}

func (o *ChannelMember) SetChannelMuted(muted bool) {
	if o.IsChannelMuted() {
		o.NotifyProps[MARK_UNREAD_NOTIFY_PROP] = CHANNEL_MARK_UNREAD_ALL
	} else {
		o.NotifyProps[MARK_UNREAD_NOTIFY_PROP] = CHANNEL_MARK_UNREAD_MENTION
	}
}

func (o *ChannelMember) IsChannelMuted() bool {
	return o.NotifyProps[MARK_UNREAD_NOTIFY_PROP] == CHANNEL_MARK_UNREAD_MENTION
}

func IsChannelNotifyLevelValid(notifyLevel string) bool {
	return notifyLevel == CHANNEL_NOTIFY_DEFAULT ||
		notifyLevel == CHANNEL_NOTIFY_ALL ||
		notifyLevel == CHANNEL_NOTIFY_MENTION ||
		notifyLevel == CHANNEL_NOTIFY_NONE
}

func IsChannelMarkUnreadLevelValid(markUnreadLevel string) bool {
	return markUnreadLevel == CHANNEL_MARK_UNREAD_ALL || markUnreadLevel == CHANNEL_MARK_UNREAD_MENTION
}

func IsSendEmailValid(sendEmail string) bool {
	return sendEmail == CHANNEL_NOTIFY_DEFAULT || sendEmail == "true" || sendEmail == "false"
}

func IsIgnoreChannelMentionsValid(ignoreChannelMentions string) bool {
	return ignoreChannelMentions == IGNORE_CHANNEL_MENTIONS_ON || ignoreChannelMentions == IGNORE_CHANNEL_MENTIONS_OFF || ignoreChannelMentions == IGNORE_CHANNEL_MENTIONS_DEFAULT
}

func GetDefaultChannelNotifyProps() StringMap {
	return StringMap{
		DESKTOP_NOTIFY_PROP:                 CHANNEL_NOTIFY_DEFAULT,
		MARK_UNREAD_NOTIFY_PROP:             CHANNEL_MARK_UNREAD_ALL,
		PUSH_NOTIFY_PROP:                    CHANNEL_NOTIFY_DEFAULT,
		EMAIL_NOTIFY_PROP:                   CHANNEL_NOTIFY_DEFAULT,
		IGNORE_CHANNEL_MENTIONS_NOTIFY_PROP: IGNORE_CHANNEL_MENTIONS_DEFAULT,
	}
}
