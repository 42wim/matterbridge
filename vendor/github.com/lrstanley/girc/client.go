// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Client contains all of the information necessary to run a single IRC
// client.
type Client struct {
	// Config represents the configuration. Please take extra caution in that
	// entries in this are not edited while the client is connected, to prevent
	// data races. This is NOT concurrent safe to update.
	Config Config
	// rx is a buffer of events waiting to be processed.
	rx chan *Event
	// tx is a buffer of events waiting to be sent.
	tx chan *Event
	// state represents the throw-away state for the irc session.
	state *state
	// initTime represents the creation time of the client.
	initTime time.Time
	// Handlers is a handler which manages internal and external handlers.
	Handlers *Caller
	// CTCP is a handler which manages internal and external CTCP handlers.
	CTCP *CTCP
	// Cmd contains various helper methods to interact with the server.
	Cmd *Commands
	// mu is the mux used for connections/disconnections from the server,
	// so multiple threads aren't trying to connect at the same time, and
	// vice versa.
	mu sync.RWMutex
	// stop is used to communicate with Connect(), letting it know that the
	// client wishes to cancel/close.
	stop context.CancelFunc
	// conn is a net.Conn reference to the IRC server. If this is nil, it is
	// safe to assume that we're not connected. If this is not nil, this
	// means we're either connected, connecting, or cleaning up. This should
	// be guarded with Client.mu.
	conn *ircConn
	// debug is used if a writer is supplied for Client.Config.Debugger.
	debug *log.Logger
}

// Config contains configuration options for an IRC client
type Config struct {
	// Server is a host/ip of the server you want to connect to. This only
	// has an affect during the dial process
	Server string
	// ServerPass is the server password used to authenticate. This only has
	// an affect during the dial process.
	ServerPass string
	// Port is the port that will be used during server connection. This only
	// has an affect during the dial process.
	Port int
	// Nick is an rfc-valid nickname used during connection. This only has an
	// affect during the dial process.
	Nick string
	// User is the username/ident to use on connect. Ignored if an identd
	// server is used. This only has an affect during the dial process.
	User string
	// Name is the "realname" that's used during connection. This only has an
	// affect during the dial process.
	Name string
	// SASL contains the necessary authentication data to authenticate
	// with SASL. See the documentation for SASLMech for what is currently
	// supported. Capability tracking must be enabled for this to work, as
	// this requires IRCv3 CAP handling.
	SASL SASLMech
	// Bind is used to bind to a specific host or ip during the dial process
	// when connecting to the server. This can be a hostname, however it must
	// resolve to an IPv4/IPv6 address bindable on your system. Otherwise,
	// you can simply use a IPv4/IPv6 address directly. This only has an
	// affect during the dial process and will not work with DialerConnect().
	Bind string
	// SSL allows dialing via TLS. See TLSConfig to set your own TLS
	// configuration (e.g. to not force hostname checking). This only has an
	// affect during the dial process.
	SSL bool
	// TLSConfig is an optional user-supplied tls configuration, used during
	// socket creation to the server. SSL must be enabled for this to be used.
	// This only has an affect during the dial process.
	TLSConfig *tls.Config
	// AllowFlood allows the client to bypass the rate limit of outbound
	// messages.
	AllowFlood bool
	// GlobalFormat enables passing through all events which have trailing
	// text through the color Fmt() function, so you don't have to wrap
	// every response in the Fmt() method.
	//
	// Note that this only actually applies to PRIVMSG, NOTICE and TOPIC
	// events, to ensure it doesn't clobber unwanted events.
	GlobalFormat bool
	// Debug is an optional, user supplied location to log the raw lines
	// sent from the server, or other useful debug logs. Defaults to
	// ioutil.Discard. For quick debugging, this could be set to os.Stdout.
	Debug io.Writer
	// Out is used to write out a prettified version of incoming events. For
	// example, channel JOIN/PART, PRIVMSG/NOTICE, KICk, etc. Useful to get
	// a brief output of the activity of the client. If you are looking to
	// log raw messages, look at a handler and girc.ALLEVENTS and the relevant
	// Event.Bytes() or Event.String() methods.
	Out io.Writer
	// RecoverFunc is called when a handler throws a panic. If RecoverFunc is
	// set, the panic will be considered recovered, otherwise the client will
	// panic. Set this to DefaultRecoverHandler if you don't want the client
	// to panic, however you don't want to handle the panic yourself.
	// DefaultRecoverHandler will log the panic to Debug or os.Stdout if
	// Debug is unset.
	RecoverFunc func(c *Client, e *HandlerError)
	// SupportedCaps are the IRCv3 capabilities you would like the client to
	// support on top of the ones which the client already supports (see
	// cap.go for which ones the client enables by default). Only use this
	// if you have not called DisableTracking(). The keys value gets passed
	// to the server if supported.
	SupportedCaps map[string][]string
	// Version is the application version information that will be used in
	// response to a CTCP VERSION, if default CTCP replies have not been
	// overwritten or a VERSION handler was already supplied.
	Version string
	// PingDelay is the frequency between when the client sends a keep-alive
	// PING to the server, and awaits a response (and times out if the server
	// doesn't respond in time). This should be between 20-600 seconds. See
	// Client.Latency() if you want to determine the delay between the server
	// and the client. If this is set to -1, the client will not attempt to
	// send client -> server PING requests.
	PingDelay time.Duration

	// disableTracking disables all channel and user-level tracking. Useful
	// for highly embedded scripts with single purposes. This has an exported
	// method which enables this and ensures prop cleanup, see
	// Client.DisableTracking().
	disableTracking bool
	// HandleNickCollide when set, allows the client to handle nick collisions
	// in a custom way. If unset, the client will attempt to append a
	// underscore to the end of the nickname, in order to bypass using
	// an invalid nickname. For example, if "test" is already in use, or is
	// blocked by the network/a service, the client will try and use "test_",
	// then it will attempt "test__", "test___", and so on.
	HandleNickCollide func(oldNick string) (newNick string)
}

// ErrInvalidConfig is returned when the configuration passed to the client
// is invalid.
type ErrInvalidConfig struct {
	Conf Config // Conf is the configuration that was not valid.
	err  error
}

func (e ErrInvalidConfig) Error() string { return "invalid configuration: " + e.err.Error() }

// isValid checks some basic settings to ensure the config is valid.
func (conf *Config) isValid() error {
	if conf.Server == "" {
		return &ErrInvalidConfig{Conf: *conf, err: errors.New("empty server")}
	}

	// Default port to 6667 (the standard IRC port).
	if conf.Port == 0 {
		conf.Port = 6667
	}

	if conf.Port < 1 || conf.Port > 65535 {
		return &ErrInvalidConfig{Conf: *conf, err: errors.New("port outside valid range (1-65535)")}
	}

	if !IsValidNick(conf.Nick) {
		return &ErrInvalidConfig{Conf: *conf, err: errors.New("bad nickname specified")}
	}
	if !IsValidUser(conf.User) {
		return &ErrInvalidConfig{Conf: *conf, err: errors.New("bad user/ident specified")}
	}

	return nil
}

// ErrNotConnected is returned if a method is used when the client isn't
// connected.
var ErrNotConnected = errors.New("client is not connected to server")

// New creates a new IRC client with the specified server, name and config.
func New(config Config) *Client {
	c := &Client{
		Config:   config,
		rx:       make(chan *Event, 25),
		tx:       make(chan *Event, 25),
		CTCP:     newCTCP(),
		initTime: time.Now(),
	}

	c.Cmd = &Commands{c: c}

	if c.Config.PingDelay >= 0 && c.Config.PingDelay < (20*time.Second) {
		c.Config.PingDelay = 20 * time.Second
	} else if c.Config.PingDelay > (600 * time.Second) {
		c.Config.PingDelay = 600 * time.Second
	}

	if c.Config.Debug == nil {
		c.debug = log.New(ioutil.Discard, "", 0)
	} else {
		c.debug = log.New(c.Config.Debug, "debug:", log.Ltime|log.Lshortfile)
		c.debug.Print("initializing debugging")
	}

	// Setup the caller.
	c.Handlers = newCaller(c.debug)

	// Give ourselves a new state.
	c.state = &state{}
	c.state.reset()

	// Register builtin handlers.
	c.registerBuiltins()

	// Register default CTCP responses.
	c.CTCP.addDefaultHandlers()

	return c
}

// String returns a brief description of the current client state.
func (c *Client) String() string {
	connected := c.IsConnected()

	return fmt.Sprintf(
		"<Client init:%q handlers:%d connected:%t>", c.initTime.String(), c.Handlers.Len(), connected,
	)
}

// TLSConnectionState returns the TLS connection state from tls.Conn{}, which
// is useful to return needed TLS fingerprint info, certificates, verify cert
// expiration dates, etc. Will only return an error if the underlying
// connection wasn't established using TLS (see ErrConnNotTLS), or if the
// client isn't connected.
func (c *Client) TLSConnectionState() (*tls.ConnectionState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil {
		return nil, ErrNotConnected
	}

	c.conn.mu.RLock()
	defer c.conn.mu.RUnlock()

	if !c.conn.connected {
		return nil, ErrNotConnected
	}

	if tlsConn, ok := c.conn.sock.(*tls.Conn); ok {
		cs := tlsConn.ConnectionState()
		return &cs, nil
	}

	return nil, ErrConnNotTLS
}

// ErrConnNotTLS is returned when Client.TLSConnectionState() is called, and
// the connection to the server wasn't made with TLS.
var ErrConnNotTLS = errors.New("underlying connection is not tls")

// Close closes the network connection to the server, and sends a STOPPED
// event. This should cause Connect() to return with nil. This should be
// safe to call multiple times. See Connect()'s documentation on how
// handlers and goroutines are handled when disconnected from the server.
func (c *Client) Close() {
	c.mu.RLock()
	if c.stop != nil {
		c.debug.Print("requesting client to stop")
		c.stop()
	}
	c.mu.RUnlock()
}

// ErrEvent is an error returned when the server (or library) sends an ERROR
// message response. The string returned contains the trailing text from the
// message.
type ErrEvent struct {
	Event *Event
}

func (e *ErrEvent) Error() string {
	if e.Event == nil {
		return "unknown error occurred"
	}

	return e.Event.Trailing
}

func (c *Client) execLoop(ctx context.Context, errs chan error, wg *sync.WaitGroup) {
	c.debug.Print("starting execLoop")
	defer c.debug.Print("closing execLoop")

	var event *Event

	for {
		select {
		case <-ctx.Done():
			// We've been told to exit, however we shouldn't bail on the
			// current events in the queue that should be processed, as one
			// may want to handle an ERROR, QUIT, etc.
			c.debug.Printf("received signal to close, flushing %d events and executing", len(c.rx))
			for {
				select {
				case event = <-c.rx:
					c.RunHandlers(event)
				default:
					goto done
				}
			}

		done:
			wg.Done()
			return
		case event = <-c.rx:
			if event != nil && event.Command == ERROR {
				// Handles incoming ERROR responses. These are only ever sent
				// by the server (with the exception that this library may use
				// them as a lower level way of signalling to disconnect due
				// to some other client-choosen error), and should always be
				// followed up by the server disconnecting the client. If for
				// some reason the server doesn't disconnect the client, or
				// if this library is the source of the error, this should
				// signal back up to the main connect loop, to disconnect.
				errs <- &ErrEvent{Event: event}

				// Make sure to not actually exit, so we can let any handlers
				// actually handle the ERROR event.
			}

			c.RunHandlers(event)
		}
	}
}

// DisableTracking disables all channel/user-level/CAP tracking, and clears
// all internal handlers. Useful for highly embedded scripts with single
// purposes. This cannot be un-done on a client.
func (c *Client) DisableTracking() {
	c.debug.Print("disabling tracking")
	c.Config.disableTracking = true
	c.Handlers.clearInternal()

	c.state.Lock()
	c.state.channels = nil
	c.state.Unlock()
	c.state.notify(c, UPDATE_STATE)

	c.registerBuiltins()
}

// Server returns the string representation of host+port pair for net.Conn.
func (c *Client) Server() string {
	return fmt.Sprintf("%s:%d", c.Config.Server, c.Config.Port)
}

// Lifetime returns the amount of time that has passed since the client was
// created.
func (c *Client) Lifetime() time.Duration {
	return time.Since(c.initTime)
}

// Uptime is the time at which the client successfully connected to the
// server.
func (c *Client) Uptime() (up *time.Time, err error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	c.mu.RLock()
	c.conn.mu.RLock()
	up = c.conn.connTime
	c.conn.mu.RUnlock()
	c.mu.RUnlock()

	return up, nil
}

// ConnSince is the duration that has past since the client successfully
// connected to the server.
func (c *Client) ConnSince() (since *time.Duration, err error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	c.mu.RLock()
	c.conn.mu.RLock()
	timeSince := time.Since(*c.conn.connTime)
	c.conn.mu.RUnlock()
	c.mu.RUnlock()

	return &timeSince, nil
}

// IsConnected returns true if the client is connected to the server.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	if c.conn == nil {
		c.mu.RUnlock()
		return false
	}

	c.conn.mu.RLock()
	connected := c.conn.connected
	c.conn.mu.RUnlock()
	c.mu.RUnlock()

	return connected
}

// GetNick returns the current nickname of the active connection. Panics if
// tracking is disabled.
func (c *Client) GetNick() string {
	c.panicIfNotTracking()

	c.state.RLock()
	defer c.state.RUnlock()

	if c.state.nick == "" {
		return c.Config.Nick
	}
	return c.state.nick
}

// GetIdent returns the current ident of the active connection. Panics if
// tracking is disabled. May be empty, as this is obtained from when we join
// a channel, as there is no other more efficient method to return this info.
func (c *Client) GetIdent() string {
	c.panicIfNotTracking()

	c.state.RLock()
	defer c.state.RUnlock()

	if c.state.ident == "" {
		return c.Config.User
	}
	return c.state.ident
}

// GetHost returns the current host of the active connection. Panics if
// tracking is disabled. May be empty, as this is obtained from when we join
// a channel, as there is no other more efficient method to return this info.
func (c *Client) GetHost() (host string) {
	c.panicIfNotTracking()

	c.state.RLock()
	host = c.state.host
	c.state.RUnlock()
	return host
}

// ChannelList returns the (sorted) active list of channel names that the client
// is in. Panics if tracking is disabled.
func (c *Client) ChannelList() []string {
	c.panicIfNotTracking()

	c.state.RLock()
	channels := make([]string, 0, len(c.state.channels))
	for channel := range c.state.channels {
		channels = append(channels, c.state.channels[channel].Name)
	}
	c.state.RUnlock()
	sort.Strings(channels)
	return channels
}

// Channels returns the (sorted) active channels that the client is in. Panics
// if tracking is disabled.
func (c *Client) Channels() []*Channel {
	c.panicIfNotTracking()

	c.state.RLock()
	channels := make([]*Channel, 0, len(c.state.channels))
	for channel := range c.state.channels {
		channels = append(channels, c.state.channels[channel].Copy())
	}
	c.state.RUnlock()

	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
	return channels
}

// UserList returns the (sorted) active list of nicknames that the client is
// tracking across all channels. Panics if tracking is disabled.
func (c *Client) UserList() []string {
	c.panicIfNotTracking()

	c.state.RLock()
	users := make([]string, 0, len(c.state.users))
	for user := range c.state.users {
		users = append(users, c.state.users[user].Nick)
	}
	c.state.RUnlock()
	sort.Strings(users)
	return users
}

// Users returns the (sorted) active users that the client is tracking across
// all channels. Panics if tracking is disabled.
func (c *Client) Users() []*User {
	c.panicIfNotTracking()

	c.state.RLock()
	users := make([]*User, 0, len(c.state.users))
	for user := range c.state.users {
		users = append(users, c.state.users[user].Copy())
	}
	c.state.RUnlock()

	sort.Slice(users, func(i, j int) bool {
		return users[i].Nick < users[j].Nick
	})
	return users
}

// LookupChannel looks up a given channel in state. If the channel doesn't
// exist, nil is returned. Panics if tracking is disabled.
func (c *Client) LookupChannel(name string) (channel *Channel) {
	c.panicIfNotTracking()
	if name == "" {
		return nil
	}

	c.state.RLock()
	channel = c.state.lookupChannel(name).Copy()
	c.state.RUnlock()
	return channel
}

// LookupUser looks up a given user in state. If the user doesn't exist, nil
// is returned. Panics if tracking is disabled.
func (c *Client) LookupUser(nick string) (user *User) {
	c.panicIfNotTracking()
	if nick == "" {
		return nil
	}

	c.state.RLock()
	user = c.state.lookupUser(nick).Copy()
	c.state.RUnlock()
	return user
}

// IsInChannel returns true if the client is in channel. Panics if tracking
// is disabled.
func (c *Client) IsInChannel(channel string) (in bool) {
	c.panicIfNotTracking()

	c.state.RLock()
	_, in = c.state.channels[ToRFC1459(channel)]
	c.state.RUnlock()
	return in
}

// GetServerOption retrieves a server capability setting that was retrieved
// during client connection. This is also known as ISUPPORT (or RPL_PROTOCTL).
// Will panic if used when tracking has been disabled. Examples of usage:
//
//   nickLen, success := GetServerOption("MAXNICKLEN")
//
func (c *Client) GetServerOption(key string) (result string, ok bool) {
	c.panicIfNotTracking()

	c.state.RLock()
	result, ok = c.state.serverOptions[key]
	c.state.RUnlock()
	return result, ok
}

// NetworkName returns the network identifier. E.g. "EsperNet", "ByteIRC".
// May be empty if the server does not support RPL_ISUPPORT (or RPL_PROTOCTL).
// Will panic if used when tracking has been disabled.
func (c *Client) NetworkName() (name string) {
	c.panicIfNotTracking()

	name, _ = c.GetServerOption("NETWORK")
	return name
}

// ServerVersion returns the server software version, if the server has
// supplied this information during connection. May be empty if the server
// does not support RPL_MYINFO. Will panic if used when tracking has been
// disabled.
func (c *Client) ServerVersion() (version string) {
	c.panicIfNotTracking()

	version, _ = c.GetServerOption("VERSION")
	return version
}

// ServerMOTD returns the servers message of the day, if the server has sent
// it upon connect. Will panic if used when tracking has been disabled.
func (c *Client) ServerMOTD() (motd string) {
	c.panicIfNotTracking()

	c.state.RLock()
	motd = c.state.motd
	c.state.RUnlock()
	return motd
}

// Latency is the latency between the server and the client. This is measured
// by determining the difference in time between when we ping the server, and
// when we receive a pong.
func (c *Client) Latency() (delta time.Duration) {
	c.mu.RLock()
	c.conn.mu.RLock()
	delta = c.conn.lastPong.Sub(c.conn.lastPing)
	c.conn.mu.RUnlock()
	c.mu.RUnlock()

	if delta < 0 {
		return 0
	}

	return delta
}

// HasCapability checks if the client connection has the given capability. If
// you want the full list of capabilities, listen for the girc.CAP_ACK event.
// Will panic if used when tracking has been disabled.
func (c *Client) HasCapability(name string) (has bool) {
	c.panicIfNotTracking()

	if !c.IsConnected() {
		return false
	}

	name = strings.ToLower(name)

	c.state.RLock()
	for i := 0; i < len(c.state.enabledCap); i++ {
		if strings.ToLower(c.state.enabledCap[i]) == name {
			has = true
			break
		}
	}
	c.state.RUnlock()

	return has
}

// panicIfNotTracking will throw a panic when it's called, and tracking is
// disabled. Adds useful info like what function specifically, and where it
// was called from.
func (c *Client) panicIfNotTracking() {
	if !c.Config.disableTracking {
		return
	}

	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	_, file, line, _ := runtime.Caller(2)

	panic(fmt.Sprintf("%s used when tracking is disabled (caller %s:%d)", fn.Name(), file, line))
}
