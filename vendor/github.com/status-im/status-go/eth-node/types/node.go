package types

import (
	"fmt"

	"go.uber.org/zap"

	enstypes "github.com/status-im/status-go/eth-node/types/ens"
)

// EnodeID is a unique identifier for each node.
type EnodeID [32]byte

// ID prints as a long hexadecimal number.
func (n EnodeID) String() string {
	return fmt.Sprintf("%x", n[:])
}

type Node interface {
	NewENSVerifier(logger *zap.Logger) enstypes.ENSVerifier
	GetWaku(ctx interface{}) (Waku, error)
	GetWakuV2(ctx interface{}) (Waku, error)
	AddPeer(url string) error
	RemovePeer(url string) error
	PeersCount() int
}
