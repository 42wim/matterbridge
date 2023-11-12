package udp

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

// Maintains a mapping of transaction IDs to handlers.
type Dispatcher struct {
	mu           sync.RWMutex
	transactions map[TransactionId]Transaction
}

// The caller owns b.
func (me *Dispatcher) Dispatch(b []byte, addr net.Addr) error {
	buf := bytes.NewBuffer(b)
	var rh ResponseHeader
	err := Read(buf, &rh)
	if err != nil {
		return err
	}
	me.mu.RLock()
	defer me.mu.RUnlock()
	if t, ok := me.transactions[rh.TransactionId]; ok {
		t.h(DispatchedResponse{
			Header: rh,
			Body:   append([]byte(nil), buf.Bytes()...),
			Addr:   addr,
		})
		return nil
	} else {
		return fmt.Errorf("unknown transaction id %v", rh.TransactionId)
	}
}

func (me *Dispatcher) forgetTransaction(id TransactionId) {
	me.mu.Lock()
	defer me.mu.Unlock()
	delete(me.transactions, id)
}

func (me *Dispatcher) NewTransaction(h TransactionResponseHandler) Transaction {
	me.mu.Lock()
	defer me.mu.Unlock()
	for {
		id := RandomTransactionId()
		if _, ok := me.transactions[id]; ok {
			continue
		}
		t := Transaction{
			d:  me,
			h:  h,
			id: id,
		}
		if me.transactions == nil {
			me.transactions = make(map[TransactionId]Transaction)
		}
		me.transactions[id] = t
		return t
	}
}

type DispatchedResponse struct {
	Header ResponseHeader
	// Response payload, after the header.
	Body []byte
	// Response source address
	Addr net.Addr
}
