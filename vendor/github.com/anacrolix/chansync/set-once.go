package chansync

import (
	"sync"
	"sync/atomic"

	"github.com/anacrolix/chansync/events"
)

// SetOnce is a boolean value that can only be flipped from false to true.
type SetOnce struct {
	ch chan struct{}
	// Could be faster than trying to receive from ch
	closed    uint32
	initOnce  sync.Once
	closeOnce sync.Once
}

// Returns a channel that is closed when the event is flagged.
func (me *SetOnce) Done() events.Done {
	me.init()
	return me.ch
}

func (me *SetOnce) init() {
	me.initOnce.Do(func() {
		me.ch = make(chan struct{})
	})
}

// Set only returns true the first time it is called.
func (me *SetOnce) Set() (first bool) {
	me.closeOnce.Do(func() {
		me.init()
		first = true
		atomic.StoreUint32(&me.closed, 1)
		close(me.ch)
	})
	return
}

func (me *SetOnce) IsSet() bool {
	return atomic.LoadUint32(&me.closed) != 0
}
