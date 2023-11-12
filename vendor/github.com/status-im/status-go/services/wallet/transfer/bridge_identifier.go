package transfer

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/status-im/status-go/rpc/chain"
	w_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/token"
)

// TODO: Find proper way to uniquely match Origin and Destination transactions (some sort of hash or uniqueID)
// Current approach is not failsafe (for example, if multiple identical bridge operations are triggered
// at the same time)
// Recipient + Relayer + Data should match in both Origin and Destination transactions
func getHopBridgeFromL1CrossTxID(recipient common.Address, relayer common.Address, logData []byte) string {
	return fmt.Sprintf("FromL1_%s_%s_%s", recipient.String(), relayer.String(), hex.EncodeToString(logData))
}

func getHopBridgeFromL2CrossTxID(transferID *big.Int) string {
	return fmt.Sprintf("FromL2_0x%s", transferID.Text(16))
}

type originTxParams struct {
	fromNetworkID uint64
	fromTxHash    common.Hash
	fromAddress   common.Address
	fromAsset     string
	fromAmount    *big.Int
	toNetworkID   uint64
	toAddress     common.Address
	crossTxID     string
	timestamp     uint64
}

func upsertHopBridgeOriginTx(ctx context.Context, transactionManager *TransactionManager, params originTxParams) (*MultiTransaction, error) {
	// Try to find "destination" half of the multiTx
	multiTx, err := transactionManager.GetBridgeDestinationMultiTransaction(ctx, params.toNetworkID, params.crossTxID)
	if err != nil {
		return nil, err
	}

	if multiTx == nil {
		multiTx = &MultiTransaction{
			// Data from "origin" transaction
			FromNetworkID: params.fromNetworkID,
			FromTxHash:    params.fromTxHash,
			FromAddress:   params.fromAddress,
			FromAsset:     params.fromAsset,
			FromAmount:    (*hexutil.Big)(params.fromAmount),
			ToNetworkID:   params.toNetworkID,
			ToAddress:     params.toAddress,
			// To be replaced by "destination" transaction, need to be non-null
			ToAsset:  params.fromAsset,
			ToAmount: (*hexutil.Big)(params.fromAmount),
			// Common data
			Type:      MultiTransactionBridge,
			CrossTxID: params.crossTxID,
			Timestamp: params.timestamp,
		}

		_, err := transactionManager.InsertMultiTransaction(multiTx)
		if err != nil {
			return nil, err
		}

	} else {
		multiTx.FromNetworkID = params.fromNetworkID
		multiTx.FromTxHash = params.fromTxHash
		multiTx.FromAddress = params.fromAddress
		multiTx.FromAsset = params.fromAsset
		multiTx.FromAmount = (*hexutil.Big)(params.fromAmount)
		multiTx.Timestamp = params.timestamp

		err := transactionManager.UpdateMultiTransaction(multiTx)
		if err != nil {
			return nil, err
		}
	}
	return multiTx, nil
}

type destinationTxParams struct {
	toNetworkID uint64
	toTxHash    common.Hash
	toAddress   common.Address
	toAsset     string
	toAmount    *big.Int
	crossTxID   string
	timestamp   uint64
}

func upsertHopBridgeDestinationTx(ctx context.Context, transactionManager *TransactionManager, params destinationTxParams) (*MultiTransaction, error) {
	// Try to find "origin" half of the multiTx
	multiTx, err := transactionManager.GetBridgeOriginMultiTransaction(ctx, params.toNetworkID, params.crossTxID)
	if err != nil {
		return nil, err
	}

	if multiTx == nil {
		multiTx = &MultiTransaction{
			// To be replaced by "origin" transaction, need to be non-null
			FromAddress: params.toAddress,
			FromAsset:   params.toAsset,
			FromAmount:  (*hexutil.Big)(params.toAmount),
			// Data from "destination" transaction
			ToNetworkID: params.toNetworkID,
			ToTxHash:    params.toTxHash,
			ToAddress:   params.toAddress,
			ToAsset:     params.toAsset,
			ToAmount:    (*hexutil.Big)(params.toAmount),
			// Common data
			Type:      MultiTransactionBridge,
			CrossTxID: params.crossTxID,
			Timestamp: params.timestamp,
		}

		_, err := transactionManager.InsertMultiTransaction(multiTx)
		if err != nil {
			return nil, err
		}
	} else {
		multiTx.ToTxHash = params.toTxHash
		multiTx.ToAsset = params.toAsset
		multiTx.ToAmount = (*hexutil.Big)(params.toAmount)
		multiTx.Timestamp = params.timestamp

		err := transactionManager.UpdateMultiTransaction(multiTx)
		if err != nil {
			return nil, err
		}
	}
	return multiTx, nil
}

func buildHopBridgeMultitransaction(ctx context.Context, client chain.ClientInterface, transactionManager *TransactionManager, tokenManager *token.Manager, subTx *Transfer) (*MultiTransaction, error) {
	// Identify if it's from/to transaction
	switch w_common.GetEventType(subTx.Log) {
	case w_common.HopBridgeTransferSentToL2EventType:
		// L1->L2 Origin transaction
		toChainID, recipient, relayer, fromAmount, err := w_common.ParseHopBridgeTransferSentToL2Log(subTx.Log)
		if err != nil {
			return nil, err
		}

		params := originTxParams{
			fromNetworkID: subTx.NetworkID,
			fromTxHash:    subTx.Receipt.TxHash,
			fromAddress:   subTx.From,
			fromAsset:     "ETH",
			fromAmount:    fromAmount,
			toNetworkID:   toChainID,
			toAddress:     recipient,
			crossTxID:     getHopBridgeFromL1CrossTxID(recipient, relayer, subTx.Log.Data),
			timestamp:     subTx.Timestamp,
		}

		return upsertHopBridgeOriginTx(ctx, transactionManager, params)

	case w_common.HopBridgeTransferFromL1CompletedEventType:
		// L1->L2 Destination transaction
		recipient, relayer, toAmount, err := w_common.ParseHopBridgeTransferFromL1CompletedLog(subTx.Log)
		if err != nil {
			return nil, err
		}

		params := destinationTxParams{
			toNetworkID: subTx.NetworkID,
			toTxHash:    subTx.Receipt.TxHash,
			toAddress:   recipient,
			toAsset:     "ETH",
			toAmount:    toAmount,
			crossTxID:   getHopBridgeFromL1CrossTxID(recipient, relayer, subTx.Log.Data),
			timestamp:   subTx.Timestamp,
		}

		return upsertHopBridgeDestinationTx(ctx, transactionManager, params)

	case w_common.HopBridgeTransferSentEventType:
		// L2->L1 / L2->L2 Origin transaction
		transferID, toChainID, recipient, fromAmount, _, _, _, _, _, err := w_common.ParseHopBridgeTransferSentLog(subTx.Log)
		if err != nil {
			return nil, err
		}

		params := originTxParams{
			fromNetworkID: subTx.NetworkID,
			fromTxHash:    subTx.Receipt.TxHash,
			fromAddress:   subTx.From,
			fromAsset:     "ETH",
			fromAmount:    fromAmount,
			toNetworkID:   toChainID,
			toAddress:     recipient,
			crossTxID:     getHopBridgeFromL2CrossTxID(transferID),
			timestamp:     subTx.Timestamp,
		}

		return upsertHopBridgeOriginTx(ctx, transactionManager, params)

	case w_common.HopBridgeWithdrawalBondedEventType:
		// L2->L1 / L2->L2 Destination transaction
		transferID, toAmount, err := w_common.ParseHopWithdrawalBondedLog(subTx.Log)
		if err != nil {
			return nil, err
		}

		params := destinationTxParams{
			toNetworkID: subTx.NetworkID,
			toTxHash:    subTx.Receipt.TxHash,
			toAddress:   subTx.Address,
			toAsset:     "ETH",
			toAmount:    toAmount,
			crossTxID:   getHopBridgeFromL2CrossTxID(transferID),
			timestamp:   subTx.Timestamp,
		}

		return upsertHopBridgeDestinationTx(ctx, transactionManager, params)
	}
	return nil, nil
}
