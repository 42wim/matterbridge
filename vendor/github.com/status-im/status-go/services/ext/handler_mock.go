package ext

import (
	"github.com/status-im/status-go/eth-node/types"
)

type failureMessage struct {
	IDs   [][]byte
	Error error
}

func NewHandlerMock(buf int) HandlerMock {
	return HandlerMock{
		confirmations:     make(chan [][]byte, buf),
		expirations:       make(chan failureMessage, buf),
		requestsCompleted: make(chan types.Hash, buf),
		requestsExpired:   make(chan types.Hash, buf),
		requestsFailed:    make(chan types.Hash, buf),
	}
}

type HandlerMock struct {
	confirmations     chan [][]byte
	expirations       chan failureMessage
	requestsCompleted chan types.Hash
	requestsExpired   chan types.Hash
	requestsFailed    chan types.Hash
}

func (t HandlerMock) EnvelopeSent(ids [][]byte) {
	t.confirmations <- ids
}

func (t HandlerMock) EnvelopeExpired(ids [][]byte, err error) {
	t.expirations <- failureMessage{IDs: ids, Error: err}
}

func (t HandlerMock) MailServerRequestCompleted(requestID types.Hash, lastEnvelopeHash types.Hash, cursor []byte, err error) {
	if err == nil {
		t.requestsCompleted <- requestID
	} else {
		t.requestsFailed <- requestID
	}
}

func (t HandlerMock) MailServerRequestExpired(hash types.Hash) {
	t.requestsExpired <- hash
}
