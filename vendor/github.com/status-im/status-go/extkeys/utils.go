package extkeys

import (
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
)

// errors
var (
	ErrInvalidSecretKey = errors.New("generated secret key cannot be used")
)

func splitHMAC(seed, salt []byte) (secretKey, chainCode []byte, err error) {
	data := hmac.New(sha512.New, salt)
	if _, err = data.Write(seed); err != nil {
		return
	}
	I := data.Sum(nil)

	// Split I into two 32-byte sequences, IL and IR.
	// IL = master secret key
	// IR = master chain code
	secretKey = I[:32]
	chainCode = I[32:]

	// IL (secretKey) is expected to be a 256-bit integer (it is used as parse256(IL)),
	// and consequently that integer must be within range for SECP256k1 private key.
	//
	// There's tiny possibility (<1 in 2^127) this invariant is violated:
	//   error is returned in that case, and simple resolution is to request another child with i incremented.
	keyBigInt := new(big.Int).SetBytes(secretKey)
	if keyBigInt.Cmp(btcec.S256().N) >= 0 || keyBigInt.Sign() == 0 {
		err = ErrInvalidSecretKey
	}

	return
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
// nolint: unparam
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}
