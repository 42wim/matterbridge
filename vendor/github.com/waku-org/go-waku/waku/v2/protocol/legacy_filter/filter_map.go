package legacy_filter

import (
	"sync"

	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/timesource"
)

type FilterMap struct {
	sync.RWMutex
	timesource  timesource.Timesource
	items       map[string]Filter
	broadcaster relay.Broadcaster
}

type FilterMapItem struct {
	Key   string
	Value Filter
}

func NewFilterMap(broadcaster relay.Broadcaster, timesource timesource.Timesource) *FilterMap {
	return &FilterMap{
		timesource:  timesource,
		items:       make(map[string]Filter),
		broadcaster: broadcaster,
	}
}

func (fm *FilterMap) Set(key string, value Filter) {
	fm.Lock()
	defer fm.Unlock()

	fm.items[key] = value
}

func (fm *FilterMap) Get(key string) (Filter, bool) {
	fm.Lock()
	defer fm.Unlock()

	value, ok := fm.items[key]

	return value, ok
}

func (fm *FilterMap) Delete(key string) {
	fm.Lock()
	defer fm.Unlock()

	_, ok := fm.items[key]
	if !ok {
		return
	}

	close(fm.items[key].Chan)
	delete(fm.items, key)
}

func (fm *FilterMap) RemoveAll() {
	fm.Lock()
	defer fm.Unlock()

	for k, v := range fm.items {
		close(v.Chan)
		delete(fm.items, k)
	}
}

func (fm *FilterMap) Items() <-chan FilterMapItem {
	c := make(chan FilterMapItem)

	f := func() {
		fm.RLock()
		defer fm.RUnlock()

		for k, v := range fm.items {
			c <- FilterMapItem{k, v}
		}
		close(c)
	}
	go f()

	return c
}

// Notify is used to push a received message from a filter subscription to
// any content filter registered on this node and to the broadcast subscribers
func (fm *FilterMap) Notify(msg *pb.WakuMessage, requestID string) {
	fm.RLock()
	defer fm.RUnlock()

	filter, ok := fm.items[requestID]
	if !ok {
		// We do this because the key for the filter is set to the requestID received from the filter protocol.
		// This means we do not need to check the content filter explicitly as all MessagePushs already contain
		// the requestID of the coresponding filter.
		return
	}

	envelope := protocol.NewEnvelope(msg, fm.timesource.Now().UnixNano(), filter.Topic)

	// Broadcasting message so it's stored
	fm.broadcaster.Submit(envelope)

	// TODO: In case of no topics we should either trigger here for all messages,
	// or we should not allow such filter to exist in the first place.
	for _, contentTopic := range filter.ContentFilters {
		if msg.ContentTopic == contentTopic {
			filter.Chan <- envelope
			break
		}
	}
}
