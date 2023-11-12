package types

// Envelope represents a clear-text data packet to transmit through the Whisper
// network. Its contents may or may not be encrypted and signed.
type Envelope interface {
	Wrapped

	Hash() Hash // cached hash of the envelope to avoid rehashing every time
	Bloom() []byte
	PoW() float64
	Expiry() uint32
	TTL() uint32
	Topic() TopicType
	Size() int
}

// EventType used to define known envelope events.
type EventType string

// NOTE: This list of event names is extracted from Geth. It must be kept in sync, or otherwise a mapping layer needs to be created
const (
	// EventEnvelopeSent fires when envelope was sent to a peer.
	EventEnvelopeSent EventType = "envelope.sent"
	// EventEnvelopeExpired fires when envelop expired
	EventEnvelopeExpired EventType = "envelope.expired"
	// EventEnvelopeReceived is sent once envelope was received from a peer.
	// EventEnvelopeReceived must be sent to the feed even if envelope was previously in the cache.
	// And event, ideally, should contain information about peer that sent envelope to us.
	EventEnvelopeReceived EventType = "envelope.received"
	// EventBatchAcknowledged is sent when batch of envelopes was acknowledged by a peer.
	EventBatchAcknowledged EventType = "batch.acknowledged"
	// EventEnvelopeAvailable fires when envelop is available for filters
	EventEnvelopeAvailable EventType = "envelope.available"
	// EventMailServerRequestSent fires when such request is sent.
	EventMailServerRequestSent EventType = "mailserver.request.sent"
	// EventMailServerRequestCompleted fires after mailserver sends all the requested messages
	EventMailServerRequestCompleted EventType = "mailserver.request.completed"
	// EventMailServerRequestExpired fires after mailserver the request TTL ends.
	// This event is independent and concurrent to EventMailServerRequestCompleted.
	// Request should be considered as expired only if expiry event was received first.
	EventMailServerRequestExpired EventType = "mailserver.request.expired"
	// EventMailServerEnvelopeArchived fires after an envelope has been archived
	EventMailServerEnvelopeArchived EventType = "mailserver.envelope.archived"
	// EventMailServerSyncFinished fires when the sync of messages is finished.
	EventMailServerSyncFinished EventType = "mailserver.sync.finished"
)

const (
	// EnvelopeTimeNotSynced represents the code passed to notify of a clock skew situation
	EnvelopeTimeNotSynced uint = 1000
	// EnvelopeOtherError represents the code passed to notify of a generic error situation
	EnvelopeOtherError
)

// EnvelopeEvent used for envelopes events.
type EnvelopeEvent struct {
	Event EventType
	Hash  Hash
	Batch Hash
	Peer  EnodeID
	Data  interface{}
}

// EnvelopeError code and optional description of the error.
type EnvelopeError struct {
	Hash        Hash
	Code        uint
	Description string
}

// Subscription represents a stream of events. The carrier of the events is typically a
// channel, but isn't part of the interface.
//
// Subscriptions can fail while established. Failures are reported through an error
// channel. It receives a value if there is an issue with the subscription (e.g. the
// network connection delivering the events has been closed). Only one value will ever be
// sent.
//
// The error channel is closed when the subscription ends successfully (i.e. when the
// source of events is closed). It is also closed when Unsubscribe is called.
//
// The Unsubscribe method cancels the sending of events. You must call Unsubscribe in all
// cases to ensure that resources related to the subscription are released. It can be
// called any number of times.
type Subscription interface {
	Err() <-chan error // returns the error channel
	Unsubscribe()      // cancels sending of events, closing the error channel
}
