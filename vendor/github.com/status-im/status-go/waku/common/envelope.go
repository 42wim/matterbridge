// Copyright 2019 The Waku Library Authors.
//
// The Waku library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Waku library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty off
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Waku library. If not, see <http://www.gnu.org/licenses/>.
//
// This software uses the go-ethereum library, which is licensed
// under the GNU Lesser General Public Library, version 3 or any later.

package common

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/rlp"
)

// Envelope represents a clear-text data packet to transmit through the Waku
// network. Its contents may or may not be encrypted and signed.
type Envelope struct {
	Expiry uint32
	TTL    uint32
	Topic  TopicType
	Data   []byte
	Nonce  uint64

	pow float64 // Message-specific PoW as described in the Waku specification.

	// the following variables should not be accessed directly, use the corresponding function instead: Hash(), Bloom()
	hash  common.Hash // Cached hash of the envelope to avoid rehashing every time.
	bloom []byte
}

// Size returns the size of envelope as it is sent (i.e. public fields only)
func (e *Envelope) Size() int {
	return EnvelopeHeaderLength + len(e.Data)
}

// rlpWithoutNonce returns the RLP encoded envelope contents, except the nonce.
func (e *Envelope) rlpWithoutNonce() []byte {
	res, _ := rlp.EncodeToBytes([]interface{}{e.Expiry, e.TTL, e.Topic, e.Data})
	return res
}

// NewEnvelope wraps a Waku message with expiration and destination data
// included into an envelope for network forwarding.
func NewEnvelope(ttl uint32, topic TopicType, msg *SentMessage, now time.Time) *Envelope {
	env := Envelope{
		Expiry: uint32(now.Add(time.Second * time.Duration(ttl)).Unix()),
		TTL:    ttl,
		Topic:  topic,
		Data:   msg.Raw,
		Nonce:  0,
	}

	return &env
}

// Seal closes the envelope by spending the requested amount of time as a proof
// of work on hashing the data.
func (e *Envelope) Seal(options *MessageParams) error {
	if options.PoW == 0 {
		// PoW is not required
		return nil
	}

	var target, bestLeadingZeros int
	if options.PoW < 0 {
		// target is not set - the function should run for a period
		// of time specified in WorkTime param. Since we can predict
		// the execution time, we can also adjust Expiry.
		e.Expiry += options.WorkTime
	} else {
		target = e.powToFirstBit(options.PoW)
	}

	rwn := e.rlpWithoutNonce()
	buf := make([]byte, len(rwn)+8)
	copy(buf, rwn)
	asAnInt := new(big.Int)

	finish := time.Now().Add(time.Duration(options.WorkTime) * time.Second).UnixNano()
	for nonce := uint64(0); time.Now().UnixNano() < finish; {
		for i := 0; i < 1024; i++ {
			binary.BigEndian.PutUint64(buf[len(rwn):], nonce)
			h := crypto.Keccak256(buf)
			asAnInt.SetBytes(h)
			leadingZeros := 256 - asAnInt.BitLen()
			if leadingZeros > bestLeadingZeros {
				e.Nonce, bestLeadingZeros = nonce, leadingZeros
				if target > 0 && bestLeadingZeros >= target {
					return nil
				}
			}
			nonce++
		}
	}

	if target > 0 && bestLeadingZeros < target {
		return fmt.Errorf("failed to reach the PoW target, specified pow time (%d seconds) was insufficient", options.WorkTime)
	}

	return nil
}

// PoW computes (if necessary) and returns the proof of work target
// of the envelope.
func (e *Envelope) PoW() float64 {
	if e.pow == 0 {
		e.CalculatePoW(0)
	}
	return e.pow
}

func (e *Envelope) CalculatePoW(diff uint32) {
	rwn := e.rlpWithoutNonce()
	buf := make([]byte, len(rwn)+8)
	copy(buf, rwn)
	binary.BigEndian.PutUint64(buf[len(rwn):], e.Nonce)
	powHash := new(big.Int).SetBytes(crypto.Keccak256(buf))
	leadingZeroes := 256 - powHash.BitLen()
	x := math.Pow(2, float64(leadingZeroes))
	x /= float64(len(rwn))
	x /= float64(e.TTL + diff)
	e.pow = x
}

func (e *Envelope) powToFirstBit(pow float64) int {
	x := pow
	x *= float64(e.Size())
	x *= float64(e.TTL)
	bits := math.Log2(x)
	bits = math.Ceil(bits)
	res := int(bits)
	if res < 1 {
		res = 1
	}
	return res
}

// Hash returns the SHA3 hash of the envelope, calculating it if not yet done.
func (e *Envelope) Hash() common.Hash {
	if (e.hash == common.Hash{}) {
		encoded, _ := rlp.EncodeToBytes(e)
		e.hash = crypto.Keccak256Hash(encoded)
	}
	return e.hash
}

// DecodeRLP decodes an Envelope from an RLP data stream.
func (e *Envelope) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}
	// The decoding of Envelope uses the struct fields but also needs
	// to compute the hash of the whole RLP-encoded envelope. This
	// type has the same structure as Envelope but is not an
	// rlp.Decoder (does not implement DecodeRLP function).
	// Only public members will be encoded.
	type rlpenv Envelope
	if err := rlp.DecodeBytes(raw, (*rlpenv)(e)); err != nil {
		return err
	}
	e.hash = crypto.Keccak256Hash(raw)
	return nil
}

// OpenAsymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *Envelope) OpenAsymmetric(key *ecdsa.PrivateKey) (*ReceivedMessage, error) {
	message := &ReceivedMessage{Raw: e.Data}
	err := message.decryptAsymmetric(key)
	switch err {
	case nil:
		return message, nil
	case ecies.ErrInvalidPublicKey: // addressed to somebody else
		return nil, err
	default:
		return nil, fmt.Errorf("unable to open envelope, decrypt failed: %v", err)
	}
}

// OpenSymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *Envelope) OpenSymmetric(key []byte) (msg *ReceivedMessage, err error) {
	msg = &ReceivedMessage{Raw: e.Data}
	err = msg.decryptSymmetric(key)
	if err != nil {
		msg = nil
	}
	return msg, err
}

// Open tries to decrypt an envelope, and populates the message fields in case of success.
func (e *Envelope) Open(watcher *Filter) (msg *ReceivedMessage) {
	if watcher == nil {
		return nil
	}

	// The API interface forbids filters doing both symmetric and asymmetric encryption.
	if watcher.expectsAsymmetricEncryption() && watcher.expectsSymmetricEncryption() {
		return nil
	}

	if watcher.expectsAsymmetricEncryption() {
		msg, _ = e.OpenAsymmetric(watcher.KeyAsym)
		if msg != nil {
			msg.Dst = &watcher.KeyAsym.PublicKey
		}
	} else if watcher.expectsSymmetricEncryption() {
		msg, _ = e.OpenSymmetric(watcher.KeySym)
		if msg != nil {
			msg.SymKeyHash = crypto.Keccak256Hash(watcher.KeySym)
		}
	}

	if msg != nil {
		ok := msg.ValidateAndParse()
		if !ok {
			return nil
		}
		msg.Topic = e.Topic
		msg.PoW = e.PoW()
		msg.TTL = e.TTL
		msg.Sent = e.Expiry - e.TTL
		msg.EnvelopeHash = e.Hash()
	}
	return msg
}

// Bloom maps 4-bytes Topic into 64-byte bloom filter with 3 bits set (at most).
func (e *Envelope) Bloom() []byte {
	if e.bloom == nil {
		e.bloom = e.Topic.ToBloom()
	}
	return e.bloom
}

// EnvelopeError code and optional description of the error.
type EnvelopeError struct {
	Hash        common.Hash
	Code        uint
	Description string
}

// ErrorToEnvelopeError converts common golang error into EnvelopeError with a code.
func ErrorToEnvelopeError(hash common.Hash, err error) EnvelopeError {
	code := EnvelopeOtherError
	switch err.(type) {
	case TimeSyncError:
		code = EnvelopeTimeNotSynced
	}
	return EnvelopeError{
		Hash:        hash,
		Code:        code,
		Description: err.Error(),
	}
}
