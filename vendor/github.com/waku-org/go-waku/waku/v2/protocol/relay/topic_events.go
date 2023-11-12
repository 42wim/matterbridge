package relay

import (
	"context"
	"errors"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/logging"
	"go.uber.org/zap"
)

// EvtRelaySubscribed is an event emitted when a new subscription to a pubsub topic is created
type EvtRelaySubscribed struct {
	Topic     string
	TopicInst *pubsub.Topic
}

// EvtRelayUnsubscribed is an event emitted when a subscription to a pubsub topic is closed
type EvtRelayUnsubscribed struct {
	Topic string
}

type PeerTopicState int

const (
	PEER_JOINED = iota
	PEER_LEFT
)

type EvtPeerTopic struct {
	PubsubTopic string
	PeerID      peer.ID
	State       PeerTopicState
}

// Events returns the event bus on which WakuRelay events will be emitted
func (w *WakuRelay) Events() event.Bus {
	return w.events
}

func (w *WakuRelay) addPeerTopicEventListener(topic *pubsub.Topic) (*pubsub.TopicEventHandler, error) {
	handler, err := topic.EventHandler()
	if err != nil {
		return nil, err
	}
	w.WaitGroup().Add(1)
	go w.topicEventPoll(topic.String(), handler)
	return handler, nil
}

func (w *WakuRelay) topicEventPoll(topic string, handler *pubsub.TopicEventHandler) {
	defer w.WaitGroup().Done()
	for {
		evt, err := handler.NextPeerEvent(w.Context())
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				break
			}
			w.log.Error("failed to get next peer event", zap.String("topic", topic), zap.Error(err))
			continue
		}
		if evt.Peer.Validate() != nil { //Empty peerEvent is returned when context passed in done.
			break
		}
		if evt.Type == pubsub.PeerJoin {
			w.log.Debug("received a PeerJoin event", zap.String("topic", topic), logging.HostID("peerID", evt.Peer))
			err = w.emitters.EvtPeerTopic.Emit(EvtPeerTopic{PubsubTopic: topic, PeerID: evt.Peer, State: PEER_JOINED})
			if err != nil {
				w.log.Error("failed to emit PeerJoin", zap.String("topic", topic), zap.Error(err))
			}
		} else if evt.Type == pubsub.PeerLeave {
			w.log.Debug("received a PeerLeave event", zap.String("topic", topic), logging.HostID("peerID", evt.Peer))
			err = w.emitters.EvtPeerTopic.Emit(EvtPeerTopic{PubsubTopic: topic, PeerID: evt.Peer, State: PEER_LEFT})
			if err != nil {
				w.log.Error("failed to emit PeerLeave", zap.String("topic", topic), zap.Error(err))
			}
		} else {
			w.log.Error("unknown event type received", zap.String("topic", topic),
				zap.Int("eventType", int(evt.Type)))
		}
	}
}

func (w *WakuRelay) CreateEventEmitters() error {
	var err error
	w.emitters.EvtRelaySubscribed, err = w.events.Emitter(new(EvtRelaySubscribed))
	if err != nil {
		return err
	}
	w.emitters.EvtRelayUnsubscribed, err = w.events.Emitter(new(EvtRelayUnsubscribed))
	if err != nil {
		return err
	}

	w.emitters.EvtPeerTopic, err = w.events.Emitter(new(EvtPeerTopic))
	if err != nil {
		return err
	}
	return nil
}
