package connection

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

// Client expects a single event with all states
type StatusNotification map[string]State // id -> State

type StatusNotifier struct {
	statuses  map[string]*Status // id -> Status
	eventType walletevent.EventType
	feed      *event.Feed
}

func NewStatusNotifier(statuses map[string]*Status, eventType walletevent.EventType, feed *event.Feed) *StatusNotifier {
	n := StatusNotifier{
		statuses:  statuses,
		eventType: eventType,
		feed:      feed,
	}

	for _, status := range statuses {
		status.SetStateChangeCb(n.notify)
	}

	return &n
}

func (n *StatusNotifier) notify(state State) {
	// state is ignored, as client expects all valid states in
	// a single event, so we fetch them from the map
	if n.feed != nil {
		statusMap := make(StatusNotification)
		for id, status := range n.statuses {
			state := status.GetState()
			if state.Value == StateValueUnknown {
				continue
			}
			statusMap[id] = state
		}

		encodedMessage, err := json.Marshal(statusMap)
		if err != nil {
			return
		}

		n.feed.Send(walletevent.Event{
			Type:     n.eventType,
			Accounts: []common.Address{},
			Message:  string(encodedMessage),
			At:       time.Now().Unix(),
		})
	}
}
