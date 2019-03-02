package brocketchat

import (
	"github.com/42wim/matterbridge/bridge/config"
)

func (b *Brocketchat) handleRocket() {
	messages := make(chan *config.Message)
	if b.GetString("WebhookBindAddress") != "" {
		b.Log.Debugf("Choosing webhooks based receiving")
		go b.handleRocketHook(messages)
	} else {
		b.Log.Debugf("Choosing login/password based receiving")
		go b.handleRocketClient(messages)
	}
	for message := range messages {
		message.Account = b.Account
		b.Log.Debugf("<= Sending message from %s on %s to gateway", message.Username, b.Account)
		b.Log.Debugf("<= Message is %#v", message)
		b.Remote <- *message
	}
}

func (b *Brocketchat) handleRocketHook(messages chan *config.Message) {
	for {
		message := b.rh.Receive()
		b.Log.Debugf("Receiving from rockethook %#v", message)
		// do not loop
		if message.UserName == b.GetString("Nick") {
			continue
		}
		messages <- &config.Message{
			UserID:   message.UserID,
			Username: message.UserName,
			Text:     message.Text,
			Channel:  message.ChannelName,
		}
	}
}

func (b *Brocketchat) handleRocketClient(messages chan *config.Message) {
	for message := range b.messageChan {
		// skip messages with same ID, apparently messages get duplicated for an unknown reason
		if _, ok := b.cache.Get(message.ID); ok {
			continue
		}
		b.cache.Add(message.ID, true)
		b.Log.Debugf("message %#v", message)
		m := message
		if b.skipMessage(&m) {
			b.Log.Debugf("Skipped message: %#v", message)
			continue
		}

		rmsg := &config.Message{Text: message.Msg,
			Username: message.User.UserName,
			Channel:  b.getChannelName(message.RoomID),
			Account:  b.Account,
			UserID:   message.User.ID,
			ID:       message.ID,
		}
		messages <- rmsg
	}
}

func (b *Brocketchat) handleUploadFile(msg *config.Message) error {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if err := b.uploadFile(&fi, b.getChannelID(msg.Channel)); err != nil {
			return err
		}
	}
	return nil
}
