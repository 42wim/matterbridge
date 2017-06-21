package steam

import (
	"math/rand"
	"time"

	"github.com/Philipp15b/go-steam/netutil"
)

// CMServers contains a list of worlwide servers
var CMServers = [][]string{
	{ // North American Servers
		// Chicago
		"162.254.193.44:27018",
		"162.254.193.44:27019",
		"162.254.193.44:27020",
		"162.254.193.44:27021",
		"162.254.193.45:27017",
		"162.254.193.45:27018",
		"162.254.193.45:27019",
		"162.254.193.45:27021",
		"162.254.193.46:27017",
		"162.254.193.46:27018",
		"162.254.193.46:27019",
		"162.254.193.46:27020",
		"162.254.193.46:27021",
		"162.254.193.47:27019",
		"162.254.193.47:27020",

		// Ashburn
		"208.78.164.9:27017",
		"208.78.164.9:27018",
		"208.78.164.9:27019",
		"208.78.164.10:27017",
		"208.78.164.10:27018",
		"208.78.164.10:27019",
		"208.78.164.11:27017",
		"208.78.164.11:27018",
		"208.78.164.11:27019",
		"208.78.164.12:27017",
		"208.78.164.12:27018",
		"208.78.164.12:27019",
		"208.78.164.13:27017",
		"208.78.164.13:27018",
		"208.78.164.13:27019",
		"208.78.164.14:27017",
		"208.78.164.14:27018",
		"208.78.164.14:27019",
	},
	{ // Europe Servers
		// Luxembourg
		"146.66.152.10:27017",
		"146.66.152.10:27018",
		"146.66.152.10:27019",
		"146.66.152.10:27020",
		"146.66.152.11:27017",
		"146.66.152.11:27018",
		"146.66.152.11:27019",
		"146.66.152.11:27020",

		// Poland
		"155.133.242.8:27017",
		"155.133.242.8:27018",
		"155.133.242.8:27019",
		"155.133.242.8:27020",
		"155.133.242.9:27017",
		"155.133.242.9:27018",
		"155.133.242.9:27019",
		"155.133.242.9:27020",

		// Vienna
		"146.66.155.8:27017",
		"146.66.155.8:27018",
		"146.66.155.8:27019",
		"146.66.155.8:27020",
		"185.25.182.10:27017",
		"185.25.182.10:27018",
		"185.25.182.10:27019",
		"185.25.182.10:27020",

		// London
		"162.254.196.40:27017",
		"162.254.196.40:27018",
		"162.254.196.40:27019",
		"162.254.196.40:27020",
		"162.254.196.40:27021",
		"162.254.196.41:27017",
		"162.254.196.41:27018",
		"162.254.196.41:27019",
		"162.254.196.41:27020",
		"162.254.196.41:27021",
		"162.254.196.42:27017",
		"162.254.196.42:27018",
		"162.254.196.42:27019",
		"162.254.196.42:27020",
		"162.254.196.42:27021",
		"162.254.196.43:27017",
		"162.254.196.43:27018",
		"162.254.196.43:27019",
		"162.254.196.43:27020",
		"162.254.196.43:27021",

		// Stockholm
		"185.25.180.14:27017",
		"185.25.180.14:27018",
		"185.25.180.14:27019",
		"185.25.180.14:27020",
		"185.25.180.15:27017",
		"185.25.180.15:27018",
		"185.25.180.15:27019",
		"185.25.180.15:27020",
	},
}

// GetRandomCM returns back a random server anywhere
func GetRandomCM() *netutil.PortAddr {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	servers := append(CMServers[0], CMServers[1]...)
	addr := netutil.ParsePortAddr(servers[rng.Int31n(int32(len(servers)))])
	if addr == nil {
		panic("invalid address in CMServers slice")
	}
	return addr
}

// GetRandomNorthAmericaCM returns back a random server in north america
func GetRandomNorthAmericaCM() *netutil.PortAddr {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	addr := netutil.ParsePortAddr(CMServers[0][rng.Int31n(int32(len(CMServers[0])))])
	if addr == nil {
		panic("invalid address in CMServers slice")
	}
	return addr
}

// GetRandomEuropeCM returns back a random server in europe
func GetRandomEuropeCM() *netutil.PortAddr {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	addr := netutil.ParsePortAddr(CMServers[1][rng.Int31n(int32(len(CMServers[1])))])
	if addr == nil {
		panic("invalid address in CMServers slice")
	}
	return addr
}
