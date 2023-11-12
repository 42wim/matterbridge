package discovery

import (
	"time"

	"github.com/ethereum/go-ethereum/p2p/discv5"
)

const (
	// EthereumV5 is kademlia-based discovery from go-ethereum repository.
	EthereumV5 = "ethv5"
	// RendezvousV1 is req/rep based discovery that uses ENR for records.
	RendezvousV1 = "ethvousv1"
)

// Discovery is an abstract interface for using different discovery providers.
type Discovery interface {
	Running() bool
	Start() error
	Stop() error
	Register(topic string, stop chan struct{}) error
	Discover(topic string, period <-chan time.Duration, found chan<- *discv5.Node, lookup chan<- bool) error
}
