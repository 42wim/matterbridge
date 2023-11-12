package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common/shard"
)

type SetCommunityShard struct {
	CommunityID types.HexBytes  `json:"communityId"`
	Shard       *shard.Shard    `json:"shard,omitempty"`
	PrivateKey  *types.HexBytes `json:"privateKey,omitempty"`
}

func (s *SetCommunityShard) Validate() error {
	if s == nil {
		return errors.New("invalid request")
	}
	if s.Shard != nil {
		// TODO: for now only MainStatusShard(16) is accepted
		if s.Shard.Cluster != shard.MainStatusShardCluster {
			return errors.New("invalid shard cluster")
		}
		if s.Shard.Index > 1023 {
			return errors.New("invalid shard index. Only 0-1023 is allowed")
		}
	}
	return nil
}
