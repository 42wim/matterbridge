package collectibles

import (
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/jmoiron/sqlx"

	"github.com/status-im/status-go/services/wallet/bigint"
	w_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/sqlite"
)

const InvalidTimestamp = int64(-1)

type OwnershipDB struct {
	db *sql.DB
	mu sync.Mutex
}

func NewOwnershipDB(sqlDb *sql.DB) *OwnershipDB {
	return &OwnershipDB{
		db: sqlDb,
	}
}

const unknownUpdateTimestamp = int64(math.MaxInt64)

const selectOwnershipColumns = "chain_id, contract_address, token_id"

const ownershipTimestampColumns = "owner_address, chain_id, timestamp"
const selectOwnershipTimestampColumns = "timestamp"

func insertTmpOwnership(
	db *sql.DB,
	chainID w_common.ChainID,
	ownerAddress common.Address,
	balancesPerContractAdddress thirdparty.TokenBalancesPerContractAddress,
) error {
	// Put old/new ownership data into temp tables
	// NOTE: Temp table CREATE doesn't work with prepared statements,
	// so we have to use Exec directly
	_, err := db.Exec(`
		DROP TABLE IF EXISTS temp.old_collectibles_ownership_cache; 
		CREATE TABLE temp.old_collectibles_ownership_cache(
			contract_address VARCHAR NOT NULL,
			token_id BLOB NOT NULL,
			balance BLOB NOT NULL
		);
		DROP TABLE IF EXISTS temp.new_collectibles_ownership_cache; 
		CREATE TABLE temp.new_collectibles_ownership_cache(
			contract_address VARCHAR NOT NULL,
			token_id BLOB NOT NULL,
			balance BLOB NOT NULL
		);`)
	if err != nil {
		return err
	}

	insertTmpOldOwnership, err := db.Prepare(`
			INSERT INTO temp.old_collectibles_ownership_cache
			SELECT contract_address, token_id, balance FROM collectibles_ownership_cache
			WHERE chain_id = ? AND owner_address = ?`)
	if err != nil {
		return err
	}
	defer insertTmpOldOwnership.Close()

	_, err = insertTmpOldOwnership.Exec(chainID, ownerAddress)
	if err != nil {
		return err
	}

	insertTmpNewOwnership, err := db.Prepare(`
			INSERT INTO temp.new_collectibles_ownership_cache (contract_address, token_id, balance) 
			VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer insertTmpNewOwnership.Close()

	for contractAddress, balances := range balancesPerContractAdddress {
		for _, balance := range balances {
			_, err = insertTmpNewOwnership.Exec(
				contractAddress,
				(*bigint.SQLBigIntBytes)(balance.TokenID.Int),
				(*bigint.SQLBigIntBytes)(balance.Balance.Int),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func removeOldAddressOwnership(
	creator sqlite.StatementCreator,
	chainID w_common.ChainID,
	ownerAddress common.Address,
) ([]thirdparty.CollectibleUniqueID, error) {
	// Find collectibles in the DB that are not in the temp table
	removedQuery, err := creator.Prepare(fmt.Sprintf(`
	SELECT %d, tOld.contract_address, tOld.token_id 
		FROM temp.old_collectibles_ownership_cache tOld
		LEFT JOIN temp.new_collectibles_ownership_cache tNew ON
			tOld.contract_address = tNew.contract_address AND tOld.token_id = tNew.token_id
		WHERE 
			tNew.contract_address IS NULL
	`, chainID))
	if err != nil {
		return nil, err
	}
	defer removedQuery.Close()

	removedRows, err := removedQuery.Query()
	if err != nil {
		return nil, err
	}

	defer removedRows.Close()
	removedIDs, err := thirdparty.RowsToCollectibles(removedRows)
	if err != nil {
		return nil, err
	}

	removeOwnership, err := creator.Prepare("DELETE FROM collectibles_ownership_cache WHERE chain_id = ? AND owner_address = ? AND contract_address = ? AND token_id = ?")
	if err != nil {
		return nil, err
	}
	defer removeOwnership.Close()

	for _, id := range removedIDs {
		_, err = removeOwnership.Exec(
			chainID,
			ownerAddress,
			id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int),
		)
		if err != nil {
			return nil, err
		}
	}

	return removedIDs, nil
}

func updateChangedAddressOwnership(
	creator sqlite.StatementCreator,
	chainID w_common.ChainID,
	ownerAddress common.Address,
) ([]thirdparty.CollectibleUniqueID, error) {
	// Find collectibles in the temp table that are in the DB and have a different balance
	updatedQuery, err := creator.Prepare(fmt.Sprintf(`
		SELECT %d, tNew.contract_address, tNew.token_id 
		FROM temp.new_collectibles_ownership_cache tNew
		LEFT JOIN temp.old_collectibles_ownership_cache tOld ON
			tOld.contract_address = tNew.contract_address AND tOld.token_id = tNew.token_id
		WHERE 
			tOld.contract_address IS NOT NULL AND tOld.balance != tNew.balance
	`, chainID))
	if err != nil {
		return nil, err
	}
	defer updatedQuery.Close()

	updatedRows, err := updatedQuery.Query()
	if err != nil {
		return nil, err
	}
	defer updatedRows.Close()

	updatedIDs, err := thirdparty.RowsToCollectibles(updatedRows)
	if err != nil {
		return nil, err
	}

	updateOwnership, err := creator.Prepare(`
		UPDATE collectibles_ownership_cache
		SET balance = (SELECT tNew.balance
			FROM temp.new_collectibles_ownership_cache tNew
			WHERE tNew.contract_address = collectibles_ownership_cache.contract_address AND tNew.token_id = collectibles_ownership_cache.token_id)
		WHERE chain_id = ? AND owner_address = ? AND contract_address = ? AND token_id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer updateOwnership.Close()

	for _, id := range updatedIDs {
		_, err = updateOwnership.Exec(
			chainID,
			ownerAddress,
			id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int))
		if err != nil {
			return nil, err
		}
	}

	return updatedIDs, nil
}

func insertNewAddressOwnership(
	creator sqlite.StatementCreator,
	chainID w_common.ChainID,
	ownerAddress common.Address,
) ([]thirdparty.CollectibleUniqueID, error) {
	// Find collectibles in the temp table that are not in the DB
	insertedQuery, err := creator.Prepare(fmt.Sprintf(`
		SELECT %d, tNew.contract_address, tNew.token_id 
		FROM temp.new_collectibles_ownership_cache tNew
		LEFT JOIN temp.old_collectibles_ownership_cache tOld ON
			tOld.contract_address = tNew.contract_address AND tOld.token_id = tNew.token_id
		WHERE 
			tOld.contract_address IS NULL
	`, chainID))
	if err != nil {
		return nil, err
	}
	defer insertedQuery.Close()

	insertedRows, err := insertedQuery.Query()
	if err != nil {
		return nil, err
	}
	defer insertedRows.Close()

	insertedIDs, err := thirdparty.RowsToCollectibles(insertedRows)
	if err != nil {
		return nil, err
	}

	insertOwnership, err := creator.Prepare(fmt.Sprintf(`
		INSERT INTO collectibles_ownership_cache
		SELECT
			%d, tNew.contract_address, tNew.token_id, X'%s', tNew.balance, NULL
		FROM temp.new_collectibles_ownership_cache tNew
		WHERE
			tNew.contract_address = ? AND tNew.token_id = ?
	`, chainID, ownerAddress.Hex()[2:]))
	if err != nil {
		return nil, err
	}
	defer insertOwnership.Close()

	for _, id := range insertedIDs {
		_, err = insertOwnership.Exec(
			id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int))
		if err != nil {
			return nil, err
		}
	}

	return insertedIDs, nil
}

func updateAddressOwnership(
	tx sqlite.StatementCreator,
	chainID w_common.ChainID,
	ownerAddress common.Address,
) (removedIDs, updatedIDs, insertedIDs []thirdparty.CollectibleUniqueID, err error) {
	removedIDs, err = removeOldAddressOwnership(tx, chainID, ownerAddress)
	if err != nil {
		return
	}

	updatedIDs, err = updateChangedAddressOwnership(tx, chainID, ownerAddress)
	if err != nil {
		return
	}

	insertedIDs, err = insertNewAddressOwnership(tx, chainID, ownerAddress)
	if err != nil {
		return
	}

	return
}

func updateAddressOwnershipTimestamp(creator sqlite.StatementCreator, ownerAddress common.Address, chainID w_common.ChainID, timestamp int64) error {
	updateTimestamp, err := creator.Prepare(fmt.Sprintf(`INSERT OR REPLACE INTO collectibles_ownership_update_timestamps (%s) 
																				VALUES (?, ?, ?)`, ownershipTimestampColumns))
	if err != nil {
		return err
	}
	defer updateTimestamp.Close()

	_, err = updateTimestamp.Exec(ownerAddress, chainID, timestamp)

	return err
}

// Returns the list of added/removed IDs when comparing the given list of IDs with the ones in the DB.
// Call before Update for the result to be useful.
func (o *OwnershipDB) GetIDsNotInDB(
	ownerAddress common.Address,
	newIDs []thirdparty.CollectibleUniqueID) ([]thirdparty.CollectibleUniqueID, error) {
	ret := make([]thirdparty.CollectibleUniqueID, 0, len(newIDs))

	exists, err := o.db.Prepare(`SELECT EXISTS (
			SELECT 1 FROM collectibles_ownership_cache
			WHERE chain_id=? AND contract_address=? AND token_id=? AND owner_address=?
		)`)
	if err != nil {
		return nil, err
	}

	for _, id := range newIDs {
		row := exists.QueryRow(
			id.ContractID.ChainID,
			id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int),
			ownerAddress,
		)
		var exists bool
		err = row.Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			ret = append(ret, id)
		}
	}

	return ret, nil
}

func (o *OwnershipDB) GetIsFirstOfCollection(onwerAddress common.Address, newIDs []thirdparty.CollectibleUniqueID) (map[thirdparty.CollectibleUniqueID]bool, error) {
	ret := make(map[thirdparty.CollectibleUniqueID]bool)

	exists, err := o.db.Prepare(`SELECT count(*) FROM collectibles_ownership_cache
			WHERE chain_id=? AND contract_address=? AND owner_address=?`)
	if err != nil {
		return nil, err
	}

	for _, id := range newIDs {
		row := exists.QueryRow(
			id.ContractID.ChainID,
			id.ContractID.Address,
			onwerAddress,
		)
		var count int
		err = row.Scan(&count)
		if err != nil {
			return nil, err
		}
		ret[id] = count <= 1
	}
	return ret, nil
}

func (o *OwnershipDB) Update(chainID w_common.ChainID, ownerAddress common.Address, balances thirdparty.TokenBalancesPerContractAddress, timestamp int64) (removedIDs, updatedIDs, insertedIDs []thirdparty.CollectibleUniqueID, err error) {
	// Ensure all steps are done atomically
	o.mu.Lock()
	defer o.mu.Unlock()

	err = insertTmpOwnership(o.db, chainID, ownerAddress, balances)
	if err != nil {
		return
	}

	var (
		tx *sql.Tx
	)
	tx, err = o.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	// Compare tmp and current ownership tables and update the current one
	removedIDs, updatedIDs, insertedIDs, err = updateAddressOwnership(tx, chainID, ownerAddress)
	if err != nil {
		return
	}

	// Update timestamp
	err = updateAddressOwnershipTimestamp(tx, ownerAddress, chainID, timestamp)

	return
}

func (o *OwnershipDB) GetOwnedCollectibles(chainIDs []w_common.ChainID, ownerAddresses []common.Address, offset int, limit int) ([]thirdparty.CollectibleUniqueID, error) {
	query, args, err := sqlx.In(fmt.Sprintf(`SELECT DISTINCT %s
		FROM collectibles_ownership_cache
		WHERE chain_id IN (?) AND owner_address IN (?)
		LIMIT ? OFFSET ?`, selectOwnershipColumns), chainIDs, ownerAddresses, limit, offset)
	if err != nil {
		return nil, err
	}

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return thirdparty.RowsToCollectibles(rows)
}

func (o *OwnershipDB) GetOwnedCollectible(chainID w_common.ChainID, ownerAddresses common.Address, contractAddress common.Address, tokenID *big.Int) (*thirdparty.CollectibleUniqueID, error) {
	query := fmt.Sprintf(`SELECT %s
		FROM collectibles_ownership_cache
		WHERE chain_id = ? AND owner_address = ? AND contract_address = ? AND token_id = ?`, selectOwnershipColumns)

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(chainID, ownerAddresses, contractAddress, (*bigint.SQLBigIntBytes)(tokenID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids, err := thirdparty.RowsToCollectibles(rows)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return &ids[0], nil
}

func (o *OwnershipDB) GetOwnershipUpdateTimestamp(owner common.Address, chainID w_common.ChainID) (int64, error) {
	query := fmt.Sprintf(`SELECT %s
		FROM collectibles_ownership_update_timestamps
		WHERE owner_address = ? AND chain_id = ?`, selectOwnershipTimestampColumns)

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return InvalidTimestamp, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(owner, chainID)

	var timestamp int64

	err = row.Scan(&timestamp)

	if err == sql.ErrNoRows {
		return InvalidTimestamp, nil
	} else if err != nil {
		return InvalidTimestamp, err
	}

	return timestamp, nil
}

func (o *OwnershipDB) GetLatestOwnershipUpdateTimestamp(chainID w_common.ChainID) (int64, error) {
	query := `SELECT MAX(timestamp)
		FROM collectibles_ownership_update_timestamps
		WHERE chain_id = ?`

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return InvalidTimestamp, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(chainID)

	var timestamp sql.NullInt64

	err = row.Scan(&timestamp)

	if err != nil {
		return InvalidTimestamp, err
	}
	if timestamp.Valid {
		return timestamp.Int64, nil
	}

	return InvalidTimestamp, nil
}

func (o *OwnershipDB) GetOwnership(id thirdparty.CollectibleUniqueID) ([]thirdparty.AccountBalance, error) {
	query := fmt.Sprintf(`SELECT c.owner_address, c.balance, COALESCE(t.timestamp, %d)
		FROM collectibles_ownership_cache c
		LEFT JOIN transfers t ON
			c.transfer_id = t.hash
		WHERE
		c.chain_id = ? AND c.contract_address = ? AND c.token_id = ?`, unknownUpdateTimestamp)

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id.ContractID.ChainID, id.ContractID.Address, (*bigint.SQLBigIntBytes)(id.TokenID.Int))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ret []thirdparty.AccountBalance
	for rows.Next() {
		accountBalance := thirdparty.AccountBalance{
			Balance: &bigint.BigInt{Int: big.NewInt(0)},
		}
		err = rows.Scan(
			&accountBalance.Address,
			(*bigint.SQLBigIntBytes)(accountBalance.Balance.Int),
			&accountBalance.TxTimestamp,
		)
		if err != nil {
			return nil, err
		}

		ret = append(ret, accountBalance)
	}

	return ret, nil
}

func (o *OwnershipDB) SetTransferID(ownerAddress common.Address, id thirdparty.CollectibleUniqueID, transferID common.Hash) (bool, error) {
	query := `UPDATE collectibles_ownership_cache
		SET transfer_id = ?
		WHERE chain_id = ? AND contract_address = ? AND token_id = ? AND owner_address = ?`

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(transferID, id.ContractID.ChainID, id.ContractID.Address, (*bigint.SQLBigIntBytes)(id.TokenID.Int), ownerAddress)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected > 0 {
		return true, nil
	}

	return false, nil
}

func (o *OwnershipDB) GetTransferID(ownerAddress common.Address, id thirdparty.CollectibleUniqueID) (*common.Hash, error) {
	query := `SELECT transfer_id
		FROM collectibles_ownership_cache
		WHERE chain_id = ? AND contract_address = ? AND token_id = ? AND owner_address = ?
		LIMIT 1`

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(id.ContractID.ChainID, id.ContractID.Address, (*bigint.SQLBigIntBytes)(id.TokenID.Int), ownerAddress)

	var dbTransferID []byte

	err = row.Scan(&dbTransferID)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(dbTransferID) > 0 {
		transferID := common.BytesToHash(dbTransferID)
		return &transferID, nil
	}

	return nil, nil
}

func (o *OwnershipDB) GetCollectiblesWithNoTransferID(account common.Address, chainID w_common.ChainID) ([]thirdparty.CollectibleUniqueID, error) {
	query := `SELECT contract_address, token_id
		FROM collectibles_ownership_cache
		WHERE chain_id = ? AND owner_address = ? AND transfer_id IS NULL`

	stmt, err := o.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(chainID, account)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ret []thirdparty.CollectibleUniqueID
	for rows.Next() {
		id := thirdparty.CollectibleUniqueID{
			ContractID: thirdparty.ContractID{
				ChainID: chainID,
			},
			TokenID: &bigint.BigInt{Int: big.NewInt(0)},
		}
		err = rows.Scan(
			&id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int),
		)
		if err != nil {
			return nil, err
		}

		ret = append(ret, id)
	}

	return ret, nil
}
