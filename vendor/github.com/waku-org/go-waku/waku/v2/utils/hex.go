package utils

import (
	"encoding/hex"
	"strings"
)

// DecodeHexString decodes input string into a hex string.
// Note that if the string is prefixed by 0x , it is trimmed
func DecodeHexString(input string) ([]byte, error) {
	input = strings.TrimPrefix(input, "0x")
	return hex.DecodeString(input)
}
