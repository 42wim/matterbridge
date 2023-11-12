package ethapi

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/internal/ethapi"
)

type LimitedPersonalAPI struct {
	privateAPI *ethapi.PrivateAccountAPI
}

func NewLimitedPersonalAPI(am *accounts.Manager) *LimitedPersonalAPI {
	return &LimitedPersonalAPI{ethapi.NewSubsetOfPrivateAccountAPI(am)}
}

func (s *LimitedPersonalAPI) Sign(ctx context.Context, data hexutil.Bytes, addr common.Address, passwd string) (hexutil.Bytes, error) {
	return s.privateAPI.Sign(ctx, data, addr, passwd)
}

func (s *LimitedPersonalAPI) EcRecover(ctx context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	return s.privateAPI.EcRecover(ctx, data, sig)
}
