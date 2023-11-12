package protocol

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/status-im/status-go/deprecation"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	userimage "github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
	v1protocol "github.com/status-im/status-go/protocol/v1"
	"github.com/status-im/status-go/services/utils"
)

var chatColors = []string{
	"#fa6565", // red
	"#887af9", // blue
	"#FE8F59", // orange
	"#7cda00", // green
	"#51d0f0", // light-blue
	"#d37ef4", // purple
}

type ChatType int

const (
	ChatTypeOneToOne ChatType = iota + 1
	ChatTypePublic
	ChatTypePrivateGroupChat
	// Deprecated: CreateProfileChat shouldn't be used
	// and is only left here in case profile chat feature is re-introduced.
	ChatTypeProfile
	// Deprecated: ChatTypeTimeline shouldn't be used
	// and is only left here in case profile chat feature is re-introduced.
	ChatTypeTimeline
	ChatTypeCommunityChat
)

const (
	FirstMessageTimestampUndefined = 0
	FirstMessageTimestampNoMessage = 1
)

const (
	MuteFor1MinDuration   = time.Minute
	MuteFor15MinsDuration = 15 * time.Minute
	MuteFor1HrsDuration   = time.Hour
	MuteFor8HrsDuration   = 8 * time.Hour
	MuteFor1WeekDuration  = 7 * 24 * time.Hour
)

const (
	MuteFor15Min requests.MutingVariation = iota + 1
	MuteFor1Hr
	MuteFor8Hr
	MuteFor1Week
	MuteTillUnmuted
	MuteTill1Min
	Unmuted
)

const pkStringLength = 68

// timelineChatID is a magic constant id for your own timeline
// Deprecated: timeline chats are no more supported
const timelineChatID = "@timeline70bd746ddcc12beb96b2c9d572d0784ab137ffc774f5383e50585a932080b57cca0484b259e61cecbaa33a4c98a300a"

type Chat struct {
	// ID is the id of the chat, for public chats it is the name e.g. status, for one-to-one
	// is the hex encoded public key and for group chats is a random uuid appended with
	// the hex encoded pk of the creator of the chat
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Emoji       string `json:"emoji"`
	// Active indicates whether the chat has been soft deleted
	Active bool `json:"active"`

	ChatType ChatType `json:"chatType"`

	// Timestamp indicates the last time this chat has received/sent a message
	Timestamp int64 `json:"timestamp"`
	// LastClockValue indicates the last clock value to be used when sending messages
	LastClockValue uint64 `json:"lastClockValue"`
	// DeletedAtClockValue indicates the clock value at time of deletion, messages
	// with lower clock value of this should be discarded
	DeletedAtClockValue uint64 `json:"deletedAtClockValue"`
	// ReadMessagesAtClockValue indicates the clock value of time till all
	// messages are considered as read
	ReadMessagesAtClockValue uint64
	// Denormalized fields
	UnviewedMessagesCount uint            `json:"unviewedMessagesCount"`
	UnviewedMentionsCount uint            `json:"unviewedMentionsCount"`
	LastMessage           *common.Message `json:"lastMessage"`

	// Group chat fields
	// Members are the members who have been invited to the group chat
	Members []ChatMember `json:"members"`
	// MembershipUpdates is all the membership events in the chat
	MembershipUpdates []v1protocol.MembershipUpdateEvent `json:"membershipUpdateEvents"`

	// Generated username name of the chat for one-to-ones
	Alias string `json:"alias,omitempty"`
	// Identicon generated from public key
	Identicon string `json:"identicon"`

	// Muted is used to check whether we want to receive
	// push notifications for this chat
	Muted bool `json:"muted"`

	// Time in which chat was muted
	MuteTill time.Time `json:"muteTill,omitempty"`

	// Public key of administrator who created invitation link
	InvitationAdmin string `json:"invitationAdmin,omitempty"`

	// Public key of administrator who sent us group invitation
	ReceivedInvitationAdmin string `json:"receivedInvitationAdmin,omitempty"`

	// Public key of user profile
	Profile string `json:"profile,omitempty"`

	// CommunityID is the id of the community it belongs to
	CommunityID string `json:"communityId,omitempty"`

	// CategoryID is the id of the community category this chat belongs to.
	CategoryID string `json:"categoryId,omitempty"`

	// Joined is a timestamp that indicates when the chat was joined
	Joined int64 `json:"joined,omitempty"`

	// SyncedTo is the time up until it has synced with a mailserver
	SyncedTo uint32 `json:"syncedTo,omitempty"`

	// SyncedFrom is the time from when it was synced with a mailserver
	SyncedFrom uint32 `json:"syncedFrom,omitempty"`

	// FirstMessageTimestamp is the time when first message was sent/received on the chat
	// valid only for community chats
	// 0 - undefined
	// 1 - no messages
	FirstMessageTimestamp uint32 `json:"firstMessageTimestamp,omitempty"`

	// Highlight is used for highlight chats
	Highlight bool `json:"highlight,omitempty"`

	// Image of the chat in Base64 format
	Base64Image string `json:"image,omitempty"`
}

type ChatPreview struct {
	// ID is the id of the chat, for public chats it is the name e.g. status, for one-to-one
	// is the hex encoded public key and for group chats is a random uuid appended with
	// the hex encoded pk of the creator of the chat
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Emoji       string `json:"emoji"`
	// Active indicates whether the chat has been soft deleted
	Active bool `json:"active"`

	ChatType ChatType `json:"chatType"`

	// Timestamp indicates the last time this chat has received/sent a message
	Timestamp int64 `json:"timestamp"`
	// LastClockValue indicates the last clock value to be used when sending messages
	LastClockValue uint64 `json:"lastClockValue"`
	// DeletedAtClockValue indicates the clock value at time of deletion, messages
	// with lower clock value of this should be discarded
	DeletedAtClockValue uint64 `json:"deletedAtClockValue"`

	// Denormalized fields
	UnviewedMessagesCount uint `json:"unviewedMessagesCount"`
	UnviewedMentionsCount uint `json:"unviewedMentionsCount"`

	// Generated username name of the chat for one-to-ones
	Alias string `json:"alias,omitempty"`
	// Identicon generated from public key
	Identicon string `json:"identicon"`

	// Muted is used to check whether we want to receive
	// push notifications for this chat
	Muted bool `json:"muted,omitempty"`

	// Time in which chat will be  ummuted
	MuteTill time.Time `json:"muteTill,omitempty"`

	// Public key of user profile
	Profile string `json:"profile,omitempty"`

	// CommunityID is the id of the community it belongs to
	CommunityID string `json:"communityId,omitempty"`

	// CategoryID is the id of the community category this chat belongs to.
	CategoryID string `json:"categoryId,omitempty"`

	// Joined is a timestamp that indicates when the chat was joined
	Joined int64 `json:"joined,omitempty"`

	// SyncedTo is the time up until it has synced with a mailserver
	SyncedTo uint32 `json:"syncedTo,omitempty"`

	// SyncedFrom is the time from when it was synced with a mailserver
	SyncedFrom uint32 `json:"syncedFrom,omitempty"`

	// ParsedText is the parsed markdown for displaying
	ParsedText json.RawMessage `json:"parsedText,omitempty"`

	Text string `json:"text,omitempty"`

	ContentType protobuf.ChatMessage_ContentType `json:"contentType,omitempty"`

	// Highlight is used for highlight chats
	Highlight bool `json:"highlight,omitempty"`

	// Used for display invited community's name in the last message
	ContentCommunityID string `json:"contentCommunityId,omitempty"`

	// Members array to represent how many there are for chats preview of group chats
	Members []ChatMember `json:"members"`

	OutgoingStatus   string `json:"outgoingStatus,omitempty"`
	ResponseTo       string `json:"responseTo"`
	AlbumImagesCount uint32 `json:"albumImagesCount,omitempty"`
	From             string `json:"from"`
	Deleted          bool   `json:"deleted"`
	DeletedForMe     bool   `json:"deletedForMe"`
}

func (c *Chat) PublicKey() (*ecdsa.PublicKey, error) {
	// For one to one chatID is an encoded public key
	if c.ChatType != ChatTypeOneToOne {
		return nil, nil
	}
	return common.HexToPubkey(c.ID)
}

func (c *Chat) Public() bool {
	return c.ChatType == ChatTypePublic
}

// Deprecated: ProfileUpdates shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
func (c *Chat) ProfileUpdates() bool {
	return c.ChatType == ChatTypeProfile || len(c.Profile) > 0
}

// Deprecated: Timeline shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
func (c *Chat) Timeline() bool {
	return c.ChatType == ChatTypeTimeline
}

func (c *Chat) OneToOne() bool {
	return c.ChatType == ChatTypeOneToOne
}

func (c *Chat) CommunityChat() bool {
	return c.ChatType == ChatTypeCommunityChat
}

func (c *Chat) PrivateGroupChat() bool {
	return c.ChatType == ChatTypePrivateGroupChat
}

func (c *Chat) IsActivePersonalChat() bool {
	return c.Active && (c.OneToOne() || c.PrivateGroupChat() || c.Public()) && c.CommunityID == ""
}

func (c *Chat) shouldBeSynced() bool {
	isPublicChat := !c.Timeline() && !c.ProfileUpdates() && c.Public()
	return isPublicChat || c.OneToOne() || c.PrivateGroupChat()
}

func (c *Chat) CommunityChatID() string {
	if c.ChatType != ChatTypeCommunityChat {
		return c.ID
	}

	// Strips out the local prefix of the community-id
	return c.ID[pkStringLength:]
}

func (c *Chat) Validate() error {
	if c.ID == "" {
		return errors.New("chatID can't be blank")
	}

	if c.OneToOne() {
		_, err := c.PublicKey()
		return err
	}
	return nil
}

func (c *Chat) MembersAsPublicKeys() ([]*ecdsa.PublicKey, error) {
	publicKeys := make([]string, len(c.Members))
	for idx, item := range c.Members {
		publicKeys[idx] = item.ID
	}
	return stringSliceToPublicKeys(publicKeys)
}

func (c *Chat) HasMember(memberID string) bool {
	for _, member := range c.Members {
		if memberID == member.ID {
			return true
		}
	}

	return false
}

func (c *Chat) RemoveMember(memberID string) {
	members := c.Members
	c.Members = []ChatMember{}
	for _, member := range members {
		if memberID != member.ID {
			c.Members = append(c.Members, member)
		}
	}
}

func (c *Chat) updateChatFromGroupMembershipChanges(g *v1protocol.Group) {

	// ID
	c.ID = g.ChatID()

	// Name
	c.Name = g.Name()

	// Color
	color := g.Color()
	if color != "" {
		c.Color = g.Color()
	}

	// Image
	base64Image, err := userimage.GetPayloadDataURI(g.Image())
	if err == nil {
		c.Base64Image = base64Image
	}

	// Members
	members := g.Members()
	admins := g.Admins()
	chatMembers := make([]ChatMember, 0, len(members))
	for _, m := range members {

		chatMember := ChatMember{
			ID: m,
		}
		chatMember.Admin = stringSliceContains(admins, m)
		chatMembers = append(chatMembers, chatMember)
	}
	c.Members = chatMembers

	// MembershipUpdates
	c.MembershipUpdates = g.Events()
}

// NextClockAndTimestamp returns the next clock value
// and the current timestamp
func (c *Chat) NextClockAndTimestamp(timesource common.TimeSource) (uint64, uint64) {
	clock := c.LastClockValue
	timestamp := timesource.GetCurrentTime()
	if clock == 0 || clock < timestamp {
		clock = timestamp
	} else {
		clock = clock + 1
	}
	c.LastClockValue = clock

	return clock, timestamp
}

func (c *Chat) UpdateFromMessage(message *common.Message, timesource common.TimeSource) error {
	c.Timestamp = int64(timesource.GetCurrentTime())

	// If the clock of the last message is lower, we set the message
	if c.LastMessage == nil || c.LastMessage.Clock <= message.Clock {
		c.LastMessage = message
	}
	// If the clock is higher we set the clock
	if c.LastClockValue < message.Clock {
		c.LastClockValue = message.Clock
	}
	return nil
}

func (c *Chat) UpdateFirstMessageTimestamp(timestamp uint32) bool {
	if timestamp == c.FirstMessageTimestamp {
		return false
	}

	// Do not allow to assign `Undefined`` or `NoMessage` to already set timestamp
	if timestamp == FirstMessageTimestampUndefined ||
		(timestamp == FirstMessageTimestampNoMessage &&
			c.FirstMessageTimestamp != FirstMessageTimestampUndefined) {
		return false
	}

	if c.FirstMessageTimestamp == FirstMessageTimestampUndefined ||
		c.FirstMessageTimestamp == FirstMessageTimestampNoMessage ||
		timestamp < c.FirstMessageTimestamp {
		c.FirstMessageTimestamp = timestamp
		return true
	}

	return false
}

// ChatMembershipUpdate represent an event on membership of the chat
type ChatMembershipUpdate struct {
	// Unique identifier for the event
	ID string `json:"id"`
	// Type indicates the kind of event
	Type protobuf.MembershipUpdateEvent_EventType `json:"type"`
	// Name represents the name in the event of changing name events
	Name string `json:"name,omitempty"`
	// Clock value of the event
	ClockValue uint64 `json:"clockValue"`
	// Signature of the event
	Signature string `json:"signature"`
	// Hex encoded public key of the creator of the event
	From string `json:"from"`
	// Target of the event for single-target events
	Member string `json:"member,omitempty"`
	// Target of the event for multi-target events
	Members []string `json:"members,omitempty"`
}

// ChatMember represents a member who participates in a group chat
type ChatMember struct {
	// ID is the hex encoded public key of the member
	ID string `json:"id"`
	// Admin indicates if the member is an admin of the group chat
	Admin bool `json:"admin"`
}

func (c ChatMember) PublicKey() (*ecdsa.PublicKey, error) {
	return common.HexToPubkey(c.ID)
}

func oneToOneChatID(publicKey *ecdsa.PublicKey) string {
	return types.EncodeHex(crypto.FromECDSAPub(publicKey))
}

func OneToOneFromPublicKey(pk *ecdsa.PublicKey, timesource common.TimeSource) *Chat {
	chatID := types.EncodeHex(crypto.FromECDSAPub(pk))
	newChat := CreateOneToOneChat(chatID[:8], pk, timesource)

	return newChat
}

func CreateOneToOneChat(name string, publicKey *ecdsa.PublicKey, timesource common.TimeSource) *Chat {
	timestamp := timesource.GetCurrentTime()
	return &Chat{
		ID:                       oneToOneChatID(publicKey),
		Name:                     name,
		Timestamp:                int64(timestamp),
		ReadMessagesAtClockValue: 0,
		Active:                   true,
		Joined:                   int64(timestamp),
		ChatType:                 ChatTypeOneToOne,
		Highlight:                true,
	}
}

func CreateCommunityChat(orgID, chatID string, orgChat *protobuf.CommunityChat, timesource common.TimeSource) *Chat {
	color := orgChat.Identity.Color
	if color == "" {
		color = chatColors[rand.Intn(len(chatColors))] // nolint: gosec
	}

	timestamp := timesource.GetCurrentTime()
	return &Chat{
		CommunityID:              orgID,
		CategoryID:               orgChat.CategoryId,
		Name:                     orgChat.Identity.DisplayName,
		Description:              orgChat.Identity.Description,
		Active:                   true,
		Color:                    color,
		Emoji:                    orgChat.Identity.Emoji,
		ID:                       orgID + chatID,
		Timestamp:                int64(timestamp),
		Joined:                   int64(timestamp),
		ReadMessagesAtClockValue: 0,
		ChatType:                 ChatTypeCommunityChat,
		FirstMessageTimestamp:    orgChat.Identity.FirstMessageTimestamp,
	}
}

func (c *Chat) DeepLink() string {
	if c.OneToOne() {
		return "status-app://p/" + c.ID
	}
	if c.PrivateGroupChat() {
		return "status-app://g/args?a2=" + c.ID
	}

	if c.CommunityChat() {
		communityChannelID := strings.TrimPrefix(c.ID, c.CommunityID)
		pubkey, err := types.DecodeHex(c.CommunityID)
		if err != nil {
			return ""
		}

		serializedCommunityID, err := utils.SerializePublicKey(pubkey)

		if err != nil {
			return ""
		}

		return "status-app://cc/" + communityChannelID + "#" + serializedCommunityID
	}

	if c.Public() {
		return "status-app://" + c.ID
	}

	return ""
}

func CreateCommunityChats(org *communities.Community, timesource common.TimeSource) []*Chat {
	var chats []*Chat
	orgID := org.IDString()

	for chatID, chat := range org.Chats() {
		chats = append(chats, CreateCommunityChat(orgID, chatID, chat, timesource))
	}
	return chats
}

func CreatePublicChat(name string, timesource common.TimeSource) *Chat {
	timestamp := timesource.GetCurrentTime()
	return &Chat{
		ID:                       name,
		Name:                     name,
		Active:                   true,
		Timestamp:                int64(timestamp),
		Joined:                   int64(timestamp),
		ReadMessagesAtClockValue: 0,
		Color:                    chatColors[rand.Intn(len(chatColors))], // nolint: gosec
		ChatType:                 ChatTypePublic,
	}
}

// Deprecated: buildProfileChatID shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
func buildProfileChatID(publicKeyString string) string {
	return "@" + publicKeyString
}

// Deprecated: CreateProfileChat shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
func CreateProfileChat(pubkey string, timesource common.TimeSource) *Chat {
	// Return nil to prevent usage of deprecated function
	if deprecation.ChatProfileDeprecated {
		return nil
	}

	id := buildProfileChatID(pubkey)
	return &Chat{
		ID:        id,
		Name:      id,
		Active:    true,
		Timestamp: int64(timesource.GetCurrentTime()),
		Joined:    int64(timesource.GetCurrentTime()),
		Color:     chatColors[rand.Intn(len(chatColors))], // nolint: gosec
		ChatType:  ChatTypeProfile,
		Profile:   pubkey,
	}
}

func CreateGroupChat(timesource common.TimeSource) Chat {
	timestamp := timesource.GetCurrentTime()
	synced := uint32(timestamp / 1000)

	return Chat{
		Active:                   true,
		Color:                    chatColors[rand.Intn(len(chatColors))], // nolint: gosec
		Timestamp:                int64(timestamp),
		ReadMessagesAtClockValue: 0,
		SyncedTo:                 synced,
		SyncedFrom:               synced,
		ChatType:                 ChatTypePrivateGroupChat,
		Highlight:                true,
	}
}

// Deprecated: CreateTimelineChat shouldn't be used
// and is only left here in case profile chat feature is re-introduced.
func CreateTimelineChat(timesource common.TimeSource) *Chat {
	// Return nil to prevent usage of deprecated function
	if deprecation.ChatTimelineDeprecated {
		return nil
	}

	return &Chat{
		ID:        timelineChatID,
		Name:      "#" + timelineChatID,
		Timestamp: int64(timesource.GetCurrentTime()),
		Active:    true,
		ChatType:  ChatTypeTimeline,
	}
}

func stringSliceToPublicKeys(slice []string) ([]*ecdsa.PublicKey, error) {
	result := make([]*ecdsa.PublicKey, len(slice))
	for idx, item := range slice {
		var err error
		result[idx], err = common.HexToPubkey(item)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func stringSliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
