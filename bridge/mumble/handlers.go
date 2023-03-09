package bmumble

import (
	"strconv"
	"time"

	"layeh.com/gumble/gumble"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
)

func (b *Bmumble) handleServerConfig(event *gumble.ServerConfigEvent) {
	b.serverConfigUpdate <- *event
}

func (b *Bmumble) handleTextMessage(event *gumble.TextMessageEvent) {
	sender := "unknown"
	if event.TextMessage.Sender != nil {
		sender = event.TextMessage.Sender.Name
	}
	// If the text message is received before receiving a ServerSync
	// and UserState, Client.Self or Self.Channel are nil
	if event.Client.Self == nil || event.Client.Self.Channel == nil {
		b.Log.Warn("Connection bootstrap not finished, discarding text message")
		return
	}
	// Convert Mumble HTML messages to markdown
	parts, err := b.convertHTMLtoMarkdown(event.TextMessage.Message)
	if err != nil {
		b.Log.Error(err)
	}
	now := time.Now().UTC()
	for i, part := range parts {
		// Construct matterbridge message and pass on to the gateway
		rmsg := config.Message{
			Channel:  strconv.FormatUint(uint64(event.Client.Self.Channel.ID), 10),
			Username: sender,
			UserID:   sender + "@" + b.Host,
			Account:  b.Account,
		}
		if part.Image == nil {
			rmsg.Text = part.Text
		} else {
			fileExt := part.FileExtension
			if fileExt == ".jfif" {
				fileExt = ".jpg"
			}
			if fileExt == ".jpe" {
				fileExt = ".jpg"
			}
			fname := b.Account + "_" + strconv.FormatInt(now.UnixNano(), 10) + "_" + strconv.Itoa(i) + fileExt
			rmsg.Extra = make(map[string][]interface{})
			if err = helper.HandleDownloadSize(b.Log, &rmsg, fname, int64(len(part.Image)), b.General); err != nil {
				b.Log.WithError(err).Warn("not including image in message")
				continue
			}
			helper.HandleDownloadData(b.Log, &rmsg, fname, "", "", &part.Image, b.General)
		}
		b.Log.Debugf("Sending message to gateway: %+v", rmsg)
		b.Remote <- rmsg
	}
}

func (b *Bmumble) handleConnect(event *gumble.ConnectEvent) {
	// Set the user's "bio"/comment
	if comment := b.GetString("UserComment"); comment != "" && event.Client.Self != nil {
		event.Client.Self.SetComment(comment)
	}
	// No need to talk or listen
	event.Client.Self.SetSelfDeafened(true)
	// if the Channel variable is set, this is a reconnect -> rejoin channel
	if b.Channel != nil {
		if err := b.doJoin(event.Client, *b.Channel); err != nil {
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

func (b *Bmumble) handleJoinLeave(event *gumble.UserChangeEvent) {
	// Ignore events happening before setup is done
	if b.Channel == nil {
		return
	}
	if b.GetBool("nosendjoinpart") {
		return
	}
	b.Log.Debugf("Received gumble user change event: %+v", event)

	text := ""
	switch {
	case event.Type&gumble.UserChangeKicked > 0:
		text = " was kicked"
	case event.Type&gumble.UserChangeBanned > 0:
		text = " was banned"
	case event.Type&gumble.UserChangeDisconnected > 0:
		if event.User.Channel != nil && event.User.Channel.ID == *b.Channel {
			text = " left"
		}
	case event.Type&gumble.UserChangeConnected > 0:
		if event.User.Channel != nil && event.User.Channel.ID == *b.Channel {
			text = " joined"
		}
	case event.Type&gumble.UserChangeChannel > 0:
		// Treat Mumble channel changes the same as connects/disconnects; as far as matterbridge is concerned, they are identical
		if event.User.Channel != nil && event.User.Channel.ID == *b.Channel {
			text = " joined"
		} else {
			text = " left"
		}
	}

	if text != "" {
		b.Remote <- config.Message{
			Username: "system",
			Text:     event.User.Name + text,
			Channel:  strconv.FormatUint(uint64(*b.Channel), 10),
			Account:  b.Account,
			Event:    config.EventJoinLeave,
		}
	}
}

func (b *Bmumble) handleUserModified(event *gumble.UserChangeEvent) {
	// Ignore events happening before setup is done
	if b.Channel == nil {
		return
	}

	if event.Type&gumble.UserChangeChannel > 0 {
		// Someone attempted to move the user out of the configured channel; attempt to join back
		if err := b.doJoin(event.Client, *b.Channel); err != nil {
			b.Log.Error(err)
		}
	}
}

func (b *Bmumble) handleUserChange(event *gumble.UserChangeEvent) {
	// The UserChangeEvent is used for both the gumble client itself as well as other clients
	if event.User != event.Client.Self {
		// other users
		b.handleJoinLeave(event)
	} else {
		// gumble user
		b.handleUserModified(event)
	}
}

func (b *Bmumble) handleDisconnect(event *gumble.DisconnectEvent) {
	b.connected <- *event
}
