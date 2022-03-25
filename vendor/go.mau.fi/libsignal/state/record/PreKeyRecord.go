package record

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/util/bytehelper"
	"go.mau.fi/libsignal/util/optional"
)

// PreKeySerializer is an interface for serializing and deserializing
// PreKey objects into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type PreKeySerializer interface {
	Serialize(preKey *PreKeyStructure) []byte
	Deserialize(serialized []byte) (*PreKeyStructure, error)
}

// NewPreKeyFromBytes will return a prekey record from the given bytes using the given serializer.
func NewPreKeyFromBytes(serialized []byte, serializer PreKeySerializer) (*PreKey, error) {
	// Use the given serializer to decode the signal message.
	preKeyStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewPreKeyFromStruct(preKeyStructure, serializer)
}

// NewPreKeyFromStruct returns a PreKey record using the given serializable structure.
func NewPreKeyFromStruct(structure *PreKeyStructure, serializer PreKeySerializer) (*PreKey, error) {
	// Create the prekey record from the structure.
	preKey := &PreKey{
		structure:  *structure,
		serializer: serializer,
	}

	// Generate the ECC key from bytes.
	publicKey := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(structure.PublicKey))
	privateKey := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(structure.PrivateKey))
	keyPair := ecc.NewECKeyPair(publicKey, privateKey)
	preKey.keyPair = keyPair

	return preKey, nil
}

// NewPreKey record returns a new pre key record that can
// be stored in a PreKeyStore.
func NewPreKey(id uint32, keyPair *ecc.ECKeyPair, serializer PreKeySerializer) *PreKey {
	return &PreKey{
		structure: PreKeyStructure{
			ID:         id,
			PublicKey:  keyPair.PublicKey().Serialize(),
			PrivateKey: bytehelper.ArrayToSlice(keyPair.PrivateKey().Serialize()),
		},
		keyPair:    keyPair,
		serializer: serializer,
	}
}

// PreKeyStructure is a structure for serializing PreKey records.
type PreKeyStructure struct {
	ID         uint32
	PublicKey  []byte
	PrivateKey []byte
}

// PreKey record is a structure for storing pre keys inside
// a PreKeyStore.
type PreKey struct {
	structure  PreKeyStructure
	keyPair    *ecc.ECKeyPair
	serializer PreKeySerializer
}

// ID returns the pre key record's id.
func (p *PreKey) ID() *optional.Uint32 {
	// TODO: manually set this to empty if empty
	return optional.NewOptionalUint32(p.structure.ID)
}

// KeyPair returns the pre key record's key pair.
func (p *PreKey) KeyPair() *ecc.ECKeyPair {
	return p.keyPair
}

// Serialize uses the PreKey serializer to return the PreKey
// as serialized bytes.
func (p *PreKey) Serialize() []byte {
	structure := p.structure
	return p.serializer.Serialize(&structure)
}
