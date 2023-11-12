package pushnotificationserver

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/crypto/ecies"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

const encryptedPayloadKeyLength = 16
const defaultGorushURL = "https://gorush.status.im"

var errUnhandledPushNotificationType = errors.New("unhandled push notification type")

type Config struct {
	Enabled bool
	// Identity is our identity key
	Identity *ecdsa.PrivateKey
	// GorushUrl is the url for the gorush service
	GorushURL string

	Logger *zap.Logger
}

type Server struct {
	persistence   Persistence
	config        *Config
	messageSender *common.MessageSender
	// SentRequests keeps track of the requests sent to gorush, for testing only
	SentRequests int64
}

func New(config *Config, persistence Persistence, messageSender *common.MessageSender) *Server {
	if len(config.GorushURL) == 0 {
		config.GorushURL = defaultGorushURL

	}
	return &Server{persistence: persistence, config: config, messageSender: messageSender}
}

func (s *Server) Start() error {
	if s.config.Logger == nil {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return errors.Wrap(err, "failed to create a logger")
		}
		s.config.Logger = logger
	}

	s.config.Logger.Info("starting push notification server")
	if s.config.Identity == nil {
		s.config.Logger.Debug("Identity nil")
		// Pull identity from database
		identity, err := s.persistence.GetIdentity()
		if err != nil {
			return err
		}
		if identity == nil {
			identity, err = crypto.GenerateKey()
			if err != nil {
				return err
			}
			if err := s.persistence.SaveIdentity(identity); err != nil {
				return err
			}
		}
		s.config.Identity = identity
	}

	pks, err := s.persistence.GetPushNotificationRegistrationPublicKeys()
	if err != nil {
		return err
	}
	// listen to all topics for users registered
	for _, pk := range pks {
		if err := s.listenToPublicKeyQueryTopic(pk); err != nil {
			return err
		}
	}

	s.config.Logger.Info("started push notification server", zap.String("identity", types.EncodeHex(crypto.FromECDSAPub(&s.config.Identity.PublicKey))))

	return nil
}

// HandlePushNotificationRegistration builds a response for the registration and sends it back to the user
func (s *Server) HandlePushNotificationRegistration(publicKey *ecdsa.PublicKey, payload []byte) error {
	response := s.buildPushNotificationRegistrationResponse(publicKey, payload)
	if response == nil {
		return nil
	}
	encodedMessage, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_PUSH_NOTIFICATION_REGISTRATION_RESPONSE,
		// we skip encryption as might be sent from an ephemeral key
		SkipEncryptionLayer: true,
	}

	_, err = s.messageSender.SendPrivate(context.Background(), publicKey, &rawMessage)
	return err
}

// HandlePushNotificationQuery builds a response for the query and sends it back to the user
func (s *Server) HandlePushNotificationQuery(publicKey *ecdsa.PublicKey, messageID []byte, query *protobuf.PushNotificationQuery) error {
	if query == nil {
		return nil
	}

	response := s.buildPushNotificationQueryResponse(query)
	if response == nil {
		return nil
	}
	response.MessageId = messageID
	encodedMessage, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_PUSH_NOTIFICATION_QUERY_RESPONSE,
		// we skip encryption as sent from an ephemeral key
		SkipEncryptionLayer: true,
	}

	_, err = s.messageSender.SendPrivate(context.Background(), publicKey, &rawMessage)
	return err
}

// HandlePushNotificationRequest will send a gorush notification and send a response back to the user
func (s *Server) HandlePushNotificationRequest(publicKey *ecdsa.PublicKey,
	messageID []byte,
	request *protobuf.PushNotificationRequest) error {
	s.config.Logger.Debug("handling pn request", zap.Binary("message-id", messageID))

	// This is at-most-once semantic for now
	exists, err := s.persistence.PushNotificationExists(messageID)
	if err != nil {
		return err
	}

	if exists {
		s.config.Logger.Debug("already handled")
		return nil
	}

	response, requestsAndRegistrations := s.buildPushNotificationRequestResponse(request)
	//AndSendNotification(&request)
	if response == nil {
		return nil
	}
	err = s.sendPushNotification(requestsAndRegistrations)
	if err != nil {
		s.config.Logger.Error("failed to send go rush notification", zap.Error(err))
		return err
	}
	encodedMessage, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload:     encodedMessage,
		MessageType: protobuf.ApplicationMetadataMessage_PUSH_NOTIFICATION_RESPONSE,
		// We skip encryption here as the message has been sent from an ephemeral key
		SkipEncryptionLayer: true,
	}

	_, err = s.messageSender.SendPrivate(context.Background(), publicKey, &rawMessage)
	return err
}

// buildGrantSignatureMaterial builds a grant for a specific server.
// We use 3 components:
// 1) The client public key. Not sure this applies to our signature scheme, but best to be conservative. https://crypto.stackexchange.com/questions/15538/given-a-message-and-signature-find-a-public-key-that-makes-the-signature-valid
// 2) The server public key
// 3) The access token
// By verifying this signature, a client can trust the server was instructed to store this access token.
func (s *Server) buildGrantSignatureMaterial(clientPublicKey *ecdsa.PublicKey, serverPublicKey *ecdsa.PublicKey, accessToken string) []byte {
	var signatureMaterial []byte
	signatureMaterial = append(signatureMaterial, crypto.CompressPubkey(clientPublicKey)...)
	signatureMaterial = append(signatureMaterial, crypto.CompressPubkey(serverPublicKey)...)
	signatureMaterial = append(signatureMaterial, []byte(accessToken)...)
	a := crypto.Keccak256(signatureMaterial)
	return a
}

func (s *Server) verifyGrantSignature(clientPublicKey *ecdsa.PublicKey, accessToken string, grant []byte) error {
	signatureMaterial := s.buildGrantSignatureMaterial(clientPublicKey, &s.config.Identity.PublicKey, accessToken)
	recoveredPublicKey, err := crypto.SigToPub(signatureMaterial, grant)
	if err != nil {
		return err
	}

	if !common.IsPubKeyEqual(recoveredPublicKey, clientPublicKey) {
		return errors.New("pubkey mismatch")
	}
	return nil

}

func (s *Server) generateSharedKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	return ecies.ImportECDSA(s.config.Identity).GenerateShared(
		ecies.ImportECDSAPublic(publicKey),
		encryptedPayloadKeyLength,
		encryptedPayloadKeyLength,
	)
}

func (s *Server) validateUUID(u string) error {
	if len(u) == 0 {
		return errors.New("empty uuid")
	}
	_, err := uuid.Parse(u)
	return err
}

func (s *Server) decryptRegistration(publicKey *ecdsa.PublicKey, payload []byte) ([]byte, error) {
	sharedKey, err := s.generateSharedKey(publicKey)
	if err != nil {
		return nil, err
	}

	return common.Decrypt(payload, sharedKey)
}

// validateRegistration validates a new message against the last one received for a given installationID and and public key
// and return the decrypted message
func (s *Server) validateRegistration(publicKey *ecdsa.PublicKey, payload []byte) (*protobuf.PushNotificationRegistration, error) {
	if payload == nil {
		return nil, ErrEmptyPushNotificationRegistrationPayload
	}

	if publicKey == nil {
		return nil, ErrEmptyPushNotificationRegistrationPublicKey
	}

	decryptedPayload, err := s.decryptRegistration(publicKey, payload)
	if err != nil {
		return nil, err
	}

	registration := &protobuf.PushNotificationRegistration{}

	if err := proto.Unmarshal(decryptedPayload, registration); err != nil {
		return nil, ErrCouldNotUnmarshalPushNotificationRegistration
	}

	if registration.Version < 1 {
		return nil, ErrInvalidPushNotificationRegistrationVersion
	}

	if err := s.validateUUID(registration.InstallationId); err != nil {
		return nil, ErrMalformedPushNotificationRegistrationInstallationID
	}

	previousVersion, err := s.persistence.GetPushNotificationRegistrationVersion(common.HashPublicKey(publicKey), registration.InstallationId)
	if err != nil {
		return nil, err
	}

	if registration.Version <= previousVersion {
		return nil, ErrInvalidPushNotificationRegistrationVersion
	}

	// unregistering message
	if registration.Unregister {
		return registration, nil
	}

	if err := s.validateUUID(registration.AccessToken); err != nil {
		return nil, ErrMalformedPushNotificationRegistrationAccessToken
	}

	if len(registration.Grant) == 0 {
		return nil, ErrMalformedPushNotificationRegistrationGrant
	}

	if err := s.verifyGrantSignature(publicKey, registration.AccessToken, registration.Grant); err != nil {

		s.config.Logger.Error("failed to verify grant", zap.Error(err))
		return nil, ErrMalformedPushNotificationRegistrationGrant
	}

	if len(registration.DeviceToken) == 0 {
		return nil, ErrMalformedPushNotificationRegistrationDeviceToken
	}

	if registration.TokenType == protobuf.PushNotificationRegistration_UNKNOWN_TOKEN_TYPE {
		return nil, ErrUnknownPushNotificationRegistrationTokenType
	}

	return registration, nil
}

// buildPushNotificationQueryResponse check if we have the client information and send them back
func (s *Server) buildPushNotificationQueryResponse(query *protobuf.PushNotificationQuery) *protobuf.PushNotificationQueryResponse {

	s.config.Logger.Debug("handling push notification query")
	response := &protobuf.PushNotificationQueryResponse{}
	if query == nil || len(query.PublicKeys) == 0 {
		return response
	}

	registrations, err := s.persistence.GetPushNotificationRegistrationByPublicKeys(query.PublicKeys)
	if err != nil {
		s.config.Logger.Error("failed to retrieve registration", zap.Error(err))
		return response
	}

	for _, idAndResponse := range registrations {

		registration := idAndResponse.Registration

		info := &protobuf.PushNotificationQueryInfo{
			PublicKey:      idAndResponse.ID,
			Grant:          registration.Grant,
			Version:        registration.Version,
			InstallationId: registration.InstallationId,
		}

		// if instructed to only allow from contacts, send back a list
		if registration.AllowFromContactsOnly {
			info.AllowedKeyList = registration.AllowedKeyList
		} else {
			info.AccessToken = registration.AccessToken
		}
		response.Info = append(response.Info, info)
	}

	response.Success = true
	return response
}

func (s *Server) contains(list [][]byte, chatID []byte) bool {
	for _, list := range list {
		if bytes.Equal(list, chatID) {
			return true
		}
	}
	return false
}

type reportResult struct {
	sendNotification bool
	report           *protobuf.PushNotificationReport
}

// buildPushNotificationReport checks the request against the registration and
// returns whether we should send the notification and what the response should be
func (s *Server) buildPushNotificationReport(pn *protobuf.PushNotification, registration *protobuf.PushNotificationRegistration) (*reportResult, error) {
	response := &reportResult{}
	report := &protobuf.PushNotificationReport{
		PublicKey:      pn.PublicKey,
		InstallationId: pn.InstallationId,
	}

	if pn.Type == protobuf.PushNotification_UNKNOWN_PUSH_NOTIFICATION_TYPE {
		s.config.Logger.Warn("unhandled type")
		return nil, errUnhandledPushNotificationType
	}

	if registration == nil {
		s.config.Logger.Warn("empty registration")
		report.Error = protobuf.PushNotificationReport_NOT_REGISTERED
	} else if registration.AccessToken != pn.AccessToken {
		s.config.Logger.Debug("invalid token")
		report.Error = protobuf.PushNotificationReport_WRONG_TOKEN
	} else if (s.isMessageNotification(pn) && !s.isValidMessageNotification(pn, registration)) || (s.isMentionNotification(pn) && !s.isValidMentionNotification(pn, registration)) || (s.isRequestToJoinCommunityNotification(pn) && !s.isValidRequestToJoinCommunityNotification(pn, registration)) {
		s.config.Logger.Debug("filtered notification")
		// We report as successful but don't send the notification
		// for privacy reasons, as otherwise we would disclose that
		// the sending client has been blocked or that the registering
		// client has not joined a given public chat
		report.Success = true
	} else {
		response.sendNotification = true
		s.config.Logger.Debug("sending push notification")
		report.Success = true
	}

	response.report = report

	return response, nil
}

// buildPushNotificationRequestResponse will build a response
func (s *Server) buildPushNotificationRequestResponse(request *protobuf.PushNotificationRequest) (*protobuf.PushNotificationResponse, []*RequestAndRegistration) {
	response := &protobuf.PushNotificationResponse{}
	// We don't even send a response in this case
	if request == nil || len(request.MessageId) == 0 {
		s.config.Logger.Warn("empty message id")
		return nil, nil
	}

	response.MessageId = request.MessageId

	// collect successful requests & registrations
	var requestAndRegistrations []*RequestAndRegistration

	for _, pn := range request.Requests {
		registration, err := s.persistence.GetPushNotificationRegistrationByPublicKeyAndInstallationID(pn.PublicKey, pn.InstallationId)
		var report *protobuf.PushNotificationReport
		if err != nil {
			report = &protobuf.PushNotificationReport{
				PublicKey:      pn.PublicKey,
				Error:          protobuf.PushNotificationReport_UNKNOWN_ERROR_TYPE,
				InstallationId: pn.InstallationId,
			}
		} else {
			response, err := s.buildPushNotificationReport(pn, registration)
			if err != nil {
				s.config.Logger.Warn("unhandled type")
				continue
			}

			if response.sendNotification {
				requestAndRegistrations = append(requestAndRegistrations, &RequestAndRegistration{
					Request:      pn,
					Registration: registration,
				})
			}
			report = response.report

		}

		response.Reports = append(response.Reports, report)
	}

	s.config.Logger.Debug("built pn request")
	if len(requestAndRegistrations) == 0 {
		s.config.Logger.Warn("no request and registration")
		return response, nil
	}

	return response, requestAndRegistrations
}

func (s *Server) sendPushNotification(requestAndRegistrations []*RequestAndRegistration) error {
	if len(requestAndRegistrations) == 0 {
		return nil
	}
	s.SentRequests++
	goRushRequest := PushNotificationRegistrationToGoRushRequest(requestAndRegistrations)
	return sendGoRushNotification(goRushRequest, s.config.GorushURL, s.config.Logger)
}

// listenToPublicKeyQueryTopic listen to a topic derived from the hashed public key
func (s *Server) listenToPublicKeyQueryTopic(hashedPublicKey []byte) error {
	if s.messageSender == nil {
		return nil
	}
	encodedPublicKey := hex.EncodeToString(hashedPublicKey)
	_, err := s.messageSender.JoinPublic(encodedPublicKey)
	return err
}

// buildPushNotificationRegistrationResponse will check the registration is valid, save it, and listen to the topic for the queries
func (s *Server) buildPushNotificationRegistrationResponse(publicKey *ecdsa.PublicKey, payload []byte) *protobuf.PushNotificationRegistrationResponse {

	s.config.Logger.Debug("handling push notification registration")
	response := &protobuf.PushNotificationRegistrationResponse{
		RequestId: common.Shake256(payload),
	}

	registration, err := s.validateRegistration(publicKey, payload)

	if err != nil {
		if err == ErrInvalidPushNotificationRegistrationVersion {
			response.Error = protobuf.PushNotificationRegistrationResponse_VERSION_MISMATCH
		} else {
			response.Error = protobuf.PushNotificationRegistrationResponse_MALFORMED_MESSAGE
		}
		s.config.Logger.Warn("registration did not validate", zap.Error(err))
		return response
	}

	if registration.Unregister {
		s.config.Logger.Debug("unregistering client")
		// We save an empty registration, only keeping version and installation-id
		if err := s.persistence.UnregisterPushNotificationRegistration(common.HashPublicKey(publicKey), registration.InstallationId, registration.Version); err != nil {
			response.Error = protobuf.PushNotificationRegistrationResponse_INTERNAL_ERROR
			s.config.Logger.Error("failed to unregister ", zap.Error(err))
			return response
		}

	} else if err := s.persistence.SavePushNotificationRegistration(common.HashPublicKey(publicKey), registration); err != nil {
		response.Error = protobuf.PushNotificationRegistrationResponse_INTERNAL_ERROR
		s.config.Logger.Error("failed to save registration", zap.Error(err))
		return response
	}

	if err := s.listenToPublicKeyQueryTopic(common.HashPublicKey(publicKey)); err != nil {
		response.Error = protobuf.PushNotificationRegistrationResponse_INTERNAL_ERROR
		s.config.Logger.Error("failed to listen to topic", zap.Error(err))
		return response

	}
	response.Success = true

	s.config.Logger.Debug("handled push notification registration successfully")

	return response
}

func (s *Server) isMentionNotification(pn *protobuf.PushNotification) bool {
	return pn.Type == protobuf.PushNotification_MENTION
}

// isValidMentionNotification checks:
// this is a mention
// mentions are enabled
// the user joined the public chat
// the author is not blocked
func (s *Server) isValidMentionNotification(pn *protobuf.PushNotification, registration *protobuf.PushNotificationRegistration) bool {
	return s.isMentionNotification(pn) && !registration.BlockMentions && s.contains(registration.AllowedMentionsChatList, pn.ChatId) && !s.contains(registration.BlockedChatList, pn.Author)
}

func (s *Server) isMessageNotification(pn *protobuf.PushNotification) bool {
	return pn.Type == protobuf.PushNotification_MESSAGE
}

// isValidMessageNotification checks:
// this is a message
// the chat is not muted
// the author is not blocked
func (s *Server) isValidMessageNotification(pn *protobuf.PushNotification, registration *protobuf.PushNotificationRegistration) bool {
	return s.isMessageNotification(pn) && !s.contains(registration.BlockedChatList, pn.ChatId) && !s.contains(registration.MutedChatList, pn.ChatId) && !s.contains(registration.BlockedChatList, pn.Author)
}

func (s *Server) isRequestToJoinCommunityNotification(pn *protobuf.PushNotification) bool {
	return pn.Type == protobuf.PushNotification_REQUEST_TO_JOIN_COMMUNITY
}

// isValidRequestToJoinCommunityNotification checks:
// this is a request to join a community
// the author is not blocked
func (s *Server) isValidRequestToJoinCommunityNotification(pn *protobuf.PushNotification, registration *protobuf.PushNotificationRegistration) bool {
	return s.isRequestToJoinCommunityNotification(pn) && !s.contains(registration.BlockedChatList, pn.Author)
}
