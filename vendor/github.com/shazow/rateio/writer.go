package rateio

import "io"

type writer struct {
	io.Writer
	Limiter
}

// Write writes the contents of p into the buffer.
// Returns ErrRateExceeded error if our specified read is exceeded.
func (w *writer) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		return
	}

	err = w.Limiter.Count(n)
	return
}

// NewWriter proxies an io.Writer but keeps track of bytes read based on our Limiter.
func NewWriter(w io.Writer, limiter Limiter) io.Writer {
	return &writer{w, limiter}
}
