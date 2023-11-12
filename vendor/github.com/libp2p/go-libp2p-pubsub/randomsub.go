package pubsub

import (
	"context"
	"math"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	RandomSubID = protocol.ID("/randomsub/1.0.0")
)

var (
	RandomSubD = 6
)

// NewRandomSub returns a new PubSub object using RandomSubRouter as the router.
func NewRandomSub(ctx context.Context, h host.Host, size int, opts ...Option) (*PubSub, error) {
	rt := &RandomSubRouter{
		size:  size,
		peers: make(map[peer.ID]protocol.ID),
	}
	return NewPubSub(ctx, h, rt, opts...)
}

// RandomSubRouter is a router that implements a random propagation strategy.
// For each message, it selects the square root of the network size peers, with a min of RandomSubD,
// and forwards the message to them.
type RandomSubRouter struct {
	p      *PubSub
	peers  map[peer.ID]protocol.ID
	size   int
	tracer *pubsubTracer
}

func (rs *RandomSubRouter) Protocols() []protocol.ID {
	return []protocol.ID{RandomSubID, FloodSubID}
}

func (rs *RandomSubRouter) Attach(p *PubSub) {
	rs.p = p
	rs.tracer = p.tracer
}

func (rs *RandomSubRouter) AddPeer(p peer.ID, proto protocol.ID) {
	rs.tracer.AddPeer(p, proto)
	rs.peers[p] = proto
}

func (rs *RandomSubRouter) RemovePeer(p peer.ID) {
	rs.tracer.RemovePeer(p)
	delete(rs.peers, p)
}

func (rs *RandomSubRouter) EnoughPeers(topic string, suggested int) bool {
	// check all peers in the topic
	tmap, ok := rs.p.topics[topic]
	if !ok {
		return false
	}

	fsPeers := 0
	rsPeers := 0

	// count floodsub and randomsub peers
	for p := range tmap {
		switch rs.peers[p] {
		case FloodSubID:
			fsPeers++
		case RandomSubID:
			rsPeers++
		}
	}

	if suggested == 0 {
		suggested = RandomSubD
	}

	if fsPeers+rsPeers >= suggested {
		return true
	}

	if rsPeers >= RandomSubD {
		return true
	}

	return false
}

func (rs *RandomSubRouter) AcceptFrom(peer.ID) AcceptStatus {
	return AcceptAll
}

func (rs *RandomSubRouter) HandleRPC(rpc *RPC) {}

func (rs *RandomSubRouter) Publish(msg *Message) {
	from := msg.ReceivedFrom

	tosend := make(map[peer.ID]struct{})
	rspeers := make(map[peer.ID]struct{})
	src := peer.ID(msg.GetFrom())

	topic := msg.GetTopic()
	tmap, ok := rs.p.topics[topic]
	if !ok {
		return
	}

	for p := range tmap {
		if p == from || p == src {
			continue
		}

		if rs.peers[p] == FloodSubID {
			tosend[p] = struct{}{}
		} else {
			rspeers[p] = struct{}{}
		}
	}

	if len(rspeers) > RandomSubD {
		target := RandomSubD
		sqrt := int(math.Ceil(math.Sqrt(float64(rs.size))))
		if sqrt > target {
			target = sqrt
		}
		if target > len(rspeers) {
			target = len(rspeers)
		}
		xpeers := peerMapToList(rspeers)
		shufflePeers(xpeers)
		xpeers = xpeers[:target]
		for _, p := range xpeers {
			tosend[p] = struct{}{}
		}
	} else {
		for p := range rspeers {
			tosend[p] = struct{}{}
		}
	}

	out := rpcWithMessages(msg.Message)
	for p := range tosend {
		mch, ok := rs.p.peers[p]
		if !ok {
			continue
		}

		select {
		case mch <- out:
			rs.tracer.SendRPC(out, p)
		default:
			log.Infof("dropping message to peer %s: queue full", p)
			rs.tracer.DropRPC(out, p)
		}
	}
}

func (rs *RandomSubRouter) Join(topic string) {
	rs.tracer.Join(topic)
}

func (rs *RandomSubRouter) Leave(topic string) {
	rs.tracer.Join(topic)
}
