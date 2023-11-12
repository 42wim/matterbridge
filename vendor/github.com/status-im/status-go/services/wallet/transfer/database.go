package transfer

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/services/wallet/bigint"
	w_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/sqlite"
)

// DBHeader fields from header that are stored in database.
type DBHeader struct {
	Number                *big.Int
	Hash                  common.Hash
	Timestamp             uint64
	PreloadedTransactions []*PreloadedTransaction
	Network               uint64
	Address               common.Address
	// Head is true if the block was a head at the time it was pulled from chain.
	Head bool
	// Loaded is true if transfers from this block have been already fetched
	Loaded bool
}

func toDBHeader(header *types.Header, blockHash common.Hash, account common.Address) *DBHeader {
	return &DBHeader{
		Hash:      blockHash,
		Number:    header.Number,
		Timestamp: header.Time,
		Loaded:    false,
		Address:   account,
	}
}

// SyncOption is used to specify that application processed transfers for that block.
type SyncOption uint

// JSONBlob type for marshaling/unmarshaling inner type to json.
type JSONBlob struct {
	data interface{}
}

// Scan implements interface.
func (blob *JSONBlob) Scan(value interface{}) error {
	if value == nil || reflect.ValueOf(blob.data).IsNil() {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("not a byte slice")
	}
	if len(bytes) == 0 {
		return nil
	}
	err := json.Unmarshal(bytes, blob.data)
	return err
}

// Value implements interface.
func (blob *JSONBlob) Value() (driver.Value, error) {
	if blob.data == nil || reflect.ValueOf(blob.data).IsNil() {
		return nil, nil
	}
	return json.Marshal(blob.data)
}

func NewDB(client *sql.DB) *Database {
	return &Database{client: client}
}

// Database sql wrapper for operations with wallet objects.
type Database struct {
	client *sql.DB
}

// Close closes database.
func (db *Database) Close() error {
	return db.client.Close()
}

func (db *Database) SaveBlocks(chainID uint64, headers []*DBHeader) (err error) {
	var (
		tx *sql.Tx
	)
	tx, err = db.client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	err = insertBlocksWithTransactions(chainID, tx, headers)
	if err != nil {
		return
	}

	return
}

func saveTransfersMarkBlocksLoaded(creator statementCreator, chainID uint64, address common.Address, transfers []Transfer, blocks []*big.Int) (err error) {
	err = updateOrInsertTransfers(chainID, creator, transfers)
	if err != nil {
		return
	}

	err = markBlocksAsLoaded(chainID, creator, address, blocks)
	if err != nil {
		return
	}

	return
}

// GetTransfersInRange loads transfers for a given address between two blocks.
func (db *Database) GetTransfersInRange(chainID uint64, address common.Address, start, end *big.Int) (rst []Transfer, err error) {
	query := newTransfersQuery().FilterNetwork(chainID).FilterAddress(address).FilterStart(start).FilterEnd(end).FilterLoaded(1)
	rows, err := db.client.Query(query.String(), query.Args()...)
	if err != nil {
		return
	}
	defer rows.Close()
	return query.TransferScan(rows)
}

// GetTransfersByAddress loads transfers for a given address between two blocks.
func (db *Database) GetTransfersByAddress(chainID uint64, address common.Address, toBlock *big.Int, limit int64) (rst []Transfer, err error) {
	query := newTransfersQuery().
		FilterNetwork(chainID).
		FilterAddress(address).
		FilterEnd(toBlock).
		FilterLoaded(1).
		SortByBlockNumberAndHash().
		Limit(limit)

	rows, err := db.client.Query(query.String(), query.Args()...)
	if err != nil {
		return
	}
	defer rows.Close()
	return query.TransferScan(rows)
}

// GetTransfersByAddressAndBlock loads transfers for a given address and block.
func (db *Database) GetTransfersByAddressAndBlock(chainID uint64, address common.Address, block *big.Int, limit int64) (rst []Transfer, err error) {
	query := newTransfersQuery().
		FilterNetwork(chainID).
		FilterAddress(address).
		FilterBlockNumber(block).
		FilterLoaded(1).
		SortByBlockNumberAndHash().
		Limit(limit)

	rows, err := db.client.Query(query.String(), query.Args()...)
	if err != nil {
		return
	}
	defer rows.Close()
	return query.TransferScan(rows)
}

// GetTransfers load transfers transfer between two blocks.
func (db *Database) GetTransfers(chainID uint64, start, end *big.Int) (rst []Transfer, err error) {
	query := newTransfersQuery().FilterNetwork(chainID).FilterStart(start).FilterEnd(end).FilterLoaded(1)
	rows, err := db.client.Query(query.String(), query.Args()...)
	if err != nil {
		return
	}
	defer rows.Close()
	return query.TransferScan(rows)
}

func (db *Database) GetTransfersForIdentities(ctx context.Context, identities []TransactionIdentity) (rst []Transfer, err error) {
	query := newTransfersQuery()
	for _, identity := range identities {
		subQuery := newSubQuery()
		subQuery = subQuery.FilterNetwork(uint64(identity.ChainID)).FilterTransactionID(identity.Hash).FilterAddress(identity.Address)
		query.addSubQuery(subQuery, OrSeparator)
	}
	rows, err := db.client.QueryContext(ctx, query.String(), query.Args()...)
	if err != nil {
		return
	}
	defer rows.Close()
	return query.TransferScan(rows)
}

func (db *Database) GetTransactionsToLoad(chainID uint64, address common.Address, blockNumber *big.Int) (rst []*PreloadedTransaction, err error) {
	query := newTransfersQueryForPreloadedTransactions().
		FilterNetwork(chainID).
		FilterLoaded(0)

	if address != (common.Address{}) {
		query.FilterAddress(address)
	}

	if blockNumber != nil {
		query.FilterBlockNumber(blockNumber)
	}

	rows, err := db.client.Query(query.String(), query.Args()...)
	if err != nil {
		return
	}
	defer rows.Close()
	return query.PreloadedTransactionScan(rows)
}

// statementCreator allows to pass transaction or database to use in consumer.
type statementCreator interface {
	Prepare(query string) (*sql.Stmt, error)
}

// Only used by status-mobile
func (db *Database) InsertBlock(chainID uint64, account common.Address, blockNumber *big.Int, blockHash common.Hash) error {
	var (
		tx *sql.Tx
	)
	tx, err := db.client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	blockDB := blockDBFields{
		chainID:     chainID,
		account:     account,
		blockNumber: blockNumber,
		blockHash:   blockHash,
	}
	return insertBlockDBFields(tx, blockDB)
}

type blockDBFields struct {
	chainID     uint64
	account     common.Address
	blockNumber *big.Int
	blockHash   common.Hash
}

func insertBlockDBFields(creator statementCreator, block blockDBFields) error {
	insert, err := creator.Prepare("INSERT OR IGNORE INTO blocks(network_id, address, blk_number, blk_hash, loaded) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = insert.Exec(block.chainID, block.account, (*bigint.SQLBigInt)(block.blockNumber), block.blockHash, true)
	return err
}

func insertBlocksWithTransactions(chainID uint64, creator statementCreator, headers []*DBHeader) error {
	insert, err := creator.Prepare("INSERT OR IGNORE INTO blocks(network_id, address, blk_number, blk_hash, loaded) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	updateTx, err := creator.Prepare(`UPDATE transfers
	SET log = ?, log_index = ?
	WHERE network_id = ? AND address = ? AND hash = ?`)
	if err != nil {
		return err
	}

	insertTx, err := creator.Prepare(`INSERT OR IGNORE
	INTO transfers (network_id, address, sender, hash, blk_number, blk_hash, type, timestamp, log, loaded, log_index, token_id, amount_padded128hex)
	VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, 0, ?, ?, ?)`)
	if err != nil {
		return err
	}

	for _, header := range headers {
		_, err = insert.Exec(chainID, header.Address, (*bigint.SQLBigInt)(header.Number), header.Hash, header.Loaded)
		if err != nil {
			return err
		}
		for _, transaction := range header.PreloadedTransactions {
			var logIndex *uint
			if transaction.Log != nil {
				logIndex = new(uint)
				*logIndex = transaction.Log.Index
			}
			res, err := updateTx.Exec(&JSONBlob{transaction.Log}, logIndex, chainID, header.Address, transaction.ID)
			if err != nil {
				return err
			}
			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected > 0 {
				continue
			}

			tokenID := (*bigint.SQLBigIntBytes)(transaction.TokenID)
			txValue := sqlite.BigIntToPadded128BitsStr(transaction.Value)
			// Is that correct to set sender as account address?
			_, err = insertTx.Exec(chainID, header.Address, header.Address, transaction.ID, (*bigint.SQLBigInt)(header.Number), header.Hash, transaction.Type, &JSONBlob{transaction.Log}, logIndex, tokenID, txValue)
			if err != nil {
				log.Error("error saving token transfer", "err", err)
				return err
			}
		}
	}
	return nil
}

func updateOrInsertTransfers(chainID uint64, creator statementCreator, transfers []Transfer) error {
	txsDBFields := make([]transferDBFields, 0, len(transfers))
	for _, t := range transfers {
		var receiptType *uint8
		var txHash, blockHash *common.Hash
		var receiptStatus, cumulativeGasUsed, gasUsed *uint64
		var contractAddress *common.Address
		var transactionIndex, logIndex *uint

		if t.Receipt != nil {
			receiptType = &t.Receipt.Type
			receiptStatus = &t.Receipt.Status
			txHash = &t.Receipt.TxHash
			if t.Log != nil {
				logIndex = new(uint)
				*logIndex = t.Log.Index
			}
			blockHash = &t.Receipt.BlockHash
			cumulativeGasUsed = &t.Receipt.CumulativeGasUsed
			contractAddress = &t.Receipt.ContractAddress
			gasUsed = &t.Receipt.GasUsed
			transactionIndex = &t.Receipt.TransactionIndex
		}

		var txProtected *bool
		var txGas, txNonce, txSize *uint64
		var txGasPrice, txGasTipCap, txGasFeeCap *big.Int
		var txType *uint8
		var txValue *big.Int
		var tokenAddress *common.Address
		var tokenID *big.Int
		var txFrom *common.Address
		var txTo *common.Address
		if t.Transaction != nil {
			if t.Log != nil {
				_, tokenAddress, txFrom, txTo = w_common.ExtractTokenTransferData(t.Type, t.Log, t.Transaction)
				tokenID = t.TokenID
				// Zero tokenID can be used for ERC721 and ERC1155 transfers but when serialzed/deserialized it becomes nil
				// as 0 value of big.Int bytes is nil.
				if tokenID == nil && (t.Type == w_common.Erc721Transfer || t.Type == w_common.Erc1155Transfer) {
					tokenID = big.NewInt(0)
				}
				txValue = t.TokenValue
			} else {
				txValue = new(big.Int).Set(t.Transaction.Value())
				txFrom = &t.From
				txTo = t.Transaction.To()
			}

			txType = new(uint8)
			*txType = t.Transaction.Type()
			txProtected = new(bool)
			*txProtected = t.Transaction.Protected()
			txGas = new(uint64)
			*txGas = t.Transaction.Gas()
			txGasPrice = t.Transaction.GasPrice()
			txGasTipCap = t.Transaction.GasTipCap()
			txGasFeeCap = t.Transaction.GasFeeCap()
			txNonce = new(uint64)
			*txNonce = t.Transaction.Nonce()
			txSize = new(uint64)
			*txSize = t.Transaction.Size()
		}

		dbFields := transferDBFields{
			chainID:            chainID,
			id:                 t.ID,
			blockHash:          t.BlockHash,
			blockNumber:        t.BlockNumber,
			timestamp:          t.Timestamp,
			address:            t.Address,
			transaction:        t.Transaction,
			sender:             t.From,
			receipt:            t.Receipt,
			log:                t.Log,
			transferType:       t.Type,
			baseGasFees:        t.BaseGasFees,
			multiTransactionID: t.MultiTransactionID,
			receiptStatus:      receiptStatus,
			receiptType:        receiptType,
			txHash:             txHash,
			logIndex:           logIndex,
			receiptBlockHash:   blockHash,
			cumulativeGasUsed:  cumulativeGasUsed,
			contractAddress:    contractAddress,
			gasUsed:            gasUsed,
			transactionIndex:   transactionIndex,
			txType:             txType,
			txProtected:        txProtected,
			txGas:              txGas,
			txGasPrice:         txGasPrice,
			txGasTipCap:        txGasTipCap,
			txGasFeeCap:        txGasFeeCap,
			txValue:            txValue,
			txNonce:            txNonce,
			txSize:             txSize,
			tokenAddress:       tokenAddress,
			tokenID:            tokenID,
			txFrom:             txFrom,
			txTo:               txTo,
		}
		txsDBFields = append(txsDBFields, dbFields)
	}

	return updateOrInsertTransfersDBFields(creator, txsDBFields)
}

type transferDBFields struct {
	chainID            uint64
	id                 common.Hash
	blockHash          common.Hash
	blockNumber        *big.Int
	timestamp          uint64
	address            common.Address
	transaction        *types.Transaction
	sender             common.Address
	receipt            *types.Receipt
	log                *types.Log
	transferType       w_common.Type
	baseGasFees        string
	multiTransactionID MultiTransactionIDType
	receiptStatus      *uint64
	receiptType        *uint8
	txHash             *common.Hash
	logIndex           *uint
	receiptBlockHash   *common.Hash
	cumulativeGasUsed  *uint64
	contractAddress    *common.Address
	gasUsed            *uint64
	transactionIndex   *uint
	txType             *uint8
	txProtected        *bool
	txGas              *uint64
	txGasPrice         *big.Int
	txGasTipCap        *big.Int
	txGasFeeCap        *big.Int
	txValue            *big.Int
	txNonce            *uint64
	txSize             *uint64
	tokenAddress       *common.Address
	tokenID            *big.Int
	txFrom             *common.Address
	txTo               *common.Address
}

func updateOrInsertTransfersDBFields(creator statementCreator, transfers []transferDBFields) error {
	insert, err := creator.Prepare(`INSERT OR REPLACE INTO transfers
        (network_id, hash, blk_hash, blk_number, timestamp, address, tx, sender, receipt, log, type, loaded, base_gas_fee, multi_transaction_id,
		status, receipt_type, tx_hash, log_index, block_hash, cumulative_gas_used, contract_address, gas_used, tx_index,
		tx_type, protected, gas_limit, gas_price_clamped64, gas_tip_cap_clamped64, gas_fee_cap_clamped64, amount_padded128hex, account_nonce, size, token_address, token_id, tx_from_address, tx_to_address)
	VALUES
        (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	for _, t := range transfers {
		txGasPrice := sqlite.BigIntToClampedInt64(t.txGasPrice)
		txGasTipCap := sqlite.BigIntToClampedInt64(t.txGasTipCap)
		txGasFeeCap := sqlite.BigIntToClampedInt64(t.txGasFeeCap)
		txValue := sqlite.BigIntToPadded128BitsStr(t.txValue)

		_, err = insert.Exec(t.chainID, t.id, t.blockHash, (*bigint.SQLBigInt)(t.blockNumber), t.timestamp, t.address, &JSONBlob{t.transaction}, t.sender, &JSONBlob{t.receipt}, &JSONBlob{t.log}, t.transferType, t.baseGasFees, t.multiTransactionID,
			t.receiptStatus, t.receiptType, t.txHash, t.logIndex, t.receiptBlockHash, t.cumulativeGasUsed, t.contractAddress, t.gasUsed, t.transactionIndex,
			t.txType, t.txProtected, t.txGas, txGasPrice, txGasTipCap, txGasFeeCap, txValue, t.txNonce, t.txSize, t.tokenAddress, (*bigint.SQLBigIntBytes)(t.tokenID), t.txFrom, t.txTo)
		if err != nil {
			log.Error("can't save transfer", "b-hash", t.blockHash, "b-n", t.blockNumber, "a", t.address, "h", t.id)
			return err
		}

		err = removeGasOnlyEthTransfer(creator, t)
		if err != nil {
			log.Error("can't remove gas only eth transfer", "b-hash", t.blockHash, "b-n", t.blockNumber, "a", t.address, "h", t.id, "err", err)
			// no return err, since it's not critical
		}
	}
	return nil
}

func removeGasOnlyEthTransfer(creator statementCreator, t transferDBFields) error {
	if t.transferType != w_common.EthTransfer {
		query, err := creator.Prepare(`DELETE FROM transfers WHERE tx_hash = ? AND address = ? AND network_id = ?
		 AND account_nonce = ? AND type = 'eth' AND amount_padded128hex = '00000000000000000000000000000000'`)
		if err != nil {
			return err
		}

		_, err = query.Exec(t.txHash, t.address, t.chainID, t.txNonce)
		if err != nil {
			return err
		}
	}
	return nil
}

// markBlocksAsLoaded(chainID, tx, address, blockNumbers)
// In case block contains both ETH and token transfers, it will be marked as loaded on ETH transfer processing.
// This is not a problem since for token transfers we have preloaded transactions and blocks 'loaded' flag is needed
// for ETH transfers only.
func markBlocksAsLoaded(chainID uint64, creator statementCreator, address common.Address, blocks []*big.Int) error {
	update, err := creator.Prepare("UPDATE blocks SET loaded=? WHERE address=? AND blk_number=? AND network_id=?")
	if err != nil {
		return err
	}

	for _, block := range blocks {
		_, err := update.Exec(true, address, (*bigint.SQLBigInt)(block), chainID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetOwnedMultiTransactionID returns sql.ErrNoRows if no transaction is found for the given identity
func GetOwnedMultiTransactionID(tx *sql.Tx, chainID w_common.ChainID, id common.Hash, address common.Address) (mTID int64, err error) {
	row := tx.QueryRow(`SELECT COALESCE(multi_transaction_id, 0) FROM transfers WHERE network_id = ? AND hash = ? AND address = ?`, chainID, id, address)
	err = row.Scan(&mTID)
	if err != nil {
		return 0, err
	}
	return mTID, nil
}

func (db *Database) GetLatestCollectibleTransfer(address common.Address, id thirdparty.CollectibleUniqueID) (*Transfer, error) {
	query := newTransfersQuery().
		FilterAddress(address).
		FilterNetwork(uint64(id.ContractID.ChainID)).
		FilterTokenAddress(id.ContractID.Address).
		FilterTokenID(id.TokenID.Int).
		FilterLoaded(1).
		SortByTimestamp(false).
		Limit(1)
	rows, err := db.client.Query(query.String(), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers, err := query.TransferScan(rows)
	if err == sql.ErrNoRows || len(transfers) == 0 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &transfers[0], nil
}

// Delete blocks for address and chainID
// Transfers will be deleted by cascade
func deleteBlocks(creator statementCreator, address common.Address) error {
	delete, err := creator.Prepare("DELETE FROM blocks WHERE address = ?")
	if err != nil {
		return err
	}

	_, err = delete.Exec(address)
	return err
}

func getAddresses(creator statementCreator) (rst []common.Address, err error) {
	stmt, err := creator.Prepare(`SELECT address FROM transfers UNION SELECT address FROM blocks UNION
		SELECT address FROM blocks_ranges_sequential UNION SELECT address FROM blocks_ranges`)
	if err != nil {
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	address := common.Address{}
	for rows.Next() {
		err = rows.Scan(&address)
		if err != nil {
			return nil, err
		}
		rst = append(rst, address)
	}

	return rst, nil
}
