package pubsub

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/protocol"
)

// GossipSubFeatureTest is a feature test function; it takes a feature and a protocol ID and
// should return true if the feature is supported by the protocol
type GossipSubFeatureTest = func(GossipSubFeature, protocol.ID) bool

// GossipSubFeature is a feature discriminant enum
type GossipSubFeature int

const (
	// Protocol supports basic GossipSub Mesh -- gossipsub-v1.0 compatible
	GossipSubFeatureMesh = iota
	// Protocol supports Peer eXchange on prune -- gossipsub-v1.1 compatible
	GossipSubFeaturePX
)

// GossipSubDefaultProtocols is the default gossipsub router protocol list
var GossipSubDefaultProtocols = []protocol.ID{GossipSubID_v11, GossipSubID_v10, FloodSubID}

// GossipSubDefaultFeatures is the feature test function for the default gossipsub protocols
func GossipSubDefaultFeatures(feat GossipSubFeature, proto protocol.ID) bool {
	switch feat {
	case GossipSubFeatureMesh:
		return proto == GossipSubID_v11 || proto == GossipSubID_v10
	case GossipSubFeaturePX:
		return proto == GossipSubID_v11
	default:
		return false
	}
}

// WithGossipSubProtocols is a gossipsub router option that configures a custom protocol list
// and feature test function
func WithGossipSubProtocols(protos []protocol.ID, feature GossipSubFeatureTest) Option {
	return func(ps *PubSub) error {
		gs, ok := ps.rt.(*GossipSubRouter)
		if !ok {
			return fmt.Errorf("pubsub router is not gossipsub")
		}

		gs.protos = protos
		gs.feature = feature

		return nil
	}
}
