package parser

// Pos represents a position in the file set.
type Pos int

// NoPos represents an invalid position.
const NoPos Pos = 0

// IsValid returns true if the position is valid.
func (p Pos) IsValid() bool {
	return p != NoPos
}
