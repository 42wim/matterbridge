// Package reedsolomon ...
// ref to doc: https://en.wikiversity.org/wiki/Reed%E2%80%93Solomon_codes_for_coders#Polynomial_division
// ref to project: github.com/skip2/go-qrcode/reedsolomon
package reedsolomon

import (
	"github.com/yeqown/reedsolomon/binary"
)

type word byte // 8bit as a word

// Encode ...
func Encode(bin *binary.Binary, numECWords int) *binary.Binary {
	if bin.Len()%8 != 0 {
		panic("could not deal with binary times 8bits")
	}
	// generate polynomial
	generator := rsGenPoly(numECWords)

	// poly div
	remainder := polyDiv(bin.Bytes(), generator)

	// append error correction stream
	bout := bin.Copy()
	bout.AppendBytes(remainder...)
	return bout
}

// Decode ...
// TODO: finish this ~
func Decode(bin *binary.Binary, numECWords int) *binary.Binary {
	return nil
}
