package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrDeclineRequestToJoinCommunityInvalidID = errors.New("accept-request-to-join-community: invalid id")

type DeclineRequestToJoinCommunity struct {
	ID types.HexBytes `json:"id"`
}

func (j *DeclineRequestToJoinCommunity) Validate() error {
	if len(j.ID) == 0 {
		return ErrDeclineRequestToJoinCommunityInvalidID
	}

	return nil
}
