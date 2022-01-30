package ratchet

import (
	"crypto/hmac"
	"crypto/sha256"
)

var messageKeySeed = []byte{0x01}
var chainKeySeed = []byte{0x02}

// NewSenderChainKey will return a new SenderChainKey.
func NewSenderChainKey(iteration uint32, chainKey []byte) *SenderChainKey {
	return &SenderChainKey{
		iteration: iteration,
		chainKey:  chainKey,
	}
}

// NewSenderChainKeyFromStruct will return a new chain key object from the
// given serializeable structure.
func NewSenderChainKeyFromStruct(structure *SenderChainKeyStructure) *SenderChainKey {
	return &SenderChainKey{
		iteration: structure.Iteration,
		chainKey:  structure.ChainKey,
	}
}

// NewStructFromSenderChainKeys returns a serializeable structure of chain keys.
func NewStructFromSenderChainKey(key *SenderChainKey) *SenderChainKeyStructure {
	return &SenderChainKeyStructure{
		Iteration: key.iteration,
		ChainKey:  key.chainKey,
	}
}

// SenderChainKeyStructure is a serializeable structure of SenderChainKeys.
type SenderChainKeyStructure struct {
	Iteration uint32
	ChainKey  []byte
}

type SenderChainKey struct {
	iteration uint32
	chainKey  []byte
}

func (k *SenderChainKey) Iteration() uint32 {
	return k.iteration
}

func (k *SenderChainKey) SenderMessageKey() (*SenderMessageKey, error) {
	return NewSenderMessageKey(k.iteration, k.getDerivative(messageKeySeed, k.chainKey))
}

func (k *SenderChainKey) Next() *SenderChainKey {
	return NewSenderChainKey(k.iteration+1, k.getDerivative(chainKeySeed, k.chainKey))
}

func (k *SenderChainKey) Seed() []byte {
	return k.chainKey
}

func (k *SenderChainKey) getDerivative(seed []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key[:])
	mac.Write(seed)

	return mac.Sum(nil)
}
