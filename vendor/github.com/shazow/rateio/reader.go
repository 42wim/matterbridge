package rateio

import "io"

type reader struct {
	io.Reader
	Limiter
}

// Read reads data into p.
// Returns ErrRateExceeded error if our specified read is exceeded.
func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil {
		return
	}

	err = r.Limiter.Count(n)
	return
}

// NewReader proxies an io.Reader but keeps track of bytes read based on our Limiter.
func NewReader(r io.Reader, limiter Limiter) io.Reader {
	return &reader{r, limiter}
}
