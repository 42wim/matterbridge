package pubsub

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	FloodSubID              = protocol.ID("/floodsub/1.0.0")
	FloodSubTopicSearchSize = 5
)

// NewFloodsubWithProtocols returns a new floodsub-enabled PubSub objecting using the protocols specified in ps.
func NewFloodsubWithProtocols(ctx context.Context, h host.Host, ps []protocol.ID, opts ...Option) (*PubSub, error) {
	rt := &FloodSubRouter{
		protocols: ps,
	}
	return NewPubSub(ctx, h, rt, opts...)
}

// NewFloodSub returns a new PubSub object using the FloodSubRouter.
func NewFloodSub(ctx context.Context, h host.Host, opts ...Option) (*PubSub, error) {
	return NewFloodsubWithProtocols(ctx, h, []protocol.ID{FloodSubID}, opts...)
}

type FloodSubRouter struct {
	p         *PubSub
	protocols []protocol.ID
	tracer    *pubsubTracer
}

func (fs *FloodSubRouter) Protocols() []protocol.ID {
	return fs.protocols
}

func (fs *FloodSubRouter) Attach(p *PubSub) {
	fs.p = p
	fs.tracer = p.tracer
}

func (fs *FloodSubRouter) AddPeer(p peer.ID, proto protocol.ID) {
	fs.tracer.AddPeer(p, proto)
}

func (fs *FloodSubRouter) RemovePeer(p peer.ID) {
	fs.tracer.RemovePeer(p)
}

func (fs *FloodSubRouter) EnoughPeers(topic string, suggested int) bool {
	// check all peers in the topic
	tmap, ok := fs.p.topics[topic]
	if !ok {
		return false
	}

	if suggested == 0 {
		suggested = FloodSubTopicSearchSize
	}

	if len(tmap) >= suggested {
		return true
	}

	return false
}

func (fs *FloodSubRouter) AcceptFrom(peer.ID) AcceptStatus {
	return AcceptAll
}

func (fs *FloodSubRouter) HandleRPC(rpc *RPC) {}

func (fs *FloodSubRouter) Publish(msg *Message) {
	from := msg.ReceivedFrom
	topic := msg.GetTopic()

	out := rpcWithMessages(msg.Message)
	for pid := range fs.p.topics[topic] {
		if pid == from || pid == peer.ID(msg.GetFrom()) {
			continue
		}

		mch, ok := fs.p.peers[pid]
		if !ok {
			continue
		}

		select {
		case mch <- out:
			fs.tracer.SendRPC(out, pid)
		default:
			log.Infof("dropping message to peer %s: queue full", pid)
			fs.tracer.DropRPC(out, pid)
			// Drop it. The peer is too slow.
		}
	}
}

func (fs *FloodSubRouter) Join(topic string) {
	fs.tracer.Join(topic)
}

func (fs *FloodSubRouter) Leave(topic string) {
	fs.tracer.Leave(topic)
}
