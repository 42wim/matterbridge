package transport

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/types"
)

// EnvelopeState in local tracker
type EnvelopeState int

const (
	// NotRegistered returned if asked hash wasn't registered in the tracker.
	NotRegistered EnvelopeState = -1
	// EnvelopePosted is set when envelope was added to a local waku queue.
	EnvelopePosted EnvelopeState = iota + 1
	// EnvelopeSent is set when envelope is sent to at least one peer.
	EnvelopeSent
)

type EnvelopesMonitorConfig struct {
	EnvelopeEventsHandler            EnvelopeEventsHandler
	MaxAttempts                      int
	AwaitOnlyMailServerConfirmations bool
	IsMailserver                     func(types.EnodeID) bool
	Logger                           *zap.Logger
}

// EnvelopeEventsHandler used for two different event types.
type EnvelopeEventsHandler interface {
	EnvelopeSent([][]byte)
	EnvelopeExpired([][]byte, error)
	MailServerRequestCompleted(types.Hash, types.Hash, []byte, error)
	MailServerRequestExpired(types.Hash)
}

// NewEnvelopesMonitor returns a pointer to an instance of the EnvelopesMonitor.
func NewEnvelopesMonitor(w types.Waku, config EnvelopesMonitorConfig) *EnvelopesMonitor {
	logger := config.Logger

	if logger == nil {
		logger = zap.NewNop()
	}

	var api types.PublicWakuAPI
	if w != nil {
		api = w.PublicWakuAPI()
	}

	return &EnvelopesMonitor{
		w:                                w,
		api:                              api,
		handler:                          config.EnvelopeEventsHandler,
		awaitOnlyMailServerConfirmations: config.AwaitOnlyMailServerConfirmations,
		maxAttempts:                      config.MaxAttempts,
		isMailserver:                     config.IsMailserver,
		logger:                           logger.With(zap.Namespace("EnvelopesMonitor")),

		// key is envelope hash (event.Hash)
		envelopes: map[types.Hash]*monitoredEnvelope{},

		// key is hash of the batch (event.Batch)
		batches: map[types.Hash]map[types.Hash]struct{}{},

		// key is stringified message identifier
		identifierHashes: make(map[string][]types.Hash),
	}
}

type monitoredEnvelope struct {
	state       EnvelopeState
	attempts    int
	message     *types.NewMessage
	identifiers [][]byte
}

// EnvelopesMonitor is responsible for monitoring waku envelopes state.
type EnvelopesMonitor struct {
	w           types.Waku
	api         types.PublicWakuAPI
	handler     EnvelopeEventsHandler
	maxAttempts int

	mu sync.Mutex

	envelopes        map[types.Hash]*monitoredEnvelope
	batches          map[types.Hash]map[types.Hash]struct{}
	identifierHashes map[string][]types.Hash

	awaitOnlyMailServerConfirmations bool

	wg           sync.WaitGroup
	quit         chan struct{}
	isMailserver func(peer types.EnodeID) bool

	logger *zap.Logger
}

// Start processing events.
func (m *EnvelopesMonitor) Start() {
	m.quit = make(chan struct{})
	m.wg.Add(1)
	go func() {
		m.handleEnvelopeEvents()
		m.wg.Done()
	}()
}

// Stop process events.
func (m *EnvelopesMonitor) Stop() {
	close(m.quit)
	m.wg.Wait()
}

// Add hashes to a tracker.
// Identifiers may be backed by multiple envelopes. It happens when message is split in segmentation layer.
func (m *EnvelopesMonitor) Add(identifiers [][]byte, envelopeHashes []types.Hash, messages []*types.NewMessage) error {
	if len(envelopeHashes) != len(messages) {
		return errors.New("hashes don't match messages")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, identifier := range identifiers {
		m.identifierHashes[string(identifier)] = envelopeHashes
	}

	for i, envelopeHash := range envelopeHashes {
		if _, ok := m.envelopes[envelopeHash]; !ok {
			m.envelopes[envelopeHash] = &monitoredEnvelope{
				state:       EnvelopePosted,
				attempts:    1,
				message:     messages[i],
				identifiers: identifiers,
			}
		}
	}

	m.processIdentifiers(identifiers)

	return nil
}

func (m *EnvelopesMonitor) GetState(hash types.Hash) EnvelopeState {
	m.mu.Lock()
	defer m.mu.Unlock()
	envelope, exist := m.envelopes[hash]
	if !exist {
		return NotRegistered
	}
	return envelope.state
}

// handleEnvelopeEvents processes waku envelope events
func (m *EnvelopesMonitor) handleEnvelopeEvents() {
	events := make(chan types.EnvelopeEvent, 100) // must be buffered to prevent blocking waku
	sub := m.w.SubscribeEnvelopeEvents(events)
	defer func() {
		close(events)
		sub.Unsubscribe()
	}()
	for {
		select {
		case <-m.quit:
			return
		case event := <-events:
			m.handleEvent(event)
		}
	}
}

// handleEvent based on type of the event either triggers
// confirmation handler or removes hash from tracker
func (m *EnvelopesMonitor) handleEvent(event types.EnvelopeEvent) {
	handlers := map[types.EventType]func(types.EnvelopeEvent){
		types.EventEnvelopeSent:      m.handleEventEnvelopeSent,
		types.EventEnvelopeExpired:   m.handleEventEnvelopeExpired,
		types.EventBatchAcknowledged: m.handleAcknowledgedBatch,
		types.EventEnvelopeReceived:  m.handleEventEnvelopeReceived,
	}
	if handler, ok := handlers[event.Event]; ok {
		handler(event)
	}
}

func (m *EnvelopesMonitor) handleEventEnvelopeSent(event types.EnvelopeEvent) {
	// Mailserver confirmations for WakuV2 are disabled
	if (m.w == nil || m.w.Version() < 2) && m.awaitOnlyMailServerConfirmations {
		if !m.isMailserver(event.Peer) {
			return
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	confirmationExpected := event.Batch != (types.Hash{})

	envelope, ok := m.envelopes[event.Hash]

	// If confirmations are not expected, we keep track of the envelope
	// being sent
	if !ok && !confirmationExpected {
		m.envelopes[event.Hash] = &monitoredEnvelope{state: EnvelopeSent}
		return
	}

	// if message was already confirmed - skip it
	if envelope.state == EnvelopeSent {
		return
	}
	m.logger.Debug("envelope is sent", zap.String("hash", event.Hash.String()), zap.String("peer", event.Peer.String()))
	if confirmationExpected {
		if _, ok := m.batches[event.Batch]; !ok {
			m.batches[event.Batch] = map[types.Hash]struct{}{}
		}
		m.batches[event.Batch][event.Hash] = struct{}{}
		m.logger.Debug("waiting for a confirmation", zap.String("batch", event.Batch.String()))
	} else {
		m.logger.Debug("confirmation not expected, marking as sent")
		envelope.state = EnvelopeSent
		m.processIdentifiers(envelope.identifiers)
	}
}

func (m *EnvelopesMonitor) handleAcknowledgedBatch(event types.EnvelopeEvent) {

	if m.awaitOnlyMailServerConfirmations && !m.isMailserver(event.Peer) {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	envelopes, ok := m.batches[event.Batch]
	if !ok {
		m.logger.Debug("batch is not found", zap.String("batch", event.Batch.String()))
	}
	m.logger.Debug("received a confirmation", zap.String("batch", event.Batch.String()), zap.String("peer", event.Peer.String()))
	envelopeErrors, ok := event.Data.([]types.EnvelopeError)
	if event.Data != nil && !ok {
		m.logger.Error("received unexpected data in the the confirmation event", zap.Any("data", event.Data))
	}
	failedEnvelopes := map[types.Hash]struct{}{}
	for i := range envelopeErrors {
		envelopeError := envelopeErrors[i]
		_, exist := m.envelopes[envelopeError.Hash]
		if exist {
			m.logger.Warn("envelope that was posted by us is discarded", zap.String("hash", envelopeError.Hash.String()), zap.String("peer", event.Peer.String()), zap.String("error", envelopeError.Description))
			var err error
			switch envelopeError.Code {
			case types.EnvelopeTimeNotSynced:
				err = errors.New("envelope wasn't delivered due to time sync issues")
			}
			m.handleEnvelopeFailure(envelopeError.Hash, err)
		}
		failedEnvelopes[envelopeError.Hash] = struct{}{}
	}

	for hash := range envelopes {
		if _, exist := failedEnvelopes[hash]; exist {
			continue
		}
		envelope, ok := m.envelopes[hash]
		if !ok || envelope.state == EnvelopeSent {
			continue
		}
		envelope.state = EnvelopeSent
		m.processIdentifiers(envelope.identifiers)
	}
	delete(m.batches, event.Batch)
}

func (m *EnvelopesMonitor) handleEventEnvelopeExpired(event types.EnvelopeEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handleEnvelopeFailure(event.Hash, errors.New("envelope expired due to connectivity issues"))
}

// handleEnvelopeFailure is a common code path for processing envelopes failures. not thread safe, lock
// must be used on a higher level.
func (m *EnvelopesMonitor) handleEnvelopeFailure(hash types.Hash, err error) {
	if envelope, ok := m.envelopes[hash]; ok {
		m.clearMessageState(hash)
		if envelope.state == EnvelopeSent {
			return
		}
		if envelope.attempts < m.maxAttempts {
			m.logger.Debug("retrying to send a message", zap.String("hash", hash.String()), zap.Int("attempt", envelope.attempts+1))
			hex, err := m.api.Post(context.TODO(), *envelope.message)
			if err != nil {
				m.logger.Error("failed to retry sending message", zap.String("hash", hash.String()), zap.Int("attempt", envelope.attempts+1), zap.Error(err))
				if m.handler != nil {
					m.handler.EnvelopeExpired(envelope.identifiers, err)
				}

			}
			envelopeID := types.BytesToHash(hex)
			m.envelopes[envelopeID] = &monitoredEnvelope{
				state:       EnvelopePosted,
				attempts:    envelope.attempts + 1,
				message:     envelope.message,
				identifiers: envelope.identifiers,
			}
		} else {
			m.logger.Debug("envelope expired", zap.String("hash", hash.String()))
			if m.handler != nil {
				m.handler.EnvelopeExpired(envelope.identifiers, err)
			}
		}
	}
}

func (m *EnvelopesMonitor) handleEventEnvelopeReceived(event types.EnvelopeEvent) {
	if m.awaitOnlyMailServerConfirmations && !m.isMailserver(event.Peer) {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	envelope, ok := m.envelopes[event.Hash]
	if !ok || envelope.state != EnvelopePosted {
		return
	}
	m.logger.Debug("expected envelope received", zap.String("hash", event.Hash.String()), zap.String("peer", event.Peer.String()))
	envelope.state = EnvelopeSent
	m.processIdentifiers(envelope.identifiers)
}

func (m *EnvelopesMonitor) processIdentifiers(identifiers [][]byte) {
	sentIdentifiers := make([][]byte, 0, len(identifiers))

	for _, identifier := range identifiers {
		hashes, ok := m.identifierHashes[string(identifier)]
		if !ok {
			continue
		}

		sent := true
		// Consider identifier as sent if all corresponding envelopes are in EnvelopeSent state
		for _, hash := range hashes {
			envelope, ok := m.envelopes[hash]
			if !ok || envelope.state != EnvelopeSent {
				sent = false
				break
			}
		}
		if sent {
			sentIdentifiers = append(sentIdentifiers, identifier)
		}
	}

	if len(sentIdentifiers) > 0 && m.handler != nil {
		m.handler.EnvelopeSent(sentIdentifiers)
	}
}

// clearMessageState removes all message and envelope state.
// not thread-safe, should be protected on a higher level.
func (m *EnvelopesMonitor) clearMessageState(envelopeID types.Hash) {
	envelope, ok := m.envelopes[envelopeID]
	if !ok {
		return
	}
	delete(m.envelopes, envelopeID)
	for _, identifier := range envelope.identifiers {
		delete(m.identifierHashes, string(identifier))
	}
}
