package bbolt

import (
	bolt "go.etcd.io/bbolt"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for bbolt (formerly known as Bolt / Bolt DB).
type Store struct {
	db         *bolt.DB
	bucketName string
	codec      encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s Store) Set(k string, v interface{}) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that bbolt can handle
	data, err := s.codec.Marshal(v)
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		return b.Put([]byte(k), data)
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
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		txData := b.Get([]byte(k))
		// txData is only valid during the transaction.
		// Its value must be copied to make it valid outside of the tx.
		// TODO: Benchmark if it's faster to copy + close tx,
		// or to keep the tx open until unmarshalling is done.
		if txData != nil {
			// `data = append([]byte{}, txData...)` would also work, but the following is more explicit
			data = make([]byte, len(txData))
			copy(data, txData)
		}
		return nil
	})
	if err != nil {
		return false, nil
	}

	// If no value was found return false
	if data == nil {
		return false, nil
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

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		return b.Delete([]byte(k))
	})
}

// Close closes the store.
// It must be called to make sure that all open transactions finish and to release all DB resources.
func (s Store) Close() error {
	return s.db.Close()
}

// Options are the options for the bbolt store.
type Options struct {
	// Bucket name for storing the key-value pairs.
	// Optional ("default" by default).
	BucketName string
	// Path of the DB file.
	// Optional ("bbolt.db" by default).
	Path string
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// BucketName: "default", Path: "bbolt.db", Codec: encoding.JSON
var DefaultOptions = Options{
	BucketName: "default",
	Path:       "bbolt.db",
	Codec:      encoding.JSON,
}

// NewStore creates a new bbolt store.
// Note: bbolt uses an exclusive write lock on the database file so it cannot be shared by multiple processes.
// So when creating multiple clients you should always use a new database file (by setting a different Path in the options).
//
// You must call the Close() method on the store when you're done working with it.
func NewStore(options Options) (Store, error) {
	result := Store{}

	// Set default values
	if options.BucketName == "" {
		options.BucketName = DefaultOptions.BucketName
	}
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	// Open DB
	db, err := bolt.Open(options.Path, 0600, nil)
	if err != nil {
		return result, err
	}

	// Create a bucket if it doesn't exist yet.
	// In bbolt key/value pairs are stored to and read from buckets.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(options.BucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	result.db = db
	result.bucketName = options.BucketName
	result.codec = options.Codec

	return result, nil
}
