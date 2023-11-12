package transport

import (
	"context"
	"database/sql"
	"strings"
)

type ProcessedMessageIDsCache struct {
	db *sql.DB
}

func NewProcessedMessageIDsCache(db *sql.DB) *ProcessedMessageIDsCache {
	return &ProcessedMessageIDsCache{db: db}
}

func (c *ProcessedMessageIDsCache) Clear() error {
	_, err := c.db.Exec("DELETE FROM transport_message_cache")
	return err
}

func (c *ProcessedMessageIDsCache) Hits(ids []string) (map[string]bool, error) {
	hits := make(map[string]bool)

	// Split the results into batches of 999 items.
	// To prevent excessive memory allocations, the maximum value of a host parameter number
	// is SQLITE_MAX_VARIABLE_NUMBER, which defaults to 999
	batch := 999
	for i := 0; i < len(ids); i += batch {
		j := i + batch
		if j > len(ids) {
			j = len(ids)
		}

		currentBatch := ids[i:j]

		idsArgs := make([]interface{}, 0, len(currentBatch))
		for _, id := range currentBatch {
			idsArgs = append(idsArgs, id)
		}

		inVector := strings.Repeat("?, ", len(currentBatch)-1) + "?"
		query := "SELECT id FROM transport_message_cache WHERE id IN (" + inVector + ")" // nolint: gosec

		rows, err := c.db.Query(query, idsArgs...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			err := rows.Scan(&id)
			if err != nil {
				return nil, err
			}
			hits[id] = true
		}
	}

	return hits, nil
}

func (c *ProcessedMessageIDsCache) Add(ids []string, timestamp uint64) (err error) {
	var tx *sql.Tx
	tx, err = c.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	for _, id := range ids {

		var stmt *sql.Stmt
		stmt, err = tx.Prepare(`INSERT INTO transport_message_cache(id,timestamp) VALUES (?, ?)`)
		if err != nil {
			return
		}

		_, err = stmt.Exec(id, timestamp)
		if err != nil {
			return
		}
	}

	return
}

func (c *ProcessedMessageIDsCache) Clean(timestamp uint64) error {
	_, err := c.db.Exec(`DELETE FROM transport_message_cache WHERE timestamp < ?`, timestamp)
	return err
}
