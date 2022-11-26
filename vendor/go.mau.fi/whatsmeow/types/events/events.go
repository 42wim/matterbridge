// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package events contains all the events that whatsmeow.Client emits to functions registered with AddEventHandler.
package events

import (
	"fmt"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// QR is emitted after connecting when there's no session data in the device store.
//
// The QR codes are available in the Codes slice. You should render the strings as QR codes one by
// one, switching to the next one whenever enough time has passed. WhatsApp web seems to show the
// first code for 60 seconds and all other codes for 20 seconds.
//
// When the QR code has been scanned and pairing is complete, PairSuccess will be emitted. If you
// run out of codes before scanning, the server will close the websocket, and you will have to
// reconnect to get more codes.
type QR struct {
	Codes []string
}

// PairSuccess is emitted after the QR code has been scanned with the phone and the handshake has
// been completed. Note that this is generally followed by a websocket reconnection, so you should
// wait for the Connected before trying to send anything.
type PairSuccess struct {
	ID           types.JID
	BusinessName string
	Platform     string
}

// PairError is emitted when a pair-success event is received from the server, but finishing the pairing locally fails.
type PairError struct {
	ID           types.JID
	BusinessName string
	Platform     string
	Error        error
}

// QRScannedWithoutMultidevice is emitted when the pairing QR code is scanned, but the phone didn't have multidevice enabled.
// The same QR code can still be scanned after this event, which means the user can just be told to enable multidevice and re-scan the code.
type QRScannedWithoutMultidevice struct{}

// Connected is emitted when the client has successfully connected to the WhatsApp servers
// and is authenticated. The user who the client is authenticated as will be in the device store
// at this point, which is why this event doesn't contain any data.
type Connected struct{}

// KeepAliveTimeout is emitted when the keepalive ping request to WhatsApp web servers times out.
//
// Currently, there's no automatic handling for these, but it's expected that the TCP connection will
// either start working again or notice it's dead on its own eventually. Clients may use this event to
// decide to force a disconnect+reconnect faster.
type KeepAliveTimeout struct {
	ErrorCount  int
	LastSuccess time.Time
}

// KeepAliveRestored is emitted if the keepalive pings start working again after some KeepAliveTimeout events.
// Note that if the websocket disconnects before the pings start working, this event will not be emitted.
type KeepAliveRestored struct{}

// LoggedOut is emitted when the client has been unpaired from the phone.
//
// This can happen while connected (stream:error messages) or right after connecting (connect failure messages).
//
// This will not be emitted when the logout is initiated by this client (using Client.LogOut()).
type LoggedOut struct {
	// OnConnect is true if the event was triggered by a connect failure message.
	// If it's false, the event was triggered by a stream:error message.
	OnConnect bool
	// If OnConnect is true, then this field contains the reason code.
	Reason ConnectFailureReason
}

// StreamReplaced is emitted when the client is disconnected by another client connecting with the same keys.
//
// This can happen if you accidentally start another process with the same session
// or otherwise try to connect twice with the same session.
type StreamReplaced struct{}

// TempBanReason is an error code included in temp ban error events.
type TempBanReason int

const (
	TempBanBlockedByUsers         TempBanReason = 101
	TempBanSentToTooManyPeople    TempBanReason = 102
	TempBanCreatedTooManyGroups   TempBanReason = 103
	TempBanSentTooManySameMessage TempBanReason = 104
	TempBanBroadcastList          TempBanReason = 106
)

var tempBanReasonMessage = map[TempBanReason]string{
	TempBanBlockedByUsers:         "too many people blocked you",
	TempBanSentToTooManyPeople:    "you sent too many messages to people who don't have you in their address books",
	TempBanCreatedTooManyGroups:   "you created too many groups with people who don't have you in their address books",
	TempBanSentTooManySameMessage: "you sent the same message to too many people",
	TempBanBroadcastList:          "you sent too many messages to a broadcast list",
}

// String returns the reason code and a human-readable description of the ban reason.
func (tbr TempBanReason) String() string {
	msg, ok := tempBanReasonMessage[tbr]
	if !ok {
		msg = "you may have violated the terms of service (unknown error)"
	}
	return fmt.Sprintf("%d: %s", int(tbr), msg)
}

// TemporaryBan is emitted when there's a connection failure with the ConnectFailureTempBanned reason code.
type TemporaryBan struct {
	Code   TempBanReason
	Expire time.Time
}

func (tb *TemporaryBan) String() string {
	if tb.Expire.IsZero() {
		return fmt.Sprintf("You've been temporarily banned: %v", tb.Code)
	}
	return fmt.Sprintf("You've been temporarily banned: %v. The ban expires at %v", tb.Code, tb.Expire)
}

// ConnectFailureReason is an error code included in connection failure events.
type ConnectFailureReason int

const (
	ConnectFailureLoggedOut      ConnectFailureReason = 401
	ConnectFailureTempBanned     ConnectFailureReason = 402
	ConnectFailureMainDeviceGone ConnectFailureReason = 403
	ConnectFailureUnknownLogout  ConnectFailureReason = 406

	ConnectFailureClientOutdated ConnectFailureReason = 405
	ConnectFailureBadUserAgent   ConnectFailureReason = 409

	// 400, 500 and 501 are also existing codes, but the meaning is unknown
)

var connectFailureReasonMessage = map[ConnectFailureReason]string{
	ConnectFailureLoggedOut:      "logged out from another device",
	ConnectFailureTempBanned:     "account temporarily banned",
	ConnectFailureMainDeviceGone: "primary device was logged out", // seems to happen for both bans and switching phones
	ConnectFailureUnknownLogout:  "logged out for unknown reason",
	ConnectFailureClientOutdated: "client is out of date",
	ConnectFailureBadUserAgent:   "client user agent was rejected",
}

// IsLoggedOut returns true if the client should delete session data due to this connect failure.
func (cfr ConnectFailureReason) IsLoggedOut() bool {
	return cfr == ConnectFailureLoggedOut || cfr == ConnectFailureMainDeviceGone || cfr == ConnectFailureUnknownLogout
}

// String returns the reason code and a short human-readable description of the error.
func (cfr ConnectFailureReason) String() string {
	msg, ok := connectFailureReasonMessage[cfr]
	if !ok {
		msg = "unknown error"
	}
	return fmt.Sprintf("%d: %s", int(cfr), msg)
}

// ConnectFailure is emitted when the WhatsApp server sends a <failure> node with an unknown reason.
//
// Known reasons are handled internally and emitted as different events (e.g. LoggedOut and TemporaryBan).
type ConnectFailure struct {
	Reason ConnectFailureReason
	Raw    *waBinary.Node
}

// ClientOutdated is emitted when the WhatsApp server rejects the connection with the ConnectFailureClientOutdated code.
type ClientOutdated struct{}

// StreamError is emitted when the WhatsApp server sends a <stream:error> node with an unknown code.
//
// Known codes are handled internally and emitted as different events (e.g. LoggedOut).
type StreamError struct {
	Code string
	Raw  *waBinary.Node
}

// Disconnected is emitted when the websocket is closed by the server.
type Disconnected struct{}

// HistorySync is emitted when the phone has sent a blob of historical messages.
type HistorySync struct {
	Data *waProto.HistorySync
}

// UndecryptableMessage is emitted when receiving a new message that failed to decrypt.
//
// The library will automatically ask the sender to retry. If the sender resends the message,
// and it's decryptable, then it will be emitted as a normal Message event.
//
// The UndecryptableMessage event may also be repeated if the resent message is also undecryptable.
type UndecryptableMessage struct {
	Info types.MessageInfo

	// IsUnavailable is true if the recipient device didn't send a ciphertext to this device at all
	// (as opposed to sending a ciphertext, but the ciphertext not being decryptable).
	IsUnavailable bool
}

// Message is emitted when receiving a new message.
type Message struct {
	Info    types.MessageInfo // Information about the message like the chat and sender IDs
	Message *waProto.Message  // The actual message struct

	IsEphemeral           bool // True if the message was unwrapped from an EphemeralMessage
	IsViewOnce            bool // True if the message was unwrapped from a ViewOnceMessage or ViewOnceMessageV2
	IsViewOnceV2          bool // True if the message was unwrapped from a ViewOnceMessage
	IsDocumentWithCaption bool // True if the message was unwrapped from a DocumentWithCaptionMessage
	IsEdit                bool // True if the message was unwrapped from an EditedMessage

	// The raw message struct. This is the raw unmodified data, which means the actual message might
	// be wrapped in DeviceSentMessage, EphemeralMessage or ViewOnceMessage.
	RawMessage *waProto.Message
}

// UnwrapRaw fills the Message, IsEphemeral and IsViewOnce fields based on the raw message in the RawMessage field.
func (evt *Message) UnwrapRaw() *Message {
	evt.Message = evt.RawMessage
	if evt.Message.GetDeviceSentMessage().GetMessage() != nil {
		evt.Info.DeviceSentMeta = &types.DeviceSentMeta{
			DestinationJID: evt.Message.GetDeviceSentMessage().GetDestinationJid(),
			Phash:          evt.Message.GetDeviceSentMessage().GetPhash(),
		}
		evt.Message = evt.Message.GetDeviceSentMessage().GetMessage()
	}
	if evt.Message.GetEphemeralMessage().GetMessage() != nil {
		evt.Message = evt.Message.GetEphemeralMessage().GetMessage()
		evt.IsEphemeral = true
	}
	if evt.Message.GetViewOnceMessage().GetMessage() != nil {
		evt.Message = evt.Message.GetViewOnceMessage().GetMessage()
		evt.IsViewOnce = true
	}
	if evt.Message.GetViewOnceMessageV2().GetMessage() != nil {
		evt.Message = evt.Message.GetViewOnceMessageV2().GetMessage()
		evt.IsViewOnce = true
		evt.IsViewOnceV2 = true
	}
	if evt.Message.GetDocumentWithCaptionMessage().GetMessage() != nil {
		evt.Message = evt.Message.GetDocumentWithCaptionMessage().GetMessage()
		evt.IsDocumentWithCaption = true
	}
	if evt.Message.GetEditedMessage().GetMessage() != nil {
		evt.Message = evt.Message.GetEditedMessage().GetMessage()
		evt.IsEdit = true
	}
	return evt
}

// ReceiptType represents the type of a Receipt event.
type ReceiptType string

const (
	// ReceiptTypeDelivered means the message was delivered to the device (but the user might not have noticed).
	ReceiptTypeDelivered ReceiptType = ""
	// ReceiptTypeSender is sent by your other devices when a message you sent is delivered to them.
	ReceiptTypeSender ReceiptType = "sender"
	// ReceiptTypeRetry means the message was delivered to the device, but decrypting the message failed.
	ReceiptTypeRetry ReceiptType = "retry"
	// ReceiptTypeRead means the user opened the chat and saw the message.
	ReceiptTypeRead ReceiptType = "read"
	// ReceiptTypeReadSelf means the current user read a message from a different device, and has read receipts disabled in privacy settings.
	ReceiptTypeReadSelf ReceiptType = "read-self"
	// ReceiptTypePlayed means the user opened a view-once media message.
	//
	// This is dispatched for both incoming and outgoing messages when played. If the current user opened the media,
	// it means the media should be removed from all devices. If a recipient opened the media, it's just a notification
	// for the sender that the media was viewed.
	ReceiptTypePlayed ReceiptType = "played"
)

// GoString returns the name of the Go constant for the ReceiptType value.
func (rt ReceiptType) GoString() string {
	switch rt {
	case ReceiptTypeRead:
		return "events.ReceiptTypeRead"
	case ReceiptTypeReadSelf:
		return "events.ReceiptTypeReadSelf"
	case ReceiptTypeDelivered:
		return "events.ReceiptTypeDelivered"
	case ReceiptTypePlayed:
		return "events.ReceiptTypePlayed"
	default:
		return fmt.Sprintf("events.ReceiptType(%#v)", string(rt))
	}
}

// Receipt is emitted when an outgoing message is delivered to or read by another user, or when another device reads an incoming message.
//
// N.B. WhatsApp on Android sends message IDs from newest message to oldest, but WhatsApp on iOS sends them in the opposite order (oldest first).
type Receipt struct {
	types.MessageSource
	MessageIDs []types.MessageID
	Timestamp  time.Time
	Type       ReceiptType
}

// ChatPresence is emitted when a chat state update (also known as typing notification) is received.
//
// Note that WhatsApp won't send you these updates unless you mark yourself as online:
//
//	client.SendPresence(types.PresenceAvailable)
type ChatPresence struct {
	types.MessageSource
	State types.ChatPresence      // The current state, either composing or paused
	Media types.ChatPresenceMedia // When composing, the type of message
}

// Presence is emitted when a presence update is received.
//
// Note that WhatsApp only sends you presence updates for individual users after you subscribe to them:
//
//	client.SubscribePresence(user JID)
type Presence struct {
	// The user whose presence event this is
	From types.JID
	// True if the user is now offline
	Unavailable bool
	// The time when the user was last online. This may be the zero value if the user has hid their last seen time.
	LastSeen time.Time
}

// JoinedGroup is emitted when you join or are added to a group.
type JoinedGroup struct {
	Reason    string          // If the event was triggered by you using an invite link, this will be "invite".
	Type      string          // "new" if it's a newly created group.
	CreateKey types.MessageID // If you created the group, this is the same message ID you passed to CreateGroup.
	types.GroupInfo
}

// GroupInfo is emitted when the metadata of a group changes.
type GroupInfo struct {
	JID       types.JID  // The group ID in question
	Notify    string     // Seems like a top-level type for the invite
	Sender    *types.JID // The user who made the change. Doesn't seem to be present when notify=invite
	Timestamp time.Time  // The time when the change occurred

	Name      *types.GroupName      // Group name change
	Topic     *types.GroupTopic     // Group topic (description) change
	Locked    *types.GroupLocked    // Group locked status change (can only admins edit group info?)
	Announce  *types.GroupAnnounce  // Group announce status change (can only admins send messages?)
	Ephemeral *types.GroupEphemeral // Disappearing messages change

	NewInviteLink *string // Group invite link change

	PrevParticipantVersionID string
	ParticipantVersionID     string

	JoinReason string // This will be "invite" if the user joined via invite link

	Join  []types.JID // Users who joined or were added the group
	Leave []types.JID // Users who left or were removed from the group

	Promote []types.JID // Users who were promoted to admins
	Demote  []types.JID // Users who were demoted to normal users

	UnknownChanges []*waBinary.Node
}

// Picture is emitted when a user's profile picture or group's photo is changed.
//
// You can use Client.GetProfilePictureInfo to get the actual image URL after this event.
type Picture struct {
	JID       types.JID // The user or group ID where the picture was changed.
	Author    types.JID // The user who changed the picture.
	Timestamp time.Time // The timestamp when the picture was changed.
	Remove    bool      // True if the picture was removed.
	PictureID string    // The new picture ID if it was not removed.
}

// IdentityChange is emitted when another user changes their primary device.
type IdentityChange struct {
	JID       types.JID
	Timestamp time.Time

	// Implicit will be set to true if the event was triggered by an untrusted identity error,
	// rather than an identity change notification from the server.
	Implicit bool
}

// PrivacySettings is emitted when the user changes their privacy settings.
type PrivacySettings struct {
	NewSettings         types.PrivacySettings
	GroupAddChanged     bool
	LastSeenChanged     bool
	StatusChanged       bool
	ProfileChanged      bool
	ReadReceiptsChanged bool
}

// OfflineSyncPreview is emitted right after connecting if the server is going to send events that the client missed during downtime.
type OfflineSyncPreview struct {
	Total int

	AppDataChanges int
	Messages       int
	Notifications  int
	Receipts       int
}

// OfflineSyncCompleted is emitted after the server has finished sending missed events.
type OfflineSyncCompleted struct {
	Count int
}

type MediaRetryError struct {
	Code int
}

// MediaRetry is emitted when the phone sends a response to a media retry request.
type MediaRetry struct {
	Ciphertext []byte
	IV         []byte

	// Sometimes there's an unencrypted media retry error. In these cases, Ciphertext and IV will be nil.
	Error *MediaRetryError

	Timestamp time.Time // The time of the response.

	MessageID types.MessageID // The ID of the message.
	ChatID    types.JID       // The chat ID where the message was sent.
	SenderID  types.JID       // The user who sent the message. Only present in groups.
	FromMe    bool            // Whether the message was sent by the current user or someone else.
}
