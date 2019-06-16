package charset

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

func init() {
	registerClass("ascii", fromASCII, toASCII)
}

const errorByte = '?'

type translateFromASCII bool

type codePointError struct {
	i       int
	cp      rune
	charset string
}

func (e *codePointError) Error() string {
	return fmt.Sprintf("Parse error at index %d: Code point %d is undefined in %s", e.i, e.cp, e.charset)
}

func (strict translateFromASCII) Translate(data []byte, eof bool) (int, []byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	for i, c := range data {
		if c > 0 && c < 128 {
			buf.WriteByte(c)
			if c < 32 && c != 10 && c != 13 && c != 9 {
				// badly formed
			}
		} else {
			if strict {
				return 0, nil, &codePointError{i, rune(c), "US-ASCII"}
			}
			buf.WriteRune(utf8.RuneError)
		}
	}
	return len(data), buf.Bytes(), nil
}

type translateToASCII bool

func (strict translateToASCII) Translate(data []byte, eof bool) (int, []byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	for _, c := range data {
		if c > 0 && c < 128 {
			buf.WriteByte(c)
		} else {
			buf.WriteByte(errorByte)
		}
	}
	return len(data), buf.Bytes(), nil
}

func fromASCII(arg string) (Translator, error) {
	return new(translateFromASCII), nil
}

func toASCII(arg string) (Translator, error) {
	return new(translateToASCII), nil
}
