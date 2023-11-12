package balance

import (
	"math"
	"math/big"
)

func NewSimpleCacher() Cacher {
	return newCacherImpl(newSimpleCache())
}

// implements cacheIface for plain map internal storage
type mapCache[K comparable, V any] struct {
	cache map[K]V
}

func (c *mapCache[K, V]) get(key K) V {
	return c.cache[key]
}

func (c *mapCache[K, V]) set(key K, value V) {
	c.cache[key] = value
}

func (c *mapCache[K, V]) len() int {
	return len(c.cache)
}

func (c *mapCache[K, V]) keys() []K {
	keys := make([]K, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}

func (c *mapCache[K, V]) init() {
	c.cache = make(map[K]V)
}

func (c *mapCache[K, V]) clear() {
	c.cache = make(map[K]V)
}

// specializes generic cache
type simpleCache struct {
	genericCache[*mapCache[uint64, *big.Int], *mapCache[uint64, *int64], *mapCache[int64, nonceRange]]
}

func newSimpleCache() *simpleCache {
	return &simpleCache{
		genericCache: genericCache[*mapCache[uint64, *big.Int], *mapCache[uint64, *int64], *mapCache[int64, nonceRange]]{

			balances:        make(addressChainMap[*mapCache[uint64, *big.Int]]),
			nonces:          make(addressChainMap[*mapCache[uint64, *int64]]),
			nonceRangeCache: *newNonceRangeCache[*mapCache[int64, nonceRange]](),
		},
	}
}

// Doesn't remove all entries, but keeps max and min to use on next iterations of transfer blocks searching
func (c *simpleCache) Clear() {
	c.rw.Lock()
	defer c.rw.Unlock()

	for _, chainCache := range c.balances {
		for _, cache := range chainCache {
			if cache.len() == 0 {
				continue
			}

			var maxBlock uint64 = 0
			var minBlock uint64 = math.MaxUint64
			for _, key := range cache.keys() {
				if key > maxBlock {
					maxBlock = key
				}
				if key < minBlock {
					minBlock = key
				}
			}
			maxBlockValue := cache.get(maxBlock)
			minBlockValue := cache.get(maxBlock)
			cache.clear()

			if maxBlockValue != nil {
				cache.set(maxBlock, maxBlockValue)
			}

			if minBlockValue != nil {
				cache.set(minBlock, minBlockValue)
			}
		}
	}

	for _, chainCache := range c.nonces {
		for _, cache := range chainCache {
			if cache.len() == 0 {
				continue
			}

			var maxBlock uint64 = 0
			var minBlock uint64 = math.MaxUint64
			for _, key := range cache.keys() {
				if key > maxBlock {
					maxBlock = key
				}
				if key < minBlock {
					minBlock = key
				}
			}
			maxBlockValue := cache.get(maxBlock)
			minBlockValue := cache.get(maxBlock)
			cache.clear()

			if maxBlockValue != nil {
				cache.set(maxBlock, maxBlockValue)
			}

			if minBlockValue != nil {
				cache.set(minBlock, minBlockValue)
			}
		}
	}

	c.nonceRangeCache.clear()
}
