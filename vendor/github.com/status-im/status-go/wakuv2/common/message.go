package common

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/waku-org/go-waku/waku/v2/payload"
	"github.com/waku-org/go-waku/waku/v2/protocol"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

// MessageType represents where this message comes from
type MessageType int

const (
	RelayedMessageType MessageType = iota
	StoreMessageType
)

// MessageParams specifies the exact way a message should be wrapped
// into an Envelope.
type MessageParams struct {
	Src     *ecdsa.PrivateKey
	Dst     *ecdsa.PublicKey
	KeySym  []byte
	Topic   TopicType
	Payload []byte
	Padding []byte
}

// ReceivedMessage represents a data packet to be received through the
// WakuV2 protocol and successfully decrypted.
type ReceivedMessage struct {
	Envelope *protocol.Envelope // Wrapped Waku Message

	MsgType MessageType

	Data      []byte
	Padding   []byte
	Signature []byte

	Sent uint32           // Time when the message was posted into the network
	Src  *ecdsa.PublicKey // Message recipient (identity used to decode the message)
	Dst  *ecdsa.PublicKey // Message recipient (identity used to decode the message)

	PubsubTopic  string
	ContentTopic TopicType

	SymKeyHash common.Hash // The Keccak256Hash of the key

	hash common.Hash

	Processed atomic.Bool
}

// MessagesRequest contains details of a request for historic messages.
type MessagesRequest struct {
	// ID of the request. The current implementation requires ID to be 32-byte array,
	// however, it's not enforced for future implementation.
	ID []byte `json:"id"`

	// From is a lower bound of time range.
	From uint32 `json:"from"`

	// To is a upper bound of time range.
	To uint32 `json:"to"`

	// Limit determines the number of messages sent by the mail server
	// for the current paginated request.
	Limit uint32 `json:"limit"`

	// Cursor is used as starting point for paginated requests.
	Cursor []byte `json:"cursor"`

	// Topics is a list of topics. A returned message should
	// belong to one of the topics from the list.
	Topics [][]byte `json:"topics"`
}

func (r MessagesRequest) Validate() error {
	if len(r.ID) != common.HashLength {
		return errors.New("invalid 'ID', expected a 32-byte slice")
	}

	if r.From > r.To {
		return errors.New("invalid 'From' value which is greater than To")
	}

	if r.Limit > MaxLimitInMessagesRequest {
		return fmt.Errorf("invalid 'Limit' value, expected value lower than %d", MaxLimitInMessagesRequest)
	}

	return nil
}

// EnvelopeError code and optional description of the error.
type EnvelopeError struct {
	Hash        common.Hash
	Code        uint
	Description string
}

// ErrorToEnvelopeError converts common golang error into EnvelopeError with a code.
func ErrorToEnvelopeError(hash common.Hash, err error) EnvelopeError {
	code := EnvelopeOtherError
	switch err.(type) {
	case TimeSyncError:
		code = EnvelopeTimeNotSynced
	}
	return EnvelopeError{
		Hash:        hash,
		Code:        code,
		Description: err.Error(),
	}
}

// MessagesResponse sent as a response after processing batch of envelopes.
type MessagesResponse struct {
	// Hash is a hash of all envelopes sent in the single batch.
	Hash common.Hash
	// Per envelope error.
	Errors []EnvelopeError
}

func (msg *ReceivedMessage) isSymmetricEncryption() bool {
	return msg.SymKeyHash != common.Hash{}
}

func (msg *ReceivedMessage) isAsymmetricEncryption() bool {
	return msg.Dst != nil
}

// MessageStore defines interface for temporary message store.
type MessageStore interface {
	Add(*ReceivedMessage) error
	Pop() ([]*ReceivedMessage, error)
}

// NewMemoryMessageStore returns pointer to an instance of the MemoryMessageStore.
func NewMemoryMessageStore() *MemoryMessageStore {
	return &MemoryMessageStore{
		messages: map[common.Hash]*ReceivedMessage{},
	}
}

// MemoryMessageStore represents messages stored in a memory hash table.
type MemoryMessageStore struct {
	mu       sync.Mutex
	messages map[common.Hash]*ReceivedMessage
}

func NewReceivedMessage(env *protocol.Envelope, msgType MessageType) *ReceivedMessage {
	ct, err := ExtractTopicFromContentTopic(env.Message().ContentTopic)
	if err != nil {
		log.Debug("failed to extract content topic from message", "topic", env.Message().ContentTopic, "err", err)
		return nil
	}

	return &ReceivedMessage{
		Envelope:     env,
		MsgType:      msgType,
		Sent:         uint32(env.Message().GetTimestamp() / int64(time.Second)),
		ContentTopic: ct,
		PubsubTopic:  env.PubsubTopic(),
	}
}

// Hash returns the SHA3 hash of the envelope, calculating it if not yet done.
func (msg *ReceivedMessage) Hash() common.Hash {
	if (msg.hash == common.Hash{}) {
		msg.hash = common.BytesToHash(msg.Envelope.Hash())
	}
	return msg.hash
}

// Add adds message to store.
func (store *MemoryMessageStore) Add(msg *ReceivedMessage) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exist := store.messages[msg.Hash()]; !exist {
		store.messages[msg.Hash()] = msg
	}
	return nil
}

// Pop returns all available messages and cleans the store.
func (store *MemoryMessageStore) Pop() ([]*ReceivedMessage, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	all := make([]*ReceivedMessage, 0, len(store.messages))
	for hash, msg := range store.messages {
		delete(store.messages, hash)
		all = append(all, msg)
	}
	return all, nil
}

// Open tries to decrypt an message, and populates the message fields in case of success.
func (msg *ReceivedMessage) Open(watcher *Filter) (result *ReceivedMessage) {
	if watcher == nil {
		return nil
	}

	// The API interface forbids filters doing both symmetric and asymmetric encryption.
	if watcher.expectsAsymmetricEncryption() && watcher.expectsSymmetricEncryption() {
		return nil
	}

	// TODO: should we update msg instead of creating a new received message?
	result = new(ReceivedMessage)

	keyInfo := new(payload.KeyInfo)
	if watcher.expectsAsymmetricEncryption() {
		keyInfo.Kind = payload.Asymmetric
		keyInfo.PrivKey = watcher.KeyAsym
		msg.Dst = &watcher.KeyAsym.PublicKey
	} else if watcher.expectsSymmetricEncryption() {
		keyInfo.Kind = payload.Symmetric
		keyInfo.SymKey = watcher.KeySym
		msg.SymKeyHash = crypto.Keccak256Hash(watcher.KeySym)
	}

	raw, err := payload.DecodePayload(msg.Envelope.Message(), keyInfo)

	if err != nil {
		log.Error("failed to decode message", "err", err)
		return nil
	}

	result.Envelope = msg.Envelope
	result.Data = raw.Data
	result.Padding = raw.Padding
	result.Signature = raw.Signature
	result.Src = raw.PubKey

	result.Sent = uint32(msg.Envelope.Message().GetTimestamp() / int64(time.Second))

	ct, err := ExtractTopicFromContentTopic(msg.Envelope.Message().ContentTopic)
	if err != nil {
		log.Error("failed to decode message", "err", err)
		return nil
	}

	result.PubsubTopic = watcher.PubsubTopic
	result.ContentTopic = ct

	return result
}
