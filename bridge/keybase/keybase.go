package bmatrix

import (
	"strconv"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

type Bkeybase struct {
	kbc     *kbchat.API
	user    string
	channel string
	team    string
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger { // idk what this does
	b := &Bkeybase{Config: cfg}
	b.team = b.Config.GetString("Team")
	return b
}

func (b *Bkeybase) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Team"))
	b.kbc, err = kbchat.Start(kbchat.RunOptions{})
	if err != nil {
		return err
	}
	b.user = b.kbc.GetUsername()
	b.Log.Info("Connection succeeded")
	go b.handleKeybase()
	return nil
}

func (b *Bkeybase) Disconnect() error {
	return nil
}

func (b *Bkeybase) JoinChannel(channel config.ChannelInfo) error {
	b.Lock()
	b.channel = channel.Name
	b.Unlock()
	return nil
}

func (b *Bkeybase) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// TODO: /me handling
	// if msg.Event == config.EventUserAction {
	// 	m := matrix.TextMessage{
	// 		MsgType: "m.emote",
	// 		Body:    msg.Username + msg.Text,
	// 	}
	// 	resp, err := b.mc.SendMessageEvent(channel, "m.room.message", m)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	return resp.EventID, err
	// }

	// Delete message not supported by keybase go library yet

	// Upload a file if it exists
	// kbchat does not support attachments yet

	// Edit message if we have an ID
	// matrix has no editing support

	// Use notices to send join/leave events
	// if msg.Event == config.EventJoinLeave {
	// 	resp, err := b.mc.SendNotice(channel, msg.Username+msg.Text)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	return resp.EventID, err
	// }

	// resp, err := b.mc.SendHTML(channel, msg.Username+msg.Text, username+helper.ParseMarkdown(msg.Text))
	resp, err := b.kbc.SendMessageByTeamName(b.team, msg.Username+msg.Text, &b.channel)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(resp.Result.MsgID), err
}

func (b *Bkeybase) handleKeybase() {
	sub, err := b.kbc.ListenForNewTextMessages()
	if err != nil {
		b.Log.Error("Error listening: %s", err.Error())
	}
	// syncer.OnEventType("m.room.redaction", b.handleEvent)
	// syncer.OnEventType("m.room.message", b.handleEvent)
	go func() {
		for {
			msg, err := sub.Read()
			if err != nil {
				b.Log.Error("failed to read message: %s", err.Error())
			}

			if msg.Message.Content.Type != "text" {
				continue
			}

			if msg.Message.Sender.Username == b.kbc.GetUsername() {
				continue
			}

			b.handleEvent(msg.Message)

		}
	}()
}

func (b *Bkeybase) handleEvent(msg kbchat.Message) {
	b.Log.Debugf("== Receiving event: %#v", msg)
	if msg.Sender.Username != b.kbc.GetUsername() {

		// TODO download avatar

		// Create our message
		rmsg := config.Message{Username: msg.Sender.Username, Channel: msg.Channel.Name, ID: strconv.Itoa(msg.MsgID)}

		// Text must be a string
		if msg.Content.Type != "text" {
			b.Log.Errorf("message is not text")
			return
		}

		// Delete event TODO
		// if ev.Type == "m.room.redaction" {
		// 	rmsg.Event = config.EventMsgDelete
		// 	rmsg.ID = ev.Redacts
		// 	rmsg.Text = config.EventMsgDelete
		// 	b.Remote <- rmsg
		// 	return
		// }

		// Do we have a /me action
		// if ev.Content["msgtype"].(string) == "m.emote" {
		// 	rmsg.Event = config.EventUserAction
		// }

		// Do we have attachments
		// doesn't matter because we can't handle it yet

		b.Log.Debugf("<= Sending message from %s on %s to gateway", msg.Sender.Username, msg.Channel.Name)
		b.Remote <- rmsg
	}
}
