package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var (
	ErrCheckAllCommunityChannelsPermissionsInvalidID = errors.New("check-community-channel-permissions: invalid id")
)

type CheckAllCommunityChannelsPermissions struct {
	CommunityID types.HexBytes
	Addresses   []string `json:"addresses"`
}

func (u *CheckAllCommunityChannelsPermissions) Validate() error {
	if len(u.CommunityID) == 0 {
		return ErrCheckAllCommunityChannelsPermissionsInvalidID
	}

	return nil
}
