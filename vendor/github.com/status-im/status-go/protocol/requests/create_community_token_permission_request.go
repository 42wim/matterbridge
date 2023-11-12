package requests

import (
	"errors"
	"strconv"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

const maxTokenCriteriaPerPermission = 5

var (
	ErrCreateCommunityTokenPermissionInvalidCommunityID    = errors.New("create community token permission needs a valid community id")
	ErrCreateCommunityTokenPermissionTooManyTokenCriteria  = errors.New("too many token criteria")
	ErrCreateCommunityTokenPermissionInvalidPermissionType = errors.New("invalid community token permission type")
	ErrCreateCommunityTokenPermissionInvalidTokenCriteria  = errors.New("invalid community permission token criteria data")
)

type CreateCommunityTokenPermission struct {
	CommunityID   types.HexBytes                         `json:"communityId"`
	Type          protobuf.CommunityTokenPermission_Type `json:"type"`
	TokenCriteria []*protobuf.TokenCriteria              `json:"tokenCriteria"`
	IsPrivate     bool                                   `json:"isPrivate"`
	ChatIds       []string                               `json:"chat_ids"`
}

func (p *CreateCommunityTokenPermission) Validate() error {
	if len(p.CommunityID) == 0 {
		return ErrCreateCommunityTokenPermissionInvalidCommunityID
	}

	if len(p.TokenCriteria) > maxTokenCriteriaPerPermission {
		return ErrCreateCommunityTokenPermissionTooManyTokenCriteria
	}

	if p.Type == protobuf.CommunityTokenPermission_UNKNOWN_TOKEN_PERMISSION {
		return ErrCreateCommunityTokenPermissionInvalidPermissionType
	}

	for _, c := range p.TokenCriteria {
		if c.EnsPattern == "" && len(c.ContractAddresses) == 0 {
			return ErrCreateCommunityTokenPermissionInvalidTokenCriteria
		}

		floatAmount, _ := strconv.ParseFloat(c.Amount, 32)
		if len(c.ContractAddresses) > 0 && floatAmount == 0 {
			return ErrCreateCommunityTokenPermissionInvalidTokenCriteria
		}
	}

	return nil
}

func (p *CreateCommunityTokenPermission) ToCommunityTokenPermission() protobuf.CommunityTokenPermission {
	return protobuf.CommunityTokenPermission{
		Type:          p.Type,
		TokenCriteria: p.TokenCriteria,
		IsPrivate:     p.IsPrivate,
		ChatIds:       p.ChatIds,
	}
}
