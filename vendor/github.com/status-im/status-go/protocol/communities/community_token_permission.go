package communities

import (
	"reflect"

	"github.com/status-im/status-go/protocol/protobuf"
)

type TokenPermissionState uint8

const (
	TokenPermissionApproved TokenPermissionState = iota
	TokenPermissionAdditionPending
	TokenPermissionUpdatePending
	TokenPermissionRemovalPending
)

type CommunityTokenPermission struct {
	*protobuf.CommunityTokenPermission
	State TokenPermissionState `json:"state,omitempty"`
}

func NewCommunityTokenPermission(base *protobuf.CommunityTokenPermission) *CommunityTokenPermission {
	return &CommunityTokenPermission{
		CommunityTokenPermission: base,
		State:                    TokenPermissionApproved,
	}
}

func (p *CommunityTokenPermission) Equals(other *CommunityTokenPermission) bool {
	if p.Id != other.Id ||
		p.Type != other.Type ||
		len(p.TokenCriteria) != len(other.TokenCriteria) ||
		len(p.ChatIds) != len(other.ChatIds) ||
		p.IsPrivate != other.IsPrivate ||
		p.State != other.State {
		return false
	}

	for i := range p.TokenCriteria {
		if !compareTokenCriteria(p.TokenCriteria[i], other.TokenCriteria[i]) {
			return false
		}
	}

	return reflect.DeepEqual(p.ChatIds, other.ChatIds)
}

func compareTokenCriteria(a, b *protobuf.TokenCriteria) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return a.Type == b.Type &&
		a.Symbol == b.Symbol &&
		a.Name == b.Name &&
		a.Amount == b.Amount &&
		a.EnsPattern == b.EnsPattern &&
		a.Decimals == b.Decimals &&
		reflect.DeepEqual(a.ContractAddresses, b.ContractAddresses) &&
		reflect.DeepEqual(a.TokenIds, b.TokenIds)
}
