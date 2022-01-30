// Package ratchet provides the methods necessary to establish a new double
// ratchet session.
package ratchet

import (
	"encoding/base64"
	"encoding/binary"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/kdf"
	"go.mau.fi/libsignal/keys/chain"
	"go.mau.fi/libsignal/keys/root"
	"go.mau.fi/libsignal/keys/session"
)

var b64 = base64.StdEncoding.EncodeToString

func genDiscontinuity() [32]byte {
	var discontinuity [32]byte
	for i := range discontinuity {
		discontinuity[i] = 0xFF
	}
	return discontinuity
}

// CalculateSenderSession calculates the key agreement for a recipient. This
// should be used when we are trying to send a message to someone for the
// first time.
func CalculateSenderSession(parameters *SenderParameters) (*session.KeyPair, error) {
	var secret [32]byte
	var publicKey [32]byte
	var privateKey [32]byte
	masterSecret := []byte{} // Create a master shared secret that is 5 different 32-byte values
	discontinuity := genDiscontinuity()
	masterSecret = append(masterSecret, discontinuity[:]...)

	// Calculate the agreement using their signed prekey and our identity key.
	publicKey = parameters.TheirSignedPreKey().PublicKey()
	privateKey = parameters.OurIdentityKey().PrivateKey().Serialize()
	secret = kdf.CalculateSharedSecret(
		publicKey,
		privateKey,
	)
	masterSecret = append(masterSecret, secret[:]...)

	// Calculate the agreement using their identity key and our base key.
	publicKey = parameters.TheirIdentityKey().PublicKey().PublicKey()
	privateKey = parameters.OurBaseKey().PrivateKey().Serialize()
	secret = kdf.CalculateSharedSecret(
		publicKey,
		privateKey,
	)
	masterSecret = append(masterSecret, secret[:]...)

	// Calculate the agreement using their signed prekey and our base key.
	publicKey = parameters.TheirSignedPreKey().PublicKey()
	privateKey = parameters.OurBaseKey().PrivateKey().Serialize()
	secret = kdf.CalculateSharedSecret(
		publicKey,
		privateKey,
	)
	masterSecret = append(masterSecret, secret[:]...)

	// If they have a one-time prekey, use it to calculate the shared secret with their
	// one time key and our base key.
	if parameters.TheirOneTimePreKey() != nil {
		publicKey = parameters.TheirOneTimePreKey().PublicKey()
		privateKey = parameters.OurBaseKey().PrivateKey().Serialize()
		secret = kdf.CalculateSharedSecret(
			publicKey,
			privateKey,
		)
		masterSecret = append(masterSecret, secret[:]...)

	}

	// Derive the root and chain keys based on the master secret.
	derivedKeysBytes, err := kdf.DeriveSecrets(masterSecret, nil, []byte("WhisperText"), root.DerivedSecretsSize)
	if err != nil {
		return nil, err
	}
	derivedKeys := session.NewDerivedSecrets(derivedKeysBytes)
	chainKey := chain.NewKey(kdf.DeriveSecrets, derivedKeys.ChainKey(), 0)
	rootKey := root.NewKey(kdf.DeriveSecrets, derivedKeys.RootKey())

	// Add the root and chain keys to a structure that will hold both keys.
	sessionKeys := session.NewKeyPair(rootKey, chainKey)

	return sessionKeys, nil
}

// CalculateReceiverSession calculates the key agreement for a sender. This should
// be used when we are receiving a message from someone for the first time.
func CalculateReceiverSession(parameters *ReceiverParameters) (*session.KeyPair, error) {
	var secret [32]byte
	var publicKey [32]byte
	var privateKey [32]byte
	masterSecret := []byte{} // Create a master shared secret that is 5 different 32-byte values

	discontinuity := genDiscontinuity()
	masterSecret = append(masterSecret, discontinuity[:]...)

	// Calculate the agreement using their identity key and our signed pre key.
	publicKey = parameters.TheirIdentityKey().PublicKey().PublicKey()
	privateKey = parameters.OurSignedPreKey().PrivateKey().Serialize()
	secret = kdf.CalculateSharedSecret(
		publicKey,
		privateKey,
	)
	masterSecret = append(masterSecret, secret[:]...)

	// Calculate the agreement using their base key and our identity key.
	publicKey = parameters.TheirBaseKey().PublicKey()
	privateKey = parameters.OurIdentityKeyPair().PrivateKey().Serialize()
	secret = kdf.CalculateSharedSecret(
		publicKey,
		privateKey,
	)
	masterSecret = append(masterSecret, secret[:]...)

	// Calculate the agreement using their base key and our signed prekey.
	publicKey = parameters.TheirBaseKey().PublicKey()
	privateKey = parameters.OurSignedPreKey().PrivateKey().Serialize()
	secret = kdf.CalculateSharedSecret(
		publicKey,
		privateKey,
	)
	masterSecret = append(masterSecret, secret[:]...)

	// If we had a one-time prekey, use it to calculate the shared secret with our
	// one time key and their base key.
	if parameters.OurOneTimePreKey() != nil {
		publicKey = parameters.TheirBaseKey().PublicKey()
		privateKey = parameters.OurOneTimePreKey().PrivateKey().Serialize()
		secret = kdf.CalculateSharedSecret(
			publicKey,
			privateKey,
		)
		masterSecret = append(masterSecret, secret[:]...)

	}

	// Derive the root and chain keys based on the master secret.
	derivedKeysBytes, err := kdf.DeriveSecrets(masterSecret, nil, []byte("WhisperText"), root.DerivedSecretsSize)
	if err != nil {
		return nil, err
	}
	derivedKeys := session.NewDerivedSecrets(derivedKeysBytes)
	chainKey := chain.NewKey(kdf.DeriveSecrets, derivedKeys.ChainKey(), 0)
	rootKey := root.NewKey(kdf.DeriveSecrets, derivedKeys.RootKey())

	// Add the root and chain keys to a structure that will hold both keys.
	sessionKeys := session.NewKeyPair(rootKey, chainKey)

	return sessionKeys, nil
}

// CalculateSymmetricSession calculates the key agreement between two users. This
// works by both clients exchanging KeyExchange messages to first establish a session.
// This is useful for establishing a session if both users are online.
func CalculateSymmetricSession(parameters *SymmetricParameters) (*session.KeyPair, error) {
	// Compare the base public keys so we can deterministically know whether we should
	// be setting up a sender or receiver session. If our key converted to an integer is
	// less than the other user's, act as a sender.
	if isSender(parameters.OurBaseKey.PublicKey(), parameters.TheirBaseKey) {
		senderParameters := &SenderParameters{
			ourBaseKey:         parameters.OurBaseKey,
			ourIdentityKeyPair: parameters.OurIdentityKeyPair,
			theirRatchetKey:    parameters.TheirRatchetKey,
			theirIdentityKey:   parameters.TheirIdentityKey,
			theirSignedPreKey:  parameters.TheirBaseKey,
		}

		return CalculateSenderSession(senderParameters)
	}

	// If our base public key was larger than the other user's, act as a receiver.
	receiverParameters := &ReceiverParameters{
		ourIdentityKeyPair: parameters.OurIdentityKeyPair,
		ourRatchetKey:      parameters.OurRatchetKey,
		ourSignedPreKey:    parameters.OurBaseKey,
		theirBaseKey:       parameters.TheirBaseKey,
		theirIdentityKey:   parameters.TheirIdentityKey,
	}

	return CalculateReceiverSession(receiverParameters)
}

// isSender is a private method for determining if a symmetric session should
// be calculated as the sender or receiver. It does so by converting the given
// keys into integers and comparing the size of those integers.
func isSender(ourKey, theirKey ecc.ECPublicKeyable) bool {
	ourKeyInt := binary.BigEndian.Uint32(ourKey.Serialize())
	theirKeyInt := binary.BigEndian.Uint32(theirKey.Serialize())

	return ourKeyInt < theirKeyInt
}
