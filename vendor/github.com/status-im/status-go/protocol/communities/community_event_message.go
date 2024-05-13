package communities

import (
	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/protocol/protobuf"
)

type CommunityEventsMessage struct {
	CommunityID                    []byte           `json:"communityId"`
	EventsBaseCommunityDescription []byte           `json:"eventsBaseCommunityDescription"`
	Events                         []CommunityEvent `json:"events,omitempty"`
}

func (m *CommunityEventsMessage) ToProtobuf() *protobuf.CommunityEventsMessage {
	result := protobuf.CommunityEventsMessage{
		CommunityId:                    m.CommunityID,
		EventsBaseCommunityDescription: m.EventsBaseCommunityDescription,
		SignedEvents:                   []*protobuf.SignedCommunityEvent{},
	}

	for _, event := range m.Events {
		signedEvent := &protobuf.SignedCommunityEvent{
			Signature: event.Signature,
			Payload:   event.Payload,
		}
		result.SignedEvents = append(result.SignedEvents, signedEvent)
	}

	return &result
}

func CommunityEventsMessageFromProtobuf(msg *protobuf.CommunityEventsMessage) (*CommunityEventsMessage, error) {
	result := &CommunityEventsMessage{
		CommunityID:                    msg.CommunityId,
		EventsBaseCommunityDescription: msg.EventsBaseCommunityDescription,
		Events:                         []CommunityEvent{},
	}

	for _, signedEvent := range msg.SignedEvents {
		event, err := communityEventFromProtobuf(signedEvent)
		if err != nil {
			return nil, err
		}
		result.Events = append(result.Events, *event)
	}

	return result, nil
}

func (m *CommunityEventsMessage) Marshal() ([]byte, error) {
	pb := m.ToProtobuf()
	return proto.Marshal(pb)
}
