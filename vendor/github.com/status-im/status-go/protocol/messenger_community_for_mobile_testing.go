package protocol

import (
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
)

func (m *Messenger) createCommunityChat(communityID types.HexBytes, name string) (*MessengerResponse, error) {
	return m.CreateCommunityChat(communityID, &protobuf.CommunityChat{
		Permissions: &protobuf.CommunityPermissions{
			Access: protobuf.CommunityPermissions_AUTO_ACCEPT,
		},
		Identity: &protobuf.ChatIdentity{
			DisplayName: name,
			Description: name,
		},
	})
}

func (m *Messenger) CreateClosedCommunity() (*MessengerResponse, error) {
	response, err := m.CreateCommunity(&requests.CreateCommunity{
		Name:                         "closed community",
		Description:                  "closed community to check membership requests",
		Color:                        "#887af9",
		HistoryArchiveSupportEnabled: true,
		Membership:                   protobuf.CommunityPermissions_MANUAL_ACCEPT,
		PinMessageAllMembersEnabled:  true,
	}, true)
	if err != nil {
		return nil, err
	}
	community := response.Communities()[0]
	cid := community.ID()

	var (
		catsChannelID  string
		dogsChannelID  string
		rulesChannelID string
	)
	response2, err := m.createCommunityChat(cid, "cats")
	if err != nil {
		return nil, err
	}
	catsChannelID = response2.Chats()[0].CommunityChatID()
	if err = response.Merge(response2); err != nil {
		return nil, err
	}

	response2, err = m.createCommunityChat(cid, "dogs")
	if err != nil {
		return nil, err
	}
	dogsChannelID = response2.Chats()[0].CommunityChatID()
	if err = response.Merge(response2); err != nil {
		return nil, err
	}
	response2, err = m.createCommunityChat(cid, "rules")
	if err != nil {
		return nil, err
	}
	rulesChannelID = response2.Chats()[0].CommunityChatID()
	if err = response.Merge(response2); err != nil {
		return nil, err
	}

	response2, err = m.CreateCommunityCategory(&requests.CreateCommunityCategory{
		CommunityID:  cid,
		CategoryName: "pets",
		ChatIDs:      []string{catsChannelID, dogsChannelID},
	})
	if err != nil {
		return nil, err
	}
	if err = response.Merge(response2); err != nil {
		return nil, err
	}
	response2, err = m.CreateCommunityCategory(&requests.CreateCommunityCategory{
		CommunityID:  cid,
		CategoryName: "household",
		ChatIDs:      []string{rulesChannelID},
	})
	if err != nil {
		return nil, err
	}
	if err = response.Merge(response2); err != nil {
		return nil, err
	}
	return response, nil
}

func (m *Messenger) CreateOpenCommunity() (*MessengerResponse, error) {
	response, err := m.CreateCommunity(&requests.CreateCommunity{
		Name:                         "open community",
		Description:                  "open community to join with no requests",
		Color:                        "#26a69a",
		HistoryArchiveSupportEnabled: true,
		Membership:                   protobuf.CommunityPermissions_AUTO_ACCEPT,
		PinMessageAllMembersEnabled:  false,
	}, true)
	return response, err
}

func (m *Messenger) CreateTokenGatedCommunity() (*MessengerResponse, error) {
	response, err := m.CreateCommunity(&requests.CreateCommunity{
		Name:                         "SNT community",
		Description:                  "require 10 SNT Goerli to use",
		Color:                        "#eab700",
		HistoryArchiveSupportEnabled: true,
		Membership:                   protobuf.CommunityPermissions_MANUAL_ACCEPT,
		PinMessageAllMembersEnabled:  false,
	}, true)
	if err != nil {
		return nil, err
	}
	community := response.Communities()[0]
	cid := community.ID()
	generalChatID := response.Chats()[0].CommunityChatID()

	return m.CreateCommunityTokenPermission(&requests.CreateCommunityTokenPermission{
		CommunityID: cid,
		Type:        protobuf.CommunityTokenPermission_BECOME_MEMBER,
		TokenCriteria: []*protobuf.TokenCriteria{{
			ContractAddresses: map[uint64]string{params.GoerliNetworkID: "0x3D6AFAA395C31FCd391fE3D562E75fe9E8ec7E6a"},
			Type:              protobuf.CommunityTokenType_ERC20,
			Symbol:            "STT",
			Name:              "Status Test Token",
			AmountInWei:       "10000000000000000000",
			Decimals:          18,
		}},
		ChatIds: []string{generalChatID},
	})
}
