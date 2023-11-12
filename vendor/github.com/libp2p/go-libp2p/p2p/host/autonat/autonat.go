package autonat

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"

	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

var log = logging.Logger("autonat")

const maxConfidence = 3

// AmbientAutoNAT is the implementation of ambient NAT autodiscovery
type AmbientAutoNAT struct {
	host host.Host

	*config

	ctx               context.Context
	ctxCancel         context.CancelFunc // is closed when Close is called
	backgroundRunning chan struct{}      // is closed when the background go routine exits

	inboundConn   chan network.Conn
	dialResponses chan error
	// status is an autoNATResult reflecting current status.
	status atomic.Pointer[network.Reachability]
	// Reflects the confidence on of the NATStatus being private, as a single
	// dialback may fail for reasons unrelated to NAT.
	// If it is <3, then multiple autoNAT peers may be contacted for dialback
	// If only a single autoNAT peer is known, then the confidence increases
	// for each failure until it reaches 3.
	confidence   int
	lastInbound  time.Time
	lastProbeTry time.Time
	lastProbe    time.Time
	recentProbes map[peer.ID]time.Time

	service *autoNATService

	emitReachabilityChanged event.Emitter
	subscriber              event.Subscription
}

// StaticAutoNAT is a simple AutoNAT implementation when a single NAT status is desired.
type StaticAutoNAT struct {
	host         host.Host
	reachability network.Reachability
	service      *autoNATService
}

// New creates a new NAT autodiscovery system attached to a host
func New(h host.Host, options ...Option) (AutoNAT, error) {
	var err error
	conf := new(config)
	conf.host = h
	conf.dialPolicy.host = h

	if err = defaults(conf); err != nil {
		return nil, err
	}
	if conf.addressFunc == nil {
		conf.addressFunc = h.Addrs
	}

	for _, o := range options {
		if err = o(conf); err != nil {
			return nil, err
		}
	}
	emitReachabilityChanged, _ := h.EventBus().Emitter(new(event.EvtLocalReachabilityChanged), eventbus.Stateful)

	var service *autoNATService
	if (!conf.forceReachability || conf.reachability == network.ReachabilityPublic) && conf.dialer != nil {
		service, err = newAutoNATService(conf)
		if err != nil {
			return nil, err
		}
		service.Enable()
	}

	if conf.forceReachability {
		emitReachabilityChanged.Emit(event.EvtLocalReachabilityChanged{Reachability: conf.reachability})

		return &StaticAutoNAT{
			host:         h,
			reachability: conf.reachability,
			service:      service,
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	as := &AmbientAutoNAT{
		ctx:               ctx,
		ctxCancel:         cancel,
		backgroundRunning: make(chan struct{}),
		host:              h,
		config:            conf,
		inboundConn:       make(chan network.Conn, 5),
		dialResponses:     make(chan error, 1),

		emitReachabilityChanged: emitReachabilityChanged,
		service:                 service,
		recentProbes:            make(map[peer.ID]time.Time),
	}
	reachability := network.ReachabilityUnknown
	as.status.Store(&reachability)

	subscriber, err := as.host.EventBus().Subscribe(
		[]any{new(event.EvtLocalAddressesUpdated), new(event.EvtPeerIdentificationCompleted)},
		eventbus.Name("autonat"),
	)
	if err != nil {
		return nil, err
	}
	as.subscriber = subscriber

	h.Network().Notify(as)
	go as.background()

	return as, nil
}

// Status returns the AutoNAT observed reachability status.
func (as *AmbientAutoNAT) Status() network.Reachability {
	s := as.status.Load()
	return *s
}

func (as *AmbientAutoNAT) emitStatus() {
	status := *as.status.Load()
	as.emitReachabilityChanged.Emit(event.EvtLocalReachabilityChanged{Reachability: status})
	if as.metricsTracer != nil {
		as.metricsTracer.ReachabilityStatus(status)
	}
}

func ipInList(candidate ma.Multiaddr, list []ma.Multiaddr) bool {
	candidateIP, _ := manet.ToIP(candidate)
	for _, i := range list {
		if ip, err := manet.ToIP(i); err == nil && ip.Equal(candidateIP) {
			return true
		}
	}
	return false
}

func (as *AmbientAutoNAT) background() {
	defer close(as.backgroundRunning)
	// wait a bit for the node to come online and establish some connections
	// before starting autodetection
	delay := as.config.bootDelay

	subChan := as.subscriber.Out()
	defer as.subscriber.Close()
	defer as.emitReachabilityChanged.Close()

	timer := time.NewTimer(delay)
	defer timer.Stop()
	timerRunning := true
	retryProbe := false
	for {
		select {
		// new inbound connection.
		case conn := <-as.inboundConn:
			localAddrs := as.host.Addrs()
			if manet.IsPublicAddr(conn.RemoteMultiaddr()) &&
				!ipInList(conn.RemoteMultiaddr(), localAddrs) {
				as.lastInbound = time.Now()
			}

		case e := <-subChan:
			switch e := e.(type) {
			case event.EvtLocalAddressesUpdated:
				// On local address update, reduce confidence from maximum so that we schedule
				// the next probe sooner
				if as.confidence == maxConfidence {
					as.confidence--
				}
			case event.EvtPeerIdentificationCompleted:
				if s, err := as.host.Peerstore().SupportsProtocols(e.Peer, AutoNATProto); err == nil && len(s) > 0 {
					currentStatus := *as.status.Load()
					if currentStatus == network.ReachabilityUnknown {
						as.tryProbe(e.Peer)
					}
				}
			default:
				log.Errorf("unknown event type: %T", e)
			}

		// probe finished.
		case err, ok := <-as.dialResponses:
			if !ok {
				return
			}
			if IsDialRefused(err) {
				retryProbe = true
			} else {
				as.handleDialResponse(err)
			}
		case <-timer.C:
			peer := as.getPeerToProbe()
			as.tryProbe(peer)
			timerRunning = false
			retryProbe = false
		case <-as.ctx.Done():
			return
		}

		// Drain the timer channel if it hasn't fired in preparation for Resetting it.
		if timerRunning && !timer.Stop() {
			<-timer.C
		}
		timer.Reset(as.scheduleProbe(retryProbe))
		timerRunning = true
	}
}

func (as *AmbientAutoNAT) cleanupRecentProbes() {
	fixedNow := time.Now()
	for k, v := range as.recentProbes {
		if fixedNow.Sub(v) > as.throttlePeerPeriod {
			delete(as.recentProbes, k)
		}
	}
}

// scheduleProbe calculates when the next probe should be scheduled for.
func (as *AmbientAutoNAT) scheduleProbe(retryProbe bool) time.Duration {
	// Our baseline is a probe every 'AutoNATRefreshInterval'
	// This is modulated by:
	// * if we are in an unknown state, have low confidence, or we want to retry because a probe was refused that
	//   should drop to 'AutoNATRetryInterval'
	// * recent inbound connections (implying continued connectivity) should decrease the retry when public
	// * recent inbound connections when not public mean we should try more actively to see if we're public.
	fixedNow := time.Now()
	currentStatus := *as.status.Load()

	nextProbe := fixedNow
	// Don't look for peers in the peer store more than once per second.
	if !as.lastProbeTry.IsZero() {
		backoff := as.lastProbeTry.Add(time.Second)
		if backoff.After(nextProbe) {
			nextProbe = backoff
		}
	}
	if !as.lastProbe.IsZero() {
		untilNext := as.config.refreshInterval
		if retryProbe {
			untilNext = as.config.retryInterval
		} else if currentStatus == network.ReachabilityUnknown {
			untilNext = as.config.retryInterval
		} else if as.confidence < maxConfidence {
			untilNext = as.config.retryInterval
		} else if currentStatus == network.ReachabilityPublic && as.lastInbound.After(as.lastProbe) {
			untilNext *= 2
		} else if currentStatus != network.ReachabilityPublic && as.lastInbound.After(as.lastProbe) {
			untilNext /= 5
		}

		if as.lastProbe.Add(untilNext).After(nextProbe) {
			nextProbe = as.lastProbe.Add(untilNext)
		}
	}
	if as.metricsTracer != nil {
		as.metricsTracer.NextProbeTime(nextProbe)
	}
	return nextProbe.Sub(fixedNow)
}

// handleDialResponse updates the current status based on dial response.
func (as *AmbientAutoNAT) handleDialResponse(dialErr error) {
	var observation network.Reachability
	switch {
	case dialErr == nil:
		observation = network.ReachabilityPublic
	case IsDialError(dialErr):
		observation = network.ReachabilityPrivate
	default:
		observation = network.ReachabilityUnknown
	}

	as.recordObservation(observation)
}

// recordObservation updates NAT status and confidence
func (as *AmbientAutoNAT) recordObservation(observation network.Reachability) {

	currentStatus := *as.status.Load()

	if observation == network.ReachabilityPublic {
		changed := false
		if currentStatus != network.ReachabilityPublic {
			// Aggressively switch to public from other states ignoring confidence
			log.Debugf("NAT status is public")

			// we are flipping our NATStatus, so confidence drops to 0
			as.confidence = 0
			if as.service != nil {
				as.service.Enable()
			}
			changed = true
		} else if as.confidence < maxConfidence {
			as.confidence++
		}
		as.status.Store(&observation)
		if changed {
			as.emitStatus()
		}
	} else if observation == network.ReachabilityPrivate {
		if currentStatus != network.ReachabilityPrivate {
			if as.confidence > 0 {
				as.confidence--
			} else {
				log.Debugf("NAT status is private")

				// we are flipping our NATStatus, so confidence drops to 0
				as.confidence = 0
				as.status.Store(&observation)
				if as.service != nil {
					as.service.Disable()
				}
				as.emitStatus()
			}
		} else if as.confidence < maxConfidence {
			as.confidence++
			as.status.Store(&observation)
		}
	} else if as.confidence > 0 {
		// don't just flip to unknown, reduce confidence first
		as.confidence--
	} else {
		log.Debugf("NAT status is unknown")
		as.status.Store(&observation)
		if currentStatus != network.ReachabilityUnknown {
			if as.service != nil {
				as.service.Enable()
			}
			as.emitStatus()
		}
	}
	if as.metricsTracer != nil {
		as.metricsTracer.ReachabilityStatusConfidence(as.confidence)
	}
}

func (as *AmbientAutoNAT) tryProbe(p peer.ID) bool {
	as.lastProbeTry = time.Now()
	if p.Validate() != nil {
		return false
	}

	if lastTime, ok := as.recentProbes[p]; ok {
		if time.Since(lastTime) < as.throttlePeerPeriod {
			return false
		}
	}
	as.cleanupRecentProbes()

	info := as.host.Peerstore().PeerInfo(p)

	if !as.config.dialPolicy.skipPeer(info.Addrs) {
		as.recentProbes[p] = time.Now()
		as.lastProbe = time.Now()
		go as.probe(&info)
		return true
	}
	return false
}

func (as *AmbientAutoNAT) probe(pi *peer.AddrInfo) {
	cli := NewAutoNATClient(as.host, as.config.addressFunc, as.metricsTracer)
	ctx, cancel := context.WithTimeout(as.ctx, as.config.requestTimeout)
	defer cancel()

	err := cli.DialBack(ctx, pi.ID)
	log.Debugf("Dialback through peer %s completed: err: %s", pi.ID, err)

	select {
	case as.dialResponses <- err:
	case <-as.ctx.Done():
		return
	}
}

func (as *AmbientAutoNAT) getPeerToProbe() peer.ID {
	peers := as.host.Network().Peers()
	if len(peers) == 0 {
		return ""
	}

	candidates := make([]peer.ID, 0, len(peers))

	for _, p := range peers {
		info := as.host.Peerstore().PeerInfo(p)
		// Exclude peers which don't support the autonat protocol.
		if proto, err := as.host.Peerstore().SupportsProtocols(p, AutoNATProto); len(proto) == 0 || err != nil {
			continue
		}

		// Exclude peers in backoff.
		if lastTime, ok := as.recentProbes[p]; ok {
			if time.Since(lastTime) < as.throttlePeerPeriod {
				continue
			}
		}

		if as.config.dialPolicy.skipPeer(info.Addrs) {
			continue
		}
		candidates = append(candidates, p)
	}

	if len(candidates) == 0 {
		return ""
	}

	return candidates[rand.Intn(len(candidates))]
}

func (as *AmbientAutoNAT) Close() error {
	as.ctxCancel()
	if as.service != nil {
		as.service.Disable()
	}
	<-as.backgroundRunning
	return nil
}

// Status returns the AutoNAT observed reachability status.
func (s *StaticAutoNAT) Status() network.Reachability {
	return s.reachability
}

func (s *StaticAutoNAT) Close() error {
	if s.service != nil {
		s.service.Disable()
	}
	return nil
}
