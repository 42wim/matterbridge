package autorelay

import (
	"context"
	"errors"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// AutoRelay will call this function when it needs new candidates because it is
// not connected to the desired number of relays or we get disconnected from one
// of the relays. Implementations must send *at most* numPeers, and close the
// channel when they don't intend to provide any more peers. AutoRelay will not
// call the callback again until the channel is closed. Implementations should
// send new peers, but may send peers they sent before. AutoRelay implements a
// per-peer backoff (see WithBackoff). See WithMinInterval for setting the
// minimum interval between calls to the callback. The context.Context passed
// may be canceled when AutoRelay feels satisfied, it will be canceled when the
// node is shutting down. If the context is canceled you MUST close the output
// channel at some point.
type PeerSource func(ctx context.Context, num int) <-chan peer.AddrInfo

type config struct {
	clock      ClockWithInstantTimer
	peerSource PeerSource
	// minimum interval used to call the peerSource callback
	minInterval time.Duration
	// see WithMinCandidates
	minCandidates int
	// see WithMaxCandidates
	maxCandidates int
	// Delay until we obtain reservations with relays, if we have less than minCandidates candidates.
	// See WithBootDelay.
	bootDelay time.Duration
	// backoff is the time we wait after failing to obtain a reservation with a candidate
	backoff time.Duration
	// Number of relays we strive to obtain a reservation with.
	desiredRelays int
	// see WithMaxCandidateAge
	maxCandidateAge  time.Duration
	setMinCandidates bool
	// see WithMetricsTracer
	metricsTracer MetricsTracer
}

var defaultConfig = config{
	clock:           RealClock{},
	minCandidates:   4,
	maxCandidates:   20,
	bootDelay:       3 * time.Minute,
	backoff:         time.Hour,
	desiredRelays:   2,
	maxCandidateAge: 30 * time.Minute,
	minInterval:     30 * time.Second,
}

var (
	errAlreadyHavePeerSource = errors.New("can only use a single WithPeerSource or WithStaticRelays")
)

type Option func(*config) error

func WithStaticRelays(static []peer.AddrInfo) Option {
	return func(c *config) error {
		if c.peerSource != nil {
			return errAlreadyHavePeerSource
		}

		WithPeerSource(func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
			if len(static) < numPeers {
				numPeers = len(static)
			}
			c := make(chan peer.AddrInfo, numPeers)
			defer close(c)

			for i := 0; i < numPeers; i++ {
				c <- static[i]
			}
			return c
		})(c)
		WithMinCandidates(len(static))(c)
		WithMaxCandidates(len(static))(c)
		WithNumRelays(len(static))(c)

		return nil
	}
}

// WithPeerSource defines a callback for AutoRelay to query for more relay candidates.
func WithPeerSource(f PeerSource) Option {
	return func(c *config) error {
		if c.peerSource != nil {
			return errAlreadyHavePeerSource
		}
		c.peerSource = f
		return nil
	}
}

// WithNumRelays sets the number of relays we strive to obtain reservations with.
func WithNumRelays(n int) Option {
	return func(c *config) error {
		c.desiredRelays = n
		return nil
	}
}

// WithMaxCandidates sets the number of relay candidates that we buffer.
func WithMaxCandidates(n int) Option {
	return func(c *config) error {
		c.maxCandidates = n
		if c.minCandidates > n {
			c.minCandidates = n
		}
		return nil
	}
}

// WithMinCandidates sets the minimum number of relay candidates we collect before to get a reservation
// with any of them (unless we've been running for longer than the boot delay).
// This is to make sure that we don't just randomly connect to the first candidate that we discover.
func WithMinCandidates(n int) Option {
	return func(c *config) error {
		if n > c.maxCandidates {
			n = c.maxCandidates
		}
		c.minCandidates = n
		c.setMinCandidates = true
		return nil
	}
}

// WithBootDelay set the boot delay for finding relays.
// We won't attempt any reservation if we've have less than a minimum number of candidates.
// This prevents us to connect to the "first best" relay, and allows us to carefully select the relay.
// However, in case we haven't found enough relays after the boot delay, we use what we have.
func WithBootDelay(d time.Duration) Option {
	return func(c *config) error {
		c.bootDelay = d
		return nil
	}
}

// WithBackoff sets the time we wait after failing to obtain a reservation with a candidate.
func WithBackoff(d time.Duration) Option {
	return func(c *config) error {
		c.backoff = d
		return nil
	}
}

// WithMaxCandidateAge sets the maximum age of a candidate.
// When we are connected to the desired number of relays, we don't ask the peer source for new candidates.
// This can lead to AutoRelay's candidate list becoming outdated, and means we won't be able
// to quickly establish a new relay connection if our existing connection breaks, if all the candidates
// have become stale.
func WithMaxCandidateAge(d time.Duration) Option {
	return func(c *config) error {
		c.maxCandidateAge = d
		return nil
	}
}

// InstantTimer is a timer that triggers at some instant rather than some duration
type InstantTimer interface {
	Reset(d time.Time) bool
	Stop() bool
	Ch() <-chan time.Time
}

// ClockWithInstantTimer is a clock that can create timers that trigger at some
// instant rather than some duration
type ClockWithInstantTimer interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	InstantTimer(when time.Time) InstantTimer
}

type RealTimer struct{ t *time.Timer }

var _ InstantTimer = (*RealTimer)(nil)

func (t RealTimer) Ch() <-chan time.Time {
	return t.t.C
}

func (t RealTimer) Reset(d time.Time) bool {
	return t.t.Reset(time.Until(d))
}

func (t RealTimer) Stop() bool {
	return t.t.Stop()
}

type RealClock struct{}

var _ ClockWithInstantTimer = RealClock{}

func (RealClock) Now() time.Time {
	return time.Now()
}
func (RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}
func (RealClock) InstantTimer(when time.Time) InstantTimer {
	t := time.NewTimer(time.Until(when))
	return &RealTimer{t}
}

func WithClock(cl ClockWithInstantTimer) Option {
	return func(c *config) error {
		c.clock = cl
		return nil
	}
}

// WithMinInterval sets the minimum interval after which peerSource callback will be called for more
// candidates even if AutoRelay needs new candidates.
func WithMinInterval(interval time.Duration) Option {
	return func(c *config) error {
		c.minInterval = interval
		return nil
	}
}

// WithMetricsTracer configures autorelay to use mt to track metrics
func WithMetricsTracer(mt MetricsTracer) Option {
	return func(c *config) error {
		c.metricsTracer = mt
		return nil
	}
}
