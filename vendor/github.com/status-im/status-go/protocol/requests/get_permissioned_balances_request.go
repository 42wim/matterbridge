package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrGetPermissionedBalancesMissingID = errors.New("GetPermissionedBalances: missing community ID")

type GetPermissionedBalances struct {
	CommunityID types.HexBytes `json:"communityId"`
}

func (r *GetPermissionedBalances) Validate() error {
	if len(r.CommunityID) == 0 {
		return ErrGetPermissionedBalancesMissingID
	}

	return nil
}
