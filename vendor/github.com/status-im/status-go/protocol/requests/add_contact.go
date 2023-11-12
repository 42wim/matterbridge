package requests

import (
	"errors"
)

var ErrAddContactInvalidID = errors.New("add-contact: invalid id")

type AddContact struct {
	ID          string `json:"id"`
	Nickname    string `json:"nickname"`
	DisplayName string `json:"displayName"`
	ENSName     string `json:"ensName"`
}

func (a *AddContact) Validate() error {
	if len(a.ID) == 0 {
		return ErrAddContactInvalidID
	}

	return nil
}

func (a *AddContact) HexID() (string, error) {
	return ConvertCompressedToLegacyKey(a.ID)
}
