package pubsub

import (
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var _ network.Notifiee = (*PubSubNotif)(nil)

type PubSubNotif PubSub

func (p *PubSubNotif) OpenedStream(n network.Network, s network.Stream) {
}

func (p *PubSubNotif) ClosedStream(n network.Network, s network.Stream) {
}

func (p *PubSubNotif) Connected(n network.Network, c network.Conn) {
	// ignore transient connections
	if c.Stat().Transient {
		return
	}

	go func() {
		p.newPeersPrioLk.RLock()
		p.newPeersMx.Lock()
		p.newPeersPend[c.RemotePeer()] = struct{}{}
		p.newPeersMx.Unlock()
		p.newPeersPrioLk.RUnlock()

		select {
		case p.newPeers <- struct{}{}:
		default:
		}
	}()
}

func (p *PubSubNotif) Disconnected(n network.Network, c network.Conn) {
}

func (p *PubSubNotif) Listen(n network.Network, _ ma.Multiaddr) {
}

func (p *PubSubNotif) ListenClose(n network.Network, _ ma.Multiaddr) {
}

func (p *PubSubNotif) Initialize() {
	isTransient := func(pid peer.ID) bool {
		for _, c := range p.host.Network().ConnsToPeer(pid) {
			if !c.Stat().Transient {
				return false
			}
		}

		return true
	}

	p.newPeersPrioLk.RLock()
	p.newPeersMx.Lock()
	for _, pid := range p.host.Network().Peers() {
		if isTransient(pid) {
			continue
		}

		p.newPeersPend[pid] = struct{}{}
	}
	p.newPeersMx.Unlock()
	p.newPeersPrioLk.RUnlock()

	select {
	case p.newPeers <- struct{}{}:
	default:
	}
}
