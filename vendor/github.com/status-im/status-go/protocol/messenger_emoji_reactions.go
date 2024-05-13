package protocol

import (
	"context"

	"github.com/pkg/errors"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

func (m *Messenger) SendEmojiReaction(ctx context.Context, chatID, messageID string, emojiID protobuf.EmojiReaction_Type) (*MessengerResponse, error) {
	var response MessengerResponse

	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	emojiR := &EmojiReaction{
		EmojiReaction: &protobuf.EmojiReaction{
			Clock:     clock,
			MessageId: messageID,
			ChatId:    chatID,
			Type:      emojiID,
		},
		LocalChatID: chatID,
		From:        types.EncodeHex(crypto.FromECDSAPub(&m.identity.PublicKey)),
	}
	encodedMessage, err := m.encodeChatEntity(chat, emojiR)
	if err != nil {
		return nil, err
	}

	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:          chatID,
		Payload:              encodedMessage,
		SkipGroupMessageWrap: true,
		MessageType:          protobuf.ApplicationMetadataMessage_EMOJI_REACTION,
		// Don't resend using datasync, that would create quite a lot
		// of traffic if clicking too eagelry
		ResendAutomatically: false,
	})
	if err != nil {
		return nil, err
	}

	response.AddEmojiReaction(emojiR)
	response.AddChat(chat)

	err = m.persistence.SaveEmojiReaction(emojiR)
	if err != nil {
		return nil, errors.Wrap(err, "Can't save emoji reaction in db")
	}

	return &response, nil
}

func (m *Messenger) EmojiReactionsByChatID(chatID string, cursor string, limit int) ([]*EmojiReaction, error) {
	chat, err := m.persistence.Chat(chatID)
	if err != nil {
		return nil, err
	}

	if chat.Timeline() {
		var chatIDs = []string{"@" + contactIDFromPublicKey(&m.identity.PublicKey)}
		m.allContacts.Range(func(contactID string, contact *Contact) (shouldContinue bool) {
			if contact.added() {
				chatIDs = append(chatIDs, "@"+contact.ID)
			}
			return true
		})
		return m.persistence.EmojiReactionsByChatIDs(chatIDs, cursor, limit)
	}
	return m.persistence.EmojiReactionsByChatID(chatID, cursor, limit)
}

func (m *Messenger) EmojiReactionsByChatIDMessageID(chatID string, messageID string) ([]*EmojiReaction, error) {
	_, err := m.persistence.Chat(chatID)
	if err != nil {
		return nil, err
	}

	return m.persistence.EmojiReactionsByChatIDMessageID(chatID, messageID)
}

func (m *Messenger) SendEmojiReactionRetraction(ctx context.Context, emojiReactionID string) (*MessengerResponse, error) {
	emojiR, err := m.persistence.EmojiReactionByID(emojiReactionID)
	if err != nil {
		return nil, err
	}

	// Check that the sender is the key owner
	pk := types.EncodeHex(crypto.FromECDSAPub(&m.identity.PublicKey))
	if emojiR.From != pk {
		return nil, errors.Errorf("identity mismatch, "+
			"emoji reactions can only be retracted by the reaction sender, "+
			"emoji reaction sent by '%s', current identity '%s'",
			emojiR.From, pk,
		)
	}

	// Get chat and clock
	chat, ok := m.allChats.Load(emojiR.GetChatId())
	if !ok {
		return nil, ErrChatNotFound
	}
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	// Update the relevant fields
	emojiR.Clock = clock
	emojiR.Retracted = true

	encodedMessage, err := m.encodeChatEntity(chat, emojiR)
	if err != nil {
		return nil, err
	}

	// Send the marshalled EmojiReactionRetraction protobuf
	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:          emojiR.GetChatId(),
		Payload:              encodedMessage,
		SkipGroupMessageWrap: true,
		MessageType:          protobuf.ApplicationMetadataMessage_EMOJI_REACTION,
		// Don't resend using datasync, that would create quite a lot
		// of traffic if clicking too eagelry
		ResendAutomatically: false,
	})
	if err != nil {
		return nil, err
	}

	// Update MessengerResponse
	response := MessengerResponse{}
	emojiR.Retracted = true
	response.AddEmojiReaction(emojiR)
	response.AddChat(chat)

	// Persist retraction state for emoji reaction
	err = m.persistence.SaveEmojiReaction(emojiR)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
