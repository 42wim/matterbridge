package peermanager

import (
	"runtime"
	"sync"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/zap"
)

// ConnectionGater is the implementation of the connection gater used to limit
// the number of connections per IP address
type ConnectionGater struct {
	sync.Mutex
	logger  *zap.Logger
	limiter map[string]int
}

const maxConnsPerIP = 10

// NewConnectionGater creates a new instance of ConnectionGater
func NewConnectionGater(logger *zap.Logger) *ConnectionGater {
	c := &ConnectionGater{
		logger:  logger.Named("connection-gater"),
		limiter: make(map[string]int),
	}

	return c
}

// InterceptPeerDial is called on an imminent outbound peer dial request, prior
// to the addresses of that peer being available/resolved. Blocking connections
// at this stage is typical for blacklisting scenarios.
func (c *ConnectionGater) InterceptPeerDial(_ peer.ID) (allow bool) {
	return true
}

// InterceptAddrDial is called on an imminent outbound dial to a peer on a
// particular address. Blocking connections at this stage is typical for
// address filtering.
func (c *ConnectionGater) InterceptAddrDial(pid peer.ID, m multiaddr.Multiaddr) (allow bool) {
	return true
}

// InterceptAccept is called as soon as a transport listener receives an
// inbound connection request, before any upgrade takes place. Transports who
// accept already secure and/or multiplexed connections (e.g. possibly QUIC)
// MUST call this method regardless, for correctness/consistency.
func (c *ConnectionGater) InterceptAccept(n network.ConnMultiaddrs) (allow bool) {
	if !c.validateInboundConn(n.RemoteMultiaddr()) {
		runtime.Gosched() // Allow other go-routines to run in the event
		c.logger.Info("exceeds allowed inbound connections from this ip", zap.String("multiaddr", n.RemoteMultiaddr().String()))
		return false
	}

	return true
}

// InterceptSecured is called for both inbound and outbound connections,
// after a security handshake has taken place and we've authenticated the peer
func (c *ConnectionGater) InterceptSecured(_ network.Direction, _ peer.ID, _ network.ConnMultiaddrs) (allow bool) {
	return true
}

// InterceptUpgraded is called for inbound and outbound connections, after
// libp2p has finished upgrading the connection entirely to a secure,
// multiplexed channel.
func (c *ConnectionGater) InterceptUpgraded(_ network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

// NotifyDisconnect is called when a connection disconnects.
func (c *ConnectionGater) NotifyDisconnect(addr multiaddr.Multiaddr) {
	ip, err := manet.ToIP(addr)
	if err != nil {
		return
	}

	c.Lock()
	defer c.Unlock()

	currConnections, ok := c.limiter[ip.String()]
	if ok {
		currConnections--
		if currConnections <= 0 {
			delete(c.limiter, ip.String())
		} else {
			c.limiter[ip.String()] = currConnections
		}
	}
}

func (c *ConnectionGater) validateInboundConn(addr multiaddr.Multiaddr) bool {
	ip, err := manet.ToIP(addr)
	if err != nil {
		return false
	}

	c.Lock()
	defer c.Unlock()

	if currConnections := c.limiter[ip.String()]; currConnections+1 > maxConnsPerIP {
		return false
	}

	c.limiter[ip.String()]++
	return true
}
