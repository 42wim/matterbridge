package communities

import (
	"github.com/golang/protobuf/proto"

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
				channelID: proto.Clone(channel).(*protobuf.CommunityChat),
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
			Members: description.Members,
			Chats:   description.Chats,
		}

		keyIDSeqNo, encryptedDescription, err := encryptor.encryptCommunityDescription(community, descriptionToEncrypt)
		if err != nil {
			return err
		}

		// Set private data and cleanup unencrypted members and chats
		description.PrivateData[keyIDSeqNo] = encryptedDescription
		description.Members = make(map[string]*protobuf.CommunityMember)
		description.Chats = make(map[string]*protobuf.CommunityChat)
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

		for pk, member := range decryptedDescription.Members {
			if description.Members == nil {
				description.Members = make(map[string]*protobuf.CommunityMember)
			}
			description.Members[pk] = member
		}

		for id, decryptedChannel := range decryptedDescription.Chats {
			if description.Chats == nil {
				description.Chats = make(map[string]*protobuf.CommunityChat)
			}

			if channel := description.Chats[id]; channel != nil {
				if len(channel.Members) == 0 {
					channel.Members = decryptedChannel.Members
				}
			} else {
				description.Chats[id] = decryptedChannel
			}
		}
	}

	return failedToDecrypt, nil
}
