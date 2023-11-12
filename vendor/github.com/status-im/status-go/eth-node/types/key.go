package types

import (
	"crypto/ecdsa"

	"github.com/google/uuid"

	"github.com/status-im/status-go/extkeys"
)

type Key struct {
	ID uuid.UUID // Version 4 "random" for unique id not derived from key data
	// to simplify lookups we also store the address
	Address Address
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	PrivateKey *ecdsa.PrivateKey
	// ExtendedKey is the extended key of the PrivateKey itself, and it's used
	// to derive child keys.
	ExtendedKey *extkeys.ExtendedKey
	// SubAccountIndex is DEPRECATED
	// It was use in Status to keep track of the number of sub-account created
	// before having multi-account support.
	SubAccountIndex uint32
}
