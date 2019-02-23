package runtime

import (
	"errors"
)

// ErrStackOverflow is a stack overflow error.
var ErrStackOverflow = errors.New("stack overflow")
