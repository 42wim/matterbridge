package valid

import (
	"bytes"
)

var URIs = [][]byte{
	[]byte("http://"),
	[]byte("https://"),
	[]byte("ftp://"),
	[]byte("mailto:"),
}

var Paths = [][]byte{
	[]byte("/"),
	[]byte("./"),
	[]byte("../"),
}

// TODO: documentation
func IsSafeURL(url []byte) bool {
	nLink := len(url)
	for _, path := range Paths {
		nPath := len(path)
		linkPrefix := url[:nPath]
		if nLink >= nPath && bytes.Equal(linkPrefix, path) {
			if nLink == nPath {
				return true
			} else if isAlnum(url[nPath]) {
				return true
			}
		}
	}

	for _, prefix := range URIs {
		// TODO: handle unicode here
		// case-insensitive prefix test
		nPrefix := len(prefix)
		if nLink > nPrefix {
			linkPrefix := bytes.ToLower(url[:nPrefix])
			if bytes.Equal(linkPrefix, prefix) && isAlnum(url[nPrefix]) {
				return true
			}
		}
	}

	return false
}

// isAlnum returns true if c is a digit or letter
// TODO: check when this is looking for ASCII alnum and when it should use unicode
func isAlnum(c byte) bool {
	return (c >= '0' && c <= '9') || isLetter(c)
}

// isLetter returns true if c is ascii letter
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
