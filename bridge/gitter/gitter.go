package bgitter

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"github.com/sromku/go-gitter"
	"strings"
)

type Bgitter struct {
	c       *gitter.Gitter
	Config  *config.Protocol
	Remote  chan config.Message
	Account string
	Users   []gitter.User
	Rooms   []gitter.Room
}

var flog *log.Entry
var protocol = "gitter"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Bgitter {
	b := &Bgitter{}
	b.Config = &cfg
	b.Remote = c
	b.Account = account
	return b
}

func (b *Bgitter) Connect() error {
	var err error
	flog.Info("Connecting")
	b.c = gitter.New(b.Config.Token)
	_, err = b.c.GetUser()
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	flog.Info("Connection succeeded")
	b.Rooms, _ = b.c.GetRooms()
	return nil
}

func (b *Bgitter) Disconnect() error {
	return nil

}

func (b *Bgitter) JoinChannel(channel string) error {
	room := channel
	roomID := b.getRoomID(room)
	if roomID == "" {
		return fmt.Errorf("Could not find roomID for %v. Please create the room on gitter.im", channel)
	}
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
				// check for ZWSP to see if it's not an echo
				if !strings.HasSuffix(ev.Message.Text, "​") {
					flog.Debugf("Sending message from %s on %s to gateway", ev.Message.From.Username, b.Account)
					b.Remote <- config.Message{Username: ev.Message.From.Username, Text: ev.Message.Text, Channel: room,
						Account: b.Account, Avatar: b.getAvatar(ev.Message.From.Username)}
				}
			case *gitter.GitterConnectionClosed:
				flog.Errorf("connection with gitter closed for room %s", room)
			}
		}
	}(stream, room)
	return nil
}

func (b *Bgitter) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	roomID := b.getRoomID(msg.Channel)
	if roomID == "" {
		flog.Errorf("Could not find roomID for %v", msg.Channel)
		return nil
	}
	// add ZWSP because gitter echoes our own messages
	return b.c.SendMessage(roomID, msg.Username+msg.Text+" ​")
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
