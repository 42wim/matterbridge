package bmatrix

import (
	"html"
	"regexp"
	"sync"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

type Bkeybase struct {
	kbc    *kbchat.API
	UserID string
	sync.RWMutex
	htmlTag *regexp.Regexp
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger { // idk what this does
	b := &Bkeybase{Config: cfg}
	b.htmlTag = regexp.MustCompile("</.*?>")
	return b
}

func (b *Bkeybase) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	b.mc, err = matrix.NewClient(b.GetString("Server"), "", "")
	if err != nil {
		return err
	}
	resp, err := b.mc.Login(&matrix.ReqLogin{
		Type:     "m.login.password",
		User:     b.GetString("Login"),
		Password: b.GetString("Password"),
	})
	if err != nil {
		return err
	}
	b.mc.SetCredentials(resp.UserID, resp.AccessToken)
	b.UserID = resp.UserID
	b.Log.Info("Connection succeeded")
	go b.handlematrix()
	return nil
}

func (b *Bmatrix) Disconnect() error {
	return nil
}

func (b *Bmatrix) JoinChannel(channel config.ChannelInfo) error {
	resp, err := b.mc.JoinRoom(channel.Name, "", nil)
	if err != nil {
		return err
	}
	b.Lock()
	b.RoomMap[resp.RoomID] = channel.Name
	b.Unlock()
	return err
}

func (b *Bmatrix) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	channel := b.getRoomID(msg.Channel)
	b.Log.Debugf("Channel %s maps to channel id %s", msg.Channel, channel)

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		m := matrix.TextMessage{
			MsgType: "m.emote",
			Body:    msg.Username + msg.Text,
		}
		resp, err := b.mc.SendMessageEvent(channel, "m.room.message", m)
		if err != nil {
			return "", err
		}
		return resp.EventID, err
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}
		resp, err := b.mc.RedactEvent(channel, msg.ID, &matrix.ReqRedact{})
		if err != nil {
			return "", err
		}
		return resp.EventID, err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			if _, err := b.mc.SendText(channel, rmsg.Username+rmsg.Text); err != nil {
				b.Log.Errorf("sendText failed: %s", err)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFiles(&msg, channel)
		}
	}

	// Edit message if we have an ID
	// matrix has no editing support

	// Use notices to send join/leave events
	if msg.Event == config.EventJoinLeave {
		resp, err := b.mc.SendNotice(channel, msg.Username+msg.Text)
		if err != nil {
			return "", err
		}
		return resp.EventID, err
	}

	username := html.EscapeString(msg.Username)
	// check if we have a </tag>. if we have, we don't escape HTML. #696
	if b.htmlTag.MatchString(msg.Username) {
		username = msg.Username
	}
	// Post normal message with HTML support (eg riot.im)
	resp, err := b.mc.SendHTML(channel, msg.Username+msg.Text, username+helper.ParseMarkdown(msg.Text))
	if err != nil {
		return "", err
	}
	return resp.EventID, err
}

func (b *Bmatrix) getRoomID(channel string) string {
	b.RLock()
	defer b.RUnlock()
	for ID, name := range b.RoomMap {
		if name == channel {
			return ID
		}
	}
	return ""
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
	if ev.Sender != b.UserID {

		// TODO download avatar

		// Create our message
		rmsg := config.Message{Username: msg.Sender.Username, Channel: channel, Account: b.Account, UserID: ev.Sender, ID: ev.ID}

		// Text must be a string
		if rmsg.Text, ok = ev.Content["body"].(string); !ok {
			b.Log.Errorf("Content[body] is not a string: %T\n%#v",
				ev.Content["body"], ev.Content)
			return
		}

		// Remove homeserver suffix if configured
		if b.GetBool("NoHomeServerSuffix") {
			re := regexp.MustCompile("(.*?):.*")
			rmsg.Username = re.ReplaceAllString(rmsg.Username, `$1`)
		}

		// Delete event
		if ev.Type == "m.room.redaction" {
			rmsg.Event = config.EventMsgDelete
			rmsg.ID = ev.Redacts
			rmsg.Text = config.EventMsgDelete
			b.Remote <- rmsg
			return
		}

		// Do we have a /me action
		if ev.Content["msgtype"].(string) == "m.emote" {
			rmsg.Event = config.EventUserAction
		}

		// Do we have attachments
		if b.containsAttachment(ev.Content) {
			err := b.handleDownloadFile(&rmsg, ev.Content)
			if err != nil {
				b.Log.Errorf("download failed: %#v", err)
			}
		}

		b.Log.Debugf("<= Sending message from %s on %s to gateway", ev.Sender, b.Account)
		b.Remote <- rmsg
	}
}
