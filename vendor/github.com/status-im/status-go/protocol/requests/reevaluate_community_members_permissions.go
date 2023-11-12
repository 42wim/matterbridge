package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

type ReevaluateCommunityMembersPermissions struct {
	CommunityID types.HexBytes `json:"communityId"`
}

func (r *ReevaluateCommunityMembersPermissions) Validate() error {
	if r.CommunityID == nil || len(r.CommunityID) == 0 {
		return errors.New("reevaluate community members permissions does not contain communityID")
	}

	return nil
}
