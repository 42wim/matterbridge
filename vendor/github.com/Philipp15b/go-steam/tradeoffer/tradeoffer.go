/*
Implements methods to interact with the official Trade Offer API.

See: https://developer.valvesoftware.com/wiki/Steam_Web_API/IEconService
*/
package tradeoffer

import (
	"encoding/json"
	"github.com/Philipp15b/go-steam/economy/inventory"
	"github.com/Philipp15b/go-steam/steamid"
)

type TradeOfferState uint

const (
	TradeOfferState_Invalid                  TradeOfferState = 1  // Invalid
	TradeOfferState_Active                                   = 2  // This trade offer has been sent, neither party has acted on it yet.
	TradeOfferState_Accepted                                 = 3  // The trade offer was accepted by the recipient and items were exchanged.
	TradeOfferState_Countered                                = 4  // The recipient made a counter offer
	TradeOfferState_Expired                                  = 5  // The trade offer was not accepted before the expiration date
	TradeOfferState_Canceled                                 = 6  // The sender cancelled the offer
	TradeOfferState_Declined                                 = 7  // The recipient declined the offer
	TradeOfferState_InvalidItems                             = 8  // Some of the items in the offer are no longer available (indicated by the missing flag in the output)
	TradeOfferState_CreatedNeedsConfirmation                 = 9  // The offer hasn't been sent yet and is awaiting email/mobile confirmation. The offer is only visible to the sender.
	TradeOfferState_CanceledBySecondFactor                   = 10 // Either party canceled the offer via email/mobile. The offer is visible to both parties, even if the sender canceled it before it was sent.
	TradeOfferState_InEscrow                                 = 11 // The trade has been placed on hold. The items involved in the trade have all been removed from both parties' inventories and will be automatically delivered in the future.
)

type TradeOfferConfirmationMethod uint

const (
	TradeOfferConfirmationMethod_Invalid   TradeOfferConfirmationMethod = 0
	TradeOfferConfirmationMethod_Email                                  = 1
	TradeOfferConfirmationMethod_MobileApp                              = 2
)

type Asset struct {
	AppId      uint32 `json:",string"`
	ContextId  uint64 `json:",string"`
	AssetId    uint64 `json:",string"`
	CurrencyId uint64 `json:",string"`
	ClassId    uint64 `json:",string"`
	InstanceId uint64 `json:",string"`
	Amount     uint64 `json:",string"`
	Missing    bool
}

type TradeOffer struct {
	TradeOfferId       uint64                       `json:",string"`
	TradeId            uint64                       `json:",string"`
	OtherAccountId     uint32                       `json:"accountid_other"`
	OtherSteamId       steamid.SteamId              `json:"-"`
	Message            string                       `json:"message"`
	ExpirationTime     uint32                       `json:"expiraton_time"`
	State              TradeOfferState              `json:"trade_offer_state"`
	ToGive             []*Asset                     `json:"items_to_give"`
	ToReceive          []*Asset                     `json:"items_to_receive"`
	IsOurOffer         bool                         `json:"is_our_offer"`
	TimeCreated        uint32                       `json:"time_created"`
	TimeUpdated        uint32                       `json:"time_updated"`
	EscrowEndDate      uint32                       `json:"escrow_end_date"`
	ConfirmationMethod TradeOfferConfirmationMethod `json:"confirmation_method"`
}

func (t *TradeOffer) UnmarshalJSON(data []byte) error {
	type Alias TradeOffer
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if t.OtherAccountId == 0 {
		t.OtherSteamId = steamid.SteamId(0)
		return nil
	}
	t.OtherSteamId = steamid.SteamId(uint64(t.OtherAccountId) + 76561197960265728)
	return nil
}

type TradeOffersResult struct {
	Sent         []*TradeOffer `json:"trade_offers_sent"`
	Received     []*TradeOffer `json:"trade_offers_received"`
	Descriptions []*Description
}

type TradeOfferResult struct {
	Offer        *TradeOffer
	Descriptions []*Description
}
type Description struct {
	AppId      uint32 `json:"appid"`
	ClassId    uint64 `json:"classid,string"`
	InstanceId uint64 `json:"instanceid,string"`

	IconUrl      string `json:"icon_url"`
	IconUrlLarge string `json:"icon_url_large"`

	Name           string
	MarketName     string `json:"market_name"`
	MarketHashName string `json:"market_hash_name"`

	// Colors in hex, for example `B2B2B2`
	NameColor       string `json:"name_color"`
	BackgroundColor string `json:"background_color"`

	Type string

	Tradable                  bool   `json:"tradable"`
	Commodity                 bool   `json:"commodity"`
	MarketTradableRestriction uint32 `json:"market_tradable_restriction"`

	Descriptions inventory.DescriptionLines `json:"descriptions"`
	Actions      []*inventory.Action        `json:"actions"`
}
