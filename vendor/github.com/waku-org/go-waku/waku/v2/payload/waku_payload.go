package payload

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"

	"errors"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
)

// KeyKind indicates the type of encryption to apply
type KeyKind string

const (
	Symmetric  KeyKind = "Symmetric"
	Asymmetric KeyKind = "Asymmetric"
	None       KeyKind = "None"
)

const Unencrypted = 0
const V1Encryption = 1

// Payload contains the data of the message to encode
type Payload struct {
	Data    []byte   // Raw message payload
	Padding []byte   // Used to align data size, since data size alone might reveal important metainformation.
	Key     *KeyInfo // Contains the type of encryption to apply and the private key to use for signing the message
}

// DecodedPayload contains the data of the received message after decrypting it
type DecodedPayload struct {
	Data      []byte           // Decoded message payload
	Padding   []byte           // Used to align data size, since data size alone might reveal important metainformation.
	PubKey    *ecdsa.PublicKey // The public key that signed the payload
	Signature []byte
}

type KeyInfo struct {
	Kind    KeyKind           // Indicates the type of encryption to use
	SymKey  []byte            // If the encryption is Symmetric, a Symmetric key must be specified
	PubKey  ecdsa.PublicKey   // If the encryption is Asymmetric, the public key of the message receptor must be specified
	PrivKey *ecdsa.PrivateKey // Set a privkey if the message requires a signature
}

// Encode encodes a payload depending on the version parameter.
// 0 for raw unencrypted data, and 1 for using WakuV1 encoding.
func (payload Payload) Encode(version uint32) ([]byte, error) {
	switch version {
	case 0:
		return payload.Data, nil
	case 1:
		data, err := payload.v1Data()
		if err != nil {
			return nil, err
		}

		if payload.Key.PrivKey != nil {
			data, err = sign(data, *payload.Key.PrivKey)
			if err != nil {
				return nil, err
			}
		}

		switch payload.Key.Kind {
		case Symmetric:
			encoded, err := encryptSymmetric(data, payload.Key.SymKey)
			if err != nil {
				return nil, fmt.Errorf("couldn't encrypt using symmetric key: %w", err)
			}

			return encoded, nil
		case Asymmetric:
			encoded, err := encryptAsymmetric(data, &payload.Key.PubKey)
			if err != nil {
				return nil, fmt.Errorf("couldn't encrypt using asymmetric key: %w", err)
			}
			return encoded, nil
		case None:
			return nil, errors.New("non supported KeyKind")
		}
	}
	return nil, errors.New("unsupported wakumessage version")
}

func EncodeWakuMessage(message *pb.WakuMessage, keyInfo *KeyInfo) error {
	msgPayload := message.Payload
	payload := Payload{
		Data: msgPayload,
		Key:  keyInfo,
	}

	encodedBytes, err := payload.Encode(message.GetVersion())
	if err != nil {
		return err
	}

	message.Payload = encodedBytes
	return nil
}

// DecodePayload decodes a WakuMessage depending on the version parameter.
// 0 for raw unencrypted data, and 1 for using WakuV1 decoding
func DecodePayload(message *pb.WakuMessage, keyInfo *KeyInfo) (*DecodedPayload, error) {
	switch message.GetVersion() {
	case uint32(0):
		return &DecodedPayload{Data: message.Payload}, nil
	case uint32(1):
		switch keyInfo.Kind {
		case Symmetric:
			if keyInfo.SymKey == nil {
				return nil, errors.New("symmetric key is required")
			}

			decodedData, err := decryptSymmetric(message.Payload, keyInfo.SymKey)
			if err != nil {
				return nil, fmt.Errorf("couldn't decrypt using symmetric key: %w", err)
			}

			decodedPayload, err := validateAndParse(decodedData)
			if err != nil {
				return nil, err
			}

			return decodedPayload, nil
		case Asymmetric:
			if keyInfo.PrivKey == nil {
				return nil, errors.New("private key is required")
			}

			decodedData, err := decryptAsymmetric(message.Payload, keyInfo.PrivKey)
			if err != nil {
				return nil, fmt.Errorf("couldn't decrypt using asymmetric key: %w", err)
			}

			decodedPayload, err := validateAndParse(decodedData)
			if err != nil {
				return nil, err
			}

			return decodedPayload, nil
		case None:
			return nil, errors.New("non supported KeyKind")
		}
	}
	return nil, errors.New("unsupported wakumessage version")
}

func DecodeWakuMessage(message *pb.WakuMessage, keyInfo *KeyInfo) error {
	decodedPayload, err := DecodePayload(message, keyInfo)
	if err != nil {
		return err
	}

	message.Payload = decodedPayload.Data
	return nil
}

const aesNonceLength = 12
const aesKeyLength = 32
const signatureFlag = byte(4)
const flagsLength = 1
const padSizeLimit = 256 // just an arbitrary number, could be changed without breaking the protocol
const signatureLength = 65
const sizeMask = byte(3)

// Decrypts a message with a topic key, using AES-GCM-256.
// nonce size should be 12 bytes (see cipher.gcmStandardNonceSize).
func decryptSymmetric(payload []byte, key []byte) ([]byte, error) {
	// symmetric messages are expected to contain the 12-byte nonce at the end of the payload
	if len(payload) < aesNonceLength {
		return nil, errors.New("missing salt or invalid payload in symmetric message")
	}

	salt := payload[len(payload)-aesNonceLength:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	decrypted, err := aesgcm.Open(nil, salt, payload[:len(payload)-aesNonceLength], nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// Decrypts an encrypted payload with a private key.
func decryptAsymmetric(payload []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	decrypted, err := ecies.ImportECDSA(key).Decrypt(payload, nil, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, err
}

// ValidatePublicKey checks the format of the given public key.
func validatePublicKey(k *ecdsa.PublicKey) bool {
	return k != nil && k.X != nil && k.Y != nil && k.X.Sign() != 0 && k.Y.Sign() != 0
}

// Encrypts and returns with a public key.
func encryptAsymmetric(rawPayload []byte, key *ecdsa.PublicKey) ([]byte, error) {
	if !validatePublicKey(key) {
		return nil, errors.New("invalid public key provided for asymmetric encryption")
	}

	encrypted, err := ecies.Encrypt(crand.Reader, ecies.ImportECDSAPublic(key), rawPayload, nil, nil)
	if err == nil {
		return encrypted, nil
	}
	return nil, err
}

// Encrypts a payload with a topic key, using AES-GCM-256.
// nonce size should be 12 bytes (see cipher.gcmStandardNonceSize).
func encryptSymmetric(rawPayload []byte, key []byte) ([]byte, error) {
	if !validateDataIntegrity(key, aesKeyLength) {
		return nil, errors.New("invalid key provided for symmetric encryption, size: " + strconv.Itoa(len(key)))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	salt, err := generateSecureRandomData(aesNonceLength) // never use more than 2^32 random nonces with a given key
	if err != nil {
		return nil, err
	}
	encrypted := aesgcm.Seal(nil, salt, rawPayload, nil)
	return append(encrypted, salt...), nil
}

// validateDataIntegrity returns false if the data have the wrong or contains all zeros,
// which is the simplest and the most common bug.
func validateDataIntegrity(k []byte, expectedSize int) bool {
	if len(k) != expectedSize {
		return false
	}
	if expectedSize > 3 && containsOnlyZeros(k) {
		return false
	}
	return true
}

// containsOnlyZeros checks if the data contain only zeros.
func containsOnlyZeros(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// generateSecureRandomData generates random data where extra security is required.
// The purpose of this function is to prevent some bugs in software or in hardware
// from delivering not-very-random data. This is especially useful for AES nonce,
// where true randomness does not really matter, but it is very important to have
// a unique nonce for every message.
func generateSecureRandomData(length int) ([]byte, error) {
	x := make([]byte, length)
	y := make([]byte, length)
	res := make([]byte, length)

	_, err := crand.Read(x)
	if err != nil {
		return nil, err
	} else if !validateDataIntegrity(x, length) {
		return nil, errors.New("crypto/rand failed to generate secure random data")
	}
	_, err = crand.Read(y)
	if err != nil {
		return nil, err
	} else if !validateDataIntegrity(y, length) {
		return nil, errors.New("math/rand failed to generate secure random data")
	}
	for i := 0; i < length; i++ {
		res[i] = x[i] ^ y[i]
	}
	if !validateDataIntegrity(res, length) {
		return nil, errors.New("failed to generate secure random data")
	}
	return res, nil
}

func isMessageSigned(flags byte) bool {
	return (flags & signatureFlag) != 0
}

// sign calculates the cryptographic signature for the message,
// also setting the sign flag.
func sign(data []byte, privKey ecdsa.PrivateKey) ([]byte, error) {
	result := make([]byte, len(data))
	copy(result, data)

	if isMessageSigned(result[0]) {
		// this should not happen, but no reason to panic
		return result, nil
	}

	result[0] |= signatureFlag // it is important to set this flag before signing
	hash := crypto.Keccak256(result)
	signature, err := crypto.Sign(hash, &privKey)

	if err != nil {
		result[0] &= (0xFF ^ signatureFlag) // clear the flag
		return nil, err
	}
	result = append(result, signature...)

	return result, nil
}

func (payload Payload) v1Data() ([]byte, error) {
	const payloadSizeFieldMaxSize = 4
	result := make([]byte, 1, flagsLength+payloadSizeFieldMaxSize+len(payload.Data)+len(payload.Padding)+signatureLength+padSizeLimit)
	result[0] = 0 // set all the flags to zero
	result = payload.addPayloadSizeField(result)
	result = append(result, payload.Data...)
	result, err := payload.appendPadding(result)
	return result, err
}

// addPayloadSizeField appends the auxiliary field containing the size of payload
func (payload Payload) addPayloadSizeField(input []byte) []byte {
	fieldSize := getSizeOfPayloadSizeField(payload.Data)
	field := make([]byte, 4)
	binary.LittleEndian.PutUint32(field, uint32(len(payload.Data)))
	field = field[:fieldSize]
	result := append(input, field...)
	result[0] |= byte(fieldSize)
	return result
}

// getSizeOfPayloadSizeField returns the number of bytes necessary to encode the size of payload
func getSizeOfPayloadSizeField(payload []byte) int {
	s := 1
	for i := len(payload); i >= 256; i /= 256 {
		s++
	}
	return s
}

// appendPadding appends the padding specified in params.
// If no padding is provided in params, then random padding is generated.
func (payload Payload) appendPadding(input []byte) ([]byte, error) {
	if len(payload.Padding) != 0 {
		// padding data was provided by the Dapp, just use it as is
		result := append(input, payload.Padding...)
		return result, nil
	}

	rawSize := flagsLength + getSizeOfPayloadSizeField(payload.Data) + len(payload.Data)
	if payload.Key.PrivKey != nil {
		rawSize += signatureLength
	}
	odd := rawSize % padSizeLimit
	paddingSize := padSizeLimit - odd
	pad := make([]byte, paddingSize)
	_, err := crand.Read(pad)
	if err != nil {
		return nil, err
	}
	if !validateDataIntegrity(pad, paddingSize) {
		return nil, errors.New("failed to generate random padding of size " + strconv.Itoa(paddingSize))
	}
	result := append(input, pad...)
	return result, nil
}

func validateAndParse(input []byte) (*DecodedPayload, error) {
	end := len(input)
	if end < 1 {
		return nil, errors.New("invalid message length")
	}

	msg := new(DecodedPayload)

	if isMessageSigned(input[0]) {
		end -= signatureLength
		if end <= 1 {
			return nil, errors.New("invalid message length")
		}
		msg.Signature = input[end : end+signatureLength]

		var err error
		msg.PubKey, err = msg.sigToPubKey(input)
		if err != nil {
			return nil, err
		}
	}

	beg := 1
	payloadSize := 0
	sizeOfPayloadSizeField := int(input[0] & sizeMask) // number of bytes indicating the size of payload

	if sizeOfPayloadSizeField != 0 {
		if end < beg+sizeOfPayloadSizeField {
			return nil, errors.New("invalid message length")
		}
		payloadSize = int(bytesToUintLittleEndian(input[beg : beg+sizeOfPayloadSizeField]))
		beg += sizeOfPayloadSizeField
		if beg+payloadSize > end {
			return nil, errors.New("invalid message length")
		}
		msg.Data = input[beg : beg+payloadSize]
	}

	beg += payloadSize
	msg.Padding = input[beg:end]

	return msg, nil
}

// SigToPubKey returns the public key associated to the message's
// signature.
func (p *DecodedPayload) sigToPubKey(input []byte) (*ecdsa.PublicKey, error) {
	defer func() { _ = recover() }() // in case of invalid signature
	hash := crypto.Keccak256(input[0 : len(input)-signatureLength])
	pub, err := crypto.SigToPub(hash, p.Signature)
	if err != nil {
		return nil, err
	}

	return pub, nil
}

// bytesToUintLittleEndian converts the slice to 64-bit unsigned integer.
func bytesToUintLittleEndian(b []byte) (res uint64) {
	mul := uint64(1)
	for i := 0; i < len(b); i++ {
		res += uint64(b[i]) * mul
		mul *= 256
	}
	return res
}
