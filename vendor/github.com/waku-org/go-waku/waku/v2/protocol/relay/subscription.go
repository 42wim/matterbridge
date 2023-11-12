package relay

import (
	"context"

	"github.com/waku-org/go-waku/waku/v2/protocol"
	"golang.org/x/exp/slices"
)

// Subscription handles the details of a particular Topic subscription. There may be many subscriptions for a given topic.
type Subscription struct {
	ID            int
	Unsubscribe   func() //for internal use only. For relay Subscription use relay protocol's unsubscribe
	Ch            chan *protocol.Envelope
	contentFilter protocol.ContentFilter
	subType       SubscriptionType
	noConsume     bool
}

type SubscriptionType int

const (
	SpecificContentTopics SubscriptionType = iota
	AllContentTopics
)

// Submit allows a message to be submitted for a subscription
func (s *Subscription) Submit(ctx context.Context, msg *protocol.Envelope) {
	//Filter and notify
	// - if contentFilter doesn't have a contentTopic
	// - if contentFilter has contentTopics and it matches with message
	if !s.noConsume && (len(s.contentFilter.ContentTopicsList()) == 0 ||
		(len(s.contentFilter.ContentTopicsList()) > 0 && slices.Contains[string](s.contentFilter.ContentTopicsList(), msg.Message().ContentTopic))) {
		select {
		case <-ctx.Done():
			return
		case s.Ch <- msg:
		}
	}
}

// NewSubscription creates a subscription that will only receive messages based on the contentFilter
func NewSubscription(contentFilter protocol.ContentFilter) *Subscription {
	ch := make(chan *protocol.Envelope)
	var subType SubscriptionType
	if len(contentFilter.ContentTopicsList()) == 0 {
		subType = AllContentTopics
	}
	return &Subscription{
		Unsubscribe: func() {
			close(ch)
		},
		Ch:            ch,
		contentFilter: contentFilter,
		subType:       subType,
	}
}
