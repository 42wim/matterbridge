package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

// Information whether the MarketMarketItem is available.
const (
	MarketItemAvailable = iota
	MarketItemRemoved
	MarketItemUnavailable
)

// MarketCurrency struct.
type MarketCurrency struct {
	ID    int    `json:"id"`    // Currency ID
	Name  string `json:"name"`  // Currency sign
	Title string `json:"title"` // Currency Title
}

// MarketMarketAlbum struct.
type MarketMarketAlbum struct {
	Count       int         `json:"count"`    // Items number
	ID          int         `json:"id"`       // Market album ID
	OwnerID     int         `json:"owner_id"` // Market album owner's ID
	Photo       PhotosPhoto `json:"photo"`
	Title       string      `json:"title"`        // Market album title
	UpdatedTime int         `json:"updated_time"` // Date when album has been updated last time in Unixtime
	IsMain      BaseBoolInt `json:"is_main"`
	IsHidden    BaseBoolInt `json:"is_hidden"`
}

// ToAttachment return attachment format.
func (marketAlbum MarketMarketAlbum) ToAttachment() string {
	return fmt.Sprintf("market_album%d_%d", marketAlbum.OwnerID, marketAlbum.ID)
}

// MarketMarketCategory struct.
type MarketMarketCategory struct {
	ID      int           `json:"id"`   // Category ID
	Name    string        `json:"name"` // Category name
	Section MarketSection `json:"section"`
}

// MarketMarketItem struct.
type MarketMarketItem struct {
	AccessKey    string               `json:"access_key"`   // Access key for the market item
	Availability int                  `json:"availability"` // Information whether the item is available
	Category     MarketMarketCategory `json:"category"`

	// Date when the item has been created in Unixtime.
	Date               int                        `json:"date,omitempty"`
	Description        string                     `json:"description"` // Item description
	ID                 int                        `json:"id"`          // Item ID
	OwnerID            int                        `json:"owner_id"`    // Item owner's ID
	Price              MarketPrice                `json:"price"`
	ThumbPhoto         string                     `json:"thumb_photo"` // URL of the preview image
	Title              string                     `json:"title"`       // Item title
	CanComment         BaseBoolInt                `json:"can_comment"`
	CanRepost          BaseBoolInt                `json:"can_repost"`
	IsFavorite         BaseBoolInt                `json:"is_favorite"`
	IsMainVariant      BaseBoolInt                `json:"is_main_variant"`
	AlbumsIDs          []int                      `json:"albums_ids"`
	Photos             []PhotosPhoto              `json:"photos"`
	Likes              BaseLikesInfo              `json:"likes"`
	Reposts            BaseRepostsInfo            `json:"reposts"`
	ViewsCount         int                        `json:"views_count,omitempty"`
	URL                string                     `json:"url"` // URL to item
	ButtonTitle        string                     `json:"button_title"`
	ExternalID         string                     `json:"external_id"`
	Dimensions         MarketDimensions           `json:"dimensions"`
	Weight             int                        `json:"weight"`
	VariantsGroupingID int                        `json:"variants_grouping_id"`
	PropertyValues     []MarketMarketItemProperty `json:"property_values"`
	CartQuantity       int                        `json:"cart_quantity"`
	SKU                string                     `json:"sku"`
}

// UnmarshalJSON MarketMarketItem.
//
// BUG(VK): https://github.com/SevereCloud/vksdk/issues/147
func (market *MarketMarketItem) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("false")) {
		return nil
	}

	type renamedMarketMarketItem MarketMarketItem

	var r renamedMarketMarketItem

	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	*market = MarketMarketItem(r)

	return nil
}

// DecodeMsgpack MarketMarketItem.
//
// BUG(VK): https://github.com/SevereCloud/vksdk/issues/147
func (market *MarketMarketItem) DecodeMsgpack(dec *msgpack.Decoder) error {
	data, err := dec.DecodeRaw()
	if err != nil {
		return err
	}

	if bytes.Equal(data, []byte{msgpcode.False}) {
		return nil
	}

	type renamedMarketMarketItem MarketMarketItem

	var r renamedMarketMarketItem

	d := msgpack.NewDecoder(bytes.NewReader(data))
	d.SetCustomStructTag("json")

	err = d.Decode(&r)
	if err != nil {
		return err
	}

	*market = MarketMarketItem(r)

	return nil
}

// MarketMarketItemProperty struct.
type MarketMarketItemProperty struct {
	VariantID    int    `json:"variant_id"`
	VariantName  string `json:"variant_name"`
	PropertyName string `json:"property_name"`
}

// MarketDimensions struct.
type MarketDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Length int `json:"length"`
}

// ToAttachment return attachment format.
func (market MarketMarketItem) ToAttachment() string {
	return fmt.Sprintf("market%d_%d", market.OwnerID, market.ID)
}

// MarketPrice struct.
type MarketPrice struct {
	Amount        string         `json:"amount"` // Amount
	Currency      MarketCurrency `json:"currency"`
	DiscountRate  int            `json:"discount_rate"`
	OldAmount     string         `json:"old_amount"`
	Text          string         `json:"text"` // Text
	OldAmountText string         `json:"old_amount_text"`
}

// UnmarshalJSON MarketPrice.
//
// BUG(VK): unavailable product, in fave.get return [].
func (m *MarketPrice) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("[]")) {
		return nil
	}

	type renamedMarketPrice MarketPrice

	var r renamedMarketPrice

	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	*m = MarketPrice(r)

	return nil
}

// DecodeMsgpack MarketPrice.
//
// BUG(VK): unavailable product, in fave.get return [].
func (m *MarketPrice) DecodeMsgpack(dec *msgpack.Decoder) error {
	data, err := dec.DecodeRaw()
	if err != nil {
		return err
	}

	if bytes.Equal(data, []byte{msgpcode.FixedArrayLow}) {
		return nil
	}

	type renamedMarketPrice MarketPrice

	var r renamedMarketPrice

	d := msgpack.NewDecoder(bytes.NewReader(data))
	d.SetCustomStructTag("json")

	err = d.Decode(&r)
	if err != nil {
		return err
	}

	*m = MarketPrice(r)

	return nil
}

// MarketSection struct.
type MarketSection struct {
	ID   int    `json:"id"`   // Section ID
	Name string `json:"name"` // Section name
}

// MarketOrderStatus order status.
type MarketOrderStatus int

// Possible values.
const (
	MarketOrderNew MarketOrderStatus = iota
	MarketOrderPending
	MarketOrderProcessing
	MarketOrderShipped
	MarketOrderComplete
	MarketOrderCanceled
	MarketOrderRefund
)

// MarketOrder struct.
type MarketOrder struct {
	ID                int                  `json:"id"`
	GroupID           int                  `json:"group_id"`
	UserID            int                  `json:"user_id"`
	Date              int                  `json:"date"`
	Status            MarketOrderStatus    `json:"status"`
	ItemsCount        int                  `json:"items_count"`
	TotalPrice        MarketPrice          `json:"total_price"`
	DisplayOrderID    string               `json:"display_order_id"`
	Comment           string               `json:"comment"`
	PreviewOrderItems []MarketOrderItem    `json:"preview_order_items"`
	PriceDetails      []MarketPriceDetail  `json:"price_details"`
	Delivery          MarketOrderDelivery  `json:"delivery"`
	Recipient         MarketOrderRecipient `json:"recipient"`
}

// MarketOrderDelivery struct.
type MarketOrderDelivery struct {
	TrackNumber   string              `json:"track_number"`
	TrackLink     string              `json:"track_link"`
	Address       string              `json:"address"`
	Type          string              `json:"type"`
	DeliveryPoint MarketDeliveryPoint `json:"delivery_point,omitempty"`
}

// MarketDeliveryPoint struct.
type MarketDeliveryPoint struct {
	ID           int                        `json:"id"`
	ExternalID   string                     `json:"external_id"`
	OutpostOnly  BaseBoolInt                `json:"outpost_only"`
	CashOnly     BaseBoolInt                `json:"cash_only"`
	Address      MarketDeliveryPointAddress `json:"address"`
	DisplayTitle string                     `json:"display_title"`
	ServiceID    int                        `json:"service_id"`
}

// MarketDeliveryPointAddress struct.
type MarketDeliveryPointAddress struct {
	ID             int     `json:"id"`
	Address        string  `json:"address"`
	CityID         int     `json:"city_id"`
	CountryID      int     `json:"country_id"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Phone          string  `json:"phone"`
	Title          string  `json:"title"`
	WorkInfoStatus string  `json:"work_info_status"`
}

// MarketOrderRecipient struct.
type MarketOrderRecipient struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	DisplayText string `json:"display_text"`
}

// MarketOrderItem struct.
type MarketOrderItem struct {
	OwnerID  int              `json:"owner_id"`
	ItemID   int              `json:"item_id"`
	Price    MarketPrice      `json:"price"`
	Quantity int              `json:"quantity"`
	Item     MarketMarketItem `json:"item"`
	Title    string           `json:"title"`
	Photo    PhotosPhoto      `json:"photo"`
	Variants []string         `json:"variants"`
}

// MarketPriceDetail struct.
type MarketPriceDetail struct {
	Title    string      `json:"title"`
	Price    MarketPrice `json:"price"`
	IsAccent BaseBoolInt `json:"is_accent,omitempty"`
}
