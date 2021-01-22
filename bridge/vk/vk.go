package bvk

import (
	"context"
	"strconv"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	longpoll "github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type user struct {
	lastname, firstname, avatar string
}

type Bvk struct {
	c            *api.VK
	usernamesMap map[int]user
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bvk{usernamesMap: make(map[int]user), Config: cfg}
}

func (b *Bvk) Connect() error {
	b.Log.Info("Connecting")
	b.c = api.NewVK(b.GetString("Token"))
	lp, err := longpoll.NewLongPoll(b.c, b.GetInt("GroupID"))
	if err != nil {
		b.Log.Error(err)
		return err
	}

	lp.MessageNew(b.handleMessage)

	go lp.Run()

	return nil
}

func (b *Bvk) Disconnect() error {
	return nil
}

func (b *Bvk) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Bvk) Send(msg config.Message) (string, error) {
	b.Log.Debug(msg.Text)

	peerId, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return "", err
	}

	if msg.Text != "" {
		text := msg.Username + msg.Text

		res, err := b.c.MessagesSend(api.Params{
			"peer_id":   peerId,
			"message":   text,
			"random_id": time.Now().Unix(),
		})

		if err != nil {
			return "", err
		}

		return string(res), nil
	}

	return "", nil
}

func (b *Bvk) getUser(id int) user {
	u, found := b.usernamesMap[id]
	if !found {
		b.Log.Debug("Fetching username for ", id)

		result, _ := b.c.UsersGet(api.Params{
			"user_ids": id,
			"fields":   "photo_200",
		})

		resUser := result[0]
		u = user{lastname: resUser.LastName, firstname: resUser.FirstName, avatar: resUser.Photo200}
		b.usernamesMap[id] = u
	}

	return u
}

func (b *Bvk) handleMessage(ctx context.Context, obj events.MessageNewObject) {
	msg := obj.Message
	b.Log.Debug("ChatID: ", msg.PeerID)
	u := b.getUser(msg.FromID)

	rmsg := config.Message{
		Event:    config.EventUserAction,
		Text:     msg.Text,
		Username: u.firstname + " " + u.lastname,
		Avatar:   u.avatar,
		Channel:  strconv.Itoa(msg.PeerID),
		Account:  b.Account,
		UserID:   strconv.Itoa(msg.FromID),
		ID:       strconv.Itoa(msg.ConversationMessageID),
	}

	b.Remote <- rmsg
}
