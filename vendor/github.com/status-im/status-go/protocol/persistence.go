package protocol

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mat/besticon/besticon"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/images"
	userimage "github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/identity"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/services/browsers"
)

var (
	// ErrMsgAlreadyExist returned if msg already exist.
	ErrMsgAlreadyExist = errors.New("message with given ID already exist")
	HoursInTwoWeeks    = 336
)

// sqlitePersistence wrapper around sql db with operations common for a client.
type sqlitePersistence struct {
	*common.RawMessagesPersistence
	db *sql.DB
}

func newSQLitePersistence(db *sql.DB) *sqlitePersistence {
	return &sqlitePersistence{common.NewRawMessagesPersistence(db), db}
}

func (db sqlitePersistence) SaveChat(chat Chat) error {
	err := chat.Validate()
	if err != nil {
		return err
	}
	return db.saveChat(nil, chat)
}

func (db sqlitePersistence) SaveChats(chats []*Chat) error {
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

	for _, chat := range chats {
		err := db.saveChat(tx, *chat)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db sqlitePersistence) SaveContacts(contacts []*Contact) error {
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

	for _, contact := range contacts {
		err := db.SaveContact(contact, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db sqlitePersistence) saveChat(tx *sql.Tx, chat Chat) error {
	var err error
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

	// Encode members
	var encodedMembers bytes.Buffer
	memberEncoder := gob.NewEncoder(&encodedMembers)

	if err := memberEncoder.Encode(chat.Members); err != nil {
		return err
	}

	// Encode membership updates
	var encodedMembershipUpdates bytes.Buffer
	membershipUpdatesEncoder := gob.NewEncoder(&encodedMembershipUpdates)

	if err := membershipUpdatesEncoder.Encode(chat.MembershipUpdates); err != nil {
		return err
	}

	// encode last message
	var encodedLastMessage []byte
	if chat.LastMessage != nil {
		encodedLastMessage, err = json.Marshal(chat.LastMessage)
		if err != nil {
			return err
		}
	}

	// Insert record
	stmt, err := tx.Prepare(`INSERT INTO chats(id, name, color, emoji, active, type, timestamp,  deleted_at_clock_value, unviewed_message_count, unviewed_mentions_count, last_clock_value, last_message, members, membership_updates, muted, muted_till, invitation_admin, profile, community_id, joined, synced_from, synced_to, first_message_timestamp, description, highlight, read_messages_at_clock_value, received_invitation_admin, image_payload)
	    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?, ?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var imagePayload []byte
	if len(chat.Base64Image) > 0 {
		imagePayload, err = userimage.GetPayloadFromURI(chat.Base64Image)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec(
		chat.ID,
		chat.Name,
		chat.Color,
		chat.Emoji,
		chat.Active,
		chat.ChatType,
		chat.Timestamp,
		chat.DeletedAtClockValue,
		chat.UnviewedMessagesCount,
		chat.UnviewedMentionsCount,
		chat.LastClockValue,
		encodedLastMessage,
		encodedMembers.Bytes(),
		encodedMembershipUpdates.Bytes(),
		chat.Muted,
		chat.MuteTill,
		chat.InvitationAdmin,
		chat.Profile,
		chat.CommunityID,
		chat.Joined,
		chat.SyncedFrom,
		chat.SyncedTo,
		chat.FirstMessageTimestamp,
		chat.Description,
		chat.Highlight,
		chat.ReadMessagesAtClockValue,
		chat.ReceivedInvitationAdmin,
		imagePayload,
	)

	if err != nil {
		return err
	}

	return err
}

func (db sqlitePersistence) SetSyncTimestamps(syncedFrom, syncedTo uint32, chatID string) error {
	_, err := db.db.Exec(`UPDATE chats SET synced_from = ?, synced_to = ? WHERE id = ?`, syncedFrom, syncedTo, chatID)
	return err
}

func (db sqlitePersistence) DeleteChat(chatID string) (err error) {
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

	_, err = tx.Exec("DELETE FROM chats WHERE id = ?", chatID)
	if err != nil {
		return
	}

	_, err = tx.Exec(`DELETE FROM user_messages WHERE local_chat_id = ?`, chatID)
	return
}

func (db sqlitePersistence) MuteChat(chatID string, mutedTill time.Time) error {
	mutedTillFormatted := mutedTill.Format(time.RFC3339)
	_, err := db.db.Exec("UPDATE chats SET muted = 1, muted_till = ? WHERE id = ?", mutedTillFormatted, chatID)
	return err
}

func (db sqlitePersistence) UnmuteChat(chatID string) error {
	_, err := db.db.Exec("UPDATE chats SET muted = 0, muted_till = 0 WHERE id = ?", chatID)
	return err
}

func (db sqlitePersistence) Chats() ([]*Chat, error) {
	return db.chats(nil)
}

func (db sqlitePersistence) chats(tx *sql.Tx) (chats []*Chat, err error) {
	if tx == nil {
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
	}

	rows, err := tx.Query(`
		SELECT
			chats.id,
			chats.name,
			chats.color,
			chats.emoji,
			chats.active,
			chats.type,
			chats.timestamp,
			chats.deleted_at_clock_value,
			chats.read_messages_at_clock_value,
			chats.unviewed_message_count,
			chats.unviewed_mentions_count,
			chats.last_clock_value,
			chats.last_message,
			chats.members,
			chats.membership_updates,
			chats.muted,
			chats.muted_till,
			chats.invitation_admin,
			chats.profile,
			chats.community_id,
			chats.joined,
			chats.synced_from,
			chats.synced_to,
			chats.first_message_timestamp,
		    chats.description,
			contacts.alias,
			chats.highlight,
			chats.received_invitation_admin,
			chats.image_payload
		FROM chats LEFT JOIN contacts ON chats.id = contacts.id
		ORDER BY chats.timestamp DESC
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			alias                    sql.NullString
			invitationAdmin          sql.NullString
			profile                  sql.NullString
			syncedFrom               sql.NullInt64
			syncedTo                 sql.NullInt64
			firstMessageTimestamp    sql.NullInt64
			MuteTill                 sql.NullTime
			chat                     Chat
			encodedMembers           []byte
			encodedMembershipUpdates []byte
			lastMessageBytes         []byte
			imagePayload             []byte
		)
		err = rows.Scan(
			&chat.ID,
			&chat.Name,
			&chat.Color,
			&chat.Emoji,
			&chat.Active,
			&chat.ChatType,
			&chat.Timestamp,
			&chat.DeletedAtClockValue,
			&chat.ReadMessagesAtClockValue,
			&chat.UnviewedMessagesCount,
			&chat.UnviewedMentionsCount,
			&chat.LastClockValue,
			&lastMessageBytes,
			&encodedMembers,
			&encodedMembershipUpdates,
			&chat.Muted,
			&MuteTill,
			&invitationAdmin,
			&profile,
			&chat.CommunityID,
			&chat.Joined,
			&syncedFrom,
			&syncedTo,
			&firstMessageTimestamp,
			&chat.Description,
			&alias,
			&chat.Highlight,
			&chat.ReceivedInvitationAdmin,
			&imagePayload,
		)

		if err != nil {
			return
		}

		if invitationAdmin.Valid {
			chat.InvitationAdmin = invitationAdmin.String
		}

		if profile.Valid {
			chat.Profile = profile.String
		}

		// Restore members
		membersDecoder := gob.NewDecoder(bytes.NewBuffer(encodedMembers))
		err = membersDecoder.Decode(&chat.Members)
		if err != nil {
			return
		}

		// Restore membership updates
		membershipUpdatesDecoder := gob.NewDecoder(bytes.NewBuffer(encodedMembershipUpdates))
		err = membershipUpdatesDecoder.Decode(&chat.MembershipUpdates)
		if err != nil {
			return
		}

		if syncedFrom.Valid {
			chat.SyncedFrom = uint32(syncedFrom.Int64)
		}

		if syncedTo.Valid {
			chat.SyncedTo = uint32(syncedTo.Int64)
		}

		if firstMessageTimestamp.Valid {
			chat.FirstMessageTimestamp = uint32(firstMessageTimestamp.Int64)
		}

		if imagePayload != nil {
			base64Image, err := userimage.GetPayloadDataURI(imagePayload)
			if err == nil {
				chat.Base64Image = base64Image
			}
		}

		// Restore last message
		if lastMessageBytes != nil {
			message := common.NewMessage()
			if err = json.Unmarshal(lastMessageBytes, message); err != nil {
				return
			}
			chat.LastMessage = message
		}

		if MuteTill.Valid {
			chat.MuteTill = MuteTill.Time
		}

		chat.Alias = alias.String

		chats = append(chats, &chat)
	}

	return
}

func (db sqlitePersistence) Chat(chatID string) (*Chat, error) {
	var (
		chat                     Chat
		encodedMembers           []byte
		encodedMembershipUpdates []byte
		lastMessageBytes         []byte
		invitationAdmin          sql.NullString
		profile                  sql.NullString
		syncedFrom               sql.NullInt64
		syncedTo                 sql.NullInt64
		firstMessageTimestamp    sql.NullInt64
		MuteTill                 sql.NullTime
		imagePayload             []byte
	)

	var unviewedMessagesCount int
	var unviewedMentionsCount int

	err := db.db.QueryRow(`
		SELECT
			id,
			name,
			color,
			emoji,
			active,
			type,
			timestamp,
			read_messages_at_clock_value,
			deleted_at_clock_value,
			unviewed_message_count,
			unviewed_mentions_count,
			last_clock_value,
			last_message,
			members,
			membership_updates,
			muted,
			muted_till,
			invitation_admin,
			profile,
			community_id,
			joined,
			description,
			highlight,
			received_invitation_admin,
			synced_from,
			synced_to,
			first_message_timestamp,
			image_payload
		FROM chats
		WHERE id = ?
	`, chatID).Scan(&chat.ID,
		&chat.Name,
		&chat.Color,
		&chat.Emoji,
		&chat.Active,
		&chat.ChatType,
		&chat.Timestamp,
		&chat.ReadMessagesAtClockValue,
		&chat.DeletedAtClockValue,
		&unviewedMessagesCount,
		&unviewedMentionsCount,
		&chat.LastClockValue,
		&lastMessageBytes,
		&encodedMembers,
		&encodedMembershipUpdates,
		&chat.Muted,
		&MuteTill,
		&invitationAdmin,
		&profile,
		&chat.CommunityID,
		&chat.Joined,
		&chat.Description,
		&chat.Highlight,
		&chat.ReceivedInvitationAdmin,
		&syncedFrom,
		&syncedTo,
		&firstMessageTimestamp,
		&imagePayload,
	)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		if syncedFrom.Valid {
			chat.SyncedFrom = uint32(syncedFrom.Int64)
		}
		if syncedTo.Valid {
			chat.SyncedTo = uint32(syncedTo.Int64)
		}
		if firstMessageTimestamp.Valid {
			chat.FirstMessageTimestamp = uint32(firstMessageTimestamp.Int64)
		}
		if invitationAdmin.Valid {
			chat.InvitationAdmin = invitationAdmin.String
		}
		if profile.Valid {
			chat.Profile = profile.String
		}

		// Set UnviewedCounts and make sure they are above 0
		// Since Chat's UnviewedMessagesCount is uint and the SQL column is INT, it can create a discrepancy
		if unviewedMessagesCount < 0 {
			unviewedMessagesCount = 0
		}
		if unviewedMentionsCount < 0 {
			unviewedMentionsCount = 0
		}
		chat.UnviewedMessagesCount = uint(unviewedMessagesCount)
		chat.UnviewedMentionsCount = uint(unviewedMentionsCount)

		// Restore members
		membersDecoder := gob.NewDecoder(bytes.NewBuffer(encodedMembers))
		err = membersDecoder.Decode(&chat.Members)
		if err != nil {
			return nil, err
		}

		// Restore membership updates
		membershipUpdatesDecoder := gob.NewDecoder(bytes.NewBuffer(encodedMembershipUpdates))
		err = membershipUpdatesDecoder.Decode(&chat.MembershipUpdates)
		if err != nil {
			return nil, err
		}

		// Restore last message
		if lastMessageBytes != nil {
			message := common.NewMessage()
			if err = json.Unmarshal(lastMessageBytes, message); err != nil {
				return nil, err
			}
			chat.LastMessage = message
		}

		if imagePayload != nil {
			base64Image, err := userimage.GetPayloadDataURI(imagePayload)
			if err == nil {
				chat.Base64Image = base64Image
			}
		}

		if MuteTill.Valid {
			chat.MuteTill = MuteTill.Time
		}

		return &chat, nil
	}

	return nil, err

}

func (db sqlitePersistence) Contacts() ([]*Contact, error) {
	allContacts := make(map[string]*Contact)

	rows, err := db.db.Query(`
		SELECT
			c.id,
			c.address,
			v.name,
			v.verified,
			c.alias,
			c.display_name,
			c.identicon,
			c.last_updated,
			c.last_updated_locally,
			c.blocked,
			c.removed,
			c.bio,
			c.local_nickname,
			c.contact_request_state,
                        c.contact_request_local_clock,
			c.contact_request_remote_state,
			c.contact_request_remote_clock,
			i.image_type,
			i.payload,
                        i.clock_value,
			COALESCE(c.verification_status, 0) as verification_status,
			COALESCE(t.trust_status, 0) as trust_status
		FROM contacts c
		LEFT JOIN chat_identity_contacts i ON c.id = i.contact_id
		LEFT JOIN ens_verification_records v ON c.id = v.public_key
		LEFT JOIN trusted_users t ON c.id = t.id;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var (
			contact                   Contact
			nickname                  sql.NullString
			contactRequestLocalState  sql.NullInt64
			contactRequestLocalClock  sql.NullInt64
			contactRequestRemoteState sql.NullInt64
			contactRequestRemoteClock sql.NullInt64
			displayName               sql.NullString
			imageType                 sql.NullString
			ensName                   sql.NullString
			ensVerified               sql.NullBool
			blocked                   sql.NullBool
			removed                   sql.NullBool
			bio                       sql.NullString
			lastUpdatedLocally        sql.NullInt64
			identityImageClock        sql.NullInt64
			imagePayload              []byte
		)

		contact.Images = make(map[string]images.IdentityImage)

		err := rows.Scan(
			&contact.ID,
			&contact.Address,
			&ensName,
			&ensVerified,
			&contact.Alias,
			&displayName,
			&contact.Identicon,
			&contact.LastUpdated,
			&lastUpdatedLocally,
			&blocked,
			&removed,
			&bio,
			&nickname,
			&contactRequestLocalState,
			&contactRequestLocalClock,
			&contactRequestRemoteState,
			&contactRequestRemoteClock,
			&imageType,
			&imagePayload,
			&identityImageClock,
			&contact.VerificationStatus,
			&contact.TrustStatus,
		)
		if err != nil {
			return nil, err
		}

		if nickname.Valid {
			contact.LocalNickname = nickname.String
		}

		if bio.Valid {
			contact.Bio = bio.String
		}

		if contactRequestLocalState.Valid {
			contact.ContactRequestLocalState = ContactRequestState(contactRequestLocalState.Int64)
		}

		if contactRequestLocalClock.Valid {
			contact.ContactRequestLocalClock = uint64(contactRequestLocalClock.Int64)
		}

		if contactRequestRemoteState.Valid {
			contact.ContactRequestRemoteState = ContactRequestState(contactRequestRemoteState.Int64)
		}

		if contactRequestRemoteClock.Valid {
			contact.ContactRequestRemoteClock = uint64(contactRequestRemoteClock.Int64)
		}

		if displayName.Valid {
			contact.DisplayName = displayName.String
		}

		if ensName.Valid {
			contact.EnsName = ensName.String
		}

		if ensVerified.Valid {
			contact.ENSVerified = ensVerified.Bool
		}

		if blocked.Valid {
			contact.Blocked = blocked.Bool
		}

		if removed.Valid {
			contact.Removed = removed.Bool
		}

		if lastUpdatedLocally.Valid {
			contact.LastUpdatedLocally = uint64(lastUpdatedLocally.Int64)
		}

		previousContact, ok := allContacts[contact.ID]
		if !ok {
			if imageType.Valid {
				contact.Images[imageType.String] = images.IdentityImage{Name: imageType.String, Payload: imagePayload, Clock: uint64(identityImageClock.Int64)}
			}

			allContacts[contact.ID] = &contact

		} else if imageType.Valid {
			previousContact.Images[imageType.String] = images.IdentityImage{Name: imageType.String, Payload: imagePayload, Clock: uint64(identityImageClock.Int64)}
			allContacts[contact.ID] = previousContact

		}
	}

	// Read social links
	for _, contact := range allContacts {
		rows, err := db.db.Query(`SELECT link_text, link_url FROM chat_identity_social_links WHERE chat_id = ?`, contact.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				text sql.NullString
				url  sql.NullString
			)
			err := rows.Scan(
				&text, &url,
			)
			if err != nil {
				return nil, err
			}

			link := &identity.SocialLink{}
			if text.Valid {
				link.Text = text.String
			}
			if url.Valid {
				link.URL = url.String
			}
			contact.SocialLinks = append(contact.SocialLinks, link)
		}
	}

	var response []*Contact
	for key := range allContacts {
		response = append(response, allContacts[key])

	}
	return response, nil
}

func extractImageTypes(images map[string]*protobuf.IdentityImage) []string {
	uniqueImageTypesMap := make(map[string]struct{})
	for key := range images {
		uniqueImageTypesMap[key] = struct{}{}
	}

	var uniqueImageTypes []string
	for key := range uniqueImageTypesMap {
		uniqueImageTypes = append(uniqueImageTypes, key)
	}

	return uniqueImageTypes
}

func generatePlaceholders(count int) string {
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ", ")
}

func (db sqlitePersistence) UpdateContactChatIdentity(contactID string, chatIdentity *protobuf.ChatIdentity) (clockUpdated, imagesUpdated bool, err error) {
	if chatIdentity.Clock == 0 {
		return false, false, errors.New("clock value unset")
	}

	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return false, false, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	extractedImageTypes := extractImageTypes(chatIdentity.Images)

	query := "DELETE FROM chat_identity_contacts WHERE contact_id = ?"
	if len(extractedImageTypes) > 0 {
		query += " AND image_type NOT IN (" + generatePlaceholders(len(extractedImageTypes)) + ")"
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return false, false, err
	}
	defer stmt.Close()

	args := make([]interface{}, len(extractedImageTypes)+1)
	args[0] = contactID
	for i, v := range extractedImageTypes {
		args[i+1] = v
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		return false, false, err
	}

	imagesUpdated = false
	if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected > 0 {
		imagesUpdated = true
	}

	updateClock := func() (updated bool, err error) {
		var newerClockEntryExists bool
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM chat_identity_last_received WHERE chat_id = ? AND clock_value >= ?)`, contactID, chatIdentity.Clock).Scan(&newerClockEntryExists)
		if err != nil {
			return false, err
		}
		if newerClockEntryExists {
			return false, nil
		}

		stmt, err := tx.Prepare("INSERT INTO chat_identity_last_received (chat_id, clock_value) VALUES (?, ?)")
		if err != nil {
			return false, err
		}
		defer stmt.Close()
		_, err = stmt.Exec(
			contactID,
			chatIdentity.Clock,
		)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	clockUpdated, err = updateClock()
	if err != nil {
		return false, false, err
	}

	for imageType, image := range chatIdentity.Images {
		var exists bool
		err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM chat_identity_contacts WHERE contact_id = ? AND image_type = ? AND clock_value >= ?)`, contactID, imageType, chatIdentity.Clock).Scan(&exists)
		if err != nil {
			return clockUpdated, false, err
		}

		if exists {
			continue
		}

		stmt, err := tx.Prepare(`INSERT INTO chat_identity_contacts (contact_id, image_type, clock_value, payload) VALUES (?, ?, ?, ?)`)
		if err != nil {
			return clockUpdated, false, err
		}
		defer stmt.Close()
		if image.Payload == nil {
			continue
		}

		// TODO implement something that doesn't reject all images if a single image fails validation
		// Validate image URI to make sure it's serializable
		_, err = images.GetPayloadDataURI(image.Payload)
		if err != nil {
			return clockUpdated, false, err
		}

		_, err = stmt.Exec(
			contactID,
			imageType,
			chatIdentity.Clock,
			image.Payload,
		)
		if err != nil {
			return false, false, err
		}
		imagesUpdated = true
	}

	if clockUpdated && chatIdentity.SocialLinks != nil {
		stmt, err := tx.Prepare(`INSERT INTO chat_identity_social_links (chat_id, link_text, link_url) VALUES (?, ?, ?)`)
		if err != nil {
			return clockUpdated, imagesUpdated, err
		}
		defer stmt.Close()

		for _, link := range chatIdentity.SocialLinks {
			_, err = stmt.Exec(
				contactID,
				link.Text,
				link.Url,
			)
			if err != nil {
				return clockUpdated, imagesUpdated, err
			}
		}
	}

	return
}

func (db sqlitePersistence) ExpiredMessagesIDs(maxSendCount int) ([]string, error) {
	ids := []string{}

	rows, err := db.db.Query(`
			SELECT
			  id
			FROM
				raw_messages
			WHERE
			(message_type IN (?, ?) OR resend_automatically) AND sent = ? AND send_count <= ?`,
		protobuf.ApplicationMetadataMessage_CHAT_MESSAGE,
		protobuf.ApplicationMetadataMessage_EMOJI_REACTION,
		false,
		maxSendCount)
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

func (db sqlitePersistence) SaveContact(contact *Contact, tx *sql.Tx) (err error) {
	if tx == nil {
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
	}

	// Insert record
	// NOTE: name, photo and tribute_to_talk are not used anymore, but it's not nullable
	// Removing it requires copying over the table which might be expensive
	// when there are many contacts, so best avoiding it
	stmt, err := tx.Prepare(`
		INSERT INTO contacts(
			id,
			address,
			alias,
			display_name,
			identicon,
			last_updated,
			last_updated_locally,
			local_nickname,
			contact_request_state,
                        contact_request_local_clock,
			contact_request_remote_state,
			contact_request_remote_clock,
			blocked,
			removed,
			verification_status,
			bio,
			name,
			photo,
			tribute_to_talk
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		contact.ID,
		contact.Address,
		contact.Alias,
		contact.DisplayName,
		contact.Identicon,
		contact.LastUpdated,
		contact.LastUpdatedLocally,
		contact.LocalNickname,
		contact.ContactRequestLocalState,
		contact.ContactRequestLocalClock,
		contact.ContactRequestRemoteState,
		contact.ContactRequestRemoteClock,
		contact.Blocked,
		contact.Removed,
		contact.VerificationStatus,
		contact.Bio,
		//TODO we need to drop these columns
		"",
		"",
		"",
	)
	return
}

func (db sqlitePersistence) SaveTransactionToValidate(transaction *TransactionToValidate) error {
	compressedKey := crypto.CompressPubkey(transaction.From)

	_, err := db.db.Exec(`INSERT INTO messenger_transactions_to_validate(
		command_id,
		message_id,
		transaction_hash,
		retry_count,
		first_seen,
		public_key,
		signature,
		to_validate)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		transaction.CommandID,
		transaction.MessageID,
		transaction.TransactionHash,
		transaction.RetryCount,
		transaction.FirstSeen,
		compressedKey,
		transaction.Signature,
		transaction.Validate,
	)

	return err
}

func (db sqlitePersistence) UpdateTransactionToValidate(transaction *TransactionToValidate) error {
	_, err := db.db.Exec(`UPDATE messenger_transactions_to_validate
			      SET retry_count = ?, to_validate = ?
			      WHERE transaction_hash = ?`,
		transaction.RetryCount,
		transaction.Validate,
		transaction.TransactionHash,
	)
	return err
}

func (db sqlitePersistence) TransactionsToValidate() ([]*TransactionToValidate, error) {
	var transactions []*TransactionToValidate
	rows, err := db.db.Query(`
		SELECT
		command_id,
			message_id,
			transaction_hash,
			retry_count,
			first_seen,
			public_key,
			signature,
			to_validate
		FROM messenger_transactions_to_validate
		WHERE to_validate = 1;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t TransactionToValidate
		var pkBytes []byte
		err = rows.Scan(
			&t.CommandID,
			&t.MessageID,
			&t.TransactionHash,
			&t.RetryCount,
			&t.FirstSeen,
			&pkBytes,
			&t.Signature,
			&t.Validate,
		)
		if err != nil {
			return nil, err
		}

		publicKey, err := crypto.DecompressPubkey(pkBytes)
		if err != nil {
			return nil, err
		}
		t.From = publicKey

		transactions = append(transactions, &t)
	}

	return transactions, nil
}

func (db sqlitePersistence) GetWhenChatIdentityLastPublished(chatID string) (t int64, hash []byte, err error) {
	rows, err := db.db.Query("SELECT clock_value, hash FROM chat_identity_last_published WHERE chat_id = ?", chatID)
	if err != nil {
		return t, nil, err
	}
	defer func() {
		err = rows.Close()
	}()

	for rows.Next() {
		err = rows.Scan(&t, &hash)
		if err != nil {
			return t, nil, err
		}
	}

	return t, hash, nil
}

func (db sqlitePersistence) SaveWhenChatIdentityLastPublished(chatID string, hash []byte) (err error) {
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

	stmt, err := tx.Prepare("INSERT INTO chat_identity_last_published (chat_id, clock_value, hash) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(chatID, time.Now().Unix(), hash)
	if err != nil {
		return err
	}

	return nil
}

func (db sqlitePersistence) ResetWhenChatIdentityLastPublished(chatID string) (err error) {
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

	stmt, err := tx.Prepare("INSERT INTO chat_identity_last_published (chat_id, clock_value, hash) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(chatID, 0, []byte("."))
	if err != nil {
		return err
	}

	return nil
}

func (db sqlitePersistence) InsertStatusUpdate(userStatus UserStatus) error {
	_, err := db.db.Exec(`INSERT INTO status_updates(
		public_key,
		status_type,
		clock,
		custom_text)
		VALUES (?, ?, ?, ?)`,
		userStatus.PublicKey,
		userStatus.StatusType,
		userStatus.Clock,
		userStatus.CustomText,
	)

	return err
}

func (db sqlitePersistence) CleanOlderStatusUpdates() error {
	now := time.Now()
	twoWeeksAgo := now.Add(time.Duration(-1*HoursInTwoWeeks) * time.Hour)
	_, err := db.db.Exec(`DELETE FROM status_updates WHERE clock < ?`,
		uint64(twoWeeksAgo.Unix()),
	)

	return err
}

func (db sqlitePersistence) StatusUpdates() (statusUpdates []UserStatus, err error) {
	rows, err := db.db.Query(`
		SELECT
			public_key,
			status_type,
			clock,
			custom_text
		FROM status_updates
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userStatus UserStatus
		err = rows.Scan(
			&userStatus.PublicKey,
			&userStatus.StatusType,
			&userStatus.Clock,
			&userStatus.CustomText,
		)
		if err != nil {
			return
		}
		statusUpdates = append(statusUpdates, userStatus)
	}

	return
}

func (db sqlitePersistence) DeleteSwitcherCard(cardID string) error {
	_, err := db.db.Exec("DELETE from switcher_cards WHERE card_id = ?", cardID)
	return err
}

func (db sqlitePersistence) UpsertSwitcherCard(switcherCard SwitcherCard) error {
	_, err := db.db.Exec(`INSERT INTO switcher_cards(
		card_id,
		type,
		clock,
		screen_id)
		VALUES (?, ?, ?, ?)`,
		switcherCard.CardID,
		switcherCard.Type,
		switcherCard.Clock,
		switcherCard.ScreenID,
	)

	return err
}

func (db sqlitePersistence) SwitcherCards() (switcherCards []SwitcherCard, err error) {
	rows, err := db.db.Query(`
		SELECT
			card_id,
			type,
			clock,
			screen_id
		FROM switcher_cards
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var switcherCard SwitcherCard
		err = rows.Scan(
			&switcherCard.CardID,
			&switcherCard.Type,
			&switcherCard.Clock,
			&switcherCard.ScreenID,
		)
		if err != nil {
			return
		}
		switcherCards = append(switcherCards, switcherCard)
	}

	return
}

func (db sqlitePersistence) NextHigherClockValueOfAutomaticStatusUpdates(clock uint64) (uint64, error) {
	var nextClock uint64

	err := db.db.QueryRow(`
		SELECT clock
		FROM status_updates
		WHERE clock > ? AND status_type = ?
		LIMIT 1
	`, clock, protobuf.StatusUpdate_AUTOMATIC).Scan(&nextClock)

	switch err {
	case sql.ErrNoRows:
		return 0, common.ErrRecordNotFound
	case nil:
		return nextClock, nil
	default:
		return 0, err
	}
}

func (db sqlitePersistence) DeactivatedAutomaticStatusUpdates(fromClock uint64, tillClock uint64) (statusUpdates []UserStatus, err error) {
	rows, err := db.db.Query(`
		SELECT
			public_key,
			?,
			clock + 1,
			custom_text
		FROM status_updates
		WHERE clock > ? AND clock <= ? AND status_type = ?
	`, protobuf.StatusUpdate_INACTIVE, fromClock, tillClock, protobuf.StatusUpdate_AUTOMATIC)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userStatus UserStatus
		err = rows.Scan(
			&userStatus.PublicKey,
			&userStatus.StatusType,
			&userStatus.Clock,
			&userStatus.CustomText,
		)
		if err != nil {
			return
		}
		statusUpdates = append(statusUpdates, userStatus)
	}

	return
}

func (db *sqlitePersistence) AddBookmark(bookmark browsers.Bookmark) (browsers.Bookmark, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return bookmark, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()
	insert, err := tx.Prepare("INSERT OR REPLACE INTO bookmarks (url, name, image_url, removed, clock) VALUES (?, ?, ?, ?, ?)")

	if err != nil {
		return bookmark, err
	}

	// Get the right icon
	finder := besticon.IconFinder{}
	icons, iconError := finder.FetchIcons(bookmark.URL)

	if iconError == nil && len(icons) > 0 {
		icon := finder.IconInSizeRange(besticon.SizeRange{Min: 48, Perfect: 48, Max: 100})
		if icon != nil {
			bookmark.ImageURL = icon.URL
		} else {
			bookmark.ImageURL = icons[0].URL
		}
	} else {
		log.Error("error getting the bookmark icon", "iconError", iconError)
	}

	_, err = insert.Exec(bookmark.URL, bookmark.Name, bookmark.ImageURL, bookmark.Removed, bookmark.Clock)
	return bookmark, err
}

func (db *sqlitePersistence) AddBrowser(browser browsers.Browser) (err error) {
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
	insert, err := tx.Prepare("INSERT OR REPLACE INTO browsers(id, name, timestamp, dapp, historyIndex) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return
	}

	_, err = insert.Exec(browser.ID, browser.Name, browser.Timestamp, browser.Dapp, browser.HistoryIndex)
	insert.Close()
	if err != nil {
		return
	}

	if len(browser.History) == 0 {
		return
	}
	bhInsert, err := tx.Prepare("INSERT INTO browsers_history(browser_id, history) VALUES(?, ?)")
	if err != nil {
		return
	}
	defer bhInsert.Close()
	for _, history := range browser.History {
		_, err = bhInsert.Exec(browser.ID, history)
		if err != nil {
			return
		}
	}
	return
}

func (db *sqlitePersistence) InsertBrowser(browser browsers.Browser) (err error) {
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

	bInsert, err := tx.Prepare("INSERT OR REPLACE INTO browsers(id, name, timestamp, dapp, historyIndex) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return
	}
	_, err = bInsert.Exec(browser.ID, browser.Name, browser.Timestamp, browser.Dapp, browser.HistoryIndex)
	bInsert.Close()
	if err != nil {
		return
	}

	if len(browser.History) == 0 {
		return
	}
	bhInsert, err := tx.Prepare("INSERT INTO browsers_history(browser_id, history) VALUES(?, ?)")
	if err != nil {
		return
	}
	defer bhInsert.Close()
	for _, history := range browser.History {
		_, err = bhInsert.Exec(browser.ID, history)
		if err != nil {
			return
		}
	}

	return
}

func (db *sqlitePersistence) RemoveBookmark(url string, deletedAt uint64) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`UPDATE bookmarks SET removed = 1, deleted_at = ? WHERE url = ?`, deletedAt, url)
	return err
}

func (db *sqlitePersistence) GetBrowsers() (rst []*browsers.Browser, err error) {
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

	// FULL and RIGHT joins are not supported
	bRows, err := tx.Query("SELECT id, name, timestamp, dapp, historyIndex FROM browsers ORDER BY timestamp DESC")
	if err != nil {
		return
	}
	defer bRows.Close()
	browsersArr := map[string]*browsers.Browser{}
	for bRows.Next() {
		browser := browsers.Browser{}
		err = bRows.Scan(&browser.ID, &browser.Name, &browser.Timestamp, &browser.Dapp, &browser.HistoryIndex)
		if err != nil {
			return nil, err
		}
		browsersArr[browser.ID] = &browser
		rst = append(rst, &browser)
	}

	bhRows, err := tx.Query("SELECT browser_id, history from browsers_history")
	if err != nil {
		return
	}
	defer bhRows.Close()
	var (
		id      string
		history string
	)
	for bhRows.Next() {
		err = bhRows.Scan(&id, &history)
		if err != nil {
			return
		}
		browsersArr[id].History = append(browsersArr[id].History, history)
	}

	return rst, nil
}

func (db *sqlitePersistence) DeleteBrowser(id string) error {
	_, err := db.db.Exec("DELETE from browsers WHERE id = ?", id)
	return err
}

func (db *sqlitePersistence) GetBookmarkByURL(url string) (*browsers.Bookmark, error) {
	bookmark := browsers.Bookmark{}
	err := db.db.QueryRow(`SELECT url, name, image_url, removed, clock, deleted_at FROM bookmarks WHERE url = ?`, url).Scan(&bookmark.URL, &bookmark.Name, &bookmark.ImageURL, &bookmark.Removed, &bookmark.Clock, &bookmark.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

func (db *sqlitePersistence) UpdateBookmark(oldURL string, bookmark browsers.Bookmark) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	insert, err := tx.Prepare("UPDATE bookmarks SET url = ?, name = ?, image_url = ?, removed = ?, clock = ?, deleted_at = ? WHERE url = ?")
	if err != nil {
		return err
	}
	_, err = insert.Exec(bookmark.URL, bookmark.Name, bookmark.ImageURL, bookmark.Removed, bookmark.Clock, bookmark.DeletedAt, oldURL)
	return err
}

func (db *sqlitePersistence) DeleteSoftRemovedBookmarks(threshold uint64) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()
	_, err = tx.Exec(`DELETE from bookmarks WHERE removed = 1 AND deleted_at < ?`, threshold)
	return err
}

func (db *sqlitePersistence) InsertWalletConnectSession(session *WalletConnectSession) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	sessionInsertPreparedStatement, err := tx.Prepare("INSERT OR REPLACE INTO wallet_connect_v1_sessions(peer_id, dapp_name, dapp_url, info) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer sessionInsertPreparedStatement.Close()
	_, err = sessionInsertPreparedStatement.Exec(session.PeerID, session.DAppName, session.DAppURL, session.Info)
	return err
}

func (db *sqlitePersistence) GetWalletConnectSession() ([]WalletConnectSession, error) {
	var sessions []WalletConnectSession

	rows, err := db.db.Query("SELECT peer_id, dapp_name, dapp_url, info FROM wallet_connect_v1_sessions ORDER BY dapp_name")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		session := WalletConnectSession{}
		err = rows.Scan(&session.PeerID, &session.DAppName, &session.DAppURL, &session.Info)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (db *sqlitePersistence) DeleteWalletConnectSession(peerID string) error {
	deleteStatement, err := db.db.Prepare("DELETE FROM wallet_connect_v1_sessions where peer_id=?")
	if err != nil {
		return err
	}
	defer deleteStatement.Close()
	_, err = deleteStatement.Exec(peerID)
	return err
}
