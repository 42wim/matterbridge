package rendezvous

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type rendezvousDiscovery struct {
	rp           RendezvousPoint
	peerCache    map[string]*discoveryCache
	peerCacheMux sync.RWMutex
	rng          *rand.Rand
	rngMux       sync.Mutex
}

type discoveryCache struct {
	recs   map[peer.ID]*peerRecord
	cookie []byte
	mux    sync.Mutex
}

type peerRecord struct {
	peer   peer.AddrInfo
	expire int64
}

func NewRendezvousDiscovery(host host.Host, rendezvousPeer peer.ID) discovery.Discovery {
	rp := NewRendezvousPoint(host, rendezvousPeer)
	return &rendezvousDiscovery{rp: rp, peerCache: make(map[string]*discoveryCache), rng: rand.New(rand.NewSource(rand.Int63()))}
}

func (c *rendezvousDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	// Get options
	var options discovery.Options
	err := options.Apply(opts...)
	if err != nil {
		return 0, err
	}

	ttl := options.Ttl
	var ttlSeconds int

	if ttl == 0 {
		ttlSeconds = 7200
	} else {
		ttlSeconds = int(math.Round(ttl.Seconds()))
	}

	if rttl, err := c.rp.Register(ctx, ns, ttlSeconds); err != nil {
		return 0, err
	} else {
		return rttl, nil
	}
}

func (c *rendezvousDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	// Get options
	var options discovery.Options
	err := options.Apply(opts...)
	if err != nil {
		return nil, err
	}

	const maxLimit = 1000
	limit := options.Limit
	if limit == 0 || limit > maxLimit {
		limit = maxLimit
	}

	// Get cached peers
	var cache *discoveryCache

	c.peerCacheMux.RLock()
	cache, ok := c.peerCache[ns]
	c.peerCacheMux.RUnlock()
	if !ok {
		c.peerCacheMux.Lock()
		cache, ok = c.peerCache[ns]
		if !ok {
			cache = &discoveryCache{recs: make(map[peer.ID]*peerRecord)}
			c.peerCache[ns] = cache
		}
		c.peerCacheMux.Unlock()
	}

	cache.mux.Lock()
	defer cache.mux.Unlock()

	// Remove all expired entries from cache
	currentTime := time.Now().Unix()
	newCacheSize := len(cache.recs)

	for p := range cache.recs {
		rec := cache.recs[p]
		if rec.expire < currentTime {
			newCacheSize--
			delete(cache.recs, p)
		}
	}

	cookie := cache.cookie

	// Discover new records if we don't have enough
	if newCacheSize < limit {
		// TODO: Should we return error even if we have valid cached results?
		var regs []Registration
		var newCookie []byte
		if regs, newCookie, err = c.rp.Discover(ctx, ns, limit, cookie); err == nil {
			for _, reg := range regs {
				rec := &peerRecord{peer: reg.Peer, expire: int64(reg.Ttl) + currentTime}
				cache.recs[rec.peer.ID] = rec
			}
			cache.cookie = newCookie
		}
	}

	// Randomize and fill channel with available records
	count := len(cache.recs)
	if limit < count {
		count = limit
	}

	chPeer := make(chan peer.AddrInfo, count)

	c.rngMux.Lock()
	perm := c.rng.Perm(len(cache.recs))[0:count]
	c.rngMux.Unlock()

	permSet := make(map[int]int)
	for i, v := range perm {
		permSet[v] = i
	}

	sendLst := make([]*peer.AddrInfo, count)
	iter := 0
	for k := range cache.recs {
		if sendIndex, ok := permSet[iter]; ok {
			sendLst[sendIndex] = &cache.recs[k].peer
		}
		iter++
	}

	for _, send := range sendLst {
		chPeer <- *send
	}

	close(chPeer)
	return chPeer, err
}
