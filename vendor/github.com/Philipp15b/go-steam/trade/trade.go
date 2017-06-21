package trade

import (
	"errors"
	"time"

	"github.com/Philipp15b/go-steam/steamid"
	"github.com/Philipp15b/go-steam/trade/tradeapi"
)

const pollTimeout = time.Second

type Trade struct {
	ThemId steamid.SteamId

	MeReady, ThemReady bool

	lastPoll     time.Time
	queuedEvents []interface{}
	api          *tradeapi.Trade
}

func New(sessionId, steamLogin, steamLoginSecure string, other steamid.SteamId) *Trade {
	return &Trade{
		other,
		false, false,
		time.Unix(0, 0),
		nil,
		tradeapi.New(sessionId, steamLogin, steamLoginSecure, other),
	}
}

func (t *Trade) Version() uint {
	return t.api.Version
}

// Returns all queued events and removes them from the queue without performing a HTTP request, like Poll() would.
func (t *Trade) Events() []interface{} {
	qe := t.queuedEvents
	t.queuedEvents = nil
	return qe
}

func (t *Trade) onStatus(status *tradeapi.Status) error {
	if !status.Success {
		return errors.New("trade: returned status not successful! error message: " + status.Error)
	}

	if status.NewVersion {
		t.api.Version = status.Version

		t.MeReady = status.Me.Ready == true
		t.ThemReady = status.Them.Ready == true
	}

	switch status.TradeStatus {
	case tradeapi.TradeStatus_Complete:
		t.addEvent(&TradeEndedEvent{TradeEndReason_Complete})
	case tradeapi.TradeStatus_Cancelled:
		t.addEvent(&TradeEndedEvent{TradeEndReason_Cancelled})
	case tradeapi.TradeStatus_Timeout:
		t.addEvent(&TradeEndedEvent{TradeEndReason_Timeout})
	case tradeapi.TradeStatus_Failed:
		t.addEvent(&TradeEndedEvent{TradeEndReason_Failed})
	case tradeapi.TradeStatus_Open:
		// nothing
	default:
		// ignore too
	}

	t.updateEvents(status.Events)
	return nil
}

func (t *Trade) updateEvents(events tradeapi.EventList) {
	if len(events) == 0 {
		return
	}

	var lastLogPos uint
	for i, event := range events {
		if i < t.api.LogPos {
			continue
		}
		if event.SteamId != t.ThemId {
			continue
		}

		if lastLogPos < i {
			lastLogPos = i
		}

		switch event.Action {
		case tradeapi.Action_AddItem:
			t.addEvent(&ItemAddedEvent{newItem(event)})
		case tradeapi.Action_RemoveItem:
			t.addEvent(&ItemRemovedEvent{newItem(event)})
		case tradeapi.Action_Ready:
			t.ThemReady = true
			t.addEvent(new(ReadyEvent))
		case tradeapi.Action_Unready:
			t.ThemReady = false
			t.addEvent(new(UnreadyEvent))
		case tradeapi.Action_SetCurrency:
			t.addEvent(&SetCurrencyEvent{
				newCurrency(event),
				event.OldAmount,
				event.NewAmount,
			})
		case tradeapi.Action_ChatMessage:
			t.addEvent(&ChatEvent{
				event.Text,
			})
		}
	}

	t.api.LogPos = uint(lastLogPos) + 1
}

func (t *Trade) addEvent(event interface{}) {
	t.queuedEvents = append(t.queuedEvents, event)
}
