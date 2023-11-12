package dnsdisc

import (
	"context"
	"net"
)

// GetResolver returns a *net.Resolver object using a custom nameserver, or
// the default system resolver if no nameserver is specified
func GetResolver(ctx context.Context, nameserver string) *net.Resolver {
	if nameserver == "" {
		return net.DefaultResolver
	}

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, net.JoinHostPort(nameserver, "53"))
		},
	}
}
