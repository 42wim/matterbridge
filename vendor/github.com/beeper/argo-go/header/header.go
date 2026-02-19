package header

import (
	"fmt"

	"github.com/beeper/argo-go/pkg/bitset"
	"github.com/beeper/argo-go/pkg/buf"
)

// Public constants for header flags
const (
	HeaderInlineEverythingFlag      = 0
	HeaderSelfDescribingFlag        = 1
	HeaderOutOfBandFieldErrorsFlag  = 2
	HeaderSelfDescribingErrorsFlag  = 3
	HeaderNullTerminatedStringsFlag = 4
	HeaderNoDeduplicationFlag       = 5
	HeaderHasUserFlagsFlag          = 6
)

// Header represents the Argo message header.
type Header struct {
	flags     *bitset.BitSet
	userFlags *bitset.BitSet
}

// NewHeader creates a new Header.
func NewHeader() *Header {
	return &Header{
		flags: bitset.NewBitSet(),
	}
}

// GetFlag returns the boolean value of a given header flag.
func (h *Header) GetFlag(flag int) bool {
	if h.flags == nil {
		return false
	}
	return h.flags.GetBit(flag)
}

// SetFlag sets the boolean value of a given header flag.
func (h *Header) SetFlag(flag int, value bool) {
	if h.flags == nil {
		h.flags = bitset.NewBitSet()
	}
	if value {
		h.flags.SetBit(flag)
	} else {
		h.flags.UnsetBit(flag)
	}
}

// Read reads the header from the provided Read buffer.
// It updates the Header's internal state (flags, userFlags).
// It also advances the position of the buffer.
func (h *Header) Read(reader buf.Read) error {
	if reader == nil {
		return fmt.Errorf("reader is nil, cannot read header")
	}

	var vbs bitset.VarBitSet
	_, flags, err := vbs.Read(reader) // Read standard flags using VarBitSet.Read
	if err != nil {
		return fmt.Errorf("failed to read standard header flags: %w", err)
	}
	h.flags = flags

	// Check based on the just-read standard flags
	if h.GetFlag(HeaderHasUserFlagsFlag) { // Use new GetFlag method
		_, userFlags, err := vbs.Read(reader) // Read user flags using VarBitSet.Read
		if err != nil {
			return fmt.Errorf("failed to read user header flags: %w", err)
		}
		h.userFlags = userFlags
	} else {
		h.userFlags = nil // Ensure userFlags is nil if HasUserFlags is false
	}
	return nil
}

// Write writes the header to the provided Write buffer.
func (h *Header) Write(writer buf.Write) error {
	if writer == nil {
		return fmt.Errorf("writer is nil, cannot write header")
	}
	flagBytes, err := (&bitset.VarBitSet{}).Write(h.flags, 0)
	if err != nil {
		return fmt.Errorf("failed to write flags: %w", err)
	}

	var userFlagBytes []byte
	if h.GetFlag(HeaderHasUserFlagsFlag) { // Use new GetFlag method
		userFlagBytes, err = (&bitset.VarBitSet{}).Write(h.userFlags, 0)
		if err != nil {
			return fmt.Errorf("failed to write userFlags: %w", err)
		}
	}

	if _, err := writer.Write(flagBytes); err != nil {
		return fmt.Errorf("buffer write error for flags: %w", err)
	}
	if userFlagBytes != nil {
		if _, err := writer.Write(userFlagBytes); err != nil {
			return fmt.Errorf("buffer write error for userFlags: %w", err)
		}
	}
	return nil
}

// AsBytes serializes the header to a byte slice.
func (h *Header) AsBytes() ([]byte, error) {
	tempBuf := buf.NewBuf(0) // Use argo.Buf which implements Write
	err := h.Write(tempBuf)
	if err != nil {
		return nil, err // Propagate error from Write
	}
	return tempBuf.Bytes(), nil
}

// UserFlags returns the user flags BitSet.
func (h *Header) UserFlags() *bitset.BitSet {
	return h.userFlags
}

// SetUserFlags sets the user flags BitSet.
// This also sets or unsets the HeaderHasUserFlagsFlag bit accordingly.
func (h *Header) SetUserFlags(bs *bitset.BitSet) {
	h.userFlags = bs
	if h.flags == nil { // Ensure flags is initialized
		h.flags = bitset.NewBitSet()
	}
	hasUserFlagsSet := bs != nil && len(bs.Bytes()) > 0
	h.SetFlag(HeaderHasUserFlagsFlag, hasUserFlagsSet) // Use new SetFlag method
}
