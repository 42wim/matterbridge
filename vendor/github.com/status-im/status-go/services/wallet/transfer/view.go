package transfer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	w_common "github.com/status-im/status-go/services/wallet/common"
)

// View stores only fields used by a client and ensures that all relevant fields are
// encoded in hex.
type View struct {
	ID                   common.Hash    `json:"id"`
	Type                 w_common.Type  `json:"type"`
	Address              common.Address `json:"address"`
	BlockNumber          *hexutil.Big   `json:"blockNumber"`
	BlockHash            common.Hash    `json:"blockhash"`
	Timestamp            hexutil.Uint64 `json:"timestamp"`
	GasPrice             *hexutil.Big   `json:"gasPrice"`
	MaxFeePerGas         *hexutil.Big   `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big   `json:"maxPriorityFeePerGas"`
	EffectiveTip         *hexutil.Big   `json:"effectiveTip"`
	EffectiveGasPrice    *hexutil.Big   `json:"effectiveGasPrice"`
	GasLimit             hexutil.Uint64 `json:"gasLimit"`
	GasUsed              hexutil.Uint64 `json:"gasUsed"`
	Nonce                hexutil.Uint64 `json:"nonce"`
	TxStatus             hexutil.Uint64 `json:"txStatus"`
	Input                hexutil.Bytes  `json:"input"`
	TxHash               common.Hash    `json:"txHash"`
	Value                *hexutil.Big   `json:"value"`   // Only used for Type EthTransfer and Erc20Transfer
	TokenID              *hexutil.Big   `json:"tokenId"` // Only used for Type Erc721Transfer
	From                 common.Address `json:"from"`
	To                   common.Address `json:"to"`
	Contract             common.Address `json:"contract"`
	NetworkID            uint64         `json:"networkId"`
	MultiTransactionID   int64          `json:"multiTransactionID"`
	BaseGasFees          string         `json:"base_gas_fee"`
}

func castToTransferViews(transfers []Transfer) []View {
	views := make([]View, 0, len(transfers))
	for _, tx := range transfers {
		switch tx.Type {
		case w_common.EthTransfer, w_common.Erc20Transfer, w_common.Erc721Transfer:
			view := CastToTransferView(tx)
			views = append(views, view)
		}
	}
	return views
}

func CastToTransferView(t Transfer) View {
	view := View{}
	view.ID = t.ID
	view.Type = getFixedTransferType(t)
	view.Address = t.Address
	view.BlockNumber = (*hexutil.Big)(t.BlockNumber)
	view.BlockHash = t.BlockHash
	view.Timestamp = hexutil.Uint64(t.Timestamp)
	view.GasPrice = (*hexutil.Big)(t.Transaction.GasPrice())
	if t.BaseGasFees != "" {
		baseFee := new(big.Int)
		baseFee.SetString(t.BaseGasFees[2:], 16)
		tip := t.Transaction.EffectiveGasTipValue(baseFee)

		view.EffectiveTip = (*hexutil.Big)(tip)
		price := new(big.Int).Add(baseFee, tip)
		view.EffectiveGasPrice = (*hexutil.Big)(price)
	}
	view.MaxFeePerGas = (*hexutil.Big)(t.Transaction.GasFeeCap())
	view.MaxPriorityFeePerGas = (*hexutil.Big)(t.Transaction.GasTipCap())
	view.GasLimit = hexutil.Uint64(t.Transaction.Gas())
	view.GasUsed = hexutil.Uint64(t.Receipt.GasUsed)
	view.BaseGasFees = t.BaseGasFees
	view.Nonce = hexutil.Uint64(t.Transaction.Nonce())
	view.TxStatus = hexutil.Uint64(t.Receipt.Status)
	view.Input = hexutil.Bytes(t.Transaction.Data())
	view.TxHash = t.Transaction.Hash()
	view.NetworkID = t.NetworkID

	value := new(hexutil.Big)
	tokenID := new(hexutil.Big)

	switch view.Type {
	case w_common.EthTransfer:
		view.From = t.From
		if t.Transaction.To() != nil {
			view.To = *t.Transaction.To()
		}
		value = (*hexutil.Big)(t.Transaction.Value())
		view.Contract = t.Receipt.ContractAddress
	case w_common.Erc20Transfer:
		view.Contract = t.Log.Address
		from, to, valueInt := w_common.ParseErc20TransferLog(t.Log)
		view.From, view.To, value = from, to, (*hexutil.Big)(valueInt)
	case w_common.Erc721Transfer:
		view.Contract = t.Log.Address
		from, to, tokenIDInt := w_common.ParseErc721TransferLog(t.Log)
		view.From, view.To, tokenID = from, to, (*hexutil.Big)(tokenIDInt)
	}

	view.MultiTransactionID = int64(t.MultiTransactionID)
	view.Value = value
	view.TokenID = tokenID

	return view
}

func getFixedTransferType(tx Transfer) w_common.Type {
	// erc721 transfers share signature with erc20 ones, so they both used to be categorized as erc20
	// by the Downloader. We fix this here since they might be mis-categorized in the db.
	if tx.Type == w_common.Erc20Transfer {
		eventType := w_common.GetEventType(tx.Log)
		return w_common.EventTypeToSubtransactionType(eventType)
	}
	return tx.Type
}
