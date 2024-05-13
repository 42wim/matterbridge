package balance

import (
	"context"
	"math/big"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Reader interface for reading balance at a specified address.
type Reader interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	CallBlockHashByTransaction(ctx context.Context, blockNumber *big.Int, index uint) (common.Hash, error)
	NetworkID() uint64
}

// Cacher interface for caching balance to BalanceCache. Requires BalanceReader to fetch balance.
type Cacher interface {
	BalanceAt(ctx context.Context, client Reader, account common.Address, blockNumber *big.Int) (*big.Int, error)
	NonceAt(ctx context.Context, client Reader, account common.Address, blockNumber *big.Int) (*int64, error)
	Clear()
	Cache() CacheIface
}

// Interface for cache of balances.
type CacheIface interface {
	GetBalance(account common.Address, chainID uint64, blockNumber *big.Int) *big.Int
	GetNonce(account common.Address, chainID uint64, blockNumber *big.Int) *int64
	AddBalance(account common.Address, chainID uint64, blockNumber *big.Int, balance *big.Int)
	AddNonce(account common.Address, chainID uint64, blockNumber *big.Int, nonce *int64)
	BalanceSize(account common.Address, chainID uint64) int
	NonceSize(account common.Address, chainID uint64) int
	Clear()
}

type addressChainMap[T any] map[common.Address]map[uint64]T // address->chainID

type cacheIface[K comparable, V any] interface {
	get(K) V
	set(K, V)
	len() int
	keys() []K
	clear()
	init()
}

// genericCache is a generic implementation of CacheIface
type genericCache[B cacheIface[uint64, *big.Int], N cacheIface[uint64, *int64], NR cacheIface[int64, nonceRange]] struct {
	nonceRangeCache[NR]

	// balances maps an address and chain to a cache of a block number and the balance of this particular address on the chain
	balances addressChainMap[B]
	nonces   addressChainMap[N]
	rw       sync.RWMutex
}

func (b *genericCache[_, _, _]) GetBalance(account common.Address, chainID uint64, blockNumber *big.Int) *big.Int {
	b.rw.RLock()
	defer b.rw.RUnlock()

	_, exists := b.balances[account]
	if !exists {
		return nil
	}

	_, exists = b.balances[account][chainID]
	if !exists {
		return nil
	}

	return b.balances[account][chainID].get(blockNumber.Uint64())
}

func (b *genericCache[B, _, _]) AddBalance(account common.Address, chainID uint64, blockNumber *big.Int, balance *big.Int) {
	b.rw.Lock()
	defer b.rw.Unlock()

	_, exists := b.balances[account]
	if !exists {
		b.balances[account] = make(map[uint64]B)
	}

	_, exists = b.balances[account][chainID]
	if !exists {
		b.balances[account][chainID] = reflect.New(reflect.TypeOf(b.balances[account][chainID]).Elem()).Interface().(B)
		b.balances[account][chainID].init()
	}

	b.balances[account][chainID].set(blockNumber.Uint64(), balance)
}

func (b *genericCache[_, _, _]) GetNonce(account common.Address, chainID uint64, blockNumber *big.Int) *int64 {
	b.rw.RLock()
	defer b.rw.RUnlock()

	_, exists := b.nonces[account]
	if !exists {
		return nil
	}

	_, exists = b.nonces[account][chainID]
	if !exists {
		return nil
	}

	nonce := b.nonces[account][chainID].get(blockNumber.Uint64())
	if nonce != nil {
		return nonce
	}

	return b.findNonceInRange(account, chainID, blockNumber)
}

func (b *genericCache[_, N, _]) AddNonce(account common.Address, chainID uint64, blockNumber *big.Int, nonce *int64) {
	b.rw.Lock()
	defer b.rw.Unlock()

	_, exists := b.nonces[account]
	if !exists {
		b.nonces[account] = make(map[uint64]N)
	}

	_, exists = b.nonces[account][chainID]
	if !exists {
		b.nonces[account][chainID] = reflect.New(reflect.TypeOf(b.nonces[account][chainID]).Elem()).Interface().(N)
		b.nonces[account][chainID].init()
	}

	b.nonces[account][chainID].set(blockNumber.Uint64(), nonce)
	b.updateNonceRange(account, chainID, blockNumber, nonce)
}

func (b *genericCache[_, _, _]) BalanceSize(account common.Address, chainID uint64) int {
	b.rw.RLock()
	defer b.rw.RUnlock()

	_, exists := b.balances[account]
	if !exists {
		return 0
	}

	_, exists = b.balances[account][chainID]
	if !exists {
		return 0
	}

	return b.balances[account][chainID].len()
}

func (b *genericCache[_, N, _]) NonceSize(account common.Address, chainID uint64) int {
	b.rw.RLock()
	defer b.rw.RUnlock()

	_, exists := b.nonces[account]
	if !exists {
		return 0
	}

	_, exists = b.nonces[account][chainID]
	if !exists {
		return 0
	}

	return b.nonces[account][chainID].len()
}

// implements Cacher interface that caches balance and nonce in memory.
type cacherImpl struct {
	cache CacheIface
}

func newCacherImpl(cache CacheIface) *cacherImpl {
	return &cacherImpl{
		cache: cache,
	}
}

func (b *cacherImpl) BalanceAt(ctx context.Context, client Reader, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	cachedBalance := b.cache.GetBalance(account, client.NetworkID(), blockNumber)
	if cachedBalance != nil {
		return cachedBalance, nil
	}

	balance, err := client.BalanceAt(ctx, account, blockNumber)
	if err != nil {
		return nil, err
	}
	b.cache.AddBalance(account, client.NetworkID(), blockNumber, balance)
	return balance, nil
}

func (b *cacherImpl) NonceAt(ctx context.Context, client Reader, account common.Address, blockNumber *big.Int) (*int64, error) {
	cachedNonce := b.cache.GetNonce(account, client.NetworkID(), blockNumber)
	if cachedNonce != nil {
		return cachedNonce, nil
	}

	nonce, err := client.NonceAt(ctx, account, blockNumber)
	if err != nil {
		return nil, err
	}
	int64Nonce := int64(nonce)
	b.cache.AddNonce(account, client.NetworkID(), blockNumber, &int64Nonce)

	return &int64Nonce, nil
}

func (b *cacherImpl) Clear() {
	b.cache.Clear()
}

func (b *cacherImpl) Cache() CacheIface {
	return b.cache
}
