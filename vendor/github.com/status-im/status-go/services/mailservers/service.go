package mailservers

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

func NewService(db *Database) *Service {
	return &Service{db: db}
}

type Service struct {
	db *Database
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Stop() error {
	return nil
}

func (s *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "mailservers",
			Version:   "0.1.0",
			Service:   NewAPI(s.db),
		},
	}
}

func (s *Service) Protocols() []p2p.Protocol {
	return nil
}
