package accounts

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

var (
	errKeycardDbTransactionIsNil         = errors.New("keycard: database transaction is nil")
	errCannotAddKeycardForUnknownKeypair = errors.New("keycard: cannot add keycard for an unknown keyapir")
	ErrNoKeycardForPassedKeycardUID      = errors.New("keycard: no keycard for the passed keycard uid")
)

type Keycard struct {
	KeycardUID        string          `json:"keycard-uid"`
	KeycardName       string          `json:"keycard-name"`
	KeycardLocked     bool            `json:"keycard-locked"`
	AccountsAddresses []types.Address `json:"accounts-addresses"`
	KeyUID            string          `json:"key-uid"`
	Position          uint64
}

func (kp *Keycard) ToSyncKeycard() *protobuf.SyncKeycard {
	kc := &protobuf.SyncKeycard{
		Uid:      kp.KeycardUID,
		Name:     kp.KeycardName,
		Locked:   kp.KeycardLocked,
		KeyUid:   kp.KeyUID,
		Position: kp.Position,
	}

	for _, addr := range kp.AccountsAddresses {
		kc.Addresses = append(kc.Addresses, addr.Bytes())
	}

	return kc
}

func (kp *Keycard) FromSyncKeycard(kc *protobuf.SyncKeycard) {
	kp.KeycardUID = kc.Uid
	kp.KeycardName = kc.Name
	kp.KeycardLocked = kc.Locked
	kp.KeyUID = kc.KeyUid
	kp.Position = kc.Position

	for _, addr := range kc.Addresses {
		kp.AccountsAddresses = append(kp.AccountsAddresses, types.BytesToAddress(addr))
	}
}

func containsAddress(addresses []types.Address, address types.Address) bool {
	for _, addr := range addresses {
		if addr == address {
			return true
		}
	}
	return false
}

func (db *Database) processResult(rows *sql.Rows) ([]*Keycard, error) {
	keycards := []*Keycard{}
	for rows.Next() {
		keycard := &Keycard{}
		var accAddress sql.NullString
		err := rows.Scan(&keycard.KeycardUID, &keycard.KeycardName, &keycard.KeycardLocked, &accAddress, &keycard.KeyUID,
			&keycard.Position)
		if err != nil {
			return nil, err
		}

		addr := types.Address{}
		if accAddress.Valid {
			addr = types.BytesToAddress([]byte(accAddress.String))
		}

		foundAtIndex := -1
		for i := range keycards {
			if keycards[i].KeycardUID == keycard.KeycardUID {
				foundAtIndex = i
				break
			}
		}
		if foundAtIndex == -1 {
			keycard.AccountsAddresses = append(keycard.AccountsAddresses, addr)
			keycards = append(keycards, keycard)
		} else {
			if containsAddress(keycards[foundAtIndex].AccountsAddresses, addr) {
				continue
			}
			keycards[foundAtIndex].AccountsAddresses = append(keycards[foundAtIndex].AccountsAddresses, addr)
		}
	}

	return keycards, nil
}

func (db *Database) getKeycards(tx *sql.Tx, keyUID string, keycardUID string) ([]*Keycard, error) {
	query := `
		SELECT
			kc.keycard_uid,
			kc.keycard_name,
			kc.keycard_locked,
			ka.account_address,
			kc.key_uid,
			kc.position
		FROM
			keycards AS kc
		LEFT JOIN
			keycards_accounts AS ka
		ON
			kc.keycard_uid = ka.keycard_uid
		LEFT JOIN
			keypairs_accounts AS kpa
		ON
			ka.account_address = kpa.address
		%s
		ORDER BY
			kc.position, kpa.position`

	var where string
	var args []interface{}

	if keyUID != "" {
		where = "WHERE kc.key_uid = ?"
		args = append(args, keyUID)
		if keycardUID != "" {
			where += " AND kc.keycard_uid = ?"
			args = append(args, keycardUID)
		}
	} else if keycardUID != "" {
		where = "WHERE kc.keycard_uid = ?"
		args = append(args, keycardUID)
	}

	query = fmt.Sprintf(query, where)

	var (
		stmt *sql.Stmt
		err  error
	)
	if tx == nil {
		stmt, err = db.db.Prepare(query)
	} else {
		stmt, err = tx.Prepare(query)
	}
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.processResult(rows)
}

func (db *Database) getKeycardByKeycardUID(tx *sql.Tx, keycardUID string) (*Keycard, error) {
	keycards, err := db.getKeycards(tx, "", keycardUID)
	if err != nil {
		return nil, err
	}

	if len(keycards) == 0 {
		return nil, ErrNoKeycardForPassedKeycardUID
	}

	return keycards[0], nil
}

func (db *Database) GetAllKnownKeycards() ([]*Keycard, error) {
	return db.getKeycards(nil, "", "")
}

func (db *Database) GetKeycardsWithSameKeyUID(keyUID string) ([]*Keycard, error) {
	return db.getKeycards(nil, keyUID, "")
}

func (db *Database) GetKeycardByKeycardUID(keycardUID string) (*Keycard, error) {
	return db.getKeycardByKeycardUID(nil, keycardUID)
}

func (db *Database) saveOrUpdateKeycardAccounts(tx *sql.Tx, kcUID string, accountsAddresses []types.Address) (err error) {
	if tx == nil {
		return errKeycardDbTransactionIsNil
	}

	for i := range accountsAddresses {
		addr := accountsAddresses[i]

		_, err = tx.Exec(`
			INSERT OR IGNORE INTO
				keycards_accounts
				(
					keycard_uid,
					account_address
				)
			VALUES
				(?, ?);
			`, kcUID, addr)

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) deleteKeycard(tx *sql.Tx, kcUID string) (err error) {
	if tx == nil {
		return errKeycardDbTransactionIsNil
	}

	delete, err := tx.Prepare(`
		DELETE
		FROM
			keycards
		WHERE
			keycard_uid = ?
	`)
	if err != nil {
		return err
	}
	defer delete.Close()

	_, err = delete.Exec(kcUID)

	return err
}

func (db *Database) deleteAllKeycardsWithKeyUID(tx *sql.Tx, keyUID string) (err error) {
	if tx == nil {
		return errKeycardDbTransactionIsNil
	}

	delete, err := tx.Prepare(`
		DELETE
		FROM
			keycards
		WHERE
			key_uid = ?
	`)
	if err != nil {
		return err
	}
	defer delete.Close()

	_, err = delete.Exec(keyUID)
	return err
}

func (db *Database) deleteKeycardAccounts(tx *sql.Tx, kcUID string, accountAddresses []types.Address) (err error) {
	if tx == nil {
		return errKeycardDbTransactionIsNil
	}

	inVector := strings.Repeat(",?", len(accountAddresses)-1)
	query := `
		DELETE
		FROM
			keycards_accounts
		WHERE
			keycard_uid = ?
		AND
			account_address	IN (?` + inVector + `)`

	delete, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer delete.Close()

	args := make([]interface{}, len(accountAddresses)+1)
	args[0] = kcUID
	for i, addr := range accountAddresses {
		args[i+1] = addr
	}

	_, err = delete.Exec(args...)

	return err
}

func (db *Database) SaveOrUpdateKeycard(keycard Keycard, clock uint64, updateKeypairClock bool) error {
	tx, err := db.db.Begin()
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

	relatedKeypairExists, err := db.keypairExists(tx, keycard.KeyUID)
	if err != nil {
		return err
	}

	if !relatedKeypairExists {
		return errCannotAddKeycardForUnknownKeypair
	}

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO
			keycards
			(
				keycard_uid,
				keycard_name,
				key_uid
			)
		VALUES
			(?, ?, ?);

		UPDATE
			keycards
		SET
			keycard_name = ?,
			keycard_locked = ?,
			position = ?
		WHERE
			keycard_uid = ?;
		`, keycard.KeycardUID, keycard.KeycardName, keycard.KeyUID,
		keycard.KeycardName, keycard.KeycardLocked, keycard.Position, keycard.KeycardUID)
	if err != nil {
		return err
	}

	err = db.saveOrUpdateKeycardAccounts(tx, keycard.KeycardUID, keycard.AccountsAddresses)
	if err != nil {
		return err
	}

	if updateKeypairClock {
		return db.updateKeypairClock(tx, keycard.KeyUID, clock)
	}

	return nil
}

func (db *Database) execKeycardUpdateQuery(kcUID string, clock uint64, field string, value interface{}) (err error) {
	tx, err := db.db.Begin()
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

	keycard, err := db.getKeycardByKeycardUID(tx, kcUID)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`UPDATE keycards SET %s = ? WHERE keycard_uid = ?`, field) // nolint: gosec
	_, err = tx.Exec(sql, value, kcUID)
	if err != nil {
		return err
	}

	return db.updateKeypairClock(tx, keycard.KeyUID, clock)
}

func (db *Database) KeycardLocked(kcUID string, clock uint64) (err error) {
	return db.execKeycardUpdateQuery(kcUID, clock, "keycard_locked", true)
}

func (db *Database) KeycardUnlocked(kcUID string, clock uint64) (err error) {
	return db.execKeycardUpdateQuery(kcUID, clock, "keycard_locked", false)
}

func (db *Database) UpdateKeycardUID(oldKcUID string, newKcUID string, clock uint64) (err error) {
	return db.execKeycardUpdateQuery(oldKcUID, clock, "keycard_uid", newKcUID)
}

func (db *Database) SetKeycardName(kcUID string, kpName string, clock uint64) (err error) {
	return db.execKeycardUpdateQuery(kcUID, clock, "keycard_name", kpName)
}

func (db *Database) DeleteKeycardAccounts(kcUID string, accountAddresses []types.Address, clock uint64) (err error) {
	tx, err := db.db.Begin()
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

	keycard, err := db.getKeycardByKeycardUID(tx, kcUID)
	if err != nil {
		return err
	}

	err = db.deleteKeycardAccounts(tx, kcUID, accountAddresses)
	if err != nil {
		return err
	}

	return db.updateKeypairClock(tx, keycard.KeyUID, clock)
}

func (db *Database) DeleteKeycard(kcUID string, clock uint64) (err error) {
	tx, err := db.db.Begin()
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

	keycard, err := db.getKeycardByKeycardUID(tx, kcUID)
	if err != nil {
		return err
	}

	err = db.deleteKeycard(tx, kcUID)
	if err != nil {
		return err
	}

	return db.updateKeypairClock(tx, keycard.KeyUID, clock)
}

func (db *Database) DeleteAllKeycardsWithKeyUID(keyUID string, clock uint64) (err error) {
	tx, err := db.db.Begin()
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

	err = db.deleteAllKeycardsWithKeyUID(tx, keyUID)
	if err != nil {
		return err
	}

	return db.updateKeypairClock(tx, keyUID, clock)
}

func (db *Database) GetPositionForNextNewKeycard() (uint64, error) {
	var pos sql.NullInt64
	err := db.db.QueryRow("SELECT MAX(position) FROM keycards").Scan(&pos)
	if err != nil {
		return 0, err
	}
	if pos.Valid {
		return uint64(pos.Int64) + 1, nil
	}
	return 0, nil
}
