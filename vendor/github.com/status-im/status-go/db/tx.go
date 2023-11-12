package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBTx doesn't provide any read isolation. It allows committing all writes atomically (put/delete).
type LevelDBTx struct {
	batch *leveldb.Batch
	db    LevelDBStorage
}

// Put adds key/value to associated batch.
func (tx LevelDBTx) Put(key, buf []byte) error {
	tx.batch.Put(key, buf)
	return nil
}

// Delete adds delete operation to associated batch.
func (tx LevelDBTx) Delete(key []byte) error {
	tx.batch.Delete(key)
	return nil
}

// Get reads from currently committed state.
func (tx LevelDBTx) Get(key []byte) ([]byte, error) {
	return tx.db.Get(key)
}

// NewIterator returns iterator.Iterator that will read from currently committed state.
func (tx LevelDBTx) NewIterator(slice *util.Range) iterator.Iterator {
	return tx.db.NewIterator(slice)
}

// Commit writes batch atomically.
func (tx LevelDBTx) Commit() error {
	return tx.db.db.Write(tx.batch, nil)
}
