package ecc

// ECPrivateKeyable is an interface for all elliptic curve private keys.
type ECPrivateKeyable interface {
	Serialize() [32]byte
	Type() int
}
