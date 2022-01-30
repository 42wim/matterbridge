// Package chain provides chain keys used in double ratchet sessions.
package chain

import (
	"crypto/hmac"
	"crypto/sha256"
	"go.mau.fi/libsignal/kdf"
	"go.mau.fi/libsignal/keys/message"
)

var messageKeySeed = []byte{0x01}
var chainKeySeed = []byte{0x02}

// NewKey returns a new chain key with the given kdf, key, and index
func NewKey(kdf kdf.HKDF, key []byte, index uint32) *Key {
	chainKey := Key{
		kdf:   kdf,
		key:   key,
		index: index,
	}

	return &chainKey
}

// NewKeyFromStruct will return a chain key built from the given structure.
func NewKeyFromStruct(structure *KeyStructure, kdf kdf.HKDF) *Key {
	return NewKey(
		kdf,
		structure.Key,
		structure.Index,
	)
}

// NewStructFromKey will return a chain key structure for serialization.
func NewStructFromKey(key *Key) *KeyStructure {
	return &KeyStructure{
		Key:   key.key,
		Index: key.index,
	}
}

// KeyStructure is a serializeable structure for chain keys.
type KeyStructure struct {
	Key   []byte
	Index uint32
}

// Key is used for generating message keys. This key "ratchets" every time a
// message key is generated. Every time the chain key ratchets forward, its index
// increases by one.
type Key struct {
	kdf   kdf.HKDF
	key   []byte
	index uint32 // Index's maximum size: 4,294,967,295
}

// Current returns the current ChainKey struct.
func (c *Key) Current() *Key {
	return c
}

// Key returns the ChainKey's key material.
func (c *Key) Key() []byte {
	return c.key
}

// SetKey will set the ChainKey's key material.
func (c *Key) SetKey(key []byte) {
	c.key = key
}

// Index returns how many times the ChainKey has been "ratcheted" forward.
func (c *Key) Index() uint32 {
	return c.index
}

// SetIndex sets how many times the ChainKey has been "ratcheted" forward.
func (c *Key) SetIndex(index uint32) {
	c.index = index
}

// NextKey uses the key derivation function to generate a new ChainKey.
func (c *Key) NextKey() *Key {
	nextKey := c.BaseMaterial(chainKeySeed)
	return NewKey(c.kdf, nextKey, c.index+1)
}

// MessageKeys returns message keys, which includes the cipherkey, mac, iv, and index.
func (c *Key) MessageKeys() *message.Keys {
	inputKeyMaterial := c.BaseMaterial(messageKeySeed)
	keyMaterialBytes, _ := c.kdf(inputKeyMaterial, nil, []byte(message.KdfSalt), message.DerivedSecretsSize)
	keyMaterial := newKeyMaterial(keyMaterialBytes)

	// Use the key material returned from the key derivation function for our cipherkey, mac, and iv.
	messageKeys := message.NewKeys(
		keyMaterial.CipherKey, // Use the first 32 bytes of the key material for the CipherKey
		keyMaterial.MacKey,    // Use bytes 32-64 of the key material for the MacKey
		keyMaterial.IV,        // Use the last 16 bytes for the IV.
		c.Index(),             // Attach the chain key's index to the message keys.
	)

	return messageKeys
}

// BaseMaterial uses hmac to derive the base material used in the key derivation function for a new key.
func (c *Key) BaseMaterial(seed []byte) []byte {
	mac := hmac.New(sha256.New, c.key[:])
	mac.Write(seed)

	return mac.Sum(nil)
}

// NewKeyMaterial takes an 80-byte slice derived from a key derivation function and splits
// it into the cipherkey, mac, and iv.
func newKeyMaterial(keyMaterialBytes []byte) *kdf.KeyMaterial {
	cipherKey := keyMaterialBytes[:32] // Use the first 32 bytes of the key material for the CipherKey
	macKey := keyMaterialBytes[32:64]  // Use bytes 32-64 of the key material for the MacKey
	iv := keyMaterialBytes[64:80]      // Use the last 16 bytes for the IV.

	keyMaterial := kdf.KeyMaterial{
		CipherKey: cipherKey,
		MacKey:    macKey,
		IV:        iv,
	}

	return &keyMaterial
}
