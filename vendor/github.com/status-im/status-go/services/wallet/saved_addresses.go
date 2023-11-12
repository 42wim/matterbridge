package wallet

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	multiAccCommon "github.com/status-im/status-go/multiaccounts/common"
)

type savedAddressMeta struct {
	UpdateClock uint64 // wall clock used to deconflict concurrent updates
}

type SavedAddress struct {
	Address common.Address `json:"address"`
	// TODO: Add Emoji
	// Emoji	string		   `json:"emoji"`
	Name            string                            `json:"name"`
	ChainShortNames string                            `json:"chainShortNames"` // used with address only, not with ENSName
	ENSName         string                            `json:"ens"`
	ColorID         multiAccCommon.CustomizationColor `json:"colorId"`
	IsTest          bool                              `json:"isTest"`
	CreatedAt       int64                             `json:"createdAt"`
	Removed         bool                              `json:"removed"`
	savedAddressMeta
}

func (s *SavedAddress) ID() string {
	return fmt.Sprintf("%s-%t", s.Address.Hex(), s.IsTest)
}

type SavedAddressesManager struct {
	db *sql.DB
}

func NewSavedAddressesManager(db *sql.DB) *SavedAddressesManager {
	return &SavedAddressesManager{db: db}
}

const rawQueryColumnsOrder = "address, name, removed, update_clock, chain_short_names, ens_name, is_test, created_at, color"

// getSavedAddressesFromDBRows retrieves all data based on SELECT Query using rawQueryColumnsOrder
func getSavedAddressesFromDBRows(rows *sql.Rows) ([]*SavedAddress, error) {
	var addresses []*SavedAddress
	for rows.Next() {
		sa := &SavedAddress{}
		// based on rawQueryColumnsOrder
		err := rows.Scan(
			&sa.Address,
			&sa.Name,
			&sa.Removed,
			&sa.UpdateClock,
			&sa.ChainShortNames,
			&sa.ENSName,
			&sa.IsTest,
			&sa.CreatedAt,
			&sa.ColorID,
		)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, sa)
	}

	return addresses, nil
}

func (sam *SavedAddressesManager) getSavedAddresses(condition string) ([]*SavedAddress, error) {
	var whereCondition string
	if condition != "" {
		whereCondition = fmt.Sprintf("WHERE %s", condition)
	}

	rows, err := sam.db.Query(fmt.Sprintf("SELECT %s FROM saved_addresses %s", rawQueryColumnsOrder, whereCondition))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses, err := getSavedAddressesFromDBRows(rows)
	return addresses, err
}

func (sam *SavedAddressesManager) GetSavedAddresses() ([]*SavedAddress, error) {
	return sam.getSavedAddresses("removed != 1")
}

// GetRawSavedAddresses provides access to the soft-delete and sync metadata
func (sam *SavedAddressesManager) GetRawSavedAddresses() ([]*SavedAddress, error) {
	return sam.getSavedAddresses("")
}

func (sam *SavedAddressesManager) upsertSavedAddress(sa SavedAddress, tx *sql.Tx) (err error) {
	if tx == nil {
		tx, err = sam.db.Begin()
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
	}
	rows, err := tx.Query(
		fmt.Sprintf("SELECT %s FROM saved_addresses WHERE address = ? AND is_test = ?", rawQueryColumnsOrder),
		sa.Address, sa.IsTest,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	savedAddresses, err := getSavedAddressesFromDBRows(rows)
	if err != nil {
		return err
	}
	sa.CreatedAt = time.Now().Unix()
	for _, savedAddress := range savedAddresses {
		if savedAddress.Address == sa.Address && savedAddress.IsTest == sa.IsTest {
			sa.CreatedAt = savedAddress.CreatedAt
			break
		}
	}
	sqlStatement := `
	INSERT OR REPLACE
	INTO
		saved_addresses (
			address,
			name,
			removed,
			update_clock,
			chain_short_names,
			ens_name,
			is_test,
			created_at,
			color
		)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	insert, err := tx.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer insert.Close()
	_, err = insert.Exec(sa.Address, sa.Name, sa.Removed, sa.UpdateClock, sa.ChainShortNames, sa.ENSName,
		sa.IsTest, sa.CreatedAt, sa.ColorID)
	return err
}

func (sam *SavedAddressesManager) UpdateMetadataAndUpsertSavedAddress(sa SavedAddress) error {
	return sam.upsertSavedAddress(sa, nil)
}

func (sam *SavedAddressesManager) startTransactionAndCheckIfNewerChange(address common.Address, isTest bool, updateClock uint64) (newer bool, tx *sql.Tx, err error) {
	tx, err = sam.db.Begin()
	if err != nil {
		return false, nil, err
	}
	row := tx.QueryRow("SELECT update_clock FROM saved_addresses WHERE address = ? AND is_test = ?", address, isTest)
	if err != nil {
		return false, tx, err
	}

	var dbUpdateClock uint64
	err = row.Scan(&dbUpdateClock)
	if err != nil {
		return err == sql.ErrNoRows, tx, err
	}
	return dbUpdateClock < updateClock, tx, nil
}

func (sam *SavedAddressesManager) AddSavedAddressIfNewerUpdate(sa SavedAddress) (insertedOrUpdated bool, err error) {
	newer, tx, err := sam.startTransactionAndCheckIfNewerChange(sa.Address, sa.IsTest, sa.UpdateClock)
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()
	if !newer {
		return false, err
	}

	err = sam.upsertSavedAddress(sa, tx)
	if err != nil {
		return false, err
	}

	return true, err
}

func (sam *SavedAddressesManager) DeleteSavedAddress(address common.Address, isTest bool, updateClock uint64) (deleted bool, err error) {
	if err != nil {
		return false, err
	}
	newer, tx, err := sam.startTransactionAndCheckIfNewerChange(address, isTest, updateClock)
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()
	if !newer {
		return false, err
	}

	update, err := tx.Prepare(`UPDATE saved_addresses SET removed = 1, update_clock = ? WHERE address = ? AND is_test = ?`)
	if err != nil {
		return false, err
	}
	defer update.Close()
	res, err := update.Exec(updateClock, address, isTest)
	if err != nil {
		return false, err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return nRows > 0, nil
}

func (sam *SavedAddressesManager) DeleteSoftRemovedSavedAddresses(threshold uint64) error {
	_, err := sam.db.Exec(`DELETE FROM saved_addresses WHERE removed = 1 AND update_clock < ?`, threshold)
	return err
}
