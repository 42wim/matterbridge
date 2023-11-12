package utils

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// GetPeerID is used to extract the peerID from a multiaddress
func GetPeerID(m multiaddr.Multiaddr) (peer.ID, error) {
	peerIDStr, err := m.ValueForProtocol(multiaddr.P_P2P)
	if err != nil {
		return "", err
	}

	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return "", err
	}

	return peerID, nil
}
