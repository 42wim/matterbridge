package bgitter

import (
	"fmt"
	"github.com/42wim/go-gitter"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"strings"
)

type Bgitter struct {
	c       *gitter.Gitter
	Config  *config.Protocol
	Remote  chan config.Message
	Account string
	User    *gitter.User
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
	b.User, err = b.c.GetUser()
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
				if ev.Message.From.ID != b.User.ID {
					flog.Debugf("Sending message from %s on %s to gateway", ev.Message.From.Username, b.Account)
					rmsg := config.Message{Username: ev.Message.From.Username, Text: ev.Message.Text, Channel: room,
						Account: b.Account, Avatar: b.getAvatar(ev.Message.From.Username), UserID: ev.Message.From.ID,
						ID: ev.Message.ID}
					if strings.HasPrefix(ev.Message.Text, "@"+ev.Message.From.Username) {
						rmsg.Event = config.EVENT_USER_ACTION
						rmsg.Text = strings.Replace(rmsg.Text, "@"+ev.Message.From.Username+" ", "", -1)
					}
					flog.Debugf("Message is %#v", rmsg)
					b.Remote <- rmsg
				}
			case *gitter.GitterConnectionClosed:
				flog.Errorf("connection with gitter closed for room %s", room)
			}
		}
	}(stream, room.Name)
	return nil
}

func (b *Bgitter) Send(msg config.Message) (string, error) {
	flog.Debugf("Receiving %#v", msg)
	roomID := b.getRoomID(msg.Channel)
	if roomID == "" {
		flog.Errorf("Could not find roomID for %v", msg.Channel)
		return "", nil
	}
	if msg.ID != "" {
		flog.Debugf("updating message with id %s", msg.ID)
		_, err := b.c.UpdateMessage(roomID, msg.ID, msg.Username+msg.Text)
		if err != nil {
			return "", err
		}
		return "", nil
	}
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
