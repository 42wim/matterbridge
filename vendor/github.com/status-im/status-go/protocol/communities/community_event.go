package communities

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/protobuf"
)

type CommunityEvent struct {
	CommunityEventClock uint64                             `json:"communityEventClock"`
	Type                protobuf.CommunityEvent_EventType  `json:"type"`
	CommunityConfig     *protobuf.CommunityConfig          `json:"communityConfig,omitempty"`
	TokenPermission     *protobuf.CommunityTokenPermission `json:"tokenPermissions,omitempty"`
	CategoryData        *protobuf.CategoryData             `json:"categoryData,omitempty"`
	ChannelData         *protobuf.ChannelData              `json:"channelData,omitempty"`
	MemberToAction      string                             `json:"memberToAction,omitempty"`
	RequestToJoin       *protobuf.CommunityRequestToJoin   `json:"requestToJoin,omitempty"`
	TokenMetadata       *protobuf.CommunityTokenMetadata   `json:"tokenMetadata,omitempty"`
	Payload             []byte                             `json:"payload"`
	Signature           []byte                             `json:"signature"`
}

func (e *CommunityEvent) ToProtobuf() *protobuf.CommunityEvent {
	var acceptedRequestsToJoin map[string]*protobuf.CommunityRequestToJoin
	var rejectedRequestsToJoin map[string]*protobuf.CommunityRequestToJoin

	switch e.Type {
	case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT:
		acceptedRequestsToJoin = make(map[string]*protobuf.CommunityRequestToJoin)
		acceptedRequestsToJoin[e.MemberToAction] = e.RequestToJoin
	case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT:
		rejectedRequestsToJoin = make(map[string]*protobuf.CommunityRequestToJoin)
		rejectedRequestsToJoin[e.MemberToAction] = e.RequestToJoin
	}

	return &protobuf.CommunityEvent{
		CommunityEventClock:    e.CommunityEventClock,
		Type:                   e.Type,
		CommunityConfig:        e.CommunityConfig,
		TokenPermission:        e.TokenPermission,
		CategoryData:           e.CategoryData,
		ChannelData:            e.ChannelData,
		MemberToAction:         e.MemberToAction,
		RejectedRequestsToJoin: rejectedRequestsToJoin,
		AcceptedRequestsToJoin: acceptedRequestsToJoin,
		TokenMetadata:          e.TokenMetadata,
	}
}

func communityEventFromProtobuf(msg *protobuf.SignedCommunityEvent) (*CommunityEvent, error) {
	decodedEvent := protobuf.CommunityEvent{}
	err := proto.Unmarshal(msg.Payload, &decodedEvent)
	if err != nil {
		return nil, err
	}

	memberToAction := decodedEvent.MemberToAction
	var requestToJoin *protobuf.CommunityRequestToJoin

	switch decodedEvent.Type {
	case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT:
		for member, request := range decodedEvent.AcceptedRequestsToJoin {
			memberToAction = member
			requestToJoin = request
			break
		}
	case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT:
		for member, request := range decodedEvent.RejectedRequestsToJoin {
			memberToAction = member
			requestToJoin = request
			break
		}
	}

	return &CommunityEvent{
		CommunityEventClock: decodedEvent.CommunityEventClock,
		Type:                decodedEvent.Type,
		CommunityConfig:     decodedEvent.CommunityConfig,
		TokenPermission:     decodedEvent.TokenPermission,
		CategoryData:        decodedEvent.CategoryData,
		ChannelData:         decodedEvent.ChannelData,
		MemberToAction:      memberToAction,
		RequestToJoin:       requestToJoin,
		TokenMetadata:       decodedEvent.TokenMetadata,
		Payload:             msg.Payload,
		Signature:           msg.Signature,
	}, nil
}

func (e *CommunityEvent) RecoverSigner() (*ecdsa.PublicKey, error) {
	if e.Signature == nil || len(e.Signature) == 0 {
		return nil, errors.New("missing signature")
	}

	signer, err := crypto.SigToPub(
		crypto.Keccak256(e.Payload),
		e.Signature,
	)
	if err != nil {
		return nil, errors.New("failed to recover signer")
	}

	return signer, nil
}

func (e *CommunityEvent) Sign(pk *ecdsa.PrivateKey) error {
	sig, err := crypto.Sign(crypto.Keccak256(e.Payload), pk)
	if err != nil {
		return err
	}

	e.Signature = sig
	return nil
}

func (e *CommunityEvent) Validate() error {
	switch e.Type {
	case protobuf.CommunityEvent_COMMUNITY_EDIT:
		if e.CommunityConfig == nil || e.CommunityConfig.Identity == nil ||
			e.CommunityConfig.Permissions == nil || e.CommunityConfig.AdminSettings == nil {
			return errors.New("invalid config change admin event")
		}

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE:
		if e.TokenPermission == nil || len(e.TokenPermission.Id) == 0 {
			return errors.New("invalid token permission change event")
		}

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE:
		if e.TokenPermission == nil || len(e.TokenPermission.Id) == 0 {
			return errors.New("invalid token permission delete event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_CREATE:
		if e.CategoryData == nil || len(e.CategoryData.CategoryId) == 0 {
			return errors.New("invalid community category create event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_DELETE:
		if e.CategoryData == nil || len(e.CategoryData.CategoryId) == 0 {
			return errors.New("invalid community category delete event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_EDIT:
		if e.CategoryData == nil || len(e.CategoryData.CategoryId) == 0 {
			return errors.New("invalid community category edit event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_CREATE:
		if e.ChannelData == nil || len(e.ChannelData.ChannelId) == 0 ||
			e.ChannelData.Channel == nil {
			return errors.New("invalid community channel create event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_DELETE:
		if e.ChannelData == nil || len(e.ChannelData.ChannelId) == 0 {
			return errors.New("invalid community channel delete event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_EDIT:
		if e.ChannelData == nil || len(e.ChannelData.ChannelId) == 0 ||
			e.ChannelData.Channel == nil {
			return errors.New("invalid community channel edit event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_REORDER:
		if e.ChannelData == nil || len(e.ChannelData.ChannelId) == 0 {
			return errors.New("invalid community channel reorder event")
		}

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_REORDER:
		if e.CategoryData == nil || len(e.CategoryData.CategoryId) == 0 {
			return errors.New("invalid community category reorder event")
		}

	case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT, protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT:
		if len(e.MemberToAction) == 0 || e.RequestToJoin == nil {
			return errors.New("invalid community request to join event")
		}

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK:
		if len(e.MemberToAction) == 0 {
			return errors.New("invalid community member kick event")
		}

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN:
		if len(e.MemberToAction) == 0 {
			return errors.New("invalid community member ban event")
		}

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN:
		if len(e.MemberToAction) == 0 {
			return errors.New("invalid community member unban event")
		}

	case protobuf.CommunityEvent_COMMUNITY_TOKEN_ADD:
		if e.TokenMetadata == nil || len(e.TokenMetadata.ContractAddresses) == 0 {
			return errors.New("invalid add community token event")
		}
	case protobuf.CommunityEvent_COMMUNITY_DELETE_BANNED_MEMBER_MESSAGES:
		if len(e.MemberToAction) == 0 {
			return errors.New("invalid delete all community member messages event")
		}
	}
	return nil
}

// EventTypeID constructs a unique identifier for an event and its associated target.
func (e *CommunityEvent) EventTypeID() string {
	switch e.Type {
	case protobuf.CommunityEvent_COMMUNITY_EDIT:
		return fmt.Sprintf("%d", e.Type)

	case protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE,
		protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE:
		return fmt.Sprintf("%d-%s", e.Type, e.TokenPermission.Id)

	case protobuf.CommunityEvent_COMMUNITY_CATEGORY_CREATE,
		protobuf.CommunityEvent_COMMUNITY_CATEGORY_DELETE,
		protobuf.CommunityEvent_COMMUNITY_CATEGORY_EDIT,
		protobuf.CommunityEvent_COMMUNITY_CATEGORY_REORDER:
		return fmt.Sprintf("%d-%s", e.Type, e.CategoryData.CategoryId)

	case protobuf.CommunityEvent_COMMUNITY_CHANNEL_CREATE,
		protobuf.CommunityEvent_COMMUNITY_CHANNEL_DELETE,
		protobuf.CommunityEvent_COMMUNITY_CHANNEL_EDIT,
		protobuf.CommunityEvent_COMMUNITY_CHANNEL_REORDER:
		return fmt.Sprintf("%d-%s", e.Type, e.ChannelData.ChannelId)

	case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT,
		protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT,
		protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK,
		protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN,
		protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN,
		protobuf.CommunityEvent_COMMUNITY_DELETE_BANNED_MEMBER_MESSAGES:
		return fmt.Sprintf("%d-%s", e.Type, e.MemberToAction)

	case protobuf.CommunityEvent_COMMUNITY_TOKEN_ADD:
		return fmt.Sprintf("%d-%s", e.Type, e.TokenMetadata.Name)
	}

	return ""
}

func communityEventsToJSONEncodedBytes(communityEvents []CommunityEvent) ([]byte, error) {
	return json.Marshal(communityEvents)
}

func communityEventsFromJSONEncodedBytes(jsonEncodedRawEvents []byte) ([]CommunityEvent, error) {
	var events []CommunityEvent
	err := json.Unmarshal(jsonEncodedRawEvents, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}
