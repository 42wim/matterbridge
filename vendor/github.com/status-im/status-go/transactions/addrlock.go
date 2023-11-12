// copy of go-ethereum/internal/ethapi/addrlock.go

package transactions

import (
	"sync"

	"github.com/status-im/status-go/eth-node/types"
)

// AddrLocker provides locks for addresses
type AddrLocker struct {
	mu    sync.Mutex
	locks map[types.Address]*sync.Mutex
}

// lock returns the lock of the given address.
func (l *AddrLocker) lock(address types.Address) *sync.Mutex {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.locks == nil {
		l.locks = make(map[types.Address]*sync.Mutex)
	}
	if _, ok := l.locks[address]; !ok {
		l.locks[address] = new(sync.Mutex)
	}
	return l.locks[address]
}

// LockAddr locks an account's mutex. This is used to prevent another tx getting the
// same nonce until the lock is released. The mutex prevents the (an identical nonce) from
// being read again during the time that the first transaction is being signed.
func (l *AddrLocker) LockAddr(address types.Address) {
	l.lock(address).Lock()
}

// UnlockAddr unlocks the mutex of the given account.
func (l *AddrLocker) UnlockAddr(address types.Address) {
	l.lock(address).Unlock()
}
