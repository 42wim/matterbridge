// modernc.org/sqlite depends on modernc.org/libc which doesn't work for JS (and probably wasm but I
// think JS is the stronger signal).
//go:build !js && !nosqlite
// +build !js,!nosqlite

package storage

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/anacrolix/torrent/metainfo"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// sqlite is always the default when available.
func NewDefaultPieceCompletionForDir(dir string) (PieceCompletion, error) {
	return NewSqlitePieceCompletion(dir)
}

type sqlitePieceCompletion struct {
	mu     sync.Mutex
	closed bool
	db     *sqlite.Conn
}

var _ PieceCompletion = (*sqlitePieceCompletion)(nil)

func NewSqlitePieceCompletion(dir string) (ret *sqlitePieceCompletion, err error) {
	p := filepath.Join(dir, ".torrent.db")
	db, err := sqlite.OpenConn(p, 0)
	if err != nil {
		return
	}
	err = sqlitex.ExecScript(db, `create table if not exists piece_completion(infohash, "index", complete, unique(infohash, "index"))`)
	if err != nil {
		db.Close()
		return
	}
	ret = &sqlitePieceCompletion{db: db}
	return
}

func (me *sqlitePieceCompletion) Get(pk metainfo.PieceKey) (c Completion, err error) {
	me.mu.Lock()
	defer me.mu.Unlock()
	err = sqlitex.Exec(
		me.db, `select complete from piece_completion where infohash=? and "index"=?`,
		func(stmt *sqlite.Stmt) error {
			c.Complete = stmt.ColumnInt(0) != 0
			c.Ok = true
			return nil
		},
		pk.InfoHash.HexString(), pk.Index)
	return
}

func (me *sqlitePieceCompletion) Set(pk metainfo.PieceKey, b bool) error {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.closed {
		return errors.New("closed")
	}
	return sqlitex.Exec(
		me.db,
		`insert or replace into piece_completion(infohash, "index", complete) values(?, ?, ?)`,
		nil,
		pk.InfoHash.HexString(), pk.Index, b)
}

func (me *sqlitePieceCompletion) Close() (err error) {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.closed {
		return
	}
	err = me.db.Close()
	me.db = nil
	me.closed = true
	return
}
