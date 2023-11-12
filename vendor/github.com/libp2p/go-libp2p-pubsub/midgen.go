package pubsub

import (
	"sync"

	pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

// msgIDGenerator handles computing IDs for msgs
// It allows setting custom generators(MsgIdFunction) per topic
type msgIDGenerator struct {
	Default MsgIdFunction

	topicGensLk sync.RWMutex
	topicGens   map[string]MsgIdFunction
}

func newMsgIdGenerator() *msgIDGenerator {
	return &msgIDGenerator{
		Default:   DefaultMsgIdFn,
		topicGens: make(map[string]MsgIdFunction),
	}
}

// Set sets custom id generator(MsgIdFunction) for topic.
func (m *msgIDGenerator) Set(topic string, gen MsgIdFunction) {
	m.topicGensLk.Lock()
	m.topicGens[topic] = gen
	m.topicGensLk.Unlock()
}

// ID computes ID for the msg or short-circuits with the cached value.
func (m *msgIDGenerator) ID(msg *Message) string {
	if msg.ID != "" {
		return msg.ID
	}

	msg.ID = m.RawID(msg.Message)
	return msg.ID
}

// RawID computes ID for the proto 'msg'.
func (m *msgIDGenerator) RawID(msg *pb.Message) string {
	m.topicGensLk.RLock()
	gen, ok := m.topicGens[msg.GetTopic()]
	m.topicGensLk.RUnlock()
	if !ok {
		gen = m.Default
	}

	return gen(msg)
}
