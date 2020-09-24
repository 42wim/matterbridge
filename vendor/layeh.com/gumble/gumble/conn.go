package gumble

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"layeh.com/gumble/gumble/MumbleProto"
	"layeh.com/gumble/gumble/varint"
)

// DefaultPort is the default port on which Mumble servers listen.
const DefaultPort = 64738

// Conn represents a control protocol connection to a Mumble client/server.
type Conn struct {
	sync.Mutex
	net.Conn

	MaximumPacketBytes int
	Timeout            time.Duration

	buffer []byte
}

// NewConn creates a new Conn with the given net.Conn.
func NewConn(conn net.Conn) *Conn {
	return &Conn{
		Conn:               conn,
		Timeout:            time.Second * 20,
		MaximumPacketBytes: 1024 * 1024 * 10,
	}
}

// ReadPacket reads a packet from the server. Returns the packet type, the
// packet data, and nil on success.
//
// This function should only be called by a single go routine.
func (c *Conn) ReadPacket() (uint16, []byte, error) {
	c.Conn.SetReadDeadline(time.Now().Add(c.Timeout))
	var header [6]byte
	if _, err := io.ReadFull(c.Conn, header[:]); err != nil {
		return 0, nil, err
	}
	pType := binary.BigEndian.Uint16(header[:])
	pLength := binary.BigEndian.Uint32(header[2:])
	pLengthInt := int(pLength)
	if pLengthInt > c.MaximumPacketBytes {
		return 0, nil, errors.New("gumble: packet larger than maximum allowed size")
	}
	if pLengthInt > len(c.buffer) {
		c.buffer = make([]byte, pLengthInt)
	}
	if _, err := io.ReadFull(c.Conn, c.buffer[:pLengthInt]); err != nil {
		return 0, nil, err
	}
	return pType, c.buffer[:pLengthInt], nil
}

// WriteAudio writes an audio packet to the connection.
func (c *Conn) WriteAudio(format, target byte, sequence int64, final bool, data []byte, X, Y, Z *float32) error {
	var buff [1 + varint.MaxVarintLen*2]byte
	buff[0] = (format << 5) | target
	n := varint.Encode(buff[1:], sequence)
	if n == 0 {
		return errors.New("gumble: varint out of range")
	}
	l := int64(len(data))
	if final {
		l |= 0x2000
	}
	m := varint.Encode(buff[1+n:], l)
	if m == 0 {
		return errors.New("gumble: varint out of range")
	}
	header := buff[:1+n+m]

	var positionalLength int
	if X != nil {
		positionalLength = 3 * 4
	}

	c.Lock()
	defer c.Unlock()

	if err := c.writeHeader(1, uint32(len(header)+len(data)+positionalLength)); err != nil {
		return err
	}
	if _, err := c.Conn.Write(header); err != nil {
		return err
	}
	if _, err := c.Conn.Write(data); err != nil {
		return err
	}

	if positionalLength > 0 {
		if err := binary.Write(c.Conn, binary.LittleEndian, *X); err != nil {
			return err
		}
		if err := binary.Write(c.Conn, binary.LittleEndian, *Y); err != nil {
			return err
		}
		if err := binary.Write(c.Conn, binary.LittleEndian, *Z); err != nil {
			return err
		}
	}

	return nil
}

// WritePacket writes a data packet of the given type to the connection.
func (c *Conn) WritePacket(ptype uint16, data []byte) error {
	c.Lock()
	defer c.Unlock()
	if err := c.writeHeader(uint16(ptype), uint32(len(data))); err != nil {
		return err
	}
	if _, err := c.Conn.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Conn) writeHeader(pType uint16, pLength uint32) error {
	var header [6]byte
	binary.BigEndian.PutUint16(header[:], pType)
	binary.BigEndian.PutUint32(header[2:], pLength)
	if _, err := c.Conn.Write(header[:]); err != nil {
		return err
	}
	return nil
}

// WriteProto writes a protocol buffer message to the connection.
func (c *Conn) WriteProto(message proto.Message) error {
	var protoType uint16
	switch message.(type) {
	case *MumbleProto.Version:
		protoType = 0
	case *MumbleProto.Authenticate:
		protoType = 2
	case *MumbleProto.Ping:
		protoType = 3
	case *MumbleProto.Reject:
		protoType = 4
	case *MumbleProto.ServerSync:
		protoType = 5
	case *MumbleProto.ChannelRemove:
		protoType = 6
	case *MumbleProto.ChannelState:
		protoType = 7
	case *MumbleProto.UserRemove:
		protoType = 8
	case *MumbleProto.UserState:
		protoType = 9
	case *MumbleProto.BanList:
		protoType = 10
	case *MumbleProto.TextMessage:
		protoType = 11
	case *MumbleProto.PermissionDenied:
		protoType = 12
	case *MumbleProto.ACL:
		protoType = 13
	case *MumbleProto.QueryUsers:
		protoType = 14
	case *MumbleProto.CryptSetup:
		protoType = 15
	case *MumbleProto.ContextActionModify:
		protoType = 16
	case *MumbleProto.ContextAction:
		protoType = 17
	case *MumbleProto.UserList:
		protoType = 18
	case *MumbleProto.VoiceTarget:
		protoType = 19
	case *MumbleProto.PermissionQuery:
		protoType = 20
	case *MumbleProto.CodecVersion:
		protoType = 21
	case *MumbleProto.UserStats:
		protoType = 22
	case *MumbleProto.RequestBlob:
		protoType = 23
	case *MumbleProto.ServerConfig:
		protoType = 24
	case *MumbleProto.SuggestConfig:
		protoType = 25
	default:
		return errors.New("gumble: unknown message type")
	}
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return c.WritePacket(protoType, data)
}
