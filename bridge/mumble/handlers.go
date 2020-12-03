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
			fname := b.Account + "_" + strconv.FormatInt(now.UnixNano(), 10) + "_" + strconv.Itoa(i) + part.FileExtension
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
	event.Client.Self.SetSelfMuted(true)
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

func (b *Bmumble) handleUserChange(event *gumble.UserChangeEvent) {
	// Only care about changes to self
	if event.User != event.Client.Self {
		return
	}
	// Someone attempted to move the user out of the configured channel; attempt to join back
	if b.Channel != nil {
		if err := b.doJoin(event.Client, *b.Channel); err != nil {
			b.Log.Error(err)
		}
	}
}

func (b *Bmumble) handleDisconnect(event *gumble.DisconnectEvent) {
	b.connected <- *event
}
