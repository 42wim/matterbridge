package trade

import (
	"github.com/Philipp15b/go-steam/economy/inventory"
	"github.com/Philipp15b/go-steam/trade/tradeapi"
	"time"
)

type Slot uint

func (t *Trade) action(status *tradeapi.Status, err error) error {
	if err != nil {
		return err
	}
	t.onStatus(status)
	return nil
}

// Returns the next batch of events to process. These can be queued from calls to methods
// like `AddItem` or, if there are no queued events, from a new HTTP request to Steam's API (blocking!).
// If the latter is the case, this method may also sleep before the request
// to conform to the polling interval of the official Steam client.
func (t *Trade) Poll() ([]interface{}, error) {
	if t.queuedEvents != nil {
		return t.Events(), nil
	}

	if d := time.Since(t.lastPoll); d < pollTimeout {
		time.Sleep(pollTimeout - d)
	}
	t.lastPoll = time.Now()

	err := t.action(t.api.GetStatus())
	if err != nil {
		return nil, err
	}

	return t.Events(), nil
}

func (t *Trade) GetTheirInventory(contextId uint64, appId uint32) (*inventory.Inventory, error) {
	return inventory.GetFullInventory(func() (*inventory.PartialInventory, error) {
		return t.api.GetForeignInventory(contextId, appId, nil)
	}, func(start uint) (*inventory.PartialInventory, error) {
		return t.api.GetForeignInventory(contextId, appId, &start)
	})
}

func (t *Trade) GetOwnInventory(contextId uint64, appId uint32) (*inventory.Inventory, error) {
	return t.api.GetOwnInventory(contextId, appId)
}

func (t *Trade) GetMain() (*tradeapi.Main, error) {
	return t.api.GetMain()
}

func (t *Trade) AddItem(slot Slot, item *Item) error {
	return t.action(t.api.AddItem(uint(slot), item.AssetId, item.ContextId, item.AppId))
}

func (t *Trade) RemoveItem(slot Slot, item *Item) error {
	return t.action(t.api.RemoveItem(uint(slot), item.AssetId, item.ContextId, item.AppId))
}

func (t *Trade) Chat(message string) error {
	return t.action(t.api.Chat(message))
}

func (t *Trade) SetCurrency(amount uint, currency *Currency) error {
	return t.action(t.api.SetCurrency(amount, currency.CurrencyId, currency.ContextId, currency.AppId))
}

func (t *Trade) SetReady(ready bool) error {
	return t.action(t.api.SetReady(ready))
}

// This may only be called after a successful `SetReady(true)`.
func (t *Trade) Confirm() error {
	return t.action(t.api.Confirm())
}

func (t *Trade) Cancel() error {
	return t.action(t.api.Cancel())
}
