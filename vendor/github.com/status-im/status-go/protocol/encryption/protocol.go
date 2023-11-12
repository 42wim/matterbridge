package encryption

import (
	"bytes"
	"crypto/ecdsa"
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"

	"github.com/status-im/status-go/protocol/encryption/multidevice"
	"github.com/status-im/status-go/protocol/encryption/publisher"
	"github.com/status-im/status-go/protocol/encryption/sharedsecret"
)

//go:generate protoc --go_out=. ./protocol_message.proto

const (
	protocolVersion                = 1
	sharedSecretNegotiationVersion = 1
	partitionedTopicMinVersion     = 1
	defaultMinVersion              = 0
)

type PartitionTopicMode int

const (
	PartitionTopicNoSupport PartitionTopicMode = iota
	PartitionTopicV1
)

type ProtocolMessageSpec struct {
	Message *ProtocolMessage
	// Installations is the targeted devices
	Installations []*multidevice.Installation
	// SharedSecret is a shared secret established among the installations
	SharedSecret *sharedsecret.Secret
	// AgreedSecret indicates whether the shared secret has been agreed
	AgreedSecret bool
	// Public means that the spec contains a public wrapped message
	Public bool
}

func (p *ProtocolMessageSpec) MinVersion() uint32 {
	if len(p.Installations) == 0 {
		return defaultMinVersion
	}

	version := p.Installations[0].Version

	for _, installation := range p.Installations[1:] {
		if installation.Version < version {
			version = installation.Version
		}
	}
	return version
}

func (p *ProtocolMessageSpec) PartitionedTopicMode() PartitionTopicMode {
	if p.MinVersion() >= partitionedTopicMinVersion {
		return PartitionTopicV1
	}
	return PartitionTopicNoSupport
}

type Protocol struct {
	encryptor     *encryptor
	secret        *sharedsecret.SharedSecret
	multidevice   *multidevice.Multidevice
	publisher     *publisher.Publisher
	subscriptions *Subscriptions

	logger *zap.Logger
}

var (
	// ErrNoPayload means that there was no payload found in the received protocol message.
	ErrNoPayload    = errors.New("no payload")
	ErrNoRatchetKey = errors.New("no ratchet key for given keyID")
)

// New creates a new ProtocolService instance
func New(
	db *sql.DB,
	installationID string,
	logger *zap.Logger,
) *Protocol {
	return NewWithEncryptorConfig(
		db,
		installationID,
		defaultEncryptorConfig(installationID, logger),
		logger,
	)
}

// DB and migrations are shared between encryption package
// and its sub-packages.
func NewWithEncryptorConfig(
	db *sql.DB,
	installationID string,
	encryptorConfig encryptorConfig,
	logger *zap.Logger,
) *Protocol {
	return &Protocol{
		encryptor: newEncryptor(db, encryptorConfig),
		secret:    sharedsecret.New(db, logger),
		multidevice: multidevice.New(db, &multidevice.Config{
			MaxInstallations: 3,
			ProtocolVersion:  protocolVersion,
			InstallationID:   installationID,
		}),
		publisher: publisher.New(logger),
		logger:    logger.With(zap.Namespace("Protocol")),
	}
}

type Subscriptions struct {
	SharedSecrets   []*sharedsecret.Secret
	SendContactCode <-chan struct{}
	Quit            chan struct{}
}

func (p *Protocol) Start(myIdentity *ecdsa.PrivateKey) (*Subscriptions, error) {
	// Propagate currently cached shared secrets.
	secrets, err := p.secret.All()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all secrets")
	}
	p.subscriptions = &Subscriptions{
		SharedSecrets:   secrets,
		SendContactCode: p.publisher.Start(),
		Quit:            make(chan struct{}),
	}
	return p.subscriptions, nil
}

func (p *Protocol) Stop() error {
	p.publisher.Stop()
	if p.subscriptions != nil {
		close(p.subscriptions.Quit)
	}
	return nil
}

func (p *Protocol) addBundle(myIdentityKey *ecdsa.PrivateKey, msg *ProtocolMessage) error {
	// Get a bundle
	installations, err := p.multidevice.GetOurActiveInstallations(&myIdentityKey.PublicKey)
	if err != nil {
		return err
	}

	bundle, err := p.encryptor.CreateBundle(myIdentityKey, installations)
	if err != nil {
		return err
	}

	msg.Bundles = []*Bundle{bundle}

	return nil
}

// BuildPublicMessage marshals a public chat message given the user identity private key and a payload
func (p *Protocol) BuildPublicMessage(myIdentityKey *ecdsa.PrivateKey, payload []byte) (*ProtocolMessageSpec, error) {
	// Build message not encrypted
	message := &ProtocolMessage{
		InstallationId: p.encryptor.config.InstallationID,
		PublicMessage:  payload,
	}

	err := p.addBundle(myIdentityKey, message)
	if err != nil {
		return nil, err
	}

	return &ProtocolMessageSpec{Message: message, Public: true}, nil
}

// BuildEncryptedMessage returns a 1:1 chat message and optionally a negotiated topic given the user identity private key, the recipient's public key, and a payload
func (p *Protocol) BuildEncryptedMessage(myIdentityKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, payload []byte) (*ProtocolMessageSpec, error) {

	// Get recipients installations.
	activeInstallations, err := p.multidevice.GetActiveInstallations(publicKey)
	if err != nil {
		return nil, err
	}

	// Encrypt payload
	encryptedMessagesByInstalls, installations, err := p.encryptor.EncryptPayload(publicKey, myIdentityKey, activeInstallations, payload)
	if err != nil {
		return nil, err
	}

	// Build message
	message := &ProtocolMessage{
		InstallationId:   p.encryptor.config.InstallationID,
		EncryptedMessage: encryptedMessagesByInstalls,
	}

	err = p.addBundle(myIdentityKey, message)
	if err != nil {
		return nil, err
	}

	// Check who we are sending the message to, and see if we have a shared secret
	// across devices
	var installationIDs []string
	for installationID := range message.GetEncryptedMessage() {
		if installationID != noInstallationID {
			installationIDs = append(installationIDs, installationID)
		}
	}

	sharedSecret, agreed, err := p.secret.Agreed(myIdentityKey, p.encryptor.config.InstallationID, publicKey, installationIDs)
	if err != nil {
		return nil, err
	}

	spec := &ProtocolMessageSpec{
		SharedSecret:  sharedSecret,
		AgreedSecret:  agreed,
		Message:       message,
		Installations: installations,
	}
	return spec, nil
}

func (p *Protocol) GenerateHashRatchetKey(groupID []byte) (*HashRatchetKeyCompatibility, error) {
	return p.encryptor.GenerateHashRatchetKey(groupID)
}

func (p *Protocol) GetAllHREncodedKeys(groupID []byte) ([]byte, error) {
	keys, err := p.encryptor.persistence.GetKeysForGroup(groupID)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}

	return p.GetMarshaledHREncodedKeys(groupID, keys)
}

// GetKeyIDsForGroup returns a slice of key IDs belonging to a given group ID
func (p *Protocol) GetKeysForGroup(groupID []byte) ([]*HashRatchetKeyCompatibility, error) {
	return p.encryptor.persistence.GetKeysForGroup(groupID)
}

func (p *Protocol) GetHREncodedKeys(groupID []byte, ratchets []*HashRatchetKeyCompatibility) *HRKeys {
	keys := &HRKeys{}
	for _, ratchet := range ratchets {
		key := &HRKey{
			DeprecatedKeyId: ratchet.DeprecatedKeyID(),
			Key:             ratchet.Key,
			Timestamp:       ratchet.Timestamp,
		}
		keys.Keys = append(keys.Keys, key)
	}

	return keys
}

func (p *Protocol) GetMarshaledHREncodedKeys(groupID []byte, ratchets []*HashRatchetKeyCompatibility) ([]byte, error) {
	keys := p.GetHREncodedKeys(groupID, ratchets)
	return proto.Marshal(keys)
}

// BuildHashRatchetRekeyGroup builds a public message
// with the new key
func (p *Protocol) BuildHashRatchetReKeyGroupMessage(myIdentityKey *ecdsa.PrivateKey, recipients []*ecdsa.PublicKey, groupID []byte, payload []byte, ratchet *HashRatchetKeyCompatibility) (*ProtocolMessageSpec, error) {

	var err error
	if ratchet == nil {
		ratchet, err = p.GenerateHashRatchetKey(groupID)
		if err != nil {
			return nil, err
		}
	}

	message, err := buildGroupRekeyMessage(myIdentityKey, groupID, ratchet.Timestamp, ratchet.Key, recipients)
	if err != nil {
		return nil, err
	}

	keys := &HRKeys{
		RekeyGroup: message,
	}
	spec := &ProtocolMessageSpec{
		Public: true,
		Message: &ProtocolMessage{
			InstallationId: p.encryptor.config.InstallationID,
			EncryptedMessage: map[string]*EncryptedMessageProtocol{noInstallationID: &EncryptedMessageProtocol{
				HRHeader: &HRHeader{
					SeqNo:   0,
					GroupId: groupID,
					Keys:    keys,
				},
				Payload: payload,
			},
			},
		},
	}

	return spec, nil
}

// BuildHashRatchetKeyExchangeMessage builds a 1:1 message
// containing newly generated hash ratchet key
func (p *Protocol) BuildHashRatchetKeyExchangeMessage(myIdentityKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, groupID []byte, ratchets []*HashRatchetKeyCompatibility) (*ProtocolMessageSpec, error) {

	keys := p.GetHREncodedKeys(groupID, ratchets)

	encodedKeys, err := proto.Marshal(keys)
	if err != nil {
		return nil, err
	}

	response, err := p.BuildEncryptedMessage(myIdentityKey, publicKey, encodedKeys)
	if err != nil {
		return nil, err
	}

	// Loop through installations and assign HRHeader
	// SeqNo=0 has a special meaning for HandleMessage
	// and signifies a message with hash ratchet key payload
	for _, v := range response.Message.EncryptedMessage {
		v.HRHeader = &HRHeader{
			SeqNo:   0,
			GroupId: groupID,
			Keys:    keys,
		}

	}

	return response, err
}

func (p *Protocol) BuildHashRatchetKeyExchangeMessageWithPayload(myIdentityKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, groupID []byte, ratchets []*HashRatchetKeyCompatibility, payload []byte) (*ProtocolMessageSpec, error) {

	keys := p.GetHREncodedKeys(groupID, ratchets)

	response, err := p.BuildEncryptedMessage(myIdentityKey, publicKey, payload)
	if err != nil {
		return nil, err
	}

	// Loop through installations and assign HRHeader
	// SeqNo=0 has a special meaning for HandleMessage
	// and signifies a message with hash ratchet key payload
	for _, v := range response.Message.EncryptedMessage {
		v.HRHeader = &HRHeader{
			SeqNo:   0,
			GroupId: groupID,
			Keys:    keys,
		}

	}

	return response, err
}

func (p *Protocol) GetCurrentKeyForGroup(groupID []byte) (*HashRatchetKeyCompatibility, error) {
	return p.encryptor.persistence.GetCurrentKeyForGroup(groupID)

}

// BuildHashRatchetMessage returns a hash ratchet chat message
func (p *Protocol) BuildHashRatchetMessage(groupID []byte, payload []byte) (*ProtocolMessageSpec, error) {

	ratchet, err := p.encryptor.persistence.GetCurrentKeyForGroup(groupID)
	if err != nil {
		return nil, err
	}

	// Encrypt payload
	encryptedMessagesByInstalls, err := p.encryptor.EncryptHashRatchetPayload(ratchet, payload)
	if err != nil {
		return nil, err
	}

	// Build message
	message := &ProtocolMessage{
		InstallationId:   p.encryptor.config.InstallationID,
		EncryptedMessage: encryptedMessagesByInstalls,
	}

	spec := &ProtocolMessageSpec{
		Message: message,
	}
	return spec, nil
}

func (p *Protocol) GetKeyExMessageSpecs(groupID []byte, identity *ecdsa.PrivateKey, recipients []*ecdsa.PublicKey, forceRekey bool) ([]*ProtocolMessageSpec, error) {
	var ratchets []*HashRatchetKeyCompatibility
	var err error
	if !forceRekey {
		ratchets, err = p.encryptor.persistence.GetKeysForGroup(groupID)
		if err != nil {
			return nil, err
		}
	}
	if len(ratchets) == 0 || forceRekey {
		ratchet, err := p.GenerateHashRatchetKey(groupID)
		if err != nil {
			return nil, err
		}
		ratchets = []*HashRatchetKeyCompatibility{ratchet}
	}
	specs := make([]*ProtocolMessageSpec, len(recipients))
	for i, recipient := range recipients {
		keyExMsg, err := p.BuildHashRatchetKeyExchangeMessage(identity, recipient, groupID, ratchets)
		if err != nil {
			return nil, err
		}
		specs[i] = keyExMsg

	}

	return specs, nil
}

// BuildDHMessage builds a message with DH encryption so that it can be decrypted by any other device.
func (p *Protocol) BuildDHMessage(myIdentityKey *ecdsa.PrivateKey, destination *ecdsa.PublicKey, payload []byte) (*ProtocolMessageSpec, error) {
	// Encrypt payload
	encryptionResponse, err := p.encryptor.EncryptPayloadWithDH(destination, payload)
	if err != nil {
		return nil, err
	}

	// Build message
	message := &ProtocolMessage{
		InstallationId:   p.encryptor.config.InstallationID,
		EncryptedMessage: encryptionResponse,
	}

	err = p.addBundle(myIdentityKey, message)
	if err != nil {
		return nil, err
	}

	return &ProtocolMessageSpec{Message: message}, nil
}

// ProcessPublicBundle processes a received X3DH bundle.
func (p *Protocol) ProcessPublicBundle(myIdentityKey *ecdsa.PrivateKey, bundle *Bundle) ([]*multidevice.Installation, error) {
	logger := p.logger.With(zap.String("site", "ProcessPublicBundle"))

	if err := p.encryptor.ProcessPublicBundle(myIdentityKey, bundle); err != nil {
		return nil, err
	}

	installations, enabled, err := p.recoverInstallationsFromBundle(myIdentityKey, bundle)
	if err != nil {
		return nil, err
	}

	// TODO(adam): why do we add installations using identity obtained from GetIdentity()
	// instead of the output of crypto.CompressPubkey()? I tried the second option
	// and the unit tests TestTopic and TestMaxDevices fail.
	identityFromBundle := bundle.GetIdentity()
	theirIdentity, err := ExtractIdentity(bundle)
	if err != nil {
		logger.Panic("unrecoverable error extracting identity", zap.Error(err))
	}
	compressedIdentity := crypto.CompressPubkey(theirIdentity)
	if !bytes.Equal(identityFromBundle, compressedIdentity) {
		logger.Panic("identity from bundle and compressed are not equal")
	}

	return p.multidevice.AddInstallations(bundle.GetIdentity(), bundle.GetTimestamp(), installations, enabled)
}

func (p *Protocol) GetMultiDevice() *multidevice.Multidevice {
	return p.multidevice
}

// recoverInstallationsFromBundle extracts installations from the bundle.
// It returns extracted installations and true if the installations
// are ours, i.e. the bundle was created by our identity key.
func (p *Protocol) recoverInstallationsFromBundle(myIdentityKey *ecdsa.PrivateKey, bundle *Bundle) ([]*multidevice.Installation, bool, error) {
	var installations []*multidevice.Installation

	theirIdentity, err := ExtractIdentity(bundle)
	if err != nil {
		return nil, false, err
	}

	myIdentityStr := fmt.Sprintf("0x%x", crypto.FromECDSAPub(&myIdentityKey.PublicKey))
	theirIdentityStr := fmt.Sprintf("0x%x", crypto.FromECDSAPub(theirIdentity))
	// Any device from other peers will be considered enabled, ours needs to
	// be explicitly enabled.
	enabled := theirIdentityStr != myIdentityStr
	signedPreKeys := bundle.GetSignedPreKeys()

	for installationID, signedPreKey := range signedPreKeys {
		if installationID != p.multidevice.InstallationID() {
			installations = append(installations, &multidevice.Installation{
				Identity: theirIdentityStr,
				ID:       installationID,
				Version:  signedPreKey.GetProtocolVersion(),
			})
		}
	}

	return installations, enabled, nil
}

// GetBundle retrieves or creates a X3DH bundle, given a private identity key.
func (p *Protocol) GetBundle(myIdentityKey *ecdsa.PrivateKey) (*Bundle, error) {
	installations, err := p.multidevice.GetOurActiveInstallations(&myIdentityKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return p.encryptor.CreateBundle(myIdentityKey, installations)
}

// EnableInstallation enables an installation for multi-device sync.
func (p *Protocol) EnableInstallation(myIdentityKey *ecdsa.PublicKey, installationID string) error {
	return p.multidevice.EnableInstallation(myIdentityKey, installationID)
}

// DisableInstallation disables an installation for multi-device sync.
func (p *Protocol) DisableInstallation(myIdentityKey *ecdsa.PublicKey, installationID string) error {
	return p.multidevice.DisableInstallation(myIdentityKey, installationID)
}

// GetOurInstallations returns all the installations available given an identity
func (p *Protocol) GetOurInstallations(myIdentityKey *ecdsa.PublicKey) ([]*multidevice.Installation, error) {
	return p.multidevice.GetOurInstallations(myIdentityKey)
}

// GetOurActiveInstallations returns all the active installations available given an identity
func (p *Protocol) GetOurActiveInstallations(myIdentityKey *ecdsa.PublicKey) ([]*multidevice.Installation, error) {
	return p.multidevice.GetOurActiveInstallations(myIdentityKey)
}

// SetInstallationMetadata sets the metadata for our own installation
func (p *Protocol) SetInstallationMetadata(myIdentityKey *ecdsa.PublicKey, installationID string, data *multidevice.InstallationMetadata) error {
	return p.multidevice.SetInstallationMetadata(myIdentityKey, installationID, data)
}

// SetInstallationName sets the metadata for our own installation
func (p *Protocol) SetInstallationName(myIdentityKey *ecdsa.PublicKey, installationID string, name string) error {
	return p.multidevice.SetInstallationName(myIdentityKey, installationID, name)
}

// GetPublicBundle retrieves a public bundle given an identity
func (p *Protocol) GetPublicBundle(theirIdentityKey *ecdsa.PublicKey) (*Bundle, error) {
	installations, err := p.multidevice.GetActiveInstallations(theirIdentityKey)
	if err != nil {
		return nil, err
	}
	return p.encryptor.GetPublicBundle(theirIdentityKey, installations)
}

// ConfirmMessageProcessed confirms and deletes message keys for the given messages
func (p *Protocol) ConfirmMessageProcessed(messageID []byte) error {
	logger := p.logger.With(zap.String("site", "ConfirmMessageProcessed"))
	logger.Debug("confirming message", zap.String("messageID", types.EncodeHex(messageID)))
	return p.encryptor.ConfirmMessageProcessed(messageID)
}

type HashRatchetInfo struct {
	GroupID []byte
	KeyID   []byte
}
type DecryptMessageResponse struct {
	DecryptedMessage []byte
	Installations    []*multidevice.Installation
	SharedSecrets    []*sharedsecret.Secret
	HashRatchetInfo  []*HashRatchetInfo
}

func (p *Protocol) HandleHashRatchetKeysPayload(groupID, encodedKeys []byte, myIdentityKey *ecdsa.PrivateKey, theirIdentityKey *ecdsa.PublicKey) ([]*HashRatchetInfo, error) {

	keys := &HRKeys{}
	err := proto.Unmarshal(encodedKeys, keys)
	if err != nil {
		return nil, err
	}
	return p.HandleHashRatchetKeys(groupID, keys, myIdentityKey, theirIdentityKey)
}

func (p *Protocol) HandleHashRatchetKeys(groupID []byte, keys *HRKeys, myIdentityKey *ecdsa.PrivateKey, theirIdentityKey *ecdsa.PublicKey) ([]*HashRatchetInfo, error) {
	if keys == nil {
		return nil, nil
	}

	var info []*HashRatchetInfo

	for _, key := range keys.Keys {
		ratchet := &HashRatchetKeyCompatibility{
			GroupID:   groupID,
			Timestamp: key.Timestamp,
			Key:       key.Key,
		}

		// If there's no timestamp, is coming from an older client
		if key.Timestamp == 0 {
			ratchet.Timestamp = uint64(key.DeprecatedKeyId)
		}
		keyID, err := ratchet.GetKeyID()
		if err != nil {
			return nil, err
		}
		p.logger.Debug("retrieved keys", zap.String("keyID", types.Bytes2Hex(keyID)))

		// Payload contains hash ratchet key
		err = p.encryptor.persistence.SaveHashRatchetKey(ratchet)
		if err != nil {
			return nil, err
		}
		info = append(info, &HashRatchetInfo{GroupID: groupID, KeyID: keyID})
	}

	if keys.RekeyGroup != nil {
		if keys.RekeyGroup.Timestamp == 0 {
			return nil, errors.New("timestamp can't be nil")
		}

		encryptionKey, err := decryptGroupRekeyMessage(myIdentityKey, theirIdentityKey, keys.RekeyGroup)
		if err != nil {
			return nil, err
		}

		if len(encryptionKey) != 0 {

			ratchet := &HashRatchetKeyCompatibility{
				GroupID:   groupID,
				Timestamp: keys.RekeyGroup.Timestamp,
				Key:       encryptionKey,
			}

			keyID, err := ratchet.GetKeyID()
			if err != nil {
				return nil, err
			}
			p.logger.Debug("retrieved group keys", zap.String("keyID", types.Bytes2Hex(keyID)))
			// Payload contains hash ratchet key
			err = p.encryptor.persistence.SaveHashRatchetKey(ratchet)
			if err != nil {
				return nil, err
			}

			info = append(info, &HashRatchetInfo{GroupID: groupID, KeyID: keyID})

		}
	}

	return info, nil
}

// HandleMessage unmarshals a message and processes it, decrypting it if it is a 1:1 message.
func (p *Protocol) HandleMessage(
	myIdentityKey *ecdsa.PrivateKey,
	theirPublicKey *ecdsa.PublicKey,
	protocolMessage *ProtocolMessage,
	messageID []byte,
) (*DecryptMessageResponse, error) {
	logger := p.logger.With(zap.String("site", "HandleMessage"))
	response := &DecryptMessageResponse{}

	logger.Debug("received a protocol message",
		zap.String("sender-public-key",
			types.EncodeHex(crypto.FromECDSAPub(theirPublicKey))),
		zap.String("my-installation-id", p.encryptor.config.InstallationID),
		zap.String("messageID", types.EncodeHex(messageID)))

	if p.encryptor == nil {
		return nil, errors.New("encryption service not initialized")
	}

	// Process bundles
	for _, bundle := range protocolMessage.GetBundles() {
		// Should we stop processing if the bundle cannot be verified?
		newInstallations, err := p.ProcessPublicBundle(myIdentityKey, bundle)
		if err != nil {
			return nil, err
		}
		response.Installations = newInstallations
	}

	// Check if it's a public message
	if publicMessage := protocolMessage.GetPublicMessage(); publicMessage != nil {
		// Nothing to do, as already in cleartext
		response.DecryptedMessage = publicMessage
		return response, nil
	}

	// Decrypt message
	if encryptedMessage := protocolMessage.GetEncryptedMessage(); encryptedMessage != nil {
		message, err := p.encryptor.DecryptPayload(
			myIdentityKey,
			theirPublicKey,
			protocolMessage.GetInstallationId(),
			encryptedMessage,
			messageID,
		)

		if err == ErrHashRatchetGroupIDNotFound {
			msg := p.encryptor.GetMessage(encryptedMessage)

			if msg != nil {
				if header := msg.GetHRHeader(); header != nil {
					response.HashRatchetInfo = append(response.HashRatchetInfo, &HashRatchetInfo{GroupID: header.GroupId, KeyID: header.KeyId})
				}
			}
			return response, err
		}

		if err != nil {
			return nil, err
		}

		dmProtocol := encryptedMessage[p.encryptor.config.InstallationID]
		if dmProtocol == nil {
			dmProtocol = encryptedMessage[noInstallationID]
		}
		if dmProtocol != nil {
			hrHeader := dmProtocol.HRHeader
			if hrHeader != nil && hrHeader.SeqNo == 0 {
				var hashRatchetKeys []*HashRatchetInfo
				if hrHeader.Keys != nil {
					hashRatchetKeys, err = p.HandleHashRatchetKeys(hrHeader.GroupId, hrHeader.Keys, myIdentityKey, theirPublicKey)
					if err != nil {
						return nil, err
					}

				} else {
					// For backward compatibility
					hashRatchetKeys, err = p.HandleHashRatchetKeysPayload(hrHeader.GroupId, message, myIdentityKey, theirPublicKey)
					if err != nil {
						return nil, err
					}
				}
				response.HashRatchetInfo = hashRatchetKeys
			}
		}

		bundles := protocolMessage.GetBundles()
		version := getProtocolVersion(bundles, protocolMessage.GetInstallationId())
		if version >= sharedSecretNegotiationVersion {
			sharedSecret, err := p.secret.Generate(myIdentityKey, theirPublicKey, protocolMessage.GetInstallationId())
			if err != nil {
				return nil, err
			}

			response.SharedSecrets = []*sharedsecret.Secret{sharedSecret}
		}
		response.DecryptedMessage = message
		return response, nil
	}

	// Return error
	return nil, ErrNoPayload
}

func (p *Protocol) ShouldAdvertiseBundle(publicKey *ecdsa.PublicKey, time int64) (bool, error) {
	return p.publisher.ShouldAdvertiseBundle(publicKey, time)
}

func (p *Protocol) ConfirmBundleAdvertisement(publicKey *ecdsa.PublicKey, time int64) {
	p.publisher.SetLastAck(publicKey, time)
}

func (p *Protocol) BuildBundleAdvertiseMessage(myIdentityKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (*ProtocolMessageSpec, error) {
	return p.BuildDHMessage(myIdentityKey, publicKey, nil)
}

func getProtocolVersion(bundles []*Bundle, installationID string) uint32 {
	if installationID == "" {
		return defaultMinVersion
	}

	for _, bundle := range bundles {
		if bundle != nil {
			signedPreKeys := bundle.GetSignedPreKeys()
			if signedPreKeys == nil {
				continue
			}

			signedPreKey := signedPreKeys[installationID]
			if signedPreKey == nil {
				return defaultMinVersion
			}

			return signedPreKey.GetProtocolVersion()
		}
	}

	return defaultMinVersion
}

func (p *Protocol) EncryptWithHashRatchet(groupID []byte, payload []byte) ([]byte, *HashRatchetKeyCompatibility, uint32, error) {
	ratchet, err := p.encryptor.persistence.GetCurrentKeyForGroup(groupID)
	if err != nil {
		return nil, nil, 0, err
	}

	encryptedPayload, newSeqNo, err := p.encryptor.EncryptWithHR(ratchet, payload)
	if err != nil {
		return nil, nil, 0, err
	}

	return encryptedPayload, ratchet, newSeqNo, nil
}

func (p *Protocol) DecryptWithHashRatchet(keyID []byte, seqNo uint32, payload []byte) ([]byte, error) {
	ratchet, err := p.encryptor.persistence.GetHashRatchetKeyByID(keyID)
	if err != nil {
		return nil, err
	}
	if ratchet == nil {
		return nil, ErrNoRatchetKey
	}

	return p.encryptor.DecryptWithHR(ratchet, seqNo, payload)
}
