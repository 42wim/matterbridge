package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

var ErrRemoveRoleFromMemberInvalidCommunityID = errors.New("remove-role-from-member: invalid community id")
var ErrRemoveRoleFromMemberInvalidUser = errors.New("remove-role-from-member: invalid user id")
var ErrRemoveRoleFromMemberInvalidRole = errors.New("remove-role-from-member: invalid role")

type RemoveRoleFromMember struct {
	CommunityID types.HexBytes                 `json:"communityId"`
	User        types.HexBytes                 `json:"user"`
	Role        protobuf.CommunityMember_Roles `json:"role"`
}

func (r *RemoveRoleFromMember) Validate() error {
	if len(r.CommunityID) == 0 {
		return ErrRemoveRoleFromMemberInvalidCommunityID
	}

	if len(r.User) == 0 {
		return ErrRemoveRoleFromMemberInvalidUser
	}

	if r.Role == protobuf.CommunityMember_ROLE_NONE {
		return ErrRemoveRoleFromMemberInvalidRole
	}

	return nil
}
