package node

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/signal"
	"github.com/status-im/status-go/transactions"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p/enode"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/appmetrics"
	"github.com/status-im/status-go/common"
	gethbridge "github.com/status-im/status-go/eth-node/bridge/geth"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/logutils"
	"github.com/status-im/status-go/mailserver"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	accountssvc "github.com/status-im/status-go/services/accounts"
	"github.com/status-im/status-go/services/accounts/settingsevent"
	appmetricsservice "github.com/status-im/status-go/services/appmetrics"
	"github.com/status-im/status-go/services/browsers"
	"github.com/status-im/status-go/services/chat"
	"github.com/status-im/status-go/services/communitytokens"
	"github.com/status-im/status-go/services/ens"
	"github.com/status-im/status-go/services/ext"
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
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/web3provider"
	"github.com/status-im/status-go/timesource"
	"github.com/status-im/status-go/waku"
	wakucommon "github.com/status-im/status-go/waku/common"
	"github.com/status-im/status-go/wakuv2"
)

var (
	// ErrWakuClearIdentitiesFailure clearing whisper identities has failed.
	ErrWakuClearIdentitiesFailure = errors.New("failed to clear waku identities")
	// ErrRPCClientUnavailable is returned if an RPC client can't be retrieved.
	// This is a normal situation when a node is stopped.
	ErrRPCClientUnavailable = errors.New("JSON-RPC client is unavailable")
)

func (b *StatusNode) initServices(config *params.NodeConfig, mediaServer *server.MediaServer) error {
	accountsFeed := &event.Feed{}
	settingsFeed := &event.Feed{}
	accDB, err := accounts.NewDB(b.appDB)
	if err != nil {
		return err
	}

	setSettingsNotifier(accDB, settingsFeed)

	services := []common.StatusService{}
	services = appendIf(config.UpstreamConfig.Enabled, services, b.rpcFiltersService())
	services = append(services, b.subscriptionService())
	services = append(services, b.rpcStatsService())
	services = append(services, b.appmetricsService())
	services = append(services, b.peerService())
	services = append(services, b.personalService())
	services = append(services, b.statusPublicService())
	services = append(services, b.pendingTrackerService(&b.walletFeed))
	services = append(services, b.ensService(b.timeSourceNow()))
	services = append(services, b.CommunityTokensService())
	services = append(services, b.stickersService(accDB))
	services = append(services, b.updatesService())
	services = appendIf(b.appDB != nil && b.multiaccountsDB != nil, services, b.accountsService(accountsFeed, accDB, mediaServer))
	services = appendIf(config.BrowsersConfig.Enabled, services, b.browsersService())
	services = appendIf(config.PermissionsConfig.Enabled, services, b.permissionsService())
	services = appendIf(config.MailserversConfig.Enabled, services, b.mailserversService())
	services = appendIf(config.Web3ProviderConfig.Enabled, services, b.providerService(accDB))
	services = append(services, b.gifService(accDB))
	services = append(services, b.ChatService(accDB))

	// Wallet Service is used by wakuExtSrvc/wakuV2ExtSrvc
	// Keep this initialization before the other two
	if config.WalletConfig.Enabled {
		walletService := b.walletService(accDB, b.appDB, accountsFeed, settingsFeed, &b.walletFeed)
		services = append(services, walletService)
	}

	// CollectiblesManager needs the WakuExt service to get metadata for
	// Community collectibles.
	// Messenger needs the CollectiblesManager to get the list of collectibles owned
	// by a certain account and check community entry permissions.
	// We handle circular dependency between the two by delaying ininitalization of the CommunityCollectibleInfoProvider
	// in the CollectiblesManager.
	if config.WakuConfig.Enabled {
		wakuService, err := b.wakuService(&config.WakuConfig, &config.ClusterConfig)
		if err != nil {
			return err
		}

		services = append(services, wakuService)

		wakuext, err := b.wakuExtService(config)
		if err != nil {
			return err
		}

		b.wakuExtSrvc = wakuext

		services = append(services, wakuext)

		b.SetWalletCommunityInfoProvider(wakuext)
	}

	if config.WakuV2Config.Enabled {
		telemetryServerURL := ""
		if accDB.DB() != nil {
			telemetryServerURL, err = accDB.GetTelemetryServerURL()
			if err != nil {
				return err
			}
		}

		waku2Service, err := b.wakuV2Service(config, telemetryServerURL)
		if err != nil {
			return err
		}
		services = append(services, waku2Service)

		wakuext, err := b.wakuV2ExtService(config)
		if err != nil {
			return err
		}

		b.wakuV2ExtSrvc = wakuext

		services = append(services, wakuext)

		b.SetWalletCommunityInfoProvider(wakuext)
	}

	// We ignore for now local notifications flag as users who are upgrading have no mean to enable it
	lns, err := b.localNotificationsService(config.NetworkID)
	if err != nil {
		return err
	}
	services = append(services, lns)

	b.peerSrvc.SetDiscoverer(b)

	for i := range services {
		b.RegisterLifecycle(services[i])
	}

	b.services = services

	return nil
}

func (b *StatusNode) RegisterLifecycle(s common.StatusService) {
	b.addPublicMethods(s.APIs())
	b.gethNode.RegisterAPIs(s.APIs())
	b.gethNode.RegisterProtocols(s.Protocols())
	b.gethNode.RegisterLifecycle(s)
}

// Add through reflection a list of public methods so we can check when the
// user makes a call if they are allowed
func (b *StatusNode) addPublicMethods(apis []gethrpc.API) {
	for _, api := range apis {
		if api.Public {
			addSuitableCallbacks(reflect.ValueOf(api.Service), api.Namespace, b.publicMethods)
		}
	}
}

func (b *StatusNode) nodeBridge() types.Node {
	return gethbridge.NewNodeBridge(b.gethNode, b.wakuSrvc, b.wakuV2Srvc)
}

func (b *StatusNode) wakuExtService(config *params.NodeConfig) (*wakuext.Service, error) {
	if b.gethNode == nil {
		return nil, errors.New("geth node not initialized")
	}

	if b.wakuExtSrvc == nil {
		b.wakuExtSrvc = wakuext.New(*config, b.nodeBridge(), b.rpcClient, ext.EnvelopeSignalHandler{}, b.db)
	}

	b.wakuExtSrvc.SetP2PServer(b.gethNode.Server())
	return b.wakuExtSrvc, nil
}

func (b *StatusNode) wakuV2ExtService(config *params.NodeConfig) (*wakuv2ext.Service, error) {
	if b.gethNode == nil {
		return nil, errors.New("geth node not initialized")
	}
	if b.wakuV2ExtSrvc == nil {
		b.wakuV2ExtSrvc = wakuv2ext.New(*config, b.nodeBridge(), b.rpcClient, ext.EnvelopeSignalHandler{}, b.db)
	}

	b.wakuV2ExtSrvc.SetP2PServer(b.gethNode.Server())
	return b.wakuV2ExtSrvc, nil
}

func (b *StatusNode) statusPublicService() *status.Service {
	if b.statusPublicSrvc == nil {
		b.statusPublicSrvc = status.New()
	}
	return b.statusPublicSrvc
}

func (b *StatusNode) StatusPublicService() *status.Service {
	return b.statusPublicSrvc
}

func (b *StatusNode) AccountService() *accountssvc.Service {
	return b.accountsSrvc
}

func (b *StatusNode) BrowserService() *browsers.Service {
	return b.browsersSrvc
}

func (b *StatusNode) EnsService() *ens.Service {
	return b.ensSrvc
}

func (b *StatusNode) WakuService() *waku.Waku {
	return b.wakuSrvc
}

func (b *StatusNode) WakuExtService() *wakuext.Service {
	return b.wakuExtSrvc
}

func (b *StatusNode) WakuV2ExtService() *wakuv2ext.Service {
	return b.wakuV2ExtSrvc
}
func (b *StatusNode) WakuV2Service() *wakuv2.Waku {
	return b.wakuV2Srvc
}

func (b *StatusNode) wakuService(wakuCfg *params.WakuConfig, clusterCfg *params.ClusterConfig) (*waku.Waku, error) {
	if b.wakuSrvc == nil {
		cfg := &waku.Config{
			MaxMessageSize:         wakucommon.DefaultMaxMessageSize,
			BloomFilterMode:        wakuCfg.BloomFilterMode,
			FullNode:               wakuCfg.FullNode,
			SoftBlacklistedPeerIDs: wakuCfg.SoftBlacklistedPeerIDs,
			MinimumAcceptedPoW:     params.WakuMinimumPoW,
			EnableConfirmations:    wakuCfg.EnableConfirmations,
		}

		if wakuCfg.MaxMessageSize > 0 {
			cfg.MaxMessageSize = wakuCfg.MaxMessageSize
		}
		if wakuCfg.MinimumPoW > 0 {
			cfg.MinimumAcceptedPoW = wakuCfg.MinimumPoW
		}

		w := waku.New(cfg, logutils.ZapLogger())

		if wakuCfg.EnableRateLimiter {
			r := wakuRateLimiter(wakuCfg, clusterCfg)
			w.RegisterRateLimiter(r)
		}

		if timesource := b.timeSource(); timesource != nil {
			w.SetTimeSource(timesource.Now)
		}

		// enable mail service
		if wakuCfg.EnableMailServer {
			if err := registerWakuMailServer(w, wakuCfg); err != nil {
				return nil, fmt.Errorf("failed to register WakuMailServer: %v", err)
			}
		}

		if wakuCfg.LightClient {
			emptyBloomFilter := make([]byte, 64)
			if err := w.SetBloomFilter(emptyBloomFilter); err != nil {
				return nil, err
			}
		}
		b.wakuSrvc = w
	}
	return b.wakuSrvc, nil

}

func (b *StatusNode) wakuV2Service(nodeConfig *params.NodeConfig, telemetryServerURL string) (*wakuv2.Waku, error) {
	if b.wakuV2Srvc == nil {
		cfg := &wakuv2.Config{
			MaxMessageSize:          wakucommon.DefaultMaxMessageSize,
			Host:                    nodeConfig.WakuV2Config.Host,
			Port:                    nodeConfig.WakuV2Config.Port,
			LightClient:             nodeConfig.WakuV2Config.LightClient,
			KeepAliveInterval:       nodeConfig.WakuV2Config.KeepAliveInterval,
			Rendezvous:              nodeConfig.Rendezvous,
			WakuNodes:               nodeConfig.ClusterConfig.WakuNodes,
			EnableStore:             nodeConfig.WakuV2Config.EnableStore,
			StoreCapacity:           nodeConfig.WakuV2Config.StoreCapacity,
			StoreSeconds:            nodeConfig.WakuV2Config.StoreSeconds,
			DiscoveryLimit:          nodeConfig.WakuV2Config.DiscoveryLimit,
			DiscV5BootstrapNodes:    nodeConfig.ClusterConfig.DiscV5BootstrapNodes,
			Nameserver:              nodeConfig.WakuV2Config.Nameserver,
			UDPPort:                 nodeConfig.WakuV2Config.UDPPort,
			AutoUpdate:              nodeConfig.WakuV2Config.AutoUpdate,
			DefaultShardPubsubTopic: shard.DefaultShardPubsubTopic(),
			UseShardAsDefaultTopic:  nodeConfig.WakuV2Config.UseShardAsDefaultTopic,
			TelemetryServerURL:      telemetryServerURL,
			ClusterID:               nodeConfig.ClusterConfig.ClusterID,
		}

		// Configure peer exchange and discv5 settings based on node type
		if cfg.LightClient {
			cfg.EnablePeerExchangeServer = false
			cfg.EnablePeerExchangeClient = true
			cfg.EnableDiscV5 = false
		} else {
			cfg.EnablePeerExchangeServer = true
			cfg.EnablePeerExchangeClient = false
			cfg.EnableDiscV5 = true
		}

		if nodeConfig.WakuV2Config.MaxMessageSize > 0 {
			cfg.MaxMessageSize = nodeConfig.WakuV2Config.MaxMessageSize
		}

		w, err := wakuv2.New(nodeConfig.NodeKey, nodeConfig.ClusterConfig.Fleet, cfg, logutils.ZapLogger(), b.appDB, b.timeSource(), signal.SendHistoricMessagesRequestFailed, signal.SendPeerStats)

		if err != nil {
			return nil, err
		}
		b.wakuV2Srvc = w
	}

	return b.wakuV2Srvc, nil
}

func setSettingsNotifier(db *accounts.Database, feed *event.Feed) {
	db.SetSettingsNotifier(func(setting settings.SettingField, val interface{}) {
		feed.Send(settingsevent.Event{
			Type:    settingsevent.EventTypeChanged,
			Setting: setting,
			Value:   val,
		})
	})
}

func wakuRateLimiter(wakuCfg *params.WakuConfig, clusterCfg *params.ClusterConfig) *wakucommon.PeerRateLimiter {
	enodes := append(
		parseNodes(clusterCfg.StaticNodes),
		parseNodes(clusterCfg.TrustedMailServers)...,
	)
	var (
		ips     []string
		peerIDs []enode.ID
	)
	for _, item := range enodes {
		ips = append(ips, item.IP().String())
		peerIDs = append(peerIDs, item.ID())
	}
	return wakucommon.NewPeerRateLimiter(
		&wakucommon.PeerRateLimiterConfig{
			PacketLimitPerSecIP:     wakuCfg.PacketRateLimitIP,
			PacketLimitPerSecPeerID: wakuCfg.PacketRateLimitPeerID,
			BytesLimitPerSecIP:      wakuCfg.BytesRateLimitIP,
			BytesLimitPerSecPeerID:  wakuCfg.BytesRateLimitPeerID,
			WhitelistedIPs:          ips,
			WhitelistedPeerIDs:      peerIDs,
		},
		&wakucommon.MetricsRateLimiterHandler{},
		&wakucommon.DropPeerRateLimiterHandler{
			Tolerance: wakuCfg.RateLimitTolerance,
		},
	)
}

func (b *StatusNode) rpcFiltersService() *rpcfilters.Service {
	if b.rpcFiltersSrvc == nil {
		b.rpcFiltersSrvc = rpcfilters.New(b)
	}
	return b.rpcFiltersSrvc
}

func (b *StatusNode) subscriptionService() *subscriptions.Service {
	if b.subscriptionsSrvc == nil {

		b.subscriptionsSrvc = subscriptions.New(func() *rpc.Client { return b.RPCClient() })
	}
	return b.subscriptionsSrvc
}

func (b *StatusNode) rpcStatsService() *rpcstats.Service {
	if b.rpcStatsSrvc == nil {
		b.rpcStatsSrvc = rpcstats.New()
	}

	return b.rpcStatsSrvc
}

func (b *StatusNode) accountsService(accountsFeed *event.Feed, accDB *accounts.Database, mediaServer *server.MediaServer) *accountssvc.Service {
	if b.accountsSrvc == nil {
		b.accountsSrvc = accountssvc.NewService(
			accDB,
			b.multiaccountsDB,
			b.gethAccountManager,
			b.config,
			accountsFeed,
			mediaServer,
		)
	}

	return b.accountsSrvc
}

func (b *StatusNode) browsersService() *browsers.Service {
	if b.browsersSrvc == nil {
		b.browsersSrvc = browsers.NewService(browsers.NewDB(b.appDB))
	}
	return b.browsersSrvc
}

func (b *StatusNode) ensService(timesource func() time.Time) *ens.Service {
	if b.ensSrvc == nil {
		b.ensSrvc = ens.NewService(b.rpcClient, b.gethAccountManager, b.pendingTracker, b.config, b.appDB, timesource)
	}
	return b.ensSrvc
}

func (b *StatusNode) pendingTrackerService(walletFeed *event.Feed) *transactions.PendingTxTracker {
	if b.pendingTracker == nil {
		b.pendingTracker = transactions.NewPendingTxTracker(b.walletDB, b.rpcClient, b.rpcFiltersSrvc, walletFeed, transactions.PendingCheckInterval)
		if b.transactor != nil {
			b.transactor.SetPendingTracker(b.pendingTracker)
		}
	}
	return b.pendingTracker
}

func (b *StatusNode) CommunityTokensService() *communitytokens.Service {
	if b.communityTokensSrvc == nil {
		b.communityTokensSrvc = communitytokens.NewService(b.rpcClient, b.gethAccountManager, b.pendingTracker, b.config, b.appDB)
	}
	return b.communityTokensSrvc
}

func (b *StatusNode) stickersService(accountDB *accounts.Database) *stickers.Service {
	if b.stickersSrvc == nil {
		b.stickersSrvc = stickers.NewService(accountDB, b.rpcClient, b.gethAccountManager, b.config, b.downloader, b.httpServer, b.pendingTracker)
	}
	return b.stickersSrvc
}

func (b *StatusNode) updatesService() *updates.Service {
	if b.updatesSrvc == nil {
		b.updatesSrvc = updates.NewService(b.ensService(b.timeSourceNow()))
	}

	return b.updatesSrvc
}

func (b *StatusNode) gifService(accountsDB *accounts.Database) *gif.Service {
	if b.gifSrvc == nil {
		b.gifSrvc = gif.NewService(accountsDB)
	}
	return b.gifSrvc
}

func (b *StatusNode) ChatService(accountsDB *accounts.Database) *chat.Service {
	if b.chatSrvc == nil {
		b.chatSrvc = chat.NewService(accountsDB)
	}
	return b.chatSrvc
}

func (b *StatusNode) permissionsService() *permissions.Service {
	if b.permissionsSrvc == nil {
		b.permissionsSrvc = permissions.NewService(permissions.NewDB(b.appDB))
	}
	return b.permissionsSrvc
}

func (b *StatusNode) mailserversService() *mailservers.Service {
	if b.mailserversSrvc == nil {

		b.mailserversSrvc = mailservers.NewService(mailservers.NewDB(b.appDB))
	}
	return b.mailserversSrvc
}

func (b *StatusNode) providerService(accountsDB *accounts.Database) *web3provider.Service {
	web3S := web3provider.NewService(b.appDB, accountsDB, b.rpcClient, b.config, b.gethAccountManager, b.rpcFiltersSrvc, b.transactor)
	if b.providerSrvc == nil {
		b.providerSrvc = web3S
	}
	return b.providerSrvc
}

func (b *StatusNode) appmetricsService() common.StatusService {
	if b.appMetricsSrvc == nil {
		b.appMetricsSrvc = appmetricsservice.NewService(appmetrics.NewDB(b.appDB))
	}
	return b.appMetricsSrvc
}

func (b *StatusNode) WalletService() *wallet.Service {
	return b.walletSrvc
}

func (b *StatusNode) SetWalletCommunityInfoProvider(provider thirdparty.CommunityInfoProvider) {
	if b.walletSrvc != nil {
		b.walletSrvc.SetWalletCommunityInfoProvider(provider)
	}
}

func (b *StatusNode) walletService(accountsDB *accounts.Database, appDB *sql.DB, accountsFeed *event.Feed, settingsFeed *event.Feed, walletFeed *event.Feed) *wallet.Service {
	if b.walletSrvc == nil {
		b.walletSrvc = wallet.NewService(
			b.walletDB, accountsDB, appDB, b.rpcClient, accountsFeed, settingsFeed, b.gethAccountManager, b.transactor, b.config,
			b.ensService(b.timeSourceNow()),
			b.stickersService(accountsDB),
			b.pendingTracker,
			walletFeed,
			b.httpServer,
		)
	}
	return b.walletSrvc
}

func (b *StatusNode) localNotificationsService(network uint64) (*localnotifications.Service, error) {
	var err error
	if b.localNotificationsSrvc == nil {
		b.localNotificationsSrvc, err = localnotifications.NewService(b.appDB, transfer.NewDB(b.walletDB), network)
		if err != nil {
			return nil, err
		}
	}
	return b.localNotificationsSrvc, nil
}

func (b *StatusNode) peerService() *peer.Service {
	if b.peerSrvc == nil {
		b.peerSrvc = peer.New()
	}
	return b.peerSrvc
}

func registerWakuMailServer(wakuService *waku.Waku, config *params.WakuConfig) (err error) {
	var mailServer mailserver.WakuMailServer
	wakuService.RegisterMailServer(&mailServer)

	return mailServer.Init(wakuService, config)
}

func appendIf(condition bool, services []common.StatusService, service common.StatusService) []common.StatusService {
	if !condition {
		return services
	}
	return append(services, service)
}

func (b *StatusNode) RPCFiltersService() *rpcfilters.Service {
	return b.rpcFiltersSrvc
}

func (b *StatusNode) PendingTracker() *transactions.PendingTxTracker {
	return b.pendingTracker
}

func (b *StatusNode) StopLocalNotifications() error {
	if b.localNotificationsSrvc == nil {
		return nil
	}

	if b.localNotificationsSrvc.IsStarted() {
		err := b.localNotificationsSrvc.Stop()
		if err != nil {
			b.log.Error("LocalNotifications service stop failed on StopLocalNotifications", "error", err)
			return nil
		}
	}

	return nil
}

func (b *StatusNode) StartLocalNotifications() error {
	if b.localNotificationsSrvc == nil {
		return nil
	}

	if b.walletSrvc == nil {
		return nil
	}

	if !b.localNotificationsSrvc.IsStarted() {
		err := b.localNotificationsSrvc.Start()

		if err != nil {
			b.log.Error("LocalNotifications service start failed on StartLocalNotifications", "error", err)
			return nil
		}
	}

	err := b.localNotificationsSrvc.SubscribeWallet(&b.walletFeed)

	if err != nil {
		b.log.Error("LocalNotifications service could not subscribe to wallet on StartLocalNotifications", "error", err)
		return nil
	}

	return nil
}

// `personal_sign` and `personal_ecRecover` methods are important to
// keep DApps working.
// Usually, they are provided by an ETH or a LES service, but when using
// upstream, we don't start any of these, so we need to start our own
// implementation.

func (b *StatusNode) personalService() *personal.Service {
	if b.personalSrvc == nil {
		b.personalSrvc = personal.New(b.accountsManager)
	}
	return b.personalSrvc
}

func (b *StatusNode) timeSource() *timesource.NTPTimeSource {

	if b.timeSourceSrvc == nil {
		b.timeSourceSrvc = timesource.Default()
	}
	return b.timeSourceSrvc
}

func (b *StatusNode) timeSourceNow() func() time.Time {
	return b.timeSource().Now
}

func (b *StatusNode) Cleanup() error {
	if b.wakuSrvc != nil {
		if err := b.wakuSrvc.DeleteKeyPairs(); err != nil {
			return fmt.Errorf("%s: %v", ErrWakuClearIdentitiesFailure, err)
		}
	}

	if b.Config() != nil && b.Config().WalletConfig.Enabled {
		if b.walletSrvc != nil {
			if b.walletSrvc.IsStarted() {
				err := b.walletSrvc.Stop()
				if err != nil {
					return err
				}
			}
		}
	}

	if b.ensSrvc != nil {
		err := b.ensSrvc.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

type RPCCall struct {
	Method string `json:"method"`
}

func (b *StatusNode) CallPrivateRPC(inputJSON string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.rpcClient == nil {
		return "", ErrRPCClientUnavailable
	}

	return b.rpcClient.CallRaw(inputJSON), nil
}

// CallRPC calls public methods on the node, we register public methods
// in a map and check if they can be called in this function
func (b *StatusNode) CallRPC(inputJSON string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.rpcClient == nil {
		return "", ErrRPCClientUnavailable
	}

	rpcCall := &RPCCall{}
	err := json.Unmarshal([]byte(inputJSON), rpcCall)
	if err != nil {
		return "", err
	}

	if rpcCall.Method == "" || !b.publicMethods[rpcCall.Method] {
		return ErrRPCMethodUnavailable, nil
	}

	return b.rpcClient.CallRaw(inputJSON), nil
}
