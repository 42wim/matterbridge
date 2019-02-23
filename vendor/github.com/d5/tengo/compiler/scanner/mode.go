package scanner

// Mode represents a scanner mode.
type Mode int

// List of scanner modes.
const (
	ScanComments Mode = 1 << iota
	DontInsertSemis
)
