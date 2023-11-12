package mailservers

import (
	"sort"

	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/status-im/status-go/eth-node/types"
)

// GetFirstConnected returns first connected peer that is also added to a peer store.
// Raises ErrNoConnected if no peers are added to a peer store.
func GetFirstConnected(provider PeersProvider, store *PeerStore) (*enode.Node, error) {
	peers := provider.Peers()
	for _, p := range peers {
		if store.Exist(types.EnodeID(p.ID())) {
			return p.Node(), nil
		}
	}
	return nil, ErrNoConnected
}

// NodesNotifee interface to be notified when new nodes are received.
type NodesNotifee interface {
	Notify([]*enode.Node)
}

// EnsureUsedRecordsAddedFirst checks if any nodes were marked as connected before app went offline.
func EnsureUsedRecordsAddedFirst(ps *PeerStore, conn NodesNotifee) error {
	records, err := ps.cache.LoadAll()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].LastUsed.After(records[j].LastUsed)
	})
	all := recordsToNodes(records)
	if !records[0].LastUsed.IsZero() {
		conn.Notify(all[:1])
	}
	conn.Notify(all)
	return nil
}

func recordsToNodes(records []PeerRecord) []*enode.Node {
	nodes := make([]*enode.Node, len(records))
	for i := range records {
		nodes[i] = records[i].Node()
	}
	return nodes
}
