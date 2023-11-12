package protocol

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

// SendPinMessage sends the PinMessage to the corresponding chat
func (m *Messenger) SendPinMessage(ctx context.Context, message *common.PinMessage) (*MessengerResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.sendPinMessage(ctx, message)
}

func (m *Messenger) sendPinMessage(ctx context.Context, message *common.PinMessage) (*MessengerResponse, error) {
	var response MessengerResponse

	// A valid added chat is required.
	chat, ok := m.allChats.Load(message.ChatId)
	if !ok {
		return nil, errors.New("chat not found")
	}

	if chat.CommunityChat() {
		community, err := m.communitiesManager.GetByIDString(chat.CommunityID)
		if err != nil {
			return nil, err
		}

		hasPermission := community.IsPrivilegedMember(&m.identity.PublicKey)
		pinMessageAllowed := community.AllowsAllMembersToPinMessage()
		canPost, err := community.CanPost(&m.identity.PublicKey, chat.CommunityChatID())
		if err != nil {
			return nil, err
		}

		if !canPost && !pinMessageAllowed && !hasPermission {
			return nil, errors.New("can't pin message")
		}
	}

	err := m.handleStandaloneChatIdentity(chat)
	if err != nil {
		return nil, err
	}

	err = extendPinMessageFromChat(message, chat, &m.identity.PublicKey, m.getTimesource())
	if err != nil {
		return nil, err
	}

	message.ID, err = generatePinMessageID(&m.identity.PublicKey, message, chat)
	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.encodeChatEntity(chat, message)
	if err != nil {
		return nil, err
	}

	rawMessage := common.RawMessage{
		LocalChatID:          chat.ID,
		Payload:              encodedMessage,
		MessageType:          protobuf.ApplicationMetadataMessage_PIN_MESSAGE,
		SkipGroupMessageWrap: true,
		ResendAutomatically:  true,
	}
	_, err = m.dispatchMessage(ctx, rawMessage)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SavePinMessages([]*common.PinMessage{message})
	if err != nil {
		return nil, err
	}

	if message.Pinned {
		id, err := generatePinMessageNotificationID(&m.identity.PublicKey, message, chat)
		if err != nil {
			return nil, err
		}
		chatMessage := &common.Message{
			ChatMessage: &protobuf.ChatMessage{
				Clock:       message.Clock,
				Timestamp:   m.getTimesource().GetCurrentTime(),
				ChatId:      chat.ID,
				MessageType: message.MessageType,
				ResponseTo:  message.MessageId,
				ContentType: protobuf.ChatMessage_SYSTEM_MESSAGE_PINNED_MESSAGE,
			},
			WhisperTimestamp: m.getTimesource().GetCurrentTime(),
			ID:               id,
			LocalChatID:      chat.ID,
			From:             m.myHexIdentity(),
		}

		msg := []*common.Message{chatMessage}
		err = m.persistence.SaveMessages(msg)
		if err != nil {
			return nil, err
		}

		msg, err = m.pullMessagesAndResponsesFromDB(msg)
		if err != nil {
			return nil, err
		}

		response.SetMessages(msg)
		err = m.prepareMessages(response.messages)
		if err != nil {
			return nil, err
		}
	}

	response.AddPinMessage(message)
	response.AddChat(chat)
	return &response, m.saveChat(chat)
}

func (m *Messenger) PinnedMessageByChatID(chatID, cursor string, limit int) ([]*common.PinnedMessage, string, error) {
	pinnedMsgs, cursor, err := m.persistence.PinnedMessageByChatID(chatID, cursor, limit)

	if err != nil {
		return nil, "", err
	}

	if m.httpServer != nil {
		for idx := range pinnedMsgs {
			msg := pinnedMsgs[idx].Message
			err = m.prepareMessage(msg, m.httpServer)
			if err != nil {
				return nil, "", err
			}
			pinnedMsgs[idx].Message = msg
		}
	}
	return pinnedMsgs, cursor, nil
}

func (m *Messenger) SavePinMessages(messages []*common.PinMessage) error {
	return m.persistence.SavePinMessages(messages)
}

func generatePinMessageID(pubKey *ecdsa.PublicKey, pm *common.PinMessage, chat *Chat) (string, error) {
	data, err := pinMessageBaseID(pubKey, pm, chat)
	if err != nil {
		return "", err
	}

	id := sha256.Sum256(data)
	idString := fmt.Sprintf("%x", id)

	return idString, nil
}

func pinMessageBaseID(pubKey *ecdsa.PublicKey, pm *common.PinMessage, chat *Chat) ([]byte, error) {
	data := gethcommon.FromHex(pm.MessageId)

	switch {
	case chat.ChatType == ChatTypeOneToOne:
		ourPubKey := crypto.FromECDSAPub(pubKey)
		tmpPubKey, err := chat.PublicKey()
		if err != nil {
			return nil, err
		}
		theirPubKey := crypto.FromECDSAPub(tmpPubKey)

		if bytes.Compare(ourPubKey, theirPubKey) < 0 {
			data = append(data, ourPubKey...)   // our key
			data = append(data, theirPubKey...) // their key
		} else {
			data = append(data, theirPubKey...) // their key
			data = append(data, ourPubKey...)   // our key
		}
	default:
		data = append(data, []byte(chat.ID)...)
	}

	return data, nil
}

func generatePinMessageNotificationID(pubKey *ecdsa.PublicKey, pm *common.PinMessage, chat *Chat) (string, error) {
	data, err := pinMessageBaseID(pubKey, pm, chat)
	if err != nil {
		return "", err
	}

	clockBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(clockBytes, pm.Clock)
	data = append(data, clockBytes...)

	id := sha256.Sum256(data)
	idString := fmt.Sprintf("%x", id)

	return idString, nil
}
