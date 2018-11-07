package birc

import (
	"strings"
)

/*
func tableformatter(nicks []string, nicksPerRow int, continued bool) string {
	result := "|IRC users"
	if continued {
		result = "|(continued)"
	}
	for i := 0; i < 2; i++ {
		for j := 1; j <= nicksPerRow && j <= len(nicks); j++ {
			if i == 0 {
				result += "|"
			} else {
				result += ":-|"
			}
		}
		result += "\r\n|"
	}
	result += nicks[0] + "|"
	for i := 1; i < len(nicks); i++ {
		if i%nicksPerRow == 0 {
			result += "\r\n|" + nicks[i] + "|"
		} else {
			result += nicks[i] + "|"
		}
	}
	return result
}
*/

func plainformatter(nicks []string) string {
	return strings.Join(nicks, ", ") + " currently on IRC"
}

func IsMarkup(message string) bool {
	switch message[0] {
	case '|':
		fallthrough
	case '#':
		fallthrough
	case '_':
		fallthrough
	case '*':
		fallthrough
	case '~':
		fallthrough
	case '-':
		fallthrough
	case ':':
		fallthrough
	case '>':
		fallthrough
	case '=':
		return true
	}
	return false
}
