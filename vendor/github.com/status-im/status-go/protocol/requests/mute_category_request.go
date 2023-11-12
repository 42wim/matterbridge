package requests

import (
	"errors"
)

var ErrInvalidMuteCategoryParams = errors.New("mute-category: invalid params")

type MuteCategory struct {
	CommunityID string
	CategoryID  string
	MutedType   MutingVariation
}

func (a *MuteCategory) Validate() error {
	if len(a.CommunityID) == 0 || len(a.CategoryID) == 0 {
		return ErrInvalidMuteCategoryParams
	}

	if a.MutedType < 0 {
		return ErrInvalidMuteCategoryParams
	}

	return nil
}
