package record

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/util/optional"
)

// NewUnackPreKeyMessageItems returns message items that are unacknowledged.
func NewUnackPreKeyMessageItems(preKeyID *optional.Uint32, signedPreKeyID uint32,
	baseKey ecc.ECPublicKeyable) *UnackPreKeyMessageItems {

	return &UnackPreKeyMessageItems{
		preKeyID:       preKeyID,
		signedPreKeyID: signedPreKeyID,
		baseKey:        baseKey,
	}
}

// NewUnackPreKeyMessageItemsFromStruct will return a new unacknowledged prekey
// message items object from the given structure.
func NewUnackPreKeyMessageItemsFromStruct(structure *UnackPreKeyMessageItemsStructure) *UnackPreKeyMessageItems {
	baseKey, _ := ecc.DecodePoint(structure.BaseKey, 0)
	return NewUnackPreKeyMessageItems(
		structure.PreKeyID,
		structure.SignedPreKeyID,
		baseKey,
	)
}

// UnackPreKeyMessageItemsStructure is a serializable structure for unackowledged
// prekey message items.
type UnackPreKeyMessageItemsStructure struct {
	PreKeyID       *optional.Uint32
	SignedPreKeyID uint32
	BaseKey        []byte
}

// UnackPreKeyMessageItems is a structure for messages that have not been
// acknowledged.
type UnackPreKeyMessageItems struct {
	preKeyID       *optional.Uint32
	signedPreKeyID uint32
	baseKey        ecc.ECPublicKeyable
}

// PreKeyID returns the prekey id of the unacknowledged message.
func (u *UnackPreKeyMessageItems) PreKeyID() *optional.Uint32 {
	return u.preKeyID
}

// SignedPreKeyID returns the signed prekey id of the unacknowledged message.
func (u *UnackPreKeyMessageItems) SignedPreKeyID() uint32 {
	return u.signedPreKeyID
}

// BaseKey returns the ECC public key of the unacknowledged message.
func (u *UnackPreKeyMessageItems) BaseKey() ecc.ECPublicKeyable {
	return u.baseKey
}

// structure will return a serializable base structure
// for unacknowledged prekey message items.
func (u *UnackPreKeyMessageItems) structure() *UnackPreKeyMessageItemsStructure {
	return &UnackPreKeyMessageItemsStructure{
		PreKeyID:       u.preKeyID,
		SignedPreKeyID: u.signedPreKeyID,
		BaseKey:        u.baseKey.Serialize(),
	}
}
