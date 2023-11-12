package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrReorderCommunityChatInvalidCommunityID = errors.New("reorder-community-chat: invalid community id")
var ErrReorderCommunityChatInvalidChatID = errors.New("reorder-community-chat: invalid chat id")
var ErrReorderCommunityChatInvalidPosition = errors.New("reorder-community-chat: invalid position")

type ReorderCommunityChat struct {
	CommunityID types.HexBytes `json:"communityId"`
	CategoryID  string         `json:"categoryId"`
	ChatID      string         `json:"chatId"`
	Position    int            `json:"position"`
}

func (j *ReorderCommunityChat) Validate() error {
	if len(j.CommunityID) == 0 {
		return ErrReorderCommunityChatInvalidCommunityID
	}

	if len(j.ChatID) == 0 {
		return ErrReorderCommunityChatInvalidChatID
	}

	if j.Position < 0 {
		return ErrReorderCommunityCategoryInvalidPosition
	}

	return nil
}
