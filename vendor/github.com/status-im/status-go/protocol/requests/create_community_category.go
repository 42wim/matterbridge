package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrCreateCommunityCategoryInvalidCommunityID = errors.New("create-community-category: invalid community id")
var ErrCreateCommunityCategoryInvalidName = errors.New("create-community-category: invalid category name")

type CreateCommunityCategory struct {
	CommunityID  types.HexBytes `json:"communityId"`
	CategoryName string         `json:"categoryName"`
	ChatIDs      []string       `json:"chatIds"`
	ThirdPartyID string         `json:"thirdPartyID,omitempty"`
}

func (j *CreateCommunityCategory) Validate() error {
	if len(j.CommunityID) == 0 {
		return ErrCreateCommunityCategoryInvalidCommunityID
	}

	if len(j.CategoryName) == 0 {
		return ErrCreateCommunityCategoryInvalidName
	}

	return nil
}
