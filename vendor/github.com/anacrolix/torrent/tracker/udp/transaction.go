package udp

import "math/rand"

func RandomTransactionId() TransactionId {
	return TransactionId(rand.Uint32())
}

type TransactionResponseHandler func(dr DispatchedResponse)

type Transaction struct {
	id int32
	d  *Dispatcher
	h  TransactionResponseHandler
}

func (t *Transaction) Id() TransactionId {
	return t.id
}

func (t *Transaction) End() {
	t.d.forgetTransaction(t.id)
}
