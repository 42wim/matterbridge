package transport

import (
	"database/sql"
	"fmt"
)

type SqlitePersistence struct {
	db        *sql.DB
	tableName string
}

func newSQLitePersistence(db *sql.DB, tableName string) *SqlitePersistence {
	return &SqlitePersistence{db: db, tableName: tableName}
}

func (s *SqlitePersistence) Add(chatID string, key []byte) error {
	// tableName controlled by us
	statement := fmt.Sprintf("INSERT INTO %s(chat_id, key) VALUES(?, ?)", s.tableName) // nolint:gosec
	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(chatID, key)
	return err
}

func (s *SqlitePersistence) All() (map[string][]byte, error) {
	keys := make(map[string][]byte)

	// tableName controlled by us
	statement := fmt.Sprintf("SELECT chat_id, key FROM %s", s.tableName) // nolint: gosec

	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			chatID string
			key    []byte
		)

		err := rows.Scan(&chatID, &key)
		if err != nil {
			return nil, err
		}
		keys[chatID] = key
	}

	return keys, nil
}
