package melody

import (
	"sync"
	"sync/atomic"
)

type sessionSet struct {
	mu      sync.RWMutex
	members map[*Session]struct{}
}

func (ss *sessionSet) add(s *Session) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.members[s] = struct{}{}
}

func (ss *sessionSet) del(s *Session) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.members, s)
}

func (ss *sessionSet) clear() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.members = make(map[*Session]struct{})
}

func (ss *sessionSet) each(cb func(*Session)) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	for s := range ss.members {
		cb(s)
	}
}

func (ss *sessionSet) len() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return len(ss.members)
}

func (ss *sessionSet) all() []*Session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	s := make([]*Session, 0, len(ss.members))
	for k := range ss.members {
		s = append(s, k)
	}

	return s
}

type hub struct {
	sessions   sessionSet
	broadcast  chan envelope
	register   chan *Session
	unregister chan *Session
	exit       chan envelope
	open       atomic.Bool
}

func newHub() *hub {
	return &hub{
		sessions: sessionSet{
			members: make(map[*Session]struct{}),
		},
		broadcast:  make(chan envelope),
		register:   make(chan *Session),
		unregister: make(chan *Session),
		exit:       make(chan envelope),
	}
}

func (h *hub) run() {
	h.open.Store(true)

loop:
	for {
		select {
		case s := <-h.register:
			h.sessions.add(s)
		case s := <-h.unregister:
			h.sessions.del(s)
		case m := <-h.broadcast:
			h.sessions.each(func(s *Session) {
				if m.filter == nil {
					s.writeMessage(m)
				} else if m.filter(s) {
					s.writeMessage(m)
				}
			})
		case m := <-h.exit:
			h.open.Store(false)

			h.sessions.each(func(s *Session) {
				s.writeMessage(m)
				s.Close()
			})

			h.sessions.clear()

			break loop
		}
	}
}

func (h *hub) closed() bool {
	return !h.open.Load()
}

func (h *hub) len() int {
	return h.sessions.len()
}

func (h *hub) all() []*Session {
	return h.sessions.all()
}
