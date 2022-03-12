//go:build !go1.13
// +build !go1.13

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
		return fmt.Errorf("marshal failed: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("failed new request: %v", err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post webhook: %v", err)
	}
	defer resp.Body.Close()

	return checkStatusCode(resp, discard{})
}
