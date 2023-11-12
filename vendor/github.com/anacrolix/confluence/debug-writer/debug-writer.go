package debug_writer

import "io"

type Interface interface {
	WriteDebug(w io.Writer)
}
