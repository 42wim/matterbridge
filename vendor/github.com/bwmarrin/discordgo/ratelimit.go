package discordgo

import (
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter holds all ratelimit buckets
type RateLimiter struct {
	sync.Mutex
	global          *int64
	buckets         map[string]*Bucket
	globalRateLimit time.Duration
}

// NewRatelimiter returns a new RateLimiter
func NewRatelimiter() *RateLimiter {

	return &RateLimiter{
		buckets: make(map[string]*Bucket),
		global:  new(int64),
	}
}

// getBucket retrieves or creates a bucket
func (r *RateLimiter) getBucket(key string) *Bucket {
	r.Lock()
	defer r.Unlock()

	if bucket, ok := r.buckets[key]; ok {
		return bucket
	}

	b := &Bucket{
		remaining: 1,
		Key:       key,
		global:    r.global,
	}

	r.buckets[key] = b
	return b
}

// LockBucket Locks until a request can be made
func (r *RateLimiter) LockBucket(bucketID string) *Bucket {

	b := r.getBucket(bucketID)

	b.Lock()

	// If we ran out of calls and the reset time is still ahead of us
	// then we need to take it easy and relax a little
	if b.remaining < 1 && b.reset.After(time.Now()) {
		time.Sleep(b.reset.Sub(time.Now()))

	}

	// Check for global ratelimits
	sleepTo := time.Unix(0, atomic.LoadInt64(r.global))
	if now := time.Now(); now.Before(sleepTo) {
		time.Sleep(sleepTo.Sub(now))
	}

	b.remaining--
	return b
}

// Bucket represents a ratelimit bucket, each bucket gets ratelimited individually (-global ratelimits)
type Bucket struct {
	sync.Mutex
	Key       string
	remaining int
	limit     int
	reset     time.Time
	global    *int64
}

// Release unlocks the bucket and reads the headers to update the buckets ratelimit info
// and locks up the whole thing in case if there's a global ratelimit.
func (b *Bucket) Release(headers http.Header) error {

	defer b.Unlock()
	if headers == nil {
		return nil
	}

	remaining := headers.Get("X-RateLimit-Remaining")
	reset := headers.Get("X-RateLimit-Reset")
	global := headers.Get("X-RateLimit-Global")
	retryAfter := headers.Get("Retry-After")

	// Update global and per bucket reset time if the proper headers are available
	// If global is set, then it will block all buckets until after Retry-After
	// If Retry-After without global is provided it will use that for the new reset
	// time since it's more accurate than X-RateLimit-Reset.
	// If Retry-After after is not proided, it will update the reset time from X-RateLimit-Reset
	if retryAfter != "" {
		parsedAfter, err := strconv.ParseInt(retryAfter, 10, 64)
		if err != nil {
			return err
		}

		resetAt := time.Now().Add(time.Duration(parsedAfter) * time.Millisecond)

		// Lock either this single bucket or all buckets
		if global != "" {
			atomic.StoreInt64(b.global, resetAt.UnixNano())
		} else {
			b.reset = resetAt
		}
	} else if reset != "" {
		// Calculate the reset time by using the date header returned from discord
		discordTime, err := http.ParseTime(headers.Get("Date"))
		if err != nil {
			return err
		}

		unix, err := strconv.ParseInt(reset, 10, 64)
		if err != nil {
			return err
		}

		// Calculate the time until reset and add it to the current local time
		// some extra time is added because without it i still encountered 429's.
		// The added amount is the lowest amount that gave no 429's
		// in 1k requests
		delta := time.Unix(unix, 0).Sub(discordTime) + time.Millisecond*250
		b.reset = time.Now().Add(delta)
	}

	// Udpate remaining if header is present
	if remaining != "" {
		parsedRemaining, err := strconv.ParseInt(remaining, 10, 32)
		if err != nil {
			return err
		}
		b.remaining = int(parsedRemaining)
	}

	return nil
}
