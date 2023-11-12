package legacy_filter

import (
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"go.uber.org/zap"
)

type (
	FilterSubscribeParameters struct {
		host              host.Host
		selectedPeer      peer.ID
		peerSelectionType peermanager.PeerSelection
		preferredPeers    peer.IDSlice
		log               *zap.Logger
	}

	FilterSubscribeOption func(*FilterSubscribeParameters)

	FilterParameters struct {
		Timeout time.Duration
		pm      *peermanager.PeerManager
	}

	Option func(*FilterParameters)
)

func WithTimeout(timeout time.Duration) Option {
	return func(params *FilterParameters) {
		params.Timeout = timeout
	}
}

func WithPeerManager(pm *peermanager.PeerManager) Option {
	return func(params *FilterParameters) {
		params.pm = pm
	}
}

func WithPeer(p peer.ID) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) {
		params.selectedPeer = p
	}
}

// WithAutomaticPeerSelection is an option used to randomly select a peer from the peer store.
// If a list of specific peers is passed, the peer will be chosen from that list assuming it
// supports the chosen protocol, otherwise it will chose a peer from the node peerstore
func WithAutomaticPeerSelection(fromThesePeers ...peer.ID) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) {
		params.peerSelectionType = peermanager.Automatic
		params.preferredPeers = fromThesePeers
	}
}

// WithFastestPeerSelection is an option used to select a peer from the peer store
// with the lowest ping If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a
// peer from the node peerstore
func WithFastestPeerSelection(fromThesePeers ...peer.ID) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) {
		params.peerSelectionType = peermanager.LowestRTT
	}
}

func DefaultOptions() []Option {
	return []Option{
		WithTimeout(24 * time.Hour),
	}
}

func DefaultSubscribtionOptions() []FilterSubscribeOption {
	return []FilterSubscribeOption{
		WithAutomaticPeerSelection(),
	}
}
