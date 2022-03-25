package ecc

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"

	"go.mau.fi/libsignal/logger"
)

// DjbType is the Diffie-Hellman curve type (curve25519) created by D. J. Bernstein.
const DjbType = 0x05

var ErrBadKeyType = errors.New("bad key type")

// DecodePoint will take the given bytes and offset and return an ECPublicKeyable object.
// This is used to check the byte at the given offset in the byte array for a special
// "type" byte that will determine the key type. Currently only DJB EC keys are supported.
func DecodePoint(bytes []byte, offset int) (ECPublicKeyable, error) {
	keyType := bytes[offset] & 0xFF

	switch keyType {
	case DjbType:
		keyBytes := [32]byte{}
		copy(keyBytes[:], bytes[offset+1:])
		return NewDjbECPublicKey(keyBytes), nil
	default:
		return nil, fmt.Errorf("%w %d", ErrBadKeyType, keyType)
	}
}

func CreateKeyPair(privateKey []byte) *ECKeyPair {
	var private, public [32]byte
	copy(private[:], privateKey)

	private[0] &= 248
	private[31] &= 127
	private[31] |= 64

	curve25519.ScalarBaseMult(&public, &private)

	// Put data into our keypair struct
	djbECPub := NewDjbECPublicKey(public)
	djbECPriv := NewDjbECPrivateKey(private)
	keypair := NewECKeyPair(djbECPub, djbECPriv)

	logger.Debug("Returning keypair: ", keypair)
	return keypair
}

// GenerateKeyPair returns an EC Key Pair.
func GenerateKeyPair() (*ECKeyPair, error) {
	// logger.Debug("Generating EC Key Pair...")
	// Get cryptographically secure random numbers.
	random := rand.Reader

	// Create a byte array for our public and private keys.
	var private, public [32]byte

	// Generate some random data
	_, err := io.ReadFull(random, private[:])
	if err != nil {
		return nil, err
	}

	// Documented at: http://cr.yp.to/ecdh.html
	private[0] &= 248
	private[31] &= 127
	private[31] |= 64

	curve25519.ScalarBaseMult(&public, &private)

	// Put data into our keypair struct
	djbECPub := NewDjbECPublicKey(public)
	djbECPriv := NewDjbECPrivateKey(private)
	keypair := NewECKeyPair(djbECPub, djbECPriv)

	// logger.Debug("Returning keypair: ", keypair)

	return keypair, nil
}

// VerifySignature verifies that the message was signed with the given key.
func VerifySignature(signingKey ECPublicKeyable, message []byte, signature [64]byte) bool {
	logger.Debug("Verifying signature of bytes: ", message)
	publicKey := signingKey.PublicKey()
	valid := verify(publicKey, message, &signature)
	logger.Debug("Signature valid: ", valid)
	return valid
}

// CalculateSignature signs a message with the given private key.
func CalculateSignature(signingKey ECPrivateKeyable, message []byte) [64]byte {
	logger.Debug("Signing bytes with signing key")
	// Get cryptographically secure random numbers.
	var random [64]byte
	r := rand.Reader
	io.ReadFull(r, random[:])

	// Get the private key.
	privateKey := signingKey.Serialize()

	// Sign the message.
	signature := sign(&privateKey, message, random)
	return *signature
}
