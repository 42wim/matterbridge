package ssdp

import (
	"fmt"
	"net"
)

// LocationProvider provides address for Location header which can be reached from
// "from" address network.
type LocationProvider interface {
	// Location provides an address be reachable from the network located
	// by "from" address or "ifi" interface.
	// One of "from" or "ifi" must not be nil.
	Location(from net.Addr, ifi *net.Interface) string
}

// LocationProviderFunc type is an adapter to allow the use of ordinary
// functions are location providers.
type LocationProviderFunc func(net.Addr, *net.Interface) string

func (f LocationProviderFunc) Location(from net.Addr, ifi *net.Interface) string {
	return f(from, ifi)
}

type fixedLocation string

func (s fixedLocation) Location(net.Addr, *net.Interface) string {
	return string(s)
}

func toLocationProvider(v interface{}) (LocationProvider, error) {
	switch w := v.(type) {
	case string:
		return fixedLocation(w), nil
	case LocationProvider:
		return w, nil
	default:
		return nil, fmt.Errorf("location should be a string or a ssdp.LocationProvider but got %T", w)
	}
}
