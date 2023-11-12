package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/sha3"

	types "github.com/status-im/status-go/eth-node/types"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	aesNonceLength = 12
)

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given digest cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(digestHash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	return gethcrypto.Sign(digestHash, prv)
}

// SignBytes signs the hash of arbitrary data.
func SignBytes(data []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	return Sign(Keccak256(data), prv)
}

// SignBytesAsHex signs the Keccak256 hash of arbitrary data and returns its hex representation.
func SignBytesAsHex(data []byte, identity *ecdsa.PrivateKey) (string, error) {
	signature, err := SignBytes(data, identity)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil
}

// SignStringAsHex signs the Keccak256 hash of arbitrary string and returns its hex representation.
func SignStringAsHex(data string, identity *ecdsa.PrivateKey) (string, error) {
	return SignBytesAsHex([]byte(data), identity)
}

// VerifySignatures verifies tuples of signatures content/hash/public key
func VerifySignatures(signaturePairs [][3]string) error {
	for _, signaturePair := range signaturePairs {
		content := Keccak256([]byte(signaturePair[0]))

		signature, err := hex.DecodeString(signaturePair[1])
		if err != nil {
			return err
		}

		publicKeyBytes, err := hex.DecodeString(signaturePair[2])
		if err != nil {
			return err
		}

		publicKey, err := UnmarshalPubkey(publicKeyBytes)
		if err != nil {
			return err
		}

		recoveredKey, err := SigToPub(
			content,
			signature,
		)
		if err != nil {
			return err
		}

		if PubkeyToAddress(*recoveredKey) != PubkeyToAddress(*publicKey) {
			return errors.New("identity key and signature mismatch")
		}
	}

	return nil
}

// ExtractSignatures extract from tuples of signatures content a public key
// DEPRECATED: use ExtractSignature
func ExtractSignatures(signaturePairs [][2]string) ([]string, error) {
	response := make([]string, len(signaturePairs))
	for i, signaturePair := range signaturePairs {
		content := Keccak256([]byte(signaturePair[0]))

		signature, err := hex.DecodeString(signaturePair[1])
		if err != nil {
			return nil, err
		}

		recoveredKey, err := SigToPub(
			content,
			signature,
		)
		if err != nil {
			return nil, err
		}

		response[i] = fmt.Sprintf("%x", FromECDSAPub(recoveredKey))
	}

	return response, nil
}

// ExtractSignature returns a public key for a given data and signature.
func ExtractSignature(data, signature []byte) (*ecdsa.PublicKey, error) {
	dataHash := Keccak256(data)
	return SigToPub(dataHash, signature)
}

func EncryptSymmetric(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	salt, err := generateSecureRandomData(aesNonceLength)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	encrypted := aesgcm.Seal(nil, salt, plaintext, nil)
	return append(encrypted, salt...), nil
}

func DecryptSymmetric(key []byte, cyphertext []byte) ([]byte, error) {
	// symmetric messages are expected to contain the 12-byte nonce at the end of the payload
	if len(cyphertext) < aesNonceLength {
		return nil, errors.New("missing salt or invalid payload in symmetric message")
	}
	salt := cyphertext[len(cyphertext)-aesNonceLength:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	decrypted, err := aesgcm.Open(nil, salt, cyphertext[:len(cyphertext)-aesNonceLength], nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func containsOnlyZeros(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

func validateDataIntegrity(k []byte, expectedSize int) bool {
	if len(k) != expectedSize {
		return false
	}
	if containsOnlyZeros(k) {
		return false
	}
	return true
}

func generateSecureRandomData(length int) ([]byte, error) {
	res := make([]byte, length)

	_, err := rand.Read(res)
	if err != nil {
		return nil, err
	}

	if !validateDataIntegrity(res, length) {
		return nil, errors.New("crypto/rand failed to generate secure random data")
	}

	return res, nil
}

// TextHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calulcated as
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func TextHash(data []byte) []byte {
	hash, _ := TextAndHash(data)
	return hash
}

// TextAndHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calulcated as
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func TextAndHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

func EcRecover(ctx context.Context, data types.HexBytes, sig types.HexBytes) (types.Address, error) {
	// Returns the address for the Account that was used to create the signature.
	//
	// Note, this function is compatible with eth_sign and personal_sign. As such it recovers
	// the address of:
	// hash = keccak256("\x19${byteVersion}Ethereum Signed Message:\n${message length}${message}")
	// addr = ecrecover(hash, signature)
	//
	// Note, the signature must conform to the secp256k1 curve R, S and V values, where
	// the V value must be be 27 or 28 for legacy reasons.
	//
	// https://github.com/ethereum/go-ethereum/wiki/Management-APIs#personal_ecRecover
	if len(sig) != 65 {
		return types.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return types.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1
	hash := TextHash(data)
	rpk, err := SigToPub(hash, sig)
	if err != nil {
		return types.Address{}, err
	}
	return PubkeyToAddress(*rpk), nil
}
