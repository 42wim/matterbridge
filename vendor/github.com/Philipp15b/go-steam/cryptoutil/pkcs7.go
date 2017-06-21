package cryptoutil

import (
	"crypto/aes"
)

// Returns a new byte array padded with PKCS7 and prepended
// with empty space of the AES block size (16 bytes) for the IV.
func padPKCS7WithIV(src []byte) []byte {
	missing := aes.BlockSize - (len(src) % aes.BlockSize)
	newSize := len(src) + aes.BlockSize + missing
	dest := make([]byte, newSize, newSize)
	copy(dest[aes.BlockSize:], src)

	padding := byte(missing)
	for i := newSize - missing; i < newSize; i++ {
		dest[i] = padding
	}
	return dest
}

func unpadPKCS7(src []byte) []byte {
	padLen := src[len(src)-1]
	return src[:len(src)-int(padLen)]
}
