package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/v1"
)

const maxChatMessageTextLength = 4096
const maxStatusMessageText = 128

// maxWhisperDrift is how many milliseconds we allow the clock value to differ
// from whisperTimestamp
const maxWhisperFutureDriftMs uint64 = 120000

func validateClockValue(clock uint64, whisperTimestamp uint64) error {
	if clock == 0 {
		return errors.New("clock can't be 0")
	}

	if clock > whisperTimestamp && clock-whisperTimestamp > maxWhisperFutureDriftMs {
		return errors.New("clock value too high")
	}

	return nil
}

func ValidateMembershipUpdateMessage(message *protocol.MembershipUpdateMessage, timeNowMs uint64) error {

	for _, e := range message.Events {
		if err := validateClockValue(e.ClockValue, timeNowMs); err != nil {
			return err
		}

	}
	return nil
}

func ValidateStatusUpdate(message *protobuf.StatusUpdate) error {
	if message == nil {
		return errors.New("message can't be nil")
	}
	if message.Clock == 0 {
		return errors.New("clock can't be 0")
	}

	if message.StatusType == protobuf.StatusUpdate_UNKNOWN_STATUS_TYPE {
		return errors.New("unknown status type")
	}

	if len([]rune(message.CustomText)) > maxStatusMessageText {
		return fmt.Errorf("custom text shouldn't be longer than %d", maxStatusMessageText)
	}

	return nil

}

func ValidateEditMessage(message *protobuf.EditMessage) error {
	if message == nil {
		return errors.New("message can't be nil")
	}
	if message.Clock == 0 {
		return errors.New("clock can't be 0")
	}
	if len(message.ChatId) == 0 {
		return errors.New("chat-id can't be empty")
	}
	if len(message.MessageId) == 0 {
		return errors.New("message-id can't be empty")
	}

	if message.MessageType == protobuf.MessageType_UNKNOWN_MESSAGE_TYPE || message.MessageType == protobuf.MessageType_SYSTEM_MESSAGE_PRIVATE_GROUP {
		return errors.New("unknown message type")
	}

	return ValidateText(message.Text)
}

func ValidateDeleteMessage(message *protobuf.DeleteMessage) error {
	if message == nil {
		return errors.New("message can't be nil")
	}

	if len(message.ChatId) == 0 {
		return errors.New("chat-id can't be empty")
	}
	if len(message.MessageId) == 0 {
		return errors.New("message-id can't be empty")
	}

	if message.MessageType == protobuf.MessageType_UNKNOWN_MESSAGE_TYPE || message.MessageType == protobuf.MessageType_SYSTEM_MESSAGE_PRIVATE_GROUP {
		return errors.New("unknown message type")
	}

	return nil
}

func ValidateDeleteForMeMessage(message *protobuf.SyncDeleteForMeMessage) error {
	if message == nil {
		return errors.New("message can't be nil")
	}

	if len(message.MessageId) == 0 {
		return errors.New("message-id can't be empty")
	}

	return nil
}

func ValidateReceivedPairInstallation(message *protobuf.SyncPairInstallation, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(strings.TrimSpace(message.Name)) == 0 {
		return errors.New("name can't be empty")
	}

	if len(strings.TrimSpace(message.DeviceType)) == 0 {
		return errors.New("device type can't be empty")
	}

	if len(strings.TrimSpace(message.InstallationId)) == 0 {
		return errors.New("installationId can't be empty")
	}

	return nil
}

func ValidateReceivedSendTransaction(message *protobuf.SendTransaction, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(strings.TrimSpace(message.TransactionHash)) == 0 {
		return errors.New("transaction hash can't be empty")
	}

	if message.Signature == nil {
		return errors.New("signature can't be nil")
	}

	return nil
}

func ValidateReceivedRequestAddressForTransaction(message *protobuf.RequestAddressForTransaction, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(strings.TrimSpace(message.Value)) == 0 {
		return errors.New("value can't be empty")
	}

	_, err := strconv.ParseFloat(message.Value, 64)
	if err != nil {
		return err
	}

	return nil
}

func ValidateReceivedRequestTransaction(message *protobuf.RequestTransaction, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(strings.TrimSpace(message.Value)) == 0 {
		return errors.New("value can't be empty")
	}

	if len(strings.TrimSpace(message.Address)) == 0 {
		return errors.New("address can't be empty")
	}

	_, err := strconv.ParseFloat(message.Value, 64)
	if err != nil {
		return err
	}

	return nil
}

func ValidateReceivedAcceptRequestAddressForTransaction(message *protobuf.AcceptRequestAddressForTransaction, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(message.Id) == 0 {
		return errors.New("messageID can't be empty")
	}

	if len(strings.TrimSpace(message.Address)) == 0 {
		return errors.New("address can't be empty")
	}

	return nil
}

func ValidateReceivedDeclineRequestAddressForTransaction(message *protobuf.DeclineRequestAddressForTransaction, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(message.Id) == 0 {
		return errors.New("messageID can't be empty")
	}

	return nil
}

func ValidateReceivedDeclineRequestTransaction(message *protobuf.DeclineRequestTransaction, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(message.Id) == 0 {
		return errors.New("messageID can't be empty")
	}

	return nil
}

func ValidateText(text string) error {
	if len(strings.TrimSpace(text)) == 0 {
		return errors.New("text can't be empty")
	}

	if len([]rune(text)) > maxChatMessageTextLength {
		return fmt.Errorf("text shouldn't be longer than %d", maxChatMessageTextLength)
	}

	return nil
}

func ValidateReceivedChatMessage(message *protobuf.ChatMessage, whisperTimestamp uint64) error {
	if err := validateClockValue(message.Clock, whisperTimestamp); err != nil {
		return err
	}

	if message.Timestamp == 0 {
		return errors.New("timestamp can't be 0")
	}

	if message.ContentType != protobuf.ChatMessage_DISCORD_MESSAGE &&
		message.ContentType != protobuf.ChatMessage_BRIDGE_MESSAGE &&
		(message.ContentType != protobuf.ChatMessage_IMAGE || message.Text != "") {
		if err := ValidateText(message.Text); err != nil {
			return err
		}
	}

	if len(message.ChatId) == 0 {
		return errors.New("chatId can't be empty")
	}

	if message.MessageType == protobuf.MessageType_UNKNOWN_MESSAGE_TYPE || message.MessageType == protobuf.MessageType_SYSTEM_MESSAGE_PRIVATE_GROUP {
		return errors.New("unknown message type")
	}

	switch message.ContentType {
	case protobuf.ChatMessage_UNKNOWN_CONTENT_TYPE:
		return errors.New("unknown content type")

	case protobuf.ChatMessage_TRANSACTION_COMMAND:
		return errors.New("can't receive request address for transaction from others")

	case protobuf.ChatMessage_STICKER:
		if message.Payload == nil {
			return errors.New("no sticker content")
		}
		sticker := message.GetSticker()
		if sticker == nil {
			return errors.New("no sticker content")
		}
		if len(sticker.Hash) == 0 {
			return errors.New("sticker hash not set")
		}

	case protobuf.ChatMessage_IMAGE:
		if message.Payload == nil {
			return errors.New("no image content")
		}
		image := message.GetImage()
		if image == nil {
			return errors.New("no image content")
		}
		if len(image.Payload) == 0 {
			return errors.New("image payload empty")
		}
		if image.Format == protobuf.ImageFormat_UNKNOWN_IMAGE_FORMAT {
			return errors.New("image type unknown")
		}

	case protobuf.ChatMessage_BRIDGE_MESSAGE:
		if message.Payload == nil {
			return errors.New("no bridge message content")
		}
		bridgeMessage := message.GetBridgeMessage()
		if bridgeMessage == nil {
			return errors.New("no bridge message content")
		}
		if len(bridgeMessage.UserName) == 0 {
			return errors.New("no username")
		}
		if len(bridgeMessage.BridgeName) == 0 {
			return errors.New("no bridge name")
		}
		if len(bridgeMessage.Content) == 0 {
			return errors.New("no bridge message content text")
		}
	}

	if message.ContentType == protobuf.ChatMessage_AUDIO {
		if message.Payload == nil {
			return errors.New("no audio content")
		}
		audio := message.GetAudio()
		if audio == nil {
			return errors.New("no audio content")
		}
		if len(audio.Payload) == 0 {
			return errors.New("audio payload empty")
		}

		if audio.Type == protobuf.AudioMessage_UNKNOWN_AUDIO_TYPE {
			return errors.New("audio type unknown")
		}
	}

	if message.ContentType == protobuf.ChatMessage_SYSTEM_MESSAGE_CONTENT_PRIVATE_GROUP {
		return errors.New("private group system message content type not allowed")
	}

	if message.ContentType == protobuf.ChatMessage_SYSTEM_MESSAGE_MUTUAL_EVENT_SENT ||
		message.ContentType == protobuf.ChatMessage_SYSTEM_MESSAGE_MUTUAL_EVENT_ACCEPTED ||
		message.ContentType == protobuf.ChatMessage_SYSTEM_MESSAGE_MUTUAL_EVENT_REMOVED {
		return errors.New("mutual state event system message content type not allowed")
	}

	if err := ValidateDisplayName(&message.DisplayName); err != nil {
		return err
	}

	return nil
}

func ValidateReceivedEmojiReaction(emoji *protobuf.EmojiReaction, whisperTimestamp uint64) error {
	if err := validateClockValue(emoji.Clock, whisperTimestamp); err != nil {
		return err
	}

	if len(emoji.MessageId) == 0 {
		return errors.New("message-id can't be empty")
	}

	if len(emoji.ChatId) == 0 {
		return errors.New("chat-id can't be empty")
	}

	if emoji.Type == protobuf.EmojiReaction_UNKNOWN_EMOJI_REACTION_TYPE {
		return errors.New("unknown emoji reaction type")
	}

	if emoji.MessageType == protobuf.MessageType_UNKNOWN_MESSAGE_TYPE {
		return errors.New("unknown message type")
	}

	return nil
}

func ValidateReceivedGroupChatInvitation(invitation *protobuf.GroupChatInvitation) error {

	if len(invitation.ChatId) == 0 {
		return errors.New("chat-id can't be empty")
	}

	return nil
}
