package pubsub

import (
	"context"
	"encoding/binary"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerMetadataStore is an interface for storing and retrieving per peer metadata
type PeerMetadataStore interface {
	// Get retrieves the metadata associated with a peer;
	// It should return nil if there is no metadata associated with the peer and not an error.
	Get(context.Context, peer.ID) ([]byte, error)
	// Put sets the metadata associated with a peer.
	Put(context.Context, peer.ID, []byte) error
}

// BasicSeqnoValidator is a basic validator, usable as a default validator, that ignores replayed
// messages outside the seen cache window. The validator uses the message seqno as a peer-specific
// nonce to decide whether the message should be propagated, comparing to the maximal nonce store
// in the peer metadata store. This is useful to ensure that there can be no infinitely propagating
// messages in the network regardless of the seen cache span and network diameter.
// It requires that pubsub is instantiated with a strict message signing policy and that seqnos
// are not disabled, ie it doesn't support anonymous mode.
//
// Warning: See https://github.com/libp2p/rust-libp2p/issues/3453
// TL;DR: rust is currently violating the spec by issuing a random seqno, which creates an
// interoperability hazard. We expect this issue to be addressed in the not so distant future,
// but keep this in mind if you are in a mixed environment with (older) rust nodes.
type BasicSeqnoValidator struct {
	mx   sync.RWMutex
	meta PeerMetadataStore
}

// NewBasicSeqnoValidator constructs a BasicSeqnoValidator using the givven PeerMetadataStore.
func NewBasicSeqnoValidator(meta PeerMetadataStore) ValidatorEx {
	val := &BasicSeqnoValidator{
		meta: meta,
	}
	return val.validate
}

func (v *BasicSeqnoValidator) validate(ctx context.Context, _ peer.ID, m *Message) ValidationResult {
	p := m.GetFrom()

	v.mx.RLock()
	nonceBytes, err := v.meta.Get(ctx, p)
	v.mx.RUnlock()

	if err != nil {
		log.Warn("error retrieving peer nonce: %s", err)
		return ValidationIgnore
	}

	var nonce uint64
	if len(nonceBytes) > 0 {
		nonce = binary.BigEndian.Uint64(nonceBytes)
	}

	var seqno uint64
	seqnoBytes := m.GetSeqno()
	if len(seqnoBytes) > 0 {
		seqno = binary.BigEndian.Uint64(seqnoBytes)
	}

	// compare against the largest seen nonce
	if seqno <= nonce {
		return ValidationIgnore
	}

	// get the nonce and compare again with an exclusive lock before commiting (cf concurrent validation)
	v.mx.Lock()
	defer v.mx.Unlock()

	nonceBytes, err = v.meta.Get(ctx, p)
	if err != nil {
		log.Warn("error retrieving peer nonce: %s", err)
		return ValidationIgnore
	}

	if len(nonceBytes) > 0 {
		nonce = binary.BigEndian.Uint64(nonceBytes)
	}

	if seqno <= nonce {
		return ValidationIgnore
	}

	// update the nonce
	nonceBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, seqno)

	err = v.meta.Put(ctx, p, nonceBytes)
	if err != nil {
		log.Warn("error storing peer nonce: %s", err)
	}

	return ValidationAccept
}
