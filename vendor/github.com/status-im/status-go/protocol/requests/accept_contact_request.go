package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrAcceptContactRequestInvalidID = errors.New("accept-contact-request: invalid id")

type AcceptContactRequest struct {
	ID types.HexBytes `json:"id"`
}

func (a *AcceptContactRequest) Validate() error {
	if len(a.ID) == 0 {
		return ErrAcceptContactRequestInvalidID
	}

	return nil
}
