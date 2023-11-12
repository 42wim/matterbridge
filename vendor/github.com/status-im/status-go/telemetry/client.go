package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/transport"
	v1protocol "github.com/status-im/status-go/protocol/v1"

	v2protocol "github.com/waku-org/go-waku/waku/v2/protocol"
)

type Client struct {
	serverURL  string
	httpClient *http.Client
	logger     *zap.Logger
	keyUID     string
	nodeName   string
}

func NewClient(logger *zap.Logger, serverURL string, keyUID string, nodeName string) *Client {
	return &Client{
		serverURL:  serverURL,
		httpClient: &http.Client{Timeout: time.Minute},
		logger:     logger,
		keyUID:     keyUID,
		nodeName:   nodeName,
	}
}

func (c *Client) PushReceivedMessages(filter transport.Filter, sshMessage *types.Message, messages []*v1protocol.StatusMessage) {
	c.logger.Debug("Pushing received messages to telemetry server")
	url := fmt.Sprintf("%s/received-messages", c.serverURL)
	var postBody []map[string]interface{}
	for _, message := range messages {
		postBody = append(postBody, map[string]interface{}{
			"chatId":         filter.ChatID,
			"messageHash":    types.EncodeHex(sshMessage.Hash),
			"messageId":      message.ApplicationLayer.ID,
			"sentAt":         sshMessage.Timestamp,
			"pubsubTopic":    filter.PubsubTopic,
			"topic":          filter.ContentTopic.String(),
			"messageType":    message.ApplicationLayer.Type.String(),
			"receiverKeyUID": c.keyUID,
			"nodeName":       c.nodeName,
			"messageSize":    len(sshMessage.Payload),
		})
	}
	body, _ := json.Marshal(postBody)
	_, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("Error sending message to telemetry server", zap.Error(err))
	}
}

func (c *Client) PushReceivedEnvelope(envelope *v2protocol.Envelope) {
	url := fmt.Sprintf("%s/received-envelope", c.serverURL)
	postBody := map[string]interface{}{
		"messageHash":    types.EncodeHex(envelope.Hash()),
		"sentAt":         uint32(envelope.Message().GetTimestamp() / int64(time.Second)),
		"pubsubTopic":    envelope.PubsubTopic(),
		"topic":          envelope.Message().ContentTopic,
		"receiverKeyUID": c.keyUID,
		"nodeName":       c.nodeName,
	}
	body, _ := json.Marshal(postBody)
	_, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("Error sending envelope to telemetry server", zap.Error(err))
	}
}

func (c *Client) UpdateEnvelopeProcessingError(shhMessage *types.Message, processingError error) {
	c.logger.Debug("Pushing envelope update to telemetry server", zap.String("hash", types.EncodeHex(shhMessage.Hash)))
	url := fmt.Sprintf("%s/update-envelope", c.serverURL)
	var errorString = ""
	if processingError != nil {
		errorString = processingError.Error()
	}
	postBody := map[string]interface{}{
		"messageHash":     types.EncodeHex(shhMessage.Hash),
		"sentAt":          shhMessage.Timestamp,
		"pubsubTopic":     shhMessage.PubsubTopic,
		"topic":           shhMessage.Topic,
		"receiverKeyUID":  c.keyUID,
		"nodeName":        c.nodeName,
		"processingError": errorString,
	}
	body, _ := json.Marshal(postBody)
	_, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("Error sending envelope update to telemetry server", zap.Error(err))
	}
}
