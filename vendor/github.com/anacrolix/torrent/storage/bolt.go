//go:build !noboltdb && !wasm
// +build !noboltdb,!wasm

package storage

import (
	"encoding/binary"
	"path/filepath"
	"time"

	"github.com/anacrolix/missinggo/expect"
	"go.etcd.io/bbolt"

	"github.com/anacrolix/torrent/metainfo"
)

const (
	// Chosen to match the usual chunk size in a torrent client. This way, most chunk writes are to
	// exactly one full item in bbolt DB.
	chunkSize = 1 << 14
)

type boltClient struct {
	db *bbolt.DB
}

type boltTorrent struct {
	cl *boltClient
	ih metainfo.Hash
}

func NewBoltDB(filePath string) ClientImplCloser {
	db, err := bbolt.Open(filepath.Join(filePath, "bolt.db"), 0o600, &bbolt.Options{
		Timeout: time.Second,
	})
	expect.Nil(err)
	db.NoSync = true
	return &boltClient{db}
}

func (me *boltClient) Close() error {
	return me.db.Close()
}

func (me *boltClient) OpenTorrent(_ *metainfo.Info, infoHash metainfo.Hash) (TorrentImpl, error) {
	t := &boltTorrent{me, infoHash}
	return TorrentImpl{
		Piece: t.Piece,
		Close: t.Close,
	}, nil
}

func (me *boltTorrent) Piece(p metainfo.Piece) PieceImpl {
	ret := &boltPiece{
		p:  p,
		db: me.cl.db,
		ih: me.ih,
	}
	copy(ret.key[:], me.ih[:])
	binary.BigEndian.PutUint32(ret.key[20:], uint32(p.Index()))
	return ret
}

func (boltTorrent) Close() error { return nil }
