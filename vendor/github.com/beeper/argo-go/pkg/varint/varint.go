// Package varint implements ULEB128 (Unsigned Little Endian Base 128)
// and ZigZag encoding/decoding for arbitrary-precision integers (*big.Int)
// and standard integer types (uint64, int64).
// ULEB128 is a variable-length encoding for unsigned integers.
// ZigZag encoding maps signed integers to unsigned integers so they
// can be efficiently encoded using ULEB128.
package varint

import (
	"errors"
	"math/big"
)

// Pre-allocate common big.Int values to reduce allocations in loops/checks.
var (
	big0    = big.NewInt(0)
	big1    = big.NewInt(1)
	big7f   = big.NewInt(0x7f) // 127
	big80   = big.NewInt(0x80) // 128
	bigNeg1 = big.NewInt(-1)   // Used for ~0 equivalent in Xor

	// Constants for BytesNeeded
	uMax1 = big.NewInt(0x7f)
	uMax2 = big.NewInt(0x3fff)
	uMax3 = big.NewInt(0x1fffff)
	uMax4 = big.NewInt(0xfffffff)
	uMax5 = big.NewInt(0x7ffffffff)
	uMax6 = big.NewInt(0x3ffffffffff)
	uMax7 = big.NewInt(0x1ffffffffffff)
	uMax8 = big.NewInt(0xffffffffffffff)
	uMax9 = big.NewInt(0x7fffffffffffffff) // This is MaxInt64, last one before needing more than 9 bytes for typical uint64 range
)

// UnsignedBytesNeeded calculates the number of bytes required to ULEB128 encode n.
func UnsignedBytesNeeded(n *big.Int) int {
	if n.Cmp(uMax1) <= 0 {
		return 1
	}
	if n.Cmp(uMax2) <= 0 {
		return 2
	}
	if n.Cmp(uMax3) <= 0 {
		return 3
	}
	if n.Cmp(uMax4) <= 0 {
		return 4
	}
	if n.Cmp(uMax5) <= 0 {
		return 5
	}
	if n.Cmp(uMax6) <= 0 {
		return 6
	}
	if n.Cmp(uMax7) <= 0 {
		return 7
	}
	if n.Cmp(uMax8) <= 0 {
		return 8
	}
	if n.Cmp(uMax9) <= 0 {
		return 9
	} // Max value for 9 bytes is 2^63-1

	// General case for numbers larger than 2^63-1
	needed := 0
	// Create a copy to avoid modifying the input n
	tempN := new(big.Int).Set(n)
	for tempN.Cmp(big0) > 0 {
		needed++
		tempN.Rsh(tempN, 7)
	}
	return needed
}

// UnsignedEncode ULEB128 encodes n into a new byte slice.
func UnsignedEncode(n *big.Int) []byte {
	length := UnsignedBytesNeeded(n)
	buf := make([]byte, length)
	UnsignedEncodeInto(n, buf, 0)
	return buf
}

// UnsignedEncodeInto ULEB128 encodes n into the provided buffer buf starting at offset.
// It returns the number of bytes written.
// Panics if the buffer is too small. Callers should ensure buf has enough space,
// e.g., by using UnsignedBytesNeeded.
func UnsignedEncodeInto(n *big.Int, buf []byte, offset int) int {
	// Create a copy to avoid modifying the input n
	tempN := new(big.Int).Set(n)
	pos := offset

	// Use a temporary big.Int for calculations to avoid allocations in the loop
	octet := new(big.Int)

	for {
		octet.And(tempN, big7f) // octet = tempN & 0x7f
		tempN.Rsh(tempN, 7)     // tempN = tempN >> 7

		if pos >= len(buf) {
			panic("varint: buffer too small for UnsignedEncodeInto")
		}

		if tempN.Cmp(big0) == 0 { // No more bits to encode
			buf[pos] = byte(octet.Uint64()) // No continuation bit
			pos++
			break
		} else {
			// There are more bits, so set the continuation bit (0x80)
			octet.Or(octet, big80)
			buf[pos] = byte(octet.Uint64())
			pos++
		}
	}
	return pos - offset
}

// UnsignedDecode ULEB128 decodes a number from buf starting at offset.
// It returns the decoded number, the number of bytes read, and an error if any.
func UnsignedDecode(buf []byte, offset int) (result *big.Int, length int, err error) {
	result = big.NewInt(0)
	var shift uint = 0
	pos := offset

	// Use a temporary big.Int for calculations to avoid allocations in the loop
	octetVal := new(big.Int)

	for {
		if pos >= len(buf) {
			return nil, 0, errors.New("varint: buffer too short for UnsignedDecode")
		}
		octet := buf[pos]
		pos++

		// Get the 7-bit value part
		octetVal.SetInt64(int64(octet & 0x7F))

		// Shift it and OR into the result
		octetVal.Lsh(octetVal, shift)
		result.Or(result, octetVal)

		if (octet & 0x80) == 0 { // Check continuation bit
			return result, pos - offset, nil
		}

		shift += 7
		// Protect against varints that are too long for a 256-bit number (max 37 bytes).
		// Max shift for data in the 37th byte is (37-1)*7 = 252.
		// shift > 252 indicates processing beyond the 37th byte.
		// (pos - offset) > 37 indicates more than 37 bytes read.
		if shift > 63 && (pos-offset) > 9 {
			return nil, 0, errors.New("varint: varint too large for 64-bit")
		}
	}
}

// --- ZigZag Encoding ---

// toZigZag converts a signed integer n to an unsigned integer suitable for ULEB128 encoding.
func toZigZag(n *big.Int) *big.Int {
	if n.BitLen() > 63 { // keep parity with Erlang’s 64-bit limit
		panic("varint: value out of int64 range")
	}
	res := new(big.Int).Lsh(n, 1)
	if n.Sign() < 0 {
		res.Xor(res, bigNeg1)
	}
	return res.And(res, big.NewInt(0).SetUint64(^uint64(0))) // mask 64 bits
}

// fromZigZag converts an unsigned integer n (decoded from ULEB128) back to a signed integer.
func fromZigZag(n *big.Int) *big.Int {
	res := new(big.Int)
	temp := new(big.Int).And(n, big1)

	res.Rsh(n, 1)
	if temp.Cmp(big1) == 0 { // If the last bit of n is 1 (sign bit for negative original)
		// ^ (~0) which is XOR with -1 for big.Int
		res.Xor(res, bigNeg1)
	}
	return res
}

// ZigZagEncode encodes a signed integer n using ZigZag and ULEB128.
func ZigZagEncode(n *big.Int) []byte {
	zz := toZigZag(n)
	return UnsignedEncode(zz)
}

// ZigZagEncodeInto encodes a signed integer n into buf using ZigZag and ULEB128.
// Returns the number of bytes written.
// Panics if the buffer is too small.
func ZigZagEncodeInto(n *big.Int, buf []byte, offset int) int {
	zz := toZigZag(n)
	return UnsignedEncodeInto(zz, buf, offset)
}

// ZigZagDecode decodes a ZigZag-ULEB128 encoded number from buf.
// Returns the decoded signed number, bytes read, and an error.
func ZigZagDecode(buf []byte, offset int) (result *big.Int, length int, err error) {
	unsignedVal, l, err := UnsignedDecode(buf, offset)
	if err != nil {
		return nil, 0, err
	}
	return fromZigZag(unsignedVal), l, nil
}

// --- Convenience functions for int64/uint64 if needed ---
// These are often useful as big.Int can be overkill for common integer sizes.

// UnsignedEncodeUint64 ULEB128 encodes a uint64.
func UnsignedEncodeUint64(val uint64) []byte {
	n := new(big.Int).SetUint64(val)
	return UnsignedEncode(n)
}

// UnsignedDecodeToUint64 ULEB128 decodes to a uint64.
// Returns an error if the decoded value overflows uint64.
func UnsignedDecodeToUint64(buf []byte, offset int) (result uint64, length int, err error) {
	bn, l, err := UnsignedDecode(buf, offset)
	if err != nil {
		return 0, 0, err
	}
	if !bn.IsUint64() {
		return 0, l, errors.New("varint: value overflows uint64")
	}
	return bn.Uint64(), l, nil
}

// ZigZagEncodeInt64 encodes an int64 using ZigZag and ULEB128.
func ZigZagEncodeInt64(val int64) []byte {
	n := big.NewInt(val)
	return ZigZagEncode(n)
}

// ZigZagDecodeToInt64 decodes a ZigZag-ULEB128 encoded number to int64.
// Returns an error if the decoded value overflows int64.
func ZigZagDecodeToInt64(buf []byte, offset int) (result int64, length int, err error) {
	bn, l, err := ZigZagDecode(buf, offset)
	if err != nil {
		return 0, 0, err
	}
	if !bn.IsInt64() {
		return 0, l, errors.New("varint: value overflows int64")
	}
	return bn.Int64(), l, nil
}
