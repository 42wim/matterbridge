package multicast

import (
	"net"
)

type InterfacesProviderFunc func() []net.Interface

// InterfacesProvider specify a function to list all interfaces to multicast.
// If no provider are given, all possible interfaces will be used.
var InterfacesProvider InterfacesProviderFunc

// interfaces gets list of net.Interface to multicast UDP packet.
func interfaces() ([]net.Interface, error) {
	if p := InterfacesProvider; p != nil {
		if list := p(); len(list) > 0 {
			return list, nil
		}
	}
	return interfacesIPv4()
}

// interfacesIPv4 lists net.Interface on IPv4.
func interfacesIPv4() ([]net.Interface, error) {
	iflist, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	list := make([]net.Interface, 0, len(iflist))
	for _, ifi := range iflist {
		if !hasLinkUp(&ifi) || !hasMulticast(&ifi) || !hasIPv4Address(&ifi) {
			continue
		}
		list = append(list, ifi)
	}
	return list, nil
}

// hasLinkUp checks an I/F have link-up or not.
func hasLinkUp(ifi *net.Interface) bool {
	return ifi.Flags&net.FlagUp != 0
}

// hasMulticast checks an I/F supports multicast or not.
func hasMulticast(ifi *net.Interface) bool {
	return ifi.Flags&net.FlagMulticast != 0
}

// hasIPv4Address checks an I/F have IPv4 address.
func hasIPv4Address(ifi *net.Interface) bool {
	addrs, err := ifi.Addrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			continue
		}
		if len(ip.To4()) == net.IPv4len && !ip.IsUnspecified() {
			return true
		}
	}
	return false
}
