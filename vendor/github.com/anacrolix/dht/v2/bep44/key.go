package bep44

import (
	"crypto/ed25519"
)

func Sign(k ed25519.PrivateKey, salt []byte, seq int64, bv []byte) []byte {
	return ed25519.Sign(k, bufferToSign(salt, bv, seq))
}

func Verify(k ed25519.PublicKey, salt []byte, seq int64, bv []byte, sig []byte) bool {
	return ed25519.Verify(k, bufferToSign(salt, bv, seq), sig)
}
