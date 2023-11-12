package doubleratchet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// DefaultCrypto is an implementation of Crypto with cryptographic primitives recommended
// by the Double Ratchet Algorithm specification. However, some details are different,
// see function comments for details.
type DefaultCrypto struct{}

// GenerateDH creates a new Diffie-Hellman key pair.
func (c DefaultCrypto) GenerateDH() (DHPair, error) {
	var privKey [32]byte
	if _, err := io.ReadFull(rand.Reader, privKey[:]); err != nil {
		return dhPair{}, fmt.Errorf("couldn't generate privKey: %s", err)
	}
	privKey[0] &= 248
	privKey[31] &= 127
	privKey[31] |= 64

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &privKey)
	return dhPair{
		privateKey: privKey[:],
		publicKey:  pubKey[:],
	}, nil
}

// DH returns the output from the Diffie-Hellman calculation between
// the private key from the DH key pair dhPair and the DH public key dbPub.
func (c DefaultCrypto) DH(dhPair DHPair, dhPub Key) (Key, error) {
	var (
		dhOut   [32]byte
		privKey [32]byte
		pubKey  [32]byte
	)
	if len(dhPair.PrivateKey()) != 32 {
		return nil, fmt.Errorf("Invalid private key length: %d", len(dhPair.PrivateKey()))
	}

	if len(dhPub) != 32 {
		return nil, fmt.Errorf("Invalid private key length: %d", len(dhPair.PrivateKey()))
	}

	copy(privKey[:], dhPair.PrivateKey()[:32])
	copy(pubKey[:], dhPub[:32])

	curve25519.ScalarMult(&dhOut, &privKey, &pubKey)
	return dhOut[:], nil
}

// KdfRK returns a pair (32-byte root key, 32-byte chain key) as the output of applying
// a KDF keyed by a 32-byte root key rk to a Diffie-Hellman output dhOut.
func (c DefaultCrypto) KdfRK(rk, dhOut Key) (Key, Key, Key) {
	var (
		r   = hkdf.New(sha256.New, dhOut, rk, []byte("rsZUpEuXUqqwXBvSy3EcievAh4cMj6QL"))
		buf = make([]byte, 96)
	)

	// The only error here is an entropy limit which won't be reached for such a short buffer.
	_, _ = io.ReadFull(r, buf)

	rootKey := make(Key, 32)
	headerKey := make(Key, 32)
	chainKey := make(Key, 32)

	copy(rootKey[:], buf[:32])
	copy(chainKey[:], buf[32:64])
	copy(headerKey[:], buf[64:96])
	return rootKey, chainKey, headerKey
}

// KdfCK returns a pair (32-byte chain key, 32-byte message key) as the output of applying
// a KDF keyed by a 32-byte chain key ck to some constant.
func (c DefaultCrypto) KdfCK(ck Key) (Key, Key) {
	const (
		ckInput = 15
		mkInput = 16
	)

	chainKey := make(Key, 32)
	msgKey := make(Key, 32)

	h := hmac.New(sha256.New, ck[:])

	_, _ = h.Write([]byte{ckInput})
	copy(chainKey[:], h.Sum(nil))
	h.Reset()

	_, _ = h.Write([]byte{mkInput})
	copy(msgKey[:], h.Sum(nil))

	return chainKey, msgKey
}

// Encrypt uses a slightly different approach than in the algorithm specification:
// it uses AES-256-CTR instead of AES-256-CBC for security, ciphertext length and implementation
// complexity considerations.
func (c DefaultCrypto) Encrypt(mk Key, plaintext, ad []byte) ([]byte, error) {
	encKey, authKey, iv := c.deriveEncKeys(mk)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext, iv[:])

	var (
		block, _ = aes.NewCipher(encKey[:]) // No error will occur here as encKey is guaranteed to be 32 bytes.
		stream   = cipher.NewCTR(block, iv[:])
	)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return append(ciphertext, c.computeSignature(authKey[:], ciphertext, ad)...), nil
}

// Decrypt returns the AEAD decryption of ciphertext with message key mk.
func (c DefaultCrypto) Decrypt(mk Key, authCiphertext, ad []byte) ([]byte, error) {
	var (
		l          = len(authCiphertext)
		ciphertext = authCiphertext[:l-sha256.Size]
		signature  = authCiphertext[l-sha256.Size:]
	)

	// Check the signature.
	encKey, authKey, _ := c.deriveEncKeys(mk)

	if s := c.computeSignature(authKey[:], ciphertext, ad); !bytes.Equal(s, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decrypt.
	var (
		block, _  = aes.NewCipher(encKey[:]) // No error will occur here as encKey is guaranteed to be 32 bytes.
		stream    = cipher.NewCTR(block, ciphertext[:aes.BlockSize])
		plaintext = make([]byte, len(ciphertext[aes.BlockSize:]))
	)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	return plaintext, nil
}

// deriveEncKeys derive keys for message encryption and decryption. Returns (encKey, authKey, iv, err).
func (c DefaultCrypto) deriveEncKeys(mk Key) (Key, Key, [16]byte) {
	// First, derive encryption and authentication key out of mk.
	salt := make([]byte, 32)
	var (
		r   = hkdf.New(sha256.New, mk[:], salt, []byte("pcwSByyx2CRdryCffXJwy7xgVZWtW5Sh"))
		buf = make([]byte, 80)
	)

	// The only error here is an entropy limit which won't be reached for such a short buffer.
	_, _ = io.ReadFull(r, buf)

	var encKey Key = make(Key, 32)
	var authKey Key = make(Key, 32)
	var iv [16]byte

	copy(encKey[:], buf[0:32])
	copy(authKey[:], buf[32:64])
	copy(iv[:], buf[64:80])

	return encKey, authKey, iv
}

func (c DefaultCrypto) computeSignature(authKey, ciphertext, associatedData []byte) []byte {
	h := hmac.New(sha256.New, authKey)
	_, _ = h.Write(associatedData)
	_, _ = h.Write(ciphertext)
	return h.Sum(nil)
}

type dhPair struct {
	privateKey Key
	publicKey  Key
}

func (p dhPair) PrivateKey() Key {
	return p.privateKey
}

func (p dhPair) PublicKey() Key {
	return p.publicKey
}

func (p dhPair) String() string {
	return fmt.Sprintf("{privateKey: %s publicKey: %s}", p.privateKey, p.publicKey)
}
