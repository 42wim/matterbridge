package srtp

import "github.com/pion/rtp"

// cipher represents a implementation of one
// of the SRTP Specific ciphers
type srtpCipher interface {
	// authTagLen returns auth key length of the cipher.
	// See the note below.
	authTagLen() int
	// aeadAuthTagLen returns AEAD auth key length of the cipher.
	// See the note below.
	aeadAuthTagLen() int
	getRTCPIndex([]byte) uint32

	encryptRTP([]byte, *rtp.Header, []byte, uint32) ([]byte, error)
	encryptRTCP([]byte, []byte, uint32, uint32) ([]byte, error)

	decryptRTP([]byte, []byte, *rtp.Header, int, uint32) ([]byte, error)
	decryptRTCP([]byte, []byte, uint32, uint32) ([]byte, error)
}

/*
NOTE: Auth tag and AEAD auth tag are placed at the different position in SRTCP

In non-AEAD cipher, the authentication tag is placed *after* the ESRTCP word
(Encrypted-flag and SRTCP index).

> AES_128_CM_HMAC_SHA1_80
> | RTCP Header | Encrypted payload |E| SRTCP Index | Auth tag |
>                                   ^               |----------|
>                                   |                ^
>                                   |                authTagLen=10
>                                   aeadAuthTagLen=0

In AEAD cipher, the AEAD authentication tag is embedded in the ciphertext.
It is *before* the ESRTCP word (Encrypted-flag and SRTCP index).

> AEAD_AES_128_GCM
> | RTCP Header | Encrypted payload | AEAD auth tag |E| SRTCP Index |
>                                   |---------------|               ^
>                                    ^                              authTagLen=0
>                                    aeadAuthTagLen=16

See https://tools.ietf.org/html/rfc7714 for the full specifications.
*/
