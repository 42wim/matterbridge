package pubsub

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	manet "github.com/multiformats/go-multiaddr/net"
)

var (
	DefaultPeerGaterRetainStats     = 6 * time.Hour
	DefaultPeerGaterQuiet           = time.Minute
	DefaultPeerGaterDuplicateWeight = 0.125
	DefaultPeerGaterIgnoreWeight    = 1.0
	DefaultPeerGaterRejectWeight    = 16.0
	DefaultPeerGaterThreshold       = 0.33
	DefaultPeerGaterGlobalDecay     = ScoreParameterDecay(2 * time.Minute)
	DefaultPeerGaterSourceDecay     = ScoreParameterDecay(time.Hour)
)

// PeerGaterParams groups together parameters that control the operation of the peer gater
type PeerGaterParams struct {
	// when the ratio of throttled/validated messages exceeds this threshold, the gater turns on
	Threshold float64
	// (linear) decay parameter for gater counters
	GlobalDecay float64 // global counter decay
	SourceDecay float64 // per IP counter decay
	// decay interval
	DecayInterval time.Duration
	// counter zeroing threshold
	DecayToZero float64
	// how long to retain stats
	RetainStats time.Duration
	// quiet interval before turning off the gater; if there are no validation throttle events
	// for this interval, the gater turns off
	Quiet time.Duration
	// weight of duplicate message deliveries
	DuplicateWeight float64
	// weight of ignored messages
	IgnoreWeight float64
	// weight of rejected messages
	RejectWeight float64

	// priority topic delivery weights
	TopicDeliveryWeights map[string]float64
}

func (p *PeerGaterParams) validate() error {
	if p.Threshold <= 0 {
		return fmt.Errorf("invalid Threshold; must be > 0")
	}
	if p.GlobalDecay <= 0 || p.GlobalDecay >= 1 {
		return fmt.Errorf("invalid GlobalDecay; must be between 0 and 1")
	}
	if p.SourceDecay <= 0 || p.SourceDecay >= 1 {
		return fmt.Errorf("invalid SourceDecay; must be between 0 and 1")
	}
	if p.DecayInterval < time.Second {
		return fmt.Errorf("invalid DecayInterval; must be at least 1s")
	}
	if p.DecayToZero <= 0 || p.DecayToZero >= 1 {
		return fmt.Errorf("invalid DecayToZero; must be between 0 and 1")
	}
	// no need to check stats retention; a value of 0 means we don't retain stats
	if p.Quiet < time.Second {
		return fmt.Errorf("invalud Quiet interval; must be at least 1s")
	}
	if p.DuplicateWeight <= 0 {
		return fmt.Errorf("invalid DuplicateWeight; must be > 0")
	}
	if p.IgnoreWeight < 1 {
		return fmt.Errorf("invalid IgnoreWeight; must be >= 1")
	}
	if p.RejectWeight < 1 {
		return fmt.Errorf("invalud RejectWeight; must be >= 1")
	}

	return nil
}

// WithTopicDeliveryWeights is a fluid setter for the priority topic delivery weights
func (p *PeerGaterParams) WithTopicDeliveryWeights(w map[string]float64) *PeerGaterParams {
	p.TopicDeliveryWeights = w
	return p
}

// NewPeerGaterParams creates a new PeerGaterParams struct, using the specified threshold and decay
// parameters and default values for all other parameters.
func NewPeerGaterParams(threshold, globalDecay, sourceDecay float64) *PeerGaterParams {
	return &PeerGaterParams{
		Threshold:       threshold,
		GlobalDecay:     globalDecay,
		SourceDecay:     sourceDecay,
		DecayToZero:     DefaultDecayToZero,
		DecayInterval:   DefaultDecayInterval,
		RetainStats:     DefaultPeerGaterRetainStats,
		Quiet:           DefaultPeerGaterQuiet,
		DuplicateWeight: DefaultPeerGaterDuplicateWeight,
		IgnoreWeight:    DefaultPeerGaterIgnoreWeight,
		RejectWeight:    DefaultPeerGaterRejectWeight,
	}
}

// DefaultPeerGaterParams creates a new PeerGaterParams struct using default values
func DefaultPeerGaterParams() *PeerGaterParams {
	return NewPeerGaterParams(DefaultPeerGaterThreshold, DefaultPeerGaterGlobalDecay, DefaultPeerGaterSourceDecay)
}

// the gater object.
type peerGater struct {
	sync.Mutex

	host host.Host

	// gater parameters
	params *PeerGaterParams

	// counters
	validate, throttle float64

	// time of last validation throttle
	lastThrottle time.Time

	// stats per peer.ID -- multiple peer IDs may share the same stats object if they are
	// colocated in the same IP
	peerStats map[peer.ID]*peerGaterStats
	// stats per IP
	ipStats map[string]*peerGaterStats

	// for unit tests
	getIP func(peer.ID) string
}

type peerGaterStats struct {
	// number of connected peer IDs mapped to this stat object
	connected int
	// stats expiration time -- only valid if connected = 0
	expire time.Time

	// counters
	deliver, duplicate, ignore, reject float64
}

// WithPeerGater is a gossipsub router option that enables reactive validation queue
// management.
// The Gater is activated if the ratio of throttled/validated messages exceeds the specified
// threshold.
// Once active, the Gater probabilistically throttles peers _before_ they enter the validation
// queue, performing Random Early Drop.
// The throttle decision is randomized, with the probability of allowing messages to enter the
// validation queue controlled by the statistical observations of the performance of all peers
// in the IP address of the gated peer.
// The Gater deactivates if there is no validation throttlinc occurring for the specified quiet
// interval.
func WithPeerGater(params *PeerGaterParams) Option {
	return func(ps *PubSub) error {
		gs, ok := ps.rt.(*GossipSubRouter)
		if !ok {
			return fmt.Errorf("pubsub router is not gossipsub")
		}

		err := params.validate()
		if err != nil {
			return err
		}

		gs.gate = newPeerGater(ps.ctx, ps.host, params)

		// hook the tracer
		if ps.tracer != nil {
			ps.tracer.raw = append(ps.tracer.raw, gs.gate)
		} else {
			ps.tracer = &pubsubTracer{
				raw:   []RawTracer{gs.gate},
				pid:   ps.host.ID(),
				idGen: ps.idGen,
			}
		}

		return nil
	}
}

func newPeerGater(ctx context.Context, host host.Host, params *PeerGaterParams) *peerGater {
	pg := &peerGater{
		params:    params,
		peerStats: make(map[peer.ID]*peerGaterStats),
		ipStats:   make(map[string]*peerGaterStats),
		host:      host,
	}
	go pg.background(ctx)
	return pg
}

func (pg *peerGater) background(ctx context.Context) {
	tick := time.NewTicker(pg.params.DecayInterval)

	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			pg.decayStats()
		case <-ctx.Done():
			return
		}
	}
}

func (pg *peerGater) decayStats() {
	pg.Lock()
	defer pg.Unlock()

	pg.validate *= pg.params.GlobalDecay
	if pg.validate < pg.params.DecayToZero {
		pg.validate = 0
	}

	pg.throttle *= pg.params.GlobalDecay
	if pg.throttle < pg.params.DecayToZero {
		pg.throttle = 0
	}

	now := time.Now()
	for ip, st := range pg.ipStats {
		if st.connected > 0 {
			st.deliver *= pg.params.SourceDecay
			if st.deliver < pg.params.DecayToZero {
				st.deliver = 0
			}

			st.duplicate *= pg.params.SourceDecay
			if st.duplicate < pg.params.DecayToZero {
				st.duplicate = 0
			}

			st.ignore *= pg.params.SourceDecay
			if st.ignore < pg.params.DecayToZero {
				st.ignore = 0
			}

			st.reject *= pg.params.SourceDecay
			if st.reject < pg.params.DecayToZero {
				st.reject = 0
			}
		} else if st.expire.Before(now) {
			delete(pg.ipStats, ip)
		}
	}
}

func (pg *peerGater) getPeerStats(p peer.ID) *peerGaterStats {
	st, ok := pg.peerStats[p]
	if !ok {
		st = pg.getIPStats(p)
		pg.peerStats[p] = st
	}
	return st
}

func (pg *peerGater) getIPStats(p peer.ID) *peerGaterStats {
	ip := pg.getPeerIP(p)
	st, ok := pg.ipStats[ip]
	if !ok {
		st = &peerGaterStats{}
		pg.ipStats[ip] = st
	}
	return st
}

func (pg *peerGater) getPeerIP(p peer.ID) string {
	if pg.getIP != nil {
		return pg.getIP(p)
	}

	connToIP := func(c network.Conn) string {
		remote := c.RemoteMultiaddr()
		ip, err := manet.ToIP(remote)
		if err != nil {
			log.Warnf("error determining IP for remote peer in %s: %s", remote, err)
			return "<unknown>"
		}
		return ip.String()
	}

	conns := pg.host.Network().ConnsToPeer(p)
	switch len(conns) {
	case 0:
		return "<unknown>"
	case 1:
		return connToIP(conns[0])
	default:
		// we have multiple connections -- order by number of streams and use the one with the
		// most streams; it's a nightmare to track multiple IPs per peer, so pick the best one.
		streams := make(map[string]int)
		for _, c := range conns {
			if c.Stat().Transient {
				// ignore transient
				continue
			}
			streams[c.ID()] = len(c.GetStreams())
		}
		sort.Slice(conns, func(i, j int) bool {
			return streams[conns[i].ID()] > streams[conns[j].ID()]
		})
		return connToIP(conns[0])
	}
}

// router interface
func (pg *peerGater) AcceptFrom(p peer.ID) AcceptStatus {
	if pg == nil {
		return AcceptAll
	}

	pg.Lock()
	defer pg.Unlock()

	// check the quiet period; if the validation queue has not throttled for more than the Quiet
	// interval, we turn off the circuit breaker and accept.
	if time.Since(pg.lastThrottle) > pg.params.Quiet {
		return AcceptAll
	}

	// no throttle events -- or they have decayed; accept.
	if pg.throttle == 0 {
		return AcceptAll
	}

	// check the throttle/validate ration; if it is below threshold we accept.
	if pg.validate != 0 && pg.throttle/pg.validate < pg.params.Threshold {
		return AcceptAll
	}

	st := pg.getPeerStats(p)

	// compute the goodput of the peer; the denominator is the weighted mix of message counters
	total := st.deliver + pg.params.DuplicateWeight*st.duplicate + pg.params.IgnoreWeight*st.ignore + pg.params.RejectWeight*st.reject
	if total == 0 {
		return AcceptAll
	}

	// we make a randomized decision based on the goodput of the peer.
	// the probabiity is biased by adding 1 to the delivery counter so that we don't unconditionally
	// throttle in the first negative event; it also ensures that a peer always has a chance of being
	// accepted; this is not a sinkhole/blacklist.
	threshold := (1 + st.deliver) / (1 + total)
	if rand.Float64() < threshold {
		return AcceptAll
	}

	log.Debugf("throttling peer %s with threshold %f", p, threshold)
	return AcceptControl
}

// -- RawTracer interface methods
var _ RawTracer = (*peerGater)(nil)

// tracer interface
func (pg *peerGater) AddPeer(p peer.ID, proto protocol.ID) {
	pg.Lock()
	defer pg.Unlock()

	st := pg.getPeerStats(p)
	st.connected++
}

func (pg *peerGater) RemovePeer(p peer.ID) {
	pg.Lock()
	defer pg.Unlock()

	st := pg.getPeerStats(p)
	st.connected--
	st.expire = time.Now().Add(pg.params.RetainStats)

	delete(pg.peerStats, p)
}

func (pg *peerGater) Join(topic string)             {}
func (pg *peerGater) Leave(topic string)            {}
func (pg *peerGater) Graft(p peer.ID, topic string) {}
func (pg *peerGater) Prune(p peer.ID, topic string) {}

func (pg *peerGater) ValidateMessage(msg *Message) {
	pg.Lock()
	defer pg.Unlock()

	pg.validate++
}

func (pg *peerGater) DeliverMessage(msg *Message) {
	pg.Lock()
	defer pg.Unlock()

	st := pg.getPeerStats(msg.ReceivedFrom)

	topic := msg.GetTopic()
	weight := pg.params.TopicDeliveryWeights[topic]

	if weight == 0 {
		weight = 1
	}

	st.deliver += weight
}

func (pg *peerGater) RejectMessage(msg *Message, reason string) {
	pg.Lock()
	defer pg.Unlock()

	switch reason {
	case RejectValidationQueueFull:
		fallthrough
	case RejectValidationThrottled:
		pg.lastThrottle = time.Now()
		pg.throttle++

	case RejectValidationIgnored:
		st := pg.getPeerStats(msg.ReceivedFrom)
		st.ignore++

	default:
		st := pg.getPeerStats(msg.ReceivedFrom)
		st.reject++
	}
}

func (pg *peerGater) DuplicateMessage(msg *Message) {
	pg.Lock()
	defer pg.Unlock()

	st := pg.getPeerStats(msg.ReceivedFrom)
	st.duplicate++
}

func (pg *peerGater) ThrottlePeer(p peer.ID) {}

func (pg *peerGater) RecvRPC(rpc *RPC) {}

func (pg *peerGater) SendRPC(rpc *RPC, p peer.ID) {}

func (pg *peerGater) DropRPC(rpc *RPC, p peer.ID) {}

func (pg *peerGater) UndeliverableMessage(msg *Message) {}
