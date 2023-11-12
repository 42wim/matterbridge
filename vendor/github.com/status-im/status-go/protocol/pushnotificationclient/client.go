package pushnotificationclient

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	mrand "math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/crypto/ecies"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

// How does sending notifications work?
// 1) Every time a message is scheduled for sending, it will be received on a channel.
//    we keep track on whether we should send a push notification for this message.
// 2) Every time a message is dispatched, we check whether we should send a notification.
//    If so, we query the user info if necessary, check which installations we should be targeting
//    and notify the server if we have information about the user (i.e a token).
//    The logic is complicated by the fact that sometimes messages are batched together (datasync)
//    and the fact that sometimes we send messages to all devices (dh messages).
// 3) The server will notify us if the wrong token is used, in which case a loop will be started that
//    will re-query and re-send the notification, up to a maximum.

// How does registering works?
// We register with the server asynchronously, through a loop, that will try to make sure that
// we have registered with all the servers added, until eventually it gives up.

// A lot of the logic is complicated by the fact that waku/whisper is not req/response, so we just fire a message
// hoping to get a reply at some later stages.

const encryptedPayloadKeyLength = 16
const accessTokenKeyLength = 16
const staleQueryTimeInSeconds = 86400
const mentionInstallationID = "mention"
const oneToOneChatIDLength = 132

// maxRegistrationRetries is the maximum number of attempts we do before giving up registering with a server
const maxRegistrationRetries int64 = 12

// maxPushNotificationRetries is the maximum number of attempts before we give up sending a push notification
const maxPushNotificationRetries int64 = 4

// pushNotificationBackoffTime is the step of the exponential backoff
const pushNotificationBackoffTime int64 = 2

// RegistrationBackoffTime is the step of the exponential backoff
const RegistrationBackoffTime int64 = 15

// defaultPushNotificationsServerCount is how many push notification servers we should register with if none is selected
const defaultPushNotificationsServersCount = 3

type ServerType int

const (
	ServerTypeDefault = iota + 1
	ServerTypeCustom
)

type PushNotificationServer struct {
	PublicKey     *ecdsa.PublicKey `json:"-"`
	Registered    bool             `json:"registered,omitempty"`
	RegisteredAt  int64            `json:"registeredAt,omitempty"`
	LastRetriedAt int64            `json:"lastRetriedAt,omitempty"`
	RetryCount    int64            `json:"retryCount,omitempty"`
	AccessToken   string           `json:"accessToken,omitempty"`
	Type          ServerType       `json:"type,omitempty"`
}

func (s *PushNotificationServer) MarshalJSON() ([]byte, error) {
	type ServerAlias PushNotificationServer
	item := struct {
		*ServerAlias
		PublicKeyString string `json:"publicKey"`
	}{
		ServerAlias:     (*ServerAlias)(s),
		PublicKeyString: types.EncodeHex(crypto.FromECDSAPub(s.PublicKey)),
	}

	return json.Marshal(item)
}

type PushNotificationInfo struct {
	AccessToken     string
	InstallationID  string
	PublicKey       *ecdsa.PublicKey
	ServerPublicKey *ecdsa.PublicKey
	RetrievedAt     int64
	Version         uint64
}

type SentNotification struct {
	PublicKey        *ecdsa.PublicKey
	InstallationID   string
	LastTriedAt      int64
	RetryCount       int64
	MessageID        []byte
	ChatID           string
	NotificationType protobuf.PushNotification_PushNotificationType
	Success          bool
	Error            protobuf.PushNotificationReport_ErrorType
}

type RegistrationOptions struct {
	PublicChatIDs  []string
	MutedChatIDs   []string
	BlockedChatIDs []string
	ContactIDs     []*ecdsa.PublicKey
}

func (s *SentNotification) HashedPublicKey() []byte {
	return common.HashPublicKey(s.PublicKey)
}

type Config struct {
	// Identity is our identity key
	Identity *ecdsa.PrivateKey
	// SendEnabled indicates whether we should be sending push notifications
	SendEnabled bool
	// RemoteNotificationsEnabled is whether we should register with a remote server for push notifications
	RemoteNotificationsEnabled bool

	// AllowyFromContactsOnly indicates whether we should be receiving push notifications
	// only from contacts
	AllowFromContactsOnly bool

	// BlockMentions indicates whether we should not receive notification for mentions
	BlockMentions bool

	// InstallationID is the installation-id for this device
	InstallationID string

	Logger *zap.Logger

	// DefaultServers holds the push notification servers used by
	// default if none is selected
	DefaultServers []*ecdsa.PublicKey
}

type MessagePersistence interface {
	MessageByID(string) (*common.Message, error)
}

type Client struct {
	persistence        *Persistence
	messagePersistence MessagePersistence

	config *Config

	// lastPushNotificationRegistration is the latest known push notification version
	lastPushNotificationRegistration *protobuf.PushNotificationRegistration

	// lastContactIDs is the latest contact ids array
	lastContactIDs []*ecdsa.PublicKey

	// AccessToken is the access token that is currently being used
	AccessToken string
	// deviceToken is the device token for this device
	deviceToken string
	// TokenType is the type of token
	tokenType protobuf.PushNotificationRegistration_TokenType
	// APNTopic is the topic of the apn topic for push notification
	apnTopic string

	// randomReader only used for testing so we have deterministic encryption
	reader io.Reader

	//messageSender used to send and being notified of messages
	messageSender *common.MessageSender

	// registrationLoopQuitChan is a channel to indicate to the registration loop that should be terminating
	registrationLoopQuitChan chan struct{}

	// resendingLoopQuitChan is a channel to indicate to the send loop that should be terminating
	resendingLoopQuitChan chan struct{}

	quit chan struct{}

	// registrationSubscriptions is a list of chan of client subscribed to the registration event
	registrationSubscriptions []chan struct{}

	// pendingRegistrations is a map of pending registrations.
	// in theory we should store them in the database, but for now we can keep them in memory at
	// the cost of having to register multiple times in case the program stops
	pendingRegistrations map[string]bool
}

func New(persistence *Persistence, config *Config, sender *common.MessageSender, messagePersistence MessagePersistence) *Client {
	return &Client{
		quit:                 make(chan struct{}),
		config:               config,
		messageSender:        sender,
		messagePersistence:   messagePersistence,
		persistence:          persistence,
		pendingRegistrations: make(map[string]bool),
		reader:               rand.Reader,
	}
}

func (c *Client) Start() error {
	if c.messageSender == nil {
		return errors.New("can't start, missing message sender")
	}

	err := c.loadLastPushNotificationRegistration()
	if err != nil {
		return err
	}

	c.subscribeForMessageEvents()

	// We start even if push notifications are disabled, as we might
	// actually be sending an unregister message
	c.startRegistrationLoop()

	c.startResendingLoop()

	return nil
}

func (c *Client) Offline() {
	c.stopRegistrationLoop()
	c.stopResendingLoop()
}

func (c *Client) Online() {
	c.startRegistrationLoop()
	c.startResendingLoop()
}

func (c *Client) publishOnRegistrationSubscriptions() {
	// Publish on channels, drop if buffer is full
	for _, s := range c.registrationSubscriptions {
		select {
		case s <- struct{}{}:
		default:
			c.config.Logger.Warn("subscription channel full, dropping message")
		}
	}
}

func (c *Client) quitRegistrationSubscriptions() {
	for _, s := range c.registrationSubscriptions {
		close(s)
	}
}

func (c *Client) Stop() error {
	close(c.quit)
	c.stopRegistrationLoop()
	c.stopResendingLoop()
	c.quitRegistrationSubscriptions()
	return nil
}

// Unregister unregisters from all the servers
func (c *Client) Unregister() error {
	// stop registration loop
	c.stopRegistrationLoop()

	c.config.RemoteNotificationsEnabled = false

	registration := c.buildPushNotificationUnregisterMessage()
	err := c.saveLastPushNotificationRegistration(registration, nil)
	if err != nil {
		return err
	}

	// reset servers
	err = c.resetServers()
	if err != nil {
		return err
	}

	// and asynchronously register
	c.startRegistrationLoop()
	return nil
}

// Registered returns true if we registered with all the servers
func (c *Client) Registered() (bool, error) {
	servers, err := c.persistence.GetServers()
	if err != nil {
		return false, err
	}

	for _, s := range servers {
		if !s.Registered {
			return false, nil
		}
	}

	return true, nil
}

func (c *Client) SubscribeToRegistrations() chan struct{} {
	s := make(chan struct{}, 100)
	c.registrationSubscriptions = append(c.registrationSubscriptions, s)
	return s
}

func (c *Client) GetSentNotification(hashedPublicKey []byte, installationID string, messageID []byte) (*SentNotification, error) {
	return c.persistence.GetSentNotification(hashedPublicKey, installationID, messageID)
}

func (c *Client) GetServers() ([]*PushNotificationServer, error) {
	return c.persistence.GetServers()
}

func (c *Client) Reregister(options *RegistrationOptions) error {
	c.config.Logger.Debug("re-registering")
	if len(c.deviceToken) == 0 {
		c.config.Logger.Info("no device token, not registering")
		return nil
	}

	if !c.config.RemoteNotificationsEnabled {
		c.config.Logger.Info("remote notifications not enabled, not registering")
		return nil
	}

	return c.Register(c.deviceToken, c.apnTopic, c.tokenType, options)
}

// pickDefaultServesr picks n servers at random
func (c *Client) pickDefaultServers(servers []*ecdsa.PublicKey) []*ecdsa.PublicKey {
	// shuffle and pick n at random
	shuffledServers := make([]*ecdsa.PublicKey, len(servers))
	copy(shuffledServers, c.config.DefaultServers)
	mrand.Seed(time.Now().Unix())
	mrand.Shuffle(len(shuffledServers), func(i, j int) {
		shuffledServers[i], shuffledServers[j] = shuffledServers[j], shuffledServers[i]
	})
	// Take the min not to get an out of bounds slice
	min := len(c.config.DefaultServers)
	if min > defaultPushNotificationsServersCount {
		min = defaultPushNotificationsServersCount
	}

	return shuffledServers[:min]
}

// Register registers with all the servers
func (c *Client) Register(deviceToken, apnTopic string, tokenType protobuf.PushNotificationRegistration_TokenType, options *RegistrationOptions) error {
	// stop registration loop
	c.stopRegistrationLoop()

	c.config.RemoteNotificationsEnabled = true

	// check if we need to fallback on default servers
	currentServers, err := c.persistence.GetServers()
	if err != nil {
		return err
	}
	if len(currentServers) == 0 && len(c.config.DefaultServers) != 0 {
		c.config.Logger.Debug("servers empty, checking default servers")
		for _, s := range c.pickDefaultServers(c.config.DefaultServers) {
			err = c.AddPushNotificationsServer(s, ServerTypeDefault)
			if err != nil {
				return err
			}
		}
	}

	// reset servers
	err = c.resetServers()
	if err != nil {
		return err
	}

	c.deviceToken = deviceToken
	c.apnTopic = apnTopic
	c.tokenType = tokenType

	registration, err := c.buildPushNotificationRegistrationMessage(options)
	if err != nil {
		return err
	}

	err = c.saveLastPushNotificationRegistration(registration, options.ContactIDs)
	if err != nil {
		return err
	}

	c.startRegistrationLoop()

	return nil
}

// HandlePushNotificationRegistrationResponse should check whether the response was successful or not, retry if necessary otherwise store the result in the database
func (c *Client) HandlePushNotificationRegistrationResponse(publicKey *ecdsa.PublicKey, response *protobuf.PushNotificationRegistrationResponse) error {
	if response == nil {
		return nil
	}

	c.config.Logger.Debug("received push notification registration response", zap.Any("response", response))

	if len(response.RequestId) == 0 {
		return errors.New("empty requestId")
	}

	if !c.pendingRegistrations[hex.EncodeToString(response.RequestId)] {
		return errors.New("not for one of our requests")
	}

	// Not successful ignore for now
	if !response.Success {
		return errors.New("response was not successful")
	}

	servers, err := c.persistence.GetServersByPublicKey([]*ecdsa.PublicKey{publicKey})
	if err != nil {
		return err
	}

	// we haven't registered with this server
	if len(servers) != 1 {
		return errors.New("not registered with this server, ignoring")
	}

	server := servers[0]
	server.Registered = true
	server.RegisteredAt = time.Now().Unix()

	err = c.persistence.UpsertServer(server)
	if err != nil {
		return err
	}
	c.publishOnRegistrationSubscriptions()

	return nil
}

// processQueryInfo takes info about push notifications and validates them
func (c *Client) processQueryInfo(clientPublicKey *ecdsa.PublicKey, serverPublicKey *ecdsa.PublicKey, info *protobuf.PushNotificationQueryInfo) error {
	// make sure the public key matches
	if !bytes.Equal(info.PublicKey, common.HashPublicKey(clientPublicKey)) {
		c.config.Logger.Warn("reply for different key, ignoring")
		return errors.New("reply for a different key, ignoring")
	}

	accessToken := info.AccessToken

	// the user wants notification from contacts only, try to decrypt the access token to see if we are in their contacts
	if len(accessToken) == 0 && len(info.AllowedKeyList) != 0 {
		accessToken = c.handleAllowedKeyList(clientPublicKey, info.AllowedKeyList)

	}

	// no luck
	if len(accessToken) == 0 {
		c.config.Logger.Debug("not in the allowed key list")
		return nil
	}

	// We check the user has allowed this server to store this particular
	// access token, otherwise anyone could reply with a fake token
	// and receive notifications for a user
	if err := c.handleGrant(clientPublicKey, serverPublicKey, info.Grant, accessToken); err != nil {
		c.config.Logger.Warn("grant verification failed, ignoring", zap.Error(err))
		return err
	}

	pushNotificationInfo := &PushNotificationInfo{
		PublicKey:       clientPublicKey,
		ServerPublicKey: serverPublicKey,
		AccessToken:     accessToken,
		InstallationID:  info.InstallationId,
		Version:         info.Version,
		RetrievedAt:     time.Now().Unix(),
	}

	err := c.persistence.SavePushNotificationInfo([]*PushNotificationInfo{pushNotificationInfo})
	if err != nil {
		c.config.Logger.Error("failed to save push notifications", zap.Error(err))
		return err
	}
	return nil
}

// HandlePushNotificationQueryResponse should update the data in the database for a given user
func (c *Client) HandlePushNotificationQueryResponse(serverPublicKey *ecdsa.PublicKey, response *protobuf.PushNotificationQueryResponse) error {
	c.config.Logger.Debug("received push notification query response", zap.Any("response", response))
	if response == nil || len(response.Info) == 0 {
		return errors.New("empty response from the server")
	}

	// get the public key associated with this query
	clientPublicKey, err := c.persistence.GetQueryPublicKey(response.MessageId)
	if err != nil {
		c.config.Logger.Error("failed to query client publicKey", zap.Error(err))
		return err
	}
	if clientPublicKey == nil {
		c.config.Logger.Debug("query not found")
		return nil
	}

	// process query, make sure to validate grant as coming from the server
	for _, info := range response.Info {
		err := c.processQueryInfo(clientPublicKey, serverPublicKey, info)
		if err != nil {

			c.config.Logger.Warn("failed to process info", zap.Any("info", info), zap.Error(err))
			continue
		}
	}
	return nil

}

// HandleContactCodeAdvertisement checks if there are any info and process them
func (c *Client) HandleContactCodeAdvertisement(clientPublicKey *ecdsa.PublicKey, message *protobuf.ContactCodeAdvertisement) error {
	if message == nil {
		return nil
	}
	// nothing to do for our own pubkey
	if common.IsPubKeyEqual(clientPublicKey, &c.config.Identity.PublicKey) {
		return nil
	}

	c.config.Logger.Debug("received contact code advertisement", zap.Any("advertisement", message))
	for _, info := range message.PushNotificationInfo {
		c.config.Logger.Debug("handling push notification query info")
		serverPublicKey, err := crypto.DecompressPubkey(info.ServerPublicKey)
		if err != nil {
			c.config.Logger.Error("could not unmarshal server pubkey", zap.Binary("server-key", info.ServerPublicKey))
			return err
		}
		err = c.processQueryInfo(clientPublicKey, serverPublicKey, info)
		if err != nil {
			return err
		}
	}

	// Save query so that we won't query again to early
	// NOTE: this is not very accurate as we might fetch an historical message,
	// prolonging the time that we fetch new info.
	// Most of the times it should work fine, as if the info are stale they'd be
	// fetched again because of an error response from the push notification server
	return c.persistence.SavePushNotificationQuery(clientPublicKey, []byte(uuid.New().String()))
}

// HandlePushNotificationResponse should set the request as processed
func (c *Client) HandlePushNotificationResponse(serverKey *ecdsa.PublicKey, response *protobuf.PushNotificationResponse) error {
	if response == nil {
		return nil
	}

	messageID := response.MessageId
	c.config.Logger.Debug("received response for", zap.String("messageID", types.EncodeHex(messageID)))
	for _, report := range response.Reports {
		c.config.Logger.Debug("received response", zap.Any("report", report))
		err := c.persistence.UpdateNotificationResponse(messageID, report)
		if err != nil {
			return err
		}
	}

	// Restart resending loop, in case we need to resend some notifications
	c.stopResendingLoop()
	c.startResendingLoop()
	return nil
}

func (c *Client) RemovePushNotificationServer(publicKey *ecdsa.PublicKey) error {
	c.config.Logger.Debug("removing push notification server", zap.Any("public-key", publicKey))
	//TODO: this needs implementing. It requires unregistering from the server and
	// likely invalidate the device token of the user
	return errors.New("not implemented")
}

func (c *Client) AddPushNotificationsServer(publicKey *ecdsa.PublicKey, serverType ServerType) error {
	c.config.Logger.Debug("adding push notifications server", zap.Any("public-key", publicKey))
	currentServers, err := c.persistence.GetServers()
	if err != nil {
		return err
	}

	for _, server := range currentServers {
		if common.IsPubKeyEqual(server.PublicKey, publicKey) {
			return errors.New("push notification server already added")
		}
	}

	err = c.persistence.UpsertServer(&PushNotificationServer{
		PublicKey: publicKey,
		Type:      serverType,
	})
	if err != nil {
		return err
	}

	if c.config.RemoteNotificationsEnabled {
		c.startRegistrationLoop()
	}
	return nil
}

func (c *Client) GetPushNotificationInfo(publicKey *ecdsa.PublicKey, installationIDs []string) ([]*PushNotificationInfo, error) {
	if len(installationIDs) == 0 {
		return c.persistence.GetPushNotificationInfoByPublicKey(publicKey)
	}
	return c.persistence.GetPushNotificationInfo(publicKey, installationIDs)
}

func (c *Client) Enabled() bool {
	return c.config.RemoteNotificationsEnabled
}

func (c *Client) EnableSending() {
	c.config.SendEnabled = true
}

func (c *Client) DisableSending() {
	c.config.SendEnabled = false
}

func (c *Client) EnablePushNotificationsFromContactsOnly(options *RegistrationOptions) error {
	c.config.Logger.Debug("enabling push notification from contacts only")
	c.config.AllowFromContactsOnly = true
	if c.lastPushNotificationRegistration != nil && c.config.RemoteNotificationsEnabled {
		c.config.Logger.Debug("re-registering after enabling push notifications from contacts only")
		return c.Register(c.deviceToken, c.apnTopic, c.tokenType, options)
	}
	return nil
}

func (c *Client) DisablePushNotificationsFromContactsOnly(options *RegistrationOptions) error {
	c.config.Logger.Debug("disabling push notification from contacts only")
	c.config.AllowFromContactsOnly = false
	if c.lastPushNotificationRegistration != nil && c.config.RemoteNotificationsEnabled {
		c.config.Logger.Debug("re-registering after disabling push notifications from contacts only")
		return c.Register(c.deviceToken, c.apnTopic, c.tokenType, options)
	}
	return nil
}

func (c *Client) EnablePushNotificationsBlockMentions(options *RegistrationOptions) error {
	c.config.Logger.Debug("disabling push notifications for mentions")
	c.config.BlockMentions = true
	if c.lastPushNotificationRegistration != nil && c.config.RemoteNotificationsEnabled {
		c.config.Logger.Debug("re-registering after disabling push notifications for mentions")
		return c.Register(c.deviceToken, c.apnTopic, c.tokenType, options)
	}
	return nil
}

func (c *Client) DisablePushNotificationsBlockMentions(options *RegistrationOptions) error {
	c.config.Logger.Debug("enabling push notifications for mentions")
	c.config.BlockMentions = false
	if c.lastPushNotificationRegistration != nil && c.config.RemoteNotificationsEnabled {
		c.config.Logger.Debug("re-registering after enabling push notifications for mentions")
		return c.Register(c.deviceToken, c.apnTopic, c.tokenType, options)
	}
	return nil
}

func encryptAccessToken(plaintext []byte, key []byte, reader io.Reader) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (c *Client) encryptRegistration(publicKey *ecdsa.PublicKey, payload []byte) ([]byte, error) {
	sharedKey, err := c.generateSharedKey(publicKey)
	if err != nil {
		return nil, err
	}

	return common.Encrypt(payload, sharedKey, c.reader)
}

func (c *Client) generateSharedKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	return ecies.ImportECDSA(c.config.Identity).GenerateShared(
		ecies.ImportECDSAPublic(publicKey),
		encryptedPayloadKeyLength,
		encryptedPayloadKeyLength,
	)
}

// subscribeForMessageEvents subscribes for newly sent/scheduled messages so we can check if we need to send a push notification
func (c *Client) subscribeForMessageEvents() {
	go func() {
		c.config.Logger.Debug("subscribing for message events")
		messageEventsSubscription := c.messageSender.SubscribeToMessageEvents()
		for {
			select {
			case m, more := <-messageEventsSubscription:
				if !more {
					c.config.Logger.Debug("no more message events, quitting")
					return
				}
				switch m.Type {
				case common.MessageScheduled:
					c.config.Logger.Debug("handling message scheduled")
					if err := c.handleMessageScheduled(m); err != nil {
						c.config.Logger.Error("failed to handle message", zap.Error(err))
					}
				case common.MessageSent:
					c.config.Logger.Debug("handling message sent")
					if err := c.handleMessageSent(m); err != nil {
						c.config.Logger.Error("failed to handle message", zap.Error(err))
					}
				default:
					c.config.Logger.Warn("message event type not supported")
				}
			case <-c.quit:
				return
			}

		}
	}()
}

// loadLastPushNotificationRegistration loads from the database the last registration
func (c *Client) loadLastPushNotificationRegistration() error {
	lastRegistration, lastContactIDs, err := c.persistence.GetLastPushNotificationRegistration()
	if err != nil {
		return err
	}
	if lastRegistration == nil {
		lastRegistration = &protobuf.PushNotificationRegistration{}
	}
	c.lastContactIDs = lastContactIDs
	c.lastPushNotificationRegistration = lastRegistration
	c.deviceToken = lastRegistration.DeviceToken
	c.apnTopic = lastRegistration.ApnTopic
	c.tokenType = lastRegistration.TokenType
	return nil
}

func (c *Client) stopRegistrationLoop() {
	// stop old registration loop
	if c.registrationLoopQuitChan != nil {
		close(c.registrationLoopQuitChan)
		c.registrationLoopQuitChan = nil
	}
}

func (c *Client) stopResendingLoop() {
	// stop old registration loop
	if c.resendingLoopQuitChan != nil {
		close(c.resendingLoopQuitChan)
		c.resendingLoopQuitChan = nil
	}
}

func (c *Client) startRegistrationLoop() {
	c.stopRegistrationLoop()
	c.registrationLoopQuitChan = make(chan struct{})
	go func() {
		err := c.registrationLoop()
		if err != nil {
			c.config.Logger.Error("registration loop exited with an error", zap.Error(err))
		}
	}()
}

func (c *Client) startResendingLoop() {
	c.stopResendingLoop()
	c.resendingLoopQuitChan = make(chan struct{})
	go func() {
		err := c.resendingLoop()
		if err != nil {
			c.config.Logger.Error("resending loop exited with an error", zap.Error(err))
		}
	}()
}

// queryNotificationInfo will block and query for the client token, if force is set it
// will ignore the cool off period
func (c *Client) queryNotificationInfo(publicKey *ecdsa.PublicKey, force bool) error {
	c.config.Logger.Debug("retrieving queried at")

	// Check if we queried recently
	queriedAt, err := c.persistence.GetQueriedAt(publicKey)
	if err != nil {
		c.config.Logger.Error("failed to retrieve queried at", zap.Error(err))
		return err
	}
	c.config.Logger.Debug("checking if querying necessary")

	// Naively query again if too much time has passed.
	// Here it might not be necessary
	if force || time.Now().Unix()-queriedAt > staleQueryTimeInSeconds {
		c.config.Logger.Debug("querying info")
		err := c.queryPushNotificationInfo(publicKey)
		if err != nil {
			c.config.Logger.Error("could not query pn info", zap.Error(err))
			return err
		}
		// This is just horrible, but for now will do,
		// the issue is that we don't really know how long it will
		// take to reply, as there might be multiple servers
		// replying to us.
		// The only time we are 100% certain that we can proceed is
		// when we have non-stale info for each device, but
		// most devices are not going to be registered, so we'd still
		// have to wait the maximum amount of time allowed.
		// A better way to handle this is to set a maximum timer of say
		// 3 seconds, but act at a tick every 200ms.
		// That way we still are able to batch multiple push notifications
		// but we don't have to wait every time 3 seconds, which is wasteful
		// This probably will have to be addressed before released
		time.Sleep(3 * time.Second)
	}
	return nil
}

// handleMessageSent is called every time a message is sent
func (c *Client) handleMessageSent(e *common.MessageEvent) error {

	sentMessage := e.SentMessage
	// Ignore if we are not sending notifications
	if !c.config.SendEnabled {
		return nil
	}

	// check if it's for one of our devices, do nothing in that case
	if e.Recipient != nil && common.IsPubKeyEqual(e.Recipient, &c.config.Identity.PublicKey) {
		return nil
	}

	if sentMessage.PublicKey == nil {
		return c.handlePublicMessageSent(sentMessage)
	}
	return c.handleDirectMessageSent(sentMessage)
}

// saving to the database might happen after we fetch the message, so we retry
// for a reasonable amount of time before giving up
func (c *Client) getMessage(messageID string) (*common.Message, error) {
	retries := 0
	for retries < 10 {
		message, err := c.messagePersistence.MessageByID(messageID)
		if err == common.ErrRecordNotFound {
			retries++
			time.Sleep(300 * time.Millisecond)
			continue
		} else if err != nil {
			return nil, err
		}

		return message, nil
	}
	return nil, common.ErrRecordNotFound
}

// handlePublicMessageSent handles public messages, we notify only on mentions
func (c *Client) handlePublicMessageSent(sentMessage *common.SentMessage) error {
	// We always expect a single message, as we never batch them
	if len(sentMessage.MessageIDs) != 1 {
		return errors.New("batched public messages not handled")
	}

	messageID := sentMessage.MessageIDs[0]
	c.config.Logger.Debug("handling public messages", zap.Binary("messageID", messageID))
	tracked, err := c.persistence.TrackedMessage(messageID)
	if err != nil {
		return err
	}

	if !tracked {
		c.config.Logger.Debug("messageID not tracked, nothing to do", zap.Binary("messageID", messageID))
	}

	c.config.Logger.Debug("messageID tracked", zap.Binary("messageID", messageID))

	message, err := c.getMessage(types.EncodeHex(messageID))
	if err != nil {
		c.config.Logger.Error("could not retrieve message", zap.Error(err))
	}

	// This might happen if the user deleted their messages for example
	if message == nil {
		c.config.Logger.Warn("message not retrieved")
		return nil
	}

	c.config.Logger.Debug("message found", zap.Binary("messageID", messageID))
	for _, pkString := range message.Mentions {
		c.config.Logger.Debug("handling mention", zap.String("publickey", pkString))
		pubkeyBytes, err := types.DecodeHex(pkString)
		if err != nil {
			return err
		}

		publicKey, err := crypto.UnmarshalPubkey(pubkeyBytes)
		if err != nil {
			return err
		}

		// we use a synthetic installationID for mentions, as all devices need to be notified
		shouldNotify, err := c.shouldNotifyOn(publicKey, mentionInstallationID, messageID)
		if err != nil {
			return err
		}

		c.config.Logger.Debug("should no mention", zap.Any("publickey", shouldNotify))
		// we send the notifications and return the info of the devices notified
		infos, err := c.SendNotification(publicKey, nil, messageID, message.LocalChatID, protobuf.PushNotification_MENTION)
		if err != nil {
			return err
		}

		// mark message as sent so we don't notify again
		for _, i := range infos {
			c.config.Logger.Debug("marking as sent ", zap.Binary("mid", messageID), zap.String("id", i.InstallationID))
			if err := c.notifiedOn(publicKey, i.InstallationID, messageID, message.LocalChatID, protobuf.PushNotification_MESSAGE); err != nil {
				return err
			}

		}

	}

	return nil
}

// handleDirectMessageSent handles one to ones and private group chat messages
// It will check if we need to notify on the message, and if so it will try to
// dispatch a push notification messages might be batched, if coming
// from datasync for example.
func (c *Client) handleDirectMessageSent(sentMessage *common.SentMessage) error {
	c.config.Logger.Debug("handling direct messages", zap.Any("messageIDs", sentMessage.MessageIDs))

	publicKey := sentMessage.PublicKey

	// Collect the messageIDs we want to notify on
	var trackedMessageIDs [][]byte

	for _, messageID := range sentMessage.MessageIDs {
		tracked, err := c.persistence.TrackedMessage(messageID)
		if err != nil {
			return err
		}
		if tracked {
			trackedMessageIDs = append(trackedMessageIDs, messageID)
		}
	}

	// Nothing to do
	if len(trackedMessageIDs) == 0 {
		c.config.Logger.Debug("nothing to do for", zap.Any("messageIDs", sentMessage.MessageIDs))
		return nil
	}

	// sendToAllDevices indicates whether the message has been sent using public key encryption only
	// i.e not through the double ratchet. In that case, any device will have received it.
	sendToAllDevices := len(sentMessage.Spec.Installations) == 0

	var installationIDs []string

	anyActionableMessage := sendToAllDevices

	// Check if we should be notifiying those installations
	for _, messageID := range trackedMessageIDs {
		for _, installation := range sentMessage.Spec.Installations {
			installationID := installation.ID
			shouldNotify, err := c.shouldNotifyOn(publicKey, installationID, messageID)
			if err != nil {
				return err
			}
			if shouldNotify {
				anyActionableMessage = true
				installationIDs = append(installationIDs, installation.ID)
			}
		}
	}

	// Is there anything we should be notifying on?
	if !anyActionableMessage {
		c.config.Logger.Debug("no actionable installation IDs")
		return nil
	}

	c.config.Logger.Debug("actionable messages", zap.Any("messageIDs", trackedMessageIDs), zap.Any("installation-ids", installationIDs))

	// Get message to check chatID. Again we use the first message for simplicity, but we should send one for each chatID. Messages though are very rarely batched.
	message, err := c.getMessage(types.EncodeHex(trackedMessageIDs[0]))
	if err != nil {
		return err
	}

	// This is not the prettiest.
	// because chatIDs are asymettric, we need to check if it's a one-to-one message or a group chat message.
	// to do that we fingerprint the chatID.
	// If it's a public key, we use our own public key as chatID, which correspond to the chatID used by the other peer
	// otherwise we use the group chat ID
	var chatID string
	if len(message.ChatId) == oneToOneChatIDLength {
		chatID = types.EncodeHex(crypto.FromECDSAPub(&c.config.Identity.PublicKey))
	} else {
		// this is a group chat
		chatID = message.ChatId
	}

	// we send the notifications and return the info of the devices notified
	infos, err := c.SendNotification(publicKey, installationIDs, trackedMessageIDs[0], chatID, protobuf.PushNotification_MESSAGE)
	if err != nil {
		return err
	}

	// mark message as sent so we don't notify again
	for _, i := range infos {
		for _, messageID := range trackedMessageIDs {

			c.config.Logger.Debug("marking as sent ", zap.Binary("mid", messageID), zap.String("id", i.InstallationID))
			if err := c.notifiedOn(publicKey, i.InstallationID, messageID, chatID, protobuf.PushNotification_MESSAGE); err != nil {
				return err
			}

		}
	}

	return nil
}

// handleMessageScheduled keeps track of the message to make sure we notify on it
func (c *Client) handleMessageScheduled(e *common.MessageEvent) error {
	message := e.RawMessage
	if !message.SendPushNotification {
		return nil
	}

	// check if it's for one of our devices, do nothing in that case
	if e.Recipient != nil && common.IsPubKeyEqual(e.Recipient, &c.config.Identity.PublicKey) {
		return nil
	}

	messageID, err := types.DecodeHex(message.ID)
	if err != nil {
		return err
	}
	return c.persistence.TrackPushNotification(message.LocalChatID, messageID)
}

// shouldNotifyOn check whether we should notify a particular public-key/installation-id/message-id combination
func (c *Client) shouldNotifyOn(publicKey *ecdsa.PublicKey, installationID string, messageID []byte) (bool, error) {

	if publicKey != nil && common.IsPubKeyEqual(publicKey, &c.config.Identity.PublicKey) {
		return false, nil
	}

	if len(installationID) == 0 {
		return c.persistence.ShouldSendNotificationToAllInstallationIDs(publicKey, messageID)
	}
	return c.persistence.ShouldSendNotificationFor(publicKey, installationID, messageID)
}

// notifiedOn marks a combination of publickey/installationid/messageID/chatID/type as notified
func (c *Client) notifiedOn(publicKey *ecdsa.PublicKey, installationID string, messageID []byte, chatID string, notificationType protobuf.PushNotification_PushNotificationType) error {
	return c.persistence.UpsertSentNotification(&SentNotification{
		PublicKey:        publicKey,
		LastTriedAt:      time.Now().Unix(),
		InstallationID:   installationID,
		MessageID:        messageID,
		ChatID:           chatID,
		NotificationType: notificationType,
	})
}

func (c *Client) chatIDsHashes(chatIDs []string) [][]byte {
	var mutedChatListHashes [][]byte

	for _, chatID := range chatIDs {
		mutedChatListHashes = append(mutedChatListHashes, common.Shake256([]byte(chatID)))
	}

	return mutedChatListHashes
}

func (c *Client) encryptToken(publicKey *ecdsa.PublicKey, token []byte) ([]byte, error) {
	sharedKey, err := ecies.ImportECDSA(c.config.Identity).GenerateShared(
		ecies.ImportECDSAPublic(publicKey),
		accessTokenKeyLength,
		accessTokenKeyLength,
	)
	if err != nil {
		return nil, err
	}
	encryptedToken, err := encryptAccessToken(token, sharedKey, c.reader)
	if err != nil {
		return nil, err
	}
	return encryptedToken, nil
}

func (c *Client) decryptToken(publicKey *ecdsa.PublicKey, token []byte) ([]byte, error) {
	sharedKey, err := ecies.ImportECDSA(c.config.Identity).GenerateShared(
		ecies.ImportECDSAPublic(publicKey),
		accessTokenKeyLength,
		accessTokenKeyLength,
	)
	if err != nil {
		return nil, err
	}
	decryptedToken, err := common.Decrypt(token, sharedKey)
	if err != nil {
		return nil, err
	}
	return decryptedToken, nil
}

// allowedKeyList builds up a list of encrypted tokens, used for registering with the server
func (c *Client) allowedKeyList(token []byte, contactIDs []*ecdsa.PublicKey) ([][]byte, error) {
	// If we allow everyone, don't set the list
	if !c.config.AllowFromContactsOnly {
		return nil, nil
	}
	var encryptedTokens [][]byte
	for _, publicKey := range contactIDs {
		encryptedToken, err := c.encryptToken(publicKey, token)
		if err != nil {
			return nil, err
		}

		encryptedTokens = append(encryptedTokens, encryptedToken)

	}
	return encryptedTokens, nil
}

// getToken checks if we need to refresh the token
// and return a new one in that case. A token is refreshed only if it's not set
// or if a contact has been removed
func (c *Client) getToken(contactIDs []*ecdsa.PublicKey) string {
	if c.lastPushNotificationRegistration == nil || len(c.lastPushNotificationRegistration.AccessToken) == 0 || c.shouldRefreshToken(c.lastContactIDs, contactIDs, c.lastPushNotificationRegistration.AllowFromContactsOnly, c.config.AllowFromContactsOnly) {
		c.config.Logger.Info("refreshing access token")
		return uuid.New().String()
	}
	return c.lastPushNotificationRegistration.AccessToken
}

func (c *Client) getVersion() uint64 {
	if c.lastPushNotificationRegistration == nil {
		return 1
	}
	return c.lastPushNotificationRegistration.Version + 1
}

func (c *Client) buildPushNotificationRegistrationMessage(options *RegistrationOptions) (*protobuf.PushNotificationRegistration, error) {
	token := c.getToken(options.ContactIDs)
	allowedKeyList, err := c.allowedKeyList([]byte(token), options.ContactIDs)
	if err != nil {
		return nil, err
	}

	return &protobuf.PushNotificationRegistration{
		AccessToken:             token,
		TokenType:               c.tokenType,
		ApnTopic:                c.apnTopic,
		Version:                 c.getVersion(),
		InstallationId:          c.config.InstallationID,
		DeviceToken:             c.deviceToken,
		AllowFromContactsOnly:   c.config.AllowFromContactsOnly,
		Enabled:                 c.config.RemoteNotificationsEnabled,
		BlockedChatList:         c.chatIDsHashes(options.BlockedChatIDs),
		BlockMentions:           c.config.BlockMentions,
		AllowedMentionsChatList: c.chatIDsHashes(options.PublicChatIDs),
		AllowedKeyList:          allowedKeyList,
		MutedChatList:           c.chatIDsHashes(options.MutedChatIDs),
	}, nil
}

func (c *Client) buildPushNotificationUnregisterMessage() *protobuf.PushNotificationRegistration {
	options := &protobuf.PushNotificationRegistration{
		Version:        c.getVersion(),
		InstallationId: c.config.InstallationID,
		Unregister:     true,
	}
	return options
}

// shouldRefreshToken tells us whether we should create a new token,
// that's only necessary when a contact is removed
// or allowFromContactsOnly is enabled.
// In both cases we want to invalidate any existing token
func (c *Client) shouldRefreshToken(oldContactIDs, newContactIDs []*ecdsa.PublicKey, oldAllowFromContactsOnly, newAllowFromContactsOnly bool) bool {

	// Check if allowFromContactsOnly has just been enabled
	if !oldAllowFromContactsOnly && newAllowFromContactsOnly {
		return true
	}

	newContactIDsMap := make(map[string]bool)
	for _, pk := range newContactIDs {
		newContactIDsMap[types.EncodeHex(crypto.FromECDSAPub(pk))] = true
	}

	for _, pk := range oldContactIDs {
		if ok := newContactIDsMap[types.EncodeHex(crypto.FromECDSAPub(pk))]; !ok {
			return true
		}

	}
	return false
}

func nextServerRetry(server *PushNotificationServer) int64 {
	return server.LastRetriedAt + RegistrationBackoffTime*server.RetryCount*int64(math.Exp2(float64(server.RetryCount)))
}

func nextPushNotificationRetry(pn *SentNotification) int64 {
	return pn.LastTriedAt + pushNotificationBackoffTime*pn.RetryCount*int64(math.Exp2(float64(pn.RetryCount)))
}

// We calculate if it's too early to retry, by exponentially backing off
func shouldRetryRegisteringWithServer(server *PushNotificationServer) bool {
	return time.Now().Unix() >= nextServerRetry(server)
}

// We calculate if it's too early to retry, by exponentially backing off
func shouldRetryPushNotification(pn *SentNotification) bool {
	if pn.RetryCount > maxPushNotificationRetries {
		return false
	}
	return time.Now().Unix() >= nextPushNotificationRetry(pn)
}

func (c *Client) resetServers() error {
	servers, err := c.persistence.GetServers()
	if err != nil {
		return err
	}
	for _, server := range servers {

		// Reset server registration data
		server.Registered = false
		server.RegisteredAt = 0
		server.RetryCount = 0
		server.LastRetriedAt = time.Now().Unix()
		server.AccessToken = ""

		if err := c.persistence.UpsertServer(server); err != nil {
			return err
		}
	}

	return nil
}

// registerWithServer will register with a push notification server. This will use
// the user identity key for dispatching, as the content is in any case signed, so identity needs to be revealed.
func (c *Client) registerWithServer(registration *protobuf.PushNotificationRegistration, server *PushNotificationServer) error {
	// reset server registration data
	server.Registered = false
	server.RegisteredAt = 0
	server.RetryCount++
	server.LastRetriedAt = time.Now().Unix()
	server.AccessToken = registration.AccessToken

	// save
	if err := c.persistence.UpsertServer(server); err != nil {
		return err
	}

	// build grant for this specific server
	grant, err := c.buildGrantSignature(server.PublicKey, registration.AccessToken)
	if err != nil {
		c.config.Logger.Error("failed to build grant", zap.Error(err))
		return err
	}

	registration.Grant = grant

	// marshal message
	marshaledRegistration, err := proto.Marshal(registration)
	if err != nil {
		return err
	}

	// encrypt and dispatch message
	encryptedRegistration, err := c.encryptRegistration(server.PublicKey, marshaledRegistration)
	if err != nil {
		return err
	}
	rawMessage := common.RawMessage{
		Payload:     encryptedRegistration,
		MessageType: protobuf.ApplicationMetadataMessage_PUSH_NOTIFICATION_REGISTRATION,
		// We send on personal topic to avoid a lot of traffic on the partitioned topic
		SendOnPersonalTopic: true,
		SkipEncryptionLayer: true,
	}

	_, err = c.messageSender.SendPrivate(context.Background(), server.PublicKey, &rawMessage)

	if err != nil {
		return err
	}

	c.pendingRegistrations[hex.EncodeToString(common.Shake256(encryptedRegistration))] = true
	return nil
}

// SendNotification sends an actual notification to the push notification server.
// the notification is sent using an ephemeral key to shield the real identity of the sender
func (c *Client) SendNotification(publicKey *ecdsa.PublicKey, installationIDs []string, messageID []byte, chatID string, notificationType protobuf.PushNotification_PushNotificationType) ([]*PushNotificationInfo, error) {

	if common.IsPubKeyEqual(publicKey, &c.config.Identity.PublicKey) {
		return nil, nil
	}

	// get latest push notification infos
	err := c.queryNotificationInfo(publicKey, false)
	if err != nil {
		return nil, err
	}
	c.config.Logger.Debug("queried info")

	// retrieve info from the database
	info, err := c.GetPushNotificationInfo(publicKey, installationIDs)
	if err != nil {
		c.config.Logger.Error("could not get pn info", zap.Error(err))
		return nil, err
	}

	// naively dispatch to the first server for now
	// push notifications are only retried for now if a WRONG_TOKEN response is returned.
	// we should also retry if no response at all is received after a timeout.
	// also we send a single notification for multiple message ids, need to check with UI what's the desired behavior

	// shuffle so we don't hit the same servers all the times
	// NOTE: here's is a tradeoff, ideally we want to randomly pick a server,
	// but hit the same servers for batched notifications, for now naively
	// hit a random server
	mrand.Seed(time.Now().Unix())
	mrand.Shuffle(len(info), func(i, j int) {
		info[i], info[j] = info[j], info[i]
	})

	installationIDsMap := make(map[string]bool)

	// one info per installation id, grouped by server
	actionableInfos := make(map[string][]*PushNotificationInfo)

	for _, i := range info {

		if !installationIDsMap[i.InstallationID] {
			serverKey := hex.EncodeToString(crypto.CompressPubkey(i.ServerPublicKey))
			actionableInfos[serverKey] = append(actionableInfos[serverKey], i)
			installationIDsMap[i.InstallationID] = true
		}

	}

	c.config.Logger.Debug("actionable info", zap.Int("count", len(actionableInfos)))

	// add ephemeral key and listen to it
	ephemeralKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	_, err = c.messageSender.AddEphemeralKey(ephemeralKey)
	if err != nil {
		return nil, err
	}

	var actionedInfo []*PushNotificationInfo
	for _, infos := range actionableInfos {
		var pushNotifications []*protobuf.PushNotification
		for _, i := range infos {
			pushNotifications = append(pushNotifications, &protobuf.PushNotification{
				Type: notificationType,
				// For now we set the ChatID to our own identity key, this will work fine for blocked users
				// and muted 1-to-1 chats, but not for group chats.
				ChatId:         common.Shake256([]byte(chatID)),
				Author:         common.Shake256([]byte(types.EncodeHex(crypto.FromECDSAPub(&c.config.Identity.PublicKey)))),
				AccessToken:    i.AccessToken,
				PublicKey:      common.HashPublicKey(publicKey),
				InstallationId: i.InstallationID,
			})

		}
		request := &protobuf.PushNotificationRequest{
			MessageId: messageID,
			Requests:  pushNotifications,
		}
		serverPublicKey := infos[0].ServerPublicKey

		payload, err := proto.Marshal(request)
		if err != nil {
			return nil, err
		}

		rawMessage := common.RawMessage{
			Payload: payload,
			Sender:  ephemeralKey,
			// we skip encryption as we don't want to save any key material
			// for an ephemeral key, no need to use pfs as these are throw away keys
			SkipEncryptionLayer: true,
			MessageType:         protobuf.ApplicationMetadataMessage_PUSH_NOTIFICATION_REQUEST,
		}

		_, err = c.messageSender.SendPrivate(context.Background(), serverPublicKey, &rawMessage)

		if err != nil {
			return nil, err
		}
		actionedInfo = append(actionedInfo, infos...)
	}
	return actionedInfo, nil
}

func (c *Client) resendNotification(pn *SentNotification) error {
	c.config.Logger.Debug("resending notification")
	pn.RetryCount++
	pn.LastTriedAt = time.Now().Unix()
	err := c.persistence.UpsertSentNotification(pn)
	if err != nil {
		c.config.Logger.Error("failed to upsert notification", zap.Error(err))
		return err
	}

	// re-fetch push notification info
	err = c.queryNotificationInfo(pn.PublicKey, true)
	if err != nil {
		c.config.Logger.Error("failed to query notification info", zap.Error(err))
		return err
	}

	if err != nil {
		c.config.Logger.Error("could not get pn info", zap.Error(err))
		return err
	}

	_, err = c.SendNotification(pn.PublicKey, []string{pn.InstallationID}, pn.MessageID, pn.ChatID, pn.NotificationType)
	return err
}

// resendingLoop is a loop that is running when push notifications need to be resent, it only runs when needed, it will quit if no work is necessary.
func (c *Client) resendingLoop() error {
	for {
		c.config.Logger.Debug("running resending loop")
		var lowestNextRetry int64

		// fetch retriable notifications
		retriableNotifications, err := c.persistence.GetRetriablePushNotifications()
		if err != nil {
			c.config.Logger.Error("failed retrieving notifications, quitting resending loop", zap.Error(err))
			return err
		}

		if len(retriableNotifications) == 0 {
			c.config.Logger.Debug("no retriable notifications, quitting")
			return nil
		}

		c.config.Logger.Debug("have some retriable notifications", zap.Int("retryable-notifications", len(retriableNotifications)))

		for _, pn := range retriableNotifications {

			// check if we should retry the notification
			if shouldRetryPushNotification(pn) {
				c.config.Logger.Debug("retrying pn")
				err := c.resendNotification(pn)
				if err != nil {
					return err
				}
			}
			// set the lowest next retry if necessary
			nextRetry := nextPushNotificationRetry(pn)
			if lowestNextRetry == 0 || nextRetry < lowestNextRetry {
				lowestNextRetry = nextRetry
			}
		}

		nextRetry := lowestNextRetry - time.Now().Unix()

		// Give some room, sleep at least a second
		if nextRetry < 1 {
			nextRetry = 1
		}

		// how long should we sleep for?
		waitFor := time.Duration(nextRetry)

		select {

		case <-time.After(waitFor * time.Second):
		case <-c.resendingLoopQuitChan:
			return nil
		}
	}
}

// registrationLoop is a loop that is running when we need to register with a push notification server, it only runs when needed, it will quit if no work is necessary.
func (c *Client) registrationLoop() error {
	if c.lastPushNotificationRegistration == nil {
		return nil
	}
	for {
		c.config.Logger.Debug("running registration loop")
		servers, err := c.persistence.GetServers()
		if err != nil {
			c.config.Logger.Error("failed retrieving servers, quitting registration loop", zap.Error(err))
			return err
		}
		if len(servers) == 0 {
			c.config.Logger.Debug("nothing to do, quitting registration loop")
			return nil
		}

		var nonRegisteredServers []*PushNotificationServer
		for _, server := range servers {
			if !server.Registered && server.RetryCount < maxRegistrationRetries {
				nonRegisteredServers = append(nonRegisteredServers, server)
			}
		}

		if len(nonRegisteredServers) == 0 {
			c.config.Logger.Debug("registered with all servers, quitting registration loop")
			return nil
		}

		c.config.Logger.Debug("Trying to register with", zap.Int("servers", len(nonRegisteredServers)))

		var lowestNextRetry int64

		for _, server := range nonRegisteredServers {
			if shouldRetryRegisteringWithServer(server) {
				c.config.Logger.Debug("registering with server", zap.Any("server", server))
				err := c.registerWithServer(c.lastPushNotificationRegistration, server)
				if err != nil {
					return err
				}
			}
			nextRetry := nextServerRetry(server)
			if lowestNextRetry == 0 || nextRetry < lowestNextRetry {
				lowestNextRetry = nextRetry
			}
		}

		nextRetry := lowestNextRetry - time.Now().Unix()
		waitFor := time.Duration(nextRetry)
		c.config.Logger.Debug("Waiting for", zap.Any("wait for", waitFor))
		select {

		case <-time.After(waitFor * time.Second):
		case <-c.registrationLoopQuitChan:
			return nil
		}
	}
}

func (c *Client) saveLastPushNotificationRegistration(registration *protobuf.PushNotificationRegistration, contactIDs []*ecdsa.PublicKey) error {
	// stop registration loop
	c.stopRegistrationLoop()

	err := c.persistence.SaveLastPushNotificationRegistration(registration, contactIDs)
	if err != nil {
		return err
	}
	c.lastPushNotificationRegistration = registration
	c.lastContactIDs = contactIDs

	return nil
}

// buildGrantSignatureMaterial builds a grant for a specific server.
// We use 3 components:
// 1) The client public key. Not sure this applies to our signature scheme, but best to be conservative. https://crypto.stackexchange.com/questions/15538/given-a-message-and-signature-find-a-public-key-that-makes-the-signature-valid
// 2) The server public key
// 3) The access token
// By verifying this signature, a client can trust the server was instructed to store this access token.

func (c *Client) buildGrantSignatureMaterial(clientPublicKey *ecdsa.PublicKey, serverPublicKey *ecdsa.PublicKey, accessToken string) []byte {
	var signatureMaterial []byte
	signatureMaterial = append(signatureMaterial, crypto.CompressPubkey(clientPublicKey)...)
	signatureMaterial = append(signatureMaterial, crypto.CompressPubkey(serverPublicKey)...)
	signatureMaterial = append(signatureMaterial, []byte(accessToken)...)
	return crypto.Keccak256(signatureMaterial)
}

func (c *Client) buildGrantSignature(serverPublicKey *ecdsa.PublicKey, accessToken string) ([]byte, error) {
	signatureMaterial := c.buildGrantSignatureMaterial(&c.config.Identity.PublicKey, serverPublicKey, accessToken)
	return crypto.Sign(signatureMaterial, c.config.Identity)
}

func (c *Client) handleGrant(clientPublicKey *ecdsa.PublicKey, serverPublicKey *ecdsa.PublicKey, grant []byte, accessToken string) error {
	signatureMaterial := c.buildGrantSignatureMaterial(clientPublicKey, serverPublicKey, accessToken)
	extractedPublicKey, err := crypto.SigToPub(signatureMaterial, grant)
	if err != nil {
		return err
	}

	if !common.IsPubKeyEqual(clientPublicKey, extractedPublicKey) {
		return errors.New("invalid grant")
	}
	return nil
}

// handleAllowedKeyList will try to decrypt a token from the list, to see if we are allowed to send push notification to a given user
func (c *Client) handleAllowedKeyList(publicKey *ecdsa.PublicKey, allowedKeyList [][]byte) string {
	c.config.Logger.Debug("handling allowed key list")
	for _, encryptedToken := range allowedKeyList {
		token, err := c.decryptToken(publicKey, encryptedToken)
		if err != nil {
			c.config.Logger.Warn("could not decrypt token", zap.Error(err))
			continue
		}
		c.config.Logger.Debug("decrypted token")
		return string(token)
	}
	return ""
}

func (c *Client) MyPushNotificationQueryInfo() ([]*protobuf.PushNotificationQueryInfo, error) {

	// Nothing to do
	if c.lastPushNotificationRegistration == nil || c.lastPushNotificationRegistration.Unregister {
		return nil, nil

	}
	var response []*protobuf.PushNotificationQueryInfo
	servers, err := c.persistence.GetServers()
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		// ignore non-registered servers
		if !server.Registered {
			continue
		}
		// build grant for this specific server
		grant, err := c.buildGrantSignature(server.PublicKey, c.lastPushNotificationRegistration.AccessToken)
		if err != nil {
			c.config.Logger.Error("failed to build grant", zap.Error(err))
			return nil, err
		}

		queryInfo := &protobuf.PushNotificationQueryInfo{
			InstallationId: c.config.InstallationID,
			// is this the right key?
			PublicKey:       common.HashPublicKey(&c.config.Identity.PublicKey),
			Version:         c.lastPushNotificationRegistration.Version,
			Grant:           grant,
			ServerPublicKey: crypto.CompressPubkey(server.PublicKey),
		}
		if c.lastPushNotificationRegistration.AllowFromContactsOnly {
			queryInfo.AllowedKeyList = c.lastPushNotificationRegistration.AllowedKeyList
		} else {
			queryInfo.AccessToken = c.lastPushNotificationRegistration.AccessToken
		}
		response = append(response, queryInfo)
	}
	return response, nil
}

// queryPushNotificationInfo sends a message to any server who has the given user registered.
// it uses an ephemeral key so the identity of the client querying is not disclosed
func (c *Client) queryPushNotificationInfo(publicKey *ecdsa.PublicKey) error {
	hashedPublicKey := common.HashPublicKey(publicKey)
	query := &protobuf.PushNotificationQuery{
		PublicKeys: [][]byte{hashedPublicKey},
	}
	encodedMessage, err := proto.Marshal(query)
	if err != nil {
		return err
	}

	ephemeralKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload: encodedMessage,
		Sender:  ephemeralKey,
		// we don't want to wrap in an encryption layer message
		SkipEncryptionLayer: true,
		MessageType:         protobuf.ApplicationMetadataMessage_PUSH_NOTIFICATION_QUERY,
	}

	_, err = c.messageSender.AddEphemeralKey(ephemeralKey)
	if err != nil {
		return err
	}

	// this is the topic of message
	encodedPublicKey := hex.EncodeToString(hashedPublicKey)
	messageID, err := c.messageSender.SendPublic(context.Background(), encodedPublicKey, rawMessage)

	if err != nil {
		return err
	}

	return c.persistence.SavePushNotificationQuery(publicKey, messageID)
}
