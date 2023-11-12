//go:build cgo && !disable_libutp
// +build cgo,!disable_libutp

package torrent

import (
	utp "github.com/anacrolix/go-libutp"
	"github.com/anacrolix/log"
)

func NewUtpSocket(network, addr string, fc firewallCallback, logger log.Logger) (utpSocket, error) {
	s, err := utp.NewSocket(network, addr, utp.WithLogger(logger))
	if s == nil {
		return nil, err
	}
	if err != nil {
		return s, err
	}
	if fc != nil {
		s.SetSyncFirewallCallback(utp.FirewallCallback(fc))
	}
	return s, err
}
