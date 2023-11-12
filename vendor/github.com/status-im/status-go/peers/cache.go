package peers

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/status-im/status-go/db"
)

// NewCache returns instance of PeersDatabase
func NewCache(db *leveldb.DB) *Cache {
	return &Cache{db: db}
}

// Cache maintains list of peers that were discovered.
type Cache struct {
	db *leveldb.DB
}

func makePeerKey(peerID enode.ID, topic discv5.Topic) []byte {
	return db.Key(db.PeersCache, []byte(topic), peerID.Bytes())
}

// AddPeer stores peer with a following key: <topic><peer ID>
func (d *Cache) AddPeer(peer *discv5.Node, topic discv5.Topic) error {
	data, err := peer.MarshalText()
	if err != nil {
		return err
	}
	pk, err := peer.ID.Pubkey()
	if err != nil {
		return err
	}
	return d.db.Put(makePeerKey(enode.PubkeyToIDV4(pk), topic), data, nil)
}

// RemovePeer deletes a peer from database.
func (d *Cache) RemovePeer(nodeID enode.ID, topic discv5.Topic) error {
	return d.db.Delete(makePeerKey(nodeID, topic), nil)
}

// GetPeersRange returns peers for a given topic with a limit.
func (d *Cache) GetPeersRange(topic discv5.Topic, limit int) (nodes []*discv5.Node) {
	key := db.Key(db.PeersCache, []byte(topic))
	// it is important to set Limit on the range passed to iterator, so that
	// we limit reads only to particular topic.
	iterator := d.db.NewIterator(util.BytesPrefix(key), nil)
	defer iterator.Release()
	count := 0
	for iterator.Next() && count < limit {
		node := discv5.Node{}
		value := iterator.Value()
		if err := node.UnmarshalText(value); err != nil {
			log.Error("can't unmarshal node", "value", value, "error", err)
			continue
		}
		nodes = append(nodes, &node)
		count++
	}
	return nodes
}
