package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrCancelRequestToJoinCommunityInvalidID = errors.New("cancel-request-to-join-community: invalid id")

type CancelRequestToJoinCommunity struct {
	ID types.HexBytes `json:"id"`
}

func (j *CancelRequestToJoinCommunity) Validate() error {
	if len(j.ID) == 0 {
		return ErrCancelRequestToJoinCommunityInvalidID
	}

	return nil
}
