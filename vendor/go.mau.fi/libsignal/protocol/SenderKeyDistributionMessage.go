package protocol

import (
	"fmt"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/signalerror"
)

// SenderKeyDistributionMessageSerializer is an interface for serializing and deserializing
// SenderKeyDistributionMessages into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type SenderKeyDistributionMessageSerializer interface {
	Serialize(signalMessage *SenderKeyDistributionMessageStructure) []byte
	Deserialize(serialized []byte) (*SenderKeyDistributionMessageStructure, error)
}

// NewSenderKeyDistributionMessageFromBytes will return a Signal Ciphertext message from the given
// bytes using the given serializer.
func NewSenderKeyDistributionMessageFromBytes(serialized []byte,
	serializer SenderKeyDistributionMessageSerializer) (*SenderKeyDistributionMessage, error) {

	// Use the given serializer to decode the signal message.
	signalMessageStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewSenderKeyDistributionMessageFromStruct(signalMessageStructure, serializer)
}

// NewSenderKeyDistributionMessageFromStruct returns a Signal Ciphertext message from the
// given serializable structure.
func NewSenderKeyDistributionMessageFromStruct(structure *SenderKeyDistributionMessageStructure,
	serializer SenderKeyDistributionMessageSerializer) (*SenderKeyDistributionMessage, error) {

	// Throw an error if the given message structure is an unsupported version.
	if structure.Version <= UnsupportedVersion {
		return nil, fmt.Errorf("%w %d (sender key distribution)", signalerror.ErrOldMessageVersion, structure.Version)
	}

	// Throw an error if the given message structure is a future version.
	if structure.Version > CurrentVersion {
		return nil, fmt.Errorf("%w %d (sender key distribution)", signalerror.ErrUnknownMessageVersion, structure.Version)
	}

	// Throw an error if the structure is missing critical fields.
	if structure.SigningKey == nil || structure.ChainKey == nil {
		return nil, fmt.Errorf("%w (sender key distribution)", signalerror.ErrIncompleteMessage)
	}

	// Get the signing key object from bytes.
	signingKey, err := ecc.DecodePoint(structure.SigningKey, 0)
	if err != nil {
		return nil, err
	}

	// Create the signal message object from the structure.
	message := &SenderKeyDistributionMessage{
		id:           structure.ID,
		iteration:    structure.Iteration,
		chainKey:     structure.ChainKey,
		version:      structure.Version,
		signatureKey: signingKey,
		serializer:   serializer,
	}

	// Generate the ECC key from bytes.
	message.signatureKey, err = ecc.DecodePoint(structure.SigningKey, 0)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// NewSenderKeyDistributionMessage returns a Signal Ciphertext message.
func NewSenderKeyDistributionMessage(id uint32, iteration uint32,
	chainKey []byte, signatureKey ecc.ECPublicKeyable,
	serializer SenderKeyDistributionMessageSerializer) *SenderKeyDistributionMessage {

	return &SenderKeyDistributionMessage{
		id:           id,
		iteration:    iteration,
		chainKey:     chainKey,
		signatureKey: signatureKey,
		serializer:   serializer,
	}
}

// SenderKeyDistributionMessageStructure is a serializeable structure for senderkey
// distribution messages.
type SenderKeyDistributionMessageStructure struct {
	ID         uint32
	Iteration  uint32
	ChainKey   []byte
	SigningKey []byte
	Version    uint32
}

// SenderKeyDistributionMessage is a structure for senderkey distribution messages.
type SenderKeyDistributionMessage struct {
	id           uint32
	iteration    uint32
	chainKey     []byte
	version      uint32
	signatureKey ecc.ECPublicKeyable
	serializer   SenderKeyDistributionMessageSerializer
}

// ID will return the message's id.
func (p *SenderKeyDistributionMessage) ID() uint32 {
	return p.id
}

// Iteration will return the message's iteration.
func (p *SenderKeyDistributionMessage) Iteration() uint32 {
	return p.iteration
}

// ChainKey will return the message's chain key in bytes.
func (p *SenderKeyDistributionMessage) ChainKey() []byte {
	return p.chainKey
}

// SignatureKey will return the message's signature public key
func (p *SenderKeyDistributionMessage) SignatureKey() ecc.ECPublicKeyable {
	return p.signatureKey
}

// Serialize will use the given serializer and return the message as
// bytes.
func (p *SenderKeyDistributionMessage) Serialize() []byte {
	structure := &SenderKeyDistributionMessageStructure{
		ID:         p.id,
		Iteration:  p.iteration,
		ChainKey:   p.chainKey,
		SigningKey: p.signatureKey.Serialize(),
		Version:    CurrentVersion,
	}
	return p.serializer.Serialize(structure)
}

// Type will return the message's type.
func (p *SenderKeyDistributionMessage) Type() uint32 {
	return SENDERKEY_DISTRIBUTION_TYPE
}
