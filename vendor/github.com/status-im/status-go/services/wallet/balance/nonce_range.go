package balance

import (
	"math/big"
	"reflect"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type nonceRange struct {
	nonce int64
	max   *big.Int
	min   *big.Int
}

type sortedNonceRangesCacheType addressChainMap[[]nonceRange] // address->chainID->[]nonceRange

type nonceRangeCache[T cacheIface[int64, nonceRange]] struct {
	nonceRanges  addressChainMap[T]
	sortedRanges sortedNonceRangesCacheType
	rw           sync.RWMutex
}

func newNonceRangeCache[T cacheIface[int64, nonceRange]]() *nonceRangeCache[T] {
	return &nonceRangeCache[T]{
		nonceRanges:  make(addressChainMap[T]),
		sortedRanges: make(sortedNonceRangesCacheType),
	}
}

func (b *nonceRangeCache[T]) updateNonceRange(account common.Address, chainID uint64, blockNumber *big.Int, nonce *int64) {
	b.rw.Lock()
	defer b.rw.Unlock()

	_, exists := b.nonceRanges[account]
	if !exists {
		b.nonceRanges[account] = make(map[uint64]T)
	}
	_, exists = b.nonceRanges[account][chainID]
	if !exists {
		b.nonceRanges[account][chainID] = reflect.New(reflect.TypeOf(b.nonceRanges[account][chainID]).Elem()).Interface().(T)
		b.nonceRanges[account][chainID].init()
	}

	nr := b.nonceRanges[account][chainID].get(*nonce)
	if nr == reflect.Zero(reflect.TypeOf(nr)).Interface() {
		nr = nonceRange{
			max:   big.NewInt(0).Set(blockNumber),
			min:   big.NewInt(0).Set(blockNumber),
			nonce: *nonce,
		}
	} else {
		if nr.max.Cmp(blockNumber) == -1 {
			nr.max.Set(blockNumber)
		}

		if nr.min.Cmp(blockNumber) == 1 {
			nr.min.Set(blockNumber)
		}
	}

	b.nonceRanges[account][chainID].set(*nonce, nr)
	b.sortRanges(account, chainID)
}

func (b *nonceRangeCache[_]) findNonceInRange(account common.Address, chainID uint64, block *big.Int) *int64 {
	b.rw.RLock()
	defer b.rw.RUnlock()

	for k := range b.sortedRanges[account][chainID] {
		nr := b.sortedRanges[account][chainID][k]
		cmpMin := nr.min.Cmp(block)
		if cmpMin == 1 {
			return nil
		} else if cmpMin == 0 {
			return &nr.nonce
		} else {
			cmpMax := nr.max.Cmp(block)
			if cmpMax >= 0 {
				return &nr.nonce
			}
		}
	}

	return nil
}

func (b *nonceRangeCache[T]) sortRanges(account common.Address, chainID uint64) {
	// DO NOT LOCK HERE - this function is called from a locked function

	keys := b.nonceRanges[account][chainID].keys()

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	ranges := []nonceRange{}
	for _, k := range keys {
		r := b.nonceRanges[account][chainID].get(k)
		ranges = append(ranges, r)
	}

	_, exists := b.sortedRanges[account]
	if !exists {
		b.sortedRanges[account] = make(map[uint64][]nonceRange)
	}

	b.sortedRanges[account][chainID] = ranges
}

func (b *nonceRangeCache[T]) clear() {
	b.rw.Lock()
	defer b.rw.Unlock()

	b.nonceRanges = make(addressChainMap[T])
	b.sortedRanges = make(sortedNonceRangesCacheType)
}

func (b *nonceRangeCache[T]) size(account common.Address, chainID uint64) int {
	b.rw.RLock()
	defer b.rw.RUnlock()

	_, exists := b.nonceRanges[account]
	if !exists {
		return 0
	}

	_, exists = b.nonceRanges[account][chainID]
	if !exists {
		return 0
	}

	return b.nonceRanges[account][chainID].len()
}
