package verifier

import (
	"context"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// LocalVerifier verifies nodes based on a provided local list.
type LocalVerifier struct {
	KnownPeers map[enode.ID]struct{}
}

// NewLocalVerifier returns a new LocalVerifier instance.
func NewLocalVerifier(peers []enode.ID) *LocalVerifier {
	knownPeers := make(map[enode.ID]struct{})
	for _, peer := range peers {
		knownPeers[peer] = struct{}{}
	}

	return &LocalVerifier{KnownPeers: knownPeers}
}

// VerifyNode checks if a given node is trusted using a local list.
func (v *LocalVerifier) VerifyNode(_ context.Context, nodeID enode.ID) bool {
	if _, ok := v.KnownPeers[nodeID]; ok {
		return true
	}
	return false
}
