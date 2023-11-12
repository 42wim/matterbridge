package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrDeclineContactRequestInvalidID = errors.New("decline-contact-request: invalid id")

type DeclineContactRequest struct {
	ID types.HexBytes `json:"id"`
}

func (a *DeclineContactRequest) Validate() error {
	if len(a.ID) == 0 {
		return ErrDeclineContactRequestInvalidID
	}

	return nil
}
