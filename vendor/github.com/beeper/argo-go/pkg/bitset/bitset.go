// Package bitset provides a BitSet data structure backed by *big.Int,
// along with helpers for reading and writing bitsets in variable-length
// (self-delimiting) and fixed-length formats.
package bitset

import (
	"fmt"
	"math/big"

	"github.com/beeper/argo-go/pkg/buf"
)

// BitSet implements a growable set of bits, backed by a *big.Int.
// It provides methods to get, set, and unset individual bits.
// The zero value of a BitSet is not ready for use; always use NewBitSet.
type BitSet struct {
	val *big.Int
}

// NewBitSet creates and initializes a new BitSet to zero.
func NewBitSet() *BitSet {
	return &BitSet{val: big.NewInt(0)}
}

// GetBit returns the boolean value of the bit at the given index.
// It panics if the index is negative.
func (bs *BitSet) GetBit(index int) bool {
	if index < 0 {
		panic("Bitset index must be positive")
	}
	return bs.val.Bit(index) == 1
}

// SetBit sets the bit at the given index to 1 (true).
// It panics if the index is negative.
// Returns the BitSet pointer for chaining.
func (bs *BitSet) SetBit(index int) *BitSet {
	if index < 0 {
		panic("Bitset index must be positive")
	}
	bs.val.SetBit(bs.val, index, 1)
	return bs
}

// UnsetBit sets the bit at the given index to 0 (false).
// It panics if the index is negative.
// Returns the BitSet pointer for chaining.
func (bs *BitSet) UnsetBit(index int) *BitSet {
	if index < 0 {
		panic("Bitset index must be positive")
	}
	bs.val.SetBit(bs.val, index, 0)
	return bs
}

// Bytes returns the raw byte representation of the BitSet's underlying *big.Int.
// The bytes are in big-endian order. Returns nil if the BitSet or its internal value is nil.
func (bs *BitSet) Bytes() []byte {
	if bs == nil || bs.val == nil {
		return nil
	}
	return bs.val.Bytes()
}

// VarBitSet provides methods for reading and writing variable-length, self-delimiting bitsets.
// In this format, each byte encodes 7 bits of data, with the least significant bit (LSB)
// acting as a continuation flag (1 means more bytes follow, 0 means this is the last byte).
type VarBitSet struct{}

// Read reads a variable-length, self-delimiting bitset from a buf.Read.
// It reconstructs the BitSet from bytes where each byte contributes 7 data bits
// and 1 continuation bit. It returns the number of bytes read, the resulting BitSet,
// and any error encountered during reading.
func (v *VarBitSet) Read(buf buf.Read) (int, *BitSet, error) {
	bytesRead := 0
	bitset := NewBitSet()
	more := true
	bitPos := 0
	for more {
		byteVal, err := buf.ReadByte()
		if err != nil {
			return bytesRead, nil, fmt.Errorf("failed to read byte for var bitset: %w", err)
		}
		bytesRead++
		// byteVal is already a byte, no need to check if <0 or >255
		shiftedVal := new(big.Int).Lsh(big.NewInt(int64((byteVal&0xff)>>1)), uint(bitPos))
		bitset.val.Or(bitset.val, shiftedVal)
		bitPos += 7
		more = (byteVal & 1) == 1
	}
	return bytesRead, bitset, nil
}

// Write encodes a BitSet into a variable-length, self-delimiting byte slice.
// Each byte in the output slice contains 7 bits of data from the BitSet and one
// continuation bit (LSB). If the LSB is 1, more bytes follow; if 0, it's the last byte.
// The 'padToLength' argument ensures the output byte slice has at least that many bytes,
// padding with zero bytes (0x00, which represents zero value and no continuation) if necessary.
// Returns an error if the BitSet contains negative values (which is not typical for bitsets).
func (v *VarBitSet) Write(bs *BitSet, padToLength int) ([]byte, error) {
	if bs.val.Sign() < 0 {
		return nil, fmt.Errorf("Bitsets must only contain positive values")
	}
	var bytes []byte
	// Make a copy to avoid modifying the original
	valCopy := new(big.Int).Set(bs.val)
	more := valCopy.Sign() > 0

	for more {
		byteValBig := new(big.Int).And(valCopy, big.NewInt(0x7f)) // Get the lowest 7 bits
		byteVal := byteValBig.Int64() << 1                        // Shift left to make space for the 'more' bit

		valCopy.Rsh(valCopy, 7) // Shift right by 7 bits
		more = valCopy.Sign() > 0
		if more {
			byteVal = byteVal | 1 // Set the 'more' bit
		}
		bytes = append(bytes, byte(byteVal))
	}

	if len(bytes) == 0 { // If original value was 0, loop didn't run.
		bytes = append(bytes, 0x00)
	}

	if padToLength > len(bytes) {
		padding := make([]byte, padToLength-len(bytes))
		bytes = append(bytes, padding...)
	}
	return bytes, nil
}

// FixedBitSet provides methods for reading and writing fixed-length, undelimited bitsets.
// In this format, all 8 bits of each byte are used for data.
// The total length of the bitset in bytes must be known beforehand for reading.
type FixedBitSet struct{}

// Write encodes a BitSet into a fixed-length, undelimited byte slice.
// All 8 bits of each byte in the output are used for data from the BitSet.
// The 'padToLength' argument ensures the output byte slice has at least that many bytes,
// padding with zero bytes if necessary. The least significant bytes of the BitSet appear first.
// Returns an error if the BitSet contains negative values.
func (f *FixedBitSet) Write(bs *BitSet, padToLength int) ([]byte, error) {
	if bs.val.Sign() < 0 {
		return nil, fmt.Errorf("Bitsets must only contain positive values")
	}
	var bytes []byte
	valCopy := new(big.Int).Set(bs.val)
	more := valCopy.Sign() > 0

	for more {
		byteValBig := new(big.Int).And(valCopy, big.NewInt(0xff)) // Get the lowest 8 bits
		byteVal := byteValBig.Uint64()
		valCopy.Rsh(valCopy, 8) // Shift right by 8 bits
		more = valCopy.Sign() > 0
		bytes = append(bytes, byte(byteVal))
	}

	if len(bytes) == 0 && padToLength > 0 { // Handle case where bitset is 0
		bytes = append(bytes, 0)
	}

	if padToLength > len(bytes) {
		padding := make([]byte, padToLength-len(bytes))
		bytes = append(bytes, padding...)
	}
	return bytes, nil
}

// Read reconstructs a BitSet from a fixed-length, undelimited sequence of bytes
// within a larger byte slice. 'pos' specifies the starting position in 'bytes',
// and 'length' specifies how many bytes to read to form the bitset.
// All 8 bits of each read byte contribute to the BitSet.
func (f *FixedBitSet) Read(bytes []byte, pos int, length int) (*BitSet, error) {
	if pos+length > len(bytes) {
		return nil, fmt.Errorf("not enough bytes to read fixed bitset of length %d from pos %d", length, pos)
	}

	bitset := NewBitSet()
	bitPos := 0
	for i := 0; i < length; i++ {
		byteVal := bytes[pos+i]
		shiftedVal := new(big.Int).Lsh(big.NewInt(int64(byteVal&0xff)), uint(bitPos))
		bitset.val.Or(bitset.val, shiftedVal)
		bitPos += 8
	}
	return bitset, nil
}

// BytesNeededForNumBits calculates the minimum number of bytes required to store
// a fixed-length bitset containing 'numBits' of information.
// This is equivalent to ceil(numBits / 8).
func (f *FixedBitSet) BytesNeededForNumBits(numBits int) int {
	if numBits <= 0 {
		return 0
	}
	return (numBits + 7) / 8 // Equivalent to Math.ceil(numBits / 8)
}
