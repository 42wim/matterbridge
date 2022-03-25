package record

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/groups/ratchet"
	"go.mau.fi/libsignal/util/bytehelper"
)

const maxMessageKeys = 2000

// SenderKeyStateSerializer is an interface for serializing and deserializing
// a Signal State into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type SenderKeyStateSerializer interface {
	Serialize(state *SenderKeyStateStructure) []byte
	Deserialize(serialized []byte) (*SenderKeyStateStructure, error)
}

// NewSenderKeyStateFromBytes will return a Signal State from the given
// bytes using the given serializer.
func NewSenderKeyStateFromBytes(serialized []byte, serializer SenderKeyStateSerializer) (*SenderKeyState, error) {
	// Use the given serializer to decode the signal message.
	stateStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewSenderKeyStateFromStructure(stateStructure, serializer)
}

// NewSenderKeyState returns a new SenderKeyState.
func NewSenderKeyState(keyID uint32, iteration uint32, chainKey []byte,
	signatureKey *ecc.ECKeyPair, serializer SenderKeyStateSerializer) *SenderKeyState {

	return &SenderKeyState{
		keys:           make([]*ratchet.SenderMessageKey, 0, maxMessageKeys/2),
		keyID:          keyID,
		senderChainKey: ratchet.NewSenderChainKey(iteration, chainKey),
		signingKeyPair: signatureKey,
		serializer:     serializer,
	}
}

// NewSenderKeyStateFromPublicKey returns a new SenderKeyState with the given publicKey.
func NewSenderKeyStateFromPublicKey(keyID uint32, iteration uint32, chainKey []byte,
	signatureKey ecc.ECPublicKeyable, serializer SenderKeyStateSerializer) *SenderKeyState {

	keyPair := ecc.NewECKeyPair(signatureKey, nil)

	return &SenderKeyState{
		keys:           make([]*ratchet.SenderMessageKey, 0, maxMessageKeys/2),
		keyID:          keyID,
		senderChainKey: ratchet.NewSenderChainKey(iteration, chainKey),
		signingKeyPair: keyPair,
		serializer:     serializer,
	}
}

// NewSenderKeyStateFromStructure will return a new session state with the
// given state structure. This structure is given back from an
// implementation of the sender key state serializer.
func NewSenderKeyStateFromStructure(structure *SenderKeyStateStructure,
	serializer SenderKeyStateSerializer) (*SenderKeyState, error) {

	// Convert our ecc keys from bytes into object form.
	signingKeyPublic, err := ecc.DecodePoint(structure.SigningKeyPublic, 0)
	if err != nil {
		return nil, err
	}
	signingKeyPrivate := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(structure.SigningKeyPrivate))

	// Build our sender message keys from structure
	senderMessageKeys := make([]*ratchet.SenderMessageKey, len(structure.Keys))
	for i := range structure.Keys {
		senderMessageKeys[i] = ratchet.NewSenderMessageKeyFromStruct(structure.Keys[i])
	}

	// Build our state object.
	state := &SenderKeyState{
		keys:           senderMessageKeys,
		keyID:          structure.KeyID,
		senderChainKey: ratchet.NewSenderChainKeyFromStruct(structure.SenderChainKey),
		signingKeyPair: ecc.NewECKeyPair(signingKeyPublic, signingKeyPrivate),
		serializer:     serializer,
	}

	return state, nil
}

// SenderKeyStateStructure is a serializeable structure of SenderKeyState.
type SenderKeyStateStructure struct {
	Keys              []*ratchet.SenderMessageKeyStructure
	KeyID             uint32
	SenderChainKey    *ratchet.SenderChainKeyStructure
	SigningKeyPrivate []byte
	SigningKeyPublic  []byte
}

// SenderKeyState is a structure for maintaining a senderkey session state.
type SenderKeyState struct {
	keys           []*ratchet.SenderMessageKey
	keyID          uint32
	senderChainKey *ratchet.SenderChainKey
	signingKeyPair *ecc.ECKeyPair
	serializer     SenderKeyStateSerializer
}

// SigningKey returns the signing key pair of the sender key state.
func (k *SenderKeyState) SigningKey() *ecc.ECKeyPair {
	return k.signingKeyPair
}

// SenderChainKey returns the sender chain key of the state.
func (k *SenderKeyState) SenderChainKey() *ratchet.SenderChainKey {
	return k.senderChainKey
}

// KeyID returns the state's key id.
func (k *SenderKeyState) KeyID() uint32 {
	return k.keyID
}

// HasSenderMessageKey will return true if the state has a key with the
// given iteration.
func (k *SenderKeyState) HasSenderMessageKey(iteration uint32) bool {
	for i := 0; i < len(k.keys); i++ {
		if k.keys[i].Iteration() == iteration {
			return true
		}
	}
	return false
}

// AddSenderMessageKey will add the given sender message key to the state.
func (k *SenderKeyState) AddSenderMessageKey(senderMsgKey *ratchet.SenderMessageKey) {
	k.keys = append(k.keys, senderMsgKey)

	if len(k.keys) > maxMessageKeys {
		k.keys = k.keys[1:]
	}
}

// SetSenderChainKey will set the state's sender chain key with the given key.
func (k *SenderKeyState) SetSenderChainKey(senderChainKey *ratchet.SenderChainKey) {
	k.senderChainKey = senderChainKey
}

// RemoveSenderMessageKey will remove the key in this state with the given iteration number.
func (k *SenderKeyState) RemoveSenderMessageKey(iteration uint32) *ratchet.SenderMessageKey {
	for i := 0; i < len(k.keys); i++ {
		if k.keys[i].Iteration() == iteration {
			removed := k.keys[i]
			k.keys = append(k.keys[0:i], k.keys[i+1:]...)
			return removed
		}
	}

	return nil
}

// Serialize will return the state as bytes using the given serializer.
func (k *SenderKeyState) Serialize() []byte {
	return k.serializer.Serialize(k.structure())
}

// structure will return a serializable structure of the
// the given state so it can be persistently stored.
func (k *SenderKeyState) structure() *SenderKeyStateStructure {
	// Convert our sender message keys into a serializeable structure
	keys := make([]*ratchet.SenderMessageKeyStructure, len(k.keys))
	for i := range k.keys {
		keys[i] = ratchet.NewStructFromSenderMessageKey(k.keys[i])
	}

	// Build and return our state structure.
	s := &SenderKeyStateStructure{
		Keys:             keys,
		KeyID:            k.keyID,
		SenderChainKey:   ratchet.NewStructFromSenderChainKey(k.senderChainKey),
		SigningKeyPublic: k.signingKeyPair.PublicKey().Serialize(),
	}
	if k.signingKeyPair.PrivateKey() != nil {
		s.SigningKeyPrivate = bytehelper.ArrayToSlice(k.signingKeyPair.PrivateKey().Serialize())
	}
	return s
}
