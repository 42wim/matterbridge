package udp

import (
	"context"
	"net"

	"github.com/anacrolix/log"

	"github.com/anacrolix/missinggo/v2"
)

type NewConnClientOpts struct {
	// The network to operate to use, such as "udp4", "udp", "udp6".
	Network string
	// Tracker address
	Host string
	// If non-nil, forces either IPv4 or IPv6 in the UDP tracker wire protocol.
	Ipv6 *bool
	// Logger to use for internal errors.
	Logger log.Logger
}

// Manages a Client with a specific connection.
type ConnClient struct {
	Client  Client
	conn    net.PacketConn
	d       Dispatcher
	readErr error
	closed  bool
	newOpts NewConnClientOpts
}

func (cc *ConnClient) reader() {
	b := make([]byte, 0x800)
	for {
		n, addr, err := cc.conn.ReadFrom(b)
		if err != nil {
			// TODO: Do bad things to the dispatcher, and incoming calls to the client if we have a
			// read error.
			cc.readErr = err
			if !cc.closed {
				panic(err)
			}
			break
		}
		err = cc.d.Dispatch(b[:n], addr)
		if err != nil {
			cc.newOpts.Logger.Levelf(log.Debug, "dispatching packet received on %v: %v", cc.conn.LocalAddr(), err)
		}
	}
}

func ipv6(opt *bool, network string, remoteAddr net.Addr) bool {
	if opt != nil {
		return *opt
	}
	switch network {
	case "udp4":
		return false
	case "udp6":
		return true
	}
	rip := missinggo.AddrIP(remoteAddr)
	return rip.To16() != nil && rip.To4() == nil
}

// Allows a UDP Client to write packets to an endpoint without knowing about the network specifics.
type clientWriter struct {
	pc      net.PacketConn
	network string
	address string
}

func (me clientWriter) Write(p []byte) (n int, err error) {
	addr, err := net.ResolveUDPAddr(me.network, me.address)
	if err != nil {
		return
	}
	return me.pc.WriteTo(p, addr)
}

func NewConnClient(opts NewConnClientOpts) (cc *ConnClient, err error) {
	conn, err := net.ListenPacket(opts.Network, ":0")
	if err != nil {
		return
	}
	if opts.Logger.IsZero() {
		opts.Logger = log.Default
	}
	cc = &ConnClient{
		Client: Client{
			Writer: clientWriter{
				pc:      conn,
				network: opts.Network,
				address: opts.Host,
			},
		},
		conn:    conn,
		newOpts: opts,
	}
	cc.Client.Dispatcher = &cc.d
	go cc.reader()
	return
}

func (cc *ConnClient) Close() error {
	cc.closed = true
	return cc.conn.Close()
}

func (cc *ConnClient) Announce(
	ctx context.Context, req AnnounceRequest, opts Options,
) (
	h AnnounceResponseHeader, nas AnnounceResponsePeers, err error,
) {
	return cc.Client.Announce(ctx, req, opts, func(addr net.Addr) bool {
		return ipv6(cc.newOpts.Ipv6, cc.newOpts.Network, addr)
	})
}

func (cc *ConnClient) LocalAddr() net.Addr {
	return cc.conn.LocalAddr()
}
