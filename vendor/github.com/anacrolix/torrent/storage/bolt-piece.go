//go:build !noboltdb && !wasm
// +build !noboltdb,!wasm

package storage

import (
	"encoding/binary"
	"io"

	"go.etcd.io/bbolt"

	"github.com/anacrolix/torrent/metainfo"
)

type boltPiece struct {
	db  *bbolt.DB
	p   metainfo.Piece
	ih  metainfo.Hash
	key [24]byte
}

var (
	_             PieceImpl = (*boltPiece)(nil)
	dataBucketKey           = []byte("data")
)

func (me *boltPiece) pc() PieceCompletionGetSetter {
	return boltPieceCompletion{me.db}
}

func (me *boltPiece) pk() metainfo.PieceKey {
	return metainfo.PieceKey{me.ih, me.p.Index()}
}

func (me *boltPiece) Completion() Completion {
	c, err := me.pc().Get(me.pk())
	switch err {
	case bbolt.ErrDatabaseNotOpen:
		return Completion{}
	case nil:
	default:
		panic(err)
	}
	return c
}

func (me *boltPiece) MarkComplete() error {
	return me.pc().Set(me.pk(), true)
}

func (me *boltPiece) MarkNotComplete() error {
	return me.pc().Set(me.pk(), false)
}

func (me *boltPiece) ReadAt(b []byte, off int64) (n int, err error) {
	err = me.db.View(func(tx *bbolt.Tx) error {
		db := tx.Bucket(dataBucketKey)
		if db == nil {
			return io.EOF
		}
		ci := off / chunkSize
		off %= chunkSize
		for len(b) != 0 {
			ck := me.chunkKey(int(ci))
			_b := db.Get(ck[:])
			// If the chunk is the wrong size, assume it's missing as we can't rely on the data.
			if len(_b) != chunkSize {
				return io.EOF
			}
			n1 := copy(b, _b[off:])
			off = 0
			ci++
			b = b[n1:]
			n += n1
		}
		return nil
	})
	return
}

func (me *boltPiece) chunkKey(index int) (ret [26]byte) {
	copy(ret[:], me.key[:])
	binary.BigEndian.PutUint16(ret[24:], uint16(index))
	return
}

func (me *boltPiece) WriteAt(b []byte, off int64) (n int, err error) {
	err = me.db.Update(func(tx *bbolt.Tx) error {
		db, err := tx.CreateBucketIfNotExists(dataBucketKey)
		if err != nil {
			return err
		}
		ci := off / chunkSize
		off %= chunkSize
		for len(b) != 0 {
			_b := make([]byte, chunkSize)
			ck := me.chunkKey(int(ci))
			copy(_b, db.Get(ck[:]))
			n1 := copy(_b[off:], b)
			db.Put(ck[:], _b)
			if n1 > len(b) {
				break
			}
			b = b[n1:]
			off = 0
			ci++
			n += n1
		}
		return nil
	})
	return
}
