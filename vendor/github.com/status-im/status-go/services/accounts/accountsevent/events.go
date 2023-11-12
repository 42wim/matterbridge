package accountsevent

import "github.com/ethereum/go-ethereum/common"

// EventType type for event types.
type EventType string

// Event is a type for accounts events.
type Event struct {
	Type     EventType        `json:"type"`
	Accounts []common.Address `json:"accounts"`
}

const (
	// EventTypeAdded is emitted when a new account is added.
	EventTypeAdded EventType = "added"
	// EventTypeRemoved is emitted when an account is removed.
	EventTypeRemoved EventType = "removed"
)
