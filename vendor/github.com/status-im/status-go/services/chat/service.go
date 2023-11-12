package chat

import (
	"github.com/ethereum/go-ethereum/p2p"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/protocol"
)

func NewService(accountsDB *accounts.Database) *Service {
	return &Service{
		accountsDB: accountsDB,
	}
}

type Service struct {
	messenger  *protocol.Messenger
	accountsDB *accounts.Database
}

func (s *Service) Init(messenger *protocol.Messenger) {
	s.messenger = messenger
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Stop() error {
	return nil
}

func (s *Service) APIs() []gethrpc.API {
	return []gethrpc.API{
		{
			Namespace: "chat",
			Version:   "0.1.0",
			Service:   NewAPI(s),
		},
	}
}

func (s *Service) Protocols() []p2p.Protocol {
	return nil
}
