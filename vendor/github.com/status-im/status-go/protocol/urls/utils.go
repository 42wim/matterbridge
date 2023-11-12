package urls

import (
	"bytes"
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/andybalholm/brotli"
)

const (
	htmlTagStart = 60 // Unicode `<`
	htmlTagEnd   = 62 // Unicode `>`
)

// Taken from https://stackoverflow.com/a/64701836
// Aggressively strips HTML tags from a string.
// It will only keep anything between `>` and `<`.
func stripHTMLTags(s string) string {
	// Setup a string builder and allocate enough memory for the new string.
	var builder strings.Builder
	builder.Grow(len(s) + utf8.UTFMax)

	in := false // True if we are inside an HTML tag.
	start := 0  // The index of the previous start tag character `<`
	end := 0    // The index of the previous end tag character `>`

	for i, c := range s {
		// If this is the last character and we are not in an HTML tag, save it.
		if (i+1) == len(s) && end >= start {
			builder.WriteString(s[end:])
		}

		// Keep going if the character is not `<` or `>`
		if c != htmlTagStart && c != htmlTagEnd {
			continue
		}

		if c == htmlTagStart {
			// Only update the start if we are not in a tag.
			// This make sure we strip out `<<br>` not just `<br>`
			if !in {
				start = i
			}
			in = true

			// Write the valid string between the close and start of the two tags.
			builder.WriteString(s[end:start])
			continue
		}
		// else c == htmlTagEnd
		in = false
		end = i + 1
	}
	s = builder.String()
	return s
}

func encodeDataURL(data []byte) (string, error) {
	bb := bytes.NewBuffer([]byte{})
	writer := brotli.NewWriter(bb)
	_, err := writer.Write(data)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bb.Bytes()), nil
}

func decodeDataURL(data string) ([]byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	output := make([]byte, 4096)
	bb := bytes.NewBuffer(decoded)
	reader := brotli.NewReader(bb)
	n, err := reader.Read(output)
	if err != nil {
		return nil, err
	}

	return output[:n], nil
}
