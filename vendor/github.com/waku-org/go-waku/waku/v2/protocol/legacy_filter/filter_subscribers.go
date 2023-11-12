package legacy_filter

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/waku/v2/protocol/legacy_filter/pb"
)

type Subscriber struct {
	peer      peer.ID
	requestID string
	filter    *pb.FilterRequest // @TODO MAKE THIS A SEQUENCE AGAIN?
}

func (sub Subscriber) HasContentTopic(topic string) bool {
	if len(sub.filter.ContentFilters) == 0 {
		return true // When the subscriber has no specific ContentTopic filter
	}

	for _, filter := range sub.filter.ContentFilters {
		if filter.ContentTopic == topic {
			return true
		}
	}
	return false
}

type Subscribers struct {
	sync.RWMutex
	subscribers []Subscriber
	timeout     time.Duration
	failedPeers map[peer.ID]time.Time
}

func NewSubscribers(timeout time.Duration) *Subscribers {
	return &Subscribers{
		timeout:     timeout,
		failedPeers: make(map[peer.ID]time.Time),
	}
}

func (sub *Subscribers) Clear() {
	sub.Lock()
	defer sub.Unlock()

	sub.subscribers = nil
	sub.failedPeers = make(map[peer.ID]time.Time)
}

func (sub *Subscribers) Append(s Subscriber) int {
	sub.Lock()
	defer sub.Unlock()

	sub.subscribers = append(sub.subscribers, s)
	return len(sub.subscribers)
}

func (sub *Subscribers) Items(contentTopic *string) <-chan Subscriber {
	c := make(chan Subscriber)

	f := func() {
		sub.RLock()
		defer sub.RUnlock()
		for _, s := range sub.subscribers {
			if contentTopic == nil || s.HasContentTopic(*contentTopic) {
				c <- s
			}
		}
		close(c)
	}
	go f()

	return c
}

func (sub *Subscribers) Length() int {
	sub.RLock()
	defer sub.RUnlock()

	return len(sub.subscribers)
}

func (sub *Subscribers) IsFailedPeer(peerID peer.ID) bool {
	sub.RLock()
	defer sub.RUnlock()
	_, ok := sub.failedPeers[peerID]
	return ok
}

func (sub *Subscribers) FlagAsSuccess(peerID peer.ID) {
	sub.Lock()
	defer sub.Unlock()

	_, ok := sub.failedPeers[peerID]
	if ok {
		delete(sub.failedPeers, peerID)
	}
}

func (sub *Subscribers) FlagAsFailure(peerID peer.ID) {
	sub.Lock()
	defer sub.Unlock()

	lastFailure, ok := sub.failedPeers[peerID]
	if ok {
		elapsedTime := time.Since(lastFailure)
		if elapsedTime > sub.timeout {
			var tmpSubs []Subscriber
			for _, s := range sub.subscribers {
				if s.peer != peerID {
					tmpSubs = append(tmpSubs, s)
				}
			}
			sub.subscribers = tmpSubs

			delete(sub.failedPeers, peerID)
		}
	} else {
		sub.failedPeers[peerID] = time.Now()
	}
}

// RemoveContentFilters removes a set of content filters registered for an specific peer
func (sub *Subscribers) RemoveContentFilters(peerID peer.ID, requestID string, contentFilters []*pb.FilterRequest_ContentFilter) {
	sub.Lock()
	defer sub.Unlock()

	var peerIdsToRemove []peer.ID

	for subIndex, subscriber := range sub.subscribers {
		if subscriber.peer != peerID || subscriber.requestID != requestID {
			continue
		}

		// make sure we delete the content filter
		// if no more topics are left
		for _, contentFilter := range contentFilters {
			subCfs := subscriber.filter.ContentFilters
			for i, cf := range subCfs {
				if cf.ContentTopic == contentFilter.ContentTopic {
					l := len(subCfs) - 1
					subCfs[i] = subCfs[l]
					subscriber.filter.ContentFilters = subCfs[:l]
				}
			}
			sub.subscribers[subIndex] = subscriber
		}

		if len(subscriber.filter.ContentFilters) == 0 {
			peerIdsToRemove = append(peerIdsToRemove, subscriber.peer)
		}
	}

	// make sure we delete the subscriber
	// if no more content filters left
	for _, peerID := range peerIdsToRemove {
		for i, s := range sub.subscribers {
			if s.peer == peerID && s.requestID == requestID {
				l := len(sub.subscribers) - 1
				sub.subscribers[i] = sub.subscribers[l]
				sub.subscribers = sub.subscribers[:l]
			}
		}
	}
}
