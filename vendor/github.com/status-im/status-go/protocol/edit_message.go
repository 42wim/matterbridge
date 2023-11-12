package protocol

import (
	"crypto/ecdsa"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/protocol/protobuf"
)

// EditMessage represents an edit of a message from a user in the application layer, used for persistence, querying and
// signaling
type EditMessage struct {
	*protobuf.EditMessage

	// ID is the ID of the message that has been edited
	ID string `json:"id,omitempty"`

	// From is a public key of the author of the edit reaction.
	From string `json:"from,omitempty"`

	// SigPubKey is the ecdsa encoded public key of the edit author
	SigPubKey *ecdsa.PublicKey `json:"-"`

	// LocalChatID is the chatID of the local chat (one-to-one are not symmetric)
	LocalChatID string `json:"localChatId"`
}

func NewEditMessage() *EditMessage {
	return &EditMessage{EditMessage: &protobuf.EditMessage{}}
}

// GetSigPubKey returns an ecdsa encoded public key
// this function is required to implement the ChatEntity interface
func (e *EditMessage) GetSigPubKey() *ecdsa.PublicKey {
	return e.SigPubKey
}

// GetProtoBuf returns the struct's embedded protobuf struct
// this function is required to implement the ChatEntity interface
func (e *EditMessage) GetProtobuf() proto.Message {
	return e.EditMessage
}

// SetMessageType a setter for the MessageType field
// this function is required to implement the ChatEntity interface
func (e *EditMessage) SetMessageType(messageType protobuf.MessageType) {
	e.MessageType = messageType
}

// WrapGroupMessage indicates whether we should wrap this in membership information
func (e *EditMessage) WrapGroupMessage() bool {
	return false
}
