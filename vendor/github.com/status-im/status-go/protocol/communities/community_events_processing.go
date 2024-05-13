package communities

import (
	"crypto/ecdsa"
	"errors"
	"sort"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	utils "github.com/status-im/status-go/common"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

var ErrInvalidCommunityEventClock = errors.New("clock for admin event message is outdated")

func (o *Community) processEvents(message *CommunityEventsMessage, lastlyAppliedEvents map[string]uint64) error {
	processor := &eventsProcessor{
		community:           o,
		message:             message,
		logger:              o.config.Logger.Named("eventsProcessor"),
		lastlyAppliedEvents: lastlyAppliedEvents,
	}
	return processor.exec()
}

type eventsProcessor struct {
	community           *Community
	message             *CommunityEventsMessage
	logger              *zap.Logger
	lastlyAppliedEvents map[string]uint64

	eventsToApply []CommunityEvent
}

func (e *eventsProcessor) exec() error {
	e.community.mutex.Lock()
	defer e.community.mutex.Unlock()

	err := e.validateDescription()
	if err != nil {
		return err
	}

	e.filterEvents()
	e.mergeEvents()
	e.retainNewestEventsPerEventTypeID()
	e.sortEvents()
	e.applyEvents()

	return nil
}

func (e *eventsProcessor) validateDescription() error {
	description, err := validateAndGetEventsMessageCommunityDescription(e.message.EventsBaseCommunityDescription, e.community.ControlNode())
	if err != nil {
		return err
	}

	// Control node is the only entity that can apply events from past description.
	// In this case, events are compared against the clocks of the most recently applied events.
	if e.community.IsControlNode() && description.Clock < e.community.config.CommunityDescription.Clock {
		return nil
	}

	if description.Clock != e.community.config.CommunityDescription.Clock {
		return ErrInvalidCommunityEventClock
	}

	return nil
}

func (e *eventsProcessor) validateEvent(event *CommunityEvent) error {
	if e.lastlyAppliedEvents != nil {
		if clock, found := e.lastlyAppliedEvents[event.EventTypeID()]; found && clock >= event.CommunityEventClock {
			return errors.New("event outdated")
		}
	}

	signer, err := event.RecoverSigner()
	if err != nil {
		return err
	}

	return e.community.validateEvent(event, signer)
}

// Filter invalid and outdated events.
func (e *eventsProcessor) filterEvents() {
	for _, ev := range e.message.Events {
		event := ev
		if err := e.validateEvent(&event); err == nil {
			e.eventsToApply = append(e.eventsToApply, event)
		} else {
			e.logger.Warn("invalid community event", zap.String("EventTypeID", event.EventTypeID()), zap.Uint64("clock", event.CommunityEventClock), zap.Error(err))
		}
	}
}

// Merge message's events with community's events.
func (e *eventsProcessor) mergeEvents() {
	if e.community.config.EventsData != nil {
		for _, ev := range e.community.config.EventsData.Events {
			event := ev
			if err := e.validateEvent(&event); err == nil {
				e.eventsToApply = append(e.eventsToApply, event)
			} else {
				// NOTE: this should not happen, events should be validated before they are saved in the db.
				// It has been identified that an invalid event is saved to the database for some reason.
				// The code flow leading to this behavior is not yet known.
				// https://github.com/status-im/status-desktop/issues/14106
				e.logger.Error("invalid community event read from db", zap.String("EventTypeID", event.EventTypeID()), zap.Uint64("clock", event.CommunityEventClock), zap.Error(err))
			}
		}
	}
}

// Keep only the newest event per PropertyTypeID.
func (e *eventsProcessor) retainNewestEventsPerEventTypeID() {
	eventsMap := make(map[string]CommunityEvent)

	for _, event := range e.eventsToApply {
		if existingEvent, found := eventsMap[event.EventTypeID()]; !found || event.CommunityEventClock > existingEvent.CommunityEventClock {
			eventsMap[event.EventTypeID()] = event
		}
	}

	e.eventsToApply = []CommunityEvent{}
	for _, event := range eventsMap {
		e.eventsToApply = append(e.eventsToApply, event)
	}
}

// Sorts events by clock.
func (e *eventsProcessor) sortEvents() {
	sort.Slice(e.eventsToApply, func(i, j int) bool {
		if e.eventsToApply[i].CommunityEventClock == e.eventsToApply[j].CommunityEventClock {
			return e.eventsToApply[i].Type < e.eventsToApply[j].Type
		}
		return e.eventsToApply[i].CommunityEventClock < e.eventsToApply[j].CommunityEventClock
	})
}

func (e *eventsProcessor) applyEvents() {
	if e.community.config.EventsData == nil {
		e.community.config.EventsData = &EventsData{
			EventsBaseCommunityDescription: e.message.EventsBaseCommunityDescription,
		}
	}
	e.community.config.EventsData.Events = e.eventsToApply

	e.community.applyEvents()
}

func (o *Community) applyEvents() {
	if o.config.EventsData == nil {
		return
	}

	for _, event := range o.config.EventsData.Events {
		err := o.applyEvent(event)
		if err != nil {
			o.config.Logger.Warn("failed to apply event", zap.String("EventTypeID", event.EventTypeID()), zap.Uint64("clock", event.CommunityEventClock), zap.Error(err))
		}
	}
}

func (o *Community) applyEvent(communityEvent CommunityEvent) error {
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
			o.banUserFromCommunity(pk, &protobuf.CommunityBanInfo{DeleteAllMessages: false})
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
	case protobuf.CommunityEvent_COMMUNITY_DELETE_BANNED_MEMBER_MESSAGES:
		if o.IsControlNode() {
			pk, err := common.HexToPubkey(communityEvent.MemberToAction)
			if err != nil {
				return err
			}

			err = o.deleteBannedMemberAllMessages(pk)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *Community) addNewCommunityEvent(event *CommunityEvent) error {
	err := event.Validate()
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

func (o *Community) toCommunityEventsMessage() *CommunityEventsMessage {
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
