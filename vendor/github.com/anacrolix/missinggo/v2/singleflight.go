package missinggo

import "sync"

type ongoing struct {
	do    sync.Mutex
	users int
}

type SingleFlight struct {
	mu      sync.Mutex
	ongoing map[string]*ongoing
}

type Operation struct {
	sf *SingleFlight
	id string
}

func (op Operation) Unlock() {
	op.sf.Unlock(op.id)
}

func (me *SingleFlight) Lock(id string) Operation {
	me.mu.Lock()
	on, ok := me.ongoing[id]
	if !ok {
		on = new(ongoing)
		if me.ongoing == nil {
			me.ongoing = make(map[string]*ongoing)
		}
		me.ongoing[id] = on
	}
	on.users++
	me.mu.Unlock()
	on.do.Lock()
	return Operation{me, id}
}

func (me *SingleFlight) Unlock(id string) {
	me.mu.Lock()
	on := me.ongoing[id]
	on.do.Unlock()
	on.users--
	if on.users == 0 {
		delete(me.ongoing, id)
	}
	me.mu.Unlock()
}
