package web3provider

import (
	"database/sql"

	"github.com/ethereum/go-ethereum/p2p"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/transactions"

	"github.com/status-im/status-go/multiaccounts/accounts"

	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/permissions"
	"github.com/status-im/status-go/services/rpcfilters"
)

func NewService(appDB *sql.DB, accountsDB *accounts.Database, rpcClient *rpc.Client, config *params.NodeConfig, accountsManager *account.GethManager, rpcFiltersSrvc *rpcfilters.Service, transactor *transactions.Transactor) *Service {
	return &Service{
		permissionsDB:   permissions.NewDB(appDB),
		accountsDB:      accountsDB,
		rpcClient:       rpcClient,
		rpcFiltersSrvc:  rpcFiltersSrvc,
		config:          config,
		accountsManager: accountsManager,
		transactor:      transactor,
	}
}

type Service struct {
	permissionsDB   *permissions.Database
	accountsDB      *accounts.Database
	rpcClient       *rpc.Client
	rpcFiltersSrvc  *rpcfilters.Service
	accountsManager *account.GethManager
	config          *params.NodeConfig
	transactor      *transactions.Transactor
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
			Namespace: "provider",
			Version:   "0.1.0",
			Service:   NewAPI(s),
		},
	}
}

func (s *Service) Protocols() []p2p.Protocol {
	return nil
}
