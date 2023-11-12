package requests

import (
	"errors"
)

var ErrInvalidMuteChatParams = errors.New("mute-chat: invalid params")

type MutingVariation int

type MuteChat struct {
	ChatID    string
	MutedType MutingVariation
}

func (a *MuteChat) Validate() error {
	if len(a.ChatID) == 0 {
		return ErrInvalidMuteChatParams
	}

	if a.MutedType < 0 {
		return ErrInvalidMuteChatParams
	}

	return nil
}
