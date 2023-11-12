package wallet

import (
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/services/ens"
	"github.com/status-im/status-go/services/stickers"
	"github.com/status-im/status-go/services/wallet/activity"
	"github.com/status-im/status-go/services/wallet/balance"
	"github.com/status-im/status-go/services/wallet/blockchainstate"
	"github.com/status-im/status-go/services/wallet/collectibles"
	"github.com/status-im/status-go/services/wallet/community"
	"github.com/status-im/status-go/services/wallet/currency"
	"github.com/status-im/status-go/services/wallet/history"
	"github.com/status-im/status-go/services/wallet/market"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/thirdparty/alchemy"
	"github.com/status-im/status-go/services/wallet/thirdparty/coingecko"
	"github.com/status-im/status-go/services/wallet/thirdparty/cryptocompare"
	"github.com/status-im/status-go/services/wallet/thirdparty/opensea"
	"github.com/status-im/status-go/services/wallet/thirdparty/rarible"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletconnect"
	"github.com/status-im/status-go/services/wallet/walletevent"
	"github.com/status-im/status-go/transactions"
)

const (
	EventBlockchainStatusChanged walletevent.EventType = "wallet-blockchain-status-changed"
)

// NewService initializes service instance.
func NewService(
	db *sql.DB,
	accountsDB *accounts.Database,
	appDB *sql.DB,
	rpcClient *rpc.Client,
	accountFeed *event.Feed,
	settingsFeed *event.Feed,
	gethManager *account.GethManager,
	transactor *transactions.Transactor,
	config *params.NodeConfig,
	ens *ens.Service,
	stickers *stickers.Service,
	pendingTxManager *transactions.PendingTxTracker,
	feed *event.Feed,
	mediaServer *server.MediaServer,
) *Service {
	cryptoOnRampManager := NewCryptoOnRampManager(&CryptoOnRampOptions{
		dataSourceType: DataSourceStatic,
	})

	signals := &walletevent.SignalsTransmitter{
		Publisher: feed,
	}
	blockchainStatus := make(map[uint64]string)
	mutex := sync.Mutex{}
	rpcClient.SetWalletNotifier(func(chainID uint64, message string) {
		mutex.Lock()
		defer mutex.Unlock()

		if len(blockchainStatus) == 0 {
			networks, err := rpcClient.NetworkManager.Get(false)
			if err != nil {
				return
			}

			for _, network := range networks {
				blockchainStatus[network.ChainID] = "up"
			}
		}

		blockchainStatus[chainID] = message
		encodedmessage, err := json.Marshal(blockchainStatus)
		if err != nil {
			return
		}

		feed.Send(walletevent.Event{
			Type:     EventBlockchainStatusChanged,
			Accounts: []common.Address{},
			Message:  string(encodedmessage),
			At:       time.Now().Unix(),
			ChainID:  chainID,
		})
	})

	communityManager := community.NewManager(db, mediaServer, feed)
	balanceCacher := balance.NewCacherWithTTL(5 * time.Minute)
	tokenManager := token.NewTokenManager(db, rpcClient, communityManager, rpcClient.NetworkManager, appDB, mediaServer, feed)
	savedAddressesManager := &SavedAddressesManager{db: db}
	transactionManager := transfer.NewTransactionManager(db, gethManager, transactor, config, accountsDB, pendingTxManager, feed)
	blockChainState := blockchainstate.NewBlockChainState()
	transferController := transfer.NewTransferController(db, accountsDB, rpcClient, accountFeed, feed, transactionManager, pendingTxManager,
		tokenManager, balanceCacher, blockChainState)
	transferController.Start()
	cryptoCompare := cryptocompare.NewClient()
	coingecko := coingecko.NewClient()
	marketManager := market.NewManager(cryptoCompare, coingecko, feed)
	reader := NewReader(rpcClient, tokenManager, marketManager, communityManager, accountsDB, NewPersistence(db), feed)
	history := history.NewService(db, accountsDB, feed, rpcClient, tokenManager, marketManager, balanceCacher.Cache())
	currency := currency.NewService(db, feed, tokenManager, marketManager)

	openseaHTTPClient := opensea.NewHTTPClient()
	openseaV2Client := opensea.NewClientV2(config.WalletConfig.OpenseaAPIKey, openseaHTTPClient)
	raribleClient := rarible.NewClient(config.WalletConfig.RaribleMainnetAPIKey, config.WalletConfig.RaribleTestnetAPIKey)
	alchemyClient := alchemy.NewClient(config.WalletConfig.AlchemyAPIKeys)

	// Try OpenSea, Infura, Alchemy in that order
	contractOwnershipProviders := []thirdparty.CollectibleContractOwnershipProvider{
		raribleClient,
		alchemyClient,
	}

	accountOwnershipProviders := []thirdparty.CollectibleAccountOwnershipProvider{
		raribleClient,
		alchemyClient,
		openseaV2Client,
	}

	collectibleDataProviders := []thirdparty.CollectibleDataProvider{
		raribleClient,
		alchemyClient,
		openseaV2Client,
	}

	collectionDataProviders := []thirdparty.CollectionDataProvider{
		raribleClient,
		alchemyClient,
		openseaV2Client,
	}

	collectiblesManager := collectibles.NewManager(db, rpcClient, communityManager, contractOwnershipProviders, accountOwnershipProviders, collectibleDataProviders, collectionDataProviders, mediaServer, feed)
	collectibles := collectibles.NewService(db, feed, accountsDB, accountFeed, settingsFeed, communityManager, rpcClient.NetworkManager, collectiblesManager)

	activity := activity.NewService(db, tokenManager, collectiblesManager, feed, pendingTxManager)

	walletconnect := walletconnect.NewService(db, rpcClient.NetworkManager, accountsDB, transactionManager, gethManager, feed, config)

	return &Service{
		db:                    db,
		accountsDB:            accountsDB,
		rpcClient:             rpcClient,
		tokenManager:          tokenManager,
		communityManager:      communityManager,
		savedAddressesManager: savedAddressesManager,
		transactionManager:    transactionManager,
		pendingTxManager:      pendingTxManager,
		transferController:    transferController,
		cryptoOnRampManager:   cryptoOnRampManager,
		collectiblesManager:   collectiblesManager,
		collectibles:          collectibles,
		feesManager:           &FeeManager{rpcClient},
		gethManager:           gethManager,
		marketManager:         marketManager,
		transactor:            transactor,
		ens:                   ens,
		stickers:              stickers,
		feed:                  feed,
		signals:               signals,
		reader:                reader,
		history:               history,
		currency:              currency,
		activity:              activity,
		decoder:               NewDecoder(),
		blockChainState:       blockChainState,
		keycardPairings:       NewKeycardPairings(),
		walletConnect:         walletconnect,
	}
}

// Service is a wallet service.
type Service struct {
	db                    *sql.DB
	accountsDB            *accounts.Database
	rpcClient             *rpc.Client
	savedAddressesManager *SavedAddressesManager
	tokenManager          *token.Manager
	communityManager      *community.Manager
	transactionManager    *transfer.TransactionManager
	pendingTxManager      *transactions.PendingTxTracker
	cryptoOnRampManager   *CryptoOnRampManager
	transferController    *transfer.Controller
	feesManager           *FeeManager
	marketManager         *market.Manager
	started               bool
	collectiblesManager   *collectibles.Manager
	collectibles          *collectibles.Service
	gethManager           *account.GethManager
	transactor            *transactions.Transactor
	ens                   *ens.Service
	stickers              *stickers.Service
	feed                  *event.Feed
	signals               *walletevent.SignalsTransmitter
	reader                *Reader
	history               *history.Service
	currency              *currency.Service
	activity              *activity.Service
	decoder               *Decoder
	blockChainState       *blockchainstate.BlockChainState
	keycardPairings       *KeycardPairings
	walletConnect         *walletconnect.Service
}

// Start signals transmitter.
func (s *Service) Start() error {
	s.transferController.Start()
	s.currency.Start()
	err := s.signals.Start()
	s.history.Start()
	s.collectibles.Start()
	s.started = true
	return err
}

// Set external Collectibles community info provider
func (s *Service) SetWalletCommunityInfoProvider(provider thirdparty.CommunityInfoProvider) {
	s.communityManager.SetCommunityInfoProvider(provider)
}

// Stop reactor and close db.
func (s *Service) Stop() error {
	log.Info("wallet will be stopped")
	s.signals.Stop()
	s.transferController.Stop()
	s.currency.Stop()
	s.reader.Stop()
	s.history.Stop()
	s.activity.Stop()
	s.collectibles.Stop()
	s.started = false
	log.Info("wallet stopped")
	return nil
}

// APIs returns list of available RPC APIs.
func (s *Service) APIs() []gethrpc.API {
	return []gethrpc.API{
		{
			Namespace: "wallet",
			Version:   "0.1.0",
			Service:   NewAPI(s),
			Public:    true,
		},
	}
}

// Protocols returns list of p2p protocols.
func (s *Service) Protocols() []p2p.Protocol {
	return nil
}

func (s *Service) IsStarted() bool {
	return s.started
}

func (s *Service) KeycardPairings() *KeycardPairings {
	return s.keycardPairings
}
