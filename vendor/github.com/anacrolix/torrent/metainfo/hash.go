package metainfo

import (
	"crypto/sha1"
	"encoding"
	"encoding/hex"
	"fmt"
)

const HashSize = 20

// 20-byte SHA1 hash used for info and pieces.
type Hash [HashSize]byte

var _ fmt.Formatter = (*Hash)(nil)

func (h Hash) Format(f fmt.State, c rune) {
	// TODO: I can't figure out a nice way to just override the 'x' rune, since it's meaningless
	// with the "default" 'v', or .String() already returning the hex.
	f.Write([]byte(h.HexString()))
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) AsString() string {
	return string(h[:])
}

func (h Hash) String() string {
	return h.HexString()
}

func (h Hash) HexString() string {
	return fmt.Sprintf("%x", h[:])
}

func (h *Hash) FromHexString(s string) (err error) {
	if len(s) != 2*HashSize {
		err = fmt.Errorf("hash hex string has bad length: %d", len(s))
		return
	}
	n, err := hex.Decode(h[:], []byte(s))
	if err != nil {
		return
	}
	if n != HashSize {
		panic(n)
	}
	return
}

var (
	_ encoding.TextUnmarshaler = (*Hash)(nil)
	_ encoding.TextMarshaler   = Hash{}
)

func (h *Hash) UnmarshalText(b []byte) error {
	return h.FromHexString(string(b))
}

func (h Hash) MarshalText() (text []byte, err error) {
	return []byte(h.HexString()), nil
}

func NewHashFromHex(s string) (h Hash) {
	err := h.FromHexString(s)
	if err != nil {
		panic(err)
	}
	return
}

func HashBytes(b []byte) (ret Hash) {
	hasher := sha1.New()
	hasher.Write(b)
	copy(ret[:], hasher.Sum(nil))
	return
}
