package netutil

import (
	"net"
	"strconv"
	"strings"
)

// An addr that is neither restricted to TCP nor UDP, but has an IP and a port.
type PortAddr struct {
	IP   net.IP
	Port uint16
}

// Parses an IP address with a port, for example "209.197.29.196:27017".
// If the given string is not valid, this function returns nil.
func ParsePortAddr(addr string) *PortAddr {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return nil
	}
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return nil
	}
	port, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return nil
	}
	return &PortAddr{ip, uint16(port)}
}

func (p *PortAddr) ToTCPAddr() *net.TCPAddr {
	return &net.TCPAddr{p.IP, int(p.Port), ""}
}

func (p *PortAddr) ToUDPAddr() *net.UDPAddr {
	return &net.UDPAddr{p.IP, int(p.Port), ""}
}

func (p *PortAddr) String() string {
	return p.IP.String() + ":" + strconv.FormatUint(uint64(p.Port), 10)
}
