package mailserver

import (
	"time"

	"github.com/status-im/status-go/eth-node/types"
)

// every this many seconds check real envelopes count
const envelopeCountCheckInterval = 60

// DB is an interface to abstract interactions with the db so that the mailserver
// is agnostic to the underlying technology used
type DB interface {
	Close() error
	// SaveEnvelope stores an envelope
	SaveEnvelope(types.Envelope) error
	// GetEnvelope returns an rlp encoded envelope from the datastore
	GetEnvelope(*DBKey) ([]byte, error)
	// Prune removes envelopes older than time
	Prune(time.Time, int) (int, error)
	// BuildIterator returns an iterator over envelopes
	BuildIterator(query CursorQuery) (Iterator, error)
}

type Iterator interface {
	Next() bool
	DBKey() (*DBKey, error)
	Release() error
	Error() error
	GetEnvelopeByBloomFilter(bloom []byte) ([]byte, error)
	GetEnvelopeByTopicsMap(topics map[types.TopicType]bool) ([]byte, error)
}

type CursorQuery struct {
	start  []byte
	end    []byte
	cursor []byte
	limit  uint32
	bloom  []byte
	topics [][]byte
}
