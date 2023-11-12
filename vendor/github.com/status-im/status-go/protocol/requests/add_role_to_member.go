package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

var ErrAddRoleToMemberInvalidCommunityID = errors.New("add-role-to-member: invalid community id")
var ErrAddRoleToMemberInvalidUser = errors.New("add-role-to-member: invalid user id")
var ErrAddRoleToMemberInvalidRole = errors.New("add-role-to-member: invalid role")

type AddRoleToMember struct {
	CommunityID types.HexBytes                 `json:"communityId"`
	User        types.HexBytes                 `json:"user"`
	Role        protobuf.CommunityMember_Roles `json:"role"`
}

func (a *AddRoleToMember) Validate() error {
	if len(a.CommunityID) == 0 {
		return ErrAddRoleToMemberInvalidCommunityID
	}

	if len(a.User) == 0 {
		return ErrAddRoleToMemberInvalidUser
	}

	if a.Role == protobuf.CommunityMember_ROLE_NONE {
		return ErrAddRoleToMemberInvalidRole
	}

	return nil
}
