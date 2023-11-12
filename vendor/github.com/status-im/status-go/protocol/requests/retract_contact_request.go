package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrRetractContactRequestInvalidContactID = errors.New("retract-contact-request: invalid id")

type RetractContactRequest struct {
	ID types.HexBytes `json:"id"`
}

func (a *RetractContactRequest) Validate() error {
	if len(a.ID) == 0 {
		return ErrRetractContactRequestInvalidContactID
	}

	return nil
}
