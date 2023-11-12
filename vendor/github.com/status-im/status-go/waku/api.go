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

package waku

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/status-im/status-go/waku/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
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
	Messages       int     `json:"messages"`       // Number of floating messages.
	MinPow         float64 `json:"minPow"`         // Minimal accepted PoW
	MaxMessageSize uint32  `json:"maxMessageSize"` // Maximum accepted message size
}

// Context is used higher up the food-chain and without significant refactoring is not a simple thing to remove / change

// Info returns diagnostic information about the waku node.
func (api *PublicWakuAPI) Info(ctx context.Context) Info {
	return Info{
		Messages:       len(api.w.msgQueue) + len(api.w.p2pMsgQueue),
		MinPow:         api.w.MinPow(),
		MaxMessageSize: api.w.MaxMessageSize(),
	}
}

// SetMaxMessageSize sets the maximum message size that is accepted.
// Upper limit is defined by MaxMessageSize.
func (api *PublicWakuAPI) SetMaxMessageSize(ctx context.Context, size uint32) (bool, error) {
	return true, api.w.SetMaxMessageSize(size)
}

// SetMinPoW sets the minimum PoW, and notifies the peers.
func (api *PublicWakuAPI) SetMinPoW(ctx context.Context, pow float64) (bool, error) {
	return true, api.w.SetMinimumPoW(pow, true)
}

// SetBloomFilter sets the new value of bloom filter, and notifies the peers.
func (api *PublicWakuAPI) SetBloomFilter(ctx context.Context, bloom hexutil.Bytes) (bool, error) {
	return true, api.w.SetBloomFilter(bloom)
}

func (api *PublicWakuAPI) BloomFilter() []byte {
	return api.w.BloomFilter()
}

// MarkTrustedPeer marks a peer trusted, which will allow it to send historic (expired) messages.
// Note: This function is not adding new nodes, the node needs to exists as a peer.
func (api *PublicWakuAPI) MarkTrustedPeer(ctx context.Context, url string) (bool, error) {
	n, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		return false, err
	}
	return true, api.w.AllowP2PMessagesFromPeer(n.ID().Bytes())
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

// MakeLightClient turns the node into light client, which does not forward
// any incoming messages, and sends only messages originated in this node.
func (api *PublicWakuAPI) MakeLightClient(ctx context.Context) bool {
	api.w.SetLightClientMode(true)
	return api.w.LightClientMode()
}

// CancelLightClient cancels light client mode.
func (api *PublicWakuAPI) CancelLightClient(ctx context.Context) bool {
	api.w.SetLightClientMode(false)
	return !api.w.LightClientMode()
}

//go:generate gencodec -type NewMessage -field-override newMessageOverride -out gen_newmessage_json.go

// NewMessage represents a new waku message that is posted through the RPC.
type NewMessage struct {
	SymKeyID   string           `json:"symKeyID"`
	PublicKey  []byte           `json:"pubKey"`
	Sig        string           `json:"sig"`
	TTL        uint32           `json:"ttl"`
	Topic      common.TopicType `json:"topic"`
	Payload    []byte           `json:"payload"`
	Padding    []byte           `json:"padding"`
	PowTime    uint32           `json:"powTime"`
	PowTarget  float64          `json:"powTarget"`
	TargetPeer string           `json:"targetPeer"`
	Ephemeral  bool             `json:"ephemeral"`
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

	params := &common.MessageParams{
		TTL:      req.TTL,
		Payload:  req.Payload,
		Padding:  req.Padding,
		WorkTime: req.PowTime,
		PoW:      req.PowTarget,
		Topic:    req.Topic,
	}

	// Set key that is used to sign the message
	if len(req.Sig) > 0 {
		if params.Src, err = api.w.GetPrivateKey(req.Sig); err != nil {
			return nil, err
		}
	}

	// Set symmetric key that is used to encrypt the message
	if symKeyGiven {
		if params.Topic == (common.TopicType{}) { // topics are mandatory with symmetric encryption
			return nil, ErrNoTopics
		}
		if params.KeySym, err = api.w.GetSymKey(req.SymKeyID); err != nil {
			return nil, err
		}
		if !common.ValidateDataIntegrity(params.KeySym, common.AESKeyLength) {
			return nil, ErrInvalidSymmetricKey
		}
	}

	// Set asymmetric key that is used to encrypt the message
	if pubKeyGiven {
		if params.Dst, err = crypto.UnmarshalPubkey(req.PublicKey); err != nil {
			return nil, ErrInvalidPublicKey
		}
	}

	// encrypt and sent message
	msg, err := common.NewSentMessage(params)
	if err != nil {
		return nil, err
	}

	var result []byte
	env, err := msg.Wrap(params, api.w.CurrentTime())
	if err != nil {
		return nil, err
	}

	// send to specific node (skip PoW check)
	if len(req.TargetPeer) > 0 {
		n, err := enode.Parse(enode.ValidSchemes, req.TargetPeer)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target peer: %s", err)
		}
		err = api.w.SendP2PMessages(n.ID().Bytes(), env)
		if err == nil {
			hash := env.Hash()
			result = hash[:]
		}
		return result, err
	}

	// ensure that the message PoW meets the node's minimum accepted PoW
	if req.PowTarget < api.w.MinPow() {
		return nil, ErrTooLowPoW
	}

	err = api.w.Send(env)
	if err == nil {
		hash := env.Hash()
		result = hash[:]
	}
	return result, err
}

// UninstallFilter is alias for Unsubscribe
func (api *PublicWakuAPI) UninstallFilter(id string) {
	api.w.Unsubscribe(id) // nolint: errcheck
}

// Unsubscribe disables and removes an existing filter.
func (api *PublicWakuAPI) Unsubscribe(id string) {
	api.w.Unsubscribe(id) // nolint: errcheck
}

// Criteria holds various filter options for inbound messages.
type Criteria struct {
	SymKeyID     string             `json:"symKeyID"`
	PrivateKeyID string             `json:"privateKeyID"`
	Sig          []byte             `json:"sig"`
	MinPow       float64            `json:"minPow"`
	Topics       []common.TopicType `json:"topics"`
	AllowP2P     bool               `json:"allowP2P"`
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
		PoW:      crit.MinPow,
		Messages: common.NewMemoryMessageStore(),
		AllowP2P: crit.AllowP2P,
	}

	if len(crit.Sig) > 0 {
		if filter.Src, err = crypto.UnmarshalPubkey(crit.Sig); err != nil {
			return nil, ErrInvalidSigningPubKey
		}
	}

	for _, bt := range crit.Topics {
		filter.Topics = append(filter.Topics, bt[:])
	}

	// listen for message that are encrypted with the given symmetric key
	if symKeyGiven {
		if len(filter.Topics) == 0 {
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
				_ = api.w.Unsubscribe(id)
				return
			}
		}
	}()

	return rpcSub, nil
}

//go:generate gencodec -type Message -field-override messageOverride -out gen_message_json.go

// Message is the RPC representation of a waku message.
type Message struct {
	Sig       []byte           `json:"sig,omitempty"`
	TTL       uint32           `json:"ttl"`
	Timestamp uint32           `json:"timestamp"`
	Topic     common.TopicType `json:"topic"`
	Payload   []byte           `json:"payload"`
	Padding   []byte           `json:"padding"`
	PoW       float64          `json:"pow"`
	Hash      []byte           `json:"hash"`
	Dst       []byte           `json:"recipientPublicKey,omitempty"`
	P2P       bool             `json:"bool,omitempty"`
}

// ToWakuMessage converts an internal message into an API version.
func ToWakuMessage(message *common.ReceivedMessage) *Message {
	msg := Message{
		Payload:   message.Payload,
		Padding:   message.Padding,
		Timestamp: message.Sent,
		TTL:       message.TTL,
		PoW:       message.PoW,
		Hash:      message.EnvelopeHash.Bytes(),
		Topic:     message.Topic,
		P2P:       message.P2P,
	}

	if message.Dst != nil {
		b := crypto.FromECDSAPub(message.Dst)
		if b != nil {
			msg.Dst = b
		}
	}

	if common.IsMessageSigned(message.Raw[0]) {
		b := crypto.FromECDSAPub(message.SigToPubKey())
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
	logger := api.w.logger.With(zap.String("site", "getFilterMessages"), zap.String("filterId", id))
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

		logger.Debug("retrieved filter message", zap.String("hash", msg.EnvelopeHash.String()), zap.Bool("isP2P", msg.P2P), zap.String("topic", msg.Topic.String()))
		messages = append(messages, ToWakuMessage(msg))
	}

	return messages, nil
}

// DeleteMessageFilter deletes a filter.
func (api *PublicWakuAPI) DeleteMessageFilter(id string) (bool, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	delete(api.lastUsed, id)
	return true, api.w.Unsubscribe(id)
}

// NewMessageFilter creates a new filter that can be used to poll for
// (new) messages that satisfy the given criteria.
func (api *PublicWakuAPI) NewMessageFilter(req Criteria) (string, error) {
	var (
		src     *ecdsa.PublicKey
		keySym  []byte
		keyAsym *ecdsa.PrivateKey
		topics  [][]byte

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

	if len(req.Topics) > 0 {
		topics = make([][]byte, len(req.Topics))
		for i, topic := range req.Topics {
			topics[i] = make([]byte, common.TopicLength)
			copy(topics[i], topic[:])
		}
	}

	f := &common.Filter{
		Src:      src,
		KeySym:   keySym,
		KeyAsym:  keyAsym,
		PoW:      req.MinPow,
		AllowP2P: req.AllowP2P,
		Topics:   topics,
		Messages: common.NewMemoryMessageStore(),
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
