package perf

import (
	"sync"

	"github.com/anacrolix/missinggo"
)

type TimedLocker struct {
	L    sync.Locker
	Desc string
}

func (me *TimedLocker) Lock() {
	tr := NewTimer()
	me.L.Lock()
	tr.Mark(me.Desc)
}

func (me *TimedLocker) Unlock() {
	me.L.Unlock()
}

type TimedRWLocker struct {
	RWL       missinggo.RWLocker
	WriteDesc string
	ReadDesc  string
}

func (me *TimedRWLocker) Lock() {
	tr := NewTimer()
	me.RWL.Lock()
	tr.Mark(me.WriteDesc)
}

func (me *TimedRWLocker) Unlock() {
	me.RWL.Unlock()
}

func (me *TimedRWLocker) RLock() {
	tr := NewTimer()
	me.RWL.RLock()
	tr.Mark(me.ReadDesc)
}

func (me *TimedRWLocker) RUnlock() {
	me.RWL.RUnlock()
}
