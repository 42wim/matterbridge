package object // import "github.com/SevereCloud/vksdk/v2/object"

// GiftsGift Privacy type.
const (
	GiftsGiftPrivacyForAll        = iota // name and message for all
	GiftsGiftPrivacyNameForAll           // name for all
	GiftsGiftPrivacyRecipientOnly        // name and message for recipient only
)

// GiftsGift struct.
type GiftsGift struct {
	Date     int         `json:"date"`    // Date when gist has been sent in Unixtime
	FromID   int         `json:"from_id"` // Gift sender ID
	Gift     GiftsLayout `json:"gift"`
	GiftHash string      `json:"gift_hash"` // Hash
	ID       int         `json:"id"`        // Gift ID
	Message  string      `json:"message"`   // Comment text
	Privacy  int         `json:"privacy"`

	Description string `json:"description"`
	PaymentType string `json:"payment_type"`
	Price       int    `json:"price"`
	PriceStr    string `json:"price_str"`
}

// GiftsLayout struct.
type GiftsLayout struct {
	ID                int         `json:"id"`
	Thumb256          string      `json:"thumb_256"` // URL of the preview image with 256 px in width
	Thumb48           string      `json:"thumb_48"`  // URL of the preview image with 48 px in width
	Thumb96           string      `json:"thumb_96"`  // URL of the preview image with 96 px in width
	StickersProductID int         `json:"stickers_product_id"`
	IsStickersStyle   BaseBoolInt `json:"is_stickers_style"`
}
