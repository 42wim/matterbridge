package common

import (
	"crypto/ecdsa"

	"github.com/status-im/status-go/protocol/protobuf"
)

type CommKeyExMsgType uint8

const (
	KeyExMsgNone  CommKeyExMsgType = 0
	KeyExMsgReuse CommKeyExMsgType = 1
	KeyExMsgRekey CommKeyExMsgType = 2
)

// RawMessage represent a sent or received message, kept for being able
// to re-send/propagate
type RawMessage struct {
	ID                    string
	LocalChatID           string
	LastSent              uint64
	SendCount             int
	Sent                  bool
	ResendAutomatically   bool
	SkipEncryptionLayer   bool // don't wrap message into ProtocolMessage
	SendPushNotification  bool
	MessageType           protobuf.ApplicationMetadataMessage_Type
	Payload               []byte
	Sender                *ecdsa.PrivateKey
	Recipients            []*ecdsa.PublicKey
	SkipGroupMessageWrap  bool
	SkipApplicationWrap   bool
	SendOnPersonalTopic   bool
	CommunityID           []byte
	CommunityKeyExMsgType CommKeyExMsgType
	Ephemeral             bool
	BeforeDispatch        func(*RawMessage) error
	HashRatchetGroupID    []byte
	PubsubTopic           string
}
