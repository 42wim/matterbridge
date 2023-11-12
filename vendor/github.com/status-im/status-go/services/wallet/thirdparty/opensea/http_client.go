package opensea

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

const requestTimeout = 5 * time.Second
const getRequestRetryMaxCount = 15
const getRequestWaitTime = 300 * time.Millisecond

type HTTPClient struct {
	client         *http.Client
	getRequestLock sync.RWMutex
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (o *HTTPClient) doGetRequest(ctx context.Context, url string, apiKey string) ([]byte, error) {
	// Ensure only one thread makes a request at a time
	o.getRequestLock.Lock()
	defer o.getRequestLock.Unlock()

	retryCount := 0
	statusCode := http.StatusOK

	// Try to do the request without an apiKey first
	tmpAPIKey := ""

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:96.0) Gecko/20100101 Firefox/96.0")
		if len(tmpAPIKey) > 0 {
			req.Header.Set("X-API-KEY", tmpAPIKey)
		}

		resp, err := o.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Error("failed to close opensea request body", "err", err)
			}
		}()

		statusCode = resp.StatusCode
		switch resp.StatusCode {
		case http.StatusOK:
			body, err := ioutil.ReadAll(resp.Body)
			return body, err
		case http.StatusBadRequest:
			// The OpenSea v2 API will return error 400 if the account holds no collectibles on
			// the requested chain. This shouldn't be treated as an error, return an empty body.
			return nil, nil
		case http.StatusTooManyRequests:
			if retryCount < getRequestRetryMaxCount {
				// sleep and retry
				time.Sleep(getRequestWaitTime)
				retryCount++
				continue
			}
			// break and error
		case http.StatusForbidden:
			// Request requires an apiKey, set it and retry
			if tmpAPIKey == "" && apiKey != "" {
				tmpAPIKey = apiKey
				// sleep and retry
				time.Sleep(getRequestWaitTime)
				continue
			}
			// break and error
		default:
			// break and error
		}
		break
	}
	return nil, fmt.Errorf("unsuccessful request: %d %s", statusCode, http.StatusText(statusCode))
}
