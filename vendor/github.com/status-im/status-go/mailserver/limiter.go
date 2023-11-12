package mailserver

import (
	"sync"
	"time"
)

type rateLimiter struct {
	sync.RWMutex

	lifespan time.Duration // duration of the limit
	db       map[string]time.Time

	period time.Duration
	cancel chan struct{}
}

func newRateLimiter(duration time.Duration) *rateLimiter {
	return &rateLimiter{
		lifespan: duration,
		db:       make(map[string]time.Time),
		period:   time.Second,
	}
}

func (l *rateLimiter) Start() {
	cancel := make(chan struct{})

	l.Lock()
	l.cancel = cancel
	l.Unlock()

	go l.cleanUp(l.period, cancel)
}

func (l *rateLimiter) Stop() {
	l.Lock()
	defer l.Unlock()

	if l.cancel == nil {
		return
	}
	close(l.cancel)
	l.cancel = nil
}

func (l *rateLimiter) Add(id string) {
	l.Lock()
	l.db[id] = time.Now()
	l.Unlock()
}

func (l *rateLimiter) IsAllowed(id string) bool {
	l.RLock()
	defer l.RUnlock()

	if lastRequestTime, ok := l.db[id]; ok {
		return lastRequestTime.Add(l.lifespan).Before(time.Now())
	}

	return true
}

func (l *rateLimiter) cleanUp(period time.Duration, cancel <-chan struct{}) {
	t := time.NewTicker(period)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			l.deleteExpired()
		case <-cancel:
			return
		}
	}
}

func (l *rateLimiter) deleteExpired() {
	l.Lock()
	defer l.Unlock()

	now := time.Now()
	for id, lastRequestTime := range l.db {
		if lastRequestTime.Add(l.lifespan).Before(now) {
			delete(l.db, id)
		}
	}
}
