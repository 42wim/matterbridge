package torrent

import (
	"context"
	"net"
)

// Dialers have the network locked in.
type Dialer interface {
	Dial(_ context.Context, addr string) (net.Conn, error)
	DialerNetwork() string
}

// An interface to ease wrapping dialers that explicitly include a network parameter.
type DialContexter interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

// Used by wrappers of standard library network types.
var DefaultNetDialer = &net.Dialer{}

// Adapts a DialContexter to the Dial interface in this package.
type NetworkDialer struct {
	Network string
	Dialer  DialContexter
}

func (me NetworkDialer) DialerNetwork() string {
	return me.Network
}

func (me NetworkDialer) Dial(ctx context.Context, addr string) (_ net.Conn, err error) {
	return me.Dialer.DialContext(ctx, me.Network, addr)
}
