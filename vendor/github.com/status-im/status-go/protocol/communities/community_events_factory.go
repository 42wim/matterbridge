package communities

import "github.com/status-im/status-go/protocol/protobuf"

func (o *Community) ToCreateChannelCommunityEvent(channelID string, channel *protobuf.CommunityChat) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CHANNEL_CREATE,
		ChannelData: &protobuf.ChannelData{
			ChannelId: channelID,
			Channel:   channel,
		},
	}
}

func (o *Community) ToEditChannelCommunityEvent(channelID string, channel *protobuf.CommunityChat) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CHANNEL_EDIT,
		ChannelData: &protobuf.ChannelData{
			ChannelId: channelID,
			Channel:   channel,
		},
	}
}

func (o *Community) ToDeleteChannelCommunityEvent(channelID string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CHANNEL_DELETE,
		ChannelData: &protobuf.ChannelData{
			ChannelId: channelID,
		},
	}
}

func (o *Community) ToReorderChannelCommunityEvent(categoryID string, channelID string, position int) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
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
		CommunityEventClock: o.nextEventClock(),
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
		CommunityEventClock: o.nextEventClock(),
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
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CATEGORY_DELETE,
		CategoryData: &protobuf.CategoryData{
			CategoryId: categoryID,
		},
	}
}

func (o *Community) ToReorderCategoryCommunityEvent(categoryID string, position int) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_CATEGORY_REORDER,
		CategoryData: &protobuf.CategoryData{
			CategoryId: categoryID,
			Position:   int32(position),
		},
	}
}

func (o *Community) ToBanCommunityMemberCommunityEvent(pubkey string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN,
		MemberToAction:      pubkey,
	}
}

func (o *Community) ToDeleteAllMemberMessagesEvent(pubkey string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_DELETE_BANNED_MEMBER_MESSAGES,
		MemberToAction:      pubkey,
	}
}

func (o *Community) ToUnbanCommunityMemberCommunityEvent(pubkey string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN,
		MemberToAction:      pubkey,
	}
}

func (o *Community) ToKickCommunityMemberCommunityEvent(pubkey string) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK,
		MemberToAction:      pubkey,
	}
}

func (o *Community) ToCommunityEditCommunityEvent(description *protobuf.CommunityDescription) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
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
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE,
		TokenPermission:     permission,
	}
}

func (o *Community) ToCommunityTokenPermissionDeleteCommunityEvent(permission *protobuf.CommunityTokenPermission) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE,
		TokenPermission:     permission,
	}
}

func (o *Community) ToCommunityRequestToJoinAcceptCommunityEvent(member string, request *protobuf.CommunityRequestToJoin) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT,
		MemberToAction:      member,
		RequestToJoin:       request,
	}
}

func (o *Community) ToCommunityRequestToJoinRejectCommunityEvent(member string, request *protobuf.CommunityRequestToJoin) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT,
		MemberToAction:      member,
		RequestToJoin:       request,
	}
}

func (o *Community) ToAddTokenMetadataCommunityEvent(tokenMetadata *protobuf.CommunityTokenMetadata) *CommunityEvent {
	return &CommunityEvent{
		CommunityEventClock: o.nextEventClock(),
		Type:                protobuf.CommunityEvent_COMMUNITY_TOKEN_ADD,
		TokenMetadata:       tokenMetadata,
	}
}

func (o *Community) nextEventClock() uint64 {
	latestEventClock := uint64(0)
	if o.config.EventsData != nil {
		for _, event := range o.config.EventsData.Events {
			if event.CommunityEventClock > latestEventClock {
				latestEventClock = event.CommunityEventClock
			}
		}
	}

	clock := o.config.CommunityDescription.Clock
	if latestEventClock > clock {
		clock = latestEventClock
	}

	// lamport timestamp
	timestamp := o.timesource.GetCurrentTime()
	if clock == 0 || clock < timestamp {
		clock = timestamp
	} else {
		clock = clock + 1
	}

	return clock
}
