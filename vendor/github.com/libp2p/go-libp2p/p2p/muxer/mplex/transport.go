package mplex

import (
	"net"

	"github.com/libp2p/go-libp2p/core/network"

	mp "github.com/libp2p/go-mplex"
)

// DefaultTransport has default settings for Transport
var DefaultTransport = &Transport{}

const ID = "/mplex/6.7.0"

var _ network.Multiplexer = &Transport{}

// Transport implements mux.Multiplexer that constructs
// mplex-backed muxed connections.
type Transport struct{}

func (t *Transport) NewConn(nc net.Conn, isServer bool, scope network.PeerScope) (network.MuxedConn, error) {
	m, err := mp.NewMultiplex(nc, isServer, scope)
	if err != nil {
		return nil, err
	}
	return NewMuxedConn(m), nil
}
