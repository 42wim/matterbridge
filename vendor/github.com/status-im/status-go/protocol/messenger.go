package protocol

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/appmetrics"
	"github.com/status-im/status-go/connection"
	"github.com/status-im/status-go/contracts"
	"github.com/status-im/status-go/deprecation"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	sociallinkssettings "github.com/status-im/status-go/multiaccounts/settings_social_links"
	"github.com/status-im/status-go/protocol/anonmetrics"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/encryption"
	"github.com/status-im/status-go/protocol/encryption/multidevice"
	"github.com/status-im/status-go/protocol/encryption/sharedsecret"
	"github.com/status-im/status-go/protocol/ens"
	"github.com/status-im/status-go/protocol/identity"
	"github.com/status-im/status-go/protocol/identity/alias"
	"github.com/status-im/status-go/protocol/identity/identicon"
	"github.com/status-im/status-go/protocol/peersyncing"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/pushnotificationclient"
	"github.com/status-im/status-go/protocol/pushnotificationserver"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/protocol/sqlite"
	"github.com/status-im/status-go/protocol/storenodes"
	"github.com/status-im/status-go/protocol/transport"
	v1protocol "github.com/status-im/status-go/protocol/v1"
	"github.com/status-im/status-go/protocol/verification"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/services/browsers"
	"github.com/status-im/status-go/services/communitytokens"
	ensservice "github.com/status-im/status-go/services/ens"
	"github.com/status-im/status-go/services/ext/mailservers"
	localnotifications "github.com/status-im/status-go/services/local-notifications"
	mailserversDB "github.com/status-im/status-go/services/mailservers"
	"github.com/status-im/status-go/services/wallet"
	"github.com/status-im/status-go/services/wallet/community"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/signal"
	"github.com/status-im/status-go/telemetry"
)

const (
	PubKeyStringLength = 132

	transactionSentTxt = "Transaction sent"

	publicChat  ChatContext = "public-chat"
	privateChat ChatContext = "private-chat"
)

var communityAdvertiseIntervalSecond int64 = 60 * 60

// messageCacheIntervalMs is how long we should keep processed messages in the cache, in ms
var messageCacheIntervalMs uint64 = 1000 * 60 * 60 * 48

// Messenger is a entity managing chats and messages.
// It acts as a bridge between the application and encryption
// layers.
// It needs to expose an interface to manage installations
// because installations are managed by the user.
// Similarly, it needs to expose an interface to manage
// mailservers because they can also be managed by the user.
type Messenger struct {
	node                      types.Node
	server                    *p2p.Server
	peerStore                 *mailservers.PeerStore
	config                    *config
	identity                  *ecdsa.PrivateKey
	persistence               *sqlitePersistence
	transport                 *transport.Transport
	encryptor                 *encryption.Protocol
	sender                    *common.MessageSender
	ensVerifier               *ens.Verifier
	anonMetricsClient         *anonmetrics.Client
	anonMetricsServer         *anonmetrics.Server
	pushNotificationClient    *pushnotificationclient.Client
	pushNotificationServer    *pushnotificationserver.Server
	communitiesManager        *communities.Manager
	communitiesKeyDistributor communities.KeyDistributor
	accountsManager           account.Manager
	mentionsManager           *MentionManager
	storeNodeRequestsManager  *StoreNodeRequestManager
	logger                    *zap.Logger

	outputCSV bool
	csvFile   *os.File

	verifyTransactionClient    EthClient
	featureFlags               common.FeatureFlags
	shutdownTasks              []func() error
	shouldPublishContactCode   bool
	systemMessagesTranslations *systemMessageTranslationsMap
	allChats                   *chatMap
	selfContact                *Contact
	selfContactSubscriptions   []chan *SelfContactChangeEvent
	allContacts                *contactMap
	allInstallations           *installationMap
	modifiedInstallations      *stringBoolMap
	installationID             string
	mailserverCycle            mailserverCycle
	communityStorenodes        *storenodes.CommunityStorenodes
	database                   *sql.DB
	multiAccounts              *multiaccounts.Database
	settings                   *accounts.Database
	account                    *multiaccounts.Account
	mailserversDatabase        *mailserversDB.Database
	browserDatabase            *browsers.Database
	httpServer                 *server.MediaServer

	started           bool
	quit              chan struct{}
	ctx               context.Context
	cancel            context.CancelFunc
	shutdownWaitGroup sync.WaitGroup

	importingCommunities map[string]bool
	importingChannels    map[string]bool
	importRateLimiter    *rate.Limiter
	importDelayer        struct {
		wait chan struct{}
		once sync.Once
	}

	connectionState       connection.State
	telemetryClient       *telemetry.Client
	contractMaker         *contracts.ContractMaker
	verificationDatabase  *verification.Persistence
	savedAddressesManager *wallet.SavedAddressesManager
	walletAPI             *wallet.API

	// TODO(samyoul) Determine if/how the remaining usage of this mutex can be removed
	mutex                     sync.Mutex
	mailPeersMutex            sync.RWMutex
	handleMessagesMutex       sync.Mutex
	handleImportMessagesMutex sync.Mutex

	// flag to disable checking #hasPairedDevices
	localPairing bool
	// flag to enable backedup messages processing, false by default
	processBackedupMessages bool

	communityTokensService communitytokens.ServiceInterface

	// used to track dispatched messages
	dispatchMessageTestCallback func(common.RawMessage)

	// used to track unhandled messages
	unhandledMessagesTracker func(*v1protocol.StatusMessage, error)

	// enables control over chat messages iteration
	retrievedMessagesIteratorFactory func(map[transport.Filter][]*types.Message) MessagesIterator

	peersyncing         *peersyncing.PeerSyncing
	peersyncingOffers   map[string]uint64
	peersyncingRequests map[string]uint64
}

type connStatus int

const (
	disconnected connStatus = iota + 1
	connecting
	connected
)

type peerStatus struct {
	status                connStatus
	canConnectAfter       time.Time
	lastConnectionAttempt time.Time
	mailserver            mailserversDB.Mailserver
}
type mailserverCycle struct {
	sync.RWMutex
	allMailservers            []mailserversDB.Mailserver
	activeMailserver          *mailserversDB.Mailserver
	peers                     map[string]peerStatus
	events                    chan *p2p.PeerEvent
	subscription              event.Subscription
	availabilitySubscriptions []chan struct{}
}

type EnvelopeEventsInterceptor struct {
	EnvelopeEventsHandler transport.EnvelopeEventsHandler
	Messenger             *Messenger
}

func (m *Messenger) GetOwnPrimaryName() (string, error) {
	ensName, err := m.settings.ENSName()
	if err != nil {
		return ensName, nil
	}
	return m.settings.DisplayName()
}

func (m *Messenger) ResolvePrimaryName(mentionID string) (string, error) {
	if mentionID == m.myHexIdentity() {
		return m.GetOwnPrimaryName()
	}
	contact, ok := m.allContacts.Load(mentionID)
	if !ok {
		var err error
		contact, err = buildContactFromPkString(mentionID)
		if err != nil {
			return mentionID, err
		}
	}
	return contact.PrimaryName(), nil
}

// EnvelopeSent triggered when envelope delivered at least to 1 peer.
func (interceptor EnvelopeEventsInterceptor) EnvelopeSent(identifiers [][]byte) {
	if interceptor.Messenger != nil {
		var ids []string
		for _, identifierBytes := range identifiers {
			ids = append(ids, types.EncodeHex(identifierBytes))
		}

		err := interceptor.Messenger.processSentMessages(ids)
		if err != nil {
			interceptor.Messenger.logger.Info("messenger failed to process sent messages", zap.Error(err))
		}

		// We notify the client, regardless whether we were able to mark them as sent
		interceptor.EnvelopeEventsHandler.EnvelopeSent(identifiers)
	} else {
		// NOTE(rasom): In case if interceptor.Messenger is not nil and
		// some error occurred on processing sent message we don't want
		// to send envelop.sent signal to the client, thus `else` cause
		// is necessary.
		interceptor.EnvelopeEventsHandler.EnvelopeSent(identifiers)
	}
}

// EnvelopeExpired triggered when envelope is expired but wasn't delivered to any peer.
func (interceptor EnvelopeEventsInterceptor) EnvelopeExpired(identifiers [][]byte, err error) {
	//we don't track expired events in Messenger, so just redirect to handler
	interceptor.EnvelopeEventsHandler.EnvelopeExpired(identifiers, err)
}

// MailServerRequestCompleted triggered when the mailserver sends a message to notify that the request has been completed
func (interceptor EnvelopeEventsInterceptor) MailServerRequestCompleted(requestID types.Hash, lastEnvelopeHash types.Hash, cursor []byte, err error) {
	//we don't track mailserver requests in Messenger, so just redirect to handler
	interceptor.EnvelopeEventsHandler.MailServerRequestCompleted(requestID, lastEnvelopeHash, cursor, err)
}

// MailServerRequestExpired triggered when the mailserver request expires
func (interceptor EnvelopeEventsInterceptor) MailServerRequestExpired(hash types.Hash) {
	//we don't track mailserver requests in Messenger, so just redirect to handler
	interceptor.EnvelopeEventsHandler.MailServerRequestExpired(hash)
}

func NewMessenger(
	nodeName string,
	identity *ecdsa.PrivateKey,
	node types.Node,
	installationID string,
	peerStore *mailservers.PeerStore,
	opts ...Option,
) (*Messenger, error) {
	var messenger *Messenger

	c := messengerDefaultConfig()

	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return nil, err
		}
	}

	logger := c.logger
	if c.logger == nil {
		var err error
		if logger, err = zap.NewDevelopment(); err != nil {
			return nil, errors.Wrap(err, "failed to create a logger")
		}
	}

	if c.systemMessagesTranslations == nil {
		c.systemMessagesTranslations = defaultSystemMessagesTranslations
	}

	// Configure the database.
	if c.appDb == nil {
		return nil, errors.New("database instance or database path needs to be provided")
	}
	database := c.appDb

	// Apply any post database creation changes to the database
	for _, opt := range c.afterDbCreatedHooks {
		if err := opt(&c); err != nil {
			return nil, err
		}
	}

	// Apply migrations for all components.
	err := sqlite.Migrate(database)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply migrations")
	}

	// Initialize transport layer.
	var transp *transport.Transport

	if waku, err := node.GetWaku(nil); err == nil && waku != nil {
		transp, err = transport.NewTransport(
			waku,
			identity,
			database,
			"waku_keys",
			nil,
			c.envelopesMonitorConfig,
			logger,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create  Transport")
		}
	} else {
		logger.Info("failed to find Waku service; trying WakuV2", zap.Error(err))
		wakuV2, err := node.GetWakuV2(nil)
		if err != nil || wakuV2 == nil {
			return nil, errors.Wrap(err, "failed to find Whisper and Waku V1/V2 services")
		}
		transp, err = transport.NewTransport(
			wakuV2,
			identity,
			database,
			"wakuv2_keys",
			nil,
			c.envelopesMonitorConfig,
			logger,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create  Transport")
		}
	}

	// Initialize encryption layer.
	encryptionProtocol := encryption.New(
		database,
		installationID,
		logger,
	)

	sender, err := common.NewMessageSender(
		identity,
		database,
		encryptionProtocol,
		transp,
		logger,
		c.featureFlags,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create messageSender")
	}

	// Initialise anon metrics client
	var anonMetricsClient *anonmetrics.Client
	if c.anonMetricsClientConfig != nil &&
		c.anonMetricsClientConfig.ShouldSend &&
		c.anonMetricsClientConfig.Active == anonmetrics.ActiveClientPhrase {

		anonMetricsClient = anonmetrics.NewClient(sender)
		anonMetricsClient.Config = c.anonMetricsClientConfig
		anonMetricsClient.Identity = identity
		anonMetricsClient.DB = appmetrics.NewDB(database)
		anonMetricsClient.Logger = logger
	}

	// Initialise anon metrics server
	var anonMetricsServer *anonmetrics.Server
	if c.anonMetricsServerConfig != nil &&
		c.anonMetricsServerConfig.Enabled &&
		c.anonMetricsServerConfig.Active == anonmetrics.ActiveServerPhrase {

		server, err := anonmetrics.NewServer(c.anonMetricsServerConfig.PostgresURI)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create anonmetrics.Server")
		}

		anonMetricsServer = server
		anonMetricsServer.Config = c.anonMetricsServerConfig
		anonMetricsServer.Logger = logger
	}

	var telemetryClient *telemetry.Client
	if c.telemetryServerURL != "" {
		telemetryClient = telemetry.NewClient(logger, c.telemetryServerURL, c.account.KeyUID, nodeName)
		if c.wakuService != nil {
			c.wakuService.SetStatusTelemetryClient(telemetryClient)
		}
	}

	// Initialize push notification server
	var pushNotificationServer *pushnotificationserver.Server
	if c.pushNotificationServerConfig != nil && c.pushNotificationServerConfig.Enabled {
		c.pushNotificationServerConfig.Identity = identity
		pushNotificationServerPersistence := pushnotificationserver.NewSQLitePersistence(database)
		pushNotificationServer = pushnotificationserver.New(c.pushNotificationServerConfig, pushNotificationServerPersistence, sender)
	}

	// Initialize push notification client
	pushNotificationClientPersistence := pushnotificationclient.NewPersistence(database)
	pushNotificationClientConfig := c.pushNotificationClientConfig
	if pushNotificationClientConfig == nil {
		pushNotificationClientConfig = &pushnotificationclient.Config{}
	}

	sqlitePersistence := newSQLitePersistence(database)
	// Overriding until we handle different identities
	pushNotificationClientConfig.Identity = identity
	pushNotificationClientConfig.Logger = logger
	pushNotificationClientConfig.InstallationID = installationID

	pushNotificationClient := pushnotificationclient.New(pushNotificationClientPersistence, pushNotificationClientConfig, sender, sqlitePersistence)

	ensVerifier := ens.New(node, logger, transp, database, c.verifyENSURL, c.verifyENSContractAddress)

	managerOptions := []communities.ManagerOption{
		communities.WithAccountManager(c.accountsManager),
	}

	var walletAPI *wallet.API
	if c.walletService != nil {
		walletAPI = wallet.NewAPI(c.walletService)
		managerOptions = append(managerOptions, communities.WithCollectiblesManager(walletAPI))
	} else if c.collectiblesManager != nil {
		managerOptions = append(managerOptions, communities.WithCollectiblesManager(c.collectiblesManager))
	}

	if c.tokenManager != nil {
		managerOptions = append(managerOptions, communities.WithTokenManager(c.tokenManager))
	} else if c.rpcClient != nil {
		tokenManager := token.NewTokenManager(c.walletDb, c.rpcClient, community.NewManager(database, c.httpServer, nil), c.rpcClient.NetworkManager, database, c.httpServer, nil)
		managerOptions = append(managerOptions, communities.WithTokenManager(communities.NewDefaultTokenManager(tokenManager)))
	}

	if c.walletConfig != nil {
		managerOptions = append(managerOptions, communities.WithWalletConfig(c.walletConfig))
	}

	if c.communityTokensService != nil {
		managerOptions = append(managerOptions, communities.WithCommunityTokensService(c.communityTokensService))
	}

	communitiesKeyDistributor := &CommunitiesKeyDistributorImpl{
		sender:    sender,
		encryptor: encryptionProtocol,
	}

	communitiesManager, err := communities.NewManager(identity, installationID, database, encryptionProtocol, logger, ensVerifier, c.communityTokensService, transp, transp, communitiesKeyDistributor, c.torrentConfig, managerOptions...)
	if err != nil {
		return nil, err
	}

	settings, err := accounts.NewDB(database)
	if err != nil {
		return nil, err
	}

	savedAddressesManager := wallet.NewSavedAddressesManager(c.walletDb)

	selfContact, err := buildSelfContact(identity, settings, c.multiAccount, c.account)
	if err != nil {
		return nil, fmt.Errorf("failed to build contact of ourself: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	messenger = &Messenger{
		config:                     &c,
		node:                       node,
		identity:                   identity,
		persistence:                sqlitePersistence,
		transport:                  transp,
		encryptor:                  encryptionProtocol,
		sender:                     sender,
		anonMetricsClient:          anonMetricsClient,
		anonMetricsServer:          anonMetricsServer,
		telemetryClient:            telemetryClient,
		communityTokensService:     c.communityTokensService,
		pushNotificationClient:     pushNotificationClient,
		pushNotificationServer:     pushNotificationServer,
		communitiesManager:         communitiesManager,
		communitiesKeyDistributor:  communitiesKeyDistributor,
		accountsManager:            c.accountsManager,
		ensVerifier:                ensVerifier,
		featureFlags:               c.featureFlags,
		systemMessagesTranslations: c.systemMessagesTranslations,
		allChats:                   new(chatMap),
		selfContact:                selfContact,
		allContacts: &contactMap{
			logger: logger,
			me:     selfContact,
		},
		allInstallations:        new(installationMap),
		installationID:          installationID,
		modifiedInstallations:   new(stringBoolMap),
		verifyTransactionClient: c.verifyTransactionClient,
		database:                database,
		multiAccounts:           c.multiAccount,
		settings:                settings,
		peersyncing:             peersyncing.New(peersyncing.Config{Database: database, Timesource: transp}),
		peersyncingOffers:       make(map[string]uint64),
		peersyncingRequests:     make(map[string]uint64),
		peerStore:               peerStore,
		verificationDatabase:    verification.NewPersistence(database),
		mailserverCycle: mailserverCycle{
			peers:                     make(map[string]peerStatus),
			availabilitySubscriptions: make([]chan struct{}, 0),
		},
		mailserversDatabase:  c.mailserversDatabase,
		communityStorenodes:  storenodes.NewCommunityStorenodes(storenodes.NewDB(database), logger),
		account:              c.account,
		quit:                 make(chan struct{}),
		ctx:                  ctx,
		cancel:               cancel,
		importingCommunities: make(map[string]bool),
		importingChannels:    make(map[string]bool),
		importRateLimiter:    rate.NewLimiter(rate.Every(importSlowRate), 1),
		importDelayer: struct {
			wait chan struct{}
			once sync.Once
		}{wait: make(chan struct{})},
		browserDatabase: c.browserDatabase,
		httpServer:      c.httpServer,
		shutdownTasks: []func() error{
			ensVerifier.Stop,
			pushNotificationClient.Stop,
			communitiesManager.Stop,
			encryptionProtocol.Stop,
			func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := transp.ResetFilters(ctx)
				if err != nil {
					logger.Warn("could not reset filters", zap.Error(err))
				}
				// We don't want to thrown an error in this case, this is a soft
				// fail
				return nil
			},
			transp.Stop,
			func() error { sender.Stop(); return nil },
			// Currently this often fails, seems like it's safe to ignore them
			// https://github.com/uber-go/zap/issues/328
			func() error { _ = logger.Sync; return nil },
			database.Close,
		},
		logger:                           logger,
		savedAddressesManager:            savedAddressesManager,
		retrievedMessagesIteratorFactory: NewDefaultMessagesIterator,
	}

	if c.rpcClient != nil {
		contractMaker, err := contracts.NewContractMaker(c.rpcClient)
		if err != nil {
			return nil, err
		}
		messenger.contractMaker = contractMaker
	}

	messenger.mentionsManager = NewMentionManager(messenger)
	messenger.storeNodeRequestsManager = NewStoreNodeRequestManager(messenger)

	if c.walletService != nil {
		messenger.walletAPI = walletAPI
	}

	if c.outputMessagesCSV {
		messenger.outputCSV = c.outputMessagesCSV
		csvFile, err := os.Create("messages-" + fmt.Sprint(time.Now().Unix()) + ".csv")
		if err != nil {
			return nil, err
		}

		_, err = csvFile.Write([]byte("timestamp\tmessageID\tfrom\ttopic\tchatID\tmessageType\tmessage\n"))
		if err != nil {
			return nil, err
		}

		messenger.csvFile = csvFile
		messenger.shutdownTasks = append(messenger.shutdownTasks, csvFile.Close)
	}

	if anonMetricsClient != nil {
		messenger.shutdownTasks = append(messenger.shutdownTasks, anonMetricsClient.Stop)
	}
	if anonMetricsServer != nil {
		messenger.shutdownTasks = append(messenger.shutdownTasks, anonMetricsServer.Stop)
	}

	if c.envelopesMonitorConfig != nil {
		interceptor := EnvelopeEventsInterceptor{c.envelopesMonitorConfig.EnvelopeEventsHandler, messenger}
		err := messenger.transport.SetEnvelopeEventsHandler(interceptor)
		if err != nil {
			logger.Info("Unable to set envelopes event handler", zap.Error(err))
		}
	}

	return messenger, nil
}

func (m *Messenger) SetP2PServer(server *p2p.Server) {
	m.server = server
}

func (m *Messenger) EnableBackedupMessagesProcessing() {
	m.processBackedupMessages = true
}

func (m *Messenger) processSentMessages(ids []string) error {
	if m.connectionState.Offline {
		return errors.New("Can't mark message as sent while offline")
	}

	for _, id := range ids {
		rawMessage, err := m.persistence.RawMessageByID(id)
		// If we have no raw message, we create a temporary one, so that
		// the sent status is preserved
		if err == sql.ErrNoRows || rawMessage == nil {
			rawMessage = &common.RawMessage{
				ID:          id,
				MessageType: protobuf.ApplicationMetadataMessage_CHAT_MESSAGE,
			}
		} else if err != nil {
			return errors.Wrapf(err, "Can't get raw message with id %v", id)
		}

		rawMessage.Sent = true

		err = m.persistence.SaveRawMessage(rawMessage)
		if err != nil {
			return errors.Wrapf(err, "Can't save raw message marked as sent")
		}

		err = m.UpdateMessageOutgoingStatus(id, common.OutgoingStatusSent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Messenger) shouldResendMessage(message *common.RawMessage, t common.TimeSource) (bool, error) {
	if m.featureFlags.ResendRawMessagesDisabled {
		return false, nil
	}
	//exponential backoff depends on how many attempts to send message already made
	power := math.Pow(2, float64(message.SendCount-1))
	backoff := uint64(power) * uint64(m.config.messageResendMinDelay.Milliseconds())
	backoffElapsed := t.GetCurrentTime() > (message.LastSent + backoff)
	return backoffElapsed, nil
}

func (m *Messenger) resendExpiredMessages() error {
	if m.connectionState.Offline {
		return errors.New("offline")
	}

	ids, err := m.persistence.ExpiredMessagesIDs(m.config.messageResendMaxCount)
	if err != nil {
		return errors.Wrapf(err, "Can't get expired reactions from db")
	}

	for _, id := range ids {
		rawMessage, err := m.persistence.RawMessageByID(id)
		if err != nil {
			return errors.Wrapf(err, "Can't get raw message with id %v", id)
		}

		chat, ok := m.allChats.Load(rawMessage.LocalChatID)
		if !ok {
			return ErrChatNotFound
		}

		if !(chat.Public() || chat.CommunityChat()) {
			return errors.New("Only public chats and community chats messages are resent")
		}

		ok, err = m.shouldResendMessage(rawMessage, m.getTimesource())
		if err != nil {
			return err
		}

		if ok {
			err = m.persistence.SaveRawMessage(rawMessage)
			if err != nil {
				return errors.Wrapf(err, "Can't save raw message marked as non-expired")
			}

			err = m.reSendRawMessage(context.Background(), rawMessage.ID)
			if err != nil {
				return errors.Wrapf(err, "Can't resend expired message with id %v", rawMessage.ID)
			}
		}
	}
	return nil
}

func (m *Messenger) ToForeground() {
	if m.httpServer != nil {
		m.httpServer.ToForeground()
	}
}

func (m *Messenger) ToBackground() {
	if m.httpServer != nil {
		m.httpServer.ToBackground()
	}
}

func (m *Messenger) Start() (*MessengerResponse, error) {
	if m.started {
		return nil, errors.New("messenger already started")
	}
	m.started = true

	now := time.Now().UnixMilli()
	if err := m.settings.CheckAndDeleteExpiredKeypairsAndAccounts(uint64(now)); err != nil {
		return nil, err
	}

	m.logger.Info("starting messenger", zap.String("identity", types.EncodeHex(crypto.FromECDSAPub(&m.identity.PublicKey))))
	// Start push notification server
	if m.pushNotificationServer != nil {
		if err := m.pushNotificationServer.Start(); err != nil {
			return nil, err
		}
	}

	// Start push notification client
	if m.pushNotificationClient != nil {
		m.handlePushNotificationClientRegistrations(m.pushNotificationClient.SubscribeToRegistrations())

		if err := m.pushNotificationClient.Start(); err != nil {
			return nil, err
		}
	}

	// Start anonymous metrics client
	if m.anonMetricsClient != nil {
		if err := m.anonMetricsClient.Start(); err != nil {
			return nil, err
		}
	}

	ensSubscription := m.ensVerifier.Subscribe()

	// Subscrbe
	if err := m.ensVerifier.Start(); err != nil {
		return nil, err
	}

	if err := m.communitiesManager.Start(); err != nil {
		return nil, err
	}

	// set shared secret handles
	m.sender.SetHandleSharedSecrets(m.handleSharedSecrets)
	if err := m.sender.StartDatasync(m.sendDataSync); err != nil {
		return nil, err
	}

	subscriptions, err := m.encryptor.Start(m.identity)
	if err != nil {
		return nil, err
	}

	// handle stored shared secrets
	err = m.handleSharedSecrets(subscriptions.SharedSecrets)
	if err != nil {
		return nil, err
	}

	m.handleEncryptionLayerSubscriptions(subscriptions)
	m.handleCommunitiesSubscription(m.communitiesManager.Subscribe())
	m.handleCommunitiesHistoryArchivesSubscription(m.communitiesManager.Subscribe())
	m.updateCommunitiesActiveMembersPeriodically()
	m.handleENSVerificationSubscription(ensSubscription)
	m.watchConnectionChange()
	m.watchChatsAndCommunitiesToUnmute()
	m.watchCommunitiesToUnmute()
	m.watchExpiredMessages()
	m.watchIdentityImageChanges()
	m.watchWalletBalances()
	m.watchPendingCommunityRequestToJoin()
	m.broadcastLatestUserStatus()
	m.timeoutAutomaticStatusUpdates()
	if !m.config.featureFlags.DisableCheckingForBackup {
		m.startBackupLoop()
	}
	if !m.config.featureFlags.DisableAutoMessageLoop {
		err = m.startAutoMessageLoop()
		if err != nil {
			return nil, err
		}
	}
	m.startPeerSyncingLoop()
	m.startSyncSettingsLoop()
	m.startSettingsChangesLoop()
	m.startCommunityRekeyLoop()
	if m.config.codeControlFlags.CuratedCommunitiesUpdateLoopEnabled {
		m.startCuratedCommunitiesUpdateLoop()
	}
	m.startMessageSegmentsCleanupLoop()

	if err := m.cleanTopics(); err != nil {
		return nil, err
	}
	response := &MessengerResponse{}

	mailservers, err := m.allMailservers()
	if err != nil {
		return nil, err
	}

	response.Mailservers = mailservers
	err = m.StartMailserverCycle(mailservers)
	if err != nil {
		return nil, err
	}

	if err := m.communityStorenodes.ReloadFromDB(); err != nil {
		return nil, err
	}

	controlledCommunities, err := m.communitiesManager.Controlled()
	if err != nil {
		return nil, err
	}

	if m.torrentClientReady() {
		available := m.SubscribeMailserverAvailable()
		go func() {
			<-available
			m.InitHistoryArchiveTasks(controlledCommunities)
		}()
	}

	for _, c := range controlledCommunities {
		if c.Joined() && c.HasTokenPermissions() {
			go m.communitiesManager.ReevaluateMembersPeriodically(c.ID())
		}
	}

	joinedCommunities, err := m.communitiesManager.Joined()
	if err != nil {
		return nil, err
	}

	for _, joinedCommunity := range joinedCommunities {
		// resume importing message history archives in case
		// imports have been interrupted previously
		err := m.resumeHistoryArchivesImport(joinedCommunity.ID())
		if err != nil {
			return nil, err
		}
	}
	m.enableHistoryArchivesImportAfterDelay()

	if m.httpServer != nil {
		err = m.httpServer.Start()
		if err != nil {
			return nil, err
		}
	}

	err = m.GarbageCollectRemovedBookmarks()
	if err != nil {
		return nil, err
	}

	err = m.garbageCollectRemovedSavedAddresses()
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (m *Messenger) IdentityPublicKey() *ecdsa.PublicKey {
	return &m.identity.PublicKey
}

func (m *Messenger) IdentityPublicKeyCompressed() []byte {
	return crypto.CompressPubkey(m.IdentityPublicKey())
}

func (m *Messenger) IdentityPublicKeyString() string {
	return types.EncodeHex(crypto.FromECDSAPub(m.IdentityPublicKey()))
}

// cleanTopics remove any topic that does not have a Listen flag set
func (m *Messenger) cleanTopics() error {
	if m.mailserversDatabase == nil {
		return nil
	}
	var filters []*transport.Filter
	for _, f := range m.transport.Filters() {
		if f.Listen && !f.Ephemeral {
			filters = append(filters, f)
		}
	}

	m.logger.Debug("keeping topics", zap.Any("filters", filters))

	return m.mailserversDatabase.SetTopics(filters)
}

// handle connection change is called each time we go from offline/online or viceversa
func (m *Messenger) handleConnectionChange(online bool) {
	// Update pushNotificationClient
	if m.pushNotificationClient != nil {
		if online {
			m.pushNotificationClient.Online()
		} else {
			m.pushNotificationClient.Offline()
		}
	}

	// Publish contact code
	if online && m.shouldPublishContactCode {
		if err := m.publishContactCode(); err != nil {
			m.logger.Error("could not publish on contact code", zap.Error(err))
		}
		m.shouldPublishContactCode = false
	}

	// Start fetching messages from store nodes
	if online && m.config.codeControlFlags.AutoRequestHistoricMessages {
		m.asyncRequestAllHistoricMessages()
	}

	// Update ENS verifier
	m.ensVerifier.SetOnline(online)
}

func (m *Messenger) Online() bool {
	switch m.transport.WakuVersion() {
	case 2:
		return m.transport.PeerCount() > 0
	default:
		return m.node.PeersCount() > 0
	}
}

func (m *Messenger) buildContactCodeAdvertisement() (*protobuf.ContactCodeAdvertisement, error) {
	if m.pushNotificationClient == nil || !m.pushNotificationClient.Enabled() {
		return nil, nil
	}
	m.logger.Debug("adding push notification info to contact code bundle")
	info, err := m.pushNotificationClient.MyPushNotificationQueryInfo()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, nil
	}
	return &protobuf.ContactCodeAdvertisement{
		PushNotificationInfo: info,
	}, nil
}

// publishContactCode sends a public message wrapped in the encryption
// layer, which will propagate our bundle
func (m *Messenger) publishContactCode() error {
	var payload []byte
	m.logger.Debug("sending contact code")
	contactCodeAdvertisement, err := m.buildContactCodeAdvertisement()
	if err != nil {
		m.logger.Error("could not build contact code advertisement", zap.Error(err))
	}

	if contactCodeAdvertisement == nil {
		contactCodeAdvertisement = &protobuf.ContactCodeAdvertisement{}
	}

	err = m.attachChatIdentity(contactCodeAdvertisement)
	if err != nil {
		return err
	}

	if contactCodeAdvertisement.ChatIdentity != nil {
		m.logger.Debug("attached chat identity", zap.Int("images len", len(contactCodeAdvertisement.ChatIdentity.Images)))
	} else {
		m.logger.Debug("no attached chat identity")
	}

	payload, err = proto.Marshal(contactCodeAdvertisement)
	if err != nil {
		return err
	}

	contactCodeTopic := transport.ContactCodeTopic(&m.identity.PublicKey)
	rawMessage := common.RawMessage{
		LocalChatID: contactCodeTopic,
		MessageType: protobuf.ApplicationMetadataMessage_CONTACT_CODE_ADVERTISEMENT,
		Payload:     payload,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = m.sender.SendPublic(ctx, contactCodeTopic, rawMessage)
	if err != nil {
		m.logger.Warn("failed to send a contact code", zap.Error(err))
	}

	joinedCommunities, err := m.communitiesManager.Joined()
	if err != nil {
		return err
	}
	for _, community := range joinedCommunities {
		rawMessage.LocalChatID = community.MemberUpdateChannelID()
		rawMessage.PubsubTopic = community.PubsubTopic()
		_, err = m.sender.SendPublic(ctx, rawMessage.LocalChatID, rawMessage)
		if err != nil {
			return err
		}
	}

	m.logger.Debug("contact code sent")
	return err
}

// contactCodeAdvertisement attaches a protobuf.ChatIdentity to the given protobuf.ContactCodeAdvertisement,
// if the `shouldPublish` conditions are met
func (m *Messenger) attachChatIdentity(cca *protobuf.ContactCodeAdvertisement) error {
	contactCodeTopic := transport.ContactCodeTopic(&m.identity.PublicKey)
	shouldPublish, err := m.shouldPublishChatIdentity(contactCodeTopic)
	if err != nil {
		return err
	}

	if !shouldPublish {
		return nil
	}

	cca.ChatIdentity, err = m.createChatIdentity(privateChat)
	if err != nil {
		return err
	}

	img, err := m.multiAccounts.GetIdentityImage(m.account.KeyUID, images.SmallDimName)
	if err != nil {
		return err
	}

	displayName, err := m.settings.DisplayName()
	if err != nil {
		return err
	}

	bio, err := m.settings.Bio()
	if err != nil {
		return err
	}

	socialLinks, err := m.settings.GetSocialLinks()
	if err != nil {
		return err
	}

	profileShowcase, err := m.GetProfileShowcaseForSelfIdentity()
	if err != nil {
		return err
	}

	identityHash, err := m.getIdentityHash(displayName, bio, img, socialLinks, profileShowcase)
	if err != nil {
		return err
	}

	err = m.persistence.SaveWhenChatIdentityLastPublished(contactCodeTopic, identityHash)
	if err != nil {
		return err
	}

	return nil
}

// handleStandaloneChatIdentity sends a standalone ChatIdentity message to a public or private channel if the publish criteria is met
func (m *Messenger) handleStandaloneChatIdentity(chat *Chat) error {
	if chat.ChatType != ChatTypePublic && chat.ChatType != ChatTypeOneToOne {
		return nil
	}
	shouldPublishChatIdentity, err := m.shouldPublishChatIdentity(chat.ID)
	if err != nil {
		return err
	}
	if !shouldPublishChatIdentity {
		return nil
	}

	chatContext := GetChatContextFromChatType(chat.ChatType)

	ci, err := m.createChatIdentity(chatContext)
	if err != nil {
		return err
	}

	payload, err := proto.Marshal(ci)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID: chat.ID,
		MessageType: protobuf.ApplicationMetadataMessage_CHAT_IDENTITY,
		Payload:     payload,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if chat.ChatType == ChatTypePublic {
		_, err = m.sender.SendPublic(ctx, chat.ID, rawMessage)
		if err != nil {
			return err
		}
	} else {
		pk, err := chat.PublicKey()
		if err != nil {
			return err
		}
		_, err = m.sender.SendPrivate(ctx, pk, &rawMessage)
		if err != nil {
			return err
		}

	}

	img, err := m.multiAccounts.GetIdentityImage(m.account.KeyUID, images.SmallDimName)
	if err != nil {
		return err
	}

	displayName, err := m.settings.DisplayName()
	if err != nil {
		return err
	}

	bio, err := m.settings.Bio()
	if err != nil {
		return err
	}

	socialLinks, err := m.settings.GetSocialLinks()
	if err != nil {
		return err
	}

	profileShowcase, err := m.GetProfileShowcaseForSelfIdentity()
	if err != nil {
		return err
	}

	identityHash, err := m.getIdentityHash(displayName, bio, img, socialLinks, profileShowcase)
	if err != nil {
		return err
	}

	err = m.persistence.SaveWhenChatIdentityLastPublished(chat.ID, identityHash)
	if err != nil {
		return err
	}

	return nil
}

func (m *Messenger) getIdentityHash(displayName, bio string, img *images.IdentityImage, socialLinks identity.SocialLinks, profileShowcase *protobuf.ProfileShowcase) ([]byte, error) {
	socialLinksData, err := socialLinks.Serialize()
	if err != nil {
		return []byte{}, err
	}

	profileShowcaseData, err := proto.Marshal(profileShowcase)
	if err != nil {
		return []byte{}, err
	}

	if img == nil {
		return crypto.Keccak256([]byte(displayName), []byte(bio), socialLinksData, profileShowcaseData), nil
	}

	return crypto.Keccak256(img.Payload, []byte(displayName), []byte(bio), socialLinksData, profileShowcaseData), nil
}

// shouldPublishChatIdentity returns true if the last time the ChatIdentity was attached was more than 24 hours ago
func (m *Messenger) shouldPublishChatIdentity(chatID string) (bool, error) {
	if m.account == nil {
		return false, nil
	}

	// Check we have at least one image or a display name
	img, err := m.multiAccounts.GetIdentityImage(m.account.KeyUID, images.SmallDimName)
	if err != nil {
		return false, err
	}

	displayName, err := m.settings.DisplayName()
	if err != nil {
		return false, err
	}

	if img == nil && displayName == "" {
		return false, nil
	}

	lp, hash, err := m.persistence.GetWhenChatIdentityLastPublished(chatID)
	if err != nil {
		return false, err
	}

	bio, err := m.settings.Bio()
	if err != nil {
		return false, err
	}

	socialLinks, err := m.settings.GetSocialLinks()
	if err != nil {
		return false, err
	}

	profileShowcase, err := m.GetProfileShowcaseForSelfIdentity()
	if err != nil {
		return false, err
	}

	identityHash, err := m.getIdentityHash(displayName, bio, img, socialLinks, profileShowcase)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(hash, identityHash) {
		return true, nil
	}

	// Note: If Alice does not add bob as a contact she will not update her contact code with images
	return lp == 0 || time.Now().Unix()-lp > 24*60*60, nil
}

// createChatIdentity creates a context based protobuf.ChatIdentity.
// context 'public-chat' will attach only the 'thumbnail' IdentityImage
// context 'private-chat' will attach all IdentityImage
func (m *Messenger) createChatIdentity(context ChatContext) (*protobuf.ChatIdentity, error) {
	m.logger.Info(fmt.Sprintf("account keyUID '%s'", m.account.KeyUID))
	m.logger.Info(fmt.Sprintf("context '%s'", context))

	displayName, err := m.settings.DisplayName()
	if err != nil {
		return nil, err
	}

	bio, err := m.settings.Bio()
	if err != nil {
		return nil, err
	}

	socialLinks, err := m.settings.GetSocialLinks()
	if err != nil {
		return nil, err
	}

	profileShowcase, err := m.GetProfileShowcaseForSelfIdentity()
	if err != nil {
		return nil, err
	}

	ci := &protobuf.ChatIdentity{
		Clock:           m.transport.GetCurrentTime(),
		EnsName:         "", // TODO add ENS name handling to dedicate PR
		DisplayName:     displayName,
		Description:     bio,
		SocialLinks:     socialLinks.ToProtobuf(),
		ProfileShowcase: profileShowcase,
	}

	err = m.attachIdentityImagesToChatIdentity(context, ci)
	if err != nil {
		return nil, err
	}

	return ci, nil
}

// adaptIdentityImageToProtobuf Adapts a images.IdentityImage to protobuf.IdentityImage
func (m *Messenger) adaptIdentityImageToProtobuf(img *images.IdentityImage) *protobuf.IdentityImage {
	return &protobuf.IdentityImage{
		Payload:     img.Payload,
		SourceType:  protobuf.IdentityImage_RAW_PAYLOAD, // TODO add ENS avatar handling to dedicated PR
		ImageFormat: images.GetProtobufImageFormat(img.Payload),
	}
}

func (m *Messenger) attachIdentityImagesToChatIdentity(context ChatContext, ci *protobuf.ChatIdentity) error {
	s, err := m.getSettings()
	if err != nil {
		return err
	}

	if s.ProfilePicturesShowTo == settings.ProfilePicturesShowToNone {
		m.logger.Info(fmt.Sprintf("settings.ProfilePicturesShowTo is set to '%d', skipping attaching IdentityImages", s.ProfilePicturesShowTo))
		return nil
	}

	ciis := make(map[string]*protobuf.IdentityImage)

	switch context {
	case publicChat:
		m.logger.Info(fmt.Sprintf("handling %s ChatIdentity", context))

		img, err := m.multiAccounts.GetIdentityImage(m.account.KeyUID, images.SmallDimName)
		if err != nil {
			return err
		}

		if img == nil {
			return nil
		}

		m.logger.Debug(fmt.Sprintf("%s images.IdentityImage '%s'", context, spew.Sdump(img)))

		ciis[images.SmallDimName] = m.adaptIdentityImageToProtobuf(img)
		m.logger.Debug(fmt.Sprintf("%s protobuf.IdentityImage '%s'", context, spew.Sdump(ciis)))
		ci.Images = ciis

	case privateChat:
		m.logger.Info(fmt.Sprintf("handling %s ChatIdentity", context))

		imgs, err := m.multiAccounts.GetIdentityImages(m.account.KeyUID)
		if err != nil {
			return err
		}

		m.logger.Debug(fmt.Sprintf("%s images.IdentityImage '%s'", context, spew.Sdump(imgs)))

		for _, img := range imgs {
			ciis[img.Name] = m.adaptIdentityImageToProtobuf(img)
		}
		m.logger.Debug(fmt.Sprintf("%s protobuf.IdentityImage '%s'", context, spew.Sdump(ciis)))
		ci.Images = ciis

	default:
		return fmt.Errorf("unknown ChatIdentity context '%s'", context)
	}

	if s.ProfilePicturesShowTo == settings.ProfilePicturesShowToContactsOnly {
		err := EncryptIdentityImagesWithContactPubKeys(ci.Images, m)
		if err != nil {
			return err
		}
	}

	return nil
}

// handleSharedSecrets process the negotiated secrets received from the encryption layer
func (m *Messenger) handleSharedSecrets(secrets []*sharedsecret.Secret) error {
	for _, secret := range secrets {
		fSecret := types.NegotiatedSecret{
			PublicKey: secret.Identity,
			Key:       secret.Key,
		}
		_, err := m.transport.ProcessNegotiatedSecret(fSecret)
		if err != nil {
			return err
		}
	}
	return nil
}

// handleInstallations adds the installations in the installations map
func (m *Messenger) handleInstallations(installations []*multidevice.Installation) {
	for _, installation := range installations {
		if installation.Identity == contactIDFromPublicKey(&m.identity.PublicKey) {
			if _, ok := m.allInstallations.Load(installation.ID); !ok {
				m.allInstallations.Store(installation.ID, installation)
				m.modifiedInstallations.Store(installation.ID, true)
			}
		}
	}
}

// handleEncryptionLayerSubscriptions handles events from the encryption layer
func (m *Messenger) handleEncryptionLayerSubscriptions(subscriptions *encryption.Subscriptions) {
	go func() {
		for {
			select {
			case <-subscriptions.SendContactCode:
				if err := m.publishContactCode(); err != nil {
					m.logger.Error("failed to publish contact code", zap.Error(err))
				}
				// we also piggy-back to clean up cached messages
				if err := m.transport.CleanMessagesProcessed(m.getTimesource().GetCurrentTime() - messageCacheIntervalMs); err != nil {
					m.logger.Error("failed to clean processed messages", zap.Error(err))
				}

			case keys := <-subscriptions.NewHashRatchetKeys:
				if m.communitiesManager == nil {
					continue
				}
				if err := m.communitiesManager.NewHashRatchetKeys(keys); err != nil {
					m.logger.Error("failed to invalidate cache for decrypted communities", zap.Error(err))
				}
			case <-subscriptions.Quit:
				m.logger.Debug("quitting encryption subscription loop")
				return
			}
		}
	}()
}

func (m *Messenger) handleENSVerified(records []*ens.VerificationRecord) {
	var contacts []*Contact
	for _, record := range records {
		m.logger.Info("handling record", zap.Any("record", record))
		contact, ok := m.allContacts.Load(record.PublicKey)
		if !ok {
			m.logger.Info("contact not found")
			continue
		}

		contact.ENSVerified = record.Verified
		contact.EnsName = record.Name
		contacts = append(contacts, contact)
	}

	m.logger.Info("handled records", zap.Any("contacts", contacts))
	if len(contacts) != 0 {
		if err := m.persistence.SaveContacts(contacts); err != nil {
			m.logger.Error("failed to save contacts", zap.Error(err))
			return
		}
	}

	m.PublishMessengerResponse(&MessengerResponse{Contacts: contacts})
}

func (m *Messenger) handleENSVerificationSubscription(c chan []*ens.VerificationRecord) {
	go func() {
		for {
			select {
			case records, more := <-c:
				if !more {
					m.logger.Info("No more records, quitting")
					return
				}
				if len(records) != 0 {
					m.logger.Info("handling records", zap.Any("records", records))
					m.handleENSVerified(records)
				}
			case <-m.quit:
				return
			}
		}
	}()
}

// watchConnectionChange checks the connection status and call handleConnectionChange when this changes
func (m *Messenger) watchConnectionChange() {
	state := false

	processNewState := func(newState bool) {
		if state == newState {
			return
		}
		state = newState
		m.logger.Debug("connection changed", zap.Bool("online", state))
		m.handleConnectionChange(state)
	}

	pollConnectionStatus := func() {
		func() {
			for {
				select {
				case <-time.After(200 * time.Millisecond):
					processNewState(m.Online())
				case <-m.quit:
					return
				}
			}
		}()
	}

	subscribedConnectionStatus := func(subscription *types.ConnStatusSubscription) {
		defer subscription.Unsubscribe()
		for {
			select {
			case status := <-subscription.C:
				processNewState(status.IsOnline)
			case <-m.quit:
				return
			}
		}
	}

	m.logger.Debug("watching connection changes")
	m.Online()
	m.handleConnectionChange(state)

	waku, err := m.node.GetWakuV2(nil)

	if err != nil {
		// No waku v2, we can't watch connection changes
		// Instead we will poll the connection status.
		m.logger.Warn("using WakuV1, can't watch connection changes, this might be have side-effects")
		go pollConnectionStatus()
		return
	}

	subscription, err := waku.SubscribeToConnStatusChanges()
	if err != nil {
		// Log error and fallback to polling
		m.logger.Error("failed to subscribe to connection status changes", zap.Error(err))
		go pollConnectionStatus()
		return
	}

	go subscribedConnectionStatus(subscription)
}

// watchChatsAndCommunitiesToUnmute regularly checks for chats and communities that should be unmuted
func (m *Messenger) watchChatsAndCommunitiesToUnmute() {
	m.logger.Debug("watching unmuted chats")
	go func() {
		for {
			select {
			case <-time.After(1 * time.Minute):
				response := &MessengerResponse{}
				m.allChats.Range(func(chatID string, c *Chat) bool {
					chatMuteTill, _ := time.Parse(time.RFC3339, c.MuteTill.Format(time.RFC3339))
					currTime, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

					if currTime.After(chatMuteTill) && !chatMuteTill.Equal(time.Time{}) && c.Muted {
						err := m.persistence.UnmuteChat(c.ID)
						if err != nil {
							m.logger.Info("err", zap.Any("Couldn't unmute chat", err))
							return false
						}
						c.Muted = false
						c.MuteTill = time.Time{}
						response.AddChat(c)
					}
					return true
				})

				if !response.IsEmpty() {
					signal.SendNewMessages(response)
				}
			case <-m.quit:
				return
			}
		}
	}()
}

// watchCommunitiesToUnmute regularly checks for communities that should be unmuted
func (m *Messenger) watchCommunitiesToUnmute() {
	m.logger.Debug("watching unmuted communities")
	go func() {
		for {
			select {
			case <-time.After(1 * time.Minute):
				response, err := m.CheckCommunitiesToUnmute()
				if err != nil {
					return
				}

				if !response.IsEmpty() {
					signal.SendNewMessages(response)
				}
			case <-m.quit:
				return
			}
		}
	}()
}

// watchExpiredMessages regularly checks for expired emojis and invoke their resending
func (m *Messenger) watchExpiredMessages() {
	m.logger.Debug("watching expired messages")
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				if m.Online() {
					err := m.resendExpiredMessages()
					if err != nil {
						m.logger.Debug("failed to resend expired message", zap.Error(err))
					}
				}
			case <-m.quit:
				return
			}
		}
	}()
}

// watchIdentityImageChanges checks for identity images changes and publishes to the contact code when it happens
func (m *Messenger) watchIdentityImageChanges() {
	m.logger.Debug("watching identity image changes")
	if m.multiAccounts == nil {
		return
	}

	channel := m.multiAccounts.SubscribeToIdentityImageChanges()

	go func() {
		for {
			select {
			case change := <-channel:
				identityImages, err := m.multiAccounts.GetIdentityImages(m.account.KeyUID)
				if err != nil {
					m.logger.Error("failed to get profile pictures to save self contact", zap.Error(err))
					break
				}

				identityImagesMap := make(map[string]images.IdentityImage)
				for _, img := range identityImages {
					identityImagesMap[img.Name] = *img
				}
				m.selfContact.Images = identityImagesMap
				m.publishSelfContactSubscriptions(&SelfContactChangeEvent{ImagesChanged: true})

				if change.PublishExpected {
					err = m.syncProfilePictures(m.dispatchMessage, identityImages)
					if err != nil {
						m.logger.Error("failed to sync profile pictures to paired devices", zap.Error(err))
					}
					err = m.PublishIdentityImage()
					if err != nil {
						m.logger.Error("failed to publish identity image", zap.Error(err))
					}
				}
			case <-m.quit:
				return
			}
		}
	}()
}

func (m *Messenger) watchPendingCommunityRequestToJoin() {
	m.logger.Debug("watching community request to join")

	go func() {
		for {
			select {
			case <-time.After(time.Minute * 10):
				_, err := m.CheckAndDeletePendingRequestToJoinCommunity(context.Background(), false)
				if err != nil {
					m.logger.Error("failed to check and delete pending request to join community", zap.Error(err))
				}
			case <-m.quit:
				return
			}
		}
	}()
}

func (m *Messenger) PublishIdentityImage() error {
	// Reset last published time for ChatIdentity so new contact can receive data
	err := m.resetLastPublishedTimeForChatIdentity()
	if err != nil {
		m.logger.Error("failed to reset publish time", zap.Error(err))
		return err
	}

	// If not online, we schedule it
	if !m.Online() {
		m.shouldPublishContactCode = true
		return nil
	}

	return m.publishContactCode()
}

// handlePushNotificationClientRegistration handles registration events
func (m *Messenger) handlePushNotificationClientRegistrations(c chan struct{}) {
	go func() {
		for {
			_, more := <-c
			if !more {
				return
			}
			if err := m.publishContactCode(); err != nil {
				m.logger.Error("failed to publish contact code", zap.Error(err))
			}

		}
	}()
}

// Init analyzes chats and contacts in order to setup filters
// which are responsible for retrieving messages.
func (m *Messenger) Init() error {

	// Seed the for color generation
	rand.Seed(time.Now().Unix())

	logger := m.logger.With(zap.String("site", "Init"))

	if m.useShards() {
		// Community requests will arrive in this pubsub topic
		err := m.SubscribeToPubsubTopic(shard.DefaultNonProtectedPubsubTopic(), nil)
		if err != nil {
			return err
		}
	}

	var (
		filtersToInit []transport.FiltersToInitialize
		publicKeys    []*ecdsa.PublicKey
	)

	joinedCommunities, err := m.communitiesManager.Joined()
	if err != nil {
		return err
	}
	for _, org := range joinedCommunities {
		// the org advertise on the public topic derived by the pk
		filtersToInit = append(filtersToInit, m.DefaultFilters(org)...)

		// This is for status-go versions that didn't have `CommunitySettings`
		// We need to ensure communities that existed before community settings
		// were introduced will have community settings as well
		exists, err := m.communitiesManager.CommunitySettingsExist(org.ID())
		if err != nil {
			logger.Warn("failed to check if community settings exist", zap.Error(err))
			continue
		}

		if !exists {
			communitySettings := communities.CommunitySettings{
				CommunityID:                  org.IDString(),
				HistoryArchiveSupportEnabled: true,
			}

			err = m.communitiesManager.SaveCommunitySettings(communitySettings)
			if err != nil {
				logger.Warn("failed to save community settings", zap.Error(err))
			}
			continue
		}

		// In case we do have settings, but the history archive support is disabled
		// for this community, we enable it, as this should be the default for all
		// non-admin communities
		communitySettings, err := m.communitiesManager.GetCommunitySettingsByID(org.ID())
		if err != nil {
			logger.Warn("failed to fetch community settings", zap.Error(err))
			continue
		}

		if !org.IsControlNode() && !communitySettings.HistoryArchiveSupportEnabled {
			communitySettings.HistoryArchiveSupportEnabled = true
			err = m.communitiesManager.UpdateCommunitySettings(*communitySettings)
			if err != nil {
				logger.Warn("failed to update community settings", zap.Error(err))
			}
		}
	}

	spectatedCommunities, err := m.communitiesManager.Spectated()
	if err != nil {
		return err
	}
	for _, org := range spectatedCommunities {
		filtersToInit = append(filtersToInit, m.DefaultFilters(org)...)
	}

	// Get chat IDs and public keys from the existing chats.
	// TODO: Get only active chats by the query.
	chats, err := m.persistence.Chats()
	if err != nil {
		return err
	}
	for _, chat := range chats {
		if err := chat.Validate(); err != nil {
			logger.Warn("failed to validate chat", zap.Error(err))
			continue
		}

		if err = m.initChatFirstMessageTimestamp(chat); err != nil {
			logger.Warn("failed to init first message timestamp", zap.Error(err))
			continue
		}

		m.allChats.Store(chat.ID, chat)

		if !chat.Active || chat.Timeline() {
			continue
		}

		communityInfo := make(map[string]*communities.Community)

		switch chat.ChatType {
		case ChatTypePublic, ChatTypeProfile:
			filtersToInit = append(filtersToInit, transport.FiltersToInitialize{ChatID: chat.ID})
		case ChatTypeCommunityChat:
			communityID, err := hexutil.Decode(chat.CommunityID)
			if err != nil {
				return err
			}

			community, ok := communityInfo[chat.CommunityID]
			if !ok {
				community, err = m.communitiesManager.GetByID(communityID)
				if err != nil {
					return err
				}
				communityInfo[chat.CommunityID] = community
			}

			filtersToInit = append(filtersToInit, transport.FiltersToInitialize{ChatID: chat.ID, PubsubTopic: community.PubsubTopic()})
		case ChatTypeOneToOne:
			pk, err := chat.PublicKey()
			if err != nil {
				return err
			}
			publicKeys = append(publicKeys, pk)
		case ChatTypePrivateGroupChat:
			for _, member := range chat.Members {
				publicKey, err := member.PublicKey()
				if err != nil {
					return errors.Wrapf(err, "invalid public key for member %s in chat %s", member.ID, chat.Name)
				}
				publicKeys = append(publicKeys, publicKey)
			}
		default:
			return errors.New("invalid chat type")
		}
	}

	// Timeline and profile chats are deprecated.
	// This code can be removed after some reasonable time.

	// upsert timeline chat
	if !deprecation.ChatProfileDeprecated {
		err = m.ensureTimelineChat()
		if err != nil {
			return err
		}
	}

	// upsert profile chat
	if !deprecation.ChatTimelineDeprecated {
		err = m.ensureMyOwnProfileChat()
		if err != nil {
			return err
		}
	}

	// Get chat IDs and public keys from the contacts.
	contacts, err := m.persistence.Contacts()
	if err != nil {
		return err
	}
	for idx, contact := range contacts {
		if err = m.updateContactImagesURL(contact); err != nil {
			return err
		}
		m.allContacts.Store(contact.ID, contacts[idx])
		// We only need filters for contacts added by us and not blocked.
		if !contact.added() || contact.Blocked {
			continue
		}
		publicKey, err := contact.PublicKey()
		if err != nil {
			logger.Error("failed to get contact's public key", zap.Error(err))
			continue
		}
		publicKeys = append(publicKeys, publicKey)
	}

	installations, err := m.encryptor.GetOurInstallations(&m.identity.PublicKey)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		m.allInstallations.Store(installation.ID, installation)
	}

	err = m.setInstallationHostname()
	if err != nil {
		return err
	}

	_, err = m.transport.InitFilters(filtersToInit, publicKeys)
	if err != nil {
		return err
	}

	// Init filters for the communities we control
	var communityFiltersToInitialize []transport.CommunityFilterToInitialize
	controlledCommunities, err := m.communitiesManager.Controlled()
	if err != nil {
		return err
	}

	for _, c := range controlledCommunities {
		communityFiltersToInitialize = append(communityFiltersToInitialize, transport.CommunityFilterToInitialize{
			Shard:   c.Shard(),
			PrivKey: c.PrivateKey(),
		})
	}

	_, err = m.InitCommunityFilters(communityFiltersToInitialize)
	if err != nil {
		return err
	}

	return nil
}

// Shutdown takes care of ensuring a clean shutdown of Messenger
func (m *Messenger) Shutdown() (err error) {
	if m == nil {
		return nil
	}

	select {
	case _, ok := <-m.quit:
		if !ok {
			return errors.New("messenger already shutdown")
		}
	default:
	}

	close(m.quit)
	m.cancel()
	m.shutdownWaitGroup.Wait()
	for i, task := range m.shutdownTasks {
		m.logger.Debug("running shutdown task", zap.Int("n", i))
		if tErr := task(); tErr != nil {
			m.logger.Info("shutdown task failed", zap.Error(tErr))
			if err == nil {
				// First error appeared.
				err = tErr
			} else {
				// We return all errors. They will be concatenated in the order of occurrence,
				// however, they will also be returned as a single error.
				err = errors.Wrap(err, tErr.Error())
			}
		}
	}
	return
}

func (m *Messenger) EnableInstallation(id string) error {
	installation, ok := m.allInstallations.Load(id)
	if !ok {
		return errors.New("no installation found")
	}

	err := m.encryptor.EnableInstallation(&m.identity.PublicKey, id)
	if err != nil {
		return err
	}
	installation.Enabled = true
	// TODO(samyoul) remove storing of an updated reference pointer?
	m.allInstallations.Store(id, installation)
	return nil
}

func (m *Messenger) DisableInstallation(id string) error {
	installation, ok := m.allInstallations.Load(id)
	if !ok {
		return errors.New("no installation found")
	}

	err := m.encryptor.DisableInstallation(&m.identity.PublicKey, id)
	if err != nil {
		return err
	}
	installation.Enabled = false
	// TODO(samyoul) remove storing of an updated reference pointer?
	m.allInstallations.Store(id, installation)
	return nil
}

func (m *Messenger) Installations() []*multidevice.Installation {
	installations := make([]*multidevice.Installation, m.allInstallations.Len())

	var i = 0
	m.allInstallations.Range(func(installationID string, installation *multidevice.Installation) (shouldContinue bool) {
		installations[i] = installation
		i++
		return true
	})
	return installations
}

func (m *Messenger) setInstallationMetadata(id string, data *multidevice.InstallationMetadata) error {
	installation, ok := m.allInstallations.Load(id)
	if !ok {
		return errors.New("no installation found")
	}

	installation.InstallationMetadata = data
	return m.encryptor.SetInstallationMetadata(m.IdentityPublicKey(), id, data)
}

func (m *Messenger) SetInstallationMetadata(id string, data *multidevice.InstallationMetadata) error {
	return m.setInstallationMetadata(id, data)
}

func (m *Messenger) SetInstallationName(id string, name string) error {
	installation, ok := m.allInstallations.Load(id)
	if !ok {
		return errors.New("no installation found")
	}

	installation.InstallationMetadata.Name = name
	return m.encryptor.SetInstallationName(m.IdentityPublicKey(), id, name)
}

// NOT IMPLEMENTED
func (m *Messenger) SelectMailserver(id string) error {
	return ErrNotImplemented
}

// NOT IMPLEMENTED
func (m *Messenger) AddMailserver(enode string) error {
	return ErrNotImplemented
}

// NOT IMPLEMENTED
func (m *Messenger) RemoveMailserver(id string) error {
	return ErrNotImplemented
}

// NOT IMPLEMENTED
func (m *Messenger) Mailservers() ([]string, error) {
	return nil, ErrNotImplemented
}

func (m *Messenger) initChatFirstMessageTimestamp(chat *Chat) error {
	if !chat.CommunityChat() || chat.FirstMessageTimestamp != FirstMessageTimestampUndefined {
		return nil
	}

	oldestMessageTimestamp, hasAnyMessage, err := m.persistence.OldestMessageWhisperTimestampByChatID(chat.ID)
	if err != nil {
		return err
	}

	if hasAnyMessage {
		if oldestMessageTimestamp == FirstMessageTimestampUndefined {
			return nil
		}
		return m.updateChatFirstMessageTimestamp(chat, whisperToUnixTimestamp(oldestMessageTimestamp), &MessengerResponse{})
	}

	return m.updateChatFirstMessageTimestamp(chat, FirstMessageTimestampNoMessage, &MessengerResponse{})
}

func (m *Messenger) addMessagesAndChat(chat *Chat, messages []*common.Message, response *MessengerResponse) (*MessengerResponse, error) {
	response.AddChat(chat)
	response.AddMessages(messages)
	err := m.persistence.SaveMessages(response.Messages())
	if err != nil {
		return nil, err
	}

	return response, m.saveChat(chat)
}

func (m *Messenger) reregisterForPushNotifications() error {
	m.logger.Info("contact state changed, re-registering for push notification")
	if m.pushNotificationClient == nil {
		return nil
	}

	return m.pushNotificationClient.Reregister(m.pushNotificationOptions())
}

// pull a message from the database and send it again
func (m *Messenger) reSendRawMessage(ctx context.Context, messageID string) error {
	message, err := m.persistence.RawMessageByID(messageID)
	if err != nil {
		return err
	}

	chat, ok := m.allChats.Load(message.LocalChatID)
	if !ok {
		return errors.New("chat not found")
	}

	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             message.Payload,
		PubsubTopic:         message.PubsubTopic,
		MessageType:         message.MessageType,
		Recipients:          message.Recipients,
		ResendAutomatically: message.ResendAutomatically,
		SendCount:           message.SendCount,
	})
	return err
}

// ReSendChatMessage pulls a message from the database and sends it again
func (m *Messenger) ReSendChatMessage(ctx context.Context, messageID string) error {
	return m.reSendRawMessage(ctx, messageID)
}

func (m *Messenger) SetLocalPairing(localPairing bool) {
	m.localPairing = localPairing
}
func (m *Messenger) hasPairedDevices() bool {
	logger := m.logger.Named("hasPairedDevices")

	if m.localPairing {
		return true
	}

	var count int
	m.allInstallations.Range(func(installationID string, installation *multidevice.Installation) (shouldContinue bool) {
		if installation.Enabled {
			count++
		}
		return true
	})
	logger.Debug("installations info",
		zap.Int("Number of installations", m.allInstallations.Len()),
		zap.Int("Number of enabled installations", count))
	return count > 1
}

func (m *Messenger) HasPairedDevices() bool {
	return m.hasPairedDevices()
}

// sendToPairedDevices will check if we have any paired devices and send to them if necessary
func (m *Messenger) sendToPairedDevices(ctx context.Context, spec common.RawMessage) error {
	hasPairedDevices := m.hasPairedDevices()
	// We send a message to any paired device
	if hasPairedDevices {
		_, err := m.sender.SendPrivate(ctx, &m.identity.PublicKey, &spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Messenger) dispatchPairInstallationMessage(ctx context.Context, spec common.RawMessage) (common.RawMessage, error) {
	var err error
	var id []byte

	id, err = m.sender.SendPairInstallation(ctx, &m.identity.PublicKey, spec)

	if err != nil {
		return spec, err
	}
	spec.ID = types.EncodeHex(id)
	spec.SendCount++
	err = m.persistence.SaveRawMessage(&spec)
	if err != nil {
		return spec, err
	}

	return spec, nil
}

func (m *Messenger) dispatchMessage(ctx context.Context, rawMessage common.RawMessage) (common.RawMessage, error) {
	var err error
	var id []byte
	logger := m.logger.With(zap.String("site", "dispatchMessage"), zap.String("chatID", rawMessage.LocalChatID))
	chat, ok := m.allChats.Load(rawMessage.LocalChatID)
	if !ok {
		return rawMessage, errors.New("no chat found")
	}

	switch chat.ChatType {
	case ChatTypeOneToOne:
		publicKey, err := chat.PublicKey()
		if err != nil {
			return rawMessage, err
		}

		//SendPrivate will alter message identity and possibly datasyncid, so we save an unchanged
		//message for sending to paired devices later
		specCopyForPairedDevices := rawMessage
		if !common.IsPubKeyEqual(publicKey, &m.identity.PublicKey) || rawMessage.SkipEncryptionLayer {
			id, err = m.sender.SendPrivate(ctx, publicKey, &rawMessage)

			if err != nil {
				return rawMessage, err
			}
		}

		err = m.sendToPairedDevices(ctx, specCopyForPairedDevices)

		if err != nil {
			return rawMessage, err
		}

	case ChatTypePublic, ChatTypeProfile:
		logger.Debug("sending public message", zap.String("chatName", chat.Name))
		id, err = m.sender.SendPublic(ctx, chat.ID, rawMessage)
		if err != nil {
			return rawMessage, err
		}

	case ChatTypeCommunityChat:
		community, err := m.communitiesManager.GetByIDString(chat.CommunityID)
		if err != nil {
			return rawMessage, err
		}
		rawMessage.PubsubTopic = community.PubsubTopic()

		canPost, err := m.communitiesManager.CanPost(&m.identity.PublicKey, chat.CommunityID, chat.CommunityChatID(), rawMessage.MessageType)
		if err != nil {
			return rawMessage, err
		}

		if !canPost {
			m.logger.Error("can't post on chat",
				zap.String("chatID", chat.ID),
				zap.String("chatName", chat.Name),
				zap.Any("messageType", rawMessage.MessageType),
			)
			return rawMessage, fmt.Errorf("can't post message type '%d' on chat '%s'", rawMessage.MessageType, chat.ID)
		}

		logger.Debug("sending community chat message", zap.String("chatName", chat.Name))
		isCommunityEncrypted, err := m.communitiesManager.IsEncrypted(chat.CommunityID)
		if err != nil {
			return rawMessage, err
		}
		isChannelEncrypted, err := m.communitiesManager.IsChannelEncrypted(chat.CommunityID, chat.ID)
		if err != nil {
			return rawMessage, err
		}
		isEncrypted := isCommunityEncrypted || isChannelEncrypted
		if !isEncrypted {
			id, err = m.sender.SendPublic(ctx, chat.ID, rawMessage)
			if err != nil {
				return rawMessage, err
			}
		} else {
			rawMessage.CommunityID, err = types.DecodeHex(chat.CommunityID)
			if err != nil {
				return rawMessage, err
			}

			if isChannelEncrypted {
				rawMessage.HashRatchetGroupID = []byte(chat.ID)
			} else {
				rawMessage.HashRatchetGroupID = rawMessage.CommunityID
			}

			id, err = m.sender.SendCommunityMessage(ctx, rawMessage)
			if err != nil {
				return rawMessage, err
			}
		}
	case ChatTypePrivateGroupChat:
		logger.Debug("sending group message", zap.String("chatName", chat.Name))
		if rawMessage.Recipients == nil {
			rawMessage.Recipients, err = chat.MembersAsPublicKeys()
			if err != nil {
				return rawMessage, err
			}
		}

		hasPairedDevices := m.hasPairedDevices()

		if !hasPairedDevices {

			// Filter out my key from the recipients
			n := 0
			for _, recipient := range rawMessage.Recipients {
				if !common.IsPubKeyEqual(recipient, &m.identity.PublicKey) {
					rawMessage.Recipients[n] = recipient
					n++
				}
			}
			rawMessage.Recipients = rawMessage.Recipients[:n]
		}

		// We won't really send the message out if there's no recipients
		if len(rawMessage.Recipients) == 0 {
			rawMessage.Sent = true
		}

		// We skip wrapping in some cases (emoji reactions for example)
		if !rawMessage.SkipGroupMessageWrap {
			rawMessage.MessageType = protobuf.ApplicationMetadataMessage_MEMBERSHIP_UPDATE_MESSAGE
		}

		id, err = m.sender.SendGroup(ctx, rawMessage.Recipients, rawMessage)
		if err != nil {
			return rawMessage, err
		}

	default:
		return rawMessage, errors.New("chat type not supported")
	}
	rawMessage.ID = types.EncodeHex(id)
	rawMessage.SendCount++
	rawMessage.LastSent = m.getTimesource().GetCurrentTime()
	err = m.persistence.SaveRawMessage(&rawMessage)
	if err != nil {
		return rawMessage, err
	}

	if m.dispatchMessageTestCallback != nil {
		m.dispatchMessageTestCallback(rawMessage)
	}
	return rawMessage, nil
}

// SendChatMessage takes a minimal message and sends it based on the corresponding chat
func (m *Messenger) SendChatMessage(ctx context.Context, message *common.Message) (*MessengerResponse, error) {
	return m.sendChatMessage(ctx, message)
}

// SendChatMessages takes a array of messages and sends it based on the corresponding chats
func (m *Messenger) SendChatMessages(ctx context.Context, messages []*common.Message) (*MessengerResponse, error) {
	var response MessengerResponse

	generatedAlbumID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	imagesCount := uint32(0)
	for _, message := range messages {
		if message.ContentType == protobuf.ChatMessage_IMAGE {
			imagesCount++
		}

	}

	for _, message := range messages {
		if message.ContentType == protobuf.ChatMessage_IMAGE && len(messages) > 1 {
			err = message.SetAlbumIDAndImagesCount(generatedAlbumID.String(), imagesCount)
			if err != nil {
				return nil, err
			}
		}
		messageResponse, err := m.SendChatMessage(ctx, message)
		if err != nil {
			return nil, err
		}
		err = response.Merge(messageResponse)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}

// sendChatMessage takes a minimal message and sends it based on the corresponding chat
func (m *Messenger) sendChatMessage(ctx context.Context, message *common.Message) (*MessengerResponse, error) {
	displayName, err := m.settings.DisplayName()
	if err != nil {
		return nil, err
	}

	message.DisplayName = displayName

	replacedText, err := m.mentionsManager.ReplaceWithPublicKey(message.ChatId, message.Text)
	if err == nil {
		message.Text = replacedText
	} else {
		m.logger.Error("failed to replace text with public key", zap.String("chatID", message.ChatId), zap.String("text", message.Text))
	}

	if len(message.ImagePath) != 0 {

		err := message.LoadImage()
		if err != nil {
			return nil, err
		}

	} else if len(message.CommunityID) != 0 {
		community, err := m.communitiesManager.GetByIDString(message.CommunityID)
		if err != nil {
			return nil, err
		}

		wrappedCommunity, err := community.ToProtocolMessageBytes()
		if err != nil {
			return nil, err
		}

		message.Payload = &protobuf.ChatMessage_Community{Community: wrappedCommunity}
		message.Shard = community.Shard().Protobuffer()

		message.ContentType = protobuf.ChatMessage_COMMUNITY
	} else if len(message.AudioPath) != 0 {
		err := message.LoadAudio()
		if err != nil {
			return nil, err
		}
	}

	// We consider link previews non-critical data, so we do not want to block
	// messages from being sent.

	unfurledLinks, err := message.ConvertLinkPreviewsToProto()
	if err != nil {
		m.logger.Error("failed to convert link previews", zap.Error(err))
	} else {
		message.UnfurledLinks = unfurledLinks
	}

	unfurledStatusLinks, err := message.ConvertStatusLinkPreviewsToProto()
	if err != nil {
		m.logger.Error("failed to convert status link previews", zap.Error(err))
	} else {
		message.UnfurledStatusLinks = unfurledStatusLinks
	}

	var response MessengerResponse

	// A valid added chat is required.
	chat, ok := m.allChats.Load(message.ChatId)
	if !ok {
		return nil, errors.New("Chat not found")
	}

	err = m.handleStandaloneChatIdentity(chat)
	if err != nil {
		return nil, err
	}

	err = extendMessageFromChat(message, chat, &m.identity.PublicKey, m.getTimesource())
	if err != nil {
		return nil, err
	}

	err = m.addContactRequestPropagatedState(message)
	if err != nil {
		return nil, err
	}

	encodedMessage, err := m.encodeChatEntity(chat, message)
	if err != nil {
		return nil, err
	}

	rawMessage := common.RawMessage{
		LocalChatID:          chat.ID,
		SendPushNotification: m.featureFlags.PushNotifications,
		Payload:              encodedMessage,
		MessageType:          protobuf.ApplicationMetadataMessage_CHAT_MESSAGE,
		ResendAutomatically:  true,
	}

	// We want to save the raw message before dispatching it, to avoid race conditions
	// since it might get dispatched and confirmed before it's saved.
	// This is not the best solution, probably it would be better to split
	// the sent status in a different table and join on query for messages,
	// but that's a much larger change and it would require an expensive migration of clients
	rawMessage.BeforeDispatch = func(rawMessage *common.RawMessage) error {

		if rawMessage.Sent {
			message.OutgoingStatus = common.OutgoingStatusSent
		}
		message.ID = rawMessage.ID
		err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
		if err != nil {
			return err
		}

		err = chat.UpdateFromMessage(message, m.getTimesource())
		if err != nil {
			return err
		}

		err := m.persistence.SaveMessages([]*common.Message{message})
		if err != nil {
			return err
		}

		var syncMessageType peersyncing.SyncMessageType
		if chat.OneToOne() {
			syncMessageType = peersyncing.SyncMessageOneToOneType
		} else if chat.CommunityChat() {
			syncMessageType = peersyncing.SyncMessageCommunityType
		} else if chat.PrivateGroupChat() {
			syncMessageType = peersyncing.SyncMessagePrivateGroup

		}

		wrappedMessage, err := v1protocol.WrapMessageV1(rawMessage.Payload, rawMessage.MessageType, rawMessage.Sender)
		if err != nil {
			return errors.Wrap(err, "failed to wrap message")
		}

		syncMessage := peersyncing.SyncMessage{
			Type:      syncMessageType,
			ID:        types.Hex2Bytes(rawMessage.ID),
			GroupID:   []byte(chat.ID),
			Payload:   wrappedMessage,
			Timestamp: m.transport.GetCurrentTime() / 1000,
		}

		// If the chat type is not supported, skip saving it
		if syncMessageType == 0 {
			return nil
		}

		// ensure that the message is saved only once
		rawMessage.BeforeDispatch = nil

		return m.peersyncing.Add(syncMessage)
	}

	rawMessage, err = m.dispatchMessage(ctx, rawMessage)
	if err != nil {
		return nil, err
	}

	msg, err := m.pullMessagesAndResponsesFromDB([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	if err := m.updateChatFirstMessageTimestamp(chat, whisperToUnixTimestamp(message.WhisperTimestamp), &response); err != nil {
		return nil, err
	}

	response.SetMessages(msg)
	response.AddChat(chat)

	m.logger.Debug("inside sendChatMessage",
		zap.String("id", message.ID),
		zap.String("text", message.Text),
		zap.String("from", message.From),
		zap.String("displayName", message.DisplayName),
		zap.String("ChatId", message.ChatId),
		zap.String("Clock", strconv.FormatUint(message.Clock, 10)),
		zap.String("Timestamp", strconv.FormatUint(message.Timestamp, 10)),
	)
	err = m.prepareMessages(response.messages)

	if err != nil {
		return nil, err
	}

	return &response, m.saveChat(chat)
}

func whisperToUnixTimestamp(whisperTimestamp uint64) uint32 {
	return uint32(whisperTimestamp / 1000)
}

func (m *Messenger) updateChatFirstMessageTimestamp(chat *Chat, timestamp uint32, response *MessengerResponse) error {
	// Currently supported only for communities
	if !chat.CommunityChat() {
		return nil
	}

	community, err := m.communitiesManager.GetByIDString(chat.CommunityID)
	if err != nil {
		return err
	}

	if community.IsControlNode() && chat.UpdateFirstMessageTimestamp(timestamp) {
		community, changes, err := m.communitiesManager.EditChatFirstMessageTimestamp(community.ID(), chat.ID, chat.FirstMessageTimestamp)
		if err != nil {
			return err
		}

		response.AddCommunity(community)
		response.CommunityChanges = append(response.CommunityChanges, changes)
	}

	return nil
}

func (m *Messenger) ShareImageMessage(request *requests.ShareImageMessage) (*MessengerResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	response := &MessengerResponse{}

	msg, err := m.persistence.MessageByID(request.MessageID)
	if err != nil {
		return nil, err
	}

	var messages []*common.Message
	for _, pk := range request.Users {
		message := common.NewMessage()
		message.ChatId = pk.String()
		message.Payload = msg.Payload
		message.Text = "This message has been shared with you"
		message.ContentType = protobuf.ChatMessage_IMAGE
		messages = append(messages, message)

		r, err := m.CreateOneToOneChat(&requests.CreateOneToOneChat{ID: pk})
		if err != nil {
			return nil, err
		}

		if err := response.Merge(r); err != nil {
			return nil, err
		}
	}

	sendMessagesResponse, err := m.SendChatMessages(context.Background(), messages)
	if err != nil {
		return nil, err
	}

	if err := response.Merge(sendMessagesResponse); err != nil {
		return nil, err
	}

	return response, nil
}

func (m *Messenger) syncProfilePicturesFromDatabase(rawMessageHandler RawMessageHandler) error {
	keyUID := m.account.KeyUID
	identityImages, err := m.multiAccounts.GetIdentityImages(keyUID)
	if err != nil {
		return err
	}
	return m.syncProfilePictures(rawMessageHandler, identityImages)
}

func (m *Messenger) syncProfilePictures(rawMessageHandler RawMessageHandler, identityImages []*images.IdentityImage) error {
	if !m.hasPairedDevices() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pictures := make([]*protobuf.SyncProfilePicture, len(identityImages))
	clock, chat := m.getLastClockWithRelatedChat()
	for i, image := range identityImages {
		p := &protobuf.SyncProfilePicture{}
		p.Name = image.Name
		p.Payload = image.Payload
		p.Width = uint32(image.Width)
		p.Height = uint32(image.Height)
		p.FileSize = uint32(image.FileSize)
		p.ResizeTarget = uint32(image.ResizeTarget)
		if image.Clock == 0 {
			p.Clock = clock
		} else {
			p.Clock = image.Clock
		}
		pictures[i] = p
	}

	message := &protobuf.SyncProfilePictures{}
	message.KeyUid = m.account.KeyUID
	message.Pictures = pictures

	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_PROFILE_PICTURES,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

// SyncDevices sends all public chats and contacts to paired devices
// TODO remove use of photoPath in contacts
func (m *Messenger) SyncDevices(ctx context.Context, ensName, photoPath string, rawMessageHandler RawMessageHandler) (err error) {
	if rawMessageHandler == nil {
		rawMessageHandler = m.dispatchMessage
	}

	myID := contactIDFromPublicKey(&m.identity.PublicKey)

	displayName, err := m.settings.DisplayName()
	if err != nil {
		return err
	}

	if _, err = m.sendContactUpdate(ctx, myID, displayName, ensName, photoPath, rawMessageHandler); err != nil {
		return err
	}

	m.allChats.Range(func(chatID string, chat *Chat) bool {
		if !chat.shouldBeSynced() {
			return true

		}
		err = m.syncChat(ctx, chat, rawMessageHandler)
		return err == nil
	})
	if err != nil {
		return err
	}

	m.allContacts.Range(func(contactID string, contact *Contact) bool {
		if contact.ID != myID &&
			(contact.LocalNickname != "" || contact.added() || contact.Blocked) {
			if err = m.syncContact(ctx, contact, rawMessageHandler); err != nil {
				return false
			}
		}
		return true
	})

	cs, err := m.communitiesManager.JoinedAndPendingCommunitiesWithRequests()
	if err != nil {
		return err
	}
	for _, c := range cs {
		if err = m.syncCommunity(ctx, c, rawMessageHandler); err != nil {
			return err
		}
	}

	bookmarks, err := m.browserDatabase.GetBookmarks()
	if err != nil {
		return err
	}
	for _, b := range bookmarks {
		if err = m.SyncBookmark(ctx, b, rawMessageHandler); err != nil {
			return err
		}
	}

	trustedUsers, err := m.verificationDatabase.GetAllTrustStatus()
	if err != nil {
		return err
	}
	for id, ts := range trustedUsers {
		if err = m.SyncTrustedUser(ctx, id, ts, rawMessageHandler); err != nil {
			return err
		}
	}

	verificationRequests, err := m.verificationDatabase.GetVerificationRequests()
	if err != nil {
		return err
	}
	for i := range verificationRequests {
		if err = m.SyncVerificationRequest(ctx, &verificationRequests[i], rawMessageHandler); err != nil {
			return err
		}
	}

	err = m.syncSettings(rawMessageHandler)
	if err != nil {
		return err
	}

	err = m.syncProfilePicturesFromDatabase(rawMessageHandler)
	if err != nil {
		return err
	}

	ids, err := m.persistence.LatestContactRequestIDs()

	if err != nil {
		return err
	}

	for id, state := range ids {
		if state == common.ContactRequestStateAccepted || state == common.ContactRequestStateDismissed {
			accepted := state == common.ContactRequestStateAccepted
			err := m.syncContactRequestDecision(ctx, id, accepted, rawMessageHandler)
			if err != nil {
				return err
			}
		}
	}

	// we have to sync deleted keypairs as well
	keypairs, err := m.settings.GetAllKeypairs()
	if err != nil {
		return err
	}

	for _, kp := range keypairs {
		err = m.syncKeypair(kp, rawMessageHandler)
		if err != nil {
			return err
		}
	}

	// we have to sync deleted watch only accounts as well
	woAccounts, err := m.settings.GetAllWatchOnlyAccounts()
	if err != nil {
		return err
	}

	for _, woAcc := range woAccounts {
		err = m.syncWalletAccount(woAcc, rawMessageHandler)
		if err != nil {
			return err
		}
	}

	savedAddresses, err := m.savedAddressesManager.GetRawSavedAddresses()
	if err != nil {
		return err
	}

	for i := range savedAddresses {
		sa := savedAddresses[i]

		err = m.syncSavedAddress(ctx, sa, rawMessageHandler)
		if err != nil {
			return err
		}
	}

	if err = m.syncEnsUsernameDetails(ctx, rawMessageHandler); err != nil {
		return err
	}

	if err = m.syncDeleteForMeMessage(ctx, rawMessageHandler); err != nil {
		return err
	}

	err = m.syncAccountsPositions(rawMessageHandler)
	if err != nil {
		return err
	}

	err = m.syncSocialLinks(context.Background(), rawMessageHandler)
	if err != nil {
		return err
	}

	err = m.syncProfileShowcasePreferences(context.Background(), rawMessageHandler)
	if err != nil {
		return err
	}

	return nil
}

func (m *Messenger) syncContactRequestDecision(ctx context.Context, requestID string, accepted bool, rawMessageHandler RawMessageHandler) error {
	m.logger.Info("syncContactRequestDecision", zap.Any("from", requestID))
	if !m.hasPairedDevices() {
		return nil
	}

	clock, chat := m.getLastClockWithRelatedChat()

	var status protobuf.SyncContactRequestDecision_DecisionStatus
	if accepted {
		status = protobuf.SyncContactRequestDecision_ACCEPTED
	} else {
		status = protobuf.SyncContactRequestDecision_DECLINED
	}

	message := &protobuf.SyncContactRequestDecision{
		RequestId:      requestID,
		Clock:          clock,
		DecisionStatus: status,
	}

	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_CONTACT_REQUEST_DECISION,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	return nil
}

func (m *Messenger) getLastClockWithRelatedChat() (uint64, *Chat) {
	chatID := contactIDFromPublicKey(&m.identity.PublicKey)

	chat, ok := m.allChats.Load(chatID)
	if !ok {
		chat = OneToOneFromPublicKey(&m.identity.PublicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	m.allChats.Store(chat.ID, chat)
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	return clock, chat
}

// SendPairInstallation sends a pair installation message
func (m *Messenger) SendPairInstallation(ctx context.Context, rawMessageHandler RawMessageHandler) (*MessengerResponse, error) {
	var err error
	var response MessengerResponse

	installation, ok := m.allInstallations.Load(m.installationID)
	if !ok {
		return nil, errors.New("no installation found")
	}

	if installation.InstallationMetadata == nil {
		return nil, errors.New("no installation metadata")
	}

	clock, chat := m.getLastClockWithRelatedChat()

	pairMessage := &protobuf.SyncPairInstallation{
		Clock:          clock,
		Name:           installation.InstallationMetadata.Name,
		InstallationId: installation.ID,
		DeviceType:     installation.InstallationMetadata.DeviceType,
		Version:        installation.Version}
	encodedMessage, err := proto.Marshal(pairMessage)
	if err != nil {
		return nil, err
	}

	if rawMessageHandler == nil {
		rawMessageHandler = m.dispatchPairInstallationMessage
	}
	_, err = rawMessageHandler(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_PAIR_INSTALLATION,
		ResendAutomatically: true,
	})
	if err != nil {
		return nil, err
	}

	response.AddChat(chat)

	chat.LastClockValue = clock
	err = m.saveChat(chat)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// syncChat sync a chat with paired devices
func (m *Messenger) syncChat(ctx context.Context, chatToSync *Chat, rawMessageHandler RawMessageHandler) error {
	var err error
	if !m.hasPairedDevices() {
		return nil
	}
	clock, chat := m.getLastClockWithRelatedChat()

	syncMessage := &protobuf.SyncChat{
		Clock:    clock,
		Id:       chatToSync.ID,
		Name:     chatToSync.Name,
		ChatType: uint32(chatToSync.ChatType),
		Active:   chatToSync.Active,
	}
	chatMuteTill, _ := time.Parse(time.RFC3339, chatToSync.MuteTill.Format(time.RFC3339))
	if chatToSync.Muted && chatMuteTill.Equal(time.Time{}) {
		// Only set Muted if it is "permanently" muted
		syncMessage.Muted = true
	}
	if chatToSync.OneToOne() {
		syncMessage.Name = "" // The Name is useless in 1-1 chats
	}
	if chatToSync.PrivateGroupChat() {
		syncMessage.MembershipUpdateEvents = make([]*protobuf.MembershipUpdateEvents, len(chatToSync.MembershipUpdates))
		for i, membershipUpdate := range chatToSync.MembershipUpdates {
			syncMessage.MembershipUpdateEvents[i] = &protobuf.MembershipUpdateEvents{
				Clock:      membershipUpdate.ClockValue,
				Type:       uint32(membershipUpdate.Type),
				Members:    membershipUpdate.Members,
				Name:       membershipUpdate.Name,
				Signature:  membershipUpdate.Signature,
				ChatId:     membershipUpdate.ChatID,
				From:       membershipUpdate.From,
				RawPayload: membershipUpdate.RawPayload,
				Color:      membershipUpdate.Color,
				Image:      membershipUpdate.Image,
			}
		}
	}
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_CHAT,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) syncClearHistory(ctx context.Context, publicChat *Chat, rawMessageHandler RawMessageHandler) error {
	var err error
	if !m.hasPairedDevices() {
		return nil
	}
	clock, chat := m.getLastClockWithRelatedChat()

	syncMessage := &protobuf.SyncClearHistory{
		ChatId:    publicChat.ID,
		ClearedAt: publicChat.DeletedAtClockValue,
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_CLEAR_HISTORY,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) syncChatRemoving(ctx context.Context, id string, rawMessageHandler RawMessageHandler) error {
	var err error
	if !m.hasPairedDevices() {
		return nil
	}
	clock, chat := m.getLastClockWithRelatedChat()

	syncMessage := &protobuf.SyncChatRemoved{
		Clock: clock,
		Id:    id,
	}
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_CHAT_REMOVED,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

// syncContact sync as contact with paired devices
func (m *Messenger) syncContact(ctx context.Context, contact *Contact, rawMessageHandler RawMessageHandler) error {
	var err error
	if contact.IsSyncing {
		return nil
	}
	if !m.hasPairedDevices() {
		return nil
	}
	clock, chat := m.getLastClockWithRelatedChat()

	syncMessage := m.buildSyncContactMessage(contact)

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_INSTALLATION_CONTACT_V2,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) propagateSyncInstallationCommunityWithHRKeys(msg *protobuf.SyncInstallationCommunity, c *communities.Community) error {
	communityKeys, err := m.encryptor.GetAllHRKeysMarshaledV1(c.ID())
	if err != nil {
		return err
	}
	msg.EncryptionKeysV1 = communityKeys

	communityAndChannelKeys := [][]byte{}
	communityKeys, err = m.encryptor.GetAllHRKeysMarshaledV2(c.ID())
	if err != nil {
		return err
	}
	if len(communityKeys) > 0 {
		communityAndChannelKeys = append(communityAndChannelKeys, communityKeys)
	}

	for channelID := range c.Chats() {
		channelKeys, err := m.encryptor.GetAllHRKeysMarshaledV2([]byte(c.IDString() + channelID))
		if err != nil {
			return err
		}
		if len(channelKeys) > 0 {
			communityAndChannelKeys = append(communityAndChannelKeys, channelKeys)
		}
	}
	msg.EncryptionKeysV2 = communityAndChannelKeys

	return nil
}

func (m *Messenger) syncCommunity(ctx context.Context, community *communities.Community, rawMessageHandler RawMessageHandler) error {
	logger := m.logger.Named("syncCommunity")
	if !m.hasPairedDevices() {
		logger.Debug("device has no paired devices")
		return nil
	}
	logger.Debug("device has paired device(s)")

	clock, chat := m.getLastClockWithRelatedChat()

	communitySettings, err := m.communitiesManager.GetCommunitySettingsByID(community.ID())
	if err != nil {
		return err
	}

	syncControlNode, err := m.communitiesManager.GetSyncControlNode(community.ID())
	if err != nil {
		return err
	}

	syncMessage, err := community.ToSyncInstallationCommunityProtobuf(clock, communitySettings, syncControlNode)
	if err != nil {
		return err
	}

	err = m.propagateSyncInstallationCommunityWithHRKeys(syncMessage, community)
	if err != nil {
		return err
	}

	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_INSTALLATION_COMMUNITY,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}
	logger.Debug("message dispatched")

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) SyncBookmark(ctx context.Context, bookmark *browsers.Bookmark, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	clock, chat := m.getLastClockWithRelatedChat()

	syncMessage := &protobuf.SyncBookmark{
		Clock:     clock,
		Url:       bookmark.URL,
		Name:      bookmark.Name,
		ImageUrl:  bookmark.ImageURL,
		Removed:   bookmark.Removed,
		DeletedAt: bookmark.DeletedAt,
	}
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_BOOKMARK,
		ResendAutomatically: true,
	}
	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) SyncEnsNamesWithDispatchMessage(ctx context.Context, usernameDetail *ensservice.UsernameDetail) error {
	return m.syncEnsUsernameDetail(ctx, usernameDetail, m.dispatchMessage)
}

func (m *Messenger) syncEnsUsernameDetails(ctx context.Context, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	ensNameDetails, err := m.getEnsUsernameDetails()
	if err != nil {
		return err
	}
	for _, d := range ensNameDetails {
		if err = m.syncEnsUsernameDetail(ctx, d, rawMessageHandler); err != nil {
			return err
		}
	}
	return nil
}

func (m *Messenger) saveEnsUsernameDetailProto(syncMessage *protobuf.SyncEnsUsernameDetail) (*ensservice.UsernameDetail, error) {
	ud := &ensservice.UsernameDetail{
		Username: syncMessage.Username,
		Clock:    syncMessage.Clock,
		ChainID:  syncMessage.ChainId,
		Removed:  syncMessage.Removed,
	}
	db := ensservice.NewEnsDatabase(m.database)
	err := db.SaveOrUpdateEnsUsername(ud)
	if err != nil {
		return nil, err
	}
	return ud, nil
}

func (m *Messenger) HandleSyncEnsUsernameDetail(state *ReceivedMessageState, syncMessage *protobuf.SyncEnsUsernameDetail, statusMessage *v1protocol.StatusMessage) error {
	ud, err := m.saveEnsUsernameDetailProto(syncMessage)
	if err != nil {
		return err
	}
	state.Response.AddEnsUsernameDetail(ud)
	return nil
}

func (m *Messenger) syncEnsUsernameDetail(ctx context.Context, usernameDetail *ensservice.UsernameDetail, rawMessageHandler RawMessageHandler) error {
	syncMessage := &protobuf.SyncEnsUsernameDetail{
		Clock:    usernameDetail.Clock,
		Username: usernameDetail.Username,
		ChainId:  usernameDetail.ChainID,
		Removed:  usernameDetail.Removed,
	}
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	_, chat := m.getLastClockWithRelatedChat()
	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ENS_USERNAME_DETAIL,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	return err
}

func (m *Messenger) syncAccountCustomizationColor(ctx context.Context, acc *multiaccounts.Account) error {
	if !m.hasPairedDevices() {
		return nil
	}

	_, chat := m.getLastClockWithRelatedChat()

	message := &protobuf.SyncAccountCustomizationColor{
		KeyUid:             acc.KeyUID,
		CustomizationColor: string(acc.CustomizationColor),
		UpdatedAt:          acc.CustomizationColorClock,
	}

	encodedMessage, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_ACCOUNT_CUSTOMIZATION_COLOR,
		ResendAutomatically: true,
	}

	_, err = m.dispatchMessage(ctx, rawMessage)
	return err
}

func (m *Messenger) SyncTrustedUser(ctx context.Context, publicKey string, ts verification.TrustStatus, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	clock, chat := m.getLastClockWithRelatedChat()

	syncMessage := &protobuf.SyncTrustedUser{
		Clock:  clock,
		Id:     publicKey,
		Status: protobuf.SyncTrustedUser_TrustStatus(ts),
	}
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_TRUSTED_USER,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) SyncVerificationRequest(ctx context.Context, vr *verification.Request, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	clock, chat := m.getLastClockWithRelatedChat()

	syncMessage := &protobuf.SyncVerificationRequest{
		Id:                 vr.ID,
		Clock:              clock,
		From:               vr.From,
		To:                 vr.To,
		Challenge:          vr.Challenge,
		Response:           vr.Response,
		RequestedAt:        vr.RequestedAt,
		RepliedAt:          vr.RepliedAt,
		VerificationStatus: protobuf.SyncVerificationRequest_VerificationStatus(vr.RequestStatus),
	}
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_VERIFICATION_REQUEST,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)
	if err != nil {
		return err
	}

	chat.LastClockValue = clock
	return m.saveChat(chat)
}

// RetrieveAll retrieves messages from all filters, processes them and returns a
// MessengerResponse to the client
func (m *Messenger) RetrieveAll() (*MessengerResponse, error) {
	chatWithMessages, err := m.transport.RetrieveRawAll()
	if err != nil {
		return nil, err
	}

	return m.handleRetrievedMessages(chatWithMessages, true, false)
}

func (m *Messenger) StartRetrieveMessagesLoop(tick time.Duration, cancel <-chan struct{}) {
	m.shutdownWaitGroup.Add(1)
	go func() {
		defer m.shutdownWaitGroup.Done()
		ticker := time.NewTicker(tick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.ProcessAllMessages()
			case <-cancel:
				return
			}
		}
	}()
}

func (m *Messenger) ProcessAllMessages() {
	response, err := m.RetrieveAll()
	if err != nil {
		m.logger.Error("failed to retrieve raw messages", zap.Error(err))
		return
	}
	m.PublishMessengerResponse(response)
}

func (m *Messenger) PublishMessengerResponse(response *MessengerResponse) {
	if response.IsEmpty() {
		return
	}

	notifications := response.Notifications()
	// Clear notifications as not used for now
	response.ClearNotifications()
	signal.SendNewMessages(response)
	localnotifications.PushMessages(notifications)
}

func (m *Messenger) GetStats() types.StatsSummary {
	return m.transport.GetStats()
}

func (m *Messenger) GetTransport() *transport.Transport {
	return m.transport
}

type CurrentMessageState struct {
	// Message is the protobuf message received
	Message *protobuf.ChatMessage
	// MessageID is the ID of the message
	MessageID string
	// WhisperTimestamp is the whisper timestamp of the message
	WhisperTimestamp uint64
	// Contact is the contact associated with the author of the message
	Contact *Contact
	// PublicKey is the public key of the author of the message
	PublicKey *ecdsa.PublicKey

	StatusMessage *v1protocol.StatusMessage
}

type ReceivedMessageState struct {
	// State on the message being processed
	CurrentMessageState *CurrentMessageState
	// AllChats in memory
	AllChats *chatMap
	// All contacts in memory
	AllContacts *contactMap
	// List of contacts modified
	ModifiedContacts *stringBoolMap
	// All installations in memory
	AllInstallations *installationMap
	// List of communities modified
	ModifiedInstallations *stringBoolMap
	// Map of existing messages
	ExistingMessagesMap map[string]bool
	// EmojiReactions is a list of emoji reactions for the current batch
	// indexed by from-message-id-emoji-type
	EmojiReactions map[string]*EmojiReaction
	// GroupChatInvitations is a list of invitation requests or rejections
	GroupChatInvitations map[string]*GroupChatInvitation
	// Response to the client
	Response           *MessengerResponse
	ResolvePrimaryName func(string) (string, error)
	// Timesource is a time source for clock values/timestamps.
	Timesource              common.TimeSource
	AllBookmarks            map[string]*browsers.Bookmark
	AllVerificationRequests []*verification.Request
	AllTrustStatus          map[string]verification.TrustStatus
}

// addNewMessageNotification takes a common.Message and generates a new NotificationBody and appends it to the
// []Response.Notifications if the message is m.New
func (r *ReceivedMessageState) addNewMessageNotification(publicKey ecdsa.PublicKey, m *common.Message, responseTo *common.Message, profilePicturesVisibility int) error {
	if !m.New {
		return nil
	}

	pubKey, err := m.GetSenderPubKey()
	if err != nil {
		return err
	}
	contactID := contactIDFromPublicKey(pubKey)

	chat, ok := r.AllChats.Load(m.LocalChatID)
	if !ok {
		return fmt.Errorf("chat ID '%s' not present", m.LocalChatID)
	}

	contact, ok := r.AllContacts.Load(contactID)
	if !ok {
		return fmt.Errorf("contact ID '%s' not present", contactID)
	}

	if !chat.Muted {
		if showMessageNotification(publicKey, m, chat, responseTo) {
			notification, err := NewMessageNotification(m.ID, m, chat, contact, r.ResolvePrimaryName, profilePicturesVisibility)
			if err != nil {
				return err
			}
			r.Response.AddNotification(notification)
		}
	}

	return nil
}

// updateExistingActivityCenterNotification updates AC notification if it exists and hasn't been read yet
func (r *ReceivedMessageState) updateExistingActivityCenterNotification(publicKey ecdsa.PublicKey, m *Messenger, message *common.Message, responseTo *common.Message) error {
	notification, err := m.persistence.GetActivityCenterNotificationByID(types.FromHex(message.ID))
	if err != nil {
		return err
	}

	if notification == nil || notification.Read {
		return nil
	}

	notification.Message = message
	notification.ReplyMessage = responseTo
	notification.UpdatedAt = m.GetCurrentTimeInMillis()

	err = m.addActivityCenterNotification(r.Response, notification, nil)
	if err != nil {
		return err
	}

	return nil
}

// addNewActivityCenterNotification takes a common.Message and generates a new ActivityCenterNotification and appends it to the
// []Response.ActivityCenterNotifications if the message is m.New
func (r *ReceivedMessageState) addNewActivityCenterNotification(publicKey ecdsa.PublicKey, m *Messenger, message *common.Message, responseTo *common.Message) error {
	if !message.New {
		return nil
	}

	chat, ok := r.AllChats.Load(message.LocalChatID)
	if !ok {
		return fmt.Errorf("chat ID '%s' not present", message.LocalChatID)
	}

	isNotification, notificationType := showMentionOrReplyActivityCenterNotification(publicKey, message, chat, responseTo)
	if !isNotification {
		return nil
	}

	if chat.CommunityChat() {
		joinedClock, err := m.communitiesManager.GetCommunityRequestToJoinClock(&publicKey, message.CommunityID)
		if err != nil {
			return err
		}

		// Ignore mentions & replies in community before joining
		if message.Clock < joinedClock {
			return nil
		}
	}

	// Use albumId as notificationId to prevent multiple notifications
	// for same message with multiple images
	var notificationID string

	image := message.GetImage()
	var albumMessages = []*common.Message{}
	if image != nil && image.GetAlbumId() != "" {
		notificationID = image.GetAlbumId()
		album, err := m.persistence.albumMessages(message.LocalChatID, image.AlbumId)
		if err != nil {
			return err
		}
		if m.httpServer != nil {
			for _, msg := range album {
				err = m.prepareMessage(msg, m.httpServer)

				if err != nil {
					return err
				}
			}
		}

		albumMessages = album
	} else {
		notificationID = message.ID
	}

	notification := &ActivityCenterNotification{
		ID:            types.FromHex(notificationID),
		Name:          chat.Name,
		Message:       message,
		ReplyMessage:  responseTo,
		Type:          notificationType,
		Timestamp:     message.WhisperTimestamp,
		ChatID:        chat.ID,
		CommunityID:   chat.CommunityID,
		Author:        message.From,
		UpdatedAt:     m.GetCurrentTimeInMillis(),
		AlbumMessages: albumMessages,
		Read:          message.Seen,
	}

	return m.addActivityCenterNotification(r.Response, notification, nil)
}

func (m *Messenger) buildMessageState() *ReceivedMessageState {
	return &ReceivedMessageState{
		AllChats:              m.allChats,
		AllContacts:           m.allContacts,
		ModifiedContacts:      new(stringBoolMap),
		AllInstallations:      m.allInstallations,
		ModifiedInstallations: m.modifiedInstallations,
		ExistingMessagesMap:   make(map[string]bool),
		EmojiReactions:        make(map[string]*EmojiReaction),
		GroupChatInvitations:  make(map[string]*GroupChatInvitation),
		Response:              &MessengerResponse{},
		Timesource:            m.getTimesource(),
		ResolvePrimaryName:    m.ResolvePrimaryName,
		AllBookmarks:          make(map[string]*browsers.Bookmark),
		AllTrustStatus:        make(map[string]verification.TrustStatus),
	}
}

func (m *Messenger) outputToCSV(timestamp uint32, messageID types.HexBytes, from string, topic types.TopicType, chatID string, msgType protobuf.ApplicationMetadataMessage_Type, parsedMessage interface{}) {
	if !m.outputCSV {
		return
	}

	msgJSON, err := json.Marshal(parsedMessage)
	if err != nil {
		m.logger.Error("could not marshall message", zap.Error(err))
		return
	}

	line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\n", timestamp, messageID.String(), from, topic.String(), chatID, msgType, msgJSON)
	_, err = m.csvFile.Write([]byte(line))
	if err != nil {
		m.logger.Error("could not write to csv", zap.Error(err))
		return
	}
}

func (m *Messenger) shouldSkipDuplicate(messageType protobuf.ApplicationMetadataMessage_Type) bool {
	// Permit re-processing of ApplicationMetadataMessage_COMMUNITY_DESCRIPTION messages,
	// as they may be queued pending receipt of decryption keys.
	allowedDuplicateTypes := map[protobuf.ApplicationMetadataMessage_Type]struct{}{
		protobuf.ApplicationMetadataMessage_COMMUNITY_DESCRIPTION: struct{}{},
	}
	if _, isAllowedDuplicate := allowedDuplicateTypes[messageType]; isAllowedDuplicate {
		return false
	}

	return true
}

func (m *Messenger) handleImportedMessages(messagesToHandle map[transport.Filter][]*types.Message) error {

	messageState := m.buildMessageState()

	logger := m.logger.With(zap.String("site", "handleImportedMessages"))

	for filter, messages := range messagesToHandle {
		for _, shhMessage := range messages {

			handleMessageResponse, err := m.sender.HandleMessages(shhMessage)
			if err != nil {
				logger.Info("failed to decode messages", zap.Error(err))
				continue
			}
			statusMessages := handleMessageResponse.StatusMessages

			for _, msg := range statusMessages {
				logger := logger.With(zap.String("message-id", msg.TransportLayer.Message.ThirdPartyID))
				logger.Debug("processing message")

				publicKey := msg.SigPubKey()
				senderID := contactIDFromPublicKey(publicKey)

				if len(msg.EncryptionLayer.HashRatchetInfo) != 0 {
					err := m.communitiesManager.NewHashRatchetKeys(msg.EncryptionLayer.HashRatchetInfo)
					if err != nil {
						m.logger.Warn("failed to invalidate communities description cache", zap.Error(err))
					}

				}
				// Don't process duplicates
				messageID := msg.TransportLayer.Message.ThirdPartyID
				exists, err := m.messageExists(messageID, messageState.ExistingMessagesMap)
				if err != nil {
					logger.Warn("failed to check message exists", zap.Error(err))
				}
				if exists && m.shouldSkipDuplicate(msg.ApplicationLayer.Type) {
					logger.Debug("skipping duplicate", zap.String("messageID", messageID))
					continue
				}

				var contact *Contact
				if c, ok := messageState.AllContacts.Load(senderID); ok {
					contact = c
				} else {
					c, err := buildContact(senderID, publicKey)
					if err != nil {
						logger.Info("failed to build contact", zap.Error(err))
						continue
					}
					contact = c
					messageState.AllContacts.Store(senderID, contact)
				}
				messageState.CurrentMessageState = &CurrentMessageState{
					MessageID:        messageID,
					WhisperTimestamp: uint64(msg.TransportLayer.Message.Timestamp) * 1000,
					Contact:          contact,
					PublicKey:        publicKey,
					StatusMessage:    msg,
				}

				if msg.ApplicationLayer.Payload != nil {

					logger.Debug("Handling parsed message")

					switch msg.ApplicationLayer.Type {

					case protobuf.ApplicationMetadataMessage_CHAT_MESSAGE:
						err = m.handleChatMessageProtobuf(messageState, msg.ApplicationLayer.Payload, msg, filter, true)
						if err != nil {
							logger.Warn("failed to handle ChatMessage", zap.Error(err))
							continue
						}

					case protobuf.ApplicationMetadataMessage_PIN_MESSAGE:
						err = m.handlePinMessageProtobuf(messageState, msg.ApplicationLayer.Payload, msg, filter, true)
						if err != nil {
							logger.Warn("failed to handle PinMessage", zap.Error(err))
						}
					}
				}
			}
		}
	}

	importMessageAuthors := messageState.Response.DiscordMessageAuthors()
	if len(importMessageAuthors) > 0 {
		err := m.persistence.SaveDiscordMessageAuthors(importMessageAuthors)
		if err != nil {
			return err
		}
	}

	importMessagesToSave := messageState.Response.DiscordMessages()
	if len(importMessagesToSave) > 0 {
		m.communitiesManager.LogStdout(fmt.Sprintf("saving %d discord messages", len(importMessagesToSave)))
		m.handleImportMessagesMutex.Lock()
		err := m.persistence.SaveDiscordMessages(importMessagesToSave)
		if err != nil {
			m.communitiesManager.LogStdout("failed to save discord messages", zap.Error(err))
			m.handleImportMessagesMutex.Unlock()
			return err
		}
		m.handleImportMessagesMutex.Unlock()
	}

	messageAttachmentsToSave := messageState.Response.DiscordMessageAttachments()
	if len(messageAttachmentsToSave) > 0 {
		m.communitiesManager.LogStdout(fmt.Sprintf("saving %d discord message attachments", len(messageAttachmentsToSave)))
		m.handleImportMessagesMutex.Lock()
		err := m.persistence.SaveDiscordMessageAttachments(messageAttachmentsToSave)
		if err != nil {
			m.communitiesManager.LogStdout("failed to save discord message attachments", zap.Error(err))
			m.handleImportMessagesMutex.Unlock()
			return err
		}
		m.handleImportMessagesMutex.Unlock()
	}

	messagesToSave := messageState.Response.Messages()
	if len(messagesToSave) > 0 {
		m.communitiesManager.LogStdout(fmt.Sprintf("saving %d app messages", len(messagesToSave)))
		m.handleMessagesMutex.Lock()
		err := m.SaveMessages(messagesToSave)
		if err != nil {
			m.handleMessagesMutex.Unlock()
			return err
		}
		m.handleMessagesMutex.Unlock()
	}

	// Save chats if they were modified
	if len(messageState.Response.chats) > 0 {
		err := m.saveChats(messageState.Response.Chats())
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Messenger) handleRetrievedMessages(chatWithMessages map[transport.Filter][]*types.Message, storeWakuMessages bool, fromArchive bool) (*MessengerResponse, error) {

	m.handleMessagesMutex.Lock()
	defer m.handleMessagesMutex.Unlock()

	messageState := m.buildMessageState()

	logger := m.logger.With(zap.String("site", "RetrieveAll"))

	controlledCommunitiesChatIDs, err := m.communitiesManager.GetOwnedCommunitiesChatIDs()
	if err != nil {
		logger.Info("failed to retrieve admin communities", zap.Error(err))
	}

	iterator := m.retrievedMessagesIteratorFactory(chatWithMessages)
	for iterator.HasNext() {
		filter, messages := iterator.Next()

		var processedMessages []string
		for _, shhMessage := range messages {
			logger := logger.With(zap.String("hash", types.EncodeHex(shhMessage.Hash)))
			// Indicates tha all messages in the batch have been processed correctly
			allMessagesProcessed := true

			if controlledCommunitiesChatIDs[filter.ChatID] && storeWakuMessages {
				logger.Debug("storing waku message")
				err := m.communitiesManager.StoreWakuMessage(shhMessage)
				if err != nil {
					logger.Warn("failed to store waku message", zap.Error(err))
				}
			}

			handleMessagesResponse, err := m.sender.HandleMessages(shhMessage)
			if err != nil {
				if m.telemetryClient != nil {
					go m.telemetryClient.UpdateEnvelopeProcessingError(shhMessage, err)
				}
				logger.Info("failed to decode messages", zap.Error(err))
				continue
			}

			if handleMessagesResponse == nil {
				continue
			}

			statusMessages := handleMessagesResponse.StatusMessages

			if m.telemetryClient != nil {
				go m.telemetryClient.PushReceivedMessages(filter, shhMessage, statusMessages)
			}

			err = m.handleDatasyncMetadata(handleMessagesResponse)
			if err != nil {
				m.logger.Warn("failed to handle datasync metadata", zap.Error(err))
			}

			logger.Debug("processing messages further", zap.Int("count", len(statusMessages)))

			for _, msg := range statusMessages {
				logger := logger.With(zap.String("message-id", msg.ApplicationLayer.ID.String()))
				logger.Info("processing message")
				publicKey := msg.SigPubKey()

				m.handleInstallations(msg.EncryptionLayer.Installations)
				err := m.handleSharedSecrets(msg.EncryptionLayer.SharedSecrets)
				if err != nil {
					// log and continue, non-critical error
					logger.Warn("failed to handle shared secrets")
				}

				senderID := contactIDFromPublicKey(publicKey)
				ownID := contactIDFromPublicKey(m.IdentityPublicKey())
				m.logger.Info("processing message", zap.Any("type", msg.ApplicationLayer.Type), zap.String("senderID", senderID))

				if senderID == ownID {
					// Skip own messages of certain types
					if msg.ApplicationLayer.Type == protobuf.ApplicationMetadataMessage_CONTACT_CODE_ADVERTISEMENT {
						continue
					}
				}

				contact, contactFound := messageState.AllContacts.Load(senderID)

				// Check for messages from blocked users
				if contactFound && contact.Blocked {
					continue
				}

				// Don't process duplicates
				messageID := types.EncodeHex(msg.ApplicationLayer.ID)
				exists, err := m.messageExists(messageID, messageState.ExistingMessagesMap)
				if err != nil {
					logger.Warn("failed to check message exists", zap.Error(err))
				}
				if exists && m.shouldSkipDuplicate(msg.ApplicationLayer.Type) {
					logger.Debug("skipping duplicate", zap.String("messageID", messageID))
					continue
				}

				if !contactFound {
					c, err := buildContact(senderID, publicKey)
					if err != nil {
						logger.Info("failed to build contact", zap.Error(err))
						allMessagesProcessed = false
						continue
					}
					contact = c
					if msg.ApplicationLayer.Type != protobuf.ApplicationMetadataMessage_PUSH_NOTIFICATION_QUERY {
						messageState.AllContacts.Store(senderID, contact)
					}
				}
				messageState.CurrentMessageState = &CurrentMessageState{
					MessageID:        messageID,
					WhisperTimestamp: uint64(msg.TransportLayer.Message.Timestamp) * 1000,
					Contact:          contact,
					PublicKey:        publicKey,
					StatusMessage:    msg,
				}

				if msg.ApplicationLayer.Payload != nil {

					err := m.dispatchToHandler(messageState, msg.ApplicationLayer.Payload, msg, filter, fromArchive)
					if err != nil {
						allMessagesProcessed = false
						logger.Warn("failed to process protobuf", zap.Error(err))
						if m.unhandledMessagesTracker != nil {
							m.unhandledMessagesTracker(msg, err)
						}
						continue
					}
					logger.Debug("Handled parsed message")

				} else {
					logger.Debug("parsed message is nil")
				}
			}

			m.processCommunityChanges(messageState)

			// NOTE: for now we confirm messages as processed regardless whether we
			// actually processed them, this is because we need to differentiate
			// from messages that we want to retry to process and messages that
			// are never going to be processed
			m.transport.MarkP2PMessageAsProcessed(gethcommon.BytesToHash(shhMessage.Hash))

			if allMessagesProcessed {
				processedMessages = append(processedMessages, types.EncodeHex(shhMessage.Hash))
			}
		}

		if len(processedMessages) != 0 {
			if err := m.transport.ConfirmMessagesProcessed(processedMessages, m.getTimesource().GetCurrentTime()); err != nil {
				logger.Warn("failed to confirm processed messages", zap.Error(err))
			}
		}
	}

	return m.saveDataAndPrepareResponse(messageState)
}

func (m *Messenger) saveDataAndPrepareResponse(messageState *ReceivedMessageState) (*MessengerResponse, error) {
	var err error
	var contactsToSave []*Contact
	messageState.ModifiedContacts.Range(func(id string, value bool) (shouldContinue bool) {
		contact, ok := messageState.AllContacts.Load(id)
		if ok {
			contactsToSave = append(contactsToSave, contact)
			messageState.Response.AddContact(contact)
		}
		return true
	})

	// Hydrate chat alias and identicon
	for id := range messageState.Response.chats {
		chat, _ := messageState.AllChats.Load(id)
		if chat == nil {
			continue
		}
		if chat.OneToOne() {
			contact, ok := m.allContacts.Load(chat.ID)
			if ok {
				chat.Alias = contact.Alias
				chat.Identicon = contact.Identicon
			}
		}

		messageState.Response.AddChat(chat)
	}

	messageState.ModifiedInstallations.Range(func(id string, value bool) (shouldContinue bool) {
		installation, _ := messageState.AllInstallations.Load(id)
		messageState.Response.Installations = append(messageState.Response.Installations, installation)
		if installation.InstallationMetadata != nil {
			err = m.setInstallationMetadata(id, installation.InstallationMetadata)
			if err != nil {
				return false
			}
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	if len(messageState.Response.chats) > 0 {
		err = m.saveChats(messageState.Response.Chats())
		if err != nil {
			return nil, err
		}
	}

	messagesToSave := messageState.Response.Messages()
	if len(messagesToSave) > 0 {
		err = m.SaveMessages(messagesToSave)
		if err != nil {
			return nil, err
		}
	}

	for _, emojiReaction := range messageState.EmojiReactions {
		messageState.Response.AddEmojiReaction(emojiReaction)
	}

	for _, groupChatInvitation := range messageState.GroupChatInvitations {
		messageState.Response.Invitations = append(messageState.Response.Invitations, groupChatInvitation)
	}

	if len(contactsToSave) > 0 {
		err = m.persistence.SaveContacts(contactsToSave)
		if err != nil {
			return nil, err
		}
	}

	newMessagesIds := map[string]struct{}{}
	for _, message := range messagesToSave {
		if message.New {
			newMessagesIds[message.ID] = struct{}{}
		}
	}

	messagesWithResponses, err := m.pullMessagesAndResponsesFromDB(messagesToSave)
	if err != nil {
		return nil, err
	}
	messagesByID := map[string]*common.Message{}
	for _, message := range messagesWithResponses {
		messagesByID[message.ID] = message
	}
	messageState.Response.SetMessages(messagesWithResponses)

	notificationsEnabled, err := m.settings.GetNotificationsEnabled()
	if err != nil {
		return nil, err
	}

	profilePicturesVisibility, err := m.settings.GetProfilePicturesVisibility()
	if err != nil {
		return nil, err
	}

	err = m.prepareMessages(messageState.Response.messages)
	if err != nil {
		return nil, err
	}

	for _, message := range messageState.Response.messages {
		if _, ok := newMessagesIds[message.ID]; ok {
			message.New = true

			if notificationsEnabled {
				// Create notification body to be eventually passed to `localnotifications.SendMessageNotifications()`
				if err = messageState.addNewMessageNotification(m.identity.PublicKey, message, messagesByID[message.ResponseTo], profilePicturesVisibility); err != nil {
					return nil, err
				}
			}

			// Create activity center notification body to be eventually passed to `activitycenter.SendActivityCenterNotifications()`
			if err = messageState.addNewActivityCenterNotification(m.identity.PublicKey, m, message, messagesByID[message.ResponseTo]); err != nil {
				return nil, err
			}
		}
	}

	// Reset installations
	m.modifiedInstallations = new(stringBoolMap)

	if len(messageState.AllBookmarks) > 0 {
		bookmarks, err := m.storeSyncBookmarks(messageState.AllBookmarks)
		if err != nil {
			return nil, err
		}
		messageState.Response.AddBookmarks(bookmarks)
	}

	if len(messageState.AllVerificationRequests) > 0 {
		for _, vr := range messageState.AllVerificationRequests {
			messageState.Response.AddVerificationRequest(vr)
		}
	}

	if len(messageState.AllTrustStatus) > 0 {
		messageState.Response.AddTrustStatuses(messageState.AllTrustStatus)
	}

	// Hydrate pinned messages
	for _, pinnedMessage := range messageState.Response.PinMessages() {
		if pinnedMessage.Pinned {
			pinnedMessage.Message = &common.PinnedMessage{
				Message:  messageState.Response.GetMessage(pinnedMessage.MessageId),
				PinnedBy: pinnedMessage.From,
				PinnedAt: pinnedMessage.Clock,
			}
		}
	}

	return messageState.Response, nil
}

func (m *Messenger) storeSyncBookmarks(bookmarkMap map[string]*browsers.Bookmark) ([]*browsers.Bookmark, error) {
	var bookmarks []*browsers.Bookmark
	for _, bookmark := range bookmarkMap {
		bookmarks = append(bookmarks, bookmark)
	}
	return m.browserDatabase.StoreSyncBookmarks(bookmarks)
}

func (m *Messenger) MessageByID(id string) (*common.Message, error) {
	return m.persistence.MessageByID(id)
}

func (m *Messenger) MessagesExist(ids []string) (map[string]bool, error) {
	return m.persistence.MessagesExist(ids)
}

func (m *Messenger) FirstUnseenMessageID(chatID string) (string, error) {
	return m.persistence.FirstUnseenMessageID(chatID)
}

func (m *Messenger) latestIncomingMessageClock(chatID string) (uint64, error) {
	return m.persistence.latestIncomingMessageClock(chatID)
}

func (m *Messenger) MessageByChatID(chatID, cursor string, limit int) ([]*common.Message, string, error) {
	chat, err := m.persistence.Chat(chatID)
	if err != nil {
		return nil, "", err
	}

	if chat == nil {
		return nil, "", ErrChatNotFound
	}

	var msgs []*common.Message
	var nextCursor string

	if chat.Timeline() {
		var chatIDs = []string{"@" + contactIDFromPublicKey(&m.identity.PublicKey)}
		m.allContacts.Range(func(contactID string, contact *Contact) (shouldContinue bool) {
			if contact.added() {
				chatIDs = append(chatIDs, "@"+contact.ID)
			}
			return true
		})
		msgs, nextCursor, err = m.persistence.MessageByChatIDs(chatIDs, cursor, limit)
		if err != nil {
			return nil, "", err
		}
	} else {
		msgs, nextCursor, err = m.persistence.MessageByChatID(chatID, cursor, limit)
		if err != nil {
			return nil, "", err
		}

	}

	if m.httpServer != nil {
		for _, msg := range msgs {
			err = m.prepareMessage(msg, m.httpServer)
			if err != nil {
				return nil, "", err
			}
		}
	}

	return msgs, nextCursor, nil
}

func (m *Messenger) prepareMessages(messages map[string]*common.Message) error {
	if m.httpServer == nil {
		return nil
	}
	for idx := range messages {
		err := m.prepareMessage(messages[idx], m.httpServer)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractQuotedImages(messages []*common.Message, s *server.MediaServer) []string {
	var quotedImages []string

	for _, message := range messages {
		if message.ChatMessage != nil && message.ChatMessage.ContentType == protobuf.ChatMessage_IMAGE {
			quotedImages = append(quotedImages, s.MakeImageURL(message.ID))
		}
	}
	return quotedImages
}

func (m *Messenger) prepareTokenData(tokenData *ActivityTokenData, s *server.MediaServer) error {
	if tokenData.TokenType == int(protobuf.CommunityTokenType_ERC721) {
		tokenData.ImageURL = s.MakeWalletCollectibleImagesURL(tokenData.CollectibleID)
	} else if tokenData.TokenType == int(protobuf.CommunityTokenType_ERC20) {
		tokenData.ImageURL = s.MakeCommunityTokenImagesURL(tokenData.CommunityID, tokenData.ChainID, tokenData.Symbol)
	}
	return nil
}

func (m *Messenger) prepareMessage(msg *common.Message, s *server.MediaServer) error {
	if msg.QuotedMessage != nil && msg.QuotedMessage.ContentType == int64(protobuf.ChatMessage_IMAGE) {
		msg.QuotedMessage.ImageLocalURL = s.MakeImageURL(msg.QuotedMessage.ID)

		quotedMessage, err := m.MessageByID(msg.QuotedMessage.ID)
		if err != nil {
			return err
		}
		if quotedMessage == nil {
			return errors.New("message not found")
		}

		if quotedMessage.ChatMessage != nil {
			image := quotedMessage.ChatMessage.GetImage()
			albumID := quotedMessage.ChatMessage.GetImage().AlbumId

			if image != nil && image.GetAlbumId() != "" {
				albumMessages, err := m.persistence.albumMessages(quotedMessage.LocalChatID, albumID)
				if err != nil {
					return err
				}

				quotedImages := extractQuotedImages(albumMessages, s)
				quotedImagesJSON, err := json.Marshal(quotedImages)
				if err != nil {
					return err
				}

				msg.QuotedMessage.AlbumImages = quotedImagesJSON
			}
		}
	}
	if msg.QuotedMessage != nil && msg.QuotedMessage.ContentType == int64(protobuf.ChatMessage_AUDIO) {
		msg.QuotedMessage.AudioLocalURL = s.MakeAudioURL(msg.QuotedMessage.ID)
	}
	if msg.QuotedMessage != nil && msg.QuotedMessage.ContentType == int64(protobuf.ChatMessage_STICKER) {
		msg.QuotedMessage.HasSticker = true
	}
	if msg.QuotedMessage != nil && msg.QuotedMessage.ContentType == int64(protobuf.ChatMessage_DISCORD_MESSAGE) {
		dm := msg.QuotedMessage.DiscordMessage
		exists, err := m.persistence.HasDiscordMessageAuthorImagePayload(dm.Author.Id)
		if err != nil {
			return err
		}

		if exists {
			msg.QuotedMessage.DiscordMessage.Author.LocalUrl = s.MakeDiscordAuthorAvatarURL(dm.Author.Id)
		}
	}

	if msg.ContentType == protobuf.ChatMessage_IMAGE {
		msg.ImageLocalURL = s.MakeImageURL(msg.ID)
	}

	if msg.ContentType == protobuf.ChatMessage_DISCORD_MESSAGE {

		dm := msg.GetDiscordMessage()
		exists, err := m.persistence.HasDiscordMessageAuthorImagePayload(dm.Author.Id)
		if err != nil {
			return err
		}

		if exists {
			dm.Author.LocalUrl = s.MakeDiscordAuthorAvatarURL(dm.Author.Id)
		}

		for idx, attachment := range dm.Attachments {
			if strings.Contains(attachment.ContentType, "image") {
				hasPayload, err := m.persistence.HasDiscordMessageAttachmentPayload(attachment.Id, dm.Id)
				if err != nil {
					m.logger.Error("failed to check if message attachment exist", zap.Error(err))
					continue
				}
				if hasPayload {
					localURL := s.MakeDiscordAttachmentURL(dm.Id, attachment.Id)
					dm.Attachments[idx].LocalUrl = localURL
				}
			}
		}
		msg.Payload = &protobuf.ChatMessage_DiscordMessage{
			DiscordMessage: dm,
		}
	}
	if msg.ContentType == protobuf.ChatMessage_AUDIO {
		msg.AudioLocalURL = s.MakeAudioURL(msg.ID)
	}
	if msg.ContentType == protobuf.ChatMessage_STICKER {
		msg.StickerLocalURL = s.MakeStickerURL(msg.GetSticker().Hash)
	}
	msg.LinkPreviews = msg.ConvertFromProtoToLinkPreviews(s.MakeLinkPreviewThumbnailURL, s.MakeLinkPreviewFaviconURL)
	msg.StatusLinkPreviews = msg.ConvertFromProtoToStatusLinkPreviews(s.MakeStatusLinkPreviewThumbnailURL)

	return nil
}

func (m *Messenger) AllMessageByChatIDWhichMatchTerm(chatID string, searchTerm string, caseSensitive bool) ([]*common.Message, error) {
	_, err := m.persistence.Chat(chatID)
	if err != nil {
		return nil, err
	}

	return m.persistence.AllMessageByChatIDWhichMatchTerm(chatID, searchTerm, caseSensitive)
}

func (m *Messenger) AllMessagesFromChatsAndCommunitiesWhichMatchTerm(communityIds []string, chatIds []string, searchTerm string, caseSensitive bool) ([]*common.Message, error) {
	return m.persistence.AllMessagesFromChatsAndCommunitiesWhichMatchTerm(communityIds, chatIds, searchTerm, caseSensitive)
}

func (m *Messenger) SaveMessages(messages []*common.Message) error {
	return m.persistence.SaveMessages(messages)
}

func (m *Messenger) DeleteMessage(id string) error {
	return m.persistence.DeleteMessage(id)
}

func (m *Messenger) DeleteMessagesByChatID(id string) error {
	return m.persistence.DeleteMessagesByChatID(id)
}

func (m *Messenger) markMessageAsUnreadImpl(chatID string, messageID string) (uint64, uint64, error) {
	count, countWithMentions, err := m.persistence.MarkMessageAsUnread(chatID, messageID)

	if err != nil {
		return 0, 0, err
	}

	chat, err := m.persistence.Chat(chatID)
	if err != nil {
		return 0, 0, err
	}
	m.allChats.Store(chatID, chat)
	return count, countWithMentions, nil
}

func (m *Messenger) MarkMessageAsUnread(chatID string, messageID string) (*MessengerResponse, error) {
	count, countWithMentions, err := m.markMessageAsUnreadImpl(chatID, messageID)
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}
	response.AddSeenAndUnseenMessages(&SeenUnseenMessages{
		ChatID:            chatID,
		Count:             count,
		CountWithMentions: countWithMentions,
		Seen:              false,
	})

	ids, err := m.persistence.GetMessageIdsWithGreaterTimestamp(chatID, messageID)
	if err != nil {
		return nil, err
	}

	hexBytesIds := []types.HexBytes{}
	for _, id := range ids {
		hexBytesIds = append(hexBytesIds, types.FromHex(id))
	}

	updatedAt := m.GetCurrentTimeInMillis()
	notifications, err := m.persistence.MarkActivityCenterNotificationsUnread(hexBytesIds, updatedAt)
	if err != nil {
		return nil, err
	}

	response.AddActivityCenterNotifications(notifications)

	return response, nil
}

// MarkMessagesSeen marks messages with `ids` as seen in the chat `chatID`.
// It returns the number of affected messages or error. If there is an error,
// the number of affected messages is always zero.
func (m *Messenger) markMessagesSeenImpl(chatID string, ids []string) (uint64, uint64, *Chat, error) {
	count, countWithMentions, err := m.persistence.MarkMessagesSeen(chatID, ids)
	if err != nil {
		return 0, 0, nil, err
	}
	chat, err := m.persistence.Chat(chatID)
	if err != nil {
		return 0, 0, nil, err
	}
	m.allChats.Store(chatID, chat)
	return count, countWithMentions, chat, nil
}

// Deprecated: Use MarkMessagesRead instead
func (m *Messenger) MarkMessagesSeen(chatID string, ids []string) (uint64, uint64, []*ActivityCenterNotification, error) {
	count, countWithMentions, _, err := m.markMessagesSeenImpl(chatID, ids)
	if err != nil {
		return 0, 0, nil, err
	}

	hexBytesIds := []types.HexBytes{}
	for _, id := range ids {
		hexBytesIds = append(hexBytesIds, types.FromHex(id))
	}

	// Mark notifications as read in the database
	updatedAt := m.GetCurrentTimeInMillis()
	err = m.persistence.MarkActivityCenterNotificationsRead(hexBytesIds, updatedAt)
	if err != nil {
		return 0, 0, nil, err
	}

	notifications, err := m.persistence.GetActivityCenterNotificationsByID(hexBytesIds)
	if err != nil {
		return 0, 0, nil, err
	}

	return count, countWithMentions, notifications, nil
}

func (m *Messenger) MarkMessagesRead(chatID string, ids []string) (*MessengerResponse, error) {
	count, countWithMentions, _, err := m.markMessagesSeenImpl(chatID, ids)
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}
	response.AddSeenAndUnseenMessages(&SeenUnseenMessages{
		ChatID:            chatID,
		Count:             count,
		CountWithMentions: countWithMentions,
		Seen:              true,
	})

	hexBytesIds := []types.HexBytes{}
	for _, id := range ids {
		hexBytesIds = append(hexBytesIds, types.FromHex(id))
	}

	// Mark notifications as read in the database
	updatedAt := m.GetCurrentTimeInMillis()
	err = m.persistence.MarkActivityCenterNotificationsRead(hexBytesIds, updatedAt)
	if err != nil {
		return nil, err
	}

	notifications, err := m.persistence.GetActivityCenterNotificationsByID(hexBytesIds)
	if err != nil {
		return nil, err
	}

	response.AddActivityCenterNotifications(notifications)

	return response, nil
}

func (m *Messenger) syncChatMessagesRead(ctx context.Context, chatID string, clock uint64, rawMessageHandler RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	_, chat := m.getLastClockWithRelatedChat()

	syncMessage := &protobuf.SyncChatMessagesRead{
		Clock: clock,
		Id:    chatID,
	}
	encodedMessage, err := proto.Marshal(syncMessage)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_CHAT_MESSAGES_READ,
		ResendAutomatically: true,
	}

	_, err = rawMessageHandler(ctx, rawMessage)

	return err
}

func (m *Messenger) markAllRead(chatID string, clock uint64, shouldBeSynced bool) error {
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return errors.New("chat not found")
	}

	_, _, err := m.persistence.MarkAllRead(chatID, clock)
	if err != nil {
		return err
	}

	if shouldBeSynced {
		err := m.syncChatMessagesRead(context.Background(), chatID, clock, m.dispatchMessage)
		if err != nil {
			return err
		}
	}

	chat.ReadMessagesAtClockValue = clock
	chat.Highlight = false

	chat.UnviewedMessagesCount = 0
	chat.UnviewedMentionsCount = 0

	if chat.LastMessage != nil {
		chat.LastMessage.Seen = true
	}

	// TODO(samyoul) remove storing of an updated reference pointer?
	m.allChats.Store(chat.ID, chat)
	return m.persistence.SaveChats([]*Chat{chat})
}

func (m *Messenger) MarkAllRead(ctx context.Context, chatID string) (*MessengerResponse, error) {
	response := &MessengerResponse{}

	notifications, err := m.DismissAllActivityCenterNotificationsFromChatID(ctx, chatID, m.GetCurrentTimeInMillis())
	if err != nil {
		return nil, err
	}
	response.AddActivityCenterNotifications(notifications)

	clock, _ := m.latestIncomingMessageClock(chatID)

	if clock == 0 {
		chat, ok := m.allChats.Load(chatID)
		if !ok {
			return nil, errors.New("chat not found")
		}
		clock, _ = chat.NextClockAndTimestamp(m.getTimesource())
	}

	err = m.markAllRead(chatID, clock, true)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (m *Messenger) MarkAllReadInCommunity(ctx context.Context, communityID string) (*MessengerResponse, error) {
	response := &MessengerResponse{}

	notifications, err := m.DismissAllActivityCenterNotificationsFromCommunity(ctx, communityID, m.GetCurrentTimeInMillis())
	if err != nil {
		return nil, err
	}
	response.AddActivityCenterNotifications(notifications)

	chatIDs, err := m.persistence.AllChatIDsByCommunity(nil, communityID)
	if err != nil {
		return nil, err
	}

	err = m.persistence.MarkAllReadMultiple(chatIDs)
	if err != nil {
		return nil, err
	}

	for _, chatID := range chatIDs {
		chat, ok := m.allChats.Load(chatID)

		if ok {
			chat.UnviewedMessagesCount = 0
			chat.UnviewedMentionsCount = 0
			m.allChats.Store(chat.ID, chat)
			response.AddChat(chat)
		} else {
			err = errors.New(fmt.Sprintf("chat with chatID %s not found", chatID))
		}
	}
	return response, err
}

// MuteChat signals to the messenger that we don't want to be notified
// on new messages from this chat
func (m *Messenger) MuteChat(request *requests.MuteChat) (time.Time, error) {
	chat, ok := m.allChats.Load(request.ChatID)
	if !ok {
		// Only one to one chan be muted when it's not in the database
		publicKey, err := common.HexToPubkey(request.ChatID)
		if err != nil {
			return time.Time{}, err
		}

		// Create a one to one chat and set active to false
		chat = CreateOneToOneChat(request.ChatID, publicKey, m.getTimesource())
		chat.Active = false
		err = m.initChatSyncFields(chat)
		if err != nil {
			return time.Time{}, err
		}
		err = m.saveChat(chat)
		if err != nil {
			return time.Time{}, err
		}
	}

	var contact *Contact
	if chat.OneToOne() {
		contact, _ = m.allContacts.Load(request.ChatID)
	}

	var MuteTill time.Time

	switch request.MutedType {
	case MuteTill1Min:
		MuteTill = time.Now().Add(MuteFor1MinDuration)
	case MuteFor15Min:
		MuteTill = time.Now().Add(MuteFor15MinsDuration)
	case MuteFor1Hr:
		MuteTill = time.Now().Add(MuteFor1HrsDuration)
	case MuteFor8Hr:
		MuteTill = time.Now().Add(MuteFor8HrsDuration)
	case MuteFor1Week:
		MuteTill = time.Now().Add(MuteFor1WeekDuration)
	default:
		MuteTill = time.Time{}
	}
	err := m.saveChat(chat)
	if err != nil {
		return time.Time{}, err
	}

	muteTillTimeRemoveMs, err := time.Parse(time.RFC3339, MuteTill.Format(time.RFC3339))

	if err != nil {
		return time.Time{}, err
	}

	return m.muteChat(chat, contact, muteTillTimeRemoveMs)
}

func (m *Messenger) MuteChatV2(muteParams *requests.MuteChat) (time.Time, error) {
	return m.MuteChat(muteParams)
}

func (m *Messenger) muteChat(chat *Chat, contact *Contact, mutedTill time.Time) (time.Time, error) {
	err := m.persistence.MuteChat(chat.ID, mutedTill)
	if err != nil {
		return time.Time{}, err
	}

	chat.Muted = true
	chat.MuteTill = mutedTill
	// TODO(samyoul) remove storing of an updated reference pointer?
	m.allChats.Store(chat.ID, chat)

	if contact != nil {
		err := m.syncContact(context.Background(), contact, m.dispatchMessage)
		if err != nil {
			return time.Time{}, err
		}
	}

	if !chat.MuteTill.IsZero() {
		err := m.reregisterForPushNotifications()
		if err != nil {
			return time.Time{}, err
		}
		return mutedTill, nil
	}

	return time.Time{}, m.reregisterForPushNotifications()
}

// UnmuteChat signals to the messenger that we want to be notified
// on new messages from this chat
func (m *Messenger) UnmuteChat(chatID string) error {
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return errors.New("chat not found")
	}

	var contact *Contact
	if chat.OneToOne() {
		contact, _ = m.allContacts.Load(chatID)
	}

	return m.unmuteChat(chat, contact)
}

func (m *Messenger) unmuteChat(chat *Chat, contact *Contact) error {
	err := m.persistence.UnmuteChat(chat.ID)
	if err != nil {
		return err
	}

	chat.Muted = false
	chat.MuteTill = time.Time{}
	// TODO(samyoul) remove storing of an updated reference pointer?
	m.allChats.Store(chat.ID, chat)

	if chat.CommunityChat() {
		community, err := m.communitiesManager.GetByIDString(chat.CommunityID)
		if err != nil {
			return err
		}

		err = m.communitiesManager.SetMuted(community.ID(), false)
		if err != nil {
			return err
		}
	}

	if contact != nil {
		err := m.syncContact(context.Background(), contact, m.dispatchMessage)
		if err != nil {
			return err
		}
	}
	return m.reregisterForPushNotifications()
}

func (m *Messenger) UpdateMessageOutgoingStatus(id, newOutgoingStatus string) error {
	return m.persistence.UpdateMessageOutgoingStatus(id, newOutgoingStatus)
}

// Identicon returns an identicon based on the input string
func Identicon(id string) (string, error) {
	return identicon.GenerateBase64(id)
}

// GenerateAlias name returns the generated name given a public key hex encoded prefixed with 0x
func GenerateAlias(id string) (string, error) {
	return alias.GenerateFromPublicKeyString(id)
}

func (m *Messenger) RequestTransaction(ctx context.Context, chatID, value, contract, address string) (*MessengerResponse, error) {
	var response MessengerResponse

	// A valid added chat is required.
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, errors.New("Chat not found")
	}
	if chat.ChatType != ChatTypeOneToOne {
		return nil, errors.New("Need to be a one-to-one chat")
	}

	message := common.NewMessage()
	err := extendMessageFromChat(message, chat, &m.identity.PublicKey, m.transport)
	if err != nil {
		return nil, err
	}

	message.MessageType = protobuf.MessageType_ONE_TO_ONE
	message.ContentType = protobuf.ChatMessage_TRANSACTION_COMMAND
	message.Seen = true
	message.Text = "Request transaction"

	request := &protobuf.RequestTransaction{
		Clock:    message.Clock,
		Address:  address,
		Value:    value,
		Contract: contract,
		ChatId:   chatID,
	}
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_REQUEST_TRANSACTION,
		ResendAutomatically: true,
	})

	message.CommandParameters = &common.CommandParameters{
		ID:           rawMessage.ID,
		Value:        value,
		Address:      address,
		Contract:     contract,
		CommandState: common.CommandStateRequestTransaction,
	}

	if err != nil {
		return nil, err
	}
	messageID := rawMessage.ID

	message.ID = messageID
	message.CommandParameters.ID = messageID
	err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}

	err = chat.UpdateFromMessage(message, m.transport)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	return m.addMessagesAndChat(chat, []*common.Message{message}, &response)
}

func (m *Messenger) RequestAddressForTransaction(ctx context.Context, chatID, from, value, contract string) (*MessengerResponse, error) {
	var response MessengerResponse

	// A valid added chat is required.
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, errors.New("Chat not found")
	}
	if chat.ChatType != ChatTypeOneToOne {
		return nil, errors.New("Need to be a one-to-one chat")
	}

	message := common.NewMessage()
	err := extendMessageFromChat(message, chat, &m.identity.PublicKey, m.transport)
	if err != nil {
		return nil, err
	}

	message.MessageType = protobuf.MessageType_ONE_TO_ONE
	message.ContentType = protobuf.ChatMessage_TRANSACTION_COMMAND
	message.Seen = true
	message.Text = "Request address for transaction"

	request := &protobuf.RequestAddressForTransaction{
		Clock:    message.Clock,
		Value:    value,
		Contract: contract,
		ChatId:   chatID,
	}
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_REQUEST_ADDRESS_FOR_TRANSACTION,
		ResendAutomatically: true,
	})

	message.CommandParameters = &common.CommandParameters{
		ID:           rawMessage.ID,
		From:         from,
		Value:        value,
		Contract:     contract,
		CommandState: common.CommandStateRequestAddressForTransaction,
	}

	if err != nil {
		return nil, err
	}
	messageID := rawMessage.ID

	message.ID = messageID
	message.CommandParameters.ID = messageID
	err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}

	err = chat.UpdateFromMessage(message, m.transport)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	return m.addMessagesAndChat(chat, []*common.Message{message}, &response)
}

func (m *Messenger) AcceptRequestAddressForTransaction(ctx context.Context, messageID, address string) (*MessengerResponse, error) {
	var response MessengerResponse

	message, err := m.MessageByID(messageID)
	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, errors.New("message not found")
	}

	chatID := message.LocalChatID

	// A valid added chat is required.
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, errors.New("Chat not found")
	}
	if chat.ChatType != ChatTypeOneToOne {
		return nil, errors.New("Need to be a one-to-one chat")
	}

	clock, timestamp := chat.NextClockAndTimestamp(m.transport)
	message.Clock = clock
	message.WhisperTimestamp = timestamp
	message.Timestamp = timestamp
	message.Text = "Request address for transaction accepted"
	message.Seen = true
	message.OutgoingStatus = common.OutgoingStatusSending

	// Hide previous message
	previousMessage, err := m.persistence.MessageByCommandID(chatID, messageID)
	if err != nil {
		return nil, err
	}

	if previousMessage == nil {
		return nil, errors.New("No previous message found")
	}

	err = m.persistence.HideMessage(previousMessage.ID)
	if err != nil {
		return nil, err
	}

	message.Replace = previousMessage.ID

	request := &protobuf.AcceptRequestAddressForTransaction{
		Clock:   message.Clock,
		Id:      messageID,
		Address: address,
		ChatId:  chatID,
	}
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_ACCEPT_REQUEST_ADDRESS_FOR_TRANSACTION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	message.ID = rawMessage.ID
	message.CommandParameters.Address = address
	message.CommandParameters.CommandState = common.CommandStateRequestAddressForTransactionAccepted

	err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}

	err = chat.UpdateFromMessage(message, m.transport)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	return m.addMessagesAndChat(chat, []*common.Message{message}, &response)
}

func (m *Messenger) DeclineRequestTransaction(ctx context.Context, messageID string) (*MessengerResponse, error) {
	var response MessengerResponse

	message, err := m.MessageByID(messageID)
	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, errors.New("message not found")
	}

	chatID := message.LocalChatID

	// A valid added chat is required.
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, errors.New("Chat not found")
	}
	if chat.ChatType != ChatTypeOneToOne {
		return nil, errors.New("Need to be a one-to-one chat")
	}

	clock, timestamp := chat.NextClockAndTimestamp(m.transport)
	message.Clock = clock
	message.WhisperTimestamp = timestamp
	message.Timestamp = timestamp
	message.Text = "Transaction request declined"
	message.Seen = true
	message.OutgoingStatus = common.OutgoingStatusSending
	message.Replace = messageID

	err = m.persistence.HideMessage(messageID)
	if err != nil {
		return nil, err
	}

	request := &protobuf.DeclineRequestTransaction{
		Clock:  message.Clock,
		Id:     messageID,
		ChatId: chatID,
	}
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_DECLINE_REQUEST_TRANSACTION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	message.ID = rawMessage.ID
	message.CommandParameters.CommandState = common.CommandStateRequestTransactionDeclined

	err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}

	err = chat.UpdateFromMessage(message, m.transport)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	return m.addMessagesAndChat(chat, []*common.Message{message}, &response)
}

func (m *Messenger) DeclineRequestAddressForTransaction(ctx context.Context, messageID string) (*MessengerResponse, error) {
	var response MessengerResponse

	message, err := m.MessageByID(messageID)
	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, errors.New("message not found")
	}

	chatID := message.LocalChatID

	// A valid added chat is required.
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, errors.New("Chat not found")
	}
	if chat.ChatType != ChatTypeOneToOne {
		return nil, errors.New("Need to be a one-to-one chat")
	}

	clock, timestamp := chat.NextClockAndTimestamp(m.transport)
	message.Clock = clock
	message.WhisperTimestamp = timestamp
	message.Timestamp = timestamp
	message.Text = "Request address for transaction declined"
	message.Seen = true
	message.OutgoingStatus = common.OutgoingStatusSending
	message.Replace = messageID

	err = m.persistence.HideMessage(messageID)
	if err != nil {
		return nil, err
	}

	request := &protobuf.DeclineRequestAddressForTransaction{
		Clock:  message.Clock,
		Id:     messageID,
		ChatId: chatID,
	}
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_DECLINE_REQUEST_ADDRESS_FOR_TRANSACTION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	message.ID = rawMessage.ID
	message.CommandParameters.CommandState = common.CommandStateRequestAddressForTransactionDeclined

	err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}

	err = chat.UpdateFromMessage(message, m.transport)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	return m.addMessagesAndChat(chat, []*common.Message{message}, &response)
}

func (m *Messenger) AcceptRequestTransaction(ctx context.Context, transactionHash, messageID string, signature []byte) (*MessengerResponse, error) {
	var response MessengerResponse

	message, err := m.MessageByID(messageID)
	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, errors.New("message not found")
	}

	chatID := message.LocalChatID

	// A valid added chat is required.
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, errors.New("Chat not found")
	}
	if chat.ChatType != ChatTypeOneToOne {
		return nil, errors.New("Need to be a one-to-one chat")
	}

	clock, timestamp := chat.NextClockAndTimestamp(m.transport)
	message.Clock = clock
	message.WhisperTimestamp = timestamp
	message.Timestamp = timestamp
	message.Seen = true
	message.Text = transactionSentTxt
	message.OutgoingStatus = common.OutgoingStatusSending

	// Hide previous message
	previousMessage, err := m.persistence.MessageByCommandID(chatID, messageID)
	if err != nil && err != common.ErrRecordNotFound {
		return nil, err
	}

	if previousMessage != nil {
		err = m.persistence.HideMessage(previousMessage.ID)
		if err != nil {
			return nil, err
		}
		message.Replace = previousMessage.ID
	}

	err = m.persistence.HideMessage(messageID)
	if err != nil {
		return nil, err
	}

	request := &protobuf.SendTransaction{
		Clock:           message.Clock,
		Id:              messageID,
		TransactionHash: transactionHash,
		Signature:       signature,
		ChatId:          chatID,
	}
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SEND_TRANSACTION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	message.ID = rawMessage.ID
	message.CommandParameters.TransactionHash = transactionHash
	message.CommandParameters.Signature = signature
	message.CommandParameters.CommandState = common.CommandStateTransactionSent

	err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}

	err = chat.UpdateFromMessage(message, m.transport)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	return m.addMessagesAndChat(chat, []*common.Message{message}, &response)
}

func (m *Messenger) SendTransaction(ctx context.Context, chatID, value, contract, transactionHash string, signature []byte) (*MessengerResponse, error) {
	var response MessengerResponse

	// A valid added chat is required.
	chat, ok := m.allChats.Load(chatID)
	if !ok {
		return nil, errors.New("Chat not found")
	}
	if chat.ChatType != ChatTypeOneToOne {
		return nil, errors.New("Need to be a one-to-one chat")
	}

	message := common.NewMessage()
	err := extendMessageFromChat(message, chat, &m.identity.PublicKey, m.transport)
	if err != nil {
		return nil, err
	}

	message.MessageType = protobuf.MessageType_ONE_TO_ONE
	message.ContentType = protobuf.ChatMessage_TRANSACTION_COMMAND
	message.LocalChatID = chatID

	clock, timestamp := chat.NextClockAndTimestamp(m.transport)
	message.Clock = clock
	message.WhisperTimestamp = timestamp
	message.Seen = true
	message.Timestamp = timestamp
	message.Text = transactionSentTxt

	request := &protobuf.SendTransaction{
		Clock:           message.Clock,
		TransactionHash: transactionHash,
		Signature:       signature,
		ChatId:          chatID,
	}
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SEND_TRANSACTION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	message.ID = rawMessage.ID
	message.CommandParameters = &common.CommandParameters{
		TransactionHash: transactionHash,
		Value:           value,
		Contract:        contract,
		Signature:       signature,
		CommandState:    common.CommandStateTransactionSent,
	}

	err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}

	err = chat.UpdateFromMessage(message, m.transport)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{message})
	if err != nil {
		return nil, err
	}

	return m.addMessagesAndChat(chat, []*common.Message{message}, &response)
}

func (m *Messenger) ValidateTransactions(ctx context.Context, addresses []types.Address) (*MessengerResponse, error) {
	if m.verifyTransactionClient == nil {
		return nil, nil
	}

	logger := m.logger.With(zap.String("site", "ValidateTransactions"))
	logger.Debug("Validating transactions")
	txs, err := m.persistence.TransactionsToValidate()
	if err != nil {
		logger.Error("Error pulling", zap.Error(err))
		return nil, err
	}
	logger.Debug("Txs", zap.Int("count", len(txs)), zap.Any("txs", txs))
	var response MessengerResponse
	validator := NewTransactionValidator(addresses, m.persistence, m.verifyTransactionClient, m.logger)
	responses, err := validator.ValidateTransactions(ctx)
	if err != nil {
		logger.Error("Error validating", zap.Error(err))
		return nil, err
	}
	for _, validationResult := range responses {
		var message *common.Message
		chatID := contactIDFromPublicKey(validationResult.Transaction.From)
		chat, ok := m.allChats.Load(chatID)
		if !ok {
			chat = OneToOneFromPublicKey(validationResult.Transaction.From, m.transport)
		}
		if validationResult.Message != nil {
			message = validationResult.Message
		} else {
			message = common.NewMessage()
			err := extendMessageFromChat(message, chat, &m.identity.PublicKey, m.transport)
			if err != nil {
				return nil, err
			}
		}

		message.MessageType = protobuf.MessageType_ONE_TO_ONE
		message.ContentType = protobuf.ChatMessage_TRANSACTION_COMMAND
		message.LocalChatID = chatID
		message.OutgoingStatus = ""

		clock, timestamp := chat.NextClockAndTimestamp(m.transport)
		message.Clock = clock
		message.Timestamp = timestamp
		message.WhisperTimestamp = timestamp
		message.Text = "Transaction received"
		message.Seen = false

		message.ID = validationResult.Transaction.MessageID
		if message.CommandParameters == nil {
			message.CommandParameters = &common.CommandParameters{}
		} else {
			message.CommandParameters = validationResult.Message.CommandParameters
		}

		message.CommandParameters.Value = validationResult.Value
		message.CommandParameters.Contract = validationResult.Contract
		message.CommandParameters.Address = validationResult.Address
		message.CommandParameters.CommandState = common.CommandStateTransactionSent
		message.CommandParameters.TransactionHash = validationResult.Transaction.TransactionHash

		err = message.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
		if err != nil {
			return nil, err
		}

		err = chat.UpdateFromMessage(message, m.transport)
		if err != nil {
			return nil, err
		}

		if len(message.CommandParameters.ID) != 0 {
			// Hide previous message
			previousMessage, err := m.persistence.MessageByCommandID(chatID, message.CommandParameters.ID)
			if err != nil && err != common.ErrRecordNotFound {
				return nil, err
			}

			if previousMessage != nil {
				err = m.persistence.HideMessage(previousMessage.ID)
				if err != nil {
					return nil, err
				}
				message.Replace = previousMessage.ID
			}
		}

		response.AddMessage(message)
		m.allChats.Store(chat.ID, chat)
		response.AddChat(chat)

		contact, err := m.getOrBuildContactFromMessage(message)
		if err != nil {
			return nil, err
		}

		notificationsEnabled, err := m.settings.GetNotificationsEnabled()
		if err != nil {
			return nil, err
		}

		profilePicturesVisibility, err := m.settings.GetProfilePicturesVisibility()
		if err != nil {
			return nil, err
		}

		if notificationsEnabled {
			notification, err := NewMessageNotification(message.ID, message, chat, contact, m.ResolvePrimaryName, profilePicturesVisibility)
			if err != nil {
				return nil, err
			}
			response.AddNotification(notification)
		}

	}

	if len(response.messages) > 0 {
		err = m.SaveMessages(response.Messages())
		if err != nil {
			return nil, err
		}
	}
	return &response, nil
}

// pullMessagesAndResponsesFromDB pulls all the messages and the one that have
// been replied to from the database
func (m *Messenger) pullMessagesAndResponsesFromDB(messages []*common.Message) ([]*common.Message, error) {
	var messageIDs []string
	for _, message := range messages {
		messageIDs = append(messageIDs, message.ID)
		if len(message.ResponseTo) != 0 {
			messageIDs = append(messageIDs, message.ResponseTo)
		}

	}
	// We pull from the database all the messages & replies involved,
	// so we let the db build the correct messages
	return m.persistence.MessagesByIDs(messageIDs)
}

func (m *Messenger) SignMessage(message string) ([]byte, error) {
	hash := crypto.TextHash([]byte(message))
	return crypto.Sign(hash, m.identity)
}

func (m *Messenger) CreateCommunityTokenDeploymentSignature(ctx context.Context, chainID uint64, addressFrom string, communityID string) ([]byte, error) {
	return m.communitiesManager.CreateCommunityTokenDeploymentSignature(ctx, chainID, addressFrom, communityID)
}

func (m *Messenger) getTimesource() common.TimeSource {
	return m.transport
}

func (m *Messenger) GetCurrentTimeInMillis() uint64 {
	return m.getTimesource().GetCurrentTime()
}

// AddPushNotificationsServer adds a push notification server
func (m *Messenger) AddPushNotificationsServer(ctx context.Context, publicKey *ecdsa.PublicKey, serverType pushnotificationclient.ServerType) error {
	if m.pushNotificationClient == nil {
		return errors.New("push notification client not enabled")
	}
	return m.pushNotificationClient.AddPushNotificationsServer(publicKey, serverType)
}

// RemovePushNotificationServer removes a push notification server
func (m *Messenger) RemovePushNotificationServer(ctx context.Context, publicKey *ecdsa.PublicKey) error {
	if m.pushNotificationClient == nil {
		return errors.New("push notification client not enabled")
	}
	return m.pushNotificationClient.RemovePushNotificationServer(publicKey)
}

// UnregisterFromPushNotifications unregister from any server
func (m *Messenger) UnregisterFromPushNotifications(ctx context.Context) error {
	return m.pushNotificationClient.Unregister()
}

// DisableSendingPushNotifications signals the client not to send any push notification
func (m *Messenger) DisableSendingPushNotifications() error {
	if m.pushNotificationClient == nil {
		return errors.New("push notification client not enabled")
	}
	m.pushNotificationClient.DisableSending()
	return nil
}

// EnableSendingPushNotifications signals the client to send push notifications
func (m *Messenger) EnableSendingPushNotifications() error {
	if m.pushNotificationClient == nil {
		return errors.New("push notification client not enabled")
	}
	m.pushNotificationClient.EnableSending()
	return nil
}

func (m *Messenger) pushNotificationOptions() *pushnotificationclient.RegistrationOptions {
	var contactIDs []*ecdsa.PublicKey
	var mutedChatIDs []string
	var publicChatIDs []string
	var blockedChatIDs []string

	m.allContacts.Range(func(contactID string, contact *Contact) (shouldContinue bool) {
		if contact.added() && !contact.Blocked {
			pk, err := contact.PublicKey()
			if err != nil {
				m.logger.Warn("could not parse contact public key")
				return true
			}
			contactIDs = append(contactIDs, pk)
		} else if contact.Blocked {
			blockedChatIDs = append(blockedChatIDs, contact.ID)
		}
		return true
	})

	m.allChats.Range(func(chatID string, chat *Chat) (shouldContinue bool) {
		if chat.Muted {
			mutedChatIDs = append(mutedChatIDs, chat.ID)
			return true
		}
		if chat.Active && (chat.Public() || chat.CommunityChat()) {
			publicChatIDs = append(publicChatIDs, chat.ID)
		}
		return true
	})

	return &pushnotificationclient.RegistrationOptions{
		ContactIDs:     contactIDs,
		MutedChatIDs:   mutedChatIDs,
		PublicChatIDs:  publicChatIDs,
		BlockedChatIDs: blockedChatIDs,
	}
}

// RegisterForPushNotification register deviceToken with any push notification server enabled
func (m *Messenger) RegisterForPushNotifications(ctx context.Context, deviceToken, apnTopic string, tokenType protobuf.PushNotificationRegistration_TokenType) error {
	if m.pushNotificationClient == nil {
		return errors.New("push notification client not enabled")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	err := m.pushNotificationClient.Register(deviceToken, apnTopic, tokenType, m.pushNotificationOptions())
	if err != nil {
		m.logger.Error("failed to register for push notifications", zap.Error(err))
		return err
	}
	return nil
}

// RegisteredForPushNotifications returns whether we successfully registered with all the servers
func (m *Messenger) RegisteredForPushNotifications() (bool, error) {
	if m.pushNotificationClient == nil {
		return false, errors.New("no push notification client")
	}
	return m.pushNotificationClient.Registered()
}

// EnablePushNotificationsFromContactsOnly is used to indicate that we want to received push notifications only from contacts
func (m *Messenger) EnablePushNotificationsFromContactsOnly() error {
	if m.pushNotificationClient == nil {
		return errors.New("no push notification client")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.pushNotificationClient.EnablePushNotificationsFromContactsOnly(m.pushNotificationOptions())
}

// DisablePushNotificationsFromContactsOnly is used to indicate that we want to received push notifications from anyone
func (m *Messenger) DisablePushNotificationsFromContactsOnly() error {
	if m.pushNotificationClient == nil {
		return errors.New("no push notification client")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.pushNotificationClient.DisablePushNotificationsFromContactsOnly(m.pushNotificationOptions())
}

// EnablePushNotificationsBlockMentions is used to indicate that we dont want to received push notifications for mentions
func (m *Messenger) EnablePushNotificationsBlockMentions() error {
	if m.pushNotificationClient == nil {
		return errors.New("no push notification client")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.pushNotificationClient.EnablePushNotificationsBlockMentions(m.pushNotificationOptions())
}

// DisablePushNotificationsBlockMentions is used to indicate that we want to received push notifications for mentions
func (m *Messenger) DisablePushNotificationsBlockMentions() error {
	if m.pushNotificationClient == nil {
		return errors.New("no push notification client")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.pushNotificationClient.DisablePushNotificationsBlockMentions(m.pushNotificationOptions())
}

// GetPushNotificationsServers returns the servers used for push notifications
func (m *Messenger) GetPushNotificationsServers() ([]*pushnotificationclient.PushNotificationServer, error) {
	if m.pushNotificationClient == nil {
		return nil, errors.New("no push notification client")
	}
	return m.pushNotificationClient.GetServers()
}

// StartPushNotificationsServer initialize and start a push notification server, using the current messenger identity key
func (m *Messenger) StartPushNotificationsServer() error {
	if m.pushNotificationServer == nil {
		pushNotificationServerPersistence := pushnotificationserver.NewSQLitePersistence(m.database)
		config := &pushnotificationserver.Config{
			Enabled:  true,
			Logger:   m.logger,
			Identity: m.identity,
		}
		m.pushNotificationServer = pushnotificationserver.New(config, pushNotificationServerPersistence, m.sender)
	}

	return m.pushNotificationServer.Start()
}

// StopPushNotificationServer stops the push notification server if running
func (m *Messenger) StopPushNotificationsServer() error {
	m.pushNotificationServer = nil
	return nil
}

func generateAliasAndIdenticon(pk string) (string, string, error) {
	identicon, err := identicon.GenerateBase64(pk)
	if err != nil {
		return "", "", err
	}

	name, err := alias.GenerateFromPublicKeyString(pk)
	if err != nil {
		return "", "", err
	}
	return name, identicon, nil

}

func (m *Messenger) encodeChatEntity(chat *Chat, message common.ChatEntity) ([]byte, error) {
	var encodedMessage []byte
	var err error
	l := m.logger.With(zap.String("site", "Send"), zap.String("chatID", chat.ID))

	switch chat.ChatType {
	case ChatTypeOneToOne:
		l.Debug("sending private message")
		message.SetMessageType(protobuf.MessageType_ONE_TO_ONE)
		encodedMessage, err = proto.Marshal(message.GetProtobuf())
		if err != nil {
			return nil, err
		}

	case ChatTypePublic, ChatTypeProfile:
		l.Debug("sending public message", zap.String("chatName", chat.Name))
		message.SetMessageType(protobuf.MessageType_PUBLIC_GROUP)
		encodedMessage, err = proto.Marshal(message.GetProtobuf())
		if err != nil {
			return nil, err
		}

	case ChatTypeCommunityChat:
		l.Debug("sending community chat message", zap.String("chatName", chat.Name))
		message.SetMessageType(protobuf.MessageType_COMMUNITY_CHAT)
		encodedMessage, err = proto.Marshal(message.GetProtobuf())
		if err != nil {
			return nil, err
		}

	case ChatTypePrivateGroupChat:
		message.SetMessageType(protobuf.MessageType_PRIVATE_GROUP)
		l.Debug("sending group message", zap.String("chatName", chat.Name))
		if !message.WrapGroupMessage() {
			encodedMessage, err = proto.Marshal(message.GetProtobuf())
			if err != nil {
				return nil, err
			}
		} else {

			group, err := newProtocolGroupFromChat(chat)
			if err != nil {
				return nil, err
			}

			// NOTE(cammellos): Disabling for now since the optimiziation is not
			// applicable anymore after we changed group rules to allow
			// anyone to change group details
			encodedMessage, err = m.sender.EncodeMembershipUpdate(group, message)
			if err != nil {
				return nil, err
			}
		}

	default:
		return nil, errors.New("chat type not supported")
	}

	return encodedMessage, nil
}

func (m *Messenger) getOrBuildContactFromMessage(msg *common.Message) (*Contact, error) {
	if c, ok := m.allContacts.Load(msg.From); ok {
		return c, nil
	}

	senderPubKey, err := msg.GetSenderPubKey()
	if err != nil {
		return nil, err
	}
	senderID := contactIDFromPublicKey(senderPubKey)
	c, err := buildContact(senderID, senderPubKey)
	if err != nil {
		return nil, err
	}

	// TODO(samyoul) remove storing of an updated reference pointer?
	m.allContacts.Store(msg.From, c)
	return c, nil
}

func (m *Messenger) BloomFilter() []byte {
	return m.transport.BloomFilter()
}

func (m *Messenger) getSettings() (settings.Settings, error) {
	sDB, err := accounts.NewDB(m.database)
	if err != nil {
		return settings.Settings{}, err
	}
	return sDB.GetSettings()
}

func (m *Messenger) getEnsUsernameDetails() (result []*ensservice.UsernameDetail, err error) {
	db := ensservice.NewEnsDatabase(m.database)
	return db.GetEnsUsernames(nil)
}

func ToVerificationRequest(message *protobuf.SyncVerificationRequest) *verification.Request {
	return &verification.Request{
		From:          message.From,
		To:            message.To,
		Challenge:     message.Challenge,
		Response:      message.Response,
		RequestedAt:   message.RequestedAt,
		RepliedAt:     message.RepliedAt,
		RequestStatus: verification.RequestStatus(message.VerificationStatus),
	}
}

func (m *Messenger) HandleSyncVerificationRequest(state *ReceivedMessageState, message *protobuf.SyncVerificationRequest, statusMessage *v1protocol.StatusMessage) error {
	verificationRequest := ToVerificationRequest(message)

	err := m.verificationDatabase.SaveVerificationRequest(verificationRequest)
	if err != nil {
		return err
	}

	myPubKey := hexutil.Encode(crypto.FromECDSAPub(&m.identity.PublicKey))

	state.AllVerificationRequests = append(state.AllVerificationRequests, verificationRequest)

	if message.From == myPubKey { // Verification requests we sent
		contact, ok := m.allContacts.Load(message.To)
		if !ok {
			m.logger.Info("contact not found")
			return nil
		}

		contact.VerificationStatus = VerificationStatus(message.VerificationStatus)
		if err := m.persistence.SaveContact(contact, nil); err != nil {
			return err
		}

		m.allContacts.Store(contact.ID, contact)
		state.ModifiedContacts.Store(contact.ID, true)

		// TODO: create activity center notif

	}
	// else { // Verification requests we received
	// // TODO: activity center notif
	//}

	return nil
}

func (m *Messenger) ImageServerURL() string {
	return m.httpServer.MakeImageServerURL()
}

func (m *Messenger) myHexIdentity() string {
	return common.PubkeyToHex(&m.identity.PublicKey)
}

func (m *Messenger) GetMentionsManager() *MentionManager {
	return m.mentionsManager
}

func (m *Messenger) getOtherMessagesInAlbum(message *common.Message, chatID string) ([]*common.Message, error) {
	var connectedMessages []*common.Message
	// In case of Image messages, we need to delete all the images in the album
	if message.ContentType == protobuf.ChatMessage_IMAGE {
		image := message.GetImage()
		if image != nil && image.AlbumId != "" {
			messagesInTheAlbum, err := m.persistence.albumMessages(chatID, image.GetAlbumId())
			if err != nil {
				return nil, err
			}
			connectedMessages = append(connectedMessages, messagesInTheAlbum...)
			return connectedMessages, nil
		}
	}
	return append(connectedMessages, message), nil
}

func (m *Messenger) withChatClock(callback func(string, uint64) error) error {
	clock, chat := m.getLastClockWithRelatedChat()
	err := callback(chat.ID, clock)
	if err != nil {
		return err
	}
	chat.LastClockValue = clock
	return m.saveChat(chat)
}

func (m *Messenger) syncDeleteForMeMessage(ctx context.Context, rawMessageDispatcher RawMessageHandler) error {
	deleteForMes, err := m.persistence.GetDeleteForMeMessages()
	if err != nil {
		return err
	}

	return m.withChatClock(func(chatID string, _ uint64) error {
		for _, deleteForMe := range deleteForMes {
			encodedMessage, err2 := proto.Marshal(deleteForMe)
			if err2 != nil {
				return err2
			}
			rawMessage := common.RawMessage{
				LocalChatID:         chatID,
				Payload:             encodedMessage,
				MessageType:         protobuf.ApplicationMetadataMessage_SYNC_DELETE_FOR_ME_MESSAGE,
				ResendAutomatically: true,
			}
			_, err2 = rawMessageDispatcher(ctx, rawMessage)
			if err2 != nil {
				return err2
			}
		}
		return nil
	})
}

func (m *Messenger) syncSocialLinks(ctx context.Context, rawMessageDispatcher RawMessageHandler) error {
	if !m.hasPairedDevices() {
		return nil
	}

	dbSocialLinks, err := m.settings.GetSocialLinks()
	if err != nil {
		return err
	}

	dbClock, err := m.settings.GetSocialLinksClock()
	if err != nil {
		return err
	}

	_, chat := m.getLastClockWithRelatedChat()
	encodedMessage, err := proto.Marshal(dbSocialLinks.ToSyncProtobuf(dbClock))
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_SYNC_SOCIAL_LINKS,
		ResendAutomatically: true,
	}

	_, err = rawMessageDispatcher(ctx, rawMessage)
	return err
}

func (m *Messenger) HandleSyncSocialLinks(state *ReceivedMessageState, message *protobuf.SyncSocialLinks, statusMessage *v1protocol.StatusMessage) error {
	return m.handleSyncSocialLinks(message, func(links identity.SocialLinks) {
		state.Response.SocialLinksInfo = &identity.SocialLinksInfo{
			Links:   links,
			Removed: len(links) == 0,
		}
	})
}

func (m *Messenger) handleSyncSocialLinks(message *protobuf.SyncSocialLinks, callback func(identity.SocialLinks)) error {
	if message == nil {
		return nil
	}
	var (
		links identity.SocialLinks
		err   error
	)
	for _, sl := range message.SocialLinks {
		link := &identity.SocialLink{
			Text: sl.Text,
			URL:  sl.Url,
		}
		err = ValidateSocialLink(link)
		if err != nil {
			return err
		}

		links = append(links, link)
	}

	err = m.settings.AddOrReplaceSocialLinksIfNewer(links, message.Clock)
	if err != nil {
		if err == sociallinkssettings.ErrOlderSocialLinksProvided {
			return nil
		}
		return err
	}

	callback(links)

	return nil
}

func (m *Messenger) GetDeleteForMeMessages() ([]*protobuf.SyncDeleteForMeMessage, error) {
	return m.persistence.GetDeleteForMeMessages()
}

func (m *Messenger) startMessageSegmentsCleanupLoop() {
	logger := m.logger.Named("messageSegmentsCleanupLoop")

	go func() {
		// Delay by a few minutes to minimize messenger's startup time
		var interval time.Duration = 5 * time.Minute
		for {
			select {
			case <-time.After(interval):
				// Set the regular interval after the first execution
				interval = 1 * time.Hour

				err := m.sender.CleanupSegments()
				if err != nil {
					logger.Error("failed to cleanup segments", zap.Error(err))
				}

			case <-m.quit:
				return
			}
		}
	}()
}
