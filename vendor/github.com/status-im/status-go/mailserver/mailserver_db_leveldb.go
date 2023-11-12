package mailserver

import (
	"fmt"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/status-im/status-go/eth-node/types"
	waku "github.com/status-im/status-go/waku/common"
)

type LevelDB struct {
	// We can't embed as there are some state problems with go-routines
	ldb  *leveldb.DB
	name string
	done chan struct{}
}

type LevelDBIterator struct {
	iterator.Iterator
}

func (i *LevelDBIterator) DBKey() (*DBKey, error) {
	return &DBKey{
		raw: i.Key(),
	}, nil
}

func (i *LevelDBIterator) GetEnvelopeByTopicsMap(topics map[types.TopicType]bool) ([]byte, error) {
	rawValue := make([]byte, len(i.Value()))
	copy(rawValue, i.Value())

	key, err := i.DBKey()
	if err != nil {
		return nil, err
	}

	if !topics[key.Topic()] {
		return nil, nil
	}

	return rawValue, nil
}

func (i *LevelDBIterator) GetEnvelopeByBloomFilter(bloom []byte) ([]byte, error) {
	var envelopeBloom []byte
	rawValue := make([]byte, len(i.Value()))
	copy(rawValue, i.Value())

	key, err := i.DBKey()
	if err != nil {
		return nil, err
	}

	if len(key.Bytes()) != DBKeyLength {
		var err error
		envelopeBloom, err = extractBloomFromEncodedEnvelope(rawValue)
		if err != nil {
			return nil, err
		}
	} else {
		envelopeBloom = types.TopicToBloom(key.Topic())
	}
	if !types.BloomFilterMatch(bloom, envelopeBloom) {
		return nil, nil
	}
	return rawValue, nil
}

func (i *LevelDBIterator) Release() error {
	i.Iterator.Release()
	return nil
}

func NewLevelDB(dataDir string) (*LevelDB, error) {
	// Open opens an existing leveldb database
	db, err := leveldb.OpenFile(dataDir, nil)
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		log.Info("database is corrupted trying to recover", "path", dataDir)
		db, err = leveldb.RecoverFile(dataDir, nil)
	}

	instance := LevelDB{
		ldb:  db,
		name: dataDir, // name is used for metrics labels
		done: make(chan struct{}),
	}

	// initialize the metric value
	instance.updateArchivedEnvelopesCount()
	// checking count on every insert is inefficient
	go func() {
		for {
			select {
			case <-instance.done:
				return
			case <-time.After(time.Second * envelopeCountCheckInterval):
				instance.updateArchivedEnvelopesCount()
			}
		}
	}()
	return &instance, err
}

// GetEnvelope get an envelope by its key
func (db *LevelDB) GetEnvelope(key *DBKey) ([]byte, error) {
	defer recoverLevelDBPanics("GetEnvelope")
	return db.ldb.Get(key.Bytes(), nil)
}

func (db *LevelDB) updateArchivedEnvelopesCount() {
	if count, err := db.envelopesCount(); err != nil {
		log.Warn("db query for envelopes count failed", "err", err)
	} else {
		archivedEnvelopesGauge.WithLabelValues(db.name).Set(float64(count))
	}
}

// Build iterator returns an iterator given a start/end and a cursor
func (db *LevelDB) BuildIterator(query CursorQuery) (Iterator, error) {
	defer recoverLevelDBPanics("BuildIterator")

	i := db.ldb.NewIterator(&util.Range{Start: query.start, Limit: query.end}, nil)

	envelopeQueriesCounter.WithLabelValues("unknown", "unknown").Inc()
	// seek to the end as we want to return envelopes in a descending order
	if len(query.cursor) == CursorLength {
		i.Seek(query.cursor)
	}
	return &LevelDBIterator{i}, nil
}

// Prune removes envelopes older than time
func (db *LevelDB) Prune(t time.Time, batchSize int) (int, error) {
	defer recoverLevelDBPanics("Prune")

	var zero types.Hash
	var emptyTopic types.TopicType
	kl := NewDBKey(0, emptyTopic, zero)
	ku := NewDBKey(uint32(t.Unix()), emptyTopic, zero)
	query := CursorQuery{
		start: kl.Bytes(),
		end:   ku.Bytes(),
	}
	i, err := db.BuildIterator(query)
	if err != nil {
		return 0, err
	}
	defer func() { _ = i.Release() }()

	batch := leveldb.Batch{}
	removed := 0

	for i.Next() {
		dbKey, err := i.DBKey()
		if err != nil {
			return 0, err
		}

		batch.Delete(dbKey.Bytes())

		if batch.Len() == batchSize {
			if err := db.ldb.Write(&batch, nil); err != nil {
				return removed, err
			}

			removed = removed + batch.Len()
			batch.Reset()
		}
	}

	if batch.Len() > 0 {
		if err := db.ldb.Write(&batch, nil); err != nil {
			return removed, err
		}

		removed = removed + batch.Len()
	}

	return removed, nil
}

func (db *LevelDB) envelopesCount() (int, error) {
	defer recoverLevelDBPanics("envelopesCount")
	iterator, err := db.BuildIterator(CursorQuery{})
	if err != nil {
		return 0, err
	}
	// LevelDB does not have API for getting a count
	var count int
	for iterator.Next() {
		count++
	}
	return count, nil
}

// SaveEnvelope stores an envelope in leveldb and increments the metrics
func (db *LevelDB) SaveEnvelope(env types.Envelope) error {
	defer recoverLevelDBPanics("SaveEnvelope")

	key := NewDBKey(env.Expiry()-env.TTL(), env.Topic(), env.Hash())
	rawEnvelope, err := rlp.EncodeToBytes(env.Unwrap())
	if err != nil {
		log.Error(fmt.Sprintf("rlp.EncodeToBytes failed: %s", err))
		archivedErrorsCounter.WithLabelValues(db.name).Inc()
		return err
	}

	if err = db.ldb.Put(key.Bytes(), rawEnvelope, nil); err != nil {
		log.Error(fmt.Sprintf("Writing to DB failed: %s", err))
		archivedErrorsCounter.WithLabelValues(db.name).Inc()
	}
	archivedEnvelopesGauge.WithLabelValues(db.name).Inc()
	archivedEnvelopeSizeMeter.WithLabelValues(db.name).Observe(
		float64(waku.EnvelopeHeaderLength + env.Size()))
	return err
}

func (db *LevelDB) Close() error {
	select {
	case <-db.done:
	default:
		close(db.done)
	}
	return db.ldb.Close()
}

func recoverLevelDBPanics(calleMethodName string) {
	// Recover from possible goleveldb panics
	if r := recover(); r != nil {
		if errString, ok := r.(string); ok {
			log.Error(fmt.Sprintf("recovered from panic in %s: %s", calleMethodName, errString))
		}
	}
}
