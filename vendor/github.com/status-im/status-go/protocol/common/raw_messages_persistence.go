package common

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/gob"
	"strings"
	"time"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

type RawMessageConfirmation struct {
	// DataSyncID is the ID of the datasync message sent
	DataSyncID []byte
	// MessageID is the message id of the message
	MessageID []byte
	// PublicKey is the compressed receiver public key
	PublicKey []byte
	// ConfirmedAt is the unix timestamp in seconds of when the message was confirmed
	ConfirmedAt int64
}

type RawMessagesPersistence struct {
	db *sql.DB
}

func NewRawMessagesPersistence(db *sql.DB) *RawMessagesPersistence {
	return &RawMessagesPersistence{db: db}
}

func (db RawMessagesPersistence) SaveRawMessage(message *RawMessage) error {
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

	var pubKeys [][]byte
	for _, pk := range message.Recipients {
		pubKeys = append(pubKeys, crypto.CompressPubkey(pk))
	}
	// Encode recipients
	var encodedRecipients bytes.Buffer
	encoder := gob.NewEncoder(&encodedRecipients)

	if err := encoder.Encode(pubKeys); err != nil {
		return err
	}

	// If the message is not sent, we check whether there's a record
	// in the database already and preserve the state
	if !message.Sent {
		oldMessage, err := db.rawMessageByID(tx, message.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if oldMessage != nil {
			message.Sent = oldMessage.Sent
		}
	}

	_, err = tx.Exec(`
		 INSERT INTO
		 raw_messages
		 (
		   id,
		   local_chat_id,
		   last_sent,
		   send_count,
		   sent,
		   message_type,
		   resend_automatically,
		   recipients,
		   skip_encryption,
	           send_push_notification,
		   skip_group_message_wrap,
		   send_on_personal_topic,
		   payload
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		message.ID,
		message.LocalChatID,
		message.LastSent,
		message.SendCount,
		message.Sent,
		message.MessageType,
		message.ResendAutomatically,
		encodedRecipients.Bytes(),
		message.SkipEncryptionLayer,
		message.SendPushNotification,
		message.SkipGroupMessageWrap,
		message.SendOnPersonalTopic,
		message.Payload)
	return err
}

func (db RawMessagesPersistence) RawMessageByID(id string) (*RawMessage, error) {
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

	return db.rawMessageByID(tx, id)
}

func (db RawMessagesPersistence) rawMessageByID(tx *sql.Tx, id string) (*RawMessage, error) {
	var rawPubKeys [][]byte
	var encodedRecipients []byte
	var skipGroupMessageWrap sql.NullBool
	var sendOnPersonalTopic sql.NullBool
	message := &RawMessage{}

	err := tx.QueryRow(`
			SELECT
			  id,
			  local_chat_id,
			  last_sent,
			  send_count,
			  sent,
			  message_type,
			  resend_automatically,
			  recipients,
			  skip_encryption,
		          send_push_notification,
			  skip_group_message_wrap,
			  send_on_personal_topic,
		          payload
			FROM
				raw_messages
			WHERE
				id = ?`,
		id,
	).Scan(
		&message.ID,
		&message.LocalChatID,
		&message.LastSent,
		&message.SendCount,
		&message.Sent,
		&message.MessageType,
		&message.ResendAutomatically,
		&encodedRecipients,
		&message.SkipEncryptionLayer,
		&message.SendPushNotification,
		&skipGroupMessageWrap,
		&sendOnPersonalTopic,
		&message.Payload,
	)
	if err != nil {
		return nil, err
	}

	if rawPubKeys != nil {
		// Restore recipients
		decoder := gob.NewDecoder(bytes.NewBuffer(encodedRecipients))
		err = decoder.Decode(&rawPubKeys)
		if err != nil {
			return nil, err
		}
		for _, pkBytes := range rawPubKeys {
			pubkey, err := crypto.UnmarshalPubkey(pkBytes)
			if err != nil {
				return nil, err
			}
			message.Recipients = append(message.Recipients, pubkey)
		}
	}

	if skipGroupMessageWrap.Valid {
		message.SkipGroupMessageWrap = skipGroupMessageWrap.Bool
	}

	if sendOnPersonalTopic.Valid {
		message.SendOnPersonalTopic = sendOnPersonalTopic.Bool
	}

	return message, nil
}

func (db RawMessagesPersistence) RawMessagesIDsByType(t protobuf.ApplicationMetadataMessage_Type) ([]string, error) {
	ids := []string{}

	rows, err := db.db.Query(`
			SELECT
			  id
			FROM
				raw_messages
			WHERE
			message_type = ?`,
		t)
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// MarkAsConfirmed marks all the messages with dataSyncID as confirmed and returns
// the messageIDs that can be considered confirmed.
// If atLeastOne is set it will return messageid if at least once of the messages
// sent has been confirmed
func (db RawMessagesPersistence) MarkAsConfirmed(dataSyncID []byte, atLeastOne bool) (messageID types.HexBytes, err error) {
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

	confirmedAt := time.Now().Unix()
	_, err = tx.Exec(`UPDATE raw_message_confirmations SET confirmed_at = ? WHERE datasync_id = ? AND confirmed_at = 0`, confirmedAt, dataSyncID)
	if err != nil {
		return
	}

	// Select any tuple that has a message_id with a datasync_id = ? and that has just been confirmed
	rows, err := tx.Query(`SELECT message_id,confirmed_at FROM raw_message_confirmations WHERE message_id = (SELECT message_id FROM raw_message_confirmations WHERE datasync_id = ? LIMIT 1)`, dataSyncID)
	if err != nil {
		return
	}
	defer rows.Close()

	confirmedResult := true

	for rows.Next() {
		var confirmedAt int64
		err = rows.Scan(&messageID, &confirmedAt)
		if err != nil {
			return
		}
		confirmed := confirmedAt > 0

		if atLeastOne && confirmed {
			// We return, as at least one was confirmed
			return
		}

		confirmedResult = confirmedResult && confirmed
	}

	if !confirmedResult {
		messageID = nil
		return
	}

	return
}

func (db RawMessagesPersistence) InsertPendingConfirmation(confirmation *RawMessageConfirmation) error {

	_, err := db.db.Exec(`INSERT INTO raw_message_confirmations
		 (datasync_id, message_id, public_key)
		 VALUES
		 (?,?,?)`,
		confirmation.DataSyncID,
		confirmation.MessageID,
		confirmation.PublicKey,
	)
	return err
}

func (db RawMessagesPersistence) SaveHashRatchetMessage(groupID []byte, keyID []byte, m *types.Message) error {
	_, err := db.db.Exec(`INSERT INTO hash_ratchet_encrypted_messages(hash, sig, TTL, timestamp, topic, payload, dst, p2p, padding, group_id, key_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, m.Hash, m.Sig, m.TTL, m.Timestamp, types.TopicTypeToByteArray(m.Topic), m.Payload, m.Dst, m.P2P, m.Padding, groupID, keyID)
	return err
}

func (db RawMessagesPersistence) GetHashRatchetMessages(keyID []byte) ([]*types.Message, error) {
	var messages []*types.Message

	rows, err := db.db.Query(`SELECT hash, sig, TTL, timestamp, topic, payload, dst, p2p, padding FROM hash_ratchet_encrypted_messages WHERE key_id = ?`, keyID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var topic []byte
		message := &types.Message{}

		err := rows.Scan(&message.Hash, &message.Sig, &message.TTL, &message.Timestamp, &topic, &message.Payload, &message.Dst, &message.P2P, &message.Padding)
		if err != nil {
			return nil, err
		}

		message.Topic = types.BytesToTopic(topic)
		messages = append(messages, message)
	}

	return messages, nil
}

func (db RawMessagesPersistence) DeleteHashRatchetMessages(ids [][]byte) error {
	if len(ids) == 0 {
		return nil
	}

	idsArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsArgs = append(idsArgs, id)
	}
	inVector := strings.Repeat("?, ", len(ids)-1) + "?"

	_, err := db.db.Exec("DELETE FROM hash_ratchet_encrypted_messages WHERE hash IN ("+inVector+")", idsArgs...) // nolint: gosec

	return err
}

func (db *RawMessagesPersistence) IsMessageAlreadyCompleted(hash []byte) (bool, error) {
	var alreadyCompleted int
	err := db.db.QueryRow("SELECT COUNT(*) FROM message_segments_completed WHERE hash = ?", hash).Scan(&alreadyCompleted)
	if err != nil {
		return false, err
	}
	return alreadyCompleted > 0, nil
}

func (db *RawMessagesPersistence) SaveMessageSegment(segment *protobuf.SegmentMessage, sigPubKey *ecdsa.PublicKey, timestamp int64) error {
	sigPubKeyBlob := crypto.CompressPubkey(sigPubKey)

	_, err := db.db.Exec("INSERT INTO message_segments (hash, segment_index, segments_count, sig_pub_key, payload, timestamp) VALUES (?, ?, ?, ?, ?, ?)",
		segment.EntireMessageHash, segment.Index, segment.SegmentsCount, sigPubKeyBlob, segment.Payload, timestamp)

	return err
}

// Get ordered message segments for given hash
func (db *RawMessagesPersistence) GetMessageSegments(hash []byte, sigPubKey *ecdsa.PublicKey) ([]*protobuf.SegmentMessage, error) {
	sigPubKeyBlob := crypto.CompressPubkey(sigPubKey)

	rows, err := db.db.Query("SELECT hash, segment_index, segments_count, payload FROM message_segments WHERE hash = ? AND sig_pub_key = ? ORDER BY segment_index", hash, sigPubKeyBlob)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []*protobuf.SegmentMessage
	for rows.Next() {
		var segment protobuf.SegmentMessage
		err := rows.Scan(&segment.EntireMessageHash, &segment.Index, &segment.SegmentsCount, &segment.Payload)
		if err != nil {
			return nil, err
		}
		segments = append(segments, &segment)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return segments, nil
}

func (db *RawMessagesPersistence) RemoveMessageSegmentsOlderThan(timestamp int64) error {
	_, err := db.db.Exec("DELETE FROM message_segments WHERE timestamp < ?", timestamp)
	return err
}

func (db *RawMessagesPersistence) CompleteMessageSegments(hash []byte, sigPubKey *ecdsa.PublicKey, timestamp int64) error {
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

	sigPubKeyBlob := crypto.CompressPubkey(sigPubKey)

	_, err = tx.Exec("DELETE FROM message_segments WHERE hash = ? AND sig_pub_key = ?", hash, sigPubKeyBlob)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO message_segments_completed (hash, sig_pub_key, timestamp) VALUES (?,?,?)", hash, sigPubKeyBlob, timestamp)
	if err != nil {
		return err
	}

	return err
}

func (db *RawMessagesPersistence) RemoveMessageSegmentsCompletedOlderThan(timestamp int64) error {
	_, err := db.db.Exec("DELETE FROM message_segments_completed WHERE timestamp < ?", timestamp)
	return err
}
