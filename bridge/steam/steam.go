package bsteam

import (
	"fmt"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/Philipp15b/go-steam/steamid"
)

type Bsteam struct {
	c         *steam.Client
	connected chan struct{}
	userMap   map[steamid.SteamId]string
	sync.RWMutex
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bsteam{Config: cfg}
	b.userMap = make(map[steamid.SteamId]string)
	b.connected = make(chan struct{})
	return b
}

func (b *Bsteam) Connect() error {
	b.Log.Info("Connecting")
	b.c = steam.NewClient()
	go b.handleEvents()
	go b.c.Connect()
	select {
	case <-b.connected:
		b.Log.Info("Connection succeeded")
	case <-time.After(time.Second * 30):
		return fmt.Errorf("connection timed out")
	}
	return nil
}

func (b *Bsteam) Disconnect() error {
	b.c.Disconnect()
	return nil

}

func (b *Bsteam) JoinChannel(channel config.ChannelInfo) error {
	id, err := steamid.NewId(channel.Name)
	if err != nil {
		return err
	}
	b.c.Social.JoinChat(id)
	return nil
}

func (b *Bsteam) Send(msg config.Message) (string, error) {
	// ignore delete messages
	if msg.Event == config.EventMsgDelete {
		return "", nil
	}
	id, err := steamid.NewId(msg.Channel)
	if err != nil {
		return "", err
	}

	// Handle files
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.c.Social.SendMessage(id, steamlang.EChatEntryType_ChatMsg, rmsg.Username+rmsg.Text)
		}
		for i := range msg.Extra["file"] {
			if err := b.handleFileInfo(&msg, msg.Extra["file"][i]); err != nil {
				b.Log.Error(err)
			}
			b.c.Social.SendMessage(id, steamlang.EChatEntryType_ChatMsg, msg.Username+msg.Text)
		}
		return "", nil
	}

	b.c.Social.SendMessage(id, steamlang.EChatEntryType_ChatMsg, msg.Username+msg.Text)
	return "", nil
}

func (b *Bsteam) getNick(id steamid.SteamId) string {
	b.RLock()
	defer b.RUnlock()
	if name, ok := b.userMap[id]; ok {
		return name
	}
	return "unknown"
}
