package db

import (
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/ethereum/go-ethereum/log"
)

type storagePrefix byte

const (
	// PeersCache is used for the db entries used for peers DB
	PeersCache storagePrefix = iota
	// DeduplicatorCache is used for the db entries used for messages
	// deduplication cache
	DeduplicatorCache
	// MailserversCache is a list of mail servers provided by users.
	MailserversCache
	// TopicHistoryBucket isolated bucket for storing history metadata.
	TopicHistoryBucket
	// HistoryRequestBucket isolated bucket for storing list of pending requests.
	HistoryRequestBucket
)

// NewMemoryDB returns leveldb with memory backend prefixed with a bucket.
func NewMemoryDB() (*leveldb.DB, error) {
	return leveldb.Open(storage.NewMemStorage(), nil)
}

// NewDBNamespace returns instance that ensures isolated operations.
func NewDBNamespace(db Storage, prefix storagePrefix) LevelDBNamespace {
	return LevelDBNamespace{
		db:     db,
		prefix: prefix,
	}
}

// NewMemoryDBNamespace wraps in memory leveldb with provided bucket.
// Mostly used for tests. Including tests in other packages.
func NewMemoryDBNamespace(prefix storagePrefix) (pdb LevelDBNamespace, err error) {
	db, err := NewMemoryDB()
	if err != nil {
		return pdb, err
	}
	return NewDBNamespace(LevelDBStorage{db: db}, prefix), nil
}

// Key creates a DB key for a specified service with specified data
func Key(prefix storagePrefix, data ...[]byte) []byte {
	keyLength := 1
	for _, d := range data {
		keyLength += len(d)
	}
	key := make([]byte, keyLength)
	key[0] = byte(prefix)
	startPos := 1
	for _, d := range data {
		copy(key[startPos:], d[:])
		startPos += len(d)
	}

	return key
}

// Create returns status pointer to leveldb.DB.
func Create(path, dbName string) (*leveldb.DB, error) {
	// Create euphemeral storage if the node config path isn't provided
	if path == "" {
		return leveldb.Open(storage.NewMemStorage(), nil)
	}

	path = filepath.Join(path, dbName)
	return Open(path, &opt.Options{OpenFilesCacheCapacity: 5})
}

// Open opens an existing leveldb database
func Open(path string, opts *opt.Options) (db *leveldb.DB, err error) {
	db, err = leveldb.OpenFile(path, opts)
	if _, iscorrupted := err.(*errors.ErrCorrupted); iscorrupted {
		log.Info("database is corrupted trying to recover", "path", path)
		db, err = leveldb.RecoverFile(path, nil)
	}
	return
}

// LevelDBNamespace database where all operations will be prefixed with a certain bucket.
type LevelDBNamespace struct {
	db     Storage
	prefix storagePrefix
}

func (db LevelDBNamespace) prefixedKey(key []byte) []byte {
	endkey := make([]byte, len(key)+1)
	endkey[0] = byte(db.prefix)
	copy(endkey[1:], key)
	return endkey
}

func (db LevelDBNamespace) Put(key, value []byte) error {
	return db.db.Put(db.prefixedKey(key), value)
}

func (db LevelDBNamespace) Get(key []byte) ([]byte, error) {
	return db.db.Get(db.prefixedKey(key))
}

// Range returns leveldb util.Range prefixed with a single byte.
// If prefix is nil range will iterate over all records in a given bucket.
func (db LevelDBNamespace) Range(prefix, limit []byte) *util.Range {
	if limit == nil {
		return util.BytesPrefix(db.prefixedKey(prefix))
	}
	return &util.Range{Start: db.prefixedKey(prefix), Limit: db.prefixedKey(limit)}
}

// Delete removes key from database.
func (db LevelDBNamespace) Delete(key []byte) error {
	return db.db.Delete(db.prefixedKey(key))
}

// NewIterator returns iterator for a given slice.
func (db LevelDBNamespace) NewIterator(slice *util.Range) NamespaceIterator {
	return NamespaceIterator{db.db.NewIterator(slice)}
}

// NamespaceIterator wraps leveldb iterator, works mostly the same way.
// The only difference is that first byte of the key is dropped.
type NamespaceIterator struct {
	iter iterator.Iterator
}

// Key returns key of the current item.
func (iter NamespaceIterator) Key() []byte {
	return iter.iter.Key()[1:]
}

// Value returns actual value of the current item.
func (iter NamespaceIterator) Value() []byte {
	return iter.iter.Value()
}

// Error returns accumulated error.
func (iter NamespaceIterator) Error() error {
	return iter.iter.Error()
}

// Prev moves cursor backward.
func (iter NamespaceIterator) Prev() bool {
	return iter.iter.Prev()
}

// Next moves cursor forward.
func (iter NamespaceIterator) Next() bool {
	return iter.iter.Next()
}
