package stickers

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/contracts/snt"
	"github.com/status-im/status-go/contracts/stickers"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/services/wallet/bigint"
)

func (api *API) BuyPrepareTxCallMsg(chainID uint64, from types.Address, packID *bigint.BigInt) (ethereum.CallMsg, error) {
	callOpts := &bind.CallOpts{Context: api.ctx, Pending: false}

	stickerType, err := api.contractMaker.NewStickerType(chainID)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	packInfo, err := stickerType.GetPackData(callOpts, packID.Int)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	stickerMarketABI, err := abi.JSON(strings.NewReader(stickers.StickerMarketABI))
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	extraData, err := stickerMarketABI.Pack("buyToken", packID.Int, from, packInfo.Price)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	sntABI, err := abi.JSON(strings.NewReader(snt.SNTABI))
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	stickerMarketAddress, err := stickers.StickerMarketContractAddress(chainID)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	data, err := sntABI.Pack("approveAndCall", stickerMarketAddress, packInfo.Price, extraData)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	sntAddress, err := snt.ContractAddress(chainID)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	return ethereum.CallMsg{
		From:  common.Address(from),
		To:    &sntAddress,
		Value: big.NewInt(0),
		Data:  data,
	}, nil
}

func (api *API) BuyPrepareTx(ctx context.Context, chainID uint64, from types.Address, packID *bigint.BigInt) (interface{}, error) {
	callMsg, err := api.BuyPrepareTxCallMsg(chainID, from, packID)
	if err != nil {
		return nil, err
	}

	return toCallArg(callMsg), nil
}

func (api *API) BuyEstimate(ctx context.Context, chainID uint64, from types.Address, packID *bigint.BigInt) (uint64, error) {
	callMsg, err := api.BuyPrepareTxCallMsg(chainID, from, packID)
	if err != nil {
		return 0, err
	}
	ethClient, err := api.contractMaker.RPCClient.EthClient(chainID)
	if err != nil {
		return 0, err
	}

	return ethClient.EstimateGas(ctx, callMsg)
}

func (api *API) StickerMarketAddress(ctx context.Context, chainID uint64) (common.Address, error) {
	return stickers.StickerMarketContractAddress(chainID)
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
