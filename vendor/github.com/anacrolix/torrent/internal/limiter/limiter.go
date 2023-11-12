package limiter

import "sync"

type Key = interface{}

// Manages resources with a limited number of concurrent slots for use for each key.
type Instance struct {
	SlotsPerKey int

	mu sync.Mutex
	// Limits concurrent use of a resource. Push into the channel to use a slot, and receive to free
	// up a slot.
	active map[Key]*activeValueType
}

type activeValueType struct {
	ch   chan struct{}
	refs int
}

type ActiveValueRef struct {
	v *activeValueType
	k Key
	i *Instance
}

// Returns the limiting channel. Send to it to obtain a slot, and receive to release the slot.
func (me ActiveValueRef) C() chan struct{} {
	return me.v.ch
}

// Drop the reference to a key, this allows keys to be reclaimed when they're no longer in use.
func (me ActiveValueRef) Drop() {
	me.i.mu.Lock()
	defer me.i.mu.Unlock()
	me.v.refs--
	if me.v.refs == 0 {
		delete(me.i.active, me.k)
	}
}

// Get a reference to the values for a key. You should make sure to call Drop exactly once on the
// returned value when done.
func (i *Instance) GetRef(key Key) ActiveValueRef {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.active == nil {
		i.active = make(map[Key]*activeValueType)
	}
	v, ok := i.active[key]
	if !ok {
		v = &activeValueType{
			ch: make(chan struct{}, i.SlotsPerKey),
		}
		i.active[key] = v
	}
	v.refs++
	return ActiveValueRef{
		v: v,
		k: key,
		i: i,
	}
}
