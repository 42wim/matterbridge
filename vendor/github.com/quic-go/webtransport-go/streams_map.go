package webtransport

import (
	"sync"

	"github.com/quic-go/quic-go"
)

type closeFunc func()

// The streamsMap manages the streams of a single QUIC connection.
// Note that several WebTransport sessions can share one QUIC connection.
type streamsMap struct {
	mx sync.Mutex
	m  map[quic.StreamID]closeFunc
}

func newStreamsMap() *streamsMap {
	return &streamsMap{m: make(map[quic.StreamID]closeFunc)}
}

func (s *streamsMap) AddStream(id quic.StreamID, close closeFunc) {
	s.mx.Lock()
	s.m[id] = close
	s.mx.Unlock()
}

func (s *streamsMap) RemoveStream(id quic.StreamID) {
	s.mx.Lock()
	delete(s.m, id)
	s.mx.Unlock()
}

func (s *streamsMap) CloseSession() {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, cl := range s.m {
		cl()
	}
	s.m = nil
}
