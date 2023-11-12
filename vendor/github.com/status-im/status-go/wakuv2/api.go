// Copyright 2019 The Waku Library Authors.
//
// The Waku library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Waku library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty off
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Waku library. If not, see <http://www.gnu.org/licenses/>.
//
// This software uses the go-ethereum library, which is licensed
// under the GNU Lesser General Public Library, version 3 or any later.

package wakuv2

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/waku-org/go-waku/waku/v2/payload"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"

	"github.com/status-im/status-go/wakuv2/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"google.golang.org/protobuf/proto"
)

// List of errors
var (
	ErrSymAsym              = errors.New("specify either a symmetric or an asymmetric key")
	ErrInvalidSymmetricKey  = errors.New("invalid symmetric key")
	ErrInvalidPublicKey     = errors.New("invalid public key")
	ErrInvalidSigningPubKey = errors.New("invalid signing public key")
	ErrTooLowPoW            = errors.New("message rejected, PoW too low")
	ErrNoTopics             = errors.New("missing topic(s)")
)

// PublicWakuAPI provides the waku RPC service that can be
// use publicly without security implications.
type PublicWakuAPI struct {
	w *Waku

	mu       sync.Mutex
	lastUsed map[string]time.Time // keeps track when a filter was polled for the last time.
}

// NewPublicWakuAPI create a new RPC waku service.
func NewPublicWakuAPI(w *Waku) *PublicWakuAPI {
	api := &PublicWakuAPI{
		w:        w,
		lastUsed: make(map[string]time.Time),
	}
	return api
}

// Info contains diagnostic information.
type Info struct {
	Messages       int    `json:"messages"`       // Number of floating messages.
	MaxMessageSize uint32 `json:"maxMessageSize"` // Maximum accepted message size
}

// Context is used higher up the food-chain and without significant refactoring is not a simple thing to remove / change

// Info returns diagnostic information about the waku node.
func (api *PublicWakuAPI) Info(ctx context.Context) Info {
	return Info{
		Messages:       len(api.w.msgQueue),
		MaxMessageSize: api.w.MaxMessageSize(),
	}
}

// NewKeyPair generates a new public and private key pair for message decryption and encryption.
// It returns an ID that can be used to refer to the keypair.
func (api *PublicWakuAPI) NewKeyPair(ctx context.Context) (string, error) {
	return api.w.NewKeyPair()
}

// AddPrivateKey imports the given private key.
func (api *PublicWakuAPI) AddPrivateKey(ctx context.Context, privateKey hexutil.Bytes) (string, error) {
	key, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return "", err
	}
	return api.w.AddKeyPair(key)
}

// DeleteKeyPair removes the key with the given key if it exists.
func (api *PublicWakuAPI) DeleteKeyPair(ctx context.Context, key string) (bool, error) {
	if ok := api.w.DeleteKeyPair(key); ok {
		return true, nil
	}
	return false, fmt.Errorf("key pair %s not found", key)
}

// HasKeyPair returns an indication if the node has a key pair that is associated with the given id.
func (api *PublicWakuAPI) HasKeyPair(ctx context.Context, id string) bool {
	return api.w.HasKeyPair(id)
}

// GetPublicKey returns the public key associated with the given key. The key is the hex
// encoded representation of a key in the form specified in section 4.3.6 of ANSI X9.62.
func (api *PublicWakuAPI) GetPublicKey(ctx context.Context, id string) (hexutil.Bytes, error) {
	key, err := api.w.GetPrivateKey(id)
	if err != nil {
		return hexutil.Bytes{}, err
	}
	return crypto.FromECDSAPub(&key.PublicKey), nil
}

// GetPrivateKey returns the private key associated with the given key. The key is the hex
// encoded representation of a key in the form specified in section 4.3.6 of ANSI X9.62.
func (api *PublicWakuAPI) GetPrivateKey(ctx context.Context, id string) (hexutil.Bytes, error) {
	key, err := api.w.GetPrivateKey(id)
	if err != nil {
		return hexutil.Bytes{}, err
	}
	return crypto.FromECDSA(key), nil
}

// NewSymKey generate a random symmetric key.
// It returns an ID that can be used to refer to the key.
// Can be used encrypting and decrypting messages where the key is known to both parties.
func (api *PublicWakuAPI) NewSymKey(ctx context.Context) (string, error) {
	return api.w.GenerateSymKey()
}

// AddSymKey import a symmetric key.
// It returns an ID that can be used to refer to the key.
// Can be used encrypting and decrypting messages where the key is known to both parties.
func (api *PublicWakuAPI) AddSymKey(ctx context.Context, key hexutil.Bytes) (string, error) {
	return api.w.AddSymKeyDirect([]byte(key))
}

// GenerateSymKeyFromPassword derive a key from the given password, stores it, and returns its ID.
func (api *PublicWakuAPI) GenerateSymKeyFromPassword(ctx context.Context, passwd string) (string, error) {
	return api.w.AddSymKeyFromPassword(passwd)
}

// HasSymKey returns an indication if the node has a symmetric key associated with the given key.
func (api *PublicWakuAPI) HasSymKey(ctx context.Context, id string) bool {
	return api.w.HasSymKey(id)
}

// GetSymKey returns the symmetric key associated with the given id.
func (api *PublicWakuAPI) GetSymKey(ctx context.Context, id string) (hexutil.Bytes, error) {
	return api.w.GetSymKey(id)
}

// DeleteSymKey deletes the symmetric key that is associated with the given id.
func (api *PublicWakuAPI) DeleteSymKey(ctx context.Context, id string) bool {
	return api.w.DeleteSymKey(id)
}

func (api *PublicWakuAPI) BloomFilter() []byte {
	return nil
}

//go:generate gencodec -type NewMessage -field-override newMessageOverride -out gen_newmessage_json.go

// NewMessage represents a new waku message that is posted through the RPC.
type NewMessage struct {
	SymKeyID     string           `json:"symKeyID"`
	PublicKey    []byte           `json:"pubKey"`
	Sig          string           `json:"sig"`
	PubsubTopic  string           `json:"pubsubTopic"`
	ContentTopic common.TopicType `json:"topic"`
	Payload      []byte           `json:"payload"`
	Padding      []byte           `json:"padding"`
	TargetPeer   string           `json:"targetPeer"`
	Ephemeral    bool             `json:"ephemeral"`
}

// Post posts a message on the Waku network.
// returns the hash of the message in case of success.
func (api *PublicWakuAPI) Post(ctx context.Context, req NewMessage) (hexutil.Bytes, error) {
	var (
		symKeyGiven = len(req.SymKeyID) > 0
		pubKeyGiven = len(req.PublicKey) > 0
		err         error
	)

	// user must specify either a symmetric or an asymmetric key
	if (symKeyGiven && pubKeyGiven) || (!symKeyGiven && !pubKeyGiven) {
		return nil, ErrSymAsym
	}

	var keyInfo *payload.KeyInfo = new(payload.KeyInfo)

	// Set key that is used to sign the message
	if len(req.Sig) > 0 {
		privKey, err := api.w.GetPrivateKey(req.Sig)
		if err != nil {
			return nil, err
		}
		keyInfo.PrivKey = privKey
	}

	// Set symmetric key that is used to encrypt the message
	if symKeyGiven {
		keyInfo.Kind = payload.Symmetric

		if req.ContentTopic == (common.TopicType{}) { // topics are mandatory with symmetric encryption
			return nil, ErrNoTopics
		}
		if keyInfo.SymKey, err = api.w.GetSymKey(req.SymKeyID); err != nil {
			return nil, err
		}
		if !common.ValidateDataIntegrity(keyInfo.SymKey, common.AESKeyLength) {
			return nil, ErrInvalidSymmetricKey
		}
	}

	// Set asymmetric key that is used to encrypt the message
	if pubKeyGiven {
		keyInfo.Kind = payload.Asymmetric

		var pubK *ecdsa.PublicKey
		if pubK, err = crypto.UnmarshalPubkey(req.PublicKey); err != nil {
			return nil, ErrInvalidPublicKey
		}
		keyInfo.PubKey = *pubK
	}

	var version uint32 = 1 // Use wakuv1 encryption

	p := new(payload.Payload)
	p.Data = req.Payload
	p.Key = keyInfo

	payload, err := p.Encode(version)
	if err != nil {
		return nil, err
	}

	wakuMsg := &pb.WakuMessage{
		Payload:      payload,
		Version:      &version,
		ContentTopic: req.ContentTopic.ContentTopic(),
		Timestamp:    proto.Int64(api.w.timestamp()),
		Meta:         []byte{}, // TODO: empty for now. Once we use Waku Archive v2, we should deprecate the timestamp and use an ULID here
		Ephemeral:    &req.Ephemeral,
	}

	hash, err := api.w.Send(req.PubsubTopic, wakuMsg)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

// UninstallFilter is alias for Unsubscribe
func (api *PublicWakuAPI) UninstallFilter(ctx context.Context, id string) {
	api.w.Unsubscribe(ctx, id) // nolint: errcheck
}

// Unsubscribe disables and removes an existing filter.
func (api *PublicWakuAPI) Unsubscribe(ctx context.Context, id string) {
	api.w.Unsubscribe(ctx, id) // nolint: errcheck
}

// Criteria holds various filter options for inbound messages.
type Criteria struct {
	SymKeyID      string             `json:"symKeyID"`
	PrivateKeyID  string             `json:"privateKeyID"`
	Sig           []byte             `json:"sig"`
	PubsubTopic   string             `json:"pubsubTopic"`
	ContentTopics []common.TopicType `json:"topics"`
}

// Messages set up a subscription that fires events when messages arrive that match
// the given set of criteria.
func (api *PublicWakuAPI) Messages(ctx context.Context, crit Criteria) (*rpc.Subscription, error) {
	var (
		symKeyGiven = len(crit.SymKeyID) > 0
		pubKeyGiven = len(crit.PrivateKeyID) > 0
		err         error
	)

	// ensure that the RPC connection supports subscriptions
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}

	// user must specify either a symmetric or an asymmetric key
	if (symKeyGiven && pubKeyGiven) || (!symKeyGiven && !pubKeyGiven) {
		return nil, ErrSymAsym
	}

	filter := common.Filter{
		Messages: common.NewMemoryMessageStore(),
	}

	if len(crit.Sig) > 0 {
		if filter.Src, err = crypto.UnmarshalPubkey(crit.Sig); err != nil {
			return nil, ErrInvalidSigningPubKey
		}
	}

	filter.PubsubTopic = crit.PubsubTopic
	filter.ContentTopics = common.NewTopicSet(crit.ContentTopics)

	// listen for message that are encrypted with the given symmetric key
	if symKeyGiven {
		if len(filter.ContentTopics) == 0 {
			return nil, ErrNoTopics
		}
		key, err := api.w.GetSymKey(crit.SymKeyID)
		if err != nil {
			return nil, err
		}
		if !common.ValidateDataIntegrity(key, common.AESKeyLength) {
			return nil, ErrInvalidSymmetricKey
		}
		filter.KeySym = key
		filter.SymKeyHash = crypto.Keccak256Hash(filter.KeySym)
	}

	// listen for messages that are encrypted with the given public key
	if pubKeyGiven {
		filter.KeyAsym, err = api.w.GetPrivateKey(crit.PrivateKeyID)
		if err != nil || filter.KeyAsym == nil {
			return nil, ErrInvalidPublicKey
		}
	}

	id, err := api.w.Subscribe(&filter)
	if err != nil {
		return nil, err
	}

	// create subscription and start waiting for message events
	rpcSub := notifier.CreateSubscription()
	go func() {
		// for now poll internally, refactor waku internal for channel support
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if filter := api.w.GetFilter(id); filter != nil {
					for _, rpcMessage := range toMessage(filter.Retrieve()) {
						if err := notifier.Notify(rpcSub.ID, rpcMessage); err != nil {
							log.Error("Failed to send notification", "err", err)
						}
					}
				}
			case <-rpcSub.Err():
				_ = api.w.Unsubscribe(context.Background(), id)
				return
			}
		}
	}()

	return rpcSub, nil
}

//go:generate gencodec -type Message -field-override messageOverride -out gen_message_json.go

// Message is the RPC representation of a waku message.
type Message struct {
	Sig          []byte           `json:"sig,omitempty"`
	Timestamp    uint32           `json:"timestamp"`
	PubsubTopic  string           `json:"pubsubTopic"`
	ContentTopic common.TopicType `json:"topic"`
	Payload      []byte           `json:"payload"`
	Padding      []byte           `json:"padding"`
	Hash         []byte           `json:"hash"`
	Dst          []byte           `json:"recipientPublicKey,omitempty"`
}

// ToWakuMessage converts an internal message into an API version.
func ToWakuMessage(message *common.ReceivedMessage) *Message {
	msg := Message{
		Payload:      message.Data,
		Padding:      message.Padding,
		Timestamp:    message.Sent,
		Hash:         message.Hash().Bytes(),
		PubsubTopic:  message.PubsubTopic,
		ContentTopic: message.ContentTopic,
	}

	if message.Dst != nil {
		b := crypto.FromECDSAPub(message.Dst)
		if b != nil {
			msg.Dst = b
		}
	}

	if message.Src != nil {
		b := crypto.FromECDSAPub(message.Src)
		if b != nil {
			msg.Sig = b
		}
	}

	return &msg
}

// toMessage converts a set of messages to its RPC representation.
func toMessage(messages []*common.ReceivedMessage) []*Message {
	msgs := make([]*Message, len(messages))
	for i, msg := range messages {
		msgs[i] = ToWakuMessage(msg)
	}
	return msgs
}

// GetFilterMessages returns the messages that match the filter criteria and
// are received between the last poll and now.
func (api *PublicWakuAPI) GetFilterMessages(id string) ([]*Message, error) {
	api.mu.Lock()
	f := api.w.GetFilter(id)
	if f == nil {
		api.mu.Unlock()
		return nil, fmt.Errorf("filter not found")
	}
	api.lastUsed[id] = time.Now()
	api.mu.Unlock()

	receivedMessages := f.Retrieve()
	messages := make([]*Message, 0, len(receivedMessages))
	for _, msg := range receivedMessages {
		messages = append(messages, ToWakuMessage(msg))
	}

	return messages, nil
}

// DeleteMessageFilter deletes a filter.
func (api *PublicWakuAPI) DeleteMessageFilter(id string) (bool, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	delete(api.lastUsed, id)
	return true, api.w.Unsubscribe(context.Background(), id)
}

// NewMessageFilter creates a new filter that can be used to poll for
// (new) messages that satisfy the given criteria.
func (api *PublicWakuAPI) NewMessageFilter(req Criteria) (string, error) {
	var (
		src     *ecdsa.PublicKey
		keySym  []byte
		keyAsym *ecdsa.PrivateKey

		symKeyGiven  = len(req.SymKeyID) > 0
		asymKeyGiven = len(req.PrivateKeyID) > 0

		err error
	)

	// user must specify either a symmetric or an asymmetric key
	if (symKeyGiven && asymKeyGiven) || (!symKeyGiven && !asymKeyGiven) {
		return "", ErrSymAsym
	}

	if len(req.Sig) > 0 {
		if src, err = crypto.UnmarshalPubkey(req.Sig); err != nil {
			return "", ErrInvalidSigningPubKey
		}
	}

	if symKeyGiven {
		if keySym, err = api.w.GetSymKey(req.SymKeyID); err != nil {
			return "", err
		}
		if !common.ValidateDataIntegrity(keySym, common.AESKeyLength) {
			return "", ErrInvalidSymmetricKey
		}
	}

	if asymKeyGiven {
		if keyAsym, err = api.w.GetPrivateKey(req.PrivateKeyID); err != nil {
			return "", err
		}
	}

	f := &common.Filter{
		Src:           src,
		KeySym:        keySym,
		KeyAsym:       keyAsym,
		PubsubTopic:   req.PubsubTopic,
		ContentTopics: common.NewTopicSet(req.ContentTopics),
		Messages:      common.NewMemoryMessageStore(),
	}

	id, err := api.w.Subscribe(f)
	if err != nil {
		return "", err
	}

	api.mu.Lock()
	api.lastUsed[id] = time.Now()
	api.mu.Unlock()

	return id, nil
}
