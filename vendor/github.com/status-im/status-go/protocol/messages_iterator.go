package protocol

import (
	"golang.org/x/exp/maps"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/transport"
)

type MessagesIterator interface {
	HasNext() bool
	Next() (transport.Filter, []*types.Message)
}

type DefaultMessagesIterator struct {
	chatWithMessages map[transport.Filter][]*types.Message
	keys             []transport.Filter
	currentIndex     int
}

func NewDefaultMessagesIterator(chatWithMessages map[transport.Filter][]*types.Message) MessagesIterator {
	return &DefaultMessagesIterator{
		chatWithMessages: chatWithMessages,
		keys:             maps.Keys(chatWithMessages),
		currentIndex:     0,
	}
}

func (it *DefaultMessagesIterator) HasNext() bool {
	return it.currentIndex < len(it.keys)
}

func (it *DefaultMessagesIterator) Next() (transport.Filter, []*types.Message) {
	if it.HasNext() {
		key := it.keys[it.currentIndex]
		it.currentIndex++
		return key, it.chatWithMessages[key]
	}
	return transport.Filter{}, nil
}
