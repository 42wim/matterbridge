package slack

import (
	"encoding/json"
)

// InteractionType type of interactions
type InteractionType string

// ActionType type represents the type of action (attachment, block, etc.)
type actionType string

// action is an interface that should be implemented by all callback action types
type action interface {
	actionType() actionType
}

// Types of interactions that can be received.
const (
	InteractionTypeDialogCancellation = InteractionType("dialog_cancellation")
	InteractionTypeDialogSubmission   = InteractionType("dialog_submission")
	InteractionTypeDialogSuggestion   = InteractionType("dialog_suggestion")
	InteractionTypeInteractionMessage = InteractionType("interactive_message")
	InteractionTypeMessageAction      = InteractionType("message_action")
	InteractionTypeBlockActions       = InteractionType("block_actions")
)

// InteractionCallback is sent from slack when a user interactions with a button or dialog.
type InteractionCallback struct {
	Type            InteractionType `json:"type"`
	Token           string          `json:"token"`
	CallbackID      string          `json:"callback_id"`
	ResponseURL     string          `json:"response_url"`
	TriggerID       string          `json:"trigger_id"`
	ActionTs        string          `json:"action_ts"`
	Team            Team            `json:"team"`
	Channel         Channel         `json:"channel"`
	User            User            `json:"user"`
	OriginalMessage Message         `json:"original_message"`
	Message         Message         `json:"message"`
	Name            string          `json:"name"`
	Value           string          `json:"value"`
	MessageTs       string          `json:"message_ts"`
	AttachmentID    string          `json:"attachment_id"`
	ActionCallback  ActionCallbacks `json:"actions"`
	DialogSubmissionCallback
}

// ActionCallback is a convenience struct defined to allow dynamic unmarshalling of
// the "actions" value in Slack's JSON response, which varies depending on block type
type ActionCallbacks struct {
	AttachmentActions []*AttachmentAction
	BlockActions      []*BlockAction
}

// UnmarshalJSON implements the Marshaller interface in order to delegate
// marshalling and allow for proper type assertion when decoding the response
func (a *ActionCallbacks) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	for _, r := range raw {
		var obj map[string]interface{}
		err := json.Unmarshal(r, &obj)
		if err != nil {
			return err
		}

		if _, ok := obj["block_id"].(string); ok {
			action, err := unmarshalAction(r, &BlockAction{})
			if err != nil {
				return err
			}

			a.BlockActions = append(a.BlockActions, action.(*BlockAction))
			return nil
		}

		action, err := unmarshalAction(r, &AttachmentAction{})
		if err != nil {
			return err
		}
		a.AttachmentActions = append(a.AttachmentActions, action.(*AttachmentAction))
	}

	return nil
}

func unmarshalAction(r json.RawMessage, callbackAction action) (action, error) {
	err := json.Unmarshal(r, callbackAction)
	if err != nil {
		return nil, err
	}
	return callbackAction, nil
}
