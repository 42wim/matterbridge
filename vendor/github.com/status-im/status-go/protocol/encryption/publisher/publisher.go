package publisher

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/logutils"
)

const (
	// How often a ticker fires in seconds.
	tickerInterval = 120
	// How often we should publish a contact code in seconds.
	publishInterval = 21600
	// Cooldown period on acking messages when not targeting our device.
	deviceNotFoundAckInterval = 7200
)

var (
	errNotEnoughTimePassed = errors.New("not enough time passed")
)

type Publisher struct {
	persistence *persistence
	logger      *zap.Logger
	notifyCh    chan struct{}
	quit        chan struct{}
}

func New(logger *zap.Logger) *Publisher {
	if logger == nil {
		logger = logutils.ZapLogger()
	}

	return &Publisher{
		persistence: newPersistence(),
		logger:      logger.With(zap.Namespace("Publisher")),
	}
}

func (p *Publisher) Start() <-chan struct{} {
	logger := p.logger.With(zap.String("site", "Start"))

	logger.Info("starting publisher")

	p.notifyCh = make(chan struct{}, 100)
	p.quit = make(chan struct{})

	go p.tickerLoop()

	return p.notifyCh
}

func (p *Publisher) Stop() {
	// If hasn't started, ignore
	if p.quit == nil {
		return
	}
	select {
	case _, ok := <-p.quit:
		if !ok {
			// channel already closed
			return
		}
	default:
		close(p.quit)
	}
}

func (p *Publisher) tickerLoop() {
	ticker := time.NewTicker(tickerInterval * time.Second)

	go func() {
		logger := p.logger.With(zap.String("site", "tickerLoop"))

		for {
			select {
			case <-ticker.C:
				err := p.notify()
				switch err {
				case errNotEnoughTimePassed:
					logger.Debug("not enough time passed")
				case nil:
					// skip
				default:
					logger.Error("error while sending a contact code", zap.Error(err))
				}
			case <-p.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Publisher) notify() error {
	lastPublished := p.persistence.getLastPublished()

	now := time.Now().Unix()

	if now-lastPublished < publishInterval {
		return errNotEnoughTimePassed
	}

	select {
	case p.notifyCh <- struct{}{}:
	default:
		p.logger.Warn("publisher channel full, dropping message")
	}

	p.persistence.setLastPublished(now)
	return nil
}

func (p *Publisher) ShouldAdvertiseBundle(publicKey *ecdsa.PublicKey, now int64) (bool, error) {
	identity := crypto.CompressPubkey(publicKey)
	lastAcked := p.persistence.lastAck(identity)
	return now-lastAcked < deviceNotFoundAckInterval, nil
}

func (p *Publisher) SetLastAck(publicKey *ecdsa.PublicKey, now int64) {
	identity := crypto.CompressPubkey(publicKey)
	p.persistence.setLastAck(identity, now)
}
