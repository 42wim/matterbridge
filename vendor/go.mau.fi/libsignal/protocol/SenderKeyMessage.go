package protocol

import (
	"fmt"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/signalerror"
	"go.mau.fi/libsignal/util/bytehelper"
)

// SenderKeyMessageSerializer is an interface for serializing and deserializing
// SenderKeyMessages into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type SenderKeyMessageSerializer interface {
	Serialize(signalMessage *SenderKeyMessageStructure) []byte
	Deserialize(serialized []byte) (*SenderKeyMessageStructure, error)
}

// NewSenderKeyMessageFromBytes will return a Signal Ciphertext message from the given
// bytes using the given serializer.
func NewSenderKeyMessageFromBytes(serialized []byte,
	serializer SenderKeyMessageSerializer) (*SenderKeyMessage, error) {

	// Use the given serializer to decode the signal message.
	senderKeyMessageStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewSenderKeyMessageFromStruct(senderKeyMessageStructure, serializer)
}

// NewSenderKeyMessageFromStruct returns a Signal Ciphertext message from the
// given serializable structure.
func NewSenderKeyMessageFromStruct(structure *SenderKeyMessageStructure,
	serializer SenderKeyMessageSerializer) (*SenderKeyMessage, error) {

	// Throw an error if the given message structure is an unsupported version.
	if structure.Version <= UnsupportedVersion {
		return nil, fmt.Errorf("%w %d (sender key message)", signalerror.ErrOldMessageVersion, structure.Version)
	}

	// Throw an error if the given message structure is a future version.
	if structure.Version > CurrentVersion {
		return nil, fmt.Errorf("%w %d (sender key message)", signalerror.ErrUnknownMessageVersion, structure.Version)
	}

	// Throw an error if the structure is missing critical fields.
	if structure.CipherText == nil {
		return nil, fmt.Errorf("%w (sender key message)", signalerror.ErrIncompleteMessage)
	}

	// Create the signal message object from the structure.
	whisperMessage := &SenderKeyMessage{
		keyID:      structure.ID,
		version:    structure.Version,
		iteration:  structure.Iteration,
		ciphertext: structure.CipherText,
		signature:  structure.Signature,
		serializer: serializer,
	}

	return whisperMessage, nil
}

// NewSenderKeyMessage returns a SenderKeyMessage.
func NewSenderKeyMessage(keyID uint32, iteration uint32, ciphertext []byte,
	signatureKey ecc.ECPrivateKeyable, serializer SenderKeyMessageSerializer) *SenderKeyMessage {

	// Ensure we have a valid signature key
	if signatureKey == nil {
		panic("Signature is nil. Unable to sign new senderkey message.")
	}

	// Build our SenderKeyMessage.
	senderKeyMessage := &SenderKeyMessage{
		keyID:      keyID,
		iteration:  iteration,
		ciphertext: ciphertext,
		version:    CurrentVersion,
		serializer: serializer,
	}

	// Sign the serialized message and include it in the message. This will be included
	// in the signed serialized version of the message.
	signature := ecc.CalculateSignature(signatureKey, senderKeyMessage.Serialize())
	senderKeyMessage.signature = bytehelper.ArrayToSlice64(signature)

	return senderKeyMessage
}

// SenderKeyMessageStructure is a serializeable structure for SenderKey messages.
type SenderKeyMessageStructure struct {
	ID         uint32
	Iteration  uint32
	CipherText []byte
	Version    uint32
	Signature  []byte
}

// SenderKeyMessage is a structure for messages using senderkey groups.
type SenderKeyMessage struct {
	version    uint32
	keyID      uint32
	iteration  uint32
	ciphertext []byte
	signature  []byte
	serializer SenderKeyMessageSerializer
}

// KeyID returns the SenderKeyMessage key ID.
func (p *SenderKeyMessage) KeyID() uint32 {
	return p.keyID
}

// Iteration returns the SenderKeyMessage iteration.
func (p *SenderKeyMessage) Iteration() uint32 {
	return p.iteration
}

// Ciphertext returns the SenderKeyMessage encrypted ciphertext.
func (p *SenderKeyMessage) Ciphertext() []byte {
	return p.ciphertext
}

// Version returns the Signal message version of the message.
func (p *SenderKeyMessage) Version() uint32 {
	return p.version
}

// Serialize will use the given serializer to return the message as bytes
// excluding the signature. This should be used for signing and verifying
// message signatures.
func (p *SenderKeyMessage) Serialize() []byte {
	structure := &SenderKeyMessageStructure{
		ID:         p.keyID,
		Iteration:  p.iteration,
		CipherText: p.ciphertext,
		Version:    p.version,
	}

	return p.serializer.Serialize(structure)
}

// SignedSerialize will use the given serializer to return the message as
// bytes with the message signature included. This should be used when
// sending the message over the network.
func (p *SenderKeyMessage) SignedSerialize() []byte {
	structure := &SenderKeyMessageStructure{
		ID:         p.keyID,
		Iteration:  p.iteration,
		CipherText: p.ciphertext,
		Version:    p.version,
		Signature:  p.signature,
	}

	return p.serializer.Serialize(structure)
}

// Signature returns the SenderKeyMessage signature
func (p *SenderKeyMessage) Signature() [64]byte {
	return bytehelper.SliceToArray64(p.signature)
}

// Type returns the sender key type.
func (p *SenderKeyMessage) Type() uint32 {
	return SENDERKEY_TYPE
}
