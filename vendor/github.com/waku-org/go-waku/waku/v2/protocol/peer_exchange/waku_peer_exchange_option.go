package peer_exchange

import (
	"errors"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"go.uber.org/zap"
)

type PeerExchangeParameters struct {
	host              host.Host
	selectedPeer      peer.ID
	peerAddr          multiaddr.Multiaddr
	peerSelectionType peermanager.PeerSelection
	preferredPeers    peer.IDSlice
	pm                *peermanager.PeerManager
	log               *zap.Logger
}

type PeerExchangeOption func(*PeerExchangeParameters) error

// WithPeer is an option used to specify the peerID to fetch peers from
func WithPeer(p peer.ID) PeerExchangeOption {
	return func(params *PeerExchangeParameters) error {
		params.selectedPeer = p
		if params.peerAddr != nil {
			return errors.New("peerAddr and peerId options are mutually exclusive")
		}
		return nil
	}
}

// WithPeerAddr is an option used to specify a peerAddress to fetch peers from
// This new peer will be added to peerStore.
// Note that this option is mutually exclusive to WithPeerAddr, only one of them can be used.
func WithPeerAddr(pAddr multiaddr.Multiaddr) PeerExchangeOption {
	return func(params *PeerExchangeParameters) error {
		params.peerAddr = pAddr
		if params.selectedPeer != "" {
			return errors.New("peerAddr and peerId options are mutually exclusive")
		}
		return nil
	}
}

// WithAutomaticPeerSelection is an option used to randomly select a peer from the Waku peer store
// to obtains peers from. If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a peer
// from the node peerstore
// Note: this option can only be used if WakuNode is initialized which internally intializes the peerManager
func WithAutomaticPeerSelection(fromThesePeers ...peer.ID) PeerExchangeOption {
	return func(params *PeerExchangeParameters) error {
		params.peerSelectionType = peermanager.Automatic
		params.preferredPeers = fromThesePeers
		return nil
	}
}

// WithFastestPeerSelection is an option used to select a peer from the peer store
// with the lowest ping. If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a peer
// from the node peerstore
func WithFastestPeerSelection(fromThesePeers ...peer.ID) PeerExchangeOption {
	return func(params *PeerExchangeParameters) error {
		params.peerSelectionType = peermanager.LowestRTT
		params.preferredPeers = fromThesePeers
		return nil
	}
}

// DefaultOptions are the default options to be used when using the lightpush protocol
func DefaultOptions(host host.Host) []PeerExchangeOption {
	return []PeerExchangeOption{
		WithAutomaticPeerSelection(),
	}
}
