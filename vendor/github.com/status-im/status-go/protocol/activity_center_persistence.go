package protocol

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
)

const allFieldsForTableActivityCenterNotification = `id, timestamp, notification_type, chat_id, read, dismissed, accepted, message, author,
    reply_message, community_id, membership_status, contact_verification_status, token_data, deleted, updated_at`

var emptyNotifications = make([]*ActivityCenterNotification, 0)

func (db sqlitePersistence) DeleteActivityCenterNotificationByID(id []byte, updatedAt uint64) error {
	_, err := db.db.Exec(`UPDATE activity_center_notifications SET deleted = 1, updated_at = ? WHERE id = ? AND NOT deleted`, updatedAt, id)
	return err
}

func (db sqlitePersistence) DeleteActivityCenterNotificationForMessage(chatID string, messageID string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	params := activityCenterQueryParams{
		chatID: chatID,
	}

	_, notifications, err := db.buildActivityCenterQuery(tx, params)

	if err != nil {
		return nil, err
	}

	var ids []types.HexBytes
	var matchNotifications []*ActivityCenterNotification
	withNotification := func(a *ActivityCenterNotification) {
		a.Read = true
		a.Dismissed = true
		a.Deleted = true
		a.UpdatedAt = updatedAt
		ids = append(ids, a.ID)
		matchNotifications = append(matchNotifications, a)
	}

	for _, notification := range notifications {
		if notification.LastMessage != nil && notification.LastMessage.ID == messageID {
			withNotification(notification)
		}

		if notification.Message != nil && notification.Message.ID == messageID {
			withNotification(notification)
		}
	}

	if len(ids) > 0 {
		args := make([]interface{}, 0, len(ids)+1)
		args = append(args, updatedAt)
		for _, id := range ids {
			args = append(args, id)
		}

		inVector := strings.Repeat("?, ", len(ids)-1) + "?"
		query := "UPDATE activity_center_notifications SET read = 1, dismissed = 1, deleted = 1, updated_at = ? WHERE id IN (" + inVector + ")" // nolint: gosec
		_, err = tx.Exec(query, args...)
		return matchNotifications, err
	}

	return matchNotifications, nil
}

func (db sqlitePersistence) SaveActivityCenterNotification(notification *ActivityCenterNotification, updateState bool) (int64, error) {
	var tx *sql.Tx
	var err error

	err = notification.Valid()
	if err != nil {
		return 0, err
	}

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	// encode message
	var encodedMessage []byte
	if notification.Message != nil {
		encodedMessage, err = json.Marshal(notification.Message)
		if err != nil {
			return 0, err
		}
	}

	// encode message
	var encodedReplyMessage []byte
	if notification.ReplyMessage != nil {
		encodedReplyMessage, err = json.Marshal(notification.ReplyMessage)
		if err != nil {
			return 0, err
		}
	}

	// encode token data
	var encodedTokenData []byte
	if notification.TokenData != nil {
		encodedTokenData, err = json.Marshal(notification.TokenData)
		if err != nil {
			return 0, err
		}
	}

	result, err := tx.Exec(`
		INSERT OR REPLACE
		INTO activity_center_notifications (
			id,
			timestamp,
			notification_type,
			chat_id,
			community_id,
			membership_status,
			message,
			reply_message,
			author,
			contact_verification_status,
			read,
			accepted,
			dismissed,
			token_data,
			deleted,
		    updated_at
		)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		`,
		notification.ID,
		notification.Timestamp,
		notification.Type,
		notification.ChatID,
		notification.CommunityID,
		notification.MembershipStatus,
		encodedMessage,
		encodedReplyMessage,
		notification.Author,
		notification.ContactVerificationStatus,
		notification.Read,
		notification.Accepted,
		notification.Dismissed,
		encodedTokenData,
		notification.Deleted,
		notification.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return n, err
	}

	// When we have inserted or updated unread notification - mark whole activity_center_settings as unseen
	if updateState && n > 0 && !notification.Read {
		_, err = tx.Exec(`UPDATE activity_center_states SET has_seen = 0, updated_at = ?`, notification.UpdatedAt)
	}

	return n, nil
}

func (db sqlitePersistence) parseRowFromTableActivityCenterNotification(rows *sql.Rows, withNotification func(notification *ActivityCenterNotification)) ([]*ActivityCenterNotification, error) {
	var notifications []*ActivityCenterNotification
	defer rows.Close()
	for rows.Next() {
		var chatID sql.NullString
		var communityID sql.NullString
		var messageBytes []byte
		var replyMessageBytes []byte
		var tokenDataBytes []byte
		var author sql.NullString
		notification := &ActivityCenterNotification{}
		err := rows.Scan(
			&notification.ID,
			&notification.Timestamp,
			&notification.Type,
			&chatID,
			&notification.Read,
			&notification.Dismissed,
			&notification.Accepted,
			&messageBytes,
			&author,
			&replyMessageBytes,
			&communityID,
			&notification.MembershipStatus,
			&notification.ContactVerificationStatus,
			&tokenDataBytes,
			&notification.Deleted,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if chatID.Valid {
			notification.ChatID = chatID.String
		}

		if communityID.Valid {
			notification.CommunityID = communityID.String
		}

		if author.Valid {
			notification.Author = author.String
		}

		if len(tokenDataBytes) > 0 {
			err = json.Unmarshal(tokenDataBytes, &notification.TokenData)
			if err != nil {
				return nil, err
			}
		}

		if len(messageBytes) > 0 {
			err = json.Unmarshal(messageBytes, &notification.Message)
			if err != nil {
				return nil, err
			}
		}

		if len(replyMessageBytes) > 0 {
			err = json.Unmarshal(replyMessageBytes, &notification.ReplyMessage)
			if err != nil {
				return nil, err
			}
		}

		if withNotification != nil {
			withNotification(notification)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (db sqlitePersistence) unmarshalActivityCenterNotificationRow(row *sql.Row) (*ActivityCenterNotification, error) {
	var chatID sql.NullString
	var communityID sql.NullString
	var lastMessageBytes []byte
	var messageBytes []byte
	var replyMessageBytes []byte
	var tokenDataBytes []byte
	var name sql.NullString
	var author sql.NullString
	notification := &ActivityCenterNotification{}
	err := row.Scan(
		&notification.ID,
		&notification.Timestamp,
		&notification.Type,
		&chatID,
		&communityID,
		&notification.MembershipStatus,
		&notification.Read,
		&notification.Accepted,
		&notification.Dismissed,
		&notification.Deleted,
		&messageBytes,
		&lastMessageBytes,
		&replyMessageBytes,
		&notification.ContactVerificationStatus,
		&name,
		&author,
		&tokenDataBytes,
		&notification.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if chatID.Valid {
		notification.ChatID = chatID.String
	}

	if communityID.Valid {
		notification.CommunityID = communityID.String
	}

	if name.Valid {
		notification.Name = name.String
	}

	if author.Valid {
		notification.Author = author.String
	}

	if len(tokenDataBytes) > 0 {
		err = json.Unmarshal(tokenDataBytes, &notification.TokenData)
		if err != nil {
			return nil, err
		}
	}

	// Restore last message
	if lastMessageBytes != nil {
		lastMessage := common.NewMessage()
		if err = json.Unmarshal(lastMessageBytes, lastMessage); err != nil {
			return nil, err
		}
		notification.LastMessage = lastMessage
	}

	// Restore message
	if messageBytes != nil {
		message := common.NewMessage()
		if err = json.Unmarshal(messageBytes, message); err != nil {
			return nil, err
		}
		notification.Message = message
	}

	// Restore reply message
	if replyMessageBytes != nil {
		replyMessage := common.NewMessage()
		if err = json.Unmarshal(replyMessageBytes, replyMessage); err != nil {
			return nil, err
		}
		notification.ReplyMessage = replyMessage
	}

	return notification, nil
}

func (db sqlitePersistence) unmarshalActivityCenterNotificationRows(rows *sql.Rows) (string, []*ActivityCenterNotification, error) {
	var notifications []*ActivityCenterNotification
	latestCursor := ""
	for rows.Next() {
		var chatID sql.NullString
		var communityID sql.NullString
		var lastMessageBytes []byte
		var messageBytes []byte
		var replyMessageBytes []byte
		var tokenDataBytes []byte
		var name sql.NullString
		var author sql.NullString
		notification := &ActivityCenterNotification{}
		err := rows.Scan(
			&notification.ID,
			&notification.Timestamp,
			&notification.Type,
			&chatID,
			&communityID,
			&notification.MembershipStatus,
			&notification.Read,
			&notification.Accepted,
			&notification.Dismissed,
			&messageBytes,
			&lastMessageBytes,
			&replyMessageBytes,
			&notification.ContactVerificationStatus,
			&name,
			&author,
			&tokenDataBytes,
			&latestCursor,
			&notification.UpdatedAt)
		if err != nil {
			return "", nil, err
		}

		if chatID.Valid {
			notification.ChatID = chatID.String
		}

		if communityID.Valid {
			notification.CommunityID = communityID.String
		}

		if name.Valid {
			notification.Name = name.String
		}

		if author.Valid {
			notification.Author = author.String
		}

		if len(tokenDataBytes) > 0 {
			tokenData := &ActivityTokenData{}
			if err = json.Unmarshal(tokenDataBytes, &tokenData); err != nil {
				return "", nil, err
			}
			notification.TokenData = tokenData
		}

		// Restore last message
		if lastMessageBytes != nil {
			lastMessage := common.NewMessage()
			if err = json.Unmarshal(lastMessageBytes, lastMessage); err != nil {
				return "", nil, err
			}
			notification.LastMessage = lastMessage
		}

		// Restore message
		if messageBytes != nil {
			message := common.NewMessage()
			if err = json.Unmarshal(messageBytes, message); err != nil {
				return "", nil, err
			}
			notification.Message = message
		}

		// Restore reply message
		if replyMessageBytes != nil {
			replyMessage := common.NewMessage()
			if err = json.Unmarshal(replyMessageBytes, replyMessage); err != nil {
				return "", nil, err
			}
			notification.ReplyMessage = replyMessage
		}

		notifications = append(notifications, notification)
	}

	return latestCursor, notifications, nil

}

type activityCenterQueryParams struct {
	cursor              string
	limit               uint64
	ids                 []types.HexBytes
	chatID              string
	author              string
	read                ActivityCenterQueryParamsRead
	accepted            bool
	activityCenterTypes []ActivityCenterType
}

func (db sqlitePersistence) prepareQueryConditionsAndArgs(params activityCenterQueryParams) ([]interface{}, string) {
	var args []interface{}
	var conditions []string

	cursor := params.cursor
	ids := params.ids
	author := params.author
	activityCenterTypes := params.activityCenterTypes
	chatID := params.chatID
	read := params.read
	accepted := params.accepted

	if cursor != "" {
		conditions = append(conditions, "cursor <= ?")
		args = append(args, cursor)
	}

	if len(ids) != 0 {
		inVector := strings.Repeat("?, ", len(ids)-1) + "?"
		conditions = append(conditions, fmt.Sprintf("a.id IN (%s)", inVector))
		for _, id := range ids {
			args = append(args, id)
		}
	}

	switch read {
	case ActivityCenterQueryParamsReadRead:
		conditions = append(conditions, "a.read = 1")
	case ActivityCenterQueryParamsReadUnread:
		conditions = append(conditions, "NOT a.read")
	}

	if !accepted {
		conditions = append(conditions, "NOT a.accepted")
	}

	if chatID != "" {
		conditions = append(conditions, "a.chat_id = ?")
		args = append(args, chatID)
	}

	if author != "" {
		conditions = append(conditions, "a.author = ?")
		args = append(args, author)
	}

	if len(activityCenterTypes) > 0 {
		inVector := strings.Repeat("?, ", len(activityCenterTypes)-1) + "?"
		conditions = append(conditions, fmt.Sprintf("a.notification_type IN (%s)", inVector))
		for _, activityCenterType := range activityCenterTypes {
			args = append(args, activityCenterType)
		}
	}

	conditions = append(conditions, "NOT a.deleted")

	var conditionsString string
	if len(conditions) > 0 {
		conditionsString = " WHERE " + strings.Join(conditions, " AND ")
	}

	return args, conditionsString
}

func (db sqlitePersistence) buildActivityCenterQuery(tx *sql.Tx, params activityCenterQueryParams) (string, []*ActivityCenterNotification, error) {
	args, conditionsString := db.prepareQueryConditionsAndArgs(params)

	query := fmt.Sprintf( // nolint: gosec
		`
	SELECT
	a.id,
	a.timestamp,
	a.notification_type,
	a.chat_id,
	a.community_id,
	a.membership_status,
	a.read,
	a.accepted,
	a.dismissed,
	a.message,
	c.last_message,
	a.reply_message,
	a.contact_verification_status,
	c.name,
	a.author,
	a.token_data,
	substr('0000000000000000000000000000000000000000000000000000000000000000' || a.timestamp, -64, 64) || hex(a.id) as cursor,
	a.updated_at
	FROM activity_center_notifications a
	LEFT JOIN chats c
	ON
	c.id = a.chat_id
	%s
	ORDER BY cursor DESC`, conditionsString)

	if params.limit != 0 {
		args = append(args, params.limit)
		query += ` LIMIT ?`
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()

	return db.unmarshalActivityCenterNotificationRows(rows)
}

func (db sqlitePersistence) buildActivityCenterNotificationsCountQuery(isAccepted bool, read ActivityCenterQueryParamsRead, activityCenterTypes []ActivityCenterType) *sql.Row {
	params := activityCenterQueryParams{
		accepted:            isAccepted,
		read:                read,
		activityCenterTypes: activityCenterTypes,
	}

	args, conditionsString := db.prepareQueryConditionsAndArgs(params)
	query := fmt.Sprintf(`SELECT COUNT(1) FROM activity_center_notifications a %s`, conditionsString)

	return db.db.QueryRow(query, args...)
}

func (db sqlitePersistence) runActivityCenterIDQuery(query string) ([][]byte, error) {
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}

	var ids [][]byte

	for rows.Next() {
		var id []byte
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (db sqlitePersistence) GetNotReadActivityCenterNotificationIds() ([][]byte, error) {
	return db.runActivityCenterIDQuery("SELECT a.id FROM activity_center_notifications a WHERE NOT a.read AND NOT a.deleted")
}

func (db sqlitePersistence) GetToProcessActivityCenterNotificationIds() ([][]byte, error) {
	return db.runActivityCenterIDQuery(`
		SELECT a.id
		FROM activity_center_notifications a
		WHERE NOT a.dismissed AND NOT a.accepted AND NOT a.deleted
		`)
}

func (db sqlitePersistence) HasPendingNotificationsForChat(chatID string) (bool, error) {
	rows, err := db.db.Query(`
		SELECT 1 FROM activity_center_notifications a
		WHERE a.chat_id = ?
			AND NOT a.deleted
			AND NOT a.dismissed
			AND NOT a.accepted
		`, chatID)

	if err != nil {
		return false, err
	}

	result := false
	if rows.Next() {
		result = true
		rows.Close()
	}

	err = rows.Err()
	return result, err
}

func (db sqlitePersistence) GetActivityCenterNotificationsByID(ids []types.HexBytes) ([]*ActivityCenterNotification, error) {
	if len(ids) == 0 {
		return emptyNotifications, nil
	}
	idsArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsArgs = append(idsArgs, id)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	// nolint: gosec
	rows, err := db.db.Query(
		`
		SELECT
		a.id,
		a.timestamp,
		a.notification_type,
		a.chat_id,
		a.community_id,
		a.membership_status,
		a.read,
		a.accepted,
		a.dismissed,
		a.message,
		c.last_message,
		a.reply_message,
		a.contact_verification_status,
		c.name,
		a.author,
		a.token_data,
		substr('0000000000000000000000000000000000000000000000000000000000000000' || a.timestamp, -64, 64) || hex(a.id) as cursor,
		a.updated_at
		FROM activity_center_notifications a
		LEFT JOIN chats c
		ON
		c.id = a.chat_id
		WHERE a.id IN (`+inVector+`) AND NOT a.deleted`, idsArgs...)

	if err != nil {
		return nil, err
	}

	_, notifications, err := db.unmarshalActivityCenterNotificationRows(rows)
	if err != nil {
		return nil, nil
	}

	return notifications, nil
}

// GetActivityCenterNotificationByID returns a notification by its ID even it's deleted logically
func (db sqlitePersistence) GetActivityCenterNotificationByID(id types.HexBytes) (*ActivityCenterNotification, error) {
	row := db.db.QueryRow(`
		SELECT
		a.id,
		a.timestamp,
		a.notification_type,
		a.chat_id,
		a.community_id,
		a.membership_status,
		a.read,
		a.accepted,
		a.dismissed,
		a.deleted,
		a.message,
		c.last_message,
		a.reply_message,
		a.contact_verification_status,
		c.name,
		a.author,
		a.token_data,
		a.updated_at
		FROM activity_center_notifications a
		LEFT JOIN chats c
		ON
		c.id = a.chat_id
		WHERE a.id = ?`, id)

	notification, err := db.unmarshalActivityCenterNotificationRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return notification, err
}

func (db sqlitePersistence) activityCenterNotifications(params activityCenterQueryParams) (string, []*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error
	// We fetch limit + 1 to check for pagination
	nonIncrementedLimit := params.limit
	incrementedLimit := int(params.limit) + 1
	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	params.limit = uint64(incrementedLimit)
	latestCursor, notifications, err := db.buildActivityCenterQuery(tx, params)
	if err != nil {
		return "", nil, err
	}

	if len(notifications) == incrementedLimit {
		notifications = notifications[0:nonIncrementedLimit]
	} else {
		latestCursor = ""
	}

	return latestCursor, notifications, nil
}

func (db sqlitePersistence) DismissAllActivityCenterNotifications(updatedAt uint64) error {
	_, err := db.db.Exec(`UPDATE activity_center_notifications SET read = 1, dismissed = 1, updated_at = ? WHERE NOT dismissed AND NOT accepted AND NOT deleted`, updatedAt)
	return err
}

func (db sqlitePersistence) DismissAllActivityCenterNotificationsFromUser(userPublicKey string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	query := fmt.Sprintf(`SELECT %s FROM activity_center_notifications WHERE
                                       author = ? AND
                                       NOT deleted AND
                                       NOT dismissed AND
                                       NOT accepted`, allFieldsForTableActivityCenterNotification)
	rows, err := tx.Query(query, userPublicKey)
	if err != nil {
		return nil, err
	}
	var notifications []*ActivityCenterNotification
	notifications, err = db.parseRowFromTableActivityCenterNotification(rows, func(notification *ActivityCenterNotification) {
		notification.Read = true
		notification.Dismissed = true
		notification.UpdatedAt = updatedAt
	})
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE activity_center_notifications
		SET read = 1, dismissed = 1, updated_at = ?
		WHERE author = ?
			AND NOT deleted
			AND NOT dismissed
			AND NOT accepted
		`,
		updatedAt, userPublicKey)
	if err != nil {
		return nil, err
	}

	return notifications, updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) MarkActivityCenterNotificationsDeleted(ids []types.HexBytes, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	if len(ids) == 0 {
		return emptyNotifications, nil
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, updatedAt)
	for _, id := range ids {
		args = append(args, id)
	}

	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	// nolint: gosec
	query := fmt.Sprintf(`SELECT %s FROM activity_center_notifications WHERE id IN (%s) AND NOT deleted`,
		allFieldsForTableActivityCenterNotification,
		inVector)
	rows, err := tx.Query(query, args[1:]...)
	if err != nil {
		return nil, err
	}
	notifications, err := db.parseRowFromTableActivityCenterNotification(rows, func(notification *ActivityCenterNotification) {
		notification.Deleted = true
		notification.UpdatedAt = updatedAt
	})
	if err != nil {
		return nil, err
	}

	update := "UPDATE activity_center_notifications SET deleted = 1, updated_at = ? WHERE id IN (" + inVector + ") AND NOT deleted"
	_, err = tx.Exec(update, args...)
	if err != nil {
		return nil, err
	}

	return notifications, updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) DismissActivityCenterNotifications(ids []types.HexBytes, updatedAt uint64) error {
	if len(ids) == 0 {
		return nil
	}

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, updatedAt)
	for _, id := range ids {
		args = append(args, id)
	}

	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	query := "UPDATE activity_center_notifications SET read = 1, dismissed = 1, updated_at = ? WHERE id IN (" + inVector + ") AND not deleted" // nolint: gosec
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}
	return updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) DismissActivityCenterNotificationsByCommunity(communityID string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	query := "UPDATE activity_center_notifications SET read = 1, dismissed = 1, updated_at = ? WHERE community_id = ? AND notification_type IN (?, ?, ?, ?) AND NOT deleted" // nolint: gosec
	_, err = tx.Exec(query, updatedAt, communityID,
		ActivityCenterNotificationTypeCommunityRequest, ActivityCenterNotificationTypeCommunityKicked, ActivityCenterNotificationTypeCommunityBanned, ActivityCenterNotificationTypeCommunityUnbanned)
	if err != nil {
		return nil, err
	}

	_, notifications, err := db.buildActivityCenterQuery(tx, activityCenterQueryParams{})
	if err != nil {
		return nil, err
	}

	return notifications, updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) DismissAllActivityCenterNotificationsFromCommunity(communityID string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	chatIDs, err := db.AllChatIDsByCommunity(tx, communityID)
	if err != nil {
		return nil, err
	}

	chatIDsCount := len(chatIDs)
	if chatIDsCount == 0 {
		return nil, nil
	}

	args := make([]interface{}, 0, chatIDsCount+1)
	args = append(args, updatedAt)
	for _, chatID := range chatIDs {
		args = append(args, chatID)
	}

	inVector := strings.Repeat("?, ", chatIDsCount-1) + "?"

	// nolint: gosec
	query := fmt.Sprintf(`SELECT %s FROM activity_center_notifications WHERE chat_id IN (%s) AND NOT deleted`, allFieldsForTableActivityCenterNotification, inVector)
	rows, err := tx.Query(query, args[1:]...)
	if err != nil {
		return nil, err
	}
	notifications, err := db.parseRowFromTableActivityCenterNotification(rows, func(notification *ActivityCenterNotification) {
		notification.Read = true
		notification.Dismissed = true
		notification.UpdatedAt = updatedAt
	})
	if err != nil {
		return nil, err
	}

	query = "UPDATE activity_center_notifications SET read = 1, dismissed = 1, updated_at = ? WHERE chat_id IN (" + inVector + ") AND NOT deleted" // nolint: gosec
	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return notifications, updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) DismissAllActivityCenterNotificationsFromChatID(chatID string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	query := fmt.Sprintf(`SELECT %s FROM activity_center_notifications
          WHERE chat_id = ?
          AND NOT deleted
          AND NOT accepted
          AND notification_type != ?`, allFieldsForTableActivityCenterNotification)
	rows, err := tx.Query(query, chatID, ActivityCenterNotificationTypeContactRequest)
	if err != nil {
		return nil, err
	}
	notifications, err := db.parseRowFromTableActivityCenterNotification(rows, func(notification *ActivityCenterNotification) {
		notification.Read = true
		notification.Dismissed = true
		notification.UpdatedAt = updatedAt
	})
	if err != nil {
		return nil, err
	}

	// We exclude notifications related to contacts, since those we don't want to
	// be cleared.
	query = `
		UPDATE activity_center_notifications
		SET read = 1, dismissed = 1, updated_at = ?
		WHERE chat_id = ?
		    AND NOT deleted
			AND NOT accepted
			AND notification_type != ?
	`
	_, err = tx.Exec(query, updatedAt, chatID, ActivityCenterNotificationTypeContactRequest)
	if err != nil {
		return nil, err
	}
	return notifications, updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) AcceptAllActivityCenterNotifications(updatedAt uint64) ([]*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	_, notifications, err := db.buildActivityCenterQuery(tx, activityCenterQueryParams{})
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`UPDATE activity_center_notifications SET read = 1, accepted = 1, updated_at = ? WHERE NOT dismissed AND NOT accepted AND NOT deleted`, updatedAt)
	return notifications, err
}

func (db sqlitePersistence) AcceptActivityCenterNotifications(ids []types.HexBytes, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	if len(ids) == 0 {
		return emptyNotifications, nil
	}
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, updatedAt)
	for _, id := range ids {
		args = append(args, id)
	}

	params := activityCenterQueryParams{
		ids: ids,
	}
	_, notifications, err := db.buildActivityCenterQuery(tx, params)
	if err != nil {
		return nil, err
	}
	var updateNotifications []*ActivityCenterNotification
	for _, n := range notifications {
		n.Read = true
		n.Accepted = true
		n.UpdatedAt = updatedAt
		updateNotifications = append(updateNotifications, n)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	query := "UPDATE activity_center_notifications SET read = 1, accepted = 1, updated_at = ? WHERE id IN (" + inVector + ") AND NOT deleted" // nolint: gosec
	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return updateNotifications, updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) AcceptActivityCenterNotificationsForInvitesFromUser(userPublicKey string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	query := fmt.Sprintf(`SELECT %s FROM activity_center_notifications
          WHERE author = ?
          AND NOT deleted
          AND NOT dismissed
          AND NOT accepted
          AND notification_type = ?`, allFieldsForTableActivityCenterNotification)
	rows, err := tx.Query(query, userPublicKey, ActivityCenterNotificationTypeNewPrivateGroupChat)
	if err != nil {
		return nil, err
	}
	notifications, err := db.parseRowFromTableActivityCenterNotification(rows, func(notification *ActivityCenterNotification) {
		notification.Read = true
		notification.Accepted = true
		notification.UpdatedAt = updatedAt
	})
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE activity_center_notifications
		SET read = 1, accepted = 1, updated_at = ?
		WHERE author = ?
			AND NOT deleted
			AND NOT dismissed
		    AND NOT accepted
			AND notification_type = ?
		`,
		updatedAt, userPublicKey, ActivityCenterNotificationTypeNewPrivateGroupChat)

	if err != nil {
		return nil, err
	}

	return notifications, updateActivityCenterState(tx, updatedAt)
}

func (db sqlitePersistence) MarkAllActivityCenterNotificationsRead(updatedAt uint64) error {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()
	_, err = tx.Exec(`UPDATE activity_center_notifications SET read = 1, updated_at = ? WHERE NOT read AND NOT deleted`, updatedAt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE activity_center_states SET has_seen = 1, updated_at = ?`, updatedAt)
	return err
}

func (db sqlitePersistence) MarkActivityCenterNotificationsRead(ids []types.HexBytes, updatedAt uint64) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, updatedAt)
	for _, id := range ids {
		args = append(args, id)
	}

	var tx *sql.Tx
	var err error

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	query := "UPDATE activity_center_notifications SET read = 1, updated_at = ? WHERE id IN (" + inVector + ") AND NOT deleted" // nolint: gosec
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}
	return updateActivityCenterState(tx, updatedAt)

}

func updateActivityCenterState(tx *sql.Tx, updatedAt uint64) error {
	var unreadCount int
	err := tx.QueryRow("SELECT COUNT(1) FROM activity_center_notifications WHERE read = 0 AND deleted = 0").Scan(&unreadCount)
	if err != nil {
		return err
	}
	var hasSeen int
	if unreadCount == 0 {
		hasSeen = 1
	}

	_, err = tx.Exec(`UPDATE activity_center_states SET has_seen = ?, updated_at = ?`, hasSeen, updatedAt)
	return err
}

func (db sqlitePersistence) MarkActivityCenterNotificationsUnread(ids []types.HexBytes, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	if len(ids) == 0 {
		return emptyNotifications, nil
	}
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, updatedAt)
	for _, id := range ids {
		args = append(args, id)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	// nolint: gosec
	query := fmt.Sprintf("SELECT %s FROM activity_center_notifications WHERE id IN (%s) AND NOT deleted", allFieldsForTableActivityCenterNotification, inVector)

	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	rows, err := tx.Query(query, args[1:]...)
	if err != nil {
		return nil, err
	}
	notifications, err := db.parseRowFromTableActivityCenterNotification(rows, func(notification *ActivityCenterNotification) {
		notification.Read = false
		notification.UpdatedAt = updatedAt
	})
	if err != nil {
		return nil, err
	}

	if len(notifications) == 0 {
		return notifications, nil
	}

	query = "UPDATE activity_center_notifications SET read = 0, updated_at = ? WHERE id IN (" + inVector + ") AND NOT deleted" // nolint: gosec
	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`UPDATE activity_center_states SET has_seen = 0, updated_at = ?`, updatedAt)
	return notifications, err
}

func (db sqlitePersistence) ActivityCenterNotifications(cursor string, limit uint64, activityTypes []ActivityCenterType, readType ActivityCenterQueryParamsRead, accepted bool) (string, []*ActivityCenterNotification, error) {
	params := activityCenterQueryParams{
		activityCenterTypes: activityTypes,
		cursor:              cursor,
		limit:               limit,
		read:                readType,
		accepted:            accepted,
	}

	return db.activityCenterNotifications(params)
}

func (db sqlitePersistence) ActivityCenterNotificationsCount(activityTypes []ActivityCenterType, readType ActivityCenterQueryParamsRead, accepted bool) (uint64, error) {
	var count uint64
	err := db.buildActivityCenterNotificationsCountQuery(accepted, readType, activityTypes).Scan(&count)
	return count, err
}

func (db sqlitePersistence) ActiveContactRequestNotification(contactID string) (*ActivityCenterNotification, error) {
	// QueryRow expects a query that returns at most one row. In theory the query
	// wouldn't even need the ORDER + LIMIT 1 because we expect only one active
	// contact request per contact, but to avoid relying on the unpredictable
	// behavior of the DB engine for sorting, we sort by notification.Timestamp
	// DESC.
	query := `
		SELECT
			a.id,
			a.timestamp,
			a.notification_type,
			a.chat_id,
			a.community_id,
			a.membership_status,
			a.read,
			a.accepted,
			a.dismissed,
			a.deleted,
			a.message,
			c.last_message,
			a.reply_message,
			a.contact_verification_status,
			c.name,
			a.author,
			a.token_data,
			a.updated_at
		FROM activity_center_notifications a
		LEFT JOIN chats c ON c.id = a.chat_id
		WHERE a.author = ?
		    AND NOT a.deleted
			AND NOT a.dismissed
			AND NOT a.accepted
			AND a.notification_type = ?
		ORDER BY a.timestamp DESC
		LIMIT 1
		`
	row := db.db.QueryRow(query, contactID, ActivityCenterNotificationTypeContactRequest)
	notification, err := db.unmarshalActivityCenterNotificationRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return notification, err
}

func (db sqlitePersistence) DeleteChatContactRequestActivityCenterNotifications(chatID string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	query := fmt.Sprintf(`SELECT %s FROM activity_center_notifications WHERE chat_id = ? AND NOT deleted AND notification_type = ?`, allFieldsForTableActivityCenterNotification)
	rows, err := db.db.Query(query, chatID, ActivityCenterNotificationTypeContactRequest)
	if err != nil {
		return nil, err
	}
	notifications, err := db.parseRowFromTableActivityCenterNotification(rows, func(notification *ActivityCenterNotification) {
		notification.Deleted = true
		notification.UpdatedAt = updatedAt
	})
	if err != nil {
		return nil, err
	}

	_, err = db.db.Exec(`
				UPDATE activity_center_notifications SET deleted = 1, updated_at = ?
	WHERE
	chat_id = ?
	AND NOT deleted
	AND notification_type = ?
	`, updatedAt, chatID, ActivityCenterNotificationTypeContactRequest)
	return notifications, err
}

func (db sqlitePersistence) HasUnseenActivityCenterNotifications() (bool, uint64, error) {
	row := db.db.QueryRow(`SELECT has_seen, updated_at FROM activity_center_states`)
	hasSeen := true
	updatedAt := uint64(0)
	err := row.Scan(&hasSeen, &updatedAt)
	return !hasSeen, updatedAt, err
}

func (db sqlitePersistence) UpdateActivityCenterNotificationState(state *ActivityCenterState) (int64, error) {
	result, err := db.db.Exec(`UPDATE activity_center_states SET has_seen = ?, updated_at = ?`, state.HasSeen, state.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db sqlitePersistence) GetActivityCenterState() (*ActivityCenterState, error) {
	unseen, updatedAt, err := db.HasUnseenActivityCenterNotifications()
	if err != nil {
		return nil, err
	}

	state := &ActivityCenterState{
		HasSeen:   !unseen,
		UpdatedAt: updatedAt,
	}
	return state, nil
}

func (db sqlitePersistence) UpdateActivityCenterState(updatedAt uint64) (*ActivityCenterState, error) {
	var unreadCount int
	err := db.db.QueryRow("SELECT COUNT(1) FROM activity_center_notifications WHERE read = 0 AND deleted = 0").Scan(&unreadCount)
	if err != nil {
		return nil, err
	}
	var hasSeen int
	if unreadCount == 0 {
		hasSeen = 1
	}

	_, err = db.db.Exec(`UPDATE activity_center_states SET has_seen = ?, updated_at = ?`, hasSeen, updatedAt)
	return &ActivityCenterState{HasSeen: hasSeen == 1, UpdatedAt: updatedAt}, err
}
