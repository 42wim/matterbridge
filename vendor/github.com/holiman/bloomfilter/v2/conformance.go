// Package bloomfilter is face-meltingly fast, thread-safe,
// marshalable, unionable, probability- and
// optimal-size-calculating Bloom filter in go
//
// https://github.com/steakknife/bloomfilter
//
// Copyright © 2014, 2015, 2018 Barry Allard
//
// MIT license
//
package v2

import (
	"encoding"
	"encoding/gob"
	"encoding/json"
	"io"
)

// compile-time conformance tests
var (
	_ encoding.BinaryMarshaler   = (*Filter)(nil)
	_ encoding.BinaryUnmarshaler = (*Filter)(nil)
	_ io.ReaderFrom              = (*Filter)(nil)
	_ io.WriterTo                = (*Filter)(nil)
	_ gob.GobDecoder             = (*Filter)(nil)
	_ gob.GobEncoder             = (*Filter)(nil)
	_ json.Marshaler             = (*Filter)(nil)
	_ json.Unmarshaler           = (*Filter)(nil)
)
