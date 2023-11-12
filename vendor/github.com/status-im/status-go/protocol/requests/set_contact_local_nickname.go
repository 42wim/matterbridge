package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrSetContactLocalNicknameInvalidID = errors.New("add-contact: invalid id")

type SetContactLocalNickname struct {
	ID       types.HexBytes `json:"id"`
	Nickname string         `json:"nickname"`
}

func (a *SetContactLocalNickname) Validate() error {
	if len(a.ID) == 0 {
		return ErrSetContactLocalNicknameInvalidID
	}

	return nil
}
