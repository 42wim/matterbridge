package requests

import (
	"github.com/status-im/status-go/eth-node/types"
)

type CommunityChannelShareURL struct {
	CommunityID types.HexBytes
	ChannelID   string
}

func (r *CommunityChannelShareURL) Validate() error {
	if len(r.CommunityID) == 0 {
		return ErrCheckPermissionToJoinCommunityInvalidID
	}

	return nil
}
