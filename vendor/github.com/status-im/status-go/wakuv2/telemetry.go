package wakuv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/metrics"
	"go.uber.org/zap"
)

type BandwidthTelemetryClient struct {
	serverURL  string
	httpClient *http.Client
	hostID     string
	logger     *zap.Logger
}

func NewBandwidthTelemetryClient(logger *zap.Logger, serverURL string) *BandwidthTelemetryClient {
	return &BandwidthTelemetryClient{
		serverURL:  serverURL,
		httpClient: &http.Client{Timeout: time.Minute},
		hostID:     uuid.NewString(),
		logger:     logger.Named("bandwidth-telemetry"),
	}
}

func (c *BandwidthTelemetryClient) PushProtocolStats(relayStats metrics.Stats, storeStats metrics.Stats) {
	url := fmt.Sprintf("%s/protocol-stats", c.serverURL)
	postBody := map[string]interface{}{
		"hostID": c.hostID,
		"relay": map[string]interface{}{
			"rateIn":   relayStats.RateIn,
			"rateOut":  relayStats.RateOut,
			"totalIn":  relayStats.TotalIn,
			"totalOut": relayStats.TotalOut,
		},
		"store": map[string]interface{}{
			"rateIn":   storeStats.RateIn,
			"rateOut":  storeStats.RateOut,
			"totalIn":  storeStats.TotalIn,
			"totalOut": storeStats.TotalOut,
		},
	}

	body, _ := json.Marshal(postBody)
	_, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("Error sending message to telemetry server", zap.Error(err))
	}
}
