//go:build go1.13
// +build go1.13

package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func PostWebhookCustomHTTPContext(ctx context.Context, url string, httpClient *http.Client, msg *WebhookMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("failed new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post webhook: %w", err)
	}
	defer resp.Body.Close()

	return checkStatusCode(resp, discard{})
}
