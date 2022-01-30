// Package prekey provides prekey bundle structures for calculating
// a new Signal session with a user asyncronously.
package prekey

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/util/optional"
)

// NewBundle returns a Bundle structure that contains a remote PreKey
// and collection of associated items.
func NewBundle(registrationID, deviceID uint32, preKeyID *optional.Uint32, signedPreKeyID uint32,
	preKeyPublic, signedPreKeyPublic ecc.ECPublicKeyable, signedPreKeySig [64]byte,
	identityKey *identity.Key) *Bundle {

	bundle := Bundle{
		registrationID:        registrationID,
		deviceID:              deviceID,
		preKeyID:              preKeyID,
		preKeyPublic:          preKeyPublic,
		signedPreKeyID:        signedPreKeyID,
		signedPreKeyPublic:    signedPreKeyPublic,
		signedPreKeySignature: signedPreKeySig,
		identityKey:           identityKey,
	}

	return &bundle
}

// Bundle is a structure that contains a remote PreKey and collection
// of associated items.
type Bundle struct {
	registrationID        uint32
	deviceID              uint32
	preKeyID              *optional.Uint32
	preKeyPublic          ecc.ECPublicKeyable
	signedPreKeyID        uint32
	signedPreKeyPublic    ecc.ECPublicKeyable
	signedPreKeySignature [64]byte
	identityKey           *identity.Key
}

// DeviceID returns the device ID this PreKey belongs to.
func (b *Bundle) DeviceID() uint32 {
	return b.deviceID
}

// PreKeyID returns the unique key ID for this PreKey.
func (b *Bundle) PreKeyID() *optional.Uint32 {
	return b.preKeyID
}

// PreKey returns the public key for this PreKey.
func (b *Bundle) PreKey() ecc.ECPublicKeyable {
	return b.preKeyPublic
}

// SignedPreKeyID returns the unique key ID for this
// signed PreKey.
func (b *Bundle) SignedPreKeyID() uint32 {
	return b.signedPreKeyID
}

// SignedPreKey returns the signed PreKey for this
// PreKeyBundle.
func (b *Bundle) SignedPreKey() ecc.ECPublicKeyable {
	return b.signedPreKeyPublic
}

// SignedPreKeySignature returns the signature over the
// signed PreKey.
func (b *Bundle) SignedPreKeySignature() [64]byte {
	return b.signedPreKeySignature
}

// IdentityKey returns the Identity Key of this PreKey's owner.
func (b *Bundle) IdentityKey() *identity.Key {
	return b.identityKey
}

// RegistrationID returns the registration ID associated with
// this PreKey.
func (b *Bundle) RegistrationID() uint32 {
	return b.registrationID
}
