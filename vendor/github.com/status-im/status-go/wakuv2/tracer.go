package wakuv2

import (
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"

	"go.uber.org/zap"
)

// Trace implements EventTracer interface.
// We use custom logging, because we want to base58-encode the peerIDs. And also make the messageIDs readable.
func (w *Waku) Trace(evt *pubsub_pb.TraceEvent) {

	f := []zap.Field{
		zap.String("type", evt.Type.String()),
	}

	encode := func(peerID []byte) string {
		return base58.Encode(peerID)
	}

	encodeMeta := func(meta *pubsub_pb.TraceEvent_RPCMeta) []zap.Field {
		if meta == nil {
			return nil
		}
		var out []zap.Field
		for i, msg := range meta.Messages {
			out = append(out, zap.String(fmt.Sprintf("MessageID[%d]", i), encode(msg.GetMessageID())))
			out = append(out, zap.Stringp(fmt.Sprintf("Topic[%d]", i), msg.Topic))
		}
		for i, sub := range meta.Subscription {
			out = append(out, zap.Boolp(fmt.Sprintf("Subscribe[%d]", i), sub.Subscribe))
			out = append(out, zap.Stringp(fmt.Sprintf("Topic[%d]", i), sub.Topic))
		}
		if meta.Control != nil {
			out = append(out, zap.Any("Control", meta.Control))
		}
		return out
	}

	if evt.PublishMessage != nil {
		f = append(f, zap.String("MessageID", encode(evt.PublishMessage.MessageID)))
	}
	if evt.RejectMessage != nil {
		f = append(f, zap.String("MessageID", encode(evt.RejectMessage.MessageID)))
	}
	if evt.DuplicateMessage != nil {
		f = append(f, zap.String("MessageID", encode(evt.DuplicateMessage.MessageID)))
	}
	if evt.DeliverMessage != nil {
		f = append(f, zap.String("MessageID", encode(evt.DeliverMessage.MessageID)))
	}

	if evt.AddPeer != nil {
		f = append(f, zap.String("PeerID", encode(evt.AddPeer.GetPeerID())))
	}
	if evt.RemovePeer != nil {
		f = append(f, zap.String("PeerID", encode(evt.RemovePeer.GetPeerID())))
	}

	if evt.RecvRPC != nil {
		f = append(f, zap.String("ReceivedFrom", encode(evt.RecvRPC.GetReceivedFrom())))
		f = append(f, encodeMeta(evt.RecvRPC.Meta)...)
	}
	if evt.SendRPC != nil {
		f = append(f, zap.String("SendTo", encode(evt.SendRPC.GetSendTo())))
		f = append(f, encodeMeta(evt.SendRPC.Meta)...)
	}
	if evt.DropRPC != nil {
		f = append(f, zap.String("SendTo", encode(evt.DropRPC.GetSendTo())))
		f = append(f, encodeMeta(evt.DropRPC.Meta)...)
	}

	if evt.Join != nil {
		f = append(f, zap.String("Topic", evt.Join.GetTopic()))
	}
	if evt.Leave != nil {
		f = append(f, zap.String("Topic", evt.Leave.GetTopic()))
	}
	if evt.Graft != nil {
		f = append(f, zap.String("PeerID", encode(evt.Graft.GetPeerID())))
		f = append(f, zap.String("Topic", evt.Graft.GetTopic()))
	}
	if evt.Prune != nil {
		f = append(f, zap.String("PeerID", encode(evt.Prune.GetPeerID())))
		f = append(f, zap.String("Topic", evt.Prune.GetTopic()))
	}

	w.logger.With(f...).Debug("pubsub trace")
}
