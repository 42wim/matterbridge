package lightpush

import (
	"errors"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"go.uber.org/zap"
)

type lightPushParameters struct {
	host              host.Host
	peerAddr          multiaddr.Multiaddr
	selectedPeer      peer.ID
	peerSelectionType peermanager.PeerSelection
	preferredPeers    peer.IDSlice
	requestID         []byte
	pm                *peermanager.PeerManager
	log               *zap.Logger
	pubsubTopic       string
}

// Option is the type of options accepted when performing LightPush protocol requests
type Option func(*lightPushParameters) error

// WithPeer is an option used to specify the peerID to push a waku message to
func WithPeer(p peer.ID) Option {
	return func(params *lightPushParameters) error {
		params.selectedPeer = p
		if params.peerAddr != nil {
			return errors.New("peerAddr and peerId options are mutually exclusive")
		}
		return nil
	}
}

// WithPeerAddr is an option used to specify a peerAddress
// This new peer will be added to peerStore.
// Note that this option is mutually exclusive to WithPeerAddr, only one of them can be used.
func WithPeerAddr(pAddr multiaddr.Multiaddr) Option {
	return func(params *lightPushParameters) error {
		params.peerAddr = pAddr
		if params.selectedPeer != "" {
			return errors.New("peerAddr and peerId options are mutually exclusive")
		}
		return nil
	}
}

// WithAutomaticPeerSelection is an option used to randomly select a peer from the peer store
// to push a waku message to. If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a peer
// from the node peerstore
func WithAutomaticPeerSelection(fromThesePeers ...peer.ID) Option {
	return func(params *lightPushParameters) error {
		params.peerSelectionType = peermanager.Automatic
		params.preferredPeers = fromThesePeers
		return nil
	}
}

// WithFastestPeerSelection is an option used to select a peer from the peer store
// with the lowest ping. If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a peer
// from the node peerstore
func WithFastestPeerSelection(fromThesePeers ...peer.ID) Option {
	return func(params *lightPushParameters) error {
		params.peerSelectionType = peermanager.LowestRTT
		return nil
	}
}

// WithPubSubTopic is used to specify the pubsub topic on which a WakuMessage will be broadcasted
func WithPubSubTopic(pubsubTopic string) Option {
	return func(params *lightPushParameters) error {
		params.pubsubTopic = pubsubTopic
		return nil
	}
}

// WithDefaultPubsubTopic is used to indicate that the message should be broadcasted in the default pubsub topic
func WithDefaultPubsubTopic() Option {
	return func(params *lightPushParameters) error {
		params.pubsubTopic = relay.DefaultWakuTopic
		return nil
	}
}

// WithRequestID is an option to set a specific request ID to be used when
// publishing a message
func WithRequestID(requestID []byte) Option {
	return func(params *lightPushParameters) error {
		params.requestID = requestID
		return nil
	}
}

// WithAutomaticRequestID is an option to automatically generate a request ID
// when publishing a message
func WithAutomaticRequestID() Option {
	return func(params *lightPushParameters) error {
		params.requestID = protocol.GenerateRequestID()
		return nil
	}
}

// DefaultOptions are the default options to be used when using the lightpush protocol
func DefaultOptions(host host.Host) []Option {
	return []Option{
		WithAutomaticRequestID(),
		WithAutomaticPeerSelection(),
	}
}
