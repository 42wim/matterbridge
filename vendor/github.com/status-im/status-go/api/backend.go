package api

import (
	signercore "github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/services/personal"
	"github.com/status-im/status-go/services/typeddata"
	"github.com/status-im/status-go/transactions"
)

// StatusBackend defines the contract for the Status.im service
type StatusBackend interface {
	// IsNodeRunning() bool                       // NOTE: Only used in tests
	StartNode(config *params.NodeConfig) error // NOTE: Only used in canary
	StartNodeWithKey(acc multiaccounts.Account, password string, keyHex string, conf *params.NodeConfig) error
	StartNodeWithAccount(acc multiaccounts.Account, password string, conf *params.NodeConfig) error
	StartNodeWithAccountAndInitialConfig(account multiaccounts.Account, password string, settings settings.Settings, conf *params.NodeConfig, subaccs []*accounts.Account) error
	StopNode() error
	// RestartNode() error // NOTE: Only used in tests

	GetNodeConfig() (*params.NodeConfig, error)
	UpdateRootDataDir(datadir string)

	// SelectAccount(loginParams account.LoginParams) error
	OpenAccounts() error
	GetAccounts() ([]multiaccounts.Account, error)
	LocalPairingStarted() error
	// SaveAccount(account multiaccounts.Account) error
	SaveAccountAndStartNodeWithKey(acc multiaccounts.Account, password string, settings settings.Settings, conf *params.NodeConfig, subaccs []*accounts.Account, keyHex string) error
	Recover(rpcParams personal.RecoverParams) (types.Address, error)
	Logout() error

	CallPrivateRPC(inputJSON string) (string, error)
	CallRPC(inputJSON string) (string, error)
	HashTransaction(sendArgs transactions.SendTxArgs) (transactions.SendTxArgs, types.Hash, error)
	HashTypedData(typed typeddata.TypedData) (types.Hash, error)
	HashTypedDataV4(typed signercore.TypedData) (types.Hash, error)
	ResetChainData() error
	SendTransaction(sendArgs transactions.SendTxArgs, password string) (hash types.Hash, err error)
	SendTransactionWithChainID(chainID uint64, sendArgs transactions.SendTxArgs, password string) (hash types.Hash, err error)
	SendTransactionWithSignature(sendArgs transactions.SendTxArgs, sig []byte) (hash types.Hash, err error)
	SignHash(hexEncodedHash string) (string, error)
	SignMessage(rpcParams personal.SignParams) (types.HexBytes, error)
	SignTypedData(typed typeddata.TypedData, address string, password string) (types.HexBytes, error)
	SignTypedDataV4(typed signercore.TypedData, address string, password string) (types.HexBytes, error)

	ConnectionChange(typ string, expensive bool)
	AppStateChange(state string)

	ExtractGroupMembershipSignatures(signaturePairs [][2]string) ([]string, error)
	SignGroupMembership(content string) (string, error)
}
