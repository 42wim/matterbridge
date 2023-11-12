package stickers

import (
	"context"

	"github.com/ethereum/go-ethereum/p2p"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/ipfs"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/transactions"
)

// NewService initializes service instance.
func NewService(acc *accounts.Database, rpcClient *rpc.Client, accountsManager *account.GethManager, config *params.NodeConfig, downloader *ipfs.Downloader, httpServer *server.MediaServer, pendingTracker *transactions.PendingTxTracker) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		accountsDB:      acc,
		rpcClient:       rpcClient,
		accountsManager: accountsManager,
		keyStoreDir:     config.KeyStoreDir,
		downloader:      downloader,
		httpServer:      httpServer,
		ctx:             ctx,
		cancel:          cancel,
		api:             NewAPI(ctx, acc, rpcClient, accountsManager, pendingTracker, config.KeyStoreDir, downloader, httpServer),
	}
}

// Service is a browsers service.
type Service struct {
	accountsDB      *accounts.Database
	rpcClient       *rpc.Client
	accountsManager *account.GethManager
	downloader      *ipfs.Downloader
	keyStoreDir     string
	httpServer      *server.MediaServer
	ctx             context.Context
	cancel          context.CancelFunc
	api             *API
}

// Start a service.
func (s *Service) Start() error {
	return nil
}

// Stop a service.
func (s *Service) Stop() error {
	s.cancel()
	return nil
}

func (s *Service) API() *API {
	return s.api
}

// APIs returns list of available RPC APIs.
func (s *Service) APIs() []ethRpc.API {
	return []ethRpc.API{
		{
			Namespace: "stickers",
			Version:   "0.1.0",
			Service:   s.api,
		},
	}
}

// Protocols returns list of p2p protocols.
func (s *Service) Protocols() []p2p.Protocol {
	return nil
}
