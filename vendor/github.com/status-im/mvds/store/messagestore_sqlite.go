package store

import (
	"database/sql"
	"errors"

	"github.com/status-im/mvds/state"

	"github.com/status-im/mvds/protobuf"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

type persistentMessageStore struct {
	db *sql.DB
}

func NewPersistentMessageStore(db *sql.DB) *persistentMessageStore {
	return &persistentMessageStore{db: db}
}

func (p *persistentMessageStore) Add(message *protobuf.Message) error {
	id := message.ID()
	_, err := p.db.Exec(
		`INSERT INTO mvds_messages (id, group_id, timestamp, body)
		VALUES (?, ?, ?, ?)`,
		id[:],
		message.GroupId,
		message.Timestamp,
		message.Body,
	)
	return err
}

func (p *persistentMessageStore) Get(id state.MessageID) (*protobuf.Message, error) {
	var message protobuf.Message
	row := p.db.QueryRow(
		`SELECT group_id, timestamp, body FROM mvds_messages WHERE id = ?`,
		id[:],
	)
	if err := row.Scan(
		&message.GroupId,
		&message.Timestamp,
		&message.Body,
	); err != nil {
		return nil, err
	}
	return &message, nil
}

func (p *persistentMessageStore) Has(id state.MessageID) (bool, error) {
	var result bool
	err := p.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM mvds_messages WHERE id = ?)`,
		id[:],
	).Scan(&result)
	switch err {
	case sql.ErrNoRows:
		return false, ErrMessageNotFound
	case nil:
		return result, nil
	default:
		return false, err
	}
}
