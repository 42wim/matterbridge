package gethbridge

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/wakuv2"
	wakucommon "github.com/status-im/status-go/wakuv2/common"
)

type gethPublicWakuV2APIWrapper struct {
	api *wakuv2.PublicWakuAPI
}

// NewGethPublicWakuAPIWrapper returns an object that wraps Geth's PublicWakuAPI in a types interface
func NewGethPublicWakuV2APIWrapper(api *wakuv2.PublicWakuAPI) types.PublicWakuAPI {
	if api == nil {
		panic("PublicWakuV2API cannot be nil")
	}

	return &gethPublicWakuV2APIWrapper{
		api: api,
	}
}

// AddPrivateKey imports the given private key.
func (w *gethPublicWakuV2APIWrapper) AddPrivateKey(ctx context.Context, privateKey types.HexBytes) (string, error) {
	return w.api.AddPrivateKey(ctx, hexutil.Bytes(privateKey))
}

// GenerateSymKeyFromPassword derives a key from the given password, stores it, and returns its ID.
func (w *gethPublicWakuV2APIWrapper) GenerateSymKeyFromPassword(ctx context.Context, passwd string) (string, error) {
	return w.api.GenerateSymKeyFromPassword(ctx, passwd)
}

// DeleteKeyPair removes the key with the given key if it exists.
func (w *gethPublicWakuV2APIWrapper) DeleteKeyPair(ctx context.Context, key string) (bool, error) {
	return w.api.DeleteKeyPair(ctx, key)
}

func (w *gethPublicWakuV2APIWrapper) BloomFilter() []byte {
	return w.api.BloomFilter()
}

// NewMessageFilter creates a new filter that can be used to poll for
// (new) messages that satisfy the given criteria.
func (w *gethPublicWakuV2APIWrapper) NewMessageFilter(req types.Criteria) (string, error) {
	topics := make([]wakucommon.TopicType, len(req.Topics))
	for index, tt := range req.Topics {
		topics[index] = wakucommon.TopicType(tt)
	}

	criteria := wakuv2.Criteria{
		SymKeyID:      req.SymKeyID,
		PrivateKeyID:  req.PrivateKeyID,
		Sig:           req.Sig,
		PubsubTopic:   req.PubsubTopic,
		ContentTopics: topics,
	}
	return w.api.NewMessageFilter(criteria)
}

// GetFilterMessages returns the messages that match the filter criteria and
// are received between the last poll and now.
func (w *gethPublicWakuV2APIWrapper) GetFilterMessages(id string) ([]*types.Message, error) {
	msgs, err := w.api.GetFilterMessages(id)
	if err != nil {
		return nil, err
	}

	wrappedMsgs := make([]*types.Message, len(msgs))
	for index, msg := range msgs {
		wrappedMsgs[index] = &types.Message{
			Sig:         msg.Sig,
			Timestamp:   msg.Timestamp,
			PubsubTopic: msg.PubsubTopic,
			Topic:       types.TopicType(msg.ContentTopic),
			Payload:     msg.Payload,
			Padding:     msg.Padding,
			Hash:        msg.Hash,
			Dst:         msg.Dst,
		}
	}
	return wrappedMsgs, nil
}

// Post posts a message on the network.
// returns the hash of the message in case of success.
func (w *gethPublicWakuV2APIWrapper) Post(ctx context.Context, req types.NewMessage) ([]byte, error) {
	msg := wakuv2.NewMessage{
		SymKeyID:     req.SymKeyID,
		PublicKey:    req.PublicKey,
		Sig:          req.SigID, // Sig is really a SigID
		PubsubTopic:  req.PubsubTopic,
		ContentTopic: wakucommon.TopicType(req.Topic),
		Payload:      req.Payload,
		Padding:      req.Padding,
		TargetPeer:   req.TargetPeer,
		Ephemeral:    req.Ephemeral,
	}
	return w.api.Post(ctx, msg)
}
