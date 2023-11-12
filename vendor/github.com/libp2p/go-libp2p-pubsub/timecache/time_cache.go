package timecache

import (
	"time"

	logger "github.com/ipfs/go-log/v2"
)

var log = logger.Logger("pubsub/timecache")

// Stategy is the TimeCache expiration strategy to use.
type Strategy uint8

const (
	// Strategy_FirstSeen expires an entry from the time it was added.
	Strategy_FirstSeen Strategy = iota
	// Stategy_LastSeen expires an entry from the last time it was touched by an Add or Has.
	Strategy_LastSeen
)

// TimeCache is a cahe of recently seen messages (by id).
type TimeCache interface {
	// Add adds an id into the cache, if it is not already there.
	// Returns true if the id was newly added to the cache.
	// Depending on the implementation strategy, it may or may not update the expiry of
	// an existing entry.
	Add(string) bool
	// Has checks the cache for the presence of an id.
	// Depending on the implementation strategy, it may or may not update the expiry of
	// an existing entry.
	Has(string) bool
	// Done signals that the user is done with this cache, which it may stop background threads
	// and relinquish resources.
	Done()
}

// NewTimeCache defaults to the original ("first seen") cache implementation
func NewTimeCache(ttl time.Duration) TimeCache {
	return NewTimeCacheWithStrategy(Strategy_FirstSeen, ttl)
}

func NewTimeCacheWithStrategy(strategy Strategy, ttl time.Duration) TimeCache {
	switch strategy {
	case Strategy_FirstSeen:
		return newFirstSeenCache(ttl)
	case Strategy_LastSeen:
		return newLastSeenCache(ttl)
	default:
		// Default to the original time cache implementation
		return newFirstSeenCache(ttl)
	}
}
