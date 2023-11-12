package subscription

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type SubscriptionsMap struct {
	sync.RWMutex
	logger   *zap.Logger
	items    map[peer.ID]*PeerSubscription
	noOfSubs map[string]map[string]int
}

var ErrNotFound = errors.New("not found")

func NewSubscriptionMap(logger *zap.Logger) *SubscriptionsMap {
	return &SubscriptionsMap{
		logger:   logger.Named("subscription-map"),
		items:    make(map[peer.ID]*PeerSubscription),
		noOfSubs: map[string]map[string]int{},
	}
}

func (m *SubscriptionsMap) Count() int {
        m.RLock()
	defer m.RUnlock()
	return len(m.items)
}

func (m *SubscriptionsMap) IsListening(pubsubTopic, contentTopic string) bool {
	m.RLock()
	defer m.RUnlock()
	return m.noOfSubs[pubsubTopic] != nil && m.noOfSubs[pubsubTopic][contentTopic] > 0
}

func (m *SubscriptionsMap) increaseSubFor(pubsubTopic, contentTopic string) {
	if m.noOfSubs[pubsubTopic] == nil {
		m.noOfSubs[pubsubTopic] = map[string]int{}
	}
	m.noOfSubs[pubsubTopic][contentTopic] = m.noOfSubs[pubsubTopic][contentTopic] + 1
}

func (m *SubscriptionsMap) decreaseSubFor(pubsubTopic, contentTopic string) {
	m.noOfSubs[pubsubTopic][contentTopic] = m.noOfSubs[pubsubTopic][contentTopic] - 1
}

func (sub *SubscriptionsMap) NewSubscription(peerID peer.ID, cf protocol.ContentFilter) *SubscriptionDetails {
	sub.Lock()
	defer sub.Unlock()

	peerSubscription, ok := sub.items[peerID]
	if !ok {
		peerSubscription = &PeerSubscription{
			PeerID:             peerID,
			SubsPerPubsubTopic: make(map[string]SubscriptionSet),
		}
		sub.items[peerID] = peerSubscription
	}

	_, ok = peerSubscription.SubsPerPubsubTopic[cf.PubsubTopic]
	if !ok {
		peerSubscription.SubsPerPubsubTopic[cf.PubsubTopic] = make(SubscriptionSet)
	}

	details := &SubscriptionDetails{
		ID:            uuid.NewString(),
		mapRef:        sub,
		PeerID:        peerID,
		C:             make(chan *protocol.Envelope, 1024),
		ContentFilter: protocol.ContentFilter{PubsubTopic: cf.PubsubTopic, ContentTopics: maps.Clone(cf.ContentTopics)},
	}

	// Increase the number of subscriptions for this (pubsubTopic, contentTopic) pair
	for contentTopic := range cf.ContentTopics {
		sub.increaseSubFor(cf.PubsubTopic, contentTopic)
	}

	sub.items[peerID].SubsPerPubsubTopic[cf.PubsubTopic][details.ID] = details

	return details
}

func (sub *SubscriptionsMap) IsSubscribedTo(peerID peer.ID) bool {
	sub.RLock()
	defer sub.RUnlock()

	_, ok := sub.items[peerID]
	return ok
}

// Check if we have subscriptions for all (pubsubTopic, contentTopics[i]) pairs provided
func (sub *SubscriptionsMap) Has(peerID peer.ID, cf protocol.ContentFilter) bool {
	sub.RLock()
	defer sub.RUnlock()

	// Check if peer exits
	peerSubscription, ok := sub.items[peerID]
	if !ok {
		return false
	}
	//TODO: Handle pubsubTopic as null
	// Check if pubsub topic exists
	subscriptions, ok := peerSubscription.SubsPerPubsubTopic[cf.PubsubTopic]
	if !ok {
		return false
	}

	// Check if the content topic exists within the list of subscriptions for this peer
	for _, ct := range cf.ContentTopicsList() {
		found := false
		for _, subscription := range subscriptions {
			_, exists := subscription.ContentFilter.ContentTopics[ct]
			if exists {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
func (sub *SubscriptionsMap) Delete(subscription *SubscriptionDetails) error {
	sub.Lock()
	defer sub.Unlock()

	peerSubscription, ok := sub.items[subscription.PeerID]
	if !ok {
		return ErrNotFound
	}

	contentFilter := subscription.ContentFilter
	delete(peerSubscription.SubsPerPubsubTopic[contentFilter.PubsubTopic], subscription.ID)

	// Decrease the number of subscriptions for this (pubsubTopic, contentTopic) pair
	for contentTopic := range contentFilter.ContentTopics {
		sub.decreaseSubFor(contentFilter.PubsubTopic, contentTopic)
	}

	return nil
}

func (sub *SubscriptionsMap) clear() {
	for _, peerSubscription := range sub.items {
		for _, subscriptionSet := range peerSubscription.SubsPerPubsubTopic {
			for _, subscription := range subscriptionSet {
				subscription.CloseC()
			}
		}
	}

	sub.items = make(map[peer.ID]*PeerSubscription)
}

func (sub *SubscriptionsMap) Clear() {
	sub.Lock()
	defer sub.Unlock()
	sub.clear()
}

func (sub *SubscriptionsMap) Notify(peerID peer.ID, envelope *protocol.Envelope) {
	sub.RLock()
	defer sub.RUnlock()

	subscriptions, ok := sub.items[peerID].SubsPerPubsubTopic[envelope.PubsubTopic()]
	if ok {
		iterateSubscriptionSet(sub.logger, subscriptions, envelope)
	}
}

func iterateSubscriptionSet(logger *zap.Logger, subscriptions SubscriptionSet, envelope *protocol.Envelope) {
	for _, subscription := range subscriptions {
		func(subscription *SubscriptionDetails) {
			subscription.RLock()
			defer subscription.RUnlock()

			_, ok := subscription.ContentFilter.ContentTopics[envelope.Message().ContentTopic]
			if !ok { // only send the msg to subscriptions that have matching contentTopic
				return
			}

			if !subscription.Closed {
				select {
				case subscription.C <- envelope:
				default:
					logger.Warn("can't deliver message to subscription. subscriber too slow")
				}
			}
		}(subscription)
	}
}

func (m *SubscriptionsMap) GetSubscription(peerID peer.ID, contentFilter protocol.ContentFilter) []*SubscriptionDetails {
	m.RLock()
	defer m.RUnlock()

	var output []*SubscriptionDetails
	for _, peerSubs := range m.items {
		if peerID == "" || peerSubs.PeerID == peerID {
			for _, subs := range peerSubs.SubsPerPubsubTopic {
				for _, subscriptionDetail := range subs {
					if subscriptionDetail.isPartOf(contentFilter) {
						output = append(output, subscriptionDetail)
					}
				}
			}
		}
	}
	return output
}
