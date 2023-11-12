package protocol

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	v1protocol "github.com/status-im/status-go/protocol/v1"
)

var ErrGroupChatAddedContacts = errors.New("group-chat: can't add members who are not mutual contacts")

func (m *Messenger) validateAddedGroupMembers(members []string) error {
	for _, memberPubkey := range members {
		contactID, err := contactIDFromPublicKeyString(memberPubkey)
		if err != nil {
			return err
		}

		contact, _ := m.allContacts.Load(contactID)
		if contact == nil || !contact.mutual() {
			return ErrGroupChatAddedContacts
		}
	}
	return nil
}

func (m *Messenger) CreateGroupChatWithMembers(ctx context.Context, name string, members []string) (*MessengerResponse, error) {
	var convertedKeyMembers []string
	for _, m := range members {
		k, err := requests.ConvertCompressedToLegacyKey(m)
		if err != nil {
			return nil, err
		}
		convertedKeyMembers = append(convertedKeyMembers, k)

	}
	if err := m.validateAddedGroupMembers(convertedKeyMembers); err != nil {
		return nil, err
	}

	var response MessengerResponse
	logger := m.logger.With(zap.String("site", "CreateGroupChatWithMembers"))
	logger.Info("Creating group chat", zap.String("name", name), zap.Any("members", convertedKeyMembers))
	chat := CreateGroupChat(m.getTimesource())

	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	group, err := v1protocol.NewGroupWithCreator(name, chat.Color, clock, m.identity)
	if err != nil {
		return nil, err
	}
	chat.LastClockValue = clock

	chat.updateChatFromGroupMembershipChanges(group)
	chat.Joined = int64(m.getTimesource().GetCurrentTime())

	clock, _ = chat.NextClockAndTimestamp(m.getTimesource())

	// Add members
	if len(convertedKeyMembers) > 0 {
		event := v1protocol.NewMembersAddedEvent(convertedKeyMembers, clock)
		event.ChatID = chat.ID
		err = event.Sign(m.identity)
		if err != nil {
			return nil, err
		}

		err = group.ProcessEvent(event)
		if err != nil {
			return nil, err
		}
	}

	recipients, err := stringSliceToPublicKeys(group.Members())

	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.sender.EncodeMembershipUpdate(group, nil)
	if err != nil {
		return nil, err
	}

	m.allChats.Store(chat.ID, &chat)

	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID: chat.ID,
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
		Recipients:  recipients,
	})

	if err != nil {
		return nil, err
	}

	chat.updateChatFromGroupMembershipChanges(group)

	return m.addMessagesAndChat(&chat, buildSystemMessages(chat.MembershipUpdates, m.systemMessagesTranslations), &response)
}

func (m *Messenger) CreateGroupChatFromInvitation(name string, chatID string, adminPK string) (*MessengerResponse, error) {
	var response MessengerResponse
	logger := m.logger.With(zap.String("site", "CreateGroupChatFromInvitation"))
	logger.Info("Creating group chat from invitation", zap.String("name", name))
	chat := CreateGroupChat(m.getTimesource())
	chat.ID = chatID
	chat.Name = name
	chat.InvitationAdmin = adminPK

	response.AddChat(&chat)

	return &response, m.saveChat(&chat)
}

type removeMembersFromGroupChatResponse struct {
	oldRecipients   []*ecdsa.PublicKey
	group           *v1protocol.Group
	encodedProtobuf []byte
}

func (m *Messenger) removeMembersFromGroupChat(ctx context.Context, chat *Chat, members []string) (*removeMembersFromGroupChatResponse, error) {
	chatID := chat.ID
	logger := m.logger.With(zap.String("site", "RemoveMembersFromGroupChat"))
	logger.Info("Removing members form group chat", zap.String("chatID", chatID), zap.Any("members", members))
	group, err := newProtocolGroupFromChat(chat)
	if err != nil {
		return nil, err
	}

	// We save the initial recipients as we want to send updates to also
	// the members kicked out
	oldRecipients, err := stringSliceToPublicKeys(group.Members())
	if err != nil {
		return nil, err
	}

	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	for _, member := range members {
		// Remove member
		event := v1protocol.NewMemberRemovedEvent(member, clock)
		event.ChatID = chat.ID
		err = event.Sign(m.identity)
		if err != nil {
			return nil, err
		}

		err = group.ProcessEvent(event)
		if err != nil {
			return nil, err
		}
	}

	encoded, err := m.sender.EncodeMembershipUpdate(group, nil)
	if err != nil {
		return nil, err
	}

	return &removeMembersFromGroupChatResponse{
		oldRecipients:   oldRecipients,
		group:           group,
		encodedProtobuf: encoded,
	}, nil
}

func (m *Messenger) RemoveMembersFromGroupChat(ctx context.Context, chatID string, members []string) (*MessengerResponse, error) {
	var response MessengerResponse

	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}

	removeMembersResponse, err := m.removeMembersFromGroupChat(ctx, chat, members)
	if err != nil {
		return nil, err
	}

	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID: chat.ID,
		Payload:     removeMembersResponse.encodedProtobuf,
		MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
		Recipients:  removeMembersResponse.oldRecipients,
	})
	if err != nil {
		return nil, err
	}

	chat.updateChatFromGroupMembershipChanges(removeMembersResponse.group)

	return m.addMessagesAndChat(chat, buildSystemMessages(chat.MembershipUpdates, m.systemMessagesTranslations), &response)
}

func (m *Messenger) AddMembersToGroupChat(ctx context.Context, chatID string, members []string) (*MessengerResponse, error) {
	if err := m.validateAddedGroupMembers(members); err != nil {
		return nil, err
	}

	var response MessengerResponse
	logger := m.logger.With(zap.String("site", "AddMembersFromGroupChat"))
	logger.Info("Adding members form group chat", zap.String("chatID", chatID), zap.Any("members", members))
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}

	group, err := newProtocolGroupFromChat(chat)
	if err != nil {
		return nil, err
	}

	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
	// Add members
	event := v1protocol.NewMembersAddedEvent(members, clock)
	event.ChatID = chat.ID
	err = event.Sign(m.identity)
	if err != nil {
		return nil, err
	}

	//approve invitations
	for _, member := range members {
		logger.Info("ApproveInvitationByChatIdAndFrom", zap.String("chatID", chatID), zap.Any("member", member))

		groupChatInvitation := &GroupChatInvitation{
			GroupChatInvitation: &protobuf.GroupChatInvitation{
				ChatId: chat.ID,
			},
			From: member,
		}

		groupChatInvitation, err = m.persistence.InvitationByID(groupChatInvitation.ID())
		if err != nil && err != common.ErrRecordNotFound {
			return nil, err
		}
		if groupChatInvitation != nil {
			groupChatInvitation.State = protobuf.GroupChatInvitation_APPROVED

			err := m.persistence.SaveInvitation(groupChatInvitation)
			if err != nil {
				return nil, err
			}
			response.Invitations = append(response.Invitations, groupChatInvitation)
		}
	}

	err = group.ProcessEvent(event)
	if err != nil {
		return nil, err
	}

	recipients, err := stringSliceToPublicKeys(group.Members())
	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.sender.EncodeMembershipUpdate(group, nil)
	if err != nil {
		return nil, err
	}
	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID: chat.ID,
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
		Recipients:  recipients,
	})

	if err != nil {
		return nil, err
	}

	chat.updateChatFromGroupMembershipChanges(group)

	return m.addMessagesAndChat(chat, buildSystemMessages([]v1protocol.MembershipUpdateEvent{event}, m.systemMessagesTranslations), &response)
}

func (m *Messenger) ChangeGroupChatName(ctx context.Context, chatID string, name string) (*MessengerResponse, error) {
	logger := m.logger.With(zap.String("site", "ChangeGroupChatName"))
	logger.Info("Changing group chat name", zap.String("chatID", chatID), zap.String("name", name))

	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}

	group, err := newProtocolGroupFromChat(chat)
	if err != nil {
		return nil, err
	}

	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
	// Change name
	event := v1protocol.NewNameChangedEvent(name, clock)
	event.ChatID = chat.ID
	err = event.Sign(m.identity)
	if err != nil {
		return nil, err
	}

	// Update in-memory group
	err = group.ProcessEvent(event)
	if err != nil {
		return nil, err
	}

	recipients, err := stringSliceToPublicKeys(group.Members())
	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.sender.EncodeMembershipUpdate(group, nil)
	if err != nil {
		return nil, err
	}
	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID: chat.ID,
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
		Recipients:  recipients,
	})

	if err != nil {
		return nil, err
	}

	chat.updateChatFromGroupMembershipChanges(group)

	var response MessengerResponse

	return m.addMessagesAndChat(chat, buildSystemMessages([]v1protocol.MembershipUpdateEvent{event}, m.systemMessagesTranslations), &response)
}

func (m *Messenger) EditGroupChat(ctx context.Context, chatID string, name string, color string, image images.CroppedImage) (*MessengerResponse, error) {
	logger := m.logger.With(zap.String("site", "EditGroupChat"))
	logger.Info("Editing group chat details", zap.String("chatID", chatID), zap.String("name", name), zap.String("color", color))

	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}

	group, err := newProtocolGroupFromChat(chat)
	if err != nil {
		return nil, err
	}

	signAndProcessEvent := func(m *Messenger, event *v1protocol.MembershipUpdateEvent) error {
		err := event.Sign(m.identity)
		if err != nil {
			return err
		}

		err = group.ProcessEvent(*event)
		if err != nil {
			return err
		}

		return nil
	}

	var events []v1protocol.MembershipUpdateEvent

	if chat.Name != name {
		clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
		event := v1protocol.NewNameChangedEvent(name, clock)
		event.ChatID = chat.ID
		err = signAndProcessEvent(m, &event)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if chat.Color != color {
		clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
		event := v1protocol.NewColorChangedEvent(color, clock)
		event.ChatID = chat.ID
		err = signAndProcessEvent(m, &event)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if len(image.ImagePath) > 0 {
		payload, err := images.OpenAndAdjustImage(image, true)

		if err != nil {
			return nil, err
		}

		// prepare event
		clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
		event := v1protocol.NewImageChangedEvent(payload, clock)
		event.ChatID = chat.ID
		err = signAndProcessEvent(m, &event)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	recipients, err := stringSliceToPublicKeys(group.Members())
	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.sender.EncodeMembershipUpdate(group, nil)
	if err != nil {
		return nil, err
	}
	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID: chat.ID,
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
		Recipients:  recipients,
	})

	if err != nil {
		return nil, err
	}

	chat.updateChatFromGroupMembershipChanges(group)

	var response MessengerResponse

	return m.addMessagesAndChat(chat, buildSystemMessages(events, m.systemMessagesTranslations), &response)
}

func (m *Messenger) SendGroupChatInvitationRequest(ctx context.Context, chatID string, adminPK string,
	message string) (*MessengerResponse, error) {
	logger := m.logger.With(zap.String("site", "SendGroupChatInvitationRequest"))
	logger.Info("Sending group chat invitation request", zap.String("chatID", chatID),
		zap.String("adminPK", adminPK), zap.String("message", message))

	var response MessengerResponse

	// Get chat and clock
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	invitationR := &GroupChatInvitation{
		GroupChatInvitation: &protobuf.GroupChatInvitation{
			Clock:               clock,
			ChatId:              chatID,
			IntroductionMessage: message,
			State:               protobuf.GroupChatInvitation_REQUEST,
		},
		From: types.EncodeHex(crypto.FromECDSAPub(&m.identity.PublicKey)),
	}

	encodedMessage, err := proto.Marshal(invitationR.GetProtobuf())
	if err != nil {
		return nil, err
	}

	spec := common.RawMessage{
		LocalChatID:         adminPK,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_GROUP_CHAT_INVITATION,
		ResendAutomatically: true,
	}

	pkey, err := hex.DecodeString(adminPK[2:])
	if err != nil {
		return nil, err
	}
	// Safety check, make sure is well formed
	adminpk, err := crypto.UnmarshalPubkey(pkey)
	if err != nil {
		return nil, err
	}

	id, err := m.sender.SendPrivate(ctx, adminpk, &spec)
	if err != nil {
		return nil, err
	}

	spec.ID = types.EncodeHex(id)
	spec.SendCount++
	err = m.persistence.SaveRawMessage(&spec)
	if err != nil {
		return nil, err
	}

	response.Invitations = []*GroupChatInvitation{invitationR}

	err = m.persistence.SaveInvitation(invitationR)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (m *Messenger) GetGroupChatInvitations() ([]*GroupChatInvitation, error) {
	return m.persistence.GetGroupChatInvitations()
}

func (m *Messenger) SendGroupChatInvitationRejection(ctx context.Context, invitationRequestID string) (*MessengerResponse, error) {
	logger := m.logger.With(zap.String("site", "SendGroupChatInvitationRejection"))
	logger.Info("Sending group chat invitation reject", zap.String("invitationRequestID", invitationRequestID))

	invitationR, err := m.persistence.InvitationByID(invitationRequestID)
	if err != nil {
		return nil, err
	}

	invitationR.State = protobuf.GroupChatInvitation_REJECTED

	// Get chat and clock
	chat, ok := m.allChats.Load(invitationR.ChatId)
	if !ok {
		return nil, ErrChatNotFound
	}
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	invitationR.Clock = clock

	encodedMessage, err := proto.Marshal(invitationR.GetProtobuf())
	if err != nil {
		return nil, err
	}

	spec := common.RawMessage{
		LocalChatID:         invitationR.From,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_GROUP_CHAT_INVITATION,
		ResendAutomatically: true,
	}

	pkey, err := hex.DecodeString(invitationR.From[2:])
	if err != nil {
		return nil, err
	}
	// Safety check, make sure is well formed
	userpk, err := crypto.UnmarshalPubkey(pkey)
	if err != nil {
		return nil, err
	}

	id, err := m.sender.SendPrivate(ctx, userpk, &spec)
	if err != nil {
		return nil, err
	}

	spec.ID = types.EncodeHex(id)
	spec.SendCount++
	err = m.persistence.SaveRawMessage(&spec)
	if err != nil {
		return nil, err
	}

	var response MessengerResponse

	response.Invitations = []*GroupChatInvitation{invitationR}

	err = m.persistence.SaveInvitation(invitationR)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (m *Messenger) AddAdminsToGroupChat(ctx context.Context, chatID string, members []string) (*MessengerResponse, error) {
	var response MessengerResponse
	logger := m.logger.With(zap.String("site", "AddAdminsToGroupChat"))
	logger.Info("Add admins to group chat", zap.String("chatID", chatID), zap.Any("members", members))

	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}

	group, err := newProtocolGroupFromChat(chat)
	if err != nil {
		return nil, err
	}

	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
	// Add members
	event := v1protocol.NewAdminsAddedEvent(members, clock)
	event.ChatID = chat.ID
	err = event.Sign(m.identity)
	if err != nil {
		return nil, err
	}

	err = group.ProcessEvent(event)
	if err != nil {
		return nil, err
	}

	recipients, err := stringSliceToPublicKeys(group.Members())
	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.sender.EncodeMembershipUpdate(group, nil)
	if err != nil {
		return nil, err
	}
	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID: chat.ID,
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
		Recipients:  recipients,
	})

	if err != nil {
		return nil, err
	}

	chat.updateChatFromGroupMembershipChanges(group)
	return m.addMessagesAndChat(chat, buildSystemMessages([]v1protocol.MembershipUpdateEvent{event}, m.systemMessagesTranslations), &response)
}

// Kept only for backward compatibility (auto-join), explicit join has been removed
func (m *Messenger) ConfirmJoiningGroup(ctx context.Context, chatID string) (*MessengerResponse, error) {
	var response MessengerResponse

	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}

	_, err := m.Join(chat)
	if err != nil {
		return nil, err
	}

	group, err := newProtocolGroupFromChat(chat)
	if err != nil {
		return nil, err
	}
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
	event := v1protocol.NewMemberJoinedEvent(
		clock,
	)
	event.ChatID = chat.ID
	err = event.Sign(m.identity)
	if err != nil {
		return nil, err
	}

	err = group.ProcessEvent(event)
	if err != nil {
		return nil, err
	}

	recipients, err := stringSliceToPublicKeys(group.Members())
	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.sender.EncodeMembershipUpdate(group, nil)
	if err != nil {
		return nil, err
	}
	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID: chat.ID,
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
		Recipients:  recipients,
	})
	if err != nil {
		return nil, err
	}

	chat.updateChatFromGroupMembershipChanges(group)
	chat.Joined = int64(m.getTimesource().GetCurrentTime())

	return m.addMessagesAndChat(chat, buildSystemMessages([]v1protocol.MembershipUpdateEvent{event}, m.systemMessagesTranslations), &response)
}

func (m *Messenger) leaveGroupChat(ctx context.Context, response *MessengerResponse, chatID string, remove bool, shouldBeSynced bool) (*MessengerResponse, error) {
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, ErrChatNotFound
	}

	amIMember := chat.HasMember(common.PubkeyToHex(&m.identity.PublicKey))

	if amIMember {
		chat.RemoveMember(common.PubkeyToHex(&m.identity.PublicKey))

		group, err := newProtocolGroupFromChat(chat)
		if err != nil {
			return nil, err
		}
		clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
		event := v1protocol.NewMemberRemovedEvent(
			contactIDFromPublicKey(&m.identity.PublicKey),
			clock,
		)
		event.ChatID = chat.ID
		err = event.Sign(m.identity)
		if err != nil {
			return nil, err
		}

		err = group.ProcessEvent(event)
		if err != nil {
			return nil, err
		}

		recipients, err := stringSliceToPublicKeys(group.Members())
		if err != nil {
			return nil, err
		}

		encodedMessage, err := m.sender.EncodeMembershipUpdate(group, nil)
		if err != nil {
			return nil, err
		}

		// shouldBeSynced is false if we got here because a synced client has already
		// sent the leave group message. In that case we don't need to send it again.
		if shouldBeSynced {
			_, err = m.dispatchMessage(ctx, common.RawMessage{
				LocalChatID: chat.ID,
				Payload:     encodedMessage,
				MessageType: protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE,
				Recipients:  recipients,
			})
			if err != nil {
				return nil, err
			}
		}

		chat.updateChatFromGroupMembershipChanges(group)
		response.AddMessages(buildSystemMessages([]v1protocol.MembershipUpdateEvent{event}, m.systemMessagesTranslations))
		err = m.persistence.SaveMessages(response.Messages())
		if err != nil {
			return nil, err
		}
	}

	if remove {
		chat.Active = false
	}

	if remove && shouldBeSynced {
		err := m.syncChatRemoving(ctx, chat.ID, m.dispatchMessage)
		if err != nil {
			return nil, err
		}
	}

	response.AddChat(chat)

	return response, m.saveChat(chat)
}

func (m *Messenger) LeaveGroupChat(ctx context.Context, chatID string, remove bool) (*MessengerResponse, error) {
	_, err := m.DismissAllActivityCenterNotificationsFromChatID(ctx, chatID, m.GetCurrentTimeInMillis())
	if err != nil {
		return nil, err
	}
	var response MessengerResponse
	return m.leaveGroupChat(ctx, &response, chatID, remove, true)
}

// Decline all pending group invites from a user
func (m *Messenger) DeclineAllPendingGroupInvitesFromUser(ctx context.Context, response *MessengerResponse, userPublicKey string) (*MessengerResponse, error) {

	// Decline group invites from active chats
	chats, err := m.persistence.Chats()
	if err != nil {
		return nil, err
	}

	for _, chat := range chats {
		if chat.ChatType == ChatTypePrivateGroupChat &&
			chat.ReceivedInvitationAdmin == userPublicKey &&
			chat.Joined == 0 && chat.Active {
			response, err = m.leaveGroupChat(ctx, response, chat.ID, true, true)
			if err != nil {
				return nil, err
			}
		}
	}

	// Decline group invites from activity center notifications
	notifications, err := m.AcceptActivityCenterNotificationsForInvitesFromUser(ctx, userPublicKey, m.GetCurrentTimeInMillis())
	if err != nil {
		return nil, err
	}

	for _, notification := range notifications {
		response, err = m.leaveGroupChat(ctx, response, notification.ChatID, true, true)
		if err != nil {
			return nil, err
		}
	}
	return response, nil
}
