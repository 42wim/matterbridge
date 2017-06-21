package tradeapi

import (
	"encoding/json"
	"github.com/Philipp15b/go-steam/jsont"
	"github.com/Philipp15b/go-steam/steamid"
	"strconv"
)

type Status struct {
	Success     bool
	Error       string
	NewVersion  bool        `json:"newversion"`
	TradeStatus TradeStatus `json:"trade_status"`
	Version     uint
	LogPos      int
	Me          User
	Them        User
	Events      EventList
}

type TradeStatus uint

const (
	TradeStatus_Open      TradeStatus = 0
	TradeStatus_Complete              = 1
	TradeStatus_Empty                 = 2 // when both parties trade no items
	TradeStatus_Cancelled             = 3
	TradeStatus_Timeout               = 4 // the partner timed out
	TradeStatus_Failed                = 5
)

type EventList map[uint]*Event

// The EventList can either be an array or an object of id -> event
func (e *EventList) UnmarshalJSON(data []byte) error {
	// initialize the map if it's nil
	if *e == nil {
		*e = make(EventList)
	}

	o := make(map[string]*Event)
	err := json.Unmarshal(data, &o)
	// it's an object
	if err == nil {
		for is, event := range o {
			i, err := strconv.ParseUint(is, 10, 32)
			if err != nil {
				panic(err)
			}
			(*e)[uint(i)] = event
		}
		return nil
	}

	// it's an array
	var a []*Event
	err = json.Unmarshal(data, &a)
	if err != nil {
		return err
	}
	for i, event := range a {
		(*e)[uint(i)] = event
	}
	return nil
}

type Event struct {
	SteamId   steamid.SteamId `json:",string"`
	Action    Action          `json:",string"`
	Timestamp uint64

	AppId     uint32
	ContextId uint64 `json:",string"`
	AssetId   uint64 `json:",string"`

	Text string // only used for chat messages

	// The following is used for SetCurrency
	CurrencyId uint64 `json:",string"`
	OldAmount  uint64 `json:"old_amount,string"`
	NewAmount  uint64 `json:"amount,string"`
}

type Action uint

const (
	Action_AddItem     Action = 0
	Action_RemoveItem         = 1
	Action_Ready              = 2
	Action_Unready            = 3
	Action_Accept             = 4
	Action_SetCurrency        = 6
	Action_ChatMessage        = 7
)

type User struct {
	Ready             jsont.UintBool
	Confirmed         jsont.UintBool
	SecSinceTouch     int  `json:"sec_since_touch"`
	ConnectionPending bool `json:"connection_pending"`
	Assets            interface{}
	Currency          interface{} // either []*Currency or empty string
}

type Currency struct {
	AppId      uint64 `json:",string"`
	ContextId  uint64 `json:",string"`
	CurrencyId uint64 `json:",string"`
	Amount     uint64 `json:",string"`
}
