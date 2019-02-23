package source

import "fmt"

// FilePos represents a position information in the file.
type FilePos struct {
	Filename string // filename, if any
	Offset   int    // offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (byte count)
}

// IsValid returns true if the position is valid.
func (p FilePos) IsValid() bool {
	return p.Line > 0
}

// String returns a string in one of several forms:
//
//	file:line:column    valid position with file name
//	file:line           valid position with file name but no column (column == 0)
//	line:column         valid position without file name
//	line                valid position without file name and no column (column == 0)
//	file                invalid position with file name
//	-                   invalid position without file name
//
func (p FilePos) String() string {
	s := p.Filename

	if p.IsValid() {
		if s != "" {
			s += ":"
		}

		s += fmt.Sprintf("%d", p.Line)

		if p.Column != 0 {
			s += fmt.Sprintf(":%d", p.Column)
		}
	}

	if s == "" {
		s = "-"
	}

	return s
}
