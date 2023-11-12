// Code extracted from vendor/github.com/ethereum/go-ethereum/common/hexutil/hexutil.go

package types

import (
	"encoding/hex"
	"fmt"
	"reflect"
)

var (
	bytesT = reflect.TypeOf(HexBytes(nil))
)

// HexBytes marshals/unmarshals as a JSON string with 0x prefix.
// The empty slice marshals as "0x".
type HexBytes []byte

func (b HexBytes) Bytes() []byte {
	result := make([]byte, len(b)*2+2)
	copy(result, `0x`)
	hex.Encode(result[2:], b)
	return result
}

// MarshalText implements encoding.TextMarshaler
func (b HexBytes) MarshalText() ([]byte, error) {
	return b.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *HexBytes) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(bytesT)
	}
	return wrapTypeError(b.UnmarshalText(input[1:len(input)-1]), bytesT)
}

// UnmarshalFixedUnprefixedText decodes the input as a string with optional 0x prefix. The
// length of out determines the required input length. This function is commonly used to
// implement the UnmarshalText method for fixed-size types.
func UnmarshalFixedUnprefixedText(typname string, input, out []byte) error {
	raw, err := checkText(input, false)
	if err != nil {
		return err
	}
	if len(raw)/2 != len(out) {
		return fmt.Errorf("hex string has length %d, want %d for %s", len(raw), len(out)*2, typname)
	}
	// Pre-verify syntax before modifying out.
	for _, b := range raw {
		if decodeNibble(b) == badNibble {
			return ErrSyntax
		}
	}
	_, err = hex.Decode(out, raw)
	return err
}
