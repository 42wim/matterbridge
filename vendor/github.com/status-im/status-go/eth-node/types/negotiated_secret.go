package types

import (
	"crypto/ecdsa"
)

// NegotiatedSecret represents a negotiated secret (both public and private keys)
type NegotiatedSecret struct {
	PublicKey *ecdsa.PublicKey
	Key       []byte
}
