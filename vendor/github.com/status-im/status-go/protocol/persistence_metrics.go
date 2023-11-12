package protocol

import (
	"database/sql"
	"fmt"
	"strings"
)

const selectTimestampsQuery = "SELECT whisper_timestamp FROM user_messages WHERE %s whisper_timestamp >= ? AND whisper_timestamp <= ?"
const selectCountQuery = "SELECT COUNT(*) FROM user_messages WHERE %s whisper_timestamp >= ? AND whisper_timestamp <= ?"

func querySeveralChats(chatIDs []string) string {
	if len(chatIDs) == 0 {
		return ""
	}

	var conditions []string
	for _, chatID := range chatIDs {
		conditions = append(conditions, fmt.Sprintf("local_chat_id = '%s'", chatID))
	}
	return fmt.Sprintf("(%s) AND", strings.Join(conditions, " OR "))
}

func (db sqlitePersistence) SelectMessagesTimestampsForChatsByPeriod(chatIDs []string, startTimestamp uint64, endTimestamp uint64) ([]uint64, error) {
	query := fmt.Sprintf(selectTimestampsQuery, querySeveralChats(chatIDs))

	rows, err := db.db.Query(query, startTimestamp, endTimestamp)
	if err != nil {
		return []uint64{}, err
	}
	defer rows.Close()

	var timestamps []uint64
	for rows.Next() {
		var timestamp uint64
		err := rows.Scan(&timestamp)
		if err != nil {
			return nil, err
		}
		timestamps = append(timestamps, timestamp)
	}

	err = rows.Err()
	if err != nil {
		return []uint64{}, err
	}

	return timestamps, nil
}

func (db sqlitePersistence) SelectMessagesCountForChatsByPeriod(chatIDs []string, startTimestamp uint64, endTimestamp uint64) (int, error) {
	query := fmt.Sprintf(selectCountQuery, querySeveralChats(chatIDs))

	var count int
	if err := db.db.QueryRow(query, startTimestamp, endTimestamp).Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}
