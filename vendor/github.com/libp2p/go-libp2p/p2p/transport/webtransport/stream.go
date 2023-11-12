package libp2pwebtransport

import (
	"errors"
	"net"

	"github.com/libp2p/go-libp2p/core/network"

	"github.com/quic-go/webtransport-go"
)

const (
	reset webtransport.StreamErrorCode = 0
)

type webtransportStream struct {
	webtransport.Stream
	wsess *webtransport.Session
}

var _ net.Conn = &webtransportStream{}

func (s *webtransportStream) LocalAddr() net.Addr {
	return s.wsess.LocalAddr()
}

func (s *webtransportStream) RemoteAddr() net.Addr {
	return s.wsess.RemoteAddr()
}

type stream struct {
	webtransport.Stream
}

var _ network.MuxedStream = &stream{}

func (s *stream) Read(b []byte) (n int, err error) {
	n, err = s.Stream.Read(b)
	if err != nil && errors.Is(err, &webtransport.StreamError{}) {
		err = network.ErrReset
	}
	return n, err
}

func (s *stream) Write(b []byte) (n int, err error) {
	n, err = s.Stream.Write(b)
	if err != nil && errors.Is(err, &webtransport.StreamError{}) {
		err = network.ErrReset
	}
	return n, err
}

func (s *stream) Reset() error {
	s.Stream.CancelRead(reset)
	s.Stream.CancelWrite(reset)
	return nil
}

func (s *stream) Close() error {
	s.Stream.CancelRead(reset)
	return s.Stream.Close()
}

func (s *stream) CloseRead() error {
	s.Stream.CancelRead(reset)
	return nil
}

func (s *stream) CloseWrite() error {
	return s.Stream.Close()
}
