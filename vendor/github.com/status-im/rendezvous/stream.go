package rendezvous

import (
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var (
	ingressTrafficMeter = metrics.NewRegisteredMeter("rendezvous/InboundTraffic", nil)
	egressTrafficMeter  = metrics.NewRegisteredMeter("rendezvous/OutboundTraffic", nil)
)

// InstrumentedStream implements read writer interface and collects metrics.
type InstrumentedStream struct {
	s network.Stream
}

func (si InstrumentedStream) CloseWrite() error {
	return si.s.CloseWrite()
}

func (si InstrumentedStream) CloseRead() error {
	return si.s.CloseRead()
}

func (si InstrumentedStream) ID() string {
	return si.s.ID()
}

func (si InstrumentedStream) Write(p []byte) (int, error) {
	n, err := si.s.Write(p)
	egressTrafficMeter.Mark(int64(n))
	return n, err
}

func (si InstrumentedStream) Read(p []byte) (int, error) {
	n, err := si.s.Read(p)
	ingressTrafficMeter.Mark(int64(n))
	return n, err
}

func (si InstrumentedStream) Close() error {
	return si.s.Close()
}

func (si InstrumentedStream) Reset() error {
	return si.s.Reset()
}

func (si InstrumentedStream) SetDeadline(timeout time.Time) error {
	return si.s.SetDeadline(timeout)
}

func (si InstrumentedStream) SetReadDeadline(timeout time.Time) error {
	return si.s.SetReadDeadline(timeout)
}

func (si InstrumentedStream) SetWriteDeadline(timeout time.Time) error {
	return si.s.SetWriteDeadline(timeout)
}

func (si InstrumentedStream) Protocol() protocol.ID {
	return si.s.Protocol()
}

func (si InstrumentedStream) SetProtocol(pid protocol.ID) error {
	return si.s.SetProtocol(pid)
}

func (si InstrumentedStream) Conn() network.Conn {
	return si.s.Conn()
}

func (si InstrumentedStream) Stat() network.Stats {
	return si.s.Stat()
}

func (si InstrumentedStream) Scope() network.StreamScope {
	return si.s.Scope()
}
