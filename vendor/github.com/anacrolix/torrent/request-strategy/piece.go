package request_strategy

type ChunksIterFunc func(func(ChunkIndex))

type ChunksIter interface {
	Iter(func(ci ChunkIndex))
}

type Piece interface {
	Request() bool
	NumPendingChunks() int
}
