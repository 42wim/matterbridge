package bgitter

import (
	"github.com/42wim/go-gitter"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"strings"
)

type Bgitter struct {
	c        *gitter.Gitter
	Config   *config.Protocol
	Remote   chan config.Message
	protocol string
	origin   string
	Rooms    []gitter.Room
}

var flog *log.Entry
var protocol = "gitter"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(config config.Protocol, origin string, c chan config.Message) *Bgitter {
	b := &Bgitter{}
	b.Config = &config
	b.Remote = c
	b.protocol = protocol
	b.origin = origin
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

func (b *Bgitter) FullOrigin() string {
	return b.protocol + "." + b.origin
}

func (b *Bgitter) JoinChannel(channel string) error {
	_, err := b.c.JoinRoom(channel)
	if err != nil {
		return err
	}
	room := channel
	roomID := b.getRoomID(room)
	if roomID == "" {
		return nil
	}
	stream := b.c.Stream(roomID)
	go b.c.Listen(stream)

	go func(stream *gitter.Stream, room string) {
		for {
			event := <-stream.Event
			switch ev := event.Data.(type) {
			case *gitter.MessageReceived:
				// check for ZWSP to see if it's not an echo
				if !strings.HasSuffix(ev.Message.Text, "​") {
					flog.Debugf("Sending message from %s on %s to gateway", ev.Message.From.Username, b.FullOrigin())
					b.Remote <- config.Message{Username: ev.Message.From.Username, Text: ev.Message.Text, Channel: room,
						Origin: b.origin, Protocol: b.protocol, FullOrigin: b.FullOrigin()}
				}
			case *gitter.GitterConnectionClosed:
				flog.Errorf("connection with gitter closed for room %s", room)
			}
		}
	}(stream, room)
	return nil
}

func (b *Bgitter) Name() string {
	return b.protocol + "." + b.origin
}

func (b *Bgitter) Protocol() string {
	return b.protocol
}

func (b *Bgitter) Origin() string {
	return b.origin
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
