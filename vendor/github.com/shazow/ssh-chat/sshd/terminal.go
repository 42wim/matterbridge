package sshd

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var keepaliveInterval = time.Second * 30
var keepaliveRequest = "keepalive@ssh-chat"

var ErrNoSessionChannel = errors.New("no session channel")
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

// Extending ssh/terminal to include a closer interface
type Terminal struct {
	terminal.Terminal
	Conn    Connection
	Channel ssh.Channel

	done      chan struct{}
	closeOnce sync.Once
}

// Make new terminal from a session channel
func NewTerminal(conn *ssh.ServerConn, ch ssh.NewChannel) (*Terminal, error) {
	if ch.ChannelType() != "session" {
		return nil, ErrNotSessionChannel
	}
	channel, requests, err := ch.Accept()
	if err != nil {
		return nil, err
	}
	term := Terminal{
		Terminal: *terminal.NewTerminal(channel, "Connecting..."),
		Conn:     sshConn{conn},
		Channel:  channel,

		done: make(chan struct{}),
	}

	go term.listen(requests)

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

	return &term, nil
}

// Find session channel and make a Terminal from it
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

// Negotiate terminal type and settings
func (t *Terminal) listen(requests <-chan *ssh.Request) {
	hasShell := false

	for req := range requests {
		var width, height int
		var ok bool

		switch req.Type {
		case "shell":
			if !hasShell {
				ok = true
				hasShell = true
			}
		case "pty-req":
			width, height, ok = parsePtyRequest(req.Payload)
			if ok {
				// TODO: Hardcode width to 100000?
				err := t.SetSize(width, height)
				ok = err == nil
			}
		case "window-change":
			width, height, ok = parseWinchRequest(req.Payload)
			if ok {
				// TODO: Hardcode width to 100000?
				err := t.SetSize(width, height)
				ok = err == nil
			}
		}

		if req.WantReply {
			req.Reply(ok, nil)
		}
	}
}
