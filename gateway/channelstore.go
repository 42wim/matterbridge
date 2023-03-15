package gateway

import (
	"encoding/gob"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/bbolt"
	"github.com/philippgille/gokv/encoding"
)

type ChannelData struct {
	HasWelcomeMessage bool
	WelcomeMessage    config.Message
}

func (r *Router) getChannelStore(path string) gokv.Store {
	gob.Register(map[string]interface{}{})
	gob.Register(config.FileInfo{})

	options := bbolt.Options{
		BucketName: "ChannelData",
		Path:       path,
		Codec:      encoding.Gob,
	}

	gob.Register(map[string]interface{}{})

	store, err := bbolt.NewStore(options)
	if err != nil {
		r.logger.Errorf("Could not connect to db: %s", path)
	}

	return store
}

func (r *Router) getWelcomeMessage(Channel string) *config.Message {
	channelData := new(ChannelData)
	found, err := r.ChannelStore.Get(Channel, channelData)
	if err != nil {
		r.logger.Error(err)
	}

	if found && channelData.HasWelcomeMessage {
		return &channelData.WelcomeMessage
	}

	return nil
}

func (r *Router) setWelcomeMessage(Channel string, newWelcomeMessage *config.Message) error {
	channelData := new(ChannelData)
	r.ChannelStore.Get(Channel, channelData)

	if newWelcomeMessage == nil {
		channelData.HasWelcomeMessage = false
		channelData.WelcomeMessage = config.Message{}
	} else {
		channelData.HasWelcomeMessage = true
		channelData.WelcomeMessage = *newWelcomeMessage
	}

	err := r.ChannelStore.Set(Channel, channelData)
	if err != nil {
		r.logger.Errorf(err.Error())
	}
	return err
}
