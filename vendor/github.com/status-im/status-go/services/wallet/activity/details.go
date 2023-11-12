package activity

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/sqlite"
)

type ProtocolType = int

const (
	ProtocolHop ProtocolType = iota + 1
	ProtocolUniswap
)

type EntryChainDetails struct {
	ChainID     int64        `json:"chainId"`
	BlockNumber int64        `json:"blockNumber"`
	Hash        eth.Hash     `json:"hash"`
	Contract    *eth.Address `json:"contractAddress,omitempty"`
}

type EntryDetails struct {
	ID           string              `json:"id"`
	MultiTxID    int                 `json:"multiTxId"`
	Nonce        uint64              `json:"nonce"`
	ChainDetails []EntryChainDetails `json:"chainDetails"`
	Input        string              `json:"input"`
	ProtocolType *ProtocolType       `json:"protocolType,omitempty"`
	MaxFeePerGas *hexutil.Big        `json:"maxFeePerGas"`
	GasLimit     uint64              `json:"gasLimit"`
	TotalFees    *hexutil.Big        `json:"totalFees,omitempty"`
}

func protocolTypeFromDBType(dbType string) (protocolType *ProtocolType) {
	protocolType = new(ProtocolType)
	switch common.Type(dbType) {
	case common.UniswapV2Swap:
		fallthrough
	case common.UniswapV3Swap:
		*protocolType = ProtocolUniswap
	case common.HopBridgeFrom:
		fallthrough
	case common.HopBridgeTo:
		*protocolType = ProtocolHop
	default:
		return nil
	}
	return protocolType
}

func getMultiTxDetails(ctx context.Context, db *sql.DB, multiTxID int) (*EntryDetails, error) {
	if multiTxID <= 0 {
		return nil, errors.New("invalid tx id")
	}

	// Extracting tx only when values are not null to prevent errors during the scan.
	rows, err := db.QueryContext(ctx, `
	SELECT
		tx_hash,
		blk_number,
		network_id,
		type,
		account_nonce,
		contract_address,
		CASE 
			WHEN json_extract(tx, '$.gas') = '0x0' THEN NULL
			ELSE transfers.tx
		END as tx,
		base_gas_fee
	FROM
		transfers
	WHERE
		multi_transaction_id = ?;`, multiTxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var maxFeePerGas *hexutil.Big
	var input string
	var protocolType *ProtocolType
	var nonce, gasLimit uint64
	var totalFees *hexutil.Big
	var chainDetailsList []EntryChainDetails
	for rows.Next() {
		var contractTypeDB sql.NullString
		var chainIDDB, nonceDB, blockNumber sql.NullInt64
		var transferHashDB, contractAddressDB sql.RawBytes
		var baseGasFees string
		tx := &types.Transaction{}
		nullableTx := sqlite.JSONBlob{Data: tx}
		err := rows.Scan(&transferHashDB, &blockNumber, &chainIDDB, &contractTypeDB, &nonceDB, &contractAddressDB, &nullableTx, &baseGasFees)
		if err != nil {
			return nil, err
		}

		var chainID int64
		if chainIDDB.Valid {
			chainID = chainIDDB.Int64
		}
		chainDetails := getChainDetails(chainID, &chainDetailsList)

		if len(transferHashDB) > 0 {
			chainDetails.Hash = eth.BytesToHash(transferHashDB)
		}
		if contractTypeDB.Valid && protocolType == nil {
			protocolType = protocolTypeFromDBType(contractTypeDB.String)
		}

		if blockNumber.Valid {
			chainDetails.BlockNumber = blockNumber.Int64
		}
		if nonceDB.Valid {
			nonce = uint64(nonceDB.Int64)
		}

		if len(contractAddressDB) > 0 && chainDetails.Contract == nil {
			chainDetails.Contract = new(eth.Address)
			*chainDetails.Contract = eth.BytesToAddress(contractAddressDB)
		}

		if nullableTx.Valid {
			input = "0x" + hex.EncodeToString(tx.Data())
			maxFeePerGas = (*hexutil.Big)(tx.GasFeeCap())
			gasLimit = tx.Gas()
			baseGasFees, _ := new(big.Int).SetString(baseGasFees, 0)
			totalFees = (*hexutil.Big)(getTotalFees(tx, baseGasFees))
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if maxFeePerGas == nil {
		maxFeePerGas = (*hexutil.Big)(big.NewInt(0))
	}

	if len(input) == 0 {
		input = "0x"
	}

	return &EntryDetails{
		MultiTxID:    multiTxID,
		Nonce:        nonce,
		ProtocolType: protocolType,
		Input:        input,
		MaxFeePerGas: maxFeePerGas,
		GasLimit:     gasLimit,
		ChainDetails: chainDetailsList,
		TotalFees:    totalFees,
	}, nil
}

func getTxDetails(ctx context.Context, db *sql.DB, id string) (*EntryDetails, error) {
	if len(id) == 0 {
		return nil, errors.New("invalid tx id")
	}
	rows, err := db.QueryContext(ctx, `
	SELECT
		tx_hash,
		blk_number,
		network_id,
		account_nonce,
		tx,
		contract_address,
		base_gas_fee
	FROM
		transfers
	WHERE
		hash = ?;`, eth.HexToHash(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("Entry not found")
	}

	tx := &types.Transaction{}
	nullableTx := sqlite.JSONBlob{Data: tx}
	var transferHashDB, contractAddressDB sql.RawBytes
	var chainIDDB, nonceDB, blockNumberDB sql.NullInt64
	var baseGasFees string
	err = rows.Scan(&transferHashDB, &blockNumberDB, &chainIDDB, &nonceDB, &nullableTx, &contractAddressDB, &baseGasFees)
	if err != nil {
		return nil, err
	}

	details := &EntryDetails{
		ID: id,
	}

	var chainID int64
	if chainIDDB.Valid {
		chainID = chainIDDB.Int64
	}
	chainDetails := getChainDetails(chainID, &details.ChainDetails)

	if blockNumberDB.Valid {
		chainDetails.BlockNumber = blockNumberDB.Int64
	}

	if nonceDB.Valid {
		details.Nonce = uint64(nonceDB.Int64)
	}

	if len(transferHashDB) > 0 {
		chainDetails.Hash = eth.BytesToHash(transferHashDB)
	}

	if len(contractAddressDB) > 0 {
		chainDetails.Contract = new(eth.Address)
		*chainDetails.Contract = eth.BytesToAddress(contractAddressDB)
	}

	if nullableTx.Valid {
		details.Input = "0x" + hex.EncodeToString(tx.Data())
		details.MaxFeePerGas = (*hexutil.Big)(tx.GasFeeCap())
		details.GasLimit = tx.Gas()
		baseGasFees, _ := new(big.Int).SetString(baseGasFees, 0)
		details.TotalFees = (*hexutil.Big)(getTotalFees(tx, baseGasFees))
	}

	return details, nil
}

func getTotalFees(tx *types.Transaction, baseFee *big.Int) *big.Int {
	if tx.Type() == types.DynamicFeeTxType {
		// EIP-1559 transaction
		if baseFee == nil {
			return nil
		}
		tip := tx.GasTipCap()
		maxFee := tx.GasFeeCap()
		gasUsed := big.NewInt(int64(tx.Gas()))

		totalGasUsed := new(big.Int).Add(tip, baseFee)
		if totalGasUsed.Cmp(maxFee) > 0 {
			totalGasUsed.Set(maxFee)
		}

		return new(big.Int).Mul(totalGasUsed, gasUsed)
	}

	// Legacy transaction
	gasPrice := tx.GasPrice()
	gasUsed := big.NewInt(int64(tx.Gas()))

	return new(big.Int).Mul(gasPrice, gasUsed)
}

func getChainDetails(chainID int64, data *[]EntryChainDetails) *EntryChainDetails {
	for i, entry := range *data {
		if entry.ChainID == chainID {
			return &(*data)[i]
		}
	}
	*data = append(*data, EntryChainDetails{
		ChainID: chainID,
	})
	return &(*data)[len(*data)-1]
}
