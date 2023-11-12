package mailservers

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/status-im/status-go/eth-node/types"
)

const (
	peerEventsBuffer    = 10 // sufficient buffer to avoid blocking a p2p feed.
	whisperEventsBuffer = 20 // sufficient buffer to avod blocking a eventSub envelopes feed.
)

// PeerAdderRemover is an interface for adding or removing peers.
type PeerAdderRemover interface {
	AddPeer(node *enode.Node)
	RemovePeer(node *enode.Node)
}

// PeerEventsSubscriber interface to subscribe for p2p.PeerEvent's.
type PeerEventsSubscriber interface {
	SubscribeEvents(chan *p2p.PeerEvent) event.Subscription
}

// EnvelopeEventSubscriber interface to subscribe for types.EnvelopeEvent's.
type EnvelopeEventSubscriber interface {
	SubscribeEnvelopeEvents(chan<- types.EnvelopeEvent) types.Subscription
}

type p2pServer interface {
	PeerAdderRemover
	PeerEventsSubscriber
}

// NewConnectionManager creates an instance of ConnectionManager.
func NewConnectionManager(server p2pServer, eventSub EnvelopeEventSubscriber, target, maxFailures int, timeout time.Duration) *ConnectionManager {
	return &ConnectionManager{
		server:           server,
		eventSub:         eventSub,
		connectedTarget:  target,
		maxFailures:      maxFailures,
		notifications:    make(chan []*enode.Node),
		timeoutWaitAdded: timeout,
	}
}

// ConnectionManager manages keeps target of peers connected.
type ConnectionManager struct {
	wg   sync.WaitGroup
	quit chan struct{}

	server   p2pServer
	eventSub EnvelopeEventSubscriber

	notifications    chan []*enode.Node
	connectedTarget  int
	timeoutWaitAdded time.Duration
	maxFailures      int
}

// Notify sends a non-blocking notification about new nodes.
func (ps *ConnectionManager) Notify(nodes []*enode.Node) {
	ps.wg.Add(1)
	go func() {
		select {
		case ps.notifications <- nodes:
		case <-ps.quit:
		}
		ps.wg.Done()
	}()

}

// Start subscribes to a p2p server and handles new peers and state updates for those peers.
func (ps *ConnectionManager) Start() {
	ps.quit = make(chan struct{})
	ps.wg.Add(1)
	go func() {
		state := newInternalState(ps.server, ps.connectedTarget, ps.timeoutWaitAdded)
		events := make(chan *p2p.PeerEvent, peerEventsBuffer)
		sub := ps.server.SubscribeEvents(events)
		whisperEvents := make(chan types.EnvelopeEvent, whisperEventsBuffer)
		whisperSub := ps.eventSub.SubscribeEnvelopeEvents(whisperEvents)
		requests := map[types.Hash]struct{}{}
		failuresPerServer := map[types.EnodeID]int{}

		defer sub.Unsubscribe()
		defer whisperSub.Unsubscribe()
		defer ps.wg.Done()
		for {
			select {
			case <-ps.quit:
				return
			case err := <-sub.Err():
				log.Error("retry after error subscribing to p2p events", "error", err)
				return
			case err := <-whisperSub.Err():
				log.Error("retry after error suscribing to eventSub events", "error", err)
				return
			case newNodes := <-ps.notifications:
				state.processReplacement(newNodes, events)
			case ev := <-events:
				processPeerEvent(state, ev)
			case ev := <-whisperEvents:
				// TODO treat failed requests the same way as expired
				switch ev.Event {
				case types.EventMailServerRequestSent:
					requests[ev.Hash] = struct{}{}
				case types.EventMailServerRequestCompleted:
					// reset failures count on first success
					failuresPerServer[ev.Peer] = 0
					delete(requests, ev.Hash)
				case types.EventMailServerRequestExpired:
					_, exist := requests[ev.Hash]
					if !exist {
						continue
					}
					failuresPerServer[ev.Peer]++
					log.Debug("request to a mail server expired, disconnect a peer", "address", ev.Peer)
					if failuresPerServer[ev.Peer] >= ps.maxFailures {
						state.nodeDisconnected(ev.Peer)
					}
				}
			}
		}
	}()
}

// Stop gracefully closes all background goroutines and waits until they finish.
func (ps *ConnectionManager) Stop() {
	if ps.quit == nil {
		return
	}
	select {
	case <-ps.quit:
		return
	default:
	}
	close(ps.quit)
	ps.wg.Wait()
	ps.quit = nil
}

func (state *internalState) processReplacement(newNodes []*enode.Node, events <-chan *p2p.PeerEvent) {
	replacement := map[types.EnodeID]*enode.Node{}
	for _, n := range newNodes {
		replacement[types.EnodeID(n.ID())] = n
	}
	state.replaceNodes(replacement)
	if state.ReachedTarget() {
		log.Debug("already connected with required target", "target", state.target)
		return
	}
	if state.timeout != 0 {
		log.Debug("waiting defined timeout to establish connections",
			"timeout", state.timeout, "target", state.target)
		timer := time.NewTimer(state.timeout)
		waitForConnections(state, timer.C, events)
		timer.Stop()
	}
}

func newInternalState(srv PeerAdderRemover, target int, timeout time.Duration) *internalState {
	return &internalState{
		options:      options{target: target, timeout: timeout},
		srv:          srv,
		connected:    map[types.EnodeID]struct{}{},
		currentNodes: map[types.EnodeID]*enode.Node{},
	}
}

type options struct {
	target  int
	timeout time.Duration
}

type internalState struct {
	options
	srv PeerAdderRemover

	connected    map[types.EnodeID]struct{}
	currentNodes map[types.EnodeID]*enode.Node
}

func (state *internalState) ReachedTarget() bool {
	return len(state.connected) >= state.target
}

func (state *internalState) replaceNodes(new map[types.EnodeID]*enode.Node) {
	for nid, n := range state.currentNodes {
		if _, exist := new[nid]; !exist {
			delete(state.connected, nid)
			state.srv.RemovePeer(n)
		}
	}
	if !state.ReachedTarget() {
		for _, n := range new {
			state.srv.AddPeer(n)
		}
	}
	state.currentNodes = new
}

func (state *internalState) nodeAdded(peer types.EnodeID) {
	n, exist := state.currentNodes[peer]
	if !exist {
		return
	}
	if state.ReachedTarget() {
		state.srv.RemovePeer(n)
	} else {
		state.connected[types.EnodeID(n.ID())] = struct{}{}
	}
}

func (state *internalState) nodeDisconnected(peer types.EnodeID) {
	n, exist := state.currentNodes[peer] // unrelated event
	if !exist {
		return
	}
	_, exist = state.connected[peer] // check if already disconnected
	if !exist {
		return
	}
	if len(state.currentNodes) == 1 { // keep node connected if we don't have another choice
		return
	}
	state.srv.RemovePeer(n) // remove peer permanently, otherwise p2p.Server will try to reconnect
	delete(state.connected, peer)
	if !state.ReachedTarget() { // try to connect with any other selected (but not connected) node
		for nid, n := range state.currentNodes {
			_, exist := state.connected[nid]
			if exist || peer == nid {
				continue
			}
			state.srv.AddPeer(n)
		}
	}
}

func processPeerEvent(state *internalState, ev *p2p.PeerEvent) {
	switch ev.Type {
	case p2p.PeerEventTypeAdd:
		log.Debug("connected to a mailserver", "address", ev.Peer)
		state.nodeAdded(types.EnodeID(ev.Peer))
	case p2p.PeerEventTypeDrop:
		log.Debug("mailserver disconnected", "address", ev.Peer)
		state.nodeDisconnected(types.EnodeID(ev.Peer))
	}
}

func waitForConnections(state *internalState, timeout <-chan time.Time, events <-chan *p2p.PeerEvent) {
	for {
		select {
		case ev := <-events:
			processPeerEvent(state, ev)
			if state.ReachedTarget() {
				return
			}
		case <-timeout:
			return
		}
	}

}
