package subscription

import (
	"encoding/json"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/waku/v2/protocol"
)

// Map of SubscriptionDetails.ID to subscriptions
type SubscriptionSet map[string]*SubscriptionDetails

type PeerSubscription struct {
	PeerID             peer.ID
	SubsPerPubsubTopic map[string]SubscriptionSet
}

type PeerContentFilter struct {
	PeerID        peer.ID  `json:"peerID"`
	PubsubTopic   string   `json:"pubsubTopics"`
	ContentTopics []string `json:"contentTopics"`
}

type SubscriptionDetails struct {
	sync.RWMutex

	ID     string `json:"subscriptionID"`
	mapRef *SubscriptionsMap
	Closed bool `json:"-"`
	once   sync.Once

	PeerID        peer.ID                 `json:"peerID"`
	ContentFilter protocol.ContentFilter  `json:"contentFilters"`
	C             chan *protocol.Envelope `json:"-"`
}

func (s *SubscriptionDetails) Add(contentTopics ...string) {
	s.Lock()
	defer s.Unlock()

	for _, ct := range contentTopics {
		if _, ok := s.ContentFilter.ContentTopics[ct]; !ok {
			s.ContentFilter.ContentTopics[ct] = struct{}{}
			// Increase the number of subscriptions for this (pubsubTopic, contentTopic) pair
			s.mapRef.Lock()
			s.mapRef.increaseSubFor(s.ContentFilter.PubsubTopic, ct)
			s.mapRef.Unlock()
		}
	}
}

func (s *SubscriptionDetails) Remove(contentTopics ...string) {
	s.Lock()
	defer s.Unlock()

	for _, ct := range contentTopics {
		if _, ok := s.ContentFilter.ContentTopics[ct]; ok {
			delete(s.ContentFilter.ContentTopics, ct)
			// Decrease the number of subscriptions for this (pubsubTopic, contentTopic) pair
			s.mapRef.Lock()
			s.mapRef.decreaseSubFor(s.ContentFilter.PubsubTopic, ct)
			s.mapRef.Unlock()
		}
	}

	if len(s.ContentFilter.ContentTopics) == 0 {
		// err doesn't matter
		_ = s.mapRef.Delete(s)
	}
}

// C1 if contentFilter is empty, it means that given subscription is part of contentFilter
// C2 if not empty, check matching pubsubsTopic and atleast 1 contentTopic
func (s *SubscriptionDetails) isPartOf(contentFilter protocol.ContentFilter) bool {
	s.RLock()
	defer s.RUnlock()
	if contentFilter.PubsubTopic != "" && // C1
		s.ContentFilter.PubsubTopic != contentFilter.PubsubTopic { // C2
		return false
	}
	// C1
	if len(contentFilter.ContentTopics) == 0 {
		return true
	}
	// C2
	for cTopic := range contentFilter.ContentTopics {
		if _, ok := s.ContentFilter.ContentTopics[cTopic]; ok {
			return true
		}
	}
	return false
}

func (s *SubscriptionDetails) CloseC() {
	s.once.Do(func() {
		s.Lock()
		defer s.Unlock()

		s.Closed = true
		close(s.C)
	})
}

func (s *SubscriptionDetails) Close() error {
	s.CloseC()
	return s.mapRef.Delete(s)
}

func (s *SubscriptionDetails) MarshalJSON() ([]byte, error) {
	result := struct {
		PeerID        peer.ID  `json:"peerID"`
		PubsubTopic   string   `json:"pubsubTopics"`
		ContentTopics []string `json:"contentTopics"`
	}{
		PeerID:        s.PeerID,
		PubsubTopic:   s.ContentFilter.PubsubTopic,
		ContentTopics: s.ContentFilter.ContentTopics.ToList(),
	}

	return json.Marshal(result)
}
