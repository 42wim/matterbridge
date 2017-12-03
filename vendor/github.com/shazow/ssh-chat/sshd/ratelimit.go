package sshd

import (
	"io"
	"net"
	"time"

	"github.com/shazow/rateio"
)

type limitedConn struct {
	net.Conn
	io.Reader // Our rate-limited io.Reader for net.Conn
}

func (r *limitedConn) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}

// ReadLimitConn returns a net.Conn whose io.Reader interface is rate-limited by limiter.
func ReadLimitConn(conn net.Conn, limiter rateio.Limiter) net.Conn {
	return &limitedConn{
		Conn:   conn,
		Reader: rateio.NewReader(conn, limiter),
	}
}

// Count each read as 1 unless it exceeds some number of bytes.
type inputLimiter struct {
	// TODO: Could do all kinds of fancy things here, like be more forgiving of
	// connections that have been around for a while.

	Amount    int
	Frequency time.Duration

	remaining int
	readCap   int
	numRead   int
	timeRead  time.Time
}

// NewInputLimiter returns a rateio.Limiter with sensible defaults for
// differentiating between humans typing and bots spamming.
func NewInputLimiter() rateio.Limiter {
	grace := time.Second * 3
	return &inputLimiter{
		Amount:    2 << 14, // ~16kb, should be plenty for a high typing rate/copypasta/large key handshakes.
		Frequency: time.Minute * 1,
		readCap:   128,          // Allow up to 128 bytes per read (anecdotally, 1 character = 52 bytes over ssh)
		numRead:   -1024 * 1024, // Start with a 1mb grace
		timeRead:  time.Now().Add(grace),
	}
}

// Count applies 1 if n<readCap, else n
func (limit *inputLimiter) Count(n int) error {
	now := time.Now()
	if now.After(limit.timeRead) {
		limit.numRead = 0
		limit.timeRead = now.Add(limit.Frequency)
	}
	if n <= limit.readCap {
		limit.numRead += 1
	} else {
		limit.numRead += n
	}
	if limit.numRead > limit.Amount {
		return rateio.ErrRateExceeded
	}
	return nil
}
