package webtransport

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
)

// sessionID is the WebTransport Session ID
type sessionID uint64

const closeWebtransportSessionCapsuleType http3.CapsuleType = 0x2843

type acceptQueue[T any] struct {
	mx sync.Mutex
	// The channel is used to notify consumers (via Chan) about new incoming items.
	// Needs to be buffered to preserve the notification if an item is enqueued
	// between a call to Next and to Chan.
	c chan struct{}
	// Contains all the streams waiting to be accepted.
	// There's no explicit limit to the length of the queue, but it is implicitly
	// limited by the stream flow control provided by QUIC.
	queue []T
}

func newAcceptQueue[T any]() *acceptQueue[T] {
	return &acceptQueue[T]{c: make(chan struct{}, 1)}
}

func (q *acceptQueue[T]) Add(str T) {
	q.mx.Lock()
	q.queue = append(q.queue, str)
	q.mx.Unlock()

	select {
	case q.c <- struct{}{}:
	default:
	}
}

func (q *acceptQueue[T]) Next() T {
	q.mx.Lock()
	defer q.mx.Unlock()

	if len(q.queue) == 0 {
		return *new(T)
	}
	str := q.queue[0]
	q.queue = q.queue[1:]
	return str
}

func (q *acceptQueue[T]) Chan() <-chan struct{} { return q.c }

type Session struct {
	sessionID  sessionID
	qconn      http3.StreamCreator
	requestStr quic.Stream

	streamHdr    []byte
	uniStreamHdr []byte

	ctx      context.Context
	closeMx  sync.Mutex
	closeErr error // not nil once the session is closed
	// streamCtxs holds all the context.CancelFuncs of calls to Open{Uni}StreamSync calls currently active.
	// When the session is closed, this allows us to cancel all these contexts and make those calls return.
	streamCtxs map[int]context.CancelFunc

	bidiAcceptQueue acceptQueue[Stream]
	uniAcceptQueue  acceptQueue[ReceiveStream]

	// TODO: garbage collect streams from when they are closed
	streams streamsMap
}

func newSession(sessionID sessionID, qconn http3.StreamCreator, requestStr quic.Stream) *Session {
	tracingID := qconn.Context().Value(quic.ConnectionTracingKey).(uint64)
	ctx, ctxCancel := context.WithCancel(context.WithValue(context.Background(), quic.ConnectionTracingKey, tracingID))
	c := &Session{
		sessionID:       sessionID,
		qconn:           qconn,
		requestStr:      requestStr,
		ctx:             ctx,
		streamCtxs:      make(map[int]context.CancelFunc),
		bidiAcceptQueue: *newAcceptQueue[Stream](),
		uniAcceptQueue:  *newAcceptQueue[ReceiveStream](),
		streams:         *newStreamsMap(),
	}
	// precompute the headers for unidirectional streams
	c.uniStreamHdr = make([]byte, 0, 2+quicvarint.Len(uint64(c.sessionID)))
	c.uniStreamHdr = quicvarint.Append(c.uniStreamHdr, webTransportUniStreamType)
	c.uniStreamHdr = quicvarint.Append(c.uniStreamHdr, uint64(c.sessionID))
	// precompute the headers for bidirectional streams
	c.streamHdr = make([]byte, 0, 2+quicvarint.Len(uint64(c.sessionID)))
	c.streamHdr = quicvarint.Append(c.streamHdr, webTransportFrameType)
	c.streamHdr = quicvarint.Append(c.streamHdr, uint64(c.sessionID))

	go func() {
		defer ctxCancel()
		c.handleConn()
	}()
	return c
}

func (s *Session) handleConn() {
	var closeErr *ConnectionError
	err := s.parseNextCapsule()
	if !errors.As(err, &closeErr) {
		closeErr = &ConnectionError{Remote: true}
	}

	s.closeMx.Lock()
	defer s.closeMx.Unlock()
	// If we closed the connection, the closeErr will be set in Close.
	if s.closeErr == nil {
		s.closeErr = closeErr
	}
	for _, cancel := range s.streamCtxs {
		cancel()
	}
	s.streams.CloseSession()
}

// parseNextCapsule parses the next Capsule sent on the request stream.
// It returns a ConnectionError, if the capsule received is a CLOSE_WEBTRANSPORT_SESSION Capsule.
func (s *Session) parseNextCapsule() error {
	for {
		// TODO: enforce max size
		typ, r, err := http3.ParseCapsule(quicvarint.NewReader(s.requestStr))
		if err != nil {
			return err
		}
		switch typ {
		case closeWebtransportSessionCapsuleType:
			b := make([]byte, 4)
			if _, err := io.ReadFull(r, b); err != nil {
				return err
			}
			appErrCode := binary.BigEndian.Uint32(b)
			appErrMsg, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			return &ConnectionError{
				Remote:    true,
				ErrorCode: SessionErrorCode(appErrCode),
				Message:   string(appErrMsg),
			}
		default:
			// unknown capsule, skip it
			if _, err := io.ReadAll(r); err != nil {
				return err
			}
		}
	}
}

func (s *Session) addStream(qstr quic.Stream, addStreamHeader bool) Stream {
	var hdr []byte
	if addStreamHeader {
		hdr = s.streamHdr
	}
	str := newStream(qstr, hdr, func() { s.streams.RemoveStream(qstr.StreamID()) })
	s.streams.AddStream(qstr.StreamID(), str.closeWithSession)
	return str
}

func (s *Session) addReceiveStream(qstr quic.ReceiveStream) ReceiveStream {
	str := newReceiveStream(qstr, func() { s.streams.RemoveStream(qstr.StreamID()) })
	s.streams.AddStream(qstr.StreamID(), func() {
		str.closeWithSession()
	})
	return str
}

func (s *Session) addSendStream(qstr quic.SendStream) SendStream {
	str := newSendStream(qstr, s.uniStreamHdr, func() { s.streams.RemoveStream(qstr.StreamID()) })
	s.streams.AddStream(qstr.StreamID(), str.closeWithSession)
	return str
}

// addIncomingStream adds a bidirectional stream that the remote peer opened
func (s *Session) addIncomingStream(qstr quic.Stream) {
	s.closeMx.Lock()
	closeErr := s.closeErr
	if closeErr != nil {
		s.closeMx.Unlock()
		qstr.CancelRead(sessionCloseErrorCode)
		qstr.CancelWrite(sessionCloseErrorCode)
		return
	}
	str := s.addStream(qstr, false)
	s.closeMx.Unlock()

	s.bidiAcceptQueue.Add(str)
}

// addIncomingUniStream adds a unidirectional stream that the remote peer opened
func (s *Session) addIncomingUniStream(qstr quic.ReceiveStream) {
	s.closeMx.Lock()
	closeErr := s.closeErr
	if closeErr != nil {
		s.closeMx.Unlock()
		qstr.CancelRead(sessionCloseErrorCode)
		return
	}
	str := s.addReceiveStream(qstr)
	s.closeMx.Unlock()

	s.uniAcceptQueue.Add(str)
}

// Context returns a context that is closed when the session is closed.
func (s *Session) Context() context.Context {
	return s.ctx
}

func (s *Session) AcceptStream(ctx context.Context) (Stream, error) {
	s.closeMx.Lock()
	closeErr := s.closeErr
	s.closeMx.Unlock()
	if closeErr != nil {
		return nil, closeErr
	}

	for {
		// If there's a stream in the accept queue, return it immediately.
		if str := s.bidiAcceptQueue.Next(); str != nil {
			return str, nil
		}
		// No stream in the accept queue. Wait until we accept one.
		select {
		case <-s.ctx.Done():
			return nil, s.closeErr
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-s.bidiAcceptQueue.Chan():
		}
	}
}

func (s *Session) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	s.closeMx.Lock()
	closeErr := s.closeErr
	s.closeMx.Unlock()
	if closeErr != nil {
		return nil, s.closeErr
	}

	for {
		// If there's a stream in the accept queue, return it immediately.
		if str := s.uniAcceptQueue.Next(); str != nil {
			return str, nil
		}
		// No stream in the accept queue. Wait until we accept one.
		select {
		case <-s.ctx.Done():
			return nil, s.closeErr
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-s.uniAcceptQueue.Chan():
		}
	}
}

func (s *Session) OpenStream() (Stream, error) {
	s.closeMx.Lock()
	defer s.closeMx.Unlock()

	if s.closeErr != nil {
		return nil, s.closeErr
	}

	qstr, err := s.qconn.OpenStream()
	if err != nil {
		return nil, err
	}
	return s.addStream(qstr, true), nil
}

func (s *Session) addStreamCtxCancel(cancel context.CancelFunc) (id int) {
rand:
	id = rand.Int()
	if _, ok := s.streamCtxs[id]; ok {
		goto rand
	}
	s.streamCtxs[id] = cancel
	return id
}

func (s *Session) OpenStreamSync(ctx context.Context) (Stream, error) {
	s.closeMx.Lock()
	if s.closeErr != nil {
		s.closeMx.Unlock()
		return nil, s.closeErr
	}
	ctx, cancel := context.WithCancel(ctx)
	id := s.addStreamCtxCancel(cancel)
	s.closeMx.Unlock()

	qstr, err := s.qconn.OpenStreamSync(ctx)
	if err != nil {
		if s.closeErr != nil {
			return nil, s.closeErr
		}
		return nil, err
	}

	s.closeMx.Lock()
	defer s.closeMx.Unlock()
	delete(s.streamCtxs, id)
	// Some time might have passed. Check if the session is still alive
	if s.closeErr != nil {
		qstr.CancelWrite(sessionCloseErrorCode)
		qstr.CancelRead(sessionCloseErrorCode)
		return nil, s.closeErr
	}
	return s.addStream(qstr, true), nil
}

func (s *Session) OpenUniStream() (SendStream, error) {
	s.closeMx.Lock()
	defer s.closeMx.Unlock()

	if s.closeErr != nil {
		return nil, s.closeErr
	}
	qstr, err := s.qconn.OpenUniStream()
	if err != nil {
		return nil, err
	}
	return s.addSendStream(qstr), nil
}

func (s *Session) OpenUniStreamSync(ctx context.Context) (str SendStream, err error) {
	s.closeMx.Lock()
	if s.closeErr != nil {
		s.closeMx.Unlock()
		return nil, s.closeErr
	}
	ctx, cancel := context.WithCancel(ctx)
	id := s.addStreamCtxCancel(cancel)
	s.closeMx.Unlock()

	qstr, err := s.qconn.OpenUniStreamSync(ctx)
	if err != nil {
		if s.closeErr != nil {
			return nil, s.closeErr
		}
		return nil, err
	}

	s.closeMx.Lock()
	defer s.closeMx.Unlock()
	delete(s.streamCtxs, id)
	// Some time might have passed. Check if the session is still alive
	if s.closeErr != nil {
		qstr.CancelWrite(sessionCloseErrorCode)
		return nil, s.closeErr
	}
	return s.addSendStream(qstr), nil
}

func (s *Session) LocalAddr() net.Addr {
	return s.qconn.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.qconn.RemoteAddr()
}

func (s *Session) CloseWithError(code SessionErrorCode, msg string) error {
	first, err := s.closeWithError(code, msg)
	if err != nil || !first {
		return err
	}

	s.requestStr.CancelRead(1337)
	err = s.requestStr.Close()
	<-s.ctx.Done()
	return err
}

func (s *Session) closeWithError(code SessionErrorCode, msg string) (bool /* first call to close session */, error) {
	s.closeMx.Lock()
	defer s.closeMx.Unlock()
	// Duplicate call, or the remote already closed this session.
	if s.closeErr != nil {
		return false, nil
	}
	s.closeErr = &ConnectionError{
		ErrorCode: code,
		Message:   msg,
	}

	b := make([]byte, 4, 4+len(msg))
	binary.BigEndian.PutUint32(b, uint32(code))
	b = append(b, []byte(msg)...)

	return true, http3.WriteCapsule(
		quicvarint.NewWriter(s.requestStr),
		closeWebtransportSessionCapsuleType,
		b,
	)
}

func (c *Session) ConnectionState() quic.ConnectionState {
	return c.qconn.ConnectionState()
}
