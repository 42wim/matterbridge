package subscriptions

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type Subscriptions struct {
	mu          sync.Mutex
	subs        map[SubscriptionID]*Subscription
	checkPeriod time.Duration
	log         log.Logger
}

func NewSubscriptions(period time.Duration) *Subscriptions {
	return &Subscriptions{
		subs:        make(map[SubscriptionID]*Subscription),
		checkPeriod: period,
		log:         log.New("package", "status-go/services/subsriptions.Subscriptions"),
	}
}

func (s *Subscriptions) Create(namespace string, filter filter) (SubscriptionID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newSub := NewSubscription(namespace, filter)

	go func() {
		err := newSub.Start(s.checkPeriod)
		if err != nil {
			s.log.Error("error while starting subscription", "err", err)
		}
	}()

	s.subs[newSub.id] = newSub

	return newSub.id, nil
}

func (s *Subscriptions) Remove(id SubscriptionID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found, err := s.stopSubscription(id, true)
	if found {
		delete(s.subs, id)
	}

	return err
}

func (s *Subscriptions) removeAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	unsubscribeErrors := make(map[SubscriptionID]error)

	for id := range s.subs {
		_, err := s.stopSubscription(id, false)
		if err != nil {
			unsubscribeErrors[id] = err
		}
	}

	s.subs = make(map[SubscriptionID]*Subscription)

	if len(unsubscribeErrors) > 0 {
		return fmt.Errorf("errors while cleaning up subscriptions: %+v", unsubscribeErrors)
	}

	return nil
}

func (s *Subscriptions) stopSubscription(id SubscriptionID, uninstall bool) (bool, error) {
	sub, found := s.subs[id]
	if !found {
		return false, nil
	}
	return true, sub.Stop(uninstall)
}
