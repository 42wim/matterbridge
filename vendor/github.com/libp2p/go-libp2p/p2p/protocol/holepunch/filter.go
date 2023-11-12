package holepunch

import (
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// WithAddrFilter is a Service option that enables multiaddress filtering.
// It allows to only send a subset of observed addresses to the remote
// peer. E.g., only announce TCP or QUIC multi addresses instead of both.
// It also allows to only consider a subset of received multi addresses
// that remote peers announced to us.
// Theoretically, this API also allows to add multi addresses in both cases.
func WithAddrFilter(f AddrFilter) Option {
	return func(hps *Service) error {
		hps.filter = f
		return nil
	}
}

// AddrFilter defines the interface for the multi address filtering.
type AddrFilter interface {
	// FilterLocal filters the multi addresses that are sent to the remote peer.
	FilterLocal(remoteID peer.ID, maddrs []ma.Multiaddr) []ma.Multiaddr
	// FilterRemote filters the multi addresses received from the remote peer.
	FilterRemote(remoteID peer.ID, maddrs []ma.Multiaddr) []ma.Multiaddr
}
