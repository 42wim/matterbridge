package slack

import (
	"context"
	"encoding/json"
	"errors"
)

// InputType is the type of the dialog input type
type InputType string

const (
	// InputTypeText textfield input
	InputTypeText InputType = "text"
	// InputTypeTextArea textarea input
	InputTypeTextArea InputType = "textarea"
	// InputTypeSelect textfield input
	InputTypeSelect InputType = "select"
)

// DialogInput for dialogs input type text or menu
type DialogInput struct {
	Type        InputType `json:"type"`
	Label       string    `json:"label"`
	Name        string    `json:"name"`
	Placeholder string    `json:"placeholder"`
	Optional    bool      `json:"optional"`
}

// DialogTrigger ...
type DialogTrigger struct {
	TriggerID string `json:"trigger_id"` //Required. Must respond within 3 seconds.
	Dialog    Dialog `json:"dialog"`     //Required.
}

// Dialog as in Slack dialogs
// https://api.slack.com/dialogs#option_element_attributes#top-level_dialog_attributes
type Dialog struct {
	TriggerID      string          `json:"trigger_id"`      // Required
	CallbackID     string          `json:"callback_id"`     // Required
	State          string          `json:"state,omitempty"` // Optional
	Title          string          `json:"title"`
	SubmitLabel    string          `json:"submit_label,omitempty"`
	NotifyOnCancel bool            `json:"notify_on_cancel"`
	Elements       []DialogElement `json:"elements"`
}

// DialogElement abstract type for dialogs.
type DialogElement interface{}

// DialogCallback DEPRECATED use InteractionCallback
type DialogCallback InteractionCallback

// DialogSubmissionCallback is sent from Slack when a user submits a form from within a dialog
type DialogSubmissionCallback struct {
	State      string            `json:"state,omitempty"`
	Submission map[string]string `json:"submission"`
}

// DialogOpenResponse response from `dialog.open`
type DialogOpenResponse struct {
	SlackResponse
	DialogResponseMetadata DialogResponseMetadata `json:"response_metadata"`
}

// DialogResponseMetadata lists the error messages
type DialogResponseMetadata struct {
	Messages []string `json:"messages"`
}

// OpenDialog opens a dialog window where the triggerID originated from.
// EXPERIMENTAL: dialog functionality is currently experimental, api is not considered stable.
func (api *Client) OpenDialog(triggerID string, dialog Dialog) (err error) {
	return api.OpenDialogContext(context.Background(), triggerID, dialog)
}

// OpenDialogContext opens a dialog window where the triggerId originated from with a custom context
// EXPERIMENTAL: dialog functionality is currently experimental, api is not considered stable.
func (api *Client) OpenDialogContext(ctx context.Context, triggerID string, dialog Dialog) (err error) {
	if triggerID == "" {
		return errors.New("received empty parameters")
	}

	req := DialogTrigger{
		TriggerID: triggerID,
		Dialog:    dialog,
	}

	encoded, err := json.Marshal(req)
	if err != nil {
		return err
	}

	response := &DialogOpenResponse{}
	endpoint := APIURL + "dialog.open"
	if err := postJSON(ctx, api.httpclient, endpoint, api.token, encoded, response, api); err != nil {
		return err
	}

	return response.Err()
}
