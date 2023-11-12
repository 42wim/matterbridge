package communitytokens

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/contracts/community-tokens/ownertoken"
	communityownertokenregistry "github.com/status-im/status-go/contracts/community-tokens/registry"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/utils"
	wcommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/transactions"
)

type ServiceInterface interface {
	GetCollectibleContractData(chainID uint64, contractAddress string) (*CollectibleContractData, error)
	SetSignerPubKey(ctx context.Context, chainID uint64, contractAddress string, txArgs transactions.SendTxArgs, password string, newSignerPubKey string) (string, error)
	GetAssetContractData(chainID uint64, contractAddress string) (*AssetContractData, error)
	SafeGetSignerPubKey(ctx context.Context, chainID uint64, communityID string) (string, error)
	DeploymentSignatureDigest(chainID uint64, addressFrom string, communityID string) ([]byte, error)
}

// Collectibles service
type Service struct {
	manager         *Manager
	accountsManager *account.GethManager
	pendingTracker  *transactions.PendingTxTracker
	config          *params.NodeConfig
	db              *Database
}

// Returns a new Collectibles Service.
func NewService(rpcClient *rpc.Client, accountsManager *account.GethManager, pendingTracker *transactions.PendingTxTracker, config *params.NodeConfig, appDb *sql.DB) *Service {
	return &Service{
		manager:         &Manager{rpcClient: rpcClient},
		accountsManager: accountsManager,
		pendingTracker:  pendingTracker,
		config:          config,
		db:              NewCommunityTokensDatabase(appDb),
	}
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []ethRpc.API {
	return []ethRpc.API{
		{
			Namespace: "communitytokens",
			Version:   "0.1.0",
			Service:   NewAPI(s),
			Public:    true,
		},
	}
}

// Start is run when a service is started.
func (s *Service) Start() error {
	return nil
}

// Stop is run when a service is stopped.
func (s *Service) Stop() error {
	return nil
}

func (s *Service) NewCommunityOwnerTokenRegistryInstance(chainID uint64, contractAddress string) (*communityownertokenregistry.CommunityOwnerTokenRegistry, error) {
	backend, err := s.manager.rpcClient.EthClient(chainID)
	if err != nil {
		return nil, err
	}
	return communityownertokenregistry.NewCommunityOwnerTokenRegistry(common.HexToAddress(contractAddress), backend)
}

func (s *Service) NewOwnerTokenInstance(chainID uint64, contractAddress string) (*ownertoken.OwnerToken, error) {

	backend, err := s.manager.rpcClient.EthClient(chainID)
	if err != nil {
		return nil, err
	}
	return ownertoken.NewOwnerToken(common.HexToAddress(contractAddress), backend)

}

func (s *Service) GetSignerPubKey(ctx context.Context, chainID uint64, contractAddress string) (string, error) {

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	contractInst, err := s.NewOwnerTokenInstance(chainID, contractAddress)
	if err != nil {
		return "", err
	}
	signerPubKey, err := contractInst.SignerPublicKey(callOpts)
	if err != nil {
		return "", err
	}

	return types.ToHex(signerPubKey), nil
}

func (s *Service) SafeGetSignerPubKey(ctx context.Context, chainID uint64, communityID string) (string, error) {
	// 1. Get Owner Token contract address from deployer contract - SafeGetOwnerTokenAddress()
	ownerTokenAddr, err := s.SafeGetOwnerTokenAddress(ctx, chainID, communityID)
	if err != nil {
		return "", err
	}
	// 2. Get Signer from owner token contract - GetSignerPubKey()
	return s.GetSignerPubKey(ctx, chainID, ownerTokenAddr)
}

func (s *Service) SafeGetOwnerTokenAddress(ctx context.Context, chainID uint64, communityID string) (string, error) {
	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	deployerContractInst, err := s.manager.NewCommunityTokenDeployerInstance(chainID)
	if err != nil {
		return "", err
	}
	registryAddr, err := deployerContractInst.DeploymentRegistry(callOpts)
	if err != nil {
		return "", err
	}
	registryContractInst, err := s.NewCommunityOwnerTokenRegistryInstance(chainID, registryAddr.Hex())
	if err != nil {
		return "", err
	}
	communityEthAddress, err := convert33BytesPubKeyToEthAddress(communityID)
	if err != nil {
		return "", err
	}
	ownerTokenAddress, err := registryContractInst.GetEntry(callOpts, communityEthAddress)

	return ownerTokenAddress.Hex(), err
}

func (s *Service) GetCollectibleContractData(chainID uint64, contractAddress string) (*CollectibleContractData, error) {
	return s.manager.GetCollectibleContractData(chainID, contractAddress)
}

func (s *Service) GetAssetContractData(chainID uint64, contractAddress string) (*AssetContractData, error) {
	return s.manager.GetAssetContractData(chainID, contractAddress)
}

func (s *Service) DeploymentSignatureDigest(chainID uint64, addressFrom string, communityID string) ([]byte, error) {
	return s.manager.DeploymentSignatureDigest(chainID, addressFrom, communityID)
}

func (s *Service) SetSignerPubKey(ctx context.Context, chainID uint64, contractAddress string, txArgs transactions.SendTxArgs, password string, newSignerPubKey string) (string, error) {

	if len(newSignerPubKey) <= 0 {
		return "", fmt.Errorf("signerPubKey is empty")
	}

	transactOpts := txArgs.ToTransactOpts(utils.GetSigner(chainID, s.accountsManager, s.config.KeyStoreDir, txArgs.From, password))

	contractInst, err := s.NewOwnerTokenInstance(chainID, contractAddress)
	if err != nil {
		return "", err
	}

	tx, err := contractInst.SetSignerPublicKey(transactOpts, common.FromHex(newSignerPubKey))
	if err != nil {
		return "", err
	}

	err = s.pendingTracker.TrackPendingTransaction(
		wcommon.ChainID(chainID),
		tx.Hash(),
		common.Address(txArgs.From),
		transactions.SetSignerPublicKey,
		transactions.AutoDelete,
	)
	if err != nil {
		log.Error("TrackPendingTransaction error", "error", err)
		return "", err
	}

	return tx.Hash().Hex(), nil
}
