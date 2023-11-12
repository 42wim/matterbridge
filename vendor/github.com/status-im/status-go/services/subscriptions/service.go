package subscriptions

import (
	gethnode "github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/rpc"
)

// Make sure that Service implements gethnode.Lifecycle interface.
var _ gethnode.Lifecycle = (*Service)(nil)

// Service represents our own implementation of personal sign operations.
type Service struct {
	api *API
}

// New returns a new Service.
func New(rpcPrivateClientFunc func() *rpc.Client) *Service {
	return &Service{
		api: NewPublicAPI(rpcPrivateClientFunc),
	}
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []gethrpc.API {
	return []gethrpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   s.api,
			Public:    true,
		},
	}
}

// Start is run when a service is started.
func (s *Service) Start() error {
	return nil
}

// Stop is run when a service is stopped.
func (s *Service) Stop() error {
	return s.api.activeSubscriptions.removeAll()
}
