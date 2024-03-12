package status

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	crypto "github.com/ethereum/go-ethereum/crypto"
	api "github.com/status-im/status-go/api"
	"github.com/status-im/status-go/appdatabase"
	gethbridge "github.com/status-im/status-go/eth-node/bridge/geth"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	gonode "github.com/status-im/status-go/node"
	params "github.com/status-im/status-go/params"

	status "github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/identity/alias"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/services/ext/mailservers"
	mailserversDB "github.com/status-im/status-go/services/mailservers"

	"github.com/status-im/status-go/common/dbsetup"
	"github.com/status-im/status-go/walletdatabase"
)

type Bstatus struct {
	*bridge.Config

	// message fetching loop controls
	fetchInterval time.Duration
	fetchDone     chan bool

	// node settings
	statusListenPort int
	statusListenAddr string

	statusDataDir        string
	statusNodeConfigFile string

	privateKey *ecdsa.PrivateKey
	nodeConfig *params.NodeConfig
	statusNode *gonode.StatusNode
	messenger  *status.Messenger

	joinedCommunities []string
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bstatus{
		Config:               cfg,
		fetchDone:            make(chan bool),
		statusListenPort:     30303,
		statusListenAddr:     "0.0.0.0",
		statusDataDir:        cfg.GetString("DataDir"),
		statusNodeConfigFile: cfg.GetString("NodeConfigFile"),
		fetchInterval:        500 * time.Millisecond,
	}
}

// Generate a sane configuration for a Status Node
func (b *Bstatus) generateNodeConfig() (*params.NodeConfig, error) {
	options := []params.Option{
		b.withListenAddr(),
	}
	configFiles := []string{b.statusNodeConfigFile}
	config, err := params.NewNodeConfigWithDefaultsAndFiles(
		b.statusDataDir,
		params.MainNetworkID,
		options,
		configFiles,
	)
	if err != nil {
		return nil, err
	}

	infuraToken := os.Getenv("STATUS_BUILD_INFURA_TOKEN")
	if len(infuraToken) == 0 {
		return nil, fmt.Errorf("STATUS_BUILD_INFURA_TOKEN env variable not set")
	}

	createAccRequest := &requests.CreateAccount{
		WalletSecretsConfig: requests.WalletSecretsConfig{
			InfuraToken: infuraToken,
		},
	}
	config.Networks = api.BuildDefaultNetworks(createAccRequest)

	return config, nil
}

// get or generate new settings
func (b *Bstatus) getOrGenerateSettings(nodeConfig *params.NodeConfig, appDB *sql.DB) (*settings.Settings, error) {

	accdb, err := accounts.NewDB(appDB)
	if err != nil {
		return nil, err
	}

	prevSettings, err := accdb.GetSettings()
	if err == nil {
		// return settings if exists
		return &prevSettings, nil
	}

	// create new settings
	s := &settings.Settings{}

	// needed values
	s.BackupEnabled = false
	s.InstallationID = uuid.New().String()
	s.AutoMessageEnabled = false
	s.UseMailservers = true

	// other
	networks := make([]map[string]string, 0)
	networksJSON, err := json.Marshal(networks)
	if err != nil {
		return nil, err
	}
	networkRawMessage := json.RawMessage(networksJSON)
	s.Networks = &networkRawMessage

	err = accdb.CreateSettings(*s, *nodeConfig)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (b *Bstatus) withListenAddr() params.Option {
	if addr := b.GetString("ListenAddr"); addr != "" {
		b.statusListenAddr = addr
	}
	if port := b.GetInt("ListenPort"); port != 0 {
		b.statusListenPort = port
	}
	return func(c *params.NodeConfig) error {
		c.ListenAddr = fmt.Sprintf("%s:%d", b.statusListenAddr, b.statusListenPort)
		return nil
	}
}

// Main loop for fetching Status messages and relaying them to the bridge
func (b *Bstatus) fetchMessagesLoop() {
	ticker := time.NewTicker(b.fetchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mResp, err := b.messenger.RetrieveAll()
			if err != nil {
				b.Log.WithError(err).Error("Failed to retrieve messages")
				continue
			}
			for _, msg := range mResp.Messages() {
				b.propagateMessage(msg)
			}
		case <-b.fetchDone:
			return
		}
	}
}

func (b *Bstatus) stopMessagesLoop() {
	close(b.fetchDone)
}

func (b *Bstatus) getDisplayName(c *status.Contact) string {
	if c.ENSVerified && c.EnsName != "" {
		return c.EnsName
	}
	if c.DisplayName != "" {
		return c.DisplayName
	}
	if c.PrimaryName() != "" {
		return c.PrimaryName()
	}
	return c.Alias
}

func (b *Bstatus) propagateMessage(msg *common.Message) {
	var username string
	contact := b.messenger.GetContactByID(msg.From)
	if contact == nil {
		threeWordsName, err := alias.GenerateFromPublicKeyString(msg.From)
		if err != nil {
			username = msg.From
		} else {
			username = threeWordsName
		}
	} else {
		username = b.getDisplayName(contact)
	}

	// Send message for processing
	b.Remote <- config.Message{
		Timestamp: time.Unix(int64(msg.WhisperTimestamp), 0),
		Username:  username,
		Text:      msg.Text,
		Channel:   msg.ChatId,
		ID:        msg.ID,
		Account:   b.Account,
	}
}

// Converts a bridge message into a Status message
func (b *Bstatus) toStatusMsg(msg config.Message) *common.Message {
	message := common.NewMessage()
	message.ChatId = msg.Channel
	message.ContentType = protobuf.ChatMessage_BRIDGE_MESSAGE

	var originalID, originalParentID string
	if msg.Extra != nil {
		originalMessageIdsList := msg.Extra["OriginalMessageIds"]
		if len(originalMessageIdsList) == 1 {
			originalMessageIds, ok := originalMessageIdsList[0].(config.OriginalMessageIds)
			if ok {
				originalID = originalMessageIds.ID
				originalParentID = originalMessageIds.ParentID
			}
		}
	}

	message.Payload = &protobuf.ChatMessage_BridgeMessage{
		BridgeMessage: &protobuf.BridgeMessage{
			BridgeName:      msg.Protocol,
			UserName:        msg.Username,
			UserAvatar:      msg.Avatar,
			UserID:          msg.UserID,
			Content:         msg.Text,
			MessageID:       originalID,
			ParentMessageID: originalParentID,
		},
	}

	return message
}

func (b *Bstatus) connected() bool {
	return b.statusNode.IsRunning() && b.messenger.Online()
}

func (b *Bstatus) getCommunityIdFromFullChatId(chatId string) string {
	const communityIdLength = 68
	if len(chatId) <= communityIdLength {
		return ""
	}
	return chatId[0:communityIdLength]
}

func (b *Bstatus) createMultiAccount(privKey *ecdsa.PrivateKey) multiaccounts.Account {
	keyUID := sha256.Sum256(crypto.FromECDSAPub(&privKey.PublicKey))
	keyUIDHex := types.EncodeHex(keyUID[:])
	return multiaccounts.Account{
		KeyUID: keyUIDHex,
	}
}

// i-face functions

func (b *Bstatus) Send(msg config.Message) (string, error) {
	if !b.connected() {
		return "", fmt.Errorf("bridge %s not connected, dropping message %#v to bridge", b.Account, msg)
	}

	b.Log.Debugf("=> Sending message %#v", msg)

	_, err := b.messenger.SendChatMessage(context.Background(), b.toStatusMsg(msg))
	if err != nil {
		return "", errors.Wrap(err, "failed to send message")
	}

	return "", nil
}

func (b *Bstatus) Connect() error {
	if len(b.statusDataDir) == 0 {
		b.statusDataDir = os.TempDir() + "/matterbridge-status-data"
	}
	err := os.Mkdir(b.statusDataDir, 0750)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "Failed to create status directory")
	}

	keyHex := strings.TrimPrefix(b.GetString("Token"), "0x")
	if privKey, err := crypto.HexToECDSA(keyHex); err != nil {
		return errors.Wrap(err, "Failed to parse private key in Token field")
	} else {
		b.privateKey = privKey
	}

	b.nodeConfig, err = b.generateNodeConfig()
	if err != nil {
		return errors.Wrap(err, "Failed to generate node config")
	}

	backend := api.NewGethStatusBackend()
	b.statusNode = backend.StatusNode()

	walletDB, err := walletdatabase.InitializeDB(b.statusDataDir+"/"+"wallet.db", "", dbsetup.ReducedKDFIterationsNumber)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize wallet db")
	}
	b.statusNode.SetWalletDB(walletDB)

	appDB, err := appdatabase.InitializeDB(b.statusDataDir+"/"+"status.db", "", dbsetup.ReducedKDFIterationsNumber)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize app db")
	}

	settings, err := b.getOrGenerateSettings(b.nodeConfig, appDB)
	if err != nil {
		return errors.Wrap(err, "Failed to generate settings")
	}
	installationID := settings.InstallationID
	b.nodeConfig.ShhextConfig.InstallationID = installationID

	multiaccountsDB, err := multiaccounts.InitializeDB(b.statusDataDir + "/" + "accounts.db")
	if err != nil {
		return errors.Wrap(err, "Failed to initialize accounts db")
	}
	multiAcc := b.createMultiAccount(b.privateKey)
	multiaccountsDB.SaveAccount(multiAcc)

	b.statusNode.SetAppDB(appDB)
	b.statusNode.SetMultiaccountsDB(multiaccountsDB)

	err = backend.StartNode(b.nodeConfig)
	if err != nil {
		return errors.Wrap(err, "Failed to start status node")
	}

	// Create a custom logger to suppress DEBUG messages
	logger, _ := zap.NewProduction()

	options := []status.Option{
		status.WithDatabase(appDB),
		status.WithWalletDatabase(walletDB),
		status.WithCustomLogger(logger),
		status.WithMailserversDatabase(mailserversDB.NewDB(appDB)),
		status.WithClusterConfig(b.nodeConfig.ClusterConfig),
		status.WithCheckingForBackupDisabled(),
		status.WithAutoMessageDisabled(),
		status.WithMultiAccounts(multiaccountsDB),
		status.WithAccount(&multiAcc),
		status.WithCommunityTokensService(b.statusNode.CommunityTokensService()),
		status.WithAccountManager(backend.AccountManager()),
	}

	ldb, _ := leveldb.Open(storage.NewMemStorage(), nil)
	cache := mailservers.NewCache(ldb)
	peerStore := mailservers.NewPeerStore(cache)

	messenger, err := status.NewMessenger(
		"status bridge messenger",
		b.privateKey,
		gethbridge.NewNodeBridge(b.statusNode.GethNode(), nil, b.statusNode.WakuV2Service()),
		installationID,
		peerStore,
		options...,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create Messenger")
	}

	messenger.SetP2PServer(b.statusNode.GethNode().Server())
	messenger.EnableBackedupMessagesProcessing()

	if err := messenger.Init(); err != nil {
		return errors.Wrap(err, "Failed to init Messenger")
	}

	if _, err := messenger.Start(); err != nil {
		return errors.Wrap(err, "Failed to start Messenger")
	}
	b.messenger = messenger

	startTime := time.Now()
	for !b.connected() && time.Since(startTime) < 10*time.Second {
		time.Sleep(1 * time.Second)
	}

	if !b.connected() {
		return fmt.Errorf("failed to create Messenger")
	}

	// Start fetching messages
	go b.fetchMessagesLoop()

	return nil
}

type BridgeTimeSource struct{}

func (t *BridgeTimeSource) GetCurrentTime() uint64 {
	return uint64(time.Now().Unix()) * 1000
}

func (b *Bstatus) JoinChannel(channel config.ChannelInfo) error {
	return b.joinCommunityChannel(channel)
}

func (b *Bstatus) joinCommunityChannel(channel config.ChannelInfo) error {
	chatID := channel.Name
	communityID := b.getCommunityIdFromFullChatId(chatID)

	if communityID == "" {
		return fmt.Errorf("wrong chat id %v", chatID)
	}

	if slices.Contains(b.joinedCommunities, communityID) {
		return nil
	}
	b.joinedCommunities = append(b.joinedCommunities, communityID)

	_, err := b.messenger.FetchCommunity(&status.FetchCommunityRequest{
		CommunityKey:    communityID,
		Shard:           nil,
		TryDatabase:     true,
		WaitForResponse: true,
	})

	if err != nil {
		return errors.Wrap(err, "Failed to fetch community")
	}

	_, err = b.messenger.JoinCommunity(context.Background(), types.Hex2Bytes(communityID), true)
	if err != nil && err != communities.ErrOrgAlreadyJoined {
		return errors.Wrap(err, "Failed to join community")
	}

	return nil
}

func (b *Bstatus) Disconnect() error {
	b.stopMessagesLoop()
	if err := b.messenger.Shutdown(); err != nil {
		return errors.Wrap(err, "Failed to stop Status messenger")
	}
	if err := b.statusNode.Stop(); err != nil {
		return errors.Wrap(err, "Failed to stop Status node")
	}
	return nil
}
