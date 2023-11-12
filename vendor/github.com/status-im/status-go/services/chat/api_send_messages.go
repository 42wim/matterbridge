package chat

import (
	"context"

	"github.com/forPelevin/gomoji"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

type SendMessageResponse struct {
	Chat     *Chat             `json:"chat"`
	Messages []*common.Message `json:"messages"`
}

func (api *API) SendSticker(ctx context.Context, communityID types.HexBytes, chatID string, packID int32, hash string, responseTo string) (*SendMessageResponse, error) {
	ensName, _ := api.s.accountsDB.GetPreferredUsername()

	msg := &common.Message{
		CommunityID: string(communityID.Bytes()),
		ChatMessage: &protobuf.ChatMessage{
			ChatId:      chatID,
			ContentType: protobuf.ChatMessage_STICKER,
			Text:        "Update to latest version to see a nice sticker here!",
			Payload: &protobuf.ChatMessage_Sticker{
				Sticker: &protobuf.StickerMessage{
					Hash: hash,
					Pack: packID,
				},
			},
			ResponseTo: responseTo,
			EnsName:    ensName,
		},
	}

	response, err := api.s.messenger.SendChatMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	return api.toSendMessageResponse(response)

}

func (api *API) toSendMessageResponse(response *protocol.MessengerResponse) (*SendMessageResponse, error) {
	protocolChat := response.Chats()[0]

	community, err := api.s.messenger.GetCommunityByID(types.HexBytes(protocolChat.CommunityID))
	if err != nil {
		return nil, err
	}

	pubKey := types.EncodeHex(crypto.FromECDSAPub(api.s.messenger.IdentityPublicKey()))
	chat, err := api.toAPIChat(protocolChat, community, pubKey, false)
	if err != nil {
		return nil, err
	}
	return &SendMessageResponse{
		Chat:     chat,
		Messages: response.Messages(),
	}, nil
}

func isTextOrEmoji(text string) protobuf.ChatMessage_ContentType {
	contentType := protobuf.ChatMessage_TEXT_PLAIN
	if gomoji.RemoveEmojis(text) == "" && len(gomoji.FindAll(text)) != 0 {
		contentType = protobuf.ChatMessage_EMOJI
	}

	return contentType
}

func (api *API) SendMessage(ctx context.Context, communityID types.HexBytes, chatID string, text string, responseTo string) (*SendMessageResponse, error) {
	ensName, _ := api.s.accountsDB.GetPreferredUsername()

	msg := &common.Message{
		CommunityID: string(communityID.Bytes()),
		ChatMessage: &protobuf.ChatMessage{
			ChatId:      chatID,
			ContentType: isTextOrEmoji(text),
			Text:        text,
			ResponseTo:  responseTo,
			EnsName:     ensName,
		},
	}

	response, err := api.s.messenger.SendChatMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	return api.toSendMessageResponse(response)
}

func (api *API) SendImages(ctx context.Context, communityID types.HexBytes, chatID string, imagePaths []string, text string, responseTo string) (*SendMessageResponse, error) {
	ensName, _ := api.s.accountsDB.GetPreferredUsername()

	var messages []*common.Message

	for _, imagePath := range imagePaths {
		messages = append(messages, &common.Message{
			CommunityID: string(communityID.Bytes()),
			ChatMessage: &protobuf.ChatMessage{
				ChatId:      chatID,
				ContentType: protobuf.ChatMessage_IMAGE,
				Text:        "Update to latest version to see a nice image here!",
				ResponseTo:  responseTo,
				EnsName:     ensName,
			},
			ImagePath: imagePath,
		})
	}

	if text != "" {
		messages = append(messages, &common.Message{
			CommunityID: string(communityID.Bytes()),
			ChatMessage: &protobuf.ChatMessage{
				ChatId:      chatID,
				ContentType: isTextOrEmoji(text),
				Text:        text,
				ResponseTo:  responseTo,
				EnsName:     ensName,
			},
		})
	}

	response, err := api.s.messenger.SendChatMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	return api.toSendMessageResponse(response)

}

func (api *API) SendAudio(ctx context.Context, communityID types.HexBytes, chatID string, audioPath string, responseTo string) (*SendMessageResponse, error) {
	ensName, _ := api.s.accountsDB.GetPreferredUsername()

	msg := &common.Message{
		CommunityID: string(communityID.Bytes()),
		ChatMessage: &protobuf.ChatMessage{
			ChatId:      chatID,
			Text:        "Update to latest version to listen to an audio message here!",
			ContentType: protobuf.ChatMessage_AUDIO,
			ResponseTo:  responseTo,
			EnsName:     ensName,
		},
		AudioPath: audioPath,
	}

	response, err := api.s.messenger.SendChatMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	return api.toSendMessageResponse(response)
}
