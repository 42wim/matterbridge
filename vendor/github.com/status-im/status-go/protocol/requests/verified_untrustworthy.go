package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrVerifiedUntrustworthyInvalidID = errors.New("verified-untrustworthy: invalid id")

type VerifiedUntrustworthy struct {
	ID types.HexBytes `json:"id"`
}

func (a *VerifiedUntrustworthy) Validate() error {
	if len(a.ID) == 0 {
		return ErrVerifiedUntrustworthyInvalidID
	}

	return nil
}
