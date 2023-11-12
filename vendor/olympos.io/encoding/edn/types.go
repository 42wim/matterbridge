// Copyright 2015-2017 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
)

// RawMessage is a raw encoded, but valid, EDN value. It implements Marshaler
// and Unmarshaler and can be used to delay EDN decoding or precompute an EDN
// encoding.
type RawMessage []byte

// MarshalEDN returns m as the EDN encoding of m.
func (m RawMessage) MarshalEDN() ([]byte, error) {
	if m == nil {
		return []byte("nil"), nil
	}
	return m, nil
}

// UnmarshalEDN sets *m to a copy of data.
func (m *RawMessage) UnmarshalEDN(data []byte) error {
	if m == nil {
		return errors.New("edn.RawMessage: UnmarshalEDN on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

// A Keyword is an EDN keyword without : prepended in front.
type Keyword string

func (k Keyword) String() string {
	return fmt.Sprintf(":%s", string(k))
}

func (k Keyword) MarshalEDN() ([]byte, error) {
	return []byte(k.String()), nil
}

// A Symbol is an EDN symbol.
type Symbol string

func (s Symbol) String() string {
	return string(s)
}

func (s Symbol) MarshalEDN() ([]byte, error) {
	return []byte(s), nil
}

// A Tag is a tagged value. The Tagname represents the name of the tag, and the
// Value is the value of the element.
type Tag struct {
	Tagname string
	Value   interface{}
}

func (t Tag) String() string {
	return fmt.Sprintf("#%s %v", t.Tagname, t.Value)
}

func (t Tag) MarshalEDN() ([]byte, error) {
	str := []byte(fmt.Sprintf(`#%s `, t.Tagname))
	b, err := Marshal(t.Value)
	if err != nil {
		return nil, err
	}
	return append(str, b...), nil
}

func (t *Tag) UnmarshalEDN(bs []byte) error {
	// read actual tag, using the lexer.
	var lex lexer
	lex.reset()
	buf := bufio.NewReader(bytes.NewBuffer(bs))
	start := 0
	endTag := 0
tag:
	for {
		r, rlen, err := buf.ReadRune()
		if err != nil {
			return err
		}

		ls := lex.state(r)
		switch ls {
		case lexIgnore:
			start += rlen
			endTag += rlen
		case lexError:
			return lex.err
		case lexEndPrev:
			break tag
		case lexEnd: // unexpected, assuming tag which is not ending with lexEnd
			return errUnexpected
		case lexCont:
			endTag += rlen
		}
	}
	t.Tagname = string(bs[start+1 : endTag])
	return Unmarshal(bs[endTag:], &t.Value)
}

// A Rune type is a wrapper for a rune. It can be used to encode runes as
// characters instead of int32 values.
type Rune rune

func (r Rune) MarshalEDN() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 10))
	encodeRune(buf, rune(r))
	return buf.Bytes(), nil
}

func encodeRune(buf *bytes.Buffer, r rune) {
	const hex = "0123456789abcdef"
	if !isWhitespace(r) {
		buf.WriteByte('\\')
		buf.WriteRune(r)
	} else {
		switch r {
		case '\b':
			buf.WriteString(`\backspace`)
		case '\f':
			buf.WriteString(`\formfeed`)
		case '\n':
			buf.WriteString(`\newline`)
		case '\r':
			buf.WriteString(`\return`)
		case '\t':
			buf.WriteString(`\tab`)
		case ' ':
			buf.WriteString(`\space`)
		default:
			buf.WriteByte('\\')
			buf.WriteByte('u')
			buf.WriteByte(hex[r>>12&0xF])
			buf.WriteByte(hex[r>>8&0xF])
			buf.WriteByte(hex[r>>4&0xF])
			buf.WriteByte(hex[r&0xF])
		}
	}
}
