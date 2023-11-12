package swarm

import (
	"sort"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

// The 250ms value is from happy eyeballs RFC 8305. This is a rough estimate of 1 RTT
const (
	// duration by which TCP dials are delayed relative to the last QUIC dial
	PublicTCPDelay  = 250 * time.Millisecond
	PrivateTCPDelay = 30 * time.Millisecond

	// duration by which QUIC dials are delayed relative to previous QUIC dial
	PublicQUICDelay  = 250 * time.Millisecond
	PrivateQUICDelay = 30 * time.Millisecond

	// RelayDelay is the duration by which relay dials are delayed relative to direct addresses
	RelayDelay = 500 * time.Millisecond
)

// NoDelayDialRanker ranks addresses with no delay. This is useful for simultaneous connect requests.
func NoDelayDialRanker(addrs []ma.Multiaddr) []network.AddrDelay {
	return getAddrDelay(addrs, 0, 0, 0)
}

// DefaultDialRanker determines the ranking of outgoing connection attempts.
//
// Addresses are grouped into three distinct groups:
//
//   - private addresses (localhost and local networks (RFC 1918))
//   - public addresses
//   - relay addresses
//
// Within each group, the addresses are ranked according to the ranking logic described below.
// We then dial addresses according to this ranking, with short timeouts applied between dial attempts.
// This ranking logic dramatically reduces the number of simultaneous dial attempts, while introducing
// no additional latency in the vast majority of cases.
//
// Private and public address groups are dialed in parallel.
// Dialing relay addresses is delayed by 500 ms, if we have any non-relay alternatives.
//
// Within each group (private, public, relay addresses) we apply the following ranking logic:
//
//  1. If both IPv6 QUIC and IPv4 QUIC addresses are present, we do a Happy Eyeballs RFC 8305 style ranking.
//     First dial the IPv6 QUIC address with the lowest port. After this we dial the IPv4 QUIC address with
//     the lowest port delayed by 250ms (PublicQUICDelay) for public addresses, and 30ms (PrivateQUICDelay)
//     for local addresses. After this we dial all the rest of the addresses delayed by 250ms (PublicQUICDelay)
//     for public addresses, and 30ms (PrivateQUICDelay) for local addresses.
//  2. If only one of QUIC IPv6 or QUIC IPv4 addresses are present, dial the QUIC address with the lowest port
//     first. After this we dial the rest of the QUIC addresses delayed by 250ms (PublicQUICDelay) for public
//     addresses, and 30ms (PrivateQUICDelay) for local addresses.
//  3. If a QUIC or WebTransport address is present, TCP addresses dials are delayed relative to the last QUIC dial:
//     We prefer to end up with a QUIC connection. For public addresses, the delay introduced is 250ms (PublicTCPDelay),
//     and for private addresses 30ms (PrivateTCPDelay).
//
// We dial lowest ports first for QUIC addresses as they are more likely to be the listen port.
func DefaultDialRanker(addrs []ma.Multiaddr) []network.AddrDelay {
	relay, addrs := filterAddrs(addrs, isRelayAddr)
	pvt, addrs := filterAddrs(addrs, manet.IsPrivateAddr)
	public, addrs := filterAddrs(addrs, func(a ma.Multiaddr) bool { return isProtocolAddr(a, ma.P_IP4) || isProtocolAddr(a, ma.P_IP6) })

	var relayOffset time.Duration
	if len(public) > 0 {
		// if there is a public direct address available delay relay dials
		relayOffset = RelayDelay
	}

	res := make([]network.AddrDelay, 0, len(addrs))
	for i := 0; i < len(addrs); i++ {
		res = append(res, network.AddrDelay{Addr: addrs[i], Delay: 0})
	}

	res = append(res, getAddrDelay(pvt, PrivateTCPDelay, PrivateQUICDelay, 0)...)
	res = append(res, getAddrDelay(public, PublicTCPDelay, PublicQUICDelay, 0)...)
	res = append(res, getAddrDelay(relay, PublicTCPDelay, PublicQUICDelay, relayOffset)...)
	return res
}

// getAddrDelay ranks a group of addresses according to the ranking logic explained in
// documentation for defaultDialRanker.
// offset is used to delay all addresses by a fixed duration. This is useful for delaying all relay
// addresses relative to direct addresses.
func getAddrDelay(addrs []ma.Multiaddr, tcpDelay time.Duration, quicDelay time.Duration,
	offset time.Duration) []network.AddrDelay {

	sort.Slice(addrs, func(i, j int) bool { return score(addrs[i]) < score(addrs[j]) })

	// If the first address is (QUIC, IPv6), make the second address (QUIC, IPv4).
	happyEyeballs := false
	if len(addrs) > 0 {
		if isQUICAddr(addrs[0]) && isProtocolAddr(addrs[0], ma.P_IP6) {
			for i := 1; i < len(addrs); i++ {
				if isQUICAddr(addrs[i]) && isProtocolAddr(addrs[i], ma.P_IP4) {
					// make IPv4 address the second element
					if i > 1 {
						a := addrs[i]
						copy(addrs[2:], addrs[1:i])
						addrs[1] = a
					}
					happyEyeballs = true
					break
				}
			}
		}
	}

	res := make([]network.AddrDelay, 0, len(addrs))

	var totalTCPDelay time.Duration
	for i, addr := range addrs {
		var delay time.Duration
		switch {
		case isQUICAddr(addr):
			// For QUIC addresses we dial an IPv6 address, then after quicDelay an IPv4
			// address, then after quicDelay we dial rest of the addresses.
			if i == 1 {
				delay = quicDelay
			}
			if i > 1 && happyEyeballs {
				delay = 2 * quicDelay
			} else if i > 1 {
				delay = quicDelay
			}
			totalTCPDelay = delay + tcpDelay
		case isProtocolAddr(addr, ma.P_TCP):
			delay = totalTCPDelay
		}
		res = append(res, network.AddrDelay{Addr: addr, Delay: offset + delay})
	}
	return res
}

// score scores a multiaddress for dialing delay. Lower is better.
// The lower 16 bits of the result are the port. Low ports are ranked higher because they're
// more likely to be listen addresses.
// The addresses are ranked as:
// QUICv1 IPv6 > QUICdraft29 IPv6 > QUICv1 IPv4 > QUICdraft29 IPv4 >
// WebTransport IPv6 > WebTransport IPv4 > TCP IPv6 > TCP IPv4
func score(a ma.Multiaddr) int {
	ip4Weight := 0
	if isProtocolAddr(a, ma.P_IP4) {
		ip4Weight = 1 << 18
	}

	if _, err := a.ValueForProtocol(ma.P_WEBTRANSPORT); err == nil {
		p, _ := a.ValueForProtocol(ma.P_UDP)
		pi, _ := strconv.Atoi(p)
		return ip4Weight + (1 << 19) + pi
	}
	if _, err := a.ValueForProtocol(ma.P_QUIC); err == nil {
		p, _ := a.ValueForProtocol(ma.P_UDP)
		pi, _ := strconv.Atoi(p)
		return ip4Weight + pi + (1 << 17)
	}
	if _, err := a.ValueForProtocol(ma.P_QUIC_V1); err == nil {
		p, _ := a.ValueForProtocol(ma.P_UDP)
		pi, _ := strconv.Atoi(p)
		return ip4Weight + pi
	}
	if p, err := a.ValueForProtocol(ma.P_TCP); err == nil {
		pi, _ := strconv.Atoi(p)
		return ip4Weight + pi + (1 << 20)
	}
	return (1 << 30)
}

func isProtocolAddr(a ma.Multiaddr, p int) bool {
	found := false
	ma.ForEach(a, func(c ma.Component) bool {
		if c.Protocol().Code == p {
			found = true
			return false
		}
		return true
	})
	return found
}

func isQUICAddr(a ma.Multiaddr) bool {
	return isProtocolAddr(a, ma.P_QUIC) || isProtocolAddr(a, ma.P_QUIC_V1)
}

// filterAddrs filters an address slice in place
func filterAddrs(addrs []ma.Multiaddr, f func(a ma.Multiaddr) bool) (filtered, rest []ma.Multiaddr) {
	j := 0
	for i := 0; i < len(addrs); i++ {
		if f(addrs[i]) {
			addrs[i], addrs[j] = addrs[j], addrs[i]
			j++
		}
	}
	return addrs[:j], addrs[j:]
}
