package peer_exchange

import (
	"bufio"
	"bytes"

	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/waku-org/go-waku/waku/v2/protocol/peer_exchange/pb"
)

// simpleLRU internal uses container/list, which is ring buffer(double linked list)
type enrCache struct {
	// using lru, saves us from periodically cleaning the cache to mauintain a certain size
	data *shardLRU
}

// err on negative size
func newEnrCache(size int) *enrCache {
	inner := newShardLRU(int(size))
	return &enrCache{
		data: inner,
	}
}

// updating cache
func (c *enrCache) updateCache(node *enode.Node) error {
	currNode := c.data.Get(node.ID())
	if currNode == nil || node.Seq() > currNode.Seq() {
		return c.data.Add(node)
	}
	return nil
}

// get `numPeers` records of enr
func (c *enrCache) getENRs(neededPeers int, clusterIndex *ShardInfo) ([]*pb.PeerInfo, error) {
	//
	nodes := c.data.GetRandomNodes(clusterIndex, neededPeers)
	result := []*pb.PeerInfo{}
	for _, node := range nodes {
		//
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)
		err := node.Record().EncodeRLP(writer)
		if err != nil {
			return nil, err
		}
		writer.Flush()
		result = append(result, &pb.PeerInfo{
			Enr: b.Bytes(),
		})
	}
	return result, nil
}
