package object // import "github.com/SevereCloud/vksdk/v2/object"

import "github.com/vmihailenco/msgpack/v5"

// RawMessage is a raw encoded JSON or MessagePack value.
type RawMessage []byte

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	*m = append((*m)[0:0], data...)
	return nil
}

// EncodeMsgpack write m as the MessagePack encoding of m.
func (m RawMessage) EncodeMsgpack(enc *msgpack.Encoder) error {
	_, err := enc.Writer().Write(m)
	return err
}

// DecodeMsgpack sets *m to a copy of data.
func (m *RawMessage) DecodeMsgpack(dec *msgpack.Decoder) error {
	msg, err := dec.DecodeRaw()
	if err != nil {
		return err
	}

	*m = RawMessage(msg)

	return nil
}
