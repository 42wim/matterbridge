package personal

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
)

var (
	// ErrInvalidPersonalSignAccount is returned when the account passed to
	// personal_sign isn't equal to the currently selected account.
	ErrInvalidPersonalSignAccount = errors.New("invalid account as only the selected one can generate a signature")
)

// SignParams required to sign messages
type SignParams struct {
	Data     interface{} `json:"data"`
	Address  string      `json:"account"`
	Password string      `json:"password"`
}

// RecoverParams are for calling `personal_ecRecover`
type RecoverParams struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// PublicAPI represents a set of APIs from the `web3.personal` namespace.
type PublicAPI struct {
	rpcClient  *rpc.Client
	rpcTimeout time.Duration
}

// NewAPI creates an instance of the personal API.
func NewAPI() *PublicAPI {
	return &PublicAPI{
		rpcTimeout: 300 * time.Second,
	}
}

// SetRPC sets RPC params (client and timeout) for the API calls.
func (api *PublicAPI) SetRPC(rpcClient *rpc.Client, timeout time.Duration) {
	api.rpcClient = rpcClient
	api.rpcTimeout = timeout
}

// Recover is an implementation of `personal_ecRecover` or `web3.personal.ecRecover` API
func (api *PublicAPI) Recover(rpcParams RecoverParams) (addr types.Address, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), api.rpcTimeout)
	defer cancel()
	var gethAddr common.Address
	err = api.rpcClient.CallContextIgnoringLocalHandlers(
		ctx,
		&gethAddr,
		api.rpcClient.UpstreamChainID,
		params.PersonalRecoverMethodName,
		rpcParams.Message, rpcParams.Signature)
	addr = types.Address(gethAddr)

	return
}

// Sign is an implementation of `personal_sign` or `web3.personal.sign` API
func (api *PublicAPI) Sign(rpcParams SignParams, verifiedAccount *account.SelectedExtKey) (result types.HexBytes, err error) {
	if !strings.EqualFold(rpcParams.Address, verifiedAccount.Address.Hex()) {
		err = ErrInvalidPersonalSignAccount
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), api.rpcTimeout)
	defer cancel()
	var gethResult hexutil.Bytes
	err = api.rpcClient.CallContextIgnoringLocalHandlers(
		ctx,
		&gethResult,
		api.rpcClient.UpstreamChainID,
		params.PersonalSignMethodName,
		rpcParams.Data, rpcParams.Address, rpcParams.Password)
	result = types.HexBytes(gethResult)

	return
}
