package peer_protocol

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"
)

type Decoder struct {
	R         *bufio.Reader
	Pool      *sync.Pool
	MaxLength Integer // TODO: Should this include the length header or not?
}

// io.EOF is returned if the source terminates cleanly on a message boundary.
func (d *Decoder) Decode(msg *Message) (err error) {
	var length Integer
	err = length.Read(d.R)
	if err != nil {
		return fmt.Errorf("reading message length: %w", err)
	}
	if length > d.MaxLength {
		return errors.New("message too long")
	}
	if length == 0 {
		msg.Keepalive = true
		return
	}
	r := d.R
	readByte := func() (byte, error) {
		length--
		return d.R.ReadByte()
	}
	c, err := readByte()
	if err != nil {
		return
	}
	msg.Type = MessageType(c)
	switch msg.Type {
	case Choke, Unchoke, Interested, NotInterested, HaveAll, HaveNone:
	case Have, AllowedFast, Suggest:
		length -= 4
		err = msg.Index.Read(r)
	case Request, Cancel, Reject:
		for _, data := range []*Integer{&msg.Index, &msg.Begin, &msg.Length} {
			err = data.Read(r)
			if err != nil {
				break
			}
		}
		length -= 12
	case Bitfield:
		b := make([]byte, length)
		_, err = io.ReadFull(r, b)
		length = 0
		msg.Bitfield = unmarshalBitfield(b)
	case Piece:
		for _, pi := range []*Integer{&msg.Index, &msg.Begin} {
			err := pi.Read(r)
			if err != nil {
				return err
			}
		}
		length -= 8
		dataLen := int64(length)
		msg.Piece = *d.Pool.Get().(*[]byte)
		if int64(cap(msg.Piece)) < dataLen {
			return errors.New("piece data longer than expected")
		}
		msg.Piece = msg.Piece[:dataLen]
		_, err := io.ReadFull(r, msg.Piece)
		if err != nil {
			return fmt.Errorf("reading piece data: %w", err)
		}
		length = 0
	case Extended:
		var b byte
		b, err = readByte()
		if err != nil {
			break
		}
		msg.ExtendedID = ExtensionNumber(b)
		msg.ExtendedPayload = make([]byte, length)
		_, err = io.ReadFull(r, msg.ExtendedPayload)
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		length = 0
	case Port:
		err = binary.Read(r, binary.BigEndian, &msg.Port)
		length -= 2
	default:
		err = fmt.Errorf("unknown message type %#v", c)
	}
	if err == nil && length != 0 {
		err = fmt.Errorf("%v unused bytes in message type %v", length, msg.Type)
	}
	return
}

func readByte(r io.Reader) (b byte, err error) {
	var arr [1]byte
	n, err := r.Read(arr[:])
	b = arr[0]
	if n == 1 {
		err = nil
		return
	}
	if err == nil {
		panic(err)
	}
	return
}

func unmarshalBitfield(b []byte) (bf []bool) {
	for _, c := range b {
		for i := 7; i >= 0; i-- {
			bf = append(bf, (c>>uint(i))&1 == 1)
		}
	}
	return
}
