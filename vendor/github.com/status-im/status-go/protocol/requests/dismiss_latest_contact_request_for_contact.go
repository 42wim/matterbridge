package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrDismissLatestContactRequestForContactInvalidID = errors.New("dismiss-latest-contact-request-for-contact: invalid id")

type DismissLatestContactRequestForContact struct {
	ID types.HexBytes `json:"id"`
}

func (a *DismissLatestContactRequestForContact) Validate() error {
	if len(a.ID) == 0 {
		return ErrDismissLatestContactRequestForContactInvalidID
	}

	return nil
}
