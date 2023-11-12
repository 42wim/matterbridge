package requests

import (
	"errors"
)

var ErrDeactivateChatInvalidID = errors.New("deactivate-chat: invalid id")

type DeactivateChat struct {
	ID              string `json:"id"`
	PreserveHistory bool   `json:"preserveHistory"`
}

func (j *DeactivateChat) Validate() error {
	if len(j.ID) == 0 {
		return ErrDeactivateChatInvalidID
	}

	return nil
}
