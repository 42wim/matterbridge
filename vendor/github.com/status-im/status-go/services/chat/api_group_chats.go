package chat

import (
	"context"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/requests"
)

type GroupChatResponse struct {
	Chat     *Chat             `json:"chat"`
	Messages []*common.Message `json:"messages"`
}

type GroupChatResponseWithInvitations struct {
	Chat        *Chat                           `json:"chat"`
	Messages    []*common.Message               `json:"messages"`
	Invitations []*protocol.GroupChatInvitation `json:"invitations"`
}

type CreateOneToOneChatResponse struct {
	Chat    *Chat             `json:"chat,omitempty"`
	Contact *protocol.Contact `json:"contact,omitempty"`
}

type StartGroupChatResponse struct {
	Chat     *Chat               `json:"chat,omitempty"`
	Contacts []*protocol.Contact `json:"contacts"`
	Messages []*common.Message   `json:"messages,omitempty"`
}

func (api *API) CreateOneToOneChat(ctx context.Context, communityID types.HexBytes, ID types.HexBytes, ensName string) (*CreateOneToOneChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	pubKey := types.EncodeHex(crypto.FromECDSAPub(api.s.messenger.IdentityPublicKey()))
	response, err := api.s.messenger.CreateOneToOneChat(&requests.CreateOneToOneChat{ID: ID, ENSName: ensName})
	if err != nil {
		return nil, err
	}

	chat, err := api.toAPIChat(response.Chats()[0], nil, pubKey, false)
	if err != nil {
		return nil, err
	}

	var contact *protocol.Contact
	if ensName != "" {
		contact = response.Contacts[0]
	}

	return &CreateOneToOneChatResponse{
		Chat:    chat,
		Contact: contact,
	}, nil
}

func (api *API) CreateGroupChat(ctx context.Context, communityID types.HexBytes, name string, members []string) (*GroupChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponse(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.CreateGroupChatWithMembers(ctx, name, members)
	})
}

func (api *API) CreateGroupChatFromInvitation(communityID types.HexBytes, name string, chatID string, adminPK string) (*GroupChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponse(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.CreateGroupChatFromInvitation(name, chatID, adminPK)
	})
}

func (api *API) LeaveChat(ctx context.Context, communityID types.HexBytes, chatID string, remove bool) (*GroupChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponse(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.LeaveGroupChat(ctx, chatID, remove)
	})
}

func (api *API) AddMembers(ctx context.Context, communityID types.HexBytes, chatID string, members []string) (*GroupChatResponseWithInvitations, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponseWithInvitations(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.AddMembersToGroupChat(ctx, chatID, members)
	})
}

func (api *API) RemoveMember(ctx context.Context, communityID types.HexBytes, chatID string, member string) (*GroupChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponse(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.RemoveMembersFromGroupChat(ctx, chatID, []string{member})
	})
}

func (api *API) MakeAdmin(ctx context.Context, communityID types.HexBytes, chatID string, member string) (*GroupChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponse(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.AddAdminsToGroupChat(ctx, chatID, []string{member})
	})
}

func (api *API) RenameChat(ctx context.Context, communityID types.HexBytes, chatID string, name string) (*GroupChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponse(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.ChangeGroupChatName(ctx, chatID, name)
	})
}

func (api *API) SendGroupChatInvitationRequest(ctx context.Context, communityID types.HexBytes, chatID string, adminPK string, message string) (*GroupChatResponseWithInvitations, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	return api.execAndGetGroupChatResponseWithInvitations(func() (*protocol.MessengerResponse, error) {
		return api.s.messenger.SendGroupChatInvitationRequest(ctx, chatID, adminPK, message)
	})
}

func (api *API) GetGroupChatInvitations() ([]*protocol.GroupChatInvitation, error) {
	return api.s.messenger.GetGroupChatInvitations()
}

func (api *API) SendGroupChatInvitationRejection(ctx context.Context, invitationRequestID string) ([]*protocol.GroupChatInvitation, error) {
	response, err := api.s.messenger.SendGroupChatInvitationRejection(ctx, invitationRequestID)
	if err != nil {
		return nil, err
	}
	return response.Invitations, nil
}

func (api *API) StartGroupChat(ctx context.Context, communityID types.HexBytes, name string, members []string) (*StartGroupChatResponse, error) {
	if len(communityID) != 0 {
		return nil, ErrCommunitiesNotSupported
	}

	pubKey := types.EncodeHex(crypto.FromECDSAPub(api.s.messenger.IdentityPublicKey()))

	var response *protocol.MessengerResponse
	var err error
	if len(members) == 1 {
		memberPk, err := common.HexToPubkey(members[0])
		if err != nil {
			return nil, err
		}
		response, err = api.s.messenger.CreateOneToOneChat(&requests.CreateOneToOneChat{
			ID: types.HexBytes(crypto.FromECDSAPub(memberPk)),
		})
		if err != nil {
			return nil, err
		}
	} else {
		response, err = api.s.messenger.CreateGroupChatWithMembers(ctx, name, members)
		if err != nil {
			return nil, err
		}
	}

	chat, err := api.toAPIChat(response.Chats()[0], nil, pubKey, false)
	if err != nil {
		return nil, err
	}

	return &StartGroupChatResponse{
		Chat:     chat,
		Contacts: response.Contacts,
		Messages: response.Messages(),
	}, nil
}

func (api *API) toGroupChatResponse(pubKey string, response *protocol.MessengerResponse) (*GroupChatResponse, error) {
	chat, err := api.toAPIChat(response.Chats()[0], nil, pubKey, false)
	if err != nil {
		return nil, err
	}

	return &GroupChatResponse{
		Chat:     chat,
		Messages: response.Messages(),
	}, nil
}

func (api *API) toGroupChatResponseWithInvitations(pubKey string, response *protocol.MessengerResponse) (*GroupChatResponseWithInvitations, error) {
	g, err := api.toGroupChatResponse(pubKey, response)
	if err != nil {
		return nil, err
	}

	return &GroupChatResponseWithInvitations{
		Chat:        g.Chat,
		Messages:    g.Messages,
		Invitations: response.Invitations,
	}, nil
}

func (api *API) execAndGetGroupChatResponse(fn func() (*protocol.MessengerResponse, error)) (*GroupChatResponse, error) {
	pubKey := types.EncodeHex(crypto.FromECDSAPub(api.s.messenger.IdentityPublicKey()))
	response, err := fn()
	if err != nil {
		return nil, err
	}
	return api.toGroupChatResponse(pubKey, response)
}

func (api *API) execAndGetGroupChatResponseWithInvitations(fn func() (*protocol.MessengerResponse, error)) (*GroupChatResponseWithInvitations, error) {
	pubKey := types.EncodeHex(crypto.FromECDSAPub(api.s.messenger.IdentityPublicKey()))

	response, err := fn()
	if err != nil {
		return nil, err
	}

	return api.toGroupChatResponseWithInvitations(pubKey, response)
}
