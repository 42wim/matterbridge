package torrent

import (
	"net"
	"net/url"
)

type urlNetAddr struct {
	u *url.URL
}

func (me urlNetAddr) Network() string {
	return me.u.Scheme
}

func (me urlNetAddr) String() string {
	return me.u.Host
}

func remoteAddrFromUrl(urlStr string) net.Addr {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}
	return urlNetAddr{u}
}
