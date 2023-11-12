package bigint

import (
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Unmarshals a u256 as a fixed-length hex string with 0x prefix and leading zeros
type HexBigInt struct {
	*big.Int
}

const FixedLength = 32 // u256 -> 32 bytes

var (
	hexBigIntT = reflect.TypeOf(HexBigInt{})
)

func (b *HexBigInt) UnmarshalJSON(input []byte) error {
	var buf [FixedLength]byte
	err := hexutil.UnmarshalFixedJSON(hexBigIntT, input, buf[:])

	if err != nil {
		return err
	}

	z := new(big.Int)
	z.SetBytes(buf[:])
	b.Int = z
	return nil
}
