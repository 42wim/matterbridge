package libp2pwebtransport

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
)

var webtransportMA = ma.StringCast("/quic-v1/webtransport")

func toWebtransportMultiaddr(na net.Addr) (ma.Multiaddr, error) {
	addr, err := manet.FromNetAddr(na)
	if err != nil {
		return nil, err
	}
	if _, err := addr.ValueForProtocol(ma.P_UDP); err != nil {
		return nil, errors.New("not a UDP address")
	}
	return addr.Encapsulate(webtransportMA), nil
}

func stringToWebtransportMultiaddr(str string) (ma.Multiaddr, error) {
	host, portStr, err := net.SplitHostPort(str)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.New("failed to parse IP")
	}
	return toWebtransportMultiaddr(&net.UDPAddr{IP: ip, Port: int(port)})
}

func extractCertHashes(addr ma.Multiaddr) ([]multihash.DecodedMultihash, error) {
	certHashesStr := make([]string, 0, 2)
	ma.ForEach(addr, func(c ma.Component) bool {
		if c.Protocol().Code == ma.P_CERTHASH {
			certHashesStr = append(certHashesStr, c.Value())
		}
		return true
	})
	certHashes := make([]multihash.DecodedMultihash, 0, len(certHashesStr))
	for _, s := range certHashesStr {
		_, ch, err := multibase.Decode(s)
		if err != nil {
			return nil, fmt.Errorf("failed to multibase-decode certificate hash: %w", err)
		}
		dh, err := multihash.Decode(ch)
		if err != nil {
			return nil, fmt.Errorf("failed to multihash-decode certificate hash: %w", err)
		}
		certHashes = append(certHashes, *dh)
	}
	return certHashes, nil
}

func addrComponentForCert(hash []byte) (ma.Multiaddr, error) {
	mh, err := multihash.Encode(hash, multihash.SHA2_256)
	if err != nil {
		return nil, err
	}
	certStr, err := multibase.Encode(multibase.Base58BTC, mh)
	if err != nil {
		return nil, err
	}
	return ma.NewComponent(ma.ProtocolWithCode(ma.P_CERTHASH).Name, certStr)
}

// IsWebtransportMultiaddr returns true if the given multiaddr is a well formed
// webtransport multiaddr. Returns the number of certhashes found.
func IsWebtransportMultiaddr(multiaddr ma.Multiaddr) (bool, int) {
	const (
		init = iota
		foundUDP
		foundQuicV1
		foundWebTransport
	)
	state := init
	certhashCount := 0

	ma.ForEach(multiaddr, func(c ma.Component) bool {
		if c.Protocol().Code == ma.P_QUIC_V1 && state == init {
			state = foundUDP
		}
		if c.Protocol().Code == ma.P_QUIC_V1 && state == foundUDP {
			state = foundQuicV1
		}
		if c.Protocol().Code == ma.P_WEBTRANSPORT && state == foundQuicV1 {
			state = foundWebTransport
		}
		if c.Protocol().Code == ma.P_CERTHASH && state == foundWebTransport {
			certhashCount++
		}
		return true
	})
	return state == foundWebTransport, certhashCount
}
