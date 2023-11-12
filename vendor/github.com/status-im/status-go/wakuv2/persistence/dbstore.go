package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	gowakuPersistence "github.com/waku-org/go-waku/waku/persistence"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	storepb "github.com/waku-org/go-waku/waku/v2/protocol/store/pb"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"github.com/waku-org/go-waku/waku/v2/utils"

	"go.uber.org/zap"
)

var ErrInvalidCursor = errors.New("invalid cursor")

var ErrFutureMessage = errors.New("message timestamp in the future")
var ErrMessageTooOld = errors.New("message too old")

// MaxTimeVariance is the maximum duration in the future allowed for a message timestamp
const MaxTimeVariance = time.Duration(20) * time.Second

// DBStore is a MessageProvider that has a *sql.DB connection
type DBStore struct {
	db  *sql.DB
	log *zap.Logger

	maxMessages int
	maxDuration time.Duration

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// DBOption is an optional setting that can be used to configure the DBStore
type DBOption func(*DBStore) error

// WithDB is a DBOption that lets you use any custom *sql.DB with a DBStore.
func WithDB(db *sql.DB) DBOption {
	return func(d *DBStore) error {
		d.db = db
		return nil
	}
}

// WithRetentionPolicy is a DBOption that specifies the max number of messages
// to be stored and duration before they're removed from the message store
func WithRetentionPolicy(maxMessages int, maxDuration time.Duration) DBOption {
	return func(d *DBStore) error {
		d.maxDuration = maxDuration
		d.maxMessages = maxMessages
		return nil
	}
}

// Creates a new DB store using the db specified via options.
// It will create a messages table if it does not exist and
// clean up records according to the retention policy used
func NewDBStore(log *zap.Logger, options ...DBOption) (*DBStore, error) {
	result := new(DBStore)
	result.log = log.Named("dbstore")

	for _, opt := range options {
		err := opt(result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (d *DBStore) Start(ctx context.Context, timesource timesource.Timesource) error {
	ctx, cancel := context.WithCancel(ctx)

	d.cancel = cancel

	err := d.cleanOlderRecords()
	if err != nil {
		return err
	}

	d.wg.Add(1)
	go d.checkForOlderRecords(ctx, 60*time.Second)

	return nil
}

func (d *DBStore) Validate(env *protocol.Envelope) error {
	n := time.Unix(0, env.Index().ReceiverTime)
	upperBound := n.Add(MaxTimeVariance)
	lowerBound := n.Add(-MaxTimeVariance)

	// Ensure that messages don't "jump" to the front of the queue with future timestamps
	if env.Message().GetTimestamp() > upperBound.UnixNano() {
		return ErrFutureMessage
	}

	if env.Message().GetTimestamp() < lowerBound.UnixNano() {
		return ErrMessageTooOld
	}

	return nil
}

func (d *DBStore) cleanOlderRecords() error {
	d.log.Debug("Cleaning older records...")

	// Delete older messages
	if d.maxDuration > 0 {
		start := time.Now()
		sqlStmt := `DELETE FROM store_messages WHERE receiverTimestamp < ?`
		_, err := d.db.Exec(sqlStmt, utils.GetUnixEpochFrom(time.Now().Add(-d.maxDuration)))
		if err != nil {
			return err
		}
		elapsed := time.Since(start)
		d.log.Debug("deleting older records from the DB", zap.Duration("duration", elapsed))
	}

	// Limit number of records to a max N
	if d.maxMessages > 0 {
		start := time.Now()
		sqlStmt := `DELETE FROM store_messages WHERE id IN (SELECT id FROM store_messages ORDER BY receiverTimestamp DESC LIMIT -1 OFFSET ?)`
		_, err := d.db.Exec(sqlStmt, d.maxMessages)
		if err != nil {
			return err
		}
		elapsed := time.Since(start)
		d.log.Debug("deleting excess records from the DB", zap.Duration("duration", elapsed))
	}

	return nil
}

func (d *DBStore) checkForOlderRecords(ctx context.Context, t time.Duration) {
	defer d.wg.Done()

	ticker := time.NewTicker(t)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := d.cleanOlderRecords()
			if err != nil {
				d.log.Error("cleaning older records", zap.Error(err))
			}
		}
	}
}

// Stop closes a DB connection
func (d *DBStore) Stop() {
	if d.cancel == nil {
		return
	}

	d.cancel()
	d.wg.Wait()
	d.db.Close()
}

// Put inserts a WakuMessage into the DB
func (d *DBStore) Put(env *protocol.Envelope) error {
	stmt, err := d.db.Prepare("INSERT INTO store_messages (id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	cursor := env.Index()
	dbKey := NewDBKey(uint64(cursor.SenderTime), uint64(env.Index().ReceiverTime), env.PubsubTopic(), env.Index().Digest)
	_, err = stmt.Exec(dbKey.Bytes(), cursor.ReceiverTime, env.Message().Timestamp, env.Message().ContentTopic, env.PubsubTopic(), env.Message().Payload, env.Message().Version)
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}

// Query retrieves messages from the DB
func (d *DBStore) Query(query *storepb.HistoryQuery) (*storepb.Index, []gowakuPersistence.StoredMessage, error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		d.log.Info(fmt.Sprintf("Loading records from the DB took %s", elapsed))
	}()

	sqlQuery := `SELECT id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version 
					 FROM store_messages 
					 %s
					 ORDER BY senderTimestamp %s, id %s, pubsubTopic %s, receiverTimestamp %s `

	var conditions []string
	var parameters []interface{}
	paramCnt := 0

	if query.PubsubTopic != "" {
		paramCnt++
		conditions = append(conditions, fmt.Sprintf("pubsubTopic = $%d", paramCnt))
		parameters = append(parameters, query.PubsubTopic)
	}

	if len(query.ContentFilters) != 0 {
		var ctPlaceHolder []string
		for _, ct := range query.ContentFilters {
			if ct.ContentTopic != "" {
				paramCnt++
				ctPlaceHolder = append(ctPlaceHolder, fmt.Sprintf("$%d", paramCnt))
				parameters = append(parameters, ct.ContentTopic)
			}
		}
		conditions = append(conditions, "contentTopic IN ("+strings.Join(ctPlaceHolder, ", ")+")")
	}

	usesCursor := false
	if query.PagingInfo.Cursor != nil {
		usesCursor = true
		var exists bool
		cursorDBKey := NewDBKey(uint64(query.PagingInfo.Cursor.SenderTime), uint64(query.PagingInfo.Cursor.ReceiverTime), query.PagingInfo.Cursor.PubsubTopic, query.PagingInfo.Cursor.Digest)

		err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM store_messages WHERE id = $1)",
			cursorDBKey.Bytes(),
		).Scan(&exists)

		if err != nil {
			return nil, nil, err
		}

		if exists {
			eqOp := ">"
			if query.PagingInfo.Direction == storepb.PagingInfo_BACKWARD {
				eqOp = "<"
			}
			paramCnt++
			conditions = append(conditions, fmt.Sprintf("id %s $%d", eqOp, paramCnt))

			parameters = append(parameters, cursorDBKey.Bytes())
		} else {
			return nil, nil, ErrInvalidCursor
		}
	}

	if query.GetStartTime() != 0 {
		if !usesCursor || query.PagingInfo.Direction == storepb.PagingInfo_BACKWARD {
			paramCnt++
			conditions = append(conditions, fmt.Sprintf("id >= $%d", paramCnt))
			startTimeDBKey := NewDBKey(uint64(query.GetStartTime()), uint64(query.GetStartTime()), "", []byte{})
			parameters = append(parameters, startTimeDBKey.Bytes())
		}

	}

	if query.GetEndTime() != 0 {
		if !usesCursor || query.PagingInfo.Direction == storepb.PagingInfo_FORWARD {
			paramCnt++
			conditions = append(conditions, fmt.Sprintf("id <= $%d", paramCnt))
			endTimeDBKey := NewDBKey(uint64(query.GetEndTime()), uint64(query.GetEndTime()), "", []byte{})
			parameters = append(parameters, endTimeDBKey.Bytes())
		}
	}

	conditionStr := ""
	if len(conditions) != 0 {
		conditionStr = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderDirection := "ASC"
	if query.PagingInfo.Direction == storepb.PagingInfo_BACKWARD {
		orderDirection = "DESC"
	}

	paramCnt++
	sqlQuery += fmt.Sprintf("LIMIT $%d", paramCnt)
	sqlQuery = fmt.Sprintf(sqlQuery, conditionStr, orderDirection, orderDirection, orderDirection, orderDirection)

	stmt, err := d.db.Prepare(sqlQuery)
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()

	pageSize := query.PagingInfo.PageSize + 1

	parameters = append(parameters, pageSize)
	rows, err := stmt.Query(parameters...)
	if err != nil {
		return nil, nil, err
	}

	var result []gowakuPersistence.StoredMessage
	for rows.Next() {
		record, err := d.GetStoredMessage(rows)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, record)
	}
	defer rows.Close()

	var cursor *storepb.Index
	if len(result) != 0 {
		if len(result) > int(query.PagingInfo.PageSize) {
			result = result[0:query.PagingInfo.PageSize]
			lastMsgIdx := len(result) - 1
			cursor = protocol.NewEnvelope(result[lastMsgIdx].Message, result[lastMsgIdx].ReceiverTime, result[lastMsgIdx].PubsubTopic).Index()
		}
	}

	// The retrieved messages list should always be in chronological order
	if query.PagingInfo.Direction == storepb.PagingInfo_BACKWARD {
		for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
			result[i], result[j] = result[j], result[i]
		}
	}

	return cursor, result, nil
}

// MostRecentTimestamp returns an unix timestamp with the most recent senderTimestamp
// in the message table
func (d *DBStore) MostRecentTimestamp() (int64, error) {
	result := sql.NullInt64{}

	err := d.db.QueryRow(`SELECT max(senderTimestamp) FROM store_messages`).Scan(&result)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return result.Int64, nil
}

// Count returns the number of rows in the message table
func (d *DBStore) Count() (int, error) {
	var result int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM store_messages`).Scan(&result)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return result, nil
}

// GetAll returns all the stored WakuMessages
func (d *DBStore) GetAll() ([]gowakuPersistence.StoredMessage, error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		d.log.Info("loading records from the DB", zap.Duration("duration", elapsed))
	}()

	rows, err := d.db.Query("SELECT id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version FROM store_messages ORDER BY senderTimestamp ASC")
	if err != nil {
		return nil, err
	}

	var result []gowakuPersistence.StoredMessage

	defer rows.Close()

	for rows.Next() {
		record, err := d.GetStoredMessage(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}

	d.log.Info("DB returned records", zap.Int("count", len(result)))

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetStoredMessage is a helper function used to convert a `*sql.Rows` into a `StoredMessage`
func (d *DBStore) GetStoredMessage(row *sql.Rows) (gowakuPersistence.StoredMessage, error) {
	var id []byte
	var receiverTimestamp int64
	var senderTimestamp int64
	var contentTopic string
	var payload []byte
	var version uint32
	var pubsubTopic string

	err := row.Scan(&id, &receiverTimestamp, &senderTimestamp, &contentTopic, &pubsubTopic, &payload, &version)
	if err != nil {
		d.log.Error("scanning messages from db", zap.Error(err))
		return gowakuPersistence.StoredMessage{}, err
	}

	msg := new(pb.WakuMessage)
	msg.ContentTopic = contentTopic
	msg.Payload = payload
	msg.Timestamp = &senderTimestamp
	msg.Version = &version

	record := gowakuPersistence.StoredMessage{
		ID:           id,
		PubsubTopic:  pubsubTopic,
		ReceiverTime: receiverTimestamp,
		Message:      msg,
	}

	return record, nil
}
