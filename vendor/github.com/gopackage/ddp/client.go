package ddp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/apex/log"
	"golang.org/x/net/websocket"
)

const (
	DISCONNECTED = iota
	DIALING
	CONNECTING
	CONNECTED
)

type ConnectionListener interface {
	Connected()
}

type ConnectionNotifier interface {
	AddConnectionListener(listener ConnectionListener)
}

type StatusListener interface {
	Status(status int)
}

type StatusNotifier interface {
	AddStatusListener(listener StatusListener)
}

// Client represents a DDP client connection. The DDP client establish a DDP
// session and acts as a message pump for other tools.
type Client struct {
	// HeartbeatInterval is the time between heartbeats to send
	HeartbeatInterval time.Duration
	// HeartbeatTimeout is the time for a heartbeat ping to timeout
	HeartbeatTimeout time.Duration
	// ReconnectInterval is the time between reconnections on bad connections
	ReconnectInterval time.Duration

	// writeSocketStats controls statistics gathering for current websocket writes.
	writeSocketStats *WriterStats
	// writeStats controls statistics gathering for overall client writes.
	writeStats *WriterStats
	// readSocketStats controls statistics gathering for current websocket reads.
	readSocketStats *ReaderStats
	// readStats controls statistics gathering for overall client reads.
	readStats *ReaderStats
	// reconnects in the number of reconnections the client has made
	reconnects int64
	// pingsIn is the number of pings received from the server
	pingsIn int64
	// pingsOut is te number of pings sent by the client
	pingsOut int64

	// session contains the DDP session token (can be used for reconnects and debugging).
	session string
	// version contains the negotiated DDP protocol version in use.
	version string
	// serverID the cluster node ID for the server we connected to
	serverID string
	// ws is the underlying websocket being used.
	ws *websocket.Conn
	// encoder is a JSON encoder to send outgoing packets to the websocket.
	encoder *json.Encoder
	// url the websocket is connected to
	url string
	// origin is the origin for the websocket connection
	origin string
	// inbox is an incoming message channel
	inbox chan map[string]interface{}
	// errors is an incoming errors channel
	errors chan error
	// pingTimer is a timer for sending regular pings to the server
	pingTimer *time.Timer
	// pings tracks inflight pings based on each ping ID.
	pings map[string][]*PingTracker
	// calls tracks method invocations that are still in flight
	calls map[string]*Call
	// subs tracks active subscriptions. Map contains name->args
	subs map[string]*Call
	// collections contains all the collections currently subscribed
	collections map[string]Collection
	// connectionStatus is the current connection status of the client
	connectionStatus int
	// reconnectTimer is the timer tracking reconnections
	reconnectTimer *time.Timer
	// reconnectLock protects access to reconnection
	reconnectLock *sync.Mutex

	// statusListeners will be informed when the connection status of the client changes
	statusListeners []StatusListener
	// connectionListeners will be informed when a connection to the server is established
	connectionListeners []ConnectionListener

	// KeyManager tracks IDs for ddp messages
	KeyManager
}

// NewClient creates a default client (using an internal websocket) to the
// provided URL using the origin for the connection. The client will
// automatically connect, upgrade to a websocket, and establish a DDP
// connection session before returning the client. The client will
// automatically and internally handle heartbeats and reconnects.
//
// TBD create an option to use an external websocket (aka htt.Transport)
// TBD create an option to substitute heartbeat and reconnect behavior (aka http.Transport)
// TBD create an option to hijack the connection (aka http.Hijacker)
// TBD create profiling features (aka net/http/pprof)
func NewClient(url, origin string) *Client {
	c := &Client{
		HeartbeatInterval: time.Minute,      // Meteor impl default + 10 (we ping last)
		HeartbeatTimeout:  15 * time.Second, // Meteor impl default
		ReconnectInterval: 5 * time.Second,
		collections:       map[string]Collection{},
		url:               url,
		origin:            origin,
		inbox:             make(chan map[string]interface{}, 100),
		errors:            make(chan error, 100),
		pings:             map[string][]*PingTracker{},
		calls:             map[string]*Call{},
		subs:              map[string]*Call{},
		connectionStatus:  DISCONNECTED,
		reconnectLock:     &sync.Mutex{},

		// Stats
		writeSocketStats: NewWriterStats(nil),
		writeStats:       NewWriterStats(nil),
		readSocketStats:  NewReaderStats(nil),
		readStats:        NewReaderStats(nil),

		KeyManager: *NewKeyManager(),
	}
	c.encoder = json.NewEncoder(c.writeStats)

	// We spin off an inbox processing goroutine
	go c.inboxManager()

	return c
}

// Session returns the negotiated session token for the connection.
func (c *Client) Session() string {
	return c.session
}

// Version returns the negotiated protocol version in use by the client.
func (c *Client) Version() string {
	return c.version
}

// AddStatusListener in order to receive status change updates.
func (c *Client) AddStatusListener(listener StatusListener) {
	c.statusListeners = append(c.statusListeners, listener)
}

// AddConnectionListener in order to receive connection updates.
func (c *Client) AddConnectionListener(listener ConnectionListener) {
	c.connectionListeners = append(c.connectionListeners, listener)
}

// status updates all status listeners with the new client status.
func (c *Client) status(status int) {
	if c.connectionStatus == status {
		return
	}
	c.connectionStatus = status
	for _, listener := range c.statusListeners {
		listener.Status(status)
	}
}

// Connect attempts to connect the client to the server.
func (c *Client) Connect() error {
	c.status(DIALING)
	ws, err := websocket.Dial(c.url, "", c.origin)
	if err != nil {
		c.Close()
		log.WithError(err).Debug("dial error")
		c.reconnectLater()
		return err
	}
	log.Debug("dialed")
	// Start DDP connection
	c.start(ws, NewConnect())
	return nil
}

// Reconnect attempts to reconnect the client to the server on the existing
// DDP session.
//
// TODO needs a reconnect backoff so we don't trash a down server
// TODO reconnect should not allow more reconnects while a reconnection is already in progress.
func (c *Client) Reconnect() {
	func() {
		c.reconnectLock.Lock()
		defer c.reconnectLock.Unlock()
		if c.reconnectTimer != nil {
			c.reconnectTimer.Stop()
			c.reconnectTimer = nil
		}
	}()

	c.Close()

	c.reconnects++

	// Reconnect
	c.status(DIALING)
	ws, err := websocket.Dial(c.url, "", c.origin)
	if err != nil {
		c.Close()
		log.WithError(err).Debug("Dial error")
		c.reconnectLater()
		return
	}

	c.start(ws, NewReconnect(c.session))

	// --------------------------------------------------------------------
	// We resume inflight or ongoing subscriptions - we don't have to wait
	// for connection confirmation (messages can be pipelined).
	// --------------------------------------------------------------------

	// Send calls that haven't been confirmed - may not have been sent
	// and effects should be idempotent
	for _, call := range c.calls {
		IgnoreErr(c.Send(NewMethod(call.ID, call.ServiceMethod, call.Args.([]interface{}))), "resend method")
	}

	// Resend subscriptions and patch up collections
	for _, sub := range c.subs {
		IgnoreErr(c.Send(NewSub(sub.ID, sub.ServiceMethod, sub.Args.([]interface{}))), "resend sub")
	}
}

// Subscribe to data updates.
func (c *Client) Subscribe(subName string, done chan *Call, args ...interface{}) *Call {

	if args == nil {
		args = []interface{}{}
	}
	call := new(Call)
	call.ID = c.Next()
	call.ServiceMethod = subName
	call.Args = args
	call.Owner = c

	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel.  If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Fatal("ddp.rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	c.subs[call.ID] = call

	// Save this subscription to the client so we can reconnect
	subArgs := make([]interface{}, len(args))
	copy(subArgs, args)

	IgnoreErr(c.Send(NewSub(call.ID, subName, args)), "send sub")

	return call
}

// Sub sends a synchronous subscription request to the server.
func (c *Client) Sub(subName string, args ...interface{}) error {
	call := <-c.Subscribe(subName, make(chan *Call, 1), args...).Done
	return call.Error
}

// Go invokes the function asynchronously.  It returns the Call structure representing
// the invocation.  The done channel will signal when the call is complete by returning
// the same Call object.  If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
//
// Go and Call are modeled after the standard `net/rpc` package versions.
func (c *Client) Go(serviceMethod string, done chan *Call, args ...interface{}) *Call {

	if args == nil {
		args = []interface{}{}
	}
	call := new(Call)
	call.ID = c.Next()
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Owner = c
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel.  If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Fatal("ddp.rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	c.calls[call.ID] = call

	IgnoreErr(c.Send(NewMethod(call.ID, serviceMethod, args)), "send method")

	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (c *Client) Call(serviceMethod string, args ...interface{}) (interface{}, error) {
	call := <-c.Go(serviceMethod, make(chan *Call, 1), args...).Done
	return call.Reply, call.Error
}

// Ping sends a heartbeat signal to the server. The Ping doesn't look for
// a response but may trigger the connection to reconnect if the ping times out.
// This is primarily useful for reviving an unresponsive Client connection.
func (c *Client) Ping() {
	c.PingPong(c.Next(), c.HeartbeatTimeout, func(err error) {
		if err != nil {
			// Is there anything else we should or can do?
			c.reconnectLater()
		}
	})
}

// PingPong sends a heartbeat signal to the server and calls the provided
// function when a pong is received. An optional id can be sent to help
// track the responses - or an empty string can be used. It is the
// responsibility of the caller to respond to any errors that may occur.
func (c *Client) PingPong(id string, timeout time.Duration, handler func(error)) {
	err := c.Send(NewPing(id))
	if err != nil {
		handler(err)
		return
	}
	c.pingsOut++
	pings, ok := c.pings[id]
	if !ok {
		pings = make([]*PingTracker, 0, 5)
	}
	tracker := &PingTracker{handler: handler, timeout: timeout, timer: time.AfterFunc(timeout, func() {
		handler(fmt.Errorf("ping timeout"))
	})}
	c.pings[id] = append(pings, tracker)
}

// Send transmits messages to the server. The msg parameter must be json
// encoder compatible.
func (c *Client) Send(msg interface{}) error {
	return c.encoder.Encode(msg)
}

// Close implements the io.Closer interface.
func (c *Client) Close() {
	// Shutdown out all outstanding pings
	if c.pingTimer != nil {
		c.pingTimer.Stop()
		c.pingTimer = nil
	}

	// Close websocket
	if c.ws != nil {
		IgnoreErr(c.ws.Close(), "close ws")
		c.ws = nil
	}
	for _, collection := range c.collections {
		collection.reset()
	}
	c.status(DISCONNECTED)
}

// ResetStats resets the statistics for the client.
func (c *Client) ResetStats() {
	c.readSocketStats.Reset()
	c.readStats.Reset()
	c.writeSocketStats.Reset()
	c.writeStats.Reset()
	c.reconnects = 0
	c.pingsIn = 0
	c.pingsOut = 0
}

// Stats returns the read and write statistics of the client.
func (c *Client) Stats() *ClientStats {
	return &ClientStats{
		Reads:       c.readSocketStats.Snapshot(),
		TotalReads:  c.readStats.Snapshot(),
		Writes:      c.writeSocketStats.Snapshot(),
		TotalWrites: c.writeStats.Snapshot(),
		Reconnects:  c.reconnects,
		PingsSent:   c.pingsOut,
		PingsRecv:   c.pingsIn,
	}
}

// CollectionByName retrieves a collection by its name.
func (c *Client) CollectionByName(name string) Collection {
	collection, ok := c.collections[name]
	if !ok {
		collection = NewCollection(name)
		c.collections[name] = collection
	}
	return collection
}

// CollectionStats returns a snapshot of statistics for the currently known collections.
func (c *Client) CollectionStats() []CollectionStats {
	stats := make([]CollectionStats, 0, len(c.collections))
	for name, collection := range c.collections {
		stats = append(stats, CollectionStats{Name: name, Count: len(collection.FindAll())})
	}
	return stats
}

// start a new client connection on the provided websocket
func (c *Client) start(ws *websocket.Conn, connect *Connect) {

	c.status(CONNECTING)

	c.ws = ws
	c.writeSocketStats = NewWriterStats(c.ws)
	c.writeStats.Writer = c.writeSocketStats
	c.readSocketStats = NewReaderStats(c.ws)
	c.readStats.Reader = c.readSocketStats

	// We spin off an inbox stuffing goroutine
	go c.inboxWorker(c.readStats)

	IgnoreErr(c.Send(connect), "send connect")
}

// inboxManager pulls messages from the inbox and routes them to appropriate
// handlers.
func (c *Client) inboxManager() {
	for {
		select {
		case msg := <-c.inbox:
			// Message!
			//log.Println("Got message", msg)
			msgType, ok := msg["msg"]
			if ok {
				log.WithField("msg", msgType).Debug("recv")
				switch msgType.(string) {
				// Connection management
				case "connected":
					c.status(CONNECTED)
					for _, collection := range c.collections {
						collection.init()
					}
					c.version = "1" // "1" is the only version we support
					c.session = msg["session"].(string)
					// Start automatic heartbeats
					c.pingTimer = time.AfterFunc(c.HeartbeatInterval, func() {
						c.Ping()
						c.pingTimer.Reset(c.HeartbeatInterval)
					})
					// Notify connection listeners
					for _, listener := range c.connectionListeners {
						go listener.Connected()
					}
				case "failed":
					log.Fatalf("IM Failed to connect, we support version 1 but server supports %s", msg["version"])

				// Heartbeats
				case "ping":
					// We received a ping - need to respond with a pong
					id, ok := msg["id"]
					if ok {
						IgnoreErr(c.Send(NewPong(id.(string))), "send id ping")
					} else {
						IgnoreErr(c.Send(NewPong("")), "send empty ping")
					}
					c.pingsIn++
				case "pong":
					// We received a pong - we can clear the ping tracker and call its handler
					id, ok := msg["id"]
					var key string
					if ok {
						key = id.(string)
					}
					pings, ok := c.pings[key]
					if ok && len(pings) > 0 {
						ping := pings[0]
						pings = pings[1:]
						if len(key) == 0 || len(pings) > 0 {
							c.pings[key] = pings
						}
						ping.timer.Stop()
						ping.handler(nil)
					}

				// Live Data
				case "nosub":
					log.WithField("msg", msg).Debug("sub returned a nosub error")
					// Clear related subscriptions
					sub, ok := msg["id"]
					if ok {
						id := sub.(string)
						runningSub := c.subs[id]

						if runningSub != nil {
							runningSub.Error = errors.New("sub returned a nosub error")
							runningSub.done()
							delete(c.subs, id)
						}
					}
				case "ready":
					// Run 'done' callbacks on all ready subscriptions
					subs, ok := msg["subs"]
					if ok {
						for _, sub := range subs.([]interface{}) {
							call, ok := c.subs[sub.(string)]
							if ok {
								call.done()
							}
						}
					}
				case "added":
					c.collectionBy(msg).added(msg)
				case "changed":
					c.collectionBy(msg).changed(msg)
				case "removed":
					c.collectionBy(msg).removed(msg)
				case "addedBefore":
					c.collectionBy(msg).addedBefore(msg)
				case "movedBefore":
					c.collectionBy(msg).movedBefore(msg)

				// RPC
				case "result":
					id, ok := msg["id"]
					if ok {
						call := c.calls[id.(string)]
						delete(c.calls, id.(string))
						e, ok := msg["error"]
						if ok {
							txt, _ := json.Marshal(e)
							call.Error = fmt.Errorf(string(txt))
							call.Reply = e
						} else {
							call.Reply = msg["result"]
						}
						call.done()
					}
				case "updated":
					// We currently don't do anything with updated status

				default:
					// Ignore?
					log.WithField("msg", msg).Debug("Server sent unexpected message")
				}
			} else {
				// Current Meteor server sends an undocumented DDP message
				// (looks like clustering "hint"). We will register and
				// ignore rather than log an error.
				serverID, ok := msg["server_id"]
				if ok {
					switch ID := serverID.(type) {
					case string:
						c.serverID = ID
					default:
						log.WithField("id", serverID).Debug("Server cluster node")
					}
				} else {
					log.WithField("msg", msg).Debug("Server sent message with no `msg` field")
				}
			}
		case err := <-c.errors:
			log.WithError(err).Warn("Websocket error")
		}
	}
}

func (c *Client) collectionBy(msg map[string]interface{}) Collection {
	n, ok := msg["collection"]
	if !ok {
		return NewMockCollection()
	}
	switch name := n.(type) {
	case string:
		return c.CollectionByName(name)
	default:
		return NewMockCollection()
	}
}

// inboxWorker pulls messages from a websocket, decodes JSON packets, and
// stuffs them into a message channel.
func (c *Client) inboxWorker(ws io.Reader) {
	dec := json.NewDecoder(ws)
	for {
		var event interface{}

		if err := dec.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			c.errors <- err
		}
		if c.pingTimer != nil {
			c.pingTimer.Reset(c.HeartbeatInterval)
		}
		if event == nil {
			log.Debug("Inbox worker found nil event. May be due to broken websocket. Reconnecting.")
			break
		} else {
			c.inbox <- event.(map[string]interface{})
		}
	}

	c.reconnectLater()
}

// reconnectLater schedules a reconnect action for later. We need to make sure that we don't
// block, and that we don't reconnect more frequently than once every c.ReconnectInterval
func (c *Client) reconnectLater() {
	c.Close()
	c.reconnectLock.Lock()
	defer c.reconnectLock.Unlock()
	if c.reconnectTimer == nil {
		c.reconnectTimer = time.AfterFunc(c.ReconnectInterval, c.Reconnect)
	}
}
