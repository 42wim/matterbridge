/*
Package base36 provides a reasonably fast implementation of a binary base36 codec.
*/
package base36

// Simplified code based on https://godoc.org/github.com/mr-tron/base58
// which in turn is based on https://github.com/trezor/trezor-crypto/commit/89a7d7797b806fac

import (
	"fmt"
)

const UcAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const LcAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
const maxDigitOrdinal = 'z'
const maxDigitValueB36 = 35

var revAlphabet [maxDigitOrdinal + 1]byte

func init() {
	for i := range revAlphabet {
		revAlphabet[i] = maxDigitValueB36 + 1
	}
	for i, c := range UcAlphabet {
		revAlphabet[byte(c)] = byte(i)
		if c > '9' {
			revAlphabet[byte(c)+32] = byte(i)
		}
	}
}

// EncodeToStringUc encodes the given byte-buffer as base36 using [0-9A-Z] as
// the digit-alphabet
func EncodeToStringUc(b []byte) string { return encode(b, UcAlphabet) }

// EncodeToStringLc encodes the given byte-buffer as base36 using [0-9a-z] as
// the digit-alphabet
func EncodeToStringLc(b []byte) string { return encode(b, LcAlphabet) }

func encode(inBuf []byte, al string) string {

	bufsz := len(inBuf)
	zcnt := 0
	for zcnt < bufsz && inBuf[zcnt] == 0 {
		zcnt++
	}

	// It is crucial to make this as short as possible, especially for
	// the usual case of CIDs.
	bufsz = zcnt +
		// This is an integer simplification of
		// ceil(log(256)/log(36))
		(bufsz-zcnt)*277/179 + 1

	// Note: pools *DO NOT* help, the overhead of zeroing
	// kills any performance gain to be had
	out := make([]byte, bufsz)

	var idx, stopIdx int
	var carry uint32

	stopIdx = bufsz - 1
	for _, b := range inBuf[zcnt:] {
		idx = bufsz - 1
		for carry = uint32(b); idx > stopIdx || carry != 0; idx-- {
			carry += uint32((out[idx])) * 256
			out[idx] = byte(carry % 36)
			carry /= 36
		}
		stopIdx = idx
	}

	// Determine the additional "zero-gap" in the buffer (aside from zcnt)
	for stopIdx = zcnt; stopIdx < bufsz && out[stopIdx] == 0; stopIdx++ {
	}

	// Now encode the values with actual alphabet in-place
	vBuf := out[stopIdx-zcnt:]
	bufsz = len(vBuf)
	for idx = 0; idx < bufsz; idx++ {
		out[idx] = al[vBuf[idx]]
	}

	return string(out[:bufsz])
}

// DecodeString takes a base36 encoded string and returns a slice of the decoded
// bytes.
func DecodeString(s string) ([]byte, error) {

	if len(s) == 0 {
		return nil, fmt.Errorf("can not decode zero-length string")
	}

	zcnt := 0
	for zcnt < len(s) && s[zcnt] == '0' {
		zcnt++
	}

	// the 32bit algo stretches the result up to 2 times
	binu := make([]byte, 2*(((len(s))*179/277)+1)) // no more than 84 bytes when len(s) <= 64
	outi := make([]uint32, (len(s)+3)/4)           // no more than 16 bytes when len(s) <= 64

	for _, r := range s {
		if r > maxDigitOrdinal || revAlphabet[r] > maxDigitValueB36 {
			return nil, fmt.Errorf("invalid base36 character (%q)", r)
		}

		c := uint64(revAlphabet[r])

		for j := len(outi) - 1; j >= 0; j-- {
			t := uint64(outi[j])*36 + c
			c = (t >> 32)
			outi[j] = uint32(t & 0xFFFFFFFF)
		}
	}

	mask := (uint(len(s)%4) * 8)
	if mask == 0 {
		mask = 32
	}
	mask -= 8

	outidx := 0
	for j := 0; j < len(outi); j++ {
		for mask < 32 { // loop relies on uint overflow
			binu[outidx] = byte(outi[j] >> mask)
			mask -= 8
			outidx++
		}
		mask = 24
	}

	// find the most significant byte post-decode, if any
	for msb := zcnt; msb < outidx; msb++ {
		if binu[msb] > 0 {
			return binu[msb-zcnt : outidx : outidx], nil
		}
	}

	// it's all zeroes
	return binu[:outidx:outidx], nil
}
