package protocol

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/status-go/api/multiformat"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/protocol/urls"
	"github.com/status-im/status-go/services/utils"
)

type CommunityURLData struct {
	DisplayName  string   `json:"displayName"`
	Description  string   `json:"description"`
	MembersCount uint32   `json:"membersCount"`
	Color        string   `json:"color"`
	TagIndices   []uint32 `json:"tagIndices"`
	CommunityID  string   `json:"communityId"`
}

type CommunityChannelURLData struct {
	Emoji       string `json:"emoji"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Color       string `json:"color"`
	ChannelUUID string `json:"channelUuid"`
}

type ContactURLData struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	PublicKey   string `json:"publicKey"`
}

type URLDataResponse struct {
	Community *CommunityURLData        `json:"community"`
	Channel   *CommunityChannelURLData `json:"channel"`
	Contact   *ContactURLData          `json:"contact"`
	Shard     *shard.Shard             `json:"shard,omitempty"`
}

const baseShareURL = "https://status.app"
const userPath = "u#"
const userWithDataPath = "u/"
const communityPath = "c#"
const communityWithDataPath = "c/"
const channelPath = "cc/"

const sharedURLUserPrefix = baseShareURL + "/" + userPath
const sharedURLUserPrefixWithData = baseShareURL + "/" + userWithDataPath
const sharedURLCommunityPrefix = baseShareURL + "/" + communityPath
const sharedURLCommunityPrefixWithData = baseShareURL + "/" + communityWithDataPath
const sharedURLChannelPrefixWithData = baseShareURL + "/" + channelPath

const channelUUIDRegExp = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-5][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12}$"

var channelRegExp = regexp.MustCompile(channelUUIDRegExp)

func decodeCommunityID(serialisedPublicKey string) (string, error) {
	deserializedCommunityID, err := multiformat.DeserializeCompressedKey(serialisedPublicKey)
	if err != nil {
		return "", err
	}

	communityID, err := common.HexToPubkey(deserializedCommunityID)
	if err != nil {
		return "", err
	}

	return types.EncodeHex(crypto.CompressPubkey(communityID)), nil
}

func serializePublicKey(compressedKey types.HexBytes) (string, error) {
	return utils.SerializePublicKey(compressedKey)
}

func deserializePublicKey(compressedKey string) (types.HexBytes, error) {
	return utils.DeserializePublicKey(compressedKey)
}

func (m *Messenger) ShareCommunityURLWithChatKey(communityID types.HexBytes) (string, error) {
	shortKey, err := serializePublicKey(communityID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/c#%s", baseShareURL, shortKey), nil
}

func parseCommunityURLWithChatKey(urlData string) (*URLDataResponse, error) {
	communityID, err := decodeCommunityID(urlData)
	if err != nil {
		return nil, err
	}

	return &URLDataResponse{
		Community: &CommunityURLData{
			CommunityID: communityID,
			TagIndices:  []uint32{},
		},
		Shard: nil,
	}, nil
}

func (m *Messenger) prepareEncodedCommunityData(community *communities.Community) (string, string, error) {
	communityProto := &protobuf.Community{
		DisplayName:  community.Identity().DisplayName,
		Description:  community.DescriptionText(),
		MembersCount: uint32(community.MembersCount()),
		Color:        community.Identity().GetColor(),
		TagIndices:   community.TagsIndices(),
	}

	communityData, err := proto.Marshal(communityProto)
	if err != nil {
		return "", "", err
	}

	urlDataProto := &protobuf.URLData{
		Content: communityData,
		Shard:   community.Shard().Protobuffer(),
	}

	urlData, err := proto.Marshal(urlDataProto)
	if err != nil {
		return "", "", err
	}

	shortKey, err := serializePublicKey(community.ID())
	if err != nil {
		return "", "", err
	}

	encodedData, err := urls.EncodeDataURL(urlData)
	if err != nil {
		return "", "", err
	}

	return encodedData, shortKey, nil
}

func (m *Messenger) ShareCommunityURLWithData(communityID types.HexBytes) (string, error) {
	community, err := m.GetCommunityByID(communityID)
	if err != nil {
		return "", err
	}

	data, shortKey, err := m.prepareEncodedCommunityData(community)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/c/%s#%s", baseShareURL, data, shortKey), nil
}

func parseCommunityURLWithData(data string, chatKey string) (*URLDataResponse, error) {
	communityID, err := deserializePublicKey(chatKey)
	if err != nil {
		return nil, err
	}

	urlData, err := urls.DecodeDataURL(data)
	if err != nil {
		return nil, err
	}

	var urlDataProto protobuf.URLData
	err = proto.Unmarshal(urlData, &urlDataProto)
	if err != nil {
		return nil, err
	}

	var communityProto protobuf.Community
	err = proto.Unmarshal(urlDataProto.Content, &communityProto)
	if err != nil {
		return nil, err
	}

	var tagIndices []uint32
	if communityProto.TagIndices != nil {
		tagIndices = communityProto.TagIndices
	} else {
		tagIndices = []uint32{}
	}

	return &URLDataResponse{
		Community: &CommunityURLData{
			DisplayName:  communityProto.DisplayName,
			Description:  communityProto.Description,
			MembersCount: communityProto.MembersCount,
			Color:        communityProto.Color,
			TagIndices:   tagIndices,
			CommunityID:  types.EncodeHex(communityID),
		},
		Shard: shard.FromProtobuff(urlDataProto.Shard),
	}, nil
}

func (m *Messenger) ShareCommunityChannelURLWithChatKey(request *requests.CommunityChannelShareURL) (string, error) {
	if err := request.Validate(); err != nil {
		return "", err
	}

	shortKey, err := serializePublicKey(request.CommunityID)
	if err != nil {
		return "", err
	}

	valid, err := regexp.MatchString(channelUUIDRegExp, request.ChannelID)
	if err != nil {
		return "", err
	}

	if !valid {
		return "", fmt.Errorf("channelID should be UUID, got %s", request.ChannelID)
	}

	return fmt.Sprintf("%s/cc/%s#%s", baseShareURL, request.ChannelID, shortKey), nil
}

func parseCommunityChannelURLWithChatKey(channelID string, publicKey string) (*URLDataResponse, error) {
	valid, err := regexp.MatchString(channelUUIDRegExp, channelID)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("channelID should be UUID, got %s", channelID)
	}

	communityID, err := decodeCommunityID(publicKey)
	if err != nil {
		return nil, err
	}

	return &URLDataResponse{
		Community: &CommunityURLData{
			CommunityID: communityID,
			TagIndices:  []uint32{},
		},
		Channel: &CommunityChannelURLData{
			ChannelUUID: channelID,
		},
		Shard: nil,
	}, nil
}

func (m *Messenger) prepareEncodedCommunityChannelData(community *communities.Community, channel *protobuf.CommunityChat, channelID string) (string, string, error) {
	communityProto := &protobuf.Community{
		DisplayName:  community.Identity().DisplayName,
		Description:  community.DescriptionText(),
		MembersCount: uint32(community.MembersCount()),
		Color:        community.Identity().GetColor(),
		TagIndices:   community.TagsIndices(),
	}

	channelProto := &protobuf.Channel{
		DisplayName: channel.Identity.DisplayName,
		Description: channel.Identity.Description,
		Emoji:       channel.Identity.Emoji,
		Color:       channel.GetIdentity().Color,
		Community:   communityProto,
		Uuid:        channelID,
	}

	channelData, err := proto.Marshal(channelProto)
	if err != nil {
		return "", "", err
	}

	urlDataProto := &protobuf.URLData{
		Content: channelData,
		Shard:   community.Shard().Protobuffer(),
	}

	urlData, err := proto.Marshal(urlDataProto)
	if err != nil {
		return "", "", err
	}

	shortKey, err := serializePublicKey(community.ID())
	if err != nil {
		return "", "", err
	}
	encodedData, err := urls.EncodeDataURL(urlData)
	if err != nil {
		return "", "", err
	}

	return encodedData, shortKey, nil
}

func (m *Messenger) ShareCommunityChannelURLWithData(request *requests.CommunityChannelShareURL) (string, error) {
	if err := request.Validate(); err != nil {
		return "", err
	}

	valid, err := regexp.MatchString(channelUUIDRegExp, request.ChannelID)
	if err != nil {
		return "", err
	}

	if !valid {
		return "nil", fmt.Errorf("channelID should be UUID, got %s", request.ChannelID)
	}

	community, err := m.GetCommunityByID(request.CommunityID)
	if err != nil {
		return "", err
	}

	channel := community.Chats()[request.ChannelID]
	if channel == nil {
		return "", fmt.Errorf("channel with channelID %s not found", request.ChannelID)
	}

	data, shortKey, err := m.prepareEncodedCommunityChannelData(community, channel, request.ChannelID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/cc/%s#%s", baseShareURL, data, shortKey), nil
}

func parseCommunityChannelURLWithData(data string, chatKey string) (*URLDataResponse, error) {
	communityID, err := deserializePublicKey(chatKey)
	if err != nil {
		return nil, err
	}

	urlData, err := urls.DecodeDataURL(data)
	if err != nil {
		return nil, err
	}

	var urlDataProto protobuf.URLData
	err = proto.Unmarshal(urlData, &urlDataProto)
	if err != nil {
		return nil, err
	}

	var channelProto protobuf.Channel
	err = proto.Unmarshal(urlDataProto.Content, &channelProto)
	if err != nil {
		return nil, err
	}

	var tagIndices []uint32
	if channelProto.Community.TagIndices != nil {
		tagIndices = channelProto.Community.TagIndices
	} else {
		tagIndices = []uint32{}
	}

	return &URLDataResponse{
		Community: &CommunityURLData{
			DisplayName:  channelProto.Community.DisplayName,
			Description:  channelProto.Community.Description,
			MembersCount: channelProto.Community.MembersCount,
			Color:        channelProto.Community.Color,
			TagIndices:   tagIndices,
			CommunityID:  types.EncodeHex(communityID),
		},
		Channel: &CommunityChannelURLData{
			Emoji:       channelProto.Emoji,
			DisplayName: channelProto.DisplayName,
			Description: channelProto.Description,
			Color:       channelProto.Color,
			ChannelUUID: channelProto.Uuid,
		},
		Shard: shard.FromProtobuff(urlDataProto.Shard),
	}, nil
}

func (m *Messenger) ShareUserURLWithChatKey(contactID string) (string, error) {
	publicKey, err := common.HexToPubkey(contactID)
	if err != nil {
		return "", err
	}

	shortKey, err := serializePublicKey(crypto.CompressPubkey(publicKey))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/u#%s", baseShareURL, shortKey), nil
}

func parseUserURLWithChatKey(urlData string) (*URLDataResponse, error) {
	pubKeyBytes, err := deserializePublicKey(urlData)
	if err != nil {
		return nil, err
	}

	pubKey, err := crypto.DecompressPubkey(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	serializedPublicKey, err := multiformat.SerializeLegacyKey(common.PubkeyToHex(pubKey))
	if err != nil {
		return nil, err
	}

	return &URLDataResponse{
		Contact: &ContactURLData{
			PublicKey: serializedPublicKey,
		},
	}, nil
}

func (m *Messenger) ShareUserURLWithENS(contactID string) (string, error) {
	contact := m.GetContactByID(contactID)
	if contact == nil {
		return "", ErrContactNotFound
	}
	return fmt.Sprintf("%s/u#%s", baseShareURL, contact.EnsName), nil
}

func parseUserURLWithENS(ensName string) (*URLDataResponse, error) {
	// TODO: fetch contact by ens name
	return nil, fmt.Errorf("not implemented yet")
}

func (m *Messenger) prepareEncodedUserData(contact *Contact) (string, string, error) {
	pk, err := contact.PublicKey()
	if err != nil {
		return "", "", err
	}

	shortKey, err := serializePublicKey(crypto.CompressPubkey(pk))
	if err != nil {
		return "", "", err
	}

	userProto := &protobuf.User{
		DisplayName: contact.DisplayName,
		Description: contact.Bio,
	}

	userData, err := proto.Marshal(userProto)
	if err != nil {
		return "", "", err
	}

	urlDataProto := &protobuf.URLData{
		Content: userData,
	}

	urlData, err := proto.Marshal(urlDataProto)
	if err != nil {
		return "", "", err
	}

	encodedData, err := urls.EncodeDataURL(urlData)
	if err != nil {
		return "", "", err
	}

	return encodedData, shortKey, nil
}

func (m *Messenger) ShareUserURLWithData(contactID string) (string, error) {
	contact := m.GetContactByID(contactID)
	if contact == nil {
		return "", ErrContactNotFound
	}

	data, shortKey, err := m.prepareEncodedUserData(contact)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/u/%s#%s", baseShareURL, data, shortKey), nil
}

func parseUserURLWithData(data string, chatKey string) (*URLDataResponse, error) {
	urlData, err := urls.DecodeDataURL(data)
	if err != nil {
		return nil, err
	}

	var urlDataProto protobuf.URLData
	err = proto.Unmarshal(urlData, &urlDataProto)
	if err != nil {
		return nil, err
	}

	var userProto protobuf.User
	err = proto.Unmarshal(urlDataProto.Content, &userProto)
	if err != nil {
		return nil, err
	}

	return &URLDataResponse{
		Contact: &ContactURLData{
			DisplayName: userProto.DisplayName,
			Description: userProto.Description,
			PublicKey:   chatKey,
		},
	}, nil
}

func IsStatusSharedURL(url string) bool {
	return strings.HasPrefix(url, sharedURLUserPrefix) ||
		strings.HasPrefix(url, sharedURLUserPrefixWithData) ||
		strings.HasPrefix(url, sharedURLCommunityPrefix) ||
		strings.HasPrefix(url, sharedURLCommunityPrefixWithData) ||
		strings.HasPrefix(url, sharedURLChannelPrefixWithData)
}

func splitSharedURLData(data string) (string, string, error) {
	const count = 2
	contents := strings.SplitN(data, "#", count)
	if len(contents) != count {
		return "", "", fmt.Errorf("url should contain at least one `#` separator")
	}
	return contents[0], contents[1], nil
}

func ParseSharedURL(url string) (*URLDataResponse, error) {

	if strings.HasPrefix(url, sharedURLUserPrefix) {
		chatKey := strings.TrimPrefix(url, sharedURLUserPrefix)
		if strings.HasPrefix(chatKey, "zQ3sh") {
			return parseUserURLWithChatKey(chatKey)
		}
		return parseUserURLWithENS(chatKey)
	}

	if strings.HasPrefix(url, sharedURLUserPrefixWithData) {
		trimmedURL := strings.TrimPrefix(url, sharedURLUserPrefixWithData)
		encodedData, chatKey, err := splitSharedURLData(trimmedURL)
		if err != nil {
			return nil, err
		}
		return parseUserURLWithData(encodedData, chatKey)
	}

	if strings.HasPrefix(url, sharedURLCommunityPrefix) {
		chatKey := strings.TrimPrefix(url, sharedURLCommunityPrefix)
		return parseCommunityURLWithChatKey(chatKey)
	}

	if strings.HasPrefix(url, sharedURLCommunityPrefixWithData) {
		trimmedURL := strings.TrimPrefix(url, sharedURLCommunityPrefixWithData)
		encodedData, chatKey, err := splitSharedURLData(trimmedURL)
		if err != nil {
			return nil, err
		}
		return parseCommunityURLWithData(encodedData, chatKey)
	}

	if strings.HasPrefix(url, sharedURLChannelPrefixWithData) {
		trimmedURL := strings.TrimPrefix(url, sharedURLChannelPrefixWithData)
		encodedData, chatKey, err := splitSharedURLData(trimmedURL)
		if err != nil {
			return nil, err
		}

		if channelRegExp.MatchString(encodedData) {
			return parseCommunityChannelURLWithChatKey(encodedData, chatKey)
		}
		return parseCommunityChannelURLWithData(encodedData, chatKey)
	}

	return nil, fmt.Errorf("not a status shared url")
}
