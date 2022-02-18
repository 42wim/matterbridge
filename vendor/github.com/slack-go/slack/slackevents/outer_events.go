// outer_events.go provides EventsAPI particular outer events

package slackevents

import (
	"encoding/json"
)

// EventsAPIEvent is the base EventsAPIEvent
type EventsAPIEvent struct {
	Token        string `json:"token"`
	TeamID       string `json:"team_id"`
	Type         string `json:"type"`
	APIAppID     string `json:"api_app_id"`
	EnterpriseID string `json:"enterprise_id"`
	Data         interface{}
	InnerEvent   EventsAPIInnerEvent
}

// EventsAPIURLVerificationEvent received when configuring a EventsAPI driven app
type EventsAPIURLVerificationEvent struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

// ChallengeResponse is a response to a EventsAPIEvent URLVerification challenge
type ChallengeResponse struct {
	Challenge string
}

// EventsAPICallbackEvent is the main (outer) EventsAPI event.
type EventsAPICallbackEvent struct {
	Type         string           `json:"type"`
	Token        string           `json:"token"`
	TeamID       string           `json:"team_id"`
	APIAppID     string           `json:"api_app_id"`
	InnerEvent   *json.RawMessage `json:"event"`
	AuthedUsers  []string         `json:"authed_users"`
	AuthedTeams  []string         `json:"authed_teams"`
	EventID      string           `json:"event_id"`
	EventTime    int              `json:"event_time"`
	EventContext string           `json:"event_context"`
}

// EventsAPIAppRateLimited indicates your app's event subscriptions are being rate limited
type EventsAPIAppRateLimited struct {
	Type              string `json:"type"`
	Token             string `json:"token"`
	TeamID            string `json:"team_id"`
	MinuteRateLimited int    `json:"minute_rate_limited"`
	APIAppID          string `json:"api_app_id"`
}

const (
	// CallbackEvent is the "outer" event of an EventsAPI event.
	CallbackEvent = "event_callback"
	// URLVerification is an event used when configuring your EventsAPI app
	URLVerification = "url_verification"
	// AppRateLimited indicates your app's event subscriptions are being rate limited
	AppRateLimited = "app_rate_limited"
)

// EventsAPIEventMap maps OUTTER Event API events to their corresponding struct
// implementations. The structs should be instances of the unmarshalling
// target for the matching event type.
var EventsAPIEventMap = map[string]interface{}{
	CallbackEvent:   EventsAPICallbackEvent{},
	URLVerification: EventsAPIURLVerificationEvent{},
	AppRateLimited:  EventsAPIAppRateLimited{},
}
