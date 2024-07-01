package slackevents

import (
	"encoding/json"

	"github.com/slack-go/slack"
)

type MessageActionResponse struct {
	ResponseType    string `json:"response_type"`
	ReplaceOriginal bool   `json:"replace_original"`
	Text            string `json:"text"`
}

type MessageActionEntity struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
	Name   string `json:"name"`
}

type MessageAction struct {
	Type             string                   `json:"type"`
	Actions          []slack.AttachmentAction `json:"actions"`
	CallbackID       string                   `json:"callback_id"`
	Team             MessageActionEntity      `json:"team"`
	Channel          MessageActionEntity      `json:"channel"`
	User             MessageActionEntity      `json:"user"`
	ActionTimestamp  json.Number              `json:"action_ts"`
	MessageTimestamp json.Number              `json:"message_ts"`
	AttachmentID     json.Number              `json:"attachment_id"`
	Token            string                   `json:"token"`
	Message          slack.Message            `json:"message"`
	OriginalMessage  slack.Message            `json:"original_message"`
	ResponseURL      string                   `json:"response_url"`
	TriggerID        string                   `json:"trigger_id"`
}
