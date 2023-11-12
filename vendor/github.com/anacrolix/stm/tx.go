package stm

import (
	"fmt"
	"sort"
	"sync"
	"unsafe"
)

// A Tx represents an atomic transaction.
type Tx struct {
	reads          map[*Var]VarValue
	writes         map[*Var]interface{}
	watching       map[*Var]struct{}
	locks          txLocks
	mu             sync.Mutex
	cond           sync.Cond
	waiting        bool
	completed      bool
	tries          int
	numRetryValues int
}

// Check that none of the logged values have changed since the transaction began.
func (tx *Tx) inputsChanged() bool {
	for v, read := range tx.reads {
		if read.Changed(v.value.Load().(VarValue)) {
			return true
		}
	}
	return false
}

// Writes the values in the transaction log to their respective Vars.
func (tx *Tx) commit() {
	for v, val := range tx.writes {
		v.changeValue(val)
	}
}

func (tx *Tx) updateWatchers() {
	for v := range tx.watching {
		if _, ok := tx.reads[v]; !ok {
			delete(tx.watching, v)
			v.watchers.Delete(tx)
		}
	}
	for v := range tx.reads {
		if _, ok := tx.watching[v]; !ok {
			v.watchers.Store(tx, nil)
			tx.watching[v] = struct{}{}
		}
	}
}

// wait blocks until another transaction modifies any of the Vars read by tx.
func (tx *Tx) wait() {
	if len(tx.reads) == 0 {
		panic("not waiting on anything")
	}
	tx.updateWatchers()
	tx.mu.Lock()
	firstWait := true
	for !tx.inputsChanged() {
		if !firstWait {
			expvars.Add("wakes for unchanged versions", 1)
		}
		expvars.Add("waits", 1)
		tx.waiting = true
		tx.cond.Broadcast()
		tx.cond.Wait()
		tx.waiting = false
		firstWait = false
	}
	tx.mu.Unlock()
}

// Get returns the value of v as of the start of the transaction.
func (tx *Tx) Get(v *Var) interface{} {
	// If we previously wrote to v, it will be in the write log.
	if val, ok := tx.writes[v]; ok {
		return val
	}
	// If we haven't previously read v, record its version
	vv, ok := tx.reads[v]
	if !ok {
		vv = v.value.Load().(VarValue)
		tx.reads[v] = vv
	}
	return vv.Get()
}

// Set sets the value of a Var for the lifetime of the transaction.
func (tx *Tx) Set(v *Var, val interface{}) {
	if v == nil {
		panic("nil Var")
	}
	tx.writes[v] = val
}

type txProfileValue struct {
	*Tx
	int
}

// Retry aborts the transaction and retries it when a Var changes. You can return from this method
// to satisfy return values, but it should never actually return anything as it panics internally.
func (tx *Tx) Retry() interface{} {
	retries.Add(txProfileValue{tx, tx.numRetryValues}, 1)
	tx.numRetryValues++
	panic(retry)
	panic("unreachable")
}

// Assert is a helper function that retries a transaction if the condition is
// not satisfied.
func (tx *Tx) Assert(p bool) {
	if !p {
		tx.Retry()
	}
}

func (tx *Tx) reset() {
	tx.mu.Lock()
	for k := range tx.reads {
		delete(tx.reads, k)
	}
	for k := range tx.writes {
		delete(tx.writes, k)
	}
	tx.mu.Unlock()
	tx.removeRetryProfiles()
	tx.resetLocks()
}

func (tx *Tx) removeRetryProfiles() {
	for tx.numRetryValues > 0 {
		tx.numRetryValues--
		retries.Remove(txProfileValue{tx, tx.numRetryValues})
	}
}

func (tx *Tx) recycle() {
	for v := range tx.watching {
		delete(tx.watching, v)
		v.watchers.Delete(tx)
	}
	tx.removeRetryProfiles()
	// I don't think we can reuse Txs, because the "completed" field should/needs to be set
	// indefinitely after use.
	//txPool.Put(tx)
}

func (tx *Tx) lockAllVars() {
	tx.resetLocks()
	tx.collectAllLocks()
	tx.sortLocks()
	tx.lock()
}

func (tx *Tx) resetLocks() {
	tx.locks.clear()
}

func (tx *Tx) collectReadLocks() {
	for v := range tx.reads {
		tx.locks.append(&v.mu)
	}
}

func (tx *Tx) collectAllLocks() {
	tx.collectReadLocks()
	for v := range tx.writes {
		if _, ok := tx.reads[v]; !ok {
			tx.locks.append(&v.mu)
		}
	}
}

func (tx *Tx) sortLocks() {
	sort.Sort(&tx.locks)
}

func (tx *Tx) lock() {
	for _, l := range tx.locks.mus {
		l.Lock()
	}
}

func (tx *Tx) unlock() {
	for _, l := range tx.locks.mus {
		l.Unlock()
	}
}

func (tx *Tx) String() string {
	return fmt.Sprintf("%[1]T %[1]p", tx)
}

// Dedicated type avoids reflection in sort.Slice.
type txLocks struct {
	mus []*sync.Mutex
}

func (me txLocks) Len() int {
	return len(me.mus)
}

func (me txLocks) Less(i, j int) bool {
	return uintptr(unsafe.Pointer(me.mus[i])) < uintptr(unsafe.Pointer(me.mus[j]))
}

func (me txLocks) Swap(i, j int) {
	me.mus[i], me.mus[j] = me.mus[j], me.mus[i]
}

func (me *txLocks) clear() {
	me.mus = me.mus[:0]
}

func (me *txLocks) append(mu *sync.Mutex) {
	me.mus = append(me.mus, mu)
}
