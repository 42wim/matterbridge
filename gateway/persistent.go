package gateway

import (
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/badgerdb"
	"github.com/philippgille/gokv/encoding"
)

func (gw *Gateway) getMessageMapStore(path string) gokv.Store {
	options := badgerdb.Options{
		Dir:   path,
		Codec: encoding.Gob,
	}

	store, err := badgerdb.NewStore(options)
	if err != nil {
		gw.logger.Error(err)
		gw.logger.Errorf("Could not connect to db: %s", path)
	}

	return store
}

func (gw *Gateway) getCanonicalMessageFromStore(messageID string) string {
	if messageID == "" {
		return ""
	}

	canonicalMsgID := new(string)
	found, err := gw.CanonicalStore.Get(messageID, canonicalMsgID)
	if err != nil {
		gw.logger.Error(err)
	}

	if found {
		return *canonicalMsgID
	}

	return ""
}

func (gw *Gateway) setCanonicalMessageToStore(messageID string, canonicalMsgID string) {
	err := gw.CanonicalStore.Set(messageID, canonicalMsgID)
	if err != nil {
		gw.logger.Error(err)
	}
}

func (gw *Gateway) getDestMessagesFromStore(canonicalMsgID string, dest *bridge.Bridge, channel *config.ChannelInfo) string {
	if canonicalMsgID == "" {
		return ""
	}

	destMessageIds := new([]BrMsgID)
	found, err := gw.MessageStore.Get(canonicalMsgID, destMessageIds)
	if err != nil {
		gw.logger.Error(err)
	}

	if found {
		for _, id := range *destMessageIds {
			// check protocol, bridge name and channelname
			// for people that reuse the same bridge multiple times. see #342
			if dest.Protocol == id.Protocol && dest.Name == id.DestName && channel.ID == id.ChannelID {
				return id.ID
			}
		}
	}
	return ""
}

func (gw *Gateway) setDestMessagesToStore(canonicalMsgID string, msgIDs []*BrMsgID) {
	for _, msgID := range msgIDs {
		gw.setCanonicalMessageToStore(msgID.Protocol+" "+msgID.ID, canonicalMsgID)
	}

	err := gw.MessageStore.Set(canonicalMsgID, msgIDs)
	if err != nil {
		gw.logger.Error(err)
	}
}
