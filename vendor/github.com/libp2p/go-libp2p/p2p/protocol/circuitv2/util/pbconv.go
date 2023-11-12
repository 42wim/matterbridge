package util

import (
	"errors"

	"github.com/libp2p/go-libp2p/core/peer"
	pbv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/pb"

	ma "github.com/multiformats/go-multiaddr"
)

func PeerToPeerInfoV2(p *pbv2.Peer) (peer.AddrInfo, error) {
	if p == nil {
		return peer.AddrInfo{}, errors.New("nil peer")
	}

	id, err := peer.IDFromBytes(p.Id)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	addrs := make([]ma.Multiaddr, 0, len(p.Addrs))

	for _, addrBytes := range p.Addrs {
		a, err := ma.NewMultiaddrBytes(addrBytes)
		if err == nil {
			addrs = append(addrs, a)
		}
	}

	return peer.AddrInfo{ID: id, Addrs: addrs}, nil
}

func PeerInfoToPeerV2(pi peer.AddrInfo) *pbv2.Peer {
	addrs := make([][]byte, 0, len(pi.Addrs))
	for _, addr := range pi.Addrs {
		addrs = append(addrs, addr.Bytes())
	}

	return &pbv2.Peer{
		Id:    []byte(pi.ID),
		Addrs: addrs,
	}
}
