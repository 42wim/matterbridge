package requests

import (
	"errors"
)

var ErrFetchMessagesInvalidID = errors.New("fetch-messages: invalid id")

type FetchMessages struct {
	ID string `json:"id"`
}

func (c *FetchMessages) Validate() error {
	if len(c.ID) == 0 {
		return ErrFetchMessagesInvalidID
	}

	return nil
}
