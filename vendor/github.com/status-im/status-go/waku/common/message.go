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
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/log"
)

// MessageParams specifies the exact way a message should be wrapped
// into an Envelope.
type MessageParams struct {
	TTL      uint32
	Src      *ecdsa.PrivateKey
	Dst      *ecdsa.PublicKey
	KeySym   []byte
	Topic    TopicType
	WorkTime uint32
	PoW      float64
	Payload  []byte
	Padding  []byte
}

// SentMessage represents an end-user data packet to transmit through the
// Waku protocol. These are wrapped into Envelopes that need not be
// understood by intermediate nodes, just forwarded.
type SentMessage struct {
	Raw []byte
}

// ReceivedMessage represents a data packet to be received through the
// Waku protocol and successfully decrypted.
type ReceivedMessage struct {
	Raw []byte

	Payload   []byte
	Padding   []byte
	Signature []byte
	Salt      []byte

	PoW   float64          // Proof of work as described in the Waku spec
	Sent  uint32           // Time when the message was posted into the network
	TTL   uint32           // Maximum time to live allowed for the message
	Src   *ecdsa.PublicKey // Message recipient (identity used to decode the message)
	Dst   *ecdsa.PublicKey // Message recipient (identity used to decode the message)
	Topic TopicType

	SymKeyHash   common.Hash // The Keccak256Hash of the key
	EnvelopeHash common.Hash // Message envelope hash to act as a unique id

	P2P bool // is set to true if this message was received from mail server.
}

// MessagesRequest contains details of a request for historic messages.
type MessagesRequest struct {
	// ID of the request. The current implementation requires ID to be 32-byte array,
	// however, it's not enforced for future implementation.
	ID []byte `json:"id"`

	// From is a lower bound of time range.
	From uint32 `json:"from"`

	// To is a upper bound of time range.
	To uint32 `json:"to"`

	// Limit determines the number of messages sent by the mail server
	// for the current paginated request.
	Limit uint32 `json:"limit"`

	// Cursor is used as starting point for paginated requests.
	Cursor []byte `json:"cursor"`

	// Bloom is a filter to match requested messages.
	Bloom []byte `json:"bloom"`

	// Topics is a list of topics. A returned message should
	// belong to one of the topics from the list.
	Topics [][]byte `json:"topics"`
}

func (r MessagesRequest) Validate() error {
	if len(r.ID) != common.HashLength {
		return errors.New("invalid 'ID', expected a 32-byte slice")
	}

	if r.From > r.To {
		return errors.New("invalid 'From' value which is greater than To")
	}

	if r.Limit > MaxLimitInMessagesRequest {
		return fmt.Errorf("invalid 'Limit' value, expected value lower than %d", MaxLimitInMessagesRequest)
	}

	if len(r.Bloom) == 0 && len(r.Topics) == 0 {
		return errors.New("invalid 'Bloom' or 'Topics', one must be non-empty")
	}

	return nil
}

// MessagesResponse sent as a response after processing batch of envelopes.
type MessagesResponse struct {
	// Hash is a hash of all envelopes sent in the single batch.
	Hash common.Hash
	// Per envelope error.
	Errors []EnvelopeError
}

func IsMessageSigned(flags byte) bool {
	return (flags & signatureFlag) != 0
}

func (msg *ReceivedMessage) isSymmetricEncryption() bool {
	return msg.SymKeyHash != common.Hash{}
}

func (msg *ReceivedMessage) isAsymmetricEncryption() bool {
	return msg.Dst != nil
}

// NewSentMessage creates and initializes a non-signed, non-encrypted Waku message.
func NewSentMessage(params *MessageParams) (*SentMessage, error) {
	const payloadSizeFieldMaxSize = 4
	msg := SentMessage{}
	msg.Raw = make([]byte, 1,
		flagsLength+payloadSizeFieldMaxSize+len(params.Payload)+len(params.Padding)+signatureLength+padSizeLimit)
	msg.Raw[0] = 0 // set all the flags to zero
	msg.addPayloadSizeField(params.Payload)
	msg.Raw = append(msg.Raw, params.Payload...)
	err := msg.appendPadding(params)
	return &msg, err
}

// addPayloadSizeField appends the auxiliary field containing the size of payload
func (msg *SentMessage) addPayloadSizeField(payload []byte) {
	fieldSize := getSizeOfPayloadSizeField(payload)
	field := make([]byte, 4)
	binary.LittleEndian.PutUint32(field, uint32(len(payload)))
	field = field[:fieldSize]
	msg.Raw = append(msg.Raw, field...)
	msg.Raw[0] |= byte(fieldSize)
}

// getSizeOfPayloadSizeField returns the number of bytes necessary to encode the size of payload
func getSizeOfPayloadSizeField(payload []byte) int {
	s := 1
	for i := len(payload); i >= 256; i /= 256 {
		s++
	}
	return s
}

// appendPadding appends the padding specified in params.
// If no padding is provided in params, then random padding is generated.
func (msg *SentMessage) appendPadding(params *MessageParams) error {
	if len(params.Padding) != 0 {
		// padding data was provided by the Dapp, just use it as is
		msg.Raw = append(msg.Raw, params.Padding...)
		return nil
	}

	rawSize := flagsLength + getSizeOfPayloadSizeField(params.Payload) + len(params.Payload)
	if params.Src != nil {
		rawSize += signatureLength
	}
	odd := rawSize % padSizeLimit
	paddingSize := padSizeLimit - odd
	pad := make([]byte, paddingSize)
	_, err := crand.Read(pad)
	if err != nil {
		return err
	}
	if !ValidateDataIntegrity(pad, paddingSize) {
		return errors.New("failed to generate random padding of size " + strconv.Itoa(paddingSize))
	}
	msg.Raw = append(msg.Raw, pad...)
	return nil
}

// sign calculates and sets the cryptographic signature for the message,
// also setting the sign flag.
func (msg *SentMessage) sign(key *ecdsa.PrivateKey) error {
	if IsMessageSigned(msg.Raw[0]) {
		// this should not happen, but no reason to panic
		log.Error("failed to sign the message: already signed")
		return nil
	}

	msg.Raw[0] |= signatureFlag // it is important to set this flag before signing
	hash := crypto.Keccak256(msg.Raw)
	signature, err := crypto.Sign(hash, key)
	if err != nil {
		msg.Raw[0] &= 0xFF ^ signatureFlag // clear the flag
		return err
	}
	msg.Raw = append(msg.Raw, signature...)
	return nil
}

// encryptAsymmetric encrypts a message with a public key.
func (msg *SentMessage) encryptAsymmetric(key *ecdsa.PublicKey) error {
	if !ValidatePublicKey(key) {
		return errors.New("invalid public key provided for asymmetric encryption")
	}
	encrypted, err := ecies.Encrypt(crand.Reader, ecies.ImportECDSAPublic(key), msg.Raw, nil, nil)
	if err == nil {
		msg.Raw = encrypted
	}
	return err
}

// encryptSymmetric encrypts a message with a topic key, using AES-GCM-256.
// nonce size should be 12 bytes (see cipher.gcmStandardNonceSize).
func (msg *SentMessage) encryptSymmetric(key []byte) (err error) {
	if !ValidateDataIntegrity(key, AESKeyLength) {
		return errors.New("invalid key provided for symmetric encryption, size: " + strconv.Itoa(len(key)))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	salt, err := GenerateSecureRandomData(aesNonceLength) // never use more than 2^32 random nonces with a given key
	if err != nil {
		return err
	}
	encrypted := aesgcm.Seal(nil, salt, msg.Raw, nil)
	msg.Raw = append(encrypted, salt...)
	return nil
}

// Wrap bundles the message into an Envelope to transmit over the network.
func (msg *SentMessage) Wrap(options *MessageParams, now time.Time) (envelope *Envelope, err error) {
	if options.TTL == 0 {
		options.TTL = DefaultTTL
	}
	if options.Src != nil {
		if err = msg.sign(options.Src); err != nil {
			return nil, err
		}
	}
	if options.Dst != nil {
		err = msg.encryptAsymmetric(options.Dst)
	} else if options.KeySym != nil {
		err = msg.encryptSymmetric(options.KeySym)
	} else {
		err = errors.New("unable to encrypt the message: neither symmetric nor assymmetric key provided")
	}
	if err != nil {
		return nil, err
	}

	envelope = NewEnvelope(options.TTL, options.Topic, msg, now)
	if err = envelope.Seal(options); err != nil {
		return nil, err
	}
	return envelope, nil
}

// decryptSymmetric decrypts a message with a topic key, using AES-GCM-256.
// nonce size should be 12 bytes (see cipher.gcmStandardNonceSize).
func (msg *ReceivedMessage) decryptSymmetric(key []byte) error {
	// symmetric messages are expected to contain the 12-byte nonce at the end of the payload
	if len(msg.Raw) < aesNonceLength {
		return errors.New("missing salt or invalid payload in symmetric message")
	}
	salt := msg.Raw[len(msg.Raw)-aesNonceLength:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	decrypted, err := aesgcm.Open(nil, salt, msg.Raw[:len(msg.Raw)-aesNonceLength], nil)
	if err != nil {
		return err
	}
	msg.Raw = decrypted
	msg.Salt = salt
	return nil
}

// decryptAsymmetric decrypts an encrypted payload with a private key.
func (msg *ReceivedMessage) decryptAsymmetric(key *ecdsa.PrivateKey) error {
	decrypted, err := ecies.ImportECDSA(key).Decrypt(msg.Raw, nil, nil)
	if err == nil {
		msg.Raw = decrypted
	}
	return err
}

// ValidateAndParse checks the message validity and extracts the fields in case of success.
func (msg *ReceivedMessage) ValidateAndParse() bool {
	end := len(msg.Raw)
	if end < 1 {
		return false
	}

	if IsMessageSigned(msg.Raw[0]) {
		end -= signatureLength
		if end <= 1 {
			return false
		}
		msg.Signature = msg.Raw[end : end+signatureLength]
		msg.Src = msg.SigToPubKey()
		if msg.Src == nil {
			return false
		}
	}

	beg := 1
	payloadSize := 0
	sizeOfPayloadSizeField := int(msg.Raw[0] & SizeMask) // number of bytes indicating the size of payload
	if sizeOfPayloadSizeField != 0 {
		if end < beg+sizeOfPayloadSizeField {
			return false
		}
		payloadSize = int(BytesToUintLittleEndian(msg.Raw[beg : beg+sizeOfPayloadSizeField]))
		beg += sizeOfPayloadSizeField
		if beg+payloadSize > end {
			return false
		}
		msg.Payload = msg.Raw[beg : beg+payloadSize]
	}

	beg += payloadSize
	msg.Padding = msg.Raw[beg:end]
	return true
}

// SigToPubKey returns the public key associated to the message's
// signature.
func (msg *ReceivedMessage) SigToPubKey() *ecdsa.PublicKey {
	// in case of invalid signature
	defer func() { recover() }() // nolint: errcheck

	pub, err := crypto.SigToPub(msg.hash(), msg.Signature)
	if err != nil {
		log.Error("failed to recover public key from signature", "err", err)
		return nil
	}
	return pub
}

// hash calculates the SHA3 checksum of the message flags, payload size field, payload and padding.
func (msg *ReceivedMessage) hash() []byte {
	if IsMessageSigned(msg.Raw[0]) {
		sz := len(msg.Raw) - signatureLength
		return crypto.Keccak256(msg.Raw[:sz])
	}
	return crypto.Keccak256(msg.Raw)
}

// MessageStore defines interface for temporary message store.
type MessageStore interface {
	Add(*ReceivedMessage) error
	Pop() ([]*ReceivedMessage, error)
}

// NewMemoryMessageStore returns pointer to an instance of the MemoryMessageStore.
func NewMemoryMessageStore() *MemoryMessageStore {
	return &MemoryMessageStore{
		messages: map[common.Hash]*ReceivedMessage{},
	}
}

// MemoryMessageStore represents messages stored in a memory hash table.
type MemoryMessageStore struct {
	mu       sync.Mutex
	messages map[common.Hash]*ReceivedMessage
}

// Add adds message to store.
func (store *MemoryMessageStore) Add(msg *ReceivedMessage) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exist := store.messages[msg.EnvelopeHash]; !exist {
		store.messages[msg.EnvelopeHash] = msg
	}
	return nil
}

// Pop returns all available messages and cleans the store.
func (store *MemoryMessageStore) Pop() ([]*ReceivedMessage, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	all := make([]*ReceivedMessage, 0, len(store.messages))
	for hash, msg := range store.messages {
		delete(store.messages, hash)
		all = append(all, msg)
	}
	return all, nil
}
