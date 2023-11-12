package signal

const (
	// EventSignRequestAdded is triggered when send transaction request is queued
	EventSignRequestAdded = "sign-request.queued"
	// EventSignRequestFailed is triggered when send transaction request fails
	EventSignRequestFailed = "sign-request.failed"
)

// PendingRequestEvent is a signal sent when a sign request is added
type PendingRequestEvent struct {
	ID        string      `json:"id"`
	Method    string      `json:"method"`
	Args      interface{} `json:"args"`
	MessageID string      `json:"message_id"`
}

// SendSignRequestAdded sends a signal when a sign request is added.
func SendSignRequestAdded(event PendingRequestEvent) {
	send(EventSignRequestAdded, event)
}

// PendingRequestErrorEvent is a signal sent when sign request has failed
type PendingRequestErrorEvent struct {
	PendingRequestEvent
	ErrorMessage string `json:"error_message"`
	ErrorCode    int    `json:"error_code,string"`
}

// SendSignRequestFailed sends a signal of failed sign request.
func SendSignRequestFailed(event PendingRequestEvent, err error, errCode int) {
	send(EventSignRequestFailed,
		PendingRequestErrorEvent{
			PendingRequestEvent: event,
			ErrorMessage:        err.Error(),
			ErrorCode:           errCode,
		})
}
