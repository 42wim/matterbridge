package protobuf

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/status-im/mvds/state"
)

// ID creates the MessageID for a Message
func (m Message) ID() state.MessageID {
	t := make([]byte, 8)
	binary.LittleEndian.PutUint64(t, uint64(m.Timestamp))

	b := append([]byte("MESSAGE_ID"), m.GroupId[:]...)
	b = append(b, t...)
	b = append(b, m.Body...)

	return sha256.Sum256(b)
}
