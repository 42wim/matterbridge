package protocol

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"time"

	"go.uber.org/zap"

	"github.com/status-im/status-go/protocol/common"
)

const (
	DefaultRequestTimeout = 15000 * time.Millisecond

	headerAcceptJSON = "application/json; charset=utf-8"
	headerAcceptText = "text/html; charset=utf-8"

	// Without a particular user agent, many providers treat status-go as a
	// gluttony bot, and either respond more frequently with a 429 (Too Many
	// Requests), or simply refuse to return valid data. Note that using a known
	// browser UA doesn't work well with some providers, such as Spotify,
	// apparently they still flag status-go as a bad actor.
	headerUserAgent = "status-go/v0.151.15"

	// Currently set to English, but we could make this setting dynamic according
	// to the user's language of choice.
	headerAcceptLanguage = "en-US,en;q=0.5"
)

type Headers map[string]string

type Unfurler interface {
	Unfurl() (*common.LinkPreview, error)
}

func newDefaultLinkPreview(url *neturl.URL) *common.LinkPreview {
	return &common.LinkPreview{
		URL:      url.String(),
		Hostname: url.Hostname(),
	}
}

func fetchBody(logger *zap.Logger, httpClient *http.Client, url string, headers Headers) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Error("failed to close response body", zap.Error(err))
		}
	}()

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("http request failed, statusCode='%d'", res.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body bytes: %w", err)
	}

	return bodyBytes, nil
}
