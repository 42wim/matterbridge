package protocol

import (
	"context"
	"crypto/ecdsa"
	"errors"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/transport"
	v1protocol "github.com/status-im/status-go/protocol/v1"
)

func (m *Messenger) sendPublicCommunityShardInfo(community *communities.Community) error {
	if !community.IsControlNode() {
		return communities.ErrNotControlNode
	}

	publicShardInfo := &protobuf.PublicShardInfo{
		Clock:       community.Clock(),
		CommunityId: community.ID(),
		Shard:       community.Shard().Protobuffer(),
		ChainId:     communities.CommunityDescriptionTokenOwnerChainID(community.Description()),
	}

	payload, err := proto.Marshal(publicShardInfo)
	if err != nil {
		return err
	}

	signature, err := crypto.Sign(crypto.Keccak256(payload), community.PrivateKey())
	if err != nil {
		return err
	}

	signedShardInfo := &protobuf.CommunityPublicShardInfo{
		Signature: signature,
		Payload:   payload,
	}

	payload, err = proto.Marshal(signedShardInfo)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload: payload,
		Sender:  community.PrivateKey(),
		// we don't want to wrap in an encryption layer message
		SkipEncryptionLayer: true,
		MessageType:         protobuf.ApplicationMetadataMessage_COMMUNITY_PUBLIC_SHARD_INFO,
		PubsubTopic:         shard.DefaultNonProtectedPubsubTopic(), // it must be sent always to default shard pubsub topic
	}

	chatName := transport.CommunityShardInfoTopic(community.IDString())
	_, err = m.sender.SendPublic(context.Background(), chatName, rawMessage)
	return err
}

func (m *Messenger) HandleCommunityPublicShardInfo(state *ReceivedMessageState, a *protobuf.CommunityPublicShardInfo, statusMessage *v1protocol.StatusMessage) error {
	publicShardInfo := &protobuf.PublicShardInfo{}
	err := proto.Unmarshal(a.Payload, publicShardInfo)
	if err != nil {
		return err
	}

	logError := func(err error) {
		m.logger.Error("HandleCommunityPublicShardInfo failed: ", zap.Error(err), zap.String("communityID", types.EncodeHex(publicShardInfo.CommunityId)))
	}

	signer, err := recoverCommunityShardInfoSignature(a)
	if err != nil {
		logError(err)
		return err
	}

	err = m.verifyCommunityPublicShardInfo(publicShardInfo, signer)
	if err != nil {
		logError(err)
		return err
	}

	err = m.communitiesManager.SaveCommunityShard(publicShardInfo.CommunityId, shard.FromProtobuff(publicShardInfo.Shard), publicShardInfo.Clock)
	if err != nil && err != communities.ErrOldShardInfo {
		logError(err)
		return err
	}
	return nil
}

func recoverCommunityShardInfoSignature(rawShardInfo *protobuf.CommunityPublicShardInfo) (*ecdsa.PublicKey, error) {
	if rawShardInfo.Signature == nil || len(rawShardInfo.Signature) == 0 {
		return nil, errors.New("missing shard info signature")
	}

	return crypto.SigToPub(crypto.Keccak256(rawShardInfo.Payload), rawShardInfo.Signature)
}

func (m *Messenger) verifyCommunityPublicShardInfo(publicShardInfo *protobuf.PublicShardInfo, signer *ecdsa.PublicKey) error {
	pubKeyStr := common.PubkeyToHex(signer)

	var ownerPublicKey string
	if publicShardInfo.ChainId > 0 {
		owner, err := m.communitiesManager.SafeGetSignerPubKey(publicShardInfo.ChainId, types.EncodeHex(publicShardInfo.CommunityId))
		if err != nil {
			return err
		}
		ownerPublicKey = owner
	} else {
		communityPubkey, err := crypto.DecompressPubkey(publicShardInfo.CommunityId)
		if err != nil {
			return err
		}
		ownerPublicKey = common.PubkeyToHex(communityPubkey)
	}

	if pubKeyStr != ownerPublicKey {
		return errors.New("signed not by a community owner")
	}
	return nil
}
