package gumble

import (
	"layeh.com/gumble/gumble/MumbleProto"
)

// EventListener is the interface that must be implemented by a type if it
// wishes to be notified of Client events.
//
// Listener methods are executed synchronously as event happen. They also block
// network reads from happening until all handlers for an event are called.
// Therefore, it is not recommended to do any long processing from inside of
// these methods.
type EventListener interface {
	OnConnect(e *ConnectEvent)
	OnDisconnect(e *DisconnectEvent)
	OnTextMessage(e *TextMessageEvent)
	OnUserChange(e *UserChangeEvent)
	OnChannelChange(e *ChannelChangeEvent)
	OnPermissionDenied(e *PermissionDeniedEvent)
	OnUserList(e *UserListEvent)
	OnACL(e *ACLEvent)
	OnBanList(e *BanListEvent)
	OnContextActionChange(e *ContextActionChangeEvent)
	OnServerConfig(e *ServerConfigEvent)
}

// ConnectEvent is the event that is passed to EventListener.OnConnect.
type ConnectEvent struct {
	Client         *Client
	WelcomeMessage *string
	MaximumBitrate *int
}

// DisconnectType specifies why a Client disconnected from a server.
type DisconnectType int

// Client disconnect reasons.
const (
	DisconnectError DisconnectType = iota + 1
	DisconnectKicked
	DisconnectBanned
	DisconnectUser
)

// Has returns true if the DisconnectType has changeType part of its bitmask.
func (d DisconnectType) Has(changeType DisconnectType) bool {
	return d&changeType == changeType
}

// DisconnectEvent is the event that is passed to EventListener.OnDisconnect.
type DisconnectEvent struct {
	Client *Client
	Type   DisconnectType

	String string
}

// TextMessageEvent is the event that is passed to EventListener.OnTextMessage.
type TextMessageEvent struct {
	Client *Client
	TextMessage
}

// UserChangeType is a bitmask of items that changed for a user.
type UserChangeType int

// User change items.
const (
	UserChangeConnected UserChangeType = 1 << iota
	UserChangeDisconnected
	UserChangeKicked
	UserChangeBanned
	UserChangeRegistered
	UserChangeUnregistered
	UserChangeName
	UserChangeChannel
	UserChangeComment
	UserChangeAudio
	UserChangeTexture
	UserChangePrioritySpeaker
	UserChangeRecording
	UserChangeStats
)

// Has returns true if the UserChangeType has changeType part of its bitmask.
func (u UserChangeType) Has(changeType UserChangeType) bool {
	return u&changeType == changeType
}

// UserChangeEvent is the event that is passed to EventListener.OnUserChange.
type UserChangeEvent struct {
	Client *Client
	Type   UserChangeType
	User   *User
	Actor  *User

	String string
}

// ChannelChangeType is a bitmask of items that changed for a channel.
type ChannelChangeType int

// Channel change items.
const (
	ChannelChangeCreated ChannelChangeType = 1 << iota
	ChannelChangeRemoved
	ChannelChangeMoved
	ChannelChangeName
	ChannelChangeLinks
	ChannelChangeDescription
	ChannelChangePosition
	ChannelChangePermission
	ChannelChangeMaxUsers
)

// Has returns true if the ChannelChangeType has changeType part of its
// bitmask.
func (c ChannelChangeType) Has(changeType ChannelChangeType) bool {
	return c&changeType == changeType
}

// ChannelChangeEvent is the event that is passed to
// EventListener.OnChannelChange.
type ChannelChangeEvent struct {
	Client  *Client
	Type    ChannelChangeType
	Channel *Channel
}

// PermissionDeniedType specifies why a Client was denied permission to perform
// a particular action.
type PermissionDeniedType int

// Permission denied types.
const (
	PermissionDeniedOther              PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_Text)
	PermissionDeniedPermission         PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_Permission)
	PermissionDeniedSuperUser          PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_SuperUser)
	PermissionDeniedInvalidChannelName PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_ChannelName)
	PermissionDeniedTextTooLong        PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_TextTooLong)
	PermissionDeniedTemporaryChannel   PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_TemporaryChannel)
	PermissionDeniedMissingCertificate PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_MissingCertificate)
	PermissionDeniedInvalidUserName    PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_UserName)
	PermissionDeniedChannelFull        PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_ChannelFull)
	PermissionDeniedNestingLimit       PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_NestingLimit)
	PermissionDeniedChannelCountLimit  PermissionDeniedType = PermissionDeniedType(MumbleProto.PermissionDenied_ChannelCountLimit)
)

// Has returns true if the PermissionDeniedType has changeType part of its
// bitmask.
func (p PermissionDeniedType) Has(changeType PermissionDeniedType) bool {
	return p&changeType == changeType
}

// PermissionDeniedEvent is the event that is passed to
// EventListener.OnPermissionDenied.
type PermissionDeniedEvent struct {
	Client  *Client
	Type    PermissionDeniedType
	Channel *Channel
	User    *User

	Permission Permission
	String     string
}

// UserListEvent is the event that is passed to EventListener.OnUserList.
type UserListEvent struct {
	Client   *Client
	UserList RegisteredUsers
}

// ACLEvent is the event that is passed to EventListener.OnACL.
type ACLEvent struct {
	Client *Client
	ACL    *ACL
}

// BanListEvent is the event that is passed to EventListener.OnBanList.
type BanListEvent struct {
	Client  *Client
	BanList BanList
}

// ContextActionChangeType specifies how a ContextAction changed.
type ContextActionChangeType int

// ContextAction change types.
const (
	ContextActionAdd    ContextActionChangeType = ContextActionChangeType(MumbleProto.ContextActionModify_Add)
	ContextActionRemove ContextActionChangeType = ContextActionChangeType(MumbleProto.ContextActionModify_Remove)
)

// ContextActionChangeEvent is the event that is passed to
// EventListener.OnContextActionChange.
type ContextActionChangeEvent struct {
	Client        *Client
	Type          ContextActionChangeType
	ContextAction *ContextAction
}

// ServerConfigEvent is the event that is passed to
// EventListener.OnServerConfig.
type ServerConfigEvent struct {
	Client *Client

	MaximumBitrate            *int
	WelcomeMessage            *string
	AllowHTML                 *bool
	MaximumMessageLength      *int
	MaximumImageMessageLength *int
	MaximumUsers              *int

	CodecAlpha       *int32
	CodecBeta        *int32
	CodecPreferAlpha *bool
	CodecOpus        *bool

	SuggestVersion    *Version
	SuggestPositional *bool
	SuggestPushToTalk *bool
}
