package communities

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/status-im/status-go/api/multiformat"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	community_token "github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/protocol/v1"
)

const signatureLength = 65

type Config struct {
	PrivateKey                          *ecdsa.PrivateKey
	ControlNode                         *ecdsa.PublicKey
	ControlDevice                       bool // whether this device is control node
	CommunityDescription                *protobuf.CommunityDescription
	CommunityDescriptionProtocolMessage []byte // community in a wrapped & signed (by owner) protocol message
	ID                                  *ecdsa.PublicKey
	Joined                              bool
	JoinedAt                            int64
	Requested                           bool
	Verified                            bool
	Spectated                           bool
	Muted                               bool
	MuteTill                            time.Time
	Logger                              *zap.Logger
	RequestedToJoinAt                   uint64
	RequestsToJoin                      []*RequestToJoin
	MemberIdentity                      *ecdsa.PublicKey
	EventsData                          *EventsData
	Shard                               *shard.Shard
	PubsubTopicPrivateKey               *ecdsa.PrivateKey
	LastOpenedAt                        int64
}

type EventsData struct {
	EventsBaseCommunityDescription []byte
	Events                         []CommunityEvent
}

type Community struct {
	config     *Config
	mutex      sync.Mutex
	timesource common.TimeSource
	encryptor  DescriptionEncryptor
}

func New(config Config, timesource common.TimeSource, encryptor DescriptionEncryptor) (*Community, error) {
	if config.MemberIdentity == nil {
		return nil, errors.New("no member identity")
	}

	if timesource == nil {
		return nil, errors.New("no timesource")
	}

	if config.Logger == nil {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
		config.Logger = logger
	}

	if config.CommunityDescription == nil {
		config.CommunityDescription = &protobuf.CommunityDescription{}
	}

	return &Community{config: &config, timesource: timesource, encryptor: encryptor}, nil
}

type CommunityAdminSettings struct {
	PinMessageAllMembersEnabled bool `json:"pinMessageAllMembersEnabled"`
}

type CommunityChat struct {
	ID          string                               `json:"id"`
	Name        string                               `json:"name"`
	Color       string                               `json:"color"`
	Emoji       string                               `json:"emoji"`
	Description string                               `json:"description"`
	Members     map[string]*protobuf.CommunityMember `json:"members"`
	Permissions *protobuf.CommunityPermissions       `json:"permissions"`
	CanPost     bool                                 `json:"canPost"`
	Position    int                                  `json:"position"`
	CategoryID  string                               `json:"categoryID"`
}

type CommunityCategory struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"` // Position is used to sort the categories
}

type CommunityTag struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
}

type CommunityMemberState uint8

const (
	CommunityMemberBanned CommunityMemberState = iota
	CommunityMemberBanPending
	CommunityMemberUnbanPending
	CommunityMemberKickPending
)

func (o *Community) MarshalPublicAPIJSON() ([]byte, error) {
	if o.config.MemberIdentity == nil {
		return nil, errors.New("member identity not set")
	}
	communityItem := struct {
		ID                      types.HexBytes                       `json:"id"`
		Verified                bool                                 `json:"verified"`
		Chats                   map[string]CommunityChat             `json:"chats"`
		Categories              map[string]CommunityCategory         `json:"categories"`
		Name                    string                               `json:"name"`
		Description             string                               `json:"description"`
		IntroMessage            string                               `json:"introMessage"`
		OutroMessage            string                               `json:"outroMessage"`
		Tags                    []CommunityTag                       `json:"tags"`
		Images                  map[string]images.IdentityImage      `json:"images"`
		Color                   string                               `json:"color"`
		MembersCount            int                                  `json:"membersCount"`
		EnsName                 string                               `json:"ensName"`
		Link                    string                               `json:"link"`
		CommunityAdminSettings  CommunityAdminSettings               `json:"adminSettings"`
		Encrypted               bool                                 `json:"encrypted"`
		TokenPermissions        map[string]*CommunityTokenPermission `json:"tokenPermissions"`
		CommunityTokensMetadata []*protobuf.CommunityTokenMetadata   `json:"communityTokensMetadata"`
		ActiveMembersCount      uint64                               `json:"activeMembersCount"`
		PubsubTopic             string                               `json:"pubsubTopic"`
		PubsubTopicKey          string                               `json:"pubsubTopicKey"`
		Shard                   *shard.Shard                         `json:"shard"`
	}{
		ID:             o.ID(),
		Verified:       o.config.Verified,
		Chats:          make(map[string]CommunityChat),
		Categories:     make(map[string]CommunityCategory),
		Tags:           o.Tags(),
		PubsubTopic:    o.PubsubTopic(),
		PubsubTopicKey: o.PubsubTopicKey(),
		Shard:          o.Shard(),
	}

	if o.config.CommunityDescription != nil {
		for id, c := range o.config.CommunityDescription.Categories {
			category := CommunityCategory{
				ID:       id,
				Name:     c.Name,
				Position: int(c.Position),
			}
			communityItem.Categories[id] = category
			communityItem.Encrypted = o.Encrypted()
		}
		for id, c := range o.config.CommunityDescription.Chats {
			canPost, err := o.CanPost(o.config.MemberIdentity, id)
			if err != nil {
				return nil, err
			}
			chat := CommunityChat{
				ID:          id,
				Name:        c.Identity.DisplayName,
				Color:       c.Identity.Color,
				Emoji:       c.Identity.Emoji,
				Description: c.Identity.Description,
				Permissions: c.Permissions,
				Members:     c.Members,
				CanPost:     canPost,
				CategoryID:  c.CategoryId,
				Position:    int(c.Position),
			}
			communityItem.Chats[id] = chat
		}

		communityItem.TokenPermissions = o.tokenPermissions()
		communityItem.MembersCount = len(o.config.CommunityDescription.Members)

		communityItem.Link = fmt.Sprintf("https://join.status.im/c/0x%x", o.ID())
		if o.Shard() != nil {
			communityItem.Link = fmt.Sprintf("%s/%d/%d", communityItem.Link, o.Shard().Cluster, o.Shard().Index)
		}

		communityItem.IntroMessage = o.config.CommunityDescription.IntroMessage
		communityItem.OutroMessage = o.config.CommunityDescription.OutroMessage
		communityItem.CommunityTokensMetadata = o.config.CommunityDescription.CommunityTokensMetadata
		communityItem.ActiveMembersCount = o.config.CommunityDescription.ActiveMembersCount

		if o.config.CommunityDescription.Identity != nil {
			communityItem.Name = o.Name()
			communityItem.Color = o.config.CommunityDescription.Identity.Color
			communityItem.Description = o.config.CommunityDescription.Identity.Description
			for t, i := range o.config.CommunityDescription.Identity.Images {
				if communityItem.Images == nil {
					communityItem.Images = make(map[string]images.IdentityImage)
				}
				communityItem.Images[t] = images.IdentityImage{Name: t, Payload: i.Payload}

			}
		}

		communityItem.CommunityAdminSettings = CommunityAdminSettings{
			PinMessageAllMembersEnabled: false,
		}

		if o.config.CommunityDescription.AdminSettings != nil {
			communityItem.CommunityAdminSettings.PinMessageAllMembersEnabled = o.config.CommunityDescription.AdminSettings.PinMessageAllMembersEnabled
		}
	}
	return json.Marshal(communityItem)
}

func (o *Community) MarshalJSON() ([]byte, error) {
	if o.config.MemberIdentity == nil {
		return nil, errors.New("member identity not set")
	}
	communityItem := struct {
		ID                          types.HexBytes                       `json:"id"`
		MemberRole                  protobuf.CommunityMember_Roles       `json:"memberRole"`
		IsControlNode               bool                                 `json:"isControlNode"`
		Verified                    bool                                 `json:"verified"`
		Joined                      bool                                 `json:"joined"`
		JoinedAt                    int64                                `json:"joinedAt"`
		Spectated                   bool                                 `json:"spectated"`
		RequestedAccessAt           int                                  `json:"requestedAccessAt"`
		Name                        string                               `json:"name"`
		Description                 string                               `json:"description"`
		IntroMessage                string                               `json:"introMessage"`
		OutroMessage                string                               `json:"outroMessage"`
		Tags                        []CommunityTag                       `json:"tags"`
		Chats                       map[string]CommunityChat             `json:"chats"`
		Categories                  map[string]CommunityCategory         `json:"categories"`
		Images                      map[string]images.IdentityImage      `json:"images"`
		Permissions                 *protobuf.CommunityPermissions       `json:"permissions"`
		Members                     map[string]*protobuf.CommunityMember `json:"members"`
		CanRequestAccess            bool                                 `json:"canRequestAccess"`
		CanManageUsers              bool                                 `json:"canManageUsers"`              //TODO: we can remove this
		CanDeleteMessageForEveryone bool                                 `json:"canDeleteMessageForEveryone"` //TODO: we can remove this
		CanJoin                     bool                                 `json:"canJoin"`
		Color                       string                               `json:"color"`
		RequestedToJoinAt           uint64                               `json:"requestedToJoinAt,omitempty"`
		IsMember                    bool                                 `json:"isMember"`
		Muted                       bool                                 `json:"muted"`
		MuteTill                    time.Time                            `json:"muteTill,omitempty"`
		CommunityAdminSettings      CommunityAdminSettings               `json:"adminSettings"`
		Encrypted                   bool                                 `json:"encrypted"`
		PendingAndBannedMembers     map[string]CommunityMemberState      `json:"pendingAndBannedMembers"`
		TokenPermissions            map[string]*CommunityTokenPermission `json:"tokenPermissions"`
		CommunityTokensMetadata     []*protobuf.CommunityTokenMetadata   `json:"communityTokensMetadata"`
		ActiveMembersCount          uint64                               `json:"activeMembersCount"`
		PubsubTopic                 string                               `json:"pubsubTopic"`
		PubsubTopicKey              string                               `json:"pubsubTopicKey"`
		Shard                       *shard.Shard                         `json:"shard"`
		LastOpenedAt                int64                                `json:"lastOpenedAt"`
	}{
		ID:                          o.ID(),
		MemberRole:                  o.MemberRole(o.MemberIdentity()),
		IsControlNode:               o.IsControlNode(),
		Verified:                    o.config.Verified,
		Chats:                       make(map[string]CommunityChat),
		Categories:                  make(map[string]CommunityCategory),
		Joined:                      o.config.Joined,
		JoinedAt:                    o.config.JoinedAt,
		Spectated:                   o.config.Spectated,
		CanRequestAccess:            o.CanRequestAccess(o.config.MemberIdentity),
		CanJoin:                     o.canJoin(),
		CanManageUsers:              o.CanManageUsers(o.config.MemberIdentity),
		CanDeleteMessageForEveryone: o.CanDeleteMessageForEveryone(o.config.MemberIdentity),
		RequestedToJoinAt:           o.RequestedToJoinAt(),
		IsMember:                    o.isMember(),
		Muted:                       o.config.Muted,
		MuteTill:                    o.config.MuteTill,
		Tags:                        o.Tags(),
		Encrypted:                   o.Encrypted(),
		PubsubTopic:                 o.PubsubTopic(),
		PubsubTopicKey:              o.PubsubTopicKey(),
		Shard:                       o.Shard(),
		LastOpenedAt:                o.config.LastOpenedAt,
	}
	if o.config.CommunityDescription != nil {
		for id, c := range o.config.CommunityDescription.Categories {
			category := CommunityCategory{
				ID:       id,
				Name:     c.Name,
				Position: int(c.Position),
			}
			communityItem.Encrypted = o.Encrypted()
			communityItem.Categories[id] = category
		}
		for id, c := range o.config.CommunityDescription.Chats {
			canPost, err := o.CanPost(o.config.MemberIdentity, id)
			if err != nil {
				return nil, err
			}
			chat := CommunityChat{
				ID:          id,
				Name:        c.Identity.DisplayName,
				Emoji:       c.Identity.Emoji,
				Color:       c.Identity.Color,
				Description: c.Identity.Description,
				Permissions: c.Permissions,
				Members:     c.Members,
				CanPost:     canPost,
				CategoryID:  c.CategoryId,
				Position:    int(c.Position),
			}
			communityItem.Chats[id] = chat
		}
		communityItem.TokenPermissions = o.tokenPermissions()
		communityItem.PendingAndBannedMembers = o.PendingAndBannedMembers()
		communityItem.Members = o.config.CommunityDescription.Members
		communityItem.Permissions = o.config.CommunityDescription.Permissions
		communityItem.IntroMessage = o.config.CommunityDescription.IntroMessage
		communityItem.OutroMessage = o.config.CommunityDescription.OutroMessage
		communityItem.CommunityTokensMetadata = o.config.CommunityDescription.CommunityTokensMetadata
		communityItem.ActiveMembersCount = o.config.CommunityDescription.ActiveMembersCount

		if o.config.CommunityDescription.Identity != nil {
			communityItem.Name = o.Name()
			communityItem.Color = o.config.CommunityDescription.Identity.Color
			communityItem.Description = o.config.CommunityDescription.Identity.Description
			for t, i := range o.config.CommunityDescription.Identity.Images {
				if communityItem.Images == nil {
					communityItem.Images = make(map[string]images.IdentityImage)
				}
				communityItem.Images[t] = images.IdentityImage{Name: t, Payload: i.Payload}
			}
		}

		communityItem.CommunityAdminSettings = CommunityAdminSettings{
			PinMessageAllMembersEnabled: false,
		}

		if o.config.CommunityDescription.AdminSettings != nil {
			communityItem.CommunityAdminSettings.PinMessageAllMembersEnabled = o.config.CommunityDescription.AdminSettings.PinMessageAllMembersEnabled
		}
	}
	return json.Marshal(communityItem)
}

func (o *Community) Identity() *protobuf.ChatIdentity {
	return o.config.CommunityDescription.Identity
}

func (o *Community) Permissions() *protobuf.CommunityPermissions {
	return o.config.CommunityDescription.Permissions
}

func (o *Community) AdminSettings() *protobuf.CommunityAdminSettings {
	return o.config.CommunityDescription.AdminSettings
}

func (o *Community) Name() string {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil &&
		o.config.CommunityDescription.Identity != nil {
		return o.config.CommunityDescription.Identity.DisplayName
	}
	return ""
}

func (o *Community) DescriptionText() string {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil &&
		o.config.CommunityDescription.Identity != nil {
		return o.config.CommunityDescription.Identity.Description
	}
	return ""
}

func (o *Community) Shard() *shard.Shard {
	if o != nil && o.config != nil {
		return o.config.Shard
	}

	return nil
}

func (o *Community) CommunityShard() CommunityShard {
	return CommunityShard{
		CommunityID: o.IDString(),
		Shard:       o.Shard(),
	}
}

func (o *Community) IntroMessage() string {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil {
		return o.config.CommunityDescription.IntroMessage
	}
	return ""
}

func (o *Community) CommunityTokensMetadata() []*protobuf.CommunityTokenMetadata {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil {
		return o.config.CommunityDescription.CommunityTokensMetadata
	}
	return nil
}

func (o *Community) Tags() []CommunityTag {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil {
		var result []CommunityTag
		for _, t := range o.config.CommunityDescription.Tags {
			result = append(result, CommunityTag{
				Name:  t,
				Emoji: requests.TagsEmojies[t],
			})
		}
		return result
	}
	return nil
}

func (o *Community) TagsRaw() []string {
	return o.config.CommunityDescription.Tags
}

func (o *Community) TagsIndices() []uint32 {
	indices := []uint32{}
	for _, t := range o.config.CommunityDescription.Tags {
		i := uint32(0)
		for k := range requests.TagsEmojies {
			if k == t {
				indices = append(indices, i)
			}
			i++
		}
	}

	return indices
}

func (o *Community) OutroMessage() string {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil {
		return o.config.CommunityDescription.OutroMessage
	}
	return ""
}

func (o *Community) Color() string {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil &&
		o.config.CommunityDescription.Identity != nil {
		return o.config.CommunityDescription.Identity.Color
	}
	return ""
}

func (o *Community) Members() map[string]*protobuf.CommunityMember {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil {
		return o.config.CommunityDescription.Members
	}
	return nil
}

func (o *Community) MembersCount() int {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil {
		return len(o.config.CommunityDescription.Members)
	}
	return 0
}

func (o *Community) GetMemberPubkeys() []*ecdsa.PublicKey {
	if o != nil &&
		o.config != nil &&
		o.config.CommunityDescription != nil {
		pubkeys := make([]*ecdsa.PublicKey, len(o.config.CommunityDescription.Members))
		i := 0
		for hex := range o.config.CommunityDescription.Members {
			pubkeys[i], _ = common.HexToPubkey(hex)
			i++
		}
		return pubkeys
	}
	return nil
}

type CommunitySettings struct {
	CommunityID                  string `json:"communityId"`
	HistoryArchiveSupportEnabled bool   `json:"historyArchiveSupportEnabled"`
	Clock                        uint64 `json:"clock"`
}

// `CommunityAdminEventChanges contain additional changes that don't live on
// a `Community` but still have to be propagated to other admin and control nodes
type CommunityEventChanges struct {
	*CommunityChanges
	// `RejectedRequestsToJoin` is a map of signer keys to requests to join
	RejectedRequestsToJoin map[string]*protobuf.CommunityRequestToJoin `json:"rejectedRequestsToJoin"`
	// `AcceptedRequestsToJoin` is a map of signer keys to requests to join
	AcceptedRequestsToJoin map[string]*protobuf.CommunityRequestToJoin `json:"acceptedRequestsToJoin"`
}

func (o *Community) emptyCommunityChanges() *CommunityChanges {
	changes := EmptyCommunityChanges()
	changes.Community = o
	return changes
}

func (o *Community) CreateChat(chatID string, chat *protobuf.CommunityChat) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CHANNEL_CREATE)) {
		return nil, ErrNotAuthorized
	}

	err := o.createChat(chatID, chat)
	if err != nil {
		return nil, err
	}
	changes := o.emptyCommunityChanges()
	changes.ChatsAdded[chatID] = chat

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToCreateChannelCommunityEvent(chatID, chat))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) EditChat(chatID string, chat *protobuf.CommunityChat) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CHANNEL_EDIT)) {
		return nil, ErrNotAuthorized
	}

	err := o.editChat(chatID, chat)
	if err != nil {
		return nil, err
	}
	changes := o.emptyCommunityChanges()
	changes.ChatsModified[chatID] = &CommunityChatChanges{
		ChatModified: chat,
	}

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToEditChannelCommunityEvent(chatID, chat))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) DeleteChat(chatID string) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_CHANNEL_DELETE)) {
		return nil, ErrNotAuthorized
	}

	changes := o.deleteChat(chatID)

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToDeleteChannelCommunityEvent(chatID))
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (o *Community) getMember(pk *ecdsa.PublicKey) *protobuf.CommunityMember {

	key := common.PubkeyToHex(pk)
	member := o.config.CommunityDescription.Members[key]
	return member
}

func (o *Community) GetMember(pk *ecdsa.PublicKey) *protobuf.CommunityMember {
	return o.getMember(pk)
}

func (o *Community) getChatMember(pk *ecdsa.PublicKey, chatID string) *protobuf.CommunityMember {
	if !o.hasMember(pk) {
		return nil
	}

	chat, ok := o.config.CommunityDescription.Chats[chatID]
	if !ok {
		return nil
	}

	key := common.PubkeyToHex(pk)
	return chat.Members[key]
}

func (o *Community) hasMember(pk *ecdsa.PublicKey) bool {

	member := o.getMember(pk)
	return member != nil
}

func (o *Community) IsBanned(pk *ecdsa.PublicKey) bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.isBanned(pk)
}

func (o *Community) isBanned(pk *ecdsa.PublicKey) bool {
	key := common.PubkeyToHex(pk)

	for _, k := range o.config.CommunityDescription.BanList {
		if k == key {
			return true
		}
	}
	return false
}

func (o *Community) rolesOf(pk *ecdsa.PublicKey) []protobuf.CommunityMember_Roles {
	member := o.getMember(pk)
	if member == nil {
		return nil
	}

	return member.Roles
}

func (o *Community) memberHasRoles(member *protobuf.CommunityMember, roles map[protobuf.CommunityMember_Roles]bool) bool {
	for _, r := range member.Roles {
		if roles[r] {
			return true
		}
	}
	return false
}

func (o *Community) hasRoles(pk *ecdsa.PublicKey, roles map[protobuf.CommunityMember_Roles]bool) bool {
	if pk == nil || o.config == nil || o.config.ID == nil {
		return false
	}

	member := o.getMember(pk)
	if member == nil {
		return false
	}

	return o.memberHasRoles(member, roles)
}

func (o *Community) HasMember(pk *ecdsa.PublicKey) bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.hasMember(pk)
}

func (o *Community) IsMemberInChat(pk *ecdsa.PublicKey, chatID string) bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.getChatMember(pk, chatID) != nil
}

func (o *Community) RemoveUserFromChat(pk *ecdsa.PublicKey, chatID string) (*protobuf.CommunityDescription, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.IsControlNode() {
		return nil, ErrNotControlNode
	}
	if !o.hasMember(pk) {
		return o.config.CommunityDescription, nil
	}

	chat, ok := o.config.CommunityDescription.Chats[chatID]
	if !ok {
		return o.config.CommunityDescription, nil
	}

	key := common.PubkeyToHex(pk)
	delete(chat.Members, key)

	if o.IsControlNode() {
		o.increaseClock()
	}

	return o.config.CommunityDescription, nil
}

func (o *Community) removeMemberFromOrg(pk *ecdsa.PublicKey) {
	if !o.hasMember(pk) {
		return
	}

	key := common.PubkeyToHex(pk)

	// Remove from org
	delete(o.config.CommunityDescription.Members, key)

	// Remove from chats
	for _, chat := range o.config.CommunityDescription.Chats {
		delete(chat.Members, key)
	}
}

func (o *Community) RemoveOurselvesFromOrg(pk *ecdsa.PublicKey) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.removeMemberFromOrg(pk)
	o.increaseClock()
}

func (o *Community) RemoveUserFromOrg(pk *ecdsa.PublicKey) (*protobuf.CommunityDescription, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK)) {
		return nil, ErrNotAuthorized
	}

	if !o.IsControlNode() && o.IsPrivilegedMember(pk) {
		return nil, ErrCannotRemoveOwnerOrAdmin
	}

	if o.IsControlNode() {
		o.removeMemberFromOrg(pk)
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToKickCommunityMemberCommunityEvent(common.PubkeyToHex(pk)))
		if err != nil {
			return nil, err
		}
	}

	return o.config.CommunityDescription, nil
}

func (o *Community) RemoveAllUsersFromOrg() *CommunityChanges {
	o.increaseClock()

	myPublicKey := common.PubkeyToHex(o.config.MemberIdentity)
	member := o.config.CommunityDescription.Members[myPublicKey]

	membersToRemove := o.config.CommunityDescription.Members
	delete(membersToRemove, myPublicKey)

	changes := o.emptyCommunityChanges()
	changes.MembersRemoved = membersToRemove

	o.config.CommunityDescription.Members = make(map[string]*protobuf.CommunityMember)
	o.config.CommunityDescription.Members[myPublicKey] = member

	for chatID, chat := range o.config.CommunityDescription.Chats {
		chatMembersToRemove := chat.Members
		delete(chatMembersToRemove, myPublicKey)

		chat.Members = make(map[string]*protobuf.CommunityMember)
		chat.Members[myPublicKey] = member

		changes.ChatsModified[chatID] = &CommunityChatChanges{
			ChatModified:   chat,
			MembersRemoved: chatMembersToRemove,
		}
	}

	return changes
}

func (o *Community) AddCommunityTokensMetadata(token *protobuf.CommunityTokenMetadata) (*protobuf.CommunityDescription, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_TOKEN_ADD)) {
		return nil, ErrNotAuthorized
	}

	o.config.CommunityDescription.CommunityTokensMetadata = append(o.config.CommunityDescription.CommunityTokensMetadata, token)

	if o.IsControlNode() {
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToAddTokenMetadataCommunityEvent(token))
		if err != nil {
			return nil, err
		}
	}

	return o.config.CommunityDescription, nil
}

func (o *Community) UnbanUserFromCommunity(pk *ecdsa.PublicKey) (*protobuf.CommunityDescription, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN)) {
		return nil, ErrNotAuthorized
	}

	if o.IsControlNode() {
		o.unbanUserFromCommunity(pk)
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToUnbanCommunityMemberCommunityEvent(common.PubkeyToHex(pk)))
		if err != nil {
			return nil, err
		}
	}

	return o.config.CommunityDescription, nil
}

func (o *Community) BanUserFromCommunity(pk *ecdsa.PublicKey) (*protobuf.CommunityDescription, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN)) {
		return nil, ErrNotAuthorized
	}

	if !o.IsControlNode() && o.IsPrivilegedMember(pk) {
		return nil, ErrCannotBanOwnerOrAdmin
	}

	if o.IsControlNode() {
		o.banUserFromCommunity(pk)
		o.increaseClock()
	} else {
		err := o.addNewCommunityEvent(o.ToBanCommunityMemberCommunityEvent(common.PubkeyToHex(pk)))
		if err != nil {
			return nil, err
		}
	}

	return o.config.CommunityDescription, nil
}

func (o *Community) AddRoleToMember(pk *ecdsa.PublicKey, role protobuf.CommunityMember_Roles) (*protobuf.CommunityDescription, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.IsControlNode() {
		return nil, ErrNotControlNode
	}

	updated := false
	addRole := func(member *protobuf.CommunityMember) {
		roles := make(map[protobuf.CommunityMember_Roles]bool)
		roles[role] = true
		if !o.memberHasRoles(member, roles) {
			member.Roles = append(member.Roles, role)
			updated = true
		}
	}

	member := o.getMember(pk)
	if member != nil {
		addRole(member)
	}

	for channelID := range o.chats() {
		chatMember := o.getChatMember(pk, channelID)
		if chatMember != nil {
			addRole(member)
		}
	}

	if updated {
		o.increaseClock()
	}
	return o.config.CommunityDescription, nil
}

func (o *Community) RemoveRoleFromMember(pk *ecdsa.PublicKey, role protobuf.CommunityMember_Roles) (*protobuf.CommunityDescription, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.IsControlNode() {
		return nil, ErrNotControlNode
	}

	updated := false
	removeRole := func(member *protobuf.CommunityMember) {
		roles := make(map[protobuf.CommunityMember_Roles]bool)
		roles[role] = true
		if o.memberHasRoles(member, roles) {
			var newRoles []protobuf.CommunityMember_Roles
			for _, r := range member.Roles {
				if r != role {
					newRoles = append(newRoles, r)
				}
			}
			member.Roles = newRoles
			updated = true
		}
	}

	member := o.getMember(pk)
	if member != nil {
		removeRole(member)
	}

	for channelID := range o.chats() {
		chatMember := o.getChatMember(pk, channelID)
		if chatMember != nil {
			removeRole(member)
		}
	}

	if updated {
		o.increaseClock()
	}
	return o.config.CommunityDescription, nil
}

func (o *Community) Edit(description *protobuf.CommunityDescription) {
	o.config.CommunityDescription.Identity.DisplayName = description.Identity.DisplayName
	o.config.CommunityDescription.Identity.Description = description.Identity.Description
	o.config.CommunityDescription.Identity.Color = description.Identity.Color
	o.config.CommunityDescription.Tags = description.Tags
	o.config.CommunityDescription.Identity.Emoji = description.Identity.Emoji
	o.config.CommunityDescription.Identity.Images = description.Identity.Images
	o.config.CommunityDescription.IntroMessage = description.IntroMessage
	o.config.CommunityDescription.OutroMessage = description.OutroMessage
	if o.config.CommunityDescription.AdminSettings == nil {
		o.config.CommunityDescription.AdminSettings = &protobuf.CommunityAdminSettings{}
	}
	o.config.CommunityDescription.Permissions = description.Permissions
	o.config.CommunityDescription.AdminSettings.PinMessageAllMembersEnabled = description.AdminSettings.PinMessageAllMembersEnabled
}

func (o *Community) Join() {
	o.config.Joined = true
	o.config.JoinedAt = time.Now().Unix()
	o.config.Spectated = false
}

func (o *Community) UpdateLastOpenedAt(timestamp int64) {
	o.config.LastOpenedAt = timestamp
}

func (o *Community) Leave() {
	o.config.Joined = false
	o.config.Spectated = false
}

func (o *Community) Spectate() {
	o.config.Spectated = true
}

func (o *Community) Encrypted() bool {
	return len(o.TokenPermissionsByType(protobuf.CommunityTokenPermission_BECOME_MEMBER)) > 0
}

func (o *Community) Joined() bool {
	return o.config.Joined
}

func (o *Community) JoinedAt() int64 {
	return o.config.JoinedAt
}

func (o *Community) LastOpenedAt() int64 {
	return o.config.LastOpenedAt
}

func (o *Community) Spectated() bool {
	return o.config.Spectated
}

func (o *Community) Verified() bool {
	return o.config.Verified
}

func (o *Community) Muted() bool {
	return o.config.Muted
}

func (o *Community) MuteTill() time.Time {
	return o.config.MuteTill
}

func (o *Community) MemberIdentity() *ecdsa.PublicKey {
	return o.config.MemberIdentity
}

// UpdateCommunityDescription will update the community to the new community description and return a list of changes
func (o *Community) UpdateCommunityDescription(description *protobuf.CommunityDescription, rawMessage []byte, newControlNode *ecdsa.PublicKey) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// This is done in case tags are updated and a client sends unknown tags
	description.Tags = requests.RemoveUnknownAndDeduplicateTags(description.Tags)

	err := ValidateCommunityDescription(description)
	if err != nil {
		return nil, err
	}

	response := o.emptyCommunityChanges()

	if description.Clock <= o.config.CommunityDescription.Clock {
		return response, nil
	}

	originCommunity := o.CreateDeepCopy()

	o.config.CommunityDescription = description
	o.config.CommunityDescriptionProtocolMessage = rawMessage

	if newControlNode != nil {
		o.setControlNode(newControlNode)
	}

	// We only calculate changes if we joined/spectated the community or we requested access, otherwise not interested
	if o.config.Joined || o.config.Spectated || o.config.RequestedToJoinAt > 0 {
		response = EvaluateCommunityChanges(originCommunity, o)
	}

	return response, nil
}

func (o *Community) UpdateChatFirstMessageTimestamp(chatID string, timestamp uint32) (*CommunityChanges, error) {
	if !o.IsControlNode() {
		return nil, ErrNotControlNode
	}

	chat, ok := o.config.CommunityDescription.Chats[chatID]
	if !ok {
		return nil, ErrChatNotFound
	}

	chat.Identity.FirstMessageTimestamp = timestamp

	communityChanges := o.emptyCommunityChanges()
	communityChanges.ChatsModified[chatID] = &CommunityChatChanges{
		FirstMessageTimestampModified: timestamp,
	}
	return communityChanges, nil
}

// ValidateRequestToJoin validates a request, checks that the right permissions are applied
func (o *Community) ValidateRequestToJoin(signer *ecdsa.PublicKey, request *protobuf.CommunityRequestToJoin) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.IsControlNode() {
		// TODO: Enable this once mobile supports revealed addresses.
		// if len(request.RevealedAccounts) == 0 {
		// 	return errors.New("no addresses revealed")
		// }
	} else if o.HasPermissionToSendCommunityEvents() {
		if o.AutoAccept() {
			return errors.New("auto-accept community requests can only be processed by the control node")
		}
	} else {
		return ErrNotAdmin
	}

	if o.config.CommunityDescription.Permissions.EnsOnly && len(request.EnsName) == 0 {
		return ErrCantRequestAccess
	}

	if len(request.ChatId) != 0 {
		return o.validateRequestToJoinWithChatID(request)
	}

	err := o.validateRequestToJoinWithoutChatID(request)
	if err != nil {
		return err
	}

	if o.isBanned(signer) {
		return ErrCantRequestAccess
	}

	timeNow := uint64(time.Now().Unix())
	requestTimeOutClock, err := AddTimeoutToRequestToJoinClock(request.Clock)
	if err != nil {
		return err
	}
	if timeNow >= requestTimeOutClock {
		return errors.New("request is expired")
	}

	return nil
}

// ValidateRequestToJoin validates a request, checks that the right permissions are applied
func (o *Community) ValidateEditSharedAddresses(signer *ecdsa.PublicKey, request *protobuf.CommunityEditSharedAddresses) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// If we are not owner, fuggetaboutit
	if !o.IsControlNode() {
		return ErrNotOwner
	}

	if len(request.RevealedAccounts) == 0 {
		return errors.New("no addresses were shared")
	}

	if request.Clock < o.config.CommunityDescription.Members[common.PubkeyToHex(signer)].LastUpdateClock {
		return errors.New("edit request is older than the last one we have. Ignore")
	}

	return nil
}

// We treat control node as an owner with community key
func (o *Community) IsControlNode() bool {
	return o.config.PrivateKey != nil && o.config.PrivateKey.PublicKey.Equal(o.ControlNode()) && o.config.ControlDevice
}

func (o *Community) IsOwner() bool {
	return o.IsMemberOwner(o.config.MemberIdentity)
}

func (o *Community) IsTokenMaster() bool {
	return o.IsMemberTokenMaster(o.config.MemberIdentity)
}

func (o *Community) IsAdmin() bool {
	return o.IsMemberAdmin(o.config.MemberIdentity)
}

func (o *Community) GetPrivilegedMembers() []*ecdsa.PublicKey {
	privilegedMembers := make([]*ecdsa.PublicKey, 0)
	members := o.GetMemberPubkeys()
	for _, member := range members {
		if o.IsPrivilegedMember(member) {
			privilegedMembers = append(privilegedMembers, member)
		}
	}
	return privilegedMembers
}

func (o *Community) GetFilteredPrivilegedMembers(skipMembers map[string]struct{}) map[protobuf.CommunityMember_Roles][]*ecdsa.PublicKey {
	privilegedMembers := make(map[protobuf.CommunityMember_Roles][]*ecdsa.PublicKey)
	privilegedMembers[protobuf.CommunityMember_ROLE_TOKEN_MASTER] = []*ecdsa.PublicKey{}
	privilegedMembers[protobuf.CommunityMember_ROLE_ADMIN] = []*ecdsa.PublicKey{}
	privilegedMembers[protobuf.CommunityMember_ROLE_OWNER] = []*ecdsa.PublicKey{}

	members := o.GetMemberPubkeys()
	for _, member := range members {
		if len(skipMembers) > 0 {
			if _, exist := skipMembers[common.PubkeyToHex(member)]; exist {
				delete(skipMembers, common.PubkeyToHex(member))
				continue
			}
		}

		memberRole := o.MemberRole(member)
		if memberRole == protobuf.CommunityMember_ROLE_OWNER || memberRole == protobuf.CommunityMember_ROLE_ADMIN ||
			memberRole == protobuf.CommunityMember_ROLE_TOKEN_MASTER {

			privilegedMembers[memberRole] = append(privilegedMembers[memberRole], member)
		}
	}
	return privilegedMembers
}

func (o *Community) HasPermissionToSendCommunityEvents() bool {
	return !o.IsControlNode() && o.hasRoles(o.config.MemberIdentity, manageCommunityRoles())
}

func (o *Community) hasPermissionToSendCommunityEvent(event protobuf.CommunityEvent_EventType) bool {
	return !o.IsControlNode() && canRolesPerformEvent(o.rolesOf(o.config.MemberIdentity), event)
}

func (o *Community) hasPermissionToSendTokenPermissionCommunityEvent(event protobuf.CommunityEvent_EventType, permissionType protobuf.CommunityTokenPermission_Type) bool {
	roles := o.rolesOf(o.config.MemberIdentity)
	return !o.IsControlNode() && canRolesPerformEvent(roles, event) && canRolesModifyPermission(roles, permissionType)
}

func (o *Community) IsMemberOwner(publicKey *ecdsa.PublicKey) bool {
	return o.hasRoles(publicKey, ownerRole())
}

func (o *Community) IsMemberTokenMaster(publicKey *ecdsa.PublicKey) bool {
	return o.hasRoles(publicKey, tokenMasterRole())
}

func (o *Community) IsMemberAdmin(publicKey *ecdsa.PublicKey) bool {
	return o.hasRoles(publicKey, adminRole())
}

func (o *Community) IsPrivilegedMember(publicKey *ecdsa.PublicKey) bool {
	return o.hasRoles(publicKey, manageCommunityRoles())
}

func manageCommunityRoles() map[protobuf.CommunityMember_Roles]bool {
	roles := make(map[protobuf.CommunityMember_Roles]bool)
	roles[protobuf.CommunityMember_ROLE_OWNER] = true
	roles[protobuf.CommunityMember_ROLE_ADMIN] = true
	roles[protobuf.CommunityMember_ROLE_TOKEN_MASTER] = true
	return roles
}

func ownerRole() map[protobuf.CommunityMember_Roles]bool {
	roles := make(map[protobuf.CommunityMember_Roles]bool)
	roles[protobuf.CommunityMember_ROLE_OWNER] = true
	return roles
}

func adminRole() map[protobuf.CommunityMember_Roles]bool {
	roles := make(map[protobuf.CommunityMember_Roles]bool)
	roles[protobuf.CommunityMember_ROLE_ADMIN] = true
	return roles
}

func tokenMasterRole() map[protobuf.CommunityMember_Roles]bool {
	roles := make(map[protobuf.CommunityMember_Roles]bool)
	roles[protobuf.CommunityMember_ROLE_TOKEN_MASTER] = true
	return roles
}

func (o *Community) MemberRole(pubKey *ecdsa.PublicKey) protobuf.CommunityMember_Roles {
	if o.IsMemberOwner(pubKey) {
		return protobuf.CommunityMember_ROLE_OWNER
	} else if o.IsMemberTokenMaster(pubKey) {
		return protobuf.CommunityMember_ROLE_TOKEN_MASTER
	} else if o.IsMemberAdmin(pubKey) {
		return protobuf.CommunityMember_ROLE_ADMIN
	}

	return protobuf.CommunityMember_ROLE_NONE
}

func (o *Community) validateRequestToJoinWithChatID(request *protobuf.CommunityRequestToJoin) error {

	chat, ok := o.config.CommunityDescription.Chats[request.ChatId]

	if !ok {
		return ErrChatNotFound
	}

	// If chat is no permissions, access should not have been requested
	if chat.Permissions.Access != protobuf.CommunityPermissions_MANUAL_ACCEPT {
		return ErrCantRequestAccess
	}

	if chat.Permissions.EnsOnly && len(request.EnsName) == 0 {
		return ErrCantRequestAccess
	}

	return nil
}

func (o *Community) ManualAccept() bool {
	return o.config.CommunityDescription.Permissions.Access == protobuf.CommunityPermissions_MANUAL_ACCEPT
}

func (o *Community) AutoAccept() bool {
	// We no longer have the notion of "no membership", but for historical reasons
	// we use `NO_MEMBERSHIP` to determine wether requests to join should be automatically
	// accepted or not.
	return o.config.CommunityDescription.Permissions.Access == protobuf.CommunityPermissions_AUTO_ACCEPT
}

func (o *Community) validateRequestToJoinWithoutChatID(request *protobuf.CommunityRequestToJoin) error {
	// Previously, requests to join a community where only necessary when the community
	// permissions were indeed set to `ON_REQUEST`.
	// Now, users always have to request access but can get accepted automatically
	// (if permissions are set to NO_MEMBERSHIP).
	//
	// Hence, not only do we check whether the community permissions are ON_REQUEST but
	// also NO_MEMBERSHIP.
	if o.config.CommunityDescription.Permissions.Access != protobuf.CommunityPermissions_MANUAL_ACCEPT && o.config.CommunityDescription.Permissions.Access != protobuf.CommunityPermissions_AUTO_ACCEPT {
		return ErrCantRequestAccess
	}

	return nil
}

func (o *Community) ID() types.HexBytes {
	return crypto.CompressPubkey(o.config.ID)
}

func (o *Community) IDString() string {
	return types.EncodeHex(o.ID())
}

func (o *Community) UncompressedIDString() string {
	return types.EncodeHex(crypto.FromECDSAPub(o.config.ID))
}

func (o *Community) SerializedID() (string, error) {
	return multiformat.SerializeLegacyKey(o.UncompressedIDString())
}

func (o *Community) StatusUpdatesChannelID() string {
	return o.IDString() + "-ping"
}

func (o *Community) MagnetlinkMessageChannelID() string {
	return o.IDString() + "-magnetlinks"
}

func (o *Community) MemberUpdateChannelID() string {
	return o.IDString() + "-memberUpdate"
}

func (o *Community) PubsubTopic() string {
	return o.Shard().PubsubTopic()
}

func (o *Community) PubsubTopicPrivateKey() *ecdsa.PrivateKey {
	return o.config.PubsubTopicPrivateKey
}

func (o *Community) SetPubsubTopicPrivateKey(privKey *ecdsa.PrivateKey) {
	o.config.PubsubTopicPrivateKey = privKey
}

func (o *Community) PubsubTopicKey() string {
	if o.config.PubsubTopicPrivateKey == nil {
		return ""
	}
	return hexutil.Encode(crypto.FromECDSAPub(&o.config.PubsubTopicPrivateKey.PublicKey))
}

func (o *Community) PrivateKey() *ecdsa.PrivateKey {
	return o.config.PrivateKey
}

func (o *Community) setPrivateKey(pk *ecdsa.PrivateKey) {
	if pk != nil {
		o.config.PrivateKey = pk
	}
}

func (o *Community) ControlNode() *ecdsa.PublicKey {
	if o.config.ControlNode == nil {
		return o.config.ID
	}
	return o.config.ControlNode
}

func (o *Community) setControlNode(pubKey *ecdsa.PublicKey) {
	if pubKey != nil {
		o.config.ControlNode = pubKey
	}
}

func (o *Community) PublicKey() *ecdsa.PublicKey {
	return o.config.ID
}

func (o *Community) Description() *protobuf.CommunityDescription {
	return o.config.CommunityDescription
}

func (o *Community) DescriptionProtocolMessage() []byte {
	return o.config.CommunityDescriptionProtocolMessage
}

func (o *Community) marshaledDescription() ([]byte, error) {
	clone := proto.Clone(o.config.CommunityDescription).(*protobuf.CommunityDescription)

	// This is only workaround to lower the size of the message that goes over the wire,
	// see https://github.com/status-im/status-desktop/issues/12188
	dehydrateChannelsMembers(o.IDString(), clone)

	if o.encryptor != nil {
		err := encryptDescription(o.encryptor, o, clone)
		if err != nil {
			return nil, err
		}
	}

	return proto.Marshal(clone)
}

func (o *Community) MarshaledDescription() ([]byte, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.marshaledDescription()
}

func (o *Community) toProtocolMessageBytes() ([]byte, error) {
	// If we are not a control node, use the received serialized version
	if !o.IsControlNode() {
		// This should not happen, as we can only serialize on our side if we
		// created the community
		if len(o.config.CommunityDescriptionProtocolMessage) == 0 {
			return nil, ErrNotControlNode
		}

		return o.config.CommunityDescriptionProtocolMessage, nil
	}

	// serialize
	payload, err := o.marshaledDescription()
	if err != nil {
		return nil, err
	}

	// sign
	return protocol.WrapMessageV1(payload, protobuf.ApplicationMetadataMessage_COMMUNITY_DESCRIPTION, o.config.PrivateKey)
}

// ToProtocolMessageBytes returns the community in a wrapped & signed protocol message
func (o *Community) ToProtocolMessageBytes() ([]byte, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.toProtocolMessageBytes()
}

func channelHasTokenPermissions(communityID string, channelID string, permissions map[string]*protobuf.CommunityTokenPermission) bool {
	for _, tokenPermission := range permissions {
		if includes(tokenPermission.ChatIds, communityID+channelID) {
			return true
		}
	}
	return false
}

func dehydrateChannelsMembers(communityID string, description *protobuf.CommunityDescription) {
	// To save space, we don't attach members for channels without permissions,
	// otherwise the message will hit waku msg size limit.
	for channelID, channel := range description.Chats {
		if !channelHasTokenPermissions(communityID, channelID, description.TokenPermissions) {
			channel.Members = map[string]*protobuf.CommunityMember{} // clean members
		}
	}
}

func hydrateChannelsMembers(communityID string, description *protobuf.CommunityDescription) {
	for channelID, channel := range description.Chats {
		if !channelHasTokenPermissions(communityID, channelID, description.TokenPermissions) {
			channel.Members = make(map[string]*protobuf.CommunityMember)
			for pubKey, member := range description.Members {
				channel.Members[pubKey] = member
			}
		}
	}
}

func (o *Community) Chats() map[string]*protobuf.CommunityChat {
	// Why are we checking here for nil, it should be the responsibility of the caller
	if o == nil {
		return make(map[string]*protobuf.CommunityChat)
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.chats()
}

func (o *Community) chats() map[string]*protobuf.CommunityChat {
	response := make(map[string]*protobuf.CommunityChat)

	if o.config != nil && o.config.CommunityDescription != nil {
		for k, v := range o.config.CommunityDescription.Chats {
			response[k] = v
		}
	}

	return response
}

func (o *Community) Images() map[string]*protobuf.IdentityImage {
	response := make(map[string]*protobuf.IdentityImage)

	// Why are we checking here for nil, it should be the responsibility of the caller
	if o == nil {
		return response
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.config != nil && o.config.CommunityDescription != nil && o.config.CommunityDescription.Identity != nil {
		for k, v := range o.config.CommunityDescription.Identity.Images {
			response[k] = v
		}
	}

	return response
}

func (o *Community) Categories() map[string]*protobuf.CommunityCategory {
	response := make(map[string]*protobuf.CommunityCategory)

	if o == nil {
		return response
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.config != nil && o.config.CommunityDescription != nil {
		for k, v := range o.config.CommunityDescription.Categories {
			response[k] = v
		}
	}

	return response
}

func (o *Community) tokenPermissions() map[string]*CommunityTokenPermission {
	result := make(map[string]*CommunityTokenPermission, len(o.config.CommunityDescription.TokenPermissions))
	for _, tokenPermission := range o.config.CommunityDescription.TokenPermissions {
		result[tokenPermission.Id] = NewCommunityTokenPermission(tokenPermission)
	}

	// Non-privileged members should not see pending permissions
	if o.config.EventsData == nil || !o.IsPrivilegedMember(o.MemberIdentity()) {
		return result
	}

	processedPermissions := make(map[string]*struct{})
	for _, event := range o.config.EventsData.Events {
		if event.TokenPermission == nil || processedPermissions[event.TokenPermission.Id] != nil {
			continue
		}
		processedPermissions[event.TokenPermission.Id] = &struct{}{} // first permission event wins

		switch event.Type {
		case protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE:
			eventsTokenPermission := NewCommunityTokenPermission(event.TokenPermission)
			if result[event.TokenPermission.Id] != nil {
				eventsTokenPermission.State = TokenPermissionUpdatePending
			} else {
				eventsTokenPermission.State = TokenPermissionAdditionPending
			}
			result[eventsTokenPermission.Id] = eventsTokenPermission

		case protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE:
			tokenPermission := result[event.TokenPermission.Id]
			if tokenPermission != nil {
				tokenPermission.State = TokenPermissionRemovalPending
			}
		default:
		}
	}

	return result
}

func (o *Community) PendingAndBannedMembers() map[string]CommunityMemberState {
	result := make(map[string]CommunityMemberState)

	for _, bannedMemberID := range o.config.CommunityDescription.BanList {
		result[bannedMemberID] = CommunityMemberBanned
	}

	if o.config.EventsData == nil {
		return result
	}

	processedEvents := make(map[string]bool)
	for _, event := range o.config.EventsData.Events {
		if processedEvents[event.MemberToAction] {
			continue
		}

		switch event.Type {
		case protobuf.CommunityEvent_COMMUNITY_MEMBER_KICK:
			result[event.MemberToAction] = CommunityMemberKickPending
		case protobuf.CommunityEvent_COMMUNITY_MEMBER_BAN:
			result[event.MemberToAction] = CommunityMemberBanPending
		case protobuf.CommunityEvent_COMMUNITY_MEMBER_UNBAN:
			result[event.MemberToAction] = CommunityMemberUnbanPending
		default:
			continue
		}
		processedEvents[event.MemberToAction] = true
	}

	return result
}

func (o *Community) TokenPermissions() map[string]*CommunityTokenPermission {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.tokenPermissions()
}

func (o *Community) HasTokenPermissions() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return len(o.tokenPermissions()) > 0
}

func (o *Community) channelEncrypted(channelID string) bool {
	return o.channelHasTokenPermissions(o.ChatID(channelID))
}

func (o *Community) ChannelEncrypted(channelID string) bool {
	return o.ChannelHasTokenPermissions(o.ChatID(channelID))
}

func (o *Community) channelHasTokenPermissions(chatID string) bool {
	for _, tokenPermission := range o.tokenPermissions() {
		if includes(tokenPermission.ChatIds, chatID) {
			return true
		}
	}

	return false
}

func (o *Community) ChannelHasTokenPermissions(chatID string) bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.channelHasTokenPermissions(chatID)
}

func TokenPermissionsByType(permissions map[string]*CommunityTokenPermission, permissionType protobuf.CommunityTokenPermission_Type) []*CommunityTokenPermission {
	result := make([]*CommunityTokenPermission, 0)
	for _, tokenPermission := range permissions {
		if tokenPermission.Type == permissionType {
			result = append(result, tokenPermission)
		}
	}
	return result
}

func (o *Community) tokenPermissionByID(ID string) *CommunityTokenPermission {
	return o.tokenPermissions()[ID]
}

func (o *Community) TokenPermissionByID(ID string) *CommunityTokenPermission {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.tokenPermissionByID(ID)
}

func (o *Community) TokenPermissionsByType(permissionType protobuf.CommunityTokenPermission_Type) []*CommunityTokenPermission {
	return TokenPermissionsByType(o.tokenPermissions(), permissionType)
}

func (o *Community) ChannelTokenPermissionsByType(channelID string, permissionType protobuf.CommunityTokenPermission_Type) []*CommunityTokenPermission {
	permissions := make([]*CommunityTokenPermission, 0)
	for _, tokenPermission := range o.tokenPermissions() {
		if tokenPermission.Type == permissionType && includes(tokenPermission.ChatIds, channelID) {
			permissions = append(permissions, tokenPermission)
		}
	}
	return permissions
}

func includes(channelIDs []string, channelID string) bool {
	for _, id := range channelIDs {
		if id == channelID {
			return true
		}
	}
	return false
}

func (o *Community) UpsertTokenPermission(tokenPermission *protobuf.CommunityTokenPermission) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.IsControlNode() {
		changes, err := o.upsertTokenPermission(tokenPermission)
		if err != nil {
			return nil, err
		}

		o.increaseClock()

		return changes, nil
	}

	if o.hasPermissionToSendTokenPermissionCommunityEvent(protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_CHANGE, tokenPermission.Type) {
		existed := o.tokenPermissionByID(tokenPermission.Id) != nil

		err := o.addNewCommunityEvent(o.ToCommunityTokenPermissionChangeCommunityEvent(tokenPermission))
		if err != nil {
			return nil, err
		}

		permission := NewCommunityTokenPermission(tokenPermission)

		changes := o.emptyCommunityChanges()
		if existed {
			permission.State = TokenPermissionUpdatePending
			changes.TokenPermissionsModified[tokenPermission.Id] = permission
		} else {
			permission.State = TokenPermissionAdditionPending
			changes.TokenPermissionsAdded[tokenPermission.Id] = permission
		}

		return changes, nil
	}

	return nil, ErrNotAuthorized
}

func (o *Community) DeleteTokenPermission(permissionID string) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	tokenPermission, exists := o.config.CommunityDescription.TokenPermissions[permissionID]
	if !exists {
		return nil, ErrTokenPermissionNotFound
	}

	if o.IsControlNode() {
		changes, err := o.deleteTokenPermission(permissionID)
		if err != nil {
			return nil, err
		}

		o.increaseClock()

		return changes, nil
	}

	if o.hasPermissionToSendTokenPermissionCommunityEvent(protobuf.CommunityEvent_COMMUNITY_MEMBER_TOKEN_PERMISSION_DELETE, tokenPermission.Type) {
		err := o.addNewCommunityEvent(o.ToCommunityTokenPermissionDeleteCommunityEvent(tokenPermission))
		if err != nil {
			return nil, err
		}

		permission := NewCommunityTokenPermission(tokenPermission)
		permission.State = TokenPermissionRemovalPending

		changes := o.emptyCommunityChanges()
		changes.TokenPermissionsModified[permission.Id] = permission

		return changes, nil
	}

	return nil, ErrNotAuthorized
}

func (o *Community) VerifyGrantSignature(data []byte) (*protobuf.Grant, error) {
	if len(data) <= signatureLength {
		return nil, ErrInvalidGrant
	}
	signature := data[:signatureLength]
	payload := data[signatureLength:]
	grant := &protobuf.Grant{}
	err := proto.Unmarshal(payload, grant)
	if err != nil {
		return nil, err
	}

	if grant.Clock == 0 {
		return nil, ErrInvalidGrant
	}
	if grant.MemberId == nil {
		return nil, ErrInvalidGrant
	}
	if !bytes.Equal(grant.CommunityId, o.ID()) {
		return nil, ErrInvalidGrant
	}

	extractedPublicKey, err := crypto.SigToPub(crypto.Keccak256(payload), signature)
	if err != nil {
		return nil, err
	}

	if !common.IsPubKeyEqual(o.ControlNode(), extractedPublicKey) {
		return nil, ErrInvalidGrant
	}

	return grant, nil
}

func (o *Community) CanPost(pk *ecdsa.PublicKey, chatID string) (bool, error) {
	if o.config.CommunityDescription.Chats == nil {
		o.config.Logger.Debug("Community.CanPost: no-chats")
		return false, nil
	}

	chat, ok := o.config.CommunityDescription.Chats[chatID]
	if !ok {
		o.config.Logger.Debug("Community.CanPost: no chat with id", zap.String("chat-id", chatID))
		return false, nil
	}

	// community creator can always post, return immediately
	if common.IsPubKeyEqual(pk, o.ControlNode()) {
		return true, nil
	}

	if o.isBanned(pk) {
		o.config.Logger.Debug("Community.CanPost: user is banned", zap.String("chat-id", chatID))
		return false, nil
	}

	if o.config.CommunityDescription.Members == nil {
		o.config.Logger.Debug("Community.CanPost: no members in org", zap.String("chat-id", chatID))
		return false, nil
	}

	// If community member, also check chat membership next
	_, ok = o.config.CommunityDescription.Members[common.PubkeyToHex(pk)]
	if !ok {
		o.config.Logger.Debug("Community.CanPost: not a community member", zap.String("chat-id", chatID))
		return false, nil
	}

	if chat.Members == nil {
		o.config.Logger.Debug("Community.CanPost: no members in chat", zap.String("chat-id", chatID))
		return false, nil
	}

	// Need to also be a chat member to post
	if !o.IsMemberInChat(pk, chatID) {
		o.config.Logger.Debug("Community.CanPost: not a chat member", zap.String("chat-id", chatID))
		return false, nil
	}

	// all conditions satisfied, user can post after all
	return true, nil
}

func (o *Community) BuildGrant(key *ecdsa.PublicKey, chatID string) ([]byte, error) {
	return o.buildGrant(key, chatID)
}

func (o *Community) buildGrant(key *ecdsa.PublicKey, chatID string) ([]byte, error) {
	bytes := make([]byte, 0)
	if o.IsControlNode() {
		grant := &protobuf.Grant{
			CommunityId: o.ID(),
			MemberId:    crypto.CompressPubkey(key),
			ChatId:      chatID,
			Clock:       o.config.CommunityDescription.Clock,
		}
		marshaledGrant, err := proto.Marshal(grant)
		if err != nil {
			return nil, err
		}

		signatureMaterial := crypto.Keccak256(marshaledGrant)

		signature, err := crypto.Sign(signatureMaterial, o.config.PrivateKey)
		if err != nil {
			return nil, err
		}

		bytes = append(signature, marshaledGrant...)
	}
	return bytes, nil
}

func (o *Community) increaseClock() {
	o.config.CommunityDescription.Clock = o.nextClock()
}

func (o *Community) Clock() uint64 {
	return o.config.CommunityDescription.Clock
}

func (o *Community) CanRequestAccess(pk *ecdsa.PublicKey) bool {
	if o.hasMember(pk) {
		return false
	}

	if o.config.CommunityDescription == nil {
		return false
	}

	if o.config.CommunityDescription.Permissions == nil {
		return false
	}

	return o.config.CommunityDescription.Permissions.Access == protobuf.CommunityPermissions_MANUAL_ACCEPT
}

func (o *Community) CanManageUsers(pk *ecdsa.PublicKey) bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.IsPrivilegedMember(pk)
}

func (o *Community) CanDeleteMessageForEveryone(pk *ecdsa.PublicKey) bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.IsPrivilegedMember(pk)
}

func (o *Community) isMember() bool {
	return o.hasMember(o.config.MemberIdentity)
}

func (o *Community) CanMemberIdentityPost(chatID string) (bool, error) {
	return o.CanPost(o.config.MemberIdentity, chatID)
}

// CanJoin returns whether a user can join the community, only if it's
func (o *Community) canJoin() bool {
	if o.config.Joined {
		return false
	}

	if o.IsControlNode() {
		return true
	}

	if o.config.CommunityDescription.Permissions.Access == protobuf.CommunityPermissions_AUTO_ACCEPT {
		return true
	}

	return o.isMember()
}

func (o *Community) RequestedToJoinAt() uint64 {
	return o.config.RequestedToJoinAt
}

func (o *Community) nextClock() uint64 {
	// lamport timestamp
	clock := o.config.CommunityDescription.Clock
	timestamp := o.timesource.GetCurrentTime()
	if clock == 0 || clock < timestamp {
		clock = timestamp
	} else {
		clock = clock + 1
	}

	return clock
}

func (o *Community) CanManageUsersPublicKeys() ([]*ecdsa.PublicKey, error) {
	var response []*ecdsa.PublicKey
	roles := manageCommunityRoles()
	for pkString, member := range o.config.CommunityDescription.Members {
		if o.memberHasRoles(member, roles) {
			pk, err := common.HexToPubkey(pkString)
			if err != nil {
				return nil, err
			}

			response = append(response, pk)
		}

	}
	return response, nil
}

func (o *Community) AddRequestToJoin(request *RequestToJoin) {
	o.config.RequestsToJoin = append(o.config.RequestsToJoin, request)
}

func (o *Community) RequestsToJoin() []*RequestToJoin {
	return o.config.RequestsToJoin
}

func (o *Community) AddMember(publicKey *ecdsa.PublicKey, roles []protobuf.CommunityMember_Roles) (*CommunityChanges, error) {
	if !o.IsControlNode() {
		return nil, ErrNotControlNode
	}

	memberKey := common.PubkeyToHex(publicKey)
	changes := o.emptyCommunityChanges()

	if o.config.CommunityDescription.Members == nil {
		o.config.CommunityDescription.Members = make(map[string]*protobuf.CommunityMember)
	}

	if _, ok := o.config.CommunityDescription.Members[memberKey]; !ok {
		o.config.CommunityDescription.Members[memberKey] = &protobuf.CommunityMember{Roles: roles}
		changes.MembersAdded[memberKey] = o.config.CommunityDescription.Members[memberKey]
	}

	o.increaseClock()
	return changes, nil
}

func (o *Community) AddMemberToChat(chatID string, publicKey *ecdsa.PublicKey, roles []protobuf.CommunityMember_Roles) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.IsControlNode() {
		return nil, ErrNotControlNode
	}

	memberKey := common.PubkeyToHex(publicKey)
	changes := o.emptyCommunityChanges()

	chat, ok := o.config.CommunityDescription.Chats[chatID]
	if !ok {
		return nil, ErrChatNotFound
	}

	if chat.Members == nil {
		chat.Members = make(map[string]*protobuf.CommunityMember)
	}
	chat.Members[memberKey] = &protobuf.CommunityMember{
		Roles: roles,
	}
	changes.ChatsModified[chatID] = &CommunityChatChanges{
		ChatModified: chat,
		MembersAdded: map[string]*protobuf.CommunityMember{
			memberKey: chat.Members[memberKey],
		},
	}

	if o.IsControlNode() {
		o.increaseClock()
	}

	return changes, nil
}

func (o *Community) PopulateChatWithAllMembers(chatID string) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.IsControlNode() {
		return o.emptyCommunityChanges(), ErrNotControlNode
	}

	return o.populateChatWithAllMembers(chatID)
}

func (o *Community) populateChatWithAllMembers(chatID string) (*CommunityChanges, error) {
	result := o.emptyCommunityChanges()

	chat, exists := o.chats()[chatID]
	if !exists {
		return result, ErrChatNotFound
	}

	membersAdded := make(map[string]*protobuf.CommunityMember)
	for pubKey, member := range o.Members() {
		if chat.Members[pubKey] == nil {
			membersAdded[pubKey] = member
		}
	}
	result.ChatsModified[chatID] = &CommunityChatChanges{
		MembersAdded: membersAdded,
	}

	chat.Members = o.Members()
	o.increaseClock()

	return result, nil
}

func (o *Community) ChatID(channelID string) string {
	return o.IDString() + channelID
}

func (o *Community) ChatIDs() (chatIDs []string) {
	for channelID := range o.config.CommunityDescription.Chats {
		chatIDs = append(chatIDs, o.ChatID(channelID))
	}
	return chatIDs
}

func (o *Community) AllowsAllMembersToPinMessage() bool {
	return o.config.CommunityDescription.AdminSettings != nil && o.config.CommunityDescription.AdminSettings.PinMessageAllMembersEnabled
}

func (o *Community) AddMemberWithRevealedAccounts(dbRequest *RequestToJoin, roles []protobuf.CommunityMember_Roles, accounts []*protobuf.RevealedAccount) (*CommunityChanges, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.IsControlNode() {
		return nil, ErrNotControlNode
	}

	changes := o.addMemberWithRevealedAccounts(dbRequest.PublicKey, roles, accounts, dbRequest.Clock)

	o.increaseClock()

	return changes, nil
}

func (o *Community) CreateDeepCopy() *Community {
	return &Community{
		encryptor: o.encryptor,
		config: &Config{
			PrivateKey:                          o.config.PrivateKey,
			ControlNode:                         o.config.ControlNode,
			ControlDevice:                       o.config.ControlDevice,
			CommunityDescription:                proto.Clone(o.config.CommunityDescription).(*protobuf.CommunityDescription),
			CommunityDescriptionProtocolMessage: o.config.CommunityDescriptionProtocolMessage,
			ID:                                  o.config.ID,
			Joined:                              o.config.Joined,
			JoinedAt:                            o.config.JoinedAt,
			Requested:                           o.config.Requested,
			Verified:                            o.config.Verified,
			Spectated:                           o.config.Spectated,
			Muted:                               o.config.Muted,
			MuteTill:                            o.config.MuteTill,
			Logger:                              o.config.Logger,
			RequestedToJoinAt:                   o.config.RequestedToJoinAt,
			RequestsToJoin:                      o.config.RequestsToJoin,
			MemberIdentity:                      o.config.MemberIdentity,
			EventsData:                          o.config.EventsData,
			Shard:                               o.config.Shard,
			PubsubTopicPrivateKey:               o.config.PubsubTopicPrivateKey,
			LastOpenedAt:                        o.config.LastOpenedAt,
		},
		timesource: o.timesource,
	}
}

func (o *Community) SetActiveMembersCount(activeMembersCount uint64) (updated bool, err error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.IsControlNode() {
		return false, ErrNotControlNode
	}

	if activeMembersCount == o.config.CommunityDescription.ActiveMembersCount {
		return false, nil
	}

	o.config.CommunityDescription.ActiveMembersCount = activeMembersCount
	o.increaseClock()

	return true, nil
}

type sortSlice []sorterHelperIdx
type sorterHelperIdx struct {
	pos    int32
	catID  string
	chatID string
}

func (d sortSlice) Len() int {
	return len(d)
}

func (d sortSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d sortSlice) Less(i, j int) bool {
	return d[i].pos < d[j].pos
}

func (o *Community) unbanUserFromCommunity(pk *ecdsa.PublicKey) {
	key := common.PubkeyToHex(pk)
	for i, v := range o.config.CommunityDescription.BanList {
		if v == key {
			o.config.CommunityDescription.BanList =
				append(o.config.CommunityDescription.BanList[:i], o.config.CommunityDescription.BanList[i+1:]...)
			break
		}
	}
}

func (o *Community) banUserFromCommunity(pk *ecdsa.PublicKey) {
	key := common.PubkeyToHex(pk)
	if o.hasMember(pk) {
		// Remove from org
		delete(o.config.CommunityDescription.Members, key)

		// Remove from chats
		for _, chat := range o.config.CommunityDescription.Chats {
			delete(chat.Members, key)
		}
	}

	for _, u := range o.config.CommunityDescription.BanList {
		if u == key {
			return
		}
	}

	o.config.CommunityDescription.BanList = append(o.config.CommunityDescription.BanList, key)
}

func (o *Community) editChat(chatID string, chat *protobuf.CommunityChat) error {
	err := validateCommunityChat(o.config.CommunityDescription, chat)
	if err != nil {
		return err
	}

	if o.config.CommunityDescription.Chats == nil {
		o.config.CommunityDescription.Chats = make(map[string]*protobuf.CommunityChat)
	}
	if _, exists := o.config.CommunityDescription.Chats[chatID]; !exists {
		return ErrChatNotFound
	}

	o.config.CommunityDescription.Chats[chatID] = chat

	return nil
}

func (o *Community) createChat(chatID string, chat *protobuf.CommunityChat) error {
	err := validateCommunityChat(o.config.CommunityDescription, chat)
	if err != nil {
		return err
	}

	if o.config.CommunityDescription.Chats == nil {
		o.config.CommunityDescription.Chats = make(map[string]*protobuf.CommunityChat)
	}
	if _, ok := o.config.CommunityDescription.Chats[chatID]; ok {
		return ErrChatAlreadyExists
	}

	for _, c := range o.config.CommunityDescription.Chats {
		if chat.Identity.DisplayName == c.Identity.DisplayName {
			return ErrInvalidCommunityDescriptionDuplicatedName
		}
	}

	// Sets the chat position to be the last within its category
	chat.Position = 0
	for _, c := range o.config.CommunityDescription.Chats {
		if c.CategoryId == chat.CategoryId {
			chat.Position++
		}
	}

	chat.Members = make(map[string]*protobuf.CommunityMember)
	for pubKey, member := range o.config.CommunityDescription.Members {
		chat.Members[pubKey] = member
	}

	o.config.CommunityDescription.Chats[chatID] = chat

	return nil
}

func (o *Community) deleteChat(chatID string) *CommunityChanges {
	if o.config.CommunityDescription.Chats == nil {
		o.config.CommunityDescription.Chats = make(map[string]*protobuf.CommunityChat)
	}

	changes := o.emptyCommunityChanges()

	if chat, exists := o.config.CommunityDescription.Chats[chatID]; exists {
		tmpCatID := chat.CategoryId
		chat.CategoryId = ""
		o.SortCategoryChats(changes, tmpCatID)
		changes.ChatsRemoved[chatID] = chat
	}

	delete(o.config.CommunityDescription.Chats, chatID)
	return changes
}

func (o *Community) upsertTokenPermission(permission *protobuf.CommunityTokenPermission) (*CommunityChanges, error) {
	existed := o.tokenPermissionByID(permission.Id) != nil

	if o.config.CommunityDescription.TokenPermissions == nil {
		o.config.CommunityDescription.TokenPermissions = make(map[string]*protobuf.CommunityTokenPermission)
	}
	o.config.CommunityDescription.TokenPermissions[permission.Id] = permission

	changes := o.emptyCommunityChanges()
	if existed {
		changes.TokenPermissionsModified[permission.Id] = NewCommunityTokenPermission(permission)
	} else {
		changes.TokenPermissionsAdded[permission.Id] = NewCommunityTokenPermission(permission)
	}

	return changes, nil
}

func (o *Community) deleteTokenPermission(permissionID string) (*CommunityChanges, error) {
	permission, exists := o.config.CommunityDescription.TokenPermissions[permissionID]
	if !exists {
		return nil, ErrTokenPermissionNotFound
	}

	delete(o.config.CommunityDescription.TokenPermissions, permissionID)

	changes := o.emptyCommunityChanges()

	changes.TokenPermissionsRemoved[permissionID] = NewCommunityTokenPermission(permission)
	return changes, nil
}

func (o *Community) addMemberWithRevealedAccounts(memberKey string, roles []protobuf.CommunityMember_Roles, accounts []*protobuf.RevealedAccount, clock uint64) *CommunityChanges {
	changes := o.emptyCommunityChanges()

	if o.config.CommunityDescription.Members == nil {
		o.config.CommunityDescription.Members = make(map[string]*protobuf.CommunityMember)
	}

	if _, ok := o.config.CommunityDescription.Members[memberKey]; !ok {
		o.config.CommunityDescription.Members[memberKey] = &protobuf.CommunityMember{Roles: roles}
		changes.MembersAdded[memberKey] = o.config.CommunityDescription.Members[memberKey]
	}

	o.config.CommunityDescription.Members[memberKey].RevealedAccounts = accounts
	o.config.CommunityDescription.Members[memberKey].LastUpdateClock = clock
	changes.MemberWalletsAdded[memberKey] = o.config.CommunityDescription.Members[memberKey].RevealedAccounts

	return changes
}

func (o *Community) DeclineRequestToJoin(dbRequest *RequestToJoin) (adminEventCreated bool, err error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !(o.IsControlNode() || o.hasPermissionToSendCommunityEvent(protobuf.CommunityEvent_COMMUNITY_REQUEST_TO_JOIN_REJECT)) {
		return adminEventCreated, ErrNotAuthorized
	}

	if o.IsControlNode() {
		// typically, community's clock is increased implicitly when making changes
		// to it, however in this scenario there are no changes in the community, yet
		// we need to increase the clock to ensure the owner event is processed by other
		// nodes.
		o.increaseClock()
	} else {
		rejectedRequestsToJoin := make(map[string]*protobuf.CommunityRequestToJoin)
		rejectedRequestsToJoin[dbRequest.PublicKey] = dbRequest.ToCommunityRequestToJoinProtobuf()

		adminChanges := &CommunityEventChanges{
			CommunityChanges:       o.emptyCommunityChanges(),
			RejectedRequestsToJoin: rejectedRequestsToJoin,
		}
		err = o.addNewCommunityEvent(o.ToCommunityRequestToJoinRejectCommunityEvent(adminChanges))
		if err != nil {
			return adminEventCreated, err
		}

		adminEventCreated = true
	}

	return adminEventCreated, err
}

func (o *Community) ValidateEvent(event *CommunityEvent, signer *ecdsa.PublicKey) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	err := validateCommunityEvent(event)
	if err != nil {
		return err
	}

	eventSender := o.getMember(signer)
	if eventSender == nil {
		return ErrMemberNotFound
	}

	eventTargetRoles := []protobuf.CommunityMember_Roles{}
	eventTargetPk, err := common.HexToPubkey(event.MemberToAction)
	if err == nil {
		eventTarget := o.getMember(eventTargetPk)
		if eventTarget != nil {
			eventTargetRoles = eventTarget.Roles
		}
	}

	if !RolesAuthorizedToPerformEvent(eventSender.Roles, eventTargetRoles, event) {
		return ErrNotAuthorized
	}

	return nil
}

func (o *Community) MemberCanManageToken(member *ecdsa.PublicKey, token *community_token.CommunityToken) bool {
	return o.IsMemberOwner(member) || o.IsControlNode() || (o.IsMemberTokenMaster(member) &&
		token.PrivilegesLevel != community_token.OwnerLevel && token.PrivilegesLevel != community_token.MasterLevel)
}

func CommunityDescriptionTokenOwnerChainID(description *protobuf.CommunityDescription) uint64 {
	if description == nil {
		return 0
	}

	// We look in TokenPermissions for a token that grants BECOME_TOKEN_OWNER rights
	// There should be only one, and it's only a single chainID
	for _, p := range description.TokenPermissions {
		if p.Type == protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER && len(p.TokenCriteria) != 0 {

			for _, criteria := range p.TokenCriteria {
				for chainID := range criteria.ContractAddresses {
					return chainID
				}
			}
		}
	}

	return 0
}

func HasTokenOwnership(description *protobuf.CommunityDescription) bool {
	return uint64(0) != CommunityDescriptionTokenOwnerChainID(description)
}
