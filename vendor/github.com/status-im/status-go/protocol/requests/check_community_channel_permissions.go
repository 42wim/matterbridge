package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var (
	ErrCheckCommunityChannelPermissionsInvalidID     = errors.New("check-community-channel-permissions: invalid id")
	ErrCheckCommunityChannelPermissionsInvalidChatID = errors.New("check-community-channel-permissions: invalid chat id")
)

type CheckCommunityChannelPermissions struct {
	CommunityID types.HexBytes
	ChatID      string
	Addresses   []string `json:"addresses"`
}

func (u *CheckCommunityChannelPermissions) Validate() error {
	if len(u.CommunityID) == 0 {
		return ErrCheckCommunityChannelPermissionsInvalidID
	}
	if len(u.ChatID) == 0 {
		return ErrCheckCommunityChannelPermissionsInvalidChatID
	}

	return nil
}
