package socketmode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/internal/backoff"
	"github.com/slack-go/slack/internal/misc"
	"github.com/slack-go/slack/slackevents"

	"github.com/gorilla/websocket"
	"github.com/slack-go/slack/internal/timex"
)

// Run is a blocking function that connects the Slack Socket Mode API and handles all incoming
// requests and outgoing responses.
//
// The consumer of the Client and this function should read the Client.Events channel to receive
// `socketmode.Event`s that includes the client-specific events that may or may not wrap Socket Mode requests.
//
// Note that this function automatically reconnect on requested by Slack through a `disconnect` message.
// This function exists with an error only when a reconnection is failued due to some reason.
// If you want to retry even on reconnection failure, you'd need to write your own wrapper for this function
// to do so.
func (smc *Client) Run() error {
	return smc.RunContext(context.TODO())
}

// RunContext is a blocking function that connects the Slack Socket Mode API and handles all incoming
// requests and outgoing responses.
//
// The consumer of the Client and this function should read the Client.Events channel to receive
// `socketmode.Event`s that includes the client-specific events that may or may not wrap Socket Mode requests.
//
// Note that this function automatically reconnect on requested by Slack through a `disconnect` message.
// This function exists with an error only when a reconnection is failued due to some reason.
// If you want to retry even on reconnection failure, you'd need to write your own wrapper for this function
// to do so.
func (smc *Client) RunContext(ctx context.Context) error {
	for connectionCount := 0; ; connectionCount++ {
		if err := smc.run(ctx, connectionCount); err != nil {
			return err
		}

		// Continue and run the loop again to reconnect
	}
}

func (smc *Client) run(ctx context.Context, connectionCount int) error {
	messages := make(chan json.RawMessage)
	defer close(messages)

	deadmanTimer := newDeadmanTimer(smc.maxPingInterval)

	pingHandler := func(_ string) error {
		deadmanTimer.Reset()

		return nil
	}

	// Start trying to connect
	// the returned err is already passed onto the Events channel
	//
	// We also configures an additional ping handler for the deadmanTimer that triggers a timeout when
	// Slack did not send us WebSocket PING for more than Client.maxPingInterval.
	// We can use `<-smc.pingTimeout.C` to wait for the timeout.
	info, conn, err := smc.connect(ctx, connectionCount, pingHandler)
	if err != nil {
		// when the connection is unsuccessful its fatal, and we need to bail out.
		smc.Debugf("Failed to connect with Socket Mode on try %d: %s", connectionCount, err)

		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	smc.Events <- newEvent(EventTypeConnected, &ConnectedEvent{
		ConnectionCount: connectionCount,
		Info:            info,
	})

	smc.Debugf("WebSocket connection succeeded on try %d", connectionCount)

	// We're now connected so we can set up listeners

	var (
		wg           sync.WaitGroup
		firstErr     error
		firstErrOnce sync.Once
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		// The response sender sends Socket Mode responses over the WebSocket conn
		if err := smc.runResponseSender(ctx, conn); err != nil {
			firstErrOnce.Do(func() {
				firstErr = err
			})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		// The handler reads Socket Mode requests, and enqueues responses for sending by the response sender
		if err := smc.runRequestHandler(ctx, messages); err != nil {
			firstErrOnce.Do(func() {
				firstErr = err
			})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		// The receiver reads WebSocket messages, and enqueues parsed Socket Mode requests to be handled by
		// the request handler
		if err := smc.runMessageReceiver(ctx, conn, messages); err != nil {
			firstErrOnce.Do(func() {
				firstErr = err
			})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			// Detect when the connection is dead and try close connection.
			if err = conn.Close(); err != nil {
				smc.Debugf("Failed to close connection: %v", err)
			}
		case <-deadmanTimer.Elapsed():
			firstErrOnce.Do(func() {
				firstErr = errors.New("ping timeout: Slack did not send us WebSocket PING for more than Client.maxInterval")
			})

			cancel()
		}
	}()

	wg.Wait()

	if firstErr == context.Canceled {
		return firstErr
	}

	// wg.Wait() finishes only after any of the above go routines finishes.
	// Also, we can expect firstErr to be not nil, as goroutines can finish only on error.
	smc.Debugf("Reconnecting due to %v", firstErr)

	return nil
}

// connect attempts to connect to the slack websocket API. It handles any
// errors that occur while connecting and will return once a connection
// has been successfully opened.
func (smc *Client) connect(ctx context.Context, connectionCount int, additionalPingHandler func(string) error) (*slack.SocketModeConnection, *websocket.Conn, error) {
	const (
		errInvalidAuth      = "invalid_auth"
		errInactiveAccount  = "account_inactive"
		errMissingAuthToken = "not_authed"
		errTokenRevoked     = "token_revoked"
	)

	// used to provide exponential backoff wait time with jitter before trying
	// to connect to slack again
	boff := &backoff.Backoff{
		Max: 5 * time.Minute,
	}

	for {
		var (
			backoff time.Duration
		)

		// send connecting event
		smc.Events <- newEvent(EventTypeConnecting, &slack.ConnectingEvent{
			Attempt:         boff.Attempts() + 1,
			ConnectionCount: connectionCount,
		})

		// attempt to start the connection
		info, conn, err := smc.openAndDial(ctx, additionalPingHandler)
		if err == nil {
			return info, conn, nil
		}

		// check for fatal errors
		switch err.Error() {
		case errInvalidAuth, errInactiveAccount, errMissingAuthToken, errTokenRevoked:
			smc.Debugf("invalid auth when connecting with SocketMode: %s", err)
			return nil, nil, err
		default:
		}

		switch actual := err.(type) {
		case misc.StatusCodeError:
			if actual.Code == http.StatusNotFound {
				smc.Debugf("invalid auth when connecting with Socket Mode: %s", err)
				smc.Events <- newEvent(EventTypeInvalidAuth, &slack.InvalidAuthEvent{})
				return nil, nil, err
			}
		case *slack.RateLimitedError:
			backoff = actual.RetryAfter
		default:
		}

		backoff = timex.Max(backoff, boff.Duration())
		// any other errors are treated as recoverable and we try again after
		// sending the event along the Events channel
		smc.Events <- newEvent(EventTypeConnectionError, &slack.ConnectionErrorEvent{
			Attempt:  boff.Attempts(),
			Backoff:  backoff,
			ErrorObj: err,
		})

		// get time we should wait before attempting to connect again
		smc.Debugf("reconnection %d failed: %s reconnecting in %v\n", boff.Attempts(), err, backoff)

		// wait for one of the following to occur,
		// backoff duration has elapsed, disconnectCh is signalled, or
		// the smc finishes disconnecting.
		select {
		case <-time.After(backoff): // retry after the backoff.
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}
}

// openAndDial attempts to open a Socket Mode connection and dial to the connection endpoint using WebSocket.
// It returns the  full information returned by the "apps.connections.open" method on the
// Slack API.
func (smc *Client) openAndDial(ctx context.Context, additionalPingHandler func(string) error) (info *slack.SocketModeConnection, _ *websocket.Conn, err error) {
	var (
		url string
	)

	smc.Debugf("Starting SocketMode")
	info, url, err = smc.OpenContext(ctx)

	if err != nil {
		smc.Debugf("Failed to start or connect with SocketMode: %s", err)
		return nil, nil, err
	}

	smc.Debugf("Dialing to websocket on url %s", url)
	// Only use HTTPS for connections to prevent MITM attacks on the connection.
	upgradeHeader := http.Header{}
	upgradeHeader.Add("Origin", "https://api.slack.com")
	dialer := websocket.DefaultDialer
	if smc.dialer != nil {
		dialer = smc.dialer
	}
	conn, _, err := dialer.DialContext(ctx, url, upgradeHeader)
	if err != nil {
		smc.Debugf("Failed to dial to the websocket: %s", err)
		return nil, nil, err
	}

	conn.SetPingHandler(func(appData string) error {
		if additionalPingHandler != nil {
			if err := additionalPingHandler(appData); err != nil {
				return err
			}
		}

		smc.handlePing(conn, appData)

		return nil
	})

	// We don't need to conn.SetCloseHandler because the default handler is effective enough that
	// it sends back the CLOSE message to the server and let conn.ReadJSON() fail with CloseError.
	// The CloseError must be handled normally in our receiveMessagesInto function.
	//conn.SetCloseHandler(func(code int, text string) error {
	//  ...
	// })

	return info, conn, err
}

// runResponseSender runs the handler that reads Socket Mode responses enqueued onto Client.socketModeResponses channel
// and sends them one by one over the WebSocket connection.
// Gorilla WebSocket is not goroutine safe hence this needs to be the single place you write to the WebSocket connection.
func (smc *Client) runResponseSender(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		// 3. listen for messages that need to be sent
		case res := <-smc.socketModeResponses:
			smc.Debugf("Sending Socket Mode response with envelope ID %q: %v", res.EnvelopeID, res)

			if err := unsafeWriteSocketModeResponse(conn, res); err != nil {
				smc.Events <- newEvent(EventTypeErrorWriteFailed, &ErrorWriteFailed{
					Cause:    err,
					Response: res,
				})
			}

			smc.Debugf("Finished sending Socket Mode response with envelope ID %q", res.EnvelopeID)
		}
	}
}

// runRequestHandler is a blocking function that runs the Socket Mode request receiver.
//
// It reads WebSocket messages sent from Slack's Socket Mode WebSocket connection,
// parses them as Socket Mode requests, and processes them and optionally emit our own events into Client.Events channel.
func (smc *Client) runRequestHandler(ctx context.Context, websocket chan json.RawMessage) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message := <-websocket:
			smc.Debugf("Received WebSocket message: %s", message)

			// listen for incoming messages that need to be parsed
			evt, err := smc.parseEvent(message)
			if err != nil {
				smc.Events <- newEvent(EventTypeErrorBadMessage, &ErrorBadMessage{
					Cause:   err,
					Message: message,
				})
			} else if evt != nil {
				if evt.Type == EventTypeDisconnect {
					// We treat the `disconnect` request from Slack as an error internally,
					// so that we can tell the consumer of this function to reopen the connection on it.
					return errorRequestedDisconnect{}
				}

				smc.Events <- *evt
			}
		}
	}
}

// runMessageReceiver monitors the Socket Mode opened WebSocket connection for any incoming
// messages. It pushes the raw events into the channel.
// The receiver runs until the context is closed.
func (smc *Client) runMessageReceiver(ctx context.Context, conn *websocket.Conn, sink chan json.RawMessage) error {
	for {
		if err := smc.receiveMessagesInto(ctx, conn, sink); err != nil {
			return err
		}
	}
}

// unsafeWriteSocketModeResponse sends a WebSocket message back to Slack.
// WARNING: Call to this function must be serialized!
//
// Here's why - Gorilla WebSocket's Writes functions are not concurrency-safe.
// That is, we must serialize all the writes to it with e.g. a goroutine or mutex.
// We intentionally chose to use goroutine, which makes it harder to propagate write errors to the caller,
// but is more computationally efficient.
//
// See the below for more information on this topic:
// https://stackoverflow.com/questions/43225340/how-to-ensure-concurrency-in-golang-gorilla-websocket-package
func unsafeWriteSocketModeResponse(conn *websocket.Conn, res *Response) error {
	// set a write deadline on the connection
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}

	// Remove write deadline regardless of WriteJSON succeeds or not
	defer conn.SetWriteDeadline(time.Time{})

	if err := conn.WriteJSON(res); err != nil {
		return err
	}

	return nil
}

func newEvent(tpe EventType, data interface{}, req ...*Request) Event {
	evt := Event{Type: tpe, Data: data}

	if len(req) > 0 {
		evt.Request = req[0]
	}

	return evt
}

// Ack acknowledges the Socket Mode request with the payload.
//
// This tells Slack that the we have received the request denoted by the envelope ID,
// by sending back the envelope ID over the WebSocket connection.
func (smc *Client) Ack(req Request, payload ...interface{}) {
	res := Response{
		EnvelopeID: req.EnvelopeID,
	}

	if len(payload) > 0 {
		res.Payload = payload[0]
	}

	smc.Send(res)
}

// Send sends the Socket Mode response over a WebSocket connection.
// This is usually used for acknowledging requests, but if you need more control over Client.Ack().
// It's normally recommended to use Client.Ack() instead of this.
func (smc *Client) Send(res Response) {
	js, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	smc.Debugf("Scheduling Socket Mode response for envelope ID %s: %s", res.EnvelopeID, js)

	smc.socketModeResponses <- &res
}

// receiveMessagesInto attempts to receive an event from the WebSocket connection for Socket Mode.
// This will block until a frame is available from the WebSocket.
// If the read from the WebSocket results in a fatal error, this function will return non-nil.
func (smc *Client) receiveMessagesInto(ctx context.Context, conn *websocket.Conn, sink chan json.RawMessage) error {
	smc.Debugf("Starting to receive message")
	defer smc.Debugf("Finished to receive message")

	event := json.RawMessage{}
	err := conn.ReadJSON(&event)

	// check if the connection was closed.
	if websocket.IsUnexpectedCloseError(err) {
		return err
	}

	switch {
	case err == io.ErrUnexpectedEOF:
		// EOF's don't seem to signify a failed connection so instead we ignore
		// them here and detect a failed connection upon attempting to send a
		// 'PING' message

		// Unlike RTM, we don't ping from the our end as there seem to have no client ping.
		// We just continue to the next loop so that we `smc.disconnected` should be received if
		// this EOF error was actually due to disconnection.

		return nil
	case err != nil:
		// All other errors from ReadJSON come from NextReader, and should
		// kill the read loop and force a reconnect.
		smc.Events <- newEvent(EventTypeIncomingError, &slack.IncomingEventError{
			ErrorObj: err,
		})

		return err
	case len(event) == 0:
		smc.Debugln("Received empty event")
	default:
		if smc.debug {
			buf := &bytes.Buffer{}
			d := json.NewEncoder(buf)
			d.SetIndent("", "  ")
			if err := d.Encode(event); err != nil {
				smc.Debugln("Failed encoding decoded json:", err)
			}
			reencoded := buf.String()

			smc.Debugln("Incoming WebSocket message:", reencoded)
		}

		select {
		case sink <- event:
		case <-ctx.Done():
			smc.Debugln("cancelled while attempting to send raw event")

			return ctx.Err()
		}
	}

	return nil
}

// parseEvent takes a raw JSON message received from the slack websocket
// and handles the encoded event.
// returns the our own event that wraps the socket mode request.
func (smc *Client) parseEvent(wsMsg json.RawMessage) (*Event, error) {
	req := &Request{}
	err := json.Unmarshal(wsMsg, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling WebSocket message: %v", err)
	}

	var evt Event

	// See below two links for all the available message types.
	// - https://github.com/slackapi/node-slack-sdk/blob/c3f4d7109062a0356fb765d53794b7b5f6b3b5ae/packages/socket-mode/src/SocketModeClient.ts#L533
	// - https://api.slack.com/apis/connections/socket-implement
	switch req.Type {
	case RequestTypeHello:
		evt = newEvent(EventTypeHello, nil, req)
	case RequestTypeEventsAPI:
		payloadEvent := req.Payload

		eventsAPIEvent, err := slackevents.ParseEvent(payloadEvent, slackevents.OptionNoVerifyToken())
		if err != nil {
			return nil, fmt.Errorf("parsing Events API event: %v", err)
		}

		evt = newEvent(EventTypeEventsAPI, eventsAPIEvent, req)
	case RequestTypeDisconnect:
		// See https://api.slack.com/apis/connections/socket-implement#disconnect

		evt = newEvent(EventTypeDisconnect, nil, req)
	case RequestTypeSlashCommands:
		// See https://api.slack.com/apis/connections/socket-implement#command
		var cmd slack.SlashCommand

		if err := json.Unmarshal(req.Payload, &cmd); err != nil {
			return nil, fmt.Errorf("parsing slash command: %v", err)
		}

		evt = newEvent(EventTypeSlashCommand, cmd, req)
	case RequestTypeInteractive:
		// See belows:
		// - https://api.slack.com/apis/connections/socket-implement#button
		// - https://api.slack.com/apis/connections/socket-implement#home
		// - https://api.slack.com/apis/connections/socket-implement#modal
		// - https://api.slack.com/apis/connections/socket-implement#menu

		var callback slack.InteractionCallback

		if err := json.Unmarshal(req.Payload, &callback); err != nil {
			return nil, fmt.Errorf("parsing interaction callback: %v", err)
		}

		evt = newEvent(EventTypeInteractive, callback, req)
	default:
		return nil, fmt.Errorf("processing WebSocket message: encountered unsupported type %q", req.Type)
	}

	return &evt, nil
}

// handlePing handles an incoming 'PONG' message which should be in response to
// a previously sent 'PING' message. This is then used to compute the
// connection's latency.
func (smc *Client) handlePing(conn *websocket.Conn, event string) {
	smc.Debugf("WebSocket ping message received: %s", event)

	// In WebSocket, we need to respond a PING from the server with a PONG with the same payload as the PING.
	if err := conn.WriteControl(websocket.PongMessage, []byte(event), time.Now().Add(10*time.Second)); err != nil {
		smc.Debugf("Failed writing WebSocket PONG message: %v", err)
	}
}
