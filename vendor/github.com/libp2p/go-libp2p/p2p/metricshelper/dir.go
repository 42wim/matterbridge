package metricshelper

import "github.com/libp2p/go-libp2p/core/network"

func GetDirection(dir network.Direction) string {
	switch dir {
	case network.DirOutbound:
		return "outbound"
	case network.DirInbound:
		return "inbound"
	default:
		return "unknown"
	}
}
