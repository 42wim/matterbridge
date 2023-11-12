package requests

import (
	"errors"
)

var ErrClearHistoryInvalidID = errors.New("clear-history: invalid id")

type ClearHistory struct {
	ID string `json:"id"`
}

func (c *ClearHistory) Validate() error {
	if len(c.ID) == 0 {
		return ErrClearHistoryInvalidID
	}

	return nil
}
