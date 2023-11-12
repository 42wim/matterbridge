package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Storage is an interface for common db operations.
type Storage interface {
	Put([]byte, []byte) error
	Delete([]byte) error
	Get([]byte) ([]byte, error)
	NewIterator(*util.Range) iterator.Iterator
}

// CommitStorage allows to write all tx/batched values atomically.
type CommitStorage interface {
	Storage
	Commit() error
}

// TransactionalStorage adds transaction features on top of regular storage.
type TransactionalStorage interface {
	Storage
	NewTx() CommitStorage
}

// NewMemoryLevelDBStorage returns LevelDBStorage instance with in memory leveldb backend.
func NewMemoryLevelDBStorage() (LevelDBStorage, error) {
	mdb, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		return LevelDBStorage{}, err
	}
	return NewLevelDBStorage(mdb), nil
}

// NewLevelDBStorage creates new LevelDBStorage instance.
func NewLevelDBStorage(db *leveldb.DB) LevelDBStorage {
	return LevelDBStorage{db: db}
}

// LevelDBStorage wrapper around leveldb.DB.
type LevelDBStorage struct {
	db *leveldb.DB
}

// Put upserts given key/value pair.
func (db LevelDBStorage) Put(key, buf []byte) error {
	return db.db.Put(key, buf, nil)
}

// Delete removes given key from database..
func (db LevelDBStorage) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

// Get returns value for a given key.
func (db LevelDBStorage) Get(key []byte) ([]byte, error) {
	return db.db.Get(key, nil)
}

// NewIterator returns new leveldb iterator.Iterator instance for a given range.
func (db LevelDBStorage) NewIterator(slice *util.Range) iterator.Iterator {
	return db.db.NewIterator(slice, nil)
}

// NewTx is a wrapper around leveldb.Batch that allows to write atomically.
func (db LevelDBStorage) NewTx() CommitStorage {
	return LevelDBTx{
		batch: &leveldb.Batch{},
		db:    db,
	}
}
