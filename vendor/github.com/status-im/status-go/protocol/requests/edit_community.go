package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var (
	ErrEditCommunityInvalidID = errors.New("edit-community: invalid id")
)

type EditCommunity struct {
	CommunityID types.HexBytes
	CreateCommunity
}

func (u *EditCommunity) Validate() error {
	if len(u.CommunityID) == 0 {
		return ErrEditCommunityInvalidID
	}

	return u.CreateCommunity.Validate()
}
