package melody

import (
	"sync"
)

type hub struct {
	sessions   map[*Session]bool
	broadcast  chan *envelope
	register   chan *Session
	unregister chan *Session
	exit       chan *envelope
	open       bool
	rwmutex    *sync.RWMutex
}

func newHub() *hub {
	return &hub{
		sessions:   make(map[*Session]bool),
		broadcast:  make(chan *envelope),
		register:   make(chan *Session),
		unregister: make(chan *Session),
		exit:       make(chan *envelope),
		open:       true,
		rwmutex:    &sync.RWMutex{},
	}
}

func (h *hub) run() {
loop:
	for {
		select {
		case s := <-h.register:
			h.rwmutex.Lock()
			h.sessions[s] = true
			h.rwmutex.Unlock()
		case s := <-h.unregister:
			if _, ok := h.sessions[s]; ok {
				h.rwmutex.Lock()
				delete(h.sessions, s)
				h.rwmutex.Unlock()
			}
		case m := <-h.broadcast:
			h.rwmutex.RLock()
			for s := range h.sessions {
				if m.filter != nil {
					if m.filter(s) {
						s.writeMessage(m)
					}
				} else {
					s.writeMessage(m)
				}
			}
			h.rwmutex.RUnlock()
		case m := <-h.exit:
			h.rwmutex.Lock()
			for s := range h.sessions {
				s.writeMessage(m)
				delete(h.sessions, s)
				s.Close()
			}
			h.open = false
			h.rwmutex.Unlock()
			break loop
		}
	}
}

func (h *hub) closed() bool {
	h.rwmutex.RLock()
	defer h.rwmutex.RUnlock()
	return !h.open
}

func (h *hub) len() int {
	h.rwmutex.RLock()
	defer h.rwmutex.RUnlock()

	return len(h.sessions)
}
