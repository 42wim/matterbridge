package communities

import (
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

type DescriptionEncryptor interface {
	encryptCommunityDescription(community *Community, d *protobuf.CommunityDescription) (string, []byte, error)
	encryptCommunityDescriptionChannel(community *Community, channelID string, d *protobuf.CommunityDescription) (string, []byte, error)
	decryptCommunityDescription(keyIDSeqNo string, d []byte) (*DecryptCommunityResponse, error)
}

// Encrypts members and chats
func encryptDescription(encryptor DescriptionEncryptor, community *Community, description *protobuf.CommunityDescription) error {
	description.PrivateData = make(map[string][]byte)

	for channelID, channel := range description.Chats {
		if !community.channelEncrypted(channelID) {
			continue
		}

		descriptionToEncrypt := &protobuf.CommunityDescription{
			Chats: map[string]*protobuf.CommunityChat{
				channelID: channel,
			},
		}

		keyIDSeqNo, encryptedDescription, err := encryptor.encryptCommunityDescriptionChannel(community, channelID, descriptionToEncrypt)
		if err != nil {
			return err
		}

		// Set private data and cleanup unencrypted channel's members
		description.PrivateData[keyIDSeqNo] = encryptedDescription
		channel.Members = make(map[string]*protobuf.CommunityMember)
	}

	if community.Encrypted() {
		descriptionToEncrypt := &protobuf.CommunityDescription{
			Members:            description.Members,
			ActiveMembersCount: description.ActiveMembersCount,
			Chats:              description.Chats,
			Categories:         description.Categories,
		}

		keyIDSeqNo, encryptedDescription, err := encryptor.encryptCommunityDescription(community, descriptionToEncrypt)
		if err != nil {
			return err
		}

		// Set private data and cleanup unencrypted members, chats and categories
		description.PrivateData[keyIDSeqNo] = encryptedDescription
		description.Members = make(map[string]*protobuf.CommunityMember)
		description.ActiveMembersCount = 0
		description.Chats = make(map[string]*protobuf.CommunityChat)
		description.Categories = make(map[string]*protobuf.CommunityCategory)
	}

	return nil
}

type CommunityPrivateDataFailedToDecrypt struct {
	GroupID []byte
	KeyID   []byte
}

// Decrypts members and chats
func decryptDescription(id types.HexBytes, encryptor DescriptionEncryptor, description *protobuf.CommunityDescription, logger *zap.Logger) ([]*CommunityPrivateDataFailedToDecrypt, error) {
	if len(description.PrivateData) == 0 {
		return nil, nil
	}

	var failedToDecrypt []*CommunityPrivateDataFailedToDecrypt

	for keyIDSeqNo, encryptedDescription := range description.PrivateData {
		decryptedDescriptionResponse, err := encryptor.decryptCommunityDescription(keyIDSeqNo, encryptedDescription)
		if decryptedDescriptionResponse != nil && !decryptedDescriptionResponse.Decrypted {
			failedToDecrypt = append(failedToDecrypt, &CommunityPrivateDataFailedToDecrypt{GroupID: id, KeyID: decryptedDescriptionResponse.KeyID})
		}
		if err != nil {
			// ignore error, try to decrypt next data
			logger.Debug("failed to decrypt community private data", zap.String("keyIDSeqNo", keyIDSeqNo), zap.Error(err))
			continue
		}
		decryptedDescription := decryptedDescriptionResponse.Description

		if len(decryptedDescription.Members) > 0 {
			description.Members = decryptedDescription.Members
		}

		if decryptedDescription.ActiveMembersCount > 0 {
			description.ActiveMembersCount = decryptedDescription.ActiveMembersCount
		}

		for id, decryptedChannel := range decryptedDescription.Chats {
			if description.Chats == nil {
				description.Chats = make(map[string]*protobuf.CommunityChat)
			}

			if channel := description.Chats[id]; channel != nil {
				if len(decryptedChannel.Members) > 0 {
					channel.Members = decryptedChannel.Members
				}
			} else {
				description.Chats[id] = decryptedChannel
			}
		}

		if len(decryptedDescription.Categories) > 0 {
			description.Categories = decryptedDescription.Categories
		}
	}

	return failedToDecrypt, nil
}
