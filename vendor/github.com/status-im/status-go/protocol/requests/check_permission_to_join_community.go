package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var (
	ErrCheckPermissionToJoinCommunityInvalidID = errors.New("check-permission-to-join-community: invalid id")
)

type CheckPermissionToJoinCommunity struct {
	CommunityID types.HexBytes `json:"communityId"`
	Addresses   []string       `json:"addresses"`
}

func (u *CheckPermissionToJoinCommunity) Validate() error {
	if len(u.CommunityID) == 0 {
		return ErrCheckPermissionToJoinCommunityInvalidID
	}

	return nil
}
