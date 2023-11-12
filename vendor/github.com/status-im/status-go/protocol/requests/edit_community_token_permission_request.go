package requests

import (
	"errors"

	"github.com/status-im/status-go/protocol/protobuf"
)

var (
	ErrEditCommunityTokenPermissionInvalidID = errors.New("invalid community token permission id")
)

type EditCommunityTokenPermission struct {
	PermissionID string `json:"permissionId"`
	CreateCommunityTokenPermission
}

func (u *EditCommunityTokenPermission) Validate() error {
	if len(u.PermissionID) == 0 {
		return ErrEditCommunityTokenPermissionInvalidID
	}

	return u.CreateCommunityTokenPermission.Validate()
}

func (u *EditCommunityTokenPermission) ToCommunityTokenPermission() protobuf.CommunityTokenPermission {
	return protobuf.CommunityTokenPermission{
		Id:            u.PermissionID,
		Type:          u.Type,
		TokenCriteria: u.TokenCriteria,
		ChatIds:       u.ChatIds,
		IsPrivate:     u.IsPrivate,
	}
}
