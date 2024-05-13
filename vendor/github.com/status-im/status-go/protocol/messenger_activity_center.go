package protocol

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	v1protocol "github.com/status-im/status-go/protocol/v1"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
)

var errOnlyOneNotificationID = errors.New("only one notification id is supported")

func toHexBytes(b [][]byte) []types.HexBytes {
	hb := make([]types.HexBytes, len(b))

	for i, v := range b {
		hb[i] = types.HexBytes(v)
	}

	return hb
}

func fromHexBytes(hb []types.HexBytes) [][]byte {
	b := make([][]byte, len(hb))

	for i, v := range hb {
		b[i] = v
	}

	return b
}

func (m *Messenger) ActivityCenterNotifications(request ActivityCenterNotificationsRequest) (*ActivityCenterPaginationResponse, error) {
	cursor, notifications, err := m.persistence.ActivityCenterNotifications(request.Cursor, request.Limit, request.ActivityTypes, request.ReadType, true)
	if err != nil {
		return nil, err
	}

	if m.httpServer != nil {
		for _, notification := range notifications {
			if notification.Message != nil {
				err = m.prepareMessage(notification.Message, m.httpServer)

				if err != nil {
					return nil, err
				}

				image := notification.Message.GetImage()
				if image != nil && image.AlbumId != "" {
					album, err := m.persistence.albumMessages(notification.Message.LocalChatID, image.AlbumId)
					if err != nil {
						return nil, err
					}
					notification.AlbumMessages = album
				}
			}
			if notification.AlbumMessages != nil {
				for _, message := range notification.AlbumMessages {
					err = m.prepareMessage(message, m.httpServer)

					if err != nil {
						return nil, err
					}
				}
			}
			if notification.TokenData != nil {
				if notification.Type == ActivityCenterNotificationTypeCommunityTokenReceived || notification.Type == ActivityCenterNotificationTypeFirstCommunityTokenReceived {
					err = m.prepareTokenData(notification.TokenData, m.httpServer)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return &ActivityCenterPaginationResponse{
		Cursor:        cursor,
		Notifications: notifications,
	}, nil
}

func (m *Messenger) ActivityCenterNotificationsCount(request ActivityCenterCountRequest) (*ActivityCenterCountResponse, error) {
	response := make(ActivityCenterCountResponse)

	for _, activityType := range request.ActivityTypes {
		count, err := m.persistence.ActivityCenterNotificationsCount([]ActivityCenterType{activityType}, request.ReadType, true)
		if err != nil {
			return nil, err
		}

		response[activityType] = count
	}

	return &response, nil
}

func (m *Messenger) HasUnseenActivityCenterNotifications() (bool, error) {
	seen, _, err := m.persistence.HasUnseenActivityCenterNotifications()
	return seen, err
}

func (m *Messenger) GetActivityCenterState() (*ActivityCenterState, error) {
	return m.persistence.GetActivityCenterState()
}

func (m *Messenger) MarkAsSeenActivityCenterNotifications() (*MessengerResponse, error) {
	response := &MessengerResponse{}
	s := &ActivityCenterState{
		UpdatedAt: m.GetCurrentTimeInMillis(),
		HasSeen:   true,
	}
	_, err := m.persistence.UpdateActivityCenterNotificationState(s)
	if err != nil {
		return nil, err
	}

	state, err := m.persistence.GetActivityCenterState()
	if err != nil {
		return nil, err
	}

	response.SetActivityCenterState(state)
	return response, nil
}

func (m *Messenger) MarkAllActivityCenterNotificationsRead(ctx context.Context) (*MessengerResponse, error) {
	ids, err := m.persistence.GetNotReadActivityCenterNotificationIds()
	if err != nil {
		return nil, err
	}

	updateAt := m.GetCurrentTimeInMillis()
	return m.MarkActivityCenterNotificationsRead(ctx, toHexBytes(ids), updateAt, true)
}

func (m *Messenger) MarkActivityCenterNotificationsRead(ctx context.Context, ids []types.HexBytes, updatedAt uint64, sync bool) (*MessengerResponse, error) {
	// Mark notifications as read in the database
	if updatedAt == 0 {
		updatedAt = m.GetCurrentTimeInMillis()
	}
	err := m.persistence.MarkActivityCenterNotificationsRead(ids, updatedAt)
	if err != nil {
		return nil, err
	}

	notifications, err := m.persistence.GetActivityCenterNotificationsByID(ids)
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}
	repliesAndMentions := make(map[string][]string)

	// When marking as read Mention or Reply notification, the corresponding chat message should also be seen.
	for _, notification := range notifications {
		response.AddActivityCenterNotification(notification)

		if notification.Message != nil &&
			(notification.Type == ActivityCenterNotificationTypeMention || notification.Type == ActivityCenterNotificationTypeReply) {
			repliesAndMentions[notification.ChatID] = append(repliesAndMentions[notification.ChatID], notification.Message.ID)
		}
	}

	// Mark messages as seen
	for chatID, messageIDs := range repliesAndMentions {
		count, countWithMentions, chat, err := m.markMessagesSeenImpl(chatID, messageIDs)
		if err != nil {
			return nil, err
		}
		response.AddChat(chat)
		response.AddSeenAndUnseenMessages(&SeenUnseenMessages{
			ChatID:            chatID,
			Count:             count,
			CountWithMentions: countWithMentions,
			Seen:              true,
		})
	}

	state, err := m.persistence.GetActivityCenterState()
	if err != nil {
		return nil, err
	}

	response.SetActivityCenterState(state)

	if !sync {
		response2, err := m.processActivityCenterNotifications(notifications, true)
		if err != nil {
			return nil, err
		}
		if err = response2.Merge(response); err != nil {
			return nil, err
		}
		return response2, nil
	}
	return response, m.syncActivityCenterReadByIDs(ctx, ids, updatedAt)
}

func (m *Messenger) MarkActivityCenterNotificationsUnread(ctx context.Context, ids []types.HexBytes, updatedAt uint64, sync bool) (*MessengerResponse, error) {
	notifications, err := m.persistence.MarkActivityCenterNotificationsUnread(ids, updatedAt)
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}
	response.AddActivityCenterNotifications(notifications)

	// Don't mark messages unseen in chat, that looks weird

	state, err := m.persistence.GetActivityCenterState()
	if err != nil {
		return nil, err
	}
	response.SetActivityCenterState(state)

	if sync && len(notifications) > 0 {
		err = m.syncActivityCenterUnreadByIDs(ctx, ids, updatedAt)
	}
	return response, err
}

func (m *Messenger) MarkActivityCenterNotificationsDeleted(ctx context.Context, ids []types.HexBytes, updatedAt uint64, sync bool) (*MessengerResponse, error) {
	response := &MessengerResponse{}
	notifications, err := m.persistence.MarkActivityCenterNotificationsDeleted(ids, updatedAt)
	if err != nil {
		return nil, err
	}
	response.AddActivityCenterNotifications(notifications)
	state, err := m.persistence.GetActivityCenterState()
	if err != nil {
		return nil, err
	}
	response.SetActivityCenterState(state)
	if sync {
		err = m.syncActivityCenterDeletedByIDs(ctx, ids, updatedAt)
		if err != nil {
			m.logger.Error("MarkActivityCenterNotificationsDeleted, failed to sync activity center notifications as deleted", zap.Error(err))
			return nil, err
		}
	}
	return response, nil
}

func (m *Messenger) addActivityCenterNotification(response *MessengerResponse, notification *ActivityCenterNotification, syncAction func(context.Context, []types.HexBytes, uint64) error) error {
	_, err := m.persistence.SaveActivityCenterNotification(notification, true)
	if err != nil {
		m.logger.Error("failed to save notification", zap.Error(err))
		return err
	}

	state, err := m.persistence.GetActivityCenterState()
	if err != nil {
		m.logger.Error("failed to obtain activity center state", zap.Error(err))
		return err
	}
	response.AddActivityCenterNotification(notification)
	response.SetActivityCenterState(state)

	if syncAction != nil {
		//TODO a way to pass context
		err = syncAction(context.TODO(), []types.HexBytes{notification.ID}, notification.UpdatedAt)
		if err != nil {
			m.logger.Error("[addActivityCenterNotification] failed to sync activity center notification", zap.Error(err))
			return err
		}
	}
	return nil
}

func (m *Messenger) syncActivityCenterReadByIDs(ctx context.Context, ids []types.HexBytes, clock uint64) error {
	syncMessage := &protobuf.SyncActivityCenterRead{
		Clock: clock,
		Ids:   fromHexBytes(ids),
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	return m.sendToPairedDevices(ctx, common.RawMessage{
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACTIVITY_CENTER_READ,
		ResendAutomatically: true,
	})
}

func (m *Messenger) syncActivityCenterUnreadByIDs(ctx context.Context, ids []types.HexBytes, clock uint64) error {
	syncMessage := &protobuf.SyncActivityCenterUnread{
		Clock: clock,
		Ids:   fromHexBytes(ids),
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	return m.sendToPairedDevices(ctx, common.RawMessage{
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACTIVITY_CENTER_UNREAD,
		ResendAutomatically: true,
	})
}

func (m *Messenger) processActivityCenterNotifications(notifications []*ActivityCenterNotification, addNotifications bool) (*MessengerResponse, error) {
	response := &MessengerResponse{}
	var chats []*Chat
	for _, notification := range notifications {
		if notification.ChatID != "" {
			chat, ok := m.allChats.Load(notification.ChatID)
			if !ok {
				// This should not really happen, but ignore just in case it was deleted in the meantime
				m.logger.Warn("chat not found")
				continue
			}
			chat.Active = true

			if chat.PrivateGroupChat() {
				// Send Joined message for backward compatibility
				_, err := m.ConfirmJoiningGroup(context.Background(), chat.ID)
				if err != nil {
					m.logger.Error("failed to join group", zap.Error(err))
					return nil, err
				}
			}

			chats = append(chats, chat)
			response.AddChat(chat)
		}

		if addNotifications {
			response.AddActivityCenterNotification(notification)
		}
	}
	if len(chats) != 0 {
		err := m.saveChats(chats)
		if err != nil {
			return nil, err
		}
	}
	return response, nil
}

func (m *Messenger) processAcceptedActivityCenterNotifications(ctx context.Context, notifications []*ActivityCenterNotification, sync bool) (*MessengerResponse, error) {
	ids := make([]types.HexBytes, len(notifications))

	for i := range notifications {
		ids[i] = notifications[i].ID
	}

	state, err := m.persistence.GetActivityCenterState()
	if err != nil {
		return nil, err
	}

	if sync {
		err = m.syncActivityCenterAcceptedByIDs(ctx, ids, m.GetCurrentTimeInMillis())
		if err != nil {
			return nil, err
		}
	}

	response, err := m.processActivityCenterNotifications(notifications, !sync)
	if err != nil {
		return nil, err
	}
	response.SetActivityCenterState(state)
	return response, nil
}

func (m *Messenger) AcceptActivityCenterNotificationsForInvitesFromUser(ctx context.Context, userPublicKey string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	notifications, err := m.persistence.AcceptActivityCenterNotificationsForInvitesFromUser(userPublicKey, updatedAt)
	if err != nil {
		return nil, err
	}
	if len(notifications) > 0 {
		err = m.syncActivityCenterAccepted(ctx, notifications, updatedAt)
	}
	return notifications, err
}

func (m *Messenger) syncActivityCenterAccepted(ctx context.Context, notifications []*ActivityCenterNotification, updatedAt uint64) error {
	ids := make([]types.HexBytes, len(notifications))
	for _, notification := range notifications {
		ids = append(ids, notification.ID)
	}
	return m.syncActivityCenterAcceptedByIDs(ctx, ids, updatedAt)
}

func (m *Messenger) syncActivityCenterAcceptedByIDs(ctx context.Context, ids []types.HexBytes, clock uint64) error {
	syncMessage := &protobuf.SyncActivityCenterAccepted{
		Clock: clock,
		Ids:   fromHexBytes(ids),
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	return m.sendToPairedDevices(ctx, common.RawMessage{
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACTIVITY_CENTER_ACCEPTED,
		ResendAutomatically: true,
	})
}

func (m *Messenger) syncActivityCenterCommunityRequestDecisionAdapter(ctx context.Context, ids []types.HexBytes, _ uint64) error {
	if len(ids) != 1 {
		return errOnlyOneNotificationID
	}
	id := ids[0]
	notification, err := m.persistence.GetActivityCenterNotificationByID(id)
	if err != nil {
		return err
	}

	return m.syncActivityCenterCommunityRequestDecision(ctx, notification)
}

func (m *Messenger) syncActivityCenterCommunityRequestDecision(ctx context.Context, notification *ActivityCenterNotification) error {
	var decision protobuf.SyncActivityCenterCommunityRequestDecisionCommunityRequestDecision
	if notification.Accepted {
		decision = protobuf.SyncActivityCenterCommunityRequestDecision_ACCEPTED
	} else if notification.Dismissed {
		decision = protobuf.SyncActivityCenterCommunityRequestDecision_DECLINED
	} else {
		return errors.New("[syncActivityCenterCommunityRequestDecision] notification is not accepted or dismissed")
	}

	syncMessage := &protobuf.SyncActivityCenterCommunityRequestDecision{
		Clock:            notification.UpdatedAt,
		Id:               notification.ID,
		MembershipStatus: uint32(notification.MembershipStatus),
		Decision:         decision,
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	return m.sendToPairedDevices(ctx, common.RawMessage{
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACTIVITY_CENTER_COMMUNITY_REQUEST_DECISION,
		ResendAutomatically: true,
	})
}

func (m *Messenger) AcceptActivityCenterNotifications(ctx context.Context, ids []types.HexBytes, updatedAt uint64, sync bool) (*MessengerResponse, error) {
	if len(ids) == 0 {
		return nil, errors.New("notifications ids are not provided")
	}

	notifications, err := m.persistence.AcceptActivityCenterNotifications(ids, updatedAt)
	if err != nil {
		return nil, err
	}

	return m.processAcceptedActivityCenterNotifications(ctx, notifications, sync)
}

func (m *Messenger) DismissAllActivityCenterNotificationsFromUser(ctx context.Context, userPublicKey string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	notifications, err := m.persistence.DismissAllActivityCenterNotificationsFromUser(userPublicKey, updatedAt)
	if err != nil {
		return nil, err
	}
	if notifications == nil {
		return nil, nil
	}
	return notifications, m.syncActivityCenterDismissed(ctx, notifications, updatedAt)
}

func (m *Messenger) DismissActivityCenterNotificationsByCommunity(ctx context.Context, request *requests.DismissCommunityNotifications) error {
	err := request.Validate()
	if err != nil {
		return err
	}

	updatedAt := m.GetCurrentTimeInMillis()
	notifications, err := m.persistence.DismissActivityCenterNotificationsByCommunity(request.CommunityID.String(), updatedAt)
	if err != nil {
		return err
	}
	return m.syncActivityCenterDismissed(ctx, notifications, updatedAt)
}

func (m *Messenger) DismissAllActivityCenterNotificationsFromCommunity(ctx context.Context, communityID string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	notifications, err := m.persistence.DismissAllActivityCenterNotificationsFromCommunity(communityID, updatedAt)
	if err != nil {
		return nil, err
	}
	return notifications, m.syncActivityCenterDismissed(ctx, notifications, updatedAt)
}

func (m *Messenger) DismissAllActivityCenterNotificationsFromChatID(ctx context.Context, chatID string, updatedAt uint64) ([]*ActivityCenterNotification, error) {
	notifications, err := m.persistence.DismissAllActivityCenterNotificationsFromChatID(chatID, updatedAt)
	if err != nil {
		return nil, err
	}
	return notifications, m.syncActivityCenterDismissed(ctx, notifications, updatedAt)
}

func (m *Messenger) syncActivityCenterDeleted(ctx context.Context, notifications []*ActivityCenterNotification, updatedAt uint64) error {
	ids := make([]types.HexBytes, len(notifications))
	for _, notification := range notifications {
		ids = append(ids, notification.ID)
	}
	return m.syncActivityCenterDeletedByIDs(ctx, ids, updatedAt)
}

func (m *Messenger) syncActivityCenterDeletedByIDs(ctx context.Context, ids []types.HexBytes, clock uint64) error {
	syncMessage := &protobuf.SyncActivityCenterDeleted{
		Clock: clock,
		Ids:   fromHexBytes(ids),
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	return m.sendToPairedDevices(ctx, common.RawMessage{
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACTIVITY_CENTER_DELETED,
		ResendAutomatically: true,
	})
}

func (m *Messenger) syncActivityCenterDismissed(ctx context.Context, notifications []*ActivityCenterNotification, updatedAt uint64) error {
	ids := make([]types.HexBytes, len(notifications))
	for _, notification := range notifications {
		ids = append(ids, notification.ID)
	}
	return m.syncActivityCenterDismissedByIDs(ctx, ids, updatedAt)
}

func (m *Messenger) syncActivityCenterDismissedByIDs(ctx context.Context, ids []types.HexBytes, clock uint64) error {
	syncMessage := &protobuf.SyncActivityCenterDismissed{
		Clock: clock,
		Ids:   fromHexBytes(ids),
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	return m.sendToPairedDevices(ctx, common.RawMessage{
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACTIVITY_CENTER_DISMISSED,
		ResendAutomatically: true,
	})
}

func (m *Messenger) DismissActivityCenterNotifications(ctx context.Context, ids []types.HexBytes, updatedAt uint64, sync bool) (*MessengerResponse, error) {
	if updatedAt == 0 {
		updatedAt = m.GetCurrentTimeInMillis()
	}
	err := m.persistence.DismissActivityCenterNotifications(ids, updatedAt)
	if err != nil {
		return nil, err
	}

	state, err := m.persistence.GetActivityCenterState()
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}
	response.SetActivityCenterState(state)
	if !sync {
		notifications, err := m.persistence.GetActivityCenterNotificationsByID(ids)
		if err != nil {
			return nil, err
		}
		response2, err := m.processActivityCenterNotifications(notifications, true)
		if err != nil {
			return nil, err
		}
		err = response2.Merge(response)
		return response, err
	}
	return response, m.syncActivityCenterDismissedByIDs(ctx, ids, updatedAt)
}

func (m *Messenger) ActivityCenterNotification(id types.HexBytes) (*ActivityCenterNotification, error) {
	notification, err := m.persistence.GetActivityCenterNotificationByID(id)
	if err != nil {
		return nil, err
	}

	if notification.Message != nil {
		image := notification.Message.GetImage()
		if image != nil && image.AlbumId != "" {
			album, err := m.persistence.albumMessages(notification.Message.LocalChatID, image.AlbumId)
			if err != nil {
				return nil, err
			}
			notification.AlbumMessages = album
		}
	}

	return notification, nil
}

func (m *Messenger) HandleSyncActivityCenterRead(state *ReceivedMessageState, message *protobuf.SyncActivityCenterRead, statusMessage *v1protocol.StatusMessage) error {
	resp, err := m.MarkActivityCenterNotificationsRead(context.TODO(), toHexBytes(message.Ids), message.Clock, false)

	if err != nil {
		return err
	}

	return state.Response.Merge(resp)
}

func (m *Messenger) HandleSyncActivityCenterUnread(state *ReceivedMessageState, message *protobuf.SyncActivityCenterUnread, statusMessage *v1protocol.StatusMessage) error {
	resp, err := m.MarkActivityCenterNotificationsUnread(context.TODO(), toHexBytes(message.Ids), message.Clock, false)

	if err != nil {
		return err
	}

	return state.Response.Merge(resp)
}

func (m *Messenger) HandleSyncActivityCenterDeleted(state *ReceivedMessageState, message *protobuf.SyncActivityCenterDeleted, statusMessage *v1protocol.StatusMessage) error {
	response, err := m.MarkActivityCenterNotificationsDeleted(context.TODO(), toHexBytes(message.Ids), message.Clock, false)
	if err != nil {
		return err
	}
	return state.Response.Merge(response)
}

func (m *Messenger) HandleSyncActivityCenterAccepted(state *ReceivedMessageState, message *protobuf.SyncActivityCenterAccepted, statusMessage *v1protocol.StatusMessage) error {
	resp, err := m.AcceptActivityCenterNotifications(context.TODO(), toHexBytes(message.Ids), message.Clock, false)

	if err != nil {
		return err
	}

	return state.Response.Merge(resp)
}

func (m *Messenger) HandleSyncActivityCenterDismissed(state *ReceivedMessageState, message *protobuf.SyncActivityCenterDismissed, statusMessage *v1protocol.StatusMessage) error {
	resp, err := m.DismissActivityCenterNotifications(context.TODO(), toHexBytes(message.Ids), message.Clock, false)

	if err != nil {
		return err
	}

	return state.Response.Merge(resp)
}

func (m *Messenger) HandleSyncActivityCenterCommunityRequestDecision(state *ReceivedMessageState, a *protobuf.SyncActivityCenterCommunityRequestDecision, statusMessage *v1protocol.StatusMessage) error {
	notification, err := m.persistence.GetActivityCenterNotificationByID(a.Id)
	if err != nil {
		return err
	}
	if notification == nil {
		return errors.New("[HandleSyncActivityCenterCommunityRequestDecision] notification not found")
	}

	notification.MembershipStatus = ActivityCenterMembershipStatus(a.MembershipStatus)
	notification.UpdatedAt = a.Clock
	if a.Decision == protobuf.SyncActivityCenterCommunityRequestDecision_DECLINED {
		notification.Dismissed = true
	} else if a.Decision == protobuf.SyncActivityCenterCommunityRequestDecision_ACCEPTED {
		notification.Accepted = true
	} else {
		return errors.New("[HandleSyncActivityCenterCommunityRequestDecision] invalid decision")
	}
	_, err = m.persistence.SaveActivityCenterNotification(notification, false)
	if err != nil {
		return err
	}

	resp := state.Response
	resp.AddActivityCenterNotification(notification)

	s, err := m.persistence.UpdateActivityCenterState(notification.UpdatedAt)
	if err != nil {
		return err
	}
	resp.SetActivityCenterState(s)
	return nil
}
