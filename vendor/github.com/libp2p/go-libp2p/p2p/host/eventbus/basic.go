package eventbus

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/event"
)

// /////////////////////
// BUS

// basicBus is a type-based event delivery system
type basicBus struct {
	lk            sync.RWMutex
	nodes         map[reflect.Type]*node
	wildcard      *wildcardNode
	metricsTracer MetricsTracer
}

var _ event.Bus = (*basicBus)(nil)

type emitter struct {
	n             *node
	w             *wildcardNode
	typ           reflect.Type
	closed        atomic.Bool
	dropper       func(reflect.Type)
	metricsTracer MetricsTracer
}

func (e *emitter) Emit(evt interface{}) error {
	if e.closed.Load() {
		return fmt.Errorf("emitter is closed")
	}

	e.n.emit(evt)
	e.w.emit(evt)

	if e.metricsTracer != nil {
		e.metricsTracer.EventEmitted(e.typ)
	}
	return nil
}

func (e *emitter) Close() error {
	if !e.closed.CompareAndSwap(false, true) {
		return fmt.Errorf("closed an emitter more than once")
	}
	if e.n.nEmitters.Add(-1) == 0 {
		e.dropper(e.typ)
	}
	return nil
}

func NewBus(opts ...Option) event.Bus {
	bus := &basicBus{
		nodes:    map[reflect.Type]*node{},
		wildcard: &wildcardNode{},
	}
	for _, opt := range opts {
		opt(bus)
	}
	return bus
}

func (b *basicBus) withNode(typ reflect.Type, cb func(*node), async func(*node)) {
	b.lk.Lock()

	n, ok := b.nodes[typ]
	if !ok {
		n = newNode(typ, b.metricsTracer)
		b.nodes[typ] = n
	}

	n.lk.Lock()
	b.lk.Unlock()

	cb(n)

	if async == nil {
		n.lk.Unlock()
	} else {
		go func() {
			defer n.lk.Unlock()
			async(n)
		}()
	}
}

func (b *basicBus) tryDropNode(typ reflect.Type) {
	b.lk.Lock()
	n, ok := b.nodes[typ]
	if !ok { // already dropped
		b.lk.Unlock()
		return
	}

	n.lk.Lock()
	if n.nEmitters.Load() > 0 || len(n.sinks) > 0 {
		n.lk.Unlock()
		b.lk.Unlock()
		return // still in use
	}
	n.lk.Unlock()

	delete(b.nodes, typ)
	b.lk.Unlock()
}

type wildcardSub struct {
	ch            chan interface{}
	w             *wildcardNode
	metricsTracer MetricsTracer
	name          string
}

func (w *wildcardSub) Out() <-chan interface{} {
	return w.ch
}

func (w *wildcardSub) Close() error {
	w.w.removeSink(w.ch)
	if w.metricsTracer != nil {
		w.metricsTracer.RemoveSubscriber(reflect.TypeOf(event.WildcardSubscription))
	}
	return nil
}

func (w *wildcardSub) Name() string {
	return w.name
}

type namedSink struct {
	name string
	ch   chan interface{}
}

type sub struct {
	ch            chan interface{}
	nodes         []*node
	dropper       func(reflect.Type)
	metricsTracer MetricsTracer
	name          string
}

func (s *sub) Name() string {
	return s.name
}

func (s *sub) Out() <-chan interface{} {
	return s.ch
}

func (s *sub) Close() error {
	go func() {
		// drain the event channel, will return when closed and drained.
		// this is necessary to unblock publishes to this channel.
		for range s.ch {
		}
	}()

	for _, n := range s.nodes {
		n.lk.Lock()

		for i := 0; i < len(n.sinks); i++ {
			if n.sinks[i].ch == s.ch {
				n.sinks[i], n.sinks[len(n.sinks)-1] = n.sinks[len(n.sinks)-1], nil
				n.sinks = n.sinks[:len(n.sinks)-1]

				if s.metricsTracer != nil {
					s.metricsTracer.RemoveSubscriber(n.typ)
				}
				break
			}
		}

		tryDrop := len(n.sinks) == 0 && n.nEmitters.Load() == 0

		n.lk.Unlock()

		if tryDrop {
			s.dropper(n.typ)
		}
	}
	close(s.ch)
	return nil
}

var _ event.Subscription = (*sub)(nil)

// Subscribe creates new subscription. Failing to drain the channel will cause
// publishers to get blocked. CancelFunc is guaranteed to return after last send
// to the channel
func (b *basicBus) Subscribe(evtTypes interface{}, opts ...event.SubscriptionOpt) (_ event.Subscription, err error) {
	settings := newSubSettings()
	for _, opt := range opts {
		if err := opt(&settings); err != nil {
			return nil, err
		}
	}

	if evtTypes == event.WildcardSubscription {
		out := &wildcardSub{
			ch:            make(chan interface{}, settings.buffer),
			w:             b.wildcard,
			metricsTracer: b.metricsTracer,
			name:          settings.name,
		}
		b.wildcard.addSink(&namedSink{ch: out.ch, name: out.name})
		return out, nil
	}

	types, ok := evtTypes.([]interface{})
	if !ok {
		types = []interface{}{evtTypes}
	}

	if len(types) > 1 {
		for _, t := range types {
			if t == event.WildcardSubscription {
				return nil, fmt.Errorf("wildcard subscriptions must be started separately")
			}
		}
	}

	out := &sub{
		ch:    make(chan interface{}, settings.buffer),
		nodes: make([]*node, len(types)),

		dropper:       b.tryDropNode,
		metricsTracer: b.metricsTracer,
		name:          settings.name,
	}

	for _, etyp := range types {
		if reflect.TypeOf(etyp).Kind() != reflect.Ptr {
			return nil, errors.New("subscribe called with non-pointer type")
		}
	}

	for i, etyp := range types {
		typ := reflect.TypeOf(etyp)

		b.withNode(typ.Elem(), func(n *node) {
			n.sinks = append(n.sinks, &namedSink{ch: out.ch, name: out.name})
			out.nodes[i] = n
			if b.metricsTracer != nil {
				b.metricsTracer.AddSubscriber(typ.Elem())
			}
		}, func(n *node) {
			if n.keepLast {
				l := n.last
				if l == nil {
					return
				}
				out.ch <- l
			}
		})
	}

	return out, nil
}

// Emitter creates new emitter
//
// eventType accepts typed nil pointers, and uses the type information to
// select output type
//
// Example:
// emit, err := eventbus.Emitter(new(EventT))
// defer emit.Close() // MUST call this after being done with the emitter
//
// emit(EventT{})
func (b *basicBus) Emitter(evtType interface{}, opts ...event.EmitterOpt) (e event.Emitter, err error) {
	if evtType == event.WildcardSubscription {
		return nil, fmt.Errorf("illegal emitter for wildcard subscription")
	}

	var settings emitterSettings
	for _, opt := range opts {
		if err := opt(&settings); err != nil {
			return nil, err
		}
	}

	typ := reflect.TypeOf(evtType)
	if typ.Kind() != reflect.Ptr {
		return nil, errors.New("emitter called with non-pointer type")
	}
	typ = typ.Elem()

	b.withNode(typ, func(n *node) {
		n.nEmitters.Add(1)
		n.keepLast = n.keepLast || settings.makeStateful
		e = &emitter{n: n, typ: typ, dropper: b.tryDropNode, w: b.wildcard, metricsTracer: b.metricsTracer}
	}, nil)
	return
}

// GetAllEventTypes returns all the event types that this bus has emitters
// or subscribers for.
func (b *basicBus) GetAllEventTypes() []reflect.Type {
	b.lk.RLock()
	defer b.lk.RUnlock()

	types := make([]reflect.Type, 0, len(b.nodes))
	for t := range b.nodes {
		types = append(types, t)
	}
	return types
}

// /////////////////////
// NODE

type wildcardNode struct {
	sync.RWMutex
	nSinks        atomic.Int32
	sinks         []*namedSink
	metricsTracer MetricsTracer
}

func (n *wildcardNode) addSink(sink *namedSink) {
	n.nSinks.Add(1) // ok to do outside the lock
	n.Lock()
	n.sinks = append(n.sinks, sink)
	n.Unlock()

	if n.metricsTracer != nil {
		n.metricsTracer.AddSubscriber(reflect.TypeOf(event.WildcardSubscription))
	}
}

func (n *wildcardNode) removeSink(ch chan interface{}) {
	n.nSinks.Add(-1) // ok to do outside the lock
	n.Lock()
	for i := 0; i < len(n.sinks); i++ {
		if n.sinks[i].ch == ch {
			n.sinks[i], n.sinks[len(n.sinks)-1] = n.sinks[len(n.sinks)-1], nil
			n.sinks = n.sinks[:len(n.sinks)-1]
			break
		}
	}
	n.Unlock()
}

func (n *wildcardNode) emit(evt interface{}) {
	if n.nSinks.Load() == 0 {
		return
	}

	n.RLock()
	for _, sink := range n.sinks {

		// Sending metrics before sending on channel allows us to
		// record channel full events before blocking
		sendSubscriberMetrics(n.metricsTracer, sink)

		sink.ch <- evt
	}
	n.RUnlock()
}

type node struct {
	// Note: make sure to NEVER lock basicBus.lk when this lock is held
	lk sync.Mutex

	typ reflect.Type

	// emitter ref count
	nEmitters atomic.Int32

	keepLast bool
	last     interface{}

	sinks         []*namedSink
	metricsTracer MetricsTracer
}

func newNode(typ reflect.Type, metricsTracer MetricsTracer) *node {
	return &node{
		typ:           typ,
		metricsTracer: metricsTracer,
	}
}

func (n *node) emit(evt interface{}) {
	typ := reflect.TypeOf(evt)
	if typ != n.typ {
		panic(fmt.Sprintf("Emit called with wrong type. expected: %s, got: %s", n.typ, typ))
	}

	n.lk.Lock()
	if n.keepLast {
		n.last = evt
	}

	for _, sink := range n.sinks {

		// Sending metrics before sending on channel allows us to
		// record channel full events before blocking
		sendSubscriberMetrics(n.metricsTracer, sink)
		sink.ch <- evt
	}
	n.lk.Unlock()
}

func sendSubscriberMetrics(metricsTracer MetricsTracer, sink *namedSink) {
	if metricsTracer != nil {
		metricsTracer.SubscriberQueueLength(sink.name, len(sink.ch)+1)
		metricsTracer.SubscriberQueueFull(sink.name, len(sink.ch)+1 >= cap(sink.ch))
		metricsTracer.SubscriberEventQueued(sink.name)
	}
}
