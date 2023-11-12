package timecache

import (
	"context"
	"sync"
	"time"
)

var backgroundSweepInterval = time.Minute

func background(ctx context.Context, lk sync.Locker, m map[string]time.Time) {
	ticker := time.NewTicker(backgroundSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			sweep(lk, m, now)

		case <-ctx.Done():
			return
		}
	}
}

func sweep(lk sync.Locker, m map[string]time.Time, now time.Time) {
	lk.Lock()
	defer lk.Unlock()

	for k, expiry := range m {
		if expiry.Before(now) {
			delete(m, k)
		}
	}
}
