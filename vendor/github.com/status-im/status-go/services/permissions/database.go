package permissions

import (
	"database/sql"
)

// Database sql wrapper for operations with browser objects.
type Database struct {
	db *sql.DB
}

// Close closes database.
func (db Database) Close() error {
	return db.db.Close()
}

func NewDB(db *sql.DB) *Database {
	return &Database{db: db}
}

type DappPermissions struct {
	ID          int
	Name        string   `json:"dapp"`
	Permissions []string `json:"permissions,omitempty"`
	Address     string   `json:"address,omitempty"`
}

func (db *Database) AddPermissions(perms DappPermissions) (err error) {
	tx, err := db.db.Begin()
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
	dRows, err := tx.Query("SELECT id FROM dapps where name = ? AND address = ?", perms.Name, perms.Address)
	if err != nil {
		return
	}
	defer dRows.Close()

	var id int64
	if dRows.Next() {
		err = dRows.Scan(&id)
		if err != nil {
			return
		}
	} else {
		dInsert, err := tx.Prepare("INSERT INTO dapps(name, address) VALUES(?, ?)")
		if err != nil {
			return err
		}
		res, err := dInsert.Exec(perms.Name, perms.Address)
		dInsert.Close()
		if err != nil {
			return err
		}

		id, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}

	pDelete, err := tx.Prepare("DELETE FROM permissions WHERE dapp_id = ?")
	if err != nil {
		return
	}
	defer pDelete.Close()
	_, err = pDelete.Exec(id)
	if err != nil {
		return
	}

	if len(perms.Permissions) == 0 {
		return
	}

	pInsert, err := tx.Prepare("INSERT INTO permissions(dapp_id, permission) VALUES(?, ?)")
	if err != nil {
		return
	}
	defer pInsert.Close()
	for _, perm := range perms.Permissions {
		_, err = pInsert.Exec(id, perm)
		if err != nil {
			return
		}
	}
	return
}

func (db *Database) GetPermissions() (rst []DappPermissions, err error) {
	tx, err := db.db.Begin()
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

	// FULL and RIGHT joins are not supported
	dRows, err := tx.Query("SELECT id, name, address FROM dapps")
	if err != nil {
		return
	}
	defer dRows.Close()
	dapps := map[int]*DappPermissions{}
	for dRows.Next() {
		perms := DappPermissions{}
		err = dRows.Scan(&perms.ID, &perms.Name, &perms.Address)
		if err != nil {
			return nil, err
		}
		dapps[perms.ID] = &perms
	}

	pRows, err := tx.Query("SELECT dapp_id, permission from permissions")
	if err != nil {
		return
	}
	defer pRows.Close()
	var (
		id         int
		permission string
	)
	for pRows.Next() {
		err = pRows.Scan(&id, &permission)
		if err != nil {
			return
		}
		dapps[id].Permissions = append(dapps[id].Permissions, permission)
	}
	rst = make([]DappPermissions, 0, len(dapps))
	for key := range dapps {
		rst = append(rst, *dapps[key])
	}

	return rst, nil
}

func (db *Database) DeletePermission(name string, address string) error {
	_, err := db.db.Exec("DELETE FROM dapps WHERE name = ? AND address = ?", name, address)
	return err
}

func (db *Database) HasPermission(dappName string, address string, permission string) (bool, error) {
	var id int64
	err := db.db.QueryRow("SELECT id FROM dapps where name = ? AND address = ?", dappName, address).Scan(&id)
	if err != nil {
		return false, nil
	}

	var count uint64
	err = db.db.QueryRow(
		`SELECT COUNT(1) FROM permissions WHERE dapp_id = ? AND permission = ?`,
		id, permission,
	).Scan(&count)
	return count > 0, err
}
