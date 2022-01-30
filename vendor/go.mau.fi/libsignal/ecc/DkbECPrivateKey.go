package ecc

// NewDjbECPrivateKey returns a new EC private key with the given bytes.
func NewDjbECPrivateKey(key [32]byte) *DjbECPrivateKey {
	private := DjbECPrivateKey{
		privateKey: key,
	}
	return &private
}

// DjbECPrivateKey implements the ECPrivateKey interface and uses Curve25519.
type DjbECPrivateKey struct {
	privateKey [32]byte
}

// PrivateKey returns the private key as a byte-array.
func (d *DjbECPrivateKey) PrivateKey() [32]byte {
	return d.privateKey
}

// Serialize returns the private key as a byte-array.
func (d *DjbECPrivateKey) Serialize() [32]byte {
	return d.privateKey
}

// Type returns the EC type value.
func (d *DjbECPrivateKey) Type() int {
	return DjbType
}
