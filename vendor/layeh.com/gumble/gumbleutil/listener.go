package gumbleutil

import (
	"layeh.com/gumble/gumble"
)

// Listener is a struct that implements the gumble.EventListener interface. The
// corresponding event function in the struct is called if it is non-nil.
type Listener struct {
	Connect             func(e *gumble.ConnectEvent)
	Disconnect          func(e *gumble.DisconnectEvent)
	TextMessage         func(e *gumble.TextMessageEvent)
	UserChange          func(e *gumble.UserChangeEvent)
	ChannelChange       func(e *gumble.ChannelChangeEvent)
	PermissionDenied    func(e *gumble.PermissionDeniedEvent)
	UserList            func(e *gumble.UserListEvent)
	ACL                 func(e *gumble.ACLEvent)
	BanList             func(e *gumble.BanListEvent)
	ContextActionChange func(e *gumble.ContextActionChangeEvent)
	ServerConfig        func(e *gumble.ServerConfigEvent)
}

var _ gumble.EventListener = (*Listener)(nil)

// OnConnect implements gumble.EventListener.OnConnect.
func (l Listener) OnConnect(e *gumble.ConnectEvent) {
	if l.Connect != nil {
		l.Connect(e)
	}
}

// OnDisconnect implements gumble.EventListener.OnDisconnect.
func (l Listener) OnDisconnect(e *gumble.DisconnectEvent) {
	if l.Disconnect != nil {
		l.Disconnect(e)
	}
}

// OnTextMessage implements gumble.EventListener.OnTextMessage.
func (l Listener) OnTextMessage(e *gumble.TextMessageEvent) {
	if l.TextMessage != nil {
		l.TextMessage(e)
	}
}

// OnUserChange implements gumble.EventListener.OnUserChange.
func (l Listener) OnUserChange(e *gumble.UserChangeEvent) {
	if l.UserChange != nil {
		l.UserChange(e)
	}
}

// OnChannelChange implements gumble.EventListener.OnChannelChange.
func (l Listener) OnChannelChange(e *gumble.ChannelChangeEvent) {
	if l.ChannelChange != nil {
		l.ChannelChange(e)
	}
}

// OnPermissionDenied implements gumble.EventListener.OnPermissionDenied.
func (l Listener) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {
	if l.PermissionDenied != nil {
		l.PermissionDenied(e)
	}
}

// OnUserList implements gumble.EventListener.OnUserList.
func (l Listener) OnUserList(e *gumble.UserListEvent) {
	if l.UserList != nil {
		l.UserList(e)
	}
}

// OnACL implements gumble.EventListener.OnACL.
func (l Listener) OnACL(e *gumble.ACLEvent) {
	if l.ACL != nil {
		l.ACL(e)
	}
}

// OnBanList implements gumble.EventListener.OnBanList.
func (l Listener) OnBanList(e *gumble.BanListEvent) {
	if l.BanList != nil {
		l.BanList(e)
	}
}

// OnContextActionChange implements gumble.EventListener.OnContextActionChange.
func (l Listener) OnContextActionChange(e *gumble.ContextActionChangeEvent) {
	if l.ContextActionChange != nil {
		l.ContextActionChange(e)
	}
}

// OnServerConfig implements gumble.EventListener.OnServerConfig.
func (l Listener) OnServerConfig(e *gumble.ServerConfigEvent) {
	if l.ServerConfig != nil {
		l.ServerConfig(e)
	}
}
