package swarm

import "time"

// InstantTimer is a timer that triggers at some instant rather than some duration
type InstantTimer interface {
	Reset(d time.Time) bool
	Stop() bool
	Ch() <-chan time.Time
}

// Clock is a clock that can create timers that trigger at some
// instant rather than some duration
type Clock interface {
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

var _ Clock = RealClock{}

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
