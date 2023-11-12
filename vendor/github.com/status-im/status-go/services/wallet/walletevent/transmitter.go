package walletevent

import (
	"sync"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/signal"
)

type Publisher interface {
	Subscribe(interface{}) event.Subscription
}

// SignalsTransmitter transmits received events as wallet signals.
type SignalsTransmitter struct {
	Publisher

	wg   sync.WaitGroup
	quit chan struct{}
}

// Start runs loop in background.
func (tmr *SignalsTransmitter) Start() error {
	if tmr.quit != nil {
		// already running, nothing to do
		return nil
	}
	tmr.quit = make(chan struct{})
	events := make(chan Event, 10)
	sub := tmr.Publisher.Subscribe(events)

	tmr.wg.Add(1)
	go func() {
		defer tmr.wg.Done()
		for {
			select {
			case <-tmr.quit:
				sub.Unsubscribe()
				return
			case err := <-sub.Err():
				// technically event.Feed cannot send an error to subscription.Err channel.
				// the only time we will get an event is when that channel is closed.
				if err != nil {
					log.Error("wallet signals transmitter failed with", "error", err)
				}
				return
			case event := <-events:
				if !event.Type.IsInternal() {
					signal.SendWalletEvent(event)
				}
			}
		}
	}()
	return nil
}

// Stop stops the loop and waits till it exits.
func (tmr *SignalsTransmitter) Stop() {
	if tmr.quit == nil {
		return
	}
	close(tmr.quit)
	tmr.wg.Wait()
	tmr.quit = nil
}
