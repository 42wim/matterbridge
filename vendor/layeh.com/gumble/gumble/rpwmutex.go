package gumble

import "sync"

// rpwMutex is a reader-preferred RWMutex.
type rpwMutex struct {
	w sync.Mutex
	r sync.Mutex
	n int
}

func (m *rpwMutex) Lock() {
	m.w.Lock()
}

func (m *rpwMutex) Unlock() {
	m.w.Unlock()
}

func (m *rpwMutex) RLock() {
	m.r.Lock()
	m.n++
	if m.n == 1 {
		m.w.Lock()
	}
	m.r.Unlock()
}

func (m *rpwMutex) RUnlock() {
	m.r.Lock()
	m.n--
	if m.n == 0 {
		m.w.Unlock()
	}
	m.r.Unlock()
}
