package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrInvalidMuteCommunityParams = errors.New("mute-community: invalid params")

type MuteCommunity struct {
	CommunityID types.HexBytes  `json:"communityId"`
	MutedType   MutingVariation `json:"mutedType"`
}

func (a *MuteCommunity) Validate() error {
	if len(a.CommunityID) == 0 {
		return ErrInvalidMuteCommunityParams
	}

	if a.MutedType < 0 {
		return ErrInvalidMuteCommunityParams
	}

	return nil
}
