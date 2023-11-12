package signal

const (
	notificationEvent = "local-notifications"
)

// SendLocalNotifications sends event with a local notification.
func SendLocalNotifications(event interface{}) {
	send(notificationEvent, event)
}
