package subscriptions

import "github.com/status-im/status-go/signal"

type filterSignal struct {
	filterID string
}

func newFilterSignal(filterID string) *filterSignal {
	return &filterSignal{filterID}
}

func (s *filterSignal) SendError(err error) {
	signal.SendSubscriptionErrorEvent(s.filterID, err)
}

func (s *filterSignal) SendData(data []interface{}) {
	signal.SendSubscriptionDataEvent(s.filterID, data)
}
