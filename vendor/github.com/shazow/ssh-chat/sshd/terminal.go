package sshd

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/shazow/ssh-chat/sshd/terminal"
	"golang.org/x/crypto/ssh"
)

var keepaliveInterval = time.Second * 30
var keepaliveRequest = "keepalive@ssh-chat"

// ErrNoSessionChannel is returned when there is no session channel.
var ErrNoSessionChannel = errors.New("no session channel")

// ErrNotSessionChannel is returned when a channel is not a session channel.
var ErrNotSessionChannel = errors.New("terminal requires session channel")

// Connection is an interface with fields necessary to operate an sshd host.
type Connection interface {
	PublicKey() ssh.PublicKey
	RemoteAddr() net.Addr
	Name() string
	ClientVersion() []byte
	Close() error
}

type sshConn struct {
	*ssh.ServerConn
}

func (c sshConn) PublicKey() ssh.PublicKey {
	if c.Permissions == nil {
		return nil
	}

	s, ok := c.Permissions.Extensions["pubkey"]
	if !ok {
		return nil
	}

	key, err := ssh.ParsePublicKey([]byte(s))
	if err != nil {
		return nil
	}

	return key
}

func (c sshConn) Name() string {
	return c.User()
}

// EnvVar is an environment variable key-value pair
type EnvVar struct {
	Key   string
	Value string
}

func (v EnvVar) String() string {
	return v.Key + "=" + v.Value
}

// Env is a wrapper type around []EnvVar with some helper methods
type Env []EnvVar

// Get returns the latest value for a given key, or empty string if not found
func (e Env) Get(key string) string {
	for i := len(e) - 1; i >= 0; i-- {
		if e[i].Key == key {
			return e[i].Value
		}
	}
	return ""
}

// Terminal extends ssh/terminal to include a close method
type Terminal struct {
	terminal.Terminal
	Conn    Connection
	Channel ssh.Channel

	done      chan struct{}
	closeOnce sync.Once

	mu   sync.Mutex
	env  []EnvVar
	term string
}

// Make new terminal from a session channel
// TODO: For v2, make a separate `Serve(ctx context.Context) error` method to activate the Terminal
func NewTerminal(conn *ssh.ServerConn, ch ssh.NewChannel) (*Terminal, error) {
	if ch.ChannelType() != "session" {
		return nil, ErrNotSessionChannel
	}
	channel, requests, err := ch.Accept()
	if err != nil {
		return nil, err
	}
	term := Terminal{
		Terminal: *terminal.NewTerminal(channel, ""),
		Conn:     sshConn{conn},
		Channel:  channel,

		done: make(chan struct{}),
	}

	ready := make(chan struct{})
	go term.listen(requests, ready)

	go func() {
		// Keep-Alive Ticker
		ticker := time.Tick(keepaliveInterval)
		for {
			select {
			case <-ticker:
				_, err := channel.SendRequest(keepaliveRequest, true, nil)
				if err != nil {
					// Connection is gone
					logger.Printf("[%s] Keepalive failed, closing terminal: %s", term.Conn.RemoteAddr(), err)
					term.Close()
					return
				}
			case <-term.done:
				return
			}
		}
	}()

	// We need to wait for term.ready to acquire a shell before we return, this
	// gives the SSH session a chance to populate the env vars and other state.
	// TODO: Make the timeout configurable
	// TODO: Use context.Context for abort/timeout in the future, will need to change the API.
	select {
	case <-ready: // shell acquired
		return &term, nil
	case <-term.done:
		return nil, errors.New("terminal aborted")
	case <-time.NewTimer(time.Minute).C:
		return nil, errors.New("timed out starting terminal")
	}
}

// NewSession Finds a session channel and make a Terminal from it
func NewSession(conn *ssh.ServerConn, channels <-chan ssh.NewChannel) (*Terminal, error) {
	// Make a terminal from the first session found
	for ch := range channels {
		if t := ch.ChannelType(); t != "session" {
			logger.Printf("[%s] Ignored channel type: %s", conn.RemoteAddr(), t)
			ch.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		return NewTerminal(conn, ch)
	}

	return nil, ErrNoSessionChannel
}

// Close terminal and ssh connection
func (t *Terminal) Close() error {
	var err error
	t.closeOnce.Do(func() {
		close(t.done)
		t.Channel.Close()
		err = t.Conn.Close()
	})
	return err
}

// listen negotiates the terminal type and state
// ready is closed when the terminal is ready.
func (t *Terminal) listen(requests <-chan *ssh.Request, ready chan<- struct{}) {
	hasShell := false

	for req := range requests {
		var width, height int
		var ok bool

		switch req.Type {
		case "shell":
			if !hasShell {
				ok = true
				hasShell = true
				close(ready)
			}
		case "pty-req":
			var term string
			term, width, height, ok = parsePtyRequest(req.Payload)
			if ok {
				// TODO: Hardcode width to 100000?
				err := t.SetSize(width, height)
				ok = err == nil
				// Save the term:
				t.mu.Lock()
				t.term = term
				t.mu.Unlock()
			}
		case "window-change":
			width, height, ok = parseWinchRequest(req.Payload)
			if ok {
				// TODO: Hardcode width to 100000?
				err := t.SetSize(width, height)
				ok = err == nil
			}
		case "env":
			var v EnvVar
			if err := ssh.Unmarshal(req.Payload, &v); err == nil {
				t.mu.Lock()
				t.env = append(t.env, v)
				t.mu.Unlock()
				ok = true
			}
		}

		if req.WantReply {
			req.Reply(ok, nil)
		}
	}
}

// Env returns a list of environment key-values that have been set. They are
// returned in the order that they have been set, there is no deduplication or
// other pre-processing applied.
func (t *Terminal) Env() Env {
	t.mu.Lock()
	defer t.mu.Unlock()
	return Env(t.env)
}

// Term returns the terminal string value as set by the pty.
// If there was no pty request, it falls back to the TERM value passed in as an
// Env variable.
func (t *Terminal) Term() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.term != "" {
		return t.term
	}
	return Env(t.env).Get("TERM")
}
