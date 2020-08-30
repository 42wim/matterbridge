package gomatrix

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const lowerhex = "0123456789abcdef"

// encode the given byte using quoted-printable encoding (e.g "=2f")
// and writes it to the buffer
// See https://golang.org/src/mime/quotedprintable/writer.go
func encode(buf *bytes.Buffer, b byte) {
	buf.WriteByte('=')
	buf.WriteByte(lowerhex[b>>4])
	buf.WriteByte(lowerhex[b&0x0f])
}

// escape the given alpha character and writes it to the buffer
func escape(buf *bytes.Buffer, b byte) {
	buf.WriteByte('_')
	if b == '_' {
		buf.WriteByte('_') // another _
	} else {
		buf.WriteByte(b + 0x20) // ASCII shift A-Z to a-z
	}
}

func shouldEncode(b byte) bool {
	return b != '-' && b != '.' && b != '_' && !(b >= '0' && b <= '9') && !(b >= 'a' && b <= 'z') && !(b >= 'A' && b <= 'Z')
}

func shouldEscape(b byte) bool {
	return (b >= 'A' && b <= 'Z') || b == '_'
}

func isValidByte(b byte) bool {
	return isValidEscapedChar(b) || (b >= '0' && b <= '9') || b == '.' || b == '=' || b == '-'
}

func isValidEscapedChar(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z')
}

// EncodeUserLocalpart encodes the given string into Matrix-compliant user ID localpart form.
// See http://matrix.org/docs/spec/intro.html#mapping-from-other-character-sets
//
// This returns a string with only the characters "a-z0-9._=-". The uppercase range A-Z
// are encoded using leading underscores ("_"). Characters outside the aforementioned ranges
// (including literal underscores ("_") and equals ("=")) are encoded as UTF8 code points (NOT NCRs)
// and converted to lower-case hex with a leading "=". For example:
//   Alph@Bet_50up  => _alph=40_bet=5f50up
func EncodeUserLocalpart(str string) string {
	strBytes := []byte(str)
	var outputBuffer bytes.Buffer
	for _, b := range strBytes {
		if shouldEncode(b) {
			encode(&outputBuffer, b)
		} else if shouldEscape(b) {
			escape(&outputBuffer, b)
		} else {
			outputBuffer.WriteByte(b)
		}
	}
	return outputBuffer.String()
}

// DecodeUserLocalpart decodes the given string back into the original input string.
// Returns an error if the given string is not a valid user ID localpart encoding.
// See http://matrix.org/docs/spec/intro.html#mapping-from-other-character-sets
//
// This decodes quoted-printable bytes back into UTF8, and unescapes casing. For
// example:
//  _alph=40_bet=5f50up  =>  Alph@Bet_50up
// Returns an error if the input string contains characters outside the
// range "a-z0-9._=-", has an invalid quote-printable byte (e.g. not hex), or has
// an invalid _ escaped byte (e.g. "_5").
func DecodeUserLocalpart(str string) (string, error) {
	strBytes := []byte(str)
	var outputBuffer bytes.Buffer
	for i := 0; i < len(strBytes); i++ {
		b := strBytes[i]
		if !isValidByte(b) {
			return "", fmt.Errorf("Byte pos %d: Invalid byte", i)
		}

		if b == '_' { // next byte is a-z and should be upper-case or is another _ and should be a literal _
			if i+1 >= len(strBytes) {
				return "", fmt.Errorf("Byte pos %d: expected _[a-z_] encoding but ran out of string", i)
			}
			if !isValidEscapedChar(strBytes[i+1]) { // invalid escaping
				return "", fmt.Errorf("Byte pos %d: expected _[a-z_] encoding", i)
			}
			if strBytes[i+1] == '_' {
				outputBuffer.WriteByte('_')
			} else {
				outputBuffer.WriteByte(strBytes[i+1] - 0x20) // ASCII shift a-z to A-Z
			}
			i++ // skip next byte since we just handled it
		} else if b == '=' { // next 2 bytes are hex and should be buffered ready to be read as utf8
			if i+2 >= len(strBytes) {
				return "", fmt.Errorf("Byte pos: %d: expected quote-printable encoding but ran out of string", i)
			}
			dst := make([]byte, 1)
			_, err := hex.Decode(dst, strBytes[i+1:i+3])
			if err != nil {
				return "", err
			}
			outputBuffer.WriteByte(dst[0])
			i += 2 // skip next 2 bytes since we just handled it
		} else { // pass through
			outputBuffer.WriteByte(b)
		}
	}
	return outputBuffer.String(), nil
}

// ExtractUserLocalpart extracts the localpart portion of a user ID.
// See http://matrix.org/docs/spec/intro.html#user-identifiers
func ExtractUserLocalpart(userID string) (string, error) {
	if len(userID) == 0 || userID[0] != '@' {
		return "", fmt.Errorf("%s is not a valid user id", userID)
	}
	return strings.TrimPrefix(
		strings.SplitN(userID, ":", 2)[0], // @foo:bar:8448 => [ "@foo", "bar:8448" ]
		"@",                               // remove "@" prefix
	), nil
}
