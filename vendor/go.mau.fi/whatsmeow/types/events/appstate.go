// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package events

import (
	"time"

	"go.mau.fi/whatsmeow/appstate"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// Contact is emitted when an entry in the user's contact list is modified from another device.
type Contact struct {
	JID       types.JID // The contact who was modified.
	Timestamp time.Time // The time when the modification happened.'

	Action *waProto.ContactAction // The new contact info.
}

// PushName is emitted when a message is received with a different push name than the previous value cached for the same user.
type PushName struct {
	JID         types.JID          // The user whose push name changed.
	Message     *types.MessageInfo // The message where this change was first noticed.
	OldPushName string             // The previous push name from the local cache.
	NewPushName string             // The new push name that was included in the message.
}

// BusinessName is emitted when a message is received with a different verified business name than the previous value cached for the same user.
type BusinessName struct {
	JID             types.JID
	Message         *types.MessageInfo // This is only present if the change was detected in a message.
	OldBusinessName string
	NewBusinessName string
}

// Pin is emitted when a chat is pinned or unpinned from another device.
type Pin struct {
	JID       types.JID // The chat which was pinned or unpinned.
	Timestamp time.Time // The time when the (un)pinning happened.

	Action *waProto.PinAction // Whether the chat is now pinned or not.
}

// Star is emitted when a message is starred or unstarred from another device.
type Star struct {
	ChatJID   types.JID // The chat where the message was pinned.
	SenderJID types.JID // In group chats, the user who sent the message (except if the message was sent by the user).
	IsFromMe  bool      // Whether the message was sent by the user.
	MessageID string    // The message which was starred or unstarred.
	Timestamp time.Time // The time when the (un)starring happened.

	Action *waProto.StarAction // Whether the message is now starred or not.
}

// DeleteForMe is emitted when a message is deleted (for the current user only) from another device.
type DeleteForMe struct {
	ChatJID   types.JID // The chat where the message was deleted.
	SenderJID types.JID // In group chats, the user who sent the message (except if the message was sent by the user).
	IsFromMe  bool      // Whether the message was sent by the user.
	MessageID string    // The message which was deleted.
	Timestamp time.Time // The time when the deletion happened.

	Action *waProto.DeleteMessageForMeAction // Additional information for the deletion.
}

// Mute is emitted when a chat is muted or unmuted from another device.
type Mute struct {
	JID       types.JID // The chat which was muted or unmuted.
	Timestamp time.Time // The time when the (un)muting happened.

	Action *waProto.MuteAction // The current mute status of the chat.
}

// Archive is emitted when a chat is archived or unarchived from another device.
type Archive struct {
	JID       types.JID // The chat which was archived or unarchived.
	Timestamp time.Time // The time when the (un)archiving happened.

	Action *waProto.ArchiveChatAction // The current archival status of the chat.
}

// MarkChatAsRead is emitted when a whole chat is marked as read or unread from another device.
type MarkChatAsRead struct {
	JID       types.JID // The chat which was marked as read or unread.
	Timestamp time.Time // The time when the marking happened.

	Action *waProto.MarkChatAsReadAction // Whether the chat was marked as read or unread, and info about the most recent messages.
}

// DeleteChat is emitted when a chat is deleted on another device.
type DeleteChat struct {
	JID       types.JID // The chat which was deleted.
	Timestamp time.Time // The time when the deletion happened.

	Action *waProto.DeleteChatAction // Information about the deletion.
}

// PushNameSetting is emitted when the user's push name is changed from another device.
type PushNameSetting struct {
	Timestamp time.Time // The time when the push name was changed.

	Action *waProto.PushNameSetting // The new push name for the user.
}

// UnarchiveChatsSetting is emitted when the user changes the "Keep chats archived" setting from another device.
type UnarchiveChatsSetting struct {
	Timestamp time.Time // The time when the setting was changed.

	Action *waProto.UnarchiveChatsSetting // The new settings.
}

// AppState is emitted directly for new data received from app state syncing.
// You should generally use the higher-level events like events.Contact and events.Mute.
type AppState struct {
	Index []string
	*waProto.SyncActionValue
}

// AppStateSyncComplete is emitted when app state is resynced.
type AppStateSyncComplete struct {
	Name appstate.WAPatchName
}
