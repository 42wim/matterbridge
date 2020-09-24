package bmumble

import (
	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"

	"github.com/42wim/matterbridge/bridge/config"
)

func (b *Bmumble) handleServerConfig(event *gumble.ServerConfigEvent) {
	b.serverConfigUpdate <- *event
}

func (b *Bmumble) handleTextMessage(event *gumble.TextMessageEvent) {
	rmsg := config.Message{
		Text:     event.TextMessage.Message,
		Channel:  event.Client.Self.Channel.Name,
		Username: event.TextMessage.Sender.Name,
		UserID:   event.TextMessage.Sender.Name + "@" + b.Host,
		Account:  b.Account,
	}
	b.Log.Debugf("<= Remote message is %+v", rmsg)
	b.Remote <- rmsg
}

func (b *Bmumble) handleConnect(event *gumble.ConnectEvent) {
	// Set the user's "bio"/comment
	if comment := b.GetString("UserComment"); comment != "" {
		event.Client.Self.SetComment(comment)
	}
	// No need to talk or listen
	event.Client.Self.SetSelfDeafened(true)
	event.Client.Self.SetSelfMuted(true)
	// if the Channel variable is set, this is a reconnect -> rejoin channel
	if b.Channel != "" {
		if err := b.doJoin(event.Client, b.Channel); err != nil {
			b.Log.Error(err)
		}
		b.Remote <- config.Message{
			Username: "system",
			Text:     "rejoin",
			Channel:  "",
			Account:  b.Account,
			Event:    config.EventRejoinChannels,
		}
	}
}

func (b *Bmumble) handleUserChange(event *gumble.UserChangeEvent) {
	// Only care about changes to self
	if event.User != event.Client.Self {
		return
	}
	// Someone attempted to move the user out of the configured channel; attempt to join back
	if b.Channel != "" && b.Channel != event.Client.Self.Channel.Name {
		if err := b.doJoin(event.Client, b.Channel); err != nil {
			b.Log.Error(err)
		}
	}
}

func (b *Bmumble) handleDisconnect(event *gumble.DisconnectEvent) {
	b.connected <- *event
}

func (b *Bmumble) makeDebugHandler() *gumbleutil.Listener {
	handler := gumbleutil.Listener{
		Connect:             func(e *gumble.ConnectEvent) { b.Log.Debugf("Received connect event: %+v", e) },
		Disconnect:          func(e *gumble.DisconnectEvent) { b.Log.Debugf("Received disconnect event: %+v", e) },
		TextMessage:         func(e *gumble.TextMessageEvent) { b.Log.Debugf("Received textmessage event: %+v", e) },
		UserChange:          func(e *gumble.UserChangeEvent) { b.Log.Debugf("Received userchange event: %+v", e) },
		ChannelChange:       func(e *gumble.ChannelChangeEvent) { b.Log.Debugf("Received channelchange event: %+v", e) },
		PermissionDenied:    func(e *gumble.PermissionDeniedEvent) { b.Log.Debugf("Received permissiondenied event: %+v", e) },
		UserList:            func(e *gumble.UserListEvent) { b.Log.Debugf("Received userlist event: %+v", e) },
		ACL:                 func(e *gumble.ACLEvent) { b.Log.Debugf("Received acl event: %+v", e) },
		BanList:             func(e *gumble.BanListEvent) { b.Log.Debugf("Received banlist event: %+v", e) },
		ContextActionChange: func(e *gumble.ContextActionChangeEvent) { b.Log.Debugf("Received contextactionchange event: %+v", e) },
		ServerConfig:        func(e *gumble.ServerConfigEvent) { b.Log.Debugf("Received serverconfig event: %+v", e) },
	}
	return &handler
}
