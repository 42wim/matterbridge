package autorelay

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	basic "github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	circuitv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	circuitv2_proto "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/proto"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

const protoIDv2 = circuitv2_proto.ProtoIDv2Hop

// Terminology:
// Candidate: Once we connect to a node and it supports relay protocol,
// we call it a candidate, and consider using it as a relay.
// Relay: Out of the list of candidates, we select a relay to connect to.
// Currently, we just randomly select a candidate, but we can employ more sophisticated
// selection strategies here (e.g. by facotring in the RTT).

const (
	rsvpRefreshInterval = time.Minute
	rsvpExpirationSlack = 2 * time.Minute

	autorelayTag = "autorelay"
)

type candidate struct {
	added           time.Time
	supportsRelayV2 bool
	ai              peer.AddrInfo
}

// relayFinder is a Host that uses relays for connectivity when a NAT is detected.
type relayFinder struct {
	bootTime time.Time
	host     *basic.BasicHost

	conf *config

	refCount sync.WaitGroup

	ctxCancel   context.CancelFunc
	ctxCancelMx sync.Mutex

	peerSource PeerSource

	candidateFound             chan struct{} // receives every time we find a new relay candidate
	candidateMx                sync.Mutex
	candidates                 map[peer.ID]*candidate
	backoff                    map[peer.ID]time.Time
	maybeConnectToRelayTrigger chan struct{} // cap: 1
	// Any time _something_ hapens that might cause us to need new candidates.
	// This could be
	// * the disconnection of a relay
	// * the failed attempt to obtain a reservation with a current candidate
	// * a candidate is deleted due to its age
	maybeRequestNewCandidates chan struct{} // cap: 1.

	relayUpdated chan struct{}

	relayMx sync.Mutex
	relays  map[peer.ID]*circuitv2.Reservation

	cachedAddrs       []ma.Multiaddr
	cachedAddrsExpiry time.Time

	// A channel that triggers a run of `runScheduledWork`.
	triggerRunScheduledWork chan struct{}
	metricsTracer           MetricsTracer
}

var errAlreadyRunning = errors.New("relayFinder already running")

func newRelayFinder(host *basic.BasicHost, peerSource PeerSource, conf *config) *relayFinder {
	if peerSource == nil {
		panic("Can not create a new relayFinder. Need a Peer Source fn or a list of static relays. Refer to the documentation around `libp2p.EnableAutoRelay`")
	}

	return &relayFinder{
		bootTime:                   conf.clock.Now(),
		host:                       host,
		conf:                       conf,
		peerSource:                 peerSource,
		candidates:                 make(map[peer.ID]*candidate),
		backoff:                    make(map[peer.ID]time.Time),
		candidateFound:             make(chan struct{}, 1),
		maybeConnectToRelayTrigger: make(chan struct{}, 1),
		maybeRequestNewCandidates:  make(chan struct{}, 1),
		triggerRunScheduledWork:    make(chan struct{}, 1),
		relays:                     make(map[peer.ID]*circuitv2.Reservation),
		relayUpdated:               make(chan struct{}, 1),
		metricsTracer:              &wrappedMetricsTracer{conf.metricsTracer},
	}
}

type scheduledWorkTimes struct {
	leastFrequentInterval       time.Duration
	nextRefresh                 time.Time
	nextBackoff                 time.Time
	nextOldCandidateCheck       time.Time
	nextAllowedCallToPeerSource time.Time
}

func (rf *relayFinder) background(ctx context.Context) {
	peerSourceRateLimiter := make(chan struct{}, 1)
	rf.refCount.Add(1)
	go func() {
		defer rf.refCount.Done()
		rf.findNodes(ctx, peerSourceRateLimiter)
	}()

	rf.refCount.Add(1)
	go func() {
		defer rf.refCount.Done()
		rf.handleNewCandidates(ctx)
	}()

	subConnectedness, err := rf.host.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged), eventbus.Name("autorelay (relay finder)"))
	if err != nil {
		log.Error("failed to subscribe to the EvtPeerConnectednessChanged")
		return
	}
	defer subConnectedness.Close()

	now := rf.conf.clock.Now()
	bootDelayTimer := rf.conf.clock.InstantTimer(now.Add(rf.conf.bootDelay))
	defer bootDelayTimer.Stop()

	// This is the least frequent event. It's our fallback timer if we don't have any other work to do.
	leastFrequentInterval := rf.conf.minInterval
	// Check if leastFrequentInterval is 0 to avoid busy looping
	if rf.conf.backoff > leastFrequentInterval || leastFrequentInterval == 0 {
		leastFrequentInterval = rf.conf.backoff
	}
	if rf.conf.maxCandidateAge > leastFrequentInterval || leastFrequentInterval == 0 {
		leastFrequentInterval = rf.conf.maxCandidateAge
	}
	if rsvpRefreshInterval > leastFrequentInterval || leastFrequentInterval == 0 {
		leastFrequentInterval = rsvpRefreshInterval
	}

	scheduledWork := &scheduledWorkTimes{
		leastFrequentInterval:       leastFrequentInterval,
		nextRefresh:                 now.Add(rsvpRefreshInterval),
		nextBackoff:                 now.Add(rf.conf.backoff),
		nextOldCandidateCheck:       now.Add(rf.conf.maxCandidateAge),
		nextAllowedCallToPeerSource: now.Add(-time.Second), // allow immediately
	}

	workTimer := rf.conf.clock.InstantTimer(rf.runScheduledWork(ctx, now, scheduledWork, peerSourceRateLimiter))
	defer workTimer.Stop()

	for {
		select {
		case ev, ok := <-subConnectedness.Out():
			if !ok {
				return
			}
			evt := ev.(event.EvtPeerConnectednessChanged)
			if evt.Connectedness != network.NotConnected {
				continue
			}
			push := false

			rf.relayMx.Lock()
			if rf.usingRelay(evt.Peer) { // we were disconnected from a relay
				log.Debugw("disconnected from relay", "id", evt.Peer)
				delete(rf.relays, evt.Peer)
				rf.notifyMaybeConnectToRelay()
				rf.notifyMaybeNeedNewCandidates()
				push = true
			}
			rf.relayMx.Unlock()

			if push {
				rf.clearCachedAddrsAndSignalAddressChange()
				rf.metricsTracer.ReservationEnded(1)
			}
		case <-rf.candidateFound:
			rf.notifyMaybeConnectToRelay()
		case <-bootDelayTimer.Ch():
			rf.notifyMaybeConnectToRelay()
		case <-rf.relayUpdated:
			rf.clearCachedAddrsAndSignalAddressChange()
		case now := <-workTimer.Ch():
			// Note: `now` is not guaranteed to be the current time. It's the time
			// that the timer was fired. This is okay because we'll schedule
			// future work at a specific time.
			nextTime := rf.runScheduledWork(ctx, now, scheduledWork, peerSourceRateLimiter)
			workTimer.Reset(nextTime)
		case <-rf.triggerRunScheduledWork:
			// Ignore the next time because we aren't scheduling any future work here
			_ = rf.runScheduledWork(ctx, rf.conf.clock.Now(), scheduledWork, peerSourceRateLimiter)
		case <-ctx.Done():
			return
		}
	}
}

func (rf *relayFinder) clearCachedAddrsAndSignalAddressChange() {
	rf.relayMx.Lock()
	rf.cachedAddrs = nil
	rf.relayMx.Unlock()
	rf.host.SignalAddressChange()

	rf.metricsTracer.RelayAddressUpdated()
}

func (rf *relayFinder) runScheduledWork(ctx context.Context, now time.Time, scheduledWork *scheduledWorkTimes, peerSourceRateLimiter chan<- struct{}) time.Time {
	nextTime := now.Add(scheduledWork.leastFrequentInterval)

	if now.After(scheduledWork.nextRefresh) {
		scheduledWork.nextRefresh = now.Add(rsvpRefreshInterval)
		if rf.refreshReservations(ctx, now) {
			rf.clearCachedAddrsAndSignalAddressChange()
		}
	}

	if now.After(scheduledWork.nextBackoff) {
		scheduledWork.nextBackoff = rf.clearBackoff(now)
	}

	if now.After(scheduledWork.nextOldCandidateCheck) {
		scheduledWork.nextOldCandidateCheck = rf.clearOldCandidates(now)
	}

	if now.After(scheduledWork.nextAllowedCallToPeerSource) {
		select {
		case peerSourceRateLimiter <- struct{}{}:
			scheduledWork.nextAllowedCallToPeerSource = now.Add(rf.conf.minInterval)
			if scheduledWork.nextAllowedCallToPeerSource.Before(nextTime) {
				nextTime = scheduledWork.nextAllowedCallToPeerSource
			}
		default:
		}
	} else {
		// We still need to schedule this work if it's sooner than nextTime
		if scheduledWork.nextAllowedCallToPeerSource.Before(nextTime) {
			nextTime = scheduledWork.nextAllowedCallToPeerSource
		}
	}

	// Find the next time we need to run scheduled work.
	if scheduledWork.nextRefresh.Before(nextTime) {
		nextTime = scheduledWork.nextRefresh
	}
	if scheduledWork.nextBackoff.Before(nextTime) {
		nextTime = scheduledWork.nextBackoff
	}
	if scheduledWork.nextOldCandidateCheck.Before(nextTime) {
		nextTime = scheduledWork.nextOldCandidateCheck
	}
	if nextTime == now {
		// Only happens in CI with a mock clock
		nextTime = nextTime.Add(1) // avoids an infinite loop
	}

	rf.metricsTracer.ScheduledWorkUpdated(scheduledWork)

	return nextTime
}

// clearOldCandidates clears old candidates from the map. Returns the next time
// to run this function.
func (rf *relayFinder) clearOldCandidates(now time.Time) time.Time {
	// If we don't have any candidates, we should run this again in rf.conf.maxCandidateAge.
	nextTime := now.Add(rf.conf.maxCandidateAge)

	var deleted bool
	rf.candidateMx.Lock()
	defer rf.candidateMx.Unlock()
	for id, cand := range rf.candidates {
		expiry := cand.added.Add(rf.conf.maxCandidateAge)
		if expiry.After(now) {
			if expiry.Before(nextTime) {
				nextTime = expiry
			}
		} else {
			log.Debugw("deleting candidate due to age", "id", id)
			deleted = true
			rf.removeCandidate(id)
		}
	}
	if deleted {
		rf.notifyMaybeNeedNewCandidates()
	}

	return nextTime
}

// clearBackoff clears old backoff entries from the map. Returns the next time
// to run this function.
func (rf *relayFinder) clearBackoff(now time.Time) time.Time {
	nextTime := now.Add(rf.conf.backoff)

	rf.candidateMx.Lock()
	defer rf.candidateMx.Unlock()
	for id, t := range rf.backoff {
		expiry := t.Add(rf.conf.backoff)
		if expiry.After(now) {
			if expiry.Before(nextTime) {
				nextTime = expiry
			}
		} else {
			log.Debugw("removing backoff for node", "id", id)
			delete(rf.backoff, id)
		}
	}

	return nextTime
}

// findNodes accepts nodes from the channel and tests if they support relaying.
// It is run on both public and private nodes.
// It garbage collects old entries, so that nodes doesn't overflow.
// This makes sure that as soon as we need to find relay candidates, we have them available.
// peerSourceRateLimiter is used to limit how often we call the peer source.
func (rf *relayFinder) findNodes(ctx context.Context, peerSourceRateLimiter <-chan struct{}) {
	var peerChan <-chan peer.AddrInfo
	var wg sync.WaitGroup
	for {
		rf.candidateMx.Lock()
		numCandidates := len(rf.candidates)
		rf.candidateMx.Unlock()

		if peerChan == nil && numCandidates < rf.conf.minCandidates {
			rf.metricsTracer.CandidateLoopState(peerSourceRateLimited)

			select {
			case <-peerSourceRateLimiter:
				peerChan = rf.peerSource(ctx, rf.conf.maxCandidates)
				select {
				case rf.triggerRunScheduledWork <- struct{}{}:
				default:
				}
			case <-ctx.Done():
				return
			}
		}

		if peerChan == nil {
			rf.metricsTracer.CandidateLoopState(waitingForTrigger)
		} else {
			rf.metricsTracer.CandidateLoopState(waitingOnPeerChan)
		}

		select {
		case <-rf.maybeRequestNewCandidates:
			continue
		case pi, ok := <-peerChan:
			if !ok {
				wg.Wait()
				peerChan = nil
				continue
			}
			log.Debugw("found node", "id", pi.ID)
			rf.candidateMx.Lock()
			numCandidates := len(rf.candidates)
			backoffStart, isOnBackoff := rf.backoff[pi.ID]
			rf.candidateMx.Unlock()
			if isOnBackoff {
				log.Debugw("skipping node that we recently failed to obtain a reservation with", "id", pi.ID, "last attempt", rf.conf.clock.Since(backoffStart))
				continue
			}
			if numCandidates >= rf.conf.maxCandidates {
				log.Debugw("skipping node. Already have enough candidates", "id", pi.ID, "num", numCandidates, "max", rf.conf.maxCandidates)
				continue
			}
			rf.refCount.Add(1)
			wg.Add(1)
			go func() {
				defer rf.refCount.Done()
				defer wg.Done()
				if added := rf.handleNewNode(ctx, pi); added {
					rf.notifyNewCandidate()
				}
			}()
		case <-ctx.Done():
			rf.metricsTracer.CandidateLoopState(stopped)
			return
		}
	}
}

func (rf *relayFinder) notifyMaybeConnectToRelay() {
	select {
	case rf.maybeConnectToRelayTrigger <- struct{}{}:
	default:
	}
}

func (rf *relayFinder) notifyMaybeNeedNewCandidates() {
	select {
	case rf.maybeRequestNewCandidates <- struct{}{}:
	default:
	}
}

func (rf *relayFinder) notifyNewCandidate() {
	select {
	case rf.candidateFound <- struct{}{}:
	default:
	}
}

// handleNewNode tests if a peer supports circuit v2.
// This method is only run on private nodes.
// If a peer does, it is added to the candidates map.
// Note that just supporting the protocol doesn't guarantee that we can also obtain a reservation.
func (rf *relayFinder) handleNewNode(ctx context.Context, pi peer.AddrInfo) (added bool) {
	rf.relayMx.Lock()
	relayInUse := rf.usingRelay(pi.ID)
	rf.relayMx.Unlock()
	if relayInUse {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	supportsV2, err := rf.tryNode(ctx, pi)
	if err != nil {
		log.Debugf("node %s not accepted as a candidate: %s", pi.ID, err)
		if err == errProtocolNotSupported {
			rf.metricsTracer.CandidateChecked(false)
		}
		return false
	}
	rf.metricsTracer.CandidateChecked(true)

	rf.candidateMx.Lock()
	if len(rf.candidates) > rf.conf.maxCandidates {
		rf.candidateMx.Unlock()
		return false
	}
	log.Debugw("node supports relay protocol", "peer", pi.ID, "supports circuit v2", supportsV2)
	rf.addCandidate(&candidate{
		added:           rf.conf.clock.Now(),
		ai:              pi,
		supportsRelayV2: supportsV2,
	})
	rf.candidateMx.Unlock()
	return true
}

var errProtocolNotSupported = errors.New("doesn't speak circuit v2")

// tryNode checks if a peer actually supports either circuit v2.
// It does not modify any internal state.
func (rf *relayFinder) tryNode(ctx context.Context, pi peer.AddrInfo) (supportsRelayV2 bool, err error) {
	if err := rf.host.Connect(ctx, pi); err != nil {
		return false, fmt.Errorf("error connecting to relay %s: %w", pi.ID, err)
	}

	conns := rf.host.Network().ConnsToPeer(pi.ID)
	for _, conn := range conns {
		if isRelayAddr(conn.RemoteMultiaddr()) {
			return false, errors.New("not a public node")
		}
	}

	// wait for identify to complete in at least one conn so that we can check the supported protocols
	ready := make(chan struct{}, 1)
	for _, conn := range conns {
		go func(conn network.Conn) {
			select {
			case <-rf.host.IDService().IdentifyWait(conn):
				select {
				case ready <- struct{}{}:
				default:
				}
			case <-ctx.Done():
			}
		}(conn)
	}

	select {
	case <-ready:
	case <-ctx.Done():
		return false, ctx.Err()
	}

	protos, err := rf.host.Peerstore().SupportsProtocols(pi.ID, protoIDv2)
	if err != nil {
		return false, fmt.Errorf("error checking relay protocol support for peer %s: %w", pi.ID, err)
	}
	if len(protos) == 0 {
		return false, errProtocolNotSupported
	}
	return true, nil
}

// When a new node that could be a relay is found, we receive a notification on the maybeConnectToRelayTrigger chan.
// This function makes sure that we only run one instance of maybeConnectToRelay at once, and buffers
// exactly one more trigger event to run maybeConnectToRelay.
func (rf *relayFinder) handleNewCandidates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-rf.maybeConnectToRelayTrigger:
			rf.maybeConnectToRelay(ctx)
		}
	}
}

func (rf *relayFinder) maybeConnectToRelay(ctx context.Context) {
	rf.relayMx.Lock()
	numRelays := len(rf.relays)
	rf.relayMx.Unlock()
	// We're already connected to our desired number of relays. Nothing to do here.
	if numRelays == rf.conf.desiredRelays {
		return
	}

	rf.candidateMx.Lock()
	if len(rf.relays) == 0 && len(rf.candidates) < rf.conf.minCandidates && rf.conf.clock.Since(rf.bootTime) < rf.conf.bootDelay {
		// During the startup phase, we don't want to connect to the first candidate that we find.
		// Instead, we wait until we've found at least minCandidates, and then select the best of those.
		// However, if that takes too long (longer than bootDelay), we still go ahead.
		rf.candidateMx.Unlock()
		return
	}
	if len(rf.candidates) == 0 {
		rf.candidateMx.Unlock()
		return
	}
	candidates := rf.selectCandidates()
	rf.candidateMx.Unlock()

	// We now iterate over the candidates, attempting (sequentially) to get reservations with them, until
	// we reach the desired number of relays.
	for _, cand := range candidates {
		id := cand.ai.ID
		rf.relayMx.Lock()
		usingRelay := rf.usingRelay(id)
		rf.relayMx.Unlock()
		if usingRelay {
			rf.candidateMx.Lock()
			rf.removeCandidate(id)
			rf.candidateMx.Unlock()
			rf.notifyMaybeNeedNewCandidates()
			continue
		}
		rsvp, err := rf.connectToRelay(ctx, cand)
		if err != nil {
			log.Debugw("failed to connect to relay", "peer", id, "error", err)
			rf.notifyMaybeNeedNewCandidates()
			rf.metricsTracer.ReservationRequestFinished(false, err)
			continue
		}
		log.Debugw("adding new relay", "id", id)
		rf.relayMx.Lock()
		rf.relays[id] = rsvp
		numRelays := len(rf.relays)
		rf.relayMx.Unlock()
		rf.notifyMaybeNeedNewCandidates()

		rf.host.ConnManager().Protect(id, autorelayTag) // protect the connection

		select {
		case rf.relayUpdated <- struct{}{}:
		default:
		}

		rf.metricsTracer.ReservationRequestFinished(false, nil)

		if numRelays >= rf.conf.desiredRelays {
			break
		}
	}
}

func (rf *relayFinder) connectToRelay(ctx context.Context, cand *candidate) (*circuitv2.Reservation, error) {
	id := cand.ai.ID

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var rsvp *circuitv2.Reservation

	// make sure we're still connected.
	if rf.host.Network().Connectedness(id) != network.Connected {
		if err := rf.host.Connect(ctx, cand.ai); err != nil {
			rf.candidateMx.Lock()
			rf.removeCandidate(cand.ai.ID)
			rf.candidateMx.Unlock()
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	}

	rf.candidateMx.Lock()
	rf.backoff[id] = rf.conf.clock.Now()
	rf.candidateMx.Unlock()
	var err error
	if cand.supportsRelayV2 {
		rsvp, err = circuitv2.Reserve(ctx, rf.host, cand.ai)
		if err != nil {
			err = fmt.Errorf("failed to reserve slot: %w", err)
		}
	}
	rf.candidateMx.Lock()
	rf.removeCandidate(id)
	rf.candidateMx.Unlock()
	return rsvp, err
}

func (rf *relayFinder) refreshReservations(ctx context.Context, now time.Time) bool {
	rf.relayMx.Lock()

	// find reservations about to expire and refresh them in parallel
	g := new(errgroup.Group)
	for p, rsvp := range rf.relays {
		if now.Add(rsvpExpirationSlack).Before(rsvp.Expiration) {
			continue
		}

		p := p
		g.Go(func() error {
			err := rf.refreshRelayReservation(ctx, p)
			rf.metricsTracer.ReservationRequestFinished(true, err)

			return err
		})
	}
	rf.relayMx.Unlock()

	err := g.Wait()
	return err != nil
}

func (rf *relayFinder) refreshRelayReservation(ctx context.Context, p peer.ID) error {
	rsvp, err := circuitv2.Reserve(ctx, rf.host, peer.AddrInfo{ID: p})

	rf.relayMx.Lock()
	if err != nil {
		log.Debugw("failed to refresh relay slot reservation", "relay", p, "error", err)
		_, exists := rf.relays[p]
		delete(rf.relays, p)
		// unprotect the connection
		rf.host.ConnManager().Unprotect(p, autorelayTag)
		rf.relayMx.Unlock()
		if exists {
			rf.metricsTracer.ReservationEnded(1)
		}
		return err
	}

	log.Debugw("refreshed relay slot reservation", "relay", p)
	rf.relays[p] = rsvp
	rf.relayMx.Unlock()
	return nil
}

// usingRelay returns if we're currently using the given relay.
func (rf *relayFinder) usingRelay(p peer.ID) bool {
	_, ok := rf.relays[p]
	return ok
}

// addCandidates adds a candidate to the candidates set. Assumes caller holds candidateMx mutex
func (rf *relayFinder) addCandidate(cand *candidate) {
	_, exists := rf.candidates[cand.ai.ID]
	rf.candidates[cand.ai.ID] = cand
	if !exists {
		rf.metricsTracer.CandidateAdded(1)
	}
}

func (rf *relayFinder) removeCandidate(id peer.ID) {
	_, exists := rf.candidates[id]
	if exists {
		delete(rf.candidates, id)
		rf.metricsTracer.CandidateRemoved(1)
	}
}

// selectCandidates returns an ordered slice of relay candidates.
// Callers should attempt to obtain reservations with the candidates in this order.
func (rf *relayFinder) selectCandidates() []*candidate {
	now := rf.conf.clock.Now()
	candidates := make([]*candidate, 0, len(rf.candidates))
	for _, cand := range rf.candidates {
		if cand.added.Add(rf.conf.maxCandidateAge).After(now) {
			candidates = append(candidates, cand)
		}
	}

	// TODO: better relay selection strategy; this just selects random relays,
	// but we should probably use ping latency as the selection metric
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	return candidates
}

// This function is computes the NATed relay addrs when our status is private:
//   - The public addrs are removed from the address set.
//   - The non-public addrs are included verbatim so that peers behind the same NAT/firewall
//     can still dial us directly.
//   - On top of those, we add the relay-specific addrs for the relays to which we are
//     connected. For each non-private relay addr, we encapsulate the p2p-circuit addr
//     through which we can be dialed.
func (rf *relayFinder) relayAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	rf.relayMx.Lock()
	defer rf.relayMx.Unlock()

	if rf.cachedAddrs != nil && rf.conf.clock.Now().Before(rf.cachedAddrsExpiry) {
		return rf.cachedAddrs
	}

	raddrs := make([]ma.Multiaddr, 0, 4*len(rf.relays)+4)

	// only keep private addrs from the original addr set
	for _, addr := range addrs {
		if manet.IsPrivateAddr(addr) {
			raddrs = append(raddrs, addr)
		}
	}

	// add relay specific addrs to the list
	relayAddrCnt := 0
	for p := range rf.relays {
		addrs := cleanupAddressSet(rf.host.Peerstore().Addrs(p))
		relayAddrCnt += len(addrs)
		circuit := ma.StringCast(fmt.Sprintf("/p2p/%s/p2p-circuit", p.Pretty()))
		for _, addr := range addrs {
			pub := addr.Encapsulate(circuit)
			raddrs = append(raddrs, pub)
		}
	}

	rf.cachedAddrs = raddrs
	rf.cachedAddrsExpiry = rf.conf.clock.Now().Add(30 * time.Second)

	rf.metricsTracer.RelayAddressCount(relayAddrCnt)
	return raddrs
}

func (rf *relayFinder) Start() error {
	rf.ctxCancelMx.Lock()
	defer rf.ctxCancelMx.Unlock()
	if rf.ctxCancel != nil {
		return errAlreadyRunning
	}
	log.Debug("starting relay finder")

	rf.initMetrics()

	ctx, cancel := context.WithCancel(context.Background())
	rf.ctxCancel = cancel
	rf.refCount.Add(1)
	go func() {
		defer rf.refCount.Done()
		rf.background(ctx)
	}()
	return nil
}

func (rf *relayFinder) Stop() error {
	rf.ctxCancelMx.Lock()
	defer rf.ctxCancelMx.Unlock()
	log.Debug("stopping relay finder")
	if rf.ctxCancel != nil {
		rf.ctxCancel()
	}
	rf.refCount.Wait()
	rf.ctxCancel = nil

	rf.resetMetrics()
	return nil
}

func (rf *relayFinder) initMetrics() {
	rf.metricsTracer.DesiredReservations(rf.conf.desiredRelays)

	rf.relayMx.Lock()
	rf.metricsTracer.ReservationOpened(len(rf.relays))
	rf.relayMx.Unlock()

	rf.candidateMx.Lock()
	rf.metricsTracer.CandidateAdded(len(rf.candidates))
	rf.candidateMx.Unlock()
}

func (rf *relayFinder) resetMetrics() {
	rf.relayMx.Lock()
	rf.metricsTracer.ReservationEnded(len(rf.relays))
	rf.relayMx.Unlock()

	rf.candidateMx.Lock()
	rf.metricsTracer.CandidateRemoved(len(rf.candidates))
	rf.candidateMx.Unlock()

	rf.metricsTracer.RelayAddressCount(0)
	rf.metricsTracer.ScheduledWorkUpdated(&scheduledWorkTimes{})
}
