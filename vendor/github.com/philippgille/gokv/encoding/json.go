package encoding

import (
	"encoding/json"
)

// JSONcodec encodes/decodes Go values to/from JSON.
// You can use encoding.JSON instead of creating an instance of this struct.
type JSONcodec struct{}

// Marshal encodes a Go value to JSON.
func (c JSONcodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal decodes a JSON value into a Go value.
func (c JSONcodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
