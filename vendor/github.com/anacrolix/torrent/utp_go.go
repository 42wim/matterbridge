//go:build !cgo || disable_libutp
// +build !cgo disable_libutp

package torrent

import (
	"github.com/anacrolix/log"
	"github.com/anacrolix/utp"
)

func NewUtpSocket(network, addr string, _ firewallCallback, _ log.Logger) (utpSocket, error) {
	s, err := utp.NewSocket(network, addr)
	if s == nil {
		return nil, err
	} else {
		return s, err
	}
}
