package brocketchat

import (
	"errors"
	"strings"
	"sync"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/hook/rockethook"
	"github.com/42wim/matterbridge/matterhook"
	lru "github.com/hashicorp/golang-lru"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/realtime"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/rest"
)

type Brocketchat struct {
	mh    *matterhook.Client
	rh    *rockethook.Client
	c     *realtime.Client
	r     *rest.Client
	cache *lru.Cache
	*bridge.Config
	messageChan chan models.Message
	channelMap  map[string]string
	user        *models.User
	sync.RWMutex
}

const (
	sUserJoined       = "uj"
	sUserLeft         = "ul"
	sRoomChangedTopic = "room_changed_topic"
)

func New(cfg *bridge.Config) bridge.Bridger {
	newCache, err := lru.New(100)
	if err != nil {
		cfg.Log.Fatalf("Could not create LRU cache for rocketchat bridge: %v", err)
	}
	b := &Brocketchat{
		Config:      cfg,
		messageChan: make(chan models.Message),
		channelMap:  make(map[string]string),
		cache:       newCache,
	}
	b.Log.Debugf("enabling rocketchat")
	return b
}

func (b *Brocketchat) Command(cmd string) string {
	return ""
}

func (b *Brocketchat) Connect() error {
	if b.GetString("WebhookBindAddress") != "" {
		if err := b.doConnectWebhookBind(); err != nil {
			return err
		}
		go b.handleRocket()
		return nil
	}
	switch {
	case b.GetString("WebhookURL") != "":
		if err := b.doConnectWebhookURL(); err != nil {
			return err
		}
		go b.handleRocket()
		return nil
	case b.GetString("Login") != "":
		b.Log.Info("Connecting using login/password (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleRocket()
	}
	if b.GetString("WebhookBindAddress") == "" && b.GetString("WebhookURL") == "" &&
		b.GetString("Login") == "" {
		return errors.New("no connection method found. See that you have WebhookBindAddress, WebhookURL or Login/Password/Server configured")
	}
	return nil
}

func (b *Brocketchat) Disconnect() error {
	return nil
}

func (b *Brocketchat) JoinChannel(channel config.ChannelInfo) error {
	if b.c == nil {
		return nil
	}
	id, err := b.c.GetChannelId(strings.TrimPrefix(channel.Name, "#"))
	if err != nil {
		return err
	}
	b.Lock()
	b.channelMap[id] = channel.Name
	b.Unlock()
	mychannel := &models.Channel{ID: id, Name: strings.TrimPrefix(channel.Name, "#")}
	if err := b.c.JoinChannel(id); err != nil {
		return err
	}
	if err := b.c.SubscribeToMessageStream(mychannel, b.messageChan); err != nil {
		return err
	}
	return nil
}

func (b *Brocketchat) Send(msg config.Message) (string, error) {
	// strip the # if people has set this
	msg.Channel = strings.TrimPrefix(msg.Channel, "#")
	channel := &models.Channel{ID: b.getChannelID(msg.Channel), Name: msg.Channel}

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		msg.Text = "_" + msg.Text + "_"
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}
		return msg.ID, b.c.DeleteMessage(&models.Message{ID: msg.ID})
	}

	// Use webhook to send the message
	if b.GetString("WebhookURL") != "" {
		return "", b.sendWebhook(&msg)
	}

	// Prepend nick if configured
	if b.GetBool("PrefixMessagesWithNick") {
		msg.Text = msg.Username + msg.Text
	}

	// Edit message if we have an ID
	if msg.ID != "" {
		return msg.ID, b.c.EditMessage(&models.Message{ID: msg.ID, Msg: msg.Text, RoomID: b.getChannelID(msg.Channel)})
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			// strip the # if people has set this
			rmsg.Channel = strings.TrimPrefix(rmsg.Channel, "#")
			smsg := &models.Message{
				RoomID: b.getChannelID(rmsg.Channel),
				Msg:    rmsg.Username + rmsg.Text,
				PostMessage: models.PostMessage{
					Avatar: rmsg.Avatar,
					Alias:  rmsg.Username,
				},
			}
			if _, err := b.c.SendMessage(smsg); err != nil {
				b.Log.Errorf("SendMessage failed: %s", err)
			}
		}
		if len(msg.Extra["file"]) > 0 {
			return "", b.handleUploadFile(&msg)
		}
	}

	smsg := &models.Message{
		RoomID: channel.ID,
		Msg:    msg.Text,
		PostMessage: models.PostMessage{
			Avatar: msg.Avatar,
			Alias:  msg.Username,
		},
	}

	rmsg, err := b.c.SendMessage(smsg)
	if rmsg == nil {
		return "", err
	}
	return rmsg.ID, err
}
