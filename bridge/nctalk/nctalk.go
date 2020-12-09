package nctalk

import (
	"context"
	"crypto/tls"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"

	"gomod.garykim.dev/nc-talk/ocs"
	"gomod.garykim.dev/nc-talk/room"
	"gomod.garykim.dev/nc-talk/user"
)

type Btalk struct {
	user  *user.TalkUser
	rooms []Broom
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Btalk{Config: cfg}
}

type Broom struct {
	room      *room.TalkRoom
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (b *Btalk) Connect() error {
	b.Log.Info("Connecting")
	tconfig := &user.TalkUserConfig{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: b.GetBool("SkipTLSVerify"), //nolint:gosec
		},
	}
	var err error
	b.user, err = user.NewUser(b.GetString("Server"), b.GetString("Login"), b.GetString("Password"), tconfig)
	if err != nil {
		b.Log.Error("Config could not be used")
		return err
	}
	_, err = b.user.Capabilities()
	if err != nil {
		b.Log.Error("Cannot Connect")
		return err
	}
	b.Log.Info("Connected")
	return nil
}

func (b *Btalk) Disconnect() error {
	for _, r := range b.rooms {
		r.ctxCancel()
	}
	return nil
}

func (b *Btalk) JoinChannel(channel config.ChannelInfo) error {
	tr, err := room.NewTalkRoom(b.user, channel.Name)
	if err != nil {
		return err
	}
	newRoom := Broom{
		room: tr,
	}
	newRoom.ctx, newRoom.ctxCancel = context.WithCancel(context.Background())
	c, err := newRoom.room.ReceiveMessages(newRoom.ctx)
	if err != nil {
		return err
	}
	b.rooms = append(b.rooms, newRoom)

	// Config
	guestSuffix := " (Guest)"
	if b.IsKeySet("GuestSuffix") {
		guestSuffix = b.GetString("GuestSuffix")
	}

	go func() {
		for msg := range c {
			msg := msg

			if msg.Error != nil {
				b.Log.Errorf("Fatal message poll error: %s\n", msg.Error)

				return
			}

			// ignore messages that are one of the following
			// * not a message from a user
			// * from ourselves
			if msg.MessageType != ocs.MessageComment || msg.ActorID == b.user.User {
				continue
			}
			remoteMessage := config.Message{
				Text:     formatRichObjectString(msg.Message, msg.MessageParameters),
				Channel:  newRoom.room.Token,
				Username: DisplayName(msg, guestSuffix),
				UserID:   msg.ActorID,
				Account:  b.Account,
			}
			// It is possible for the ID to not be set on older versions of Talk so we only set it if
			// the ID is not blank
			if msg.ID != 0 {
				remoteMessage.ID = strconv.Itoa(msg.ID)
			}

			// Handle Files
			err = b.handleFiles(&remoteMessage, &msg)
			if err != nil {
				b.Log.Errorf("Error handling file: %#v", msg)

				continue
			}

			b.Log.Debugf("<= Message is %#v", remoteMessage)
			b.Remote <- remoteMessage
		}
	}()
	return nil
}

func (b *Btalk) Send(msg config.Message) (string, error) {
	r := b.getRoom(msg.Channel)
	if r == nil {
		b.Log.Errorf("Could not find room for %v", msg.Channel)
		return "", nil
	}

	// Talk currently only supports sending normal messages
	if msg.Event != "" {
		return "", nil
	}
	sentMessage, err := r.room.SendMessage(msg.Username + msg.Text)
	if err != nil {
		b.Log.Errorf("Could not send message to room %v from %v: %v", msg.Channel, msg.Username, err)
		return "", nil
	}
	return strconv.Itoa(sentMessage.ID), nil
}

func (b *Btalk) getRoom(token string) *Broom {
	for _, r := range b.rooms {
		if r.room.Token == token {
			return &r
		}
	}
	return nil
}

func (b *Btalk) handleFiles(mmsg *config.Message, message *ocs.TalkRoomMessageData) error {
	for _, parameter := range message.MessageParameters {
		if parameter.Type == ocs.ROSTypeFile {
			// Get the file
			file, err := b.user.DownloadFile(parameter.Path)
			if err != nil {
				return err
			}

			if mmsg.Extra == nil {
				mmsg.Extra = make(map[string][]interface{})
			}

			mmsg.Extra["file"] = append(mmsg.Extra["file"], config.FileInfo{
				Name:   parameter.Name,
				Data:   file,
				Size:   int64(len(*file)),
				Avatar: false,
			})
		}
	}

	return nil
}

// Spec: https://github.com/nextcloud/server/issues/1706#issue-182308785
func formatRichObjectString(message string, parameters map[string]ocs.RichObjectString) string {
	for id, parameter := range parameters {
		text := parameter.Name

		switch parameter.Type {
		case ocs.ROSTypeUser, ocs.ROSTypeGroup:
			text = "@" + text
		case ocs.ROSTypeFile:
			if parameter.Link != "" {
				text = parameter.Name
			}
		}

		message = strings.ReplaceAll(message, "{"+id+"}", text)
	}

	return message
}

func DisplayName(msg ocs.TalkRoomMessageData, suffix string) string {
	if msg.ActorType == ocs.ActorGuest {
		if msg.ActorDisplayName == "" {
			return "Guest"
		}

		return msg.ActorDisplayName + suffix
	}

	return msg.ActorDisplayName
}
