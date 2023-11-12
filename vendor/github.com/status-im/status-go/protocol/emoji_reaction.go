package protocol

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"

	accountJson "github.com/status-im/status-go/account/json"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

// EmojiReaction represents an emoji reaction from a user in the application layer, used for persistence, querying and
// signaling
type EmojiReaction struct {
	*protobuf.EmojiReaction

	// From is a public key of the author of the emoji reaction.
	From string `json:"from,omitempty"`

	// SigPubKey is the ecdsa encoded public key of the emoji reaction author
	SigPubKey *ecdsa.PublicKey `json:"-"`

	// LocalChatID is the chatID of the local chat (one-to-one are not symmetric)
	LocalChatID string `json:"localChatId"`
}

func NewEmojiReaction() *EmojiReaction {
	return &EmojiReaction{EmojiReaction: &protobuf.EmojiReaction{}}
}

// ID is the Keccak256() contatenation of From-MessageID-EmojiType
func (e *EmojiReaction) ID() string {
	return types.EncodeHex(crypto.Keccak256([]byte(fmt.Sprintf("%s%s%d", e.From, e.MessageId, e.Type))))
}

// GetSigPubKey returns an ecdsa encoded public key
// this function is required to implement the ChatEntity interface
func (e *EmojiReaction) GetSigPubKey() *ecdsa.PublicKey {
	return e.SigPubKey
}

// GetProtoBuf returns the struct's embedded protobuf struct
// this function is required to implement the ChatEntity interface
func (e *EmojiReaction) GetProtobuf() proto.Message {
	return e.EmojiReaction
}

// SetMessageType a setter for the MessageType field
// this function is required to implement the ChatEntity interface
func (e *EmojiReaction) SetMessageType(messageType protobuf.MessageType) {
	e.MessageType = messageType
}

func (e *EmojiReaction) MarshalJSON() ([]byte, error) {
	item := struct {
		ID          string                      `json:"id"`
		Clock       uint64                      `json:"clock,omitempty"`
		ChatID      string                      `json:"chatId,omitempty"`
		LocalChatID string                      `json:"localChatId,omitempty"`
		From        string                      `json:"from"`
		MessageID   string                      `json:"messageId,omitempty"`
		MessageType protobuf.MessageType        `json:"messageType,omitempty"`
		Retracted   bool                        `json:"retracted,omitempty"`
		EmojiID     protobuf.EmojiReaction_Type `json:"emojiId,omitempty"`
	}{

		ID:          e.ID(),
		Clock:       e.Clock,
		ChatID:      e.ChatId,
		LocalChatID: e.LocalChatID,
		From:        e.From,
		MessageID:   e.MessageId,
		MessageType: e.MessageType,
		Retracted:   e.Retracted,
		EmojiID:     e.Type,
	}

	ext, err := accountJson.ExtendStructWithPubKeyData(item.From, item)
	if err != nil {
		return nil, err
	}

	return json.Marshal(ext)
}

// WrapGroupMessage indicates whether we should wrap this in membership information
func (e *EmojiReaction) WrapGroupMessage() bool {
	return false
}
