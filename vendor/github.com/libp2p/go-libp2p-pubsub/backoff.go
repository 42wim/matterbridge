package pubsub

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	MinBackoffDelay        = 100 * time.Millisecond
	MaxBackoffDelay        = 10 * time.Second
	TimeToLive             = 10 * time.Minute
	BackoffCleanupInterval = 1 * time.Minute
	BackoffMultiplier      = 2
	MaxBackoffJitterCoff   = 100
	MaxBackoffAttempts     = 4
)

type backoffHistory struct {
	duration  time.Duration
	lastTried time.Time
	attempts  int
}

type backoff struct {
	mu          sync.Mutex
	info        map[peer.ID]*backoffHistory
	ct          int           // size threshold that kicks off the cleaner
	ci          time.Duration // cleanup intervals
	maxAttempts int           // maximum backoff attempts prior to ejection
}

func newBackoff(ctx context.Context, sizeThreshold int, cleanupInterval time.Duration, maxAttempts int) *backoff {
	b := &backoff{
		mu:          sync.Mutex{},
		ct:          sizeThreshold,
		ci:          cleanupInterval,
		maxAttempts: maxAttempts,
		info:        make(map[peer.ID]*backoffHistory),
	}

	rand.Seed(time.Now().UnixNano()) // used for jitter
	go b.cleanupLoop(ctx)

	return b
}

func (b *backoff) updateAndGet(id peer.ID) (time.Duration, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	h, ok := b.info[id]
	switch {
	case !ok || time.Since(h.lastTried) > TimeToLive:
		// first request goes immediately.
		h = &backoffHistory{
			duration: time.Duration(0),
			attempts: 0,
		}
	case h.attempts >= b.maxAttempts:
		return 0, fmt.Errorf("peer %s has reached its maximum backoff attempts", id)

	case h.duration < MinBackoffDelay:
		h.duration = MinBackoffDelay

	case h.duration < MaxBackoffDelay:
		jitter := rand.Intn(MaxBackoffJitterCoff)
		h.duration = (BackoffMultiplier * h.duration) + time.Duration(jitter)*time.Millisecond
		if h.duration > MaxBackoffDelay || h.duration < 0 {
			h.duration = MaxBackoffDelay
		}
	}

	h.attempts += 1
	h.lastTried = time.Now()
	b.info[id] = h
	return h.duration, nil
}

func (b *backoff) cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for id, h := range b.info {
		if time.Since(h.lastTried) > TimeToLive {
			delete(b.info, id)
		}
	}
}

func (b *backoff) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(b.ci)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return // pubsub shutting down
		case <-ticker.C:
			b.cleanup()
		}
	}
}
