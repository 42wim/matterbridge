package badgerdb

import (
	"github.com/dgraph-io/badger"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for BadgerDB.
type Store struct {
	db    *badger.DB
	codec encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s Store) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that BadgerDB can handle
	data, err := s.codec.Marshal(v)
	if err != nil {
		return err
	}

	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(k), data)
	})
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (s Store) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	var data []byte
	err = s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		// item.Value() is only valid within the transaction.
		// We can either copy it ourselves or use the ValueCopy() method.
		// TODO: Benchmark if it's faster to copy + close tx,
		// or to keep the tx open until unmarshalling is done.
		data, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	// If no value was found return false
	if err == badger.ErrKeyNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, s.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (s Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(k))
	})
}

// Close closes the store.
// It must be called to make sure that all pending updates make their way to disk.
func (s Store) Close() error {
	return s.db.Close()
}

// Options are the options for the BadgerDB store.
type Options struct {
	// Directory for storing the DB files.
	// Optional ("BadgerDB" by default).
	Dir string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Dir: "BadgerDB", Codec: encoding.JSON
var DefaultOptions = Options{
	Dir:   "BadgerDB",
	Codec: encoding.JSON,
}

// NewStore creates a new BadgerDB store.
// Note: BadgerDB uses an exclusive write lock on the database directory so it cannot be shared by multiple processes.
// So when creating multiple clients you should always use a new database directory (by setting a different Path in the options).
//
// You must call the Close() method on the store when you're done working with it.
func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.Dir == "" {
		options.Dir = DefaultOptions.Dir
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	// Open the Badger database located in the options.Dir directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions(options.Dir)
	db, err := badger.Open(opts)
	if err != nil {
		return result, err
	}

	result.db = db
	result.codec = options.Codec

	return result, nil
}
