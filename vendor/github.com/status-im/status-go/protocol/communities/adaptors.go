package communities

import (
	"github.com/status-im/status-go/protocol/protobuf"
)

func (o *Community) ToSyncInstallationCommunityProtobuf(clock uint64, communitySettings *CommunitySettings, syncControlNode *protobuf.SyncCommunityControlNode) (*protobuf.SyncInstallationCommunity, error) {
	wrappedCommunity, err := o.ToProtocolMessageBytes()
	if err != nil {
		return nil, err
	}

	var rtjs []*protobuf.SyncCommunityRequestsToJoin
	reqs := o.RequestsToJoin()
	for _, req := range reqs {
		rtjs = append(rtjs, req.ToSyncProtobuf())
	}

	settings := &protobuf.SyncCommunitySettings{
		Clock:                        clock,
		CommunityId:                  o.IDString(),
		HistoryArchiveSupportEnabled: true,
	}

	if communitySettings != nil {
		settings.HistoryArchiveSupportEnabled = communitySettings.HistoryArchiveSupportEnabled
	}

	return &protobuf.SyncInstallationCommunity{
		Clock:          clock,
		Id:             o.ID(),
		Description:    wrappedCommunity,
		Joined:         o.Joined(),
		JoinedAt:       o.JoinedAt(),
		Verified:       o.Verified(),
		Muted:          o.Muted(),
		RequestsToJoin: rtjs,
		Settings:       settings,
		ControlNode:    syncControlNode,
		LastOpenedAt:   o.LastOpenedAt(),
	}, nil
}
