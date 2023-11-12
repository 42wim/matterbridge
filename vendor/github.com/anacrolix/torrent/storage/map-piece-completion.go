package storage

import (
	"sync"

	"github.com/anacrolix/torrent/metainfo"
)

type mapPieceCompletion struct {
	// TODO: Generics
	m sync.Map
}

var _ PieceCompletion = (*mapPieceCompletion)(nil)

func NewMapPieceCompletion() PieceCompletion {
	return &mapPieceCompletion{}
}

func (*mapPieceCompletion) Close() error { return nil }

func (me *mapPieceCompletion) Get(pk metainfo.PieceKey) (c Completion, err error) {
	v, ok := me.m.Load(pk)
	if ok {
		c.Complete = v.(bool)
	}
	c.Ok = ok
	return
}

func (me *mapPieceCompletion) Set(pk metainfo.PieceKey, b bool) error {
	me.m.Store(pk, b)
	return nil
}
