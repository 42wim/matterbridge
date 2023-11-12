package balance

import (
	"context"
	"math/big"
	"time"

	"github.com/jellydator/ttlcache/v3"

	"github.com/ethereum/go-ethereum/log"
)

var (
	defaultTTLValue = 5 * time.Minute
)

func NewCacherWithTTL(ttl time.Duration) Cacher {
	return newCacherImpl(newCacheWithTTL(ttl))
}

// TTL cache implementation of cacheIface
type ttlCache[K comparable, V any] struct {
	cache *ttlcache.Cache[K, V]
}

//nolint:golint,unused // linter does not detect using it via reflect
func (c *ttlCache[K, V]) get(key K) V {
	item := c.cache.Get(key)
	if item == nil {
		var v V
		return v
	}
	return item.Value()
}

//nolint:golint,unused // linter does not detect using it via reflect
func (c *ttlCache[K, V]) set(key K, value V) {
	_ = c.cache.Set(key, value, ttlcache.DefaultTTL)
}

//nolint:golint,unused // linter does not detect using it via reflect
func (c *ttlCache[K, V]) len() int {
	return c.cache.Len()
}

//nolint:golint,unused // linter does not detect using it via reflect
func (c *ttlCache[K, V]) keys() []K {
	return c.cache.Keys()
}

//nolint:golint,unused // linter does not detect using it via reflect
func (c *ttlCache[K, V]) init() {
	c.cache = ttlcache.New[K, V](
		ttlcache.WithTTL[K, V](defaultTTLValue),
	)
	c.cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[K, V]) {
		log.Debug("Evicting item from balance/nonce cache", "reason", reason, "key", item.Key, "value", item.Value)
	})
	go c.cache.Start() // starts automatic expired item deletion
}

//nolint:golint,unused // linter does not detect using it via reflect
func (c *ttlCache[K, V]) clear() {
	c.cache.DeleteAll()
}

// specializes generic cache
type cacheWithTTL struct {
	// TODO: use ttlCache instead of mapCache for nonceRangeCache. For that we need to update sortedRanges on item eviction
	// For now, nonceRanges cache is not updated on nonces items eviction, but it should not be as big as nonceCache is
	genericCache[*ttlCache[uint64, *big.Int], *ttlCache[uint64, *int64], *mapCache[int64, nonceRange]]
}

func newCacheWithTTL(ttl time.Duration) *cacheWithTTL {
	defaultTTLValue = ttl

	return &cacheWithTTL{
		genericCache: genericCache[*ttlCache[uint64, *big.Int], *ttlCache[uint64, *int64], *mapCache[int64, nonceRange]]{
			balances:        make(addressChainMap[*ttlCache[uint64, *big.Int]]),
			nonces:          make(addressChainMap[*ttlCache[uint64, *int64]]),
			nonceRangeCache: *newNonceRangeCache[*mapCache[int64, nonceRange]](),
		},
	}
}

func (c *cacheWithTTL) Clear() {
	c.rw.Lock()
	defer c.rw.Unlock()

	// TTL cache removes expired items automatically
	// but in case we want to clear it manually we can do it here
	for _, chainCache := range c.balances {
		for _, cache := range chainCache {
			cache.clear()
		}
	}

	for _, chainCache := range c.nonces {
		for _, cache := range chainCache {
			cache.clear()
		}
	}

	c.nonceRangeCache.clear()
}
