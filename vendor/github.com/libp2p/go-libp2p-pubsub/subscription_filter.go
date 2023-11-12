package pubsub

import (
	"errors"
	"regexp"

	pb "github.com/libp2p/go-libp2p-pubsub/pb"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ErrTooManySubscriptions may be returned by a SubscriptionFilter to signal that there are too many
// subscriptions to process.
var ErrTooManySubscriptions = errors.New("too many subscriptions")

// SubscriptionFilter is a function that tells us whether we are interested in allowing and tracking
// subscriptions for a given topic.
//
// The filter is consulted whenever a subscription notification is received by another peer; if the
// filter returns false, then the notification is ignored.
//
// The filter is also consulted when joining topics; if the filter returns false, then the Join
// operation will result in an error.
type SubscriptionFilter interface {
	// CanSubscribe returns true if the topic is of interest and we can subscribe to it
	CanSubscribe(topic string) bool

	// FilterIncomingSubscriptions is invoked for all RPCs containing subscription notifications.
	// It should filter only the subscriptions of interest and my return an error if (for instance)
	// there are too many subscriptions.
	FilterIncomingSubscriptions(peer.ID, []*pb.RPC_SubOpts) ([]*pb.RPC_SubOpts, error)
}

// WithSubscriptionFilter is a pubsub option that specifies a filter for subscriptions
// in topics of interest.
func WithSubscriptionFilter(subFilter SubscriptionFilter) Option {
	return func(ps *PubSub) error {
		ps.subFilter = subFilter
		return nil
	}
}

// NewAllowlistSubscriptionFilter creates a subscription filter that only allows explicitly
// specified topics for local subscriptions and incoming peer subscriptions.
func NewAllowlistSubscriptionFilter(topics ...string) SubscriptionFilter {
	allow := make(map[string]struct{})
	for _, topic := range topics {
		allow[topic] = struct{}{}
	}

	return &allowlistSubscriptionFilter{allow: allow}
}

type allowlistSubscriptionFilter struct {
	allow map[string]struct{}
}

var _ SubscriptionFilter = (*allowlistSubscriptionFilter)(nil)

func (f *allowlistSubscriptionFilter) CanSubscribe(topic string) bool {
	_, ok := f.allow[topic]
	return ok
}

func (f *allowlistSubscriptionFilter) FilterIncomingSubscriptions(from peer.ID, subs []*pb.RPC_SubOpts) ([]*pb.RPC_SubOpts, error) {
	return FilterSubscriptions(subs, f.CanSubscribe), nil
}

// NewRegexpSubscriptionFilter creates a subscription filter that only allows topics that
// match a regular expression for local subscriptions and incoming peer subscriptions.
//
// Warning: the user should take care to match start/end of string in the supplied regular
// expression, otherwise the filter might match unwanted topics unexpectedly.
func NewRegexpSubscriptionFilter(rx *regexp.Regexp) SubscriptionFilter {
	return &rxSubscriptionFilter{allow: rx}
}

type rxSubscriptionFilter struct {
	allow *regexp.Regexp
}

var _ SubscriptionFilter = (*rxSubscriptionFilter)(nil)

func (f *rxSubscriptionFilter) CanSubscribe(topic string) bool {
	return f.allow.MatchString(topic)
}

func (f *rxSubscriptionFilter) FilterIncomingSubscriptions(from peer.ID, subs []*pb.RPC_SubOpts) ([]*pb.RPC_SubOpts, error) {
	return FilterSubscriptions(subs, f.CanSubscribe), nil
}

// FilterSubscriptions filters (and deduplicates) a list of subscriptions.
// filter should return true if a topic is of interest.
func FilterSubscriptions(subs []*pb.RPC_SubOpts, filter func(string) bool) []*pb.RPC_SubOpts {
	accept := make(map[string]*pb.RPC_SubOpts)

	for _, sub := range subs {
		topic := sub.GetTopicid()

		if !filter(topic) {
			continue
		}

		otherSub, ok := accept[topic]
		if ok {
			if sub.GetSubscribe() != otherSub.GetSubscribe() {
				delete(accept, topic)
			}
		} else {
			accept[topic] = sub
		}
	}

	if len(accept) == 0 {
		return nil
	}

	result := make([]*pb.RPC_SubOpts, 0, len(accept))
	for _, sub := range accept {
		result = append(result, sub)
	}

	return result
}

// WrapLimitSubscriptionFilter wraps a subscription filter with a hard limit in the number of
// subscriptions allowed in an RPC message.
func WrapLimitSubscriptionFilter(filter SubscriptionFilter, limit int) SubscriptionFilter {
	return &limitSubscriptionFilter{filter: filter, limit: limit}
}

type limitSubscriptionFilter struct {
	filter SubscriptionFilter
	limit  int
}

var _ SubscriptionFilter = (*limitSubscriptionFilter)(nil)

func (f *limitSubscriptionFilter) CanSubscribe(topic string) bool {
	return f.filter.CanSubscribe(topic)
}

func (f *limitSubscriptionFilter) FilterIncomingSubscriptions(from peer.ID, subs []*pb.RPC_SubOpts) ([]*pb.RPC_SubOpts, error) {
	if len(subs) > f.limit {
		return nil, ErrTooManySubscriptions
	}

	return f.filter.FilterIncomingSubscriptions(from, subs)
}
