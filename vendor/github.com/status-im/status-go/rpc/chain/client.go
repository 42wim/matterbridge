package chain

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/services/rpcstats"
)

type BatchCallClient interface {
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
}

type ClientInterface interface {
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	CallBlockHashByTransaction(ctx context.Context, blockNumber *big.Int, index uint) (common.Hash, error)
	GetBaseFeeFromBlock(ctx context.Context, blockNumber *big.Int) (string, error)
	NetworkID() uint64
	ToBigInt() *big.Int
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
	GetWalletNotifier() func(chainId uint64, message string)
	SetWalletNotifier(notifier func(chainId uint64, message string))
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BlockNumber(ctx context.Context) (uint64, error)
	SetIsConnected(value bool)
	GetIsConnected() bool
	bind.ContractCaller
	bind.ContractTransactor
	bind.ContractFilterer
}

type ClientWithFallback struct {
	ChainID         uint64
	main            *ethclient.Client
	fallback        *ethclient.Client
	mainLimiter     *RPCLimiter
	fallbackLimiter *RPCLimiter

	mainRPC     *rpc.Client
	fallbackRPC *rpc.Client

	WalletNotifier func(chainId uint64, message string)

	IsConnected     bool
	IsConnectedLock sync.RWMutex
	LastCheckedAt   int64

	circuitBreakerCmdName string
}

// Don't mark connection as failed if we get one of these errors
var propagateErrors = []error{
	vm.ErrOutOfGas,
	vm.ErrCodeStoreOutOfGas,
	vm.ErrDepth,
	vm.ErrInsufficientBalance,
	vm.ErrContractAddressCollision,
	vm.ErrExecutionReverted,
	vm.ErrMaxCodeSizeExceeded,
	vm.ErrInvalidJump,
	vm.ErrWriteProtection,
	vm.ErrReturnDataOutOfBounds,
	vm.ErrGasUintOverflow,
	vm.ErrInvalidCode,
	vm.ErrNonceUintOverflow,

	// Used by balance history to check state
	ethereum.NotFound,
	bind.ErrNoCode,
}

type CommandResult struct {
	res []any
	err error
}

func NewSimpleClient(mainLimiter *RPCLimiter, main *rpc.Client, chainID uint64) *ClientWithFallback {
	circuitBreakerCmdName := fmt.Sprintf("ethClient_%d", chainID)
	hystrix.ConfigureCommand(circuitBreakerCmdName, hystrix.CommandConfig{
		Timeout:               10000,
		MaxConcurrentRequests: 100,
		SleepWindow:           300000,
		ErrorPercentThreshold: 25,
	})

	return &ClientWithFallback{
		ChainID:               chainID,
		main:                  ethclient.NewClient(main),
		fallback:              nil,
		mainLimiter:           mainLimiter,
		fallbackLimiter:       nil,
		mainRPC:               main,
		fallbackRPC:           nil,
		IsConnected:           true,
		LastCheckedAt:         time.Now().Unix(),
		circuitBreakerCmdName: circuitBreakerCmdName,
	}
}

func NewClient(mainLimiter *RPCLimiter, main *rpc.Client, fallbackLimiter *RPCLimiter, fallback *rpc.Client, chainID uint64) *ClientWithFallback {
	circuitBreakerCmdName := fmt.Sprintf("ethClient_%d", chainID)
	hystrix.ConfigureCommand(circuitBreakerCmdName, hystrix.CommandConfig{
		Timeout:               20000,
		MaxConcurrentRequests: 100,
		SleepWindow:           300000,
		ErrorPercentThreshold: 25,
	})

	var fallbackEthClient *ethclient.Client
	if fallback != nil {
		fallbackEthClient = ethclient.NewClient(fallback)
	}
	return &ClientWithFallback{
		ChainID:               chainID,
		main:                  ethclient.NewClient(main),
		fallback:              fallbackEthClient,
		mainLimiter:           mainLimiter,
		fallbackLimiter:       fallbackLimiter,
		mainRPC:               main,
		fallbackRPC:           fallback,
		IsConnected:           true,
		LastCheckedAt:         time.Now().Unix(),
		circuitBreakerCmdName: circuitBreakerCmdName,
	}
}

func (c *ClientWithFallback) Close() {
	c.main.Close()
	if c.fallback != nil {
		c.fallback.Close()
	}
}

func isVMError(err error) bool {
	if strings.HasPrefix(err.Error(), "execution reverted") {
		return true
	}
	if strings.Contains(err.Error(), core.ErrInsufficientFunds.Error()) {
		return true
	}
	for _, vmError := range propagateErrors {
		if err == vmError {
			return true
		}

	}
	return false
}

func isRPSLimitError(err error) bool {
	return strings.Contains(err.Error(), "backoff_seconds")
}

func (c *ClientWithFallback) SetIsConnected(value bool) {
	c.IsConnectedLock.Lock()
	defer c.IsConnectedLock.Unlock()
	c.LastCheckedAt = time.Now().Unix()
	if !value {
		if c.IsConnected {
			if c.WalletNotifier != nil {
				c.WalletNotifier(c.ChainID, "down")
			}
			c.IsConnected = false
		}

	} else {
		if !c.IsConnected {
			c.IsConnected = true
			if c.WalletNotifier != nil {
				c.WalletNotifier(c.ChainID, "up")
			}
		}
	}
}

func (c *ClientWithFallback) GetIsConnected() bool {
	c.IsConnectedLock.RLock()
	defer c.IsConnectedLock.RUnlock()
	return c.IsConnected
}

func (c *ClientWithFallback) makeCall(ctx context.Context, main func() ([]any, error), fallback func() ([]any, error)) ([]any, error) {
	resultChan := make(chan CommandResult, 1)
	c.LastCheckedAt = time.Now().Unix()
	errChan := hystrix.Go(c.circuitBreakerCmdName, func() error {
		err := c.mainLimiter.WaitForRequestsAvailability(1)
		if err != nil {
			return err
		}

		res, err := main()
		if err != nil {
			if isRPSLimitError(err) {
				c.mainLimiter.ReduceLimit()
			} else if isVMError(err) {
				resultChan <- CommandResult{err: err}
				return nil
			}
			return err
		}
		resultChan <- CommandResult{res: res}
		return nil
	}, func(err error) error {
		if c.fallback == nil {
			return err
		}

		err = c.fallbackLimiter.WaitForRequestsAvailability(1)
		if err != nil {
			return err
		}

		res, err := fallback()
		if err != nil {
			if isRPSLimitError(err) {
				c.fallbackLimiter.ReduceLimit()
			} else if isVMError(err) {
				resultChan <- CommandResult{err: err}
				return nil
			}
			return err
		}
		resultChan <- CommandResult{res: res}
		return nil
	})

	select {
	case result := <-resultChan:
		if result.err != nil {
			return nil, result.err
		}
		return result.res, nil
	case err := <-errChan:
		return nil, err
	}
}

func (c *ClientWithFallback) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	rpcstats.CountCall("eth_BlockByHash")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.BlockByHash(ctx, hash); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.BlockByHash(ctx, hash); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*types.Block), nil
}

func (c *ClientWithFallback) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	rpcstats.CountCall("eth_BlockByNumber")
	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.BlockByNumber(ctx, number); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.BlockByNumber(ctx, number); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*types.Block), nil
}

func (c *ClientWithFallback) BlockNumber(ctx context.Context) (uint64, error) {
	rpcstats.CountCall("eth_BlockNumber")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.BlockNumber(ctx); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.BlockNumber(ctx); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return 0, err
	}

	return res[0].(uint64), nil
}

func (c *ClientWithFallback) PeerCount(ctx context.Context) (uint64, error) {
	rpcstats.CountCall("eth_PeerCount")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.PeerCount(ctx); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.PeerCount(ctx); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return 0, err
	}

	return res[0].(uint64), nil
}

func (c *ClientWithFallback) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	rpcstats.CountCall("eth_HeaderByHash")
	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.HeaderByHash(ctx, hash); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.HeaderByHash(ctx, hash); return []any{a}, err },
	)

	if err != nil {
		return nil, err
	}

	return res[0].(*types.Header), nil
}

func (c *ClientWithFallback) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	rpcstats.CountCall("eth_HeaderByNumber")
	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.HeaderByNumber(ctx, number); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.HeaderByNumber(ctx, number); return []any{a}, err },
	)

	if err != nil {
		return nil, err
	}

	return res[0].(*types.Header), nil
}

func (c *ClientWithFallback) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	rpcstats.CountCall("eth_TransactionByHash")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, b, err := c.main.TransactionByHash(ctx, hash); return []any{a, b}, err },
		func() ([]any, error) { a, b, err := c.fallback.TransactionByHash(ctx, hash); return []any{a, b}, err },
	)

	if err != nil {
		return nil, false, err
	}

	return res[0].(*types.Transaction), res[1].(bool), nil
}

func (c *ClientWithFallback) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	rpcstats.CountCall("eth_TransactionSender")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.TransactionSender(ctx, tx, block, index); return []any{a}, err },
		func() ([]any, error) {
			a, err := c.fallback.TransactionSender(ctx, tx, block, index)
			return []any{a}, err
		},
	)

	c.toggleConnectionState(err)

	return res[0].(common.Address), err
}

func (c *ClientWithFallback) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	rpcstats.CountCall("eth_TransactionCount")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.TransactionCount(ctx, blockHash); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.TransactionCount(ctx, blockHash); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return 0, err
	}

	return res[0].(uint), nil
}

func (c *ClientWithFallback) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	rpcstats.CountCall("eth_TransactionInBlock")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) {
			a, err := c.main.TransactionInBlock(ctx, blockHash, index)
			return []any{a}, err
		},
		func() ([]any, error) {
			a, err := c.fallback.TransactionInBlock(ctx, blockHash, index)
			return []any{a}, err
		},
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*types.Transaction), nil
}

func (c *ClientWithFallback) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	rpcstats.CountCall("eth_TransactionReceipt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.TransactionReceipt(ctx, txHash); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.TransactionReceipt(ctx, txHash); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*types.Receipt), nil
}

func (c *ClientWithFallback) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	rpcstats.CountCall("eth_SyncProgress")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.SyncProgress(ctx); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.SyncProgress(ctx); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*ethereum.SyncProgress), nil
}

func (c *ClientWithFallback) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	rpcstats.CountCall("eth_SubscribeNewHead")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.SubscribeNewHead(ctx, ch); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.SubscribeNewHead(ctx, ch); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(ethereum.Subscription), nil
}

func (c *ClientWithFallback) NetworkID() uint64 {
	return c.ChainID
}

func (c *ClientWithFallback) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	rpcstats.CountCall("eth_BalanceAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.BalanceAt(ctx, account, blockNumber); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.BalanceAt(ctx, account, blockNumber); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}

func (c *ClientWithFallback) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	rpcstats.CountCall("eth_StorageAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) {
			a, err := c.main.StorageAt(ctx, account, key, blockNumber)
			return []any{a}, err
		},
		func() ([]any, error) {
			a, err := c.fallback.StorageAt(ctx, account, key, blockNumber)
			return []any{a}, err
		},
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]byte), nil
}

func (c *ClientWithFallback) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	rpcstats.CountCall("eth_CodeAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.CodeAt(ctx, account, blockNumber); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.CodeAt(ctx, account, blockNumber); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]byte), nil
}

func (c *ClientWithFallback) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	rpcstats.CountCall("eth_NonceAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.NonceAt(ctx, account, blockNumber); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.NonceAt(ctx, account, blockNumber); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return 0, err
	}

	return res[0].(uint64), nil
}

func (c *ClientWithFallback) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	rpcstats.CountCall("eth_FilterLogs")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.FilterLogs(ctx, q); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.FilterLogs(ctx, q); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]types.Log), nil
}

func (c *ClientWithFallback) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	rpcstats.CountCall("eth_SubscribeFilterLogs")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.SubscribeFilterLogs(ctx, q, ch); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.SubscribeFilterLogs(ctx, q, ch); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(ethereum.Subscription), nil
}

func (c *ClientWithFallback) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	rpcstats.CountCall("eth_PendingBalanceAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.PendingBalanceAt(ctx, account); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.PendingBalanceAt(ctx, account); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}

func (c *ClientWithFallback) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	rpcstats.CountCall("eth_PendingStorageAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.PendingStorageAt(ctx, account, key); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.PendingStorageAt(ctx, account, key); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]byte), nil
}

func (c *ClientWithFallback) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	rpcstats.CountCall("eth_PendingCodeAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.PendingCodeAt(ctx, account); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.PendingCodeAt(ctx, account); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]byte), nil
}

func (c *ClientWithFallback) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	rpcstats.CountCall("eth_PendingNonceAt")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.PendingNonceAt(ctx, account); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.PendingNonceAt(ctx, account); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return 0, err
	}

	return res[0].(uint64), nil
}

func (c *ClientWithFallback) PendingTransactionCount(ctx context.Context) (uint, error) {
	rpcstats.CountCall("eth_PendingTransactionCount")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.PendingTransactionCount(ctx); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.PendingTransactionCount(ctx); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return 0, err
	}

	return res[0].(uint), nil
}

func (c *ClientWithFallback) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	rpcstats.CountCall("eth_CallContract_" + msg.To.String())

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.CallContract(ctx, msg, blockNumber); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.CallContract(ctx, msg, blockNumber); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]byte), nil
}

func (c *ClientWithFallback) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) ([]byte, error) {
	rpcstats.CountCall("eth_CallContractAtHash")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.CallContractAtHash(ctx, msg, blockHash); return []any{a}, err },
		func() ([]any, error) {
			a, err := c.fallback.CallContractAtHash(ctx, msg, blockHash)
			return []any{a}, err
		},
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]byte), nil
}

func (c *ClientWithFallback) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	rpcstats.CountCall("eth_PendingCallContract")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.PendingCallContract(ctx, msg); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.PendingCallContract(ctx, msg); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].([]byte), nil
}

func (c *ClientWithFallback) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	rpcstats.CountCall("eth_SuggestGasPrice")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.SuggestGasPrice(ctx); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.SuggestGasPrice(ctx); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}

func (c *ClientWithFallback) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	rpcstats.CountCall("eth_SuggestGasTipCap")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.SuggestGasTipCap(ctx); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.SuggestGasTipCap(ctx); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}

func (c *ClientWithFallback) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	rpcstats.CountCall("eth_FeeHistory")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) {
			a, err := c.main.FeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
			return []any{a}, err
		},
		func() ([]any, error) {
			a, err := c.fallback.FeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
			return []any{a}, err
		},
	)

	c.toggleConnectionState(err)

	if err != nil {
		return nil, err
	}

	return res[0].(*ethereum.FeeHistory), nil
}

func (c *ClientWithFallback) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	rpcstats.CountCall("eth_EstimateGas")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) { a, err := c.main.EstimateGas(ctx, msg); return []any{a}, err },
		func() ([]any, error) { a, err := c.fallback.EstimateGas(ctx, msg); return []any{a}, err },
	)

	c.toggleConnectionState(err)

	if err != nil {
		return 0, err
	}

	return res[0].(uint64), nil
}

func (c *ClientWithFallback) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	rpcstats.CountCall("eth_SendTransaction")

	_, err := c.makeCall(
		ctx,
		func() ([]any, error) { return nil, c.main.SendTransaction(ctx, tx) },
		func() ([]any, error) { return nil, c.fallback.SendTransaction(ctx, tx) },
	)
	return err
}

func (c *ClientWithFallback) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	rpcstats.CountCall("eth_CallContext")

	_, err := c.makeCall(
		ctx,
		func() ([]any, error) { return nil, c.mainRPC.CallContext(ctx, result, method, args...) },
		func() ([]any, error) { return nil, c.fallbackRPC.CallContext(ctx, result, method, args...) },
	)
	return err
}

func (c *ClientWithFallback) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	rpcstats.CountCall("eth_BatchCallContext")

	_, err := c.makeCall(
		ctx,
		func() ([]any, error) { return nil, c.mainRPC.BatchCallContext(ctx, b) },
		func() ([]any, error) { return nil, c.fallbackRPC.BatchCallContext(ctx, b) },
	)
	return err
}

func (c *ClientWithFallback) ToBigInt() *big.Int {
	return big.NewInt(int64(c.ChainID))
}

func (c *ClientWithFallback) GetBaseFeeFromBlock(ctx context.Context, blockNumber *big.Int) (string, error) {
	rpcstats.CountCall("eth_GetBaseFeeFromBlock")

	feeHistory, err := c.FeeHistory(ctx, 1, blockNumber, nil)

	if err != nil {
		if err.Error() == "the method eth_feeHistory does not exist/is not available" {
			return "", nil
		}
		return "", err
	}

	var baseGasFee string = ""
	if len(feeHistory.BaseFee) > 0 {
		baseGasFee = feeHistory.BaseFee[0].String()
	}

	return baseGasFee, err
}

// go-ethereum's `Transaction` items drop the blkHash obtained during the RPC call.
// This function preserves the additional data. This is the cheapest way to obtain
// the block hash for a given block number.
func (c *ClientWithFallback) CallBlockHashByTransaction(ctx context.Context, blockNumber *big.Int, index uint) (common.Hash, error) {
	rpcstats.CountCall("eth_FullTransactionByBlockNumberAndIndex")

	res, err := c.makeCall(
		ctx,
		func() ([]any, error) {
			a, err := callBlockHashByTransaction(ctx, c.mainRPC, blockNumber, index)
			return []any{a}, err
		},
		func() ([]any, error) {
			a, err := callBlockHashByTransaction(ctx, c.fallbackRPC, blockNumber, index)
			return []any{a}, err
		},
	)

	c.toggleConnectionState(err)

	if err != nil {
		return common.HexToHash(""), err
	}

	return res[0].(common.Hash), nil
}

func (c *ClientWithFallback) GetWalletNotifier() func(chainId uint64, message string) {
	return c.WalletNotifier
}

func (c *ClientWithFallback) SetWalletNotifier(notifier func(chainId uint64, message string)) {
	c.WalletNotifier = notifier
}

func (c *ClientWithFallback) toggleConnectionState(err error) {
	connected := true
	if err != nil {
		if !isVMError(err) && err != ErrRequestsOverLimit {
			connected = false
		}
	}
	c.SetIsConnected(connected)
}
