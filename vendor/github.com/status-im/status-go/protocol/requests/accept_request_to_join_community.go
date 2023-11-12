package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrAcceptRequestToJoinCommunityInvalidID = errors.New("accept-request-to-join-community: invalid id")

type AcceptRequestToJoinCommunity struct {
	ID types.HexBytes `json:"id"`
}

func (j *AcceptRequestToJoinCommunity) Validate() error {
	if len(j.ID) == 0 {
		return ErrAcceptRequestToJoinCommunityInvalidID
	}

	return nil
}
