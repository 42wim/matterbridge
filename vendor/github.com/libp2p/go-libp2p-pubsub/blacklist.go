package pubsub

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libp2p/go-libp2p-pubsub/timecache"
)

// Blacklist is an interface for peer blacklisting.
type Blacklist interface {
	Add(peer.ID) bool
	Contains(peer.ID) bool
}

// MapBlacklist is a blacklist implementation using a perfect map
type MapBlacklist map[peer.ID]struct{}

// NewMapBlacklist creates a new MapBlacklist
func NewMapBlacklist() Blacklist {
	return MapBlacklist(make(map[peer.ID]struct{}))
}

func (b MapBlacklist) Add(p peer.ID) bool {
	b[p] = struct{}{}
	return true
}

func (b MapBlacklist) Contains(p peer.ID) bool {
	_, ok := b[p]
	return ok
}

// TimeCachedBlacklist is a blacklist implementation using a time cache
type TimeCachedBlacklist struct {
	tc timecache.TimeCache
}

// NewTimeCachedBlacklist creates a new TimeCachedBlacklist with the given expiry duration
func NewTimeCachedBlacklist(expiry time.Duration) (Blacklist, error) {
	b := &TimeCachedBlacklist{tc: timecache.NewTimeCache(expiry)}
	return b, nil
}

// Add returns a bool saying whether Add of peer was successful
func (b *TimeCachedBlacklist) Add(p peer.ID) bool {
	s := p.String()
	if b.tc.Has(s) {
		return false
	}
	b.tc.Add(s)
	return true
}

func (b *TimeCachedBlacklist) Contains(p peer.ID) bool {
	return b.tc.Has(p.String())
}
