package record

import (
	"fmt"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/signalerror"
)

const maxStates = 5

// SenderKeySerializer is an interface for serializing and deserializing
// SenderKey objects into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type SenderKeySerializer interface {
	Serialize(preKey *SenderKeyStructure) []byte
	Deserialize(serialized []byte) (*SenderKeyStructure, error)
}

// NewSenderKeyFromBytes will return a prekey record from the given bytes using the given serializer.
func NewSenderKeyFromBytes(serialized []byte, serializer SenderKeySerializer,
	stateSerializer SenderKeyStateSerializer) (*SenderKey, error) {

	// Use the given serializer to decode the senderkey record
	senderKeyStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewSenderKeyFromStruct(senderKeyStructure, serializer, stateSerializer)
}

// NewSenderKeyFromStruct returns a SenderKey record using the given serializable structure.
func NewSenderKeyFromStruct(structure *SenderKeyStructure, serializer SenderKeySerializer,
	stateSerializer SenderKeyStateSerializer) (*SenderKey, error) {

	// Build our sender key states from structure.
	senderKeyStates := make([]*SenderKeyState, len(structure.SenderKeyStates))
	for i := range structure.SenderKeyStates {
		var err error
		senderKeyStates[i], err = NewSenderKeyStateFromStructure(structure.SenderKeyStates[i], stateSerializer)
		if err != nil {
			return nil, err
		}
	}

	// Build and return our session.
	senderKey := &SenderKey{
		senderKeyStates: senderKeyStates,
		serializer:      serializer,
	}

	return senderKey, nil

}

// NewSenderKey record returns a new sender key record that can
// be stored in a SenderKeyStore.
func NewSenderKey(serializer SenderKeySerializer,
	stateSerializer SenderKeyStateSerializer) *SenderKey {

	return &SenderKey{
		senderKeyStates: []*SenderKeyState{},
		serializer:      serializer,
		stateSerializer: stateSerializer,
	}
}

// SenderKeyStructure is a structure for serializing SenderKey records.
type SenderKeyStructure struct {
	SenderKeyStates []*SenderKeyStateStructure
}

// SenderKey record is a structure for storing pre keys inside
// a SenderKeyStore.
type SenderKey struct {
	senderKeyStates []*SenderKeyState
	serializer      SenderKeySerializer
	stateSerializer SenderKeyStateSerializer
}

// SenderKeyState will return the first sender key state in the record's
// list of sender key states.
func (k *SenderKey) SenderKeyState() (*SenderKeyState, error) {
	if len(k.senderKeyStates) > 0 {
		return k.senderKeyStates[0], nil
	}
	return nil, signalerror.ErrNoSenderKeyStatesInRecord
}

// GetSenderKeyStateByID will return the sender key state with the given
// key id.
func (k *SenderKey) GetSenderKeyStateByID(keyID uint32) (*SenderKeyState, error) {
	for i := 0; i < len(k.senderKeyStates); i++ {
		if k.senderKeyStates[i].KeyID() == keyID {
			return k.senderKeyStates[i], nil
		}
	}

	return nil, fmt.Errorf("%w %d", signalerror.ErrNoSenderKeyStateForID, keyID)
}

// IsEmpty will return false if there is more than one state in this
// senderkey record.
func (k *SenderKey) IsEmpty() bool {
	return len(k.senderKeyStates) == 0
}

// AddSenderKeyState will add a new state to this senderkey record with the given
// id, iteration, chainkey, and signature key.
func (k *SenderKey) AddSenderKeyState(id uint32, iteration uint32,
	chainKey []byte, signatureKey ecc.ECPublicKeyable) {

	newState := NewSenderKeyStateFromPublicKey(id, iteration, chainKey, signatureKey, k.stateSerializer)
	k.senderKeyStates = append([]*SenderKeyState{newState}, k.senderKeyStates...)

	if len(k.senderKeyStates) > maxStates {
		k.senderKeyStates = k.senderKeyStates[:len(k.senderKeyStates)-1]
	}
}

// SetSenderKeyState will  replace the current senderkey states with the given
// senderkey state.
func (k *SenderKey) SetSenderKeyState(id uint32, iteration uint32,
	chainKey []byte, signatureKey *ecc.ECKeyPair) {

	newState := NewSenderKeyState(id, iteration, chainKey, signatureKey, k.stateSerializer)
	k.senderKeyStates = make([]*SenderKeyState, 0, maxStates/2)
	k.senderKeyStates = append(k.senderKeyStates, newState)
}

// Serialize will return the record as serialized bytes so it can be
// persistently stored.
func (k *SenderKey) Serialize() []byte {
	return k.serializer.Serialize(k.Structure())
}

// Structure will return a simple serializable record structure.
// This is used for serialization to persistently
// store a session record.
func (k *SenderKey) Structure() *SenderKeyStructure {
	senderKeyStates := make([]*SenderKeyStateStructure, len(k.senderKeyStates))
	for i := range k.senderKeyStates {
		senderKeyStates[i] = k.senderKeyStates[i].structure()
	}
	return &SenderKeyStructure{
		SenderKeyStates: senderKeyStates,
	}
}
