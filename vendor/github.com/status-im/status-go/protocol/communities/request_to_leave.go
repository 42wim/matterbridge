package communities

import (
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

type RequestToLeave struct {
	ID          types.HexBytes `json:"id"`
	PublicKey   string         `json:"publicKey"`
	Clock       uint64         `json:"clock"`
	CommunityID types.HexBytes `json:"communityId"`
}

func NewRequestToLeave(publicKey string, protobuf *protobuf.CommunityRequestToLeave) *RequestToLeave {
	return &RequestToLeave{
		ID:          CalculateRequestID(publicKey, protobuf.CommunityId),
		PublicKey:   publicKey,
		Clock:       protobuf.Clock,
		CommunityID: protobuf.CommunityId,
	}
}
