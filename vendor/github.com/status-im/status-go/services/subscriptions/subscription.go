package subscriptions

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type SubscriptionID string

type Subscription struct {
	mu      sync.RWMutex
	id      SubscriptionID
	signal  *filterSignal
	quit    chan struct{}
	filter  filter
	started bool
}

func NewSubscription(namespace string, filter filter) *Subscription {
	subscriptionID := NewSubscriptionID(namespace, filter.getID())
	return &Subscription{
		id:     subscriptionID,
		signal: newFilterSignal(string(subscriptionID)),
		filter: filter,
	}
}

func (s *Subscription) Start(checkPeriod time.Duration) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("subscription already started or used")
	}
	s.started = true
	s.quit = make(chan struct{})
	quit := s.quit
	s.mu.Unlock()

	ticker := time.NewTicker(checkPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			filterData, err := s.filter.getChanges()
			if err != nil {
				s.signal.SendError(err)
			} else if len(filterData) > 0 {
				s.signal.SendData(filterData)
			}
		case <-quit:
			return nil
		}
	}
}

func (s *Subscription) Stop(uninstall bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return nil
	}
	select {
	case _, ok := <-s.quit:
		// handle a case of a closed channel
		if !ok {
			return nil
		}
	default:
		close(s.quit)
	}
	if !uninstall {
		return nil
	}
	return s.filter.uninstall()
}

func NewSubscriptionID(namespace, filterID string) SubscriptionID {
	return SubscriptionID(fmt.Sprintf("%s-%s", namespace, filterID))
}
