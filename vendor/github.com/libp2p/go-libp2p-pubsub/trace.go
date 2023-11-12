package pubsub

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

// EventTracer is a generic event tracer interface.
// This is a high level tracing interface which delivers tracing events, as defined by the protobuf
// schema in pb/trace.proto.
type EventTracer interface {
	Trace(evt *pb.TraceEvent)
}

// RawTracer is a low level tracing interface that allows an application to trace the internal
// operation of the pubsub subsystem.
//
// Note that the tracers are invoked synchronously, which means that application tracers must
// take care to not block or modify arguments.
//
// Warning: this interface is not fixed, we may be adding new methods as necessitated by the system
// in the future.
type RawTracer interface {
	// AddPeer is invoked when a new peer is added.
	AddPeer(p peer.ID, proto protocol.ID)
	// RemovePeer is invoked when a peer is removed.
	RemovePeer(p peer.ID)
	// Join is invoked when a new topic is joined
	Join(topic string)
	// Leave is invoked when a topic is abandoned
	Leave(topic string)
	// Graft is invoked when a new peer is grafted on the mesh (gossipsub)
	Graft(p peer.ID, topic string)
	// Prune is invoked when a peer is pruned from the message (gossipsub)
	Prune(p peer.ID, topic string)
	// ValidateMessage is invoked when a message first enters the validation pipeline.
	ValidateMessage(msg *Message)
	// DeliverMessage is invoked when a message is delivered
	DeliverMessage(msg *Message)
	// RejectMessage is invoked when a message is Rejected or Ignored.
	// The reason argument can be one of the named strings Reject*.
	RejectMessage(msg *Message, reason string)
	// DuplicateMessage is invoked when a duplicate message is dropped.
	DuplicateMessage(msg *Message)
	// ThrottlePeer is invoked when a peer is throttled by the peer gater.
	ThrottlePeer(p peer.ID)
	// RecvRPC is invoked when an incoming RPC is received.
	RecvRPC(rpc *RPC)
	// SendRPC is invoked when a RPC is sent.
	SendRPC(rpc *RPC, p peer.ID)
	// DropRPC is invoked when an outbound RPC is dropped, typically because of a queue full.
	DropRPC(rpc *RPC, p peer.ID)
	// UndeliverableMessage is invoked when the consumer of Subscribe is not reading messages fast enough and
	// the pressure release mechanism trigger, dropping messages.
	UndeliverableMessage(msg *Message)
}

// pubsub tracer details
type pubsubTracer struct {
	tracer EventTracer
	raw    []RawTracer
	pid    peer.ID
	idGen  *msgIDGenerator
}

func (t *pubsubTracer) PublishMessage(msg *Message) {
	if t == nil {
		return
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_PUBLISH_MESSAGE.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		PublishMessage: &pb.TraceEvent_PublishMessage{
			MessageID: []byte(t.idGen.ID(msg)),
			Topic:     msg.Message.Topic,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) ValidateMessage(msg *Message) {
	if t == nil {
		return
	}

	if msg.ReceivedFrom != t.pid {
		for _, tr := range t.raw {
			tr.ValidateMessage(msg)
		}
	}
}

func (t *pubsubTracer) RejectMessage(msg *Message, reason string) {
	if t == nil {
		return
	}

	if msg.ReceivedFrom != t.pid {
		for _, tr := range t.raw {
			tr.RejectMessage(msg, reason)
		}
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_REJECT_MESSAGE.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		RejectMessage: &pb.TraceEvent_RejectMessage{
			MessageID:    []byte(t.idGen.ID(msg)),
			ReceivedFrom: []byte(msg.ReceivedFrom),
			Reason:       &reason,
			Topic:        msg.Topic,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) DuplicateMessage(msg *Message) {
	if t == nil {
		return
	}

	if msg.ReceivedFrom != t.pid {
		for _, tr := range t.raw {
			tr.DuplicateMessage(msg)
		}
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_DUPLICATE_MESSAGE.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		DuplicateMessage: &pb.TraceEvent_DuplicateMessage{
			MessageID:    []byte(t.idGen.ID(msg)),
			ReceivedFrom: []byte(msg.ReceivedFrom),
			Topic:        msg.Topic,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) DeliverMessage(msg *Message) {
	if t == nil {
		return
	}

	if msg.ReceivedFrom != t.pid {
		for _, tr := range t.raw {
			tr.DeliverMessage(msg)
		}
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_DELIVER_MESSAGE.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		DeliverMessage: &pb.TraceEvent_DeliverMessage{
			MessageID:    []byte(t.idGen.ID(msg)),
			Topic:        msg.Topic,
			ReceivedFrom: []byte(msg.ReceivedFrom),
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) AddPeer(p peer.ID, proto protocol.ID) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.AddPeer(p, proto)
	}

	if t.tracer == nil {
		return
	}

	protoStr := string(proto)
	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_ADD_PEER.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		AddPeer: &pb.TraceEvent_AddPeer{
			PeerID: []byte(p),
			Proto:  &protoStr,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) RemovePeer(p peer.ID) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.RemovePeer(p)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_REMOVE_PEER.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		RemovePeer: &pb.TraceEvent_RemovePeer{
			PeerID: []byte(p),
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) RecvRPC(rpc *RPC) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.RecvRPC(rpc)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_RECV_RPC.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		RecvRPC: &pb.TraceEvent_RecvRPC{
			ReceivedFrom: []byte(rpc.from),
			Meta:         t.traceRPCMeta(rpc),
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) SendRPC(rpc *RPC, p peer.ID) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.SendRPC(rpc, p)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_SEND_RPC.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		SendRPC: &pb.TraceEvent_SendRPC{
			SendTo: []byte(p),
			Meta:   t.traceRPCMeta(rpc),
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) DropRPC(rpc *RPC, p peer.ID) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.DropRPC(rpc, p)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_DROP_RPC.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		DropRPC: &pb.TraceEvent_DropRPC{
			SendTo: []byte(p),
			Meta:   t.traceRPCMeta(rpc),
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) UndeliverableMessage(msg *Message) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.UndeliverableMessage(msg)
	}
}

func (t *pubsubTracer) traceRPCMeta(rpc *RPC) *pb.TraceEvent_RPCMeta {
	rpcMeta := new(pb.TraceEvent_RPCMeta)

	var msgs []*pb.TraceEvent_MessageMeta
	for _, m := range rpc.Publish {
		msgs = append(msgs, &pb.TraceEvent_MessageMeta{
			MessageID: []byte(t.idGen.RawID(m)),
			Topic:     m.Topic,
		})
	}
	rpcMeta.Messages = msgs

	var subs []*pb.TraceEvent_SubMeta
	for _, sub := range rpc.Subscriptions {
		subs = append(subs, &pb.TraceEvent_SubMeta{
			Subscribe: sub.Subscribe,
			Topic:     sub.Topicid,
		})
	}
	rpcMeta.Subscription = subs

	if rpc.Control != nil {
		var ihave []*pb.TraceEvent_ControlIHaveMeta
		for _, ctl := range rpc.Control.Ihave {
			var mids [][]byte
			for _, mid := range ctl.MessageIDs {
				mids = append(mids, []byte(mid))
			}
			ihave = append(ihave, &pb.TraceEvent_ControlIHaveMeta{
				Topic:      ctl.TopicID,
				MessageIDs: mids,
			})
		}

		var iwant []*pb.TraceEvent_ControlIWantMeta
		for _, ctl := range rpc.Control.Iwant {
			var mids [][]byte
			for _, mid := range ctl.MessageIDs {
				mids = append(mids, []byte(mid))
			}
			iwant = append(iwant, &pb.TraceEvent_ControlIWantMeta{
				MessageIDs: mids,
			})
		}

		var graft []*pb.TraceEvent_ControlGraftMeta
		for _, ctl := range rpc.Control.Graft {
			graft = append(graft, &pb.TraceEvent_ControlGraftMeta{
				Topic: ctl.TopicID,
			})
		}

		var prune []*pb.TraceEvent_ControlPruneMeta
		for _, ctl := range rpc.Control.Prune {
			peers := make([][]byte, 0, len(ctl.Peers))
			for _, pi := range ctl.Peers {
				peers = append(peers, pi.PeerID)
			}
			prune = append(prune, &pb.TraceEvent_ControlPruneMeta{
				Topic: ctl.TopicID,
				Peers: peers,
			})
		}

		rpcMeta.Control = &pb.TraceEvent_ControlMeta{
			Ihave: ihave,
			Iwant: iwant,
			Graft: graft,
			Prune: prune,
		}
	}

	return rpcMeta
}

func (t *pubsubTracer) Join(topic string) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.Join(topic)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_JOIN.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		Join: &pb.TraceEvent_Join{
			Topic: &topic,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) Leave(topic string) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.Leave(topic)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_LEAVE.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		Leave: &pb.TraceEvent_Leave{
			Topic: &topic,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) Graft(p peer.ID, topic string) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.Graft(p, topic)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_GRAFT.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		Graft: &pb.TraceEvent_Graft{
			PeerID: []byte(p),
			Topic:  &topic,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) Prune(p peer.ID, topic string) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.Prune(p, topic)
	}

	if t.tracer == nil {
		return
	}

	now := time.Now().UnixNano()
	evt := &pb.TraceEvent{
		Type:      pb.TraceEvent_PRUNE.Enum(),
		PeerID:    []byte(t.pid),
		Timestamp: &now,
		Prune: &pb.TraceEvent_Prune{
			PeerID: []byte(p),
			Topic:  &topic,
		},
	}

	t.tracer.Trace(evt)
}

func (t *pubsubTracer) ThrottlePeer(p peer.ID) {
	if t == nil {
		return
	}

	for _, tr := range t.raw {
		tr.ThrottlePeer(p)
	}
}
