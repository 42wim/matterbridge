package store

import (
	"errors"
	"sync"

	"github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"
)

type DummyStore struct {
	sync.Mutex
	ms map[state.MessageID]*protobuf.Message
}

func NewDummyStore() *DummyStore {
	return &DummyStore{ms: make(map[state.MessageID]*protobuf.Message)}
}

func (ds *DummyStore) Has(id state.MessageID) (bool, error) {
	ds.Lock()
	defer ds.Unlock()

	_, ok := ds.ms[id]
	return ok, nil
}

func (ds *DummyStore) Get(id state.MessageID) (*protobuf.Message, error) {
	ds.Lock()
	defer ds.Unlock()

	m, ok := ds.ms[id]
	if !ok {
		return nil, errors.New("message does not exist")
	}

	return m, nil
}

func (ds *DummyStore) Add(message *protobuf.Message) error {
	ds.Lock()
	defer ds.Unlock()
	ds.ms[message.ID()] = message
	return nil
}
