package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrReorderCommunityCategoryInvalidCommunityID = errors.New("reorder-community-category: invalid community id")
var ErrReorderCommunityCategoryInvalidCategoryID = errors.New("reorder-community-category: invalid category id")
var ErrReorderCommunityCategoryInvalidPosition = errors.New("reorder-community-category: invalid position")

type ReorderCommunityCategories struct {
	CommunityID types.HexBytes `json:"communityId"`
	CategoryID  string         `json:"categoryId"`
	Position    int            `json:"position"`
}

func (j *ReorderCommunityCategories) Validate() error {
	if len(j.CommunityID) == 0 {
		return ErrReorderCommunityCategoryInvalidCommunityID
	}

	if len(j.CategoryID) == 0 {
		return ErrReorderCommunityCategoryInvalidCategoryID
	}

	if j.Position < 0 {
		return ErrReorderCommunityCategoryInvalidPosition
	}

	return nil
}
