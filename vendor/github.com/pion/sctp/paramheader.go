package sctp

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type paramHeader struct {
	typ paramType
	len int
	raw []byte
}

const (
	paramHeaderLength = 4
)

var (
	errParamHeaderTooShort                  = errors.New("param header too short")
	errParamHeaderSelfReportedLengthShorter = errors.New("param self reported length is shorter than header length")
	errParamHeaderSelfReportedLengthLonger  = errors.New("param self reported length is longer than header length")
	errParamHeaderParseFailed               = errors.New("failed to parse param type")
)

func (p *paramHeader) marshal() ([]byte, error) {
	paramLengthPlusHeader := paramHeaderLength + len(p.raw)

	rawParam := make([]byte, paramLengthPlusHeader)
	binary.BigEndian.PutUint16(rawParam[0:], uint16(p.typ))
	binary.BigEndian.PutUint16(rawParam[2:], uint16(paramLengthPlusHeader))
	copy(rawParam[paramHeaderLength:], p.raw)

	return rawParam, nil
}

func (p *paramHeader) unmarshal(raw []byte) error {
	if len(raw) < paramHeaderLength {
		return errParamHeaderTooShort
	}

	paramLengthPlusHeader := binary.BigEndian.Uint16(raw[2:])
	if int(paramLengthPlusHeader) < paramHeaderLength {
		return fmt.Errorf("%w: param self reported length (%d) shorter than header length (%d)", errParamHeaderSelfReportedLengthShorter, int(paramLengthPlusHeader), paramHeaderLength)
	}
	if len(raw) < int(paramLengthPlusHeader) {
		return fmt.Errorf("%w: param length (%d) shorter than its self reported length (%d)", errParamHeaderSelfReportedLengthLonger, len(raw), int(paramLengthPlusHeader))
	}

	typ, err := parseParamType(raw[0:])
	if err != nil {
		return fmt.Errorf("%w: %v", errParamHeaderParseFailed, err)
	}
	p.typ = typ
	p.raw = raw[paramHeaderLength:paramLengthPlusHeader]
	p.len = int(paramLengthPlusHeader)

	return nil
}

func (p *paramHeader) length() int {
	return p.len
}

// String makes paramHeader printable
func (p paramHeader) String() string {
	return fmt.Sprintf("%s (%d): %s", p.typ, p.len, hex.Dump(p.raw))
}
