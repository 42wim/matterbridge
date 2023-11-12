package tracker

import (
	"context"
	"encoding/binary"

	trHttp "github.com/anacrolix/torrent/tracker/http"
	"github.com/anacrolix/torrent/tracker/udp"
)

type udpClient struct {
	cl         *udp.ConnClient
	requestUri string
}

func (c *udpClient) Close() error {
	return c.cl.Close()
}

func (c *udpClient) Announce(ctx context.Context, req AnnounceRequest, opts trHttp.AnnounceOpt) (res AnnounceResponse, err error) {
	if req.IPAddress == 0 && opts.ClientIp4 != nil {
		// I think we're taking bytes in big-endian order (all IPs), and writing it to a natively
		// ordered uint32. This will be correctly ordered when written back out by the UDP client
		// later. I'm ignoring the fact that IPv6 announces shouldn't have an IP address, we have a
		// perfectly good IPv4 address.
		req.IPAddress = binary.BigEndian.Uint32(opts.ClientIp4.To4())
	}
	h, nas, err := c.cl.Announce(ctx, req, udp.Options{RequestUri: c.requestUri})
	if err != nil {
		return
	}
	res.Interval = h.Interval
	res.Leechers = h.Leechers
	res.Seeders = h.Seeders
	for _, cp := range nas.NodeAddrs() {
		res.Peers = append(res.Peers, trHttp.Peer{}.FromNodeAddr(cp))
	}
	return
}
