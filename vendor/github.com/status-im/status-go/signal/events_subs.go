package signal

const (
	// EventSubscriptionsData is triggered when there is new data in any of the subscriptions
	EventSubscriptionsData = "subscriptions.data"
	// EventSubscriptionsError is triggered when subscriptions failed to get new data
	EventSubscriptionsError = "subscriptions.error"
)

type SubscriptionDataEvent struct {
	FilterID string        `json:"subscription_id"`
	Data     []interface{} `json:"data"`
}

type SubscriptionErrorEvent struct {
	FilterID     string `json:"subscription_id"`
	ErrorMessage string `json:"error_message"`
}

// SendSubscriptionDataEvent
func SendSubscriptionDataEvent(filterID string, data []interface{}) {
	send(EventSubscriptionsData, SubscriptionDataEvent{
		FilterID: filterID,
		Data:     data,
	})
}

// SendSubscriptionErrorEvent
func SendSubscriptionErrorEvent(filterID string, err error) {
	send(EventSubscriptionsError, SubscriptionErrorEvent{
		ErrorMessage: err.Error(),
	})
}
