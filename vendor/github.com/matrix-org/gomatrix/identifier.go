package gomatrix

// Identifier is the interface for https://matrix.org/docs/spec/client_server/r0.6.0#identifier-types
type Identifier interface {
	// Returns the identifier type
	// https://matrix.org/docs/spec/client_server/r0.6.0#identifier-types
	Type() string
}

// UserIdentifier is the Identifier for https://matrix.org/docs/spec/client_server/r0.6.0#matrix-user-id
type UserIdentifier struct {
	IDType string `json:"type"` // Set by NewUserIdentifer
	User   string `json:"user"`
}

// Type implements the Identifier interface
func (i UserIdentifier) Type() string {
	return "m.id.user"
}

// NewUserIdentifier creates a new UserIdentifier with IDType set to "m.id.user"
func NewUserIdentifier(user string) UserIdentifier {
	return UserIdentifier{
		IDType: "m.id.user",
		User:   user,
	}
}

// ThirdpartyIdentifier is the Identifier for https://matrix.org/docs/spec/client_server/r0.6.0#third-party-id
type ThirdpartyIdentifier struct {
	IDType  string `json:"type"` // Set by NewThirdpartyIdentifier
	Medium  string `json:"medium"`
	Address string `json:"address"`
}

// Type implements the Identifier interface
func (i ThirdpartyIdentifier) Type() string {
	return "m.id.thirdparty"
}

// NewThirdpartyIdentifier creates a new UserIdentifier with IDType set to "m.id.user"
func NewThirdpartyIdentifier(medium, address string) ThirdpartyIdentifier {
	return ThirdpartyIdentifier{
		IDType:  "m.id.thirdparty",
		Medium:  medium,
		Address: address,
	}
}

// PhoneIdentifier is the Identifier for https://matrix.org/docs/spec/client_server/r0.6.0#phone-number
type PhoneIdentifier struct {
	IDType  string `json:"type"` // Set by NewPhoneIdentifier
	Country string `json:"country"`
	Phone   string `json:"phone"`
}

// Type implements the Identifier interface
func (i PhoneIdentifier) Type() string {
	return "m.id.phone"
}

// NewPhoneIdentifier creates a new UserIdentifier with IDType set to "m.id.user"
func NewPhoneIdentifier(country, phone string) PhoneIdentifier {
	return PhoneIdentifier{
		IDType:  "m.id.phone",
		Country: country,
		Phone:   phone,
	}
}
