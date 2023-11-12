package rendezvous

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	dbi "github.com/waku-org/go-libp2p-rendezvous/db"
	"go.uber.org/zap"
)

type DB struct {
	db     *sql.DB
	logger *zap.Logger

	insertPeerRegistration     *sql.Stmt
	deletePeerRegistrations    *sql.Stmt
	deletePeerRegistrationsNs  *sql.Stmt
	countPeerRegistrations     *sql.Stmt
	selectPeerRegistrations    *sql.Stmt
	selectPeerRegistrationsNS  *sql.Stmt
	selectPeerRegistrationsC   *sql.Stmt
	selectPeerRegistrationsNSC *sql.Stmt
	deleteExpiredRegistrations *sql.Stmt
	getCounter                 *sql.Stmt

	nonce []byte

	cancel func()
}

func NewDB(db *sql.DB, logger *zap.Logger) *DB {
	rdb := &DB{
		db:     db,
		logger: logger.Named("rendezvous/db"),
	}

	return rdb
}

func (db *DB) Start(ctx context.Context) error {
	err := db.loadNonce()
	if err != nil {
		db.Close()
		return err
	}

	err = db.prepareStmts()
	if err != nil {
		db.Close()
		return err
	}

	bgctx, cancel := context.WithCancel(ctx)
	db.cancel = cancel
	go db.background(bgctx)

	return nil
}

func (db *DB) Close() error {
	db.cancel()
	return db.db.Close()
}

func (db *DB) insertNonce() error {
	nonce := make([]byte, 32)
	_, err := rand.Read(nonce)
	if err != nil {
		return err
	}

	_, err = db.db.Exec("INSERT INTO nonce VALUES (?)", nonce)
	if err != nil {
		return err
	}

	db.nonce = nonce
	return nil
}

func (db *DB) loadNonce() error {
	var nonce []byte
	row := db.db.QueryRow("SELECT nonce FROM nonce")
	err := row.Scan(&nonce)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.insertNonce()
		}
		return err
	}
	db.nonce = nonce
	return nil
}

func (db *DB) prepareStmts() error {
	stmt, err := db.db.Prepare("INSERT INTO registrations VALUES (NULL, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	db.insertPeerRegistration = stmt

	stmt, err = db.db.Prepare("DELETE FROM registrations WHERE peer = ?")
	if err != nil {
		return err
	}
	db.deletePeerRegistrations = stmt

	stmt, err = db.db.Prepare("DELETE FROM registrations WHERE peer = ? AND ns = ?")
	if err != nil {
		return err
	}
	db.deletePeerRegistrationsNs = stmt

	stmt, err = db.db.Prepare("SELECT COUNT(*) FROM registrations WHERE peer = ?")
	if err != nil {
		return err
	}
	db.countPeerRegistrations = stmt

	stmt, err = db.db.Prepare("SELECT * FROM registrations WHERE expire > ? LIMIT ?")
	if err != nil {
		return err
	}
	db.selectPeerRegistrations = stmt

	stmt, err = db.db.Prepare("SELECT * FROM registrations WHERE ns = ? AND expire > ? LIMIT ?")
	if err != nil {
		return err
	}
	db.selectPeerRegistrationsNS = stmt

	stmt, err = db.db.Prepare("SELECT * FROM registrations WHERE counter > ? AND expire > ? LIMIT ?")
	if err != nil {
		return err
	}
	db.selectPeerRegistrationsC = stmt

	stmt, err = db.db.Prepare("SELECT * FROM registrations WHERE counter > ? AND ns = ? AND expire > ? LIMIT ?")
	if err != nil {
		return err
	}
	db.selectPeerRegistrationsNSC = stmt

	stmt, err = db.db.Prepare("DELETE FROM registrations WHERE expire < ?")
	if err != nil {
		return err
	}
	db.deleteExpiredRegistrations = stmt

	stmt, err = db.db.Prepare("SELECT MAX(counter) FROM registrations")
	if err != nil {
		return err
	}
	db.getCounter = stmt

	return nil
}

func (db *DB) Register(p peer.ID, ns string, signedPeerRecord []byte, ttl int) (uint64, error) {
	pid := p.Pretty()
	expire := time.Now().Unix() + int64(ttl)

	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
	}

	delOld := tx.Stmt(db.deletePeerRegistrationsNs)
	insertNew := tx.Stmt(db.insertPeerRegistration)
	getCounter := tx.Stmt(db.getCounter)

	_, err = delOld.Exec(pid, ns)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	_, err = insertNew.Exec(pid, ns, expire, signedPeerRecord)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	var counter uint64
	row := getCounter.QueryRow()
	err = row.Scan(&counter)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	return counter, err
}

func (db *DB) CountRegistrations(p peer.ID) (int, error) {
	pid := p.Pretty()

	row := db.countPeerRegistrations.QueryRow(pid)

	var count int
	err := row.Scan(&count)

	return count, err
}

func (db *DB) Unregister(p peer.ID, ns string) error {
	pid := p.Pretty()

	var err error

	if ns == "" {
		_, err = db.deletePeerRegistrations.Exec(pid)
	} else {
		_, err = db.deletePeerRegistrationsNs.Exec(pid, ns)
	}

	return err
}

func (db *DB) Discover(ns string, cookie []byte, limit int) ([]dbi.RegistrationRecord, []byte, error) {
	now := time.Now().Unix()

	var (
		counter int64
		rows    *sql.Rows
		err     error
	)

	if cookie != nil {
		counter, err = unpackCookie(cookie)
		if err != nil {
			db.logger.Error("unpacking cookie", zap.Error(err))
			return nil, nil, err
		}
	}

	if counter > 0 {
		if ns == "" {
			rows, err = db.selectPeerRegistrationsC.Query(counter, now, limit)
		} else {
			rows, err = db.selectPeerRegistrationsNSC.Query(counter, ns, now, limit)
		}
	} else {
		if ns == "" {
			rows, err = db.selectPeerRegistrations.Query(now, limit)
		} else {
			rows, err = db.selectPeerRegistrationsNS.Query(ns, now, limit)
		}
	}

	if err != nil {
		db.logger.Error("query", zap.Error(err))
		return nil, nil, err
	}

	defer rows.Close()

	regs := make([]dbi.RegistrationRecord, 0, limit)
	for rows.Next() {
		var (
			reg              dbi.RegistrationRecord
			rid              string
			rns              string
			expire           int64
			signedPeerRecord []byte
			p                peer.ID
		)

		err = rows.Scan(&counter, &rid, &rns, &expire, &signedPeerRecord)
		if err != nil {
			db.logger.Error("row scan error", zap.Error(err))
			return nil, nil, err
		}

		p, err = peer.Decode(rid)
		if err != nil {
			db.logger.Error("error decoding peer id", zap.Error(err))
			continue
		}

		reg.Id = p
		reg.SignedPeerRecord = signedPeerRecord
		reg.Ttl = int(expire - now)

		if ns == "" {
			reg.Ns = rns
		}

		regs = append(regs, reg)
	}

	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	if counter > 0 {
		cookie = packCookie(counter, ns, db.nonce)
	}

	return regs, cookie, nil
}

func (db *DB) ValidCookie(ns string, cookie []byte) bool {
	return validCookie(cookie, ns, db.nonce)
}

func (db *DB) background(ctx context.Context) {
	for {
		db.cleanupExpired()

		select {
		case <-time.After(15 * time.Minute):
		case <-ctx.Done():
			return
		}
	}
}

func (db *DB) cleanupExpired() {
	now := time.Now().Unix()
	_, err := db.deleteExpiredRegistrations.Exec(now)
	if err != nil {
		db.logger.Error("deleting expired registrations", zap.Error(err))
	}
}

// cookie: counter:SHA256(nonce + ns + counter)
func packCookie(counter int64, ns string, nonce []byte) []byte {
	cbits := make([]byte, 8)
	binary.BigEndian.PutUint64(cbits, uint64(counter))

	hash := sha256.New()
	_, err := hash.Write(nonce)
	if err != nil {
		panic(err)
	}
	_, err = hash.Write([]byte(ns))
	if err != nil {
		panic(err)
	}
	_, err = hash.Write(cbits)
	if err != nil {
		panic(err)
	}

	return hash.Sum(cbits)
}

func unpackCookie(cookie []byte) (int64, error) {
	if len(cookie) < 8 {
		return 0, fmt.Errorf("bad packed cookie: not enough bytes: %v", cookie)
	}

	counter := binary.BigEndian.Uint64(cookie[:8])
	return int64(counter), nil
}

func validCookie(cookie []byte, ns string, nonce []byte) bool {
	if len(cookie) != 40 {
		return false
	}

	cbits := cookie[:8]
	hash := sha256.New()
	_, err := hash.Write(nonce)
	if err != nil {
		panic(err)
	}
	_, err = hash.Write([]byte(ns))
	if err != nil {
		panic(err)
	}
	_, err = hash.Write(cbits)
	if err != nil {
		panic(err)
	}
	hbits := hash.Sum(nil)

	return bytes.Equal(cookie[8:], hbits)
}
