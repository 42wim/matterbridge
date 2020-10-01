package gumble

import (
	"crypto/x509"
	"net"
	"time"
)

// UserStats contains additional information about a user.
type UserStats struct {
	// The owner of the stats.
	User *User

	// Stats about UDP packets sent from the client.
	FromClient UserStatsUDP
	// Stats about UDP packets sent by the server.
	FromServer UserStatsUDP

	// Number of UDP packets sent by the user.
	UDPPackets uint32
	// Average UDP ping.
	UDPPingAverage float32
	// UDP ping variance.
	UDPPingVariance float32

	// Number of TCP packets sent by the user.
	TCPPackets uint32
	// Average TCP ping.
	TCPPingAverage float32
	// TCP ping variance.
	TCPPingVariance float32

	// The user's version.
	Version Version
	// When the user connected to the server.
	Connected time.Time
	// How long the user has been idle.
	Idle time.Duration
	// How much bandwidth the user is current using.
	Bandwidth int
	// The user's certificate chain.
	Certificates []*x509.Certificate
	// Does the user have a strong certificate? A strong certificate is one that
	// is not self signed, nor expired, etc.
	StrongCertificate bool
	// A list of CELT versions supported by the user's client.
	CELTVersions []int32
	// Does the user's client supports the Opus audio codec?
	Opus bool

	// The user's IP address.
	IP net.IP
}

// UserStatsUDP contains stats about UDP packets that have been sent to or from
// the server.
type UserStatsUDP struct {
	Good   uint32
	Late   uint32
	Lost   uint32
	Resync uint32
}
