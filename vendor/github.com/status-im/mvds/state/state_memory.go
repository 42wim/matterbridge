package state

import (
	"sync"
)

type memorySyncState struct {
	sync.Mutex

	state []State
}

func NewSyncState() *memorySyncState {
	return &memorySyncState{}
}

func (s *memorySyncState) Add(newState State) error {
	s.Lock()
	defer s.Unlock()

	s.state = append(s.state, newState)

	return nil
}

func (s *memorySyncState) Remove(id MessageID, peer PeerID) error {
	s.Lock()
	defer s.Unlock()
	var newState []State

	for _, state := range s.state {
		if state.MessageID != id || state.PeerID != peer {
			newState = append(newState, state)
		}
	}

	s.state = newState

	return nil
}

func (s *memorySyncState) All(_ int64) ([]State, error) {
	s.Lock()
	defer s.Unlock()
	return s.state, nil
}

func (s *memorySyncState) Map(epoch int64, process func(State) State) error {
	s.Lock()
	defer s.Unlock()

	for i, state := range s.state {
		if state.SendEpoch > epoch {
			continue
		}

		s.state[i] = process(state)
	}

	return nil
}
