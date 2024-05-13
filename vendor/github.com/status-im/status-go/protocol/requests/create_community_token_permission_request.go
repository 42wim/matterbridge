package requests

import (
	"errors"
	"math"
	"math/big"

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

		var amountBig = new(big.Int)
		amountBig.SetString(c.AmountInWei, 10)
		if len(c.ContractAddresses) > 0 && amountBig.Cmp(big.NewInt(0)) == 0 {
			return ErrCreateCommunityTokenPermissionInvalidTokenCriteria
		}
	}

	return nil
}

func (p *CreateCommunityTokenPermission) FillDeprecatedAmount() {

	computeErc20AmountFunc := func(amountInWeis string, decimals uint64) string {
		bigfloat := new(big.Float)
		bigfloat.SetString(amountInWeis)
		multiplier := big.NewFloat(math.Pow(10, float64(decimals)))
		bigfloat.Quo(bigfloat, multiplier)
		return bigfloat.String()
	}

	for _, criteria := range p.TokenCriteria {
		if criteria.AmountInWei == "" {
			continue
		}
		// fill Amount to keep backward compatibility
		// Amount format (deprecated): "0.123"
		// AmountInWei format: "123000..000"
		if criteria.Type == protobuf.CommunityTokenType_ERC20 {
			criteria.Amount = computeErc20AmountFunc(criteria.AmountInWei, criteria.Decimals)
		} else {
			criteria.Amount = criteria.AmountInWei
		}

	}
}

func (p *CreateCommunityTokenPermission) ToCommunityTokenPermission() protobuf.CommunityTokenPermission {
	return protobuf.CommunityTokenPermission{
		Type:          p.Type,
		TokenCriteria: p.TokenCriteria,
		IsPrivate:     p.IsPrivate,
		ChatIds:       p.ChatIds,
	}
}
