package protocol

import (
	"fmt"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/signalerror"
	"go.mau.fi/libsignal/util/optional"
)

// PreKeySignalMessageSerializer is an interface for serializing and deserializing
// PreKeySignalMessages into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type PreKeySignalMessageSerializer interface {
	Serialize(signalMessage *PreKeySignalMessageStructure) []byte
	Deserialize(serialized []byte) (*PreKeySignalMessageStructure, error)
}

// NewPreKeySignalMessageFromBytes will return a Signal Ciphertext message from the given
// bytes using the given serializer.
func NewPreKeySignalMessageFromBytes(serialized []byte, serializer PreKeySignalMessageSerializer,
	msgSerializer SignalMessageSerializer) (*PreKeySignalMessage, error) {
	// Use the given serializer to decode the signal message.
	signalMessageStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewPreKeySignalMessageFromStruct(signalMessageStructure, serializer, msgSerializer)
}

// NewPreKeySignalMessageFromStruct will return a new PreKeySignalMessage from the given
// PreKeySignalMessageStructure.
func NewPreKeySignalMessageFromStruct(structure *PreKeySignalMessageStructure,
	serializer PreKeySignalMessageSerializer, msgSerializer SignalMessageSerializer) (*PreKeySignalMessage, error) {

	// Throw an error if the given message structure is an unsupported version.
	if structure.Version <= UnsupportedVersion {
		return nil, fmt.Errorf("%w %d (prekey message)", signalerror.ErrOldMessageVersion, structure.Version)
	}

	// Throw an error if the given message structure is a future version.
	if structure.Version > CurrentVersion {
		return nil, fmt.Errorf("%w %d (prekey message)", signalerror.ErrUnknownMessageVersion, structure.Version)
	}

	// Throw an error if the structure is missing critical fields.
	if structure.BaseKey == nil || structure.IdentityKey == nil || structure.Message == nil {
		return nil, fmt.Errorf("%w (prekey message)", signalerror.ErrIncompleteMessage)
	}

	// Create the signal message object from the structure.
	preKeyWhisperMessage := &PreKeySignalMessage{structure: *structure, serializer: serializer}

	// Generate the base ECC key from bytes.
	var err error
	preKeyWhisperMessage.baseKey, err = ecc.DecodePoint(structure.BaseKey, 0)
	if err != nil {
		return nil, err
	}

	// Generate the identity key from bytes
	var identityKey ecc.ECPublicKeyable
	identityKey, err = ecc.DecodePoint(structure.IdentityKey, 0)
	if err != nil {
		return nil, err
	}
	preKeyWhisperMessage.identityKey = identity.NewKey(identityKey)

	// Generate the SignalMessage object from bytes.
	preKeyWhisperMessage.message, err = NewSignalMessageFromBytes(structure.Message, msgSerializer)
	if err != nil {
		return nil, err
	}

	return preKeyWhisperMessage, nil
}

// NewPreKeySignalMessage will return a new PreKeySignalMessage object.
func NewPreKeySignalMessage(version int, registrationID uint32, preKeyID *optional.Uint32, signedPreKeyID uint32,
	baseKey ecc.ECPublicKeyable, identityKey *identity.Key, message *SignalMessage, serializer PreKeySignalMessageSerializer,
	msgSerializer SignalMessageSerializer) (*PreKeySignalMessage, error) {
	structure := &PreKeySignalMessageStructure{
		Version:        version,
		RegistrationID: registrationID,
		PreKeyID:       preKeyID,
		SignedPreKeyID: signedPreKeyID,
		BaseKey:        baseKey.Serialize(),
		IdentityKey:    identityKey.PublicKey().Serialize(),
		Message:        message.Serialize(),
	}
	return NewPreKeySignalMessageFromStruct(structure, serializer, msgSerializer)
}

// PreKeySignalMessageStructure is a serializable structure for
// PreKeySignalMessages.
type PreKeySignalMessageStructure struct {
	RegistrationID uint32
	PreKeyID       *optional.Uint32
	SignedPreKeyID uint32
	BaseKey        []byte
	IdentityKey    []byte
	Message        []byte
	Version        int
}

// PreKeySignalMessage is an encrypted Signal message that is designed
// to be used when building a session with someone for the first time.
type PreKeySignalMessage struct {
	structure   PreKeySignalMessageStructure
	baseKey     ecc.ECPublicKeyable
	identityKey *identity.Key
	message     *SignalMessage
	serializer  PreKeySignalMessageSerializer
}

func (p *PreKeySignalMessage) MessageVersion() int {
	return p.structure.Version
}

func (p *PreKeySignalMessage) IdentityKey() *identity.Key {
	return p.identityKey
}

func (p *PreKeySignalMessage) RegistrationID() uint32 {
	return p.structure.RegistrationID
}

func (p *PreKeySignalMessage) PreKeyID() *optional.Uint32 {
	return p.structure.PreKeyID
}

func (p *PreKeySignalMessage) SignedPreKeyID() uint32 {
	return p.structure.SignedPreKeyID
}

func (p *PreKeySignalMessage) BaseKey() ecc.ECPublicKeyable {
	return p.baseKey
}

func (p *PreKeySignalMessage) WhisperMessage() *SignalMessage {
	return p.message
}

func (p *PreKeySignalMessage) Serialize() []byte {
	return p.serializer.Serialize(&p.structure)
}

func (p *PreKeySignalMessage) Type() uint32 {
	return PREKEY_TYPE
}
