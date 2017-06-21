package trade

import (
	"github.com/Philipp15b/go-steam/trade/tradeapi"
)

type TradeEndedEvent struct {
	Reason TradeEndReason
}

type TradeEndReason uint

const (
	TradeEndReason_Complete  TradeEndReason = 1
	TradeEndReason_Cancelled                = 2
	TradeEndReason_Timeout                  = 3
	TradeEndReason_Failed                   = 4
)

func newItem(event *tradeapi.Event) *Item {
	return &Item{
		event.AppId,
		event.ContextId,
		event.AssetId,
	}
}

type Item struct {
	AppId     uint32
	ContextId uint64
	AssetId   uint64
}

type ItemAddedEvent struct {
	Item *Item
}

type ItemRemovedEvent struct {
	Item *Item
}

type ReadyEvent struct{}
type UnreadyEvent struct{}

func newCurrency(event *tradeapi.Event) *Currency {
	return &Currency{
		event.AppId,
		event.ContextId,
		event.CurrencyId,
	}
}

type Currency struct {
	AppId      uint32
	ContextId  uint64
	CurrencyId uint64
}

type SetCurrencyEvent struct {
	Currency  *Currency
	OldAmount uint64
	NewAmount uint64
}

type ChatEvent struct {
	Message string
}
