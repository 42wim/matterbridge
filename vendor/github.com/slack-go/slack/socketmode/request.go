package socketmode

import "encoding/json"

// Request maps to the content of each WebSocket message received via a Socket Mode WebSocket connection
//
// We call this a "request" rather than e.g. a WebSocket message or an Socket Mode "event" following python-slack-sdk:
//
// https://github.com/slackapi/python-slack-sdk/blob/3f1c4c6e27bf7ee8af57699b2543e6eb7848bcf9/slack_sdk/socket_mode/request.py#L6
//
// We know that node-slack-sdk calls it an "event", that makes it hard for us to distinguish our client's own event
// that wraps both internal events and Socket Mode "events", vs node-slack-sdk's is for the latter only.
//
// https://github.com/slackapi/node-slack-sdk/blob/main/packages/socket-mode/src/SocketModeClient.ts#L537
type Request struct {
	Type string `json:"type"`

	// `hello` type only
	NumConnections int            `json:"num_connections"`
	ConnectionInfo ConnectionInfo `json:"connection_info"`

	// `disconnect` type only

	// Reason can be "warning" or else
	Reason string `json:"reason"`

	// `hello` and `disconnect` types only
	DebugInfo DebugInfo `json:"debug_info"`

	// `events_api` type only
	EnvelopeID string `json:"envelope_id"`
	// TODO Can it really be a non-object type?
	// See https://github.com/slackapi/python-slack-sdk/blob/3f1c4c6e27bf7ee8af57699b2543e6eb7848bcf9/slack_sdk/socket_mode/request.py#L26-L31
	Payload                json.RawMessage `json:"payload"`
	AcceptsResponsePayload bool            `json:"accepts_response_payload"`
	RetryAttempt           int             `json:"retry_attempt"`
	RetryReason            string          `json:"retry_reason"`
}
