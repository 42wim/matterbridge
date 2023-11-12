package udp

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Action int32

const (
	ActionConnect Action = iota
	ActionAnnounce
	ActionScrape
	ActionError

	ConnectRequestConnectionId = 0x41727101980

	// BEP 41
	optionTypeEndOfOptions = 0
	optionTypeNOP          = 1
	optionTypeURLData      = 2
)

type TransactionId = int32

type ConnectionId = int64

type ConnectionRequest struct {
	ConnectionId  ConnectionId
	Action        Action
	TransactionId TransactionId
}

type ConnectionResponse struct {
	ConnectionId ConnectionId
}

type ResponseHeader struct {
	Action        Action
	TransactionId TransactionId
}

type RequestHeader struct {
	ConnectionId  ConnectionId
	Action        Action
	TransactionId TransactionId
} // 16 bytes

type AnnounceResponseHeader struct {
	Interval int32
	Leechers int32
	Seeders  int32
}

type InfoHash = [20]byte

func marshal(data interface{}) (b []byte, err error) {
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, data)
	b = buf.Bytes()
	return
}

func mustMarshal(data interface{}) []byte {
	b, err := marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}

func Write(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}

func Read(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.BigEndian, data)
}
