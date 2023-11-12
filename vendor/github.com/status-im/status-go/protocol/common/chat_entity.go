package common

import (
	"crypto/ecdsa"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/protocol/protobuf"
)

// ChatEntity is anything that is sendable in a chat.
// Currently it encompass a Message and EmojiReaction.
type ChatEntity interface {
	proto.Message

	GetChatId() string
	GetMessageType() protobuf.MessageType
	GetSigPubKey() *ecdsa.PublicKey
	GetProtobuf() proto.Message
	WrapGroupMessage() bool

	SetMessageType(messageType protobuf.MessageType)
}
