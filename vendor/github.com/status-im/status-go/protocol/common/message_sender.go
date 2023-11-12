package common

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"database/sql"
	"math"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	datasyncnode "github.com/status-im/mvds/node"
	datasyncproto "github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/datasync"
	datasyncpeer "github.com/status-im/status-go/protocol/datasync/peer"
	"github.com/status-im/status-go/protocol/encryption"
	"github.com/status-im/status-go/protocol/encryption/sharedsecret"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/transport"
	v1protocol "github.com/status-im/status-go/protocol/v1"
)

// Whisper message properties.
const (
	whisperTTL        = 15
	whisperDefaultPoW = 0.002
	// whisperLargeSizePoW is the PoWTarget for larger payload sizes
	whisperLargeSizePoW = 0.000002
	// largeSizeInBytes is when should we be using a lower POW.
	// Roughly this is 50KB
	largeSizeInBytes = 50000
	whisperPoWTime   = 5
)

// RekeyCompatibility indicates whether we should be sending
// keys in 1-to-1 messages as well as in the newer format
var RekeyCompatibility = true

// SentMessage reprent a message that has been passed to the transport layer
type SentMessage struct {
	PublicKey  *ecdsa.PublicKey
	Spec       *encryption.ProtocolMessageSpec
	MessageIDs [][]byte
}

type MessageEventType uint32

const (
	MessageScheduled = iota + 1
	MessageSent
)

type MessageEvent struct {
	Recipient   *ecdsa.PublicKey
	Type        MessageEventType
	SentMessage *SentMessage
	RawMessage  *RawMessage
}

type MessageSender struct {
	identity    *ecdsa.PrivateKey
	datasync    *datasync.DataSync
	database    *sql.DB
	protocol    *encryption.Protocol
	transport   *transport.Transport
	logger      *zap.Logger
	persistence *RawMessagesPersistence

	datasyncEnabled bool

	// ephemeralKeys is a map that contains the ephemeral keys of the client, used
	// to decrypt messages
	ephemeralKeys      map[string]*ecdsa.PrivateKey
	ephemeralKeysMutex sync.Mutex

	// messageEventsSubscriptions contains all the subscriptions for message events
	messageEventsSubscriptions []chan<- *MessageEvent

	featureFlags FeatureFlags

	// handleSharedSecrets is a callback that is called every time a new shared secret is negotiated
	handleSharedSecrets func([]*sharedsecret.Secret) error
}

func NewMessageSender(
	identity *ecdsa.PrivateKey,
	database *sql.DB,
	enc *encryption.Protocol,
	transport *transport.Transport,
	logger *zap.Logger,
	features FeatureFlags,
) (*MessageSender, error) {
	p := &MessageSender{
		identity:        identity,
		datasyncEnabled: features.Datasync,
		protocol:        enc,
		database:        database,
		persistence:     NewRawMessagesPersistence(database),
		transport:       transport,
		logger:          logger,
		ephemeralKeys:   make(map[string]*ecdsa.PrivateKey),
		featureFlags:    features,
	}

	return p, nil
}

func (s *MessageSender) Stop() {
	for _, c := range s.messageEventsSubscriptions {
		close(c)
	}
	s.messageEventsSubscriptions = nil
	s.StopDatasync()
}

func (s *MessageSender) SetHandleSharedSecrets(handler func([]*sharedsecret.Secret) error) {
	s.handleSharedSecrets = handler
}

func (s *MessageSender) StartDatasync(handler func(peer state.PeerID, payload *datasyncproto.Payload) error) error {
	if !s.datasyncEnabled {
		return nil
	}

	dataSyncTransport := datasync.NewNodeTransport()
	dataSyncNode, err := datasyncnode.NewPersistentNode(
		s.database,
		dataSyncTransport,
		datasyncpeer.PublicKeyToPeerID(s.identity.PublicKey),
		datasyncnode.BATCH,
		datasync.CalculateSendTime,
		s.logger,
	)
	if err != nil {
		return err
	}

	s.datasync = datasync.New(dataSyncNode, dataSyncTransport, true, s.logger)

	s.datasync.Init(handler, s.logger)
	s.datasync.Start(datasync.DatasyncTicker)

	return nil
}

// SendPrivate takes encoded data, encrypts it and sends through the wire.
func (s *MessageSender) SendPrivate(
	ctx context.Context,
	recipient *ecdsa.PublicKey,
	rawMessage *RawMessage,
) ([]byte, error) {
	s.logger.Debug(
		"sending a private message",
		zap.String("public-key", types.EncodeHex(crypto.FromECDSAPub(recipient))),
		zap.String("site", "SendPrivate"),
	)
	// Currently we don't support sending through datasync and setting custom waku fields,
	// as the datasync interface is not rich enough to propagate that information, so we
	// would have to add some complexity to handle this.
	if rawMessage.ResendAutomatically && (rawMessage.Sender != nil || rawMessage.SkipEncryptionLayer || rawMessage.SendOnPersonalTopic) {
		return nil, errors.New("setting identity, skip-encryption or personal topic and datasync not supported")
	}

	// Set sender identity if not specified
	if rawMessage.Sender == nil {
		rawMessage.Sender = s.identity
	}

	return s.sendPrivate(ctx, recipient, rawMessage)
}

// SendCommunityMessage takes encoded data, encrypts it and sends through the wire
// using the community topic and their key
func (s *MessageSender) SendCommunityMessage(
	ctx context.Context,
	rawMessage RawMessage,
) ([]byte, error) {
	s.logger.Debug(
		"sending a community message",
		zap.String("communityId", types.EncodeHex(rawMessage.CommunityID)),
		zap.String("site", "SendCommunityMessage"),
	)
	rawMessage.Sender = s.identity

	return s.sendCommunity(ctx, &rawMessage)
}

// SendPubsubTopicKey sends the protected topic key for a community to a list of recipients
func (s *MessageSender) SendPubsubTopicKey(
	ctx context.Context,
	rawMessage *RawMessage,
) ([]byte, error) {
	s.logger.Debug(
		"sending the protected topic key for a community",
		zap.String("communityId", types.EncodeHex(rawMessage.CommunityID)),
		zap.String("site", "SendPubsubTopicKey"),
	)
	rawMessage.Sender = s.identity
	messageID, err := s.getMessageID(rawMessage)
	if err != nil {
		return nil, err
	}

	rawMessage.ID = types.EncodeHex(messageID)

	// Notify before dispatching, otherwise the dispatch subscription might happen
	// earlier than the scheduled
	s.notifyOnScheduledMessage(nil, rawMessage)

	// Send to each recipients
	for _, recipient := range rawMessage.Recipients {
		_, err = s.sendPrivate(ctx, recipient, rawMessage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to send message")
		}
	}
	return messageID, nil

}

// SendGroup takes encoded data, encrypts it and sends through the wire,
// always return the messageID
func (s *MessageSender) SendGroup(
	ctx context.Context,
	recipients []*ecdsa.PublicKey,
	rawMessage RawMessage,
) ([]byte, error) {
	s.logger.Debug(
		"sending a private group message",
		zap.String("site", "SendGroup"),
	)
	// Set sender if not specified
	if rawMessage.Sender == nil {
		rawMessage.Sender = s.identity
	}

	// Calculate messageID first and set on raw message
	wrappedMessage, err := s.wrapMessageV1(&rawMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap message")
	}
	messageID := v1protocol.MessageID(&rawMessage.Sender.PublicKey, wrappedMessage)
	rawMessage.ID = types.EncodeHex(messageID)

	// We call it only once, and we nil the function after so it doesn't get called again
	if rawMessage.BeforeDispatch != nil {
		if err := rawMessage.BeforeDispatch(&rawMessage); err != nil {
			return nil, err
		}
	}

	// Send to each recipients
	for _, recipient := range recipients {
		_, err = s.sendPrivate(ctx, recipient, &rawMessage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to send message")
		}
	}
	return messageID, nil
}

func (s *MessageSender) getMessageID(rawMessage *RawMessage) (types.HexBytes, error) {
	wrappedMessage, err := s.wrapMessageV1(rawMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap message")
	}

	messageID := v1protocol.MessageID(&rawMessage.Sender.PublicKey, wrappedMessage)

	return messageID, nil
}

func ShouldCommunityMessageBeEncrypted(msgType protobuf.ApplicationMetadataMessage_Type) bool {
	return msgType == protobuf.ApplicationMetadataMessage_CHAT_MESSAGE ||
		msgType == protobuf.ApplicationMetadataMessage_EDIT_MESSAGE ||
		msgType == protobuf.ApplicationMetadataMessage_DELETE_MESSAGE ||
		msgType == protobuf.ApplicationMetadataMessage_PIN_MESSAGE ||
		msgType == protobuf.ApplicationMetadataMessage_EMOJI_REACTION
}

// sendCommunity sends a message that's to be sent in a community
// If it's a chat message, it will go to the respective topic derived by the
// chat id, if it's not a chat message, it will go to the community topic.
func (s *MessageSender) sendCommunity(
	ctx context.Context,
	rawMessage *RawMessage,
) ([]byte, error) {
	s.logger.Debug("sending community message", zap.String("recipient", types.EncodeHex(crypto.FromECDSAPub(&rawMessage.Sender.PublicKey))))

	// Set sender
	if rawMessage.Sender == nil {
		rawMessage.Sender = s.identity
	}

	messageID, err := s.getMessageID(rawMessage)
	if err != nil {
		return nil, err
	}
	rawMessage.ID = types.EncodeHex(messageID)

	if rawMessage.BeforeDispatch != nil {
		if err := rawMessage.BeforeDispatch(rawMessage); err != nil {
			return nil, err
		}
	}
	// Notify before dispatching, otherwise the dispatch subscription might happen
	// earlier than the scheduled
	s.notifyOnScheduledMessage(nil, rawMessage)

	var hashes [][]byte
	var newMessages []*types.NewMessage

	forceRekey := rawMessage.CommunityKeyExMsgType == KeyExMsgRekey

	// Check if it's a key exchange message. In this case we send it
	// to all the recipients
	if rawMessage.CommunityKeyExMsgType != KeyExMsgNone {
		// If rekeycompatibility is on, we always
		// want to execute below, otherwise we execute
		// only when we want to fill up old keys to a given user
		if RekeyCompatibility || !forceRekey {
			keyExMessageSpecs, err := s.protocol.GetKeyExMessageSpecs(rawMessage.HashRatchetGroupID, s.identity, rawMessage.Recipients, forceRekey)
			if err != nil {
				return nil, err
			}

			for i, spec := range keyExMessageSpecs {
				recipient := rawMessage.Recipients[i]
				_, _, err = s.sendMessageSpec(ctx, recipient, spec, [][]byte{messageID})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	wrappedMessage, err := s.wrapMessageV1(rawMessage)
	if err != nil {
		return nil, err
	}

	// If it's a chat message, we send it on the community chat topic
	if ShouldCommunityMessageBeEncrypted(rawMessage.MessageType) {
		messageSpec, err := s.protocol.BuildHashRatchetMessage(rawMessage.HashRatchetGroupID, wrappedMessage)
		if err != nil {
			return nil, err
		}

		payload, err := proto.Marshal(messageSpec.Message)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal")
		}
		hashes, newMessages, err = s.dispatchCommunityChatMessage(ctx, rawMessage, payload, forceRekey)
		if err != nil {
			return nil, err
		}

		sentMessage := &SentMessage{
			Spec:       messageSpec,
			MessageIDs: [][]byte{messageID},
		}

		s.notifyOnSentMessage(sentMessage)

	} else {

		pubkey, err := crypto.DecompressPubkey(rawMessage.CommunityID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decompress pubkey")
		}
		hashes, newMessages, err = s.dispatchCommunityMessage(ctx, pubkey, wrappedMessage, rawMessage.PubsubTopic, forceRekey, rawMessage)
		if err != nil {
			s.logger.Error("failed to send a community message", zap.Error(err))
			return nil, errors.Wrap(err, "failed to send a message spec")
		}
	}

	s.logger.Debug("sent community message ", zap.String("messageID", messageID.String()), zap.Strings("hashes", types.EncodeHexes(hashes)))
	s.transport.Track(messageID, hashes, newMessages)

	return messageID, nil
}

// sendPrivate sends data to the recipient identifying with a given public key.
func (s *MessageSender) sendPrivate(
	ctx context.Context,
	recipient *ecdsa.PublicKey,
	rawMessage *RawMessage,
) ([]byte, error) {
	s.logger.Debug("sending private message", zap.String("recipient", types.EncodeHex(crypto.FromECDSAPub(recipient))))

	var wrappedMessage []byte
	var err error
	if rawMessage.SkipApplicationWrap {
		wrappedMessage = rawMessage.Payload
	} else {
		wrappedMessage, err = s.wrapMessageV1(rawMessage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to wrap message")
		}
	}

	messageID := v1protocol.MessageID(&rawMessage.Sender.PublicKey, wrappedMessage)
	rawMessage.ID = types.EncodeHex(messageID)
	if rawMessage.BeforeDispatch != nil {
		if err := rawMessage.BeforeDispatch(rawMessage); err != nil {
			return nil, err
		}
	}

	// Notify before dispatching, otherwise the dispatch subscription might happen
	// earlier than the scheduled
	s.notifyOnScheduledMessage(recipient, rawMessage)

	if s.datasync != nil && s.featureFlags.Datasync && rawMessage.ResendAutomatically {
		// No need to call transport tracking.
		// It is done in a data sync dispatch step.
		datasyncID, err := s.addToDataSync(recipient, wrappedMessage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to send message with datasync")
		}
		// We don't need to receive confirmations from our own devices
		if !IsPubKeyEqual(recipient, &s.identity.PublicKey) {
			confirmation := &RawMessageConfirmation{
				DataSyncID: datasyncID,
				MessageID:  messageID,
				PublicKey:  crypto.CompressPubkey(recipient),
			}

			err = s.persistence.InsertPendingConfirmation(confirmation)
			if err != nil {
				return nil, err
			}
		}
	} else if rawMessage.SkipEncryptionLayer {

		messageBytes := wrappedMessage
		if rawMessage.CommunityKeyExMsgType == KeyExMsgReuse {
			groupID := rawMessage.HashRatchetGroupID

			ratchets, err := s.protocol.GetKeysForGroup(groupID)
			if err != nil {
				return nil, err
			}

			message, err := s.protocol.BuildHashRatchetKeyExchangeMessageWithPayload(s.identity, recipient, groupID, ratchets, wrappedMessage)
			if err != nil {
				return nil, err
			}

			messageBytes, err = proto.Marshal(message.Message)
			if err != nil {
				return nil, err
			}
		}

		// When SkipProtocolLayer is set we don't pass the message to the encryption layer
		hashes, newMessages, err := s.sendPrivateRawMessage(ctx, rawMessage, recipient, messageBytes)
		if err != nil {
			s.logger.Error("failed to send a private message", zap.Error(err))
			return nil, errors.Wrap(err, "failed to send a message spec")
		}

		s.logger.Debug("sent private message skipProtocolLayer", zap.String("messageID", messageID.String()), zap.Strings("hashes", types.EncodeHexes(hashes)))
		s.transport.Track(messageID, hashes, newMessages)

	} else {
		messageSpec, err := s.protocol.BuildEncryptedMessage(rawMessage.Sender, recipient, wrappedMessage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encrypt message")
		}

		hashes, newMessages, err := s.sendMessageSpec(ctx, recipient, messageSpec, [][]byte{messageID})
		if err != nil {
			s.logger.Error("failed to send a private message", zap.Error(err))
			return nil, errors.Wrap(err, "failed to send a message spec")
		}

		s.logger.Debug("sent private message without datasync", zap.String("messageID", messageID.String()), zap.Strings("hashes", types.EncodeHexes(hashes)))
		s.transport.Track(messageID, hashes, newMessages)
	}

	return messageID, nil
}

// sendPairInstallation sends data to the recipients, using DH
func (s *MessageSender) SendPairInstallation(
	ctx context.Context,
	recipient *ecdsa.PublicKey,
	rawMessage RawMessage,
) ([]byte, error) {
	s.logger.Debug("sending private message", zap.String("recipient", types.EncodeHex(crypto.FromECDSAPub(recipient))))

	wrappedMessage, err := s.wrapMessageV1(&rawMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap message")
	}

	messageSpec, err := s.protocol.BuildDHMessage(s.identity, recipient, wrappedMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt message")
	}

	messageID := v1protocol.MessageID(&s.identity.PublicKey, wrappedMessage)

	hashes, newMessages, err := s.sendMessageSpec(ctx, recipient, messageSpec, [][]byte{messageID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to send a message spec")
	}

	s.transport.Track(messageID, hashes, newMessages)

	return messageID, nil
}

func (s *MessageSender) encodeMembershipUpdate(
	message v1protocol.MembershipUpdateMessage,
	chatEntity ChatEntity,
) ([]byte, error) {

	if chatEntity != nil {
		chatEntityProtobuf := chatEntity.GetProtobuf()
		switch chatEntityProtobuf := chatEntityProtobuf.(type) {
		case *protobuf.ChatMessage:
			message.Message = chatEntityProtobuf
		case *protobuf.EmojiReaction:
			message.EmojiReaction = chatEntityProtobuf

		}
	}

	encodedMessage, err := v1protocol.EncodeMembershipUpdateMessage(message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode membership update message")
	}

	return encodedMessage, nil
}

// EncodeMembershipUpdate takes a group and an optional chat message and returns the protobuf representation to be sent on the wire.
// All the events in a group are encoded and added to the payload
func (s *MessageSender) EncodeMembershipUpdate(
	group *v1protocol.Group,
	chatEntity ChatEntity,
) ([]byte, error) {
	message := v1protocol.MembershipUpdateMessage{
		ChatID: group.ChatID(),
		Events: group.Events(),
	}

	return s.encodeMembershipUpdate(message, chatEntity)
}

// EncodeAbridgedMembershipUpdate takes a group and an optional chat message and returns the protobuf representation to be sent on the wire.
// Only the events relevant to the current group are encoded
func (s *MessageSender) EncodeAbridgedMembershipUpdate(
	group *v1protocol.Group,
	chatEntity ChatEntity,
) ([]byte, error) {
	message := v1protocol.MembershipUpdateMessage{
		ChatID: group.ChatID(),
		Events: group.AbridgedEvents(),
	}
	return s.encodeMembershipUpdate(message, chatEntity)
}

func (s *MessageSender) dispatchCommunityChatMessage(ctx context.Context, rawMessage *RawMessage, wrappedMessage []byte, rekey bool) ([][]byte, []*types.NewMessage, error) {
	payload := wrappedMessage
	var err error
	if rekey && len(rawMessage.HashRatchetGroupID) != 0 {

		var ratchet *encryption.HashRatchetKeyCompatibility
		// We have just rekeyed, pull the latest
		if RekeyCompatibility {
			ratchet, err = s.protocol.GetCurrentKeyForGroup(rawMessage.HashRatchetGroupID)
			if err != nil {
				return nil, nil, err
			}

		}
		// We send the message over the community topic
		spec, err := s.protocol.BuildHashRatchetReKeyGroupMessage(s.identity, rawMessage.Recipients, rawMessage.HashRatchetGroupID, wrappedMessage, ratchet)
		if err != nil {
			return nil, nil, err
		}
		payload, err = proto.Marshal(spec.Message)
		if err != nil {
			return nil, nil, err
		}
	}

	newMessage := &types.NewMessage{
		TTL:         whisperTTL,
		Payload:     payload,
		PowTarget:   calculatePoW(payload),
		PowTime:     whisperPoWTime,
		PubsubTopic: rawMessage.PubsubTopic,
	}

	if rawMessage.BeforeDispatch != nil {
		if err := rawMessage.BeforeDispatch(rawMessage); err != nil {
			return nil, nil, err
		}
	}

	// notify before dispatching
	s.notifyOnScheduledMessage(nil, rawMessage)

	newMessages, err := s.segmentMessage(newMessage)
	if err != nil {
		return nil, nil, err
	}

	hashes := make([][]byte, 0, len(newMessages))
	for _, newMessage := range newMessages {
		hash, err := s.transport.SendPublic(ctx, newMessage, rawMessage.LocalChatID)
		if err != nil {
			return nil, nil, err
		}
		hashes = append(hashes, hash)
	}

	return hashes, newMessages, nil
}

// SendPublic takes encoded data, encrypts it and sends through the wire.
func (s *MessageSender) SendPublic(
	ctx context.Context,
	chatName string,
	rawMessage RawMessage,
) ([]byte, error) {
	// Set sender
	if rawMessage.Sender == nil {
		rawMessage.Sender = s.identity
	}

	var wrappedMessage []byte
	var err error
	if rawMessage.SkipApplicationWrap {
		wrappedMessage = rawMessage.Payload
	} else {
		wrappedMessage, err = s.wrapMessageV1(&rawMessage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to wrap message")
		}
	}

	var newMessage *types.NewMessage

	messageSpec, err := s.protocol.BuildPublicMessage(s.identity, wrappedMessage)
	if err != nil {
		s.logger.Error("failed to send a public message", zap.Error(err))
		return nil, errors.Wrap(err, "failed to wrap a public message in the encryption layer")
	}

	if len(rawMessage.HashRatchetGroupID) != 0 {

		var ratchet *encryption.HashRatchetKeyCompatibility
		var err error
		// We have just rekeyed, pull the latest
		ratchet, err = s.protocol.GetCurrentKeyForGroup(rawMessage.HashRatchetGroupID)
		if err != nil {
			return nil, err
		}

		keyID, err := ratchet.GetKeyID()
		if err != nil {
			return nil, err
		}
		s.logger.Debug("adding key id to message", zap.String("keyid", types.Bytes2Hex(keyID)))
		// We send the message over the community topic
		spec, err := s.protocol.BuildHashRatchetReKeyGroupMessage(s.identity, rawMessage.Recipients, rawMessage.HashRatchetGroupID, wrappedMessage, ratchet)
		if err != nil {
			return nil, err
		}
		newMessage, err = MessageSpecToWhisper(spec)
		if err != nil {
			return nil, err
		}

	} else if !rawMessage.SkipEncryptionLayer {
		newMessage, err = MessageSpecToWhisper(messageSpec)
		if err != nil {
			return nil, err
		}
	} else {
		newMessage = &types.NewMessage{
			TTL:       whisperTTL,
			Payload:   wrappedMessage,
			PowTarget: calculatePoW(wrappedMessage),
			PowTime:   whisperPoWTime,
		}
	}

	newMessage.Ephemeral = rawMessage.Ephemeral
	newMessage.PubsubTopic = rawMessage.PubsubTopic

	messageID := v1protocol.MessageID(&rawMessage.Sender.PublicKey, wrappedMessage)
	rawMessage.ID = types.EncodeHex(messageID)

	if rawMessage.BeforeDispatch != nil {
		if err := rawMessage.BeforeDispatch(&rawMessage); err != nil {
			return nil, err
		}
	}

	// notify before dispatching
	s.notifyOnScheduledMessage(nil, &rawMessage)

	newMessages, err := s.segmentMessage(newMessage)
	if err != nil {
		return nil, err
	}

	hashes := make([][]byte, 0, len(newMessages))
	for _, newMessage := range newMessages {
		hash, err := s.transport.SendPublic(ctx, newMessage, chatName)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}

	sentMessage := &SentMessage{
		Spec:       messageSpec,
		MessageIDs: [][]byte{messageID},
	}

	s.notifyOnSentMessage(sentMessage)

	s.logger.Debug("sent public message", zap.String("messageID", messageID.String()), zap.Strings("hashes", types.EncodeHexes(hashes)))
	s.transport.Track(messageID, hashes, newMessages)

	return messageID, nil
}

// unwrapDatasyncMessage tries to unwrap message as datasync one and in case of success
// returns cloned messages with replaced payloads
func (s *MessageSender) unwrapDatasyncMessage(m *v1protocol.StatusMessage, response *handleMessageResponse) error {

	datasyncMessage, err := s.datasync.Unwrap(
		m.SigPubKey(),
		m.EncryptionLayer.Payload,
	)
	if err != nil {
		return err
	}

	response.DatasyncSender = m.SigPubKey()
	response.DatasyncAcks = append(response.DatasyncAcks, datasyncMessage.Acks...)
	response.DatasyncRequests = append(response.DatasyncRequests, datasyncMessage.Requests...)
	for _, o := range datasyncMessage.GroupOffers {
		for _, mID := range o.MessageIds {
			response.DatasyncOffers = append(response.DatasyncOffers, DatasyncOffer{GroupID: o.GroupId, MessageID: mID})
		}
	}

	for _, ds := range datasyncMessage.Messages {
		message, err := m.Clone()
		if err != nil {
			return err
		}
		message.EncryptionLayer.Payload = ds.Body
		response.DatasyncMessages = append(response.DatasyncMessages, message)

	}
	return nil
}

// HandleMessages expects a whisper message as input, and it will go through
// a series of transformations until the message is parsed into an application
// layer message, or in case of Raw methods, the processing stops at the layer
// before.
// It returns an error only if the processing of required steps failed.
func (s *MessageSender) HandleMessages(wakuMessage *types.Message) (*HandleMessageResponse, error) {
	logger := s.logger.With(zap.String("site", "HandleMessages"))
	hlogger := logger.With(zap.String("hash", types.HexBytes(wakuMessage.Hash).String()))

	response, err := s.handleMessage(wakuMessage)
	if err != nil {
		// Hash ratchet with a group id not found yet, save the message for future processing
		if err == encryption.ErrHashRatchetGroupIDNotFound && len(response.Message.EncryptionLayer.HashRatchetInfo) == 1 {
			info := response.Message.EncryptionLayer.HashRatchetInfo[0]
			return nil, s.persistence.SaveHashRatchetMessage(info.GroupID, info.KeyID, wakuMessage)
		}

		// The current message segment has been successfully retrieved.
		// However, the collection of segments is not yet complete.
		if err == ErrMessageSegmentsIncomplete {
			return nil, nil
		}

		return nil, err
	}

	// Process queued hash ratchet messages
	for _, hashRatchetInfo := range response.Message.EncryptionLayer.HashRatchetInfo {
		messages, err := s.persistence.GetHashRatchetMessages(hashRatchetInfo.KeyID)
		if err != nil {
			return nil, err
		}

		var processedIds [][]byte
		for _, message := range messages {
			hlogger.Info("handling out of order encrypted messages", zap.String("hash", types.Bytes2Hex(message.Hash)))
			r, err := s.handleMessage(message)
			if err != nil {
				hlogger.Debug("failed to handle hash ratchet message", zap.Error(err))
				continue
			}
			response.DatasyncMessages = append(response.toPublicResponse().StatusMessages, r.Messages()...)
			response.DatasyncAcks = append(response.DatasyncAcks, r.DatasyncAcks...)

			processedIds = append(processedIds, message.Hash)
		}

		err = s.persistence.DeleteHashRatchetMessages(processedIds)
		if err != nil {
			s.logger.Warn("failed to delete hash ratchet messages", zap.Error(err))
			return nil, err
		}
	}

	return response.toPublicResponse(), nil
}

type DatasyncOffer struct {
	GroupID   []byte
	MessageID []byte
}

type HandleMessageResponse struct {
	Hash             []byte
	StatusMessages   []*v1protocol.StatusMessage
	DatasyncSender   *ecdsa.PublicKey
	DatasyncAcks     [][]byte
	DatasyncOffers   []DatasyncOffer
	DatasyncRequests [][]byte
}

func (h *handleMessageResponse) toPublicResponse() *HandleMessageResponse {
	return &HandleMessageResponse{
		Hash:             h.Hash,
		StatusMessages:   h.Messages(),
		DatasyncSender:   h.DatasyncSender,
		DatasyncAcks:     h.DatasyncAcks,
		DatasyncOffers:   h.DatasyncOffers,
		DatasyncRequests: h.DatasyncRequests,
	}
}

type handleMessageResponse struct {
	Hash             []byte
	Message          *v1protocol.StatusMessage
	DatasyncMessages []*v1protocol.StatusMessage
	DatasyncSender   *ecdsa.PublicKey
	DatasyncAcks     [][]byte
	DatasyncOffers   []DatasyncOffer
	DatasyncRequests [][]byte
}

func (h *handleMessageResponse) Messages() []*v1protocol.StatusMessage {
	if len(h.DatasyncMessages) > 0 {
		return h.DatasyncMessages
	}
	return []*v1protocol.StatusMessage{h.Message}
}

func (s *MessageSender) handleMessage(wakuMessage *types.Message) (*handleMessageResponse, error) {
	logger := s.logger.With(zap.String("site", "handleMessage"))
	hlogger := logger.With(zap.String("hash", types.EncodeHex(wakuMessage.Hash)))

	message := &v1protocol.StatusMessage{}

	response := &handleMessageResponse{
		Hash:             wakuMessage.Hash,
		Message:          message,
		DatasyncMessages: []*v1protocol.StatusMessage{},
		DatasyncAcks:     [][]byte{},
	}

	err := message.HandleTransportLayer(wakuMessage)
	if err != nil {
		hlogger.Error("failed to handle transport layer message", zap.Error(err))
		return nil, err
	}

	err = s.handleSegmentationLayer(message)
	if err != nil {
		hlogger.Debug("failed to handle segmentation layer message", zap.Error(err))

		// Segments not completed yet, stop processing
		if err == ErrMessageSegmentsIncomplete {
			return nil, err
		}
		// Segments already completed, stop processing
		if err == ErrMessageSegmentsAlreadyCompleted {
			return nil, err
		}
	}

	err = s.handleEncryptionLayer(context.Background(), message)
	if err != nil {
		hlogger.Debug("failed to handle an encryption message", zap.Error(err))

		// Hash ratchet with a group id not found yet, stop processing
		if err == encryption.ErrHashRatchetGroupIDNotFound {
			return response, err
		}
	}

	if s.datasync != nil && s.datasyncEnabled {
		err := s.unwrapDatasyncMessage(message, response)
		if err != nil {
			hlogger.Debug("failed to handle datasync message", zap.Error(err))
		}
	}

	for _, msg := range response.Messages() {
		err := msg.HandleApplicationLayer()
		if err != nil {
			hlogger.Error("failed to handle application metadata layer message", zap.Error(err))
		}
	}

	return response, nil
}

// fetchDecryptionKey returns the private key associated with this public key, and returns true if it's an ephemeral key
func (s *MessageSender) fetchDecryptionKey(destination *ecdsa.PublicKey) (*ecdsa.PrivateKey, bool) {
	destinationID := types.EncodeHex(crypto.FromECDSAPub(destination))

	s.ephemeralKeysMutex.Lock()
	decryptionKey, ok := s.ephemeralKeys[destinationID]
	s.ephemeralKeysMutex.Unlock()

	// the key is not there, fallback on identity
	if !ok {
		return s.identity, false
	}
	return decryptionKey, true
}

func (s *MessageSender) handleEncryptionLayer(ctx context.Context, message *v1protocol.StatusMessage) error {
	logger := s.logger.With(zap.String("site", "handleEncryptionLayer"))
	publicKey := message.SigPubKey()

	// if it's an ephemeral key, we don't negotiate a topic
	decryptionKey, skipNegotiation := s.fetchDecryptionKey(message.TransportLayer.Dst)

	err := message.HandleEncryptionLayer(decryptionKey, publicKey, s.protocol, skipNegotiation)

	// if it's an ephemeral key, we don't have to handle a device not found error
	if err == encryption.ErrDeviceNotFound && !skipNegotiation {
		if err := s.handleErrDeviceNotFound(ctx, publicKey); err != nil {
			logger.Error("failed to handle ErrDeviceNotFound", zap.Error(err))
		}
	}
	if err != nil {
		logger.Error("failed to handle an encrypted message", zap.Error(err))
		return err
	}

	return nil
}

func (s *MessageSender) handleErrDeviceNotFound(ctx context.Context, publicKey *ecdsa.PublicKey) error {
	now := time.Now().Unix()
	advertise, err := s.protocol.ShouldAdvertiseBundle(publicKey, now)
	if err != nil {
		return err
	}
	if !advertise {
		return nil
	}

	messageSpec, err := s.protocol.BuildBundleAdvertiseMessage(s.identity, publicKey)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	// We don't pass an array of messageIDs as no action needs to be taken
	// when sending a bundle
	_, _, err = s.sendMessageSpec(ctx, publicKey, messageSpec, nil)
	if err != nil {
		return err
	}

	s.protocol.ConfirmBundleAdvertisement(publicKey, now)

	return nil
}

func (s *MessageSender) wrapMessageV1(rawMessage *RawMessage) ([]byte, error) {
	wrappedMessage, err := v1protocol.WrapMessageV1(rawMessage.Payload, rawMessage.MessageType, rawMessage.Sender)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap message")
	}
	return wrappedMessage, nil
}

func (s *MessageSender) addToDataSync(publicKey *ecdsa.PublicKey, message []byte) ([]byte, error) {
	groupID := datasync.ToOneToOneGroupID(&s.identity.PublicKey, publicKey)
	peerID := datasyncpeer.PublicKeyToPeerID(*publicKey)
	exist, err := s.datasync.IsPeerInGroup(groupID, peerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if peer is in group")
	}
	if !exist {
		if err := s.datasync.AddPeer(groupID, peerID); err != nil {
			return nil, errors.Wrap(err, "failed to add peer")
		}
	}
	id, err := s.datasync.AppendMessage(groupID, message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to append message to datasync")
	}

	return id[:], nil
}

// sendPrivateRawMessage sends a message not wrapped in an encryption layer
func (s *MessageSender) sendPrivateRawMessage(ctx context.Context, rawMessage *RawMessage, publicKey *ecdsa.PublicKey, payload []byte) ([][]byte, []*types.NewMessage, error) {
	newMessage := &types.NewMessage{
		TTL:         whisperTTL,
		Payload:     payload,
		PowTarget:   calculatePoW(payload),
		PowTime:     whisperPoWTime,
		PubsubTopic: rawMessage.PubsubTopic,
	}

	newMessages, err := s.segmentMessage(newMessage)
	if err != nil {
		return nil, nil, err
	}

	hashes := make([][]byte, 0, len(newMessages))
	var hash []byte
	for _, newMessage := range newMessages {
		if rawMessage.SendOnPersonalTopic {
			hash, err = s.transport.SendPrivateOnPersonalTopic(ctx, newMessage, publicKey)
		} else {
			hash, err = s.transport.SendPrivateWithPartitioned(ctx, newMessage, publicKey)
		}
		if err != nil {
			return nil, nil, err
		}
		hashes = append(hashes, hash)
	}

	return hashes, newMessages, nil
}

// sendCommunityMessage sends a message not wrapped in an encryption layer
// to a community
func (s *MessageSender) dispatchCommunityMessage(ctx context.Context, publicKey *ecdsa.PublicKey, wrappedMessage []byte, pubsubTopic string, rekey bool, rawMessage *RawMessage) ([][]byte, []*types.NewMessage, error) {
	payload := wrappedMessage
	if rekey && len(rawMessage.HashRatchetGroupID) != 0 {

		var ratchet *encryption.HashRatchetKeyCompatibility
		var err error
		// We have just rekeyed, pull the latest
		if RekeyCompatibility {
			ratchet, err = s.protocol.GetCurrentKeyForGroup(rawMessage.HashRatchetGroupID)
			if err != nil {
				return nil, nil, err
			}

		}
		keyID, err := ratchet.GetKeyID()
		if err != nil {
			return nil, nil, err
		}
		s.logger.Debug("adding key id to message", zap.String("keyid", types.Bytes2Hex(keyID)))
		// We send the message over the community topic
		spec, err := s.protocol.BuildHashRatchetReKeyGroupMessage(s.identity, rawMessage.Recipients, rawMessage.HashRatchetGroupID, wrappedMessage, ratchet)
		if err != nil {
			return nil, nil, err
		}
		payload, err = proto.Marshal(spec.Message)
		if err != nil {
			return nil, nil, err
		}
	}

	newMessage := &types.NewMessage{
		TTL:         whisperTTL,
		Payload:     payload,
		PowTarget:   calculatePoW(payload),
		PowTime:     whisperPoWTime,
		PubsubTopic: pubsubTopic,
	}

	newMessages, err := s.segmentMessage(newMessage)
	if err != nil {
		return nil, nil, err
	}

	hashes := make([][]byte, 0, len(newMessages))
	for _, newMessage := range newMessages {
		hash, err := s.transport.SendCommunityMessage(ctx, newMessage, publicKey)
		if err != nil {
			return nil, nil, err
		}
		hashes = append(hashes, hash)
	}

	return hashes, newMessages, nil
}

func (s *MessageSender) SendMessageSpec(ctx context.Context, publicKey *ecdsa.PublicKey, messageSpec *encryption.ProtocolMessageSpec, messageIDs [][]byte) ([][]byte, []*types.NewMessage, error) {
	return s.sendMessageSpec(ctx, publicKey, messageSpec, messageIDs)
}

// sendMessageSpec analyses the spec properties and selects a proper transport method.
func (s *MessageSender) sendMessageSpec(ctx context.Context, publicKey *ecdsa.PublicKey, messageSpec *encryption.ProtocolMessageSpec, messageIDs [][]byte) ([][]byte, []*types.NewMessage, error) {
	logger := s.logger.With(zap.String("site", "sendMessageSpec"))

	newMessage, err := MessageSpecToWhisper(messageSpec)
	if err != nil {
		return nil, nil, err
	}

	newMessages, err := s.segmentMessage(newMessage)
	if err != nil {
		return nil, nil, err
	}

	hashes := make([][]byte, 0, len(newMessages))
	var hash []byte
	for _, newMessage := range newMessages {
		// The shared secret needs to be handle before we send a message
		// otherwise the topic might not be set up before we receive a message
		if messageSpec.SharedSecret != nil && s.handleSharedSecrets != nil {
			err := s.handleSharedSecrets([]*sharedsecret.Secret{messageSpec.SharedSecret})
			if err != nil {
				return nil, nil, err
			}

		}
		// process shared secret
		if messageSpec.AgreedSecret {
			logger.Debug("sending using shared secret")
			hash, err = s.transport.SendPrivateWithSharedSecret(ctx, newMessage, publicKey, messageSpec.SharedSecret.Key)
		} else {
			logger.Debug("sending partitioned topic")
			hash, err = s.transport.SendPrivateWithPartitioned(ctx, newMessage, publicKey)
		}
		if err != nil {
			return nil, nil, err
		}
		hashes = append(hashes, hash)
	}

	sentMessage := &SentMessage{
		PublicKey:  publicKey,
		Spec:       messageSpec,
		MessageIDs: messageIDs,
	}

	s.notifyOnSentMessage(sentMessage)

	return hashes, newMessages, nil
}

func (s *MessageSender) SubscribeToMessageEvents() <-chan *MessageEvent {
	c := make(chan *MessageEvent, 100)
	s.messageEventsSubscriptions = append(s.messageEventsSubscriptions, c)
	return c
}

func (s *MessageSender) notifyOnSentMessage(sentMessage *SentMessage) {
	event := &MessageEvent{
		Type:        MessageSent,
		SentMessage: sentMessage,
	}
	// Publish on channels, drop if buffer is full
	for _, c := range s.messageEventsSubscriptions {
		select {
		case c <- event:
		default:
			s.logger.Warn("message events subscription channel full when publishing sent event, dropping message")
		}
	}

}

func (s *MessageSender) notifyOnScheduledMessage(recipient *ecdsa.PublicKey, message *RawMessage) {
	event := &MessageEvent{
		Recipient:  recipient,
		Type:       MessageScheduled,
		RawMessage: message,
	}

	// Publish on channels, drop if buffer is full
	for _, c := range s.messageEventsSubscriptions {
		select {
		case c <- event:
		default:
			s.logger.Warn("message events subscription channel full when publishing scheduled event, dropping message")
		}
	}
}

func (s *MessageSender) JoinPublic(id string) (*transport.Filter, error) {
	return s.transport.JoinPublic(id)
}

// AddEphemeralKey adds an ephemeral key that we will be listening to
// note that we never removed them from now, as waku/whisper does not
// recalculate topics on removal, so effectively there's no benefit.
// On restart they will be gone.
func (s *MessageSender) AddEphemeralKey(privateKey *ecdsa.PrivateKey) (*transport.Filter, error) {
	s.ephemeralKeysMutex.Lock()
	s.ephemeralKeys[types.EncodeHex(crypto.FromECDSAPub(&privateKey.PublicKey))] = privateKey
	s.ephemeralKeysMutex.Unlock()
	return s.transport.LoadKeyFilters(privateKey)
}

func MessageSpecToWhisper(spec *encryption.ProtocolMessageSpec) (*types.NewMessage, error) {
	var newMessage *types.NewMessage

	payload, err := proto.Marshal(spec.Message)
	if err != nil {
		return newMessage, err
	}

	newMessage = &types.NewMessage{
		TTL:       whisperTTL,
		Payload:   payload,
		PowTarget: calculatePoW(payload),
		PowTime:   whisperPoWTime,
	}
	return newMessage, nil
}

// calculatePoW returns the PoWTarget to be used.
// We check the size and arbitrarily set it to a lower PoW if the packet is
// greater than 50KB. We do this as the defaultPoW is too high for clients to send
// large messages.
func calculatePoW(payload []byte) float64 {
	if len(payload) > largeSizeInBytes {
		return whisperLargeSizePoW
	}
	return whisperDefaultPoW
}

func (s *MessageSender) StopDatasync() {
	if s.datasync != nil {
		s.datasync.Stop()
	}
}

// GetCurrentKeyForGroup returns the latest key timestampID belonging to a key group
func (s *MessageSender) GetCurrentKeyForGroup(groupID []byte) (*encryption.HashRatchetKeyCompatibility, error) {
	return s.protocol.GetCurrentKeyForGroup(groupID)
}

// GetKeyIDsForGroup returns a slice of key IDs belonging to a given group ID
func (s *MessageSender) GetKeysForGroup(groupID []byte) ([]*encryption.HashRatchetKeyCompatibility, error) {
	return s.protocol.GetKeysForGroup(groupID)
}

// Segments message into smaller chunks if the size exceeds the maximum allowed
func segmentMessage(newMessage *types.NewMessage, maxSegmentSize int) ([]*types.NewMessage, error) {
	if len(newMessage.Payload) <= maxSegmentSize {
		return []*types.NewMessage{newMessage}, nil
	}

	createSegment := func(chunkPayload []byte) (*types.NewMessage, error) {
		copy := &types.NewMessage{}
		err := copier.Copy(copy, newMessage)
		if err != nil {
			return nil, err
		}

		copy.Payload = chunkPayload
		copy.PowTarget = calculatePoW(chunkPayload)
		return copy, nil
	}

	entireMessageHash := crypto.Keccak256(newMessage.Payload)
	payloadSize := len(newMessage.Payload)
	segmentsCount := int(math.Ceil(float64(payloadSize) / float64(maxSegmentSize)))

	var segmentMessages []*types.NewMessage

	for start, index := 0, 0; start < payloadSize; start += maxSegmentSize {
		end := start + maxSegmentSize
		if end > payloadSize {
			end = payloadSize
		}

		chunk := newMessage.Payload[start:end]

		segmentMessageProto := &protobuf.SegmentMessage{
			EntireMessageHash: entireMessageHash,
			Index:             uint32(index),
			SegmentsCount:     uint32(segmentsCount),
			Payload:           chunk,
		}
		chunkPayload, err := proto.Marshal(segmentMessageProto)
		if err != nil {
			return nil, err
		}
		segmentMessage, err := createSegment(chunkPayload)
		if err != nil {
			return nil, err
		}

		segmentMessages = append(segmentMessages, segmentMessage)
		index++
	}

	return segmentMessages, nil
}

func (s *MessageSender) segmentMessage(newMessage *types.NewMessage) ([]*types.NewMessage, error) {
	// We set the max message size to 3/4 of the allowed message size, to leave
	// room for segment message metadata.
	newMessages, err := segmentMessage(newMessage, int(s.transport.MaxMessageSize()/4*3))
	s.logger.Debug("message segmented", zap.Int("segments", len(newMessages)))
	return newMessages, err
}

var ErrMessageSegmentsIncomplete = errors.New("message segments incomplete")
var ErrMessageSegmentsAlreadyCompleted = errors.New("message segments already completed")
var ErrMessageSegmentsInvalidCount = errors.New("invalid segments count")
var ErrMessageSegmentsHashMismatch = errors.New("hash of entire payload does not match")

func (s *MessageSender) handleSegmentationLayer(message *v1protocol.StatusMessage) error {
	logger := s.logger.With(zap.String("site", "handleSegmentationLayer"))
	hlogger := logger.With(zap.String("hash", types.HexBytes(message.TransportLayer.Hash).String()))

	var segmentMessage protobuf.SegmentMessage
	err := proto.Unmarshal(message.TransportLayer.Payload, &segmentMessage)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal SegmentMessage")
	}

	hlogger.Debug("handling message segment", zap.String("EntireMessageHash", types.HexBytes(segmentMessage.EntireMessageHash).String()),
		zap.Uint32("Index", segmentMessage.Index), zap.Uint32("SegmentsCount", segmentMessage.SegmentsCount))

	alreadyCompleted, err := s.persistence.IsMessageAlreadyCompleted(segmentMessage.EntireMessageHash)
	if err != nil {
		return err
	}
	if alreadyCompleted {
		return ErrMessageSegmentsAlreadyCompleted
	}

	if segmentMessage.SegmentsCount < 2 {
		return ErrMessageSegmentsInvalidCount
	}

	err = s.persistence.SaveMessageSegment(&segmentMessage, message.TransportLayer.SigPubKey, time.Now().Unix())
	if err != nil {
		return err
	}

	segments, err := s.persistence.GetMessageSegments(segmentMessage.EntireMessageHash, message.TransportLayer.SigPubKey)
	if err != nil {
		return err
	}

	if len(segments) != int(segmentMessage.SegmentsCount) {
		return ErrMessageSegmentsIncomplete
	}

	// Combine payload
	var entirePayload bytes.Buffer
	for _, segment := range segments {
		_, err := entirePayload.Write(segment.Payload)
		if err != nil {
			return errors.Wrap(err, "failed to write segment payload")
		}
	}

	// Sanity check
	entirePayloadHash := crypto.Keccak256(entirePayload.Bytes())
	if !bytes.Equal(entirePayloadHash, segmentMessage.EntireMessageHash) {
		return ErrMessageSegmentsHashMismatch
	}

	err = s.persistence.CompleteMessageSegments(segmentMessage.EntireMessageHash, message.TransportLayer.SigPubKey, time.Now().Unix())
	if err != nil {
		return err
	}

	message.TransportLayer.Payload = entirePayload.Bytes()

	return nil
}

func (s *MessageSender) CleanupSegments() error {
	weekAgo := time.Now().AddDate(0, 0, -7).Unix()
	monthAgo := time.Now().AddDate(0, -1, 0).Unix()

	err := s.persistence.RemoveMessageSegmentsOlderThan(weekAgo)
	if err != nil {
		return err
	}

	err = s.persistence.RemoveMessageSegmentsCompletedOlderThan(monthAgo)
	if err != nil {
		return err
	}

	return nil
}
