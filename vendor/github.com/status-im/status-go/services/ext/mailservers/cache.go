package mailservers

import (
	"encoding/json"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/status-im/status-go/db"
	"github.com/status-im/status-go/eth-node/types"
)

// NewPeerRecord returns instance of the peer record.
func NewPeerRecord(node *enode.Node) PeerRecord {
	return PeerRecord{node: node}
}

// PeerRecord is set data associated with each peer that is stored on disk.
// PeerRecord stored with a enode as a key in leveldb, and body marshalled as json.
type PeerRecord struct {
	node *enode.Node

	// last time it was used.
	LastUsed time.Time
}

// Encode encodes PeerRecords to bytes.
func (r PeerRecord) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// ID returns enode identity of the node.
func (r PeerRecord) ID() enode.ID {
	return r.node.ID()
}

// Node returs pointer to original object.
// enode.Node doensn't allow modification on the object.
func (r PeerRecord) Node() *enode.Node {
	return r.node
}

// EncodeKey returns bytes that will should be used as a key in persistent storage.
func (r PeerRecord) EncodeKey() ([]byte, error) {
	return r.Node().MarshalText()
}

// NewCache returns pointer to a Cache instance.
func NewCache(db *leveldb.DB) *Cache {
	return &Cache{db: db}
}

// Cache is wrapper for operations on disk with leveldb.
type Cache struct {
	db *leveldb.DB
}

// Replace deletes old and adds new records in the persistent cache.
func (c *Cache) Replace(nodes []*enode.Node) error {
	batch := new(leveldb.Batch)
	iter := createPeersIterator(c.db)
	defer iter.Release()
	newNodes := nodesToMap(nodes)
	for iter.Next() {
		record, err := unmarshalKeyValue(keyWithoutPrefix(iter.Key()), iter.Value())
		if err != nil {
			return err
		}
		if _, exist := newNodes[types.EnodeID(record.ID())]; exist {
			delete(newNodes, types.EnodeID(record.ID()))
		} else {
			batch.Delete(iter.Key())
		}
	}
	for _, n := range newNodes {
		enodeKey, err := n.MarshalText()
		if err != nil {
			return err
		}
		// we put nil as default value doesn't have any state associated with them.
		batch.Put(db.Key(db.MailserversCache, enodeKey), nil)
	}
	return c.db.Write(batch, nil)
}

// LoadAll loads all records from persistent database.
func (c *Cache) LoadAll() (rst []PeerRecord, err error) {
	iter := createPeersIterator(c.db)
	for iter.Next() {
		record, err := unmarshalKeyValue(keyWithoutPrefix(iter.Key()), iter.Value())
		if err != nil {
			return nil, err
		}
		rst = append(rst, record)
	}
	return rst, nil
}

// UpdateRecord updates single record.
func (c *Cache) UpdateRecord(record PeerRecord) error {
	enodeKey, err := record.EncodeKey()
	if err != nil {
		return err
	}
	value, err := record.Encode()
	if err != nil {
		return err
	}
	return c.db.Put(db.Key(db.MailserversCache, enodeKey), value, nil)
}

func unmarshalKeyValue(key, value []byte) (record PeerRecord, err error) {
	enodeKey := key
	node := new(enode.Node)
	err = node.UnmarshalText(enodeKey)
	if err != nil {
		return record, err
	}
	record = PeerRecord{node: node}
	if len(value) != 0 {
		err = json.Unmarshal(value, &record)
	}
	return record, err
}

func nodesToMap(nodes []*enode.Node) map[types.EnodeID]*enode.Node {
	rst := map[types.EnodeID]*enode.Node{}
	for _, n := range nodes {
		rst[types.EnodeID(n.ID())] = n
	}
	return rst
}

func createPeersIterator(level *leveldb.DB) iterator.Iterator {
	return level.NewIterator(util.BytesPrefix([]byte{byte(db.MailserversCache)}), nil)
}

// keyWithoutPrefix removes first byte from key.
func keyWithoutPrefix(key []byte) []byte {
	return key[1:]
}
