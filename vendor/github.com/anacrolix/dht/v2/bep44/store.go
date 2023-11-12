package bep44

import (
	"errors"
	"time"
)

var ErrItemNotFound = errors.New("item not found")

type Store interface {
	Put(*Item) error
	Get(Target) (*Item, error)
	Del(Target) error
}

// Wrapper is in charge of validate all new items and
// decide when to store, or ignore them depending of the BEP 44 definition.
// It is also in charge of removing expired items.
type Wrapper struct {
	s   Store
	exp time.Duration
}

func NewWrapper(s Store, exp time.Duration) *Wrapper {
	return &Wrapper{s: s, exp: exp}
}

func (w *Wrapper) Put(i *Item) error {
	if err := Check(i); err != nil {
		return err
	}

	is, err := w.s.Get(i.Target())
	if errors.Is(err, ErrItemNotFound) {
		i.created = time.Now().Local()
		return w.s.Put(i)
	}
	if err != nil {
		return err
	}

	if err := CheckIncoming(is, i); err != nil {
		return err
	}

	i.created = time.Now().Local()
	return w.s.Put(i)
}

func (w *Wrapper) Get(t Target) (*Item, error) {
	i, err := w.s.Get(t)
	if err != nil {
		return nil, err
	}

	if i.created.Add(w.exp).After(time.Now().Local()) {
		return i, nil
	}

	if err := w.s.Del(t); err != nil {
		return nil, err
	}

	return nil, ErrItemNotFound
}
