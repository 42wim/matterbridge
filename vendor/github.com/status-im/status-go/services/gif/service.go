package gif

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/multiaccounts/accounts"
)

// Service represents out own implementation of personal sign operations.
type Service struct {
	accountsDB *accounts.Database
}

// New returns a new Service.
func NewService(db *accounts.Database) *Service {
	return &Service{accountsDB: db}
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "gif",
			Version:   "0.1.0",
			Service:   NewGifAPI(s.accountsDB),
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
	return nil
}
