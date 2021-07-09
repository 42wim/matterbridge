// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// TypeMap is a mapping from event type to the content struct type.
// This is used by Content.ParseRaw() for creating the correct type of struct.
var TypeMap = map[Type]reflect.Type{
	StateMember:            reflect.TypeOf(MemberEventContent{}),
	StatePowerLevels:       reflect.TypeOf(PowerLevelsEventContent{}),
	StateCanonicalAlias:    reflect.TypeOf(CanonicalAliasEventContent{}),
	StateRoomName:          reflect.TypeOf(RoomNameEventContent{}),
	StateRoomAvatar:        reflect.TypeOf(RoomAvatarEventContent{}),
	StateTopic:             reflect.TypeOf(TopicEventContent{}),
	StateTombstone:         reflect.TypeOf(TombstoneEventContent{}),
	StateCreate:            reflect.TypeOf(CreateEventContent{}),
	StateJoinRules:         reflect.TypeOf(JoinRulesEventContent{}),
	StateHistoryVisibility: reflect.TypeOf(HistoryVisibilityEventContent{}),
	StateGuestAccess:       reflect.TypeOf(GuestAccessEventContent{}),
	StatePinnedEvents:      reflect.TypeOf(PinnedEventsEventContent{}),
	StateEncryption:        reflect.TypeOf(EncryptionEventContent{}),

	EventMessage:   reflect.TypeOf(MessageEventContent{}),
	EventSticker:   reflect.TypeOf(MessageEventContent{}),
	EventEncrypted: reflect.TypeOf(EncryptedEventContent{}),
	EventRedaction: reflect.TypeOf(RedactionEventContent{}),
	EventReaction:  reflect.TypeOf(ReactionEventContent{}),

	AccountDataRoomTags:        reflect.TypeOf(TagEventContent{}),
	AccountDataDirectChats:     reflect.TypeOf(DirectChatsEventContent{}),
	AccountDataFullyRead:       reflect.TypeOf(FullyReadEventContent{}),
	AccountDataIgnoredUserList: reflect.TypeOf(IgnoredUserListEventContent{}),

	EphemeralEventTyping:   reflect.TypeOf(TypingEventContent{}),
	EphemeralEventReceipt:  reflect.TypeOf(ReceiptEventContent{}),
	EphemeralEventPresence: reflect.TypeOf(PresenceEventContent{}),

	InRoomVerificationStart:  reflect.TypeOf(VerificationStartEventContent{}),
	InRoomVerificationReady:  reflect.TypeOf(VerificationReadyEventContent{}),
	InRoomVerificationAccept: reflect.TypeOf(VerificationAcceptEventContent{}),
	InRoomVerificationKey:    reflect.TypeOf(VerificationKeyEventContent{}),
	InRoomVerificationMAC:    reflect.TypeOf(VerificationMacEventContent{}),
	InRoomVerificationCancel: reflect.TypeOf(VerificationCancelEventContent{}),

	ToDeviceRoomKey:          reflect.TypeOf(RoomKeyEventContent{}),
	ToDeviceForwardedRoomKey: reflect.TypeOf(ForwardedRoomKeyEventContent{}),
	ToDeviceRoomKeyRequest:   reflect.TypeOf(RoomKeyRequestEventContent{}),
	ToDeviceEncrypted:        reflect.TypeOf(EncryptedEventContent{}),
	ToDeviceRoomKeyWithheld:  reflect.TypeOf(RoomKeyWithheldEventContent{}),

	ToDeviceVerificationStart:   reflect.TypeOf(VerificationStartEventContent{}),
	ToDeviceVerificationAccept:  reflect.TypeOf(VerificationAcceptEventContent{}),
	ToDeviceVerificationKey:     reflect.TypeOf(VerificationKeyEventContent{}),
	ToDeviceVerificationMAC:     reflect.TypeOf(VerificationMacEventContent{}),
	ToDeviceVerificationCancel:  reflect.TypeOf(VerificationCancelEventContent{}),
	ToDeviceVerificationRequest: reflect.TypeOf(VerificationRequestEventContent{}),

	ToDeviceOrgMatrixRoomKeyWithheld: reflect.TypeOf(RoomKeyWithheldEventContent{}),

	CallInvite:       reflect.TypeOf(CallInviteEventContent{}),
	CallCandidates:   reflect.TypeOf(CallCandidatesEventContent{}),
	CallAnswer:       reflect.TypeOf(CallAnswerEventContent{}),
	CallReject:       reflect.TypeOf(CallRejectEventContent{}),
	CallSelectAnswer: reflect.TypeOf(CallSelectAnswerEventContent{}),
	CallNegotiate:    reflect.TypeOf(CallNegotiateEventContent{}),
	CallHangup:       reflect.TypeOf(CallHangupEventContent{}),
}

// Content stores the content of a Matrix event.
//
// By default, the content is only parsed into a map[string]interface{}. However, you can call ParseRaw with the
// correct event type to parse the content into a nicer struct, which you can then access from Parsed or via the
// helper functions.
type Content struct {
	VeryRaw json.RawMessage
	Raw     map[string]interface{}
	Parsed  interface{}
}

type Relatable interface {
	GetRelatesTo() *RelatesTo
	OptionalGetRelatesTo() *RelatesTo
	SetRelatesTo(rel *RelatesTo)
}

func (content *Content) UnmarshalJSON(data []byte) error {
	content.VeryRaw = data
	err := json.Unmarshal(data, &content.Raw)
	return err
}

func (content *Content) MarshalJSON() ([]byte, error) {
	if content.Raw == nil {
		if content.Parsed == nil {
			if content.VeryRaw == nil {
				return []byte("{}"), nil
			}
			return content.VeryRaw, nil
		}
		return json.Marshal(content.Parsed)
	} else if content.Parsed != nil {
		// TODO this whole thing is incredibly hacky
		// It needs to produce JSON, where:
		// * content.Parsed is applied after content.Raw
		// * MarshalJSON() is respected inside content.Parsed
		// * Custom field inside nested objects of content.Raw are preserved,
		//   even if content.Parsed contains the higher-level objects.
		// * content.Raw is not modified

		unparsed, err := json.Marshal(content.Parsed)
		if err != nil {
			return nil, err
		}

		var rawParsed map[string]interface{}
		err = json.Unmarshal(unparsed, &rawParsed)
		if err != nil {
			return nil, err
		}

		output := make(map[string]interface{})
		for key, value := range content.Raw {
			output[key] = value
		}

		mergeMaps(output, rawParsed)
		return json.Marshal(output)
	}
	return json.Marshal(content.Raw)
}

func IsUnsupportedContentType(err error) bool {
	return errors.Is(err, UnsupportedContentType)
}

var ContentAlreadyParsed = errors.New("content is already parsed")
var UnsupportedContentType = errors.New("unsupported event type")

func (content *Content) ParseRaw(evtType Type) error {
	if content.Parsed != nil {
		return ContentAlreadyParsed
	}
	structType, ok := TypeMap[evtType]
	if !ok {
		return fmt.Errorf("%w %s", UnsupportedContentType, evtType.Repr())
	}
	content.Parsed = reflect.New(structType).Interface()
	return json.Unmarshal(content.VeryRaw, &content.Parsed)
}

func mergeMaps(into, from map[string]interface{}) {
	for key, newValue := range from {
		existingValue, ok := into[key]
		if !ok {
			into[key] = newValue
			continue
		}
		existingValueMap, okEx := existingValue.(map[string]interface{})
		newValueMap, okNew := newValue.(map[string]interface{})
		if okEx && okNew {
			mergeMaps(existingValueMap, newValueMap)
		} else {
			into[key] = newValue
		}
	}
}

func init() {
	gob.Register(&MemberEventContent{})
	gob.Register(&PowerLevelsEventContent{})
	gob.Register(&CanonicalAliasEventContent{})
	gob.Register(&EncryptionEventContent{})
	gob.Register(&RoomNameEventContent{})
	gob.Register(&RoomAvatarEventContent{})
	gob.Register(&TopicEventContent{})
	gob.Register(&TombstoneEventContent{})
	gob.Register(&CreateEventContent{})
	gob.Register(&JoinRulesEventContent{})
	gob.Register(&HistoryVisibilityEventContent{})
	gob.Register(&GuestAccessEventContent{})
	gob.Register(&PinnedEventsEventContent{})
	gob.Register(&MessageEventContent{})
	gob.Register(&MessageEventContent{})
	gob.Register(&EncryptedEventContent{})
	gob.Register(&RedactionEventContent{})
	gob.Register(&ReactionEventContent{})
	gob.Register(&TagEventContent{})
	gob.Register(&DirectChatsEventContent{})
	gob.Register(&FullyReadEventContent{})
	gob.Register(&IgnoredUserListEventContent{})
	gob.Register(&TypingEventContent{})
	gob.Register(&ReceiptEventContent{})
	gob.Register(&PresenceEventContent{})
	gob.Register(&RoomKeyEventContent{})
	gob.Register(&ForwardedRoomKeyEventContent{})
	gob.Register(&RoomKeyRequestEventContent{})
	gob.Register(&RoomKeyWithheldEventContent{})
}

// Helper cast functions below

func (content *Content) AsMember() *MemberEventContent {
	casted, ok := content.Parsed.(*MemberEventContent)
	if !ok {
		return &MemberEventContent{}
	}
	return casted
}
func (content *Content) AsPowerLevels() *PowerLevelsEventContent {
	casted, ok := content.Parsed.(*PowerLevelsEventContent)
	if !ok {
		return &PowerLevelsEventContent{}
	}
	return casted
}
func (content *Content) AsCanonicalAlias() *CanonicalAliasEventContent {
	casted, ok := content.Parsed.(*CanonicalAliasEventContent)
	if !ok {
		return &CanonicalAliasEventContent{}
	}
	return casted
}
func (content *Content) AsRoomName() *RoomNameEventContent {
	casted, ok := content.Parsed.(*RoomNameEventContent)
	if !ok {
		return &RoomNameEventContent{}
	}
	return casted
}
func (content *Content) AsRoomAvatar() *RoomAvatarEventContent {
	casted, ok := content.Parsed.(*RoomAvatarEventContent)
	if !ok {
		return &RoomAvatarEventContent{}
	}
	return casted
}
func (content *Content) AsTopic() *TopicEventContent {
	casted, ok := content.Parsed.(*TopicEventContent)
	if !ok {
		return &TopicEventContent{}
	}
	return casted
}
func (content *Content) AsTombstone() *TombstoneEventContent {
	casted, ok := content.Parsed.(*TombstoneEventContent)
	if !ok {
		return &TombstoneEventContent{}
	}
	return casted
}
func (content *Content) AsCreate() *CreateEventContent {
	casted, ok := content.Parsed.(*CreateEventContent)
	if !ok {
		return &CreateEventContent{}
	}
	return casted
}
func (content *Content) AsJoinRules() *JoinRulesEventContent {
	casted, ok := content.Parsed.(*JoinRulesEventContent)
	if !ok {
		return &JoinRulesEventContent{}
	}
	return casted
}
func (content *Content) AsHistoryVisibility() *HistoryVisibilityEventContent {
	casted, ok := content.Parsed.(*HistoryVisibilityEventContent)
	if !ok {
		return &HistoryVisibilityEventContent{}
	}
	return casted
}
func (content *Content) AsGuestAccess() *GuestAccessEventContent {
	casted, ok := content.Parsed.(*GuestAccessEventContent)
	if !ok {
		return &GuestAccessEventContent{}
	}
	return casted
}
func (content *Content) AsPinnedEvents() *PinnedEventsEventContent {
	casted, ok := content.Parsed.(*PinnedEventsEventContent)
	if !ok {
		return &PinnedEventsEventContent{}
	}
	return casted
}
func (content *Content) AsEncryption() *EncryptionEventContent {
	casted, ok := content.Parsed.(*EncryptionEventContent)
	if !ok {
		return &EncryptionEventContent{}
	}
	return casted
}
func (content *Content) AsMessage() *MessageEventContent {
	casted, ok := content.Parsed.(*MessageEventContent)
	if !ok {
		return &MessageEventContent{}
	}
	return casted
}
func (content *Content) AsEncrypted() *EncryptedEventContent {
	casted, ok := content.Parsed.(*EncryptedEventContent)
	if !ok {
		return &EncryptedEventContent{}
	}
	return casted
}
func (content *Content) AsRedaction() *RedactionEventContent {
	casted, ok := content.Parsed.(*RedactionEventContent)
	if !ok {
		return &RedactionEventContent{}
	}
	return casted
}
func (content *Content) AsReaction() *ReactionEventContent {
	casted, ok := content.Parsed.(*ReactionEventContent)
	if !ok {
		return &ReactionEventContent{}
	}
	return casted
}
func (content *Content) AsTag() *TagEventContent {
	casted, ok := content.Parsed.(*TagEventContent)
	if !ok {
		return &TagEventContent{}
	}
	return casted
}
func (content *Content) AsDirectChats() *DirectChatsEventContent {
	casted, ok := content.Parsed.(*DirectChatsEventContent)
	if !ok {
		return &DirectChatsEventContent{}
	}
	return casted
}
func (content *Content) AsFullyRead() *FullyReadEventContent {
	casted, ok := content.Parsed.(*FullyReadEventContent)
	if !ok {
		return &FullyReadEventContent{}
	}
	return casted
}
func (content *Content) AsIgnoredUserList() *IgnoredUserListEventContent {
	casted, ok := content.Parsed.(*IgnoredUserListEventContent)
	if !ok {
		return &IgnoredUserListEventContent{}
	}
	return casted
}
func (content *Content) AsTyping() *TypingEventContent {
	casted, ok := content.Parsed.(*TypingEventContent)
	if !ok {
		return &TypingEventContent{}
	}
	return casted
}
func (content *Content) AsReceipt() *ReceiptEventContent {
	casted, ok := content.Parsed.(*ReceiptEventContent)
	if !ok {
		return &ReceiptEventContent{}
	}
	return casted
}
func (content *Content) AsPresence() *PresenceEventContent {
	casted, ok := content.Parsed.(*PresenceEventContent)
	if !ok {
		return &PresenceEventContent{}
	}
	return casted
}
func (content *Content) AsRoomKey() *RoomKeyEventContent {
	casted, ok := content.Parsed.(*RoomKeyEventContent)
	if !ok {
		return &RoomKeyEventContent{}
	}
	return casted
}
func (content *Content) AsForwardedRoomKey() *ForwardedRoomKeyEventContent {
	casted, ok := content.Parsed.(*ForwardedRoomKeyEventContent)
	if !ok {
		return &ForwardedRoomKeyEventContent{}
	}
	return casted
}
func (content *Content) AsRoomKeyRequest() *RoomKeyRequestEventContent {
	casted, ok := content.Parsed.(*RoomKeyRequestEventContent)
	if !ok {
		return &RoomKeyRequestEventContent{}
	}
	return casted
}
func (content *Content) AsRoomKeyWithheld() *RoomKeyWithheldEventContent {
	casted, ok := content.Parsed.(*RoomKeyWithheldEventContent)
	if !ok {
		return &RoomKeyWithheldEventContent{}
	}
	return casted
}
func (content *Content) AsCallInvite() *CallInviteEventContent {
	casted, ok := content.Parsed.(*CallInviteEventContent)
	if !ok {
		return &CallInviteEventContent{}
	}
	return casted
}
func (content *Content) AsCallCandidates() *CallCandidatesEventContent {
	casted, ok := content.Parsed.(*CallCandidatesEventContent)
	if !ok {
		return &CallCandidatesEventContent{}
	}
	return casted
}
func (content *Content) AsCallAnswer() *CallAnswerEventContent {
	casted, ok := content.Parsed.(*CallAnswerEventContent)
	if !ok {
		return &CallAnswerEventContent{}
	}
	return casted
}
func (content *Content) AsCallReject() *CallRejectEventContent {
	casted, ok := content.Parsed.(*CallRejectEventContent)
	if !ok {
		return &CallRejectEventContent{}
	}
	return casted
}
func (content *Content) AsCallSelectAnswer() *CallSelectAnswerEventContent {
	casted, ok := content.Parsed.(*CallSelectAnswerEventContent)
	if !ok {
		return &CallSelectAnswerEventContent{}
	}
	return casted
}
func (content *Content) AsCallNegotiate() *CallNegotiateEventContent {
	casted, ok := content.Parsed.(*CallNegotiateEventContent)
	if !ok {
		return &CallNegotiateEventContent{}
	}
	return casted
}
func (content *Content) AsCallHangup() *CallHangupEventContent {
	casted, ok := content.Parsed.(*CallHangupEventContent)
	if !ok {
		return &CallHangupEventContent{}
	}
	return casted
}
