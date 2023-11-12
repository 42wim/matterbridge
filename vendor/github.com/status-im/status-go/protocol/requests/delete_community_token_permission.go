package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrDeleteCommunityTokenPermissionInvalidCommunityID = errors.New("delete community token permission needs a valid community id ")
var ErrDeleteCommunityTokenPermissionInvalidPermissionID = errors.New("invalid token permission id")

type DeleteCommunityTokenPermission struct {
	CommunityID  types.HexBytes `json:"communityId"`
	PermissionID string         `json:"permissionId"`
}

func (r *DeleteCommunityTokenPermission) Validate() error {
	if len(r.CommunityID) == 0 {
		return ErrDeleteCommunityCategoryInvalidCommunityID
	}

	if len(r.PermissionID) == 0 {
		return ErrDeleteCommunityTokenPermissionInvalidPermissionID
	}
	return nil
}
