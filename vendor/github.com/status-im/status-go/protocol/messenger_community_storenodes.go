package protocol

import (
	"context"
	"errors"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/storenodes"
	v1protocol "github.com/status-im/status-go/protocol/v1"
)

func (m *Messenger) sendCommunityPublicStorenodesInfo(community *communities.Community, snodes storenodes.Storenodes) error {
	if !community.IsControlNode() {
		return communities.ErrNotControlNode
	}

	clock, _ := m.getLastClockWithRelatedChat()
	pb := &protobuf.CommunityStorenodes{
		Clock:       clock,
		CommunityId: community.ID(),
		Storenodes:  snodes.ToProtobuf(),
		ChainId:     communities.CommunityDescriptionTokenOwnerChainID(community.Description()),
	}
	snPayload, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	signature, err := crypto.Sign(crypto.Keccak256(snPayload), community.PrivateKey())
	if err != nil {
		return err
	}
	signedStorenodesInfo := &protobuf.CommunityPublicStorenodesInfo{
		Signature: signature,
		Payload:   snPayload,
	}
	signedPayload, err := proto.Marshal(signedStorenodesInfo)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload:             signedPayload,
		Sender:              community.PrivateKey(),
		SkipEncryptionLayer: true,
		MessageType:         protobuf.ApplicationMetadataMessage_COMMUNITY_PUBLIC_STORENODES_INFO,
		PubsubTopic:         community.PubsubTopic(),
	}

	_, err = m.sender.SendPublic(context.Background(), community.IDString(), rawMessage)
	return err
}

// HandleCommunityPublicStorenodesInfo will process the control message sent by the community owner on updating the community storenodes for his community (sendCommunityPublicStorenodesInfo).
// The message will be received by many peers that are not interested on that community, so if we don't have this community in our DB we just ignore this message.
func (m *Messenger) HandleCommunityPublicStorenodesInfo(state *ReceivedMessageState, a *protobuf.CommunityPublicStorenodesInfo, statusMessage *v1protocol.StatusMessage) error {
	sn := &protobuf.CommunityStorenodes{}
	err := proto.Unmarshal(a.Payload, sn)
	if err != nil {
		return err
	}
	logger := m.logger.Named("HandleCommunityPublicStorenodesInfo").With(zap.String("communityID", types.EncodeHex(sn.CommunityId)))

	err = m.verifyCommunitySignature(a.Payload, a.Signature, sn.CommunityId, sn.ChainId)
	if err != nil {
		logger.Error("failed to verify community signature", zap.Error(err))
		return err
	}

	// verify if we are interested in this control message
	_, err = m.communitiesManager.GetByID(sn.CommunityId)
	if err != nil {
		if errors.Is(err, communities.ErrOrgNotFound) {
			logger.Debug("ignoring control message, community not found")
			return nil
		}
		logger.Error("failed get community by id", zap.Error(err))
		return err
	}

	if err := m.communityStorenodes.UpdateStorenodesInDB(sn.CommunityId, storenodes.FromProtobuf(sn.Storenodes, sn.Clock), sn.Clock); err != nil {
		logger.Error("failed to update storenodes for community", zap.Error(err))
		return err
	}
	return nil
}
