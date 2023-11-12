package communities

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/golang/protobuf/proto"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/status-im/status-go/account"
	utils "github.com/status-im/status-go/common"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	community_token "github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/protocol/encryption"
	"github.com/status-im/status-go/protocol/ens"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/protocol/transport"
	"github.com/status-im/status-go/services/communitytokens"
	"github.com/status-im/status-go/services/wallet/bigint"
	walletcommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/signal"
)

var defaultAnnounceList = [][]string{
	{"udp://tracker.opentrackr.org:1337/announce"},
	{"udp://tracker.openbittorrent.com:6969/announce"},
}
var pieceLength = 100 * 1024

const maxArchiveSizeInBytes = 30000000

var memberPermissionsCheckInterval = 1 * time.Hour
var validateInterval = 2 * time.Minute

// Used for testing only
func SetValidateInterval(duration time.Duration) {
	validateInterval = duration
}

// errors
var (
	ErrTorrentTimedout                 = errors.New("torrent has timed out")
	ErrCommunityRequestAlreadyRejected = errors.New("that user was already rejected from the community")
	ErrInvalidClock                    = errors.New("invalid clock to cancel request to join")
)

type Manager struct {
	persistence                      *Persistence
	encryptor                        *encryption.Protocol
	ensSubscription                  chan []*ens.VerificationRecord
	subscriptions                    []chan *Subscription
	ensVerifier                      *ens.Verifier
	ownerVerifier                    OwnerVerifier
	identity                         *ecdsa.PrivateKey
	installationID                   string
	accountsManager                  account.Manager
	tokenManager                     TokenManager
	collectiblesManager              CollectiblesManager
	logger                           *zap.Logger
	stdoutLogger                     *zap.Logger
	transport                        *transport.Transport
	timesource                       common.TimeSource
	quit                             chan struct{}
	torrentConfig                    *params.TorrentConfig
	torrentClient                    *torrent.Client
	walletConfig                     *params.WalletConfig
	communityTokensService           communitytokens.ServiceInterface
	historyArchiveTasksWaitGroup     sync.WaitGroup
	historyArchiveTasks              sync.Map // stores `chan struct{}`
	periodicMembersReevaluationTasks sync.Map // stores `chan struct{}`
	torrentTasks                     map[string]metainfo.Hash
	historyArchiveDownloadTasks      map[string]*HistoryArchiveDownloadTask
	stopped                          bool
	RekeyInterval                    time.Duration
	PermissionChecker                PermissionChecker
	keyDistributor                   KeyDistributor
}

type HistoryArchiveDownloadTask struct {
	CancelChan chan struct{}
	Waiter     sync.WaitGroup
	m          sync.RWMutex
	Cancelled  bool
}

func (t *HistoryArchiveDownloadTask) IsCancelled() bool {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.Cancelled
}

func (t *HistoryArchiveDownloadTask) Cancel() {
	t.m.Lock()
	defer t.m.Unlock()
	t.Cancelled = true
	close(t.CancelChan)
}

type managerOptions struct {
	accountsManager        account.Manager
	tokenManager           TokenManager
	collectiblesManager    CollectiblesManager
	walletConfig           *params.WalletConfig
	communityTokensService communitytokens.ServiceInterface
	permissionChecker      PermissionChecker
}

type TokenManager interface {
	GetBalancesByChain(ctx context.Context, accounts, tokens []gethcommon.Address, chainIDs []uint64) (map[uint64]map[gethcommon.Address]map[gethcommon.Address]*hexutil.Big, error)
	FindOrCreateTokenByAddress(ctx context.Context, chainID uint64, address gethcommon.Address) *token.Token
	GetAllChainIDs() ([]uint64, error)
}

type DefaultTokenManager struct {
	tokenManager *token.Manager
}

func NewDefaultTokenManager(tm *token.Manager) *DefaultTokenManager {
	return &DefaultTokenManager{tokenManager: tm}
}

type BalancesByChain = map[uint64]map[gethcommon.Address]map[gethcommon.Address]*hexutil.Big

func (m *DefaultTokenManager) GetAllChainIDs() ([]uint64, error) {
	networks, err := m.tokenManager.RPCClient.NetworkManager.Get(false)
	if err != nil {
		return nil, err
	}

	areTestNetworksEnabled, err := m.tokenManager.RPCClient.NetworkManager.GetTestNetworksEnabled()
	if err != nil {
		return nil, err
	}

	chainIDs := make([]uint64, 0)
	for _, network := range networks {
		if areTestNetworksEnabled == network.IsTest {
			chainIDs = append(chainIDs, network.ChainID)
		}
	}
	return chainIDs, nil
}

type CollectiblesManager interface {
	FetchBalancesByOwnerAndContractAddress(ctx context.Context, chainID walletcommon.ChainID, ownerAddress gethcommon.Address, contractAddresses []gethcommon.Address) (thirdparty.TokenBalancesPerContractAddress, error)
}

func (m *DefaultTokenManager) GetBalancesByChain(ctx context.Context, accounts, tokenAddresses []gethcommon.Address, chainIDs []uint64) (BalancesByChain, error) {
	clients, err := m.tokenManager.RPCClient.EthClients(chainIDs)
	if err != nil {
		return nil, err
	}

	resp, err := m.tokenManager.GetBalancesByChain(context.Background(), clients, accounts, tokenAddresses)
	return resp, err
}

func (m *DefaultTokenManager) FindOrCreateTokenByAddress(ctx context.Context, chainID uint64, address gethcommon.Address) *token.Token {
	return m.tokenManager.FindOrCreateTokenByAddress(ctx, chainID, address)
}

type ManagerOption func(*managerOptions)

func WithAccountManager(accountsManager account.Manager) ManagerOption {
	return func(opts *managerOptions) {
		opts.accountsManager = accountsManager
	}
}

func WithPermissionChecker(permissionChecker PermissionChecker) ManagerOption {
	return func(opts *managerOptions) {
		opts.permissionChecker = permissionChecker
	}
}

func WithCollectiblesManager(collectiblesManager CollectiblesManager) ManagerOption {
	return func(opts *managerOptions) {
		opts.collectiblesManager = collectiblesManager
	}
}

func WithTokenManager(tokenManager TokenManager) ManagerOption {
	return func(opts *managerOptions) {
		opts.tokenManager = tokenManager
	}
}

func WithWalletConfig(walletConfig *params.WalletConfig) ManagerOption {
	return func(opts *managerOptions) {
		opts.walletConfig = walletConfig
	}
}

func WithCommunityTokensService(communityTokensService communitytokens.ServiceInterface) ManagerOption {
	return func(opts *managerOptions) {
		opts.communityTokensService = communityTokensService
	}
}

type OwnerVerifier interface {
	SafeGetSignerPubKey(ctx context.Context, chainID uint64, communityID string) (string, error)
}

func NewManager(identity *ecdsa.PrivateKey, installationID string, db *sql.DB, encryptor *encryption.Protocol, logger *zap.Logger, ensverifier *ens.Verifier, ownerVerifier OwnerVerifier, transport *transport.Transport, timesource common.TimeSource, keyDistributor KeyDistributor, torrentConfig *params.TorrentConfig, opts ...ManagerOption) (*Manager, error) {
	if identity == nil {
		return nil, errors.New("empty identity")
	}

	if timesource == nil {
		return nil, errors.New("no timesource")
	}

	var err error
	if logger == nil {
		if logger, err = zap.NewDevelopment(); err != nil {
			return nil, errors.Wrap(err, "failed to create a logger")
		}
	}

	stdoutLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create archive logger")
	}

	managerConfig := managerOptions{}
	for _, opt := range opts {
		opt(&managerConfig)
	}

	manager := &Manager{
		logger:                      logger,
		stdoutLogger:                stdoutLogger,
		encryptor:                   encryptor,
		identity:                    identity,
		installationID:              installationID,
		ownerVerifier:               ownerVerifier,
		quit:                        make(chan struct{}),
		transport:                   transport,
		timesource:                  timesource,
		torrentConfig:               torrentConfig,
		torrentTasks:                make(map[string]metainfo.Hash),
		historyArchiveDownloadTasks: make(map[string]*HistoryArchiveDownloadTask),
		keyDistributor:              keyDistributor,
	}

	manager.persistence = &Persistence{
		db:                      db,
		recordBundleToCommunity: manager.dbRecordBundleToCommunity,
	}

	if managerConfig.accountsManager != nil {
		manager.accountsManager = managerConfig.accountsManager
	}

	if managerConfig.collectiblesManager != nil {
		manager.collectiblesManager = managerConfig.collectiblesManager
	}

	if managerConfig.tokenManager != nil {
		manager.tokenManager = managerConfig.tokenManager
	}

	if managerConfig.walletConfig != nil {
		manager.walletConfig = managerConfig.walletConfig
	}

	if managerConfig.communityTokensService != nil {
		manager.communityTokensService = managerConfig.communityTokensService
	}

	if ensverifier != nil {

		sub := ensverifier.Subscribe()
		manager.ensSubscription = sub
		manager.ensVerifier = ensverifier
	}

	if managerConfig.permissionChecker != nil {
		manager.PermissionChecker = managerConfig.permissionChecker
	} else {
		manager.PermissionChecker = &DefaultPermissionChecker{
			tokenManager:        manager.tokenManager,
			collectiblesManager: manager.collectiblesManager,
			logger:              logger,
			ensVerifier:         ensverifier,
		}
	}

	return manager, nil
}

func (m *Manager) LogStdout(msg string, fields ...zap.Field) {
	m.stdoutLogger.Info(msg, fields...)
	m.logger.Debug(msg, fields...)
}

type archiveMDSlice []*archiveMetadata

type archiveMetadata struct {
	hash string
	from uint64
}

func (md archiveMDSlice) Len() int {
	return len(md)
}

func (md archiveMDSlice) Swap(i, j int) {
	md[i], md[j] = md[j], md[i]
}

func (md archiveMDSlice) Less(i, j int) bool {
	return md[i].from > md[j].from
}

type Subscription struct {
	Community                                *Community
	CreatingHistoryArchivesSignal            *signal.CreatingHistoryArchivesSignal
	HistoryArchivesCreatedSignal             *signal.HistoryArchivesCreatedSignal
	NoHistoryArchivesCreatedSignal           *signal.NoHistoryArchivesCreatedSignal
	HistoryArchivesSeedingSignal             *signal.HistoryArchivesSeedingSignal
	HistoryArchivesUnseededSignal            *signal.HistoryArchivesUnseededSignal
	HistoryArchiveDownloadedSignal           *signal.HistoryArchiveDownloadedSignal
	DownloadingHistoryArchivesStartedSignal  *signal.DownloadingHistoryArchivesStartedSignal
	DownloadingHistoryArchivesFinishedSignal *signal.DownloadingHistoryArchivesFinishedSignal
	ImportingHistoryArchiveMessagesSignal    *signal.ImportingHistoryArchiveMessagesSignal
	CommunityEventsMessage                   *CommunityEventsMessage
	CommunityEventsMessageInvalidClock       *CommunityEventsMessageInvalidClockSignal
	AcceptedRequestsToJoin                   []types.HexBytes
	RejectedRequestsToJoin                   []types.HexBytes
	CommunityPrivilegedMemberSyncMessage     *CommunityPrivilegedMemberSyncMessage
	TokenCommunityValidated                  *CommunityResponse
}

type CommunityResponse struct {
	Community       *Community                             `json:"community"`
	Changes         *CommunityChanges                      `json:"changes"`
	RequestsToJoin  []*RequestToJoin                       `json:"requestsToJoin"`
	FailedToDecrypt []*CommunityPrivateDataFailedToDecrypt `json:"-"`
}

type CommunityEventsMessageInvalidClockSignal struct {
	Community              *Community
	CommunityEventsMessage *CommunityEventsMessage
}

func (m *Manager) Subscribe() chan *Subscription {
	subscription := make(chan *Subscription, 100)
	m.subscriptions = append(m.subscriptions, subscription)
	return subscription
}

func (m *Manager) Start() error {
	m.stopped = false
	if m.ensVerifier != nil {
		m.runENSVerificationLoop()
	}

	if m.ownerVerifier != nil {
		m.runOwnerVerificationLoop()
	}

	if m.torrentConfig != nil && m.torrentConfig.Enabled {
		err := m.StartTorrentClient()
		if err != nil {
			m.LogStdout("couldn't start torrent client", zap.Error(err))
		}
	}

	return nil
}

func (m *Manager) runENSVerificationLoop() {
	go func() {
		for {
			select {
			case <-m.quit:
				m.logger.Debug("quitting ens verification loop")
				return
			case records, more := <-m.ensSubscription:
				if !more {
					m.logger.Debug("no more ens records, quitting")
					return
				}
				m.logger.Info("received records", zap.Any("records", records))
			}
		}
	}()
}

// Only for testing
func (m *Manager) CommunitiesToValidate() (map[string][]communityToValidate, error) { // nolint: golint
	return m.persistence.getCommunitiesToValidate()
}

func (m *Manager) runOwnerVerificationLoop() {
	m.logger.Info("starting owner verification loop")
	go func() {
		for {
			select {
			case <-m.quit:
				m.logger.Debug("quitting owner verification loop")
				return
			case <-time.After(validateInterval):
				// If ownerverifier is nil, we skip, this is useful for testing
				if m.ownerVerifier == nil {
					continue
				}

				communitiesToValidate, err := m.persistence.getCommunitiesToValidate()

				if err != nil {
					m.logger.Error("failed to fetch communities to validate", zap.Error(err))
					continue
				}
				for id, communities := range communitiesToValidate {
					m.logger.Info("validating communities", zap.String("id", id), zap.Int("count", len(communities)))

					_, _ = m.validateCommunity(communities)
				}
			}
		}
	}()
}

func (m *Manager) ValidateCommunityByID(communityID types.HexBytes) (*CommunityResponse, error) {
	communitiesToValidate, err := m.persistence.getCommunityToValidateByID(communityID)
	if err != nil {
		m.logger.Error("failed to validate community by ID", zap.String("id", communityID.String()), zap.Error(err))
		return nil, err
	}
	return m.validateCommunity(communitiesToValidate)

}

func (m *Manager) validateCommunity(communityToValidateData []communityToValidate) (*CommunityResponse, error) {
	for _, community := range communityToValidateData {
		signer, description, err := UnwrapCommunityDescriptionMessage(community.payload)
		if err != nil {
			m.logger.Error("failed to unwrap community", zap.Error(err))
			continue
		}

		chainID := CommunityDescriptionTokenOwnerChainID(description)
		if chainID == 0 {
			// This should not happen
			m.logger.Error("chain id is 0, ignoring")
			continue
		}

		m.logger.Info("validating community", zap.String("id", types.EncodeHex(community.id)), zap.String("signer", common.PubkeyToHex(signer)))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		owner, err := m.ownerVerifier.SafeGetSignerPubKey(ctx, chainID, types.EncodeHex(community.id))
		if err != nil {
			m.logger.Error("failed to get owner", zap.Error(err))
			continue
		}

		ownerPK, err := common.HexToPubkey(owner)
		if err != nil {
			m.logger.Error("failed to convert pk string to ecdsa", zap.Error(err))
			continue
		}

		// TODO: handle shards
		response, err := m.HandleCommunityDescriptionMessage(signer, description, community.payload, ownerPK, nil)
		if err != nil {
			m.logger.Error("failed to handle community", zap.Error(err))
			err = m.persistence.DeleteCommunityToValidate(community.id, community.clock)
			if err != nil {
				m.logger.Error("failed to delete community to validate", zap.Error(err))
			}
			continue
		}

		if response != nil {

			m.logger.Info("community validated", zap.String("id", types.EncodeHex(community.id)), zap.String("signer", common.PubkeyToHex(signer)))
			m.publish(&Subscription{TokenCommunityValidated: response})
			err := m.persistence.DeleteCommunitiesToValidateByCommunityID(community.id)
			if err != nil {
				m.logger.Error("failed to delete communities to validate", zap.Error(err))
			}
			return response, nil
		}
	}

	return nil, nil
}

func (m *Manager) Stop() error {
	m.stopped = true
	close(m.quit)
	for _, c := range m.subscriptions {
		close(c)
	}
	m.StopTorrentClient()
	return nil
}

func (m *Manager) SetTorrentConfig(config *params.TorrentConfig) {
	m.torrentConfig = config
}

// getTCPandUDPport will return the same port number given if != 0,
// otherwise, it will attempt to find a free random tcp and udp port using
// the same number for both protocols
func (m *Manager) getTCPandUDPport(portNumber int) (int, error) {
	if portNumber != 0 {
		return portNumber, nil
	}

	// Find free port
	for i := 0; i < 10; i++ {
		port := func() int {
			tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort("localhost", "0"))
			if err != nil {
				m.logger.Warn("unable to resolve tcp addr: %v", zap.Error(err))
				return 0
			}

			tcpListener, err := net.ListenTCP("tcp", tcpAddr)
			if err != nil {
				m.logger.Warn("unable to listen on addr", zap.Stringer("addr", tcpAddr), zap.Error(err))
				return 0
			}
			defer tcpListener.Close()

			port := tcpListener.Addr().(*net.TCPAddr).Port

			udpAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("localhost", fmt.Sprintf("%d", port)))
			if err != nil {
				m.logger.Warn("unable to resolve udp addr: %v", zap.Error(err))
				return 0
			}

			udpListener, err := net.ListenUDP("udp", udpAddr)
			if err != nil {
				m.logger.Warn("unable to listen on addr", zap.Stringer("addr", udpAddr), zap.Error(err))
				return 0
			}
			defer udpListener.Close()

			return port
		}()

		if port != 0 {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no free port found")
}

func (m *Manager) StartTorrentClient() error {
	if m.torrentConfig == nil {
		return fmt.Errorf("can't start torrent client: missing torrentConfig")
	}

	if m.TorrentClientStarted() {
		return nil
	}

	port, err := m.getTCPandUDPport(m.torrentConfig.Port)
	if err != nil {
		return err
	}

	config := torrent.NewDefaultClientConfig()
	config.SetListenAddr(":" + fmt.Sprint(port))
	config.Seed = true

	config.DataDir = m.torrentConfig.DataDir

	if _, err := os.Stat(m.torrentConfig.DataDir); os.IsNotExist(err) {
		err := os.MkdirAll(m.torrentConfig.DataDir, 0700)
		if err != nil {
			return err
		}
	}

	m.logger.Info("Starting torrent client", zap.Any("port", port))
	// Instantiating the client will make it bootstrap and listen eagerly,
	// so no go routine is needed here
	client, err := torrent.NewClient(config)
	if err != nil {
		return err
	}
	m.torrentClient = client
	return nil
}

func (m *Manager) StopTorrentClient() []error {
	if m.TorrentClientStarted() {
		m.StopHistoryArchiveTasksIntervals()
		m.logger.Info("Stopping torrent client")
		errs := m.torrentClient.Close()
		if len(errs) > 0 {
			return errs
		}
		m.torrentClient = nil
	}
	return make([]error, 0)
}

func (m *Manager) TorrentClientStarted() bool {
	return m.torrentClient != nil
}

func (m *Manager) publish(subscription *Subscription) {
	if m.stopped {
		return
	}
	for _, s := range m.subscriptions {
		select {
		case s <- subscription:
		default:
			m.logger.Warn("subscription channel full, dropping message")
		}
	}
}

func (m *Manager) All() ([]*Community, error) {
	return m.persistence.AllCommunities(&m.identity.PublicKey)
}

type CommunityShard struct {
	CommunityID string       `json:"communityID"`
	Shard       *shard.Shard `json:"shard"`
}

type CuratedCommunities struct {
	ContractCommunities         []string
	ContractFeaturedCommunities []string
}

type KnownCommunitiesResponse struct {
	ContractCommunities         []string              `json:"contractCommunities"`
	ContractFeaturedCommunities []string              `json:"contractFeaturedCommunities"`
	Descriptions                map[string]*Community `json:"communities"`
	UnknownCommunities          []string              `json:"unknownCommunities"`
}

func (m *Manager) GetStoredDescriptionForCommunities(communityIDs []string) (*KnownCommunitiesResponse, error) {
	response := &KnownCommunitiesResponse{
		Descriptions: make(map[string]*Community),
	}

	for i := range communityIDs {
		communityID := communityIDs[i]
		communityIDBytes, err := types.DecodeHex(communityID)
		if err != nil {
			return nil, err
		}

		community, err := m.GetByID(types.HexBytes(communityIDBytes))
		if err != nil {
			return nil, err
		}

		if community != nil {
			response.Descriptions[community.IDString()] = community
		} else {
			response.UnknownCommunities = append(response.UnknownCommunities, communityID)
		}

		response.ContractCommunities = append(response.ContractCommunities, communityID)
	}

	return response, nil
}

func (m *Manager) Joined() ([]*Community, error) {
	return m.persistence.JoinedCommunities(&m.identity.PublicKey)
}

func (m *Manager) Spectated() ([]*Community, error) {
	return m.persistence.SpectatedCommunities(&m.identity.PublicKey)
}

func (m *Manager) CommunityUpdateLastOpenedAt(communityID types.HexBytes, timestamp int64) (*Community, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, err
	}

	err = m.persistence.UpdateLastOpenedAt(community.ID(), timestamp)
	if err != nil {
		return nil, err
	}
	community.UpdateLastOpenedAt(timestamp)
	return community, nil
}

func (m *Manager) JoinedAndPendingCommunitiesWithRequests() ([]*Community, error) {
	return m.persistence.JoinedAndPendingCommunitiesWithRequests(&m.identity.PublicKey)
}

func (m *Manager) DeletedCommunities() ([]*Community, error) {
	return m.persistence.DeletedCommunities(&m.identity.PublicKey)
}

func (m *Manager) Controlled() ([]*Community, error) {
	communities, err := m.persistence.CommunitiesWithPrivateKey(&m.identity.PublicKey)
	if err != nil {
		return nil, err
	}

	controlled := make([]*Community, 0, len(communities))

	for _, c := range communities {
		if c.IsControlNode() {
			controlled = append(controlled, c)
		}
	}

	return controlled, nil
}

// CreateCommunity takes a description, generates an ID for it, saves it and return it
func (m *Manager) CreateCommunity(request *requests.CreateCommunity, publish bool) (*Community, error) {

	description, err := request.ToCommunityDescription()
	if err != nil {
		return nil, err
	}

	description.Members = make(map[string]*protobuf.CommunityMember)
	description.Members[common.PubkeyToHex(&m.identity.PublicKey)] = &protobuf.CommunityMember{Roles: []protobuf.CommunityMember_Roles{protobuf.CommunityMember_ROLE_OWNER}}

	err = ValidateCommunityDescription(description)
	if err != nil {
		return nil, err
	}

	description.Clock = 1

	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	description.ID = types.EncodeHex(crypto.CompressPubkey(&key.PublicKey))

	config := Config{
		ID:                   &key.PublicKey,
		PrivateKey:           key,
		ControlNode:          &key.PublicKey,
		ControlDevice:        true,
		Logger:               m.logger,
		Joined:               true,
		JoinedAt:             time.Now().Unix(),
		MemberIdentity:       &m.identity.PublicKey,
		CommunityDescription: description,
		Shard:                nil,
		LastOpenedAt:         0,
	}

	var descriptionEncryptor DescriptionEncryptor
	if m.encryptor != nil {
		descriptionEncryptor = m
	}
	community, err := New(config, m.timesource, descriptionEncryptor)
	if err != nil {
		return nil, err
	}

	// We join any community we create
	community.Join()

	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, err
	}

	// Mark this device as the control node
	syncControlNode := &protobuf.SyncCommunityControlNode{
		Clock:          1,
		InstallationId: m.installationID,
	}
	err = m.SaveSyncControlNode(community.ID(), syncControlNode)
	if err != nil {
		return nil, err
	}

	if publish {
		m.publish(&Subscription{Community: community})
	}

	return community, nil
}

func (m *Manager) CreateCommunityTokenPermission(request *requests.CreateCommunityTokenPermission) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	// ensure key is generated before marshaling,
	// as it requires key to encrypt description
	if community.IsControlNode() && m.encryptor != nil {
		key, err := m.encryptor.GenerateHashRatchetKey(community.ID())
		if err != nil {
			return nil, nil, err
		}
		keyID, err := key.GetKeyID()
		if err != nil {
			return nil, nil, err
		}
		m.logger.Info("generate key for token", zap.String("group-id", types.Bytes2Hex(community.ID())), zap.String("key-id", types.Bytes2Hex(keyID)))
	}

	community, changes, err := m.createCommunityTokenPermission(request, community)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) EditCommunityTokenPermission(request *requests.EditCommunityTokenPermission) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	tokenPermission := request.ToCommunityTokenPermission()

	changes, err := community.UpsertTokenPermission(&tokenPermission)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) ReevaluateMembers(community *Community) (map[protobuf.CommunityMember_Roles][]*ecdsa.PublicKey, error) {
	becomeMemberPermissions := community.TokenPermissionsByType(protobuf.CommunityTokenPermission_BECOME_MEMBER)
	becomeAdminPermissions := community.TokenPermissionsByType(protobuf.CommunityTokenPermission_BECOME_ADMIN)
	becomeTokenMasterPermissions := community.TokenPermissionsByType(protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER)

	hasMemberPermissions := len(becomeMemberPermissions) > 0

	newPrivilegedRoles := make(map[protobuf.CommunityMember_Roles][]*ecdsa.PublicKey)
	newPrivilegedRoles[protobuf.CommunityMember_ROLE_TOKEN_MASTER] = []*ecdsa.PublicKey{}
	newPrivilegedRoles[protobuf.CommunityMember_ROLE_ADMIN] = []*ecdsa.PublicKey{}

	for memberKey := range community.Members() {
		memberPubKey, err := common.HexToPubkey(memberKey)
		if err != nil {
			return nil, err
		}

		if memberKey == common.PubkeyToHex(&m.identity.PublicKey) || community.IsMemberOwner(memberPubKey) {
			continue
		}

		isCurrentRoleTokenMaster := community.IsMemberTokenMaster(memberPubKey)
		isCurrentRoleAdmin := community.IsMemberAdmin(memberPubKey)
		requestID := CalculateRequestID(memberKey, community.ID())
		revealedAccounts, err := m.persistence.GetRequestToJoinRevealedAddresses(requestID)
		if err != nil {
			return nil, err
		}

		memberHasWallet := len(revealedAccounts) > 0

		// Check if user has privilege role without sharing the account to controlNode
		// or user treated as a member without wallet in closed community
		if !memberHasWallet && (hasMemberPermissions || isCurrentRoleTokenMaster || isCurrentRoleAdmin) {
			_, err = community.RemoveUserFromOrg(memberPubKey)
			if err != nil {
				return nil, err
			}
			continue
		}

		accountsAndChainIDs := revealedAccountsToAccountsAndChainIDsCombination(revealedAccounts)

		isNewRoleTokenMaster, err := m.ReevaluatePrivilegedMember(community, becomeTokenMasterPermissions, accountsAndChainIDs, memberPubKey,
			protobuf.CommunityMember_ROLE_TOKEN_MASTER, isCurrentRoleTokenMaster)

		if err != nil {
			return nil, err
		}

		if isNewRoleTokenMaster {
			if !isCurrentRoleTokenMaster {
				newPrivilegedRoles[protobuf.CommunityMember_ROLE_TOKEN_MASTER] =
					append(newPrivilegedRoles[protobuf.CommunityMember_ROLE_TOKEN_MASTER], memberPubKey)
			}
			// Skip further validation if user has TokenMaster permissions
			continue
		}

		isNewRoleAdmin, err := m.ReevaluatePrivilegedMember(community, becomeAdminPermissions, accountsAndChainIDs, memberPubKey,
			protobuf.CommunityMember_ROLE_ADMIN, isCurrentRoleAdmin)

		if err != nil {
			return nil, err
		}

		if isNewRoleAdmin {
			if !isCurrentRoleAdmin {
				newPrivilegedRoles[protobuf.CommunityMember_ROLE_ADMIN] =
					append(newPrivilegedRoles[protobuf.CommunityMember_ROLE_TOKEN_MASTER], memberPubKey)
			}
			// Skip further validation if user has Admin permissions
			continue
		}

		if hasMemberPermissions {
			permissionResponse, err := m.PermissionChecker.CheckPermissions(becomeMemberPermissions, accountsAndChainIDs, true)
			if err != nil {
				return nil, err
			}

			if !permissionResponse.Satisfied {
				_, err = community.RemoveUserFromOrg(memberPubKey)
				if err != nil {
					return nil, err
				}
				// Skip channels validation if user has been removed
				continue
			}
		}

		// Validate channel permissions
		for channelID := range community.Chats() {
			chatID := community.ChatID(channelID)

			viewOnlyPermissions := community.ChannelTokenPermissionsByType(chatID, protobuf.CommunityTokenPermission_CAN_VIEW_CHANNEL)
			viewAndPostPermissions := community.ChannelTokenPermissionsByType(chatID, protobuf.CommunityTokenPermission_CAN_VIEW_AND_POST_CHANNEL)

			if len(viewOnlyPermissions) == 0 && len(viewAndPostPermissions) == 0 {
				// ensure all members are added back if channel permissions were removed
				_, err = community.PopulateChatWithAllMembers(channelID)
				if err != nil {
					return nil, err
				}
				continue
			}

			response, err := m.checkChannelPermissions(viewOnlyPermissions, viewAndPostPermissions, accountsAndChainIDs, true)
			if err != nil {
				return nil, err
			}

			isMemberAlreadyInChannel := community.IsMemberInChat(memberPubKey, channelID)

			if response.ViewOnlyPermissions.Satisfied || response.ViewAndPostPermissions.Satisfied {
				if !isMemberAlreadyInChannel {
					_, err := community.AddMemberToChat(channelID, memberPubKey, []protobuf.CommunityMember_Roles{})
					if err != nil {
						return nil, err
					}
				}
			} else if isMemberAlreadyInChannel {
				_, err := community.RemoveUserFromChat(memberPubKey, channelID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return newPrivilegedRoles, m.saveAndPublish(community)
}

func (m *Manager) ReevaluateMembersPeriodically(communityID types.HexBytes) {
	if _, exists := m.periodicMembersReevaluationTasks.Load(communityID.String()); exists {
		return
	}

	cancel := make(chan struct{})
	m.periodicMembersReevaluationTasks.Store(communityID.String(), cancel)

	ticker := time.NewTicker(memberPermissionsCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			community, err := m.GetByID(communityID)
			if err != nil {
				m.logger.Debug("can't validate member permissions, community was not found", zap.Error(err))
				m.periodicMembersReevaluationTasks.Delete(communityID.String())
			}

			if err = m.ReevaluateCommunityMembersPermissions(community); err != nil {
				m.logger.Debug("failed to check member permissions", zap.Error(err))
				continue
			}

		case <-cancel:
			m.periodicMembersReevaluationTasks.Delete(communityID.String())
			return
		}
	}
}

func (m *Manager) DeleteCommunityTokenPermission(request *requests.DeleteCommunityTokenPermission) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	changes, err := community.DeleteTokenPermission(request.PermissionID)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) ReevaluateCommunityMembersPermissions(community *Community) error {
	if community == nil {
		return ErrOrgNotFound
	}

	// TODO: Control node needs to be notified to do a permission check if TokenMasters did airdrop
	// of the token which is using in a community permissions
	if !community.IsControlNode() {
		return ErrNotEnoughPermissions
	}

	newPrivilegedMembers, err := m.ReevaluateMembers(community)
	if err != nil {
		return err
	}

	return m.shareRequestsToJoinWithNewPrivilegedMembers(community, newPrivilegedMembers)
}

func (m *Manager) DeleteCommunity(id types.HexBytes) error {
	err := m.persistence.DeleteCommunity(id)
	if err != nil {
		return err
	}
	return m.persistence.DeleteCommunitySettings(id)
}

func (m *Manager) UpdateShard(community *Community, shard *shard.Shard, clock uint64) error {
	community.config.Shard = shard
	if shard == nil {
		return m.persistence.DeleteCommunityShard(community.ID())
	}

	return m.persistence.SaveCommunityShard(community.ID(), shard, clock)
}

// SetShard assigns a shard to a community
func (m *Manager) SetShard(communityID types.HexBytes, shard *shard.Shard) (*Community, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, err
	}

	community.increaseClock()

	err = m.UpdateShard(community, shard, community.Clock())
	if err != nil {
		return nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) UpdatePubsubTopicPrivateKey(topic string, privKey *ecdsa.PrivateKey) error {
	if privKey != nil {
		return m.transport.StorePubsubTopicKey(topic, privKey)
	}

	return m.transport.RemovePubsubTopicKey(topic)
}

// EditCommunity takes a description, updates the community with the description,
// saves it and returns it
func (m *Manager) EditCommunity(request *requests.EditCommunity) (*Community, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, err
	}

	newDescription, err := request.ToCommunityDescription()
	if err != nil {
		return nil, fmt.Errorf("Can't create community description: %v", err)
	}

	// If permissions weren't explicitly set on original request, use existing ones
	if newDescription.Permissions.Access == protobuf.CommunityPermissions_UNKNOWN_ACCESS {
		newDescription.Permissions.Access = community.config.CommunityDescription.Permissions.Access
	}
	// Use existing images for the entries that were not updated
	// NOTE: This will NOT allow deletion of the community image; it will need to
	// be handled separately.
	for imageName := range community.config.CommunityDescription.Identity.Images {
		_, exists := newDescription.Identity.Images[imageName]
		if !exists {
			// If no image was set in ToCommunityDescription then Images is nil.
			if newDescription.Identity.Images == nil {
				newDescription.Identity.Images = make(map[string]*protobuf.IdentityImage)
			}
			newDescription.Identity.Images[imageName] = community.config.CommunityDescription.Identity.Images[imageName]
		}
	}
	// TODO: handle delete image (if needed)

	err = ValidateCommunityDescription(newDescription)
	if err != nil {
		return nil, err
	}

	if !(community.IsControlNode() || community.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_EDIT)) {
		return nil, ErrNotAuthorized
	}

	// Edit the community values
	community.Edit(newDescription)
	if err != nil {
		return nil, err
	}

	if community.IsControlNode() {
		community.increaseClock()
	} else {
		err := community.addNewCommunityEvent(community.ToCommunityEditCommunityEvent(newDescription))
		if err != nil {
			return nil, err
		}
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) RemovePrivateKey(id types.HexBytes) (*Community, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return community, err
	}

	if !community.IsControlNode() {
		return community, ErrNotControlNode
	}

	community.config.PrivateKey = nil
	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return community, err
	}
	return community, nil
}

func (m *Manager) ExportCommunity(id types.HexBytes) (*ecdsa.PrivateKey, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	if !community.IsControlNode() {
		return nil, ErrNotControlNode
	}

	return community.config.PrivateKey, nil
}

func (m *Manager) ImportCommunity(key *ecdsa.PrivateKey, clock uint64) (*Community, error) {
	communityID := crypto.CompressPubkey(&key.PublicKey)

	community, err := m.GetByID(communityID)
	if err != nil && err != ErrOrgNotFound {
		return nil, err
	}

	if community == nil {
		createCommunityRequest := requests.CreateCommunity{
			Membership: protobuf.CommunityPermissions_MANUAL_ACCEPT,
			Name:       "unknown imported",
		}

		description, err := createCommunityRequest.ToCommunityDescription()
		if err != nil {
			return nil, err
		}

		err = ValidateCommunityDescription(description)
		if err != nil {
			return nil, err
		}

		description.Clock = 1
		description.ID = types.EncodeHex(communityID)

		config := Config{
			ID:                   &key.PublicKey,
			PrivateKey:           key,
			ControlNode:          &key.PublicKey,
			ControlDevice:        true,
			Logger:               m.logger,
			Joined:               true,
			JoinedAt:             time.Now().Unix(),
			MemberIdentity:       &m.identity.PublicKey,
			CommunityDescription: description,
			LastOpenedAt:         0,
		}

		var descriptionEncryptor DescriptionEncryptor
		if m.encryptor != nil {
			descriptionEncryptor = m
		}
		community, err = New(config, m.timesource, descriptionEncryptor)
		if err != nil {
			return nil, err
		}
	} else {
		community.config.PrivateKey = key
		community.config.ControlDevice = true
	}

	community.Join()
	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, err
	}

	// Mark this device as the control node
	syncControlNode := &protobuf.SyncCommunityControlNode{
		Clock:          clock,
		InstallationId: m.installationID,
	}
	err = m.SaveSyncControlNode(community.ID(), syncControlNode)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) CreateChat(communityID types.HexBytes, chat *protobuf.CommunityChat, publish bool, thirdPartyID string) (*CommunityChanges, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, err
	}
	chatID := uuid.New().String()
	if thirdPartyID != "" {
		chatID = chatID + thirdPartyID
	}

	changes, err := community.CreateChat(chatID, chat)
	if err != nil {
		return nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return changes, nil
}

func (m *Manager) EditChat(communityID types.HexBytes, chatID string, chat *protobuf.CommunityChat) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, nil, err
	}

	// Remove communityID prefix from chatID if exists
	if strings.HasPrefix(chatID, communityID.String()) {
		chatID = strings.TrimPrefix(chatID, communityID.String())
	}

	changes, err := community.EditChat(chatID, chat)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) DeleteChat(communityID types.HexBytes, chatID string) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, nil, err
	}

	// Remove communityID prefix from chatID if exists
	if strings.HasPrefix(chatID, communityID.String()) {
		chatID = strings.TrimPrefix(chatID, communityID.String())
	}
	changes, err := community.DeleteChat(chatID)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) CreateCategory(request *requests.CreateCommunityCategory, publish bool) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	categoryID := uuid.New().String()
	if request.ThirdPartyID != "" {
		categoryID = categoryID + request.ThirdPartyID
	}

	// Remove communityID prefix from chatID if exists
	for i, cid := range request.ChatIDs {
		if strings.HasPrefix(cid, request.CommunityID.String()) {
			request.ChatIDs[i] = strings.TrimPrefix(cid, request.CommunityID.String())
		}
	}

	changes, err := community.CreateCategory(categoryID, request.CategoryName, request.ChatIDs)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) EditCategory(request *requests.EditCommunityCategory) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	// Remove communityID prefix from chatID if exists
	for i, cid := range request.ChatIDs {
		if strings.HasPrefix(cid, request.CommunityID.String()) {
			request.ChatIDs[i] = strings.TrimPrefix(cid, request.CommunityID.String())
		}
	}

	changes, err := community.EditCategory(request.CategoryID, request.CategoryName, request.ChatIDs)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) EditChatFirstMessageTimestamp(communityID types.HexBytes, chatID string, timestamp uint32) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, nil, err
	}

	// Remove communityID prefix from chatID if exists
	if strings.HasPrefix(chatID, communityID.String()) {
		chatID = strings.TrimPrefix(chatID, communityID.String())
	}

	changes, err := community.UpdateChatFirstMessageTimestamp(chatID, timestamp)
	if err != nil {
		return nil, nil, err
	}

	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, nil, err
	}

	// Advertise changes
	m.publish(&Subscription{Community: community})

	return community, changes, nil
}

func (m *Manager) ReorderCategories(request *requests.ReorderCommunityCategories) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	changes, err := community.ReorderCategories(request.CategoryID, request.Position)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) ReorderChat(request *requests.ReorderCommunityChat) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	// Remove communityID prefix from chatID if exists
	if strings.HasPrefix(request.ChatID, request.CommunityID.String()) {
		request.ChatID = strings.TrimPrefix(request.ChatID, request.CommunityID.String())
	}

	changes, err := community.ReorderChat(request.CategoryID, request.ChatID, request.Position)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil
}

func (m *Manager) DeleteCategory(request *requests.DeleteCommunityCategory) (*Community, *CommunityChanges, error) {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	changes, err := community.DeleteCategory(request.CategoryID)
	if err != nil {
		return nil, nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, nil, err
	}

	return changes.Community, changes, nil
}

func (m *Manager) GenerateRequestsToJoinForAutoApprovalOnNewOwnership(communityID types.HexBytes, kickedMembers map[string]*protobuf.CommunityMember) ([]*RequestToJoin, error) {
	var requestsToJoin []*RequestToJoin
	clock := uint64(time.Now().Unix())
	for pubKeyStr := range kickedMembers {
		requestToJoin := &RequestToJoin{
			PublicKey:        pubKeyStr,
			Clock:            clock,
			CommunityID:      communityID,
			State:            RequestToJoinStateAwaitingAddresses,
			Our:              true,
			RevealedAccounts: make([]*protobuf.RevealedAccount, 0),
		}

		requestToJoin.CalculateID()

		requestsToJoin = append(requestsToJoin, requestToJoin)
	}

	return requestsToJoin, m.persistence.SaveRequestsToJoin(requestsToJoin)
}

func (m *Manager) Queue(signer *ecdsa.PublicKey, community *Community, clock uint64, payload []byte) error {

	m.logger.Info("queuing community", zap.String("id", community.IDString()), zap.String("signer", common.PubkeyToHex(signer)))

	communityToValidate := communityToValidate{
		id:         community.ID(),
		clock:      clock,
		payload:    payload,
		validateAt: uint64(time.Now().UnixNano()),
		signer:     crypto.CompressPubkey(signer),
	}
	err := m.persistence.SaveCommunityToValidate(communityToValidate)
	if err != nil {
		m.logger.Error("failed to save community", zap.Error(err))
		return err
	}

	return nil
}

func (m *Manager) HandleCommunityDescriptionMessage(signer *ecdsa.PublicKey, description *protobuf.CommunityDescription, payload []byte, verifiedOwner *ecdsa.PublicKey, communityShard *protobuf.Shard) (*CommunityResponse, error) {
	m.logger.Debug("HandleCommunityDescriptionMessage", zap.String("communityID", description.ID), zap.Uint64("clock", description.Clock))

	if signer == nil {
		return nil, errors.New("signer can't be nil")
	}

	var id []byte
	var err error
	if len(description.ID) != 0 {
		id, err = types.DecodeHex(description.ID)
		if err != nil {
			return nil, err
		}
	} else {
		// Backward compatibility
		id = crypto.CompressPubkey(signer)
	}

	failedToDecrypt, err := m.preprocessDescription(id, description)
	if err != nil {
		return nil, err
	}

	community, err := m.GetByID(id)
	if err != nil && err != ErrOrgNotFound {
		return nil, err
	}

	// We don't process failed to decrypt if the whole metadata is encrypted
	// and we joined the community already
	if community != nil && community.Joined() && len(failedToDecrypt) != 0 && description != nil && len(description.Members) == 0 {
		return &CommunityResponse{FailedToDecrypt: failedToDecrypt}, nil
	}

	// We should queue only if the community has a token owner, and the owner has been verified
	hasTokenOwnership := HasTokenOwnership(description)
	shouldQueue := hasTokenOwnership && verifiedOwner == nil

	if community == nil {
		pubKey, err := crypto.DecompressPubkey(id)
		if err != nil {
			return nil, err
		}
		config := Config{
			CommunityDescription:                description,
			Logger:                              m.logger,
			CommunityDescriptionProtocolMessage: payload,
			MemberIdentity:                      &m.identity.PublicKey,
			ID:                                  pubKey,
			ControlNode:                         signer,
			Shard:                               shard.FromProtobuff(communityShard),
		}

		var descriptionEncryptor DescriptionEncryptor
		if m.encryptor != nil {
			descriptionEncryptor = m
		}
		community, err = New(config, m.timesource, descriptionEncryptor)
		if err != nil {
			return nil, err
		}

		// A new community, we need to check if we need to validate async.
		// That would be the case if it has a contract. We queue everything and process separately.
		if shouldQueue {
			return nil, m.Queue(signer, community, description.Clock, payload)
		}
	} else {
		// only queue if already known control node is different than the signer
		// and if the clock is greater
		shouldQueue = shouldQueue && !common.IsPubKeyEqual(community.ControlNode(), signer) &&
			community.config.CommunityDescription.Clock < description.Clock
		if shouldQueue {
			return nil, m.Queue(signer, community, description.Clock, payload)
		}
	}

	if hasTokenOwnership && verifiedOwner != nil {
		// Override verified owner
		m.logger.Info("updating verified owner",
			zap.String("communityID", community.IDString()),
			zap.String("verifiedOwner", common.PubkeyToHex(verifiedOwner)),
			zap.String("signer", common.PubkeyToHex(signer)),
			zap.String("controlNode", common.PubkeyToHex(community.ControlNode())),
		)

		// If we are not the verified owner anymore, drop the private key
		if !common.IsPubKeyEqual(verifiedOwner, &m.identity.PublicKey) {
			community.config.PrivateKey = nil
		}

		// new control node will be set in the 'UpdateCommunityDescription'
		if !common.IsPubKeyEqual(verifiedOwner, signer) {
			return nil, ErrNotAuthorized
		}
	} else if !common.IsPubKeyEqual(community.ControlNode(), signer) {
		return nil, ErrNotAuthorized
	}

	r, err := m.handleCommunityDescriptionMessageCommon(community, description, payload, verifiedOwner)
	if err != nil {
		return nil, err
	}
	r.FailedToDecrypt = failedToDecrypt
	return r, nil
}

func (m *Manager) preprocessDescription(id types.HexBytes, description *protobuf.CommunityDescription) ([]*CommunityPrivateDataFailedToDecrypt, error) {
	response, err := decryptDescription(id, m, description, m.logger)
	if err != nil {
		return response, err
	}

	// Workaround for https://github.com/status-im/status-desktop/issues/12188
	hydrateChannelsMembers(types.EncodeHex(id), description)

	return response, nil
}

func (m *Manager) handleCommunityDescriptionMessageCommon(community *Community, description *protobuf.CommunityDescription, payload []byte, newControlNode *ecdsa.PublicKey) (*CommunityResponse, error) {

	changes, err := community.UpdateCommunityDescription(description, payload, newControlNode)
	if err != nil {
		return nil, err
	}

	if err = m.handleCommunityTokensMetadata(community); err != nil {
		return nil, err
	}

	hasCommunityArchiveInfo, err := m.persistence.HasCommunityArchiveInfo(community.ID())
	if err != nil {
		return nil, err
	}

	cdMagnetlinkClock := community.config.CommunityDescription.ArchiveMagnetlinkClock
	if !hasCommunityArchiveInfo {
		err = m.persistence.SaveCommunityArchiveInfo(community.ID(), cdMagnetlinkClock, 0)
		if err != nil {
			return nil, err
		}
	} else {
		magnetlinkClock, err := m.persistence.GetMagnetlinkMessageClock(community.ID())
		if err != nil {
			return nil, err
		}
		if cdMagnetlinkClock > magnetlinkClock {
			err = m.persistence.UpdateMagnetlinkMessageClock(community.ID(), cdMagnetlinkClock)
			if err != nil {
				return nil, err
			}
		}
	}

	pkString := common.PubkeyToHex(&m.identity.PublicKey)
	if m.tokenManager != nil && description.CommunityTokensMetadata != nil && len(description.CommunityTokensMetadata) > 0 {
		for _, tokenMetadata := range description.CommunityTokensMetadata {
			if tokenMetadata.TokenType != protobuf.CommunityTokenType_ERC20 {
				continue
			}

			for chainID, address := range tokenMetadata.ContractAddresses {
				_ = m.tokenManager.FindOrCreateTokenByAddress(context.Background(), chainID, gethcommon.HexToAddress(address))
			}
		}
	}

	// If the community require membership, we set whether we should leave/join the community after a state change
	if community.ManualAccept() || community.AutoAccept() {
		if changes.HasNewMember(pkString) {
			hasPendingRequest, err := m.persistence.HasPendingRequestsToJoinForUserAndCommunity(pkString, changes.Community.ID())
			if err != nil {
				return nil, err
			}
			// If there's any pending request, we should join the community
			// automatically
			changes.ShouldMemberJoin = hasPendingRequest
		}

		if changes.HasMemberLeft(pkString) {
			// If we joined previously the community, that means we have been kicked
			changes.MemberKicked = community.Joined()
		}
	}

	err = m.persistence.DeleteCommunityEvents(community.ID())
	if err != nil {
		return nil, err
	}
	community.config.EventsData = nil

	// Set Joined if we are part of the member list
	if !community.Joined() && community.hasMember(&m.identity.PublicKey) {
		changes.ShouldMemberJoin = true
	}

	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, err
	}

	// We mark our requests as completed, though maybe we should mark
	// any request for any user that has been added as completed
	if err := m.markRequestToJoinAsAccepted(&m.identity.PublicKey, community); err != nil {
		return nil, err
	}
	// Check if there's a change and we should be joining

	return &CommunityResponse{
		Community: community,
		Changes:   changes,
	}, nil
}

func (m *Manager) signEvents(community *Community) error {
	for i := range community.config.EventsData.Events {
		communityEvent := &community.config.EventsData.Events[i]
		if communityEvent.Signature == nil || len(communityEvent.Signature) == 0 {
			err := communityEvent.Sign(m.identity)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) validateAndFilterEvents(community *Community, events []CommunityEvent) []CommunityEvent {
	validatedEvents := make([]CommunityEvent, 0, len(events))

	validateEvent := func(event *CommunityEvent) error {
		signer, err := event.RecoverSigner()
		if err != nil {
			return err
		}

		err = community.ValidateEvent(event, signer)
		if err != nil {
			return err
		}

		return nil
	}

	for i := range events {
		if err := validateEvent(&events[i]); err == nil {
			validatedEvents = append(validatedEvents, events[i])
		} else {
			m.logger.Warn("invalid community event", zap.Error(err))
		}
	}

	return validatedEvents
}

func (m *Manager) HandleCommunityEventsMessage(signer *ecdsa.PublicKey, message *protobuf.CommunityEventsMessage) (*CommunityResponse, error) {
	if signer == nil {
		return nil, errors.New("signer can't be nil")
	}

	eventsMessage, err := CommunityEventsMessageFromProtobuf(message)
	if err != nil {
		return nil, err
	}

	community, err := m.GetByID(eventsMessage.CommunityID)
	if err != nil {
		return nil, err
	}

	if !community.IsPrivilegedMember(signer) {
		return nil, errors.New("user has not permissions to send events")
	}

	originCommunity := community.CreateDeepCopy()

	eventsMessage.Events = m.validateAndFilterEvents(community, eventsMessage.Events)

	err = community.UpdateCommunityByEvents(eventsMessage)
	if err != nil {
		if err == ErrInvalidCommunityEventClock && community.IsControlNode() {
			// send updated CommunityDescription to the event sender on top of which he must apply his changes
			eventsMessage.EventsBaseCommunityDescription = community.config.CommunityDescriptionProtocolMessage
			m.publish(&Subscription{
				CommunityEventsMessageInvalidClock: &CommunityEventsMessageInvalidClockSignal{
					Community:              community,
					CommunityEventsMessage: eventsMessage,
				}})
		}
		return nil, err
	}

	additionalCommunityResponse, err := m.handleAdditionalAdminChanges(community)
	if err != nil {
		return nil, err
	}

	if err = m.handleCommunityTokensMetadata(community); err != nil {
		return nil, err
	}

	// Control node applies events and publish updated CommunityDescription
	if community.IsControlNode() {
		community.config.EventsData = nil // clear events, they are already applied
		community.increaseClock()

		if m.keyDistributor != nil {
			encryptionKeyActions := EvaluateCommunityEncryptionKeyActions(originCommunity, community)
			err := m.keyDistributor.Generate(community, encryptionKeyActions)
			if err != nil {
				return nil, err
			}
		}

		err = m.persistence.SaveCommunity(community)
		if err != nil {
			return nil, err
		}

		m.publish(&Subscription{Community: community})
	} else {
		err = m.persistence.SaveCommunity(community)
		if err != nil {
			return nil, err
		}
		err := m.persistence.SaveCommunityEvents(community)
		if err != nil {
			return nil, err
		}
	}

	return &CommunityResponse{
		Community:      community,
		Changes:        EvaluateCommunityChanges(originCommunity, community),
		RequestsToJoin: additionalCommunityResponse.RequestsToJoin,
	}, nil
}

// Creates new CommunityEventsMessage by re-applying our rejected events on top of latest known CommunityDescription.
// Returns nil if none of our events were rejected.
func (m *Manager) HandleCommunityEventsMessageRejected(signer *ecdsa.PublicKey, message *protobuf.CommunityEventsMessageRejected) (*CommunityEventsMessage, error) {
	if signer == nil {
		return nil, errors.New("signer can't be nil")
	}

	id := crypto.CompressPubkey(signer)
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	eventsMessage, err := CommunityEventsMessageFromProtobuf(message.Msg)
	if err != nil {
		return nil, err
	}

	communityDescription, err := validateAndGetEventsMessageCommunityDescription(eventsMessage.EventsBaseCommunityDescription, signer)
	if err != nil {
		return nil, err
	}
	// the privileged member did not receive updated CommunityDescription so his events
	// will be send on top of outdated CommunityDescription
	if communityDescription.Clock != community.Clock() {
		return nil, errors.New("resend rejected community events aborted, client node has outdated community description")
	}

	eventsMessage.Events = m.validateAndFilterEvents(community, eventsMessage.Events)

	myRejectedEvents := make([]CommunityEvent, 0)
	for _, rejectedEvent := range eventsMessage.Events {
		rejectedEventSigner, err := rejectedEvent.RecoverSigner()
		if err != nil {
			continue
		}

		if rejectedEventSigner.Equal(m.identity.Public()) {
			myRejectedEvents = append(myRejectedEvents, rejectedEvent)
		}
	}

	if len(myRejectedEvents) == 0 {
		return nil, nil
	}

	// Re-apply rejected events on top of latest known `CommunityDescription`
	community.config.EventsData = &EventsData{
		EventsBaseCommunityDescription: community.config.CommunityDescriptionProtocolMessage,
		Events:                         myRejectedEvents,
	}
	reapplyEventsMessage := community.ToCommunityEventsMessage()

	return reapplyEventsMessage, nil
}

func (m *Manager) handleAdditionalAdminChanges(community *Community) (*CommunityResponse, error) {
	communityResponse := CommunityResponse{
		RequestsToJoin: make([]*RequestToJoin, 0),
	}

	if !(community.IsControlNode() || community.HasPermissionToSendCommunityEvents()) {
		// we're a normal user/member node, so there's nothing for us to do here
		return &communityResponse, nil
	}

	for i := range community.config.EventsData.Events {
		communityEvent := &community.config.EventsData.Events[i]
		switch communityEvent.Type {
		case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT:
			requestsToJoin, err := m.handleCommunityEventRequestAccepted(community, communityEvent)
			if err != nil {
				return nil, err
			}
			if requestsToJoin != nil {
				communityResponse.RequestsToJoin = append(communityResponse.RequestsToJoin, requestsToJoin...)
			}

		case protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT:
			requestsToJoin, err := m.handleCommunityEventRequestRejected(community, communityEvent)
			if err != nil {
				return nil, err
			}
			if requestsToJoin != nil {
				communityResponse.RequestsToJoin = append(communityResponse.RequestsToJoin, requestsToJoin...)
			}

		default:
		}
	}
	return &communityResponse, nil
}

func (m *Manager) saveOrUpdateRequestToJoin(communityID types.HexBytes, requestToJoin *RequestToJoin) (bool, error) {
	updated := false

	existingRequestToJoin, err := m.persistence.GetRequestToJoin(requestToJoin.ID)
	if err != nil && err != sql.ErrNoRows {
		return updated, err
	}

	if existingRequestToJoin != nil {
		// node already knows about this request to join, so let's compare clocks
		// and update it if necessary
		if existingRequestToJoin.Clock <= requestToJoin.Clock {
			pk, err := common.HexToPubkey(existingRequestToJoin.PublicKey)
			if err != nil {
				return updated, err
			}
			err = m.persistence.SetRequestToJoinState(common.PubkeyToHex(pk), communityID, requestToJoin.State)
			if err != nil {
				return updated, err
			}
			updated = true
		}
	} else {
		err := m.persistence.SaveRequestToJoin(requestToJoin)
		if err != nil {
			return updated, err
		}
	}

	return updated, nil
}

func (m *Manager) handleCommunityEventRequestAccepted(community *Community, communityEvent *CommunityEvent) ([]*RequestToJoin, error) {
	acceptedRequestsToJoin := make([]types.HexBytes, 0)

	requestsToJoin := make([]*RequestToJoin, 0)

	for signer, request := range communityEvent.AcceptedRequestsToJoin {
		requestToJoin := &RequestToJoin{
			PublicKey:   signer,
			Clock:       request.Clock,
			ENSName:     request.EnsName,
			CommunityID: request.CommunityId,
			State:       RequestToJoinStateAcceptedPending,
		}
		requestToJoin.CalculateID()

		existingRequestToJoin, err := m.persistence.GetRequestToJoin(requestToJoin.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if existingRequestToJoin != nil {
			alreadyProcessedByControlNode := existingRequestToJoin.State == RequestToJoinStateAccepted || existingRequestToJoin.State == RequestToJoinStateDeclined
			if alreadyProcessedByControlNode || existingRequestToJoin.State == RequestToJoinStateCanceled {
				continue
			}
		}

		requestUpdated, err := m.saveOrUpdateRequestToJoin(community.ID(), requestToJoin)
		if err != nil {
			return nil, err
		}

		// If request to join exists in control node, add request to acceptedRequestsToJoin.
		// Otherwise keep the request as RequestToJoinStateAcceptedPending,
		// as privileged users don't have revealed addresses. This can happen if control node received
		// community event message before user request to join.
		if community.IsControlNode() && requestUpdated {
			acceptedRequestsToJoin = append(acceptedRequestsToJoin, requestToJoin.ID)
		}

		requestsToJoin = append(requestsToJoin, requestToJoin)
	}
	if community.IsControlNode() {
		m.publish(&Subscription{AcceptedRequestsToJoin: acceptedRequestsToJoin})
	}
	return requestsToJoin, nil
}

func (m *Manager) handleCommunityEventRequestRejected(community *Community, communityEvent *CommunityEvent) ([]*RequestToJoin, error) {
	rejectedRequestsToJoin := make([]types.HexBytes, 0)

	requestsToJoin := make([]*RequestToJoin, 0)

	for signer, request := range communityEvent.RejectedRequestsToJoin {
		requestToJoin := &RequestToJoin{
			PublicKey:   signer,
			Clock:       request.Clock,
			ENSName:     request.EnsName,
			CommunityID: request.CommunityId,
			State:       RequestToJoinStateDeclinedPending,
		}
		requestToJoin.CalculateID()

		existingRequestToJoin, err := m.persistence.GetRequestToJoin(requestToJoin.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if existingRequestToJoin != nil {
			alreadyProcessedByControlNode := existingRequestToJoin.State == RequestToJoinStateAccepted || existingRequestToJoin.State == RequestToJoinStateDeclined
			if alreadyProcessedByControlNode || existingRequestToJoin.State == RequestToJoinStateCanceled {
				continue
			}
		}

		requestUpdated, err := m.saveOrUpdateRequestToJoin(community.ID(), requestToJoin)
		if err != nil {
			return nil, err
		}
		// If request to join exists in control node, add request to rejectedRequestsToJoin.
		// Otherwise keep the request as RequestToJoinStateDeclinedPending,
		// as privileged users don't have revealed addresses. This can happen if control node received
		// community event message before user request to join.
		if community.IsControlNode() && requestUpdated {
			rejectedRequestsToJoin = append(rejectedRequestsToJoin, requestToJoin.ID)
		}

		requestsToJoin = append(requestsToJoin, requestToJoin)
	}

	if community.IsControlNode() {
		m.publish(&Subscription{RejectedRequestsToJoin: rejectedRequestsToJoin})
	}
	return requestsToJoin, nil
}

// markRequestToJoinAsAccepted marks all the pending requests to join as completed
// if we are members
func (m *Manager) markRequestToJoinAsAccepted(pk *ecdsa.PublicKey, community *Community) error {
	if community.HasMember(pk) {
		return m.persistence.SetRequestToJoinState(common.PubkeyToHex(pk), community.ID(), RequestToJoinStateAccepted)
	}
	return nil
}

func (m *Manager) markRequestToJoinAsCanceled(pk *ecdsa.PublicKey, community *Community) error {
	return m.persistence.SetRequestToJoinState(common.PubkeyToHex(pk), community.ID(), RequestToJoinStateCanceled)
}

func (m *Manager) markRequestToJoinAsAcceptedPending(pk *ecdsa.PublicKey, community *Community) error {
	return m.persistence.SetRequestToJoinState(common.PubkeyToHex(pk), community.ID(), RequestToJoinStateAcceptedPending)
}

func (m *Manager) DeletePendingRequestToJoin(request *RequestToJoin) error {
	community, err := m.GetByID(request.CommunityID)
	if err != nil {
		return err
	}

	err = m.persistence.DeletePendingRequestToJoin(request.ID)
	if err != nil {
		return err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClockInRequestToJoin method is used for testing
func (m *Manager) UpdateClockInRequestToJoin(id types.HexBytes, clock uint64) error {
	return m.persistence.UpdateClockInRequestToJoin(id, clock)
}

func (m *Manager) SetMuted(id types.HexBytes, muted bool) error {
	return m.persistence.SetMuted(id, muted)
}

func (m *Manager) MuteCommunityTill(communityID []byte, muteTill time.Time) error {
	return m.persistence.MuteCommunityTill(communityID, muteTill)
}
func (m *Manager) CancelRequestToJoin(request *requests.CancelRequestToJoinCommunity) (*RequestToJoin, *Community, error) {
	dbRequest, err := m.persistence.GetRequestToJoin(request.ID)
	if err != nil {
		return nil, nil, err
	}

	community, err := m.GetByID(dbRequest.CommunityID)
	if err != nil {
		return nil, nil, err
	}

	pk, err := common.HexToPubkey(dbRequest.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	dbRequest.State = RequestToJoinStateCanceled
	if err := m.markRequestToJoinAsCanceled(pk, community); err != nil {
		return nil, nil, err
	}

	return dbRequest, community, nil
}

func (m *Manager) CheckPermissionToJoin(id []byte, addresses []gethcommon.Address) (*CheckPermissionToJoinResponse, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	return m.PermissionChecker.CheckPermissionToJoin(community, addresses)

}

func (m *Manager) accountsSatisfyPermissionsToJoin(community *Community, accounts []*protobuf.RevealedAccount) (bool, protobuf.CommunityMember_Roles, error) {
	accountsAndChainIDs := revealedAccountsToAccountsAndChainIDsCombination(accounts)
	becomeAdminPermissions := community.TokenPermissionsByType(protobuf.CommunityTokenPermission_BECOME_ADMIN)
	becomeMemberPermissions := community.TokenPermissionsByType(protobuf.CommunityTokenPermission_BECOME_MEMBER)
	becomeTokenMasterPermissions := community.TokenPermissionsByType(protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER)

	if m.accountsHasPrivilegedPermission(becomeTokenMasterPermissions, accountsAndChainIDs) {
		return true, protobuf.CommunityMember_ROLE_TOKEN_MASTER, nil
	}
	if m.accountsHasPrivilegedPermission(becomeAdminPermissions, accountsAndChainIDs) {
		return true, protobuf.CommunityMember_ROLE_ADMIN, nil
	}

	if len(becomeMemberPermissions) > 0 {
		permissionResponse, err := m.PermissionChecker.CheckPermissions(becomeMemberPermissions, accountsAndChainIDs, true)
		if err != nil {
			return false, protobuf.CommunityMember_ROLE_NONE, err
		}

		return permissionResponse.Satisfied, protobuf.CommunityMember_ROLE_NONE, nil
	}

	return true, protobuf.CommunityMember_ROLE_NONE, nil
}

func (m *Manager) accountsSatisfyPermissionsToJoinChannels(community *Community, accounts []*protobuf.RevealedAccount) (map[string]*protobuf.CommunityChat, error) {
	result := make(map[string]*protobuf.CommunityChat)

	accountsAndChainIDs := revealedAccountsToAccountsAndChainIDsCombination(accounts)

	for channelID, channel := range community.config.CommunityDescription.Chats {
		channelViewOnlyPermissions := community.ChannelTokenPermissionsByType(community.IDString()+channelID, protobuf.CommunityTokenPermission_CAN_VIEW_CHANNEL)
		channelViewAndPostPermissions := community.ChannelTokenPermissionsByType(community.IDString()+channelID, protobuf.CommunityTokenPermission_CAN_VIEW_AND_POST_CHANNEL)
		channelPermissions := append(channelViewOnlyPermissions, channelViewAndPostPermissions...)

		if len(channelPermissions) > 0 {
			permissionResponse, err := m.PermissionChecker.CheckPermissions(channelPermissions, accountsAndChainIDs, true)
			if err != nil {
				return nil, err
			}
			if permissionResponse.Satisfied {
				result[channelID] = channel
			}
		} else {
			result[channelID] = channel
		}
	}

	return result, nil
}

func (m *Manager) AcceptRequestToJoin(dbRequest *RequestToJoin) (*Community, error) {
	pk, err := common.HexToPubkey(dbRequest.PublicKey)
	if err != nil {
		return nil, err
	}

	community, err := m.GetByID(dbRequest.CommunityID)
	if err != nil {
		return nil, err
	}

	if community.IsControlNode() {
		revealedAccounts, err := m.persistence.GetRequestToJoinRevealedAddresses(dbRequest.ID)
		if err != nil {
			return nil, err
		}

		permissionsSatisfied, role, err := m.accountsSatisfyPermissionsToJoin(community, revealedAccounts)
		if err != nil {
			return nil, err
		}

		if !permissionsSatisfied {
			return community, ErrNoPermissionToJoin
		}

		memberRoles := []protobuf.CommunityMember_Roles{}
		if role != protobuf.CommunityMember_ROLE_NONE {
			memberRoles = []protobuf.CommunityMember_Roles{role}
		}

		_, err = community.AddMember(pk, memberRoles)
		if err != nil {
			return nil, err
		}

		channels, err := m.accountsSatisfyPermissionsToJoinChannels(community, revealedAccounts)
		if err != nil {
			return nil, err
		}

		for channelID := range channels {
			_, err = community.AddMemberToChat(channelID, pk, memberRoles)
			if err != nil {
				return nil, err
			}
		}

		dbRequest.State = RequestToJoinStateAccepted
		if err := m.markRequestToJoinAsAccepted(pk, community); err != nil {
			return nil, err
		}

		dbRequest.RevealedAccounts = revealedAccounts
		if err = m.shareAcceptedRequestToJoinWithPrivilegedMembers(community, dbRequest); err != nil {
			return nil, err
		}

		// if accepted member has a privilege role, share with him requests to join
		memberRole := community.MemberRole(pk)
		if memberRole == protobuf.CommunityMember_ROLE_OWNER || memberRole == protobuf.CommunityMember_ROLE_ADMIN ||
			memberRole == protobuf.CommunityMember_ROLE_TOKEN_MASTER {

			newPrivilegedMember := make(map[protobuf.CommunityMember_Roles][]*ecdsa.PublicKey)
			newPrivilegedMember[memberRole] = []*ecdsa.PublicKey{pk}
			if err = m.shareRequestsToJoinWithNewPrivilegedMembers(community, newPrivilegedMember); err != nil {
				return nil, err
			}
		}
	} else if community.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_ACCEPT) {
		// admins do not perform permission checks, they merely mark the
		// request as accepted (pending) and forward their decision to the control node
		acceptedRequestsToJoin := make(map[string]*protobuf.CommunityRequestToJoin)
		acceptedRequestsToJoin[dbRequest.PublicKey] = dbRequest.ToCommunityRequestToJoinProtobuf()

		adminChanges := &CommunityEventChanges{
			AcceptedRequestsToJoin: acceptedRequestsToJoin,
		}

		err := community.addNewCommunityEvent(community.ToCommunityRequestToJoinAcceptCommunityEvent(adminChanges))
		if err != nil {
			return nil, err
		}

		dbRequest.State = RequestToJoinStateAcceptedPending
		if err := m.markRequestToJoinAsAcceptedPending(pk, community); err != nil {
			return nil, err
		}
	} else {
		return nil, ErrNotAuthorized
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) GetRequestToJoin(ID types.HexBytes) (*RequestToJoin, error) {
	return m.persistence.GetRequestToJoin(ID)
}

func (m *Manager) DeclineRequestToJoin(dbRequest *RequestToJoin) (*Community, error) {
	community, err := m.GetByID(dbRequest.CommunityID)
	if err != nil {
		return nil, err
	}

	adminEventCreated, err := community.DeclineRequestToJoin(dbRequest)
	if err != nil {
		return nil, err
	}

	requestToJoinState := RequestToJoinStateDeclined
	if adminEventCreated {
		requestToJoinState = RequestToJoinStateDeclinedPending // can only be declined by control node
	}

	dbRequest.State = requestToJoinState
	err = m.persistence.SetRequestToJoinState(dbRequest.PublicKey, dbRequest.CommunityID, requestToJoinState)
	if err != nil {
		return nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) shouldUserRetainDeclined(signer *ecdsa.PublicKey, community *Community, requestClock uint64) (bool, error) {
	requestID := CalculateRequestID(common.PubkeyToHex(signer), types.HexBytes(community.IDString()))
	request, err := m.persistence.GetRequestToJoin(requestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return request.ShouldRetainDeclined(requestClock)
}

func (m *Manager) HandleCommunityCancelRequestToJoin(signer *ecdsa.PublicKey, request *protobuf.CommunityCancelRequestToJoin) (*RequestToJoin, error) {
	community, err := m.GetByID(request.CommunityId)
	if err != nil {
		return nil, err
	}

	previousRequestToJoin, err := m.GetRequestToJoinByPkAndCommunityID(signer, community.ID())
	if err != nil {
		return nil, err
	}

	if request.Clock <= previousRequestToJoin.Clock {
		return nil, ErrInvalidClock
	}

	retainDeclined, err := m.shouldUserRetainDeclined(signer, community, request.Clock)
	if err != nil {
		return nil, err
	}
	if retainDeclined {
		return nil, ErrCommunityRequestAlreadyRejected
	}

	err = m.markRequestToJoinAsCanceled(signer, community)
	if err != nil {
		return nil, err
	}

	requestToJoin, err := m.persistence.GetRequestToJoinByPk(common.PubkeyToHex(signer), community.ID(), RequestToJoinStateCanceled)
	if err != nil {
		return nil, err
	}

	if community.HasMember(signer) {
		_, err = community.RemoveUserFromOrg(signer)
		if err != nil {
			return nil, err
		}

		err = m.saveAndPublish(community)
		if err != nil {
			return nil, err
		}
	}

	return requestToJoin, nil
}

func (m *Manager) HandleCommunityRequestToJoin(signer *ecdsa.PublicKey, receiver *ecdsa.PublicKey, request *protobuf.CommunityRequestToJoin) (*Community, *RequestToJoin, error) {
	community, err := m.GetByID(request.CommunityId)
	if err != nil {
		return nil, nil, err
	}

	err = community.ValidateRequestToJoin(signer, request)
	if err != nil {
		return nil, nil, err
	}

	requestToJoin := &RequestToJoin{
		PublicKey:        common.PubkeyToHex(signer),
		Clock:            request.Clock,
		ENSName:          request.EnsName,
		CommunityID:      request.CommunityId,
		State:            RequestToJoinStatePending,
		RevealedAccounts: request.RevealedAccounts,
	}
	requestToJoin.CalculateID()

	existingRequestToJoin, err := m.persistence.GetRequestToJoin(requestToJoin.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	if existingRequestToJoin == nil {
		err = m.SaveRequestToJoin(requestToJoin)
		if err != nil {
			return nil, nil, err
		}
	} else {
		retainDeclined, err := existingRequestToJoin.ShouldRetainDeclined(request.Clock)
		if err != nil {
			return nil, nil, err
		}
		if retainDeclined {
			return nil, nil, ErrCommunityRequestAlreadyRejected
		}

		switch existingRequestToJoin.State {
		case RequestToJoinStatePending, RequestToJoinStateDeclined, RequestToJoinStateCanceled:
			// Another request have been received, save request back to pending state
			err = m.SaveRequestToJoin(requestToJoin)
			if err != nil {
				return nil, nil, err
			}
		case RequestToJoinStateAccepted:
			// if member leaved the community and tries to request to join again
			if !community.HasMember(signer) {
				err = m.SaveRequestToJoin(requestToJoin)
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}

	if community.IsControlNode() {
		// verify if revealed addresses indeed belong to requester
		for _, revealedAccount := range request.RevealedAccounts {
			recoverParams := account.RecoverParams{
				Message:   types.EncodeHex(crypto.Keccak256(crypto.CompressPubkey(signer), community.ID(), requestToJoin.ID)),
				Signature: types.EncodeHex(revealedAccount.Signature),
			}

			matching, err := m.accountsManager.CanRecover(recoverParams, types.HexToAddress(revealedAccount.Address))
			if err != nil {
				return nil, nil, err
			}
			if !matching {
				// if ownership of only one wallet address cannot be verified,
				// we mark the request as cancelled and stop
				requestToJoin.State = RequestToJoinStateDeclined
				return community, requestToJoin, nil
			}
		}

		// Save revealed addresses + signatures so they can later be added
		// to the control node's local table of known revealed addresses
		err = m.persistence.SaveRequestToJoinRevealedAddresses(requestToJoin.ID, requestToJoin.RevealedAccounts)
		if err != nil {
			return nil, nil, err
		}

		if existingRequestToJoin != nil {
			// request to join was already processed by privileged user
			// and waits to get confirmation for its decision
			if existingRequestToJoin.State == RequestToJoinStateDeclinedPending {
				requestToJoin.State = RequestToJoinStateDeclined
				return community, requestToJoin, nil
			} else if existingRequestToJoin.State == RequestToJoinStateAcceptedPending {
				requestToJoin.State = RequestToJoinStateAccepted
				return community, requestToJoin, nil

			} else if existingRequestToJoin.State == RequestToJoinStateAwaitingAddresses {
				// community ownership changed, accept request automatically
				requestToJoin.State = RequestToJoinStateAccepted
				return community, requestToJoin, nil
			}
		}

		// If user is already a member, then accept request automatically
		// It may happen when member removes itself from community and then tries to rejoin
		// More specifically, CommunityRequestToLeave may be delivered later than CommunityRequestToJoin, or not delivered at all
		acceptAutomatically := community.AutoAccept() || community.HasMember(signer)
		if acceptAutomatically {
			// Don't check permissions here,
			// it will be done further in the processing pipeline.
			requestToJoin.State = RequestToJoinStateAccepted
			return community, requestToJoin, nil
		}
	}

	return community, requestToJoin, nil
}

func (m *Manager) HandleCommunityEditSharedAddresses(signer *ecdsa.PublicKey, request *protobuf.CommunityEditSharedAddresses) error {
	community, err := m.GetByID(request.CommunityId)
	if err != nil {
		return err
	}

	if err := community.ValidateEditSharedAddresses(signer, request); err != nil {
		return err
	}

	// verify if revealed addresses indeed belong to requester
	for _, revealedAccount := range request.RevealedAccounts {
		recoverParams := account.RecoverParams{
			Message:   types.EncodeHex(crypto.Keccak256(crypto.CompressPubkey(signer), community.ID())),
			Signature: types.EncodeHex(revealedAccount.Signature),
		}

		matching, err := m.accountsManager.CanRecover(recoverParams, types.HexToAddress(revealedAccount.Address))
		if err != nil {
			return err
		}
		if !matching {
			// if ownership of only one wallet address cannot be verified we stop
			return errors.New("wrong wallet address used")
		}
	}

	requestToJoin := &RequestToJoin{
		PublicKey:        common.PubkeyToHex(signer),
		CommunityID:      community.ID(),
		RevealedAccounts: request.RevealedAccounts,
	}
	requestToJoin.CalculateID()

	err = m.persistence.RemoveRequestToJoinRevealedAddresses(requestToJoin.ID)
	if err != nil {
		return err
	}
	err = m.persistence.SaveRequestToJoinRevealedAddresses(requestToJoin.ID, requestToJoin.RevealedAccounts)
	if err != nil {
		return err
	}

	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return err
	}

	if community.IsControlNode() {
		m.publish(&Subscription{Community: community})
	}

	return nil
}

func calculateChainIDsSet(accountsAndChainIDs []*AccountChainIDsCombination, requirementsChainIDs map[uint64]bool) []uint64 {

	revealedAccountsChainIDs := make([]uint64, 0)
	revealedAccountsChainIDsMap := make(map[uint64]bool)

	// we want all chainIDs provided by revealed addresses that also exist
	// in the token requirements
	for _, accountAndChainIDs := range accountsAndChainIDs {
		for _, chainID := range accountAndChainIDs.ChainIDs {
			if requirementsChainIDs[chainID] && !revealedAccountsChainIDsMap[chainID] {
				revealedAccountsChainIDsMap[chainID] = true
				revealedAccountsChainIDs = append(revealedAccountsChainIDs, chainID)
			}
		}
	}
	return revealedAccountsChainIDs
}

type CollectiblesByChain = map[uint64]map[gethcommon.Address]thirdparty.TokenBalancesPerContractAddress

func (m *Manager) GetOwnedERC721Tokens(walletAddresses []gethcommon.Address, tokenRequirements map[uint64]map[string]*protobuf.TokenCriteria, chainIDs []uint64) (CollectiblesByChain, error) {
	if m.collectiblesManager == nil {
		return nil, errors.New("no collectibles manager")
	}

	ctx := context.Background()

	ownedERC721Tokens := make(CollectiblesByChain)

	for chainID, erc721Tokens := range tokenRequirements {

		skipChain := true
		for _, cID := range chainIDs {
			if chainID == cID {
				skipChain = false
			}
		}

		if skipChain {
			continue
		}

		contractAddresses := make([]gethcommon.Address, 0)
		for contractAddress := range erc721Tokens {
			contractAddresses = append(contractAddresses, gethcommon.HexToAddress(contractAddress))
		}

		if _, exists := ownedERC721Tokens[chainID]; !exists {
			ownedERC721Tokens[chainID] = make(map[gethcommon.Address]thirdparty.TokenBalancesPerContractAddress)
		}

		for _, owner := range walletAddresses {
			balances, err := m.collectiblesManager.FetchBalancesByOwnerAndContractAddress(ctx, walletcommon.ChainID(chainID), owner, contractAddresses)
			if err != nil {
				m.logger.Info("couldn't fetch owner assets", zap.Error(err))
				return nil, err
			}
			ownedERC721Tokens[chainID][owner] = balances
		}
	}
	return ownedERC721Tokens, nil
}

func (m *Manager) CheckChannelPermissions(communityID types.HexBytes, chatID string, addresses []gethcommon.Address) (*CheckChannelPermissionsResponse, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, err
	}

	if chatID == "" {
		return nil, errors.New(fmt.Sprintf("couldn't check channel permissions, invalid chat id: %s", chatID))
	}

	viewOnlyPermissions := community.ChannelTokenPermissionsByType(chatID, protobuf.CommunityTokenPermission_CAN_VIEW_CHANNEL)
	viewAndPostPermissions := community.ChannelTokenPermissionsByType(chatID, protobuf.CommunityTokenPermission_CAN_VIEW_AND_POST_CHANNEL)

	allChainIDs, err := m.tokenManager.GetAllChainIDs()
	if err != nil {
		return nil, err
	}
	accountsAndChainIDs := combineAddressesAndChainIDs(addresses, allChainIDs)

	response, err := m.checkChannelPermissions(viewOnlyPermissions, viewAndPostPermissions, accountsAndChainIDs, false)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveCheckChannelPermissionResponse(communityID.String(), chatID, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type CheckChannelPermissionsResponse struct {
	ViewOnlyPermissions    *CheckChannelViewOnlyPermissionsResult    `json:"viewOnlyPermissions"`
	ViewAndPostPermissions *CheckChannelViewAndPostPermissionsResult `json:"viewAndPostPermissions"`
}

type CheckChannelViewOnlyPermissionsResult struct {
	Satisfied   bool                                      `json:"satisfied"`
	Permissions map[string]*PermissionTokenCriteriaResult `json:"permissions"`
}

type CheckChannelViewAndPostPermissionsResult struct {
	Satisfied   bool                                      `json:"satisfied"`
	Permissions map[string]*PermissionTokenCriteriaResult `json:"permissions"`
}

func (m *Manager) checkChannelPermissions(viewOnlyPermissions []*CommunityTokenPermission, viewAndPostPermissions []*CommunityTokenPermission, accountsAndChainIDs []*AccountChainIDsCombination, shortcircuit bool) (*CheckChannelPermissionsResponse, error) {

	response := &CheckChannelPermissionsResponse{
		ViewOnlyPermissions: &CheckChannelViewOnlyPermissionsResult{
			Satisfied:   false,
			Permissions: make(map[string]*PermissionTokenCriteriaResult),
		},
		ViewAndPostPermissions: &CheckChannelViewAndPostPermissionsResult{
			Satisfied:   false,
			Permissions: make(map[string]*PermissionTokenCriteriaResult),
		},
	}

	viewOnlyPermissionsResponse, err := m.PermissionChecker.CheckPermissions(viewOnlyPermissions, accountsAndChainIDs, shortcircuit)
	if err != nil {
		return nil, err
	}

	viewAndPostPermissionsResponse, err := m.PermissionChecker.CheckPermissions(viewAndPostPermissions, accountsAndChainIDs, shortcircuit)
	if err != nil {
		return nil, err
	}

	hasViewOnlyPermissions := len(viewOnlyPermissions) > 0
	hasViewAndPostPermissions := len(viewAndPostPermissions) > 0

	if (hasViewAndPostPermissions && !hasViewOnlyPermissions) || (hasViewOnlyPermissions && hasViewAndPostPermissions && viewAndPostPermissionsResponse.Satisfied) {
		response.ViewOnlyPermissions.Satisfied = viewAndPostPermissionsResponse.Satisfied
	} else {
		response.ViewOnlyPermissions.Satisfied = viewOnlyPermissionsResponse.Satisfied
	}
	response.ViewOnlyPermissions.Permissions = viewOnlyPermissionsResponse.Permissions

	if (hasViewOnlyPermissions && !viewOnlyPermissionsResponse.Satisfied) ||
		(hasViewOnlyPermissions && !hasViewAndPostPermissions) {
		response.ViewAndPostPermissions.Satisfied = false
	} else {
		response.ViewAndPostPermissions.Satisfied = viewAndPostPermissionsResponse.Satisfied
	}
	response.ViewAndPostPermissions.Permissions = viewAndPostPermissionsResponse.Permissions

	return response, nil
}

func (m *Manager) CheckAllChannelsPermissions(communityID types.HexBytes, addresses []gethcommon.Address) (*CheckAllChannelsPermissionsResponse, error) {

	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, err
	}
	channels := community.Chats()

	allChainIDs, err := m.tokenManager.GetAllChainIDs()
	if err != nil {
		return nil, err
	}
	accountsAndChainIDs := combineAddressesAndChainIDs(addresses, allChainIDs)

	response := &CheckAllChannelsPermissionsResponse{
		Channels: make(map[string]*CheckChannelPermissionsResponse),
	}

	for channelID := range channels {
		viewOnlyPermissions := community.ChannelTokenPermissionsByType(community.IDString()+channelID, protobuf.CommunityTokenPermission_CAN_VIEW_CHANNEL)
		viewAndPostPermissions := community.ChannelTokenPermissionsByType(community.IDString()+channelID, protobuf.CommunityTokenPermission_CAN_VIEW_AND_POST_CHANNEL)

		checkChannelPermissionsResponse, err := m.checkChannelPermissions(viewOnlyPermissions, viewAndPostPermissions, accountsAndChainIDs, false)
		if err != nil {
			return nil, err
		}
		err = m.persistence.SaveCheckChannelPermissionResponse(community.IDString(), community.IDString()+channelID, checkChannelPermissionsResponse)
		if err != nil {
			return nil, err
		}
		response.Channels[community.IDString()+channelID] = checkChannelPermissionsResponse
	}
	return response, nil
}

func (m *Manager) GetCheckChannelPermissionResponses(communityID types.HexBytes) (*CheckAllChannelsPermissionsResponse, error) {

	response, err := m.persistence.GetCheckChannelPermissionResponses(communityID.String())
	if err != nil {
		return nil, err
	}
	return &CheckAllChannelsPermissionsResponse{Channels: response}, nil
}

type CheckAllChannelsPermissionsResponse struct {
	Channels map[string]*CheckChannelPermissionsResponse `json:"channels"`
}

func (m *Manager) HandleCommunityRequestToJoinResponse(signer *ecdsa.PublicKey, request *protobuf.CommunityRequestToJoinResponse) (*RequestToJoin, error) {
	pkString := common.PubkeyToHex(&m.identity.PublicKey)

	community, err := m.GetByID(request.CommunityId)
	if err != nil {
		return nil, err
	}

	communityDescriptionBytes, err := proto.Marshal(request.Community)
	if err != nil {
		return nil, err
	}

	// We need to wrap `request.Community` in an `ApplicationMetadataMessage`
	// of type `CommunityDescription` because `UpdateCommunityDescription` expects this.
	//
	// This is merely for marsheling/unmarsheling, hence we attaching a `Signature`
	// is not needed.
	metadataMessage := &protobuf.ApplicationMetadataMessage{
		Payload: communityDescriptionBytes,
		Type:    protobuf.ApplicationMetadataMessage_COMMUNITY_DESCRIPTION,
	}

	appMetadataMsg, err := proto.Marshal(metadataMessage)
	if err != nil {
		return nil, err
	}

	isControlNodeSigner := common.IsPubKeyEqual(community.ControlNode(), signer)
	if !isControlNodeSigner {
		return nil, ErrNotAuthorized
	}

	_, err = m.preprocessDescription(community.ID(), request.Community)
	if err != nil {
		return nil, err
	}

	_, err = community.UpdateCommunityDescription(request.Community, appMetadataMsg, nil)
	if err != nil {
		return nil, err
	}

	if err = m.handleCommunityTokensMetadata(community); err != nil {
		return nil, err
	}

	err = m.persistence.SaveCommunity(community)

	if err != nil {
		return nil, err
	}

	if request.Accepted {
		err = m.markRequestToJoinAsAccepted(&m.identity.PublicKey, community)
		if err != nil {
			return nil, err
		}
	} else {

		err = m.persistence.SetRequestToJoinState(pkString, community.ID(), RequestToJoinStateDeclined)
		if err != nil {
			return nil, err
		}
	}

	return m.persistence.GetRequestToJoinByPkAndCommunityID(pkString, community.ID())
}

func (m *Manager) HandleCommunityRequestToLeave(signer *ecdsa.PublicKey, proto *protobuf.CommunityRequestToLeave) error {
	requestToLeave := NewRequestToLeave(common.PubkeyToHex(signer), proto)
	if err := m.persistence.SaveRequestToLeave(requestToLeave); err != nil {
		return err
	}

	// Ensure corresponding requestToJoin clock is older than requestToLeave
	requestToJoin, err := m.persistence.GetRequestToJoin(requestToLeave.ID)
	if err != nil {
		return err
	}
	if requestToJoin.Clock > requestToLeave.Clock {
		return ErrOldRequestToLeave
	}

	return nil
}

func UnwrapCommunityDescriptionMessage(payload []byte) (*ecdsa.PublicKey, *protobuf.CommunityDescription, error) {

	applicationMetadataMessage := &protobuf.ApplicationMetadataMessage{}
	err := proto.Unmarshal(payload, applicationMetadataMessage)
	if err != nil {
		return nil, nil, err
	}
	if applicationMetadataMessage.Type != protobuf.ApplicationMetadataMessage_COMMUNITY_DESCRIPTION {
		return nil, nil, ErrInvalidMessage
	}
	signer, err := utils.RecoverKey(applicationMetadataMessage)
	if err != nil {
		return nil, nil, err
	}

	description := &protobuf.CommunityDescription{}

	err = proto.Unmarshal(applicationMetadataMessage.Payload, description)
	if err != nil {
		return nil, nil, err
	}

	return signer, description, nil
}

func (m *Manager) JoinCommunity(id types.HexBytes, forceJoin bool) (*Community, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !forceJoin && community.Joined() {
		// Nothing to do, we are already joined
		return community, ErrOrgAlreadyJoined
	}
	community.Join()
	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, err
	}
	return community, nil
}

func (m *Manager) SpectateCommunity(id types.HexBytes) (*Community, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}
	community.Spectate()
	if err = m.persistence.SaveCommunity(community); err != nil {
		return nil, err
	}
	return community, nil
}

func (m *Manager) GetMagnetlinkMessageClock(communityID types.HexBytes) (uint64, error) {
	return m.persistence.GetMagnetlinkMessageClock(communityID)
}

func (m *Manager) GetRequestToJoinIDByPkAndCommunityID(pk *ecdsa.PublicKey, communityID []byte) ([]byte, error) {
	return m.persistence.GetRequestToJoinIDByPkAndCommunityID(common.PubkeyToHex(pk), communityID)
}

func (m *Manager) GetCommunityRequestToJoinClock(pk *ecdsa.PublicKey, communityID string) (uint64, error) {
	request, err := m.persistence.GetRequestToJoinByPkAndCommunityID(common.PubkeyToHex(pk), []byte(communityID))
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	if request == nil || request.State != RequestToJoinStateAccepted {
		return 0, nil
	}
	return request.Clock, nil
}

func (m *Manager) GetRequestToJoinByPkAndCommunityID(pk *ecdsa.PublicKey, communityID []byte) (*RequestToJoin, error) {
	return m.persistence.GetRequestToJoinByPkAndCommunityID(common.PubkeyToHex(pk), communityID)
}

func (m *Manager) UpdateCommunityDescriptionMagnetlinkMessageClock(communityID types.HexBytes, clock uint64) error {
	community, err := m.GetByIDString(communityID.String())
	if err != nil {
		return err
	}
	community.config.CommunityDescription.ArchiveMagnetlinkClock = clock
	return m.persistence.SaveCommunity(community)
}

func (m *Manager) UpdateMagnetlinkMessageClock(communityID types.HexBytes, clock uint64) error {
	return m.persistence.UpdateMagnetlinkMessageClock(communityID, clock)
}

func (m *Manager) UpdateLastSeenMagnetlink(communityID types.HexBytes, magnetlinkURI string) error {
	return m.persistence.UpdateLastSeenMagnetlink(communityID, magnetlinkURI)
}

func (m *Manager) GetLastSeenMagnetlink(communityID types.HexBytes) (string, error) {
	return m.persistence.GetLastSeenMagnetlink(communityID)
}

func (m *Manager) LeaveCommunity(id types.HexBytes) (*Community, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	community.RemoveOurselvesFromOrg(&m.identity.PublicKey)
	community.Leave()

	if err = m.persistence.SaveCommunity(community); err != nil {
		return nil, err
	}

	return community, nil
}

// Same as LeaveCommunity, but we want to stay spectating
func (m *Manager) KickedOutOfCommunity(id types.HexBytes) (*Community, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	community.RemoveOurselvesFromOrg(&m.identity.PublicKey)
	community.Leave()
	community.Spectate()

	if err = m.persistence.SaveCommunity(community); err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) AddMemberOwnerToCommunity(communityID types.HexBytes, pk *ecdsa.PublicKey) (*Community, error) {
	community, err := m.GetByID(communityID)
	if err != nil {
		return nil, err
	}

	_, err = community.AddMember(pk, []protobuf.CommunityMember_Roles{protobuf.CommunityMember_ROLE_OWNER})
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, err
	}

	m.publish(&Subscription{Community: community})
	return community, nil
}

func (m *Manager) RemoveUserFromCommunity(id types.HexBytes, pk *ecdsa.PublicKey) (*Community, error) {
	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	_, err = community.RemoveUserFromOrg(pk)
	if err != nil {
		return nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) UnbanUserFromCommunity(request *requests.UnbanUserFromCommunity) (*Community, error) {
	id := request.CommunityID
	publicKey, err := common.HexToPubkey(request.User.String())
	if err != nil {
		return nil, err
	}

	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	_, err = community.UnbanUserFromCommunity(publicKey)
	if err != nil {
		return nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) AddRoleToMember(request *requests.AddRoleToMember) (*Community, error) {
	id := request.CommunityID
	publicKey, err := common.HexToPubkey(request.User.String())
	if err != nil {
		return nil, err
	}

	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	if !community.hasMember(publicKey) {
		return nil, ErrMemberNotFound
	}

	_, err = community.AddRoleToMember(publicKey, request.Role)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, err
	}

	m.publish(&Subscription{Community: community})

	return community, nil
}

func (m *Manager) RemoveRoleFromMember(request *requests.RemoveRoleFromMember) (*Community, error) {
	id := request.CommunityID
	publicKey, err := common.HexToPubkey(request.User.String())
	if err != nil {
		return nil, err
	}

	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	if !community.hasMember(publicKey) {
		return nil, ErrMemberNotFound
	}

	_, err = community.RemoveRoleFromMember(publicKey, request.Role)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveCommunity(community)
	if err != nil {
		return nil, err
	}

	m.publish(&Subscription{Community: community})

	return community, nil
}

func (m *Manager) BanUserFromCommunity(request *requests.BanUserFromCommunity) (*Community, error) {
	id := request.CommunityID

	publicKey, err := common.HexToPubkey(request.User.String())
	if err != nil {
		return nil, err
	}

	community, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	_, err = community.BanUserFromCommunity(publicKey)
	if err != nil {
		return nil, err
	}

	err = m.saveAndPublish(community)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (m *Manager) dbRecordBundleToCommunity(r *CommunityRecordBundle) (*Community, error) {
	var descriptionEncryptor DescriptionEncryptor
	if m.encryptor != nil {
		descriptionEncryptor = m
	}

	return recordBundleToCommunity(r, &m.identity.PublicKey, m.installationID, m.logger, m.timesource, descriptionEncryptor, func(community *Community) error {
		_, err := m.preprocessDescription(community.ID(), community.config.CommunityDescription)
		if err != nil {
			return err
		}

		err = community.updateCommunityDescriptionByEvents()
		if err != nil {
			return err
		}

		if m.transport != nil && m.transport.WakuVersion() == 2 {
			topic := community.PubsubTopic()
			privKey, err := m.transport.RetrievePubsubTopicKey(topic)
			if err != nil {
				return err
			}
			community.config.PubsubTopicPrivateKey = privKey
		}

		return nil
	})
}

func (m *Manager) GetByID(id []byte) (*Community, error) {
	community, err := m.persistence.GetByID(&m.identity.PublicKey, id)
	if err != nil {
		return nil, err
	}
	if community == nil {
		return nil, ErrOrgNotFound
	}
	return community, nil
}

func (m *Manager) GetByIDString(idString string) (*Community, error) {
	id, err := types.DecodeHex(idString)
	if err != nil {
		return nil, err
	}
	return m.GetByID(id)
}

func (m *Manager) GetCommunityShard(communityID types.HexBytes) (*shard.Shard, error) {
	return m.persistence.GetCommunityShard(communityID)
}

func (m *Manager) SaveCommunityShard(communityID types.HexBytes, shard *shard.Shard, clock uint64) error {
	return m.persistence.SaveCommunityShard(communityID, shard, clock)
}

func (m *Manager) DeleteCommunityShard(communityID types.HexBytes) error {
	return m.persistence.DeleteCommunityShard(communityID)
}

func (m *Manager) SaveRequestToJoinRevealedAddresses(requestID types.HexBytes, revealedAccounts []*protobuf.RevealedAccount) error {
	return m.persistence.SaveRequestToJoinRevealedAddresses(requestID, revealedAccounts)
}

func (m *Manager) RemoveRequestToJoinRevealedAddresses(requestID types.HexBytes) error {
	return m.persistence.RemoveRequestToJoinRevealedAddresses(requestID)
}

func (m *Manager) SaveRequestToJoinAndCommunity(requestToJoin *RequestToJoin, community *Community) (*Community, *RequestToJoin, error) {
	if err := m.persistence.SaveRequestToJoin(requestToJoin); err != nil {
		return nil, nil, err
	}
	community.config.RequestedToJoinAt = uint64(time.Now().Unix())
	community.AddRequestToJoin(requestToJoin)

	// Save revealed addresses to our own table so that we can retrieve them later when editing
	if err := m.SaveRequestToJoinRevealedAddresses(requestToJoin.ID, requestToJoin.RevealedAccounts); err != nil {
		return nil, nil, err
	}

	return community, requestToJoin, nil
}

func (m *Manager) CreateRequestToJoin(request *requests.RequestToJoinCommunity) *RequestToJoin {
	clock := uint64(time.Now().Unix())
	requestToJoin := &RequestToJoin{
		PublicKey:        common.PubkeyToHex(&m.identity.PublicKey),
		Clock:            clock,
		ENSName:          request.ENSName,
		CommunityID:      request.CommunityID,
		State:            RequestToJoinStatePending,
		Our:              true,
		RevealedAccounts: make([]*protobuf.RevealedAccount, 0),
	}

	requestToJoin.CalculateID()

	addSignature := len(request.Signatures) == len(request.AddressesToReveal)
	for i := range request.AddressesToReveal {
		revealedAcc := &protobuf.RevealedAccount{
			Address:          request.AddressesToReveal[i],
			IsAirdropAddress: types.HexToAddress(request.AddressesToReveal[i]) == types.HexToAddress(request.AirdropAddress),
		}

		if addSignature {
			revealedAcc.Signature = request.Signatures[i]
		}

		requestToJoin.RevealedAccounts = append(requestToJoin.RevealedAccounts, revealedAcc)
	}

	return requestToJoin
}

func (m *Manager) SaveRequestToJoin(request *RequestToJoin) error {
	return m.persistence.SaveRequestToJoin(request)
}

func (m *Manager) CanceledRequestsToJoinForUser(pk *ecdsa.PublicKey) ([]*RequestToJoin, error) {
	return m.persistence.CanceledRequestsToJoinForUser(common.PubkeyToHex(pk))
}

func (m *Manager) CanceledRequestToJoinForUserForCommunityID(pk *ecdsa.PublicKey, communityID []byte) (*RequestToJoin, error) {
	return m.persistence.CanceledRequestToJoinForUserForCommunityID(common.PubkeyToHex(pk), communityID)
}

func (m *Manager) PendingRequestsToJoin() ([]*RequestToJoin, error) {
	return m.persistence.PendingRequestsToJoin()
}

func (m *Manager) PendingRequestsToJoinForUser(pk *ecdsa.PublicKey) ([]*RequestToJoin, error) {
	return m.persistence.RequestsToJoinForUserByState(common.PubkeyToHex(pk), RequestToJoinStatePending)
}

func (m *Manager) PendingRequestsToJoinForCommunity(id types.HexBytes) ([]*RequestToJoin, error) {
	m.logger.Info("fetching pending invitations", zap.String("community-id", id.String()))
	return m.persistence.PendingRequestsToJoinForCommunity(id)
}

func (m *Manager) DeclinedRequestsToJoinForCommunity(id types.HexBytes) ([]*RequestToJoin, error) {
	m.logger.Info("fetching declined invitations", zap.String("community-id", id.String()))
	return m.persistence.DeclinedRequestsToJoinForCommunity(id)
}

func (m *Manager) CanceledRequestsToJoinForCommunity(id types.HexBytes) ([]*RequestToJoin, error) {
	m.logger.Info("fetching canceled invitations", zap.String("community-id", id.String()))
	return m.persistence.CanceledRequestsToJoinForCommunity(id)
}

func (m *Manager) AcceptedRequestsToJoinForCommunity(id types.HexBytes) ([]*RequestToJoin, error) {
	m.logger.Info("fetching canceled invitations", zap.String("community-id", id.String()))
	return m.persistence.AcceptedRequestsToJoinForCommunity(id)
}

func (m *Manager) AcceptedPendingRequestsToJoinForCommunity(id types.HexBytes) ([]*RequestToJoin, error) {
	return m.persistence.AcceptedPendingRequestsToJoinForCommunity(id)
}

func (m *Manager) DeclinedPendingRequestsToJoinForCommunity(id types.HexBytes) ([]*RequestToJoin, error) {
	return m.persistence.DeclinedPendingRequestsToJoinForCommunity(id)
}

func (m *Manager) AllNonApprovedCommunitiesRequestsToJoin() ([]*RequestToJoin, error) {
	m.logger.Info("fetching all non-approved invitations for all communities")
	return m.persistence.AllNonApprovedCommunitiesRequestsToJoin()
}

func (m *Manager) RequestsToJoinForCommunityAwaitingAddresses(id types.HexBytes) ([]*RequestToJoin, error) {
	m.logger.Info("fetching ownership changed invitations", zap.String("community-id", id.String()))
	return m.persistence.RequestsToJoinForCommunityAwaitingAddresses(id)
}

func (m *Manager) CanPost(pk *ecdsa.PublicKey, communityID string, chatID string) (bool, error) {
	community, err := m.GetByIDString(communityID)
	if err != nil {
		return false, err
	}
	return community.CanPost(pk, chatID)
}

func (m *Manager) IsEncrypted(communityID string) (bool, error) {
	community, err := m.GetByIDString(communityID)
	if err != nil {
		return false, err
	}

	return community.Encrypted(), nil
}

func (m *Manager) IsChannelEncrypted(communityID string, chatID string) (bool, error) {
	community, err := m.GetByIDString(communityID)
	if err != nil {
		return false, err
	}

	channelID := strings.TrimPrefix(chatID, communityID)
	return community.ChannelEncrypted(channelID), nil
}

func (m *Manager) ShouldHandleSyncCommunity(community *protobuf.SyncInstallationCommunity) (bool, error) {
	return m.persistence.ShouldHandleSyncCommunity(community)
}

func (m *Manager) ShouldHandleSyncCommunitySettings(communitySettings *protobuf.SyncCommunitySettings) (bool, error) {
	return m.persistence.ShouldHandleSyncCommunitySettings(communitySettings)
}

func (m *Manager) HandleSyncCommunitySettings(syncCommunitySettings *protobuf.SyncCommunitySettings) (*CommunitySettings, error) {
	id, err := types.DecodeHex(syncCommunitySettings.CommunityId)
	if err != nil {
		return nil, err
	}

	settings, err := m.persistence.GetCommunitySettingsByID(id)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		settings = &CommunitySettings{
			CommunityID:                  syncCommunitySettings.CommunityId,
			HistoryArchiveSupportEnabled: syncCommunitySettings.HistoryArchiveSupportEnabled,
			Clock:                        syncCommunitySettings.Clock,
		}
	}

	if syncCommunitySettings.Clock > settings.Clock {
		settings.CommunityID = syncCommunitySettings.CommunityId
		settings.HistoryArchiveSupportEnabled = syncCommunitySettings.HistoryArchiveSupportEnabled
		settings.Clock = syncCommunitySettings.Clock
	}

	err = m.persistence.SaveCommunitySettings(*settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (m *Manager) SetSyncClock(id []byte, clock uint64) error {
	return m.persistence.SetSyncClock(id, clock)
}

func (m *Manager) SetPrivateKey(id []byte, privKey *ecdsa.PrivateKey) error {
	return m.persistence.SetPrivateKey(id, privKey)
}

func (m *Manager) GetSyncedRawCommunity(id []byte) (*RawCommunityRow, error) {
	return m.persistence.getSyncedRawCommunity(id)
}

func (m *Manager) GetCommunitySettingsByID(id types.HexBytes) (*CommunitySettings, error) {
	return m.persistence.GetCommunitySettingsByID(id)
}

func (m *Manager) GetCommunitiesSettings() ([]CommunitySettings, error) {
	return m.persistence.GetCommunitiesSettings()
}

func (m *Manager) SaveCommunitySettings(settings CommunitySettings) error {
	return m.persistence.SaveCommunitySettings(settings)
}

func (m *Manager) CommunitySettingsExist(id types.HexBytes) (bool, error) {
	return m.persistence.CommunitySettingsExist(id)
}

func (m *Manager) DeleteCommunitySettings(id types.HexBytes) error {
	return m.persistence.DeleteCommunitySettings(id)
}

func (m *Manager) UpdateCommunitySettings(settings CommunitySettings) error {
	return m.persistence.UpdateCommunitySettings(settings)
}

func (m *Manager) GetOwnedCommunitiesChatIDs() (map[string]bool, error) {
	ownedCommunities, err := m.Controlled()
	if err != nil {
		return nil, err
	}

	chatIDs := make(map[string]bool)
	for _, c := range ownedCommunities {
		if c.Joined() {
			for _, id := range c.ChatIDs() {
				chatIDs[id] = true
			}
		}
	}
	return chatIDs, nil
}

func (m *Manager) GetCommunityChatsFilters(communityID types.HexBytes) ([]*transport.Filter, error) {
	chatIDs, err := m.persistence.GetCommunityChatIDs(communityID)
	if err != nil {
		return nil, err
	}

	filters := []*transport.Filter{}
	for _, cid := range chatIDs {
		filters = append(filters, m.transport.FilterByChatID(cid))
	}
	return filters, nil
}

func (m *Manager) GetCommunityChatsTopics(communityID types.HexBytes) ([]types.TopicType, error) {
	filters, err := m.GetCommunityChatsFilters(communityID)
	if err != nil {
		return nil, err
	}

	topics := []types.TopicType{}
	for _, filter := range filters {
		topics = append(topics, filter.ContentTopic)
	}

	return topics, nil
}

func (m *Manager) StoreWakuMessage(message *types.Message) error {
	return m.persistence.SaveWakuMessage(message)
}

func (m *Manager) StoreWakuMessages(messages []*types.Message) error {
	return m.persistence.SaveWakuMessages(messages)
}

func (m *Manager) GetLatestWakuMessageTimestamp(topics []types.TopicType) (uint64, error) {
	return m.persistence.GetLatestWakuMessageTimestamp(topics)
}

func (m *Manager) GetOldestWakuMessageTimestamp(topics []types.TopicType) (uint64, error) {
	return m.persistence.GetOldestWakuMessageTimestamp(topics)
}

func (m *Manager) GetLastMessageArchiveEndDate(communityID types.HexBytes) (uint64, error) {
	return m.persistence.GetLastMessageArchiveEndDate(communityID)
}

func (m *Manager) GetHistoryArchivePartitionStartTimestamp(communityID types.HexBytes) (uint64, error) {
	filters, err := m.GetCommunityChatsFilters(communityID)
	if err != nil {
		m.LogStdout("failed to get community chats filters", zap.Error(err))
		return 0, err
	}

	if len(filters) == 0 {
		// If we don't have chat filters, we likely don't have any chats
		// associated to this community, which means there's nothing more
		// to do here
		return 0, nil
	}

	topics := []types.TopicType{}

	for _, filter := range filters {
		topics = append(topics, filter.ContentTopic)
	}

	lastArchiveEndDateTimestamp, err := m.GetLastMessageArchiveEndDate(communityID)
	if err != nil {
		m.LogStdout("failed to get last archive end date", zap.Error(err))
		return 0, err
	}

	if lastArchiveEndDateTimestamp == 0 {
		// If we don't have a tracked last message archive end date, it
		// means we haven't created an archive before, which means
		// the next thing to look at is the oldest waku message timestamp for
		// this community
		lastArchiveEndDateTimestamp, err = m.GetOldestWakuMessageTimestamp(topics)
		if err != nil {
			m.LogStdout("failed to get oldest waku message timestamp", zap.Error(err))
			return 0, err
		}
		if lastArchiveEndDateTimestamp == 0 {
			// This means there's no waku message stored for this community so far
			// (even after requesting possibly missed messages), so no messages exist yet that can be archived
			m.LogStdout("can't find valid `lastArchiveEndTimestamp`")
			return 0, nil
		}
	}

	return lastArchiveEndDateTimestamp, nil
}

func (m *Manager) CreateAndSeedHistoryArchive(communityID types.HexBytes, topics []types.TopicType, startDate time.Time, endDate time.Time, partition time.Duration, encrypt bool) error {
	m.UnseedHistoryArchiveTorrent(communityID)
	_, err := m.CreateHistoryArchiveTorrentFromDB(communityID, topics, startDate, endDate, partition, encrypt)
	if err != nil {
		return err
	}
	return m.SeedHistoryArchiveTorrent(communityID)
}

func (m *Manager) StartHistoryArchiveTasksInterval(community *Community, interval time.Duration) {
	id := community.IDString()
	if _, exists := m.historyArchiveTasks.Load(id); exists {
		m.LogStdout("history archive tasks interval already in progres", zap.String("id", id))
		return
	}

	cancel := make(chan struct{})
	m.historyArchiveTasks.Store(id, cancel)
	m.historyArchiveTasksWaitGroup.Add(1)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.LogStdout("starting history archive tasks interval", zap.String("id", id))
	for {
		select {
		case <-ticker.C:
			m.LogStdout("starting archive task...", zap.String("id", id))
			lastArchiveEndDateTimestamp, err := m.GetHistoryArchivePartitionStartTimestamp(community.ID())
			if err != nil {
				m.LogStdout("failed to get last archive end date", zap.Error(err))
				continue
			}

			if lastArchiveEndDateTimestamp == 0 {
				// This means there are no waku messages for this community,
				// so nothing to do here
				m.LogStdout("couldn't determine archive start date - skipping")
				continue
			}

			topics, err := m.GetCommunityChatsTopics(community.ID())
			if err != nil {
				m.LogStdout("failed to get community chat topics ", zap.Error(err))
				continue
			}

			ts := time.Now().Unix()
			to := time.Unix(ts, 0)
			lastArchiveEndDate := time.Unix(int64(lastArchiveEndDateTimestamp), 0)

			err = m.CreateAndSeedHistoryArchive(community.ID(), topics, lastArchiveEndDate, to, interval, community.Encrypted())
			if err != nil {
				m.LogStdout("failed to create and seed history archive", zap.Error(err))
				continue
			}
		case <-cancel:
			m.UnseedHistoryArchiveTorrent(community.ID())
			m.historyArchiveTasks.Delete(id)
			m.historyArchiveTasksWaitGroup.Done()
			return
		}
	}
}

func (m *Manager) StopHistoryArchiveTasksIntervals() {
	m.historyArchiveTasks.Range(func(_, task interface{}) bool {
		close(task.(chan struct{})) // Need to cast to the chan
		return true
	})
	// Stoping archive interval tasks is async, so we need
	// to wait for all of them to be closed before we shutdown
	// the torrent client
	m.historyArchiveTasksWaitGroup.Wait()
}

func (m *Manager) StopHistoryArchiveTasksInterval(communityID types.HexBytes) {
	task, exists := m.historyArchiveTasks.Load(communityID.String())
	if exists {
		m.logger.Info("Stopping history archive tasks interval", zap.Any("id", communityID.String()))
		close(task.(chan struct{})) // Need to cast to the chan
	}
}

type EncodedArchiveData struct {
	padding int
	bytes   []byte
}

func (m *Manager) CreateHistoryArchiveTorrentFromMessages(communityID types.HexBytes, messages []*types.Message, topics []types.TopicType, startDate time.Time, endDate time.Time, partition time.Duration, encrypt bool) ([]string, error) {
	return m.CreateHistoryArchiveTorrent(communityID, messages, topics, startDate, endDate, partition, encrypt)
}

func (m *Manager) CreateHistoryArchiveTorrentFromDB(communityID types.HexBytes, topics []types.TopicType, startDate time.Time, endDate time.Time, partition time.Duration, encrypt bool) ([]string, error) {

	return m.CreateHistoryArchiveTorrent(communityID, make([]*types.Message, 0), topics, startDate, endDate, partition, encrypt)
}
func (m *Manager) CreateHistoryArchiveTorrent(communityID types.HexBytes, msgs []*types.Message, topics []types.TopicType, startDate time.Time, endDate time.Time, partition time.Duration, encrypt bool) ([]string, error) {

	loadFromDB := len(msgs) == 0

	from := startDate
	to := from.Add(partition)
	if to.After(endDate) {
		to = endDate
	}

	archiveDir := m.torrentConfig.DataDir + "/" + communityID.String()
	torrentDir := m.torrentConfig.TorrentDir
	indexPath := archiveDir + "/index"
	dataPath := archiveDir + "/data"

	wakuMessageArchiveIndexProto := &protobuf.WakuMessageArchiveIndex{}
	wakuMessageArchiveIndex := make(map[string]*protobuf.WakuMessageArchiveIndexMetadata)
	archiveIDs := make([]string, 0)

	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		err := os.MkdirAll(archiveDir, 0700)
		if err != nil {
			return archiveIDs, err
		}
	}
	if _, err := os.Stat(torrentDir); os.IsNotExist(err) {
		err := os.MkdirAll(torrentDir, 0700)
		if err != nil {
			return archiveIDs, err
		}
	}

	_, err := os.Stat(indexPath)
	if err == nil {
		wakuMessageArchiveIndexProto, err = m.LoadHistoryArchiveIndexFromFile(m.identity, communityID)
		if err != nil {
			return archiveIDs, err
		}
	}

	var offset uint64 = 0

	for hash, metadata := range wakuMessageArchiveIndexProto.Archives {
		offset = offset + metadata.Size
		wakuMessageArchiveIndex[hash] = metadata
	}

	var encodedArchives []*EncodedArchiveData
	topicsAsByteArrays := topicsAsByteArrays(topics)

	m.publish(&Subscription{CreatingHistoryArchivesSignal: &signal.CreatingHistoryArchivesSignal{
		CommunityID: communityID.String(),
	}})

	m.LogStdout("creating archives",
		zap.Any("startDate", startDate),
		zap.Any("endDate", endDate),
		zap.Duration("partition", partition),
	)
	for {
		if from.Equal(endDate) || from.After(endDate) {
			break
		}
		m.LogStdout("creating message archive",
			zap.Any("from", from),
			zap.Any("to", to),
		)

		var messages []types.Message
		if loadFromDB {
			messages, err = m.persistence.GetWakuMessagesByFilterTopic(topics, uint64(from.Unix()), uint64(to.Unix()))
			if err != nil {
				return archiveIDs, err
			}
		} else {
			for _, msg := range msgs {
				if int64(msg.Timestamp) >= from.Unix() && int64(msg.Timestamp) < to.Unix() {
					messages = append(messages, *msg)
				}
			}

		}

		if len(messages) == 0 {
			// No need to create an archive with zero messages
			m.LogStdout("no messages in this partition")
			from = to
			to = to.Add(partition)
			if to.After(endDate) {
				to = endDate
			}
			continue
		}

		// Not only do we partition messages, we also chunk them
		// roughly by size, such that each chunk will not exceed a given
		// size and archive data doesn't get too big
		messageChunks := make([][]types.Message, 0)
		currentChunkSize := 0
		currentChunk := make([]types.Message, 0)

		for _, msg := range messages {
			msgSize := len(msg.Payload) + len(msg.Sig)
			if msgSize > maxArchiveSizeInBytes {
				// we drop messages this big
				continue
			}

			if currentChunkSize+msgSize > maxArchiveSizeInBytes {
				messageChunks = append(messageChunks, currentChunk)
				currentChunk = make([]types.Message, 0)
				currentChunkSize = 0
			}
			currentChunk = append(currentChunk, msg)
			currentChunkSize = currentChunkSize + msgSize
		}
		messageChunks = append(messageChunks, currentChunk)

		for _, messages := range messageChunks {
			wakuMessageArchive := m.createWakuMessageArchive(from, to, messages, topicsAsByteArrays)
			encodedArchive, err := proto.Marshal(wakuMessageArchive)
			if err != nil {
				return archiveIDs, err
			}

			if encrypt {
				messageSpec, err := m.encryptor.BuildHashRatchetMessage(communityID, encodedArchive)
				if err != nil {
					return archiveIDs, err
				}

				encodedArchive, err = proto.Marshal(messageSpec.Message)
				if err != nil {
					return archiveIDs, err
				}
			}

			rawSize := len(encodedArchive)
			padding := 0
			size := 0

			if rawSize > pieceLength {
				size = rawSize + pieceLength - (rawSize % pieceLength)
				padding = size - rawSize
			} else {
				padding = pieceLength - rawSize
				size = rawSize + padding
			}

			wakuMessageArchiveIndexMetadata := &protobuf.WakuMessageArchiveIndexMetadata{
				Metadata: wakuMessageArchive.Metadata,
				Offset:   offset,
				Size:     uint64(size),
				Padding:  uint64(padding),
			}

			wakuMessageArchiveIndexMetadataBytes, err := proto.Marshal(wakuMessageArchiveIndexMetadata)
			if err != nil {
				return archiveIDs, err
			}

			archiveID := crypto.Keccak256Hash(wakuMessageArchiveIndexMetadataBytes).String()
			archiveIDs = append(archiveIDs, archiveID)
			wakuMessageArchiveIndex[archiveID] = wakuMessageArchiveIndexMetadata
			encodedArchives = append(encodedArchives, &EncodedArchiveData{bytes: encodedArchive, padding: padding})
			offset = offset + uint64(rawSize) + uint64(padding)
		}

		from = to
		to = to.Add(partition)
		if to.After(endDate) {
			to = endDate
		}
	}

	if len(encodedArchives) > 0 {

		dataBytes := make([]byte, 0)

		for _, encodedArchiveData := range encodedArchives {
			dataBytes = append(dataBytes, encodedArchiveData.bytes...)
			dataBytes = append(dataBytes, make([]byte, encodedArchiveData.padding)...)
		}

		wakuMessageArchiveIndexProto.Archives = wakuMessageArchiveIndex
		indexBytes, err := proto.Marshal(wakuMessageArchiveIndexProto)
		if err != nil {
			return archiveIDs, err
		}

		if encrypt {
			messageSpec, err := m.encryptor.BuildHashRatchetMessage(communityID, indexBytes)
			if err != nil {
				return archiveIDs, err
			}
			indexBytes, err = proto.Marshal(messageSpec.Message)
			if err != nil {
				return archiveIDs, err
			}
		}

		err = os.WriteFile(indexPath, indexBytes, 0644) // nolint: gosec
		if err != nil {
			return archiveIDs, err
		}

		file, err := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return archiveIDs, err
		}
		defer file.Close()

		_, err = file.Write(dataBytes)
		if err != nil {
			return archiveIDs, err
		}

		metaInfo := metainfo.MetaInfo{
			AnnounceList: defaultAnnounceList,
		}
		metaInfo.SetDefaults()
		metaInfo.CreatedBy = common.PubkeyToHex(&m.identity.PublicKey)

		info := metainfo.Info{
			PieceLength: int64(pieceLength),
		}

		err = info.BuildFromFilePath(archiveDir)
		if err != nil {
			return archiveIDs, err
		}

		metaInfo.InfoBytes, err = bencode.Marshal(info)
		if err != nil {
			return archiveIDs, err
		}

		metaInfoBytes, err := bencode.Marshal(metaInfo)
		if err != nil {
			return archiveIDs, err
		}

		err = os.WriteFile(m.torrentFile(communityID.String()), metaInfoBytes, 0644) // nolint: gosec
		if err != nil {
			return archiveIDs, err
		}

		m.LogStdout("torrent created", zap.Any("from", startDate.Unix()), zap.Any("to", endDate.Unix()))

		m.publish(&Subscription{
			HistoryArchivesCreatedSignal: &signal.HistoryArchivesCreatedSignal{
				CommunityID: communityID.String(),
				From:        int(startDate.Unix()),
				To:          int(endDate.Unix()),
			},
		})
	} else {
		m.LogStdout("no archives created")
		m.publish(&Subscription{
			NoHistoryArchivesCreatedSignal: &signal.NoHistoryArchivesCreatedSignal{
				CommunityID: communityID.String(),
				From:        int(startDate.Unix()),
				To:          int(endDate.Unix()),
			},
		})
	}

	lastMessageArchiveEndDate, err := m.persistence.GetLastMessageArchiveEndDate(communityID)
	if err != nil {
		return archiveIDs, err
	}

	if lastMessageArchiveEndDate > 0 {
		err = m.persistence.UpdateLastMessageArchiveEndDate(communityID, uint64(from.Unix()))
	} else {
		err = m.persistence.SaveLastMessageArchiveEndDate(communityID, uint64(from.Unix()))
	}
	if err != nil {
		return archiveIDs, err
	}
	return archiveIDs, nil
}

func (m *Manager) SeedHistoryArchiveTorrent(communityID types.HexBytes) error {
	m.UnseedHistoryArchiveTorrent(communityID)

	id := communityID.String()
	torrentFile := m.torrentFile(id)

	metaInfo, err := metainfo.LoadFromFile(torrentFile)
	if err != nil {
		return err
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return err
	}

	hash := metaInfo.HashInfoBytes()
	m.torrentTasks[id] = hash

	if err != nil {
		return err
	}

	torrent, err := m.torrentClient.AddTorrent(metaInfo)
	if err != nil {
		return err
	}

	torrent.DownloadAll()

	m.publish(&Subscription{
		HistoryArchivesSeedingSignal: &signal.HistoryArchivesSeedingSignal{
			CommunityID: communityID.String(),
		},
	})

	magnetLink := metaInfo.Magnet(nil, &info).String()

	m.LogStdout("seeding torrent", zap.String("id", id), zap.String("magnetLink", magnetLink))
	return nil
}

func (m *Manager) UnseedHistoryArchiveTorrent(communityID types.HexBytes) {
	id := communityID.String()

	hash, exists := m.torrentTasks[id]

	if exists {
		torrent, ok := m.torrentClient.Torrent(hash)
		if ok {
			m.logger.Debug("Unseeding and dropping torrent for community: ", zap.Any("id", id))
			torrent.Drop()
			delete(m.torrentTasks, id)

			m.publish(&Subscription{
				HistoryArchivesUnseededSignal: &signal.HistoryArchivesUnseededSignal{
					CommunityID: id,
				},
			})
		}
	}
}

func (m *Manager) IsSeedingHistoryArchiveTorrent(communityID types.HexBytes) bool {
	id := communityID.String()
	hash := m.torrentTasks[id]
	torrent, ok := m.torrentClient.Torrent(hash)
	return ok && torrent.Seeding()
}

func (m *Manager) GetHistoryArchiveDownloadTask(communityID string) *HistoryArchiveDownloadTask {
	return m.historyArchiveDownloadTasks[communityID]
}

func (m *Manager) DeleteHistoryArchiveDownloadTask(communityID string) {
	delete(m.historyArchiveDownloadTasks, communityID)
}

func (m *Manager) AddHistoryArchiveDownloadTask(communityID string, task *HistoryArchiveDownloadTask) {
	m.historyArchiveDownloadTasks[communityID] = task
}

type HistoryArchiveDownloadTaskInfo struct {
	TotalDownloadedArchivesCount int
	TotalArchivesCount           int
	Cancelled                    bool
}

func (m *Manager) DownloadHistoryArchivesByMagnetlink(communityID types.HexBytes, magnetlink string, cancelTask chan struct{}) (*HistoryArchiveDownloadTaskInfo, error) {

	id := communityID.String()

	ml, err := metainfo.ParseMagnetUri(magnetlink)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("adding torrent via magnetlink for community", zap.String("id", id), zap.String("magnetlink", magnetlink))
	torrent, err := m.torrentClient.AddMagnet(magnetlink)
	if err != nil {
		return nil, err
	}

	downloadTaskInfo := &HistoryArchiveDownloadTaskInfo{
		TotalDownloadedArchivesCount: 0,
		TotalArchivesCount:           0,
		Cancelled:                    false,
	}

	m.torrentTasks[id] = ml.InfoHash
	timeout := time.After(20 * time.Second)

	m.LogStdout("fetching torrent info", zap.String("magnetlink", magnetlink))
	select {
	case <-timeout:
		return nil, ErrTorrentTimedout
	case <-cancelTask:
		m.LogStdout("cancelled fetching torrent info")
		downloadTaskInfo.Cancelled = true
		return downloadTaskInfo, nil
	case <-torrent.GotInfo():

		files := torrent.Files()

		i, ok := findIndexFile(files)
		if !ok {
			// We're dealing with a malformed torrent, so don't do anything
			return nil, errors.New("malformed torrent data")
		}

		indexFile := files[i]
		indexFile.Download()

		m.LogStdout("downloading history archive index")
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-cancelTask:
				m.LogStdout("cancelled downloading archive index")
				downloadTaskInfo.Cancelled = true
				return downloadTaskInfo, nil
			case <-ticker.C:
				if indexFile.BytesCompleted() == indexFile.Length() {

					index, err := m.LoadHistoryArchiveIndexFromFile(m.identity, communityID)
					if err != nil {
						return nil, err
					}

					existingArchiveIDs, err := m.persistence.GetDownloadedMessageArchiveIDs(communityID)
					if err != nil {
						return nil, err
					}

					if len(existingArchiveIDs) == len(index.Archives) {
						m.LogStdout("download cancelled, no new archives")
						return downloadTaskInfo, nil
					}

					downloadTaskInfo.TotalDownloadedArchivesCount = len(existingArchiveIDs)
					downloadTaskInfo.TotalArchivesCount = len(index.Archives)

					archiveHashes := make(archiveMDSlice, 0, downloadTaskInfo.TotalArchivesCount)

					for hash, metadata := range index.Archives {
						archiveHashes = append(archiveHashes, &archiveMetadata{hash: hash, from: metadata.Metadata.From})
					}

					sort.Sort(sort.Reverse(archiveHashes))

					m.publish(&Subscription{
						DownloadingHistoryArchivesStartedSignal: &signal.DownloadingHistoryArchivesStartedSignal{
							CommunityID: communityID.String(),
						},
					})

					for _, hd := range archiveHashes {

						hash := hd.hash
						hasArchive := false

						for _, existingHash := range existingArchiveIDs {
							if existingHash == hash {
								hasArchive = true
								break
							}
						}
						if hasArchive {
							continue
						}

						metadata := index.Archives[hash]
						startIndex := int(metadata.Offset) / pieceLength
						endIndex := startIndex + int(metadata.Size)/pieceLength

						downloadMsg := fmt.Sprintf("downloading data for message archive (%d/%d)", downloadTaskInfo.TotalDownloadedArchivesCount+1, downloadTaskInfo.TotalArchivesCount)
						m.LogStdout(downloadMsg, zap.String("hash", hash))
						m.LogStdout("pieces (start, end)", zap.Any("startIndex", startIndex), zap.Any("endIndex", endIndex-1))
						torrent.DownloadPieces(startIndex, endIndex)

						piecesCompleted := make(map[int]bool)
						for i = startIndex; i < endIndex; i++ {
							piecesCompleted[i] = false
						}

						psc := torrent.SubscribePieceStateChanges()
						downloadTicker := time.NewTicker(1 * time.Second)
						defer downloadTicker.Stop()

					downloadLoop:
						for {
							select {
							case <-downloadTicker.C:
								done := true
								for i = startIndex; i < endIndex; i++ {
									piecesCompleted[i] = torrent.PieceState(i).Complete
									if !piecesCompleted[i] {
										done = false
									}
								}
								if done {
									psc.Close()
									break downloadLoop
								}
							case <-cancelTask:
								m.LogStdout("downloading archive data interrupted")
								downloadTaskInfo.Cancelled = true
								return downloadTaskInfo, nil
							}
						}
						downloadTaskInfo.TotalDownloadedArchivesCount++
						err = m.persistence.SaveMessageArchiveID(communityID, hash)
						if err != nil {
							m.LogStdout("couldn't save message archive ID", zap.Error(err))
							continue
						}
						m.publish(&Subscription{
							HistoryArchiveDownloadedSignal: &signal.HistoryArchiveDownloadedSignal{
								CommunityID: communityID.String(),
								From:        int(metadata.Metadata.From),
								To:          int(metadata.Metadata.To),
							},
						})
					}
					m.publish(&Subscription{
						HistoryArchivesSeedingSignal: &signal.HistoryArchivesSeedingSignal{
							CommunityID: communityID.String(),
						},
					})
					m.LogStdout("finished downloading archives")
					return downloadTaskInfo, nil
				}
			}
		}
	}
}

func (m *Manager) GetMessageArchiveIDsToImport(communityID types.HexBytes) ([]string, error) {
	return m.persistence.GetMessageArchiveIDsToImport(communityID)
}

func (m *Manager) ExtractMessagesFromHistoryArchive(communityID types.HexBytes, archiveID string) ([]*protobuf.WakuMessage, error) {
	id := communityID.String()

	index, err := m.LoadHistoryArchiveIndexFromFile(m.identity, communityID)
	if err != nil {
		return nil, err
	}

	dataFile, err := os.Open(m.archiveDataFile(id))
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()

	m.LogStdout("extracting messages from history archive", zap.String("archive id", archiveID))
	metadata := index.Archives[archiveID]

	_, err = dataFile.Seek(int64(metadata.Offset), 0)
	if err != nil {
		m.LogStdout("failed to seek archive data file", zap.Error(err))
		return nil, err
	}

	data := make([]byte, metadata.Size-metadata.Padding)
	m.LogStdout("loading history archive data into memory", zap.Float64("data_size_MB", float64(metadata.Size-metadata.Padding)/1024.0/1024.0))
	_, err = dataFile.Read(data)
	if err != nil {
		m.LogStdout("failed failed to read archive data", zap.Error(err))
		return nil, err
	}

	archive := &protobuf.WakuMessageArchive{}

	err = proto.Unmarshal(data, archive)
	if err != nil {
		// The archive data might eb encrypted so we try to decrypt instead first
		var protocolMessage encryption.ProtocolMessage
		err := proto.Unmarshal(data, &protocolMessage)
		if err != nil {
			m.LogStdout("failed to unmarshal protocol message", zap.Error(err))
			return nil, err
		}

		pk, err := crypto.DecompressPubkey(communityID)
		if err != nil {
			m.logger.Debug("failed to decompress community pubkey", zap.Error(err))
			return nil, err
		}
		decryptedBytes, err := m.encryptor.HandleMessage(m.identity, pk, &protocolMessage, make([]byte, 0))
		if err != nil {
			m.LogStdout("failed to decrypt message archive", zap.Error(err))
			return nil, err
		}
		err = proto.Unmarshal(decryptedBytes.DecryptedMessage, archive)
		if err != nil {
			m.LogStdout("failed to unmarshal message archive data", zap.Error(err))
			return nil, err
		}
	}
	return archive.Messages, nil
}

func (m *Manager) SetMessageArchiveIDImported(communityID types.HexBytes, hash string, imported bool) error {
	return m.persistence.SetMessageArchiveIDImported(communityID, hash, imported)
}

func (m *Manager) GetHistoryArchiveMagnetlink(communityID types.HexBytes) (string, error) {
	id := communityID.String()
	torrentFile := m.torrentFile(id)

	metaInfo, err := metainfo.LoadFromFile(torrentFile)
	if err != nil {
		return "", err
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return "", err
	}

	return metaInfo.Magnet(nil, &info).String(), nil
}

func (m *Manager) createWakuMessageArchive(from time.Time, to time.Time, messages []types.Message, topics [][]byte) *protobuf.WakuMessageArchive {
	var wakuMessages []*protobuf.WakuMessage

	for _, msg := range messages {
		topic := types.TopicTypeToByteArray(msg.Topic)
		wakuMessage := &protobuf.WakuMessage{
			Sig:          msg.Sig,
			Timestamp:    uint64(msg.Timestamp),
			Topic:        topic,
			Payload:      msg.Payload,
			Padding:      msg.Padding,
			Hash:         msg.Hash,
			ThirdPartyId: msg.ThirdPartyID,
		}
		wakuMessages = append(wakuMessages, wakuMessage)
	}

	metadata := protobuf.WakuMessageArchiveMetadata{
		From:         uint64(from.Unix()),
		To:           uint64(to.Unix()),
		ContentTopic: topics,
	}

	wakuMessageArchive := &protobuf.WakuMessageArchive{
		Metadata: &metadata,
		Messages: wakuMessages,
	}
	return wakuMessageArchive
}

func (m *Manager) LoadHistoryArchiveIndexFromFile(myKey *ecdsa.PrivateKey, communityID types.HexBytes) (*protobuf.WakuMessageArchiveIndex, error) {
	wakuMessageArchiveIndexProto := &protobuf.WakuMessageArchiveIndex{}

	indexPath := m.archiveIndexFile(communityID.String())
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(indexData, wakuMessageArchiveIndexProto)
	if err != nil {
		return nil, err
	}

	if len(wakuMessageArchiveIndexProto.Archives) == 0 && len(indexData) > 0 {
		// This means we're dealing with an encrypted index file, so we have to decrypt it first
		var protocolMessage encryption.ProtocolMessage
		err := proto.Unmarshal(indexData, &protocolMessage)
		if err != nil {
			return nil, err
		}
		pk, err := crypto.DecompressPubkey(communityID)
		if err != nil {
			return nil, err
		}
		decryptedBytes, err := m.encryptor.HandleMessage(myKey, pk, &protocolMessage, make([]byte, 0))
		if err != nil {
			return nil, err
		}
		err = proto.Unmarshal(decryptedBytes.DecryptedMessage, wakuMessageArchiveIndexProto)
		if err != nil {
			return nil, err
		}
	}

	return wakuMessageArchiveIndexProto, nil
}

func (m *Manager) TorrentFileExists(communityID string) bool {
	_, err := os.Stat(m.torrentFile(communityID))
	return err == nil
}

func (m *Manager) torrentFile(communityID string) string {
	return m.torrentConfig.TorrentDir + "/" + communityID + ".torrent"
}

func (m *Manager) archiveIndexFile(communityID string) string {
	return m.torrentConfig.DataDir + "/" + communityID + "/index"
}

func (m *Manager) archiveDataFile(communityID string) string {
	return m.torrentConfig.DataDir + "/" + communityID + "/data"
}

func topicsAsByteArrays(topics []types.TopicType) [][]byte {
	var topicsAsByteArrays [][]byte
	for _, t := range topics {
		topic := types.TopicTypeToByteArray(t)
		topicsAsByteArrays = append(topicsAsByteArrays, topic)
	}
	return topicsAsByteArrays
}

func findIndexFile(files []*torrent.File) (index int, ok bool) {
	for i, f := range files {
		if f.DisplayPath() == "index" {
			return i, true
		}
	}
	return 0, false
}

func (m *Manager) GetCommunityToken(communityID string, chainID int, address string) (*community_token.CommunityToken, error) {
	return m.persistence.GetCommunityToken(communityID, chainID, address)
}

func (m *Manager) GetCommunityTokens(communityID string) ([]*community_token.CommunityToken, error) {
	return m.persistence.GetCommunityTokens(communityID)
}

func (m *Manager) GetAllCommunityTokens() ([]*community_token.CommunityToken, error) {
	return m.persistence.GetAllCommunityTokens()
}

func (m *Manager) ImageToBase64(uri string) string {
	if uri == "" {
		return ""
	}
	file, err := os.Open(uri)
	if err != nil {
		m.logger.Error(err.Error())
		return ""
	}
	defer file.Close()

	payload, err := ioutil.ReadAll(file)
	if err != nil {
		m.logger.Error(err.Error())
		return ""
	}
	base64img, err := images.GetPayloadDataURI(payload)
	if err != nil {
		m.logger.Error(err.Error())
		return ""
	}
	return base64img
}

func (m *Manager) SaveCommunityToken(token *community_token.CommunityToken, croppedImage *images.CroppedImage) (*community_token.CommunityToken, error) {

	_, err := m.GetByIDString(token.CommunityID)
	if err != nil {
		return nil, err
	}

	if croppedImage != nil && croppedImage.ImagePath != "" {
		bytes, err := images.OpenAndAdjustImage(*croppedImage, true)
		if err != nil {
			return nil, err
		}

		base64img, err := images.GetPayloadDataURI(bytes)
		if err != nil {
			return nil, err
		}
		token.Base64Image = base64img
	} else if !images.IsPayloadDataURI(token.Base64Image) {
		// if image is already base64 do not convert (owner and master tokens have already base64 image)
		token.Base64Image = m.ImageToBase64(token.Base64Image)
	}

	return token, m.persistence.AddCommunityToken(token)
}

func (m *Manager) AddCommunityToken(token *community_token.CommunityToken, clock uint64) (*Community, error) {
	if token == nil {
		return nil, errors.New("Token is absent in database")
	}

	community, err := m.GetByIDString(token.CommunityID)
	if err != nil {
		return nil, err
	}

	if !community.MemberCanManageToken(&m.identity.PublicKey, token) {
		return nil, ErrInvalidManageTokensPermission
	}

	tokenMetadata := &protobuf.CommunityTokenMetadata{
		ContractAddresses: map[uint64]string{uint64(token.ChainID): token.Address},
		Description:       token.Description,
		Image:             token.Base64Image,
		Symbol:            token.Symbol,
		TokenType:         token.TokenType,
		Name:              token.Name,
		Decimals:          uint32(token.Decimals),
	}
	_, err = community.AddCommunityTokensMetadata(tokenMetadata)
	if err != nil {
		return nil, err
	}

	if community.IsControlNode() && (token.PrivilegesLevel == community_token.MasterLevel || token.PrivilegesLevel == community_token.OwnerLevel) {
		permissionType := protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER
		if token.PrivilegesLevel == community_token.MasterLevel {
			permissionType = protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER
		}

		contractAddresses := make(map[uint64]string)
		contractAddresses[uint64(token.ChainID)] = token.Address

		tokenCriteria := &protobuf.TokenCriteria{
			ContractAddresses: contractAddresses,
			Type:              protobuf.CommunityTokenType_ERC721,
			Symbol:            token.Symbol,
			Name:              token.Name,
			Amount:            "1",
			Decimals:          uint64(token.Decimals),
		}

		request := &requests.CreateCommunityTokenPermission{
			CommunityID:   community.ID(),
			Type:          permissionType,
			TokenCriteria: []*protobuf.TokenCriteria{tokenCriteria},
			IsPrivate:     true,
			ChatIds:       []string{},
		}

		community, _, err = m.createCommunityTokenPermission(request, community)
		if err != nil {
			return nil, err
		}

		if token.PrivilegesLevel == community_token.OwnerLevel {
			_, err = m.promoteSelfToControlNode(community, clock)
			if err != nil {
				return nil, err
			}
		}
	}

	return community, m.saveAndPublish(community)
}

func (m *Manager) UpdateCommunityTokenState(chainID int, contractAddress string, deployState community_token.DeployState) error {
	return m.persistence.UpdateCommunityTokenState(chainID, contractAddress, deployState)
}

func (m *Manager) UpdateCommunityTokenAddress(chainID int, oldContractAddress string, newContractAddress string) error {
	return m.persistence.UpdateCommunityTokenAddress(chainID, oldContractAddress, newContractAddress)
}

func (m *Manager) UpdateCommunityTokenSupply(chainID int, contractAddress string, supply *bigint.BigInt) error {
	return m.persistence.UpdateCommunityTokenSupply(chainID, contractAddress, supply)
}

func (m *Manager) RemoveCommunityToken(chainID int, contractAddress string) error {
	return m.persistence.RemoveCommunityToken(chainID, contractAddress)
}

func (m *Manager) SetCommunityActiveMembersCount(communityID string, activeMembersCount uint64) error {
	community, err := m.GetByIDString(communityID)
	if err != nil {
		return err
	}

	updated, err := community.SetActiveMembersCount(activeMembersCount)
	if err != nil {
		return err
	}

	if updated {
		if err = m.persistence.SaveCommunity(community); err != nil {
			return err
		}

		m.publish(&Subscription{Community: community})
	}

	return nil
}

// UpdateCommunity takes a Community persists it and republishes it.
// The clock is incremented meaning even a no change update will be republished by the admin, and parsed by the member.
func (m *Manager) UpdateCommunity(c *Community) error {
	c.increaseClock()

	err := m.persistence.SaveCommunity(c)
	if err != nil {
		return err
	}

	m.publish(&Subscription{Community: c})
	return nil
}

func combineAddressesAndChainIDs(addresses []gethcommon.Address, chainIDs []uint64) []*AccountChainIDsCombination {
	combinations := make([]*AccountChainIDsCombination, 0)
	for _, address := range addresses {
		combinations = append(combinations, &AccountChainIDsCombination{
			Address:  address,
			ChainIDs: chainIDs,
		})
	}
	return combinations
}

func revealedAccountsToAccountsAndChainIDsCombination(revealedAccounts []*protobuf.RevealedAccount) []*AccountChainIDsCombination {
	accountsAndChainIDs := make([]*AccountChainIDsCombination, 0)
	for _, revealedAccount := range revealedAccounts {
		accountsAndChainIDs = append(accountsAndChainIDs, &AccountChainIDsCombination{
			Address:  gethcommon.HexToAddress(revealedAccount.Address),
			ChainIDs: revealedAccount.ChainIds,
		})
	}
	return accountsAndChainIDs
}

func (m *Manager) accountsHasPrivilegedPermission(privilegedPermissions []*CommunityTokenPermission, accounts []*AccountChainIDsCombination) bool {
	if len(privilegedPermissions) > 0 {
		permissionResponse, err := m.PermissionChecker.CheckPermissions(privilegedPermissions, accounts, true)
		if err != nil {
			m.logger.Warn("check privileged permission failed: %v", zap.Error(err))
			return false
		}
		return permissionResponse.Satisfied
	}
	return false
}

func (m *Manager) saveAndPublish(community *Community) error {
	err := m.persistence.SaveCommunity(community)
	if err != nil {
		return err
	}

	if community.IsControlNode() {
		m.publish(&Subscription{Community: community})
		return nil
	} else if community.HasPermissionToSendCommunityEvents() {
		err := m.signEvents(community)
		if err != nil {
			return err
		}
		err = m.persistence.SaveCommunityEvents(community)
		if err != nil {
			return err
		}

		m.publish(&Subscription{CommunityEventsMessage: community.ToCommunityEventsMessage()})
		return nil
	}

	return nil
}

func (m *Manager) GetRevealedAddresses(communityID types.HexBytes, memberPk string) ([]*protobuf.RevealedAccount, error) {
	requestID := CalculateRequestID(memberPk, communityID)
	return m.persistence.GetRequestToJoinRevealedAddresses(requestID)
}

func (m *Manager) ReevaluatePrivilegedMember(community *Community, tokenPermissions []*CommunityTokenPermission,
	accountsAndChainIDs []*AccountChainIDsCombination, memberPubKey *ecdsa.PublicKey,
	privilegedRole protobuf.CommunityMember_Roles, alreadyHasPrivilegedRole bool) (bool, error) {

	hasPrivilegedRolePermissions := len(tokenPermissions) > 0
	removeCurrentRole := false

	if hasPrivilegedRolePermissions {
		permissionResponse, err := m.PermissionChecker.CheckPermissions(tokenPermissions, accountsAndChainIDs, true)
		if err != nil {
			return alreadyHasPrivilegedRole, err
		} else if permissionResponse.Satisfied && !alreadyHasPrivilegedRole {
			_, err = community.AddRoleToMember(memberPubKey, privilegedRole)
			if err != nil {
				return alreadyHasPrivilegedRole, err
			}
			alreadyHasPrivilegedRole = true
		} else if !permissionResponse.Satisfied && alreadyHasPrivilegedRole {
			removeCurrentRole = true
			alreadyHasPrivilegedRole = false
		}
	}

	// Remove privileged role if user does not pass role permissions check or
	// Community does not have permissions but user has a role
	if removeCurrentRole || (!hasPrivilegedRolePermissions && alreadyHasPrivilegedRole) {
		_, err := community.RemoveRoleFromMember(memberPubKey, privilegedRole)
		if err != nil {
			return alreadyHasPrivilegedRole, err
		}
		alreadyHasPrivilegedRole = false
	}

	if alreadyHasPrivilegedRole {
		// Make sure privileged user is added to every channel
		for channelID := range community.Chats() {
			if !community.IsMemberInChat(memberPubKey, channelID) {
				_, err := community.AddMemberToChat(channelID, memberPubKey, []protobuf.CommunityMember_Roles{privilegedRole})
				if err != nil {
					return alreadyHasPrivilegedRole, err
				}
			}
		}
	}

	return alreadyHasPrivilegedRole, nil
}

func (m *Manager) handleCommunityTokensMetadata(community *Community) error {
	communityID := community.IDString()
	communityTokens := community.CommunityTokensMetadata()

	if len(communityTokens) == 0 {
		return nil
	}
	for _, tokenMetadata := range communityTokens {
		for chainID, address := range tokenMetadata.ContractAddresses {
			exists, err := m.persistence.HasCommunityToken(communityID, address, int(chainID))
			if err != nil {
				return err
			}
			if !exists {
				// Fetch community token to make sure it's stored in the DB, discard result
				communityToken, err := m.FetchCommunityToken(community, tokenMetadata, chainID, address)
				if err != nil {
					return err
				}

				err = m.persistence.AddCommunityToken(communityToken)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *Manager) FetchCommunityToken(community *Community, tokenMetadata *protobuf.CommunityTokenMetadata, chainID uint64, contractAddress string) (*community_token.CommunityToken, error) {
	communityID := community.IDString()

	communityToken := &community_token.CommunityToken{
		CommunityID:        communityID,
		Address:            contractAddress,
		TokenType:          tokenMetadata.TokenType,
		Name:               tokenMetadata.Name,
		Symbol:             tokenMetadata.Symbol,
		Description:        tokenMetadata.Description,
		Transferable:       true,
		RemoteSelfDestruct: false,
		ChainID:            int(chainID),
		DeployState:        community_token.Deployed,
		Base64Image:        tokenMetadata.Image,
		Decimals:           int(tokenMetadata.Decimals),
	}

	switch tokenMetadata.TokenType {
	case protobuf.CommunityTokenType_ERC721:
		contractData, err := m.communityTokensService.GetCollectibleContractData(chainID, contractAddress)
		if err != nil {
			return nil, err
		}

		communityToken.Supply = contractData.TotalSupply
		communityToken.Transferable = contractData.Transferable
		communityToken.RemoteSelfDestruct = contractData.RemoteBurnable
		communityToken.InfiniteSupply = contractData.InfiniteSupply

	case protobuf.CommunityTokenType_ERC20:
		contractData, err := m.communityTokensService.GetAssetContractData(chainID, contractAddress)
		if err != nil {
			return nil, err
		}

		communityToken.Supply = contractData.TotalSupply
		communityToken.InfiniteSupply = contractData.InfiniteSupply
	}

	communityToken.PrivilegesLevel = getPrivilegesLevel(chainID, contractAddress, community.TokenPermissions())

	return communityToken, nil
}

func getPrivilegesLevel(chainID uint64, tokenAddress string, tokenPermissions map[string]*CommunityTokenPermission) community_token.PrivilegesLevel {
	for _, permission := range tokenPermissions {
		if permission.Type == protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER || permission.Type == protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER {
			for _, tokenCriteria := range permission.TokenCriteria {
				value, exist := tokenCriteria.ContractAddresses[chainID]
				if exist && value == tokenAddress {
					if permission.Type == protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER {
						return community_token.OwnerLevel
					}
					return community_token.MasterLevel
				}
			}
		}
	}
	return community_token.CommunityLevel
}

func (m *Manager) ValidateCommunityPrivilegedUserSyncMessage(message *protobuf.CommunityPrivilegedUserSyncMessage) error {
	if message == nil {
		return errors.New("invalid CommunityPrivilegedUserSyncMessage message")
	}

	if message.CommunityId == nil || len(message.CommunityId) == 0 {
		return errors.New("invalid CommunityId in CommunityPrivilegedUserSyncMessage message")
	}

	switch message.Type {
	case protobuf.CommunityPrivilegedUserSyncMessage_CONTROL_NODE_ACCEPT_REQUEST_TO_JOIN:
		fallthrough
	case protobuf.CommunityPrivilegedUserSyncMessage_CONTROL_NODE_REJECT_REQUEST_TO_JOIN:
		if message.RequestToJoin == nil || len(message.RequestToJoin) == 0 {
			return errors.New("invalid request to join in CommunityPrivilegedUserSyncMessage message")
		}

		for _, requestToJoinProto := range message.RequestToJoin {
			if len(requestToJoinProto.CommunityId) == 0 {
				return errors.New("no communityId in request to join in CommunityPrivilegedUserSyncMessage message")
			}
		}
	case protobuf.CommunityPrivilegedUserSyncMessage_CONTROL_NODE_ALL_SYNC_REQUESTS_TO_JOIN:
		if message.SyncRequestsToJoin == nil || len(message.SyncRequestsToJoin) == 0 {
			return errors.New("invalid sync requests to join in CommunityPrivilegedUserSyncMessage message")
		}
	}

	return nil
}

func (m *Manager) createCommunityTokenPermission(request *requests.CreateCommunityTokenPermission, community *Community) (*Community, *CommunityChanges, error) {
	if community == nil {
		return nil, nil, ErrOrgNotFound
	}

	tokenPermission := request.ToCommunityTokenPermission()
	tokenPermission.Id = uuid.New().String()
	changes, err := community.UpsertTokenPermission(&tokenPermission)
	if err != nil {
		return nil, nil, err
	}

	return community, changes, nil

}

func (m *Manager) PromoteSelfToControlNode(community *Community, clock uint64) (*CommunityChanges, error) {
	if community == nil {
		return nil, ErrOrgNotFound
	}

	ownerChanged, err := m.promoteSelfToControlNode(community, clock)
	if err != nil {
		return nil, err
	}

	if ownerChanged {
		return community.RemoveAllUsersFromOrg(), m.saveAndPublish(community)
	}

	return community.emptyCommunityChanges(), m.saveAndPublish(community)
}

func (m *Manager) promoteSelfToControlNode(community *Community, clock uint64) (bool, error) {
	ownerChanged := false
	community.setPrivateKey(m.identity)
	if !community.ControlNode().Equal(&m.identity.PublicKey) {
		ownerChanged = true
		community.setControlNode(&m.identity.PublicKey)
	}

	// Mark this device as the control node
	syncControlNode := &protobuf.SyncCommunityControlNode{
		Clock:          clock,
		InstallationId: m.installationID,
	}

	err := m.SaveSyncControlNode(community.ID(), syncControlNode)
	if err != nil {
		return false, err
	}
	community.config.ControlDevice = true

	if exists := community.HasMember(&m.identity.PublicKey); !exists {
		ownerRole := []protobuf.CommunityMember_Roles{protobuf.CommunityMember_ROLE_OWNER}
		_, err = community.AddMember(&m.identity.PublicKey, ownerRole)
		if err != nil {
			return false, err
		}

		for channelID := range community.Chats() {
			_, err = community.AddMemberToChat(channelID, &m.identity.PublicKey, ownerRole)
			if err != nil {
				return false, err
			}
		}
	} else {
		_, err = community.AddRoleToMember(&m.identity.PublicKey, protobuf.CommunityMember_ROLE_OWNER)
	}

	if err != nil {
		return false, err
	}

	community.increaseClock()

	return ownerChanged, nil
}

func (m *Manager) shareRequestsToJoinWithNewPrivilegedMembers(community *Community, newPrivilegedMembers map[protobuf.CommunityMember_Roles][]*ecdsa.PublicKey) error {
	requestsToJoin, err := m.GetCommunityRequestsToJoinWithRevealedAddresses(community.ID())
	if err != nil {
		return err
	}

	var syncRequestsWithoutRevealedAccounts []*protobuf.SyncCommunityRequestsToJoin
	var syncRequestsWithRevealedAccounts []*protobuf.SyncCommunityRequestsToJoin
	for _, request := range requestsToJoin {
		syncRequestsWithRevealedAccounts = append(syncRequestsWithRevealedAccounts, request.ToSyncProtobuf())
		requestProtoWithoutAccounts := request.ToSyncProtobuf()
		requestProtoWithoutAccounts.RevealedAccounts = []*protobuf.RevealedAccount{}
		syncRequestsWithoutRevealedAccounts = append(syncRequestsWithoutRevealedAccounts, requestProtoWithoutAccounts)
	}

	syncMsgWithoutRevealedAccounts := &protobuf.CommunityPrivilegedUserSyncMessage{
		Type:               protobuf.CommunityPrivilegedUserSyncMessage_CONTROL_NODE_ALL_SYNC_REQUESTS_TO_JOIN,
		CommunityId:        community.ID(),
		SyncRequestsToJoin: syncRequestsWithoutRevealedAccounts,
	}

	syncMsgWitRevealedAccounts := &protobuf.CommunityPrivilegedUserSyncMessage{
		Type:               protobuf.CommunityPrivilegedUserSyncMessage_CONTROL_NODE_ALL_SYNC_REQUESTS_TO_JOIN,
		CommunityId:        community.ID(),
		SyncRequestsToJoin: syncRequestsWithRevealedAccounts,
	}

	subscriptionMsg := &CommunityPrivilegedMemberSyncMessage{
		CommunityPrivateKey: community.PrivateKey(),
	}

	for role, members := range newPrivilegedMembers {
		if len(members) == 0 {
			continue
		}

		subscriptionMsg.Receivers = members

		switch role {
		case protobuf.CommunityMember_ROLE_ADMIN:
			subscriptionMsg.CommunityPrivilegedUserSyncMessage = syncMsgWithoutRevealedAccounts
		case protobuf.CommunityMember_ROLE_OWNER:
			continue
		case protobuf.CommunityMember_ROLE_TOKEN_MASTER:
			subscriptionMsg.CommunityPrivilegedUserSyncMessage = syncMsgWitRevealedAccounts
		}

		m.publish(&Subscription{CommunityPrivilegedMemberSyncMessage: subscriptionMsg})
	}

	return nil
}

func (m *Manager) shareAcceptedRequestToJoinWithPrivilegedMembers(community *Community, requestsToJoin *RequestToJoin) error {
	pk, err := common.HexToPubkey(requestsToJoin.PublicKey)
	if err != nil {
		return err
	}

	acceptedRequestsToJoinWithoutRevealedAccounts := make(map[string]*protobuf.CommunityRequestToJoin)
	acceptedRequestsToJoinWithRevealedAccounts := make(map[string]*protobuf.CommunityRequestToJoin)

	acceptedRequestsToJoinWithRevealedAccounts[requestsToJoin.PublicKey] = requestsToJoin.ToCommunityRequestToJoinProtobuf()
	requestsToJoin.RevealedAccounts = make([]*protobuf.RevealedAccount, 0)
	acceptedRequestsToJoinWithoutRevealedAccounts[requestsToJoin.PublicKey] = requestsToJoin.ToCommunityRequestToJoinProtobuf()

	msgWithRevealedAccounts := &protobuf.CommunityPrivilegedUserSyncMessage{
		Type:          protobuf.CommunityPrivilegedUserSyncMessage_CONTROL_NODE_ACCEPT_REQUEST_TO_JOIN,
		CommunityId:   community.ID(),
		RequestToJoin: acceptedRequestsToJoinWithRevealedAccounts,
	}

	msgWithoutRevealedAccounts := &protobuf.CommunityPrivilegedUserSyncMessage{
		Type:          protobuf.CommunityPrivilegedUserSyncMessage_CONTROL_NODE_ACCEPT_REQUEST_TO_JOIN,
		CommunityId:   community.ID(),
		RequestToJoin: acceptedRequestsToJoinWithoutRevealedAccounts,
	}

	// do not sent to ourself and to the accepted user
	skipMembers := make(map[string]struct{})
	skipMembers[common.PubkeyToHex(&m.identity.PublicKey)] = struct{}{}
	skipMembers[common.PubkeyToHex(pk)] = struct{}{}

	subscriptionMsg := &CommunityPrivilegedMemberSyncMessage{
		CommunityPrivateKey: community.PrivateKey(),
	}

	fileredPrivilegedMembers := community.GetFilteredPrivilegedMembers(skipMembers)
	for role, members := range fileredPrivilegedMembers {
		if len(members) == 0 {
			continue
		}

		subscriptionMsg.Receivers = members

		switch role {
		case protobuf.CommunityMember_ROLE_ADMIN:
			subscriptionMsg.CommunityPrivilegedUserSyncMessage = msgWithoutRevealedAccounts
		case protobuf.CommunityMember_ROLE_OWNER:
			fallthrough
		case protobuf.CommunityMember_ROLE_TOKEN_MASTER:
			subscriptionMsg.CommunityPrivilegedUserSyncMessage = msgWithRevealedAccounts
		}

		m.publish(&Subscription{CommunityPrivilegedMemberSyncMessage: subscriptionMsg})
	}

	return nil
}

func (m *Manager) GetCommunityRequestsToJoinWithRevealedAddresses(communityID types.HexBytes) ([]*RequestToJoin, error) {
	return m.persistence.GetCommunityRequestsToJoinWithRevealedAddresses(communityID)
}

func (m *Manager) SaveCommunity(community *Community) error {
	return m.persistence.SaveCommunity(community)
}

func (m *Manager) CreateCommunityTokenDeploymentSignature(ctx context.Context, chainID uint64, addressFrom string, communityID string) ([]byte, error) {
	community, err := m.GetByIDString(communityID)
	if err != nil {
		return nil, err
	}
	if !community.IsControlNode() {
		return nil, ErrNotControlNode
	}
	digest, err := m.communityTokensService.DeploymentSignatureDigest(chainID, addressFrom, communityID)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(digest, community.PrivateKey())
}

func (m *Manager) GetSyncControlNode(id types.HexBytes) (*protobuf.SyncCommunityControlNode, error) {
	return m.persistence.GetSyncControlNode(id)
}

func (m *Manager) SaveSyncControlNode(id types.HexBytes, syncControlNode *protobuf.SyncCommunityControlNode) error {
	return m.persistence.SaveSyncControlNode(id, syncControlNode.Clock, syncControlNode.InstallationId)
}

func (m *Manager) SetSyncControlNode(id types.HexBytes, syncControlNode *protobuf.SyncCommunityControlNode) error {
	existingSyncControlNode, err := m.GetSyncControlNode(id)
	if err != nil {
		return err
	}

	if existingSyncControlNode == nil || existingSyncControlNode.Clock < syncControlNode.Clock {
		return m.SaveSyncControlNode(id, syncControlNode)
	}

	return nil
}

func (m *Manager) GetCommunityRequestToJoinWithRevealedAddresses(pubKey string, communityID types.HexBytes) (*RequestToJoin, error) {
	return m.persistence.GetCommunityRequestToJoinWithRevealedAddresses(pubKey, communityID)
}

func (m *Manager) SafeGetSignerPubKey(chainID uint64, communityID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	return m.ownerVerifier.SafeGetSignerPubKey(ctx, chainID, communityID)
}

func (m *Manager) GetCuratedCommunities() (*CuratedCommunities, error) {
	return m.persistence.GetCuratedCommunities()
}

func (m *Manager) SetCuratedCommunities(communities *CuratedCommunities) error {
	return m.persistence.SetCuratedCommunities(communities)
}

func (m *Manager) encryptCommunityDescriptionImpl(groupID []byte, d *protobuf.CommunityDescription) (string, []byte, error) {
	payload, err := proto.Marshal(d)
	if err != nil {
		return "", nil, err
	}

	encryptedPayload, ratchet, newSeqNo, err := m.encryptor.EncryptWithHashRatchet(groupID, payload)
	if err == encryption.ErrNoEncryptionKey {
		_, err := m.encryptor.GenerateHashRatchetKey(groupID)
		if err != nil {
			return "", nil, err
		}
		encryptedPayload, ratchet, newSeqNo, err = m.encryptor.EncryptWithHashRatchet(groupID, payload)
		if err != nil {
			return "", nil, err
		}

	} else if err != nil {
		return "", nil, err
	}

	keyID, err := ratchet.GetKeyID()
	if err != nil {
		return "", nil, err
	}

	communityJSON, err := json.Marshal(d)
	if err != nil {
		return "", nil, err
	}

	m.logger.Debug("encrypting community description", zap.String("community", string(communityJSON)), zap.String("groupID", types.Bytes2Hex(groupID)), zap.String("key-id", types.Bytes2Hex(keyID)))
	keyIDSeqNo := fmt.Sprintf("%s%d", hex.EncodeToString(keyID), newSeqNo)

	return keyIDSeqNo, encryptedPayload, nil
}

func (m *Manager) encryptCommunityDescription(community *Community, d *protobuf.CommunityDescription) (string, []byte, error) {
	return m.encryptCommunityDescriptionImpl(community.ID(), d)
}

func (m *Manager) encryptCommunityDescriptionChannel(community *Community, channelID string, d *protobuf.CommunityDescription) (string, []byte, error) {
	return m.encryptCommunityDescriptionImpl([]byte(community.IDString()+channelID), d)
}

type DecryptCommunityResponse struct {
	Decrypted   bool
	Description *protobuf.CommunityDescription
	KeyID       []byte
	GroupID     []byte
}

func (m *Manager) decryptCommunityDescription(keyIDSeqNo string, d []byte) (*DecryptCommunityResponse, error) {
	const hashHexLength = 64
	if len(keyIDSeqNo) <= hashHexLength {
		return nil, errors.New("invalid keyIDSeqNo")
	}

	keyID, err := hex.DecodeString(keyIDSeqNo[:hashHexLength])
	if err != nil {
		return nil, err
	}

	seqNo, err := strconv.ParseUint(keyIDSeqNo[hashHexLength:], 10, 32)
	if err != nil {
		return nil, err
	}

	decryptedPayload, err := m.encryptor.DecryptWithHashRatchet(keyID, uint32(seqNo), d)
	if err == encryption.ErrNoRatchetKey {
		return &DecryptCommunityResponse{
			KeyID: keyID,
		}, err

	}
	if err != nil {
		return nil, err
	}

	var description protobuf.CommunityDescription
	err = proto.Unmarshal(decryptedPayload, &description)
	if err != nil {
		return nil, err
	}

	decryptCommunityResponse := &DecryptCommunityResponse{
		Decrypted:   true,
		KeyID:       keyID,
		Description: &description,
	}
	return decryptCommunityResponse, nil
}

func ToLinkPreveiwThumbnail(image images.IdentityImage) (*common.LinkPreviewThumbnail, error) {
	thumbnail := &common.LinkPreviewThumbnail{}

	if image.IsEmpty() {
		return nil, nil
	}

	width, height, err := images.GetImageDimensions(image.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to get image dimensions: %w", err)
	}

	dataURI, err := image.GetDataURI()
	if err != nil {
		return nil, fmt.Errorf("failed to get data uri: %w", err)
	}

	thumbnail.Width = width
	thumbnail.Height = height
	thumbnail.DataURI = dataURI
	return thumbnail, nil
}

func (c *Community) ToStatusLinkPreview() (*common.StatusCommunityLinkPreview, error) {
	communityLinkPreview := &common.StatusCommunityLinkPreview{}
	if image, ok := c.Images()[images.SmallDimName]; ok {
		thumbnail, err := ToLinkPreveiwThumbnail(images.IdentityImage{Payload: image.Payload})
		if err != nil {
			c.config.Logger.Warn("unfurling status link: failed to set community thumbnail", zap.Error(err))
		}
		communityLinkPreview.Icon = *thumbnail
	}

	if image, ok := c.Images()[images.BannerIdentityName]; ok {
		thumbnail, err := ToLinkPreveiwThumbnail(images.IdentityImage{Payload: image.Payload})
		if err != nil {
			c.config.Logger.Warn("unfurling status link: failed to set community thumbnail", zap.Error(err))
		}
		communityLinkPreview.Banner = *thumbnail
	}

	communityLinkPreview.CommunityID = c.IDString()
	communityLinkPreview.DisplayName = c.Name()
	communityLinkPreview.Description = c.DescriptionText()
	communityLinkPreview.MembersCount = uint32(c.MembersCount())
	communityLinkPreview.Color = c.Color()

	return communityLinkPreview, nil
}
