package node

import (
	"sync"

	"github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"
)

type payloads struct {
	sync.Mutex

	payloads map[state.PeerID]*protobuf.Payload
}

// @todo check in all the functions below that we aren't duplicating stuff

func newPayloads() payloads {
	return payloads{
		payloads: make(map[state.PeerID]*protobuf.Payload),
	}
}

func (p *payloads) AddOffers(peer state.PeerID, offers ...[]byte) {
	p.Lock()
	defer p.Unlock()

	payload := p.get(peer)

	payload.Offers = append(payload.Offers, offers...)

	p.set(peer, payload)
}

func (p *payloads) AddAcks(peer state.PeerID, acks [][]byte) {
	p.Lock()
	defer p.Unlock()

	payload := p.get(peer)

	payload.Acks = append(payload.Acks, acks...)

	p.set(peer, payload)
}

func (p *payloads) AddRequests(peer state.PeerID, request ...[]byte) {
	p.Lock()
	defer p.Unlock()

	payload := p.get(peer)

	payload.Requests = append(payload.Requests, request...)

	p.set(peer, payload)
}

func (p *payloads) AddMessages(peer state.PeerID, messages ...*protobuf.Message) {
	p.Lock()
	defer p.Unlock()

	payload := p.get(peer)
	if payload.Messages == nil {
		payload.Messages = make([]*protobuf.Message, 0)
	}

	payload.Messages = append(payload.Messages, messages...)
	p.set(peer, payload)
}

func (p *payloads) MapAndClear(f func(state.PeerID, *protobuf.Payload) error) error {
	p.Lock()
	defer p.Unlock()

	for peer, payload := range p.payloads {
		err := f(peer, payload)
		if err != nil {
			return err
		}
	}

	// TODO: this should only be called upon confirmation that the message has been sent
	p.payloads = make(map[state.PeerID]*protobuf.Payload)
	return nil
}

func (p *payloads) get(peer state.PeerID) *protobuf.Payload {
	payload := p.payloads[peer]
	if payload == nil {
		return &protobuf.Payload{}
	}
	return payload
}

func (p *payloads) set(peer state.PeerID, payload *protobuf.Payload) {
	p.payloads[peer] = payload
}
