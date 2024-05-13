package bridge

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/contracts/ierc20"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/transactions"
)

type TransferBridge struct {
	rpcClient  *rpc.Client
	transactor *transactions.Transactor
}

func NewTransferBridge(rpcClient *rpc.Client, transactor *transactions.Transactor) *TransferBridge {
	return &TransferBridge{rpcClient: rpcClient, transactor: transactor}
}

func (s *TransferBridge) Name() string {
	return "Transfer"
}

func (s *TransferBridge) Can(from, to *params.Network, token *token.Token, balance *big.Int) (bool, error) {
	return from.ChainID == to.ChainID, nil
}

func (s *TransferBridge) CalculateFees(from, to *params.Network, token *token.Token, amountIn *big.Int, nativeTokenPrice, tokenPrice float64, gasPrice *big.Float) (*big.Int, *big.Int, error) {
	return big.NewInt(0), big.NewInt(0), nil
}

func (s *TransferBridge) EstimateGas(fromNetwork *params.Network, toNetwork *params.Network, from common.Address, to common.Address, token *token.Token, amountIn *big.Int) (uint64, error) {
	estimation := uint64(0)
	var err error
	if token.Symbol == "ETH" {
		estimation, err = s.transactor.EstimateGas(fromNetwork, from, to, amountIn, []byte("eth_sendRawTransaction"))
		if err != nil {
			return 0, err
		}
	} else {
		ethClient, err := s.rpcClient.EthClient(fromNetwork.ChainID)
		if err != nil {
			return 0, err
		}

		abi, err := abi.JSON(strings.NewReader(ierc20.IERC20ABI))
		if err != nil {
			return 0, err
		}
		input, err := abi.Pack("transfer",
			to,
			amountIn,
		)

		if err != nil {
			return 0, err
		}

		ctx := context.Background()

		msg := ethereum.CallMsg{
			From: from,
			To:   &token.Address,
			Data: input,
		}

		estimation, err = ethClient.EstimateGas(ctx, msg)
		if err != nil {
			return 0, err
		}

	}
	increasedEstimation := float64(estimation) * IncreaseEstimatedGasFactor
	return uint64(increasedEstimation), nil
}

func (s *TransferBridge) BuildTx(network, _ *params.Network, fromAddress common.Address, toAddress common.Address, token *token.Token, amountIn *big.Int, bonderFee *big.Int) (*ethTypes.Transaction, error) {
	toAddr := types.Address(toAddress)
	sendArgs := &TransactionBridge{
		TransferTx: &transactions.SendTxArgs{
			From:  types.Address(fromAddress),
			To:    &toAddr,
			Value: (*hexutil.Big)(amountIn),
			Data:  types.HexBytes("0x0"),
		},
		ChainID: network.ChainID,
	}

	return s.BuildTransaction(sendArgs)
}

func (s *TransferBridge) Send(sendArgs *TransactionBridge, verifiedAccount *account.SelectedExtKey) (types.Hash, error) {
	return s.transactor.SendTransactionWithChainID(sendArgs.ChainID, *sendArgs.TransferTx, verifiedAccount)
}

func (s *TransferBridge) BuildTransaction(sendArgs *TransactionBridge) (*ethTypes.Transaction, error) {
	return s.transactor.ValidateAndBuildTransaction(sendArgs.ChainID, *sendArgs.TransferTx)
}

func (s *TransferBridge) CalculateAmountOut(from, to *params.Network, amountIn *big.Int, symbol string) (*big.Int, error) {
	return amountIn, nil
}

func (s *TransferBridge) GetContractAddress(network *params.Network, token *token.Token) *common.Address {
	return nil
}
