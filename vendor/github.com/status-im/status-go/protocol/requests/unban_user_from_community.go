package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrUnbanUserFromCommunityInvalidCommunityID = errors.New("unban-user-from-community: invalid community id")
var ErrUnbanUserFromCommunityInvalidUser = errors.New("unban-user-from-community: invalid user id")

type UnbanUserFromCommunity struct {
	CommunityID types.HexBytes `json:"communityId"`
	User        types.HexBytes `json:"user"`
}

func (b *UnbanUserFromCommunity) Validate() error {
	if len(b.CommunityID) == 0 {
		return ErrUnbanUserFromCommunityInvalidCommunityID
	}

	if len(b.User) == 0 {
		return ErrUnbanUserFromCommunityInvalidUser
	}

	return nil
}
