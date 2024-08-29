package slackevents

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/slack-go/slack"
)

// eventsMap checks both slack.EventsMapping and
// and slackevents.EventsAPIInnerEventMapping. If the event
// exists, returns the unmarshalled struct instance of
// target for the matching event type.
// TODO: Consider moving all events into its own package?
func eventsMap(t string) (interface{}, bool) {
	// Must parse EventsAPI FIRST as both RTM and EventsAPI
	// have a type: "Message" event.
	// TODO: Handle these cases more explicitly.
	v, exists := EventsAPIInnerEventMapping[EventsAPIType(t)]
	if exists {
		return v, exists
	}
	v, exists = slack.EventMapping[t]
	if exists {
		return v, exists
	}
	return v, exists
}

func parseOuterEvent(rawE json.RawMessage) (EventsAPIEvent, error) {
	e := &EventsAPIEvent{}
	err := json.Unmarshal(rawE, e)
	if err != nil {
		return EventsAPIEvent{
			"",
			"",
			"unmarshalling_error",
			"",
			"",
			&slack.UnmarshallingErrorEvent{ErrorObj: err},
			EventsAPIInnerEvent{},
		}, err
	}
	if e.Type == CallbackEvent {
		cbEvent := &EventsAPICallbackEvent{}
		err = json.Unmarshal(rawE, cbEvent)
		if err != nil {
			return EventsAPIEvent{
				"",
				"",
				"unmarshalling_error",
				"",
				"",
				&slack.UnmarshallingErrorEvent{ErrorObj: err},
				EventsAPIInnerEvent{},
			}, err
		}
		return EventsAPIEvent{
			e.Token,
			e.TeamID,
			e.Type,
			e.APIAppID,
			e.EnterpriseID,
			cbEvent,
			EventsAPIInnerEvent{},
		}, nil
	}
	urlVE := &EventsAPIURLVerificationEvent{}
	err = json.Unmarshal(rawE, urlVE)
	if err != nil {
		return EventsAPIEvent{
			"",
			"",
			"unmarshalling_error",
			"",
			"",
			&slack.UnmarshallingErrorEvent{ErrorObj: err},
			EventsAPIInnerEvent{},
		}, err
	}
	return EventsAPIEvent{
		e.Token,
		e.TeamID,
		e.Type,
		e.APIAppID,
		e.EnterpriseID,
		urlVE,
		EventsAPIInnerEvent{},
	}, nil
}

func parseInnerEvent(e *EventsAPICallbackEvent) (EventsAPIEvent, error) {
	iE := &slack.Event{}
	rawInnerJSON := e.InnerEvent
	err := json.Unmarshal(*rawInnerJSON, iE)
	if err != nil {
		return EventsAPIEvent{
			e.Token,
			e.TeamID,
			"unmarshalling_error",
			e.APIAppID,
			e.EnterpriseID,
			&slack.UnmarshallingErrorEvent{ErrorObj: err},
			EventsAPIInnerEvent{},
		}, err
	}
	v, exists := eventsMap(iE.Type)
	if !exists {
		return EventsAPIEvent{
			e.Token,
			e.TeamID,
			iE.Type,
			e.APIAppID,
			e.EnterpriseID,
			nil,
			EventsAPIInnerEvent{},
		}, fmt.Errorf("inner Event does not exist! %s", iE.Type)
	}
	t := reflect.TypeOf(v)
	recvEvent := reflect.New(t).Interface()
	err = json.Unmarshal(*rawInnerJSON, recvEvent)
	if err != nil {
		return EventsAPIEvent{
			e.Token,
			e.TeamID,
			"unmarshalling_error",
			e.APIAppID,
			e.EnterpriseID,
			&slack.UnmarshallingErrorEvent{ErrorObj: err},
			EventsAPIInnerEvent{},
		}, err
	}
	return EventsAPIEvent{
		e.Token,
		e.TeamID,
		e.Type,
		e.APIAppID,
		e.EnterpriseID,
		e,
		EventsAPIInnerEvent{iE.Type, recvEvent},
	}, nil
}

type Config struct {
	VerificationToken string
	TokenVerified     bool
}

type Option func(cfg *Config)

type verifier interface {
	Verify(token string) bool
}

func OptionVerifyToken(v verifier) Option {
	return func(cfg *Config) {
		cfg.TokenVerified = v.Verify(cfg.VerificationToken)
	}
}

// OptionNoVerifyToken skips the check of the Slack verification token
func OptionNoVerifyToken() Option {
	return func(cfg *Config) {
		cfg.TokenVerified = true
	}
}

type TokenComparator struct {
	VerificationToken string
}

func (c TokenComparator) Verify(t string) bool {
	return subtle.ConstantTimeCompare([]byte(c.VerificationToken), []byte(t)) == 1
}

// ParseEvent parses the outer and inner events (if applicable) of an events
// api event returning a EventsAPIEvent type. If the event is a url_verification event,
// the inner event is empty.
func ParseEvent(rawEvent json.RawMessage, opts ...Option) (EventsAPIEvent, error) {
	e, err := parseOuterEvent(rawEvent)
	if err != nil {
		return EventsAPIEvent{}, err
	}

	cfg := &Config{}
	cfg.VerificationToken = e.Token
	for _, opt := range opts {
		opt(cfg)
	}

	if !cfg.TokenVerified {
		return EventsAPIEvent{}, errors.New("invalid verification token")
	}

	if e.Type == CallbackEvent {
		cbEvent := e.Data.(*EventsAPICallbackEvent)
		innerEvent, err := parseInnerEvent(cbEvent)
		if err != nil {
			err := fmt.Errorf("EventsAPI Error parsing inner event: %s, %s", innerEvent.Type, err)
			return EventsAPIEvent{
				"",
				"",
				"unmarshalling_error",
				"",
				"",
				&slack.UnmarshallingErrorEvent{ErrorObj: err},
				EventsAPIInnerEvent{},
			}, err
		}
		return innerEvent, nil
	}
	urlVerificationEvent := &EventsAPIURLVerificationEvent{}
	err = json.Unmarshal(rawEvent, urlVerificationEvent)
	if err != nil {
		return EventsAPIEvent{
			"",
			"",
			"unmarshalling_error",
			"",
			"",
			&slack.UnmarshallingErrorEvent{ErrorObj: err},
			EventsAPIInnerEvent{},
		}, err
	}
	return EventsAPIEvent{
		e.Token,
		e.TeamID,
		e.Type,
		e.APIAppID,
		e.EnterpriseID,
		urlVerificationEvent,
		EventsAPIInnerEvent{},
	}, nil
}

func ParseActionEvent(payloadString string, opts ...Option) (MessageAction, error) {
	byteString := []byte(payloadString)
	action := MessageAction{}
	err := json.Unmarshal(byteString, &action)
	if err != nil {
		return MessageAction{}, errors.New("MessageAction unmarshalling failed")
	}

	cfg := &Config{}
	cfg.VerificationToken = action.Token
	for _, opt := range opts {
		opt(cfg)
	}

	if !cfg.TokenVerified {
		return MessageAction{}, errors.New("invalid verification token")
	} else {
		return action, nil
	}
}
