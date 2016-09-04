package bgitter

import (
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"github.com/sromku/go-gitter"
	"strings"
)

type Bgitter struct {
	c *gitter.Gitter
	*config.Config
	Remote chan config.Message
	Rooms  []gitter.Room
}

type Message struct {
	Text     string
	Channel  string
	Username string
}

var flog *log.Entry

func init() {
	flog = log.WithFields(log.Fields{"module": "gitter"})
}

func New(config *config.Config, c chan config.Message) *Bgitter {
	b := &Bgitter{}
	b.Config = config
	b.Remote = c
	return b
}

func (b *Bgitter) Connect() error {
	var err error
	flog.Info("Trying Gitter connection")
	b.c = gitter.New(b.Config.Gitter.Token)
	_, err = b.c.GetUser()
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	flog.Info("Connection succeeded")
	b.setupChannels()
	go b.handleGitter()
	return nil
}

func (b *Bgitter) Name() string {
	return "gitter"
}

func (b *Bgitter) Send(msg config.Message) error {
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

func (b *Bgitter) handleGitter() {
	for _, val := range b.Config.Channel {
		room := val.Gitter
		roomID := b.getRoomID(room)
		if roomID == "" {
			continue
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
						b.Remote <- config.Message{Username: ev.Message.From.Username, Text: ev.Message.Text, Channel: room, Origin: "gitter"}
					}
				case *gitter.GitterConnectionClosed:
					flog.Errorf("connection with gitter closed for room %s", room)
				}
			}
		}(stream, room)
	}
}

func (b *Bgitter) setupChannels() {
	b.Rooms, _ = b.c.GetRooms()
	for _, val := range b.Config.Channel {
		flog.Infof("Joining %s as %s", val.Gitter, b.Gitter.Nick)
		_, err := b.c.JoinRoom(val.Gitter)
		if err != nil {
			log.Errorf("Joining %s failed", val.Gitter)
		}
	}
}
