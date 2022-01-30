package record

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/util/optional"
)

// NewPendingPreKey will return a new pending pre key object.
func NewPendingPreKey(preKeyID *optional.Uint32, signedPreKeyID uint32,
	baseKey ecc.ECPublicKeyable) *PendingPreKey {

	return &PendingPreKey{
		preKeyID:       preKeyID,
		signedPreKeyID: signedPreKeyID,
		baseKey:        baseKey,
	}
}

// NewPendingPreKeyFromStruct will return a new pending prekey object from the
// given structure.
func NewPendingPreKeyFromStruct(preKey *PendingPreKeyStructure) (*PendingPreKey, error) {
	baseKey, err := ecc.DecodePoint(preKey.BaseKey, 0)
	if err != nil {
		return nil, err
	}

	pendingPreKey := NewPendingPreKey(
		preKey.PreKeyID,
		preKey.SignedPreKeyID,
		baseKey,
	)

	return pendingPreKey, nil
}

// PendingPreKeyStructure is a serializeable structure for pending
// prekeys.
type PendingPreKeyStructure struct {
	PreKeyID       *optional.Uint32
	SignedPreKeyID uint32
	BaseKey        []byte
}

// PendingPreKey is a structure for pending pre keys
// for a session state.
type PendingPreKey struct {
	preKeyID       *optional.Uint32
	signedPreKeyID uint32
	baseKey        ecc.ECPublicKeyable
}

// structure will return a serializeable structure of the pending prekey.
func (p *PendingPreKey) structure() *PendingPreKeyStructure {
	if p != nil {
		return &PendingPreKeyStructure{
			PreKeyID:       p.preKeyID,
			SignedPreKeyID: p.signedPreKeyID,
			BaseKey:        p.baseKey.Serialize(),
		}
	}
	return nil
}
