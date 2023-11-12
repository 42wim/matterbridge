package requests

import (
	"errors"
)

var ErrCreatePublicChatInvalidID = errors.New("create-public-chat: invalid id")

type CreatePublicChat struct {
	ID string `json:"id"`
}

func (c *CreatePublicChat) Validate() error {
	if len(c.ID) == 0 {
		return ErrCreatePublicChatInvalidID
	}

	return nil
}
