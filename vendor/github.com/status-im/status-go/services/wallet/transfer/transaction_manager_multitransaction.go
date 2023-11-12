package transfer

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/bridge"
	wallet_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/walletevent"
	"github.com/status-im/status-go/signal"
	"github.com/status-im/status-go/transactions"
)

const multiTransactionColumns = "from_network_id, from_tx_hash, from_address, from_asset, from_amount, to_network_id, to_tx_hash, to_address, to_asset, to_amount, type, cross_tx_id, timestamp"
const selectMultiTransactionColumns = "COALESCE(from_network_id, 0), from_tx_hash, from_address, from_asset, from_amount, COALESCE(to_network_id, 0), to_tx_hash, to_address, to_asset, to_amount, type, cross_tx_id, timestamp"

func rowsToMultiTransactions(rows *sql.Rows) ([]*MultiTransaction, error) {
	var multiTransactions []*MultiTransaction
	for rows.Next() {
		multiTransaction := &MultiTransaction{}
		var fromAmountDB, toAmountDB sql.NullString
		var fromTxHash, toTxHash sql.RawBytes
		err := rows.Scan(
			&multiTransaction.ID,
			&multiTransaction.FromNetworkID,
			&fromTxHash,
			&multiTransaction.FromAddress,
			&multiTransaction.FromAsset,
			&fromAmountDB,
			&multiTransaction.ToNetworkID,
			&toTxHash,
			&multiTransaction.ToAddress,
			&multiTransaction.ToAsset,
			&toAmountDB,
			&multiTransaction.Type,
			&multiTransaction.CrossTxID,
			&multiTransaction.Timestamp,
		)
		if len(fromTxHash) > 0 {
			multiTransaction.FromTxHash = common.BytesToHash(fromTxHash)
		}
		if len(toTxHash) > 0 {
			multiTransaction.ToTxHash = common.BytesToHash(toTxHash)
		}
		if err != nil {
			return nil, err
		}

		if fromAmountDB.Valid {
			multiTransaction.FromAmount = new(hexutil.Big)
			if _, ok := (*big.Int)(multiTransaction.FromAmount).SetString(fromAmountDB.String, 0); !ok {
				return nil, errors.New("failed to convert fromAmountDB.String to big.Int: " + fromAmountDB.String)
			}
		}

		if toAmountDB.Valid {
			multiTransaction.ToAmount = new(hexutil.Big)
			if _, ok := (*big.Int)(multiTransaction.ToAmount).SetString(toAmountDB.String, 0); !ok {
				return nil, errors.New("failed to convert fromAmountDB.String to big.Int: " + toAmountDB.String)
			}
		}

		multiTransactions = append(multiTransactions, multiTransaction)
	}

	return multiTransactions, nil
}

func getMultiTransactionTimestamp(multiTransaction *MultiTransaction) uint64 {
	if multiTransaction.Timestamp != 0 {
		return multiTransaction.Timestamp
	}
	return uint64(time.Now().Unix())
}

// insertMultiTransaction inserts a multi transaction into the database and updates multi-transaction ID and timestamp
func insertMultiTransaction(db *sql.DB, multiTransaction *MultiTransaction) (MultiTransactionIDType, error) {
	insert, err := db.Prepare(fmt.Sprintf(`INSERT INTO multi_transactions (%s)
											VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, multiTransactionColumns))
	if err != nil {
		return NoMultiTransactionID, err
	}
	timestamp := getMultiTransactionTimestamp(multiTransaction)
	result, err := insert.Exec(
		multiTransaction.FromNetworkID,
		multiTransaction.FromTxHash,
		multiTransaction.FromAddress,
		multiTransaction.FromAsset,
		multiTransaction.FromAmount.String(),
		multiTransaction.ToNetworkID,
		multiTransaction.ToTxHash,
		multiTransaction.ToAddress,
		multiTransaction.ToAsset,
		multiTransaction.ToAmount.String(),
		multiTransaction.Type,
		multiTransaction.CrossTxID,
		timestamp,
	)
	if err != nil {
		return NoMultiTransactionID, err
	}
	defer insert.Close()
	multiTransactionID, err := result.LastInsertId()

	multiTransaction.Timestamp = timestamp
	multiTransaction.ID = uint(multiTransactionID)

	return MultiTransactionIDType(multiTransactionID), err
}

func (tm *TransactionManager) InsertMultiTransaction(multiTransaction *MultiTransaction) (MultiTransactionIDType, error) {
	return tm.insertMultiTransactionAndNotify(tm.db, multiTransaction, nil)
}

func (tm *TransactionManager) insertMultiTransactionAndNotify(db *sql.DB, multiTransaction *MultiTransaction, chainIDs []uint64) (MultiTransactionIDType, error) {
	id, err := insertMultiTransaction(db, multiTransaction)
	if err != nil {
		publishMultiTransactionUpdatedEvent(db, multiTransaction, tm.eventFeed, chainIDs)
	}
	return id, err
}

// publishMultiTransactionUpdatedEvent notify listeners of new multi transaction (used in activity history)
func publishMultiTransactionUpdatedEvent(db *sql.DB, multiTransaction *MultiTransaction, eventFeed *event.Feed, chainIDs []uint64) {
	publishFn := func(chainID uint64) {
		eventFeed.Send(walletevent.Event{
			Type:     EventMTTransactionUpdate,
			ChainID:  chainID,
			Accounts: []common.Address{multiTransaction.FromAddress, multiTransaction.ToAddress},
			At:       int64(multiTransaction.Timestamp),
		})
	}
	if len(chainIDs) > 0 {
		for _, chainID := range chainIDs {
			publishFn(chainID)
		}
	} else {
		publishFn(0)
	}
}

func updateMultiTransaction(db *sql.DB, multiTransaction *MultiTransaction) error {
	if MultiTransactionIDType(multiTransaction.ID) == NoMultiTransactionID {
		return fmt.Errorf("no multitransaction ID")
	}

	update, err := db.Prepare(fmt.Sprintf(`REPLACE INTO multi_transactions (rowid, %s)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, multiTransactionColumns))

	if err != nil {
		return err
	}
	timestamp := getMultiTransactionTimestamp(multiTransaction)
	_, err = update.Exec(
		multiTransaction.ID,
		multiTransaction.FromNetworkID,
		multiTransaction.FromTxHash,
		multiTransaction.FromAddress,
		multiTransaction.FromAsset,
		multiTransaction.FromAmount.String(),
		multiTransaction.ToNetworkID,
		multiTransaction.ToTxHash,
		multiTransaction.ToAddress,
		multiTransaction.ToAsset,
		multiTransaction.ToAmount.String(),
		multiTransaction.Type,
		multiTransaction.CrossTxID,
		timestamp,
	)
	if err != nil {
		return err
	}
	return update.Close()
}

func (tm *TransactionManager) UpdateMultiTransaction(multiTransaction *MultiTransaction) error {
	return updateMultiTransaction(tm.db, multiTransaction)
}

// In case of keycard account, password should be empty
func (tm *TransactionManager) CreateMultiTransactionFromCommand(ctx context.Context, command *MultiTransactionCommand,
	data []*bridge.TransactionBridge, bridges map[string]bridge.Bridge, password string) (*MultiTransactionCommandResult, error) {

	multiTransaction := multiTransactionFromCommand(command)

	chainIDs := make([]uint64, 0, len(data))
	for _, tx := range data {
		chainIDs = append(chainIDs, tx.ChainID)
	}
	if multiTransaction.Type == MultiTransactionSend && multiTransaction.FromNetworkID == 0 && len(chainIDs) == 1 {
		multiTransaction.FromNetworkID = chainIDs[0]
	}
	multiTransactionID, err := tm.insertMultiTransactionAndNotify(tm.db, multiTransaction, chainIDs)
	if err != nil {
		return nil, err
	}

	multiTransaction.ID = uint(multiTransactionID)
	if password == "" {
		acc, err := tm.accountsDB.GetAccountByAddress(types.Address(multiTransaction.FromAddress))
		if err != nil {
			return nil, err
		}

		kp, err := tm.accountsDB.GetKeypairByKeyUID(acc.KeyUID)
		if err != nil {
			return nil, err
		}

		if !kp.MigratedToKeycard() {
			return nil, fmt.Errorf("account being used is not migrated to a keycard, password is required")
		}

		tm.multiTransactionForKeycardSigning = multiTransaction
		tm.transactionsBridgeData = data
		hashes, err := tm.buildTransactions(bridges)
		if err != nil {
			return nil, err
		}

		signal.SendTransactionsForSigningEvent(hashes)

		return nil, nil
	}

	hashes, err := tm.sendTransactions(multiTransaction, data, bridges, password)
	if err != nil {
		return nil, err
	}

	err = tm.storePendingTransactions(multiTransaction, hashes, data)
	if err != nil {
		return nil, err
	}

	return &MultiTransactionCommandResult{
		ID:     int64(multiTransactionID),
		Hashes: hashes,
	}, nil
}

func (tm *TransactionManager) ProceedWithTransactionsSignatures(ctx context.Context, signatures map[string]SignatureDetails) (*MultiTransactionCommandResult, error) {
	if tm.multiTransactionForKeycardSigning == nil {
		return nil, errors.New("no multi transaction to proceed with")
	}
	if len(tm.transactionsBridgeData) == 0 {
		return nil, errors.New("no transactions bridge data to proceed with")
	}
	if len(tm.transactionsForKeycardSingning) == 0 {
		return nil, errors.New("no transactions to proceed with")
	}
	if len(signatures) != len(tm.transactionsForKeycardSingning) {
		return nil, errors.New("not all transactions have been signed")
	}

	// check if all transactions have been signed
	for hash, desc := range tm.transactionsForKeycardSingning {
		sigDetails, ok := signatures[hash.String()]
		if !ok {
			return nil, fmt.Errorf("missing signature for transaction %s", hash)
		}

		rBytes, _ := hex.DecodeString(sigDetails.R)
		sBytes, _ := hex.DecodeString(sigDetails.S)
		vByte := byte(0)
		if sigDetails.V == "01" {
			vByte = 1
		}

		desc.signature = make([]byte, crypto.SignatureLength)
		copy(desc.signature[32-len(rBytes):32], rBytes)
		copy(desc.signature[64-len(rBytes):64], sBytes)
		desc.signature[64] = vByte
	}

	// send transactions
	hashes := make(map[uint64][]types.Hash)
	for _, desc := range tm.transactionsForKeycardSingning {
		hash, err := tm.transactor.AddSignatureToTransactionAndSend(desc.chainID, desc.builtTx, desc.signature)
		if err != nil {
			return nil, err
		}
		hashes[desc.chainID] = append(hashes[desc.chainID], hash)
	}

	err := tm.storePendingTransactions(tm.multiTransactionForKeycardSigning, hashes, tm.transactionsBridgeData)
	if err != nil {
		return nil, err
	}

	return &MultiTransactionCommandResult{
		ID:     int64(tm.multiTransactionForKeycardSigning.ID),
		Hashes: hashes,
	}, nil
}

func (tm *TransactionManager) storePendingTransactions(multiTransaction *MultiTransaction,
	hashes map[uint64][]types.Hash, data []*bridge.TransactionBridge) error {

	txs := createPendingTransactions(hashes, data, multiTransaction)
	for _, tx := range txs {
		err := tm.pendingTracker.StoreAndTrackPendingTx(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func createPendingTransactions(hashes map[uint64][]types.Hash, data []*bridge.TransactionBridge,
	multiTransaction *MultiTransaction) []*transactions.PendingTransaction {

	txs := make([]*transactions.PendingTransaction, 0)
	for _, tx := range data {
		for _, hash := range hashes[tx.ChainID] {
			pendingTransaction := &transactions.PendingTransaction{
				Hash:               common.Hash(hash),
				Timestamp:          uint64(time.Now().Unix()),
				Value:              bigint.BigInt{Int: multiTransaction.FromAmount.ToInt()},
				From:               common.Address(tx.From()),
				To:                 common.Address(tx.To()),
				Data:               tx.Data().String(),
				Type:               transactions.WalletTransfer,
				ChainID:            wallet_common.ChainID(tx.ChainID),
				MultiTransactionID: int64(multiTransaction.ID),
				Symbol:             multiTransaction.FromAsset,
				AutoDelete:         new(bool),
			}
			// Transaction downloader will delete pending transaction as soon as it is confirmed
			*pendingTransaction.AutoDelete = false
			txs = append(txs, pendingTransaction)
		}
	}
	return txs
}

func multiTransactionFromCommand(command *MultiTransactionCommand) *MultiTransaction {

	log.Info("Creating multi transaction", "command", command)

	multiTransaction := &MultiTransaction{
		FromAddress: command.FromAddress,
		ToAddress:   command.ToAddress,
		FromAsset:   command.FromAsset,
		ToAsset:     command.ToAsset,
		FromAmount:  command.FromAmount,
		ToAmount:    new(hexutil.Big),
		Type:        command.Type,
	}

	return multiTransaction
}

func (tm *TransactionManager) buildTransactions(bridges map[string]bridge.Bridge) ([]string, error) {
	tm.transactionsForKeycardSingning = make(map[common.Hash]*TransactionDescription)
	var hashes []string
	for _, bridgeTx := range tm.transactionsBridgeData {
		builtTx, err := bridges[bridgeTx.BridgeName].BuildTransaction(bridgeTx)
		if err != nil {
			return hashes, err
		}

		signer := ethTypes.NewLondonSigner(big.NewInt(int64(bridgeTx.ChainID)))
		txHash := signer.Hash(builtTx)

		tm.transactionsForKeycardSingning[txHash] = &TransactionDescription{
			chainID: bridgeTx.ChainID,
			builtTx: builtTx,
		}

		hashes = append(hashes, txHash.String())
	}

	return hashes, nil
}

func (tm *TransactionManager) sendTransactions(multiTransaction *MultiTransaction,
	data []*bridge.TransactionBridge, bridges map[string]bridge.Bridge, password string) (
	map[uint64][]types.Hash, error) {

	log.Info("Making transactions", "multiTransaction", multiTransaction)

	selectedAccount, err := tm.getVerifiedWalletAccount(multiTransaction.FromAddress.Hex(), password)
	if err != nil {
		return nil, err
	}

	hashes := make(map[uint64][]types.Hash)
	for _, tx := range data {
		hash, err := bridges[tx.BridgeName].Send(tx, selectedAccount)
		if err != nil {
			return nil, err
		}
		hashes[tx.ChainID] = append(hashes[tx.ChainID], hash)
	}
	return hashes, nil
}

func (tm *TransactionManager) GetMultiTransactions(ctx context.Context, ids []MultiTransactionIDType) ([]*MultiTransaction, error) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, v := range ids {
		placeholders[i] = "?"
		args[i] = v
	}

	stmt, err := tm.db.Prepare(fmt.Sprintf(`SELECT rowid, %s
											FROM multi_transactions
											WHERE rowid in (%s)`,
		selectMultiTransactionColumns,
		strings.Join(placeholders, ",")))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToMultiTransactions(rows)
}

func (tm *TransactionManager) getBridgeMultiTransactions(ctx context.Context, toChainID uint64, crossTxID string) ([]*MultiTransaction, error) {
	stmt, err := tm.db.Prepare(fmt.Sprintf(`SELECT rowid, %s
											FROM multi_transactions
											WHERE type=? AND to_network_id=? AND cross_tx_id=?`,
		multiTransactionColumns))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(MultiTransactionBridge, toChainID, crossTxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToMultiTransactions(rows)
}

func (tm *TransactionManager) GetBridgeOriginMultiTransaction(ctx context.Context, toChainID uint64, crossTxID string) (*MultiTransaction, error) {
	multiTxs, err := tm.getBridgeMultiTransactions(ctx, toChainID, crossTxID)
	if err != nil {
		return nil, err
	}

	for _, multiTx := range multiTxs {
		// Origin MultiTxs will have a missing "ToTxHash"
		if multiTx.ToTxHash == emptyHash {
			return multiTx, nil
		}
	}

	return nil, nil
}

func (tm *TransactionManager) GetBridgeDestinationMultiTransaction(ctx context.Context, toChainID uint64, crossTxID string) (*MultiTransaction, error) {
	multiTxs, err := tm.getBridgeMultiTransactions(ctx, toChainID, crossTxID)
	if err != nil {
		return nil, err
	}

	for _, multiTx := range multiTxs {
		// Destination MultiTxs will have a missing "FromTxHash"
		if multiTx.FromTxHash == emptyHash {
			return multiTx, nil
		}
	}

	return nil, nil
}

func (tm *TransactionManager) getVerifiedWalletAccount(address, password string) (*account.SelectedExtKey, error) {
	exists, err := tm.accountsDB.AddressExists(types.HexToAddress(address))
	if err != nil {
		log.Error("failed to query db for a given address", "address", address, "error", err)
		return nil, err
	}

	if !exists {
		log.Error("failed to get a selected account", "err", transactions.ErrInvalidTxSender)
		return nil, transactions.ErrAccountDoesntExist
	}

	key, err := tm.gethManager.VerifyAccountPassword(tm.config.KeyStoreDir, address, password)
	if err != nil {
		log.Error("failed to verify account", "account", address, "error", err)
		return nil, err
	}

	return &account.SelectedExtKey{
		Address:    key.Address,
		AccountKey: key,
	}, nil
}
