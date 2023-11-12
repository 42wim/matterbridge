package walletevent

import (
	"context"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/async"
)

type EventCb func(event Event)

// Watcher executes a given callback whenever a wallet event gets sent
type Watcher struct {
	feed     *event.Feed
	group    *async.Group
	callback EventCb
}

func NewWatcher(feed *event.Feed, callback EventCb) *Watcher {
	return &Watcher{
		feed:     feed,
		callback: callback,
	}
}

func (w *Watcher) Start() {
	if w.group != nil {
		return
	}

	w.group = async.NewGroup(context.Background())
	w.group.Add(func(ctx context.Context) error {
		return watch(ctx, w.feed, w.callback)
	})
}

func (w *Watcher) Stop() {
	if w.group != nil {
		w.group.Stop()
		w.group.Wait()
		w.group = nil
	}
}

func watch(ctx context.Context, feed *event.Feed, callback EventCb) error {
	ch := make(chan Event, 10)
	sub := feed.Subscribe(ch)
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			if err != nil {
				log.Error("wallet event watcher subscription failed", "error", err)
			}
		case ev := <-ch:
			if callback != nil {
				callback(ev)
			}
		}
	}
}
