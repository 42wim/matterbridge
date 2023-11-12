package doubleratchet

import (
	"encoding/binary"
	"fmt"
)

// MessageHE contains ciphertext and an encrypted header.
type MessageHE struct {
	Header     []byte `json:"header"`
	Ciphertext []byte `json:"ciphertext"`
}

// Message is a single message exchanged by the parties.
type Message struct {
	Header     MessageHeader `json:"header"`
	Ciphertext []byte        `json:"ciphertext"`
}

// MessageHeader that is prepended to every message.
type MessageHeader struct {
	// DHr is the sender's current ratchet public key.
	DH Key `json:"dh"`

	// N is the number of the message in the sending chain.
	N uint32 `json:"n"`

	// PN is the length of the previous sending chain.
	PN uint32 `json:"pn"`
}

// Encode the header in the binary format.
func (mh MessageHeader) Encode() MessageEncHeader {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], mh.N)
	binary.LittleEndian.PutUint32(buf[4:8], mh.PN)
	return append(buf, mh.DH[:]...)
}

// MessageEncHeader is a binary-encoded representation of a message header.
type MessageEncHeader []byte

// Decode message header out of the binary-encoded representation.
func (mh MessageEncHeader) Decode() (MessageHeader, error) {
	// n (4 bytes) + pn (4 bytes) + dh (32 bytes)
	if len(mh) != 40 {
		return MessageHeader{}, fmt.Errorf("encoded message header must be 40 bytes, %d given", len(mh))
	}
	var dh Key = make(Key, 32)
	copy(dh[:], mh[8:40])
	return MessageHeader{
		DH: dh,
		N:  binary.LittleEndian.Uint32(mh[0:4]),
		PN: binary.LittleEndian.Uint32(mh[4:8]),
	}, nil
}
