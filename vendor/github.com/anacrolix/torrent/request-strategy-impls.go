package torrent

import (
	"github.com/anacrolix/torrent/metainfo"
	request_strategy "github.com/anacrolix/torrent/request-strategy"
	"github.com/anacrolix/torrent/storage"
)

type requestStrategyInput struct {
	cl      *Client
	capFunc storage.TorrentCapacity
}

func (r requestStrategyInput) Torrent(ih metainfo.Hash) request_strategy.Torrent {
	return requestStrategyTorrent{r.cl.torrents[ih]}
}

func (r requestStrategyInput) Capacity() (int64, bool) {
	if r.capFunc == nil {
		return 0, false
	}
	return (*r.capFunc)()
}

func (r requestStrategyInput) MaxUnverifiedBytes() int64 {
	return r.cl.config.MaxUnverifiedBytes
}

var _ request_strategy.Input = requestStrategyInput{}

// Returns what is necessary to run request_strategy.GetRequestablePieces for primaryTorrent.
func (cl *Client) getRequestStrategyInput(primaryTorrent *Torrent) (input request_strategy.Input) {
	return requestStrategyInput{
		cl:      cl,
		capFunc: primaryTorrent.storage.Capacity,
	}
}

func (t *Torrent) getRequestStrategyInput() request_strategy.Input {
	return t.cl.getRequestStrategyInput(t)
}

type requestStrategyTorrent struct {
	t *Torrent
}

func (r requestStrategyTorrent) Piece(i int) request_strategy.Piece {
	return requestStrategyPiece{r.t, i}
}

func (r requestStrategyTorrent) ChunksPerPiece() uint32 {
	return r.t.chunksPerRegularPiece()
}

func (r requestStrategyTorrent) PieceLength() int64 {
	return r.t.info.PieceLength
}

var _ request_strategy.Torrent = requestStrategyTorrent{}

type requestStrategyPiece struct {
	t *Torrent
	i pieceIndex
}

func (r requestStrategyPiece) Request() bool {
	return !r.t.ignorePieceForRequests(r.i)
}

func (r requestStrategyPiece) NumPendingChunks() int {
	return int(r.t.pieceNumPendingChunks(r.i))
}

var _ request_strategy.Piece = requestStrategyPiece{}
