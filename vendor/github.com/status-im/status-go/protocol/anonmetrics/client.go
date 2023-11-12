package anonmetrics

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/appmetrics"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

const ActiveClientPhrase = "yes i am wanting the activation of the anon metrics client, please thank you lots thank you"

type ClientConfig struct {
	ShouldSend  bool
	SendAddress *ecdsa.PublicKey
	Active      string
}

type Client struct {
	Config   *ClientConfig
	DB       *appmetrics.Database
	Identity *ecdsa.PrivateKey
	Logger   *zap.Logger

	//messageSender is a message processor used to send metric batch messages
	messageSender *common.MessageSender

	IntervalInc *FibonacciIntervalIncrementer

	// mainLoopQuit is a channel that concurrently orchestrates that the main loop that should be terminated
	mainLoopQuit chan struct{}

	// deleteLoopQuit is a channel that concurrently orchestrates that the delete loop that should be terminated
	deleteLoopQuit chan struct{}

	// DBLock prevents deletion of DB items during mainloop
	DBLock sync.Mutex
}

func NewClient(sender *common.MessageSender) *Client {
	return &Client{
		messageSender: sender,
		IntervalInc: &FibonacciIntervalIncrementer{
			Last:    0,
			Current: 1,
		},
	}
}

func (c *Client) sendUnprocessedMetrics() {
	if c.Config.Active != ActiveClientPhrase {
		return
	}

	c.Logger.Debug("sendUnprocessedMetrics() triggered")

	c.DBLock.Lock()
	defer c.DBLock.Unlock()

	// Get all unsent metrics grouped by session id
	uam, err := c.DB.GetUnprocessedGroupedBySession()
	if err != nil {
		c.Logger.Error("failed to get unprocessed messages grouped by session", zap.Error(err))
	}
	c.Logger.Debug("unprocessed metrics from db", zap.Reflect("uam", uam))

	for session, batch := range uam {
		c.Logger.Debug("processing uam from session", zap.String("session", session))

		// Convert the metrics into protobuf
		amb, err := adaptModelsToProtoBatch(batch, &c.Identity.PublicKey)
		if err != nil {
			c.Logger.Error("failed to adapt models to protobuf batch", zap.Error(err))
			return
		}

		// Generate an ephemeral key per session id
		ephemeralKey, err := crypto.GenerateKey()
		if err != nil {
			c.Logger.Error("failed to generate an ephemeral key", zap.Error(err))
			return
		}

		// Prepare the protobuf message
		encodedMessage, err := proto.Marshal(amb)
		if err != nil {
			c.Logger.Error("failed to marshal protobuf", zap.Error(err))
			return
		}
		rawMessage := common.RawMessage{
			Payload:             encodedMessage,
			Sender:              ephemeralKey,
			SkipEncryptionLayer: true,
			SendOnPersonalTopic: true,
			MessageType:         protobuf.ApplicationMetadataMessage_ANONYMOUS_METRIC_BATCH,
		}

		c.Logger.Debug("rawMessage prepared from unprocessed anonymous metrics", zap.Reflect("rawMessage", rawMessage))

		// Send the metrics batch
		_, err = c.messageSender.SendPrivate(context.Background(), c.Config.SendAddress, &rawMessage)
		if err != nil {
			c.Logger.Error("failed to send metrics batch message", zap.Error(err))
			return
		}

		// Mark metrics as processed
		err = c.DB.SetToProcessed(batch)
		if err != nil {
			c.Logger.Error("failed to set metrics as processed in db", zap.Error(err))
		}
	}
}

func (c *Client) mainLoop() error {
	if c.Config.Active != ActiveClientPhrase {
		return nil
	}

	c.Logger.Debug("mainLoop() triggered")

	for {
		c.sendUnprocessedMetrics()

		waitFor := time.Duration(c.IntervalInc.Next()) * time.Second
		c.Logger.Debug("mainLoop() wait interval set", zap.Duration("waitFor", waitFor))
		select {
		case <-time.After(waitFor):
		case <-c.mainLoopQuit:
			return nil
		}
	}
}

func (c *Client) startMainLoop() {
	if c.Config.Active != ActiveClientPhrase {
		return
	}

	c.Logger.Debug("startMainLoop() triggered")

	c.stopMainLoop()
	c.mainLoopQuit = make(chan struct{})
	go func() {
		c.Logger.Debug("startMainLoop() anonymous go routine triggered")
		err := c.mainLoop()
		if err != nil {
			c.Logger.Error("main loop exited with an error", zap.Error(err))
		}
	}()
}

func (c *Client) deleteLoop() error {
	// Sleep to give the main lock time to process any old messages
	time.Sleep(time.Second * 10)

	for {
		func() {
			c.DBLock.Lock()
			defer c.DBLock.Unlock()

			oneWeekAgo := time.Now().Add(time.Hour * 24 * 7 * -1)
			err := c.DB.DeleteOlderThan(&oneWeekAgo)
			if err != nil {
				c.Logger.Error("failed to delete metrics older than given time",
					zap.Time("time given", oneWeekAgo),
					zap.Error(err))
			}
		}()

		select {
		case <-time.After(time.Hour):
		case <-c.deleteLoopQuit:
			return nil
		}
	}
}

func (c *Client) startDeleteLoop() {
	c.stopDeleteLoop()
	c.deleteLoopQuit = make(chan struct{})
	go func() {
		err := c.deleteLoop()
		if err != nil {
			c.Logger.Error("delete loop exited with an error", zap.Error(err))
		}
	}()
}

func (c *Client) Start() error {
	c.Logger.Debug("Main Start() triggered")
	if c.messageSender == nil {
		return errors.New("can't start, missing message processor")
	}

	c.startMainLoop()
	c.startDeleteLoop()
	return nil
}

func (c *Client) stopMainLoop() {
	c.Logger.Debug("stopMainLoop() triggered")

	if c.mainLoopQuit != nil {
		c.Logger.Debug("mainLoopQuit not set, attempting to close")

		close(c.mainLoopQuit)
		c.mainLoopQuit = nil
	}
}

func (c *Client) stopDeleteLoop() {
	if c.deleteLoopQuit != nil {
		close(c.deleteLoopQuit)
		c.deleteLoopQuit = nil
	}
}

func (c *Client) Stop() error {
	c.stopMainLoop()
	c.stopDeleteLoop()
	return nil
}
