package chansync

import (
	"sync"

	"github.com/anacrolix/chansync/events"
)

// Flag wraps a boolean value that starts as false (off). You can wait for it to be on or off, and
// set the value as needed.
type Flag struct {
	mu     sync.Mutex
	on     chan struct{}
	off    chan struct{}
	state  bool
	inited bool
}

func (me *Flag) Bool() bool {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.state
}

func (me *Flag) On() events.Active {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.init()
	return me.on
}

func (me *Flag) Off() events.Active {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.init()
	return me.off
}

func (me *Flag) init() {
	if me.inited {
		return
	}
	me.on = make(chan struct{})
	me.off = make(chan struct{})
	close(me.off)
	me.inited = true
}

func (me *Flag) SetBool(b bool) {
	if b {
		me.Set()
	} else {
		me.Clear()
	}
}

func (me *Flag) Set() {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.init()
	if me.state {
		return
	}
	me.state = true
	close(me.on)
	me.off = make(chan struct{})
}

func (me *Flag) Clear() {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.init()
	if !me.state {
		return
	}
	me.state = false
	close(me.off)
	me.on = make(chan struct{})
}
