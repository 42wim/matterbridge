package rateio

import (
	"errors"
	"time"
)

const minInt = -int(^uint(0)>>1) - 1

// The error returned when the read rate exceeds our specification.
var ErrRateExceeded = errors.New("Read rate exceeded.")

// Limiter is an interface for a rate limiter.
// There are a few example limiters included in the package, but feel free to go wild with your own.
type Limiter interface {
	// Apply this many bytes to the limiter, return ErrRateExceeded if the defined rate is exceeded.
	Count(int) error
}

// simpleLimiter is a rate limiter that restricts Amount bytes in Frequency duration.
type simpleLimiter struct {
	Amount    int
	Frequency time.Duration

	numRead  int
	timeRead time.Time
}

// NewSimpleLimiter creates a Limiter that restricts a given number of bytes per frequency.
func NewSimpleLimiter(amount int, frequency time.Duration) Limiter {
	return &simpleLimiter{
		Amount:    amount,
		Frequency: frequency,
	}
}

// NewGracefulLimiter returns a Limiter that is the same as a
// SimpleLimiter but adds a grace period at the start of the rate
// limiting where it allows unlimited bytes to be read during that
// period.
func NewGracefulLimiter(amount int, frequency time.Duration, grace time.Duration) Limiter {
	return &simpleLimiter{
		Amount:    amount,
		Frequency: frequency,
		numRead:   minInt,
		timeRead:  time.Now().Add(grace),
	}
}

// Count applies n bytes to the limiter.
func (limit *simpleLimiter) Count(n int) error {
	now := time.Now()
	if now.After(limit.timeRead) {
		limit.numRead = 0
		limit.timeRead = now.Add(limit.Frequency)
	}
	limit.numRead += n
	if limit.numRead > limit.Amount {
		return ErrRateExceeded
	}
	return nil
}
