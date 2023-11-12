package history

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/bigint"
)

type BalanceDB struct {
	db *sql.DB
}

func NewBalanceDB(sqlDb *sql.DB) *BalanceDB {
	return &BalanceDB{
		db: sqlDb,
	}
}

// entry represents a single row in the balance_history table
type entry struct {
	chainID      uint64
	address      common.Address
	tokenSymbol  string
	tokenAddress common.Address
	block        *big.Int
	timestamp    int64
	balance      *big.Int
}

type assetIdentity struct {
	ChainID     uint64
	Addresses   []common.Address
	TokenSymbol string
}

func (a *assetIdentity) addressesToString() string {
	var addressesStr string
	for i, address := range a.Addresses {
		addressStr := hex.EncodeToString(address[:])
		if i == 0 {
			addressesStr = "X'" + addressStr + "'"
		} else {
			addressesStr += ", X'" + addressStr + "'"
		}
	}
	return addressesStr
}

func (e *entry) String() string {
	return fmt.Sprintf("chainID: %v, address: %v, tokenSymbol: %v, tokenAddress: %v, block: %v, timestamp: %v, balance: %v",
		e.chainID, e.address, e.tokenSymbol, e.tokenAddress, e.block, e.timestamp, e.balance)
}

func (b *BalanceDB) add(entry *entry) error {
	log.Debug("Adding entry to balance_history", "entry", entry)

	_, err := b.db.Exec("INSERT OR IGNORE INTO balance_history (chain_id, address, currency, block, timestamp, balance) VALUES (?, ?, ?, ?, ?, ?)", entry.chainID, entry.address, entry.tokenSymbol, (*bigint.SQLBigInt)(entry.block), entry.timestamp, (*bigint.SQLBigIntBytes)(entry.balance))
	return err
}

func (b *BalanceDB) getEntriesWithoutBalances(chainID uint64, address common.Address) (entries []*entry, err error) {
	rows, err := b.db.Query("SELECT blk_number, tr.timestamp, token_address from transfers tr LEFT JOIN balance_history bh ON bh.block = tr.blk_number WHERE tr.network_id = ? AND tr.address = ? AND tr.type != 'erc721' AND bh.block IS NULL",
		chainID, address)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries = make([]*entry, 0)
	for rows.Next() {
		entry := &entry{
			chainID: chainID,
			address: address,
			block:   new(big.Int),
		}

		// tokenAddress can be NULL and can not unmarshal to common.Address
		tokenHexAddress := make([]byte, common.AddressLength)
		err := rows.Scan((*bigint.SQLBigInt)(entry.block), &entry.timestamp, &tokenHexAddress)
		if err != nil {
			return nil, err
		}

		tokenAddress := common.BytesToAddress(tokenHexAddress)
		if tokenAddress != (common.Address{}) {
			entry.tokenAddress = tokenAddress
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (b *BalanceDB) getNewerThan(identity *assetIdentity, timestamp uint64) (entries []*entry, err error) {
	// DISTINCT removes duplicates that can happen when a block has multiple transfers of same token
	rawQueryStr := "SELECT DISTINCT block, timestamp, balance, address FROM balance_history WHERE chain_id = ? AND address IN (%s) AND currency = ? AND timestamp > ? ORDER BY timestamp"
	queryString := fmt.Sprintf(rawQueryStr, identity.addressesToString())
	rows, err := b.db.Query(queryString, identity.ChainID, identity.TokenSymbol, timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]*entry, 0)
	for rows.Next() {
		entry := &entry{
			chainID:     identity.ChainID,
			tokenSymbol: identity.TokenSymbol,
			block:       new(big.Int),
			balance:     new(big.Int),
		}
		err := rows.Scan((*bigint.SQLBigInt)(entry.block), &entry.timestamp, (*bigint.SQLBigIntBytes)(entry.balance), &entry.address)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, nil
}

func (b *BalanceDB) getEntryPreviousTo(item *entry) (res *entry, err error) {
	res = &entry{
		chainID:     item.chainID,
		address:     item.address,
		block:       new(big.Int),
		balance:     new(big.Int),
		tokenSymbol: item.tokenSymbol,
	}

	queryStr := "SELECT block, timestamp, balance FROM balance_history WHERE chain_id = ? AND address = ? AND currency = ? AND timestamp < ? ORDER BY timestamp DESC LIMIT 1"
	row := b.db.QueryRow(queryStr, item.chainID, item.address, item.tokenSymbol, item.timestamp)

	err = row.Scan((*bigint.SQLBigInt)(res.block), &res.timestamp, (*bigint.SQLBigIntBytes)(res.balance))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return res, nil
}
