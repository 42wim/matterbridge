package settingsevent

import (
	"context"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/services/wallet/async"
)

type SettingChangeCb func(setting settings.SettingField, value interface{})

// Watcher executes a given callback whenever an account gets added/removed
type Watcher struct {
	feed     *event.Feed
	group    *async.Group
	callback SettingChangeCb
}

func NewWatcher(feed *event.Feed, callback SettingChangeCb) *Watcher {
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

func onSettingChanged(callback SettingChangeCb, setting settings.SettingField, value interface{}) {
	if callback != nil {
		callback(setting, value)
	}
}

func watch(ctx context.Context, feed *event.Feed, callback SettingChangeCb) error {
	ch := make(chan Event, 1)
	sub := feed.Subscribe(ch)
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			if err != nil {
				log.Error("settings watcher subscription failed", "error", err)
			}
		case ev := <-ch:
			if ev.Type == EventTypeChanged {
				onSettingChanged(callback, ev.Setting, ev.Value)
			}
		}
	}
}
