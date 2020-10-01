package gumbleutil

import (
	"layeh.com/gumble/gumble"
)

// ListenerFunc is a single listener function that implements the
// gumble.EventListener interface. This is useful if you would like to use a
// type-switch for handling the different event types.
//
// Example:
//  handler := func(e interface{}) {
//    switch e.(type) {
//    case *gumble.ConnectEvent:
//      println("Connected")
//    case *gumble.DisconnectEvent:
//      println("Disconnected")
//    // ...
//    }
//  }
//
//  client.Attach(gumbleutil.ListenerFunc(handler))
type ListenerFunc func(e interface{})

var _ gumble.EventListener = ListenerFunc(nil)

// OnConnect implements gumble.EventListener.OnConnect.
func (lf ListenerFunc) OnConnect(e *gumble.ConnectEvent) {
	lf(e)
}

// OnDisconnect implements gumble.EventListener.OnDisconnect.
func (lf ListenerFunc) OnDisconnect(e *gumble.DisconnectEvent) {
	lf(e)
}

// OnTextMessage implements gumble.EventListener.OnTextMessage.
func (lf ListenerFunc) OnTextMessage(e *gumble.TextMessageEvent) {
	lf(e)
}

// OnUserChange implements gumble.EventListener.OnUserChange.
func (lf ListenerFunc) OnUserChange(e *gumble.UserChangeEvent) {
	lf(e)
}

// OnChannelChange implements gumble.EventListener.OnChannelChange.
func (lf ListenerFunc) OnChannelChange(e *gumble.ChannelChangeEvent) {
	lf(e)
}

// OnPermissionDenied implements gumble.EventListener.OnPermissionDenied.
func (lf ListenerFunc) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {
	lf(e)
}

// OnUserList implements gumble.EventListener.OnUserList.
func (lf ListenerFunc) OnUserList(e *gumble.UserListEvent) {
	lf(e)
}

// OnACL implements gumble.EventListener.OnACL.
func (lf ListenerFunc) OnACL(e *gumble.ACLEvent) {
	lf(e)
}

// OnBanList implements gumble.EventListener.OnBanList.
func (lf ListenerFunc) OnBanList(e *gumble.BanListEvent) {
	lf(e)
}

// OnContextActionChange implements gumble.EventListener.OnContextActionChange.
func (lf ListenerFunc) OnContextActionChange(e *gumble.ContextActionChangeEvent) {
	lf(e)
}

// OnServerConfig implements gumble.EventListener.OnServerConfig.
func (lf ListenerFunc) OnServerConfig(e *gumble.ServerConfigEvent) {
	lf(e)
}
