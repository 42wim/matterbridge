package gethbridge

import (
	"io"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/status-im/status-go/eth-node/types"
	waku "github.com/status-im/status-go/waku/common"
)

type wakuEnvelope struct {
	env *waku.Envelope
}

// NewWakuEnvelope returns an object that wraps Geth's Waku Envelope in a types interface.
func NewWakuEnvelope(e *waku.Envelope) types.Envelope {
	return &wakuEnvelope{env: e}
}

func (w *wakuEnvelope) Unwrap() interface{} {
	return w.env
}

func (w *wakuEnvelope) Hash() types.Hash {
	return types.Hash(w.env.Hash())
}

func (w *wakuEnvelope) Bloom() []byte {
	return w.env.Bloom()
}

func (w *wakuEnvelope) PoW() float64 {
	return w.env.PoW()
}

func (w *wakuEnvelope) Expiry() uint32 {
	return w.env.Expiry
}

func (w *wakuEnvelope) TTL() uint32 {
	return w.env.TTL
}

func (w *wakuEnvelope) Topic() types.TopicType {
	return types.TopicType(w.env.Topic)
}

func (w *wakuEnvelope) Size() int {
	return len(w.env.Data)
}

func (w *wakuEnvelope) DecodeRLP(s *rlp.Stream) error {
	return w.env.DecodeRLP(s)
}

func (w *wakuEnvelope) EncodeRLP(writer io.Writer) error {
	return rlp.Encode(writer, w.env)
}
