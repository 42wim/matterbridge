// Package identity provides identity keys used for verifying the identity
// of a signal user.
package identity

import (
	"encoding/hex"
	"go.mau.fi/libsignal/ecc"
)

// NewKey generates a new IdentityKey from an ECPublicKey
func NewKey(publicKey ecc.ECPublicKeyable) *Key {
	identityKey := Key{
		publicKey: publicKey,
	}

	return &identityKey
}

// NewKeyFromBytes generates a new IdentityKey from public key bytes
func NewKeyFromBytes(publicKey [32]byte, offset int) Key {
	identityKey := Key{
		publicKey: ecc.NewDjbECPublicKey(publicKey),
	}

	return identityKey
}

// Key is a structure for representing an identity key. This same structure can
// be used for verifying recipient's identity key or storing our own identity key.
type Key struct {
	publicKey ecc.ECPublicKeyable
}

// Fingerprint gets the string fingerprint representation of the public key.
func (k *Key) Fingerprint() string {
	return hex.EncodeToString(k.publicKey.Serialize())
}

// PublicKey returns the EC Public key of the identity key
func (k *Key) PublicKey() ecc.ECPublicKeyable {
	return k.publicKey
}

// Serialize returns the serialized version of the key
func (k *Key) Serialize() []byte {
	return k.publicKey.Serialize()
}
