package pubsub

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var (
	// GossipSubConnTagBumpMessageDelivery is the amount to add to the connection manager
	// tag that tracks message deliveries. Each time a peer is the first to deliver a
	// message within a topic, we "bump" a tag by this amount, up to a maximum
	// of GossipSubConnTagMessageDeliveryCap.
	// Note that the delivery tags decay over time, decreasing by GossipSubConnTagDecayAmount
	// at every GossipSubConnTagDecayInterval.
	GossipSubConnTagBumpMessageDelivery = 1

	// GossipSubConnTagDecayInterval is the decay interval for decaying connection manager tags.
	GossipSubConnTagDecayInterval = 10 * time.Minute

	// GossipSubConnTagDecayAmount is subtracted from decaying tag values at each decay interval.
	GossipSubConnTagDecayAmount = 1

	// GossipSubConnTagMessageDeliveryCap is the maximum value for the connection manager tags that
	// track message deliveries.
	GossipSubConnTagMessageDeliveryCap = 15
)

// tagTracer is an internal tracer that applies connection manager tags to peer
// connections based on their behavior.
//
// We tag a peer's connections for the following reasons:
//   - Directly connected peers are tagged with GossipSubConnTagValueDirectPeer (default 1000).
//   - Mesh peers are tagged with a value of GossipSubConnTagValueMeshPeer (default 20).
//     If a peer is in multiple topic meshes, they'll be tagged for each.
//   - For each message that we receive, we bump a delivery tag for peer that delivered the message
//     first.
//     The delivery tags have a maximum value, GossipSubConnTagMessageDeliveryCap, and they decay at
//     a rate of GossipSubConnTagDecayAmount / GossipSubConnTagDecayInterval.
type tagTracer struct {
	sync.RWMutex

	cmgr     connmgr.ConnManager
	idGen    *msgIDGenerator
	decayer  connmgr.Decayer
	decaying map[string]connmgr.DecayingTag
	direct   map[peer.ID]struct{}

	// a map of message ids to the set of peers who delivered the message after the first delivery,
	// but before the message was finished validating
	nearFirst map[string]map[peer.ID]struct{}
}

func newTagTracer(cmgr connmgr.ConnManager) *tagTracer {
	decayer, ok := connmgr.SupportsDecay(cmgr)
	if !ok {
		log.Debugf("connection manager does not support decaying tags, delivery tags will not be applied")
	}
	return &tagTracer{
		cmgr:      cmgr,
		idGen:     newMsgIdGenerator(),
		decayer:   decayer,
		decaying:  make(map[string]connmgr.DecayingTag),
		nearFirst: make(map[string]map[peer.ID]struct{}),
	}
}

func (t *tagTracer) Start(gs *GossipSubRouter) {
	if t == nil {
		return
	}

	t.idGen = gs.p.idGen
	t.direct = gs.direct
}

func (t *tagTracer) tagPeerIfDirect(p peer.ID) {
	if t.direct == nil {
		return
	}

	// tag peer if it is a direct peer
	_, direct := t.direct[p]
	if direct {
		t.cmgr.Protect(p, "pubsub:<direct>")
	}
}

func (t *tagTracer) tagMeshPeer(p peer.ID, topic string) {
	tag := topicTag(topic)
	t.cmgr.Protect(p, tag)
}

func (t *tagTracer) untagMeshPeer(p peer.ID, topic string) {
	tag := topicTag(topic)
	t.cmgr.Unprotect(p, tag)
}

func topicTag(topic string) string {
	return fmt.Sprintf("pubsub:%s", topic)
}

func (t *tagTracer) addDeliveryTag(topic string) {
	if t.decayer == nil {
		return
	}

	name := fmt.Sprintf("pubsub-deliveries:%s", topic)
	t.Lock()
	defer t.Unlock()
	tag, err := t.decayer.RegisterDecayingTag(
		name,
		GossipSubConnTagDecayInterval,
		connmgr.DecayFixed(GossipSubConnTagDecayAmount),
		connmgr.BumpSumBounded(0, GossipSubConnTagMessageDeliveryCap))

	if err != nil {
		log.Warnf("unable to create decaying delivery tag: %s", err)
		return
	}
	t.decaying[topic] = tag
}

func (t *tagTracer) removeDeliveryTag(topic string) {
	t.Lock()
	defer t.Unlock()
	tag, ok := t.decaying[topic]
	if !ok {
		return
	}
	err := tag.Close()
	if err != nil {
		log.Warnf("error closing decaying connmgr tag: %s", err)
	}
	delete(t.decaying, topic)
}

func (t *tagTracer) bumpDeliveryTag(p peer.ID, topic string) error {
	t.RLock()
	defer t.RUnlock()

	tag, ok := t.decaying[topic]
	if !ok {
		return fmt.Errorf("no decaying tag registered for topic %s", topic)
	}
	return tag.Bump(p, GossipSubConnTagBumpMessageDelivery)
}

func (t *tagTracer) bumpTagsForMessage(p peer.ID, msg *Message) {
	topic := msg.GetTopic()
	err := t.bumpDeliveryTag(p, topic)
	if err != nil {
		log.Warnf("error bumping delivery tag: %s", err)
	}
}

// nearFirstPeers returns the peers who delivered the message while it was still validating
func (t *tagTracer) nearFirstPeers(msg *Message) []peer.ID {
	t.Lock()
	defer t.Unlock()
	peersMap, ok := t.nearFirst[t.idGen.ID(msg)]
	if !ok {
		return nil
	}
	peers := make([]peer.ID, 0, len(peersMap))
	for p := range peersMap {
		peers = append(peers, p)
	}
	return peers
}

// -- RawTracer interface methods
var _ RawTracer = (*tagTracer)(nil)

func (t *tagTracer) AddPeer(p peer.ID, proto protocol.ID) {
	t.tagPeerIfDirect(p)
}

func (t *tagTracer) Join(topic string) {
	t.addDeliveryTag(topic)
}

func (t *tagTracer) DeliverMessage(msg *Message) {
	nearFirst := t.nearFirstPeers(msg)

	t.bumpTagsForMessage(msg.ReceivedFrom, msg)
	for _, p := range nearFirst {
		t.bumpTagsForMessage(p, msg)
	}

	// delete the delivery state for this message
	t.Lock()
	delete(t.nearFirst, t.idGen.ID(msg))
	t.Unlock()
}

func (t *tagTracer) Leave(topic string) {
	t.removeDeliveryTag(topic)
}

func (t *tagTracer) Graft(p peer.ID, topic string) {
	t.tagMeshPeer(p, topic)
}

func (t *tagTracer) Prune(p peer.ID, topic string) {
	t.untagMeshPeer(p, topic)
}

func (t *tagTracer) ValidateMessage(msg *Message) {
	t.Lock()
	defer t.Unlock()

	// create map to start tracking the peers who deliver while we're validating
	id := t.idGen.ID(msg)
	if _, exists := t.nearFirst[id]; exists {
		return
	}
	t.nearFirst[id] = make(map[peer.ID]struct{})
}

func (t *tagTracer) DuplicateMessage(msg *Message) {
	t.Lock()
	defer t.Unlock()

	id := t.idGen.ID(msg)
	peers, ok := t.nearFirst[id]
	if !ok {
		return
	}
	peers[msg.ReceivedFrom] = struct{}{}
}

func (t *tagTracer) RejectMessage(msg *Message, reason string) {
	t.Lock()
	defer t.Unlock()

	// We want to delete the near-first delivery tracking for messages that have passed through
	// the validation pipeline. Other rejection reasons (missing signature, etc) skip the validation
	// queue, so we don't want to remove the state in case the message is still validating.
	switch reason {
	case RejectValidationThrottled:
		fallthrough
	case RejectValidationIgnored:
		fallthrough
	case RejectValidationFailed:
		delete(t.nearFirst, t.idGen.ID(msg))
	}
}

func (t *tagTracer) RemovePeer(peer.ID)                {}
func (t *tagTracer) ThrottlePeer(p peer.ID)            {}
func (t *tagTracer) RecvRPC(rpc *RPC)                  {}
func (t *tagTracer) SendRPC(rpc *RPC, p peer.ID)       {}
func (t *tagTracer) DropRPC(rpc *RPC, p peer.ID)       {}
func (t *tagTracer) UndeliverableMessage(msg *Message) {}
