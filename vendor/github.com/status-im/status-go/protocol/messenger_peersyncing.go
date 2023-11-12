package protocol

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"time"

	"github.com/golang/protobuf/proto"
	datasyncproto "github.com/status-im/mvds/protobuf"
	"github.com/status-im/mvds/state"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	datasyncpeer "github.com/status-im/status-go/protocol/datasync/peer"
	"github.com/status-im/status-go/protocol/encryption/sharedsecret"
	"github.com/status-im/status-go/protocol/peersyncing"
	v1protocol "github.com/status-im/status-go/protocol/v1"
)

var peerSyncingLoopInterval time.Duration = 60 * time.Second
var maxAdvertiseMessages = 40

func (m *Messenger) markDeliveredMessages(acks [][]byte) {
	for _, ack := range acks {
		//get message ID from database by datasync ID, with at-least-one
		// semantic
		messageIDBytes, err := m.persistence.MarkAsConfirmed(ack, true)
		if err != nil {
			m.logger.Info("got datasync acknowledge for message we don't have in db", zap.String("ack", hex.EncodeToString(ack)))
			continue
		}

		messageID := messageIDBytes.String()
		//mark messages as delivered

		err = m.UpdateMessageOutgoingStatus(messageID, common.OutgoingStatusDelivered)
		if err != nil {
			m.logger.Debug("Can't set message status as delivered", zap.Error(err))
		}

		//send signal to client that message status updated
		if m.config.messengerSignalsHandler != nil {
			message, err := m.persistence.MessageByID(messageID)
			if err != nil {
				m.logger.Debug("Can't get message from database", zap.Error(err))
				continue
			}
			m.config.messengerSignalsHandler.MessageDelivered(message.LocalChatID, messageID)
		}
	}
}

func (m *Messenger) handleDatasyncMetadata(response *common.HandleMessageResponse) error {
	m.OnDatasyncAcks(response.DatasyncSender, response.DatasyncAcks)

	if !m.featureFlags.Peersyncing {
		return nil
	}

	err := m.OnDatasyncOffer(response)
	if err != nil {
		return err
	}

	err = m.OnDatasyncRequests(response.DatasyncSender, response.DatasyncRequests)
	if err != nil {
		return err
	}

	return nil
}

func (m *Messenger) startPeerSyncingLoop() {
	logger := m.logger.Named("PeerSyncingLoop")

	ticker := time.NewTicker(peerSyncingLoopInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := m.sendDatasyncOffers()
				if err != nil {
					m.logger.Warn("failed to send datasync offers", zap.Error(err))
				}

			case <-m.quit:
				ticker.Stop()
				logger.Debug("peersyncing loop stopped")
				return
			}
		}
	}()
}

func (m *Messenger) sendDatasyncOffers() error {
	if !m.featureFlags.Peersyncing {
		return nil
	}

	communities, err := m.communitiesManager.Joined()
	if err != nil {
		return err
	}

	for _, community := range communities {
		var chatIDs [][]byte
		for id := range community.Chats() {
			chatIDs = append(chatIDs, []byte(community.IDString()+id))
		}

		if len(chatIDs) == 0 {
			continue
		}

		availableMessages, err := m.peersyncing.AvailableMessagesByGroupIDs(chatIDs, maxAdvertiseMessages)
		if err != nil {
			return err
		}
		availableMessagesMap := make(map[string][][]byte)
		for _, m := range availableMessages {
			groupID := types.Bytes2Hex(m.GroupID)
			availableMessagesMap[groupID] = append(availableMessagesMap[groupID], m.ID)
		}

		datasyncMessage := &datasyncproto.Payload{}
		if len(availableMessages) == 0 {
			continue
		}
		for groupID, m := range availableMessagesMap {
			datasyncMessage.GroupOffers = append(datasyncMessage.GroupOffers, &datasyncproto.Offer{GroupId: types.Hex2Bytes(groupID), MessageIds: m})
		}
		payload, err := proto.Marshal(datasyncMessage)
		if err != nil {
			return err
		}
		rawMessage := common.RawMessage{
			Payload:             payload,
			Ephemeral:           true,
			SkipApplicationWrap: true,
			PubsubTopic:         community.PubsubTopic(),
		}
		_, err = m.sender.SendPublic(context.Background(), community.IDString(), rawMessage)
		if err != nil {
			return err
		}

	}
	// Check all the group ids that need to be on offer
	// Get all the messages that need to be offered
	// Prepare datasync messages
	// Dispatch them to the right group
	return nil
}

func (m *Messenger) OnDatasyncOffer(response *common.HandleMessageResponse) error {
	sender := response.DatasyncSender
	offers := response.DatasyncOffers

	if len(offers) == 0 {
		return nil
	}

	if common.PubkeyToHex(sender) == m.myHexIdentity() {
		return nil
	}

	var offeredMessages []peersyncing.SyncMessage

	for _, o := range offers {
		offeredMessages = append(offeredMessages, peersyncing.SyncMessage{GroupID: o.GroupID, ID: o.MessageID})
	}

	messagesToFetch, err := m.peersyncing.OnOffer(offeredMessages)
	if err != nil {
		return err
	}

	if len(messagesToFetch) == 0 {
		return nil
	}

	datasyncMessage := &datasyncproto.Payload{}
	for _, msg := range messagesToFetch {
		idString := types.Bytes2Hex(msg.ID)
		lastOffered := m.peersyncingOffers[idString]
		timeNow := m.GetCurrentTimeInMillis() / 1000
		if lastOffered+30 < timeNow {
			m.peersyncingOffers[idString] = timeNow
			datasyncMessage.Requests = append(datasyncMessage.Requests, msg.ID)
		}
	}
	payload, err := proto.Marshal(datasyncMessage)
	if err != nil {
		return err
	}
	rawMessage := common.RawMessage{
		LocalChatID:         common.PubkeyToHex(sender),
		Payload:             payload,
		Ephemeral:           true,
		SkipApplicationWrap: true,
	}
	_, err = m.sender.SendPrivate(context.Background(), sender, &rawMessage)
	if err != nil {
		return err
	}

	// Check if any of the things need to be added
	// Reply if anything needs adding
	// Ack any message that is out
	return nil
}

// canSyncMessageWith checks the permission of a message
func (m *Messenger) canSyncMessageWith(message peersyncing.SyncMessage, peer *ecdsa.PublicKey) (bool, error) {
	switch message.Type {
	case peersyncing.SyncMessageCommunityType:
		chat, ok := m.allChats.Load(string(message.GroupID))
		if !ok {
			return false, nil
		}
		community, err := m.communitiesManager.GetByIDString(chat.CommunityID)
		if err != nil {
			return false, err
		}

		return m.canSyncCommunityMessageWith(chat, community, peer)

	default:
		return false, nil
	}
}

// NOTE: This is not stricly correct. It's possible that you sync a message that has been
// posted after the banning of a user from a community, but before we realized that.
// As an approximation it should be ok, but worth thinking about how to address this.
func (m *Messenger) canSyncCommunityMessageWith(chat *Chat, community *communities.Community, peer *ecdsa.PublicKey) (bool, error) {
	return community.IsMemberInChat(peer, chat.CommunityChatID()), nil
}

func (m *Messenger) OnDatasyncRequests(requester *ecdsa.PublicKey, messageIDs [][]byte) error {
	if len(messageIDs) == 0 {
		return nil
	}

	messages, err := m.peersyncing.MessagesByIDs(messageIDs)
	if err != nil {
		return err
	}
	for _, msg := range messages {
		canSync, err := m.canSyncMessageWith(msg, requester)
		if err != nil {
			return err
		}
		if !canSync {
			continue
		}
		idString := common.PubkeyToHex(requester) + types.Bytes2Hex(msg.ID)
		lastRequested := m.peersyncingRequests[idString]
		timeNow := m.GetCurrentTimeInMillis() / 1000
		if lastRequested+30 < timeNow {
			m.peersyncingRequests[idString] = timeNow

			// Check permissions
			rawMessage := common.RawMessage{
				LocalChatID:         common.PubkeyToHex(requester),
				Payload:             msg.Payload,
				Ephemeral:           true,
				SkipApplicationWrap: true,
			}
			_, err = m.sender.SendPrivate(context.Background(), requester, &rawMessage)
			if err != nil {
				return err
			}
		}

	}
	// no need of group id, since we can derive from message
	return nil
}

func (m *Messenger) OnDatasyncAcks(sender *ecdsa.PublicKey, acks [][]byte) {
	// we should make sure the sender can acknowledge those messages
	m.markDeliveredMessages(acks)
}

// sendDataSync sends a message scheduled by the data sync layer.
// Data Sync layer calls this method "dispatch" function.
func (m *Messenger) sendDataSync(receiver state.PeerID, payload *datasyncproto.Payload) error {
	ctx := context.Background()
	if !payload.IsValid() {
		m.logger.Error("payload is invalid")
		return errors.New("payload is invalid")
	}

	marshalledPayload, err := proto.Marshal(payload)
	if err != nil {
		m.logger.Error("failed to marshal payload")
		return err
	}

	publicKey, err := datasyncpeer.IDToPublicKey(receiver)
	if err != nil {
		m.logger.Error("failed to convert id to public key", zap.Error(err))
		return err
	}

	// Calculate the messageIDs
	messageIDs := make([][]byte, 0, len(payload.Messages))
	hexMessageIDs := make([]string, 0, len(payload.Messages))
	for _, payload := range payload.Messages {
		mid := v1protocol.MessageID(&m.identity.PublicKey, payload.Body)
		messageIDs = append(messageIDs, mid)
		hexMessageIDs = append(hexMessageIDs, mid.String())
	}

	messageSpec, err := m.encryptor.BuildEncryptedMessage(m.identity, publicKey, marshalledPayload)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt message")
	}

	// The shared secret needs to be handle before we send a message
	// otherwise the topic might not be set up before we receive a message
	err = m.handleSharedSecrets([]*sharedsecret.Secret{messageSpec.SharedSecret})
	if err != nil {
		return err
	}

	hashes, newMessages, err := m.sender.SendMessageSpec(ctx, publicKey, messageSpec, messageIDs)
	if err != nil {
		m.logger.Error("failed to send a datasync message", zap.Error(err))
		return err
	}

	m.logger.Debug("sent private messages", zap.Any("messageIDs", hexMessageIDs), zap.Strings("hashes", types.EncodeHexes(hashes)))
	m.transport.TrackMany(messageIDs, hashes, newMessages)

	return nil
}
