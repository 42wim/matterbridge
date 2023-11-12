package storage

import (
	"io"

	"github.com/anacrolix/torrent/metainfo"
)

type ClientImplCloser interface {
	ClientImpl
	Close() error
}

// Represents data storage for an unspecified torrent.
type ClientImpl interface {
	OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (TorrentImpl, error)
}

type TorrentCapacity *func() (cap int64, capped bool)

// Data storage bound to a torrent.
type TorrentImpl struct {
	Piece func(p metainfo.Piece) PieceImpl
	Close func() error
	// Storages that share the same space, will provide equal pointers. The function is called once
	// to determine the storage for torrents sharing the same function pointer, and mutated in
	// place.
	Capacity TorrentCapacity
}

// Interacts with torrent piece data. Optional interfaces to implement include:
//   io.WriterTo, such as when a piece supports a more efficient way to write out incomplete chunks.
//   SelfHashing, such as when a piece supports a more efficient way to hash its contents.
type PieceImpl interface {
	// These interfaces are not as strict as normally required. They can
	// assume that the parameters are appropriate for the dimensions of the
	// piece.
	io.ReaderAt
	io.WriterAt
	// Called when the client believes the piece data will pass a hash check.
	// The storage can move or mark the piece data as read-only as it sees
	// fit.
	MarkComplete() error
	MarkNotComplete() error
	// Returns true if the piece is complete.
	Completion() Completion
}

type Completion struct {
	Complete bool
	Ok       bool
}

// Allows a storage backend to override hashing (i.e. if it can do it more efficiently than the torrent client can)
type SelfHashing interface {
	SelfHash() (metainfo.Hash, error)
}
