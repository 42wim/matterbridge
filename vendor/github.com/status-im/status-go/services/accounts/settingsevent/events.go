package settingsevent

import "github.com/status-im/status-go/multiaccounts/settings"

// EventType type for event types.
type EventType string

// Event is a type for accounts events.
type Event struct {
	Type    EventType             `json:"type"`
	Setting settings.SettingField `json:"setting"`
	Value   interface{}           `json:"value"`
}

const (
	EventTypeChanged EventType = "changed"
)
