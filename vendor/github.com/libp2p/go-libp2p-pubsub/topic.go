package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/crypto"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ErrTopicClosed is returned if a Topic is utilized after it has been closed
var ErrTopicClosed = errors.New("this Topic is closed, try opening a new one")

// ErrNilSignKey is returned if a nil private key was provided
var ErrNilSignKey = errors.New("nil sign key")

// ErrEmptyPeerID is returned if an empty peer ID was provided
var ErrEmptyPeerID = errors.New("empty peer ID")

// Topic is the handle for a pubsub topic
type Topic struct {
	p     *PubSub
	topic string

	evtHandlerMux sync.RWMutex
	evtHandlers   map[*TopicEventHandler]struct{}

	mux    sync.RWMutex
	closed bool
}

// String returns the topic associated with t
func (t *Topic) String() string {
	return t.topic
}

// SetScoreParams sets the topic score parameters if the pubsub router supports peer
// scoring
func (t *Topic) SetScoreParams(p *TopicScoreParams) error {
	err := p.validate()
	if err != nil {
		return fmt.Errorf("invalid topic score parameters: %w", err)
	}

	t.mux.Lock()
	defer t.mux.Unlock()

	if t.closed {
		return ErrTopicClosed
	}

	result := make(chan error, 1)
	update := func() {
		gs, ok := t.p.rt.(*GossipSubRouter)
		if !ok {
			result <- fmt.Errorf("pubsub router is not gossipsub")
			return
		}

		if gs.score == nil {
			result <- fmt.Errorf("peer scoring is not enabled in router")
			return
		}

		err := gs.score.SetTopicScoreParams(t.topic, p)
		result <- err
	}

	select {
	case t.p.eval <- update:
		err = <-result
		return err

	case <-t.p.ctx.Done():
		return t.p.ctx.Err()
	}
}

// EventHandler creates a handle for topic specific events
// Multiple event handlers may be created and will operate independently of each other
func (t *Topic) EventHandler(opts ...TopicEventHandlerOpt) (*TopicEventHandler, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.closed {
		return nil, ErrTopicClosed
	}

	h := &TopicEventHandler{
		topic: t,
		err:   nil,

		evtLog:   make(map[peer.ID]EventType),
		evtLogCh: make(chan struct{}, 1),
	}

	for _, opt := range opts {
		err := opt(h)
		if err != nil {
			return nil, err
		}
	}

	done := make(chan struct{}, 1)

	select {
	case t.p.eval <- func() {
		tmap := t.p.topics[t.topic]
		for p := range tmap {
			h.evtLog[p] = PeerJoin
		}

		t.evtHandlerMux.Lock()
		t.evtHandlers[h] = struct{}{}
		t.evtHandlerMux.Unlock()
		done <- struct{}{}
	}:
	case <-t.p.ctx.Done():
		return nil, t.p.ctx.Err()
	}

	<-done

	return h, nil
}

func (t *Topic) sendNotification(evt PeerEvent) {
	t.evtHandlerMux.RLock()
	defer t.evtHandlerMux.RUnlock()

	for h := range t.evtHandlers {
		h.sendNotification(evt)
	}
}

// Subscribe returns a new Subscription for the topic.
// Note that subscription is not an instantaneous operation. It may take some time
// before the subscription is processed by the pubsub main loop and propagated to our peers.
func (t *Topic) Subscribe(opts ...SubOpt) (*Subscription, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.closed {
		return nil, ErrTopicClosed
	}

	sub := &Subscription{
		topic: t.topic,
		ctx:   t.p.ctx,
	}

	for _, opt := range opts {
		err := opt(sub)
		if err != nil {
			return nil, err
		}
	}

	if sub.ch == nil {
		// apply the default size
		sub.ch = make(chan *Message, 32)
	}

	out := make(chan *Subscription, 1)

	t.p.disc.Discover(sub.topic)

	select {
	case t.p.addSub <- &addSubReq{
		sub:  sub,
		resp: out,
	}:
	case <-t.p.ctx.Done():
		return nil, t.p.ctx.Err()
	}

	return <-out, nil
}

// Relay enables message relaying for the topic and returns a reference
// cancel function. Subsequent calls increase the reference counter.
// To completely disable the relay, all references must be cancelled.
func (t *Topic) Relay() (RelayCancelFunc, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.closed {
		return nil, ErrTopicClosed
	}

	out := make(chan RelayCancelFunc, 1)

	t.p.disc.Discover(t.topic)

	select {
	case t.p.addRelay <- &addRelayReq{
		topic: t.topic,
		resp:  out,
	}:
	case <-t.p.ctx.Done():
		return nil, t.p.ctx.Err()
	}

	return <-out, nil
}

// RouterReady is a function that decides if a router is ready to publish
type RouterReady func(rt PubSubRouter, topic string) (bool, error)

// ProvideKey is a function that provides a private key and its associated peer ID when publishing a new message
type ProvideKey func() (crypto.PrivKey, peer.ID)

type PublishOptions struct {
	ready     RouterReady
	customKey ProvideKey
	local     bool
}

type PubOpt func(pub *PublishOptions) error

// Publish publishes data to topic.
func (t *Topic) Publish(ctx context.Context, data []byte, opts ...PubOpt) error {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.closed {
		return ErrTopicClosed
	}

	pid := t.p.signID
	key := t.p.signKey

	pub := &PublishOptions{}
	for _, opt := range opts {
		err := opt(pub)
		if err != nil {
			return err
		}
	}

	if pub.customKey != nil && !pub.local {
		key, pid = pub.customKey()
		if key == nil {
			return ErrNilSignKey
		}
		if len(pid) == 0 {
			return ErrEmptyPeerID
		}
	}

	m := &pb.Message{
		Data:  data,
		Topic: &t.topic,
		From:  nil,
		Seqno: nil,
	}
	if pid != "" {
		m.From = []byte(pid)
		m.Seqno = t.p.nextSeqno()
	}
	if key != nil {
		m.From = []byte(pid)
		err := signMessage(pid, key, m)
		if err != nil {
			return err
		}
	}

	if pub.ready != nil {
		if t.p.disc.discovery != nil {
			t.p.disc.Bootstrap(ctx, t.topic, pub.ready)
		} else {
			// TODO: we could likely do better than polling every 200ms.
			// For example, block this goroutine on a channel,
			// and check again whenever events tell us that the number of
			// peers has increased.
			var ticker *time.Ticker
		readyLoop:
			for {
				// Check if ready for publishing.
				// Similar to what disc.Bootstrap does.
				res := make(chan bool, 1)
				select {
				case t.p.eval <- func() {
					done, _ := pub.ready(t.p.rt, t.topic)
					res <- done
				}:
					if <-res {
						break readyLoop
					}
				case <-t.p.ctx.Done():
					return t.p.ctx.Err()
				case <-ctx.Done():
					return ctx.Err()
				}
				if ticker == nil {
					ticker = time.NewTicker(200 * time.Millisecond)
					defer ticker.Stop()
				}

				select {
				case <-ticker.C:
				case <-ctx.Done():
					return fmt.Errorf("router is not ready: %w", ctx.Err())
				}
			}
		}
	}

	return t.p.val.PushLocal(&Message{m, "", t.p.host.ID(), nil, pub.local})
}

// WithReadiness returns a publishing option for only publishing when the router is ready.
// This option is not useful unless PubSub is also using WithDiscovery
func WithReadiness(ready RouterReady) PubOpt {
	return func(pub *PublishOptions) error {
		pub.ready = ready
		return nil
	}
}

// WithLocalPublication returns a publishing option to notify in-process subscribers only.
// It prevents message publication to mesh peers.
// Useful in edge cases where the msg needs to be only delivered to the in-process subscribers,
// e.g. not to spam the network with outdated msgs.
// Should not be used specifically for in-process pubsubing.
func WithLocalPublication(local bool) PubOpt {
	return func(pub *PublishOptions) error {
		pub.local = local
		return nil
	}
}

// WithSecretKeyAndPeerId returns a publishing option for providing a custom private key and its corresponding peer ID
// This option is useful when we want to send messages from "virtual", never-connectable peers in the network
func WithSecretKeyAndPeerId(key crypto.PrivKey, pid peer.ID) PubOpt {
	return func(pub *PublishOptions) error {
		pub.customKey = func() (crypto.PrivKey, peer.ID) {
			return key, pid
		}

		return nil
	}
}

// Close closes down the topic. Will return an error unless there are no active event handlers or subscriptions.
// Does not error if the topic is already closed.
func (t *Topic) Close() error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.closed {
		return nil
	}

	req := &rmTopicReq{t, make(chan error, 1)}

	select {
	case t.p.rmTopic <- req:
	case <-t.p.ctx.Done():
		return t.p.ctx.Err()
	}

	err := <-req.resp

	if err == nil {
		t.closed = true
	}

	return err
}

// ListPeers returns a list of peers we are connected to in the given topic.
func (t *Topic) ListPeers() []peer.ID {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.closed {
		return []peer.ID{}
	}

	return t.p.ListPeers(t.topic)
}

type EventType int

const (
	PeerJoin EventType = iota
	PeerLeave
)

// TopicEventHandler is used to manage topic specific events. No Subscription is required to receive events.
type TopicEventHandler struct {
	topic *Topic
	err   error

	evtLogMx sync.Mutex
	evtLog   map[peer.ID]EventType
	evtLogCh chan struct{}
}

type TopicEventHandlerOpt func(t *TopicEventHandler) error

type PeerEvent struct {
	Type EventType
	Peer peer.ID
}

// Cancel closes the topic event handler
func (t *TopicEventHandler) Cancel() {
	topic := t.topic
	t.err = fmt.Errorf("topic event handler cancelled by calling handler.Cancel()")

	topic.evtHandlerMux.Lock()
	delete(topic.evtHandlers, t)
	t.topic.evtHandlerMux.Unlock()
}

func (t *TopicEventHandler) sendNotification(evt PeerEvent) {
	t.evtLogMx.Lock()
	t.addToEventLog(evt)
	t.evtLogMx.Unlock()
}

// addToEventLog assumes a lock has been taken to protect the event log
func (t *TopicEventHandler) addToEventLog(evt PeerEvent) {
	e, ok := t.evtLog[evt.Peer]
	if !ok {
		t.evtLog[evt.Peer] = evt.Type
		// send signal that an event has been added to the event log
		select {
		case t.evtLogCh <- struct{}{}:
		default:
		}
	} else if e != evt.Type {
		delete(t.evtLog, evt.Peer)
	}
}

// pullFromEventLog assumes a lock has been taken to protect the event log
func (t *TopicEventHandler) pullFromEventLog() (PeerEvent, bool) {
	for k, v := range t.evtLog {
		evt := PeerEvent{Peer: k, Type: v}
		delete(t.evtLog, k)
		return evt, true
	}
	return PeerEvent{}, false
}

// NextPeerEvent returns the next event regarding subscribed peers
// Guarantees: Peer Join and Peer Leave events for a given peer will fire in order.
// Unless a peer both Joins and Leaves before NextPeerEvent emits either event
// all events will eventually be received from NextPeerEvent.
func (t *TopicEventHandler) NextPeerEvent(ctx context.Context) (PeerEvent, error) {
	for {
		t.evtLogMx.Lock()
		evt, ok := t.pullFromEventLog()
		if ok {
			// make sure an event log signal is available if there are events in the event log
			if len(t.evtLog) > 0 {
				select {
				case t.evtLogCh <- struct{}{}:
				default:
				}
			}
			t.evtLogMx.Unlock()
			return evt, nil
		}
		t.evtLogMx.Unlock()

		select {
		case <-t.evtLogCh:
			continue
		case <-ctx.Done():
			return PeerEvent{}, ctx.Err()
		}
	}
}
