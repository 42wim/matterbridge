package record

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/util/bytehelper"
)

// SignedPreKeySerializer is an interface for serializing and deserializing
// SignedPreKey objects into bytes. An implementation of this interface should be
// used to encode/decode the object into JSON, Protobuffers, etc.
type SignedPreKeySerializer interface {
	Serialize(signedPreKey *SignedPreKeyStructure) []byte
	Deserialize(serialized []byte) (*SignedPreKeyStructure, error)
}

// NewSignedPreKeyFromBytes will return a signed prekey record from the given
// bytes using the given serializer.
func NewSignedPreKeyFromBytes(serialized []byte, serializer SignedPreKeySerializer) (*SignedPreKey, error) {
	// Use the given serializer to decode the signal message.
	signedPreKeyStructure, err := serializer.Deserialize(serialized)
	if err != nil {
		return nil, err
	}

	return NewSignedPreKeyFromStruct(signedPreKeyStructure, serializer)
}

// NewSignedPreKeyFromStruct returns a SignedPreKey record using the given
// serializable structure.
func NewSignedPreKeyFromStruct(structure *SignedPreKeyStructure,
	serializer SignedPreKeySerializer) (*SignedPreKey, error) {

	// Create the signed prekey record from the structure.
	signedPreKey := &SignedPreKey{
		structure:  *structure,
		serializer: serializer,
		signature:  bytehelper.SliceToArray64(structure.Signature),
	}

	// Generate the ECC key from bytes.
	publicKey := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(structure.PublicKey))
	privateKey := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(structure.PrivateKey))
	keyPair := ecc.NewECKeyPair(publicKey, privateKey)
	signedPreKey.keyPair = keyPair

	return signedPreKey, nil
}

// NewSignedPreKey record creates a new signed pre key record
// with the given properties.
func NewSignedPreKey(id uint32, timestamp int64, keyPair *ecc.ECKeyPair,
	sig [64]byte, serializer SignedPreKeySerializer) *SignedPreKey {

	return &SignedPreKey{
		structure: SignedPreKeyStructure{
			ID:         id,
			Timestamp:  timestamp,
			PublicKey:  keyPair.PublicKey().Serialize(),
			PrivateKey: bytehelper.ArrayToSlice(keyPair.PrivateKey().Serialize()),
			Signature:  bytehelper.ArrayToSlice64(sig),
		},
		keyPair:    keyPair,
		signature:  sig,
		serializer: serializer,
	}
}

// SignedPreKeyStructure is a flat structure of a signed pre key, used
// for serialization and deserialization.
type SignedPreKeyStructure struct {
	ID         uint32
	PublicKey  []byte
	PrivateKey []byte
	Signature  []byte
	Timestamp  int64
}

// SignedPreKey record is a structure for storing a signed
// pre key in a SignedPreKey store.
type SignedPreKey struct {
	structure  SignedPreKeyStructure
	keyPair    *ecc.ECKeyPair
	signature  [64]byte
	serializer SignedPreKeySerializer
}

// ID returns the record's id.
func (s *SignedPreKey) ID() uint32 {
	return s.structure.ID
}

// Timestamp returns the record's timestamp
func (s *SignedPreKey) Timestamp() int64 {
	return s.structure.Timestamp
}

// KeyPair returns the signed pre key record's key pair.
func (s *SignedPreKey) KeyPair() *ecc.ECKeyPair {
	return s.keyPair
}

// Signature returns the record's signed prekey signature.
func (s *SignedPreKey) Signature() [64]byte {
	return s.signature
}

// Serialize uses the SignedPreKey serializer to return the SignedPreKey
// as serialized bytes.
func (s *SignedPreKey) Serialize() []byte {
	structure := s.structure
	return s.serializer.Serialize(&structure)
}
