package varint

import (
	"encoding/binary"
)

// Decode reads the first varint encoded number from the given buffer.
//
// On success, the function returns the varint as an int64, and the number of
// bytes read (0 if there was an error).
func Decode(b []byte) (int64, int) {
	if len(b) == 0 {
		return 0, 0
	}
	// 0xxxxxxx 7-bit positive number
	if (b[0] & 0x80) == 0 {
		return int64(b[0]), 1
	}
	// 10xxxxxx + 1 byte 14-bit positive number
	if (b[0]&0xC0) == 0x80 && len(b) >= 2 {
		return int64(b[0]&0x3F)<<8 | int64(b[1]), 2
	}
	// 110xxxxx + 2 bytes 21-bit positive number
	if (b[0]&0xE0) == 0xC0 && len(b) >= 3 {
		return int64(b[0]&0x1F)<<16 | int64(b[1])<<8 | int64(b[2]), 3
	}
	// 1110xxxx + 3 bytes 28-bit positive number
	if (b[0]&0xF0) == 0xE0 && len(b) >= 4 {
		return int64(b[0]&0xF)<<24 | int64(b[1])<<16 | int64(b[2])<<8 | int64(b[3]), 4
	}
	// 111100__ + int (32-bit) 32-bit positive number
	if (b[0]&0xFC) == 0xF0 && len(b) >= 5 {
		return int64(binary.BigEndian.Uint32(b[1:])), 5
	}
	// 111101__ + long (64-bit) 64-bit number
	if (b[0]&0xFC) == 0xF4 && len(b) >= 9 {
		return int64(binary.BigEndian.Uint64(b[1:])), 9
	}
	// 111110__ + varint Negative recursive varint
	if b[0]&0xFC == 0xF8 {
		if v, n := Decode(b[1:]); n > 0 {
			return -v, n + 1
		}
		return 0, 0
	}
	// 111111xx Byte-inverted negative two bit number (~xx)
	if b[0]&0xFC == 0xFC {
		return ^int64(b[0] & 0x03), 1
	}

	return 0, 0
}
