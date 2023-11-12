package communities

import (
	"crypto/ecdsa"

	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
)

func communityToRecord(community *Community) (*CommunityRecord, error) {
	wrappedDescription, err := community.ToProtocolMessageBytes()
	if err != nil {
		return nil, err
	}

	var shardIndex, shardCluster *uint
	if community.Shard() != nil {
		index := uint(community.Shard().Index)
		shardIndex = &index
		cluster := uint(community.Shard().Cluster)
		shardCluster = &cluster
	}

	return &CommunityRecord{
		id:           community.ID(),
		privateKey:   crypto.FromECDSA(community.PrivateKey()),
		controlNode:  crypto.FromECDSAPub(community.ControlNode()),
		description:  wrappedDescription,
		joined:       community.config.Joined,
		joinedAt:     community.config.JoinedAt,
		lastOpenedAt: community.config.LastOpenedAt,
		verified:     community.config.Verified,
		spectated:    community.config.Spectated,
		muted:        community.config.Muted,
		mutedTill:    community.config.MuteTill,
		shardCluster: shardCluster,
		shardIndex:   shardIndex,
	}, nil
}

func communityToEventsRecord(community *Community) (*EventsRecord, error) {
	if community.config.EventsData == nil {
		return nil, nil
	}

	rawEvents, err := communityEventsToJSONEncodedBytes(community.config.EventsData.Events)
	if err != nil {
		return nil, err
	}

	return &EventsRecord{
		id:             community.ID(),
		rawEvents:      rawEvents,
		rawDescription: community.config.EventsData.EventsBaseCommunityDescription,
	}, nil
}

func recordToRequestToJoin(r *RequestToJoinRecord) *RequestToJoin {
	// FIXME: fill revealed addresses
	return &RequestToJoin{
		ID:          r.id,
		PublicKey:   r.publicKey,
		Clock:       uint64(r.clock),
		ENSName:     r.ensName,
		ChatID:      r.chatID,
		CommunityID: r.communityID,
		State:       RequestToJoinState(r.state),
	}
}

func recordBundleToCommunity(r *CommunityRecordBundle, memberIdentity *ecdsa.PublicKey, installationID string,
	logger *zap.Logger, timesource common.TimeSource, encryptor DescriptionEncryptor, initializer func(*Community) error) (*Community, error) {
	var privateKey *ecdsa.PrivateKey
	var controlNode *ecdsa.PublicKey
	var err error

	if r.community.privateKey != nil {
		privateKey, err = crypto.ToECDSA(r.community.privateKey)
		if err != nil {
			return nil, err
		}
	}
	if r.community.controlNode != nil {
		controlNode, err = crypto.UnmarshalPubkey(r.community.controlNode)
		if err != nil {
			return nil, err
		}
	}

	description, err := decodeWrappedCommunityDescription(r.community.description)
	if err != nil {
		return nil, err
	}

	id, err := crypto.DecompressPubkey(r.community.id)
	if err != nil {
		return nil, err
	}

	var eventsData *EventsData
	if r.events != nil {
		eventsData, err = decodeEventsData(r.events.rawEvents, r.events.rawDescription)
		if err != nil {
			return nil, err

		}
	}

	var s *shard.Shard = nil
	if r.community.shardCluster != nil && r.community.shardIndex != nil {
		s = &shard.Shard{
			Cluster: uint16(*r.community.shardCluster),
			Index:   uint16(*r.community.shardIndex),
		}
	}

	isControlDevice := r.installationID != nil && *r.installationID == installationID

	config := Config{
		PrivateKey:                          privateKey,
		ControlNode:                         controlNode,
		ControlDevice:                       isControlDevice,
		CommunityDescription:                description,
		MemberIdentity:                      memberIdentity,
		CommunityDescriptionProtocolMessage: r.community.description,
		Logger:                              logger,
		ID:                                  id,
		Verified:                            r.community.verified,
		Muted:                               r.community.muted,
		MuteTill:                            r.community.mutedTill,
		Joined:                              r.community.joined,
		JoinedAt:                            r.community.joinedAt,
		LastOpenedAt:                        r.community.lastOpenedAt,
		Spectated:                           r.community.spectated,
		EventsData:                          eventsData,
		Shard:                               s,
	}

	community, err := New(config, timesource, encryptor)
	if err != nil {
		return nil, err
	}

	if r.requestToJoin != nil {
		community.config.RequestedToJoinAt = uint64(r.requestToJoin.clock)
		requestToJoin := recordToRequestToJoin(r.requestToJoin)
		if !requestToJoin.Empty() {
			community.AddRequestToJoin(requestToJoin)
		}
	}

	if initializer != nil {
		err = initializer(community)
		if err != nil {
			return nil, err
		}
	}

	return community, nil
}
