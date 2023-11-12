package protocol

import (
	"strings"
	"time"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	v1protocol "github.com/status-im/status-go/protocol/v1"
)

var defaultSystemMessagesTranslations = new(systemMessageTranslationsMap)

func init() {
	defaultSystemMessagesTranslationSet := map[protobuf.MembershipUpdateEvent_EventType]string{
		protobuf.MembershipUpdateEvent_CHAT_CREATED:   "{{from}} created the group {{name}}",
		protobuf.MembershipUpdateEvent_NAME_CHANGED:   "{{from}} changed the group's name to {{name}}",
		protobuf.MembershipUpdateEvent_MEMBERS_ADDED:  "{{from}} has added {{members}}",
		protobuf.MembershipUpdateEvent_ADMINS_ADDED:   "{{from}} has made {{members}} admin",
		protobuf.MembershipUpdateEvent_MEMBER_REMOVED: "{{member}} left the group",
		protobuf.MembershipUpdateEvent_ADMIN_REMOVED:  "{{member}} is not admin anymore",
		protobuf.MembershipUpdateEvent_COLOR_CHANGED:  "{{from}} changed the group's color",
		protobuf.MembershipUpdateEvent_IMAGE_CHANGED:  "{{from}} changed the group's image",
	}
	defaultSystemMessagesTranslations.Init(defaultSystemMessagesTranslationSet)
}

func tsprintf(format string, params map[string]string) string {
	for key, val := range params {
		format = strings.Replace(format, "{{"+key+"}}", val, -1)
	}
	return format
}

func eventToSystemMessage(e v1protocol.MembershipUpdateEvent, translations *systemMessageTranslationsMap) *common.Message {
	var text string
	switch e.Type {
	case protobuf.MembershipUpdateEvent_CHAT_CREATED:
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_CHAT_CREATED)
		text = tsprintf(message, map[string]string{"from": "@" + e.From, "name": e.Name})
	case protobuf.MembershipUpdateEvent_NAME_CHANGED:
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_NAME_CHANGED)
		text = tsprintf(message, map[string]string{"from": "@" + e.From, "name": e.Name})
	case protobuf.MembershipUpdateEvent_COLOR_CHANGED:
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_COLOR_CHANGED)
		text = tsprintf(message, map[string]string{"from": "@" + e.From})
	case protobuf.MembershipUpdateEvent_IMAGE_CHANGED:
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_IMAGE_CHANGED)
		text = tsprintf(message, map[string]string{"from": "@" + e.From})
	case protobuf.MembershipUpdateEvent_MEMBERS_ADDED:

		var memberMentions []string
		for _, s := range e.Members {
			memberMentions = append(memberMentions, "@"+s)
		}
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_MEMBERS_ADDED)
		text = tsprintf(message, map[string]string{"from": "@" + e.From, "members": strings.Join(memberMentions, ", ")})
	case protobuf.MembershipUpdateEvent_ADMINS_ADDED:
		var memberMentions []string
		for _, s := range e.Members {
			memberMentions = append(memberMentions, "@"+s)
		}
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_ADMINS_ADDED)
		text = tsprintf(message, map[string]string{"from": "@" + e.From, "members": strings.Join(memberMentions, ", ")})
	case protobuf.MembershipUpdateEvent_MEMBER_REMOVED:
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_MEMBER_REMOVED)
		text = tsprintf(message, map[string]string{"member": "@" + e.Members[0]})
	case protobuf.MembershipUpdateEvent_ADMIN_REMOVED:
		message, _ := translations.Load(protobuf.MembershipUpdateEvent_ADMIN_REMOVED)
		text = tsprintf(message, map[string]string{"member": "@" + e.Members[0]})

	}
	timestamp := v1protocol.TimestampInMsFromTime(time.Now())
	message := &common.Message{
		ChatMessage: &protobuf.ChatMessage{
			ChatId:      e.ChatID,
			Text:        text,
			MessageType: protobuf.MessageType_SYSTEM_MESSAGE_PRIVATE_GROUP,
			ContentType: protobuf.ChatMessage_SYSTEM_MESSAGE_CONTENT_PRIVATE_GROUP,
			Clock:       e.ClockValue,
			Timestamp:   timestamp,
		},
		From:             e.From,
		WhisperTimestamp: timestamp,
		LocalChatID:      e.ChatID,
		Seen:             true,
		ID:               types.EncodeHex(crypto.Keccak256(e.Signature)),
	}
	// We don't pass an identity here as system messages don't need the mentioned flag
	_ = message.PrepareContent("")
	return message
}

func buildSystemMessages(events []v1protocol.MembershipUpdateEvent, translations *systemMessageTranslationsMap) []*common.Message {
	var messages []*common.Message

	for _, e := range events {
		if e.Type == protobuf.MembershipUpdateEvent_MEMBER_JOINED {
			// explicit join has been removed, ignore this event
			continue
		}

		messages = append(messages, eventToSystemMessage(e, translations))
	}

	return messages
}
