package transport

import "github.com/status-im/status-go/eth-node/types"

// TODO: revise fields encoding/decoding. Some are encoded using hexutil and some using encoding/hex.
type Filter struct {
	// ChatID is the identifier of the chat
	ChatID string `json:"chatId"`
	// FilterID the whisper filter id generated
	FilterID string `json:"filterId"`
	// SymKeyID is the symmetric key id used for symmetric filters
	SymKeyID string `json:"symKeyId"`
	// OneToOne tells us if we need to use asymmetric encryption for this chat
	OneToOne bool `json:"oneToOne"`
	// Identity is the public key of the other recipient for non-public filters.
	// It's encoded using encoding/hex.
	Identity string `json:"identity"`
	// PubsubTopic is the waku2 pubsub topic
	PubsubTopic string `json:"pubsubTopic"`
	// ContentTopic is the waku topic
	ContentTopic types.TopicType `json:"topic"`
	// Discovery is whether this is a discovery topic
	Discovery bool `json:"discovery"`
	// Negotiated tells us whether is a negotiated topic
	Negotiated bool `json:"negotiated"`
	// Listen is whether we are actually listening for messages on this chat, or the filter is only created in order to be able to post on the topic
	Listen bool `json:"listen"`
	// Ephemeral indicates that this is an ephemeral filter
	Ephemeral bool `json:"ephemeral"`
	// Priority
	Priority uint64
}

func (c *Filter) IsPublic() bool {
	return !c.OneToOne
}
