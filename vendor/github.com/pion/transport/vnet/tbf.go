package vnet

import (
	"sync"
	"time"
)

const (
	// Bit is a single bit
	Bit = 1
	// KBit is a kilobit
	KBit = 1000 * Bit
	// MBit is a Megabit
	MBit = 1000 * KBit
)

// TokenBucketFilter implements a token bucket rate limit algorithm.
type TokenBucketFilter struct {
	NIC
	currentTokensInBucket int
	c                     chan Chunk
	queue                 *chunkQueue
	queueSize             int // in bytes

	mutex    sync.Mutex
	rate     int
	maxBurst int

	wg   sync.WaitGroup
	done chan struct{}
}

// TBFOption is the option type to configure a TokenBucketFilter
type TBFOption func(*TokenBucketFilter) TBFOption

// TBFQueueSizeInBytes sets the max number of bytes waiting in the queue. Can
// only be set in constructor before using the TBF.
func TBFQueueSizeInBytes(bytes int) TBFOption {
	return func(t *TokenBucketFilter) TBFOption {
		prev := t.queueSize
		t.queueSize = bytes
		return TBFQueueSizeInBytes(prev)
	}
}

// TBFRate sets the bitrate of a TokenBucketFilter
func TBFRate(rate int) TBFOption {
	return func(t *TokenBucketFilter) TBFOption {
		t.mutex.Lock()
		defer t.mutex.Unlock()
		previous := t.rate
		t.rate = rate
		return TBFRate(previous)
	}
}

// TBFMaxBurst sets the bucket size of the token bucket filter. This is the
// maximum size that can instantly leave the filter, if the bucket is full.
func TBFMaxBurst(size int) TBFOption {
	return func(t *TokenBucketFilter) TBFOption {
		t.mutex.Lock()
		defer t.mutex.Unlock()
		previous := t.maxBurst
		t.maxBurst = size
		return TBFMaxBurst(previous)
	}
}

// Set updates a setting on the token bucket filter
func (t *TokenBucketFilter) Set(opts ...TBFOption) (previous TBFOption) {
	for _, opt := range opts {
		previous = opt(t)
	}
	return previous
}

// NewTokenBucketFilter creates and starts a new TokenBucketFilter
func NewTokenBucketFilter(n NIC, opts ...TBFOption) (*TokenBucketFilter, error) {
	tbf := &TokenBucketFilter{
		NIC:                   n,
		currentTokensInBucket: 0,
		c:                     make(chan Chunk),
		queue:                 nil,
		queueSize:             50000,
		mutex:                 sync.Mutex{},
		rate:                  1 * MBit,
		maxBurst:              2 * KBit,
		wg:                    sync.WaitGroup{},
		done:                  make(chan struct{}),
	}
	tbf.Set(opts...)
	tbf.queue = newChunkQueue(0, tbf.queueSize)
	tbf.wg.Add(1)
	go tbf.run()
	return tbf, nil
}

func (t *TokenBucketFilter) onInboundChunk(c Chunk) {
	t.c <- c
}

func (t *TokenBucketFilter) run() {
	defer t.wg.Done()
	ticker := time.NewTicker(1 * time.Millisecond)

	for {
		select {
		case <-t.done:
			ticker.Stop()
			t.drainQueue()
			return
		case <-ticker.C:
			t.mutex.Lock()
			if t.currentTokensInBucket < t.maxBurst {
				// add (bitrate * S) / 1000 converted to bytes (divide by 8) S
				// is the update interval in milliseconds
				t.currentTokensInBucket += (t.rate / 1000) / 8
			}
			t.mutex.Unlock()
			t.drainQueue()
		case chunk := <-t.c:
			t.queue.push(chunk)
			t.drainQueue()
		}
	}
}

func (t *TokenBucketFilter) drainQueue() {
	for {
		next := t.queue.peek()
		if next == nil {
			break
		}
		tokens := len(next.UserData())
		if t.currentTokensInBucket < tokens {
			break
		}
		t.queue.pop()
		t.NIC.onInboundChunk(next)
		t.currentTokensInBucket -= tokens
	}
}

// Close closes and stops the token bucket filter queue
func (t *TokenBucketFilter) Close() error {
	close(t.done)
	t.wg.Wait()
	return nil
}
