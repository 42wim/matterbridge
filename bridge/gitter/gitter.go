package bgitter

import (
	"fmt"
	"strings"

	"github.com/42wim/go-gitter"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
)

type Bgitter struct {
	c     *gitter.Gitter
	User  *gitter.User
	Users []gitter.User
	Rooms []gitter.Room
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bgitter{Config: cfg}
}

func (b *Bgitter) Connect() error {
	var err error
	b.Log.Info("Connecting")
	b.c = gitter.New(b.GetString("Token"))
	b.User, err = b.c.GetUser()
	if err != nil {
		return err
	}
	b.Rooms, err = b.c.GetRooms()
	if err != nil {
		return err
	}
	b.Log.Info("Connection succeeded")
	return nil
}

func (b *Bgitter) Disconnect() error {
	return nil

}

func (b *Bgitter) JoinChannel(channel config.ChannelInfo) error {
	roomID, err := b.c.GetRoomId(channel.Name)
	if err != nil {
		return fmt.Errorf("Could not find roomID for %v. Please create the room on gitter.im", channel.Name)
	}
	room, err := b.c.GetRoom(roomID)
	if err != nil {
		return err
	}
	b.Rooms = append(b.Rooms, *room)
	user, err := b.c.GetUser()
	if err != nil {
		return err
	}
	_, err = b.c.JoinRoom(roomID, user.ID)
	if err != nil {
		return err
	}
	users, _ := b.c.GetUsersInRoom(roomID)
	b.Users = append(b.Users, users...)
	stream := b.c.Stream(roomID)
	go b.c.Listen(stream)

	go func(stream *gitter.Stream, room string) {
		for event := range stream.Event {
			switch ev := event.Data.(type) {
			case *gitter.MessageReceived:
				// ignore message sent from ourselves
				if ev.Message.From.ID != b.User.ID {
					b.Log.Debugf("<= Sending message from %s on %s to gateway", ev.Message.From.Username, b.Account)
					rmsg := config.Message{Username: ev.Message.From.Username, Text: ev.Message.Text, Channel: room,
						Account: b.Account, Avatar: b.getAvatar(ev.Message.From.Username), UserID: ev.Message.From.ID,
						ID: ev.Message.ID}
					if strings.HasPrefix(ev.Message.Text, "@"+ev.Message.From.Username) {
						rmsg.Event = config.EventUserAction
						rmsg.Text = strings.Replace(rmsg.Text, "@"+ev.Message.From.Username+" ", "", -1)
					}
					b.Log.Debugf("<= Message is %#v", rmsg)
					b.Remote <- rmsg
				}
			case *gitter.GitterConnectionClosed:
				b.Log.Errorf("connection with gitter closed for room %s", room)
			}
		}
	}(stream, room.URI)
	return nil
}

func (b *Bgitter) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)
	roomID := b.getRoomID(msg.Channel)
	if roomID == "" {
		b.Log.Errorf("Could not find roomID for %v", msg.Channel)
		return "", nil
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}
		// gitter has no delete message api so we edit message to ""
		_, err := b.c.UpdateMessage(roomID, msg.ID, "")
		if err != nil {
			return "", err
		}
		return "", nil
	}

	// Upload a file (in gitter case send the upload URL because gitter has no native upload support)
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.c.SendMessage(roomID, rmsg.Username+rmsg.Text)
		}
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg, roomID)
		}
	}

	// Edit message
	if msg.ID != "" {
		b.Log.Debugf("updating message with id %s", msg.ID)
		_, err := b.c.UpdateMessage(roomID, msg.ID, msg.Username+msg.Text)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	// Post normal message
	resp, err := b.c.SendMessage(roomID, msg.Username+msg.Text)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (b *Bgitter) getRoomID(channel string) string {
	for _, v := range b.Rooms {
		if v.URI == channel {
			return v.ID
		}
	}
	return ""
}

func (b *Bgitter) getAvatar(user string) string {
	var avatar string
	if b.Users != nil {
		for _, u := range b.Users {
			if user == u.Username {
				return u.AvatarURLSmall
			}
		}
	}
	return avatar
}

func (b *Bgitter) handleUploadFile(msg *config.Message, roomID string) (string, error) {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if fi.Comment != "" {
			msg.Text += fi.Comment + ": "
		}
		if fi.URL != "" {
			msg.Text = fi.URL
			if fi.Comment != "" {
				msg.Text = fi.Comment + ": " + fi.URL
			}
		}
		_, err := b.c.SendMessage(roomID, msg.Username+msg.Text)
		if err != nil {
			return "", err
		}
	}
	return "", nil
}
