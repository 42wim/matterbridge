package metricshelper

import ma "github.com/multiformats/go-multiaddr"

var transports = [...]int{ma.P_CIRCUIT, ma.P_WEBRTC, ma.P_WEBTRANSPORT, ma.P_QUIC, ma.P_QUIC_V1, ma.P_WSS, ma.P_WS, ma.P_TCP}

func GetTransport(a ma.Multiaddr) string {
	for _, t := range transports {
		if _, err := a.ValueForProtocol(t); err == nil {
			return ma.ProtocolWithCode(t).Name
		}
	}
	return "other"
}

func GetIPVersion(addr ma.Multiaddr) string {
	version := "unknown"
	ma.ForEach(addr, func(c ma.Component) bool {
		if c.Protocol().Code == ma.P_IP4 {
			version = "ip4"
			return false
		} else if c.Protocol().Code == ma.P_IP6 {
			version = "ip6"
			return false
		}
		return true
	})
	return version
}
