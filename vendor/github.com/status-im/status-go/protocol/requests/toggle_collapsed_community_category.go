package requests

import (
	"errors"
)

var ErrToggleCollapsedCommunityCategoryInvalidCommunityID = errors.New("toggle-collapsed-community-category: invalid community id")
var ErrToggleCollapsedCommunityCategoryInvalidName = errors.New("toggle-collapsed-community-category: invalid category name")

type ToggleCollapsedCommunityCategory struct {
	CommunityID string `json:"communityId"`
	CategoryID  string `json:"categoryId"`
	Collapsed   bool   `json:"collapsed"`
}

func (t *ToggleCollapsedCommunityCategory) Validate() error {
	if len(t.CommunityID) == 0 {
		return ErrToggleCollapsedCommunityCategoryInvalidCommunityID
	}

	if len(t.CategoryID) == 0 {
		return ErrToggleCollapsedCommunityCategoryInvalidName
	}

	return nil
}
