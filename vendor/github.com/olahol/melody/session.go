package melody

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Session wrapper around websocket connections.
type Session struct {
	Request    *http.Request
	Keys       map[string]any
	conn       *websocket.Conn
	output     chan envelope
	outputDone chan struct{}
	melody     *Melody
	open       bool
	rwmutex    *sync.RWMutex
}

func (s *Session) writeMessage(message envelope) {
	if s.closed() {
		s.melody.errorHandler(s, ErrWriteClosed)
		return
	}

	select {
	case s.output <- message:
	default:
		s.melody.errorHandler(s, ErrMessageBufferFull)
	}
}

func (s *Session) writeRaw(message envelope) error {
	if s.closed() {
		return ErrWriteClosed
	}

	s.conn.SetWriteDeadline(time.Now().Add(s.melody.Config.WriteWait))
	err := s.conn.WriteMessage(message.t, message.msg)

	if err != nil {
		return err
	}

	return nil
}

func (s *Session) closed() bool {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	return !s.open
}

func (s *Session) close() {
	s.rwmutex.Lock()
	open := s.open
	s.open = false
	s.rwmutex.Unlock()
	if open {
		s.conn.Close()
		close(s.outputDone)
	}
}

func (s *Session) ping() {
	s.writeRaw(envelope{t: websocket.PingMessage, msg: []byte{}})
}

func (s *Session) writePump() {
	ticker := time.NewTicker(s.melody.Config.PingPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case msg := <-s.output:
			err := s.writeRaw(msg)

			if err != nil {
				s.melody.errorHandler(s, err)
				break loop
			}

			if msg.t == websocket.CloseMessage {
				break loop
			}

			if msg.t == websocket.TextMessage {
				s.melody.messageSentHandler(s, msg.msg)
			}

			if msg.t == websocket.BinaryMessage {
				s.melody.messageSentHandlerBinary(s, msg.msg)
			}
		case <-ticker.C:
			s.ping()
		case _, ok := <-s.outputDone:
			if !ok {
				break loop
			}
		}
	}

	s.close()
}

func (s *Session) readPump() {
	s.conn.SetReadLimit(s.melody.Config.MaxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(s.melody.Config.PongWait))

	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(s.melody.Config.PongWait))
		s.melody.pongHandler(s)
		return nil
	})

	if s.melody.closeHandler != nil {
		s.conn.SetCloseHandler(func(code int, text string) error {
			return s.melody.closeHandler(s, code, text)
		})
	}

	for {
		t, message, err := s.conn.ReadMessage()

		if err != nil {
			s.melody.errorHandler(s, err)
			break
		}

		if s.melody.Config.ConcurrentMessageHandling {
			go s.handleMessage(t, message)
		} else {
			s.handleMessage(t, message)
		}
	}
}

func (s *Session) handleMessage(t int, message []byte) {
	switch t {
	case websocket.TextMessage:
		s.melody.messageHandler(s, message)
	case websocket.BinaryMessage:
		s.melody.messageHandlerBinary(s, message)
	}
}

// Write writes message to session.
func (s *Session) Write(msg []byte) error {
	if s.closed() {
		return ErrSessionClosed
	}

	s.writeMessage(envelope{t: websocket.TextMessage, msg: msg})

	return nil
}

// WriteBinary writes a binary message to session.
func (s *Session) WriteBinary(msg []byte) error {
	if s.closed() {
		return ErrSessionClosed
	}

	s.writeMessage(envelope{t: websocket.BinaryMessage, msg: msg})

	return nil
}

// Close closes session.
func (s *Session) Close() error {
	if s.closed() {
		return ErrSessionClosed
	}

	s.writeMessage(envelope{t: websocket.CloseMessage, msg: []byte{}})

	return nil
}

// CloseWithMsg closes the session with the provided payload.
// Use the FormatCloseMessage function to format a proper close message payload.
func (s *Session) CloseWithMsg(msg []byte) error {
	if s.closed() {
		return ErrSessionClosed
	}

	s.writeMessage(envelope{t: websocket.CloseMessage, msg: msg})

	return nil
}

// Set is used to store a new key/value pair exclusively for this session.
// It also lazy initializes s.Keys if it was not used previously.
func (s *Session) Set(key string, value any) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()

	if s.Keys == nil {
		s.Keys = make(map[string]any)
	}

	s.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (s *Session) Get(key string) (value any, exists bool) {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	if s.Keys != nil {
		value, exists = s.Keys[key]
	}

	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (s *Session) MustGet(key string) any {
	if value, exists := s.Get(key); exists {
		return value
	}

	panic("Key \"" + key + "\" does not exist")
}

// UnSet will delete the key and has no return value
func (s *Session) UnSet(key string) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	if s.Keys != nil {
		delete(s.Keys, key)
	}
}

// IsClosed returns the status of the connection.
func (s *Session) IsClosed() bool {
	return s.closed()
}

// LocalAddr returns the local addr of the connection.
func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the remote addr of the connection.
func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// WebsocketConnection returns the underlying websocket connection.
// This can be used to e.g. set/read additional websocket options or to write sychronous messages.
func (s *Session) WebsocketConnection() *websocket.Conn {
	return s.conn
}
