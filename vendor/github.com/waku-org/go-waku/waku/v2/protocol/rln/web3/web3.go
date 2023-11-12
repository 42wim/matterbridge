package web3

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/waku-org/go-waku/waku/v2/protocol/rln/contracts"
)

// RegistryContract contains an instance of the RLN Registry contract and its address
type RegistryContract struct {
	*contracts.RLNRegistry
	Address common.Address
}

// RLNContract contains an instance of the RLN contract, its address and the storage index within the registry
// that represents this contract
type RLNContract struct {
	*contracts.RLN
	Address             common.Address
	StorageIndex        uint16
	DeployedBlockNumber uint64
}

// EthClient is an interface for the ethclient.Client, so that we can pass mock client for testing
type EthClient interface {
	bind.ContractBackend
	SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	Close()
}

// Config is a helper struct that contains attributes for interaction with RLN smart contracts
type Config struct {
	configured bool

	ETHClientAddress string
	ETHClient        EthClient
	ChainID          *big.Int
	RegistryContract RegistryContract
	RLNContract      RLNContract
}

// NewConfig creates an instance of web3 Config
func NewConfig(ethClientAddress string, registryAddress common.Address) *Config {
	return &Config{
		ETHClientAddress: ethClientAddress,
		RegistryContract: RegistryContract{
			Address: registryAddress,
		},
	}
}

// BuildConfig returns an instance of Web3Config with all the required elements for interaction with the RLN smart contracts
func BuildConfig(ctx context.Context, ethClientAddress string, registryAddress common.Address) (*Config, error) {
	ethClient, err := ethclient.DialContext(ctx, ethClientAddress)
	if err != nil {
		return nil, err
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	rlnRegistry, err := contracts.NewRLNRegistry(registryAddress, ethClient)
	if err != nil {
		return nil, err
	}

	storageIndex, err := rlnRegistry.UsingStorageIndex(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, err
	}

	rlnContractAddress, err := rlnRegistry.Storages(&bind.CallOpts{Context: ctx}, storageIndex)
	if err != nil {
		return nil, err
	}

	rlnContract, err := contracts.NewRLN(rlnContractAddress, ethClient)
	if err != nil {
		return nil, err
	}

	deploymentBlockNumber, err := rlnContract.DeployedBlockNumber(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, err
	}

	return &Config{
		configured:       true,
		ETHClientAddress: ethClientAddress,
		ETHClient:        ethClient,
		ChainID:          chainID,
		RegistryContract: RegistryContract{
			RLNRegistry: rlnRegistry,
			Address:     registryAddress,
		},
		RLNContract: RLNContract{
			RLN:                 rlnContract,
			Address:             rlnContractAddress,
			StorageIndex:        storageIndex,
			DeployedBlockNumber: uint64(deploymentBlockNumber),
		},
	}, nil
}

// Build sets up the Config object by instantiating the eth client and contracts
func (w *Config) Build(ctx context.Context) error {
	if w.configured {
		return errors.New("already configured")
	}

	if w.ETHClientAddress == "" {
		return errors.New("no eth client address")
	}

	var zeroAddr common.Address
	if w.RegistryContract.Address == zeroAddr {
		return errors.New("no registry contract address")
	}

	newW, err := BuildConfig(ctx, w.ETHClientAddress, w.RegistryContract.Address)
	if err != nil {
		return err
	}

	*w = *newW
	w.configured = true

	return nil
}
