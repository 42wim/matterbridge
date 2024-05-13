package bridge

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/contracts"
	"github.com/status-im/status-go/contracts/hop"
	hopBridge "github.com/status-im/status-go/contracts/hop/bridge"
	hopWrapper "github.com/status-im/status-go/contracts/hop/wrapper"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/transactions"
)

const HopLpFeeBps = 4
const HopCanonicalTokenIndex = 0
const HophTokenIndex = 1
const HopMinBonderFeeUsd = 0.25

var HopBondTransferGasLimit = map[uint64]int64{
	1:      165000,
	5:      165000,
	10:     100000000,
	42161:  2500000,
	420:    100000000,
	421613: 2500000,
}
var HopSettlementGasLimitPerTx = map[uint64]int64{
	1:      5141,
	5:      5141,
	10:     8545,
	42161:  19843,
	420:    8545,
	421613: 19843,
}
var HopBonderFeeBps = map[string]map[uint64]int64{
	"USDC": {
		1:      14,
		5:      14,
		10:     14,
		42161:  14,
		420:    14,
		421613: 14,
	},
	"USDT": {
		1:      26,
		10:     26,
		421613: 26,
	},
	"DAI": {
		1:     26,
		10:    26,
		42161: 26,
	},
	"ETH": {
		1:      5,
		5:      5,
		10:     5,
		42161:  5,
		420:    5,
		421613: 5,
	},
	"WBTC": {
		1:     23,
		10:    23,
		42161: 23,
	},
}

type HopTxArgs struct {
	transactions.SendTxArgs
	ChainID   uint64         `json:"chainId"`
	Symbol    string         `json:"symbol"`
	Recipient common.Address `json:"recipient"`
	Amount    *hexutil.Big   `json:"amount"`
	BonderFee *hexutil.Big   `json:"bonderFee"`
}

type HopBridge struct {
	transactor    *transactions.Transactor
	tokenManager  *token.Manager
	contractMaker *contracts.ContractMaker
}

func NewHopBridge(rpcClient *rpc.Client, transactor *transactions.Transactor, tokenManager *token.Manager) *HopBridge {
	return &HopBridge{
		contractMaker: &contracts.ContractMaker{RPCClient: rpcClient},
		transactor:    transactor,
		tokenManager:  tokenManager,
	}
}

func (h *HopBridge) Name() string {
	return "Hop"
}

func (h *HopBridge) Can(from, to *params.Network, token *token.Token, balance *big.Int) (bool, error) {
	if balance.Cmp(big.NewInt(0)) == 0 {
		return false, nil
	}

	if from.ChainID == to.ChainID {
		return false, nil
	}

	fees, ok := HopBonderFeeBps[token.Symbol]
	if !ok {
		return false, nil
	}

	if _, ok := fees[from.ChainID]; !ok {
		return false, nil
	}

	if _, ok := fees[to.ChainID]; !ok {
		return false, nil
	}
	return true, nil
}

func (h *HopBridge) EstimateGas(fromNetwork *params.Network, toNetwork *params.Network, from common.Address, to common.Address, token *token.Token, amountIn *big.Int) (uint64, error) {
	var input []byte
	value := new(big.Int)

	now := time.Now()
	deadline := big.NewInt(now.Unix() + 604800)

	if token.IsNative() {
		value = amountIn
	}

	contractAddress := h.GetContractAddress(fromNetwork, token)
	if contractAddress == nil {
		return 0, errors.New("contract not found")
	}

	ctx := context.Background()

	if fromNetwork.Layer == 1 {
		ABI, err := abi.JSON(strings.NewReader(hopBridge.HopBridgeABI))
		if err != nil {
			return 0, err
		}

		input, err = ABI.Pack("sendToL2",
			big.NewInt(int64(toNetwork.ChainID)),
			to,
			amountIn,
			big.NewInt(0),
			deadline,
			common.HexToAddress("0x0"),
			big.NewInt(0))

		if err != nil {
			return 0, err
		}
	} else {
		ABI, err := abi.JSON(strings.NewReader(hopWrapper.HopWrapperABI))
		if err != nil {
			return 0, err
		}

		input, err = ABI.Pack("swapAndSend",
			big.NewInt(int64(toNetwork.ChainID)),
			to,
			amountIn,
			big.NewInt(0),
			big.NewInt(0),
			deadline,
			big.NewInt(0),
			deadline)

		if err != nil {
			return 0, err
		}
	}

	ethClient, err := h.contractMaker.RPCClient.EthClient(fromNetwork.ChainID)
	if err != nil {
		return 0, err
	}

	if code, err := ethClient.PendingCodeAt(ctx, *contractAddress); err != nil {
		return 0, err
	} else if len(code) == 0 {
		return 0, bind.ErrNoCode
	}

	msg := ethereum.CallMsg{
		From:  from,
		To:    contractAddress,
		Value: value,
		Data:  input,
	}

	estimation, err := ethClient.EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}
	increasedEstimation := float64(estimation) * IncreaseEstimatedGasFactor
	return uint64(increasedEstimation), nil
}

func (h *HopBridge) BuildTx(fromNetwork, toNetwork *params.Network, fromAddress common.Address, toAddress common.Address, token *token.Token, amountIn *big.Int, bonderFee *big.Int) (*ethTypes.Transaction, error) {
	toAddr := types.Address(toAddress)
	sendArgs := &TransactionBridge{
		HopTx: &HopTxArgs{
			SendTxArgs: transactions.SendTxArgs{
				From:  types.Address(fromAddress),
				To:    &toAddr,
				Value: (*hexutil.Big)(amountIn),
				Data:  types.HexBytes("0x0"),
			},
			Symbol:    token.Symbol,
			Recipient: toAddress,
			Amount:    (*hexutil.Big)(amountIn),
			BonderFee: (*hexutil.Big)(bonderFee),
			ChainID:   toNetwork.ChainID,
		},
		ChainID: fromNetwork.ChainID,
	}

	return h.BuildTransaction(sendArgs)
}

func (h *HopBridge) GetContractAddress(network *params.Network, token *token.Token) *common.Address {
	var address common.Address
	if network.Layer == 1 {
		address, _ = hop.L1BridgeContractAddress(network.ChainID, token.Symbol)
	} else {
		address, _ = hop.L2AmmWrapperContractAddress(network.ChainID, token.Symbol)
	}

	return &address
}

func (h *HopBridge) sendOrBuild(sendArgs *TransactionBridge, signerFn bind.SignerFn) (tx *ethTypes.Transaction, err error) {
	fromNetwork := h.contractMaker.RPCClient.NetworkManager.Find(sendArgs.ChainID)
	if fromNetwork == nil {
		return tx, fmt.Errorf("ChainID not supported %d", sendArgs.ChainID)
	}

	nonce, err := h.transactor.NextNonce(h.contractMaker.RPCClient, fromNetwork.ChainID, sendArgs.HopTx.From)
	if err != nil {
		return tx, err
	}

	argNonce := hexutil.Uint64(nonce)
	sendArgs.HopTx.Nonce = &argNonce

	token := h.tokenManager.FindToken(fromNetwork, sendArgs.HopTx.Symbol)
	if fromNetwork.Layer == 1 {
		tx, err = h.sendToL2(sendArgs.ChainID, sendArgs.HopTx, signerFn, token)
		return tx, err
	}
	tx, err = h.swapAndSend(sendArgs.ChainID, sendArgs.HopTx, signerFn, token)
	return tx, err
}

func (h *HopBridge) Send(sendArgs *TransactionBridge, verifiedAccount *account.SelectedExtKey) (hash types.Hash, err error) {
	tx, err := h.sendOrBuild(sendArgs, getSigner(sendArgs.ChainID, sendArgs.HopTx.From, verifiedAccount))
	if err != nil {
		return types.Hash{}, err
	}
	return types.Hash(tx.Hash()), nil
}

func (h *HopBridge) BuildTransaction(sendArgs *TransactionBridge) (*ethTypes.Transaction, error) {
	return h.sendOrBuild(sendArgs, nil)
}

func (h *HopBridge) sendToL2(chainID uint64, hopArgs *HopTxArgs, signerFn bind.SignerFn, token *token.Token) (tx *ethTypes.Transaction, err error) {
	bridge, err := h.contractMaker.NewHopL1Bridge(chainID, hopArgs.Symbol)
	if err != nil {
		return tx, err
	}
	txOpts := hopArgs.ToTransactOpts(signerFn)
	if token.IsNative() {
		txOpts.Value = (*big.Int)(hopArgs.Amount)
	}
	now := time.Now()
	deadline := big.NewInt(now.Unix() + 604800)
	tx, err = bridge.SendToL2(
		txOpts,
		big.NewInt(int64(hopArgs.ChainID)),
		hopArgs.Recipient,
		hopArgs.Amount.ToInt(),
		big.NewInt(0),
		deadline,
		common.HexToAddress("0x0"),
		big.NewInt(0),
	)

	return tx, err
}

func (h *HopBridge) swapAndSend(chainID uint64, hopArgs *HopTxArgs, signerFn bind.SignerFn, token *token.Token) (tx *ethTypes.Transaction, err error) {
	ammWrapper, err := h.contractMaker.NewHopL2AmmWrapper(chainID, hopArgs.Symbol)
	if err != nil {
		return tx, err
	}

	toNetwork := h.contractMaker.RPCClient.NetworkManager.Find(hopArgs.ChainID)
	if toNetwork == nil {
		return tx, err
	}

	txOpts := hopArgs.ToTransactOpts(signerFn)
	if token.IsNative() {
		txOpts.Value = (*big.Int)(hopArgs.Amount)
	}
	now := time.Now()
	deadline := big.NewInt(now.Unix() + 604800)
	amountOutMin := big.NewInt(0)
	destinationDeadline := big.NewInt(now.Unix() + 604800)
	destinationAmountOutMin := big.NewInt(0)

	if toNetwork.Layer == 1 {
		destinationDeadline = big.NewInt(0)
	}

	tx, err = ammWrapper.SwapAndSend(
		txOpts,
		new(big.Int).SetUint64(hopArgs.ChainID),
		hopArgs.Recipient,
		hopArgs.Amount.ToInt(),
		hopArgs.BonderFee.ToInt(),
		amountOutMin,
		deadline,
		destinationAmountOutMin,
		destinationDeadline,
	)

	return tx, err
}

// CalculateBonderFees logics come from: https://docs.hop.exchange/fee-calculation
func (h *HopBridge) CalculateBonderFees(from, to *params.Network, token *token.Token, amountIn *big.Int, nativeTokenPrice, tokenPrice float64, gasPrice *big.Float) (*big.Int, error) {
	amount := new(big.Float).SetInt(amountIn)
	totalFee := big.NewFloat(0)
	destinationTxFee, err := h.getDestinationTxFee(from, to, nativeTokenPrice, tokenPrice, gasPrice)
	if err != nil {
		return nil, err
	}

	bonderFeeRelative, err := h.getBonderFeeRelative(from, to, amount, token)
	if err != nil {
		return nil, err
	}
	if from.Layer != 1 {
		adjustedBonderFee, err := h.calcFromHTokenAmount(to, bonderFeeRelative, token.Symbol)
		if err != nil {
			return nil, err
		}
		adjustedDestinationTxFee, err := h.calcToHTokenAmount(to, destinationTxFee, token.Symbol)
		if err != nil {
			return nil, err
		}

		bonderFeeAbsolute := h.getBonderFeeAbsolute(tokenPrice)
		if adjustedBonderFee.Cmp(bonderFeeAbsolute) == -1 {
			adjustedBonderFee = bonderFeeAbsolute
		}

		totalFee.Add(adjustedBonderFee, adjustedDestinationTxFee)
	}
	res, _ := new(big.Float).Mul(totalFee, big.NewFloat(math.Pow(10, float64(token.Decimals)))).Int(nil)
	return res, nil
}

func (h *HopBridge) CalculateFees(from, to *params.Network, token *token.Token, amountIn *big.Int, nativeTokenPrice, tokenPrice float64, gasPrice *big.Float) (*big.Int, *big.Int, error) {
	bonderFees, err := h.CalculateBonderFees(from, to, token, amountIn, nativeTokenPrice, tokenPrice, gasPrice)
	if err != nil {
		return nil, nil, err
	}
	amountOut, err := h.amountOut(from, to, new(big.Float).SetInt(amountIn), token.Symbol)
	if err != nil {
		return nil, nil, err
	}
	amountOutInt, _ := amountOut.Int(nil)

	return bonderFees, new(big.Int).Add(
		bonderFees,
		new(big.Int).Sub(amountIn, amountOutInt),
	), nil
}

func (h *HopBridge) calcToHTokenAmount(network *params.Network, amount *big.Float, symbol string) (*big.Float, error) {
	if network.Layer == 1 || amount.Cmp(big.NewFloat(0)) == 0 {
		return amount, nil
	}

	contract, err := h.contractMaker.NewHopL2SaddlSwap(network.ChainID, symbol)
	if err != nil {
		return nil, err
	}
	amountInt, _ := amount.Int(nil)
	res, err := contract.CalculateSwap(&bind.CallOpts{Context: context.Background()}, HopCanonicalTokenIndex, HophTokenIndex, amountInt)
	if err != nil {
		return nil, err
	}

	return new(big.Float).SetInt(res), nil
}

func (h *HopBridge) calcFromHTokenAmount(network *params.Network, amount *big.Float, symbol string) (*big.Float, error) {
	if network.Layer == 1 || amount.Cmp(big.NewFloat(0)) == 0 {
		return amount, nil
	}
	contract, err := h.contractMaker.NewHopL2SaddlSwap(network.ChainID, symbol)
	if err != nil {
		return nil, err
	}
	amountInt, _ := amount.Int(nil)
	res, err := contract.CalculateSwap(&bind.CallOpts{Context: context.Background()}, HophTokenIndex, HopCanonicalTokenIndex, amountInt)
	if err != nil {
		return nil, err
	}

	return new(big.Float).SetInt(res), nil
}

func (h *HopBridge) CalculateAmountOut(from, to *params.Network, amountIn *big.Int, symbol string) (*big.Int, error) {
	amountOut, err := h.amountOut(from, to, new(big.Float).SetInt(amountIn), symbol)
	if err != nil {
		return nil, err
	}
	amountOutInt, _ := amountOut.Int(nil)
	return amountOutInt, nil
}

func (h *HopBridge) amountOut(from, to *params.Network, amountIn *big.Float, symbol string) (*big.Float, error) {
	hTokenAmount, err := h.calcToHTokenAmount(from, amountIn, symbol)
	if err != nil {
		return nil, err
	}
	return h.calcFromHTokenAmount(to, hTokenAmount, symbol)
}

func (h *HopBridge) getBonderFeeRelative(from, to *params.Network, amount *big.Float, token *token.Token) (*big.Float, error) {
	if from.Layer != 1 {
		return big.NewFloat(0), nil
	}

	hTokenAmount, err := h.calcToHTokenAmount(from, amount, token.Symbol)
	if err != nil {
		return nil, err
	}
	feeBps := HopBonderFeeBps[token.Symbol][to.ChainID]

	factor := new(big.Float).Mul(hTokenAmount, big.NewFloat(float64(feeBps)))
	return new(big.Float).Quo(
		factor,
		big.NewFloat(10000),
	), nil
}

func (h *HopBridge) getBonderFeeAbsolute(tokenPrice float64) *big.Float {
	return new(big.Float).Quo(big.NewFloat(HopMinBonderFeeUsd), big.NewFloat(tokenPrice))
}

func (h *HopBridge) getDestinationTxFee(from, to *params.Network, nativeTokenPrice, tokenPrice float64, gasPrice *big.Float) (*big.Float, error) {
	if from.Layer != 1 {
		return big.NewFloat(0), nil
	}

	bondTransferGasLimit := HopBondTransferGasLimit[to.ChainID]
	settlementGasLimit := HopSettlementGasLimitPerTx[to.ChainID]
	totalGasLimit := new(big.Int).Add(big.NewInt(bondTransferGasLimit), big.NewInt(settlementGasLimit))

	rate := new(big.Float).Quo(big.NewFloat(nativeTokenPrice), big.NewFloat(tokenPrice))

	txFeeEth := new(big.Float).Mul(gasPrice, new(big.Float).SetInt(totalGasLimit))
	return new(big.Float).Mul(txFeeEth, rate), nil
}
