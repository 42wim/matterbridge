// Package binary ...
// thanks to https://github.com/skip2/go-qrcode/blob/master/bitset/bitset.go
// I cannot do any better for now, so I just learn and write it again~
package binary

import (
	"bytes"
	"fmt"
	"log"
)

const (
	byteTrue  byte = '1'
	byteFalse byte = '0'
)

var (
	// format string
	format = "Binary length: %d, bits: %s"
)

// New ...
func New(booleans ...bool) *Binary {
	b := &Binary{
		bits:    make([]byte, 0),
		lenBits: 0,
	}
	b.AppendBools(booleans...)
	return b
}

// NewFromBinaryString ... generate Bitset from binary string
// auto get length
func NewFromBinaryString(s string) (*Binary, error) {
	var n = len(s) / 8
	if len(s)%8 != 0 {
		n++
	}

	b := &Binary{
		bits:    make([]byte, n), // prealloc memory, reducing useless space
		lenBits: 0,
	}

	for _, c := range s {
		switch c {
		case '1':
			b.AppendBools(true)
		case '0':
			b.AppendBools(false)
		case ' ':
			// skip space blank
			continue
		default:
			err := fmt.Errorf("invalid char %c in NewFromBinaryString", c)
			return nil, err
		}
	}

	return b, nil
}

// Binary struct contains bits stream and methods to be called from outside
// exsample:
// b.Len()
// b.Subset(start, end)
// b.At(pos)
type Binary struct {
	bits    []byte // 1byte = 8bit
	lenBits int    // len(bits) * 8
}

// ensureCapacity ensures the Bitset can store an additional |numBits|.
//
// The underlying array is expanded if necessary. To prevent frequent
// reallocation, expanding the underlying array at least doubles its capacity.
//
// then no need to use append ~ will no panic (out of range)
func (b *Binary) ensureCapacity(numBits int) {
	numBits += b.lenBits

	newNumBytes := numBits / 8
	if numBits%8 != 0 {
		newNumBytes++
	}

	// if larger enough
	if len(b.bits) >= newNumBytes {
		return
	}

	// larger capcity, about 3 times of current capcity
	b.bits = append(b.bits, make([]byte, newNumBytes+2*len(b.bits))...)
}

// At .get boolean value from
func (b *Binary) At(pos int) bool {
	if pos < 0 || pos >= b.lenBits {
		panic("out range of bits")
	}

	return (b.bits[pos/8]&(0x80>>uint(pos%8)) != 0)
}

// Subset do the same work like slice[start:end]
func (b *Binary) Subset(start, end int) (*Binary, error) {
	if start > end || end > b.lenBits {
		err := fmt.Errorf("Out of range start=%d end=%d lenBits=%d", start, end, b.lenBits)
		return nil, err
	}

	result := New()
	result.ensureCapacity(end - start)

	for i := start; i < end; i++ {
		if b.At(i) {
			result.bits[result.lenBits/8] |= 0x80 >> uint(result.lenBits%8)
		}
		result.lenBits++
	}

	return result, nil
}

// Append other bitset link another Bitset to after the b
func (b *Binary) Append(other *Binary) {
	b.ensureCapacity(other.Len())

	for i := 0; i < other.lenBits; i++ {
		if other.At(i) {
			b.bits[b.lenBits/8] |= 0x80 >> uint(b.lenBits%8)
		}
		b.lenBits++
	}
}

// AppendUint32 other bitset link another Bitset to after the b
func (b *Binary) AppendUint32(value uint32, numBits int) {
	b.ensureCapacity(numBits)

	if numBits > 32 {
		log.Panicf("numBits %d out of range 0-32", numBits)
	}

	for i := numBits - 1; i >= 0; i-- {
		if value&(1<<uint(i)) != 0 {
			b.bits[b.lenBits/8] |= 0x80 >> uint(b.lenBits%8)
		}

		b.lenBits++
	}
}

// AppendBytes ...
func (b *Binary) AppendBytes(byts ...byte) {
	for _, byt := range byts {
		b.AppendByte(byt, 8)
	}
}

// AppendByte ... specified num bits to append
func (b *Binary) AppendByte(byt byte, numBits int) error {
	if numBits > 8 || numBits < 0 {
		return fmt.Errorf("numBits out of range 0-8")
	}

	b.ensureCapacity(numBits)

	// append bit in byte
	for i := numBits - 1; i >= 0; i-- {
		// 0x01 << left shift count
		// 0x80 >> right shift count
		if byt&(0x01<<uint(i)) != 0 {
			b.bits[b.lenBits/8] |= 0x80 >> uint(b.lenBits%8)
		}
		b.lenBits++
	}
	return nil
}

// AppendBools append multi bool after the bit stream of b
func (b *Binary) AppendBools(booleans ...bool) {
	b.ensureCapacity(len(booleans))
	for _, bv := range booleans {
		if bv {
			b.bits[b.lenBits/8] |= 0x80 >> uint(b.lenBits%8)
		}
		b.lenBits++
	}
}

// AppendNumBools appends num bits of value value.
func (b *Binary) AppendNumBools(num int, boolean bool) {
	booleans := make([]bool, num)
	// if not false just append
	if boolean {
		for i := 0; i < num; i++ {
			booleans[i] = boolean
		}
	}
	b.AppendBools(booleans...)
}

// IterFunc used by func b.VisitAll ...
type IterFunc func(pos int, v bool)

// VisitAll loop the b.bits stream and send value into IterFunc
func (b *Binary) VisitAll(f IterFunc) {
	for pos := 0; pos < b.Len(); pos++ {
		f(pos, b.At(pos))
	}
}

// String for printing
func (b *Binary) String() string {
	var (
		bitstr []byte
		vb     byte
	)

	b.VisitAll(func(pos int, v bool) {
		vb = byteFalse
		if v {
			vb = byteTrue
		}
		bitstr = append(bitstr, vb)
	})

	return fmt.Sprintf(format, b.Len(), string(bitstr))
}

// Len ...
func (b *Binary) Len() int {
	return b.lenBits
}

// Bytes ...
func (b *Binary) Bytes() []byte {
	numBytes := b.lenBits / 8
	if b.lenBits%8 != 0 {
		numBytes++
	}
	return b.bits[:numBytes]
}

// EqualTo ...
func (b *Binary) EqualTo(other *Binary) bool {
	if b.lenBits != other.lenBits {
		return false
	}

	numByte := b.lenBits / 8
	if !bytes.Equal(b.bits[:numByte], other.bits[:numByte]) {
		return false
	}

	for pos := numByte * 8; pos < b.lenBits; pos++ {
		if b.At(pos) != other.At(pos) {
			return false
		}
	}

	return true
}

// Copy ...
func (b *Binary) Copy() *Binary {
	return &Binary{
		bits:    b.bits,
		lenBits: b.lenBits,
	}
}
