package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrVerifiedTrustedInvalidID = errors.New("verified-trusted: invalid id")

type VerifiedTrusted struct {
	ID types.HexBytes `json:"id"`
}

func (a *VerifiedTrusted) Validate() error {
	if len(a.ID) == 0 {
		return ErrVerifiedTrustedInvalidID
	}

	return nil
}
