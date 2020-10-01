package varint

import (
	"encoding/binary"
	"math"
)

// MaxVarintLen is the maximum number of bytes required to encode a varint
// number.
const MaxVarintLen = 10

// Encode encodes the given value to varint format.
func Encode(b []byte, value int64) int {
	// 111111xx Byte-inverted negative two bit number (~xx)
	if value <= -1 && value >= -4 {
		b[0] = 0xFC | byte(^value&0xFF)
		return 1
	}
	// 111110__ + varint Negative recursive varint
	if value < 0 {
		b[0] = 0xF8
		return 1 + Encode(b[1:], -value)
	}
	// 0xxxxxxx 7-bit positive number
	if value <= 0x7F {
		b[0] = byte(value)
		return 1
	}
	// 10xxxxxx + 1 byte 14-bit positive number
	if value <= 0x3FFF {
		b[0] = byte(((value >> 8) & 0x3F) | 0x80)
		b[1] = byte(value & 0xFF)
		return 2
	}
	// 110xxxxx + 2 bytes 21-bit positive number
	if value <= 0x1FFFFF {
		b[0] = byte((value>>16)&0x1F | 0xC0)
		b[1] = byte((value >> 8) & 0xFF)
		b[2] = byte(value & 0xFF)
		return 3
	}
	// 1110xxxx + 3 bytes 28-bit positive number
	if value <= 0xFFFFFFF {
		b[0] = byte((value>>24)&0xF | 0xE0)
		b[1] = byte((value >> 16) & 0xFF)
		b[2] = byte((value >> 8) & 0xFF)
		b[3] = byte(value & 0xFF)
		return 4
	}
	// 111100__ + int (32-bit) 32-bit positive number
	if value <= math.MaxInt32 {
		b[0] = 0xF0
		binary.BigEndian.PutUint32(b[1:], uint32(value))
		return 5
	}
	// 111101__ + long (64-bit) 64-bit number
	if value <= math.MaxInt64 {
		b[0] = 0xF4
		binary.BigEndian.PutUint64(b[1:], uint64(value))
		return 9
	}

	return 0
}
