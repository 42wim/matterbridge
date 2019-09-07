package slack

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type WebhookMessage struct {
	Username        string       `json:"username,omitempty"`
	IconEmoji       string       `json:"icon_emoji,omitempty"`
	IconURL         string       `json:"icon_url,omitempty"`
	Channel         string       `json:"channel,omitempty"`
	ThreadTimestamp string       `json:"thread_ts,omitempty"`
	Text            string       `json:"text,omitempty"`
	Attachments     []Attachment `json:"attachments,omitempty"`
	Parse           string       `json:"parse,omitempty"`
}

func PostWebhook(url string, msg *WebhookMessage) error {
	return PostWebhookCustomHTTP(url, http.DefaultClient, msg)
}

func PostWebhookCustomHTTP(url string, httpClient *http.Client, msg *WebhookMessage) error {
	raw, err := json.Marshal(msg)

	if err != nil {
		return errors.Wrap(err, "marshal failed")
	}

	response, err := httpClient.Post(url, "application/json", bytes.NewReader(raw))

	if err != nil {
		return errors.Wrap(err, "failed to post webhook")
	}

	return checkStatusCode(response, discard{})
}
