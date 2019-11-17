package sanitize

import "regexp"

var (
	reStripName = regexp.MustCompile("[^\\w.-]")
	reStripData = regexp.MustCompile("[^[:ascii:]]|[[:cntrl:]]")
)

const maxLength = 16

// Name returns a name with only allowed characters and a reasonable length
func Name(s string) string {
	s = reStripName.ReplaceAllString(s, "")
	nameLength := maxLength
	if len(s) <= maxLength {
		nameLength = len(s)
	}
	s = s[:nameLength]
	return s
}

// Data returns a string with only allowed characters for client-provided metadata inputs.
func Data(s string, maxlen int) string {
	if len(s) > maxlen {
		s = s[:maxlen]
	}
	return reStripData.ReplaceAllString(s, "")
}
