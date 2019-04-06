package runtime

import (
	"errors"
)

// ErrStackOverflow is a stack overflow error.
var ErrStackOverflow = errors.New("stack overflow")

// ErrObjectAllocLimit is an objects allocation limit error.
var ErrObjectAllocLimit = errors.New("object allocation limit exceeded")
