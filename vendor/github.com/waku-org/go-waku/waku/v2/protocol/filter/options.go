package filter

import (
	"errors"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"go.uber.org/zap"
)

func (old *FilterSubscribeParameters) Copy() *FilterSubscribeParameters {
	return &FilterSubscribeParameters{
		selectedPeer: old.selectedPeer,
		requestID:    old.requestID,
	}
}

type (
	FilterPingParameters struct {
		requestID []byte
	}
	FilterPingOption func(*FilterPingParameters)
)

func WithPingRequestId(requestId []byte) FilterPingOption {
	return func(params *FilterPingParameters) {
		params.requestID = requestId
	}
}

type (
	FilterSubscribeParameters struct {
		selectedPeer      peer.ID
		peerAddr          multiaddr.Multiaddr
		peerSelectionType peermanager.PeerSelection
		preferredPeers    peer.IDSlice
		requestID         []byte
		log               *zap.Logger

		// Subscribe-specific
		host host.Host
		pm   *peermanager.PeerManager

		// Unsubscribe-specific
		unsubscribeAll bool
		wg             *sync.WaitGroup
	}

	FilterParameters struct {
		Timeout        time.Duration
		MaxSubscribers int
		pm             *peermanager.PeerManager
	}

	Option func(*FilterParameters)

	FilterSubscribeOption func(*FilterSubscribeParameters) error
)

func WithTimeout(timeout time.Duration) Option {
	return func(params *FilterParameters) {
		params.Timeout = timeout
	}
}

// WithPeer is an option used to specify the peerID to request the message history.
// Note that this option is mutually exclusive to WithPeerAddr, only one of them can be used.
func WithPeer(p peer.ID) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.selectedPeer = p
		if params.peerAddr != nil {
			return errors.New("peerAddr and peerId options are mutually exclusive")
		}
		return nil
	}
}

// WithPeerAddr is an option used to specify a peerAddress.
// This new peer will be added to peerStore.
// Note that this option is mutually exclusive to WithPeerAddr, only one of them can be used.
func WithPeerAddr(pAddr multiaddr.Multiaddr) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.peerAddr = pAddr
		if params.selectedPeer != "" {
			return errors.New("peerAddr and peerId options are mutually exclusive")
		}
		return nil
	}
}

// WithAutomaticPeerSelection is an option used to randomly select a peer from the peer store.
// If a list of specific peers is passed, the peer will be chosen from that list assuming it
// supports the chosen protocol, otherwise it will chose a peer from the node peerstore
func WithAutomaticPeerSelection(fromThesePeers ...peer.ID) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.peerSelectionType = peermanager.Automatic
		params.preferredPeers = fromThesePeers
		return nil
	}
}

// WithFastestPeerSelection is an option used to select a peer from the peer store
// with the lowest ping If a list of specific peers is passed, the peer will be chosen
// from that list assuming it supports the chosen protocol, otherwise it will chose a
// peer from the node peerstore
func WithFastestPeerSelection(fromThesePeers ...peer.ID) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.peerSelectionType = peermanager.LowestRTT
		return nil
	}
}

// WithRequestID is an option to set a specific request ID to be used when
// creating/removing a filter subscription
func WithRequestID(requestID []byte) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.requestID = requestID
		return nil
	}
}

// WithAutomaticRequestID is an option to automatically generate a request ID
// when creating a filter subscription
func WithAutomaticRequestID() FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.requestID = protocol.GenerateRequestID()
		return nil
	}
}

func DefaultSubscriptionOptions() []FilterSubscribeOption {
	return []FilterSubscribeOption{
		WithAutomaticPeerSelection(),
		WithAutomaticRequestID(),
	}
}

func UnsubscribeAll() FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.unsubscribeAll = true
		return nil
	}
}

// WithWaitGroup allows specifying a waitgroup to wait until all
// unsubscribe requests are complete before the function is complete
func WithWaitGroup(wg *sync.WaitGroup) FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.wg = wg
		return nil
	}
}

// DontWait is used to fire and forget an unsubscription, and don't
// care about the results of it
func DontWait() FilterSubscribeOption {
	return func(params *FilterSubscribeParameters) error {
		params.wg = nil
		return nil
	}
}

func DefaultUnsubscribeOptions() []FilterSubscribeOption {
	return []FilterSubscribeOption{
		WithAutomaticRequestID(),
		WithWaitGroup(&sync.WaitGroup{}),
	}
}

func WithMaxSubscribers(maxSubscribers int) Option {
	return func(params *FilterParameters) {
		params.MaxSubscribers = maxSubscribers
	}
}

func WithPeerManager(pm *peermanager.PeerManager) Option {
	return func(params *FilterParameters) {
		params.pm = pm
	}
}

func DefaultOptions() []Option {
	return []Option{
		WithTimeout(24 * time.Hour),
		WithMaxSubscribers(DefaultMaxSubscriptions),
	}
}
