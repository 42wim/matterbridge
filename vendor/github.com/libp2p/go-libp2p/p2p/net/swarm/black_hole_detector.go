package swarm

import (
	"fmt"
	"sync"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

type blackHoleState int

const (
	blackHoleStateProbing blackHoleState = iota
	blackHoleStateAllowed
	blackHoleStateBlocked
)

func (st blackHoleState) String() string {
	switch st {
	case blackHoleStateProbing:
		return "Probing"
	case blackHoleStateAllowed:
		return "Allowed"
	case blackHoleStateBlocked:
		return "Blocked"
	default:
		return fmt.Sprintf("Unknown %d", st)
	}
}

type blackHoleResult int

const (
	blackHoleResultAllowed blackHoleResult = iota
	blackHoleResultProbing
	blackHoleResultBlocked
)

// blackHoleFilter provides black hole filtering for dials. This filter should be used in
// concert with a UDP of IPv6 address filter to detect UDP or IPv6 black hole. In a black
// holed environments dial requests are blocked and only periodic probes to check the
// state of the black hole are allowed.
//
// Requests are blocked if the number of successes in the last n dials is less than
// minSuccesses. If a request succeeds in Blocked state, the filter state is reset and n
// subsequent requests are allowed before reevaluating black hole state. Dials cancelled
// when some other concurrent dial succeeded are counted as failures. A sufficiently large
// n prevents false negatives in such cases.
type blackHoleFilter struct {
	// n serves the dual purpose of being the minimum number of requests after which we
	// probe the state of the black hole in blocked state and the minimum number of
	// completed dials required before evaluating black hole state.
	n int
	// minSuccesses is the minimum number of Success required in the last n dials
	// to consider we are not blocked.
	minSuccesses int
	// name for the detector.
	name string

	// requests counts number of dial requests to peers. We handle request at a peer
	// level and record results at individual address dial level.
	requests int
	// dialResults of the last `n` dials. A successful dial is true.
	dialResults []bool
	// successes is the count of successful dials in outcomes
	successes int
	// state is the current state of the detector
	state blackHoleState

	mu            sync.Mutex
	metricsTracer MetricsTracer
}

// RecordResult records the outcome of a dial. A successful dial will change the state
// of the filter to Allowed. A failed dial only blocks subsequent requests if the success
// fraction over the last n outcomes is less than the minSuccessFraction of the filter.
func (b *blackHoleFilter) RecordResult(success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == blackHoleStateBlocked && success {
		// If the call succeeds in a blocked state we reset to allowed.
		// This is better than slowly accumulating values till we cross the minSuccessFraction
		// threshold since a blackhole is a binary property.
		b.reset()
		return
	}

	if success {
		b.successes++
	}
	b.dialResults = append(b.dialResults, success)

	if len(b.dialResults) > b.n {
		if b.dialResults[0] {
			b.successes--
		}
		b.dialResults = b.dialResults[1:]
	}

	b.updateState()
	b.trackMetrics()
}

// HandleRequest returns the result of applying the black hole filter for the request.
func (b *blackHoleFilter) HandleRequest() blackHoleResult {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.requests++

	b.trackMetrics()

	if b.state == blackHoleStateAllowed {
		return blackHoleResultAllowed
	} else if b.state == blackHoleStateProbing || b.requests%b.n == 0 {
		return blackHoleResultProbing
	} else {
		return blackHoleResultBlocked
	}
}

func (b *blackHoleFilter) reset() {
	b.successes = 0
	b.dialResults = b.dialResults[:0]
	b.requests = 0
	b.updateState()
}

func (b *blackHoleFilter) updateState() {
	st := b.state

	if len(b.dialResults) < b.n {
		b.state = blackHoleStateProbing
	} else if b.successes >= b.minSuccesses {
		b.state = blackHoleStateAllowed
	} else {
		b.state = blackHoleStateBlocked
	}

	if st != b.state {
		log.Debugf("%s blackHoleDetector state changed from %s to %s", b.name, st, b.state)
	}
}

func (b *blackHoleFilter) trackMetrics() {
	if b.metricsTracer == nil {
		return
	}

	nextRequestAllowedAfter := 0
	if b.state == blackHoleStateBlocked {
		nextRequestAllowedAfter = b.n - (b.requests % b.n)
	}

	successFraction := 0.0
	if len(b.dialResults) > 0 {
		successFraction = float64(b.successes) / float64(len(b.dialResults))
	}

	b.metricsTracer.UpdatedBlackHoleFilterState(
		b.name,
		b.state,
		nextRequestAllowedAfter,
		successFraction,
	)
}

// blackHoleDetector provides UDP and IPv6 black hole detection using a `blackHoleFilter`
// for each. For details of the black hole detection logic see `blackHoleFilter`.
//
// black hole filtering is done at a peer dial level to ensure that periodic probes to
// detect change of the black hole state are actually dialed and are not skipped
// because of dial prioritisation logic.
type blackHoleDetector struct {
	udp, ipv6 *blackHoleFilter
}

// FilterAddrs filters the peer's addresses removing black holed addresses
func (d *blackHoleDetector) FilterAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	hasUDP, hasIPv6 := false, false
	for _, a := range addrs {
		if !manet.IsPublicAddr(a) {
			continue
		}
		if isProtocolAddr(a, ma.P_UDP) {
			hasUDP = true
		}
		if isProtocolAddr(a, ma.P_IP6) {
			hasIPv6 = true
		}
	}

	udpRes := blackHoleResultAllowed
	if d.udp != nil && hasUDP {
		udpRes = d.udp.HandleRequest()
	}

	ipv6Res := blackHoleResultAllowed
	if d.ipv6 != nil && hasIPv6 {
		ipv6Res = d.ipv6.HandleRequest()
	}

	return ma.FilterAddrs(
		addrs,
		func(a ma.Multiaddr) bool {
			if !manet.IsPublicAddr(a) {
				return true
			}
			// allow all UDP addresses while probing irrespective of IPv6 black hole state
			if udpRes == blackHoleResultProbing && isProtocolAddr(a, ma.P_UDP) {
				return true
			}
			// allow all IPv6 addresses while probing irrespective of UDP black hole state
			if ipv6Res == blackHoleResultProbing && isProtocolAddr(a, ma.P_IP6) {
				return true
			}

			if udpRes == blackHoleResultBlocked && isProtocolAddr(a, ma.P_UDP) {
				return false
			}
			if ipv6Res == blackHoleResultBlocked && isProtocolAddr(a, ma.P_IP6) {
				return false
			}
			return true
		},
	)
}

// RecordResult updates the state of the relevant `blackHoleFilter`s for addr
func (d *blackHoleDetector) RecordResult(addr ma.Multiaddr, success bool) {
	if !manet.IsPublicAddr(addr) {
		return
	}
	if d.udp != nil && isProtocolAddr(addr, ma.P_UDP) {
		d.udp.RecordResult(success)
	}
	if d.ipv6 != nil && isProtocolAddr(addr, ma.P_IP6) {
		d.ipv6.RecordResult(success)
	}
}

// blackHoleConfig is the config used for black hole detection
type blackHoleConfig struct {
	// Enabled enables black hole detection
	Enabled bool
	// N is the size of the sliding window used to evaluate black hole state
	N int
	// MinSuccesses is the minimum number of successes out of N required to not
	// block requests
	MinSuccesses int
}

func newBlackHoleDetector(udpConfig, ipv6Config blackHoleConfig, mt MetricsTracer) *blackHoleDetector {
	d := &blackHoleDetector{}

	if udpConfig.Enabled {
		d.udp = &blackHoleFilter{
			n:             udpConfig.N,
			minSuccesses:  udpConfig.MinSuccesses,
			name:          "UDP",
			metricsTracer: mt,
		}
	}

	if ipv6Config.Enabled {
		d.ipv6 = &blackHoleFilter{
			n:             ipv6Config.N,
			minSuccesses:  ipv6Config.MinSuccesses,
			name:          "IPv6",
			metricsTracer: mt,
		}
	}
	return d
}
