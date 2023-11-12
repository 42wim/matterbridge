package bigint

import (
	"fmt"
	"math/big"
	"strings"
)

type BigInt struct {
	*big.Int
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	return []byte("\"" + b.String() + "\""), nil
}

func (b *BigInt) UnmarshalJSON(p []byte) error {
	if string(p) == "null" {
		return nil
	}
	z := new(big.Int)
	_, ok := z.SetString(strings.Trim(string(p), "\""), 10)
	if !ok {
		return fmt.Errorf("not a valid big integer: %s", string(p))
	}
	b.Int = z
	return nil
}
