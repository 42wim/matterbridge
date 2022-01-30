package identity

import (
	"go.mau.fi/libsignal/ecc"
)

// NewKeyPair returns a new identity key with the given public and private keys.
func NewKeyPair(publicKey *Key, privateKey ecc.ECPrivateKeyable) *KeyPair {
	keyPair := KeyPair{
		publicKey:  publicKey,
		privateKey: privateKey,
	}

	return &keyPair
}

// NewKeyPairFromBytes returns a new identity key from the given serialized bytes.
//func NewKeyPairFromBytes(serialized []byte) KeyPair {
//}

// KeyPair is a holder for public and private identity key pair.
type KeyPair struct {
	publicKey  *Key
	privateKey ecc.ECPrivateKeyable
}

// PublicKey returns the identity key's public key.
func (k *KeyPair) PublicKey() *Key {
	return k.publicKey
}

// PrivateKey returns the identity key's private key.
func (k *KeyPair) PrivateKey() ecc.ECPrivateKeyable {
	return k.privateKey
}

// Serialize returns a byte array that represents the keypair.
//func (k *KeyPair) Serialize() []byte {
//}
