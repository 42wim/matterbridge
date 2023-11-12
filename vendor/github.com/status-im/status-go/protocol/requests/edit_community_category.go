package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrEditCommunityCategoryInvalidCommunityID = errors.New("edit-community-category: invalid community id")
var ErrEditCommunityCategoryInvalidCategoryID = errors.New("edit-community-category: invalid category id")
var ErrEditCommunityCategoryInvalidName = errors.New("edit-community-category: invalid category name")

type EditCommunityCategory struct {
	CommunityID  types.HexBytes `json:"communityId"`
	CategoryID   string         `json:"categoryId"`
	CategoryName string         `json:"categoryName"`
	ChatIDs      []string       `json:"chatIds"`
}

func (j *EditCommunityCategory) Validate() error {
	if len(j.CommunityID) == 0 {
		return ErrEditCommunityCategoryInvalidCommunityID
	}

	if len(j.CategoryID) == 0 {
		return ErrEditCommunityCategoryInvalidCategoryID
	}

	if len(j.CategoryName) == 0 {
		return ErrEditCommunityCategoryInvalidName
	}

	return nil
}
