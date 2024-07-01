package socketmode

import (
	"encoding/json"
	"time"

	"github.com/slack-go/slack"

	"github.com/gorilla/websocket"
)

type ConnectedEvent struct {
	ConnectionCount int // 1 = first time, 2 = second time
	Info            *slack.SocketModeConnection
}

type DebugInfo struct {
	// Host is the name of the host name on the Slack end, that can be something like `applink-7fc4fdbb64-4x5xq`
	Host string `json:"host"`

	// `hello` type only
	BuildNumber               int `json:"build_number"`
	ApproximateConnectionTime int `json:"approximate_connection_time"`
}

type ConnectionInfo struct {
	AppID string `json:"app_id"`
}

type SocketModeMessagePayload struct {
	Event json.RawMessage `json:"event"`
}

// Client is a Socket Mode client that allows programs to use [Events API](https://api.slack.com/events-api)
// and [interactive components](https://api.slack.com/interactivity) over WebSocket.
// Please see [Intro to Socket Mode](https://api.slack.com/apis/connections/socket) for more information
// on Socket Mode.
//
// The implementation is highly inspired by https://www.npmjs.com/package/@slack/socket-mode,
// but the structure and the design has been adapted as much as possible to that of our RTM client for consistency
// within the library.
//
// You can instantiate the socket mode client with
// Client's New() and call Run() to start it. Please see examples/socketmode for the usage.
type Client struct {
	// Client is the main API, embedded
	slack.Client

	// maxPingInterval is the maximum duration elapsed after the last WebSocket PING sent from Slack
	// until Client considers the WebSocket connection is dead and needs to be reopened.
	maxPingInterval time.Duration

	// Connection life-cycle
	Events              chan Event
	socketModeResponses chan *Response

	// dialer is a gorilla/websocket Dialer. If nil, use the default
	// Dialer.
	dialer *websocket.Dialer

	debug bool
	log   ilogger
}
