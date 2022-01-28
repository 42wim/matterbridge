package object // import "github.com/SevereCloud/vksdk/v2/object"

// AuthSilentTokenProfile struct.
type AuthSilentTokenProfile struct {
	Token          string      `json:"token"`
	Expires        int         `json:"expires"`
	FirstName      string      `json:"first_name"`
	LastName       string      `json:"last_name"`
	Photo50        string      `json:"photo_50"`
	Photo100       string      `json:"photo_100"`
	Photo200       string      `json:"photo_200"`
	Phone          string      `json:"phone"`
	PhoneValidated interface{} `json:"phone_validated"` // int | bool
	UserID         int         `json:"user_id"`
	IsPartial      BaseBoolInt `json:"is_partial"`
	IsService      BaseBoolInt `json:"is_service"`
}
