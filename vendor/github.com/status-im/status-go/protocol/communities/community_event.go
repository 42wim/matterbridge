package communities

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"github.com/golang/protobuf/proto"

	utils "github.com/status-im/status-go/common"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

var ErrInvalidCommunityEventClock = errors.New("clock for admin event message is outdated")

func (o *Community) ToCreateChannelCommunityEvent(channelID string, channel *protobuf.CommunityChat) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CHANNEL_CREATE,
		ChannelData: &protobuf.ChannelData{
			ChannelId: channelID,
			Channel:   channel,
		},
	}
}

func (o *Community) ToEditChannelCommunityEvent(channelID string, channel *protobuf.CommunityChat) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CHANNEL_EDIT,
		ChannelData: &protobuf.ChannelData{
			ChannelId: channelID,
			Channel:   channel,
		},
	}
}

func (o *Community) ToDeleteChannelCommunityEvent(channelID string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CHANNEL_DELETE,
		ChannelData: &protobuf.ChannelData{
			ChannelId: channelID,
		},
	}
}

func (o *Community) ToReorderChannelCommunityEvent(categoryID string, channelID string, position int) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CHANNEL_REORDER,
		ChannelData: &protobuf.ChannelData{
			CategoryId: categoryID,
			ChannelId:  channelID,
			Position:   int32(position),
		},
	}
}

func (o *Community) ToCreateCategoryCommunityEvent(categoryID string, categoryName string, channelsIds []string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CATEGORY_CREATE,
		CategoryData: &protobuf.CategoryData{
			Name:        categoryName,
			CategoryId:  categoryID,
			ChannelsIds: channelsIds,
		},
	}
}

func (o *Community) ToEditCategoryCommunityEvent(categoryID string, categoryName string, channelsIds []string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CATEGORY_EDIT,
		CategoryData: &protobuf.CategoryData{
			Name:        categoryName,
			CategoryId:  categoryID,
			ChannelsIds: channelsIds,
		},
	}
}

func (o *Community) ToDeleteCategoryCommunityEvent(categoryID string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CATEGORY_DELETE,
		CategoryData: &protobuf.CategoryData{
			CategoryId: categoryID,
		},
	}
}

func (o *Community) ToReorderCategoryCommunityEvent(categoryID string, position int) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CATEGORY_REORDER,
		CategoryData: &protobuf.CategoryData{
			CategoryId: categoryID,
			Position:   int32(position),
		},
	}
}

func (o *Community) ToBanCommunityMemberCommunityEvent(pubkey string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN,
		MemberToAction:      pubkey,
	}
}

func (o *Community) ToUnbanCommunityMemberCommunityEvent(pubkey string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN,
		MemberToAction:      pubkey,
	}
}

func (o *Community) ToKickCommunityMemberCommunityEvent(pubkey string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK,
		MemberToAction:      pubkey,
	}
}

func (o *Community) ToCommunityEditCommunityEvent(description *protobuf.CommunityDescription) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_EDIT,
		CommunityConfig: &protobuf.CommunityConfig{
			Identity:      description.Identity,
			Permissions:   description.Permissions,
			AdminSettings: description.AdminSettings,
			IntroMessage:  description.IntroMessage,
			OutroMessage:  description.OutroMessage,
			Tags:          description.Tags,
		},
	}
}

func (o *Community) ToCommunityTokenPermissionChangeCommunityEvent(permission *protobuf.CommunityTokenPermission) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE,
		TokenPermission:     permission,
	}
}

func (o *Community) ToCommunityTokenPermissionDeleteCommunityEvent(permission *protobuf.CommunityTokenPermission) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE,
		TokenPermission:     permission,
	}
}

func (o *Community) ToCommunityRequestToJoinAcceptCommunityEvent(changes *CommunityEventChanges) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock:    o.NewCommunityEventClock(),
		Type:                   protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT,
		AcceptedRequestsToJoin: changes.AcceptedRequestsToJoin,
	}
}

func (o *Community) ToCommunityRequestToJoinRejectCommunityEvent(changes *CommunityEventChanges) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock:    o.NewCommunityEventClock(),
		Type:                   protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT,
		RejectedRequestsToJoin: changes.RejectedRequestsToJoin,
	}
}

func (o *Community) ToAddTokenMetadataCommunityEvent(tokenMetadata *protobuf.CommunityTokenMetadata) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.NewCommunityEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_TOKEN_ADD,
		TokenMetadata:       tokenMetadata,
	}
}

func (o *Community) UpdateCommunityByEvents(communityEventMessage *CommunityEventsMessage) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Validate that EventsBaseCommunityDescription was signed by the control node
	description, err := validateAndGetEventsMessageCommunityDescription(communityEventMessage.EventsBaseCommunityDescription, o.ControlNode())
	if err != nil {
		return err
	}

	if description.Clock != o.config.CommunityDescription.Clock {
		return ErrInvalidCommunityEventClock
	}

	// Merge community events to existing community. Community events must be stored to the db
	// during saving the community
	o.mergeCommunityEvents(communityEventMessage)

	if o.encryptor != nil {
		_, err = decryptDescription(o.ID(), o.encryptor, description, o.config.Logger)
		if err != nil {
			return err
		}
	}

	o.config.CommunityDescription = description
	o.config.CommunityDescriptionProtocolMessage = communityEventMessage.EventsBaseCommunityDescription

	// Update the copy of the CommunityDescription by community events
	err = o.updateCommunityDescriptionByEvents()
	if err != nil {
		return err
	}

	return nil
}

func (o *Community) updateCommunityDescriptionByEvents() error {
	if o.config.EventsData == nil {
		return nil
	}

	for _, event := range o.config.EventsData.Events {
		err := o.updateCommunityDescriptionByCommunityEvent(event)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Community) updateCommunityDescriptionByCommunityEvent(communityEvent CommunityEvent) error {
	switch communityEvent.Type {
	case protobuf.CommunityEvent_COMMUNITY_EDIT:
		o.config.CommunityDescription.Identity = communityEvent.CommunityConfig.Identity
		o.config.CommunityDescription.Permissions = communityEvent.CommunityConfig.Permissions
		o.config.CommunityDescription.AdminSettings = communityEvent.CommunityConfig.AdminSettings
		o.config.CommunityDescription.IntroMessage = communityEvent.CommunityConfig.IntroMessage
		o.config.CommunityDescription.OutroMessage = communityEvent.CommunityConfig.OutroMessage
		o.config.CommunityDescription.Tags = communityEvent.CommunityConfig.Tags

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE:
		if o.IsControlNode() {
			_, err := o.upsertTokenPermission(communityEvent.TokenPermission)
			if err != nil {
				return err
			}
		}

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE:
		if o.IsControlNode() {
			_, err := o.deleteTokenPermission(communityEvent.TokenPermission.Id)
			if err != nil {
				return err
			}
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_CREATE:
		_, err := o.createCategory(communityEvent.CategoryData.CategoryId, communityEvent.CategoryData.Name, communityEvent.CategoryData.ChannelsIds)
		if err != nil {
			return err
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_DELETE:
		_, err := o.deleteCategory(communityEvent.CategoryData.CategoryId)
		if err != nil {
			return err
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_EDIT:
		_, err := o.editCategory(communityEvent.CategoryData.CategoryId, communityEvent.CategoryData.Name, communityEvent.CategoryData.ChannelsIds)
		if err != nil {
			return err
		}

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_CREATE:
		err := o.createChat(communityEvent.ChannelData.ChannelId, communityEvent.ChannelData.Channel)
		if err != nil {
			return err
		}

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_DELETE:
		o.deleteChat(communityEvent.ChannelData.ChannelId)

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_EDIT:
		err := o.editChat(communityEvent.ChannelData.ChannelId, communityEvent.ChannelData.Channel)
		if err != nil {
			return err
		}

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_REORDER:
		_, err := o.reorderChat(communityEvent.ChannelData.CategoryId, communityEvent.ChannelData.ChannelId, int(communityEvent.ChannelData.Position))
		if err != nil {
			return err
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_REORDER:
		_, err := o.reorderCategories(communityEvent.CategoryData.CategoryId, int(communityEvent.CategoryData.Position))
		if err != nil {
			return err
		}

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK:
		if o.IsControlNode() {
			pk, err := common.HexToPubkey(communityEvent.MemberToAction)
			if err != nil {
				return err
			}
			o.removeMemberFromOrg(pk)
		}
	case protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN:
		if o.IsControlNode() {
			pk, err := common.HexToPubkey(communityEvent.MemberToAction)
			if err != nil {
				return err
			}
			o.banUserFromCommunity(pk)
		}
	case protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN:
		if o.IsControlNode() {
			pk, err := common.HexToPubkey(communityEvent.MemberToAction)
			if err != nil {
				return err
			}
			o.unbanUserFromCommunity(pk)
		}
	case protobuf.CommunityEvent_COMMUNITY_TOKEN_ADD:
		o.config.CommunityDescription.CommunityTokensMetadata = append(o.config.CommunityDescription.CommunityTokensMetadata, communityEvent.TokenMetadata)
	}
	return nil
}

func (o *Community) NewCommunityEventClock() uint64 {
	return uint64(time.Now().Unix())
}

func (o *Community) addNewCommunityEvent(event *CommunityEvent) error {
	err := validateCommunityEvent(event)
	if err != nil {
		return err
	}

	// All events must be built on top of the control node CommunityDescription
	// If there were no events before, extract CommunityDescription from CommunityDescriptionProtocolMessage
	// and check the signature
	if o.config.EventsData == nil || len(o.config.EventsData.EventsBaseCommunityDescription) == 0 {
		_, err := validateAndGetEventsMessageCommunityDescription(o.config.CommunityDescriptionProtocolMessage, o.ControlNode())
		if err != nil {
			return err
		}

		o.config.EventsData = &EventsData{
			EventsBaseCommunityDescription: o.config.CommunityDescriptionProtocolMessage,
			Events:                         []CommunityEvent{},
		}
	}

	event.Payload, err = proto.Marshal(event.ToProtobuf())
	if err != nil {
		return err
	}

	o.config.EventsData.Events = append(o.config.EventsData.Events, *event)

	return nil
}

func (o *Community) ToCommunityEventsMessage() *CommunityEventsMessage {
	return &CommunityEventsMessage{
		CommunityID:                    o.ID(),
		EventsBaseCommunityDescription: o.config.EventsData.EventsBaseCommunityDescription,
		Events:                         o.config.EventsData.Events,
	}
}

func validateAndGetEventsMessageCommunityDescription(signedDescription []byte, signerPubkey *ecdsa.PublicKey) (*protobuf.CommunityDescription, error) {
	metadata := &protobuf.ApplicationMetadataMessage{}

	err := proto.Unmarshal(signedDescription, metadata)
	if err != nil {
		return nil, err
	}

	if metadata.Type != protobuf.ApplicationMetadataMessage_COMMUNITY_DESCRIPTION {
		return nil, ErrInvalidMessage
	}

	signer, err := utils.RecoverKey(metadata)
	if err != nil {
		return nil, err
	}

	if signer == nil {
		return nil, errors.New("CommunityDescription does not contain the control node signature")
	}

	if !signer.Equal(signerPubkey) {
		return nil, errors.New("CommunityDescription was not signed by an owner")
	}

	description := &protobuf.CommunityDescription{}

	err = proto.Unmarshal(metadata.Payload, description)
	if err != nil {
		return nil, err
	}

	return description, nil
}
