package ratchet

import (
	"go.mau.fi/libsignal/kdf"
	"go.mau.fi/libsignal/util/bytehelper"
)

// KdfInfo is optional bytes to include in deriving secrets with KDF.
const KdfInfo string = "WhisperGroup"

// NewSenderMessageKey will return a new sender message key using the given
// iteration and seed.
func NewSenderMessageKey(iteration uint32, seed []byte) (*SenderMessageKey, error) {
	derivative, err := kdf.DeriveSecrets(seed, nil, []byte(KdfInfo), 48)
	if err != nil {
		return nil, err
	}

	// Split our derived secrets into 2 parts
	parts := bytehelper.Split(derivative, 16, 32)

	// Build the message key.
	senderKeyMessage := &SenderMessageKey{
		iteration: iteration,
		seed:      seed,
		iv:        parts[0],
		cipherKey: parts[1],
	}

	return senderKeyMessage, nil
}

// NewSenderMessageKeyFromStruct will return a new message key object from the
// given serializeable structure.
func NewSenderMessageKeyFromStruct(structure *SenderMessageKeyStructure) *SenderMessageKey {
	return &SenderMessageKey{
		iteration: structure.Iteration,
		iv:        structure.IV,
		cipherKey: structure.CipherKey,
		seed:      structure.Seed,
	}
}

// NewStructFromSenderMessageKey returns a serializeable structure of message keys.
func NewStructFromSenderMessageKey(key *SenderMessageKey) *SenderMessageKeyStructure {
	return &SenderMessageKeyStructure{
		CipherKey: key.cipherKey,
		Iteration: key.iteration,
		IV:        key.iv,
		Seed:      key.seed,
	}
}

// SenderMessageKeyStructure is a serializeable structure of SenderMessageKeys.
type SenderMessageKeyStructure struct {
	Iteration uint32
	IV        []byte
	CipherKey []byte
	Seed      []byte
}

// SenderMessageKey is a structure for sender message keys used in group
// messaging.
type SenderMessageKey struct {
	iteration uint32
	iv        []byte
	cipherKey []byte
	seed      []byte
}

// Iteration will return the sender message key's iteration.
func (k *SenderMessageKey) Iteration() uint32 {
	return k.iteration
}

// Iv will return the sender message key's initialization vector.
func (k *SenderMessageKey) Iv() []byte {
	return k.iv
}

// CipherKey will return the key in bytes.
func (k *SenderMessageKey) CipherKey() []byte {
	return k.cipherKey
}

// Seed will return the sender message key's seed.
func (k *SenderMessageKey) Seed() []byte {
	return k.seed
}
