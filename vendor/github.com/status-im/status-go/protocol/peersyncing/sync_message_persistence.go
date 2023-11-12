package peersyncing

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
)

type SyncMessagePersistence interface {
	Add(SyncMessage) error
	All() ([]SyncMessage, error)
	Complement([]SyncMessage) ([]SyncMessage, error)
	ByGroupID([]byte, int) ([]SyncMessage, error)
	ByGroupIDs([][]byte, int) ([]SyncMessage, error)
	ByMessageIDs([][]byte) ([]SyncMessage, error)
}

type SyncMessageSQLitePersistence struct {
	db *sql.DB
}

func NewSyncMessageSQLitePersistence(db *sql.DB) *SyncMessageSQLitePersistence {
	return &SyncMessageSQLitePersistence{db: db}
}

func (p *SyncMessageSQLitePersistence) Add(message SyncMessage) error {
	if err := message.Valid(); err != nil {
		return err
	}
	_, err := p.db.Exec(`INSERT INTO peersyncing_messages (id, type, group_id, payload, timestamp) VALUES (?, ?, ?, ?, ?)`, message.ID, message.Type, message.GroupID, message.Payload, message.Timestamp)
	return err
}

func (p *SyncMessageSQLitePersistence) All() ([]SyncMessage, error) {
	var messages []SyncMessage
	rows, err := p.db.Query(`SELECT id, type, group_id, payload, timestamp FROM peersyncing_messages`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var m SyncMessage

		err := rows.Scan(&m.ID, &m.Type, &m.GroupID, &m.Payload, &m.Timestamp)
		if err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}
	return messages, nil
}

func (p *SyncMessageSQLitePersistence) ByGroupID(groupID []byte, limit int) ([]SyncMessage, error) {
	var messages []SyncMessage
	rows, err := p.db.Query(`SELECT id, type, group_id, payload, timestamp FROM peersyncing_messages WHERE group_id = ? ORDER BY timestamp DESC LIMIT ?`, groupID, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var m SyncMessage

		err := rows.Scan(&m.ID, &m.Type, &m.GroupID, &m.Payload, &m.Timestamp)
		if err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}
	return messages, nil
}

func (p *SyncMessageSQLitePersistence) Complement(messages []SyncMessage) ([]SyncMessage, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	ids := make([]interface{}, 0, len(messages))
	for _, m := range messages {
		ids = append(ids, m.ID)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	query := "SELECT id, type, group_id, payload, timestamp FROM peersyncing_messages WHERE id IN (" + inVector + ")" // nolint: gosec

	availableMessages := make(map[string]SyncMessage)
	rows, err := p.db.Query(query, ids...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var m SyncMessage

		err := rows.Scan(&m.ID, &m.Type, &m.GroupID, &m.Payload, &m.Timestamp)
		if err != nil {
			return nil, err
		}

		fmt.Printf("GOT MESSAGE: %x\n", m.ID)
		availableMessages[hex.EncodeToString(m.ID)] = m
	}

	var complement []SyncMessage
	for _, m := range messages {
		fmt.Printf("CHECKING MESSAGE: %x\n", m.ID)
		if _, ok := availableMessages[hex.EncodeToString(m.ID)]; !ok {
			complement = append(complement, m)
		}
	}

	return complement, nil
}

func (p *SyncMessageSQLitePersistence) ByGroupIDs(ids [][]byte, limit int) ([]SyncMessage, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	queryArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		queryArgs = append(queryArgs, id)
	}
	queryArgs = append(queryArgs, limit)

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	query := "SELECT id, type, group_id, payload, timestamp FROM peersyncing_messages WHERE group_id IN (" + inVector + ") ORDER BY timestamp DESC LIMIT ?" // nolint: gosec

	var messages []SyncMessage
	rows, err := p.db.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var m SyncMessage

		err := rows.Scan(&m.ID, &m.Type, &m.GroupID, &m.Payload, &m.Timestamp)
		if err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}
	return messages, nil

}

func (p *SyncMessageSQLitePersistence) ByMessageIDs(ids [][]byte) ([]SyncMessage, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	queryArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		queryArgs = append(queryArgs, id)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	query := "SELECT id, type, group_id, payload, timestamp FROM peersyncing_messages WHERE id IN (" + inVector + ")" // nolint: gosec

	var messages []SyncMessage
	rows, err := p.db.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var m SyncMessage

		err := rows.Scan(&m.ID, &m.Type, &m.GroupID, &m.Payload, &m.Timestamp)
		if err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}
	return messages, nil

}
