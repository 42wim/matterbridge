package peer

import (
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

// Make sure that Service implements node.Lifecycle interface.
var _ node.Lifecycle = (*Service)(nil)

// Discoverer manages peer discovery.
type Discoverer interface {
	Discover(topic string, max, min int) error
}

// Service it manages all endpoints for peer operations.
type Service struct {
	d Discoverer
}

// New returns a new Service.
func New() *Service {
	return &Service{}
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "peer",
			Version:   "1.0",
			Service:   NewAPI(s),
			Public:    false,
		},
	}
}

// SetDiscoverer sets discoverer for the API calls.
func (s *Service) SetDiscoverer(d Discoverer) {
	s.d = d
}

// Start is run when a service is started.
// It does nothing in this case but is required by `node.Service` interface.
func (s *Service) Start() error {
	return nil
}

// Stop is run when a service is stopped.
// It does nothing in this case but is required by `node.Service` interface.
func (s *Service) Stop() error {
	return nil
}
