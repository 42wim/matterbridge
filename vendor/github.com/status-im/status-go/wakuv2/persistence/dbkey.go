package persistence

import (
	"encoding/binary"
	"errors"

	"github.com/waku-org/go-waku/waku/v2/hash"
)

const (
	TimestampLength   = 8
	HashLength        = 32
	DigestLength      = HashLength
	PubsubTopicLength = HashLength
	DBKeyLength       = TimestampLength + PubsubTopicLength + DigestLength
)

type Hash [HashLength]byte

var (
	// ErrInvalidByteSize is returned when DBKey can't be created
	// from a byte slice because it has invalid length.
	ErrInvalidByteSize = errors.New("byte slice has invalid length")
)

// DBKey key to be stored in a db.
type DBKey struct {
	raw []byte
}

// Bytes returns a bytes representation of the DBKey.
func (k *DBKey) Bytes() []byte {
	return k.raw
}

// NewDBKey creates a new DBKey with the given values.
func NewDBKey(senderTimestamp uint64, receiverTimestamp uint64, pubsubTopic string, digest []byte) *DBKey {
	pubSubHash := make([]byte, PubsubTopicLength)
	if pubsubTopic != "" {
		pubSubHash = hash.SHA256([]byte(pubsubTopic))
	}

	var k DBKey
	k.raw = make([]byte, DBKeyLength)

	if senderTimestamp == 0 {
		binary.BigEndian.PutUint64(k.raw, receiverTimestamp)
	} else {
		binary.BigEndian.PutUint64(k.raw, senderTimestamp)
	}

	copy(k.raw[TimestampLength:], pubSubHash[:])
	copy(k.raw[TimestampLength+PubsubTopicLength:], digest)

	return &k
}
