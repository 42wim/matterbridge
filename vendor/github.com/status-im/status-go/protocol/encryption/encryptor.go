package encryption

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	dr "github.com/status-im/doubleratchet"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/crypto/ecies"
	"github.com/status-im/status-go/eth-node/types"

	"github.com/status-im/status-go/protocol/encryption/multidevice"
)

var (
	errSessionNotFound = errors.New("session not found")
	ErrDeviceNotFound  = errors.New("device not found")
	// ErrNotPairedDevice means that we received a message signed with our public key
	// but from a device that has not been paired.
	// This should not happen because the protocol forbids sending a message to
	// non-paired devices, however, in theory it is possible to receive such a message.
	ErrNotPairedDevice            = errors.New("received a message from not paired device")
	ErrHashRatchetSeqNoTooHigh    = errors.New("Hash ratchet seq no is too high")
	ErrHashRatchetGroupIDNotFound = errors.New("Hash ratchet group id not found")
	ErrNoEncryptionKey            = errors.New("no encryption key found for the community")
)

// If we have no bundles, we use a constant so that the message can reach any device.
const (
	noInstallationID         = "none"
	maxHashRatchetSeqNoDelta = 100000
)

type confirmationData struct {
	header *dr.MessageHeader
	drInfo *RatchetInfo
}

// encryptor defines a service that is responsible for the encryption aspect of the protocol.
type encryptor struct {
	persistence *sqlitePersistence
	config      encryptorConfig
	messageIDs  map[string]*confirmationData
	mutex       sync.Mutex
	logger      *zap.Logger
}

type encryptorConfig struct {
	InstallationID string
	// Max number of installations we keep synchronized.
	MaxInstallations int
	// How many consecutive messages can be skipped in the receiving chain.
	MaxSkip int
	// Any message with seqNo <= currentSeq - maxKeep will be deleted.
	MaxKeep int
	// How many keys do we store in total per session.
	MaxMessageKeysPerSession int
	// How long before we refresh the interval in milliseconds
	BundleRefreshInterval int64
	// The logging object
	Logger *zap.Logger
}

// defaultEncryptorConfig returns the default values used by the encryption service
func defaultEncryptorConfig(installationID string, logger *zap.Logger) encryptorConfig {
	if logger == nil {
		logger = zap.NewNop()
	}

	return encryptorConfig{
		MaxInstallations:         3,
		MaxSkip:                  1000,
		MaxKeep:                  3000,
		MaxMessageKeysPerSession: 2000,
		BundleRefreshInterval:    24 * 60 * 60 * 1000,
		InstallationID:           installationID,
		Logger:                   logger,
	}
}

// newEncryptor creates a new EncryptionService instance.
func newEncryptor(db *sql.DB, config encryptorConfig) *encryptor {
	return &encryptor{
		persistence: newSQLitePersistence(db),
		config:      config,
		messageIDs:  make(map[string]*confirmationData),
		logger:      config.Logger.With(zap.Namespace("encryptor")),
	}
}

func (s *encryptor) keyFromActiveX3DH(theirIdentityKey []byte, theirSignedPreKey []byte, myIdentityKey *ecdsa.PrivateKey) ([]byte, *ecdsa.PublicKey, error) {
	sharedKey, ephemeralPubKey, err := PerformActiveX3DH(theirIdentityKey, theirSignedPreKey, myIdentityKey)
	if err != nil {
		return nil, nil, err
	}

	return sharedKey, ephemeralPubKey, nil
}

func (s *encryptor) getDRSession(id []byte) (dr.Session, error) {
	sessionStorage := s.persistence.SessionStorage()
	return dr.Load(
		id,
		sessionStorage,
		dr.WithKeysStorage(s.persistence.KeysStorage()),
		dr.WithMaxSkip(s.config.MaxSkip),
		dr.WithMaxKeep(s.config.MaxKeep),
		dr.WithMaxMessageKeysPerSession(s.config.MaxMessageKeysPerSession),
		dr.WithCrypto(crypto.EthereumCrypto{}),
	)
}

func confirmationIDString(id []byte) string {
	return hex.EncodeToString(id)
}

// ConfirmMessagesProcessed confirms and deletes message keys for the given messages
func (s *encryptor) ConfirmMessageProcessed(messageID []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := confirmationIDString(messageID)
	confirmationData, ok := s.messageIDs[id]
	if !ok {
		s.logger.Debug("could not confirm message or message already confirmed", zap.String("messageID", id))
		// We are ok with this, means no key material is stored (public message, or already confirmed)
		return nil
	}

	// Load session from store first
	session, err := s.getDRSession(confirmationData.drInfo.ID)
	if err != nil {
		return err
	}

	if err := session.DeleteMk(confirmationData.header.DH, confirmationData.header.N); err != nil {
		return err
	}

	// Clean up
	delete(s.messageIDs, id)

	return nil
}

// CreateBundle retrieves or creates an X3DH bundle given a private key
func (s *encryptor) CreateBundle(privateKey *ecdsa.PrivateKey, installations []*multidevice.Installation) (*Bundle, error) {
	ourIdentityKeyC := crypto.CompressPubkey(&privateKey.PublicKey)

	bundleContainer, err := s.persistence.GetAnyPrivateBundle(ourIdentityKeyC, installations)
	if err != nil {
		return nil, err
	}

	expired := bundleContainer != nil && bundleContainer.GetBundle().Timestamp < time.Now().Add(-1*time.Duration(s.config.BundleRefreshInterval)*time.Millisecond).UnixNano()

	// If the bundle has expired we create a new one
	if expired {
		// Mark sessions has expired
		if err := s.persistence.MarkBundleExpired(bundleContainer.GetBundle().GetIdentity()); err != nil {
			return nil, err
		}

	} else if bundleContainer != nil {
		err = SignBundle(privateKey, bundleContainer)
		if err != nil {
			return nil, err
		}
		return bundleContainer.GetBundle(), nil
	}

	// needs transaction/mutex to avoid creating multiple bundles
	// although not a problem
	bundleContainer, err = NewBundleContainer(privateKey, s.config.InstallationID)
	if err != nil {
		return nil, err
	}

	if err = s.persistence.AddPrivateBundle(bundleContainer); err != nil {
		return nil, err
	}

	return s.CreateBundle(privateKey, installations)
}

// DecryptWithDH decrypts message sent with a DH key exchange, and throws away the key after decryption
func (s *encryptor) DecryptWithDH(myIdentityKey *ecdsa.PrivateKey, theirEphemeralKey *ecdsa.PublicKey, payload []byte) ([]byte, error) {
	key, err := PerformDH(
		ecies.ImportECDSA(myIdentityKey),
		ecies.ImportECDSAPublic(theirEphemeralKey),
	)
	if err != nil {
		return nil, err
	}

	return crypto.DecryptSymmetric(key, payload)

}

// keyFromPassiveX3DH decrypts message sent with a X3DH key exchange, storing the key for future exchanges
func (s *encryptor) keyFromPassiveX3DH(myIdentityKey *ecdsa.PrivateKey, theirIdentityKey *ecdsa.PublicKey, theirEphemeralKey *ecdsa.PublicKey, ourBundleID []byte) ([]byte, error) {
	bundlePrivateKey, err := s.persistence.GetPrivateKeyBundle(ourBundleID)
	if err != nil {
		s.logger.Error("could not get private bundle", zap.Error(err))
		return nil, err
	}

	if bundlePrivateKey == nil {
		return nil, errSessionNotFound
	}

	signedPreKey, err := crypto.ToECDSA(bundlePrivateKey)
	if err != nil {
		s.logger.Error("could not convert to ecdsa", zap.Error(err))
		return nil, err
	}

	key, err := PerformPassiveX3DH(
		theirIdentityKey,
		signedPreKey,
		theirEphemeralKey,
		myIdentityKey,
	)
	if err != nil {
		s.logger.Error("could not perform passive x3dh", zap.Error(err))
		return nil, err
	}
	return key, nil
}

// ProcessPublicBundle persists a bundle
func (s *encryptor) ProcessPublicBundle(myIdentityKey *ecdsa.PrivateKey, b *Bundle) error {
	return s.persistence.AddPublicBundle(b)
}

func (s *encryptor) GetMessage(msgs map[string]*EncryptedMessageProtocol) *EncryptedMessageProtocol {
	msg := msgs[s.config.InstallationID]
	if msg == nil {
		msg = msgs[noInstallationID]
	}
	return msg
}

// DecryptPayload decrypts the payload of a EncryptedMessageProtocol, given an identity private key and the sender's public key
func (s *encryptor) DecryptPayload(myIdentityKey *ecdsa.PrivateKey, theirIdentityKey *ecdsa.PublicKey, theirInstallationID string, msgs map[string]*EncryptedMessageProtocol, messageID []byte) ([]byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	msg := s.GetMessage(msgs)

	// We should not be sending a signal if it's coming from us, as we receive our own messages
	if msg == nil && !samePublicKeys(*theirIdentityKey, myIdentityKey.PublicKey) {
		s.logger.Debug("message is coming from someone else, but not targeting our installation id")
		return nil, ErrDeviceNotFound
	} else if msg == nil && theirInstallationID != s.config.InstallationID {
		s.logger.Debug("message is coming from same public key, but different installation id")
		return nil, ErrNotPairedDevice
	} else if msg == nil && theirInstallationID == s.config.InstallationID {
		s.logger.Debug("message is coming from us and is nil")
		return nil, nil
	}

	payload := msg.GetPayload()

	if x3dhHeader := msg.GetX3DHHeader(); x3dhHeader != nil {
		bundleID := x3dhHeader.GetId()
		theirEphemeralKey, err := crypto.DecompressPubkey(x3dhHeader.GetKey())

		if err != nil {
			return nil, err
		}

		symmetricKey, err := s.keyFromPassiveX3DH(myIdentityKey, theirIdentityKey, theirEphemeralKey, bundleID)
		if err != nil {
			return nil, err
		}

		theirIdentityKeyC := crypto.CompressPubkey(theirIdentityKey)
		err = s.persistence.AddRatchetInfo(symmetricKey, theirIdentityKeyC, bundleID, nil, theirInstallationID)
		if err != nil {
			return nil, err
		}
	}

	if drHeader := msg.GetDRHeader(); drHeader != nil {
		drMessage := &dr.Message{
			Header: dr.MessageHeader{
				N:  drHeader.GetN(),
				PN: drHeader.GetPn(),
				DH: drHeader.GetKey(),
			},
			Ciphertext: msg.GetPayload(),
		}

		theirIdentityKeyC := crypto.CompressPubkey(theirIdentityKey)

		drInfo, err := s.persistence.GetRatchetInfo(drHeader.GetId(), theirIdentityKeyC, theirInstallationID)
		if err != nil {
			s.logger.Error("could not get ratchet info", zap.Error(err))
			return nil, err
		}

		// We mark the exchange as successful so we stop sending x3dh header
		if err = s.persistence.RatchetInfoConfirmed(drHeader.GetId(), theirIdentityKeyC, theirInstallationID); err != nil {
			s.logger.Error("could not confirm ratchet info", zap.Error(err))
			return nil, err
		}

		if drInfo == nil {
			s.logger.Error("could not find a session")
			return nil, errSessionNotFound
		}

		confirmationData := &confirmationData{
			header: &drMessage.Header,
			drInfo: drInfo,
		}
		s.messageIDs[confirmationIDString(messageID)] = confirmationData

		return s.decryptUsingDR(theirIdentityKey, drInfo, drMessage)
	}

	// Try DH
	if header := msg.GetDHHeader(); header != nil {
		decompressedKey, err := crypto.DecompressPubkey(header.GetKey())
		if err != nil {
			return nil, err
		}
		return s.DecryptWithDH(myIdentityKey, decompressedKey, payload)
	}

	// Try Hash Ratchet
	if header := msg.GetHRHeader(); header != nil {

		ratchet := &HashRatchetKeyCompatibility{
			GroupID: header.GroupId,
			// NOTE: this would be nil in the old format
			keyID: header.KeyId,
		}

		// Old key format
		if header.DeprecatedKeyId != 0 {
			ratchet.Timestamp = uint64(header.DeprecatedKeyId)
		}

		decryptedPayload, err := s.DecryptWithHR(ratchet, header.SeqNo, payload)

		return decryptedPayload, err
	}
	return nil, errors.New("no key specified")
}

func (s *encryptor) createNewSession(drInfo *RatchetInfo, sk []byte, keyPair crypto.DHPair) (dr.Session, error) {
	var err error
	var session dr.Session

	if drInfo.PrivateKey != nil {
		session, err = dr.New(
			drInfo.ID,
			sk,
			keyPair,
			s.persistence.SessionStorage(),
			dr.WithKeysStorage(s.persistence.KeysStorage()),
			dr.WithMaxSkip(s.config.MaxSkip),
			dr.WithMaxKeep(s.config.MaxKeep),
			dr.WithMaxMessageKeysPerSession(s.config.MaxMessageKeysPerSession),
			dr.WithCrypto(crypto.EthereumCrypto{}))
	} else {
		session, err = dr.NewWithRemoteKey(
			drInfo.ID,
			sk,
			keyPair.PubKey,
			s.persistence.SessionStorage(),
			dr.WithKeysStorage(s.persistence.KeysStorage()),
			dr.WithMaxSkip(s.config.MaxSkip),
			dr.WithMaxKeep(s.config.MaxKeep),
			dr.WithMaxMessageKeysPerSession(s.config.MaxMessageKeysPerSession),
			dr.WithCrypto(crypto.EthereumCrypto{}))
	}

	return session, err
}

func (s *encryptor) encryptUsingDR(theirIdentityKey *ecdsa.PublicKey, drInfo *RatchetInfo, payload []byte) ([]byte, *DRHeader, error) {
	var err error

	var session dr.Session

	keyPair := crypto.DHPair{
		PrvKey: drInfo.PrivateKey,
		PubKey: drInfo.PublicKey,
	}

	// Load session from store first
	session, err = s.getDRSession(drInfo.ID)

	if err != nil {
		return nil, nil, err
	}

	// Create a new one
	if session == nil {
		session, err = s.createNewSession(drInfo, drInfo.Sk, keyPair)
		if err != nil {
			return nil, nil, err
		}
	}

	response, err := session.RatchetEncrypt(payload, nil)
	if err != nil {
		return nil, nil, err
	}

	header := &DRHeader{
		Id:  drInfo.BundleID,
		Key: response.Header.DH[:],
		N:   response.Header.N,
		Pn:  response.Header.PN,
	}

	return response.Ciphertext, header, nil
}

func (s *encryptor) decryptUsingDR(theirIdentityKey *ecdsa.PublicKey, drInfo *RatchetInfo, payload *dr.Message) ([]byte, error) {
	var err error

	var session dr.Session

	keyPair := crypto.DHPair{
		PrvKey: drInfo.PrivateKey,
		PubKey: drInfo.PublicKey,
	}

	session, err = s.getDRSession(drInfo.ID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		session, err = s.createNewSession(drInfo, drInfo.Sk, keyPair)
		if err != nil {
			return nil, err
		}
	}

	plaintext, err := session.RatchetDecrypt(*payload, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (s *encryptor) encryptWithDH(theirIdentityKey *ecdsa.PublicKey, payload []byte) (*EncryptedMessageProtocol, error) {
	symmetricKey, ourEphemeralKey, err := PerformActiveDH(theirIdentityKey)
	if err != nil {
		return nil, err
	}

	encryptedPayload, err := crypto.EncryptSymmetric(symmetricKey, payload)
	if err != nil {
		return nil, err
	}

	return &EncryptedMessageProtocol{
		DHHeader: &DHHeader{
			Key: crypto.CompressPubkey(ourEphemeralKey),
		},
		Payload: encryptedPayload,
	}, nil
}

func (s *encryptor) EncryptPayloadWithDH(theirIdentityKey *ecdsa.PublicKey, payload []byte) (map[string]*EncryptedMessageProtocol, error) {
	response := make(map[string]*EncryptedMessageProtocol)
	dmp, err := s.encryptWithDH(theirIdentityKey, payload)
	if err != nil {
		return nil, err
	}

	response[noInstallationID] = dmp
	return response, nil
}

// GetPublicBundle returns the active installations bundles for a given user
func (s *encryptor) GetPublicBundle(theirIdentityKey *ecdsa.PublicKey, installations []*multidevice.Installation) (*Bundle, error) {
	return s.persistence.GetPublicBundle(theirIdentityKey, installations)
}

// EncryptPayload returns a new EncryptedMessageProtocol with a given payload encrypted, given a recipient's public key and the sender private identity key
func (s *encryptor) EncryptPayload(theirIdentityKey *ecdsa.PublicKey, myIdentityKey *ecdsa.PrivateKey, installations []*multidevice.Installation, payload []byte) (map[string]*EncryptedMessageProtocol, []*multidevice.Installation, error) {
	logger := s.logger.With(
		zap.String("site", "EncryptPayload"),
		zap.String("their-identity-key", types.EncodeHex(crypto.FromECDSAPub(theirIdentityKey))))

	// Which installations we are sending the message to
	var targetedInstallations []*multidevice.Installation

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(installations) == 0 {
		// We don't have any, send a message with DH
		logger.Debug("no installations, sending to all devices")
		encryptedPayload, err := s.EncryptPayloadWithDH(theirIdentityKey, payload)
		return encryptedPayload, targetedInstallations, err
	}

	theirIdentityKeyC := crypto.CompressPubkey(theirIdentityKey)
	response := make(map[string]*EncryptedMessageProtocol)

	for _, installation := range installations {
		installationID := installation.ID
		ilogger := logger.With(zap.String("installation-id", installationID))
		ilogger.Debug("processing installation")
		if s.config.InstallationID == installationID {
			continue
		}

		bundle, err := s.persistence.GetPublicBundle(theirIdentityKey, []*multidevice.Installation{installation})
		if err != nil {
			return nil, nil, err
		}

		// See if a session is there already
		drInfo, err := s.persistence.GetAnyRatchetInfo(theirIdentityKeyC, installationID)
		if err != nil {
			return nil, nil, err
		}

		targetedInstallations = append(targetedInstallations, installation)

		if drInfo != nil {
			ilogger.Debug("found DR info for installation")
			encryptedPayload, drHeader, err := s.encryptUsingDR(theirIdentityKey, drInfo, payload)
			if err != nil {
				return nil, nil, err
			}

			dmp := EncryptedMessageProtocol{
				Payload:  encryptedPayload,
				DRHeader: drHeader,
			}

			if drInfo.EphemeralKey != nil {
				dmp.X3DHHeader = &X3DHHeader{
					Key: drInfo.EphemeralKey,
					Id:  drInfo.BundleID,
				}
			}

			response[drInfo.InstallationID] = &dmp
			continue
		}

		theirSignedPreKeyContainer := bundle.GetSignedPreKeys()[installationID]

		// This should not be nil at this point
		if theirSignedPreKeyContainer == nil {
			ilogger.Warn("could not find DR info or bundle for installation")
			continue

		}

		ilogger.Debug("DR info not found, using bundle")

		theirSignedPreKey := theirSignedPreKeyContainer.GetSignedPreKey()

		sharedKey, ourEphemeralKey, err := s.keyFromActiveX3DH(theirIdentityKeyC, theirSignedPreKey, myIdentityKey)
		if err != nil {
			return nil, nil, err
		}
		theirIdentityKeyC := crypto.CompressPubkey(theirIdentityKey)
		ourEphemeralKeyC := crypto.CompressPubkey(ourEphemeralKey)

		err = s.persistence.AddRatchetInfo(sharedKey, theirIdentityKeyC, theirSignedPreKey, ourEphemeralKeyC, installationID)
		if err != nil {
			return nil, nil, err
		}

		x3dhHeader := &X3DHHeader{
			Key: ourEphemeralKeyC,
			Id:  theirSignedPreKey,
		}

		drInfo, err = s.persistence.GetRatchetInfo(theirSignedPreKey, theirIdentityKeyC, installationID)
		if err != nil {
			return nil, nil, err
		}

		if drInfo != nil {
			encryptedPayload, drHeader, err := s.encryptUsingDR(theirIdentityKey, drInfo, payload)
			if err != nil {
				return nil, nil, err
			}

			dmp := &EncryptedMessageProtocol{
				Payload:    encryptedPayload,
				X3DHHeader: x3dhHeader,
				DRHeader:   drHeader,
			}

			response[drInfo.InstallationID] = dmp
		}
	}

	var installationIDs []string
	for _, i := range targetedInstallations {
		installationIDs = append(installationIDs, i.ID)
	}
	logger.Info(
		"built a message",
		zap.Strings("installation-ids", installationIDs),
	)

	return response, targetedInstallations, nil
}

func (s *encryptor) getNextHashRatchet(groupID []byte) (*HashRatchetKeyCompatibility, error) {
	latestKey, err := s.persistence.GetCurrentKeyForGroup(groupID)
	if err != nil {
		return nil, err
	}
	return latestKey.GenerateNext()
}

// GenerateHashRatchetKey Generates and stores a hash ratchet key given a group ID
func (s *encryptor) GenerateHashRatchetKey(groupID []byte) (*HashRatchetKeyCompatibility, error) {

	key, err := s.getNextHashRatchet(groupID)
	if err != nil {
		return nil, err
	}

	return key, s.persistence.SaveHashRatchetKey(key)
}

// EncryptHashRatchetPayload returns a new EncryptedMessageProtocol with a given payload encrypted, given a group's key
func (s *encryptor) EncryptHashRatchetPayload(ratchet *HashRatchetKeyCompatibility, payload []byte) (map[string]*EncryptedMessageProtocol, error) {
	logger := s.logger.With(
		zap.String("site", "EncryptHashRatchetPayload"),
		zap.Any("group-id", ratchet.GroupID),
		zap.Any("key-id", ratchet.keyID))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.Debug("encrypting hash ratchet message")
	encryptedPayload, newSeqNo, err := s.EncryptWithHR(ratchet, payload)
	if err != nil {
		return nil, err
	}

	keyID, err := ratchet.GetKeyID()
	if err != nil {
		return nil, err
	}

	dmp := &EncryptedMessageProtocol{
		HRHeader: &HRHeader{
			DeprecatedKeyId: ratchet.DeprecatedKeyID(),
			GroupId:         ratchet.GroupID,
			KeyId:           keyID,
			SeqNo:           newSeqNo,
		},
		Payload: encryptedPayload,
	}

	response := make(map[string]*EncryptedMessageProtocol)
	response[noInstallationID] = dmp
	return response, err
}

func samePublicKeys(pubKey1, pubKey2 ecdsa.PublicKey) bool {
	return pubKey1.X.Cmp(pubKey2.X) == 0 && pubKey1.Y.Cmp(pubKey2.Y) == 0
}

func (s *encryptor) EncryptWithHR(ratchet *HashRatchetKeyCompatibility, payload []byte) ([]byte, uint32, error) {
	hrCache, err := s.persistence.GetHashRatchetCache(ratchet, 0) // Get latest seqNo

	if err != nil {
		return nil, 0, err
	}
	if hrCache == nil {
		return nil, 0, ErrNoEncryptionKey
	}

	var dbHash []byte
	if len(hrCache.Hash) == 0 {
		dbHash = hrCache.Key
	} else {
		dbHash = hrCache.Hash
	}

	hash := crypto.Keccak256Hash(dbHash)
	encryptedPayload, err := crypto.EncryptSymmetric(hash.Bytes(), payload)
	if err != nil {
		return nil, 0, err
	}
	newSeqNo := hrCache.SeqNo + 1
	err = s.persistence.SaveHashRatchetKeyHash(ratchet, hash.Bytes(), newSeqNo)
	if err != nil {
		return nil, 0, err
	}

	return encryptedPayload, newSeqNo, nil
}

func (s *encryptor) DecryptWithHR(ratchet *HashRatchetKeyCompatibility, seqNo uint32, payload []byte) ([]byte, error) {
	// Key exchange message, nothing to decrypt
	if seqNo == 0 {
		return payload, nil
	}

	hrCache, err := s.persistence.GetHashRatchetCache(ratchet, seqNo)
	if err != nil {
		return nil, err
	}

	if hrCache == nil {
		return nil, ErrHashRatchetGroupIDNotFound
	}

	// Handle mesages with seqNo less than the one in db
	// 1. Check cache. If present for a particular seqNo, all good
	// 2. Otherwise, get the latest one for that keyId
	// 3. Every time the key is generated, it has to be saved in the cache along with the hash
	var hash []byte = hrCache.Hash
	if hrCache.SeqNo == seqNo {
		// We already have the hash for this seqNo
		hash = hrCache.Hash
	} else {
		if hrCache.SeqNo == 0 {
			// No cache records found for this keyId
			hash = hrCache.Key
		}
		// We should not have "holes" in seq numbers,
		// so a case when hrCache.SeqNo > seqNo shouldn't occur
		if seqNo-hrCache.SeqNo > maxHashRatchetSeqNoDelta {
			return nil, ErrHashRatchetSeqNoTooHigh
		}
		for i := hrCache.SeqNo; i < seqNo; i++ {
			hash = crypto.Keccak256Hash(hash).Bytes()
			err := s.persistence.SaveHashRatchetKeyHash(ratchet, hash, i+1)
			if err != nil {
				return nil, err
			}
		}
	}

	decryptedPayload, err := crypto.DecryptSymmetric(hash, payload)

	if err != nil {
		s.logger.Error("failed to decrypt hash", zap.Error(err))
		return nil, err
	}
	return decryptedPayload, nil
}
