package bytehelper

import (
	"errors"
)

// SliceToArray will convert byte slice to a 32 byte array
func SliceToArray(bytes []byte) [32]byte {
	var byteArray [32]byte
	copy(byteArray[:], bytes)
	return byteArray
}

// SliceToArray64 will convert byte slice to a 64 byte array
func SliceToArray64(bytes []byte) [64]byte {
	var byteArray [64]byte
	copy(byteArray[:], bytes)
	return byteArray
}

// ArrayToSlice will convert a 32 byte array to byte slice
func ArrayToSlice(bytes [32]byte) []byte {
	return bytes[:]
}

// ArrayToSlice64 will convert a 64 byte array to byte slice
func ArrayToSlice64(bytes [64]byte) []byte {
	return bytes[:]
}

// Split will take the given byte array and split it into half,
// with the first half being "firstLength" in size and the second
// half "secondLength" in size.
func Split(input []byte, firstLength, secondLength int) [][]byte {
	parts := make([][]byte, 2)

	parts[0] = make([]byte, firstLength)
	copy(parts[0], input[:firstLength])

	parts[1] = make([]byte, secondLength)
	copy(parts[1], input[firstLength:])

	return parts
}

// SplitThree will take the given byte array and split it into thirds,
// with the first third being "firstLength" in size, the second third
// being "secondLength" in size, and the last third being "thirdLength"
// in size.
func SplitThree(input []byte, firstLength, secondLength, thirdLength int) ([][]byte, error) {
	if input == nil || firstLength < 0 || secondLength < 0 || thirdLength < 0 ||
		len(input) < firstLength+secondLength+thirdLength {

		return nil, errors.New("Input too small: " + string(input))
	}

	parts := make([][]byte, 3)

	parts[0] = make([]byte, firstLength)
	copy(parts[0], input[:firstLength])

	parts[1] = make([]byte, secondLength)
	copy(parts[1], input[firstLength:][:secondLength])

	parts[2] = make([]byte, thirdLength)
	copy(parts[2], input[firstLength+secondLength:])

	return parts, nil
}

// Trim will trim the given byte array to the given length.
func Trim(input []byte, length int) []byte {
	result := make([]byte, length)
	copy(result, input[:length])

	return result
}

// Bytes5ToInt64 will convert the given byte array and offset to an int64.
func Bytes5ToInt64(bytes []byte, offset int) int64 {

	value := (int64(bytes[offset]&0xff) << 32) |
		(int64(bytes[offset+1]&0xff) << 24) |
		(int64(bytes[offset+2]&0xff) << 16) |
		(int64(bytes[offset+3]&0xff) << 8) |
		int64(bytes[offset+4]&0xff)

	return value
}

// CopySlice returns a copy of the given bytes.
func CopySlice(bytes []byte) []byte {
	cp := make([]byte, len(bytes))
	copy(cp, bytes)

	return cp
}
