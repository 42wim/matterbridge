package socketmode

import "time"

type deadmanTimer struct {
	timeout time.Duration
	timer   *time.Timer
}

func newDeadmanTimer(timeout time.Duration) *deadmanTimer {
	return &deadmanTimer{
		timeout: timeout,
		timer:   time.NewTimer(timeout),
	}
}

func (smc *deadmanTimer) Elapsed() <-chan time.Time {
	return smc.timer.C
}

func (smc *deadmanTimer) Reset() {
	// Note that this is the correct way to Reset a non-expired timer
	if !smc.timer.Stop() {
		select {
		case <-smc.timer.C:
		default:
		}
	}

	smc.timer.Reset(smc.timeout)
}
