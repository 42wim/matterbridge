package ecc

// NewECKeyPair returns a new elliptic curve keypair given the specified public and private keys.
func NewECKeyPair(publicKey ECPublicKeyable, privateKey ECPrivateKeyable) *ECKeyPair {
	keypair := ECKeyPair{
		publicKey:  publicKey,
		privateKey: privateKey,
	}

	return &keypair
}

// ECKeyPair is a combination of both public and private elliptic curve keys.
type ECKeyPair struct {
	publicKey  ECPublicKeyable
	privateKey ECPrivateKeyable
}

// PublicKey returns the public key from the key pair.
func (e *ECKeyPair) PublicKey() ECPublicKeyable {
	return e.publicKey
}

// PrivateKey returns the private key from the key pair.
func (e *ECKeyPair) PrivateKey() ECPrivateKeyable {
	return e.privateKey
}
