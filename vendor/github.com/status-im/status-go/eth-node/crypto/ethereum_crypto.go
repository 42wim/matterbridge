package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"

	dr "github.com/status-im/doubleratchet"
	"golang.org/x/crypto/hkdf"

	"github.com/status-im/status-go/eth-node/crypto/ecies"
)

// EthereumCrypto is an implementation of Crypto with cryptographic primitives recommended
// by the Double Ratchet Algorithm specification. However, some details are different,
// see function comments for details.
type EthereumCrypto struct{}

// See the Crypto interface.
func (c EthereumCrypto) GenerateDH() (dr.DHPair, error) {
	keys, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	return DHPair{
		PubKey: CompressPubkey(&keys.PublicKey),
		PrvKey: FromECDSA(keys),
	}, nil

}

// See the Crypto interface.
func (c EthereumCrypto) DH(dhPair dr.DHPair, dhPub dr.Key) (dr.Key, error) {
	tmpKey := dhPair.PrivateKey()
	privateKey, err := ToECDSA(tmpKey)
	if err != nil {
		return nil, err
	}

	eciesPrivate := ecies.ImportECDSA(privateKey)

	publicKey, err := DecompressPubkey(dhPub)
	if err != nil {
		return nil, err
	}
	eciesPublic := ecies.ImportECDSAPublic(publicKey)

	key, err := eciesPrivate.GenerateShared(
		eciesPublic,
		16,
		16,
	)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// See the Crypto interface.
func (c EthereumCrypto) KdfRK(rk, dhOut dr.Key) (dr.Key, dr.Key, dr.Key) {
	var (
		// We can use a non-secret constant as the last argument
		r   = hkdf.New(sha256.New, dhOut, rk, []byte("rsZUpEuXUqqwXBvSy3EcievAh4cMj6QL"))
		buf = make([]byte, 96)
	)

	rootKey := make(dr.Key, 32)
	chainKey := make(dr.Key, 32)
	headerKey := make(dr.Key, 32)

	// The only error here is an entropy limit which won't be reached for such a short buffer.
	_, _ = io.ReadFull(r, buf)

	copy(rootKey, buf[:32])
	copy(chainKey, buf[32:64])
	copy(headerKey, buf[64:96])
	return rootKey, chainKey, headerKey
}

// See the Crypto interface.
func (c EthereumCrypto) KdfCK(ck dr.Key) (dr.Key, dr.Key) {
	const (
		ckInput = 15
		mkInput = 16
	)

	chainKey := make(dr.Key, 32)
	msgKey := make(dr.Key, 32)

	h := hmac.New(sha256.New, ck)

	_, _ = h.Write([]byte{ckInput})
	copy(chainKey, h.Sum(nil))
	h.Reset()

	_, _ = h.Write([]byte{mkInput})
	copy(msgKey, h.Sum(nil))

	return chainKey, msgKey
}

// Encrypt uses a slightly different approach than in the algorithm specification:
// it uses AES-256-CTR instead of AES-256-CBC for security, ciphertext length and implementation
// complexity considerations.
func (c EthereumCrypto) Encrypt(mk dr.Key, plaintext, ad []byte) ([]byte, error) {
	encKey, authKey, iv := c.deriveEncKeys(mk)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext, iv[:])

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv[:])
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return append(ciphertext, c.computeSignature(authKey, ciphertext, ad)...), nil
}

// See the Crypto interface.
func (c EthereumCrypto) Decrypt(mk dr.Key, authCiphertext, ad []byte) ([]byte, error) {
	var (
		l          = len(authCiphertext)
		ciphertext = authCiphertext[:l-sha256.Size]
		signature  = authCiphertext[l-sha256.Size:]
	)

	// Check the signature.
	encKey, authKey, _ := c.deriveEncKeys(mk)

	if s := c.computeSignature(authKey, ciphertext, ad); !bytes.Equal(s, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decrypt.
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, ciphertext[:aes.BlockSize])
	plaintext := make([]byte, len(ciphertext[aes.BlockSize:]))

	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	return plaintext, nil
}

// deriveEncKeys derive keys for message encryption and decryption. Returns (encKey, authKey, iv, err).
func (c EthereumCrypto) deriveEncKeys(mk dr.Key) (dr.Key, dr.Key, [16]byte) {
	// First, derive encryption and authentication key out of mk.
	salt := make([]byte, 32)
	var (
		r   = hkdf.New(sha256.New, mk, salt, []byte("pcwSByyx2CRdryCffXJwy7xgVZWtW5Sh"))
		buf = make([]byte, 80)
	)

	encKey := make(dr.Key, 32)
	authKey := make(dr.Key, 32)
	var iv [16]byte

	// The only error here is an entropy limit which won't be reached for such a short buffer.
	_, _ = io.ReadFull(r, buf)

	copy(encKey, buf[0:32])
	copy(authKey, buf[32:64])
	copy(iv[:], buf[64:80])
	return encKey, authKey, iv
}

func (c EthereumCrypto) computeSignature(authKey, ciphertext, associatedData []byte) []byte {
	h := hmac.New(sha256.New, authKey)
	_, _ = h.Write(associatedData)
	_, _ = h.Write(ciphertext)
	return h.Sum(nil)
}

type DHPair struct {
	PrvKey dr.Key
	PubKey dr.Key
}

func (p DHPair) PrivateKey() dr.Key {
	return p.PrvKey
}

func (p DHPair) PublicKey() dr.Key {
	return p.PubKey
}
