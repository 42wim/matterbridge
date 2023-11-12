package request_strategy

import (
	"fmt"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/google/btree"
)

func NewPieceOrder() *PieceRequestOrder {
	return &PieceRequestOrder{
		tree: btree.New(32),
		keys: make(map[PieceRequestOrderKey]PieceRequestOrderState),
	}
}

type PieceRequestOrder struct {
	tree *btree.BTree
	keys map[PieceRequestOrderKey]PieceRequestOrderState
}

type PieceRequestOrderKey struct {
	InfoHash metainfo.Hash
	Index    int
}

type PieceRequestOrderState struct {
	Priority     piecePriority
	Partial      bool
	Availability int
}

type pieceRequestOrderItem struct {
	key   PieceRequestOrderKey
	state PieceRequestOrderState
}

func (me *pieceRequestOrderItem) Less(other btree.Item) bool {
	otherConcrete := other.(*pieceRequestOrderItem)
	return pieceOrderLess(me, otherConcrete).Less()
}

func (me *PieceRequestOrder) Add(key PieceRequestOrderKey, state PieceRequestOrderState) {
	if _, ok := me.keys[key]; ok {
		panic(key)
	}
	if me.tree.ReplaceOrInsert(&pieceRequestOrderItem{
		key:   key,
		state: state,
	}) != nil {
		panic("shouldn't already have this")
	}
	me.keys[key] = state
}

func (me *PieceRequestOrder) Update(key PieceRequestOrderKey, state PieceRequestOrderState) {
	oldState, ok := me.keys[key]
	if !ok {
		panic("key should have been added already")
	}
	if state == oldState {
		return
	}
	item := pieceRequestOrderItem{
		key:   key,
		state: oldState,
	}
	if me.tree.Delete(&item) == nil {
		panic(fmt.Sprintf("%#v", key))
	}
	item.state = state
	if me.tree.ReplaceOrInsert(&item) != nil {
		panic(key)
	}
	me.keys[key] = state
}

func (me *PieceRequestOrder) existingItemForKey(key PieceRequestOrderKey) pieceRequestOrderItem {
	return pieceRequestOrderItem{
		key:   key,
		state: me.keys[key],
	}
}

func (me *PieceRequestOrder) Delete(key PieceRequestOrderKey) {
	item := me.existingItemForKey(key)
	if me.tree.Delete(&item) == nil {
		panic(key)
	}
	delete(me.keys, key)
	// log.Printf("deleting %#v", key)
}

func (me *PieceRequestOrder) Len() int {
	return len(me.keys)
}
