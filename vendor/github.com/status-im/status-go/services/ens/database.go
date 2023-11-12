package ens

import (
	"database/sql"
)

type Database struct {
	db *sql.DB
}

type UsernameDetail struct {
	Username string `json:"username"`
	ChainID  uint64 `json:"chainId"`
	Clock    uint64 `json:"clock"`
	Removed  bool   `json:"removed"`
}

func NewEnsDatabase(db *sql.DB) *Database {
	return &Database{db: db}
}

func (db *Database) GetEnsUsernames(removed *bool) (result []*UsernameDetail, err error) {

	var sqlQuery = `SELECT username, chain_id, clock, removed  
					  FROM ens_usernames`

	var rows *sql.Rows
	if removed == nil {
		rows, err = db.db.Query(sqlQuery)
	} else {
		sqlQuery += " WHERE removed = ?"
		rows, err = db.db.Query(sqlQuery, removed)
	}

	if err != nil {
		return result, err
	}

	defer rows.Close()

	for rows.Next() {
		var ensUsername UsernameDetail
		err = rows.Scan(&ensUsername.Username, &ensUsername.ChainID, &ensUsername.Clock, &ensUsername.Removed)
		if err != nil {
			return nil, err
		}
		result = append(result, &ensUsername)
	}

	return result, nil
}

func (db *Database) AddEnsUsername(details *UsernameDetail) error {
	const sqlQuery = `INSERT OR REPLACE INTO ens_usernames(username, chain_id, clock, removed)
					  VALUES (?, ?, ?, ?)`
	_, err := db.db.Exec(sqlQuery, details.Username, details.ChainID, details.Clock, details.Removed)
	return err
}

func (db *Database) RemoveEnsUsername(details *UsernameDetail) (bool, error) {
	const sqlQuery = `UPDATE ens_usernames SET removed = 1, clock = ?  
					  WHERE username = (?) AND chain_id = ?`
	result, err := db.db.Exec(sqlQuery, details.Clock, details.Username, details.ChainID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (db *Database) SaveOrUpdateEnsUsername(details *UsernameDetail) error {
	const sqlQuery = `INSERT OR REPLACE INTO ens_usernames (username, chain_id, clock, removed)
SELECT ?, ?, ?, ?
WHERE NOT EXISTS (SELECT 1 FROM ens_usernames WHERE username = ? AND chain_id = ? AND clock >= ?);`

	_, err := db.db.Exec(sqlQuery, details.Username, details.ChainID, details.Clock, details.Removed, details.Username, details.ChainID, details.Clock)
	return err
}
