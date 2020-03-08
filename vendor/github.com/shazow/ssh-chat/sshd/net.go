package sshd

import (
	"net"
	"time"

	"github.com/shazow/rateio"
	"golang.org/x/crypto/ssh"
)

// SSHListener is the container for the connection and ssh-related configuration
type SSHListener struct {
	net.Listener
	config *ssh.ServerConfig

	RateLimit   func() rateio.Limiter
	HandlerFunc func(term *Terminal)
}

// ListenSSH makes an SSH listener socket
func ListenSSH(laddr string, config *ssh.ServerConfig) (*SSHListener, error) {
	socket, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}
	l := SSHListener{Listener: socket, config: config}
	return &l, nil
}

func (l *SSHListener) handleConn(conn net.Conn) (*Terminal, error) {
	if l.RateLimit != nil {
		// TODO: Configurable Limiter?
		conn = ReadLimitConn(conn, l.RateLimit())
	}

	// If the connection doesn't write anything back for too long before we get
	// a valid session, it should be dropped.
	var handleTimeout = 20 * time.Second
	conn.SetReadDeadline(time.Now().Add(handleTimeout))
	defer conn.SetReadDeadline(time.Time{})

	// Upgrade TCP connection to SSH connection
	sshConn, channels, requests, err := ssh.NewServerConn(conn, l.config)
	if err != nil {
		return nil, err
	}

	// FIXME: Disconnect if too many faulty requests? (Avoid DoS.)
	go ssh.DiscardRequests(requests)
	return NewSession(sshConn, channels)
}

// Serve Accepts incoming connections as terminal requests and yield them
func (l *SSHListener) Serve() {
	defer l.Close()
	for {
		conn, err := l.Accept()

		if err != nil {
			logger.Printf("Failed to accept connection: %s", err)
			break
		}

		// Goroutineify to resume accepting sockets early
		go func() {
			term, err := l.handleConn(conn)
			if err != nil {
				logger.Printf("[%s] Failed to handshake: %s", conn.RemoteAddr(), err)
				conn.Close() // Must be closed to avoid a leak
				return
			}
			l.HandlerFunc(term)
		}()
	}
}
