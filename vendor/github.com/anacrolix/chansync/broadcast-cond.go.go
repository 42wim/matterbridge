package chansync

import (
	"github.com/anacrolix/chansync/events"
	"github.com/anacrolix/sync"
)

// Can be used as zero-value. Due to the caller needing to bring their own synchronization, an
// equivalent to "sync".Cond.Signal is not provided. BroadcastCond is intended to be selected on
// with other channels.
type BroadcastCond struct {
	mu sync.Mutex
	ch chan struct{}
}

func (me *BroadcastCond) Broadcast() {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.ch != nil {
		close(me.ch)
		me.ch = nil
	}
}

// Should be called before releasing locks on resources that might trigger subsequent Broadcasts.
// The channel is closed when the condition changes.
func (me *BroadcastCond) Signaled() events.Signaled {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.ch == nil {
		me.ch = make(chan struct{})
	}
	return me.ch
}
