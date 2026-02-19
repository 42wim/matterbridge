// Package block provides reader and writer types for processing Argo data blocks.
// These types handle different encoding strategies such as length-prefixing (with or without deduplication),
// fixed-size data, and unlabeled varints.
package block

import (
	"fmt"
	"io" // For io.EOF, io.ErrUnexpectedEOF

	"github.com/beeper/argo-go/label"
	"github.com/beeper/argo-go/pkg/buf"
)

// FromBytesFunc is a generic function type that defines a conversion from a byte slice
// to a specific output type `Out`. It is used by various block readers to transform
// raw bytes read from a block into the desired data type.
// This corresponds to the `(bytes: Uint8Array) => Out` callback in the TypeScript Argo implementation.
type FromBytesFunc[Out any] func(bytes []byte) Out

// CommonState holds shared state for block readers, primarily the `DataBuf` from which
// block data is read. It also provides a default `AfterNewRead` hook.
// Embedding CommonState allows concrete reader types to share this buffer and default hook.
type CommonState struct {
	// DataBuf is the buffer from which the actual data for this block is read.
	// This is distinct from any "parent" buffer that might contain labels or references.
	DataBuf buf.Read
}

// AfterNewRead is a hook method called by some block readers after successfully reading
// and processing a new value from the block. The default implementation is a no-op.
// Concrete block reader types can provide their own `AfterNewRead` implementation if specific
// post-read actions are needed (e.g., updating internal state for deduplication).
func (cs *CommonState) AfterNewRead() {
	// Default no-op, can be overridden by embedding types.
}

// LabelBlockReader reads values from its `DataBuf` where each value is preceded by a
// length-determining label read from a separate `parentBuf`.
// It is used for types where each instance in the block is individually length-prefixed.
// `Out` is the target type of the values read.
type LabelBlockReader[Out any] struct {
	CommonState                           // Embeds DataBuf and default AfterNewRead.
	fromBytes          FromBytesFunc[Out] // Function to convert raw bytes to Out type.
	readNullTerminator bool               // Flag indicating if a null terminator should be read after the data.
}

// NewLabelBlockReader creates and returns a new LabelBlockReader.
// - dataBuf: The buffer from which the actual value data is read (corresponds to `this.buf` in some TS implementations).
// - fromBytes: A function to convert the raw bytes (read according to the label) into the target type `Out`.
// - readNullTerminator: If true, an extra null byte is read and discarded from `dataBuf` after reading the value's data (used for null-terminated strings).
func NewLabelBlockReader[Out any](dataBuf buf.Read, fromBytes FromBytesFunc[Out], readNullTerminator bool) *LabelBlockReader[Out] {
	return &LabelBlockReader[Out]{
		CommonState:        CommonState{DataBuf: dataBuf},
		fromBytes:          fromBytes,
		readNullTerminator: readNullTerminator,
	}
}

// Read decodes a single value. It first reads a label from `parentBuf` to determine the length
// of the data. Then, it reads that many bytes from its internal `DataBuf`, converts these bytes
// to the `Out` type using `fromBytes`, and optionally reads a null terminator.
// It returns the decoded value and an error if any step fails.
// This reader expects the label to indicate a length; it cannot handle backreferences, null, absent, or error labels directly.
func (r *LabelBlockReader[Out]) Read(parentBuf buf.Read) (Out, error) {
	var zero Out // Zero value of Out to return on error

	l, err := label.Read(parentBuf)
	if err != nil {
		return zero, fmt.Errorf("LabelBlockReader: failed to read label from parentBuf: %w", err)
	}

	switch l.Kind() {
	case label.LabelKindBackreference:
		return zero, fmt.Errorf("programmer error: LabelBlockReader: this type must not use backreferences (label: %s)", l)
	case label.LabelKindLength:
		lengthVal := l.Value().Int64()
		if lengthVal < 0 {
			// label.IsLength() implies non-negative, so this should be an internal inconsistency.
			return zero, fmt.Errorf("programmer error: LabelBlockReader: negative length (%d) from label %s which has KindLength", lengthVal, l)
		}

		if r.DataBuf == nil {
			return zero, fmt.Errorf("programmer error: LabelBlockReader: DataBuf is nil")
		}

		var bytesToRead []byte
		if lengthVal == 0 {
			bytesToRead = []byte{} // Empty slice for zero length
		} else {
			bytesToRead = make([]byte, int(lengthVal)) // Allocate slice for data
			n, readErr := r.DataBuf.Read(bytesToRead)
			if readErr != nil && readErr != io.EOF { // Any error other than EOF is immediately problematic
				return zero, fmt.Errorf("LabelBlockReader: DataBuf.Read failed for %d bytes: %w", lengthVal, readErr)
			}
			// If EOF was encountered or no error, ensure all requested bytes were actually read.
			if n != int(lengthVal) {
				// This indicates a short read, which is an error here.
				return zero, fmt.Errorf("LabelBlockReader: expected to read %d bytes from DataBuf, but read %d (read error: %v)", lengthVal, n, readErr)
			}
		}

		if r.readNullTerminator {
			_, err := r.DataBuf.ReadByte()
			if err != nil {
				return zero, fmt.Errorf("LabelBlockReader: failed to read null terminator: %w", err)
			}
		}

		value := r.fromBytes(bytesToRead)
		r.AfterNewRead() // Call the AfterNewRead hook (defined on CommonState by default)
		return value, nil
	case label.LabelKindNull:
		return zero, fmt.Errorf("programmer error: LabelBlockReader: reader cannot handle null labels (label: %s)", l)
	case label.LabelKindAbsent:
		return zero, fmt.Errorf("programmer error: LabelBlockReader: reader cannot handle absent labels (label: %s)", l)
	case label.LabelKindError:
		return zero, fmt.Errorf("programmer error: LabelBlockReader: reader cannot handle error labels (label: %s)", l)
	default:
		return zero, fmt.Errorf("programmer error: LabelBlockReader: unhandled label kind %s (label: %s)", l.Kind(), l)
	}
}

// DeduplicatingLabelBlockReader is similar to LabelBlockReader but adds support for deduplication.
// It reads values from its `DataBuf` that are preceded by a label (read from `parentBuf`).
// This label can either specify the length of a new value or be a backreference to a previously read value.
// New values are stored internally to be targets for future backreferences.
// `Out` is the target type of the values read.
type DeduplicatingLabelBlockReader[Out any] struct {
	CommonState                           // Embeds DataBuf and default AfterNewRead.
	fromBytes          FromBytesFunc[Out] // Function to convert raw bytes to Out type.
	values             []Out              // Stores previously seen values for backreferencing.
	readNullTerminator bool               // Flag indicating if a null terminator should be read after new data.
}

// NewDeduplicatingLabelBlockReader creates and returns a new DeduplicatingLabelBlockReader.
// - dataBuf: The buffer for reading actual value data.
// - fromBytes: Function to convert raw bytes to the `Out` type.
// - values: An initial slice of values; typically empty, will be populated as new unique values are read.
// - readNullTerminator: If true, an extra null byte is read from `dataBuf` after new string data.
func NewDeduplicatingLabelBlockReader[Out any](dataBuf buf.Read, fromBytes FromBytesFunc[Out], readNullTerminator bool) *DeduplicatingLabelBlockReader[Out] {
	return &DeduplicatingLabelBlockReader[Out]{
		CommonState:        CommonState{DataBuf: dataBuf},
		fromBytes:          fromBytes,
		values:             make([]Out, 0), // Initialize empty slice for values
		readNullTerminator: readNullTerminator,
	}
}

// Read decodes a single value, handling potential backreferences.
// It reads a label from `parentBuf`. If the label is a backreference, it returns the previously stored value.
// If the label indicates a length, it reads the data from `DataBuf`, converts it, stores it for future
// backreferences, and returns it. Optionally reads a null terminator for new string data.
// Returns the decoded value and an error if any step fails.
// This reader cannot handle null, absent, or error labels directly.
func (r *DeduplicatingLabelBlockReader[Out]) Read(parentBuf buf.Read) (Out, error) {
	var zero Out

	l, err := label.Read(parentBuf)
	if err != nil {
		return zero, fmt.Errorf("DeduplicatingLabelBlockReader: failed to read label from parentBuf: %w", err)
	}

	switch l.Kind() {
	case label.LabelKindBackreference:
		offset, offsetErr := l.ToOffset() // Converts label to backreference offset
		if offsetErr != nil {
			// l.ToOffset() already validates if it's a proper backreference kind.
			return zero, fmt.Errorf("DeduplicatingLabelBlockReader: invalid backreference label %s: %w", l, offsetErr)
		}

		idx := int(offset)
		if idx < 0 || idx >= len(r.values) {
			return zero, fmt.Errorf("DeduplicatingLabelBlockReader: invalid backreference offset %d (label: %s), current values count: %d", offset, l, len(r.values))
		}
		// The TS `if (value == undefined)` check is implicitly handled by Go's slice indexing:
		// if the index is valid, a value (possibly zero-value) exists.
		value := r.values[idx]
		return value, nil

	case label.LabelKindLength:
		lengthVal := l.Value().Int64()
		if lengthVal < 0 {
			return zero, fmt.Errorf("programmer error: DeduplicatingLabelBlockReader: negative length (%d) from label %s which has KindLength", lengthVal, l)
		}

		if r.DataBuf == nil {
			return zero, fmt.Errorf("programmer error: DeduplicatingLabelBlockReader: DataBuf is nil")
		}

		var bytesToRead []byte
		if lengthVal == 0 {
			bytesToRead = []byte{}
		} else {
			bytesToRead = make([]byte, int(lengthVal))
			n, readErr := r.DataBuf.Read(bytesToRead)
			if readErr != nil && readErr != io.EOF {
				return zero, fmt.Errorf("DeduplicatingLabelBlockReader: DataBuf.Read failed for %d bytes: %w", lengthVal, readErr)
			}
			if n != int(lengthVal) {
				return zero, fmt.Errorf("DeduplicatingLabelBlockReader: expected to read %d bytes from DataBuf, but read %d (read error: %v)", lengthVal, n, readErr)
			}
		}

		if r.readNullTerminator {
			nul, err := r.DataBuf.ReadByte()
			if err != nil {
				return zero, fmt.Errorf("DeduplicatingLabelBlockReader: failed to read null terminator: %w", err)
			}
			if nul != '\000' {
				return zero, fmt.Errorf("DeduplicatingLabelBlockReader: null terminator was not null: %b", nul)
			}
		}

		value := r.fromBytes(bytesToRead)
		r.values = append(r.values, value) // Store the new value for future backreferences
		r.AfterNewRead()
		return value, nil

	case label.LabelKindNull:
		return zero, fmt.Errorf("programmer error: DeduplicatingLabelBlockReader: reader cannot handle null labels (label: %s)", l)
	case label.LabelKindAbsent:
		return zero, fmt.Errorf("programmer error: DeduplicatingLabelBlockReader: reader cannot handle absent labels (label: %s)", l)
	case label.LabelKindError:
		return zero, fmt.Errorf("programmer error: DeduplicatingLabelBlockReader: reader cannot handle error labels (label: %s)", l)
	default:
		return zero, fmt.Errorf("programmer error: DeduplicatingLabelBlockReader: unhandled label kind %s (label: %s)", l.Kind(), l)
	}
}

// FixedSizeBlockReader reads values of a predetermined fixed byte length from its `DataBuf`.
// It does not use labels from a `parentBuf` to determine length, as the length is constant.
// `Out` is the target type of the values read.
type FixedSizeBlockReader[Out any] struct {
	CommonState                    // Embeds DataBuf.
	fromBytes   FromBytesFunc[Out] // Function to convert raw bytes to Out type.
	ByteLength  int                // The fixed number of bytes for each value.
}

// NewFixedSizeBlockReader creates and returns a new FixedSizeBlockReader.
// - dataBuf: The buffer from which fixed-size data chunks are read.
// - fromBytes: Function to convert the fixed-size byte chunks to the `Out` type.
// - byteLength: The exact number of bytes that constitutes one value. Must be non-negative, or it panics.
func NewFixedSizeBlockReader[Out any](dataBuf buf.Read, fromBytes FromBytesFunc[Out], byteLength int) *FixedSizeBlockReader[Out] {
	if byteLength < 0 {
		panic(fmt.Sprintf("FixedSizeBlockReader: byteLength cannot be negative, got %d", byteLength))
	}
	return &FixedSizeBlockReader[Out]{
		CommonState: CommonState{DataBuf: dataBuf},
		fromBytes:   fromBytes,
		ByteLength:  byteLength,
	}
}

// Read decodes a fixed-size block from r.DataBuf.
// parentBuf is ignored by this implementation but included in the signature to potentially
// match a common interface derived from the abstract BlockReader<Out>.
func (r *FixedSizeBlockReader[Out]) Read(parentBuf buf.Read) (Out, error) {
	_ = parentBuf // Explicitly acknowledge and ignore parentBuf.
	var zero Out

	if r.ByteLength < 0 {
		// This state should ideally be prevented by constructor validation.
		return zero, fmt.Errorf("FixedSizeBlockReader: invalid ByteLength %d", r.ByteLength)
	}
	if r.DataBuf == nil {
		return zero, fmt.Errorf("programmer error: FixedSizeBlockReader: DataBuf is nil")
	}

	var bytesToRead []byte
	if r.ByteLength == 0 {
		bytesToRead = []byte{}
	} else {
		bytesToRead = make([]byte, r.ByteLength)
		n, readErr := r.DataBuf.Read(bytesToRead)
		if readErr != nil && readErr != io.EOF {
			return zero, fmt.Errorf("FixedSizeBlockReader: DataBuf.Read failed for %d bytes: %w", r.ByteLength, readErr)
		}
		if n != r.ByteLength {
			return zero, fmt.Errorf("FixedSizeBlockReader: expected to read %d bytes from DataBuf, but read %d (read error: %v)", r.ByteLength, n, readErr)
		}
	}

	// The TS comment `// aligned reads` for `new Uint8Array(bytes, bytes.byteOffset, this.byteLength)`
	// is generally not a direct concern for Go's `make([]byte, ...)` which provides a new slice.
	value := r.fromBytes(bytesToRead)
	// No AfterNewRead call in the original TS for this reader.
	return value, nil
}

// UnlabeledVarIntBlockReader reads varint-encoded integers directly from its `DataBuf`.
// It does not use labels from a `parentBuf` as varints are self-delimiting.
// This reader is specifically for `int64` values after ZigZag decoding.
type UnlabeledVarIntBlockReader struct {
	CommonState // Embeds DataBuf.
}

// NewUnlabeledVarIntBlockReader creates and returns a new UnlabeledVarIntBlockReader.
// - dataBuf: The buffer from which varint-encoded data is read.
func NewUnlabeledVarIntBlockReader(dataBuf buf.Read) *UnlabeledVarIntBlockReader {
	return &UnlabeledVarIntBlockReader{
		CommonState: CommonState{DataBuf: dataBuf},
	}
}

// Read decodes a single varint-encoded integer from `DataBuf` (ignoring `parentBuf`).
// It performs ZigZag decoding to convert the unsigned varint read from the buffer
// into a signed `int64`.
// Returns the decoded `int64` and an error if reading or decoding fails.
func (r *UnlabeledVarIntBlockReader) Read(parentBuf buf.Read) (int64, error) {
	_ = parentBuf // Explicitly acknowledge and ignore parentBuf.

	if r.DataBuf == nil {
		return 0, fmt.Errorf("programmer error: UnlabeledVarIntBlockReader: DataBuf is nil")
	}

	l, err := label.Read(r.DataBuf) // Label (varint) read from r.DataBuf
	if err != nil {
		return 0, fmt.Errorf("UnlabeledVarIntBlockReader: failed to read label from DataBuf: %w", err)
	}

	// TS: `Number(Label.read(this.buf))`. This implies getting the numerical value of the label.
	// `l.Value()` returns a *big.Int. Convert to int64.
	// `IsInt64()` checks if the value fits in int64 without loss.
	if !l.Value().IsInt64() {
		return 0, fmt.Errorf("UnlabeledVarIntBlockReader: label value '%s' cannot be represented as int64", l.Value().String())
	}
	// No AfterNewRead call in the original TS for this reader.
	return l.Value().Int64(), nil
}
