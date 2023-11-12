package chansync

import (
	"sync"

	"github.com/anacrolix/chansync/events"
)

type LevelTrigger struct {
	ch       chan struct{}
	initOnce sync.Once
}

func (me *LevelTrigger) Signal() events.Signal {
	me.init()
	return me.ch
}

func (me *LevelTrigger) Active() events.Active {
	me.init()
	return me.ch
}

func (me *LevelTrigger) init() {
	me.initOnce.Do(func() {
		me.ch = make(chan struct{})
	})
}
