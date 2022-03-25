// Package kdf provides a key derivation function to calculate key output
// and negotiate shared secrets for curve X25519 keys.
package kdf

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// HKDF is a hashed key derivation function type that can be used to derive keys.
type HKDF func(inputKeyMaterial, salt, info []byte, outputLength int) ([]byte, error)

// DeriveSecrets derives the requested number of bytes using HKDF with the given
// input, salt, and info.
func DeriveSecrets(inputKeyMaterial, salt, info []byte, outputLength int) ([]byte, error) {
	kdf := hkdf.New(sha256.New, inputKeyMaterial, salt, info)

	secrets := make([]byte, outputLength)
	length, err := io.ReadFull(kdf, secrets)
	if err != nil {
		return nil, err
	}
	if length != outputLength {
		return nil, err
	}

	return secrets, nil
}

// CalculateSharedSecret uses DH Curve25519 to find a shared secret. The result of this function
// should be used in `DeriveSecrets` to output the Root and Chain keys.
func CalculateSharedSecret(theirKey, ourKey [32]byte) [32]byte {
	var sharedSecret [32]byte
	curve25519.ScalarMult(&sharedSecret, &ourKey, &theirKey)

	return sharedSecret
}

// KeyMaterial is a structure for representing a cipherkey, mac, and iv
type KeyMaterial struct {
	CipherKey []byte
	MacKey    []byte
	IV        []byte
}
