package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var (
	ErrDismissCommunityNotificationsInvalidID = errors.New("dismiss-community-notifications: invalid id")
)

type DismissCommunityNotifications struct {
	CommunityID types.HexBytes `json:"communityId"`
}

func (r *DismissCommunityNotifications) Validate() error {
	if len(r.CommunityID) == 0 {
		return ErrDismissCommunityNotificationsInvalidID
	}

	return nil
}
