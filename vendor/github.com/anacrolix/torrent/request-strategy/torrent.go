package request_strategy

type Torrent interface {
	Piece(int) Piece
	ChunksPerPiece() uint32
	PieceLength() int64
}
