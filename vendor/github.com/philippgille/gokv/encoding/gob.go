package encoding

import (
	"bytes"
	"encoding/gob"
)

// GobCodec encodes/decodes Go values to/from gob.
// You can use encoding.Gob instead of creating an instance of this struct.
type GobCodec struct{}

// Marshal encodes a Go value to gob.
func (c GobCodec) Marshal(v interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Unmarshal decodes a gob value into a Go value.
func (c GobCodec) Unmarshal(data []byte, v interface{}) error {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(v)
}
