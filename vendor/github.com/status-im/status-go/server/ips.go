package server

import (
	"net"

	"go.uber.org/zap"

	"github.com/status-im/status-go/common"
	"github.com/status-im/status-go/logutils"
)

var (
	LocalHostIP = net.IP{127, 0, 0, 1}
	Localhost   = "Localhost"
)

func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "255.255.255.255:8080")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

// addrToIPNet casts addr to IPNet.
// Returns nil if addr is not of IPNet type.
func addrToIPNet(addr net.Addr) *net.IPNet {
	switch v := addr.(type) {
	case *net.IPNet:
		return v
	default:
		return nil
	}
}

// filterAddressesForPairingServer filters private unicast addresses.
// ips is a 2-dimensional array, where each sub-array is a list of IP
// addresses for a single network interface.
func filterAddressesForPairingServer(ips [][]net.IP) []net.IP {
	var result = map[string]net.IP{}

	for _, niIps := range ips {
		var ipv4, ipv6 []net.IP

		for _, ip := range niIps {

			// Only take private global unicast addrs
			if !ip.IsGlobalUnicast() || !ip.IsPrivate() {
				continue
			}

			if v := ip.To4(); v != nil {
				ipv4 = append(ipv4, ip)
			} else {
				ipv6 = append(ipv6, ip)
			}
		}

		// Prefer IPv4 over IPv6 for shorter connection string
		if len(ipv4) == 0 {
			for _, ip := range ipv6 {
				result[ip.String()] = ip
			}
		} else {
			for _, ip := range ipv4 {
				result[ip.String()] = ip
			}
		}
	}

	var out []net.IP
	for _, v := range result {
		out = append(out, v)
	}

	return out
}

// getAndroidLocalIP uses the net dial default ip as the standard Android IP address
// patches https://github.com/status-im/status-mobile/issues/17156
// more work required for a more robust implementation, see https://github.com/wlynxg/anet
func getAndroidLocalIP() ([][]net.IP, error) {
	ip, err := GetOutboundIP()
	if err != nil {
		return nil, err
	}

	return [][]net.IP{{ip}}, nil
}

// getLocalAddresses returns an array of all addresses
// of all available network interfaces.
func getLocalAddresses() ([][]net.IP, error) {
	// TODO until we can resolve Android errors when calling net.Interfaces() just return the outbound local address.
	//  Sorry Android
	if common.OperatingSystemIs(common.AndroidPlatform) {
		return getAndroidLocalIP()
	}

	nis, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var ips [][]net.IP

	for _, ni := range nis {
		var niIps []net.IP

		addrs, err := ni.Addrs()
		if err != nil {
			logutils.ZapLogger().Warn("failed to get addresses of network interface",
				zap.String("networkInterface", ni.Name),
				zap.Error(err))
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			if ipNet := addrToIPNet(addr); ipNet == nil {
				continue
			} else {
				ip = ipNet.IP
			}
			niIps = append(niIps, ip)
		}

		if len(niIps) > 0 {
			ips = append(ips, niIps)
		}
	}

	return ips, nil
}

// GetLocalAddressesForPairingServer is a high-level func
// that returns a list of addresses to be used by local pairing server.
func GetLocalAddressesForPairingServer() ([]net.IP, error) {
	ips, err := getLocalAddresses()
	if err != nil {
		return nil, err
	}
	return filterAddressesForPairingServer(ips), nil
}

// findReachableAddresses returns a filtered remoteIps array,
// in which each IP matches one or more of given localNets.
func findReachableAddresses(remoteIPs []net.IP, localNets []net.IPNet) []net.IP {
	var result []net.IP
	for _, localNet := range localNets {
		for _, remoteIP := range remoteIPs {
			if localNet.Contains(remoteIP) {
				result = append(result, remoteIP)
			}
		}
	}
	return result
}

// getAllAvailableNetworks collects all networks
// from available network interfaces.
func getAllAvailableNetworks() ([]net.IPNet, error) {
	var localNets []net.IPNet

	nis, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, ni := range nis {
		addrs, err := ni.Addrs()
		if err != nil {
			logutils.ZapLogger().Warn("failed to get addresses of network interface",
				zap.String("networkInterface", ni.Name),
				zap.Error(err))
			continue
		}

		for _, localAddr := range addrs {
			localNets = append(localNets, *addrToIPNet(localAddr))
		}
	}
	return localNets, nil
}

// FindReachableAddressesForPairingClient is a high-level func
// that returns a reachable server's address to be used by local pairing client.
func FindReachableAddressesForPairingClient(serverIps []net.IP) ([]net.IP, error) {
	// TODO until we can resolve Android errors when calling net.Interfaces() just noop. Sorry Android
	if common.OperatingSystemIs(common.AndroidPlatform) {
		return serverIps, nil
	}

	nets, err := getAllAvailableNetworks()
	if err != nil {
		return nil, err
	}
	return findReachableAddresses(serverIps, nets), nil
}
