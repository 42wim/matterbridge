package rpcfilters

import (
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type PendingTxInfo struct {
	Hash    common.Hash
	Type    string
	From    common.Address
	ChainID uint64
}

// transactionSentToUpstreamEvent represents an event that one can subscribe to
type transactionSentToUpstreamEvent struct {
	sxMu     sync.Mutex
	sx       map[int]chan *PendingTxInfo
	listener chan *PendingTxInfo
	quit     chan struct{}
}

func newTransactionSentToUpstreamEvent() *transactionSentToUpstreamEvent {
	return &transactionSentToUpstreamEvent{
		sx:       make(map[int]chan *PendingTxInfo),
		listener: make(chan *PendingTxInfo),
	}
}

func (e *transactionSentToUpstreamEvent) Start() error {
	if e.quit != nil {
		return errors.New("latest transaction sent to upstream event is already started")
	}

	e.quit = make(chan struct{})

	go func() {
		for {
			select {
			case transactionInfo := <-e.listener:
				if e.numberOfSubscriptions() == 0 {
					continue
				}
				e.processTransactionSentToUpstream(transactionInfo)
			case <-e.quit:
				return
			}
		}
	}()

	return nil
}

func (e *transactionSentToUpstreamEvent) numberOfSubscriptions() int {
	e.sxMu.Lock()
	defer e.sxMu.Unlock()
	return len(e.sx)
}

func (e *transactionSentToUpstreamEvent) processTransactionSentToUpstream(transactionInfo *PendingTxInfo) {

	e.sxMu.Lock()
	defer e.sxMu.Unlock()

	for id, channel := range e.sx {
		select {
		case channel <- transactionInfo:
		default:
			log.Error("dropping messages %s for subscriotion %d because the channel is full", transactionInfo, id)
		}
	}
}

func (e *transactionSentToUpstreamEvent) Stop() {
	if e.quit == nil {
		return
	}

	select {
	case <-e.quit:
		return
	default:
		close(e.quit)
	}

	e.quit = nil
}

func (e *transactionSentToUpstreamEvent) Subscribe() (int, interface{}) {
	e.sxMu.Lock()
	defer e.sxMu.Unlock()

	channel := make(chan *PendingTxInfo, 512)
	id := len(e.sx)
	e.sx[id] = channel
	return id, channel
}

func (e *transactionSentToUpstreamEvent) Unsubscribe(id int) {
	e.sxMu.Lock()
	defer e.sxMu.Unlock()

	delete(e.sx, id)
}

// Trigger gets called in order to trigger the event
func (e *transactionSentToUpstreamEvent) Trigger(transactionInfo *PendingTxInfo) {
	e.listener <- transactionInfo
}
