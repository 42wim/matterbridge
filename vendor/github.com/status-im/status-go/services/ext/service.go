package ext

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"

	commongethtypes "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/api/multiformat"
	"github.com/status-im/status-go/connection"
	"github.com/status-im/status-go/db"
	coretypes "github.com/status-im/status-go/eth-node/core/types"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/protocol/anonmetrics"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/pushnotificationclient"
	"github.com/status-im/status-go/protocol/pushnotificationserver"
	"github.com/status-im/status-go/protocol/transport"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/services/browsers"
	"github.com/status-im/status-go/services/communitytokens"
	"github.com/status-im/status-go/services/ext/mailservers"
	mailserversDB "github.com/status-im/status-go/services/mailservers"
	"github.com/status-im/status-go/services/wallet"
	w_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/wakuv2"
)

const infinityString = "∞"
const providerID = "community"

// EnvelopeEventsHandler used for two different event types.
type EnvelopeEventsHandler interface {
	EnvelopeSent([][]byte)
	EnvelopeExpired([][]byte, error)
	MailServerRequestCompleted(types.Hash, types.Hash, []byte, error)
	MailServerRequestExpired(types.Hash)
}

// Service is a service that provides some additional API to whisper-based protocols like Whisper or Waku.
type Service struct {
	messenger       *protocol.Messenger
	identity        *ecdsa.PrivateKey
	cancelMessenger chan struct{}
	storage         db.TransactionalStorage
	n               types.Node
	rpcClient       *rpc.Client
	config          params.NodeConfig
	mailMonitor     *MailRequestMonitor
	server          *p2p.Server
	peerStore       *mailservers.PeerStore
	accountsDB      *accounts.Database
	multiAccountsDB *multiaccounts.Database
	account         *multiaccounts.Account
}

// Make sure that Service implements node.Service interface.
var _ node.Lifecycle = (*Service)(nil)

func New(
	config params.NodeConfig,
	n types.Node,
	rpcClient *rpc.Client,
	ldb *leveldb.DB,
	mailMonitor *MailRequestMonitor,
	eventSub mailservers.EnvelopeEventSubscriber,
) *Service {
	cache := mailservers.NewCache(ldb)
	peerStore := mailservers.NewPeerStore(cache)
	return &Service{
		storage:     db.NewLevelDBStorage(ldb),
		n:           n,
		rpcClient:   rpcClient,
		config:      config,
		mailMonitor: mailMonitor,
		peerStore:   peerStore,
	}
}

func (s *Service) NodeID() *ecdsa.PrivateKey {
	if s.server == nil {
		return nil
	}
	return s.server.PrivateKey
}

func (s *Service) GetPeer(rawURL string) (*enode.Node, error) {
	if len(rawURL) == 0 {
		return mailservers.GetFirstConnected(s.server, s.peerStore)
	}
	return enode.ParseV4(rawURL)
}

func (s *Service) InitProtocol(nodeName string, identity *ecdsa.PrivateKey, appDb, walletDb *sql.DB, httpServer *server.MediaServer, multiAccountDb *multiaccounts.Database, acc *multiaccounts.Account, accountManager *account.GethManager, rpcClient *rpc.Client, walletService *wallet.Service, communityTokensService *communitytokens.Service, wakuService *wakuv2.Waku, logger *zap.Logger) error {
	var err error
	if !s.config.ShhextConfig.PFSEnabled {
		return nil
	}

	// If Messenger has been already set up, we need to shut it down
	// before we init it again. Otherwise, it will lead to goroutines leakage
	// due to not stopped filters.
	if s.messenger != nil {
		if err := s.messenger.Shutdown(); err != nil {
			return err
		}
	}

	s.identity = identity

	dataDir := filepath.Clean(s.config.ShhextConfig.BackupDisabledDataDir)

	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return err
	}

	envelopesMonitorConfig := &transport.EnvelopesMonitorConfig{
		MaxAttempts:                      s.config.ShhextConfig.MaxMessageDeliveryAttempts,
		AwaitOnlyMailServerConfirmations: s.config.ShhextConfig.MailServerConfirmations,
		IsMailserver: func(peer types.EnodeID) bool {
			return s.peerStore.Exist(peer)
		},
		EnvelopeEventsHandler: EnvelopeSignalHandler{},
		Logger:                logger,
	}
	s.accountsDB, err = accounts.NewDB(appDb)
	if err != nil {
		return err
	}
	s.multiAccountsDB = multiAccountDb
	s.account = acc

	options, err := buildMessengerOptions(s.config, identity, appDb, walletDb, httpServer, s.rpcClient, s.multiAccountsDB, acc, envelopesMonitorConfig, s.accountsDB, walletService, communityTokensService, wakuService, logger, &MessengerSignalsHandler{}, accountManager)
	if err != nil {
		return err
	}

	messenger, err := protocol.NewMessenger(
		nodeName,
		identity,
		s.n,
		s.config.ShhextConfig.InstallationID,
		s.peerStore,
		options...,
	)
	if err != nil {
		return err
	}
	s.messenger = messenger
	s.messenger.SetP2PServer(s.server)
	if s.config.ProcessBackedupMessages {
		s.messenger.EnableBackedupMessagesProcessing()
	}
	return messenger.Init()
}

func (s *Service) StartMessenger() (*protocol.MessengerResponse, error) {
	// Start a loop that retrieves all messages and propagates them to status-mobile.
	s.cancelMessenger = make(chan struct{})
	response, err := s.messenger.Start()
	if err != nil {
		return nil, err
	}
	s.messenger.StartRetrieveMessagesLoop(time.Second, s.cancelMessenger)
	go s.verifyTransactionLoop(30*time.Second, s.cancelMessenger)

	if s.config.ShhextConfig.BandwidthStatsEnabled {
		go s.retrieveStats(5*time.Second, s.cancelMessenger)
	}

	return response, nil
}

func (s *Service) retrieveStats(tick time.Duration, cancel <-chan struct{}) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			response := s.messenger.GetStats()
			PublisherSignalHandler{}.Stats(response)
		case <-cancel:
			return
		}
	}
}

type verifyTransactionClient struct {
	chainID *big.Int
	url     string
}

func (c *verifyTransactionClient) TransactionByHash(ctx context.Context, hash types.Hash) (coretypes.Message, coretypes.TransactionStatus, error) {
	signer := gethtypes.NewLondonSigner(c.chainID)
	client, err := ethclient.Dial(c.url)
	if err != nil {
		return coretypes.Message{}, coretypes.TransactionStatusPending, err
	}

	transaction, pending, err := client.TransactionByHash(ctx, commongethtypes.BytesToHash(hash.Bytes()))
	if err != nil {
		return coretypes.Message{}, coretypes.TransactionStatusPending, err
	}

	message, err := transaction.AsMessage(signer, nil)
	if err != nil {
		return coretypes.Message{}, coretypes.TransactionStatusPending, err
	}
	from := types.BytesToAddress(message.From().Bytes())
	to := types.BytesToAddress(message.To().Bytes())

	if pending {
		return coretypes.NewMessage(
			from,
			&to,
			message.Nonce(),
			message.Value(),
			message.Gas(),
			message.GasPrice(),
			message.Data(),
			message.CheckNonce(),
		), coretypes.TransactionStatusPending, nil
	}

	receipt, err := client.TransactionReceipt(ctx, commongethtypes.BytesToHash(hash.Bytes()))
	if err != nil {
		return coretypes.Message{}, coretypes.TransactionStatusPending, err
	}

	coremessage := coretypes.NewMessage(
		from,
		&to,
		message.Nonce(),
		message.Value(),
		message.Gas(),
		message.GasPrice(),
		message.Data(),
		message.CheckNonce(),
	)

	// Token transfer, check the logs
	if len(coremessage.Data()) != 0 {
		if w_common.IsTokenTransfer(receipt.Logs) {
			return coremessage, coretypes.TransactionStatus(receipt.Status), nil
		}
		return coremessage, coretypes.TransactionStatusFailed, nil
	}

	return coremessage, coretypes.TransactionStatus(receipt.Status), nil
}

func (s *Service) verifyTransactionLoop(tick time.Duration, cancel <-chan struct{}) {
	if s.config.ShhextConfig.VerifyTransactionURL == "" {
		log.Warn("not starting transaction loop")
		return
	}

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	ctx, cancelVerifyTransaction := context.WithCancel(context.Background())

	for {
		select {
		case <-ticker.C:
			accounts, err := s.accountsDB.GetActiveAccounts()
			if err != nil {
				log.Error("failed to retrieve accounts", "err", err)
			}
			var wallets []types.Address
			for _, account := range accounts {
				if account.IsWalletNonWatchOnlyAccount() {
					wallets = append(wallets, types.BytesToAddress(account.Address.Bytes()))
				}
			}

			response, err := s.messenger.ValidateTransactions(ctx, wallets)
			if err != nil {
				log.Error("failed to validate transactions", "err", err)
				continue
			}
			s.messenger.PublishMessengerResponse(response)

		case <-cancel:
			cancelVerifyTransaction()
			return
		}
	}
}

func (s *Service) EnableInstallation(installationID string) error {
	return s.messenger.EnableInstallation(installationID)
}

// DisableInstallation disables an installation for multi-device sync.
func (s *Service) DisableInstallation(installationID string) error {
	return s.messenger.DisableInstallation(installationID)
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []gethrpc.API {
	panic("this is abstract service, use shhext or wakuext implementation")
}

func (s *Service) SetP2PServer(server *p2p.Server) {
	s.server = server
}

// Start is run when a service is started.
// It does nothing in this case but is required by `node.Service` interface.
func (s *Service) Start() error {
	return nil
}

// Stop is run when a service is stopped.
func (s *Service) Stop() error {
	log.Info("Stopping shhext service")
	if s.cancelMessenger != nil {
		select {
		case <-s.cancelMessenger:
			// channel already closed
		default:
			close(s.cancelMessenger)
			s.cancelMessenger = nil
		}
	}

	if s.messenger != nil {
		if err := s.messenger.Shutdown(); err != nil {
			log.Error("failed to stop messenger", "err", err)
			return err
		}
		s.messenger = nil
	}

	return nil
}

func buildMessengerOptions(
	config params.NodeConfig,
	identity *ecdsa.PrivateKey,
	appDb *sql.DB,
	walletDb *sql.DB,
	httpServer *server.MediaServer,
	rpcClient *rpc.Client,
	multiAccounts *multiaccounts.Database,
	account *multiaccounts.Account,
	envelopesMonitorConfig *transport.EnvelopesMonitorConfig,
	accountsDB *accounts.Database,
	walletService *wallet.Service,
	communityTokensService *communitytokens.Service,
	wakuService *wakuv2.Waku,
	logger *zap.Logger,
	messengerSignalsHandler protocol.MessengerSignalsHandler,
	accountManager account.Manager,
) ([]protocol.Option, error) {
	options := []protocol.Option{
		protocol.WithCustomLogger(logger),
		protocol.WithPushNotifications(),
		protocol.WithDatabase(appDb),
		protocol.WithWalletDatabase(walletDb),
		protocol.WithMultiAccounts(multiAccounts),
		protocol.WithMailserversDatabase(mailserversDB.NewDB(appDb)),
		protocol.WithAccount(account),
		protocol.WithBrowserDatabase(browsers.NewDB(appDb)),
		protocol.WithEnvelopesMonitorConfig(envelopesMonitorConfig),
		protocol.WithSignalsHandler(messengerSignalsHandler),
		protocol.WithENSVerificationConfig(config.ShhextConfig.VerifyENSURL, config.ShhextConfig.VerifyENSContractAddress),
		protocol.WithClusterConfig(config.ClusterConfig),
		protocol.WithTorrentConfig(&config.TorrentConfig),
		protocol.WithHTTPServer(httpServer),
		protocol.WithRPCClient(rpcClient),
		protocol.WithMessageCSV(config.OutputMessageCSVEnabled),
		protocol.WithWalletConfig(&config.WalletConfig),
		protocol.WithWalletService(walletService),
		protocol.WithCommunityTokensService(communityTokensService),
		protocol.WithWakuService(wakuService),
		protocol.WithAccountManager(accountManager),
	}

	if config.ShhextConfig.DataSyncEnabled {
		options = append(options, protocol.WithDatasync())
	}

	settings, err := accountsDB.GetSettings()
	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}

	// Generate anon metrics client config
	if settings.AnonMetricsShouldSend {
		keyBytes, err := hex.DecodeString(config.ShhextConfig.AnonMetricsSendID)
		if err != nil {
			return nil, err
		}

		key, err := crypto.UnmarshalPubkey(keyBytes)
		if err != nil {
			return nil, err
		}

		amcc := &anonmetrics.ClientConfig{
			ShouldSend:  true,
			SendAddress: key,
		}
		options = append(options, protocol.WithAnonMetricsClientConfig(amcc))
	}

	// Generate anon metrics server config
	if config.ShhextConfig.AnonMetricsServerEnabled {
		if len(config.ShhextConfig.AnonMetricsServerPostgresURI) == 0 {
			return nil, errors.New("AnonMetricsServerPostgresURI must be set")
		}

		amsc := &anonmetrics.ServerConfig{
			Enabled:     true,
			PostgresURI: config.ShhextConfig.AnonMetricsServerPostgresURI,
		}
		options = append(options, protocol.WithAnonMetricsServerConfig(amsc))
	}

	if settings.TelemetryServerURL != "" {
		options = append(options, protocol.WithTelemetry(settings.TelemetryServerURL))
	}

	if settings.PushNotificationsServerEnabled {
		config := &pushnotificationserver.Config{
			Enabled: true,
			Logger:  logger,
		}
		options = append(options, protocol.WithPushNotificationServerConfig(config))
	}

	var pushNotifServKey []*ecdsa.PublicKey
	for _, d := range config.ShhextConfig.DefaultPushNotificationsServers {
		pushNotifServKey = append(pushNotifServKey, d.PublicKey)
	}

	options = append(options, protocol.WithPushNotificationClientConfig(&pushnotificationclient.Config{
		DefaultServers:             pushNotifServKey,
		BlockMentions:              settings.PushNotificationsBlockMentions,
		SendEnabled:                settings.SendPushNotifications,
		AllowFromContactsOnly:      settings.PushNotificationsFromContactsOnly,
		RemoteNotificationsEnabled: settings.RemotePushNotificationsEnabled,
	}))

	if config.ShhextConfig.VerifyTransactionURL != "" {
		client := &verifyTransactionClient{
			url:     config.ShhextConfig.VerifyTransactionURL,
			chainID: big.NewInt(config.ShhextConfig.VerifyTransactionChainID),
		}
		options = append(options, protocol.WithVerifyTransactionClient(client))
	}

	return options, nil
}

func (s *Service) ConnectionChanged(state connection.State) {
	if s.messenger != nil {
		s.messenger.ConnectionChanged(state)
	}
}

func (s *Service) Messenger() *protocol.Messenger {
	return s.messenger
}

func tokenURIToCommunityID(tokenURI string) string {
	tmpStr := strings.Split(tokenURI, "/")

	// Community NFTs have a tokenURI of the form "compressedCommunityID/tokenID"
	if len(tmpStr) != 2 {
		return ""
	}
	compressedCommunityID := tmpStr[0]

	hexCommunityID, err := multiformat.DeserializeCompressedKey(compressedCommunityID)
	if err != nil {
		return ""
	}

	pubKey, err := common.HexToPubkey(hexCommunityID)
	if err != nil {
		return ""
	}

	communityID := types.EncodeHex(crypto.CompressPubkey(pubKey))

	return communityID
}

func (s *Service) GetCommunityID(tokenURI string) string {
	if tokenURI != "" {
		return tokenURIToCommunityID(tokenURI)
	}
	return ""
}

func (s *Service) FillCollectibleMetadata(collectible *thirdparty.FullCollectibleData) error {
	if s.messenger == nil {
		return fmt.Errorf("messenger not ready")
	}

	if collectible == nil {
		return fmt.Errorf("empty collectible")
	}

	id := collectible.CollectibleData.ID
	communityID := collectible.CollectibleData.CommunityID

	if communityID == "" {
		return fmt.Errorf("invalid communityID")
	}

	// FetchCommunityInfo should have been previously called once to ensure
	// that the latest version of the CommunityDescription is available in the DB
	community, err := s.fetchCommunity(communityID, false)

	if err != nil {
		return err
	}

	if community == nil {
		return nil
	}

	tokenMetadata, err := s.fetchCommunityCollectibleMetadata(community, id.ContractID)

	if err != nil {
		return err
	}

	if tokenMetadata == nil {
		return nil
	}

	communityToken, err := s.fetchCommunityToken(communityID, id.ContractID)
	if err != nil {
		return err
	}

	permission := fetchCommunityCollectiblePermission(community, id)

	privilegesLevel := token.CommunityLevel
	if permission != nil {
		privilegesLevel = permissionTypeToPrivilegesLevel(permission.GetType())
	}

	imagePayload, _ := images.GetPayloadFromURI(tokenMetadata.GetImage())

	collectible.CollectibleData.ContractType = w_common.ContractTypeERC721
	collectible.CollectibleData.Provider = providerID
	collectible.CollectibleData.Name = tokenMetadata.GetName()
	collectible.CollectibleData.Description = tokenMetadata.GetDescription()
	collectible.CollectibleData.ImagePayload = imagePayload
	collectible.CollectibleData.Traits = getCollectibleCommunityTraits(communityToken)

	if collectible.CollectionData == nil {
		collectible.CollectionData = &thirdparty.CollectionData{
			ID:          id.ContractID,
			CommunityID: communityID,
		}
	}
	collectible.CollectionData.ContractType = w_common.ContractTypeERC721
	collectible.CollectionData.Provider = providerID
	collectible.CollectionData.Name = tokenMetadata.GetName()
	collectible.CollectionData.ImagePayload = imagePayload

	collectible.CommunityInfo = communityToInfo(community)

	collectible.CollectibleCommunityInfo = &thirdparty.CollectibleCommunityInfo{
		PrivilegesLevel: privilegesLevel,
	}

	return nil
}

func permissionTypeToPrivilegesLevel(permissionType protobuf.CommunityTokenPermission_Type) token.PrivilegesLevel {
	switch permissionType {
	case protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER:
		return token.OwnerLevel
	case protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER:
		return token.MasterLevel
	default:
		return token.CommunityLevel
	}
}

func communityToInfo(community *communities.Community) *thirdparty.CommunityInfo {
	if community == nil {
		return nil
	}

	return &thirdparty.CommunityInfo{
		CommunityName:         community.Name(),
		CommunityColor:        community.Color(),
		CommunityImagePayload: fetchCommunityImage(community),
	}
}

func (s *Service) FetchCommunityInfo(communityID string) (*thirdparty.CommunityInfo, error) {
	community, err := s.fetchCommunity(communityID, true)
	if err != nil {
		return nil, err
	}

	return communityToInfo(community), nil
}

func (s *Service) fetchCommunity(communityID string, fetchLatest bool) (*communities.Community, error) {
	if s.messenger == nil {
		return nil, fmt.Errorf("messenger not ready")
	}

	// Try to fetch metadata from Messenger communities

	// TODO: we need the shard information in the collectible to be able to retrieve info for
	// communities that have specific shards

	if fetchLatest {
		// Try to fetch the latest version of the Community
		var shard *shard.Shard = nil // TODO: build this with info from token
		// NOTE: The community returned by this function will be nil if
		// the version we have in the DB is the latest available.
		_, err := s.messenger.FetchCommunity(&protocol.FetchCommunityRequest{
			CommunityKey:    communityID,
			Shard:           shard,
			TryDatabase:     false,
			WaitForResponse: true,
		})
		if err != nil {
			return nil, err
		}
	}

	// Get the latest successfully fetched version of the Community
	community, err := s.messenger.FindCommunityInfoFromDB(communityID)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (s *Service) fetchCommunityToken(communityID string, contractID thirdparty.ContractID) (*token.CommunityToken, error) {
	if s.messenger == nil {
		return nil, fmt.Errorf("messenger not ready")
	}

	return s.messenger.GetCommunityToken(communityID, int(contractID.ChainID), contractID.Address.String())
}

func (s *Service) fetchCommunityCollectibleMetadata(community *communities.Community, contractID thirdparty.ContractID) (*protobuf.CommunityTokenMetadata, error) {
	tokensMetadata := community.CommunityTokensMetadata()

	for _, tokenMetadata := range tokensMetadata {
		contractAddresses := tokenMetadata.GetContractAddresses()
		if contractAddresses[uint64(contractID.ChainID)] == contractID.Address.Hex() {
			return tokenMetadata, nil
		}
	}

	return nil, nil
}

func tokenCriterionContainsCollectible(tokenCriterion *protobuf.TokenCriteria, id thirdparty.CollectibleUniqueID) bool {
	// Check if token type matches
	if tokenCriterion.Type != protobuf.CommunityTokenType_ERC721 {
		return false
	}

	for chainID, contractAddressStr := range tokenCriterion.ContractAddresses {
		if chainID != uint64(id.ContractID.ChainID) {
			continue
		}

		contractAddress := commongethtypes.HexToAddress(contractAddressStr)
		if contractAddress != id.ContractID.Address {
			continue
		}

		if len(tokenCriterion.TokenIds) == 0 {
			return true
		}

		for _, tokenID := range tokenCriterion.TokenIds {
			tokenIDBigInt := new(big.Int).SetUint64(tokenID)
			if id.TokenID.Cmp(tokenIDBigInt) == 0 {
				return true
			}
		}
	}

	return false
}

func permissionContainsCollectible(permission *communities.CommunityTokenPermission, id thirdparty.CollectibleUniqueID) bool {
	// See if any token criterion contains the collectible we're looking for
	for _, tokenCriterion := range permission.TokenCriteria {
		if tokenCriterionContainsCollectible(tokenCriterion, id) {
			return true
		}
	}
	return false
}

func fetchCommunityCollectiblePermission(community *communities.Community, id thirdparty.CollectibleUniqueID) *communities.CommunityTokenPermission {
	// Permnission types of interest
	permissionTypes := []protobuf.CommunityTokenPermission_Type{
		protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER,
		protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER,
	}

	for _, permissionType := range permissionTypes {
		permissions := community.TokenPermissionsByType(permissionType)
		// See if any community permission matches the type we're looking for
		for _, permission := range permissions {
			if permissionContainsCollectible(permission, id) {
				return permission
			}
		}
	}

	return nil
}

func fetchCommunityImage(community *communities.Community) []byte {
	imageTypes := []string{
		images.LargeDimName,
		images.SmallDimName,
	}

	communityImages := community.Images()

	for _, imageType := range imageTypes {
		if pbImage, ok := communityImages[imageType]; ok {
			return pbImage.Payload
		}
	}

	return nil
}

func boolToString(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func getCollectibleCommunityTraits(token *token.CommunityToken) []thirdparty.CollectibleTrait {
	if token == nil {
		return make([]thirdparty.CollectibleTrait, 0)
	}

	totalStr := infinityString
	availableStr := infinityString
	if !token.InfiniteSupply {
		totalStr = token.Supply.String()
		// TODO: calculate available supply. See services/communitytokens/api.go
		availableStr = totalStr
	}

	transferableStr := boolToString(token.Transferable)

	destructibleStr := boolToString(token.RemoteSelfDestruct)

	return []thirdparty.CollectibleTrait{
		{
			TraitType: "Symbol",
			Value:     token.Symbol,
		},
		{
			TraitType: "Total",
			Value:     totalStr,
		},
		{
			TraitType: "Available",
			Value:     availableStr,
		},
		{
			TraitType: "Transferable",
			Value:     transferableStr,
		},
		{
			TraitType: "Destructible",
			Value:     destructibleStr,
		},
	}
}
