package buf

import (
	"errors"
	"io"
)

// BufPosition defines an interface for types that track a position.
type BufPosition interface {
	Position() int64
	SetPosition(position int64)
	IncrementPosition(numBytes int64)
}

// Read defines an interface for readable buffers.
type Read interface {
	BufPosition
	io.Reader
	io.ByteReader
	Get(position int64) (byte, error)
	Bytes() []byte
	Len() int
	Peek(n int) ([]byte, error)
}

// Write defines an interface for writable buffers.
type Write interface {
	BufPosition
	io.Writer
	io.ByteWriter
	Cap() int
}

// Buf is a dynamically-sized byte buffer, wrapping a []byte.
// It manages its own position, logical length, and capacity growth.
type Buf struct {
	data   []byte // Underlying byte slice
	pos    int64  // Current read/write position
	length int64  // Logical length of the data written (number of valid bytes in data)
}

const defaultInitialBufferSize = 32 // Or a reasonable default like 16, 32, etc.

// NewBuf creates a new Buf with an initial capacity.
func NewBuf(initialCapacity int) *Buf {
	if initialCapacity < 0 {
		initialCapacity = defaultInitialBufferSize
	}
	return &Buf{
		data:   make([]byte, 0, initialCapacity),
		pos:    0,
		length: 0,
	}
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes by slicing data to zero length.
// Reset also resets the position and logical length to 0.
func (b *Buf) Reset() {
	b.data = b.data[:0] // Keeps capacity, sets len to 0
	b.pos = 0
	b.length = 0
}

// Position returns the current position in the buffer.
func (b *Buf) Position() int64 {
	return b.pos
}

// SetPosition sets the current position in the buffer.
// No bounds checking is performed on the position value itself,
// but subsequent operations (Read/Write) will handle boundaries.
func (b *Buf) SetPosition(position int64) {
	b.pos = position
}

// IncrementPosition increments the current position by numBytes.
func (b *Buf) IncrementPosition(numBytes int64) {
	b.pos += numBytes
}

// ensureCapacity ensures that the buffer has at least minCapacity.
// If the current capacity is insufficient, it grows the buffer,
// preserving the existing content up to b.length.
// After this call, cap(b.data) will be at least minCapacity,
// and len(b.data) will remain b.length (the old logical length).
func (b *Buf) ensureCapacity(minCapacity int64) {
	currentCap := int64(cap(b.data))
	if currentCap >= minCapacity {
		return
	}

	newActualCap := currentCap
	if newActualCap == 0 {
		// If starting from zero capacity, choose a small default or minCapacity.
		// Using a small default (e.g., 16) avoids too many small allocations
		// if minCapacity is also very small.
		if minCapacity < 16 {
			newActualCap = 16
		} else {
			newActualCap = minCapacity
		}
	}

	// Grow similar to Go's append strategy: double capacity until it's sufficient.
	for newActualCap < minCapacity {
		if newActualCap == 0 { // Should have been handled by initial assignment if currentCap was 0
			newActualCap = minCapacity // Fallback if somehow 0
			break
		}
		newActualCap *= 2
		if newActualCap < 0 { // Overflow check
			newActualCap = minCapacity // Max possible if overflow
			break
		}
	}
	// Final check if doubling overshot significantly but minCapacity is larger.
	if newActualCap < minCapacity {
		newActualCap = minCapacity
	}

	// Create new slice with new capacity, copy content up to current logical length (b.length).
	// len of newData will be b.length.
	newData := make([]byte, b.length, newActualCap)
	if b.length > 0 {
		copy(newData, b.data[:b.length])
	}
	b.data = newData
}

// Read reads up to len(p) bytes into p from the current position.
// It returns the number of bytes read and any error encountered.
// EOF is returned when no more bytes are available from the current position within the logical length.
func (b *Buf) Read(p []byte) (n int, err error) {
	if b.pos < 0 {
		// Reading from a negative position is invalid.
		return 0, io.EOF // Or a specific error like "invalid position"
	}

	// Per io.Reader contract, if len(p) == 0, Read should return n == 0 and err == nil.
	if len(p) == 0 {
		return 0, nil
	}

	if b.pos >= b.length {
		// Position is at or beyond the logical end of the buffer.
		return 0, io.EOF
	}

	readableBytes := b.length - b.pos
	numToRead := int64(len(p))

	if numToRead > readableBytes {
		numToRead = readableBytes
	}

	// This check is now effectively for numToRead > 0 after potentially being capped by readableBytes.
	// If readableBytes was 0, b.pos >= b.length would have caught it.
	// If len(p) was initially > 0 but numToRead became 0 due to readableBytes, it implies EOF.
	if numToRead <= 0 {
		return 0, io.EOF
	}

	copy(p, b.data[b.pos:b.pos+numToRead])
	b.pos += numToRead
	return int(numToRead), nil
}

// WriteBuf writes the content of bb into b at the current position.
// It updates b's position and length accordingly.
// Returns the number of bytes written from bb and any error encountered.
func (b *Buf) WriteBuf(bb *Buf) (n int, err error) {
	return b.Write(bb.Bytes())
}

// Write writes len(p) bytes from p to the buffer at the current position.
// It returns the number of bytes written and any error encountered.
// If the write extends beyond the current logical length, the buffer's logical length is updated.
// If the write starts beyond the current logical length, the gap is filled with zeros.
func (b *Buf) Write(p []byte) (n int, err error) {
	numBytesToWrite := len(p)
	if numBytesToWrite == 0 {
		return 0, nil
	}

	if b.pos < 0 {
		// Writing to a negative position is invalid.
		// Consider returning an error or panicking, as this indicates misuse.
		return 0, errors.New("argo.Buf.Write: negative position")
	}

	endPos := b.pos + int64(numBytesToWrite)

	newLogicalLength := b.length
	if endPos > newLogicalLength {
		newLogicalLength = endPos
	}

	if newLogicalLength > int64(cap(b.data)) {
		b.ensureCapacity(newLogicalLength)
		// After ensureCapacity, cap(b.data) >= newLogicalLength,
		// and len(b.data) == b.length (the old logical length).
	}

	// Ensure len(b.data) is sufficient for the newLogicalLength.
	// Reslicing b.data up to newLogicalLength will zero-fill any new bytes
	// between the old b.length and newLogicalLength if newLogicalLength > len(b.data).
	// This handles padding if b.pos > old b.length.
	if newLogicalLength > int64(len(b.data)) {
		b.data = b.data[:newLogicalLength]
	}
	// Now len(b.data) == newLogicalLength, and any extended parts are zero-initialized.

	copy(b.data[b.pos:], p)

	b.length = newLogicalLength
	b.pos = endPos

	return numBytesToWrite, nil
}

// ReadByte reads and returns the next byte from the current position.
// If no byte is available, it returns error io.EOF.
func (b *Buf) ReadByte() (byte, error) {
	if b.pos < 0 {
		return 0, io.EOF
	}
	if b.pos >= b.length {
		return 0, io.EOF
	}
	val := b.data[b.pos]
	b.pos++
	return val, nil
}

// WriteByte writes a single byte to the buffer at the current position.
func (b *Buf) WriteByte(c byte) error {
	if b.pos < 0 {
		return errors.New("argo.Buf.WriteByte: negative position")
	}

	endPos := b.pos + 1
	newLogicalLength := b.length
	if endPos > newLogicalLength {
		newLogicalLength = endPos
	}

	if newLogicalLength > int64(cap(b.data)) {
		b.ensureCapacity(newLogicalLength)
	}
	if newLogicalLength > int64(len(b.data)) {
		b.data = b.data[:newLogicalLength]
	}

	b.data[b.pos] = c

	b.length = newLogicalLength
	b.pos = endPos
	return nil
}

// Get returns the byte at the given absolute position in the buffer,
// without advancing the current position.
// Position is relative to the start of the buffer's logical content.
func (b *Buf) Get(position int64) (byte, error) {
	if position < 0 || position >= b.length {
		return 0, io.EOF
	}
	return b.data[position], nil
}

// Bytes returns a slice of all logical bytes currently in the buffer.
// The returned slice is valid until the next write operation that might cause reallocation.
func (b *Buf) Bytes() []byte {
	return b.data[:b.length]
}

// Len returns the current logical length of the buffer (number of bytes written).
func (b *Buf) Len() int {
	return int(b.length)
}

// Cap returns the current capacity of the buffer's underlying storage.
func (b *Buf) Cap() int {
	return cap(b.data)
}

// Peek returns the next n bytes from the current position without advancing the reader.
// The returned slice shares the underlying array of the buffer.
// If n is larger than the available bytes, Peek returns all available bytes.
func (b *Buf) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("argo.Buf.Peek: count cannot be negative")
	}
	if n == 0 {
		return []byte{}, nil
	}

	if b.pos < 0 || b.pos >= b.length {
		return nil, io.EOF // No bytes to peek if position is invalid or at/past end.
	}

	available := b.length - b.pos

	if int64(n) > available {
		return b.data[b.pos : b.pos+available], nil
	}
	return b.data[b.pos : b.pos+int64(n)], nil
}

// BufReadonly is a read-only wrapper around a byte slice.
// It's analogous to TypeScript's BufReadonly class.
type BufReadonly struct {
	bytes []byte
	pos   int64
}

// NewBufReadonly creates a new BufReadonly wrapping the given byte slice.
// The provided slice is used directly and should not be modified externally
// if read-only behavior is to be guaranteed.
func NewBufReadonly(data []byte) *BufReadonly {
	return &BufReadonly{bytes: data, pos: 0} // Initialize pos to 0
}

// Position returns the current position in the buffer.
func (br *BufReadonly) Position() int64 {
	return br.pos
}

// SetPosition sets the current position in the buffer.
func (br *BufReadonly) SetPosition(position int64) {
	br.pos = position
}

// IncrementPosition increments the current position by numBytes.
func (br *BufReadonly) IncrementPosition(numBytes int64) {
	br.pos += numBytes
}

// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered.
// If the buffer's read position is negative, it returns (0, io.EOF).
func (br *BufReadonly) Read(p []byte) (n int, err error) {
	if br.pos < 0 {
		return 0, io.EOF
	}
	if br.pos >= int64(len(br.bytes)) {
		return 0, io.EOF
	}
	copied := copy(p, br.bytes[br.pos:])
	br.pos += int64(copied)
	return copied, nil
}

// ReadByte reads and returns the next byte from the buffer.
// If no byte is available, it returns error io.EOF.
// If the buffer's read position is negative, it returns (0, io.EOF).
func (br *BufReadonly) ReadByte() (byte, error) {
	if br.pos < 0 {
		return 0, io.EOF
	}
	if br.pos >= int64(len(br.bytes)) {
		return 0, io.EOF
	}
	val := br.bytes[br.pos]
	br.pos++
	return val, nil
}

// Get returns the byte at the given absolute position in the buffer,
// without advancing the current position.
func (br *BufReadonly) Get(position int64) (byte, error) {
	if position < 0 || position >= int64(len(br.bytes)) {
		return 0, io.EOF
	}
	return br.bytes[position], nil
}

// Bytes returns all bytes in the buffer, regardless of position.
func (br *BufReadonly) Bytes() []byte {
	return br.bytes
}

// Len returns the total length of the underlying byte slice.
func (br *BufReadonly) Len() int {
	return len(br.bytes)
}

// Peek returns the next n bytes without advancing the reader.
// If n is negative, an error is returned. If n is 0, Peek returns an empty slice and no error.
// If the buffer's read position is out of bounds (negative or >= len(br.bytes)), it returns (nil, io.EOF).
// If n is larger than the available bytes from br.pos, Peek returns all available bytes and no error.
func (br *BufReadonly) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("argo.BufReadonly.Peek: count cannot be negative")
	}
	if n == 0 {
		return []byte{}, nil
	}
	if br.pos < 0 || br.pos >= int64(len(br.bytes)) {
		return nil, io.EOF
	}

	available := int64(len(br.bytes)) - br.pos
	if available <= 0 { // Should be covered by above check, but good for safety.
		return nil, io.EOF
	}

	if int64(n) > available {
		return br.bytes[br.pos : br.pos+available], nil
	}
	return br.bytes[br.pos : br.pos+int64(n)], nil
}

// Make sure Buf implements Read and Write
var _ Read = (*Buf)(nil)
var _ Write = (*Buf)(nil)

// Make sure BufReadonly implements Read
var _ Read = (*BufReadonly)(nil)
