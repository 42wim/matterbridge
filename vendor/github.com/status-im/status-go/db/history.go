package db

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/status-im/status-go/eth-node/types"
)

var (
	// ErrEmptyKey returned if key is not expected to be empty.
	ErrEmptyKey = errors.New("TopicHistoryKey is empty")
)

// DB is a common interface for DB operations.
type DB interface {
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	Delete([]byte) error
	Range([]byte, []byte) *util.Range
	NewIterator(*util.Range) NamespaceIterator
}

// TopicHistoryKey defines bytes that are used as unique key for TopicHistory.
// first 4 bytes are types.TopicType bytes
// next 8 bytes are time.Duration encoded in big endian notation.
type TopicHistoryKey [12]byte

// LoadTopicHistoryFromKey unmarshalls key into topic and duration and loads value of topic history
// from given database.
func LoadTopicHistoryFromKey(db DB, key TopicHistoryKey) (th TopicHistory, err error) {
	if (key == TopicHistoryKey{}) {
		return th, ErrEmptyKey
	}
	topic := types.TopicType{}
	copy(topic[:], key[:4])
	duration := binary.BigEndian.Uint64(key[4:])
	th = TopicHistory{db: db, Topic: topic, Duration: time.Duration(duration)}
	return th, th.Load()
}

// TopicHistory stores necessary information.
type TopicHistory struct {
	db DB
	// whisper topic
	Topic types.TopicType

	Duration time.Duration
	// Timestamp that was used for the first request with this topic.
	// Used to identify overlapping ranges.
	First time.Time
	// Timestamp of the last synced envelope.
	Current time.Time
	End     time.Time

	RequestID types.Hash
}

// Key returns unique identifier for this TopicHistory.
func (t TopicHistory) Key() TopicHistoryKey {
	key := TopicHistoryKey{}
	copy(key[:], t.Topic[:])
	binary.BigEndian.PutUint64(key[4:], uint64(t.Duration))
	return key
}

// Value marshalls TopicHistory into bytes.
func (t TopicHistory) Value() ([]byte, error) {
	return json.Marshal(t)
}

// Load TopicHistory from db using key and unmarshalls it.
func (t *TopicHistory) Load() error {
	key := t.Key()
	if (key == TopicHistoryKey{}) {
		return errors.New("key is empty")
	}
	value, err := t.db.Get(key[:])
	if err != nil {
		return err
	}
	return json.Unmarshal(value, t)
}

// Save persists TopicHistory on disk.
func (t TopicHistory) Save() error {
	key := t.Key()
	val, err := t.Value()
	if err != nil {
		return err
	}
	return t.db.Put(key[:], val)
}

// Delete removes topic history from database.
func (t TopicHistory) Delete() error {
	key := t.Key()
	return t.db.Delete(key[:])
}

// SameRange returns true if topic has same range, which means:
// true if Current is zero and Duration is the same
// and true if Current is the same
func (t TopicHistory) SameRange(other TopicHistory) bool {
	zero := time.Time{}
	if t.Current == zero && other.Current == zero {
		return t.Duration == other.Duration
	}
	return t.Current == other.Current
}

// Pending returns true if this topic was requested from a mail server.
func (t TopicHistory) Pending() bool {
	return t.RequestID != types.Hash{}
}

// HistoryRequest is kept in the database while request is in the progress.
// Stores necessary information to identify topics with associated ranges included in the request.
type HistoryRequest struct {
	requestDB DB
	topicDB   DB

	histories []TopicHistory

	// Generated ID
	ID types.Hash
	// List of the topics
	TopicHistoryKeys []TopicHistoryKey
}

// AddHistory adds instance to internal list of instance and add instance key to the list
// which will be persisted on disk.
func (req *HistoryRequest) AddHistory(history TopicHistory) {
	req.histories = append(req.histories, history)
	req.TopicHistoryKeys = append(req.TopicHistoryKeys, history.Key())
}

// Histories returns internal lsit of topic histories.
func (req *HistoryRequest) Histories() []TopicHistory {
	// TODO Lazy load from database on first access
	return req.histories
}

// Value returns content of HistoryRequest as bytes.
func (req HistoryRequest) Value() ([]byte, error) {
	return json.Marshal(req)
}

// Save persists all attached histories and request itself on the disk.
func (req HistoryRequest) Save() error {
	for i := range req.histories {
		th := &req.histories[i]
		th.RequestID = req.ID
		if err := th.Save(); err != nil {
			return err
		}
	}
	val, err := req.Value()
	if err != nil {
		return err
	}
	return req.requestDB.Put(req.ID.Bytes(), val)
}

// Replace saves request with new ID and all data attached to the old one.
func (req HistoryRequest) Replace(id types.Hash) error {
	if (req.ID != types.Hash{}) {
		if err := req.Delete(); err != nil {
			return err
		}
	}
	req.ID = id
	return req.Save()
}

// Delete HistoryRequest from store and update every topic.
func (req HistoryRequest) Delete() error {
	return req.requestDB.Delete(req.ID.Bytes())
}

// Load reads request and topic histories content from disk and unmarshalls them.
func (req *HistoryRequest) Load() error {
	val, err := req.requestDB.Get(req.ID.Bytes())
	if err != nil {
		return err
	}
	return req.RawUnmarshall(val)
}

func (req *HistoryRequest) loadHistories() error {
	for _, hk := range req.TopicHistoryKeys {
		th, err := LoadTopicHistoryFromKey(req.topicDB, hk)
		if err != nil {
			return err
		}
		req.histories = append(req.histories, th)
	}
	return nil
}

// RawUnmarshall unmarshall given bytes into the structure.
// Used in range queries to unmarshall content of the iter.Value directly into request struct.
func (req *HistoryRequest) RawUnmarshall(val []byte) error {
	err := json.Unmarshal(val, req)
	if err != nil {
		return err
	}
	return req.loadHistories()
}

// Includes checks if TopicHistory is included into the request.
func (req *HistoryRequest) Includes(history TopicHistory) bool {
	key := history.Key()
	for i := range req.TopicHistoryKeys {
		if key == req.TopicHistoryKeys[i] {
			return true
		}
	}
	return false
}
