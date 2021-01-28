/*
Package internal unimportable
*/
package internal // import "github.com/SevereCloud/vksdk/v2/internal"

import (
	"io"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
)

// illegal is a collection of runes.
type illegal struct{}

// Contains return true.
func (i illegal) Contains(r rune) bool {
	return !(r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xDF77 ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF)
}

// XMLSanitizerReader creates an io.Reader that
// wraps another io.Reader and removes illegal xml
// characters from the io stream.
func XMLSanitizerReader(xml io.Reader) io.Reader {
	var i illegal
	t := transform.Chain(runes.Remove(i))

	return transform.NewReader(xml, t)
}

// CharsetUnknownError unknown charset.
type CharsetUnknownError struct {
	Name string
}

// Error returns the message of a CharsetUnknownError.
func (c *CharsetUnknownError) Error() string {
	return "unknown charset: " + c.Name
}

// CharsetReader if non-nil, defines a function to generate
// charset-conversion readers, converting from the provided
// non-UTF-8 charset into UTF-8. If CharsetReader is nil or
// returns an error, parsing stops with an error. One of the
// the CharsetReader's result values must be non-nil.
func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "windows-1251":
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	default:
		return nil, &CharsetUnknownError{Name: charset}
	}
}
