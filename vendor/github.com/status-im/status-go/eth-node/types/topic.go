package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// TopicLength is the expected length of the topic, in bytes
	TopicLength = 4
	// BloomFilterSize is the expected length of a bloom filter byte array, in bytes
	BloomFilterSize = 64
)

// TopicType represents a cryptographically secure, probabilistic partial
// classifications of a message, determined as the first (left) 4 bytes of the
// SHA3 hash of some arbitrary data given by the original author of the message.
type TopicType [TopicLength]byte

// BytesToTopic converts from the byte array representation of a topic
// into the TopicType type.
func BytesToTopic(b []byte) (t TopicType) {
	sz := TopicLength
	if x := len(b); x < TopicLength {
		sz = x
	}
	for i := 0; i < sz; i++ {
		t[i] = b[i]
	}
	return t
}

// String converts a topic byte array to a string representation.
func (t *TopicType) String() string {
	return EncodeHex(t[:])
}

// MarshalText returns the hex representation of t.
func (t TopicType) MarshalText() ([]byte, error) {
	return HexBytes(t[:]).MarshalText()
}

// UnmarshalText parses a hex representation to a topic.
func (t *TopicType) UnmarshalText(input []byte) error {
	return UnmarshalFixedText("Topic", input, t[:])
}

// TopicToBloom converts the topic (4 bytes) to the bloom filter (64 bytes)
func TopicToBloom(topic TopicType) []byte {
	b := make([]byte, BloomFilterSize)
	var index [3]int
	for j := 0; j < 3; j++ {
		index[j] = int(topic[j])
		if (topic[3] & (1 << uint(j))) != 0 {
			index[j] += 256
		}
	}

	for j := 0; j < 3; j++ {
		byteIndex := index[j] / 8
		bitIndex := index[j] % 8
		b[byteIndex] = (1 << uint(bitIndex))
	}
	return b
}

// BloomFilterMatch returns true if a sample matches a bloom filter
func BloomFilterMatch(filter, sample []byte) bool {
	if filter == nil {
		return true
	}

	for i := 0; i < BloomFilterSize; i++ {
		f := filter[i]
		s := sample[i]
		if (f | s) != f {
			return false
		}
	}

	return true
}

// MakeFullNodeBloom returns a bloom filter which matches all topics
func MakeFullNodeBloom() []byte {
	bloom := make([]byte, BloomFilterSize)
	for i := 0; i < BloomFilterSize; i++ {
		bloom[i] = 0xFF
	}
	return bloom
}

func StringToTopic(s string) (t TopicType) {
	str, _ := hexutil.Decode(s)
	return BytesToTopic(str)
}

func TopicTypeToByteArray(t TopicType) []byte {
	return t[:4]
}
