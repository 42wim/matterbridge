package cryptoutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"errors"
)

// Parses a DER encoded RSA public key
func ParseASN1RSAPublicKey(derBytes []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, err
	}
	pubKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return pubKey, nil
}

// Encrypts a message with the given public key using RSA-OAEP and the sha1 hash function.
func RSAEncrypt(pub *rsa.PublicKey, msg []byte) []byte {
	b, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pub, msg, nil)
	if err != nil {
		panic(err)
	}
	return b
}
