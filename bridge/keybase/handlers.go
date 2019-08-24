package bkeybase

import (
	"strconv"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

func (b *Bkeybase) handleKeybase() {
	sub, err := b.kbc.ListenForNewTextMessages()
	if err != nil {
		b.Log.Errorf("Error listening: %s", err.Error())
	}

	go func() {
		for {
			msg, err := sub.Read()
			if err != nil {
				b.Log.Errorf("failed to read message: %s", err.Error())
			}

			if msg.Message.Content.Type != "text" {
				continue
			}

			if msg.Message.Sender.Username == b.kbc.GetUsername() {
				continue
			}

			b.handleMessage(msg.Message)

		}
	}()
}

func (b *Bkeybase) handleMessage(msg kbchat.Message) {
	b.Log.Debugf("== Receiving event: %#v", msg)
	if msg.Channel.TopicName != b.channel || msg.Channel.Name != b.team {
		return
	}

	if msg.Sender.Username != b.kbc.GetUsername() {

		// TODO download avatar

		// Create our message
		rmsg := config.Message{Username: msg.Sender.Username, Text: msg.Content.Text.Body, UserID: msg.Sender.Uid, Channel: msg.Channel.TopicName, ID: strconv.Itoa(msg.MsgID), Account: b.Account}

		// Text must be a string
		if msg.Content.Type != "text" {
			b.Log.Errorf("message is not text")
			return
		}

		b.Log.Debugf("<= Sending message from %s on %s to gateway", msg.Sender.Username, msg.Channel.Name)
		b.Remote <- rmsg
	}
}
