package updates

import (
	"github.com/ethereum/go-ethereum/p2p"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/services/ens"
)

// NewService initializes service instance.
func NewService(ensService *ens.Service) *Service {
	return &Service{ensService}
}

type Service struct {
	ensService *ens.Service
}

// Start a service.
func (s *Service) Start() error {
	return nil
}

// Stop a service.
func (s *Service) Stop() error {
	return nil
}

// APIs returns list of available RPC APIs.
func (s *Service) APIs() []ethRpc.API {
	return []ethRpc.API{
		{
			Namespace: "updates",
			Version:   "0.1.0",
			Service:   NewAPI(s.ensService),
		},
	}
}

// Protocols returns list of p2p protocols.
func (s *Service) Protocols() []p2p.Protocol {
	return nil
}
