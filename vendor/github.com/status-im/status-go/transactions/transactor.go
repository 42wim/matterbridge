package transactions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/wallet/bigint"
	wallet_common "github.com/status-im/status-go/services/wallet/common"
)

const (
	// sendTxTimeout defines how many seconds to wait before returning result in sentTransaction().
	sendTxTimeout = 300 * time.Second

	defaultGas = 90000

	ValidSignatureSize = 65
)

// ErrInvalidSignatureSize is returned if a signature is not 65 bytes to avoid panic from go-ethereum
var ErrInvalidSignatureSize = errors.New("signature size must be 65")

type ErrBadNonce struct {
	nonce         uint64
	expectedNonce uint64
}

func (e *ErrBadNonce) Error() string {
	return fmt.Sprintf("bad nonce. expected %d, got %d", e.expectedNonce, e.nonce)
}

// Transactor validates, signs transactions.
// It uses upstream to propagate transactions to the Ethereum network.
type Transactor struct {
	rpcWrapper     *rpcWrapper
	pendingTracker *PendingTxTracker
	sendTxTimeout  time.Duration
	rpcCallTimeout time.Duration
	networkID      uint64
	log            log.Logger
}

// NewTransactor returns a new Manager.
func NewTransactor() *Transactor {
	return &Transactor{
		sendTxTimeout: sendTxTimeout,
		log:           log.New("package", "status-go/transactions.Manager"),
	}
}

// SetPendingTracker sets a pending tracker.
func (t *Transactor) SetPendingTracker(tracker *PendingTxTracker) {
	t.pendingTracker = tracker
}

// SetNetworkID selects a correct network.
func (t *Transactor) SetNetworkID(networkID uint64) {
	t.networkID = networkID
}

func (t *Transactor) NetworkID() uint64 {
	return t.networkID
}

// SetRPC sets RPC params, a client and a timeout
func (t *Transactor) SetRPC(rpcClient *rpc.Client, timeout time.Duration) {
	t.rpcWrapper = newRPCWrapper(rpcClient, rpcClient.UpstreamChainID)
	t.rpcCallTimeout = timeout
}

func (t *Transactor) NextNonce(rpcClient *rpc.Client, chainID uint64, from types.Address) (uint64, error) {
	wrapper := newRPCWrapper(rpcClient, chainID)
	ctx := context.Background()
	nonce, err := wrapper.PendingNonceAt(ctx, common.Address(from))
	if err != nil {
		return 0, err
	}

	// We need to take into consideration all pending transactions in case of Optimism, cause the network returns always
	// the nonce of last executed tx + 1 for the next nonce value.
	if chainID == wallet_common.OptimismMainnet ||
		chainID == wallet_common.OptimismSepolia ||
		chainID == wallet_common.OptimismGoerli {
		if t.pendingTracker != nil {
			countOfPendingTXs, err := t.pendingTracker.GetPendingTxForSuggestedNonce(wallet_common.ChainID(chainID), common.Address(from), nonce)
			if err != nil {
				return 0, err
			}
			return nonce + countOfPendingTXs, nil
		}
	}

	return nonce, err
}

func (t *Transactor) EstimateGas(network *params.Network, from common.Address, to common.Address, value *big.Int, input []byte) (uint64, error) {
	rpcWrapper := newRPCWrapper(t.rpcWrapper.RPCClient, network.ChainID)

	ctx, cancel := context.WithTimeout(context.Background(), t.rpcCallTimeout)
	defer cancel()

	msg := ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  input,
	}

	return rpcWrapper.EstimateGas(ctx, msg)
}

// SendTransaction is an implementation of eth_sendTransaction. It queues the tx to the sign queue.
func (t *Transactor) SendTransaction(sendArgs SendTxArgs, verifiedAccount *account.SelectedExtKey) (hash types.Hash, err error) {
	hash, err = t.validateAndPropagate(t.rpcWrapper, verifiedAccount, sendArgs)
	return
}

func (t *Transactor) SendTransactionWithChainID(chainID uint64, sendArgs SendTxArgs, verifiedAccount *account.SelectedExtKey) (hash types.Hash, err error) {
	wrapper := newRPCWrapper(t.rpcWrapper.RPCClient, chainID)
	hash, err = t.validateAndPropagate(wrapper, verifiedAccount, sendArgs)
	return
}

func (t *Transactor) ValidateAndBuildTransaction(chainID uint64, sendArgs SendTxArgs) (tx *gethtypes.Transaction, err error) {
	wrapper := newRPCWrapper(t.rpcWrapper.RPCClient, chainID)
	tx, err = t.validateAndBuildTransaction(wrapper, sendArgs)
	return
}

func (t *Transactor) AddSignatureToTransaction(chainID uint64, tx *gethtypes.Transaction, sig []byte) (*gethtypes.Transaction, error) {
	if len(sig) != ValidSignatureSize {
		return nil, ErrInvalidSignatureSize
	}

	rpcWrapper := newRPCWrapper(t.rpcWrapper.RPCClient, chainID)
	chID := big.NewInt(int64(rpcWrapper.chainID))

	signer := gethtypes.NewLondonSigner(chID)
	txWithSignature, err := tx.WithSignature(signer, sig)
	if err != nil {
		return nil, err
	}

	return txWithSignature, nil
}

func (t *Transactor) SendRawTransaction(chainID uint64, rawTx string) error {
	rpcWrapper := newRPCWrapper(t.rpcWrapper.RPCClient, chainID)

	ctx, cancel := context.WithTimeout(context.Background(), t.rpcCallTimeout)
	defer cancel()

	return rpcWrapper.SendRawTransaction(ctx, rawTx)
}

func createPendingTransactions(from common.Address, symbol string, chainID uint64, multiTransactionID wallet_common.MultiTransactionIDType, tx *gethtypes.Transaction) (pTx *PendingTransaction) {

	pTx = &PendingTransaction{
		Hash:               tx.Hash(),
		Timestamp:          uint64(time.Now().Unix()),
		Value:              bigint.BigInt{Int: tx.Value()},
		From:               from,
		To:                 *tx.To(),
		Nonce:              tx.Nonce(),
		Data:               string(tx.Data()),
		Type:               WalletTransfer,
		ChainID:            wallet_common.ChainID(chainID),
		MultiTransactionID: multiTransactionID,
		Symbol:             symbol,
		AutoDelete:         new(bool),
	}
	// Transaction downloader will delete pending transaction as soon as it is confirmed
	*pTx.AutoDelete = false
	return
}

func (t *Transactor) sendTransaction(rpcWrapper *rpcWrapper, from common.Address, symbol string,
	multiTransactionID wallet_common.MultiTransactionIDType, tx *gethtypes.Transaction) (hash types.Hash, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.rpcCallTimeout)
	defer cancel()

	if err := rpcWrapper.SendTransaction(ctx, tx); err != nil {
		return hash, err
	}

	if t.pendingTracker != nil {

		tx := createPendingTransactions(from, symbol, rpcWrapper.chainID, multiTransactionID, tx)

		err := t.pendingTracker.StoreAndTrackPendingTx(tx)
		if err != nil {
			return hash, err
		}
	}

	return types.Hash(tx.Hash()), nil
}

func (t *Transactor) SendTransactionWithSignature(from common.Address, symbol string,
	multiTransactionID wallet_common.MultiTransactionIDType, tx *gethtypes.Transaction) (hash types.Hash, err error) {
	rpcWrapper := newRPCWrapper(t.rpcWrapper.RPCClient, tx.ChainId().Uint64())

	return t.sendTransaction(rpcWrapper, from, symbol, multiTransactionID, tx)
}

func (t *Transactor) AddSignatureToTransactionAndSend(chainID uint64, from common.Address, symbol string,
	multiTransactionID wallet_common.MultiTransactionIDType, tx *gethtypes.Transaction, sig []byte) (hash types.Hash, err error) {
	txWithSignature, err := t.AddSignatureToTransaction(chainID, tx, sig)
	if err != nil {
		return hash, err
	}

	return t.SendTransactionWithSignature(from, symbol, multiTransactionID, txWithSignature)
}

// BuildTransactionAndSendWithSignature receive a transaction and a signature, serialize them together and propage it to the network.
// It's different from eth_sendRawTransaction because it receives a signature and not a serialized transaction with signature.
// Since the transactions is already signed, we assume it was validated and used the right nonce.
func (t *Transactor) BuildTransactionAndSendWithSignature(chainID uint64, args SendTxArgs, sig []byte) (hash types.Hash, err error) {
	txWithSignature, err := t.BuildTransactionWithSignature(chainID, args, sig)
	if err != nil {
		return hash, err
	}

	hash, err = t.SendTransactionWithSignature(common.Address(args.From), args.Symbol, args.MultiTransactionID, txWithSignature)
	return hash, err
}

func (t *Transactor) BuildTransactionWithSignature(chainID uint64, args SendTxArgs, sig []byte) (*gethtypes.Transaction, error) {
	if !args.Valid() {
		return nil, ErrInvalidSendTxArgs
	}

	if len(sig) != ValidSignatureSize {
		return nil, ErrInvalidSignatureSize
	}

	tx := t.buildTransaction(args)
	expectedNonce, err := t.NextNonce(t.rpcWrapper.RPCClient, chainID, args.From)
	if err != nil {
		return nil, err
	}

	if tx.Nonce() != expectedNonce {
		return nil, &ErrBadNonce{tx.Nonce(), expectedNonce}
	}

	txWithSignature, err := t.AddSignatureToTransaction(chainID, tx, sig)
	if err != nil {
		return nil, err
	}

	return txWithSignature, nil
}

func (t *Transactor) HashTransaction(args SendTxArgs) (validatedArgs SendTxArgs, hash types.Hash, err error) {
	if !args.Valid() {
		return validatedArgs, hash, ErrInvalidSendTxArgs
	}

	validatedArgs = args

	nonce, err := t.NextNonce(t.rpcWrapper.RPCClient, t.rpcWrapper.chainID, args.From)
	if err != nil {
		return validatedArgs, hash, err
	}

	gasPrice := (*big.Int)(args.GasPrice)
	gasFeeCap := (*big.Int)(args.MaxFeePerGas)
	gasTipCap := (*big.Int)(args.MaxPriorityFeePerGas)
	if args.GasPrice == nil && args.MaxFeePerGas == nil {
		ctx, cancel := context.WithTimeout(context.Background(), t.rpcCallTimeout)
		defer cancel()
		gasPrice, err = t.rpcWrapper.SuggestGasPrice(ctx)
		if err != nil {
			return validatedArgs, hash, err
		}
	}

	chainID := big.NewInt(int64(t.networkID))
	value := (*big.Int)(args.Value)

	var gas uint64
	if args.Gas == nil {
		ctx, cancel := context.WithTimeout(context.Background(), t.rpcCallTimeout)
		defer cancel()

		var (
			gethTo    common.Address
			gethToPtr *common.Address
		)
		if args.To != nil {
			gethTo = common.Address(*args.To)
			gethToPtr = &gethTo
		}
		if args.GasPrice == nil {
			gas, err = t.rpcWrapper.EstimateGas(ctx, ethereum.CallMsg{
				From:      common.Address(args.From),
				To:        gethToPtr,
				GasFeeCap: gasFeeCap,
				GasTipCap: gasTipCap,
				Value:     value,
				Data:      args.GetInput(),
			})
		} else {
			gas, err = t.rpcWrapper.EstimateGas(ctx, ethereum.CallMsg{
				From:     common.Address(args.From),
				To:       gethToPtr,
				GasPrice: gasPrice,
				Value:    value,
				Data:     args.GetInput(),
			})
		}
		if err != nil {
			return validatedArgs, hash, err
		}
		if gas < defaultGas {
			t.log.Info("default gas will be used because estimated is lower", "estimated", gas, "default", defaultGas)
			gas = defaultGas
		}
	} else {
		gas = uint64(*args.Gas)
	}

	newNonce := hexutil.Uint64(nonce)
	newGas := hexutil.Uint64(gas)
	validatedArgs.Nonce = &newNonce
	if args.GasPrice != nil {
		validatedArgs.GasPrice = (*hexutil.Big)(gasPrice)
	} else {
		validatedArgs.MaxPriorityFeePerGas = (*hexutil.Big)(gasTipCap)
		validatedArgs.MaxPriorityFeePerGas = (*hexutil.Big)(gasFeeCap)
	}
	validatedArgs.Gas = &newGas

	tx := t.buildTransaction(validatedArgs)
	hash = types.Hash(gethtypes.NewLondonSigner(chainID).Hash(tx))

	return validatedArgs, hash, nil
}

// make sure that only account which created the tx can complete it
func (t *Transactor) validateAccount(args SendTxArgs, selectedAccount *account.SelectedExtKey) error {
	if selectedAccount == nil {
		return account.ErrNoAccountSelected
	}

	if !bytes.Equal(args.From.Bytes(), selectedAccount.Address.Bytes()) {
		return ErrInvalidTxSender
	}

	return nil
}

func (t *Transactor) validateAndBuildTransaction(rpcWrapper *rpcWrapper, args SendTxArgs) (tx *gethtypes.Transaction, err error) {
	if !args.Valid() {
		return tx, ErrInvalidSendTxArgs
	}

	var nonce uint64
	if args.Nonce != nil {
		nonce = uint64(*args.Nonce)
	} else {
		nonce, err = t.NextNonce(rpcWrapper.RPCClient, rpcWrapper.chainID, args.From)
		if err != nil {
			return tx, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.rpcCallTimeout)
	defer cancel()

	gasPrice := (*big.Int)(args.GasPrice)
	if !args.IsDynamicFeeTx() && args.GasPrice == nil {
		gasPrice, err = rpcWrapper.SuggestGasPrice(ctx)
		if err != nil {
			return tx, err
		}
	}

	value := (*big.Int)(args.Value)
	var gas uint64
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	} else if args.Gas == nil && !args.IsDynamicFeeTx() {
		ctx, cancel = context.WithTimeout(context.Background(), t.rpcCallTimeout)
		defer cancel()

		var (
			gethTo    common.Address
			gethToPtr *common.Address
		)
		if args.To != nil {
			gethTo = common.Address(*args.To)
			gethToPtr = &gethTo
		}
		gas, err = rpcWrapper.EstimateGas(ctx, ethereum.CallMsg{
			From:     common.Address(args.From),
			To:       gethToPtr,
			GasPrice: gasPrice,
			Value:    value,
			Data:     args.GetInput(),
		})
		if err != nil {
			return tx, err
		}
		if gas < defaultGas {
			t.log.Info("default gas will be used because estimated is lower", "estimated", gas, "default", defaultGas)
			gas = defaultGas
		}
	}

	tx = t.buildTransactionWithOverrides(nonce, value, gas, gasPrice, args)
	return tx, nil
}

func (t *Transactor) validateAndPropagate(rpcWrapper *rpcWrapper, selectedAccount *account.SelectedExtKey, args SendTxArgs) (hash types.Hash, err error) {
	if err = t.validateAccount(args, selectedAccount); err != nil {
		return hash, err
	}

	tx, err := t.validateAndBuildTransaction(rpcWrapper, args)
	if err != nil {
		return hash, err
	}

	chainID := big.NewInt(int64(rpcWrapper.chainID))
	signedTx, err := gethtypes.SignTx(tx, gethtypes.NewLondonSigner(chainID), selectedAccount.AccountKey.PrivateKey)
	if err != nil {
		return hash, err
	}

	return t.sendTransaction(rpcWrapper, common.Address(args.From), args.Symbol, args.MultiTransactionID, signedTx)
}

func (t *Transactor) buildTransaction(args SendTxArgs) *gethtypes.Transaction {
	var (
		nonce    uint64
		value    *big.Int
		gas      uint64
		gasPrice *big.Int
	)
	if args.Nonce != nil {
		nonce = uint64(*args.Nonce)
	}
	if args.Value != nil {
		value = (*big.Int)(args.Value)
	}
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if args.GasPrice != nil {
		gasPrice = (*big.Int)(args.GasPrice)
	}

	return t.buildTransactionWithOverrides(nonce, value, gas, gasPrice, args)
}

func (t *Transactor) buildTransactionWithOverrides(nonce uint64, value *big.Int, gas uint64, gasPrice *big.Int, args SendTxArgs) *gethtypes.Transaction {
	var tx *gethtypes.Transaction

	if args.To != nil {
		to := common.Address(*args.To)
		var txData gethtypes.TxData

		if args.IsDynamicFeeTx() {
			gasTipCap := (*big.Int)(args.MaxPriorityFeePerGas)
			gasFeeCap := (*big.Int)(args.MaxFeePerGas)

			txData = &gethtypes.DynamicFeeTx{
				Nonce:     nonce,
				Gas:       gas,
				GasTipCap: gasTipCap,
				GasFeeCap: gasFeeCap,
				To:        &to,
				Value:     value,
				Data:      args.GetInput(),
			}
		} else {
			txData = &gethtypes.LegacyTx{
				Nonce:    nonce,
				GasPrice: gasPrice,
				Gas:      gas,
				To:       &to,
				Value:    value,
				Data:     args.GetInput(),
			}
		}
		tx = gethtypes.NewTx(txData)
		t.logNewTx(args, gas, gasPrice, value)
	} else {
		if args.IsDynamicFeeTx() {
			gasTipCap := (*big.Int)(args.MaxPriorityFeePerGas)
			gasFeeCap := (*big.Int)(args.MaxFeePerGas)

			txData := &gethtypes.DynamicFeeTx{
				Nonce:     nonce,
				Value:     value,
				Gas:       gas,
				GasTipCap: gasTipCap,
				GasFeeCap: gasFeeCap,
				Data:      args.GetInput(),
			}
			tx = gethtypes.NewTx(txData)
		} else {
			tx = gethtypes.NewContractCreation(nonce, value, gas, gasPrice, args.GetInput())
		}
		t.logNewContract(args, gas, gasPrice, value, nonce)
	}

	return tx
}

func (t *Transactor) logNewTx(args SendTxArgs, gas uint64, gasPrice *big.Int, value *big.Int) {
	t.log.Info("New transaction",
		"From", args.From,
		"To", *args.To,
		"Gas", gas,
		"GasPrice", gasPrice,
		"Value", value,
	)
}

func (t *Transactor) logNewContract(args SendTxArgs, gas uint64, gasPrice *big.Int, value *big.Int, nonce uint64) {
	t.log.Info("New contract",
		"From", args.From,
		"Gas", gas,
		"GasPrice", gasPrice,
		"Value", value,
		"Contract address", crypto.CreateAddress(args.From, nonce),
	)
}
