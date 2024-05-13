package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

var ErrDeleteCommunityMemberMessagesInvalidCommunityID = errors.New("delete-community-member-messages: invalid community id")
var ErrDeleteCommunityMemberMessagesInvalidMemberID = errors.New("delete-community-member-messages: invalid member id")
var ErrDeleteCommunityMemberMessagesInvalidData = errors.New("delete-community-member-messages: invalid data")
var ErrDeleteCommunityMemberMessagesInvalidDeleteAll = errors.New("delete-community-member-messages: invalid delete all setup")
var ErrDeleteCommunityMemberMessagesInvalidDeleteMessagesByID = errors.New("delete-community-member-messages: invalid delete messages by ID setups")
var ErrDeleteCommunityMemberMessagesInvalidMsgID = errors.New("delete-community-member-messages: invalid messages Id")
var ErrDeleteCommunityMemberMessagesInvalidMsgChatID = errors.New("delete-community-member-messages: invalid messages chatId")

type DeleteCommunityMemberMessages struct {
	CommunityID  types.HexBytes                           `json:"communityId"`
	MemberPubKey string                                   `json:"memberPubKey"`
	Messages     []*protobuf.DeleteCommunityMemberMessage `json:"messages"`
	DeleteAll    bool                                     `json:"deleteAll"`
}

func (d *DeleteCommunityMemberMessages) Validate() error {
	if len(d.CommunityID) == 0 {
		return ErrDeleteCommunityMemberMessagesInvalidCommunityID
	}

	if len(d.MemberPubKey) == 0 {
		return ErrDeleteCommunityMemberMessagesInvalidMemberID
	}

	if d.Messages != nil && len(d.Messages) > 0 && d.DeleteAll {
		return ErrDeleteCommunityMemberMessagesInvalidDeleteAll
	}

	if (d.Messages == nil || (d.Messages != nil && len(d.Messages) == 0)) && !d.DeleteAll {
		return ErrDeleteCommunityMemberMessagesInvalidDeleteMessagesByID
	}

	if d.Messages != nil {
		for _, message := range d.Messages {
			if len(message.Id) == 0 {
				return ErrDeleteCommunityMemberMessagesInvalidMsgID
			}
			if len(message.ChatId) == 0 {
				return ErrDeleteCommunityMemberMessagesInvalidMsgChatID
			}
		}
	}

	return nil
}
