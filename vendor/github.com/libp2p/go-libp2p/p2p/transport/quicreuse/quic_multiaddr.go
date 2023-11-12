package quicreuse

import (
	"errors"
	"net"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/quic-go/quic-go"
)

var (
	quicV1MA      = ma.StringCast("/quic-v1")
	quicDraft29MA = ma.StringCast("/quic")
)

func ToQuicMultiaddr(na net.Addr, version quic.VersionNumber) (ma.Multiaddr, error) {
	udpMA, err := manet.FromNetAddr(na)
	if err != nil {
		return nil, err
	}
	switch version {
	case quic.VersionDraft29:
		return udpMA.Encapsulate(quicDraft29MA), nil
	case quic.Version1:
		return udpMA.Encapsulate(quicV1MA), nil
	default:
		return nil, errors.New("unknown QUIC version")
	}
}

func FromQuicMultiaddr(addr ma.Multiaddr) (*net.UDPAddr, quic.VersionNumber, error) {
	var version quic.VersionNumber
	var partsBeforeQUIC []ma.Multiaddr
	ma.ForEach(addr, func(c ma.Component) bool {
		switch c.Protocol().Code {
		case ma.P_QUIC:
			version = quic.VersionDraft29
			return false
		case ma.P_QUIC_V1:
			version = quic.Version1
			return false
		default:
			partsBeforeQUIC = append(partsBeforeQUIC, &c)
			return true
		}
	})
	if len(partsBeforeQUIC) == 0 {
		return nil, version, errors.New("no addr before QUIC component")
	}
	if version == 0 {
		// Not found
		return nil, version, errors.New("unknown QUIC version")
	}
	netAddr, err := manet.ToNetAddr(ma.Join(partsBeforeQUIC...))
	if err != nil {
		return nil, version, err
	}
	udpAddr, ok := netAddr.(*net.UDPAddr)
	if !ok {
		return nil, 0, errors.New("not a *net.UDPAddr")
	}
	return udpAddr, version, nil
}
