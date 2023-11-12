package webtransport

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

const sessionCloseErrorCode quic.StreamErrorCode = 0x170d7b68

type SendStream interface {
	io.Writer
	io.Closer

	StreamID() quic.StreamID
	CancelWrite(StreamErrorCode)

	SetWriteDeadline(time.Time) error
}

type ReceiveStream interface {
	io.Reader

	StreamID() quic.StreamID
	CancelRead(StreamErrorCode)

	SetReadDeadline(time.Time) error
}

type Stream interface {
	SendStream
	ReceiveStream
	SetDeadline(time.Time) error
}

type sendStream struct {
	str quic.SendStream
	// WebTransport stream header.
	// Set by the constructor, set to nil once sent out.
	// Might be initialized to nil if this sendStream is part of an incoming bidirectional stream.
	streamHdr []byte

	onClose func()

	once sync.Once
}

var _ SendStream = &sendStream{}

func newSendStream(str quic.SendStream, hdr []byte, onClose func()) *sendStream {
	return &sendStream{str: str, streamHdr: hdr, onClose: onClose}
}

func (s *sendStream) maybeSendStreamHeader() (err error) {
	s.once.Do(func() {
		if _, e := s.str.Write(s.streamHdr); e != nil {
			err = e
			return
		}
		s.streamHdr = nil
	})
	return
}

func (s *sendStream) Write(b []byte) (int, error) {
	if err := s.maybeSendStreamHeader(); err != nil {
		return 0, err
	}
	n, err := s.str.Write(b)
	if err != nil && !isTimeoutError(err) {
		s.onClose()
	}
	return n, maybeConvertStreamError(err)
}

func (s *sendStream) CancelWrite(e StreamErrorCode) {
	s.str.CancelWrite(webtransportCodeToHTTPCode(e))
	s.onClose()
}

func (s *sendStream) closeWithSession() {
	s.str.CancelWrite(sessionCloseErrorCode)
}

func (s *sendStream) Close() error {
	if err := s.maybeSendStreamHeader(); err != nil {
		return err
	}
	s.onClose()
	return maybeConvertStreamError(s.str.Close())
}

func (s *sendStream) SetWriteDeadline(t time.Time) error {
	return maybeConvertStreamError(s.str.SetWriteDeadline(t))
}

func (s *sendStream) StreamID() quic.StreamID {
	return s.str.StreamID()
}

type receiveStream struct {
	str     quic.ReceiveStream
	onClose func()
}

var _ ReceiveStream = &receiveStream{}

func newReceiveStream(str quic.ReceiveStream, onClose func()) *receiveStream {
	return &receiveStream{str: str, onClose: onClose}
}

func (s *receiveStream) Read(b []byte) (int, error) {
	n, err := s.str.Read(b)
	if err != nil && !isTimeoutError(err) {
		s.onClose()
	}
	return n, maybeConvertStreamError(err)
}

func (s *receiveStream) CancelRead(e StreamErrorCode) {
	s.str.CancelRead(webtransportCodeToHTTPCode(e))
	s.onClose()
}

func (s *receiveStream) closeWithSession() {
	s.str.CancelRead(sessionCloseErrorCode)
}

func (s *receiveStream) SetReadDeadline(t time.Time) error {
	return maybeConvertStreamError(s.str.SetReadDeadline(t))
}

func (s *receiveStream) StreamID() quic.StreamID {
	return s.str.StreamID()
}

type stream struct {
	*sendStream
	*receiveStream

	mx                             sync.Mutex
	sendSideClosed, recvSideClosed bool
	onClose                        func()
}

var _ Stream = &stream{}

func newStream(str quic.Stream, hdr []byte, onClose func()) *stream {
	s := &stream{onClose: onClose}
	s.sendStream = newSendStream(str, hdr, func() { s.registerClose(true) })
	s.receiveStream = newReceiveStream(str, func() { s.registerClose(false) })
	return s
}

func (s *stream) registerClose(isSendSide bool) {
	s.mx.Lock()
	if isSendSide {
		s.sendSideClosed = true
	} else {
		s.recvSideClosed = true
	}
	isClosed := s.sendSideClosed && s.recvSideClosed
	s.mx.Unlock()

	if isClosed {
		s.onClose()
	}
}

func (s *stream) closeWithSession() {
	s.sendStream.closeWithSession()
	s.receiveStream.closeWithSession()
}

func (s *stream) SetDeadline(t time.Time) error {
	err1 := s.sendStream.SetWriteDeadline(t)
	err2 := s.receiveStream.SetReadDeadline(t)
	if err1 != nil {
		return err1
	}
	return err2
}

func (s *stream) StreamID() quic.StreamID {
	return s.receiveStream.StreamID()
}

func maybeConvertStreamError(err error) error {
	if err == nil {
		return nil
	}
	var streamErr *quic.StreamError
	if errors.As(err, &streamErr) {
		errorCode, cerr := httpCodeToWebtransportCode(streamErr.ErrorCode)
		if cerr != nil {
			return fmt.Errorf("stream reset, but failed to convert stream error %d: %w", streamErr.ErrorCode, cerr)
		}
		return &StreamError{ErrorCode: errorCode}
	}
	return err
}

func isTimeoutError(err error) bool {
	nerr, ok := err.(net.Error)
	if !ok {
		return false
	}
	return nerr.Timeout()
}
