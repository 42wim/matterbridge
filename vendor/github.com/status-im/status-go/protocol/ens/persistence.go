package ens

import (
	"context"
	"database/sql"
	"errors"
)

type Persistence struct {
	db *sql.DB
}

func NewPersistence(db *sql.DB) *Persistence {
	return &Persistence{db: db}
}

func (p *Persistence) GetENSToBeVerified(now uint64) ([]*VerificationRecord, error) {
	rows, err := p.db.Query(`SELECT public_key, name, verified, verified_at, clock, verification_retries, next_retry FROM ens_verification_records WHERE NOT(verified) AND verification_retries < ? AND next_retry <= ?`, maxRetries, now)
	if err != nil {
		return nil, err
	}

	var records []*VerificationRecord
	for rows.Next() {
		var record VerificationRecord
		err := rows.Scan(&record.PublicKey, &record.Name, &record.Verified, &record.VerifiedAt, &record.Clock, &record.VerificationRetries, &record.NextRetry)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

func (p *Persistence) UpdateRecords(records []*VerificationRecord) (err error) {
	var tx *sql.Tx
	tx, err = p.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	for _, record := range records {
		var stmt *sql.Stmt
		stmt, err = tx.Prepare(`UPDATE ens_verification_records SET verified = ?, verified_at = ?, verification_retries = ?, next_retry = ? WHERE public_key = ?`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(record.Verified, record.VerifiedAt, record.VerificationRetries, record.NextRetry, record.PublicKey)
		if err != nil {
			return err
		}

	}

	return nil
}

// AddRecord adds a record or return the latest available if already in the database and
// hasn't changed
func (p *Persistence) AddRecord(record VerificationRecord) (response *VerificationRecord, err error) {
	if !record.Valid() {
		err = errors.New("invalid ens record")
		return
	}
	var tx *sql.Tx
	tx, err = p.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	dbRecord := &VerificationRecord{PublicKey: record.PublicKey}

	err = tx.QueryRow(`SELECT name, clock, verified FROM ens_verification_records WHERE public_key = ?`, record.PublicKey).Scan(&dbRecord.Name, &dbRecord.Clock, &dbRecord.Verified)
	if err != nil && err != sql.ErrNoRows {
		return
	}

	if dbRecord.Clock >= record.Clock || dbRecord.Name == record.Name {
		response = dbRecord
		return
	}

	_, err = tx.Exec(`INSERT INTO ens_verification_records(public_key, name, clock) VALUES (?,?,?)`, record.PublicKey, record.Name, record.Clock)
	return
}

func (p *Persistence) GetVerifiedRecord(publicKey string) (*VerificationRecord, error) {
	record := &VerificationRecord{}
	err := p.db.QueryRow(`SELECT name, clock FROM ens_verification_records WHERE verified AND public_key =  ?`, publicKey).Scan(&record.Name, &record.Clock)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return record, nil
	}

	return nil, err

}
