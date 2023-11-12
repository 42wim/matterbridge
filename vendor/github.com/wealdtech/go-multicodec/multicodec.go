// Copyright © 2019 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package multicodec provides the ability to add and remove codec prefixes from data.
package multicodec

import (
	"encoding/binary"
	"fmt"
)

// AddCodec adds a codec prefix to a byte array.
// It returns a new byte array with the relevant codec prefixed.
func AddCodec(name string, data []byte) ([]byte, error) {
	id, err := ID(name)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, binary.MaxVarintLen64)
	size := binary.PutUvarint(buf, id)
	res := make([]byte, size+len(data))
	copy(res, buf[0:size])
	copy(res[size:], data)
	return res, nil
}

// RemoveCodec removes a codec prefix from a byte array.
// It returns a slice of the input byte array without the codec prefix, along with the ID of the codec that has been removed.
func RemoveCodec(data []byte) ([]byte, uint64, error) {
	id, size := binary.Uvarint(data)
	if id == 0 {
		return nil, 0, fmt.Errorf("failed to find codec prefix to remove")
	}
	return data[size:], id, nil
}

// GetCodec returns the ID of the codec prefix from a byte array
func GetCodec(data []byte) (uint64, error) {
	id, _ := binary.Uvarint(data)
	if id == 0 {
		return 0, fmt.Errorf("failed to find codec prefix to remove")
	}
	return id, nil
}

// IsCodec returns true if the data is of the supplied codec prefix, otherwise false
func IsCodec(name string, data []byte) bool {
	id, err := GetCodec(data)
	if err != nil {
		return false
	}
	codecName, err := Name(id)
	if err != nil {
		// Failed to obtain name for ID
		return false
	}
	return codecName == name
}

// ID obtains the ID of a codec from its name
func ID(name string) (uint64, error) {
	codec, exists := codecs[name]
	if !exists {
		return 0, fmt.Errorf("unknown name %s", name)
	}
	return codec.id, nil
}

// MustID obtains the ID of a codec from its name, panicking if not present.
func MustID(name string) uint64 {
	codec, exists := codecs[name]
	if !exists {
		panic(fmt.Errorf("unknown name %s", name))
	}
	return codec.id
}

// Name obtains the name of a codec from its ID
func Name(id uint64) (string, error) {
	codec, exists := reverseCodecs[id]
	if !exists {
		return "", fmt.Errorf("unknown ID 0x%x", id)
	}
	return codec.name, nil
}

// MustName obtains the name of a codec from its ID, panicking if not present.
func MustName(id uint64) string {
	codec, exists := reverseCodecs[id]
	if !exists {
		panic(fmt.Errorf("unknown ID 0x%x", id))
	}
	return codec.name
}
