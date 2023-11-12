package history

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/async"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

type TransfersLoadedCb func(chainID uint64, addresses []common.Address, block *big.Int)

// Watcher executes a given callback whenever an account gets added/removed
type Watcher struct {
	feed     *event.Feed
	group    *async.Group
	callback TransfersLoadedCb
}

func NewWatcher(feed *event.Feed, callback TransfersLoadedCb) *Watcher {
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

func onTransfersLoaded(callback TransfersLoadedCb, chainID uint64, addresses []common.Address, blockNum *big.Int) {
	if callback != nil {
		callback(chainID, addresses, blockNum)
	}
}

func watch(ctx context.Context, feed *event.Feed, callback TransfersLoadedCb) error {
	ch := make(chan walletevent.Event, 100)
	sub := feed.Subscribe(ch)
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			if err != nil {
				log.Error("history: transfers watcher subscription failed", "error", err)
			}
		case ev := <-ch:
			if ev.Type == transfer.EventNewTransfers {
				onTransfersLoaded(callback, ev.ChainID, ev.Accounts, ev.BlockNumber)
			}
		}
	}
}
