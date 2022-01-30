package ecc

// KeySize is the size of EC keys (32) with the EC type byte prepended to it.
const KeySize int = 33

// ECPublicKeyable is an interface for all elliptic curve public keys.
type ECPublicKeyable interface {
	Serialize() []byte
	Type() int
	PublicKey() [32]byte
}
