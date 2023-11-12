package publisher

import (
	"encoding/hex"
	"sync"
)

type persistence struct {
	lastAcksMutex sync.Mutex
	lastPublished int64
	lastAcks      map[string]int64
}

func newPersistence() *persistence {
	return &persistence{
		lastAcks: make(map[string]int64),
	}
}

func (s *persistence) getLastPublished() int64 {
	return s.lastPublished
}

func (s *persistence) setLastPublished(lastPublished int64) {
	s.lastPublished = lastPublished
}

func (s *persistence) lastAck(identity []byte) int64 {
	s.lastAcksMutex.Lock()
	defer s.lastAcksMutex.Unlock()
	return s.lastAcks[hex.EncodeToString(identity)]
}

func (s *persistence) setLastAck(identity []byte, lastAck int64) {
	s.lastAcksMutex.Lock()
	defer s.lastAcksMutex.Unlock()
	s.lastAcks[hex.EncodeToString(identity)] = lastAck
}
