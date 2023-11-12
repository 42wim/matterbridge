package protocol

import (
	"database/sql"
	"encoding/json"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/services/browsers"
	"github.com/status-im/status-go/services/communitytokens"
	"github.com/status-im/status-go/wakuv2"

	"go.uber.org/zap"

	"github.com/status-im/status-go/appdatabase/migrations"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol/anonmetrics"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/discord"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/pushnotificationclient"
	"github.com/status-im/status-go/protocol/pushnotificationserver"
	"github.com/status-im/status-go/protocol/transport"
	"github.com/status-im/status-go/protocol/wakusync"
	"github.com/status-im/status-go/services/mailservers"
	"github.com/status-im/status-go/services/wallet"
)

type MessageDeliveredHandler func(string, string)

type MessengerSignalsHandler interface {
	MessageDelivered(chatID string, messageID string)
	CommunityInfoFound(community *communities.Community)
	MessengerResponse(response *MessengerResponse)
	HistoryRequestStarted(numBatches int)
	HistoryRequestCompleted()

	BackupPerformed(uint64)
	HistoryArchivesProtocolEnabled()
	HistoryArchivesProtocolDisabled()
	CreatingHistoryArchives(communityID string)
	NoHistoryArchivesCreated(communityID string, from int, to int)
	HistoryArchivesCreated(communityID string, from int, to int)
	HistoryArchivesSeeding(communityID string)
	HistoryArchivesUnseeded(communityID string)
	HistoryArchiveDownloaded(communityID string, from int, to int)
	DownloadingHistoryArchivesStarted(communityID string)
	DownloadingHistoryArchivesFinished(communityID string)
	ImportingHistoryArchiveMessages(communityID string)
	StatusUpdatesTimedOut(statusUpdates *[]UserStatus)
	DiscordCategoriesAndChannelsExtracted(categories []*discord.Category, channels []*discord.Channel, oldestMessageTimestamp int64, errors map[string]*discord.ImportError)
	DiscordCommunityImportProgress(importProgress *discord.ImportProgress)
	DiscordCommunityImportFinished(communityID string)
	DiscordCommunityImportCancelled(communityID string)
	DiscordCommunityImportCleanedUp(communityID string)
	DiscordChannelImportProgress(importProgress *discord.ImportProgress)
	DiscordChannelImportFinished(communityID string, channelID string)
	DiscordChannelImportCancelled(channelID string)
	SendWakuFetchingBackupProgress(response *wakusync.WakuBackedUpDataResponse)
	SendWakuBackedUpProfile(response *wakusync.WakuBackedUpDataResponse)
	SendWakuBackedUpSettings(response *wakusync.WakuBackedUpDataResponse)
	SendWakuBackedUpKeypair(response *wakusync.WakuBackedUpDataResponse)
	SendWakuBackedUpWatchOnlyAccount(response *wakusync.WakuBackedUpDataResponse)
	SendCuratedCommunitiesUpdate(response *communities.KnownCommunitiesResponse)
}

type config struct {
	// systemMessagesTranslations holds translations for system-messages
	systemMessagesTranslations *systemMessageTranslationsMap
	// Config for the envelopes monitor
	envelopesMonitorConfig *transport.EnvelopesMonitorConfig

	featureFlags common.FeatureFlags

	appDb                  *sql.DB
	walletDb               *sql.DB
	afterDbCreatedHooks    []Option
	multiAccount           *multiaccounts.Database
	mailserversDatabase    *mailservers.Database
	account                *multiaccounts.Account
	clusterConfig          params.ClusterConfig
	browserDatabase        *browsers.Database
	torrentConfig          *params.TorrentConfig
	walletConfig           *params.WalletConfig
	walletService          *wallet.Service
	communityTokensService communitytokens.ServiceInterface
	httpServer             *server.MediaServer
	rpcClient              *rpc.Client
	tokenManager           communities.TokenManager
	accountsManager        account.Manager

	verifyTransactionClient  EthClient
	verifyENSURL             string
	verifyENSContractAddress string

	anonMetricsClientConfig *anonmetrics.ClientConfig
	anonMetricsServerConfig *anonmetrics.ServerConfig

	pushNotificationServerConfig *pushnotificationserver.Config
	pushNotificationClientConfig *pushnotificationclient.Config

	logger *zap.Logger

	outputMessagesCSV bool

	messengerSignalsHandler MessengerSignalsHandler

	telemetryServerURL string
	wakuService        *wakuv2.Waku

	messageResendMinDelay int
	messageResendMaxCount int
}

func messengerDefaultConfig() config {
	c := config{
		messageResendMinDelay: 30,
		messageResendMaxCount: 3,
	}

	c.featureFlags.AutoRequestHistoricMessages = true
	return c
}

type Option func(*config) error

// WithSystemMessagesTranslations is required for Group Chats which are currently disabled.
// nolint: unused
func WithSystemMessagesTranslations(t map[protobuf.MembershipUpdateEvent_EventType]string) Option {
	return func(c *config) error {
		c.systemMessagesTranslations.Init(t)
		return nil
	}
}

func WithCustomLogger(logger *zap.Logger) Option {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

func WithVerifyTransactionClient(client EthClient) Option {
	return func(c *config) error {
		c.verifyTransactionClient = client
		return nil
	}
}

func WithResendParams(minDelay int, maxCount int) Option {
	return func(c *config) error {
		c.messageResendMinDelay = minDelay
		c.messageResendMaxCount = maxCount
		return nil
	}
}

func WithDatabase(db *sql.DB) Option {
	return func(c *config) error {
		c.appDb = db
		return nil
	}
}

func WithWalletDatabase(db *sql.DB) Option {
	return func(c *config) error {
		c.walletDb = db
		return nil
	}
}

func WithToplevelDatabaseMigrations() Option {
	return func(c *config) error {
		c.afterDbCreatedHooks = append(c.afterDbCreatedHooks, func(c *config) error {
			return migrations.Migrate(c.appDb, nil)
		})
		return nil
	}
}

func WithAppSettings(s settings.Settings, nc params.NodeConfig) Option {
	return func(c *config) error {
		c.afterDbCreatedHooks = append(c.afterDbCreatedHooks, func(c *config) error {
			if s.Networks == nil {
				networks := new(json.RawMessage)
				if err := networks.UnmarshalJSON([]byte("net")); err != nil {
					return err
				}

				s.Networks = networks
			}

			sDB, err := accounts.NewDB(c.appDb)
			if err != nil {
				return err
			}
			return sDB.CreateSettings(s, nc)
		})
		return nil
	}
}

func WithMultiAccounts(ma *multiaccounts.Database) Option {
	return func(c *config) error {
		c.multiAccount = ma
		return nil
	}
}

func WithMailserversDatabase(ma *mailservers.Database) Option {
	return func(c *config) error {
		c.mailserversDatabase = ma
		return nil
	}
}

func WithAccount(acc *multiaccounts.Account) Option {
	return func(c *config) error {
		c.account = acc
		return nil
	}
}

func WithBrowserDatabase(bd *browsers.Database) Option {
	return func(c *config) error {
		c.browserDatabase = bd
		if c.browserDatabase == nil {
			c.afterDbCreatedHooks = append(c.afterDbCreatedHooks, func(c *config) error {
				c.browserDatabase = browsers.NewDB(c.appDb)
				return nil
			})
		}
		return nil
	}
}

func WithAnonMetricsClientConfig(anonMetricsClientConfig *anonmetrics.ClientConfig) Option {
	return func(c *config) error {
		c.anonMetricsClientConfig = anonMetricsClientConfig
		return nil
	}
}

func WithAnonMetricsServerConfig(anonMetricsServerConfig *anonmetrics.ServerConfig) Option {
	return func(c *config) error {
		c.anonMetricsServerConfig = anonMetricsServerConfig
		return nil
	}
}

func WithTelemetry(serverURL string) Option {
	return func(c *config) error {
		c.telemetryServerURL = serverURL
		return nil
	}
}

func WithPushNotificationServerConfig(pushNotificationServerConfig *pushnotificationserver.Config) Option {
	return func(c *config) error {
		c.pushNotificationServerConfig = pushNotificationServerConfig
		return nil
	}
}

func WithPushNotificationClientConfig(pushNotificationClientConfig *pushnotificationclient.Config) Option {
	return func(c *config) error {
		c.pushNotificationClientConfig = pushNotificationClientConfig
		return nil
	}
}

func WithDatasync() func(c *config) error {
	return func(c *config) error {
		c.featureFlags.Datasync = true
		return nil
	}
}

func WithPushNotifications() func(c *config) error {
	return func(c *config) error {
		c.featureFlags.PushNotifications = true
		return nil
	}
}

func WithCheckingForBackupDisabled() func(c *config) error {
	return func(c *config) error {
		c.featureFlags.DisableCheckingForBackup = true
		return nil
	}
}

func WithAutoMessageDisabled() func(c *config) error {
	return func(c *config) error {
		c.featureFlags.DisableAutoMessageLoop = true
		return nil
	}
}

func WithEnvelopesMonitorConfig(emc *transport.EnvelopesMonitorConfig) Option {
	return func(c *config) error {
		c.envelopesMonitorConfig = emc
		return nil
	}
}

func WithSignalsHandler(h MessengerSignalsHandler) Option {
	return func(c *config) error {
		c.messengerSignalsHandler = h
		return nil
	}
}

func WithENSVerificationConfig(url, address string) Option {
	return func(c *config) error {
		c.verifyENSURL = url
		c.verifyENSContractAddress = address
		return nil
	}
}

func WithClusterConfig(cc params.ClusterConfig) Option {
	return func(c *config) error {
		c.clusterConfig = cc
		return nil
	}
}

func WithTorrentConfig(tc *params.TorrentConfig) Option {
	return func(c *config) error {
		c.torrentConfig = tc
		return nil
	}
}

func WithHTTPServer(s *server.MediaServer) Option {
	return func(c *config) error {
		c.httpServer = s
		return nil
	}
}

func WithRPCClient(r *rpc.Client) Option {
	return func(c *config) error {
		c.rpcClient = r
		return nil
	}
}

func WithWalletConfig(wc *params.WalletConfig) Option {
	return func(c *config) error {
		c.walletConfig = wc
		return nil
	}
}

func WithMessageCSV(enabled bool) Option {
	return func(c *config) error {
		c.outputMessagesCSV = enabled
		return nil
	}
}

func WithWalletService(s *wallet.Service) Option {
	return func(c *config) error {
		c.walletService = s
		return nil
	}
}

func WithCommunityTokensService(s communitytokens.ServiceInterface) Option {
	return func(c *config) error {
		c.communityTokensService = s
		return nil
	}
}

func WithWakuService(s *wakuv2.Waku) Option {
	return func(c *config) error {
		c.wakuService = s
		return nil
	}
}

func WithTokenManager(tokenManager communities.TokenManager) Option {
	return func(c *config) error {
		c.tokenManager = tokenManager
		return nil
	}
}

func WithAccountManager(accountManager account.Manager) Option {
	return func(c *config) error {
		c.accountsManager = accountManager
		return nil
	}
}

func WithAutoRequestHistoricMessages(enabled bool) Option {
	return func(c *config) error {
		c.featureFlags.AutoRequestHistoricMessages = enabled
		return nil
	}
}
