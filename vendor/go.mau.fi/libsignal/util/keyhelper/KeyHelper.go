// Package keyhelper is based on: https://github.com/WhisperSystems/libsignal-protocol-java/blob/master/java/src/main/java/org/whispersystems/libsignal/util/KeyHelper.java
package keyhelper

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/state/record"
)

// GenerateIdentityKeyPair generates an identity keypair used for
// signing. Clients should only do this once at install time.
func GenerateIdentityKeyPair() (*identity.KeyPair, error) {
	keyPair, err := ecc.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	publicKey := identity.NewKey(keyPair.PublicKey())
	return identity.NewKeyPair(publicKey, keyPair.PrivateKey()), nil
}

// GeneratePreKeys generates a list of PreKeys. Client shsould do this at
// install time, and subsequently any time the list of PreKeys stored on
// the server runs low.
//
// PreKeys IDs are shorts, so they will eventually be repeated. Clients
// should store PreKeys in a circular buffer, so that they are repeated
// as infrequently as possible.
func GeneratePreKeys(start int, count int, serializer record.PreKeySerializer) ([]*record.PreKey, error) {
	var preKeys []*record.PreKey

	for i := start; i <= count; i++ {
		key, err := ecc.GenerateKeyPair()
		if err != nil {
			return nil, err
		}
		preKeys = append(preKeys, record.NewPreKey(uint32(i), key, serializer))
	}

	return preKeys, nil
}

// GenerateLastResortKey will generate the last resort PreKey. Clients should
// do this only once, at install time, and durably store it for the length
// of the install.
func GenerateLastResortKey(serializer record.PreKeySerializer) (*record.PreKey, error) {
	keyPair, err := ecc.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	return record.NewPreKey(0, keyPair, serializer), nil
}

// GenerateSignedPreKey generates a signed PreKey.
func GenerateSignedPreKey(identityKeyPair *identity.KeyPair, signedPreKeyID uint32, serializer record.SignedPreKeySerializer) (*record.SignedPreKey, error) {
	keyPair, err := ecc.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	signature := ecc.CalculateSignature(identityKeyPair.PrivateKey(), keyPair.PublicKey().Serialize())
	timestamp := time.Now().Unix()

	return record.NewSignedPreKey(signedPreKeyID, timestamp, keyPair, signature, serializer), nil
}

// GenerateRegistrationID generates a registration ID. Clients should only do
// this once, at install time.
func GenerateRegistrationID() uint32 {
	var n uint32
	binary.Read(rand.Reader, binary.LittleEndian, &n)

	return n
}

//---------- Group Stuff ----------------

func GenerateSenderSigningKey() (*ecc.ECKeyPair, error) {
	return ecc.GenerateKeyPair()
}

func GenerateSenderKey() []byte {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)
	return randBytes
}

func GenerateSenderKeyID() uint32 {
	return GenerateRegistrationID()
}

//---------- End Group Stuff --------------
