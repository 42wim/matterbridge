package types

// Account represents an Ethereum account located at a specific location defined
// by the optional URL field.
type Account struct {
	Address Address `json:"address"` // Ethereum account address derived from the key
	URL     string  `json:"url"`     // Optional resource locator within a backend
}
