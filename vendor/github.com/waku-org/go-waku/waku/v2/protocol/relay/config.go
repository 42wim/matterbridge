package relay

import (
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/waku-org/go-waku/waku/v2/hash"
	waku_proto "github.com/waku-org/go-waku/waku/v2/protocol"
)

var DefaultRelaySubscriptionBufferSize int = 1024

type RelaySubscribeParameters struct {
	dontConsume bool
	cacheSize   uint
}

type RelaySubscribeOption func(*RelaySubscribeParameters) error

// WithoutConsumer option let's a user subscribe to relay without consuming messages received.
// This is useful for a relayNode where only a subscribe is required in order to relay messages in gossipsub network.
func WithoutConsumer() RelaySubscribeOption {
	return func(params *RelaySubscribeParameters) error {
		params.dontConsume = true
		return nil
	}
}

func WithCacheSize(size uint) RelaySubscribeOption {
	return func(params *RelaySubscribeParameters) error {
		params.cacheSize = size
		return nil
	}
}

func msgIDFn(pmsg *pubsub_pb.Message) string {
	return string(hash.SHA256(pmsg.Data))
}

func (w *WakuRelay) setDefaultPeerScoreParams() {
	w.peerScoreParams = &pubsub.PeerScoreParams{
		Topics:        make(map[string]*pubsub.TopicScoreParams),
		DecayInterval: 12 * time.Second, // how often peer scoring is updated
		DecayToZero:   0.01,             // below this we consider the parameter to be zero
		RetainScore:   10 * time.Minute, // remember peer score during x after it disconnects
		// p5: application specific, unset
		AppSpecificScore: func(p peer.ID) float64 {
			return 0
		},
		AppSpecificWeight: 0.0,
		// p6: penalizes peers sharing more than threshold ips
		IPColocationFactorWeight:    -50,
		IPColocationFactorThreshold: 5.0,
		// p7: penalizes bad behaviour (weight and decay)
		BehaviourPenaltyWeight: -10,
		BehaviourPenaltyDecay:  0.986,
	}
	w.peerScoreThresholds = &pubsub.PeerScoreThresholds{
		GossipThreshold:             -100,   // no gossip is sent to peers below this score
		PublishThreshold:            -1000,  // no self-published msgs are sent to peers below this score
		GraylistThreshold:           -10000, // used to trigger disconnections + ignore peer if below this score
		OpportunisticGraftThreshold: 0,      // grafts better peers if the mesh median score drops below this. unset.
	}
}

func (w *WakuRelay) defaultPubsubOptions() []pubsub.Option {

	cfg := pubsub.DefaultGossipSubParams()
	cfg.PruneBackoff = time.Minute
	cfg.UnsubscribeBackoff = 5 * time.Second
	cfg.GossipFactor = 0.25
	cfg.D = waku_proto.GossipSubOptimalFullMeshSize
	cfg.Dlo = 4
	cfg.Dhi = 8
	cfg.Dout = 3
	cfg.Dlazy = waku_proto.GossipSubOptimalFullMeshSize
	cfg.HeartbeatInterval = time.Second
	cfg.HistoryLength = 6
	cfg.HistoryGossip = 3
	cfg.FanoutTTL = time.Minute

	w.setDefaultPeerScoreParams()

	w.setDefaultTopicParams()
	return []pubsub.Option{
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign),
		pubsub.WithNoAuthor(),
		pubsub.WithMessageIdFn(msgIDFn),
		pubsub.WithGossipSubProtocols(
			[]protocol.ID{WakuRelayID_v200, pubsub.GossipSubID_v11, pubsub.GossipSubID_v10, pubsub.FloodSubID},
			func(feat pubsub.GossipSubFeature, proto protocol.ID) bool {
				switch feat {
				case pubsub.GossipSubFeatureMesh:
					return proto == pubsub.GossipSubID_v11 || proto == pubsub.GossipSubID_v10 || proto == WakuRelayID_v200
				case pubsub.GossipSubFeaturePX:
					return proto == pubsub.GossipSubID_v11 || proto == WakuRelayID_v200
				default:
					return false
				}
			},
		),
		pubsub.WithGossipSubParams(cfg),
		pubsub.WithFloodPublish(true),
		pubsub.WithSeenMessagesTTL(2 * time.Minute),
		pubsub.WithPeerScore(w.peerScoreParams, w.peerScoreThresholds),
		pubsub.WithPeerScoreInspect(w.peerScoreInspector, 6*time.Second),
	}
}

func (w *WakuRelay) setDefaultTopicParams() {
	w.topicParams = &pubsub.TopicScoreParams{
		TopicWeight: 1,
		// p1: favours peers already in the mesh
		TimeInMeshWeight:  0.01,
		TimeInMeshQuantum: time.Second,
		TimeInMeshCap:     10.0,
		// p2: rewards fast peers
		FirstMessageDeliveriesWeight: 1.0,
		FirstMessageDeliveriesDecay:  0.5,
		FirstMessageDeliveriesCap:    10.0,
		// p3: penalizes lazy peers. safe low value
		MeshMessageDeliveriesWeight:     0,
		MeshMessageDeliveriesDecay:      0,
		MeshMessageDeliveriesCap:        0,
		MeshMessageDeliveriesThreshold:  0,
		MeshMessageDeliveriesWindow:     0,
		MeshMessageDeliveriesActivation: 0,
		// p3b: tracks history of prunes
		MeshFailurePenaltyWeight: 0,
		MeshFailurePenaltyDecay:  0,
		// p4: penalizes invalid messages. highly penalize peers sending wrong messages
		InvalidMessageDeliveriesWeight: -100.0,
		InvalidMessageDeliveriesDecay:  0.5,
	}
}
