package node

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/common"
	"github.com/status-im/status-go/connection"
	"github.com/status-im/status-go/db"
	"github.com/status-im/status-go/discovery"
	"github.com/status-im/status-go/ipfs"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/peers"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/server"
	accountssvc "github.com/status-im/status-go/services/accounts"
	appmetricsservice "github.com/status-im/status-go/services/appmetrics"
	"github.com/status-im/status-go/services/browsers"
	"github.com/status-im/status-go/services/chat"
	"github.com/status-im/status-go/services/communitytokens"
	"github.com/status-im/status-go/services/ens"
	"github.com/status-im/status-go/services/gif"
	localnotifications "github.com/status-im/status-go/services/local-notifications"
	"github.com/status-im/status-go/services/mailservers"
	"github.com/status-im/status-go/services/peer"
	"github.com/status-im/status-go/services/permissions"
	"github.com/status-im/status-go/services/personal"
	"github.com/status-im/status-go/services/rpcfilters"
	"github.com/status-im/status-go/services/rpcstats"
	"github.com/status-im/status-go/services/status"
	"github.com/status-im/status-go/services/stickers"
	"github.com/status-im/status-go/services/subscriptions"
	"github.com/status-im/status-go/services/updates"
	"github.com/status-im/status-go/services/wakuext"
	"github.com/status-im/status-go/services/wakuv2ext"
	"github.com/status-im/status-go/services/wallet"
	"github.com/status-im/status-go/services/web3provider"
	"github.com/status-im/status-go/timesource"
	"github.com/status-im/status-go/transactions"
	"github.com/status-im/status-go/waku"
	"github.com/status-im/status-go/wakuv2"
)

// errors
var (
	ErrNodeRunning            = errors.New("node is already running")
	ErrNoGethNode             = errors.New("geth node is not available")
	ErrNoRunningNode          = errors.New("there is no running node")
	ErrAccountKeyStoreMissing = errors.New("account key store is not set")
	ErrServiceUnknown         = errors.New("service unknown")
	ErrDiscoveryRunning       = errors.New("discovery is already running")
	ErrRPCMethodUnavailable   = `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"the method called does not exist/is not available"}}`
)

// StatusNode abstracts contained geth node and provides helper methods to
// interact with it.
type StatusNode struct {
	mu sync.RWMutex

	appDB           *sql.DB
	multiaccountsDB *multiaccounts.Database
	walletDB        *sql.DB

	config    *params.NodeConfig // Status node configuration
	gethNode  *node.Node         // reference to Geth P2P stack/node
	rpcClient *rpc.Client        // reference to an RPC client

	downloader *ipfs.Downloader
	httpServer *server.MediaServer

	discovery discovery.Discovery
	register  *peers.Register
	peerPool  *peers.PeerPool
	db        *leveldb.DB // used as a cache for PeerPool

	log log.Logger

	gethAccountManager *account.GethManager
	accountsManager    *accounts.Manager
	transactor         *transactions.Transactor

	// services
	services      []common.StatusService
	publicMethods map[string]bool
	// we explicitly list every service, we could use interfaces
	// and store them in a nicer way and user reflection, but for now stupid is good
	rpcFiltersSrvc         *rpcfilters.Service
	subscriptionsSrvc      *subscriptions.Service
	rpcStatsSrvc           *rpcstats.Service
	statusPublicSrvc       *status.Service
	accountsSrvc           *accountssvc.Service
	browsersSrvc           *browsers.Service
	permissionsSrvc        *permissions.Service
	mailserversSrvc        *mailservers.Service
	providerSrvc           *web3provider.Service
	appMetricsSrvc         *appmetricsservice.Service
	walletSrvc             *wallet.Service
	peerSrvc               *peer.Service
	localNotificationsSrvc *localnotifications.Service
	personalSrvc           *personal.Service
	timeSourceSrvc         *timesource.NTPTimeSource
	wakuSrvc               *waku.Waku
	wakuExtSrvc            *wakuext.Service
	wakuV2Srvc             *wakuv2.Waku
	wakuV2ExtSrvc          *wakuv2ext.Service
	ensSrvc                *ens.Service
	communityTokensSrvc    *communitytokens.Service
	gifSrvc                *gif.Service
	stickersSrvc           *stickers.Service
	chatSrvc               *chat.Service
	updatesSrvc            *updates.Service
	pendingTracker         *transactions.PendingTxTracker

	walletFeed event.Feed
}

// New makes new instance of StatusNode.
func New(transactor *transactions.Transactor) *StatusNode {
	return &StatusNode{
		gethAccountManager: account.NewGethManager(),
		transactor:         transactor,
		log:                log.New("package", "status-go/node.StatusNode"),
		publicMethods:      make(map[string]bool),
	}
}

// Config exposes reference to running node's configuration
func (n *StatusNode) Config() *params.NodeConfig {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.config
}

// GethNode returns underlying geth node.
func (n *StatusNode) GethNode() *node.Node {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.gethNode
}

func (n *StatusNode) HTTPServer() *server.MediaServer {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.httpServer
}

// Server retrieves the currently running P2P network layer.
func (n *StatusNode) Server() *p2p.Server {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.gethNode == nil {
		return nil
	}

	return n.gethNode.Server()
}

// Start starts current StatusNode, failing if it's already started.
// It accepts a list of services that should be added to the node.
func (n *StatusNode) Start(config *params.NodeConfig, accs *accounts.Manager) error {
	return n.StartWithOptions(config, StartOptions{
		StartDiscovery:  true,
		AccountsManager: accs,
	})
}

// StartOptions allows to control some parameters of Start() method.
type StartOptions struct {
	StartDiscovery  bool
	AccountsManager *accounts.Manager
}

// StartMediaServerWithoutDB starts media server without starting the node
// The server can only handle requests that don't require appdb or IPFS downloader
func (n *StatusNode) StartMediaServerWithoutDB() error {
	if n.isRunning() {
		n.log.Debug("node is already running, no need to StartMediaServerWithoutDB")
		return nil
	}

	if n.httpServer != nil {
		if err := n.httpServer.Stop(); err != nil {
			return err
		}
	}

	httpServer, err := server.NewMediaServer(nil, nil, n.multiaccountsDB, nil)
	if err != nil {
		return err
	}

	n.httpServer = httpServer

	if err := n.httpServer.Start(); err != nil {
		return err
	}

	return nil
}

// StartWithOptions starts current StatusNode, failing if it's already started.
// It takes some options that allows to further configure starting process.
func (n *StatusNode) StartWithOptions(config *params.NodeConfig, options StartOptions) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.isRunning() {
		n.log.Debug("node is already running")
		return ErrNodeRunning
	}

	n.accountsManager = options.AccountsManager

	n.log.Debug("starting with options", "ClusterConfig", config.ClusterConfig)

	db, err := db.Create(config.DataDir, params.StatusDatabase)
	if err != nil {
		return fmt.Errorf("failed to create database at %s: %v", config.DataDir, err)
	}

	n.db = db

	err = n.startWithDB(config, options.AccountsManager, db)

	// continue only if there was no error when starting node with a db
	if err == nil && options.StartDiscovery && n.discoveryEnabled() {
		err = n.startDiscovery()
	}

	if err != nil {
		if dberr := db.Close(); dberr != nil {
			n.log.Error("error while closing leveldb after node crash", "error", dberr)
		}
		n.db = nil
		return err
	}

	return nil
}

func (n *StatusNode) startWithDB(config *params.NodeConfig, accs *accounts.Manager, db *leveldb.DB) error {
	if err := n.createNode(config, accs, db); err != nil {
		return err
	}
	n.config = config

	if err := n.setupRPCClient(); err != nil {
		return err
	}

	n.downloader = ipfs.NewDownloader(config.RootDataDir)

	if n.httpServer != nil {
		if err := n.httpServer.Stop(); err != nil {
			return err
		}
	}

	httpServer, err := server.NewMediaServer(n.appDB, n.downloader, n.multiaccountsDB, n.walletDB)
	if err != nil {
		return err
	}

	n.httpServer = httpServer

	if err := n.httpServer.Start(); err != nil {
		return err
	}

	if err := n.initServices(config, n.httpServer); err != nil {
		return err
	}
	return n.startGethNode()
}

func (n *StatusNode) createNode(config *params.NodeConfig, accs *accounts.Manager, db *leveldb.DB) (err error) {
	n.gethNode, err = MakeNode(config, accs, db)
	return err
}

// startGethNode starts current StatusNode, will fail if it's already started.
func (n *StatusNode) startGethNode() error {
	return n.gethNode.Start()
}

func (n *StatusNode) setupRPCClient() (err error) {
	// setup RPC client
	gethNodeClient, err := n.gethNode.Attach()
	if err != nil {
		return
	}
	n.rpcClient, err = rpc.NewClient(gethNodeClient, n.config.NetworkID, n.config.UpstreamConfig, n.config.Networks, n.appDB)
	if err != nil {
		return
	}

	return
}

func (n *StatusNode) discoveryEnabled() bool {
	return n.config != nil && (!n.config.NoDiscovery || n.config.Rendezvous) && n.config.ClusterConfig.Enabled
}

func (n *StatusNode) discoverNode() (*enode.Node, error) {
	if !n.isRunning() {
		return nil, nil
	}

	server := n.gethNode.Server()
	discNode := server.Self()

	if n.config.AdvertiseAddr == "" {
		return discNode, nil
	}

	n.log.Info("Using AdvertiseAddr for rendezvous", "addr", n.config.AdvertiseAddr)

	r := discNode.Record()
	r.Set(enr.IP(net.ParseIP(n.config.AdvertiseAddr)))
	if err := enode.SignV4(r, server.PrivateKey); err != nil {
		return nil, err
	}
	return enode.New(enode.ValidSchemes[r.IdentityScheme()], r)
}

func (n *StatusNode) startRendezvous() (discovery.Discovery, error) {
	if !n.config.Rendezvous {
		return nil, errors.New("rendezvous is not enabled")
	}
	if len(n.config.ClusterConfig.RendezvousNodes) == 0 {
		return nil, errors.New("rendezvous node must be provided if rendezvous discovery is enabled")
	}
	maddrs := make([]ma.Multiaddr, len(n.config.ClusterConfig.RendezvousNodes))
	for i, addr := range n.config.ClusterConfig.RendezvousNodes {
		var err error
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rendezvous node %s: %v", n.config.ClusterConfig.RendezvousNodes[0], err)
		}
	}
	node, err := n.discoverNode()
	if err != nil {
		return nil, fmt.Errorf("failed to get a discover node: %v", err)
	}

	return discovery.NewRendezvous(maddrs, n.gethNode.Server().PrivateKey, node)
}

// StartDiscovery starts the peers discovery protocols depending on the node config.
func (n *StatusNode) StartDiscovery() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.discoveryEnabled() {
		return n.startDiscovery()
	}

	return nil
}

func (n *StatusNode) startDiscovery() error {
	if n.isDiscoveryRunning() {
		return ErrDiscoveryRunning
	}

	discoveries := []discovery.Discovery{}
	if !n.config.NoDiscovery {
		discoveries = append(discoveries, discovery.NewDiscV5(
			n.gethNode.Server().PrivateKey,
			n.config.ListenAddr,
			parseNodesV5(n.config.ClusterConfig.BootNodes)))
	}
	if n.config.Rendezvous {
		d, err := n.startRendezvous()
		if err != nil {
			return err
		}
		discoveries = append(discoveries, d)
	}
	if len(discoveries) == 0 {
		return errors.New("wasn't able to register any discovery")
	} else if len(discoveries) > 1 {
		n.discovery = discovery.NewMultiplexer(discoveries)
	} else {
		n.discovery = discoveries[0]
	}
	log.Debug(
		"using discovery",
		"instance", reflect.TypeOf(n.discovery),
		"registerTopics", n.config.RegisterTopics,
		"requireTopics", n.config.RequireTopics,
	)
	n.register = peers.NewRegister(n.discovery, n.config.RegisterTopics...)
	options := peers.NewDefaultOptions()
	// TODO(dshulyak) consider adding a flag to define this behaviour
	options.AllowStop = len(n.config.RegisterTopics) == 0
	options.TrustedMailServers = parseNodesToNodeID(n.config.ClusterConfig.TrustedMailServers)

	n.peerPool = peers.NewPeerPool(
		n.discovery,
		n.config.RequireTopics,
		peers.NewCache(n.db),
		options,
	)
	if err := n.discovery.Start(); err != nil {
		return err
	}
	if err := n.register.Start(); err != nil {
		return err
	}
	return n.peerPool.Start(n.gethNode.Server())
}

// Stop will stop current StatusNode. A stopped node cannot be resumed.
func (n *StatusNode) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isRunning() {
		return ErrNoRunningNode
	}

	return n.stop()
}

// stop will stop current StatusNode. A stopped node cannot be resumed.
func (n *StatusNode) stop() error {
	if n.isDiscoveryRunning() {
		if err := n.stopDiscovery(); err != nil {
			n.log.Error("Error stopping the discovery components", "error", err)
		}
		n.register = nil
		n.peerPool = nil
		n.discovery = nil
	}

	if err := n.gethNode.Close(); err != nil {
		return err
	}

	n.rpcClient = nil
	// We need to clear `gethNode` because config is passed to `Start()`
	// and may be completely different. Similarly with `config`.
	n.gethNode = nil
	n.config = nil

	err := n.httpServer.Stop()
	if err != nil {
		return err
	}
	n.httpServer = nil

	n.downloader.Stop()
	n.downloader = nil

	if n.db != nil {
		err := n.db.Close()

		n.db = nil

		return err
	}

	n.rpcFiltersSrvc = nil
	n.subscriptionsSrvc = nil
	n.rpcStatsSrvc = nil
	n.accountsSrvc = nil
	n.browsersSrvc = nil
	n.permissionsSrvc = nil
	n.mailserversSrvc = nil
	n.providerSrvc = nil
	n.appMetricsSrvc = nil
	n.walletSrvc = nil
	n.peerSrvc = nil
	n.localNotificationsSrvc = nil
	n.personalSrvc = nil
	n.timeSourceSrvc = nil
	n.wakuSrvc = nil
	n.wakuExtSrvc = nil
	n.wakuV2Srvc = nil
	n.wakuV2ExtSrvc = nil
	n.ensSrvc = nil
	n.communityTokensSrvc = nil
	n.stickersSrvc = nil
	n.publicMethods = make(map[string]bool)
	n.pendingTracker = nil
	n.log.Debug("status node stopped")
	return nil
}

func (n *StatusNode) isDiscoveryRunning() bool {
	return n.register != nil || n.peerPool != nil || n.discovery != nil
}

func (n *StatusNode) stopDiscovery() error {
	n.register.Stop()
	n.peerPool.Stop()
	return n.discovery.Stop()
}

// ResetChainData removes chain data if node is not running.
func (n *StatusNode) ResetChainData(config *params.NodeConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.isRunning() {
		return ErrNodeRunning
	}

	chainDataDir := filepath.Join(config.DataDir, config.Name, "lightchaindata")
	if _, err := os.Stat(chainDataDir); os.IsNotExist(err) {
		return err
	}
	err := os.RemoveAll(chainDataDir)
	if err == nil {
		n.log.Info("Chain data has been removed", "dir", chainDataDir)
	}
	return err
}

// IsRunning confirm that node is running.
func (n *StatusNode) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.isRunning()
}

func (n *StatusNode) isRunning() bool {
	return n.gethNode != nil && n.gethNode.Server() != nil
}

// populateStaticPeers connects current node with our publicly available LES/SHH/Swarm cluster
func (n *StatusNode) populateStaticPeers() error {
	if !n.config.ClusterConfig.Enabled {
		n.log.Info("Static peers are disabled")
		return nil
	}

	for _, enode := range n.config.ClusterConfig.StaticNodes {
		if err := n.addPeer(enode); err != nil {
			n.log.Error("Static peer addition failed", "error", err)
			return err
		}
		n.log.Info("Static peer added", "enode", enode)
	}

	return nil
}

func (n *StatusNode) removeStaticPeers() error {
	if !n.config.ClusterConfig.Enabled {
		n.log.Info("Static peers are disabled")
		return nil
	}

	for _, enode := range n.config.ClusterConfig.StaticNodes {
		if err := n.removePeer(enode); err != nil {
			n.log.Error("Static peer deletion failed", "error", err)
			return err
		}
		n.log.Info("Static peer deleted", "enode", enode)
	}
	return nil
}

// ReconnectStaticPeers removes and adds static peers to a server.
func (n *StatusNode) ReconnectStaticPeers() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isRunning() {
		return ErrNoRunningNode
	}

	if err := n.removeStaticPeers(); err != nil {
		return err
	}

	return n.populateStaticPeers()
}

// AddPeer adds new static peer node
func (n *StatusNode) AddPeer(url string) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.addPeer(url)
}

// addPeer adds new static peer node
func (n *StatusNode) addPeer(url string) error {
	parsedNode, err := enode.ParseV4(url)
	if err != nil {
		return err
	}

	if !n.isRunning() {
		return ErrNoRunningNode
	}

	n.gethNode.Server().AddPeer(parsedNode)

	return nil
}

func (n *StatusNode) removePeer(url string) error {
	parsedNode, err := enode.ParseV4(url)
	if err != nil {
		return err
	}

	if !n.isRunning() {
		return ErrNoRunningNode
	}

	n.gethNode.Server().RemovePeer(parsedNode)

	return nil
}

// PeerCount returns the number of connected peers.
func (n *StatusNode) PeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isRunning() {
		return 0
	}

	return n.gethNode.Server().PeerCount()
}

func (n *StatusNode) ConnectionChanged(state connection.State) {
	if n.wakuExtSrvc != nil {
		n.wakuExtSrvc.ConnectionChanged(state)
	}

	if n.wakuV2ExtSrvc != nil {
		n.wakuV2ExtSrvc.ConnectionChanged(state)
	}
}

// AccountManager exposes reference to node's accounts manager
func (n *StatusNode) AccountManager() (*accounts.Manager, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.gethNode == nil {
		return nil, ErrNoGethNode
	}

	return n.gethNode.AccountManager(), nil
}

// RPCClient exposes reference to RPC client connected to the running node.
func (n *StatusNode) RPCClient() *rpc.Client {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.rpcClient
}

// Discover sets up the discovery for a specific topic.
func (n *StatusNode) Discover(topic string, max, min int) (err error) {
	if n.peerPool == nil {
		return errors.New("peerPool not running")
	}
	return n.peerPool.UpdateTopic(topic, params.Limits{
		Max: max,
		Min: min,
	})
}

func (n *StatusNode) SetAppDB(db *sql.DB) {
	n.appDB = db
}

func (n *StatusNode) SetMultiaccountsDB(db *multiaccounts.Database) {
	n.multiaccountsDB = db
}

func (n *StatusNode) SetWalletDB(db *sql.DB) {
	n.walletDB = db
}
