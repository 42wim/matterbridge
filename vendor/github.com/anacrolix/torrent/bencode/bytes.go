package bencode

import (
	"errors"
)

type Bytes []byte

var (
	_ Unmarshaler = (*Bytes)(nil)
	_ Marshaler   = (*Bytes)(nil)
	_ Marshaler   = Bytes{}
)

func (me *Bytes) UnmarshalBencode(b []byte) error {
	*me = append([]byte(nil), b...)
	return nil
}

func (me Bytes) MarshalBencode() ([]byte, error) {
	if len(me) == 0 {
		return nil, errors.New("marshalled Bytes should not be zero-length")
	}
	return me, nil
}
