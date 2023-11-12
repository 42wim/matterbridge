package torrent

import (
	request_strategy "github.com/anacrolix/torrent/request-strategy"
)

func (t *Torrent) updatePieceRequestOrder(pieceIndex int) {
	if t.storage == nil {
		return
	}
	t.cl.pieceRequestOrder[t.clientPieceRequestOrderKey()].Update(
		t.pieceRequestOrderKey(pieceIndex),
		t.requestStrategyPieceOrderState(pieceIndex))
}

func (t *Torrent) clientPieceRequestOrderKey() interface{} {
	if t.storage.Capacity == nil {
		return t
	}
	return t.storage.Capacity
}

func (t *Torrent) deletePieceRequestOrder() {
	if t.storage == nil {
		return
	}
	cpro := t.cl.pieceRequestOrder
	key := t.clientPieceRequestOrderKey()
	pro := cpro[key]
	for i := 0; i < t.numPieces(); i++ {
		pro.Delete(t.pieceRequestOrderKey(i))
	}
	if pro.Len() == 0 {
		delete(cpro, key)
	}
}

func (t *Torrent) initPieceRequestOrder() {
	if t.storage == nil {
		return
	}
	if t.cl.pieceRequestOrder == nil {
		t.cl.pieceRequestOrder = make(map[interface{}]*request_strategy.PieceRequestOrder)
	}
	key := t.clientPieceRequestOrderKey()
	cpro := t.cl.pieceRequestOrder
	if cpro[key] == nil {
		cpro[key] = request_strategy.NewPieceOrder()
	}
}

func (t *Torrent) addRequestOrderPiece(i int) {
	if t.storage == nil {
		return
	}
	t.cl.pieceRequestOrder[t.clientPieceRequestOrderKey()].Add(
		t.pieceRequestOrderKey(i),
		t.requestStrategyPieceOrderState(i))
}

func (t *Torrent) getPieceRequestOrder() *request_strategy.PieceRequestOrder {
	return t.cl.pieceRequestOrder[t.clientPieceRequestOrderKey()]
}
