package peer_exchange

import (
	"container/list"
	"fmt"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	wenr "github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"github.com/waku-org/go-waku/waku/v2/utils"
)

type ShardInfo struct {
	clusterID uint16
	shard     uint16
}
type shardLRU struct {
	size       int // number of nodes allowed per shard
	idToNode   map[enode.ID][]*list.Element
	shardNodes map[ShardInfo]*list.List
	rng        *rand.Rand
	mu         sync.RWMutex
}

func newShardLRU(size int) *shardLRU {
	return &shardLRU{
		idToNode:   map[enode.ID][]*list.Element{},
		shardNodes: map[ShardInfo]*list.List{},
		size:       size,
		rng:        rand.New(rand.NewSource(rand.Int63())),
	}
}

type nodeWithShardInfo struct {
	key  ShardInfo
	node *enode.Node
}

// time complexity: O(number of previous indexes present for node.ID)
func (l *shardLRU) remove(node *enode.Node) {
	elements := l.idToNode[node.ID()]
	for _, element := range elements {
		key := element.Value.(nodeWithShardInfo).key
		l.shardNodes[key].Remove(element)
	}
	delete(l.idToNode, node.ID())
}

// if a node is removed for a list, remove it from idToNode too
func (l *shardLRU) removeFromIdToNode(ele *list.Element) {
	nodeID := ele.Value.(nodeWithShardInfo).node.ID()
	for ind, entries := range l.idToNode[nodeID] {
		if entries == ele {
			l.idToNode[nodeID] = append(l.idToNode[nodeID][:ind], l.idToNode[nodeID][ind+1:]...)
			break
		}
	}
	if len(l.idToNode[nodeID]) == 0 {
		delete(l.idToNode, nodeID)
	}
}

func nodeToRelayShard(node *enode.Node) (*protocol.RelayShards, error) {
	shard, err := wenr.RelaySharding(node.Record())
	if err != nil {
		return nil, err
	}

	if shard == nil { // if no shard info, then add to node to Cluster 0, Index 0
		shard = &protocol.RelayShards{
			ClusterID: 0,
			ShardIDs:  []uint16{0},
		}
	}

	return shard, nil
}

// time complexity: O(new number of indexes in node's shard)
func (l *shardLRU) add(node *enode.Node) error {
	shard, err := nodeToRelayShard(node)
	if err != nil {
		return err
	}

	elements := []*list.Element{}
	for _, index := range shard.ShardIDs {
		key := ShardInfo{
			shard.ClusterID,
			index,
		}
		if l.shardNodes[key] == nil {
			l.shardNodes[key] = list.New()
		}
		if l.shardNodes[key].Len() >= l.size {
			oldest := l.shardNodes[key].Back()
			l.removeFromIdToNode(oldest)
			l.shardNodes[key].Remove(oldest)
		}
		entry := l.shardNodes[key].PushFront(nodeWithShardInfo{
			key:  key,
			node: node,
		})
		elements = append(elements, entry)

	}
	l.idToNode[node.ID()] = elements

	return nil
}

// this will be called when the seq number of node is more than the one in cache
func (l *shardLRU) Add(node *enode.Node) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	// removing bcz previous node might be subscribed to different shards, we need to remove node from those shards
	l.remove(node)
	return l.add(node)
}

// clusterIndex is nil when peers for no specific shard are requested
func (l *shardLRU) GetRandomNodes(clusterIndex *ShardInfo, neededPeers int) (nodes []*enode.Node) {
	l.mu.Lock()
	defer l.mu.Unlock()

	availablePeers := l.len(clusterIndex)
	if availablePeers < neededPeers {
		neededPeers = availablePeers
	}
	// if clusterIndex is nil, then return all nodes
	var elements []*list.Element
	if clusterIndex == nil {
		elements = make([]*list.Element, 0, len(l.idToNode))
		for _, entries := range l.idToNode {
			elements = append(elements, entries[0])
		}
	} else if entries := l.shardNodes[*clusterIndex]; entries != nil && entries.Len() != 0 {
		elements = make([]*list.Element, 0, entries.Len())
		for ent := entries.Back(); ent != nil; ent = ent.Prev() {
			elements = append(elements, ent)
		}
	}
	utils.Logger().Info(fmt.Sprintf("%d", len(elements)))
	indexes := l.rng.Perm(len(elements))[0:neededPeers]
	for _, ind := range indexes {
		node := elements[ind].Value.(nodeWithShardInfo).node
		nodes = append(nodes, node)
		// this removes the node from all list (all cluster/shard pair that the node has) and adds it to the front
		l.remove(node)
		_ = l.add(node)
	}
	return nodes
}

// if clusterIndex is not nil, return len of nodes maintained for a given shard
// if clusterIndex is nil, return count of all nodes maintained
func (l *shardLRU) len(clusterIndex *ShardInfo) int {
	if clusterIndex == nil {
		return len(l.idToNode)
	}
	if entries := l.shardNodes[*clusterIndex]; entries != nil {
		return entries.Len()
	}
	return 0
}

// get the node with the given id, if it is present in cache
func (l *shardLRU) Get(id enode.ID) *enode.Node {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if elements, ok := l.idToNode[id]; ok && len(elements) > 0 {
		return elements[0].Value.(nodeWithShardInfo).node
	}
	return nil
}
