package signal

import "github.com/status-im/status-go/eth-node/types"

const (
	// EventPeerStats is sent when peer is added or removed.
	// it will be a map with capability=peer count k/v's.
	EventPeerStats = "wakuv2.peerstats"
)

// SendPeerStats sends discovery.summary signal.
func SendPeerStats(peerStats types.ConnStatus) {
	send(EventPeerStats, peerStats)
}
