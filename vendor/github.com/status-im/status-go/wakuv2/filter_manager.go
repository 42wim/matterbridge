package wakuv2

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/status-im/status-go/wakuv2/common"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	node "github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter"
	"github.com/waku-org/go-waku/waku/v2/protocol/subscription"
)

const (
	FilterEventAdded = iota
	FilterEventRemoved
	FilterEventPingResult
	FilterEventSubscribeResult
	FilterEventUnsubscribeResult
	FilterEventGetStats
)

const pingTimeout = 10 * time.Second

type FilterSubs map[string]subscription.SubscriptionSet

type FilterEvent struct {
	eventType int
	filterID  string
	success   bool
	peerID    peer.ID
	tempID    string
	sub       *subscription.SubscriptionDetails
	ch        chan FilterSubs
}

// Methods on FilterManager maintain filter peer health
//
// runFilterLoop is the main event loop
//
// Filter Install/Uninstall events are pushed onto eventChan
// Subscribe, UnsubscribeWithSubscription, IsSubscriptionAlive calls
// are invoked from goroutines and request results pushed onto eventChan
//
// filterSubs is the map of filter IDs to subscriptions

type FilterManager struct {
	ctx              context.Context
	filterSubs       FilterSubs
	eventChan        chan (FilterEvent)
	isFilterSubAlive func(sub *subscription.SubscriptionDetails) error
	getFilter        func(string) *common.Filter
	onNewEnvelopes   func(env *protocol.Envelope) error
	logger           *zap.Logger
	config           *Config
	node             *node.WakuNode
}

func newFilterManager(ctx context.Context, logger *zap.Logger, getFilterFn func(string) *common.Filter, config *Config, onNewEnvelopes func(env *protocol.Envelope) error, node *node.WakuNode) *FilterManager {
	// This fn is being mocked in test
	mgr := new(FilterManager)
	mgr.ctx = ctx
	mgr.logger = logger
	mgr.getFilter = getFilterFn
	mgr.onNewEnvelopes = onNewEnvelopes
	mgr.filterSubs = make(FilterSubs)
	mgr.eventChan = make(chan FilterEvent, 100)
	mgr.config = config
	mgr.node = node
	mgr.isFilterSubAlive = func(sub *subscription.SubscriptionDetails) error {
		ctx, cancel := context.WithTimeout(ctx, pingTimeout)
		defer cancel()
		return mgr.node.FilterLightnode().IsSubscriptionAlive(ctx, sub)
	}

	return mgr
}

func (mgr *FilterManager) runFilterLoop(wg *sync.WaitGroup) {
	defer wg.Done()
	// Use it to ping filter peer(s) periodically
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mgr.ctx.Done():
			mgr.logger.Debug("filter loop stopped")
			return
		case <-ticker.C:
			mgr.pingPeers()
		case ev := <-mgr.eventChan:
			mgr.processEvents(&ev)
		}
	}
}

func (mgr *FilterManager) processEvents(ev *FilterEvent) {
	switch ev.eventType {

	case FilterEventAdded:
		mgr.filterSubs[ev.filterID] = make(subscription.SubscriptionSet)
		mgr.resubscribe(ev.filterID)

	case FilterEventRemoved:
		for _, sub := range mgr.filterSubs[ev.filterID] {
			if sub == nil {
				// Skip temp subs
				continue
			}
			go mgr.unsubscribeFromFilter(ev.filterID, sub)
		}
		delete(mgr.filterSubs, ev.filterID)

	case FilterEventPingResult:
		if ev.success {
			break
		}
		// filterID field is only set when there are no subs to check for this filter,
		// therefore no particular peers that could be unreachable.
		if ev.filterID != "" {
			// Trigger full resubscribe, filter has too few peers
			mgr.logger.Debug("filter has too few subs", zap.String("filterId", ev.filterID))
			mgr.resubscribe(ev.filterID)
			break
		}

		// Delete subs for removed peer
		for filterID, subs := range mgr.filterSubs {
			for _, sub := range subs {
				if sub == nil {
					// Skip temp subs
					continue
				}
				if sub.PeerID == ev.peerID {
					mgr.logger.Debug("filter sub is inactive", zap.String("filterId", filterID), zap.Stringer("peerId", sub.PeerID), zap.String("subID", sub.ID))
					delete(subs, sub.ID)
					go mgr.unsubscribeFromFilter(filterID, sub)
				}
			}
			mgr.resubscribe(filterID)
		}

	case FilterEventSubscribeResult:
		subs, found := mgr.filterSubs[ev.filterID]
		if ev.success {
			if found {
				subs[ev.sub.ID] = ev.sub
				go mgr.runFilterSubscriptionLoop(ev.sub)
			} else {
				// We subscribed to a filter that is already uninstalled; invoke unsubscribe
				go mgr.unsubscribeFromFilter(ev.filterID, ev.sub)
			}
		}
		if found {
			// Delete temp subscription record
			delete(subs, ev.tempID)
		}

	case FilterEventUnsubscribeResult:
		mgr.logger.Debug("filter event unsubscribe result", zap.String("filterId", ev.filterID), zap.Stringer("peerID", ev.sub.PeerID))

	case FilterEventGetStats:
		stats := make(FilterSubs)
		for id, subs := range mgr.filterSubs {
			stats[id] = make(subscription.SubscriptionSet)
			for subID, sub := range subs {
				if sub == nil {
					// Skip temp subs
					continue
				}

				stats[id][subID] = sub
			}
		}
		ev.ch <- stats
	}
}

func (mgr *FilterManager) subscribeToFilter(filterID string, tempID string) {

	logger := mgr.logger.With(zap.String("filterId", filterID))
	f := mgr.getFilter(filterID)
	if f == nil {
		logger.Error("filter subscribeToFilter: No filter found")
		mgr.eventChan <- FilterEvent{eventType: FilterEventSubscribeResult, filterID: filterID, tempID: tempID, success: false}
		return
	}
	contentFilter := mgr.buildContentFilter(f.PubsubTopic, f.ContentTopics)
	logger.Debug("filter subscribe to filter node", zap.String("pubsubTopic", contentFilter.PubsubTopic), zap.Strings("contentTopics", contentFilter.ContentTopicsList()))
	ctx, cancel := context.WithTimeout(mgr.ctx, requestTimeout)
	defer cancel()

	subDetails, err := mgr.node.FilterLightnode().Subscribe(ctx, contentFilter, filter.WithAutomaticPeerSelection())
	var sub *subscription.SubscriptionDetails
	if err != nil {
		logger.Warn("filter could not add wakuv2 filter for peers", zap.Error(err))
	} else {
		sub = subDetails[0]
		logger.Debug("filter subscription success", zap.Stringer("peer", sub.PeerID), zap.String("pubsubTopic", contentFilter.PubsubTopic), zap.Strings("contentTopics", contentFilter.ContentTopicsList()))
	}

	success := err == nil
	mgr.eventChan <- FilterEvent{eventType: FilterEventSubscribeResult, filterID: filterID, tempID: tempID, sub: sub, success: success}
}

func (mgr *FilterManager) unsubscribeFromFilter(filterID string, sub *subscription.SubscriptionDetails) {
	mgr.logger.Debug("filter unsubscribe from filter node", zap.String("filterId", filterID), zap.String("subId", sub.ID), zap.Stringer("peer", sub.PeerID))
	// Unsubscribe on light node
	ctx, cancel := context.WithTimeout(mgr.ctx, requestTimeout)
	defer cancel()
	_, err := mgr.node.FilterLightnode().UnsubscribeWithSubscription(ctx, sub)

	if err != nil {
		mgr.logger.Warn("could not unsubscribe wakuv2 filter for peer", zap.String("filterId", filterID), zap.String("subId", sub.ID), zap.Error(err))
	}

	success := err == nil
	mgr.eventChan <- FilterEvent{eventType: FilterEventUnsubscribeResult, filterID: filterID, success: success, sub: sub}
}

// Check whether each of the installed filters
// has enough alive subscriptions to peers
func (mgr *FilterManager) pingPeers() {
	mgr.logger.Debug("filter pingPeers")

	distinctPeers := make(map[peer.ID]struct{})
	for filterID, subs := range mgr.filterSubs {
		logger := mgr.logger.With(zap.String("filterId", filterID))
		nilSubsCnt := 0
		for _, s := range subs {
			if s == nil {
				nilSubsCnt++
			}
		}
		logger.Debug("filter ping peers", zap.Int("len", len(subs)), zap.Int("len(nilSubs)", nilSubsCnt))
		if len(subs) < mgr.config.MinPeersForFilter {
			// Trigger full resubscribe
			logger.Debug("filter ping peers not enough subs")
			go func(filterID string) {
				mgr.eventChan <- FilterEvent{eventType: FilterEventPingResult, filterID: filterID, success: false}
			}(filterID)
		}
		for _, sub := range subs {
			if sub == nil {
				// Skip temp subs
				continue
			}
			_, found := distinctPeers[sub.PeerID]
			if found {
				continue
			}
			distinctPeers[sub.PeerID] = struct{}{}
			logger.Debug("filter ping peer", zap.Stringer("peerId", sub.PeerID))
			go func(sub *subscription.SubscriptionDetails) {
				err := mgr.isFilterSubAlive(sub)
				alive := err == nil

				if alive {
					logger.Debug("filter aliveness check succeeded", zap.Stringer("peerId", sub.PeerID))
				} else {
					logger.Debug("filter aliveness check failed", zap.Stringer("peerId", sub.PeerID), zap.Error(err))
				}
				mgr.eventChan <- FilterEvent{eventType: FilterEventPingResult, peerID: sub.PeerID, success: alive}
			}(sub)
		}
	}
}

func (mgr *FilterManager) buildContentFilter(pubsubTopic string, contentTopicSet common.TopicSet) protocol.ContentFilter {
	contentTopics := make([]string, len(contentTopicSet))
	for i, ct := range maps.Keys(contentTopicSet) {
		contentTopics[i] = ct.ContentTopic()
	}

	return protocol.NewContentFilter(pubsubTopic, contentTopics...)
}

func (mgr *FilterManager) resubscribe(filterID string) {
	subs, found := mgr.filterSubs[filterID]
	if !found {
		mgr.logger.Error("resubscribe filter not found", zap.String("filterId", filterID))
		return
	}
	if len(subs) > mgr.config.MinPeersForFilter {
		mgr.logger.Error("filter resubscribe too many subs", zap.String("filterId", filterID), zap.Int("len", len(subs)))
	}
	if len(subs) == mgr.config.MinPeersForFilter {
		// do nothing
		return
	}
	mgr.logger.Debug("filter resubscribe subs count:", zap.String("filterId", filterID), zap.Int("len", len(subs)))
	for i := len(subs); i < mgr.config.MinPeersForFilter; i++ {
		mgr.logger.Debug("filter check not passed, try subscribing to peers", zap.String("filterId", filterID))

		// Create sub placeholder in order to avoid potentially too many subs
		tempID := uuid.NewString()
		subs[tempID] = nil
		go mgr.subscribeToFilter(filterID, tempID)
	}
}

func (mgr *FilterManager) runFilterSubscriptionLoop(sub *subscription.SubscriptionDetails) {
	for {
		select {
		case <-mgr.ctx.Done():
			return
		case env, ok := <-sub.C:
			if ok {
				err := (mgr.onNewEnvelopes)(env)
				if err != nil {
					mgr.logger.Error("OnNewEnvelopes error", zap.Error(err))
				}
			} else {
				mgr.logger.Debug("filter sub is closed", zap.String("id", sub.ID))
				return
			}
		}
	}
}
