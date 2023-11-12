package personal

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/ethapi"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

// Make sure that Service implements node.Service interface.
var _ node.Lifecycle = (*Service)(nil)

// Service represents out own implementation of personal sign operations.
type Service struct {
	am *accounts.Manager
}

// New returns a new Service.
func New(am *accounts.Manager) *Service {
	return &Service{am}
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "personal",
			Version:   "1.0",
			Service:   ethapi.NewLimitedPersonalAPI(s.am),
			Public:    false,
		},
	}
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
