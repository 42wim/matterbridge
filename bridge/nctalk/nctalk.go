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

	go func() {
		for msg := range c {
			msg := msg

			if msg.Error != nil {
				b.Log.Errorf("Fatal message poll error: %s\n", msg.Error)

				return
			}

			// Ignore messages that are from the bot user
			if msg.ActorID == b.user.User || msg.ActorType == "bridged" {
				continue
			}

			// Handle deleting messages
			if msg.MessageType == ocs.MessageSystem && msg.Parent != nil && msg.Parent.MessageType == ocs.MessageDelete {
				b.handleDeletingMessage(&msg, &newRoom)
				continue
			}

			// Handle sending messages
			if msg.MessageType == ocs.MessageComment {
				b.handleSendingMessage(&msg, &newRoom)
				continue
			}

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

	// Standard Message Send
	if msg.Event == "" {
		// Handle sending files if they are included
		err := b.handleSendingFile(&msg, r)
		if err != nil {
			b.Log.Errorf("Could not send files in message to room %v from %v: %v", msg.Channel, msg.Username, err)

			return "", nil
		}

		sentMessage, err := b.sendText(r, &msg, msg.Text)
		if err != nil {
			b.Log.Errorf("Could not send message to room %v from %v: %v", msg.Channel, msg.Username, err)

			return "", nil
		}
		return strconv.Itoa(sentMessage.ID), nil
	}

	// Message Deletion
	if msg.Event == config.EventMsgDelete {
		messageID, err := strconv.Atoi(msg.ID)
		if err != nil {
			return "", err
		}
		data, err := r.room.DeleteMessage(messageID)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(data.ID), nil
	}

	// Message is not a type that is currently supported
	return "", nil
}

func (b *Btalk) getRoom(token string) *Broom {
	for _, r := range b.rooms {
		if r.room.Token == token {
			return &r
		}
	}
	return nil
}

func (b *Btalk) sendText(r *Broom, msg *config.Message, text string) (*ocs.TalkRoomMessageData, error) {
	messageToSend := &room.Message{Message: msg.Username + text}

	if b.GetBool("SeparateDisplayName") {
		messageToSend.Message = text
		messageToSend.ActorDisplayName = msg.Username
	}

	return r.room.SendComplexMessage(messageToSend)
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

func (b *Btalk) handleSendingFile(msg *config.Message, r *Broom) error {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if fi.URL == "" {
			continue
		}

		message := ""
		if fi.Comment != "" {
			message += fi.Comment + " "
		}
		message += fi.URL
		_, err := b.sendText(r, msg, message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Btalk) handleSendingMessage(msg *ocs.TalkRoomMessageData, r *Broom) {
	remoteMessage := config.Message{
		Text:     formatRichObjectString(msg.Message, msg.MessageParameters),
		Channel:  r.room.Token,
		Username: DisplayName(msg, b.guestSuffix()),
		UserID:   msg.ActorID,
		Account:  b.Account,
	}
	// It is possible for the ID to not be set on older versions of Talk so we only set it if
	// the ID is not blank
	if msg.ID != 0 {
		remoteMessage.ID = strconv.Itoa(msg.ID)
	}

	// Handle Files
	err := b.handleFiles(&remoteMessage, msg)
	if err != nil {
		b.Log.Errorf("Error handling file: %#v", msg)

		return
	}

	b.Log.Debugf("<= Message is %#v", remoteMessage)
	b.Remote <- remoteMessage
}

func (b *Btalk) handleDeletingMessage(msg *ocs.TalkRoomMessageData, r *Broom) {
	remoteMessage := config.Message{
		Event:   config.EventMsgDelete,
		Text:    config.EventMsgDelete,
		Channel: r.room.Token,
		ID:      strconv.Itoa(msg.Parent.ID),
		Account: b.Account,
	}
	b.Log.Debugf("<= Message being deleted is %#v", remoteMessage)
	b.Remote <- remoteMessage
}

func (b *Btalk) guestSuffix() string {
	guestSuffix := " (Guest)"
	if b.IsKeySet("GuestSuffix") {
		guestSuffix = b.GetString("GuestSuffix")
	}

	return guestSuffix
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

func DisplayName(msg *ocs.TalkRoomMessageData, suffix string) string {
	if msg.ActorType == ocs.ActorGuest {
		if msg.ActorDisplayName == "" {
			return "Guest"
		}

		return msg.ActorDisplayName + suffix
	}

	return msg.ActorDisplayName
}
