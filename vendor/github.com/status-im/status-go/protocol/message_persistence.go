package protocol

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"

	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

var basicMessagesSelectQuery = `
SELECT    %s %s
FROM      user_messages m1
LEFT JOIN user_messages m2
ON        m1.response_to = m2.id
LEFT JOIN contacts c
ON        m1.source = c.id
LEFT JOIN discord_messages dm
ON        m1.discord_message_id = dm.id
LEFT JOIN discord_message_authors dm_author
ON        dm.author_id = dm_author.id
LEFT JOIN discord_message_attachments dm_attachment
ON        dm.id = dm_attachment.discord_message_id
LEFT JOIN discord_messages m2_dm
ON        m2.discord_message_id = m2_dm.id
LEFT JOIN discord_message_authors m2_dm_author
ON        m2_dm.author_id = m2_dm_author.id
LEFT JOIN bridge_messages bm
ON        m1.id = bm.user_messages_id
LEFT JOIN bridge_messages bm_response
ON        m2.id = bm_response.user_messages_id
`

var basicInsertDiscordMessageAuthorQuery = `INSERT OR REPLACE INTO discord_message_authors(id,name,discriminator,nickname,avatar_url, avatar_image_payload) VALUES (?,?,?,?,?,?)`

var cursor = "substr('0000000000000000000000000000000000000000000000000000000000000000' || m1.clock_value, -64, 64) || m1.id"
var cursorField = cursor + " as cursor"

func (db sqlitePersistence) buildMessagesQueryWithAdditionalFields(additionalSelectFields, whereAndTheRest string) string {
	allFields := db.tableUserMessagesAllFieldsJoin()
	if additionalSelectFields != "" {
		additionalSelectFields = "," + additionalSelectFields
	}
	base := fmt.Sprintf(basicMessagesSelectQuery, allFields, additionalSelectFields)
	return base + " " + whereAndTheRest
}

func (db sqlitePersistence) buildMessagesQuery(whereAndTheRest string) string {
	return db.buildMessagesQueryWithAdditionalFields("", whereAndTheRest)
}

func (db sqlitePersistence) tableUserMessagesAllFields() string {
	return `id,
    		whisper_timestamp,
    		source,
    		text,
    		content_type,
    		username,
    		timestamp,
    		chat_id,
		local_chat_id,
    		message_type,
    		clock_value,
    		seen,
    		outgoing_status,
		parsed_text,
		sticker_pack,
		sticker_hash,
		image_payload,
		image_type,
		album_id,
		album_images,
		album_images_count,
		image_width,
		image_height,
		image_base64,
		audio_payload,
		audio_type,
		audio_duration_ms,
		audio_base64,
		community_id,
		mentions,
		links,
		unfurled_links,
		unfurled_status_links,
		command_id,
		command_value,
		command_from,
		command_address,
		command_contract,
		command_transaction_hash,
		command_state,
		command_signature,
		replace_message,
		edited_at,
		deleted,
		deleted_by,
		deleted_for_me,
		rtl,
		line_count,
		response_to,
		gap_from,
		gap_to,
		contact_request_state,
		contact_verification_status,
		mentioned,
		replied,
    discord_message_id`
}

// keep the same order as in tableUserMessagesScanAllFields
func (db sqlitePersistence) tableUserMessagesAllFieldsJoin() string {
	return `m1.id,
    		m1.whisper_timestamp,
    		m1.source,
    		m1.text,
    		m1.content_type,
    		m1.username,
    		m1.timestamp,
    		m1.chat_id,
		m1.local_chat_id,
    		m1.message_type,
    		m1.clock_value,
    		m1.seen,
    		m1.outgoing_status,
		m1.parsed_text,
		m1.sticker_pack,
		m1.sticker_hash,
		m1.image_payload,
		m1.image_type,
		COALESCE(m1.album_id, ""),
		COALESCE(m1.album_images_count, 0),
		COALESCE(m1.image_width, 0),
		COALESCE(m1.image_height, 0),
		COALESCE(m1.audio_duration_ms,0),
		m1.community_id,
		m1.mentions,
		m1.links,
		m1.unfurled_links,
		m1.unfurled_status_links,
		m1.command_id,
		m1.command_value,
		m1.command_from,
		m1.command_address,
		m1.command_contract,
		m1.command_transaction_hash,
		m1.command_state,
		m1.command_signature,
		m1.replace_message,
		m1.edited_at,
		m1.deleted,
		m1.deleted_by,
		m1.deleted_for_me,
		m1.rtl,
		m1.line_count,
		m1.response_to,
		m1.gap_from,
		m1.gap_to,
		m1.contact_request_state,
		m1.contact_verification_status,
		m1.mentioned,
		m1.replied,
    COALESCE(m1.discord_message_id, ""),
    COALESCE(dm.author_id, ""),
    COALESCE(dm.type, ""),
    COALESCE(dm.timestamp, ""),
    COALESCE(dm.timestamp_edited, ""),
    COALESCE(dm.content, ""),
    COALESCE(dm.reference_message_id, ""),
    COALESCE(dm.reference_channel_id, ""),
    COALESCE(dm_author.name, ""),
    COALESCE(dm_author.discriminator, ""),
    COALESCE(dm_author.nickname, ""),
    COALESCE(dm_author.avatar_url, ""),
    COALESCE(dm_attachment.id, ""),
    COALESCE(dm_attachment.discord_message_id, ""),
    COALESCE(dm_attachment.url, ""),
    COALESCE(dm_attachment.file_name, ""),
    COALESCE(dm_attachment.content_type, ""),
		m2.source,
		m2.text,
		m2.parsed_text,
		m2.album_images,
		m2.album_images_count,
		m2.audio_duration_ms,
		m2.community_id,
		m2.id,
        m2.content_type,
        m2.deleted,
        m2.deleted_for_me,
		c.alias,
		c.identicon,
    COALESCE(m2.discord_message_id, ""),
		COALESCE(m2_dm_author.name, ""),
		COALESCE(m2_dm_author.nickname, ""),
		COALESCE(m2_dm_author.avatar_url, ""),
		COALESCE(bm.bridge_name, ""),
		COALESCE(bm.user_name, ""),
		COALESCE(bm.user_avatar, ""),
		COALESCE(bm.user_id, ""),
		COALESCE(bm.content, ""),
		COALESCE(bm.message_id, ""),
		COALESCE(bm.parent_message_id, ""),
		COALESCE(bm_response.bridge_name, ""),
		COALESCE(bm_response.user_name, ""),
		COALESCE(bm_response.user_avatar, ""),
		COALESCE(bm_response.user_id, ""),
		COALESCE(bm_response.content, "")`
}

func (db sqlitePersistence) tableUserMessagesAllFieldsCount() int {
	return strings.Count(db.tableUserMessagesAllFields(), ",") + 1
}

type scanner interface {
	Scan(dest ...interface{}) error
}

// keep the same order as in tableUserMessagesAllFieldsJoin
func (db sqlitePersistence) tableUserMessagesScanAllFields(row scanner, message *common.Message, others ...interface{}) error {
	var quotedID sql.NullString
	var ContentType sql.NullInt64
	var quotedText sql.NullString
	var quotedParsedText []byte
	var quotedAlbumImages []byte
	var quotedAlbumImagesCount sql.NullInt64
	var quotedFrom sql.NullString
	var quotedAudioDuration sql.NullInt64
	var quotedCommunityID sql.NullString
	var quotedDeleted sql.NullBool
	var quotedDeletedForMe sql.NullBool
	var serializedMentions []byte
	var serializedLinks []byte
	var serializedUnfurledLinks []byte
	var serializedUnfurledStatusLinks []byte
	var alias sql.NullString
	var identicon sql.NullString
	var communityID sql.NullString
	var gapFrom sql.NullInt64
	var gapTo sql.NullInt64
	var editedAt sql.NullInt64
	var deleted sql.NullBool
	var deletedBy sql.NullString
	var deletedForMe sql.NullBool
	var contactRequestState sql.NullInt64
	var contactVerificationState sql.NullInt64

	sticker := &protobuf.StickerMessage{}
	command := &common.CommandParameters{}
	audio := &protobuf.AudioMessage{}
	image := &protobuf.ImageMessage{}
	discordMessage := &protobuf.DiscordMessage{
		Author:      &protobuf.DiscordMessageAuthor{},
		Reference:   &protobuf.DiscordMessageReference{},
		Attachments: []*protobuf.DiscordMessageAttachment{},
	}
	bridgeMessage := &protobuf.BridgeMessage{}

	quotedBridgeMessage := &protobuf.BridgeMessage{}

	quotedDiscordMessage := &protobuf.DiscordMessage{
		Author: &protobuf.DiscordMessageAuthor{},
	}

	attachment := &protobuf.DiscordMessageAttachment{}

	args := []interface{}{
		&message.ID,
		&message.WhisperTimestamp,
		&message.From, // source in table
		&message.Text,
		&message.ContentType,
		&message.Alias,
		&message.Timestamp,
		&message.ChatId,
		&message.LocalChatID,
		&message.MessageType,
		&message.Clock,
		&message.Seen,
		&message.OutgoingStatus,
		&message.ParsedText,
		&sticker.Pack,
		&sticker.Hash,
		&image.Payload,
		&image.Format,
		&image.AlbumId,
		&image.AlbumImagesCount,
		&image.Width,
		&image.Height,
		&audio.DurationMs,
		&communityID,
		&serializedMentions,
		&serializedLinks,
		&serializedUnfurledLinks,
		&serializedUnfurledStatusLinks,
		&command.ID,
		&command.Value,
		&command.From,
		&command.Address,
		&command.Contract,
		&command.TransactionHash,
		&command.CommandState,
		&command.Signature,
		&message.Replace,
		&editedAt,
		&deleted,
		&deletedBy,
		&deletedForMe,
		&message.RTL,
		&message.LineCount,
		&message.ResponseTo,
		&gapFrom,
		&gapTo,
		&contactRequestState,
		&contactVerificationState,
		&message.Mentioned,
		&message.Replied,
		&discordMessage.Id,
		&discordMessage.Author.Id,
		&discordMessage.Type,
		&discordMessage.Timestamp,
		&discordMessage.TimestampEdited,
		&discordMessage.Content,
		&discordMessage.Reference.MessageId,
		&discordMessage.Reference.ChannelId,
		&discordMessage.Author.Name,
		&discordMessage.Author.Discriminator,
		&discordMessage.Author.Nickname,
		&discordMessage.Author.AvatarUrl,
		&attachment.Id,
		&attachment.MessageId,
		&attachment.Url,
		&attachment.FileName,
		&attachment.ContentType,
		&quotedFrom,
		&quotedText,
		&quotedParsedText,
		&quotedAlbumImages,
		&quotedAlbumImagesCount,
		&quotedAudioDuration,
		&quotedCommunityID,
		&quotedID,
		&ContentType,
		&quotedDeleted,
		&quotedDeletedForMe,
		&alias,
		&identicon,
		&quotedDiscordMessage.Id,
		&quotedDiscordMessage.Author.Name,
		&quotedDiscordMessage.Author.Nickname,
		&quotedDiscordMessage.Author.AvatarUrl,
		&bridgeMessage.BridgeName,
		&bridgeMessage.UserName,
		&bridgeMessage.UserAvatar,
		&bridgeMessage.UserID,
		&bridgeMessage.Content,
		&bridgeMessage.MessageID,
		&bridgeMessage.ParentMessageID,
		&quotedBridgeMessage.BridgeName,
		&quotedBridgeMessage.UserName,
		&quotedBridgeMessage.UserAvatar,
		&quotedBridgeMessage.UserID,
		&quotedBridgeMessage.Content,
	}
	err := row.Scan(append(args, others...)...)
	if err != nil {
		return err
	}

	if editedAt.Valid {
		message.EditedAt = uint64(editedAt.Int64)
	}

	if deleted.Valid {
		message.Deleted = deleted.Bool
	}

	if deletedBy.Valid {
		message.DeletedBy = deletedBy.String
	}

	if deletedForMe.Valid {
		message.DeletedForMe = deletedForMe.Bool
	}

	if contactRequestState.Valid {
		message.ContactRequestState = common.ContactRequestState(contactRequestState.Int64)
	}

	if contactVerificationState.Valid {
		message.ContactVerificationState = common.ContactVerificationState(contactVerificationState.Int64)
	}

	if quotedText.Valid {
		if quotedDeleted.Bool || quotedDeletedForMe.Bool {
			message.QuotedMessage = &common.QuotedMessage{
				ID:           quotedID.String,
				From:         quotedFrom.String,
				Deleted:      quotedDeleted.Bool,
				DeletedForMe: quotedDeletedForMe.Bool,
			}
		} else {
			message.QuotedMessage = &common.QuotedMessage{
				ID:               quotedID.String,
				ContentType:      ContentType.Int64,
				From:             quotedFrom.String,
				Text:             quotedText.String,
				ParsedText:       quotedParsedText,
				AlbumImages:      quotedAlbumImages,
				AlbumImagesCount: quotedAlbumImagesCount.Int64,
				CommunityID:      quotedCommunityID.String,
				Deleted:          quotedDeleted.Bool,
			}
			if message.QuotedMessage.ContentType == int64(protobuf.ChatMessage_DISCORD_MESSAGE) {
				message.QuotedMessage.DiscordMessage = quotedDiscordMessage
			}
			if message.QuotedMessage.ContentType == int64(protobuf.ChatMessage_BRIDGE_MESSAGE) {
				message.QuotedMessage.BridgeMessage = quotedBridgeMessage
			}
		}
	}
	message.Alias = alias.String
	message.Identicon = identicon.String

	if gapFrom.Valid && gapTo.Valid {
		message.GapParameters = &common.GapParameters{
			From: uint32(gapFrom.Int64),
			To:   uint32(gapTo.Int64),
		}
	}

	if communityID.Valid {
		message.CommunityID = communityID.String
	}

	if serializedMentions != nil {
		err := json.Unmarshal(serializedMentions, &message.Mentions)
		if err != nil {
			return err
		}
	}

	if serializedLinks != nil {
		err := json.Unmarshal(serializedLinks, &message.Links)
		if err != nil {
			return err
		}
	}

	if serializedUnfurledLinks != nil {
		err = json.Unmarshal(serializedUnfurledLinks, &message.UnfurledLinks)
		if err != nil {
			return err
		}
	}

	if serializedUnfurledStatusLinks != nil {
		// use proto.Marshal, because json.Marshal doesn't support `oneof` fields
		var links protobuf.UnfurledStatusLinks
		err = proto.Unmarshal(serializedUnfurledStatusLinks, &links)
		if err != nil {
			return err
		}
		message.UnfurledStatusLinks = &links
	}

	if attachment.Id != "" {
		discordMessage.Attachments = append(discordMessage.Attachments, attachment)
	}

	switch message.ContentType {
	case protobuf.ChatMessage_STICKER:
		message.Payload = &protobuf.ChatMessage_Sticker{Sticker: sticker}

	case protobuf.ChatMessage_AUDIO:
		message.Payload = &protobuf.ChatMessage_Audio{Audio: audio}

	case protobuf.ChatMessage_TRANSACTION_COMMAND:
		message.CommandParameters = command

	case protobuf.ChatMessage_IMAGE:
		message.Payload = &protobuf.ChatMessage_Image{Image: image}

	case protobuf.ChatMessage_DISCORD_MESSAGE:
		message.Payload = &protobuf.ChatMessage_DiscordMessage{
			DiscordMessage: discordMessage,
		}

	case protobuf.ChatMessage_BRIDGE_MESSAGE:
		message.Payload = &protobuf.ChatMessage_BridgeMessage{
			BridgeMessage: bridgeMessage,
		}
	}

	return nil
}

func (db sqlitePersistence) tableUserMessagesAllValues(message *common.Message) ([]interface{}, error) {
	var gapFrom, gapTo uint32

	var albumImages []byte
	if message.QuotedMessage != nil {
		albumImages = []byte(message.QuotedMessage.AlbumImages)
	}

	sticker := message.GetSticker()
	if sticker == nil {
		sticker = &protobuf.StickerMessage{}
	}

	image := message.GetImage()
	if image == nil {
		image = &protobuf.ImageMessage{}
	}

	audio := message.GetAudio()
	if audio == nil {
		audio = &protobuf.AudioMessage{}
	}

	command := message.CommandParameters
	if command == nil {
		command = &common.CommandParameters{}
	}

	discordMessage := message.GetDiscordMessage()
	if discordMessage == nil {
		discordMessage = &protobuf.DiscordMessage{
			Author:      &protobuf.DiscordMessageAuthor{},
			Reference:   &protobuf.DiscordMessageReference{},
			Attachments: make([]*protobuf.DiscordMessageAttachment, 0),
		}
	}

	if message.GapParameters != nil {
		gapFrom = message.GapParameters.From
		gapTo = message.GapParameters.To
	}

	var serializedMentions []byte
	var err error
	if len(message.Mentions) != 0 {
		serializedMentions, err = json.Marshal(message.Mentions)
		if err != nil {
			return nil, err
		}
	}

	var serializedLinks []byte
	if len(message.Links) != 0 {
		serializedLinks, err = json.Marshal(message.Links)
		if err != nil {
			return nil, err
		}
	}

	var serializedUnfurledLinks []byte
	if links := message.GetUnfurledLinks(); links != nil {
		serializedUnfurledLinks, err = json.Marshal(links)
		if err != nil {
			return nil, err
		}
	}

	var serializedUnfurledStatusLinks []byte
	if links := message.GetUnfurledStatusLinks(); links != nil {
		// use proto.Marshal, because json.Marshal doesn't support `oneof` fields
		serializedUnfurledStatusLinks, err = proto.Marshal(links)
		if err != nil {
			return nil, err
		}
	}

	return []interface{}{
		message.ID,
		message.WhisperTimestamp,
		message.From, // source in table
		message.Text,
		message.ContentType,
		message.Alias,
		message.Timestamp,
		message.ChatId,
		message.LocalChatID,
		message.MessageType,
		message.Clock,
		message.Seen,
		message.OutgoingStatus,
		message.ParsedText,
		sticker.Pack,
		sticker.Hash,
		image.Payload,
		image.Format,
		image.AlbumId,
		albumImages,
		image.AlbumImagesCount,
		image.Width,
		image.Height,
		message.Base64Image,
		audio.Payload,
		audio.Type,
		audio.DurationMs,
		message.Base64Audio,
		message.CommunityID,
		serializedMentions,
		serializedLinks,
		serializedUnfurledLinks,
		serializedUnfurledStatusLinks,
		command.ID,
		command.Value,
		command.From,
		command.Address,
		command.Contract,
		command.TransactionHash,
		command.CommandState,
		command.Signature,
		message.Replace,
		int64(message.EditedAt),
		message.Deleted,
		message.DeletedBy,
		message.DeletedForMe,
		message.RTL,
		message.LineCount,
		message.ResponseTo,
		gapFrom,
		gapTo,
		message.ContactRequestState,
		message.ContactVerificationState,
		message.Mentioned,
		message.Replied,
		discordMessage.Id,
	}, nil
}

func (db sqlitePersistence) messageByID(tx *sql.Tx, id string) (*common.Message, error) {
	var err error
	if tx == nil {
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
	}

	query := db.buildMessagesQuery("WHERE m1.id = ?")
	rows, err := tx.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getMessageFromScanRows(db, rows)
}

func (db sqlitePersistence) albumMessages(chatID, albumID string) ([]*common.Message, error) {
	if albumID == "" {
		return nil, nil
	}
	query := db.buildMessagesQuery("WHERE m1.album_id = ? and m1.local_chat_id = ?")
	rows, err := db.db.Query(query, albumID, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getMessagesFromScanRows(db, rows, false)
}

func (db sqlitePersistence) MessageByCommandID(chatID, id string) (*common.Message, error) {

	where := `WHERE
			m1.command_id = ?
			AND
			m1.local_chat_id = ?
			ORDER BY m1.clock_value DESC
			LIMIT 1`
	query := db.buildMessagesQuery(where)
	rows, err := db.db.Query(query, id, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getMessageFromScanRows(db, rows)
}

func (db sqlitePersistence) MessageByID(id string) (*common.Message, error) {
	return db.messageByID(nil, id)
}

func (db sqlitePersistence) AlbumMessages(chatID, albumID string) ([]*common.Message, error) {
	return db.albumMessages(chatID, albumID)
}

func (db sqlitePersistence) MessagesExist(ids []string) (map[string]bool, error) {
	result := make(map[string]bool)
	if len(ids) == 0 {
		return result, nil
	}

	idsArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsArgs = append(idsArgs, id)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	query := "SELECT id FROM user_messages WHERE id IN (" + inVector + ")" // nolint: gosec
	rows, err := db.db.Query(query, idsArgs...)
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
		result[id] = true
	}

	return result, nil
}

func (db sqlitePersistence) MessagesByIDs(ids []string) ([]*common.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	idsArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsArgs = append(idsArgs, id)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"

	// nolint: gosec
	where := fmt.Sprintf("WHERE NOT(m1.hide) AND m1.id IN (%s)", inVector)
	query := db.buildMessagesQuery(where)
	rows, err := db.db.Query(query, idsArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getMessagesFromScanRows(db, rows, false)
}

func (db sqlitePersistence) MessagesByResponseTo(responseTo string) ([]*common.Message, error) {
	where := "WHERE m1.response_to = ?"
	query := db.buildMessagesQuery(where)
	rows, err := db.db.Query(query, responseTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getMessagesFromScanRows(db, rows, false)
}

// MessageByChatID returns all messages for a given chatID in descending order.
// Ordering is accomplished using two concatenated values: ClockValue and ID.
// These two values are also used to compose a cursor which is returned to the result.
func (db sqlitePersistence) MessageByChatID(chatID string, currCursor string, limit int) ([]*common.Message, string, error) {
	cursorWhere := ""
	if currCursor != "" {
		cursorWhere = "AND cursor <= ?" //nolint: goconst
	}
	args := []interface{}{chatID}
	if currCursor != "" {
		args = append(args, currCursor)
	}
	// Build a new column `cursor` at the query time by having a fixed-sized clock value at the beginning
	// concatenated with message ID. Results are sorted using this new column.
	// This new column values can also be returned as a cursor for subsequent requests.
	where := fmt.Sprintf(`
            WHERE
                NOT(m1.hide) AND m1.local_chat_id = ? %s
            ORDER BY cursor DESC
            LIMIT ?`, cursorWhere)

	query := db.buildMessagesQueryWithAdditionalFields(cursorField, where)
	rows, err := db.db.Query(
		query,
		append(args, limit+1)..., // take one more to figure our whether a cursor should be returned
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	result, cursors, err := getMessagesAndCursorsFromScanRows(db, rows)
	if err != nil {
		return nil, "", err
	}

	var newCursor string
	if len(result) > limit {
		newCursor = cursors[limit]
		result = result[:limit]
	}
	return result, newCursor, nil
}

func (db sqlitePersistence) FirstUnseenMessageID(chatID string) (string, error) {
	var id string
	err := db.db.QueryRow(`
			SELECT
				id
			FROM
				user_messages m1
			WHERE
				m1.local_chat_id = ?
				AND NOT(m1.seen) AND NOT(m1.hide) AND NOT(m1.deleted) AND NOT(m1.deleted_for_me)
			ORDER BY m1.clock_value ASC
			LIMIT 1`,
		chatID).Scan(&id)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// Get last chat message that is not hide or deleted or deleted_for_me
func (db sqlitePersistence) LatestMessageByChatID(chatID string) ([]*common.Message, error) {
	args := []interface{}{chatID}
	where := `WHERE
                NOT(m1.hide) AND NOT(m1.deleted) AND NOT(m1.deleted_for_me) AND m1.local_chat_id = ?
            ORDER BY cursor DESC
            LIMIT ?`

	query := db.buildMessagesQueryWithAdditionalFields(cursorField, where)

	rows, err := db.db.Query(
		query,
		append(args, 2)..., // take one more to figure our whether a cursor should be returned
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, _, err := getMessagesAndCursorsFromScanRows(db, rows)
	if err != nil {
		return nil, err
	}

	if len(result) > 1 {
		result = result[:1]
	}
	return result, nil
}

func (db sqlitePersistence) latestIncomingMessageClock(chatID string) (uint64, error) {
	var clock uint64
	err := db.db.QueryRow(
		fmt.Sprintf(
			`
			SELECT
                clock_value
			FROM
				user_messages m1
			WHERE
				m1.local_chat_id = ? AND m1.outgoing_status = ''
			%s DESC
			LIMIT 1
		`, cursor),
		chatID).Scan(&clock)
	if err != nil {
		return 0, err
	}
	return clock, nil
}

func (db sqlitePersistence) PendingContactRequests(currCursor string, limit int) ([]*common.Message, string, error) {
	cursorWhere := ""
	if currCursor != "" {
		cursorWhere = "AND cursor <= ?" //nolint: goconst
	}
	args := []interface{}{protobuf.ChatMessage_CONTACT_REQUEST}
	if currCursor != "" {
		args = append(args, currCursor)
	}
	// Build a new column `cursor` at the query time by having a fixed-sized clock value at the beginning
	// concatenated with message ID. Results are sorted using this new column.
	// This new column values can also be returned as a cursor for subsequent requests.
	where := fmt.Sprintf(`
            WHERE
                NOT(m1.hide) AND NOT(m1.seen) AND m1.content_type = ? %s
            ORDER BY cursor DESC
            LIMIT ?`, cursorWhere)

	query := db.buildMessagesQueryWithAdditionalFields(cursorField, where)
	rows, err := db.db.Query(
		query,
		append(args, limit+1)..., // take one more to figure our whether a cursor should be returned
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	result, cursors, err := getMessagesAndCursorsFromScanRows(db, rows)
	if err != nil {
		return nil, "", err
	}

	var newCursor string
	if len(result) > limit {
		newCursor = cursors[limit]
		result = result[:limit]
	}
	return result, newCursor, nil
}

func (db sqlitePersistence) LatestPendingContactRequestIDForContact(contactID string) (string, error) {
	var id string
	err := db.db.QueryRow(
		fmt.Sprintf(
			`
			SELECT
                                id
			FROM
				user_messages m1
			WHERE
				m1.local_chat_id = ? AND m1.content_type = ?
			ORDER BY %s DESC
			LIMIT 1
		`, cursor),
		contactID, protobuf.ChatMessage_CONTACT_REQUEST).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", err
	}
	return id, nil
}

func (db sqlitePersistence) LatestContactRequestIDs() (map[string]common.ContactRequestState, error) {
	res := map[string]common.ContactRequestState{}
	rows, err := db.db.Query(
		fmt.Sprintf(
			`
			SELECT
                                id, contact_request_state
			FROM
				user_messages m1
			WHERE
				m1.content_type = ?
			ORDER BY %s DESC
			LIMIT 20
		`, cursor), protobuf.ChatMessage_CONTACT_REQUEST)

	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		var id string
		var contactRequestState sql.NullInt64
		err := rows.Scan(&id, &contactRequestState)

		if err != nil {
			return nil, err
		}
		res[id] = common.ContactRequestState(contactRequestState.Int64)
	}

	return res, nil
}

// AllMessageByChatIDWhichMatchPattern returns all messages which match the search
// term, for a given chatID in descending order.
// Ordering is accomplished using two concatenated values: ClockValue and ID.
// These two values are also used to compose a cursor which is returned to the result.
func (db sqlitePersistence) AllMessageByChatIDWhichMatchTerm(chatID string, searchTerm string, caseSensitive bool) ([]*common.Message, error) {
	if searchTerm == "" {
		return nil, fmt.Errorf("empty search term")
	}

	searchCond := ""
	if caseSensitive {
		searchCond = "AND m1.text LIKE '%' || ? || '%'"
	} else {
		searchCond = "AND LOWER(m1.text) LIKE LOWER('%' || ? || '%')"
	}

	where := fmt.Sprintf(`
            WHERE
                NOT(m1.hide) AND m1.local_chat_id = ? %s
            ORDER BY cursor DESC`, searchCond)

	query := db.buildMessagesQueryWithAdditionalFields(cursorField, where)
	rows, err := db.db.Query(
		query,
		chatID, searchTerm,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getMessagesFromScanRows(db, rows, true)
}

// AllMessagesFromChatsAndCommunitiesWhichMatchTerm returns all messages which match the search
// term, if they belong to either any chat from the chatIds array or any channel of any community
// from communityIds array.
// Ordering is accomplished using two concatenated values: ClockValue and ID.
// These two values are also used to compose a cursor which is returned to the result.
func (db sqlitePersistence) AllMessagesFromChatsAndCommunitiesWhichMatchTerm(communityIds []string, chatIds []string, searchTerm string, caseSensitive bool) ([]*common.Message, error) {
	if searchTerm == "" {
		return nil, fmt.Errorf("empty search term")
	}

	chatsCond := ""
	if len(chatIds) > 0 {
		inVector := strings.Repeat("?, ", len(chatIds)-1) + "?"
		chatsCond = `m1.local_chat_id IN (%s)`
		chatsCond = fmt.Sprintf(chatsCond, inVector)
	}

	communitiesCond := ""
	if len(communityIds) > 0 {
		inVector := strings.Repeat("?, ", len(communityIds)-1) + "?"
		communitiesCond = `m1.local_chat_id IN (SELECT id FROM chats WHERE community_id IN (%s))`
		communitiesCond = fmt.Sprintf(communitiesCond, inVector)
	}

	searchCond := ""
	if caseSensitive {
		searchCond = "m1.text LIKE '%' || ? || '%'"
	} else {
		searchCond = "LOWER(m1.text) LIKE LOWER('%' || ? || '%')"
	}

	finalCond := "AND %s AND %s"
	if len(communityIds) > 0 && len(chatIds) > 0 {
		finalCond = "AND (%s OR %s) AND %s"
		finalCond = fmt.Sprintf(finalCond, chatsCond, communitiesCond, searchCond)
	} else if len(chatIds) > 0 {
		finalCond = fmt.Sprintf(finalCond, chatsCond, searchCond)
	} else if len(communityIds) > 0 {
		finalCond = fmt.Sprintf(finalCond, communitiesCond, searchCond)
	} else {
		return nil, fmt.Errorf("you must specify either community ids or chat ids or both")
	}

	var parameters []string
	parameters = append(parameters, chatIds...)
	parameters = append(parameters, communityIds...)
	parameters = append(parameters, searchTerm)

	idsArgs := make([]interface{}, 0, len(parameters))
	for _, param := range parameters {
		idsArgs = append(idsArgs, param)
	}

	where := fmt.Sprintf(`
            WHERE
                NOT(m1.hide) %s
            ORDER BY cursor DESC`, finalCond)

	finalQuery := db.buildMessagesQueryWithAdditionalFields(cursorField, where)
	rows, err := db.db.Query(finalQuery, idsArgs...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getMessagesFromScanRows(db, rows, true)
}

func (db sqlitePersistence) AllChatIDsByCommunity(tx *sql.Tx, communityID string) ([]string, error) {
	var err error
	var rows *sql.Rows
	query := "SELECT id FROM chats WHERE community_id = ?"
	if tx == nil {
		rows, err = db.db.Query(query, communityID)
	} else {
		rows, err = tx.Query(query, communityID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rst []string

	for rows.Next() {
		var chatID string
		err = rows.Scan(&chatID)
		if err != nil {
			return nil, err
		}
		rst = append(rst, chatID)
	}

	return rst, nil
}

func (db sqlitePersistence) CountActiveChattersInCommunity(communityID string, activeAfterTimestamp int64) (uint, error) {
	var activeChattersCount uint
	err := db.db.QueryRow(
		`
			SELECT COUNT(DISTINCT source)
			FROM user_messages
			JOIN chats ON user_messages.local_chat_id = chats.id
			WHERE chats.community_id = ?
			AND user_messages.timestamp >= ?
		`, communityID, activeAfterTimestamp).Scan(&activeChattersCount)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return activeChattersCount, nil
}

// PinnedMessageByChatID returns all pinned messages for a given chatID in descending order.
// Ordering is accomplished using two concatenated values: ClockValue and ID.
// These two values are also used to compose a cursor which is returned to the result.
func (db sqlitePersistence) PinnedMessageByChatIDs(chatIDs []string, currCursor string, limit int) ([]*common.PinnedMessage, string, error) {
	cursorWhere := ""
	if currCursor != "" {
		cursorWhere = "AND cursor <= ?" //nolint: goconst
	}
	args := make([]interface{}, len(chatIDs))
	for i, v := range chatIDs {
		args[i] = v
	}
	if currCursor != "" {
		args = append(args, currCursor)
	}

	limitStr := ""
	if limit > -1 {
		args = append(args, limit+1) // take one more to figure our whether a cursor should be returned
	}
	// Build a new column `cursor` at the query time by having a fixed-sized clock value at the beginning
	// concatenated with message ID. Results are sorted using this new column.
	// This new column values can also be returned as a cursor for subsequent requests.
	allFields := db.tableUserMessagesAllFieldsJoin()
	rows, err := db.db.Query(
		fmt.Sprintf(`
 			SELECT
 				%s,
 				pm.clock_value as pinnedAt,
 				pm.pinned_by as pinnedBy,
                                %s
 			FROM
 				pin_messages pm
 			JOIN
 				user_messages m1
 			ON
 				pm.message_id = m1.id
 			LEFT JOIN
 				user_messages m2
 			ON
 				m1.response_to = m2.id
 			LEFT JOIN
 				contacts c
 			ON
 				m1.source = c.id

       LEFT JOIN
             discord_messages dm
       ON
       m1.discord_message_id = dm.id

       LEFT JOIN
             discord_message_authors dm_author
       ON
       dm.author_id = dm_author.id

       LEFT JOIN
              discord_message_attachments dm_attachment
			 ON
       dm.id = dm_attachment.discord_message_id

			 LEFT JOIN
							discord_messages m2_dm
			 ON
			 m2.discord_message_id = m2_dm.id

				LEFT JOIN
							discord_message_authors m2_dm_author
			 ON
			 m2_dm.author_id = m2_dm_author.id

			 LEFT JOIN bridge_messages bm
			 ON m1.id = bm.user_messages_id

			LEFT JOIN bridge_messages bm_response
			ON m2.id = bm_response.user_messages_id

 			WHERE
 				pm.pinned = 1
 				AND NOT(m1.hide) AND m1.local_chat_id IN %s %s
 			ORDER BY cursor DESC
 			%s
 		`, allFields, cursorField, "(?"+strings.Repeat(",?", len(chatIDs)-1)+")", cursorWhere, limitStr),
		args..., // take one more to figure our whether a cursor should be returned
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	result, cursors, err := getPinnedMessagesAndCursorsFromScanRows(db, rows)
	if err != nil {
		return nil, "", err
	}

	var newCursor string

	if limit > -1 && len(result) > limit && cursors != nil {
		newCursor = cursors[limit]
		result = result[:limit]
	}
	return result, newCursor, nil
}

func (db sqlitePersistence) PinnedMessageByChatID(chatID string, currCursor string, limit int) ([]*common.PinnedMessage, string, error) {
	return db.PinnedMessageByChatIDs([]string{chatID}, currCursor, limit)
}

// MessageByChatIDs returns all messages for a given chatIDs in descending order.
// Ordering is accomplished using two concatenated values: ClockValue and ID.
// These two values are also used to compose a cursor which is returned to the result.
func (db sqlitePersistence) MessageByChatIDs(chatIDs []string, currCursor string, limit int) ([]*common.Message, string, error) {
	cursorWhere := ""
	if currCursor != "" {
		cursorWhere = "AND cursor <= ?" //nolint: goconst
	}
	args := make([]interface{}, len(chatIDs))
	for i, v := range chatIDs {
		args[i] = v
	}
	if currCursor != "" {
		args = append(args, currCursor)
	}
	// Build a new column `cursor` at the query time by having a fixed-sized clock value at the beginning
	// concatenated with message ID. Results are sorted using this new column.
	// This new column values can also be returned as a cursor for subsequent requests.
	where := fmt.Sprintf(`
			WHERE
				NOT(m1.hide) AND m1.local_chat_id IN %s %s
			ORDER BY cursor DESC
			LIMIT ?
		`, "(?"+strings.Repeat(",?", len(chatIDs)-1)+")", cursorWhere)
	query := db.buildMessagesQueryWithAdditionalFields(cursorField, where)
	rows, err := db.db.Query(
		query,
		append(args, limit+1)..., // take one more to figure our whether a cursor should be returned
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	result, cursors, err := getMessagesAndCursorsFromScanRows(db, rows)
	if err != nil {
		return nil, "", err
	}

	var newCursor string
	if len(result) > limit {
		newCursor = cursors[limit]
		result = result[:limit]
	}
	return result, newCursor, nil
}

func (db sqlitePersistence) OldestMessageWhisperTimestampByChatID(chatID string) (timestamp uint64, hasAnyMessage bool, err error) {
	var whisperTimestamp uint64
	err = db.db.QueryRow(
		`
			SELECT
				whisper_timestamp
			FROM
				user_messages m1
			WHERE
				m1.local_chat_id = ?
			ORDER BY substr('0000000000000000000000000000000000000000000000000000000000000000' || m1.clock_value, -64, 64) || m1.id ASC
			LIMIT 1
		`, chatID).Scan(&whisperTimestamp)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return whisperTimestamp, true, nil
}

// EmojiReactionsByChatID returns the emoji reactions for the queried messages, up to a maximum of 100, as it's a potentially unbound number.
// NOTE: This is not completely accurate, as the messages in the database might have change since the last call to `MessageByChatID`.
func (db sqlitePersistence) EmojiReactionsByChatID(chatID string, currCursor string, limit int) ([]*EmojiReaction, error) {
	cursorWhere := ""
	if currCursor != "" {
		cursorWhere = fmt.Sprintf("AND %s <= ?", cursor) //nolint: goconst
	}
	args := []interface{}{chatID, chatID}
	if currCursor != "" {
		args = append(args, currCursor)
	}
	args = append(args, limit)
	// NOTE: We match against local_chat_id for security reasons.
	// As a user could potentially send an emoji reaction for a one to
	// one/group chat that has no access to.
	// We also limit the number of emoji to a reasonable number (1000)
	// for now, as we don't want the client to choke on this.
	// The issue is that your own emoji might not be returned in such cases,
	// allowing the user to react to a post multiple times.
	// Jakubgs: Returning the whole list seems like a real overkill.
	// This will get very heavy in threads that have loads of reactions on loads of messages.
	// A more sensible response would just include a count and a bool telling you if you are in the list.
	// nolint: gosec
	query := fmt.Sprintf(`
			SELECT
			    e.clock_value,
			    e.source,
			    e.emoji_id,
			    e.message_id,
			    e.chat_id,
			    e.local_chat_id,
			    e.retracted
			FROM
				emoji_reactions e
			WHERE NOT(e.retracted)
			AND
			e.local_chat_id = ?
			AND
			e.message_id IN
			(SELECT id FROM user_messages m1 WHERE NOT(m1.hide) AND m1.local_chat_id = ? %s
			ORDER BY %s DESC LIMIT ?)
			LIMIT 1000
		`, cursorWhere, cursor)

	rows, err := db.db.Query(
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*EmojiReaction
	for rows.Next() {
		emojiReaction := NewEmojiReaction()
		err := rows.Scan(&emojiReaction.Clock,
			&emojiReaction.From,
			&emojiReaction.Type,
			&emojiReaction.MessageId,
			&emojiReaction.ChatId,
			&emojiReaction.LocalChatID,
			&emojiReaction.Retracted)
		if err != nil {
			return nil, err
		}

		result = append(result, emojiReaction)
	}

	return result, nil
}

// EmojiReactionsByChatIDMessageID returns the emoji reactions for the queried message.
func (db sqlitePersistence) EmojiReactionsByChatIDMessageID(chatID string, messageID string) ([]*EmojiReaction, error) {

	args := []interface{}{chatID, messageID}
	query := `SELECT
			    e.clock_value,
			    e.source,
			    e.emoji_id,
			    e.message_id,
			    e.chat_id,
			    e.local_chat_id,
			    e.retracted
			FROM
				emoji_reactions e
			WHERE NOT(e.retracted)
			AND
			e.local_chat_id = ?
			AND
			e.message_id = ?
			LIMIT 1000`

	rows, err := db.db.Query(
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*EmojiReaction
	for rows.Next() {
		emojiReaction := NewEmojiReaction()
		err := rows.Scan(&emojiReaction.Clock,
			&emojiReaction.From,
			&emojiReaction.Type,
			&emojiReaction.MessageId,
			&emojiReaction.ChatId,
			&emojiReaction.LocalChatID,
			&emojiReaction.Retracted)
		if err != nil {
			return nil, err
		}

		result = append(result, emojiReaction)
	}

	return result, nil
}

// EmojiReactionsByChatIDs returns the emoji reactions for the queried messages, up to a maximum of 100, as it's a potentially unbound number.
// NOTE: This is not completely accurate, as the messages in the database might have change since the last call to `MessageByChatID`.
func (db sqlitePersistence) EmojiReactionsByChatIDs(chatIDs []string, currCursor string, limit int) ([]*EmojiReaction, error) {
	cursorWhere := ""
	if currCursor != "" {
		cursorWhere = fmt.Sprintf("AND %s <= ?", cursor) //nolint: goconst
	}
	chatsLen := len(chatIDs)
	args := make([]interface{}, chatsLen*2)
	for i, v := range chatIDs {
		args[i] = v
	}
	for i, v := range chatIDs {
		args[chatsLen+i] = v
	}
	if currCursor != "" {
		args = append(args, currCursor)
	}
	args = append(args, limit)
	// NOTE: We match against local_chat_id for security reasons.
	// As a user could potentially send an emoji reaction for a one to
	// one/group chat that has no access to.
	// We also limit the number of emoji to a reasonable number (1000)
	// for now, as we don't want the client to choke on this.
	// The issue is that your own emoji might not be returned in such cases,
	// allowing the user to react to a post multiple times.
	// Jakubgs: Returning the whole list seems like a real overkill.
	// This will get very heavy in threads that have loads of reactions on loads of messages.
	// A more sensible response would just include a count and a bool telling you if you are in the list.
	// nolint: gosec
	query := fmt.Sprintf(`
			SELECT
			    e.clock_value,
			    e.source,
			    e.emoji_id,
			    e.message_id,
			    e.chat_id,
			    e.local_chat_id,
			    e.retracted
			FROM
				emoji_reactions e
			WHERE NOT(e.retracted)
			AND
			e.local_chat_id IN %s
			AND
			e.message_id IN
			(SELECT id FROM user_messages m WHERE NOT(m.hide) AND m.local_chat_id IN %s %s
			ORDER BY %s DESC LIMIT ?)
			LIMIT 1000
		`, "(?"+strings.Repeat(",?", chatsLen-1)+")", "(?"+strings.Repeat(",?", chatsLen-1)+")", cursorWhere, cursor)

	rows, err := db.db.Query(
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*EmojiReaction
	for rows.Next() {
		emojiReaction := NewEmojiReaction()
		err := rows.Scan(&emojiReaction.Clock,
			&emojiReaction.From,
			&emojiReaction.Type,
			&emojiReaction.MessageId,
			&emojiReaction.ChatId,
			&emojiReaction.LocalChatID,
			&emojiReaction.Retracted)
		if err != nil {
			return nil, err
		}

		result = append(result, emojiReaction)
	}

	return result, nil
}
func (db sqlitePersistence) SaveMessages(messages []*common.Message) (err error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	allFields := db.tableUserMessagesAllFields()
	valuesVector := strings.Repeat("?, ", db.tableUserMessagesAllFieldsCount()-1) + "?"
	query := "INSERT INTO user_messages(" + allFields + ") VALUES (" + valuesVector + ")" // nolint: gosec
	stmt, err := tx.Prepare(query)
	if err != nil {
		return
	}

	for _, msg := range messages {
		var allValues []interface{}
		allValues, err = db.tableUserMessagesAllValues(msg)
		if err != nil {
			return
		}

		_, err = stmt.Exec(allValues...)
		if err != nil {
			return
		}

		if msg.ContentType == protobuf.ChatMessage_BRIDGE_MESSAGE {
			err = db.saveBridgeMessage(tx, msg.GetBridgeMessage(), msg.ID)
			if err != nil {
				return
			}
			// handle replies
			err = db.findAndUpdateReplies(tx, msg.GetBridgeMessage().MessageID, msg.ID)
			if err != nil {
				return
			}
			parentMessageID := msg.GetBridgeMessage().ParentMessageID
			if parentMessageID != "" {
				err = db.findAndUpdateRepliedTo(tx, parentMessageID, msg.ID)
				if err != nil {
					return
				}
			}

		}
	}
	return
}

type insertPinMessagesQueries struct {
	selectStmt  string
	insertStmt  *sql.Stmt
	updateStmt  *sql.Stmt
	transaction *sql.Tx
}

func (db sqlitePersistence) buildPinMessageQueries() (*insertPinMessagesQueries, error) {
	tx, err := db.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	queries := &insertPinMessagesQueries{}
	// select
	queries.selectStmt = "SELECT clock_value FROM pin_messages WHERE id = ?"

	// insert
	allInsertFields := `id, message_id, whisper_timestamp, chat_id, local_chat_id, clock_value, pinned, pinned_by`
	insertValues := strings.Repeat("?, ", strings.Count(allInsertFields, ",")) + "?"
	insertQuery := "INSERT INTO pin_messages(" + allInsertFields + ") VALUES (" + insertValues + ")" // nolint: gosec
	insertStmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return nil, err
	}
	queries.insertStmt = insertStmt

	// update
	updateQuery := "UPDATE pin_messages SET pinned = ?, clock_value = ?, pinned_by = ? WHERE id = ?"
	updateStmt, err := tx.Prepare(updateQuery)
	if err != nil {
		return nil, err
	}
	queries.updateStmt = updateStmt
	queries.transaction = tx
	return queries, nil
}

func (db sqlitePersistence) SavePinMessages(messages []*common.PinMessage) (err error) {
	queries, err := db.buildPinMessageQueries()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = queries.transaction.Commit()
			return
		}
		// don't shadow original error
		_ = queries.transaction.Rollback()
	}()
	for _, message := range messages {
		_, err = db.savePinMessage(message, queries)
		if err != nil {
			return
		}
	}

	return
}

func (db sqlitePersistence) savePinMessage(message *common.PinMessage, queries *insertPinMessagesQueries) (inserted bool, err error) {
	tx := queries.transaction
	selectQuery := queries.selectStmt
	updateStmt := queries.updateStmt
	insertStmt := queries.insertStmt

	row := tx.QueryRow(selectQuery, message.ID)
	var existingClock uint64
	switch err = row.Scan(&existingClock); err {
	case sql.ErrNoRows:
		// not found, insert new record
		allValues := []interface{}{
			message.ID,
			message.MessageId,
			message.WhisperTimestamp,
			message.ChatId,
			message.LocalChatID,
			message.Clock,
			message.Pinned,
			message.From,
		}
		_, err = insertStmt.Exec(allValues...)
		if err != nil {
			return
		}
		inserted = true
	case nil:
		// found, update if current message is more recent, otherwise skip
		if existingClock < message.Clock {
			// update
			_, err = updateStmt.Exec(message.Pinned, message.Clock, message.From, message.ID)
			if err != nil {
				return
			}
			inserted = true
		}

	default:
		return
	}

	return
}

func (db sqlitePersistence) SavePinMessage(message *common.PinMessage) (inserted bool, err error) {
	queries, err := db.buildPinMessageQueries()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = queries.transaction.Commit()
			return
		}
		// don't shadow original error
		_ = queries.transaction.Rollback()
	}()
	return db.savePinMessage(message, queries)
}

func (db sqlitePersistence) DeleteMessage(id string) error {
	_, err := db.db.Exec(`DELETE FROM user_messages WHERE id = ?`, id)
	return err
}

func (db sqlitePersistence) DeleteMessages(ids []string) error {
	idsArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsArgs = append(idsArgs, id)
	}
	inVector := strings.Repeat("?, ", len(ids)-1) + "?"

	_, err := db.db.Exec("DELETE FROM user_messages WHERE id IN ("+inVector+")", idsArgs...) // nolint: gosec

	return err
}

func (db sqlitePersistence) HideMessage(id string) error {
	_, err := db.db.Exec(`UPDATE user_messages SET hide = 1, seen = 1 WHERE id = ?`, id)
	return err
}

// SetHideOnMessage set the hide flag, but not the seen flag, as it's needed by the client to understand whether the count should be updated
func (db sqlitePersistence) SetHideOnMessage(id string) error {
	_, err := db.db.Exec(`UPDATE user_messages SET hide = 1 WHERE id = ?`, id)
	return err
}

func (db sqlitePersistence) DeleteMessagesByCommunityID(id string) error {
	return db.deleteMessagesByCommunityID(id)
}

func (db sqlitePersistence) deleteMessagesByCommunityID(id string) (err error) {
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

	_, err = tx.Exec(`DELETE FROM user_messages WHERE community_id = ?`, id)
	if err != nil {
		return
	}

	_, err = tx.Exec(`DELETE FROM pin_messages WHERE community_id = ?`, id)

	return
}

func (db sqlitePersistence) DeleteMessagesByChatID(id string) error {
	return db.deleteMessagesByChatID(id, nil)
}

func (db sqlitePersistence) deleteMessagesByChatID(id string, tx *sql.Tx) (err error) {
	if tx == nil {
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
	}

	_, err = tx.Exec(`DELETE FROM user_messages WHERE local_chat_id = ?`, id)
	if err != nil {
		return
	}

	_, err = tx.Exec(`DELETE FROM pin_messages WHERE local_chat_id = ?`, id)

	return
}

func (db sqlitePersistence) deleteMessagesByChatIDAndClockValueLessThanOrEqual(id string, clock uint64, tx *sql.Tx) (unViewedMessages, unViewedMentions uint, err error) {
	if tx == nil {
		tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			return 0, 0, err
		}
		defer func() {
			if err == nil {
				err = tx.Commit()
				return
			}
			// don't shadow original error
			_ = tx.Rollback()
		}()
	}

	_, err = tx.Exec(`DELETE FROM user_messages WHERE local_chat_id = ? AND clock_value <= ?`, id, clock)
	if err != nil {
		return
	}

	_, err = tx.Exec(`DELETE FROM pin_messages WHERE local_chat_id = ? AND clock_value <= ?`, id, clock)
	if err != nil {
		return
	}

	_, err = tx.Exec(
		`UPDATE chats
		   SET unviewed_message_count =
		   (SELECT COUNT(1)
		   FROM user_messages
		   WHERE local_chat_id = ? AND seen = 0),
		   unviewed_mentions_count =
		   (SELECT COUNT(1)
		   FROM user_messages
		   WHERE local_chat_id = ? AND seen = 0 AND (mentioned OR replied)),
                   highlight = 0
		WHERE id = ?`, id, id, id)

	if err != nil {
		return 0, 0, err
	}

	err = tx.QueryRow(`SELECT unviewed_message_count, unviewed_mentions_count FROM chats
				WHERE id = ?`, id).Scan(&unViewedMessages, &unViewedMentions)

	return
}

func (db sqlitePersistence) MarkAllRead(chatID string, clock uint64) (int64, int64, error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	seenResult, err := tx.Exec(`UPDATE user_messages SET seen = 1 WHERE local_chat_id = ? AND seen = 0 AND clock_value <= ? AND not(mentioned) AND not(replied)`, chatID, clock)
	if err != nil {
		return 0, 0, err
	}

	seen, err := seenResult.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	mentionedOrRepliedResult, err := tx.Exec(`UPDATE user_messages SET seen = 1 WHERE local_chat_id = ? AND seen = 0 AND clock_value <= ? AND (mentioned OR replied)`, chatID, clock)
	if err != nil {
		return 0, 0, err
	}

	mentionedOrReplied, err := mentionedOrRepliedResult.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	_, err = tx.Exec(
		`UPDATE chats
		   SET unviewed_message_count = 0,
		   unviewed_mentions_count = 0,
                   highlight = 0
		WHERE id = ?`, chatID, chatID, chatID)

	if err != nil {
		return 0, 0, err
	}

	return (seen + mentionedOrReplied), mentionedOrReplied, nil
}

func (db sqlitePersistence) MarkAllReadMultiple(chatIDs []string) error {
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

	idsArgs := make([]interface{}, 0, len(chatIDs))
	for _, id := range chatIDs {
		idsArgs = append(idsArgs, id)
	}

	inVector := strings.Repeat("?, ", len(chatIDs)-1) + "?"

	q := "UPDATE user_messages SET seen = 1 WHERE local_chat_id IN (%s) AND seen != 1"
	q = fmt.Sprintf(q, inVector)
	_, err = tx.Exec(q, idsArgs...)
	if err != nil {
		return err
	}

	q = "UPDATE chats SET unviewed_mentions_count = 0, unviewed_message_count = 0, highlight = 0 WHERE id IN (%s)"
	q = fmt.Sprintf(q, inVector)
	_, err = tx.Exec(q, idsArgs...)
	return err
}

func (db sqlitePersistence) MarkMessagesSeen(chatID string, ids []string) (uint64, uint64, error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	idsArgs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsArgs = append(idsArgs, id)
	}

	inVector := strings.Repeat("?, ", len(ids)-1) + "?"
	q := "UPDATE user_messages SET seen = 1 WHERE NOT(seen) AND (mentioned OR replied) AND id IN (" + inVector + ")" // nolint: gosec
	_, err = tx.Exec(q, idsArgs...)
	if err != nil {
		return 0, 0, err
	}

	var countWithMentions uint64
	row := tx.QueryRow("SELECT changes();")
	if err := row.Scan(&countWithMentions); err != nil {
		return 0, 0, err
	}

	q = "UPDATE user_messages SET seen = 1 WHERE NOT(seen) AND NOT(mentioned) AND NOT(replied) AND id IN (" + inVector + ")" // nolint: gosec
	_, err = tx.Exec(q, idsArgs...)
	if err != nil {
		return 0, 0, err
	}

	var countNoMentions uint64
	row = tx.QueryRow("SELECT changes();")
	if err := row.Scan(&countNoMentions); err != nil {
		return 0, 0, err
	}

	// Update denormalized count
	_, err = tx.Exec(
		`UPDATE chats
              	SET unviewed_message_count =
		   (SELECT COUNT(1)
		   FROM user_messages
		   WHERE local_chat_id = ? AND seen = 0),
		   unviewed_mentions_count =
		   (SELECT COUNT(1)
		   FROM user_messages
		   WHERE local_chat_id = ? AND seen = 0 AND (mentioned OR replied)),
                   highlight = 0
		WHERE id = ?`, chatID, chatID, chatID)
	return countWithMentions + countNoMentions, countWithMentions, err
}

func (db sqlitePersistence) GetMessageIdsWithGreaterTimestamp(chatID string, messageID string) ([]string, error) {
	var err error
	var rows *sql.Rows
	query := "SELECT id FROM user_messages WHERE local_chat_id = ? AND timestamp >= (SELECT timestamp FROM user_messages WHERE id = ?)"

	rows, err = db.db.Query(query, chatID, messageID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string

	for rows.Next() {
		var messageID string
		err = rows.Scan(&messageID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, messageID)
	}

	return ids, nil
}

func (db sqlitePersistence) MarkMessageAsUnread(chatID string, messageID string) (uint64, uint64, error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	// TODO : Reduce number of queries for getting (total unread messages, total messages with mention)
	// The function expected result is a pair (total unread messages, total messages with mention)
	// Currently a 2 step operation is needed to obtain this pair
	_, err = tx.Exec(`UPDATE user_messages SET seen = 1 WHERE local_chat_id = ? AND NOT(seen)`, chatID)
	if err != nil {
		return 0, 0, err
	}

	_, err = tx.Exec(
		`UPDATE user_messages
			SET seen = 0
			WHERE local_chat_id = ?
			AND seen = 1
			AND (mentioned OR replied)
			AND timestamp >= (SELECT timestamp FROM user_messages WHERE id = ?)`, chatID, messageID)
	if err != nil {
		return 0, 0, err
	}

	var countWithMentions uint64
	row := tx.QueryRow("SELECT changes();")
	if err := row.Scan(&countWithMentions); err != nil {
		return 0, 0, err
	}

	_, err = tx.Exec(
		`UPDATE user_messages
			SET seen = 0
			WHERE local_chat_id = ?
			AND seen = 1
			AND NOT(mentioned OR replied)
			AND timestamp >= (SELECT timestamp FROM user_messages WHERE id = ?)`, chatID, messageID)
	if err != nil {
		return 0, 0, err
	}

	var countNoMentions uint64
	row = tx.QueryRow("SELECT changes();")
	if err := row.Scan(&countNoMentions); err != nil {
		return 0, 0, err
	}

	count := countWithMentions + countNoMentions

	_, err = tx.Exec(
		`UPDATE chats
            SET unviewed_message_count = ?, unviewed_mentions_count = ?,
			highlight = 0
			WHERE id = ?`, count, countWithMentions, chatID)

	return count, countWithMentions, err
}

func (db sqlitePersistence) UpdateMessageOutgoingStatus(id string, newOutgoingStatus string) error {
	_, err := db.db.Exec(`
		UPDATE user_messages
		SET outgoing_status = ?
		WHERE id = ? AND outgoing_status != ?
	`, newOutgoingStatus, id, common.OutgoingStatusDelivered)
	return err
}

// BlockContact updates a contact, deletes all the messages and 1-to-1 chat, updates the unread messages count and returns a map with the new count
func (db sqlitePersistence) BlockContact(contact *Contact, isDesktopFunc bool) ([]*Chat, error) {
	var chats []*Chat
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

	if !isDesktopFunc {
		// Delete messages
		_, err = tx.Exec(
			`DELETE
			FROM user_messages
			WHERE source = ?`,
			contact.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	// Update contact
	err = db.SaveContact(contact, tx)
	if err != nil {
		return nil, err
	}

	if !isDesktopFunc {
		// Delete one-to-one chat
		_, err = tx.Exec("DELETE FROM chats WHERE id = ?", contact.ID)
		if err != nil {
			return nil, err
		}
	}

	// Recalculate denormalized fields
	_, err = tx.Exec(`
		UPDATE chats
		SET
			unviewed_message_count = (SELECT COUNT(1) FROM user_messages WHERE seen = 0 AND local_chat_id = chats.id),
			unviewed_mentions_count = (SELECT COUNT(1) FROM user_messages WHERE seen = 0 AND local_chat_id = chats.id AND (mentioned OR replied))`)
	if err != nil {
		return nil, err
	}

	// return the updated chats
	chats, err = db.chats(tx)
	if err != nil {
		return nil, err
	}
	for _, c := range chats {
		var lastMessageID string
		row := tx.QueryRow(`SELECT id FROM user_messages WHERE local_chat_id = ? ORDER BY clock_value DESC LIMIT 1`, c.ID)
		switch err := row.Scan(&lastMessageID); err {

		case nil:
			message, err := db.messageByID(tx, lastMessageID)
			if err != nil {
				return nil, err
			}
			if message != nil {
				encodedMessage, err := json.Marshal(message)
				if err != nil {
					return nil, err
				}
				_, err = tx.Exec(`UPDATE chats SET last_message = ? WHERE id = ?`, encodedMessage, c.ID)
				if err != nil {
					return nil, err
				}
				c.LastMessage = message

			}

		case sql.ErrNoRows:
			// Reset LastMessage
			_, err = tx.Exec(`UPDATE chats SET last_message = NULL WHERE id = ?`, c.ID)
			if err != nil {
				return nil, err
			}
			c.LastMessage = nil
		default:
			return nil, err
		}
	}

	return chats, err
}

func (db sqlitePersistence) HasDiscordMessageAuthor(id string) (exists bool, err error) {
	err = db.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM discord_message_authors WHERE id = ?)`, id).Scan(&exists)
	return exists, err
}

func (db sqlitePersistence) HasDiscordMessageAuthorImagePayload(id string) (hasPayload bool, err error) {
	err = db.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM discord_message_authors WHERE id = ? AND avatar_image_payload NOT NULL)`, id).Scan(&hasPayload)
	return hasPayload, err
}

func (db sqlitePersistence) SaveDiscordMessageAuthor(author *protobuf.DiscordMessageAuthor) (err error) {
	stmt, err := db.db.Prepare(basicInsertDiscordMessageAuthorQuery)
	if err != nil {
		return
	}
	_, err = stmt.Exec(
		author.GetId(),
		author.GetName(),
		author.GetDiscriminator(),
		author.GetNickname(),
		author.GetAvatarUrl(),
		author.GetAvatarImagePayload(),
	)
	return
}

func (db sqlitePersistence) SaveDiscordMessageAuthors(authors []*protobuf.DiscordMessageAuthor) (err error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	stmt, err := tx.Prepare(basicInsertDiscordMessageAuthorQuery)
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, author := range authors {
		_, err = stmt.Exec(
			author.GetId(),
			author.GetName(),
			author.GetDiscriminator(),
			author.GetNickname(),
			author.GetAvatarUrl(),
			author.GetAvatarImagePayload(),
		)
		if err != nil {
			return
		}
	}
	return
}

func (db sqlitePersistence) UpdateDiscordMessageAuthorImage(authorID string, payload []byte) (err error) {
	query := "UPDATE discord_message_authors SET avatar_image_payload = ? WHERE id = ?"
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(payload, authorID)
	return
}

func (db sqlitePersistence) GetDiscordMessageAuthorImagePayloadByID(id string) ([]byte, error) {
	payload := make([]byte, 0)
	row := db.db.QueryRow("SELECT avatar_image_payload FROM discord_message_authors WHERE id = ?", id)
	err := row.Scan(&payload)
	return payload, err
}

func (db sqlitePersistence) GetDiscordMessageAuthorByID(id string) (*protobuf.DiscordMessageAuthor, error) {

	author := &protobuf.DiscordMessageAuthor{}

	row := db.db.QueryRow("SELECT id, name, discriminator, nickname, avatar_url FROM discord_message_authors WHERE id = ?", id)
	err := row.Scan(
		&author.Id,
		&author.Name,
		&author.Discriminator,
		&author.Nickname,
		&author.AvatarUrl)
	return author, err
}

func (db sqlitePersistence) SaveDiscordMessage(message *protobuf.DiscordMessage) (err error) {
	query := "INSERT OR REPLACE INTO discord_messages(id,type,timestamp,timestamp_edited,content,author_id, reference_message_id, reference_channel_id, reference_guild_id) VALUES (?,?,?,?,?,?,?,?,?)"
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		message.GetId(),
		message.GetType(),
		message.GetTimestamp(),
		message.GetTimestampEdited(),
		message.GetContent(),
		message.Author.GetId(),
		message.Reference.GetMessageId(),
		message.Reference.GetChannelId(),
		message.Reference.GetGuildId(),
	)
	return
}

func (db sqlitePersistence) SaveDiscordMessages(messages []*protobuf.DiscordMessage) (err error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	query := "INSERT OR REPLACE INTO discord_messages(id, author_id, type, timestamp, timestamp_edited, content, reference_message_id, reference_channel_id, reference_guild_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, msg := range messages {
		_, err = stmt.Exec(
			msg.GetId(),
			msg.Author.GetId(),
			msg.GetType(),
			msg.GetTimestamp(),
			msg.GetTimestampEdited(),
			msg.GetContent(),
			msg.Reference.GetMessageId(),
			msg.Reference.GetChannelId(),
			msg.Reference.GetGuildId(),
		)
		if err != nil {
			return
		}
	}
	return
}

func (db sqlitePersistence) HasDiscordMessageAttachmentPayload(id string, messageID string) (hasPayload bool, err error) {
	err = db.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM discord_message_attachments WHERE id = ? AND discord_message_id = ? AND payload NOT NULL)`, id, messageID).Scan(&hasPayload)
	return hasPayload, err
}

func (db sqlitePersistence) SaveDiscordMessageAttachments(attachments []*protobuf.DiscordMessageAttachment) (err error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
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

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO discord_message_attachments(id,discord_message_id,url,file_name,file_size_bytes,payload, content_type) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	if err != nil {
		return
	}

	for _, attachment := range attachments {
		_, err = stmt.Exec(
			attachment.GetId(),
			attachment.GetMessageId(),
			attachment.GetUrl(),
			attachment.GetFileName(),
			attachment.GetFileSizeBytes(),
			attachment.GetPayload(),
			attachment.GetContentType(),
		)
		if err != nil {
			return
		}
	}
	return
}

func (db sqlitePersistence) SaveEmojiReaction(emojiReaction *EmojiReaction) (err error) {
	query := "INSERT INTO emoji_reactions(id,clock_value,source,emoji_id,message_id,chat_id,local_chat_id,retracted) VALUES (?,?,?,?,?,?,?,?)"
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return
	}

	_, err = stmt.Exec(
		emojiReaction.ID(),
		emojiReaction.Clock,
		emojiReaction.From,
		emojiReaction.Type,
		emojiReaction.MessageId,
		emojiReaction.ChatId,
		emojiReaction.LocalChatID,
		emojiReaction.Retracted,
	)

	return
}

func (db sqlitePersistence) EmojiReactionByID(id string) (*EmojiReaction, error) {
	row := db.db.QueryRow(
		`SELECT
			    clock_value,
			    source,
			    emoji_id,
			    message_id,
			    chat_id,
			    local_chat_id,
			    retracted
			FROM
				emoji_reactions
			WHERE
				emoji_reactions.id = ?
		`, id)

	emojiReaction := NewEmojiReaction()
	err := row.Scan(&emojiReaction.Clock,
		&emojiReaction.From,
		&emojiReaction.Type,
		&emojiReaction.MessageId,
		&emojiReaction.ChatId,
		&emojiReaction.LocalChatID,
		&emojiReaction.Retracted,
	)

	switch err {
	case sql.ErrNoRows:
		return nil, common.ErrRecordNotFound
	case nil:
		return emojiReaction, nil
	default:
		return nil, err
	}
}

func (db sqlitePersistence) SaveInvitation(invitation *GroupChatInvitation) (err error) {
	query := "INSERT INTO group_chat_invitations(id,source,chat_id,message,state,clock) VALUES (?,?,?,?,?,?)"
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return
	}
	_, err = stmt.Exec(
		invitation.ID(),
		invitation.From,
		invitation.ChatId,
		invitation.IntroductionMessage,
		invitation.State,
		invitation.Clock,
	)

	return
}

func (db sqlitePersistence) GetGroupChatInvitations() (rst []*GroupChatInvitation, err error) {

	tx, err := db.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	bRows, err := tx.Query(`SELECT
			    source,
			    chat_id,
			    message,
			    state,
			    clock
			FROM
				group_chat_invitations`)
	if err != nil {
		return
	}
	defer bRows.Close()
	for bRows.Next() {
		invitation := GroupChatInvitation{}
		err = bRows.Scan(
			&invitation.From,
			&invitation.ChatId,
			&invitation.IntroductionMessage,
			&invitation.State,
			&invitation.Clock)
		if err != nil {
			return nil, err
		}
		rst = append(rst, &invitation)
	}

	return rst, nil
}

func (db sqlitePersistence) InvitationByID(id string) (*GroupChatInvitation, error) {
	row := db.db.QueryRow(
		`SELECT
			    source,
			    chat_id,
			    message,
			    state,
			    clock
			FROM
				group_chat_invitations
			WHERE
				group_chat_invitations.id = ?
		`, id)

	chatInvitations := NewGroupChatInvitation()
	err := row.Scan(&chatInvitations.From,
		&chatInvitations.ChatId,
		&chatInvitations.IntroductionMessage,
		&chatInvitations.State,
		&chatInvitations.Clock,
	)

	switch err {
	case sql.ErrNoRows:
		return nil, common.ErrRecordNotFound
	case nil:
		return chatInvitations, nil
	default:
		return nil, err
	}
}

// ClearHistory deletes all the messages for a chat and updates it's values
func (db sqlitePersistence) ClearHistory(chat *Chat, currentClockValue uint64) (err error) {
	var tx *sql.Tx

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
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
	err = db.clearHistory(chat, currentClockValue, tx, false)

	return
}

func (db sqlitePersistence) ClearHistoryFromSyncMessage(chat *Chat, currentClockValue uint64) (err error) {
	var tx *sql.Tx

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
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
	err = db.clearHistoryFromSyncMessage(chat, currentClockValue, tx)

	return
}

// Deactivate chat sets a chat as inactive and clear its history
func (db sqlitePersistence) DeactivateChat(chat *Chat, currentClockValue uint64, doClearHistory bool) (err error) {
	var tx *sql.Tx

	tx, err = db.db.BeginTx(context.Background(), &sql.TxOptions{})
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
	err = db.deactivateChat(chat, currentClockValue, tx, doClearHistory)

	return
}

func (db sqlitePersistence) deactivateChat(chat *Chat, currentClockValue uint64, tx *sql.Tx, doClearHistory bool) error {
	chat.Active = false
	err := db.saveChat(tx, *chat)
	if err != nil {
		return err
	}

	if !doClearHistory {
		return nil
	}
	return db.clearHistory(chat, currentClockValue, tx, true)
}

func (db sqlitePersistence) SaveDelete(deleteMessage *DeleteMessage) error {
	_, err := db.db.Exec(`INSERT INTO user_messages_deletes (clock, chat_id, message_id, source, id) VALUES(?,?,?,?,?)`, deleteMessage.Clock, deleteMessage.ChatId, deleteMessage.MessageId, deleteMessage.From, deleteMessage.ID)
	return err
}

func (db sqlitePersistence) GetDeletes(messageID string, from string) ([]*DeleteMessage, error) {
	rows, err := db.db.Query(`SELECT clock, chat_id, message_id, source, id FROM user_messages_deletes WHERE message_id = ? AND source = ? ORDER BY CLOCK DESC`, messageID, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*DeleteMessage
	for rows.Next() {
		d := NewDeleteMessage()
		err := rows.Scan(&d.Clock, &d.ChatId, &d.MessageId, &d.From, &d.ID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, d)
	}
	return messages, nil
}

func (db sqlitePersistence) SaveOrUpdateDeleteForMeMessage(deleteForMeMessage *protobuf.SyncDeleteForMeMessage) error {
	_, err := db.db.Exec(`INSERT OR REPLACE INTO user_messages_deleted_for_mes (clock, message_id)
    SELECT ?,? WHERE NOT EXISTS (SELECT 1 FROM user_messages_deleted_for_mes WHERE message_id = ? AND clock >= ?)`,
		deleteForMeMessage.Clock, deleteForMeMessage.MessageId, deleteForMeMessage.MessageId, deleteForMeMessage.Clock)
	return err
}

func (db sqlitePersistence) GetDeleteForMeMessagesByMessageID(messageID string) ([]*protobuf.SyncDeleteForMeMessage, error) {
	rows, err := db.db.Query(`SELECT clock, message_id FROM user_messages_deleted_for_mes WHERE message_id = ?`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*protobuf.SyncDeleteForMeMessage
	for rows.Next() {
		d := &protobuf.SyncDeleteForMeMessage{}
		err := rows.Scan(&d.Clock, &d.MessageId)
		if err != nil {
			return nil, err
		}
		messages = append(messages, d)
	}
	return messages, nil
}

func (db sqlitePersistence) GetDeleteForMeMessages() ([]*protobuf.SyncDeleteForMeMessage, error) {
	rows, err := db.db.Query(`SELECT clock, message_id FROM user_messages_deleted_for_mes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*protobuf.SyncDeleteForMeMessage
	for rows.Next() {
		d := &protobuf.SyncDeleteForMeMessage{}
		err := rows.Scan(&d.Clock, &d.MessageId)
		if err != nil {
			return nil, err
		}
		messages = append(messages, d)
	}
	return messages, nil
}

func (db sqlitePersistence) SaveEdit(editMessage *EditMessage) error {
	if editMessage == nil {
		return nil
	}

	_, err := db.db.Exec(`INSERT INTO user_messages_edits (clock, chat_id, message_id, text, source, id, unfurled_links, unfurled_status_links) VALUES(?,?,?,?,?,?,?,?)`, editMessage.Clock, editMessage.ChatId, editMessage.MessageId, editMessage.Text, editMessage.From, editMessage.ID, pq.Array(editMessage.UnfurledLinks), editMessage.UnfurledStatusLinks)
	return err
}

func (db sqlitePersistence) GetEdits(messageID string, from string) ([]*EditMessage, error) {
	rows, err := db.db.Query(`SELECT clock, chat_id, message_id, source, text, id, unfurled_links, unfurled_status_links FROM user_messages_edits WHERE message_id = ? AND source = ? ORDER BY CLOCK DESC`, messageID, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*EditMessage
	for rows.Next() {
		e := NewEditMessage()
		err := rows.Scan(&e.Clock, &e.ChatId, &e.MessageId, &e.From, &e.Text, &e.ID, pq.Array(&e.UnfurledLinks), &e.UnfurledStatusLinks)
		if err != nil {
			return nil, err
		}
		messages = append(messages, e)

	}
	return messages, nil
}

func (db sqlitePersistence) clearHistory(chat *Chat, currentClockValue uint64, tx *sql.Tx, deactivate bool) error {
	// Set deleted at clock value if it's not a public chat so that
	// old messages will be discarded, or if it's a straight clear history
	if !deactivate || (!chat.Public() && !chat.ProfileUpdates() && !chat.Timeline()) {
		if chat.LastMessage != nil && chat.LastMessage.Clock != 0 {
			chat.DeletedAtClockValue = chat.LastMessage.Clock
		}
		chat.DeletedAtClockValue = currentClockValue
	}

	// Reset synced-to/from
	syncedTo := uint32(currentClockValue / 1000)
	chat.SyncedTo = syncedTo
	chat.SyncedFrom = 0

	chat.LastMessage = nil
	chat.UnviewedMessagesCount = 0
	chat.UnviewedMentionsCount = 0
	chat.Highlight = true

	err := db.deleteMessagesByChatID(chat.ID, tx)
	if err != nil {
		return err
	}

	err = db.saveChat(tx, *chat)
	return err
}

func (db sqlitePersistence) clearHistoryFromSyncMessage(chat *Chat, clearedAt uint64, tx *sql.Tx) error {
	chat.DeletedAtClockValue = clearedAt

	// Reset synced-to/from
	syncedTo := uint32(clearedAt / 1000)
	chat.SyncedTo = syncedTo
	chat.SyncedFrom = 0

	unViewedMessagesCount, unViewedMentionsCount, err := db.deleteMessagesByChatIDAndClockValueLessThanOrEqual(chat.ID, clearedAt, tx)
	if err != nil {
		return err
	}

	chat.UnviewedMessagesCount = unViewedMessagesCount
	chat.UnviewedMentionsCount = unViewedMentionsCount

	if chat.LastMessage != nil && chat.LastMessage.Clock <= clearedAt {
		chat.LastMessage = nil
	}

	err = db.saveChat(tx, *chat)
	return err
}

func (db sqlitePersistence) SetContactRequestState(id string, state common.ContactRequestState) error {
	_, err := db.db.Exec(`UPDATE user_messages SET contact_request_state = ? WHERE id = ?`, state, id)
	return err
}

func getUpdatedChatMessagePayload(originalMessage *protobuf.DiscordMessage, attachmentMessage *protobuf.DiscordMessage) *protobuf.ChatMessage_DiscordMessage {
	originalMessage.Attachments = append(originalMessage.Attachments, attachmentMessage.Attachments...)
	return &protobuf.ChatMessage_DiscordMessage{
		DiscordMessage: originalMessage,
	}
}

func getMessageFromScanRows(db sqlitePersistence, rows *sql.Rows) (*common.Message, error) {
	var msg *common.Message

	for rows.Next() {
		// There's a possibility of multiple rows per message if the
		// message has a discordMessage and the discordMessage has multiple
		// attachments
		//
		// Hence, we make sure we're aggregating all attachments on a single
		// common.Message
		message := common.NewMessage()
		err := db.tableUserMessagesScanAllFields(rows, message)
		if err != nil {
			return nil, err
		}

		if msg == nil {
			msg = message
		} else if discordMessage := msg.GetDiscordMessage(); discordMessage != nil {
			msg.Payload = getUpdatedChatMessagePayload(discordMessage, message.GetDiscordMessage())
		}
	}
	if msg == nil {
		return nil, common.ErrRecordNotFound
	}
	return msg, nil
}

type HasClocks interface {
	GetClock(i int) uint64
}

func SortByClock(msgs HasClocks) {
	sort.Slice(msgs, func(i, j int) bool {
		return msgs.GetClock(j) < msgs.GetClock(i)
	})
}

func getMessagesFromScanRows(db sqlitePersistence, rows *sql.Rows, withCursor bool) ([]*common.Message, error) {
	messageIdx := make(map[string]*common.Message, 0)
	var messages common.Messages
	for rows.Next() {
		// There's a possibility of multiple rows per message if the
		// message has a discordMessage and the discordMessage has multiple
		// attachments
		//
		// Hence, we make sure we're aggregating all attachments on a single
		// common.Message
		message := common.NewMessage()

		if withCursor {
			var cursor string
			if err := db.tableUserMessagesScanAllFields(rows, message, &cursor); err != nil {
				return nil, err
			}
		} else {
			if err := db.tableUserMessagesScanAllFields(rows, message); err != nil {
				return nil, err
			}
		}

		if msg, ok := messageIdx[message.ID]; !ok {
			messageIdx[message.ID] = message
			messages = append(messages, message)
		} else if discordMessage := msg.GetDiscordMessage(); discordMessage != nil {
			msg.Payload = getUpdatedChatMessagePayload(discordMessage, message.GetDiscordMessage())
		}
	}

	SortByClock(messages)

	return messages, nil
}

func getMessagesAndCursorsFromScanRows(db sqlitePersistence, rows *sql.Rows) ([]*common.Message, []string, error) {

	var cursors []string
	var messages common.Messages
	messageIdx := make(map[string]*common.Message, 0)
	for rows.Next() {
		// There's a possibility of multiple rows per message if the
		// message has a discordMessage and the discordMessage has multiple
		// attachments
		//
		// Hence, we make sure we're aggregating all attachments on a single
		// common.Message

		var cursor string
		message := common.NewMessage()
		if err := db.tableUserMessagesScanAllFields(rows, message, &cursor); err != nil {
			return nil, nil, err
		}

		if msg, ok := messageIdx[message.ID]; !ok {
			messageIdx[message.ID] = message
			cursors = append(cursors, cursor)
			messages = append(messages, message)
		} else if discordMessage := msg.GetDiscordMessage(); discordMessage != nil {
			msg.Payload = getUpdatedChatMessagePayload(discordMessage, message.GetDiscordMessage())
		}
	}

	SortByClock(messages)

	return messages, cursors, nil
}

func getPinnedMessagesAndCursorsFromScanRows(db sqlitePersistence, rows *sql.Rows) ([]*common.PinnedMessage, []string, error) {

	var cursors []string
	var messages common.PinnedMessages
	messageIdx := make(map[string]*common.PinnedMessage, 0)

	for rows.Next() {
		var (
			pinnedAt uint64
			pinnedBy string
			cursor   string
		)
		message := common.NewMessage()
		if err := db.tableUserMessagesScanAllFields(rows, message, &pinnedAt, &pinnedBy, &cursor); err != nil {
			return nil, nil, err
		}
		if msg, ok := messageIdx[message.ID]; !ok {
			pinnedMessage := &common.PinnedMessage{
				Message:  message,
				PinnedAt: pinnedAt,
				PinnedBy: pinnedBy,
			}
			messageIdx[message.ID] = pinnedMessage
			messages = append(messages, pinnedMessage)
			cursors = append(cursors, cursor)
		} else if discordMessage := msg.Message.GetDiscordMessage(); discordMessage != nil {
			msg.Message.Payload = getUpdatedChatMessagePayload(discordMessage, message.GetDiscordMessage())
		}
	}

	SortByClock(messages)

	return messages, cursors, nil
}

func (db sqlitePersistence) saveBridgeMessage(tx *sql.Tx, message *protobuf.BridgeMessage, userMessageID string) (err error) {
	query := "INSERT INTO bridge_messages(user_messages_id,bridge_name,user_name,user_avatar,user_id,content,message_id,parent_message_id) VALUES (?,?,?,?,?,?,?,?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		userMessageID,
		message.GetBridgeName(),
		message.GetUserName(),
		message.GetUserAvatar(),
		message.GetUserID(),
		message.GetContent(),
		message.GetMessageID(),
		message.GetParentMessageID(),
	)
	return
}

func (db sqlitePersistence) GetCommunityMemberMessagesToDelete(member string, communityID string) ([]*protobuf.DeleteCommunityMemberMessage, error) {
	rows, err := db.db.Query(`SELECT m.id, m.chat_id FROM user_messages as m
		INNER JOIN chats AS ch ON ch.id = m.chat_id AND ch.community_id = ?
		WHERE m.source = ?`, communityID, member)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := []*protobuf.DeleteCommunityMemberMessage{}

	for rows.Next() {
		removeMsgsInfo := &protobuf.DeleteCommunityMemberMessage{}
		err = rows.Scan(&removeMsgsInfo.Id, &removeMsgsInfo.ChatId)
		if err != nil {
			return nil, err
		}
		result = append(result, removeMsgsInfo)
	}

	return result, nil
}

// Finds status messages id which are replies for bridgeMessageID
func (db sqlitePersistence) findStatusMessageIdsReplies(tx *sql.Tx, bridgeMessageID string) ([]string, error) {
	rows, err := tx.Query(`SELECT user_messages_id FROM bridge_messages WHERE parent_message_id = ?`, bridgeMessageID)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	var statusMessageIDs []string
	for rows.Next() {
		var statusMessageID string
		err = rows.Scan(&statusMessageID)
		if err != nil {
			return []string{}, err
		}
		statusMessageIDs = append(statusMessageIDs, statusMessageID)
	}
	return statusMessageIDs, nil
}

// Finds status messages id which are replies for bridgeMessageID
func (db sqlitePersistence) findStatusMessageIdsRepliedTo(tx *sql.Tx, parentMessageID string) (string, error) {
	rows, err := tx.Query(`SELECT user_messages_id FROM bridge_messages WHERE message_id = ?`, parentMessageID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		var statusMessageID string
		err = rows.Scan(&statusMessageID)
		if err != nil {
			return "", err
		}
		return statusMessageID, nil
	}
	return "", nil
}

func (db sqlitePersistence) updateStatusMessagesWithResponse(tx *sql.Tx, statusMessagesToUpdate []string, responseValue string) error {
	sql := "UPDATE user_messages SET response_to = ? WHERE id IN (?" + strings.Repeat(",?", len(statusMessagesToUpdate)-1) + ")"
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	args := make([]interface{}, 0, len(statusMessagesToUpdate)+1)
	args = append(args, responseValue)
	for _, msgToUpdate := range statusMessagesToUpdate {
		args = append(args, msgToUpdate)
	}
	_, err = stmt.Exec(args...)
	return err
}

// Finds if there are any messages that are replies to that message (in case replies were received earlier)
func (db sqlitePersistence) findAndUpdateReplies(tx *sql.Tx, bridgeMessageID string, statusMessageID string) error {
	replyMessageIds, err := db.findStatusMessageIdsReplies(tx, bridgeMessageID)
	if err != nil {
		return err
	}
	if len(replyMessageIds) == 0 {
		return nil
	}
	return db.updateStatusMessagesWithResponse(tx, replyMessageIds, statusMessageID)
}

func (db sqlitePersistence) findAndUpdateRepliedTo(tx *sql.Tx, discordParentMessageID string, statusMessageID string) error {
	repliedMessageID, err := db.findStatusMessageIdsRepliedTo(tx, discordParentMessageID)
	if err != nil {
		return err
	}
	if repliedMessageID == "" {
		return nil
	}
	return db.updateStatusMessagesWithResponse(tx, []string{statusMessageID}, repliedMessageID)
}

func (db sqlitePersistence) GetCommunityMemberAllMessages(member string, communityID string) ([]*common.Message, error) {
	additionalRequestData := "INNER JOIN chats AS ch ON ch.id = m1.chat_id AND ch.community_id = ? WHERE m1.source = ?"
	query := db.buildMessagesQueryWithAdditionalFields("", additionalRequestData)

	rows, err := db.db.Query(query, communityID, member)

	if err != nil {
		if err == sql.ErrNoRows {
			return []*common.Message{}, nil
		}

		return nil, err
	}

	return getMessagesFromScanRows(db, rows, false)
}
