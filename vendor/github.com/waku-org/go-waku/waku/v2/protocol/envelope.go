package protocol

import (
	"github.com/waku-org/go-waku/waku/v2/hash"
	wpb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/store/pb"
)

// Envelope contains information about the pubsub topic of a WakuMessage
// and a hash used to identify a message based on the bytes of a WakuMessage
// protobuffer
type Envelope struct {
	msg   *wpb.WakuMessage
	hash  []byte
	index *pb.Index
}

// NewEnvelope creates a new Envelope that contains a WakuMessage
// It's used as a way to know to which Pubsub topic belongs a WakuMessage
// as well as generating a hash based on the bytes that compose the message
func NewEnvelope(msg *wpb.WakuMessage, receiverTime int64, pubSubTopic string) *Envelope {
	messageHash := msg.Hash(pubSubTopic)
	digest := hash.SHA256([]byte(msg.ContentTopic), msg.Payload)
	return &Envelope{
		msg:  msg,
		hash: messageHash,
		index: &pb.Index{
			Digest:       digest[:],
			ReceiverTime: receiverTime,
			SenderTime:   msg.GetTimestamp(),
			PubsubTopic:  pubSubTopic,
		},
	}
}

// Message returns the WakuMessage associated to an Envelope
func (e *Envelope) Message() *wpb.WakuMessage {
	return e.msg
}

// PubsubTopic returns the topic on which a WakuMessage was received
func (e *Envelope) PubsubTopic() string {
	return e.index.PubsubTopic
}

// Hash returns a 32 byte hash calculated from the WakuMessage bytes
func (e *Envelope) Hash() []byte {
	return e.hash
}

func (e *Envelope) Index() *pb.Index {
	return e.index
}
