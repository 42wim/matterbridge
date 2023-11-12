package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"io"
	"math/big"

	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/crypto/ecies"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
)

const (
	nonceLength                = 12
	defaultECHDSharedKeyLength = 16
	defaultECHDMACLength       = 16
)

var (
	ErrInvalidCiphertextLength = errors.New("invalid cyphertext length")

	letterRunes       = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numberRunes       = []rune("0123456789")
	alphanumericRunes = append(numberRunes, letterRunes...)
)

func HashPublicKey(pk *ecdsa.PublicKey) []byte {
	return Shake256(crypto.CompressPubkey(pk))
}

func Decrypt(cyphertext []byte, key []byte) ([]byte, error) {
	if len(cyphertext) < nonceLength {
		return nil, ErrInvalidCiphertextLength
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

func Encrypt(plaintext []byte, key []byte, reader io.Reader) ([]byte, error) {
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

func Shake256(buf []byte) []byte {
	h := make([]byte, 64)
	sha3.ShakeSum256(h, buf)
	return h
}

// IsPubKeyEqual checks that two public keys are equal
func IsPubKeyEqual(a, b *ecdsa.PublicKey) bool {
	// the curve is always the same, just compare the points
	return a.X.Cmp(b.X) == 0 && a.Y.Cmp(b.Y) == 0
}

func PubkeyToHex(key *ecdsa.PublicKey) string {
	return types.EncodeHex(crypto.FromECDSAPub(key))
}

func PubkeyToHexBytes(key *ecdsa.PublicKey) types.HexBytes {
	return crypto.FromECDSAPub(key)
}

func HexToPubkey(pk string) (*ecdsa.PublicKey, error) {
	bytes, err := types.DecodeHex(pk)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPubkey(bytes)
}

func MakeECDHSharedKey(yourPrivateKey *ecdsa.PrivateKey, theirPubKey *ecdsa.PublicKey) ([]byte, error) {
	return ecies.ImportECDSA(yourPrivateKey).GenerateShared(
		ecies.ImportECDSAPublic(theirPubKey),
		defaultECHDSharedKeyLength,
		defaultECHDMACLength,
	)
}

func randomString(choice []rune, n int) (string, error) {
	max := big.NewInt(int64(len(choice)))
	rr := rand.Reader

	b := make([]rune, n)
	for i := range b {
		pos, err := rand.Int(rr, max)
		if err != nil {
			return "", err
		}
		b[i] = choice[pos.Int64()]
	}
	return string(b), nil
}

func RandomAlphabeticalString(n int) (string, error) {
	return randomString(letterRunes, n)
}

func RandomAlphanumericString(n int) (string, error) {
	return randomString(alphanumericRunes, n)
}
