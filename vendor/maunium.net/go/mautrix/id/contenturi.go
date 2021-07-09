// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package id

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	InvalidContentURI  = errors.New("invalid Matrix content URI")
	InputNotJSONString = errors.New("input doesn't look like a JSON string")
)

// ContentURIString is a string that's expected to be a Matrix content URI.
// It's useful for delaying the parsing of the content URI to move errors from the event content
// JSON parsing step to a later step where more appropriate errors can be produced.
type ContentURIString string

func (uriString ContentURIString) Parse() (ContentURI, error) {
	return ParseContentURI(string(uriString))
}

func (uriString ContentURIString) ParseOrIgnore() ContentURI {
	parsed, _ := ParseContentURI(string(uriString))
	return parsed
}

// ContentURI represents a Matrix content URI.
// https://matrix.org/docs/spec/client_server/r0.6.0#matrix-content-mxc-uris
type ContentURI struct {
	Homeserver string
	FileID     string
}

func MustParseContentURI(uri string) ContentURI {
	parsed, err := ParseContentURI(uri)
	if err != nil {
		panic(err)
	}
	return parsed
}

// ParseContentURI parses a Matrix content URI.
func ParseContentURI(uri string) (parsed ContentURI, err error) {
	if len(uri) == 0 {
		return
	} else if !strings.HasPrefix(uri, "mxc://") {
		err = InvalidContentURI
	} else if index := strings.IndexRune(uri[6:], '/'); index == -1 || index == len(uri)-7 {
		err = InvalidContentURI
	} else {
		parsed.Homeserver = uri[6 : 6+index]
		parsed.FileID = uri[6+index+1:]
	}
	return
}

var mxcBytes = []byte("mxc://")

func ParseContentURIBytes(uri []byte) (parsed ContentURI, err error) {
	if len(uri) == 0 {
		return
	} else if !bytes.HasPrefix(uri, mxcBytes) {
		err = InvalidContentURI
	} else if index := bytes.IndexRune(uri[6:], '/'); index == -1 || index == len(uri)-7 {
		err = InvalidContentURI
	} else {
		parsed.Homeserver = string(uri[6 : 6+index])
		parsed.FileID = string(uri[6+index+1:])
	}
	return
}

func (uri *ContentURI) UnmarshalJSON(raw []byte) (err error) {
	if string(raw) == "null" {
		*uri = ContentURI{}
		return nil
	} else if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return InputNotJSONString
	}
	parsed, err := ParseContentURIBytes(raw[1:len(raw)-1])
	if err != nil {
		return err
	}
	*uri = parsed
	return nil
}

func (uri *ContentURI) MarshalJSON() ([]byte, error) {
	if uri.IsEmpty() {
		return []byte("null"), nil
	}
	return json.Marshal(uri.String())
}

func (uri *ContentURI) UnmarshalText(raw []byte) (err error) {
	parsed, err := ParseContentURIBytes(raw)
	if err != nil {
		return err
	}
	*uri = parsed
	return nil
}

func (uri ContentURI) MarshalText() ([]byte, error) {
	if uri.IsEmpty() {
		return []byte(""), nil
	}
	return []byte(uri.String()), nil
}

func (uri *ContentURI) String() string {
	if uri.IsEmpty() {
		return ""
	}
	return fmt.Sprintf("mxc://%s/%s", uri.Homeserver, uri.FileID)
}

func (uri *ContentURI) CUString() ContentURIString {
	return ContentURIString(uri.String())
}

func (uri *ContentURI) IsEmpty() bool {
	return len(uri.Homeserver) == 0 || len(uri.FileID) == 0
}
