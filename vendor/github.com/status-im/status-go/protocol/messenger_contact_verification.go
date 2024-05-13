package protocol

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"

	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	v1protocol "github.com/status-im/status-go/protocol/v1"
	"github.com/status-im/status-go/protocol/verification"
)

const minContactVerificationMessageLen = 1
const maxContactVerificationMessageLen = 280

func (m *Messenger) SendContactVerificationRequest(ctx context.Context, contactID string, challenge string) (*MessengerResponse, error) {
	if len(challenge) < minContactVerificationMessageLen || len(challenge) > maxContactVerificationMessageLen {
		return nil, errors.New("invalid verification request challenge length")
	}

	contact, ok := m.allContacts.Load(contactID)
	if !ok || !contact.mutual() {
		return nil, errors.New("must be a mutual contact")
	}

	verifRequest := &verification.Request{
		From:          common.PubkeyToHex(&m.identity.PublicKey),
		To:            contact.ID,
		Challenge:     challenge,
		RequestStatus: verification.RequestStatusPENDING,
		RepliedAt:     0,
	}

	chat, ok := m.allChats.Load(contactID)
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return nil, err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	m.allChats.Store(chat.ID, chat)
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	request := &protobuf.RequestContactVerification{
		Clock:     clock,
		Challenge: challenge,
	}

	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_REQUEST_CONTACT_VERIFICATION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	contact.VerificationStatus = VerificationStatusVERIFYING
	contact.LastUpdatedLocally = m.getTimesource().GetCurrentTime()

	err = m.persistence.SaveContact(contact, nil)
	if err != nil {
		return nil, err
	}

	// We sync the contact with the other devices
	err = m.syncContact(context.Background(), contact, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	m.allContacts.Store(contact.ID, contact)

	verifRequest.RequestedAt = clock
	verifRequest.ID = rawMessage.ID

	err = m.verificationDatabase.SaveVerificationRequest(verifRequest)
	if err != nil {
		return nil, err
	}

	err = m.SyncVerificationRequest(context.Background(), verifRequest, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	chatMessage, err := m.createLocalContactVerificationMessage(request.Challenge, chat, rawMessage.ID, common.ContactVerificationStatePending)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{chatMessage})
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}

	response.AddVerificationRequest(verifRequest)

	err = m.createOrUpdateOutgoingContactVerificationNotification(contact, response, verifRequest, chatMessage, nil)
	if err != nil {
		return nil, err
	}

	response.AddMessage(chatMessage)

	err = m.prepareMessages(response.messages)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (m *Messenger) GetVerificationRequestSentTo(ctx context.Context, contactID string) (*verification.Request, error) {
	_, ok := m.allContacts.Load(contactID)
	if !ok {
		return nil, errors.New("contact not found")
	}

	return m.verificationDatabase.GetLatestVerificationRequestSentTo(contactID)
}

func (m *Messenger) GetReceivedVerificationRequests(ctx context.Context) ([]*verification.Request, error) {
	myPubKey := hexutil.Encode(crypto.FromECDSAPub(&m.identity.PublicKey))
	return m.verificationDatabase.GetReceivedVerificationRequests(myPubKey)
}

func (m *Messenger) CancelVerificationRequest(ctx context.Context, id string) (*MessengerResponse, error) {
	verifRequest, err := m.verificationDatabase.GetVerificationRequest(id)
	if err != nil {
		return nil, err
	}

	if verifRequest == nil {
		m.logger.Error("could not find verification request with id", zap.String("id", id))
		return nil, verification.ErrVerificationRequestNotFound
	}

	if verifRequest.From != common.PubkeyToHex(&m.identity.PublicKey) {
		return nil, errors.New("Can cancel only outgoing contact request")
	}

	contactID := verifRequest.To
	contact, ok := m.allContacts.Load(contactID)
	if !ok || !contact.mutual() {
		return nil, errors.New("Can't find contact for canceling verification request")
	}

	if verifRequest.RequestStatus != verification.RequestStatusPENDING {
		return nil, errors.New("can cancel only pending verification request")
	}

	verifRequest.RequestStatus = verification.RequestStatusCANCELED
	err = m.verificationDatabase.SaveVerificationRequest(verifRequest)
	if err != nil {
		return nil, err
	}
	contact.VerificationStatus = VerificationStatusUNVERIFIED
	contact.LastUpdatedLocally = m.getTimesource().GetCurrentTime()

	err = m.persistence.SaveContact(contact, nil)
	if err != nil {
		return nil, err
	}

	// We sync the contact with the other devices
	err = m.syncContact(context.Background(), contact, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	m.allContacts.Store(contact.ID, contact)

	// NOTE: does we need it?
	chat, ok := m.allChats.Load(verifRequest.To)
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return nil, err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	m.allChats.Store(chat.ID, chat)
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	response := &MessengerResponse{}

	response.AddVerificationRequest(verifRequest)

	response.AddContact(contact)

	err = m.SyncVerificationRequest(context.Background(), verifRequest, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	request := &protobuf.CancelContactVerification{
		Id:    id,
		Clock: clock,
	}

	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_CANCEL_CONTACT_VERIFICATION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	notification, err := m.persistence.GetActivityCenterNotificationByID(types.FromHex(id))
	if err != nil {
		return nil, err
	}

	if notification != nil {
		notification.ContactVerificationStatus = verification.RequestStatusCANCELED
		message := notification.Message
		message.ContactVerificationState = common.ContactVerificationStateCanceled
		notification.Read = true
		notification.UpdatedAt = m.GetCurrentTimeInMillis()

		err = m.addActivityCenterNotification(response, notification, m.syncActivityCenterReadByIDs)
		if err != nil {
			m.logger.Error("failed to save notification", zap.Error(err))
			return nil, err
		}
	}

	return response, nil
}

func (m *Messenger) AcceptContactVerificationRequest(ctx context.Context, id string, response string) (*MessengerResponse, error) {
	verifRequest, err := m.verificationDatabase.GetVerificationRequest(id)
	if err != nil {
		return nil, err
	}

	if verifRequest == nil {
		m.logger.Error("could not find verification request with id", zap.String("id", id))
		return nil, verification.ErrVerificationRequestNotFound
	}

	contactID := verifRequest.From

	contact, ok := m.allContacts.Load(contactID)
	if !ok || !contact.mutual() {
		return nil, errors.New("must be a mutual contact")
	}

	chat, ok := m.allChats.Load(contactID)
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return nil, err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	m.allChats.Store(chat.ID, chat)
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	err = m.verificationDatabase.AcceptContactVerificationRequest(id, response)
	if err != nil {
		return nil, err
	}

	verifRequest, err = m.verificationDatabase.GetVerificationRequest(id)
	if err != nil {
		return nil, err
	}

	err = m.SyncVerificationRequest(context.Background(), verifRequest, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	request := &protobuf.AcceptContactVerification{
		Clock:    clock,
		Id:       verifRequest.ID,
		Response: response,
	}

	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawMessage, err := m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_ACCEPT_CONTACT_VERIFICATION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	// Pull one from the db if there
	notification, err := m.persistence.GetActivityCenterNotificationByID(types.FromHex(id))
	if err != nil {
		return nil, err
	}
	resp := &MessengerResponse{}

	resp.AddVerificationRequest(verifRequest)

	replyMessage, err := m.createLocalContactVerificationMessage(response, chat, rawMessage.ID, common.ContactVerificationStateAccepted)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SaveMessages([]*common.Message{replyMessage})
	if err != nil {
		return nil, err
	}

	resp.AddMessage(replyMessage)

	if notification != nil {
		// TODO: Should we update only the message or only the notification or both?

		notification.ContactVerificationStatus = verification.RequestStatusACCEPTED
		message := notification.Message
		message.ContactVerificationState = common.ContactVerificationStateAccepted
		notification.ReplyMessage = replyMessage
		notification.Read = true
		notification.Accepted = true
		notification.UpdatedAt = m.GetCurrentTimeInMillis()
		err = m.addActivityCenterNotification(resp, notification, m.syncActivityCenterAcceptedByIDs)
		if err != nil {
			m.logger.Error("failed to save notification", zap.Error(err))
			return nil, err
		}
		resp.AddMessage(message) // <=== wasn't typo?
	}

	return resp, nil
}

func (m *Messenger) VerifiedTrusted(ctx context.Context, request *requests.VerifiedTrusted) (*MessengerResponse, error) {
	err := request.Validate()
	if err != nil {
		return nil, err
	}
	// Pull one from the db if there
	notification, err := m.persistence.GetActivityCenterNotificationByID(request.ID)
	if err != nil {
		return nil, err
	}

	if notification == nil || notification.ReplyMessage == nil {
		return nil, errors.New("could not find notification")
	}

	contactID := notification.ReplyMessage.From

	contact, ok := m.allContacts.Load(contactID)
	if !ok || !contact.mutual() {
		return nil, errors.New("must be a mutual contact")
	}

	err = m.verificationDatabase.SetTrustStatus(contactID, verification.TrustStatusTRUSTED, m.getTimesource().GetCurrentTime())
	if err != nil {
		return nil, err
	}

	err = m.SyncTrustedUser(context.Background(), contactID, verification.TrustStatusTRUSTED, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	contact.VerificationStatus = VerificationStatusVERIFIED
	contact.LastUpdatedLocally = m.getTimesource().GetCurrentTime()
	err = m.persistence.SaveContact(contact, nil)
	if err != nil {
		return nil, err
	}

	chat, ok := m.allChats.Load(contactID)
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return nil, err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	verifRequest, err := m.verificationDatabase.GetLatestVerificationRequestSentTo(contactID)
	if err != nil {
		return nil, err
	}

	if verifRequest == nil {
		return nil, errors.New("no contact verification found")
	}

	verifRequest.RequestStatus = verification.RequestStatusTRUSTED
	verifRequest.RepliedAt = clock
	err = m.verificationDatabase.SaveVerificationRequest(verifRequest)
	if err != nil {
		return nil, err
	}

	err = m.SyncVerificationRequest(context.Background(), verifRequest, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	// We sync the contact with the other devices
	err = m.syncContact(context.Background(), contact, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	// Dispatch profile message to save a contact to the encrypted profile part
	err = m.DispatchProfileShowcase()
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}

	notification.ContactVerificationStatus = verification.RequestStatusTRUSTED
	notification.Message.ContactVerificationState = common.ContactVerificationStateTrusted
	notification.Read = true
	notification.Accepted = true
	notification.UpdatedAt = m.GetCurrentTimeInMillis()

	err = m.addActivityCenterNotification(response, notification, m.syncActivityCenterAcceptedByIDs)
	if err != nil {
		m.logger.Error("failed to save notification", zap.Error(err))
		return nil, err
	}

	msg, err := m.persistence.MessageByID(notification.ReplyMessage.ID)
	if err != nil {
		return nil, err
	}
	msg.ContactVerificationState = common.ContactVerificationStateTrusted

	err = m.persistence.SaveMessages([]*common.Message{msg})
	if err != nil {
		return nil, err
	}
	response.AddMessage(msg)

	response.AddContact(contact)

	return response, nil
}

func (m *Messenger) VerifiedUntrustworthy(ctx context.Context, request *requests.VerifiedUntrustworthy) (*MessengerResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Pull one from the db if there
	notification, err := m.persistence.GetActivityCenterNotificationByID(request.ID)
	if err != nil {
		return nil, err
	}

	if notification == nil || notification.ReplyMessage == nil {
		return nil, errors.New("could not find notification")
	}

	contactID := notification.ReplyMessage.From

	contact, ok := m.allContacts.Load(contactID)
	if !ok || !contact.mutual() {
		return nil, errors.New("must be a mutual contact")
	}

	err = m.verificationDatabase.SetTrustStatus(contactID, verification.TrustStatusUNTRUSTWORTHY, m.getTimesource().GetCurrentTime())
	if err != nil {
		return nil, err
	}

	err = m.SyncTrustedUser(context.Background(), contactID, verification.TrustStatusUNTRUSTWORTHY, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	contact.VerificationStatus = VerificationStatusVERIFIED
	contact.LastUpdatedLocally = m.getTimesource().GetCurrentTime()
	err = m.persistence.SaveContact(contact, nil)
	if err != nil {
		return nil, err
	}

	chat, ok := m.allChats.Load(contactID)
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return nil, err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	verifRequest, err := m.verificationDatabase.GetLatestVerificationRequestSentTo(contactID)
	if err != nil {
		return nil, err
	}

	if verifRequest == nil {
		return nil, errors.New("no contact verification found")
	}

	verifRequest.RequestStatus = verification.RequestStatusUNTRUSTWORTHY
	verifRequest.RepliedAt = clock
	err = m.verificationDatabase.SaveVerificationRequest(verifRequest)
	if err != nil {
		return nil, err
	}

	err = m.SyncVerificationRequest(context.Background(), verifRequest, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	// We sync the contact with the other devices
	err = m.syncContact(context.Background(), contact, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	// Dispatch profile message to remove a contact from the encrypted profile part
	err = m.DispatchProfileShowcase()
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}

	notification.ContactVerificationStatus = verification.RequestStatusUNTRUSTWORTHY
	notification.Message.ContactVerificationState = common.ContactVerificationStateUntrustworthy
	notification.Read = true
	notification.Accepted = true
	notification.UpdatedAt = m.GetCurrentTimeInMillis()

	err = m.addActivityCenterNotification(response, notification, m.syncActivityCenterAcceptedByIDs)
	if err != nil {
		m.logger.Error("failed to save notification", zap.Error(err))
		return nil, err
	}

	msg, err := m.persistence.MessageByID(notification.ReplyMessage.ID)
	if err != nil {
		return nil, err
	}
	msg.ContactVerificationState = common.ContactVerificationStateUntrustworthy

	err = m.persistence.SaveMessages([]*common.Message{msg})
	if err != nil {
		return nil, err
	}

	response.AddMessage(msg)

	return response, nil
}

func (m *Messenger) DeclineContactVerificationRequest(ctx context.Context, id string) (*MessengerResponse, error) {
	verifRequest, err := m.verificationDatabase.GetVerificationRequest(id)
	if err != nil {
		return nil, err
	}

	if verifRequest == nil {
		m.logger.Error("could not find verification request with id", zap.String("id", id))
		return nil, verification.ErrVerificationRequestNotFound
	}

	contact, ok := m.allContacts.Load(verifRequest.From)
	if !ok || !contact.mutual() {
		return nil, errors.New("must be a mutual contact")
	}

	if verifRequest == nil {
		return nil, errors.New("no contact verification found")
	}

	chat, ok := m.allChats.Load(verifRequest.From)
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return nil, err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	m.allChats.Store(chat.ID, chat)
	clock, _ := chat.NextClockAndTimestamp(m.getTimesource())

	verifRequest.RequestStatus = verification.RequestStatusDECLINED
	verifRequest.RepliedAt = clock
	err = m.verificationDatabase.SaveVerificationRequest(verifRequest)
	if err != nil {
		return nil, err
	}

	response := &MessengerResponse{}

	response.AddVerificationRequest(verifRequest)

	err = m.SyncVerificationRequest(context.Background(), verifRequest, m.dispatchMessage)
	if err != nil {
		return nil, err
	}

	request := &protobuf.DeclineContactVerification{
		Id:    id,
		Clock: clock,
	}

	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	_, err = m.dispatchMessage(ctx, common.RawMessage{
		LocalChatID:         chat.ID,
		Payload:             encodedMessage,
		MessageType:         protobuf.ApplicationMetadataMessage_DECLINE_CONTACT_VERIFICATION,
		ResendAutomatically: true,
	})

	if err != nil {
		return nil, err
	}

	err = m.verificationDatabase.DeclineContactVerificationRequest(id)
	if err != nil {
		return nil, err
	}

	notification, err := m.persistence.GetActivityCenterNotificationByID(types.FromHex(id))
	if err != nil {
		return nil, err
	}

	if notification != nil {
		notification.ContactVerificationStatus = verification.RequestStatusDECLINED
		notification.Read = true
		notification.Dismissed = true
		notification.UpdatedAt = m.GetCurrentTimeInMillis()

		message := notification.Message
		message.ContactVerificationState = common.ContactVerificationStateDeclined

		err = m.addActivityCenterNotification(response, notification, m.syncActivityCenterDismissedByIDs)
		if err != nil {
			m.logger.Error("failed to save notification", zap.Error(err))
			return nil, err
		}
		response.AddMessage(message)
	}

	return response, nil
}

func (m *Messenger) MarkAsTrusted(ctx context.Context, contactID string) error {
	err := m.verificationDatabase.SetTrustStatus(contactID, verification.TrustStatusTRUSTED, m.getTimesource().GetCurrentTime())
	if err != nil {
		return err
	}

	return m.SyncTrustedUser(ctx, contactID, verification.TrustStatusTRUSTED, m.dispatchMessage)
}

func (m *Messenger) MarkAsUntrustworthy(ctx context.Context, contactID string) error {
	err := m.verificationDatabase.SetTrustStatus(contactID, verification.TrustStatusUNTRUSTWORTHY, m.getTimesource().GetCurrentTime())
	if err != nil {
		return err
	}

	return m.SyncTrustedUser(ctx, contactID, verification.TrustStatusUNTRUSTWORTHY, m.dispatchMessage)
}

func (m *Messenger) RemoveTrustStatus(ctx context.Context, contactID string) error {
	err := m.verificationDatabase.SetTrustStatus(contactID, verification.TrustStatusUNKNOWN, m.getTimesource().GetCurrentTime())
	if err != nil {
		return err
	}

	// Dispatch profile message to remove a contact from the encrypted profile part
	err = m.DispatchProfileShowcase()
	if err != nil {
		return err
	}

	return m.SyncTrustedUser(ctx, contactID, verification.TrustStatusUNKNOWN, m.dispatchMessage)
}

func (m *Messenger) GetTrustStatus(contactID string) (verification.TrustStatus, error) {
	return m.verificationDatabase.GetTrustStatus(contactID)
}

func ValidateContactVerificationRequest(request *protobuf.RequestContactVerification) error {
	challengeLen := len(strings.TrimSpace(request.Challenge))
	if challengeLen < minContactVerificationMessageLen || challengeLen > maxContactVerificationMessageLen {
		return errors.New("invalid verification request challenge length")
	}

	return nil
}

func (m *Messenger) HandleRequestContactVerification(state *ReceivedMessageState, request *protobuf.RequestContactVerification, statusMessage *v1protocol.StatusMessage) error {
	if err := ValidateContactVerificationRequest(request); err != nil {
		m.logger.Debug("Invalid verification request", zap.Error(err))
		return err
	}

	id := state.CurrentMessageState.MessageID

	if common.IsPubKeyEqual(state.CurrentMessageState.PublicKey, &m.identity.PublicKey) {
		return nil // Is ours, do nothing
	}

	myPubKey := hexutil.Encode(crypto.FromECDSAPub(&m.identity.PublicKey))
	contactID := hexutil.Encode(crypto.FromECDSAPub(state.CurrentMessageState.PublicKey))

	contact := state.CurrentMessageState.Contact
	if !contact.mutual() {
		m.logger.Debug("Received a verification request for a non added mutual contact", zap.String("contactID", contactID))
		return errors.New("must be a mutual contact")
	}

	persistedVR, err := m.verificationDatabase.GetVerificationRequest(id)
	if err != nil {
		m.logger.Debug("Error obtaining verification request", zap.Error(err))
		return err
	}

	if persistedVR != nil && persistedVR.RequestedAt > request.Clock {
		return nil // older message, ignore it
	}

	if persistedVR == nil {
		// This is a new verification request, and we have not received its acceptance/decline before
		persistedVR = &verification.Request{}
		persistedVR.ID = id
		persistedVR.From = contactID
		persistedVR.To = myPubKey
		persistedVR.RequestStatus = verification.RequestStatusPENDING
	}

	if persistedVR.From != contactID {
		return errors.New("mismatch contactID and ID")
	}

	persistedVR.Challenge = request.Challenge
	persistedVR.RequestedAt = request.Clock

	err = m.verificationDatabase.SaveVerificationRequest(persistedVR)
	if err != nil {
		m.logger.Debug("Error storing verification request", zap.Error(err))
		return err
	}
	m.logger.Info("SAVED", zap.String("id", persistedVR.ID))

	err = m.SyncVerificationRequest(context.Background(), persistedVR, m.dispatchMessage)
	if err != nil {
		return err
	}

	chat, ok := m.allChats.Load(contactID)
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	m.allChats.Store(chat.ID, chat)

	chatMessage, err := m.createContactVerificationMessage(request.Challenge, chat, state, common.ContactVerificationStatePending)
	if err != nil {
		return err
	}

	state.Response.AddMessage(chatMessage)

	state.AllVerificationRequests = append(state.AllVerificationRequests, persistedVR)

	return m.createOrUpdateIncomingContactVerificationNotification(contact, state, persistedVR, chatMessage, nil)
}

func ValidateAcceptContactVerification(request *protobuf.AcceptContactVerification) error {
	responseLen := len(strings.TrimSpace(request.Response))
	if responseLen < minContactVerificationMessageLen || responseLen > maxContactVerificationMessageLen {
		return errors.New("invalid verification request response length")
	}

	return nil
}

func (m *Messenger) HandleAcceptContactVerification(state *ReceivedMessageState, request *protobuf.AcceptContactVerification, statusMessage *v1protocol.StatusMessage) error {
	if err := ValidateAcceptContactVerification(request); err != nil {
		m.logger.Debug("Invalid AcceptContactVerification", zap.Error(err))
		return err
	}

	if common.IsPubKeyEqual(state.CurrentMessageState.PublicKey, &m.identity.PublicKey) {
		return nil // Is ours, do nothing
	}

	myPubKey := hexutil.Encode(crypto.FromECDSAPub(&m.identity.PublicKey))
	contactID := hexutil.Encode(crypto.FromECDSAPub(state.CurrentMessageState.PublicKey))

	contact := state.CurrentMessageState.Contact
	if !contact.mutual() {
		m.logger.Debug("Received a verification response for a non mutual contact", zap.String("contactID", contactID))
		return errors.New("must be a mutual contact")
	}

	persistedVR, err := m.verificationDatabase.GetVerificationRequest(request.Id)
	if err != nil {
		m.logger.Debug("Error obtaining verification request", zap.Error(err))
		return err
	}

	if persistedVR == nil {
		// This is a response for which we have not received its request before
		persistedVR = &verification.Request{}
		persistedVR.ID = request.Id
		persistedVR.From = contactID
		persistedVR.To = myPubKey
	} else {
		if persistedVR.RepliedAt > request.Clock {
			return nil // older message, ignore it
		}

		if persistedVR.RequestStatus == verification.RequestStatusCANCELED {
			return nil // Do nothing, We have already cancelled the verification request
		}
	}

	persistedVR.RequestStatus = verification.RequestStatusACCEPTED
	persistedVR.Response = request.Response
	persistedVR.RepliedAt = request.Clock

	err = m.verificationDatabase.SaveVerificationRequest(persistedVR)
	if err != nil {
		m.logger.Debug("Error storing verification request", zap.Error(err))
		return err
	}

	err = m.SyncVerificationRequest(context.Background(), persistedVR, m.dispatchMessage)
	if err != nil {
		return err
	}

	chat, ok := m.allChats.Load(contactID)
	if !ok {
		publicKey, err := contact.PublicKey()
		if err != nil {
			return err
		}
		chat = OneToOneFromPublicKey(publicKey, m.getTimesource())
		// We don't want to show the chat to the user
		chat.Active = false
	}

	m.allChats.Store(chat.ID, chat)

	chatMessage, err := m.createContactVerificationMessage(request.Response, chat, state, common.ContactVerificationStateAccepted)
	if err != nil {
		return err
	}

	state.Response.AddMessage(chatMessage)

	msg, err := m.persistence.MessageByID(request.Id)
	if err != nil {
		return err
	}
	msg.ContactVerificationState = common.ContactVerificationStateAccepted

	state.Response.AddMessage(msg)

	err = m.createOrUpdateOutgoingContactVerificationNotification(contact, state.Response, persistedVR, msg, chatMessage)
	if err != nil {
		return err
	}

	state.AllVerificationRequests = append(state.AllVerificationRequests, persistedVR)

	return nil
}

func (m *Messenger) HandleDeclineContactVerification(state *ReceivedMessageState, request *protobuf.DeclineContactVerification, statusMessage *v1protocol.StatusMessage) error {
	if common.IsPubKeyEqual(state.CurrentMessageState.PublicKey, &m.identity.PublicKey) {
		return nil // Is ours, do nothing
	}

	myPubKey := hexutil.Encode(crypto.FromECDSAPub(&m.identity.PublicKey))
	contactID := hexutil.Encode(crypto.FromECDSAPub(state.CurrentMessageState.PublicKey))

	contact := state.CurrentMessageState.Contact
	if !contact.mutual() {
		m.logger.Debug("Received a verification decline for a non mutual contact", zap.String("contactID", contactID))
		return errors.New("must be a mutual contact")
	}

	persistedVR, err := m.verificationDatabase.GetVerificationRequest(request.Id)
	if err != nil {
		m.logger.Debug("Error obtaining verification request", zap.Error(err))
		return err
	}

	if persistedVR != nil && persistedVR.RepliedAt > request.Clock {
		return nil // older message, ignore it
	}

	if persistedVR.RequestStatus == verification.RequestStatusCANCELED {
		return nil // Do nothing, We have already cancelled the verification request
	}

	contact.VerificationStatus = VerificationStatusUNVERIFIED
	contact.LastUpdatedLocally = m.getTimesource().GetCurrentTime()

	err = m.persistence.SaveContact(contact, nil)
	if err != nil {
		return err
	}

	// We sync the contact with the other devices
	err = m.syncContact(context.Background(), contact, m.dispatchMessage)
	if err != nil {
		return err
	}

	m.allContacts.Store(contact.ID, contact)

	state.Response.AddContact(contact)

	if persistedVR == nil {
		// This is a response for which we have not received its request before
		persistedVR = &verification.Request{}
		persistedVR.From = contactID
		persistedVR.To = myPubKey
	}

	persistedVR.RequestStatus = verification.RequestStatusDECLINED
	persistedVR.RepliedAt = request.Clock

	err = m.verificationDatabase.SaveVerificationRequest(persistedVR)
	if err != nil {
		m.logger.Debug("Error storing verification request", zap.Error(err))
		return err
	}

	err = m.SyncVerificationRequest(context.Background(), persistedVR, m.dispatchMessage)
	if err != nil {
		return err
	}

	state.AllVerificationRequests = append(state.AllVerificationRequests, persistedVR)

	msg, err := m.persistence.MessageByID(request.Id)
	if err != nil {
		return err
	}

	if msg != nil {
		msg.ContactVerificationState = common.ContactVerificationStateDeclined
		state.Response.AddMessage(msg)
	}

	return m.createOrUpdateOutgoingContactVerificationNotification(contact, state.Response, persistedVR, msg, nil)
}

func (m *Messenger) HandleCancelContactVerification(state *ReceivedMessageState, request *protobuf.CancelContactVerification, statusMessage *v1protocol.StatusMessage) error {
	myPubKey := hexutil.Encode(crypto.FromECDSAPub(&m.identity.PublicKey))
	contactID := hexutil.Encode(crypto.FromECDSAPub(state.CurrentMessageState.PublicKey))

	contact := state.CurrentMessageState.Contact

	persistedVR, err := m.verificationDatabase.GetVerificationRequest(request.Id)
	if err != nil {
		m.logger.Debug("Error obtaining verification request", zap.Error(err))
		return err
	}

	if persistedVR != nil && persistedVR.RequestStatus != verification.RequestStatusPENDING {
		m.logger.Debug("Only pending verification request can be canceled", zap.String("contactID", contactID))
		return errors.New("must be a pending verification request")
	}

	if persistedVR == nil {
		// This is a response for which we have not received its request before
		persistedVR = &verification.Request{}
		persistedVR.From = contactID
		persistedVR.To = myPubKey
	}

	persistedVR.RequestStatus = verification.RequestStatusCANCELED
	persistedVR.RepliedAt = request.Clock

	err = m.verificationDatabase.SaveVerificationRequest(persistedVR)
	if err != nil {
		m.logger.Debug("Error storing verification request", zap.Error(err))
		return err
	}

	err = m.SyncVerificationRequest(context.Background(), persistedVR, m.dispatchMessage)
	if err != nil {
		return err
	}

	state.AllVerificationRequests = append(state.AllVerificationRequests, persistedVR)

	msg, err := m.persistence.MessageByID(request.Id)
	if err != nil {
		return err
	}

	if msg != nil {
		msg.ContactVerificationState = common.ContactVerificationStateCanceled
		state.Response.AddMessage(msg)
	}

	return m.createOrUpdateIncomingContactVerificationNotification(contact, state, persistedVR, msg, nil)
}

func (m *Messenger) GetLatestVerificationRequestFrom(contactID string) (*verification.Request, error) {
	return m.verificationDatabase.GetLatestVerificationRequestFrom(contactID)
}

func (m *Messenger) createOrUpdateOutgoingContactVerificationNotification(contact *Contact, response *MessengerResponse, vr *verification.Request, chatMessage *common.Message, replyMessage *common.Message) error {
	notification := &ActivityCenterNotification{
		ID:                        types.FromHex(vr.ID),
		Name:                      contact.PrimaryName(),
		Type:                      ActivityCenterNotificationTypeContactVerification,
		Author:                    chatMessage.From,
		Message:                   chatMessage,
		ReplyMessage:              replyMessage,
		Timestamp:                 chatMessage.WhisperTimestamp,
		ChatID:                    contact.ID,
		ContactVerificationStatus: vr.RequestStatus,
		Read:                      vr.RequestStatus != verification.RequestStatusACCEPTED, // Mark as Unread Accepted notification because we are waiting for the asnwer
		Accepted:                  vr.RequestStatus == verification.RequestStatusTRUSTED || vr.RequestStatus == verification.RequestStatusUNTRUSTWORTHY,
		Dismissed:                 vr.RequestStatus == verification.RequestStatusDECLINED,
		UpdatedAt:                 m.GetCurrentTimeInMillis(),
	}

	return m.addActivityCenterNotification(response, notification, nil)
}

func (m *Messenger) createOrUpdateIncomingContactVerificationNotification(contact *Contact, messageState *ReceivedMessageState, vr *verification.Request, chatMessage *common.Message, replyMessage *common.Message) error {
	notification := &ActivityCenterNotification{
		ID:                        types.FromHex(vr.ID),
		Name:                      contact.PrimaryName(),
		Type:                      ActivityCenterNotificationTypeContactVerification,
		Author:                    messageState.CurrentMessageState.Contact.ID,
		Message:                   chatMessage,
		ReplyMessage:              replyMessage,
		Timestamp:                 messageState.CurrentMessageState.WhisperTimestamp,
		ChatID:                    contact.ID,
		ContactVerificationStatus: vr.RequestStatus,
		Read:                      vr.RequestStatus != verification.RequestStatusPENDING, // Unread only for pending incomming
		Accepted:                  vr.RequestStatus == verification.RequestStatusACCEPTED || vr.RequestStatus == verification.RequestStatusTRUSTED || vr.RequestStatus == verification.RequestStatusUNTRUSTWORTHY,
		Dismissed:                 vr.RequestStatus == verification.RequestStatusDECLINED,
		UpdatedAt:                 m.GetCurrentTimeInMillis(),
	}

	return m.addActivityCenterNotification(messageState.Response, notification, nil)
}

func (m *Messenger) createContactVerificationMessage(challenge string, chat *Chat, state *ReceivedMessageState, verificationStatus common.ContactVerificationState) (*common.Message, error) {
	chatMessage := common.NewMessage()
	chatMessage.ID = state.CurrentMessageState.MessageID
	chatMessage.From = state.CurrentMessageState.Contact.ID
	chatMessage.Alias = state.CurrentMessageState.Contact.Alias
	chatMessage.SigPubKey = state.CurrentMessageState.PublicKey
	chatMessage.Identicon = state.CurrentMessageState.Contact.Identicon
	chatMessage.WhisperTimestamp = state.CurrentMessageState.WhisperTimestamp

	chatMessage.ChatId = chat.ID
	chatMessage.Text = challenge
	chatMessage.ContentType = protobuf.ChatMessage_IDENTITY_VERIFICATION
	chatMessage.ContactVerificationState = verificationStatus

	err := chatMessage.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}
	return chatMessage, nil
}

func (m *Messenger) createLocalContactVerificationMessage(challenge string, chat *Chat, id string, status common.ContactVerificationState) (*common.Message, error) {

	chatMessage := common.NewMessage()
	chatMessage.ID = id
	err := extendMessageFromChat(chatMessage, chat, &m.identity.PublicKey, m.getTimesource())
	if err != nil {
		return nil, err
	}

	chatMessage.ChatId = chat.ID
	chatMessage.Text = challenge
	chatMessage.ContentType = protobuf.ChatMessage_IDENTITY_VERIFICATION
	chatMessage.ContactVerificationState = status
	err = extendMessageFromChat(chatMessage, chat, &m.identity.PublicKey, m.getTimesource())
	if err != nil {
		return nil, err
	}

	err = chatMessage.PrepareContent(common.PubkeyToHex(&m.identity.PublicKey))
	if err != nil {
		return nil, err
	}
	return chatMessage, nil
}
