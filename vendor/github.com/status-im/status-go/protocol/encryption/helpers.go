package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/crypto/ecies"
)

const keyBumpValue = uint64(10)

// GetCurrentTime64 returns the current unix time in milliseconds
func GetCurrentTime() uint64 {
	return (uint64)(time.Now().UnixNano() / int64(time.Millisecond))
}

// bumpKeyID takes a timestampID and returns its value incremented by the keyBumpValue
func bumpKeyID(timestampID uint64) uint64 {
	return timestampID + keyBumpValue
}

func generateHashRatchetKeyID(groupID []byte, timestamp uint64, keyBytes []byte) []byte {
	var keyMaterial []byte

	keyMaterial = append(keyMaterial, groupID...)

	timestampBytes := make([]byte, 8) // 8 bytes for a uint64
	binary.LittleEndian.PutUint64(timestampBytes, timestamp)
	keyMaterial = append(keyMaterial, timestampBytes...)

	keyMaterial = append(keyMaterial, keyBytes...)

	return crypto.Keccak256(keyMaterial)
}

func publicKeyMostRelevantBytes(key *ecdsa.PublicKey) uint32 {

	keyBytes := crypto.FromECDSAPub(key)

	return binary.LittleEndian.Uint32(keyBytes[1:5])
}

func encrypt(plaintext []byte, key []byte, reader io.Reader) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func generateSharedKey(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) ([]byte, error) {

	const encryptedPayloadKeyLength = 16

	return ecies.ImportECDSA(privateKey).GenerateShared(
		ecies.ImportECDSAPublic(publicKey),
		encryptedPayloadKeyLength,
		encryptedPayloadKeyLength,
	)
}

func buildGroupRekeyMessage(privateKey *ecdsa.PrivateKey, groupID []byte, timestamp uint64, keyMaterial []byte, keys []*ecdsa.PublicKey) (*RekeyGroup, error) {

	message := &RekeyGroup{
		Timestamp: timestamp,
	}

	message.Keys = make(map[uint32][]byte)

	for _, k := range keys {

		sharedKey, err := generateSharedKey(privateKey, k)
		if err != nil {
			return nil, err
		}

		encryptedKey, err := encrypt(keyMaterial, sharedKey, rand.Reader)
		if err != nil {
			return nil, err
		}

		kBytes := publicKeyMostRelevantBytes(k)

		if message.Keys[kBytes] == nil {
			message.Keys[kBytes] = encryptedKey
		} else {
			message.Keys[kBytes] = append(message.Keys[kBytes], encryptedKey...)
		}
	}

	return message, nil
}

const nonceLength = 12

func decrypt(cyphertext []byte, key []byte) ([]byte, error) {
	if len(cyphertext) < nonceLength {
		return nil, errors.New("invalid cyphertext length")
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := cyphertext[:nonceLength]
	return gcm.Open(nil, nonce, cyphertext[nonceLength:], nil)
}

const keySize = 60

func decryptGroupRekeyMessage(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, message *RekeyGroup) ([]byte, error) {
	kBytes := publicKeyMostRelevantBytes(&privateKey.PublicKey)
	if message.Keys == nil || message.Keys[kBytes] == nil {
		return nil, nil
	}

	sharedKey, err := generateSharedKey(privateKey, publicKey)
	if err != nil {
		return nil, err
	}

	keys := message.Keys[kBytes]

	nKeys := len(keys) / keySize

	var decryptedKey []byte
	for i := 0; i < nKeys; i++ {

		encryptedKey := keys[i*keySize : i*keySize+keySize]
		decryptedKey, err = decrypt(encryptedKey, sharedKey)
		if err != nil {
			continue
		} else {
			break
		}

	}

	return decryptedKey, nil
}
