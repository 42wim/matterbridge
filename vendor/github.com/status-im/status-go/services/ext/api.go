package ext

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/services/browsers"
	"github.com/status-im/status-go/services/wallet"
	"github.com/status-im/status-go/services/wallet/bigint"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/mailserver"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/common/shard"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/protocol/discord"
	"github.com/status-im/status-go/protocol/encryption/multidevice"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/pushnotificationclient"
	"github.com/status-im/status-go/protocol/requests"
	"github.com/status-im/status-go/protocol/transport"
	"github.com/status-im/status-go/protocol/urls"
	"github.com/status-im/status-go/protocol/verification"
	"github.com/status-im/status-go/services/ext/mailservers"
)

const (
	// defaultRequestTimeout is the default request timeout in seconds
	defaultRequestTimeout = 10
)

var (
	// ErrInvalidMailServerPeer is returned when it fails to parse enode from params.
	ErrInvalidMailServerPeer = errors.New("invalid mailServerPeer value")
	// ErrInvalidSymKeyID is returned when it fails to get a symmetric key.
	ErrInvalidSymKeyID = errors.New("invalid symKeyID value")
	// ErrInvalidPublicKey is returned when public key can't be extracted
	// from MailServer's nodeID.
	ErrInvalidPublicKey = errors.New("can't extract public key")
	// ErrPFSNotEnabled is returned when an endpoint PFS only is called but
	// PFS is disabled
	ErrPFSNotEnabled = errors.New("pfs not enabled")
)

// -----
// PAYLOADS
// -----

// MessagesRequest is a RequestMessages() request payload.
type MessagesRequest struct {
	// MailServerPeer is MailServer's enode address.
	MailServerPeer string `json:"mailServerPeer"`

	// From is a lower bound of time range (optional).
	// Default is 24 hours back from now.
	From uint32 `json:"from"`

	// To is a upper bound of time range (optional).
	// Default is now.
	To uint32 `json:"to"`

	// Limit determines the number of messages sent by the mail server
	// for the current paginated request
	Limit uint32 `json:"limit"`

	// Cursor is used as starting point for paginated requests
	Cursor string `json:"cursor"`

	// StoreCursor is used as starting point for WAKUV2 paginatedRequests
	StoreCursor *StoreRequestCursor `json:"storeCursor"`

	// Topic is a regular Whisper topic.
	// DEPRECATED
	Topic types.TopicType `json:"topic"`

	// Topics is a list of Whisper topics.
	Topics []types.TopicType `json:"topics"`

	// SymKeyID is an ID of a symmetric key to authenticate to MailServer.
	// It's derived from MailServer password.
	SymKeyID string `json:"symKeyID"`

	// Timeout is the time to live of the request specified in seconds.
	// Default is 10 seconds
	Timeout time.Duration `json:"timeout"`

	// Force ensures that requests will bypass enforced delay.
	Force bool `json:"force"`
}

type StoreRequestCursor struct {
	Digest       []byte  `json:"digest"`
	ReceivedTime float64 `json:"receivedTime"`
}

func (r *MessagesRequest) SetDefaults(now time.Time) {
	// set From and To defaults
	if r.To == 0 {
		r.To = uint32(now.UTC().Unix())
	}

	if r.From == 0 {
		oneDay := uint32(86400) // -24 hours
		if r.To < oneDay {
			r.From = 0
		} else {
			r.From = r.To - oneDay
		}
	}

	if r.Timeout == 0 {
		r.Timeout = defaultRequestTimeout
	}
}

// MessagesResponse is a response for requestMessages2 method.
type MessagesResponse struct {
	// Cursor from the response can be used to retrieve more messages
	// for the previous request.
	Cursor string `json:"cursor"`

	// Error indicates that something wrong happened when sending messages
	// to the requester.
	Error error `json:"error"`
}

// -----
// PUBLIC API
// -----

// PublicAPI extends whisper public API.
type PublicAPI struct {
	service  *Service
	eventSub mailservers.EnvelopeEventSubscriber
	log      log.Logger
}

// NewPublicAPI returns instance of the public API.
func NewPublicAPI(s *Service, eventSub mailservers.EnvelopeEventSubscriber) *PublicAPI {
	return &PublicAPI{
		service:  s,
		eventSub: eventSub,
		log:      log.New("package", "status-go/services/sshext.PublicAPI"),
	}
}

// RetryConfig specifies configuration for retries with timeout and max amount of retries.
type RetryConfig struct {
	BaseTimeout time.Duration
	// StepTimeout defines duration increase per each retry.
	StepTimeout time.Duration
	MaxRetries  int
}

func WaitForExpiredOrCompleted(requestID types.Hash, events chan types.EnvelopeEvent, timeout time.Duration) (*types.MailServerResponse, error) {
	expired := fmt.Errorf("request %x expired", requestID)
	after := time.NewTimer(timeout)
	defer after.Stop()
	for {
		var ev types.EnvelopeEvent
		select {
		case ev = <-events:
		case <-after.C:
			return nil, expired
		}
		if ev.Hash != requestID {
			continue
		}
		switch ev.Event {
		case types.EventMailServerRequestCompleted:
			data, ok := ev.Data.(*types.MailServerResponse)
			if ok {
				return data, nil
			}
			return nil, errors.New("invalid event data type")
		case types.EventMailServerRequestExpired:
			return nil, expired
		}
	}
}

type Author struct {
	PublicKey types.HexBytes `json:"publicKey"`
	Alias     string         `json:"alias"`
	Identicon string         `json:"identicon"`
}

type Metadata struct {
	DedupID      []byte         `json:"dedupId"`
	EncryptionID types.HexBytes `json:"encryptionId"`
	MessageID    types.HexBytes `json:"messageId"`
	Author       Author         `json:"author"`
}

func (api *PublicAPI) LeaveGroupChat(ctx Context, chatID string, remove bool) (*protocol.MessengerResponse, error) {
	return api.service.messenger.LeaveGroupChat(ctx, chatID, remove)
}

func (api *PublicAPI) CreateGroupChatWithMembers(ctx Context, name string, members []string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateGroupChatWithMembers(ctx, name, members)
}

func (api *PublicAPI) CreateGroupChatFromInvitation(name string, chatID string, adminPK string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateGroupChatFromInvitation(name, chatID, adminPK)
}

func (api *PublicAPI) AddMembersToGroupChat(ctx Context, chatID string, members []string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AddMembersToGroupChat(ctx, chatID, members)
}

func (api *PublicAPI) RemoveMemberFromGroupChat(ctx Context, chatID string, member string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RemoveMembersFromGroupChat(ctx, chatID, []string{member})
}

func (api *PublicAPI) RemoveMembersFromGroupChat(ctx Context, chatID string, members []string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RemoveMembersFromGroupChat(ctx, chatID, members)
}

func (api *PublicAPI) AddAdminsToGroupChat(ctx Context, chatID string, members []string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AddAdminsToGroupChat(ctx, chatID, members)
}

func (api *PublicAPI) ConfirmJoiningGroup(ctx context.Context, chatID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ConfirmJoiningGroup(ctx, chatID)
}

func (api *PublicAPI) ChangeGroupChatName(ctx Context, chatID string, name string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ChangeGroupChatName(ctx, chatID, name)
}

func (api *PublicAPI) SendGroupChatInvitationRequest(ctx Context, chatID string, adminPK string, message string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendGroupChatInvitationRequest(ctx, chatID, adminPK, message)
}

func (api *PublicAPI) GetGroupChatInvitations() ([]*protocol.GroupChatInvitation, error) {
	return api.service.messenger.GetGroupChatInvitations()
}

func (api *PublicAPI) SendGroupChatInvitationRejection(ctx Context, invitationRequestID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendGroupChatInvitationRejection(ctx, invitationRequestID)
}

func (api *PublicAPI) LoadFilters(parent context.Context, chats []*transport.Filter) ([]*transport.Filter, error) {
	return api.service.messenger.LoadFilters(chats)
}

func (api *PublicAPI) SaveChat(parent context.Context, chat *protocol.Chat) error {
	return api.service.messenger.SaveChat(chat)
}

func (api *PublicAPI) SaveMessages(parent context.Context, messages []*common.Message) error {
	return api.service.messenger.SaveMessages(messages)
}

func (api *PublicAPI) CreateOneToOneChat(parent context.Context, request *requests.CreateOneToOneChat) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateOneToOneChat(request)
}

func (api *PublicAPI) CreatePublicChat(parent context.Context, request *requests.CreatePublicChat) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreatePublicChat(request)
}

func (api *PublicAPI) CreateProfileChat(parent context.Context, request *requests.CreateProfileChat) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateProfileChat(request)
}

func (api *PublicAPI) Chats(parent context.Context) []*protocol.Chat {
	return api.service.messenger.Chats()
}

func (api *PublicAPI) ChatsPreview(parent context.Context) []*protocol.ChatPreview {
	return api.service.messenger.ChatsPreview()
}

func (api *PublicAPI) Chat(parent context.Context, chatID string) *protocol.Chat {
	return api.service.messenger.Chat(chatID)
}

func (api *PublicAPI) ActiveChats(parent context.Context) []*protocol.Chat {
	return api.service.messenger.ActiveChats()
}

func (api *PublicAPI) DeleteChat(parent context.Context, chatID string) error {
	return api.service.messenger.DeleteChat(chatID)
}

func (api *PublicAPI) MuteCommunityCategory(request *requests.MuteCategory) error {
	return api.service.messenger.SetMutePropertyOnChatsByCategory(request, true)
}

// Updates the lastOpenedAt key of a community
func (api *PublicAPI) CommunityUpdateLastOpenedAt(communityID string) (int64, error) {
	return api.service.messenger.CommunityUpdateLastOpenedAt(communityID)
}

func (api *PublicAPI) UnmuteCommunityCategory(communityID string, categoryID string) error {
	return api.service.messenger.SetMutePropertyOnChatsByCategory(&requests.MuteCategory{CommunityID: communityID, CategoryID: categoryID, MutedType: protocol.Unmuted}, false)
}

func (api *PublicAPI) MuteCommunityChats(request *requests.MuteCommunity) (time.Time, error) {
	return api.service.messenger.MuteAllCommunityChats(request)
}

func (api *PublicAPI) UnMuteCommunityChats(communityID string) (time.Time, error) {
	return api.service.messenger.UnMuteAllCommunityChats(communityID)
}

func (api *PublicAPI) MuteChatV2(parent context.Context, request *requests.MuteChat) (time.Time, error) {
	return api.service.messenger.MuteChatV2(request)
}

func (api *PublicAPI) MuteChat(parent context.Context, chatID string) (time.Time, error) {
	return api.service.messenger.MuteChat(&requests.MuteChat{ChatID: chatID, MutedType: 0})
}

func (api *PublicAPI) UnmuteChat(parent context.Context, chatID string) error {
	return api.service.messenger.UnmuteChat(chatID)
}

func (api *PublicAPI) BlockContact(ctx context.Context, contactID string) (*protocol.MessengerResponse, error) {
	api.log.Info("blocking contact", "contact", contactID)
	return api.service.messenger.BlockContact(ctx, contactID, false)
}

// This function is the same as the one above, but used only on the desktop side, since at the end it doesn't set
// `Added` flag to `false`, but only `Blocked` to `true`
func (api *PublicAPI) BlockContactDesktop(ctx context.Context, contactID string) (*protocol.MessengerResponse, error) {
	api.log.Info("blocking contact", "contact", contactID)
	return api.service.messenger.BlockContactDesktop(ctx, contactID)
}

func (api *PublicAPI) UnblockContact(parent context.Context, contactID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.UnblockContact(contactID)
}

func (api *PublicAPI) Contacts(parent context.Context) []*protocol.Contact {
	return api.service.messenger.Contacts()
}

func (api *PublicAPI) GetContactByID(parent context.Context, id string) *protocol.Contact {
	return api.service.messenger.GetContactByID(id)
}

func (api *PublicAPI) RequestContactInfoFromMailserver(pubkey string) (*protocol.Contact, error) {
	return api.service.messenger.FetchContact(pubkey, true)
}

func (api *PublicAPI) RemoveFilters(parent context.Context, chats []*transport.Filter) error {
	return api.service.messenger.RemoveFilters(chats)
}

// EnableInstallation enables an installation for multi-device sync.
func (api *PublicAPI) EnableInstallation(installationID string) error {
	return api.service.messenger.EnableInstallation(installationID)
}

// DisableInstallation disables an installation for multi-device sync.
func (api *PublicAPI) DisableInstallation(installationID string) error {
	return api.service.messenger.DisableInstallation(installationID)
}

// GetOurInstallations returns all the installations available given an identity
func (api *PublicAPI) GetOurInstallations() []*multidevice.Installation {
	return api.service.messenger.Installations()
}

// SetInstallationMetadata sets the metadata for our own installation
func (api *PublicAPI) SetInstallationMetadata(installationID string, data *multidevice.InstallationMetadata) error {
	return api.service.messenger.SetInstallationMetadata(installationID, data)
}

// SetInstallationName sets the only the name in metadata for a given installation
func (api *PublicAPI) SetInstallationName(installationID string, name string) error {
	return api.service.messenger.SetInstallationName(installationID, name)
}

// Communities returns a list of communities that are stored
func (api *PublicAPI) Communities(parent context.Context) ([]*communities.Community, error) {
	return api.service.messenger.Communities()
}

// JoinedCommunities returns a list of communities that the user has joined
func (api *PublicAPI) JoinedCommunities(parent context.Context) ([]*communities.Community, error) {
	return api.service.messenger.JoinedCommunities()
}

// CommunityTags return the list of possible community tags
func (api *PublicAPI) CommunityTags(parent context.Context) map[string]string {
	return requests.TagsEmojies
}

// CuratedCommunities returns the list of curated communities stored in the smart contract. If a community is
// already known by the node, its description will be returned and and will asynchronously retrieve the
// description for the communities it does not know
func (api *PublicAPI) CuratedCommunities(parent context.Context) (*communities.KnownCommunitiesResponse, error) {
	return api.service.messenger.CuratedCommunities()
}

// SpectateCommunity spectates community with the given ID
// Meaning user is only a spectator, not a member
func (api *PublicAPI) SpectateCommunity(parent context.Context, communityID types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SpectateCommunity(communityID)
}

// JoinCommunity joins a community with the given ID
func (api *PublicAPI) JoinCommunity(parent context.Context, communityID types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.JoinCommunity(parent, communityID, false)
}

// LeaveCommunity leaves a commuity with the given ID
func (api *PublicAPI) LeaveCommunity(ctx context.Context, communityID types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.LeaveCommunity(communityID)
}

// CreateCommunity creates a new community with the provided description
func (api *PublicAPI) CreateCommunity(request *requests.CreateCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateCommunity(request, true)
}

// EditCommunity edits an existing community with the provided description
func (api *PublicAPI) EditCommunity(request *requests.EditCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.EditCommunity(request)
}

// RemovePrivateKey removes the private key of the community with given ID
func (api *PublicAPI) RemovePrivateKey(id types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RemovePrivateKey(id)
}

// Sets the community shard for a community and updates all active filters for the community
func (api *PublicAPI) SetCommunityShard(request *requests.SetCommunityShard) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SetCommunityShard(request)
}

// ExportCommunity exports the private key of the community with given ID
func (api *PublicAPI) ExportCommunity(id types.HexBytes) (types.HexBytes, error) {
	key, err := api.service.messenger.ExportCommunity(id)
	if err != nil {
		return nil, err
	}
	return crypto.FromECDSA(key), nil
}

// ImportCommunity imports a community with the given private key in hex
func (api *PublicAPI) ImportCommunity(ctx context.Context, hexPrivateKey string) (*protocol.MessengerResponse, error) {
	// Strip the 0x from the beginning
	privateKey, err := crypto.HexToECDSA(hexPrivateKey[2:])
	if err != nil {
		return nil, err
	}
	return api.service.messenger.ImportCommunity(ctx, privateKey)
}

// GetCommunityPublicKeyFromPrivateKey gets the community's public key from its private key
func (api *PublicAPI) GetCommunityPublicKeyFromPrivateKey(ctx context.Context, hexPrivateKey string) string {
	publicKey := protocol.GetCommunityIDFromKey(hexPrivateKey)
	return publicKey
}

// Get community members contact list for provided wallet addresses
func (api *PublicAPI) GetCommunityMembersForWalletAddresses(communityID types.HexBytes, chainID uint64) (map[string]*protocol.Contact, error) {
	return api.service.messenger.GetCommunityMembersForWalletAddresses(communityID, chainID)
}

// Speeds up importing messages from archives
func (api *PublicAPI) SpeedupArchivesImport(ctx context.Context) {
	api.service.messenger.SpeedupArchivesImport()
}

// Slows down importing messages from archives
func (api *PublicAPI) SlowdownArchivesImport(ctx context.Context) {
	api.service.messenger.SlowdownArchivesImport()
}

// CreateCommunityChat creates a community chat in the given community
func (api *PublicAPI) CreateCommunityChat(communityID types.HexBytes, c *protobuf.CommunityChat) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateCommunityChat(communityID, c)
}

// EditCommunityChat edits a community chat in the given community
func (api *PublicAPI) EditCommunityChat(communityID types.HexBytes, chatID string, c *protobuf.CommunityChat) (*protocol.MessengerResponse, error) {
	return api.service.messenger.EditCommunityChat(communityID, chatID, c)
}

// DeleteCommunityChat deletes a community chat in the given community
func (api *PublicAPI) DeleteCommunityChat(communityID types.HexBytes, chatID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeleteCommunityChat(communityID, chatID)
}

// ShareCommunity share the community with a set of users
func (api *PublicAPI) ShareCommunity(request *requests.ShareCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ShareCommunity(request)
}

// ShareImageMessage share the selected chat image with a set of users
func (api *PublicAPI) ShareImageMessage(request *requests.ShareImageMessage) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ShareImageMessage(request)
}

// RemoveUserFromCommunity removes the user with pk from the community with ID
func (api *PublicAPI) RemoveUserFromCommunity(communityID types.HexBytes, userPublicKey string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RemoveUserFromCommunity(communityID, userPublicKey)
}

// SetCommunityMuted sets the community's muted value
func (api *PublicAPI) SetCommunityMuted(request *requests.MuteCommunity) error {
	return api.service.messenger.SetMuted(request)
}

// BanUserFromCommunity removes the user with pk from the community with ID
func (api *PublicAPI) BanUserFromCommunity(ctx context.Context, request *requests.BanUserFromCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.BanUserFromCommunity(ctx, request)
}

// UnbanUserFromCommunity removes the user's pk from the community ban list
func (api *PublicAPI) UnbanUserFromCommunity(request *requests.UnbanUserFromCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.UnbanUserFromCommunity(request)
}

func (api *PublicAPI) AddRoleToMember(request *requests.AddRoleToMember) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AddRoleToMember(request)
}

func (api *PublicAPI) RemoveRoleFromMember(request *requests.RemoveRoleFromMember) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RemoveRoleFromMember(request)
}

func (api *PublicAPI) CreateCommunityTokenPermission(request *requests.CreateCommunityTokenPermission) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateCommunityTokenPermission(request)
}

// ReevaluateCommunityMembersPermissions reevaluates community members permissions
func (api *PublicAPI) ReevaluateCommunityMembersPermissions(request *requests.ReevaluateCommunityMembersPermissions) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ReevaluateCommunityMembersPermissions(request)
}

func (api *PublicAPI) DeleteCommunityTokenPermission(request *requests.DeleteCommunityTokenPermission) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeleteCommunityTokenPermission(request)
}

func (api *PublicAPI) EditCommunityTokenPermission(request *requests.EditCommunityTokenPermission) (*protocol.MessengerResponse, error) {
	return api.service.messenger.EditCommunityTokenPermission(request)
}

// MyPendingRequestsToJoin returns the pending requests for the logged in user
func (api *PublicAPI) MyPendingRequestsToJoin() ([]*communities.RequestToJoin, error) {
	return api.service.messenger.MyPendingRequestsToJoin()
}

// MyCanceledRequestsToJoin returns the pending requests for the logged in user
func (api *PublicAPI) MyCanceledRequestsToJoin() ([]*communities.RequestToJoin, error) {
	return api.service.messenger.MyCanceledRequestsToJoin()
}

// PendingRequestsToJoinForCommunity returns the pending requests to join for a given community
func (api *PublicAPI) PendingRequestsToJoinForCommunity(id types.HexBytes) ([]*communities.RequestToJoin, error) {
	return api.service.messenger.PendingRequestsToJoinForCommunity(id)
}

// DeclinedRequestsToJoinForCommunity returns the declined requests to join for a given community
func (api *PublicAPI) DeclinedRequestsToJoinForCommunity(id types.HexBytes) ([]*communities.RequestToJoin, error) {
	return api.service.messenger.DeclinedRequestsToJoinForCommunity(id)
}

// CanceledRequestsToJoinForCommunity returns the declined requests to join for a given community
func (api *PublicAPI) CanceledRequestsToJoinForCommunity(id types.HexBytes) ([]*communities.RequestToJoin, error) {
	return api.service.messenger.CanceledRequestsToJoinForCommunity(id)
}

// AllNonApprovedCommunitiesRequestsToJoin returns the all non-approved requests to join for all communities
func (api *PublicAPI) AllNonApprovedCommunitiesRequestsToJoin() ([]*communities.RequestToJoin, error) {
	return api.service.messenger.AllNonApprovedCommunitiesRequestsToJoin()
}

// Generates a single hash for each address that needs to be revealed to a community.
// Each hash needs to be signed.
// The order of retuned hashes corresponds to the order of addresses in addressesToReveal.
func (api *PublicAPI) GenerateJoiningCommunityRequestsForSigning(memberPubKey string, communityID types.HexBytes, addressesToReveal []string) ([]account.SignParams, error) {
	return api.service.messenger.GenerateJoiningCommunityRequestsForSigning(memberPubKey, communityID, addressesToReveal)
}

// Generates a single hash for each address that needs to be revealed to a community.
// Each hash needs to be signed.
// The order of retuned hashes corresponds to the order of addresses in addressesToReveal.
func (api *PublicAPI) GenerateEditCommunityRequestsForSigning(memberPubKey string, communityID types.HexBytes, addressesToReveal []string) ([]account.SignParams, error) {
	return api.service.messenger.GenerateEditCommunityRequestsForSigning(memberPubKey, communityID, addressesToReveal)
}

// Signs the provided messages with the provided accounts and password.
// Provided accounts must not belong to a keypair that is migrated to a keycard.
// Otherwise, the signing will fail, cause such accounts should be signed with a keycard.
func (api *PublicAPI) SignData(signParams []account.SignParams) ([]string, error) {
	return api.service.messenger.SignData(signParams)
}

// CancelRequestToJoinCommunity accepts a pending request to join a community
func (api *PublicAPI) CancelRequestToJoinCommunity(ctx context.Context, request *requests.CancelRequestToJoinCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CancelRequestToJoinCommunity(ctx, request)
}

// AcceptRequestToJoinCommunity accepts a pending request to join a community
func (api *PublicAPI) AcceptRequestToJoinCommunity(request *requests.AcceptRequestToJoinCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AcceptRequestToJoinCommunity(request)
}

// DeclineRequestToJoinCommunity accepts a pending request to join a community
func (api *PublicAPI) DeclineRequestToJoinCommunity(request *requests.DeclineRequestToJoinCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeclineRequestToJoinCommunity(request)
}

// RequestToJoinCommunity requests to join a particular community
func (api *PublicAPI) RequestToJoinCommunity(request *requests.RequestToJoinCommunity) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RequestToJoinCommunity(request)
}

// EditSharedAddressesForCommunity edits the addresses that are shared with the owner of the community
func (api *PublicAPI) EditSharedAddressesForCommunity(request *requests.EditSharedAddresses) (*protocol.MessengerResponse, error) {
	return api.service.messenger.EditSharedAddressesForCommunity(request)
}

// GetRevealedAccounts gets the revealed addresses for a member in a community
func (api *PublicAPI) GetRevealedAccounts(communityID types.HexBytes, memberPk string) ([]*protobuf.RevealedAccount, error) {
	return api.service.messenger.GetRevealedAccounts(communityID, memberPk)
}

// GetRevealedAccountsForAllMembers gets the revealed addresses for all the members of a community
func (api *PublicAPI) GetRevealedAccountsForAllMembers(communityID types.HexBytes) (map[string][]*protobuf.RevealedAccount, error) {
	return api.service.messenger.GetRevealedAccountsForAllMembers(communityID)
}

// CheckAndClearPendingRequestToJoinCommunity to delete pending request to join a community which are older than 7 days
func (api *PublicAPI) CheckAndDeletePendingRequestToJoinCommunity(ctx context.Context) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CheckAndDeletePendingRequestToJoinCommunity(ctx, true)
}

// CreateCommunityCategory creates a category within a particular community
func (api *PublicAPI) CreateCommunityCategory(request *requests.CreateCommunityCategory) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateCommunityCategory(request)
}

// ReorderCommunityCategories is used to change the order of the categories of a community
func (api *PublicAPI) ReorderCommunityCategories(request *requests.ReorderCommunityCategories) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ReorderCommunityCategories(request)
}

// ReorderCommunityChat allows changing the order of the chat or switching its category
func (api *PublicAPI) ReorderCommunityChat(request *requests.ReorderCommunityChat) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ReorderCommunityChat(request)
}

// EditCommunityCategory modifies a category within a particular community
func (api *PublicAPI) EditCommunityCategory(request *requests.EditCommunityCategory) (*protocol.MessengerResponse, error) {
	return api.service.messenger.EditCommunityCategory(request)
}

// DeleteCommunityCategory deletes a category within a particular community and removes this category from any chat that has it
func (api *PublicAPI) DeleteCommunityCategory(request *requests.DeleteCommunityCategory) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeleteCommunityCategory(request)
}

func (api *PublicAPI) PromoteSelfToControlNode(communityID types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.PromoteSelfToControlNode(communityID)
}

type ApplicationMessagesResponse struct {
	Messages []*common.Message `json:"messages"`
	Cursor   string            `json:"cursor"`
}

type MarkMessageSeenResponse struct {
	Count                       uint64                                 `json:"count"`
	CountWithMentions           uint64                                 `json:"countWithMentions"`
	ActivityCenterNotifications []*protocol.ActivityCenterNotification `json:"activityCenterNotifications,omitempty"`
}

type ApplicationPinnedMessagesResponse struct {
	PinnedMessages []*common.PinnedMessage `json:"pinnedMessages"`
	Cursor         string                  `json:"cursor"`
}

type ApplicationStatusUpdatesResponse struct {
	StatusUpdates []protocol.UserStatus `json:"statusUpdates"`
}

type ApplicationSwitcherCardsResponse struct {
	SwitcherCards []protocol.SwitcherCard `json:"switcherCards"`
}

func (api *PublicAPI) ChatMessages(chatID, cursor string, limit int) (*ApplicationMessagesResponse, error) {
	messages, cursor, err := api.service.messenger.MessageByChatID(chatID, cursor, limit)
	if err != nil {
		return nil, err
	}

	return &ApplicationMessagesResponse{
		Messages: messages,
		Cursor:   cursor,
	}, nil
}

func (api *PublicAPI) MessageByMessageID(messageID string) (*common.Message, error) {
	return api.service.messenger.MessageByID(messageID)
}

func (api *PublicAPI) FirstUnseenMessageID(chatID string) (string, error) {
	return api.service.messenger.FirstUnseenMessageID(chatID)
}

func (api *PublicAPI) AllMessagesFromChatWhichMatchTerm(chatID, searchTerm string, caseSensitive bool) (*ApplicationMessagesResponse, error) {
	messages, err := api.service.messenger.AllMessageByChatIDWhichMatchTerm(chatID, searchTerm, caseSensitive)
	if err != nil {
		return nil, err
	}

	return &ApplicationMessagesResponse{
		Messages: messages,
	}, nil
}

func (api *PublicAPI) AllMessagesFromChatsAndCommunitiesWhichMatchTerm(communityIds []string, chatIds []string, searchTerm string, caseSensitive bool) (*ApplicationMessagesResponse, error) {
	messages, err := api.service.messenger.AllMessagesFromChatsAndCommunitiesWhichMatchTerm(communityIds, chatIds, searchTerm, caseSensitive)
	if err != nil {
		return nil, err
	}

	return &ApplicationMessagesResponse{
		Messages: messages,
	}, nil
}

func (api *PublicAPI) ChatPinnedMessages(chatID, cursor string, limit int) (*ApplicationPinnedMessagesResponse, error) {
	pinnedMessages, cursor, err := api.service.messenger.PinnedMessageByChatID(chatID, cursor, limit)
	if err != nil {
		return nil, err
	}

	return &ApplicationPinnedMessagesResponse{
		PinnedMessages: pinnedMessages,
		Cursor:         cursor,
	}, nil
}

func (api *PublicAPI) StatusUpdates() (*ApplicationStatusUpdatesResponse, error) {
	statusUpdates, err := api.service.messenger.StatusUpdates()
	if err != nil {
		return nil, err
	}

	return &ApplicationStatusUpdatesResponse{
		StatusUpdates: statusUpdates,
	}, nil
}

func (api *PublicAPI) UpsertSwitcherCard(request *requests.UpsertSwitcherCard) error {
	return api.service.messenger.UpsertSwitcherCard(request)
}

func (api *PublicAPI) DeleteSwitcherCard(id string) error {
	return api.service.messenger.DeleteSwitcherCard(id)
}

func (api *PublicAPI) SwitcherCards() (*ApplicationSwitcherCardsResponse, error) {
	switcherCards, err := api.service.messenger.SwitcherCards()
	if err != nil {
		return nil, err
	}

	return &ApplicationSwitcherCardsResponse{
		SwitcherCards: switcherCards,
	}, nil
}

func (api *PublicAPI) StartMessenger() (*protocol.MessengerResponse, error) {
	return api.service.StartMessenger()
}

func (api *PublicAPI) SetUserStatus(ctx context.Context, status int, customText string) error {
	return api.service.messenger.SetUserStatus(ctx, status, customText)
}

func (api *PublicAPI) DeleteMessage(id string) error {
	return api.service.messenger.DeleteMessage(id)
}

func (api *PublicAPI) DeleteMessagesByChatID(id string) error {
	return api.service.messenger.DeleteMessagesByChatID(id)
}

// Deprecated: Use MarkMessagesRead instead
func (api *PublicAPI) MarkMessagesSeen(chatID string, ids []string) (*MarkMessageSeenResponse, error) {
	count, withMentions, notifications, err := api.service.messenger.MarkMessagesSeen(chatID, ids)
	if err != nil {
		return nil, err
	}

	response := &MarkMessageSeenResponse{
		Count:                       count,
		CountWithMentions:           withMentions,
		ActivityCenterNotifications: notifications,
	}
	return response, nil
}

func (api *PublicAPI) MarkMessagesRead(chatID string, ids []string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.MarkMessagesRead(chatID, ids)
}

func (api *PublicAPI) MarkMessageAsUnread(chatID string, messageID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.MarkMessageAsUnread(chatID, messageID)
}

func (api *PublicAPI) MarkAllRead(ctx context.Context, chatID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.MarkAllRead(ctx, chatID)
}

func (api *PublicAPI) DismissActivityCenterNotificationsByCommunity(ctx context.Context, request *requests.DismissCommunityNotifications) error {
	return api.service.messenger.DismissActivityCenterNotificationsByCommunity(ctx, request)
}

func (api *PublicAPI) MarkAllReadInCommunity(ctx context.Context, communityID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.MarkAllReadInCommunity(ctx, communityID)
}

func (api *PublicAPI) SendContactRequest(ctx context.Context, request *requests.SendContactRequest) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendContactRequest(ctx, request)
}

func (api *PublicAPI) AddContact(ctx context.Context, request *requests.AddContact) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AddContact(ctx, request)
}

func (api *PublicAPI) AcceptContactRequest(ctx context.Context, request *requests.AcceptContactRequest) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AcceptContactRequest(ctx, request)
}

func (api *PublicAPI) DeclineContactRequest(ctx context.Context, request *requests.DeclineContactRequest) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeclineContactRequest(ctx, request)
}

func (api *PublicAPI) AcceptLatestContactRequestForContact(ctx context.Context, request *requests.AcceptLatestContactRequestForContact) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AcceptLatestContactRequestForContact(ctx, request)
}

func (api *PublicAPI) DismissLatestContactRequestForContact(ctx context.Context, request *requests.DismissLatestContactRequestForContact) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DismissLatestContactRequestForContact(ctx, request)
}

func (api *PublicAPI) RetractContactRequest(ctx context.Context, request *requests.RetractContactRequest) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RetractContactRequest(request)
}

func (api *PublicAPI) RemoveContact(ctx context.Context, pubKey string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RemoveContact(ctx, pubKey)
}

func (api *PublicAPI) SetContactLocalNickname(ctx context.Context, request *requests.SetContactLocalNickname) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SetContactLocalNickname(request)
}

func (api *PublicAPI) ClearHistory(request *requests.ClearHistory) (*protocol.MessengerResponse, error) {
	return api.service.messenger.ClearHistory(request)
}

func (api *PublicAPI) DeactivateChat(request *requests.DeactivateChat) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeactivateChat(request)
}

func (api *PublicAPI) UpdateMessageOutgoingStatus(id, newOutgoingStatus string) error {
	return api.service.messenger.UpdateMessageOutgoingStatus(id, newOutgoingStatus)
}

func (api *PublicAPI) SendChatMessage(ctx context.Context, message *common.Message) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendChatMessage(ctx, message)
}

func (api *PublicAPI) ReSendChatMessage(ctx context.Context, messageID string) error {
	return api.service.messenger.ReSendChatMessage(ctx, messageID)
}

func (api *PublicAPI) SendChatMessages(ctx context.Context, messages []*common.Message) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendChatMessages(ctx, messages)
}

func (api *PublicAPI) SendOneToOneMessage(request *requests.SendOneToOneMessage) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendOneToOneMessage(request)
}

func (api *PublicAPI) SendGroupChatMessage(request *requests.SendGroupChatMessage) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendGroupChatMessage(request)
}

func (api *PublicAPI) EditMessage(ctx context.Context, request *requests.EditMessage) (*protocol.MessengerResponse, error) {
	return api.service.messenger.EditMessage(ctx, request)
}

func (api *PublicAPI) DeleteMessageAndSend(ctx context.Context, messageID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeleteMessageAndSend(ctx, messageID)
}

func (api *PublicAPI) DeleteMessageForMeAndSync(ctx context.Context, chatID string, messageID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeleteMessageForMeAndSync(ctx, chatID, messageID)
}

func (api *PublicAPI) SendPinMessage(ctx context.Context, message *common.PinMessage) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendPinMessage(ctx, message)
}

func (api *PublicAPI) RequestTransaction(ctx context.Context, chatID, value, contract, address string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RequestTransaction(ctx, chatID, value, contract, address)
}

func (api *PublicAPI) RequestAddressForTransaction(ctx context.Context, chatID, from, value, contract string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RequestAddressForTransaction(ctx, chatID, from, value, contract)
}

func (api *PublicAPI) DeclineRequestAddressForTransaction(ctx context.Context, messageID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeclineRequestAddressForTransaction(ctx, messageID)
}

func (api *PublicAPI) DeclineRequestTransaction(ctx context.Context, messageID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeclineRequestTransaction(ctx, messageID)
}

func (api *PublicAPI) AcceptRequestAddressForTransaction(ctx context.Context, messageID, address string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AcceptRequestAddressForTransaction(ctx, messageID, address)
}

func (api *PublicAPI) SendTransaction(ctx context.Context, chatID, value, contract, transactionHash string, signature types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendTransaction(ctx, chatID, value, contract, transactionHash, signature)
}

func (api *PublicAPI) AcceptRequestTransaction(ctx context.Context, transactionHash, messageID string, signature types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AcceptRequestTransaction(ctx, transactionHash, messageID, signature)
}

func (api *PublicAPI) SendContactUpdates(ctx context.Context, name, picture string) error {
	return api.service.messenger.SendContactUpdates(ctx, name, picture)
}

func (api *PublicAPI) SendContactUpdate(ctx context.Context, contactID, name, picture string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendContactUpdate(ctx, contactID, name, picture)
}

func (api *PublicAPI) SetDisplayName(ctx context.Context, displayName string) error {
	return api.service.messenger.SetDisplayName(displayName)
}

func (api *PublicAPI) SetBio(ctx context.Context, bio string) error {
	return api.service.messenger.SetBio(bio)
}

func (api *PublicAPI) MarkAsTrusted(ctx context.Context, contactID string) error {
	return api.service.messenger.MarkAsTrusted(ctx, contactID)
}

func (api *PublicAPI) MarkAsUntrustworthy(ctx context.Context, contactID string) error {
	return api.service.messenger.MarkAsUntrustworthy(ctx, contactID)
}

func (api *PublicAPI) RemoveTrustStatus(ctx context.Context, contactID string) error {
	return api.service.messenger.RemoveTrustStatus(ctx, contactID)
}

func (api *PublicAPI) GetTrustStatus(ctx context.Context, contactID string) (verification.TrustStatus, error) {
	return api.service.messenger.GetTrustStatus(contactID)
}

func (api *PublicAPI) GetLatestVerificationRequestFrom(ctx context.Context, contactID string) (*verification.Request, error) {
	return api.service.messenger.GetLatestVerificationRequestFrom(contactID)
}

func (api *PublicAPI) SendContactVerificationRequest(ctx context.Context, contactID string, challenge string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendContactVerificationRequest(ctx, contactID, challenge)
}

func (api *PublicAPI) GetReceivedVerificationRequests(ctx context.Context) ([]*verification.Request, error) {
	return api.service.messenger.GetReceivedVerificationRequests(ctx)
}

func (api *PublicAPI) GetVerificationRequestSentTo(ctx context.Context, contactID string) (*verification.Request, error) {
	return api.service.messenger.GetVerificationRequestSentTo(ctx, contactID)
}

func (api *PublicAPI) CancelVerificationRequest(ctx context.Context, id string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CancelVerificationRequest(ctx, id)
}

func (api *PublicAPI) AcceptContactVerificationRequest(ctx context.Context, id string, response string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.AcceptContactVerificationRequest(ctx, id, response)
}

func (api *PublicAPI) DeclineContactVerificationRequest(ctx context.Context, id string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.DeclineContactVerificationRequest(ctx, id)
}

func (api *PublicAPI) VerifiedTrusted(ctx context.Context, request *requests.VerifiedTrusted) (*protocol.MessengerResponse, error) {
	return api.service.messenger.VerifiedTrusted(ctx, request)
}

func (api *PublicAPI) VerifiedUntrustworthy(ctx context.Context, request *requests.VerifiedUntrustworthy) (*protocol.MessengerResponse, error) {
	return api.service.messenger.VerifiedUntrustworthy(ctx, request)
}

func (api *PublicAPI) SendPairInstallation(ctx context.Context) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendPairInstallation(ctx, nil)
}

func (api *PublicAPI) SyncDevices(ctx context.Context, name, picture string) error {
	return api.service.messenger.SyncDevices(ctx, name, picture, nil)
}

func (api *PublicAPI) AddBookmark(ctx context.Context, bookmark browsers.Bookmark) error {
	return api.service.messenger.AddBookmark(ctx, bookmark)
}

func (api *PublicAPI) AddBrowser(ctx context.Context, browser browsers.Browser) error {
	return api.service.messenger.AddBrowser(ctx, browser)
}

func (api *PublicAPI) GetBrowsers(ctx context.Context) (browsers []*browsers.Browser, err error) {
	return api.service.messenger.GetBrowsers(ctx)
}

func (api *PublicAPI) DeleteBrowser(ctx context.Context, id string) error {
	return api.service.messenger.DeleteBrowser(ctx, id)
}

func (api *PublicAPI) RemoveBookmark(ctx context.Context, url string) error {
	return api.service.messenger.RemoveBookmark(ctx, url)
}

func (api *PublicAPI) UpdateBookmark(ctx context.Context, oldURL string, bookmark browsers.Bookmark) error {
	return api.service.messenger.UpdateBookmark(ctx, oldURL, bookmark)
}

func (api *PublicAPI) SignMessageWithChatKey(ctx context.Context, message string) (types.HexBytes, error) {
	return api.service.messenger.SignMessage(message)
}

func (api *PublicAPI) CreateCommunityTokenDeploymentSignature(ctx context.Context, chainID uint64, addressFrom string, communityID string) (types.HexBytes, error) {
	return api.service.messenger.CreateCommunityTokenDeploymentSignature(ctx, chainID, addressFrom, communityID)
}

// wallet connect session apis
func (api *PublicAPI) AddWalletConnectSession(ctx context.Context, request *requests.AddWalletConnectSession) error {
	return api.service.messenger.AddWalletConnectSession(request)
}

func (api *PublicAPI) GetWalletConnectSession(ctx context.Context) ([]protocol.WalletConnectSession, error) {
	return api.service.messenger.GetWalletConnectSession()
}

func (api *PublicAPI) DestroyWalletConnectSession(ctx context.Context, PeerID string) error {
	return api.service.messenger.DestroyWalletConnectSession(PeerID)
}

// Saved Addresses APIs
func (api *PublicAPI) UpsertSavedAddress(ctx context.Context, sa wallet.SavedAddress) error {
	return api.service.messenger.UpsertSavedAddress(ctx, sa)
}

func (api *PublicAPI) DeleteSavedAddress(ctx context.Context, address ethcommon.Address, isTest bool) error {
	return api.service.messenger.DeleteSavedAddress(ctx, address, isTest)
}

func (api *PublicAPI) GetSavedAddresses(ctx context.Context) ([]*wallet.SavedAddress, error) {
	return api.service.messenger.GetSavedAddresses(ctx)
}

// PushNotifications server endpoints
func (api *PublicAPI) StartPushNotificationsServer() error {
	err := api.service.accountsDB.SaveSettingField(settings.PushNotificationsServerEnabled, true)
	if err != nil {
		return err
	}

	return api.service.messenger.StartPushNotificationsServer()
}

func (api *PublicAPI) StopPushNotificationsServer() error {
	err := api.service.accountsDB.SaveSettingField(settings.PushNotificationsServerEnabled, false)
	if err != nil {
		return err
	}

	return api.service.messenger.StopPushNotificationsServer()
}

// PushNotification client endpoints

func (api *PublicAPI) RegisterForPushNotifications(ctx context.Context, deviceToken string, apnTopic string, tokenType protobuf.PushNotificationRegistration_TokenType) error {
	return api.service.messenger.RegisterForPushNotifications(ctx, deviceToken, apnTopic, tokenType)
}

func (api *PublicAPI) UnregisterFromPushNotifications(ctx context.Context) error {
	return api.service.messenger.UnregisterFromPushNotifications(ctx)
}

func (api *PublicAPI) DisableSendingNotifications(ctx context.Context) error {
	err := api.service.accountsDB.SaveSettingField(settings.SendPushNotifications, false)
	if err != nil {
		return err
	}

	return api.service.messenger.DisableSendingPushNotifications()
}

func (api *PublicAPI) EnableSendingNotifications(ctx context.Context) error {
	err := api.service.accountsDB.SaveSettingField(settings.SendPushNotifications, true)
	if err != nil {
		return err
	}
	return api.service.messenger.EnableSendingPushNotifications()
}

func (api *PublicAPI) EnablePushNotificationsFromContactsOnly(ctx context.Context) error {
	err := api.service.accountsDB.SaveSettingField(settings.PushNotificationsFromContactsOnly, true)
	if err != nil {
		return err
	}
	return api.service.messenger.EnablePushNotificationsFromContactsOnly()
}

func (api *PublicAPI) DisablePushNotificationsFromContactsOnly(ctx context.Context) error {
	err := api.service.accountsDB.SaveSettingField(settings.PushNotificationsFromContactsOnly, false)
	if err != nil {
		return err
	}
	return api.service.messenger.DisablePushNotificationsFromContactsOnly()
}

func (api *PublicAPI) EnablePushNotificationsBlockMentions(ctx context.Context) error {
	err := api.service.accountsDB.SaveSettingField(settings.PushNotificationsBlockMentions, true)
	if err != nil {
		return err
	}
	return api.service.messenger.EnablePushNotificationsBlockMentions()
}

func (api *PublicAPI) DisablePushNotificationsBlockMentions(ctx context.Context) error {
	err := api.service.accountsDB.SaveSettingField(settings.PushNotificationsBlockMentions, false)
	if err != nil {
		return err
	}
	return api.service.messenger.DisablePushNotificationsBlockMentions()
}

func (api *PublicAPI) AddPushNotificationsServer(ctx context.Context, publicKeyBytes types.HexBytes) error {
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return err
	}

	// this is coming from a user, so it has to be a custom server
	return api.service.messenger.AddPushNotificationsServer(ctx, publicKey, pushnotificationclient.ServerTypeCustom)
}

func (api *PublicAPI) RemovePushNotificationServer(ctx context.Context, publicKeyBytes types.HexBytes) error {
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return err
	}

	return api.service.messenger.RemovePushNotificationServer(ctx, publicKey)
}

func (api *PublicAPI) GetPushNotificationsServers() ([]*pushnotificationclient.PushNotificationServer, error) {
	return api.service.messenger.GetPushNotificationsServers()
}

func (api *PublicAPI) RegisteredForPushNotifications() (bool, error) {
	return api.service.messenger.RegisteredForPushNotifications()
}

// Emoji

func (api *PublicAPI) SendEmojiReaction(ctx context.Context, chatID, messageID string, emojiID protobuf.EmojiReaction_Type) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendEmojiReaction(ctx, chatID, messageID, emojiID)
}

func (api *PublicAPI) SendEmojiReactionRetraction(ctx context.Context, emojiReactionID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.SendEmojiReactionRetraction(ctx, emojiReactionID)
}

func (api *PublicAPI) EmojiReactionsByChatID(chatID string, cursor string, limit int) ([]*protocol.EmojiReaction, error) {
	return api.service.messenger.EmojiReactionsByChatID(chatID, cursor, limit)
}

func (api *PublicAPI) EmojiReactionsByChatIDMessageID(chatID string, messageID string) ([]*protocol.EmojiReaction, error) {
	return api.service.messenger.EmojiReactionsByChatIDMessageID(chatID, messageID)
}

func (api *PublicAPI) GetLinkPreviewWhitelist() []urls.Site {
	return urls.LinkPreviewWhitelist()
}

func (api *PublicAPI) GetLinkPreviewData(link string) (previewData urls.LinkPreviewData, err error) {
	return urls.GetLinkPreviewData(link)
}

// GetTextURLsToUnfurl parses text and returns a deduplicated and (somewhat) normalized
// slice of URLs. The returned URLs can be used as cache keys by clients.
// For each URL there's a corresponding metadata which should be used as to plan the unfurling.
func (api *PublicAPI) GetTextURLsToUnfurl(text string) *protocol.URLsUnfurlPlan {
	return api.service.messenger.GetTextURLsToUnfurl(text)
}

// Deprecated: GetTextURLs is deprecated in favor of more generic GetTextURLsToUnfurl.
//
// GetTextURLs parses text and returns a deduplicated and (somewhat) normalized
// slice of URLs. The returned URLs can be used as cache keys by clients.
func (api *PublicAPI) GetTextURLs(text string) []string {
	return api.service.messenger.GetURLs(text)
}

// UnfurlURLs uses a best-effort approach to unfurl each URL. Failed URLs will
// be removed from the response.
//
// This endpoint expects the client to send URLs normalized by GetTextURLs.
func (api *PublicAPI) UnfurlURLs(urls []string) (protocol.UnfurlURLsResponse, error) {
	return api.service.messenger.UnfurlURLs(nil, urls)
}

func (api *PublicAPI) EnsVerified(pk, ensName string) error {
	return api.service.messenger.ENSVerified(pk, ensName)
}

// Deprecated: RequestCommunityInfoFromMailserver is deprecated in favor of
// configurable FetchCommunity.
func (api *PublicAPI) RequestCommunityInfoFromMailserver(communityID string) (*communities.Community, error) {
	request := &protocol.FetchCommunityRequest{
		CommunityKey:    communityID,
		Shard:           nil,
		TryDatabase:     true,
		WaitForResponse: true,
	}
	return api.FetchCommunity(request)
}

// Deprecated: RequestCommunityInfoFromMailserverWithShard is deprecated in favor of
// configurable FetchCommunity.
func (api *PublicAPI) RequestCommunityInfoFromMailserverWithShard(communityID string, shard *shard.Shard) (*communities.Community, error) {
	request := &protocol.FetchCommunityRequest{
		CommunityKey:    communityID,
		Shard:           shard,
		TryDatabase:     true,
		WaitForResponse: true,
	}
	return api.FetchCommunity(request)
}

// Deprecated: RequestCommunityInfoFromMailserverAsync is deprecated in favor of
// configurable FetchCommunity.
func (api *PublicAPI) RequestCommunityInfoFromMailserverAsync(communityID string) error {
	request := &protocol.FetchCommunityRequest{
		CommunityKey:    communityID,
		Shard:           nil,
		TryDatabase:     true,
		WaitForResponse: false,
	}
	_, err := api.FetchCommunity(request)
	return err
}

// Deprecated: RequestCommunityInfoFromMailserverAsyncWithShard is deprecated in favor of
// configurable FetchCommunity.
func (api *PublicAPI) RequestCommunityInfoFromMailserverAsyncWithShard(communityID string, shard *shard.Shard) error {
	request := &protocol.FetchCommunityRequest{
		CommunityKey:    communityID,
		Shard:           shard,
		TryDatabase:     true,
		WaitForResponse: false,
	}
	_, err := api.FetchCommunity(request)
	return err
}

func (api *PublicAPI) FetchCommunity(request *protocol.FetchCommunityRequest) (*communities.Community, error) {
	return api.service.messenger.FetchCommunity(request)
}

func (api *PublicAPI) ActivityCenterNotifications(request protocol.ActivityCenterNotificationsRequest) (*protocol.ActivityCenterPaginationResponse, error) {
	return api.service.messenger.ActivityCenterNotifications(request)
}

func (api *PublicAPI) ActivityCenterNotificationsCount(request protocol.ActivityCenterCountRequest) (*protocol.ActivityCenterCountResponse, error) {
	return api.service.messenger.ActivityCenterNotificationsCount(request)
}

func (api *PublicAPI) HasUnseenActivityCenterNotifications() (bool, error) {
	return api.service.messenger.HasUnseenActivityCenterNotifications()
}

func (api *PublicAPI) GetActivityCenterState() (*protocol.ActivityCenterState, error) {
	return api.service.messenger.GetActivityCenterState()
}

func (api *PublicAPI) MarkAsSeenActivityCenterNotifications() (*protocol.MessengerResponse, error) {
	return api.service.messenger.MarkAsSeenActivityCenterNotifications()
}

func (api *PublicAPI) MarkAllActivityCenterNotificationsRead(ctx context.Context) (*protocol.MessengerResponse, error) {
	return api.service.messenger.MarkAllActivityCenterNotificationsRead(ctx)
}

func (api *PublicAPI) MarkActivityCenterNotificationsRead(ctx context.Context, ids []types.HexBytes) (*protocol.MessengerResponse, error) {
	return api.service.messenger.MarkActivityCenterNotificationsRead(ctx, ids, 0, true)
}

func (api *PublicAPI) MarkActivityCenterNotificationsUnread(ctx context.Context, ids []types.HexBytes) (*protocol.MessengerResponse, error) {
	m := api.service.messenger
	updatedAt := m.GetCurrentTimeInMillis()
	return m.MarkActivityCenterNotificationsUnread(ctx, ids, updatedAt, true)
}

func (api *PublicAPI) AcceptActivityCenterNotifications(ctx context.Context, ids []types.HexBytes) (*protocol.MessengerResponse, error) {
	m := api.service.messenger
	updatedAt := m.GetCurrentTimeInMillis()
	return api.service.messenger.AcceptActivityCenterNotifications(ctx, ids, updatedAt, true)
}

func (api *PublicAPI) DismissActivityCenterNotifications(ctx context.Context, ids []types.HexBytes) error {
	_, err := api.service.messenger.DismissActivityCenterNotifications(ctx, ids, 0, true)
	return err
}

func (api *PublicAPI) DeleteActivityCenterNotifications(ctx context.Context, ids []types.HexBytes) error {
	m := api.service.messenger
	updatedAt := m.GetCurrentTimeInMillis()
	_, err := m.MarkActivityCenterNotificationsDeleted(ctx, ids, updatedAt, true)
	return err
}

func (api *PublicAPI) RequestAllHistoricMessages(forceFetchingBackup bool) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RequestAllHistoricMessages(forceFetchingBackup)
}

func (api *PublicAPI) RequestAllHistoricMessagesWithRetries(forceFetchingBackup bool) (*protocol.MessengerResponse, error) {
	return api.service.messenger.RequestAllHistoricMessagesWithRetries(forceFetchingBackup)
}

func (api *PublicAPI) DisconnectActiveMailserver() {
	api.service.messenger.DisconnectActiveMailserver()
}

// Echo is a method for testing purposes.
func (api *PublicAPI) Echo(ctx context.Context, message string) (string, error) {
	return message, nil
}

func (api *PublicAPI) FillGaps(chatID string, messageIDs []string) error {
	return api.service.messenger.FillGaps(chatID, messageIDs)
}

func (api *PublicAPI) SyncChatFromSyncedFrom(chatID string) (uint32, error) {
	return api.service.messenger.SyncChatFromSyncedFrom(chatID)
}

// BloomFilter returns the current bloom filter bytes
func (api *PublicAPI) BloomFilter() string {
	return hexutil.Encode(api.service.messenger.BloomFilter())
}

func (api *PublicAPI) StartDiscV5() error {
	return api.service.messenger.StartDiscV5()
}

func (api *PublicAPI) StopDiscV5() error {
	return api.service.messenger.StopDiscV5()
}

func (api *PublicAPI) GetCommunitiesSettings() ([]communities.CommunitySettings, error) {
	return api.service.messenger.GetCommunitiesSettings()
}

func (api *PublicAPI) EnableCommunityHistoryArchiveProtocol() error {
	return api.service.messenger.EnableCommunityHistoryArchiveProtocol()
}

func (api *PublicAPI) DisableCommunityHistoryArchiveProtocol() error {
	return api.service.messenger.DisableCommunityHistoryArchiveProtocol()
}

func (api *PublicAPI) SubscribeToPubsubTopic(topic string, optPublicKey string) error {
	var publicKey *ecdsa.PublicKey
	if optPublicKey != "" {
		keyBytes, err := hexutil.Decode(optPublicKey)
		if err != nil {
			return err
		}

		publicKey, err = crypto.UnmarshalPubkey(keyBytes)
		if err != nil {
			return err
		}
	}

	return api.service.messenger.SubscribeToPubsubTopic(topic, publicKey)
}

func (api *PublicAPI) StorePubsubTopicKey(topic string, privKey string) error {
	keyBytes, err := hexutil.Decode(privKey)
	if err != nil {
		return err
	}

	p, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return err
	}

	return api.service.messenger.StorePubsubTopicKey(topic, p)
}

func (api *PublicAPI) AddStorePeer(address string) (string, error) {
	peerID, err := api.service.messenger.AddStorePeer(address)
	return string(peerID), err
}

func (api *PublicAPI) AddRelayPeer(address string) (string, error) {
	peerID, err := api.service.messenger.AddRelayPeer(address)
	return string(peerID), err
}

func (api *PublicAPI) DialPeer(address string) error {
	return api.service.messenger.DialPeer(address)
}

func (api *PublicAPI) DialPeerByID(peerID string) error {
	return api.service.messenger.DialPeerByID(peerID)
}

func (api *PublicAPI) DropPeer(peerID string) error {
	return api.service.messenger.DropPeer(peerID)
}

func (api *PublicAPI) Peers() map[string]types.WakuV2Peer {
	return api.service.messenger.Peers()
}

func (api *PublicAPI) ListenAddresses() ([]string, error) {
	return api.service.messenger.ListenAddresses()
}

func (api *PublicAPI) ChangeIdentityImageShowTo(showTo settings.ProfilePicturesShowToType) error {
	err := api.service.accountsDB.SaveSettingField(settings.ProfilePicturesShowTo, showTo)
	if err != nil {
		return err
	}

	return api.service.messenger.PublishIdentityImage()
}

func (api *PublicAPI) BackupData() (uint64, error) {
	return api.service.messenger.BackupData(context.Background())
}

func (api *PublicAPI) ImageServerURL() string {
	return api.service.messenger.ImageServerURL()
}

func (api *PublicAPI) ToggleUseMailservers(value bool) error {
	return api.service.messenger.ToggleUseMailservers(value)
}

func (api *PublicAPI) SetPinnedMailservers(pinnedMailservers map[string]string) error {
	return api.service.messenger.SetPinnedMailservers(pinnedMailservers)
}

func (api *PublicAPI) RequestExtractDiscordChannelsAndCategories(filesToImport []string) {
	api.service.messenger.RequestExtractDiscordChannelsAndCategories(filesToImport)
}

func (api *PublicAPI) ExtractDiscordChannelsAndCategories(filesToImport []string) (*protocol.MessengerResponse, map[string]*discord.ImportError) {
	return api.service.messenger.ExtractDiscordChannelsAndCategories(filesToImport)
}

func (api *PublicAPI) RequestImportDiscordChannel(request *requests.ImportDiscordChannel) {
	api.service.messenger.RequestImportDiscordChannel(request)
}

func (api *PublicAPI) RequestImportDiscordCommunity(request *requests.ImportDiscordCommunity) {
	api.service.messenger.RequestImportDiscordCommunity(request)
}

func (api *PublicAPI) RequestCancelDiscordCommunityImport(id string) {
	api.service.messenger.MarkDiscordCommunityImportAsCancelled(id)
}

func (api *PublicAPI) RequestCancelDiscordChannelImport(discordChannelID string) {
	api.service.messenger.MarkDiscordChannelImportAsCancelled(discordChannelID)
}

func (api *PublicAPI) BuildContact(request *requests.BuildContact) (*protocol.Contact, error) {
	return api.service.messenger.BuildContact(request)
}

func (api *PublicAPI) GetCommunityTokens(communityID string) ([]*token.CommunityToken, error) {
	return api.service.messenger.GetCommunityTokens(communityID)
}

// GetCommunityPermissionedBalances returns balances indexed by account address.
func (api *PublicAPI) GetCommunityPermissionedBalances(request *requests.GetPermissionedBalances) (map[ethcommon.Address][]communities.PermissionedBalance, error) {
	return api.service.messenger.GetCommunityPermissionedBalances(request)
}

func (api *PublicAPI) GetAllCommunityTokens() ([]*token.CommunityToken, error) {
	return api.service.messenger.GetAllCommunityTokens()
}

func (api *PublicAPI) SaveCommunityToken(token *token.CommunityToken, croppedImage *images.CroppedImage) (*token.CommunityToken, error) {
	return api.service.messenger.SaveCommunityToken(token, croppedImage)
}

func (api *PublicAPI) AddCommunityToken(communityID string, chainID int, address string) error {
	return api.service.messenger.AddCommunityToken(communityID, chainID, address)
}

func (api *PublicAPI) UpdateCommunityTokenState(chainID int, contractAddress string, deployState token.DeployState) error {
	return api.service.messenger.UpdateCommunityTokenState(chainID, contractAddress, deployState)
}

func (api *PublicAPI) UpdateCommunityTokenAddress(chainID int, oldContractAddress string, newContractAddress string) error {
	return api.service.messenger.UpdateCommunityTokenAddress(chainID, oldContractAddress, newContractAddress)
}

func (api *PublicAPI) UpdateCommunityTokenSupply(chainID int, contractAddress string, supply *bigint.BigInt) error {
	return api.service.messenger.UpdateCommunityTokenSupply(chainID, contractAddress, supply)
}

func (api *PublicAPI) RemoveCommunityToken(chainID int, contractAddress string) error {
	return api.service.messenger.RemoveCommunityToken(chainID, contractAddress)
}

func (api *PublicAPI) ToggleCollapsedCommunityCategory(request *requests.ToggleCollapsedCommunityCategory) error {
	return api.service.messenger.ToggleCollapsedCommunityCategory(request)
}

func (api *PublicAPI) CollapsedCommunityCategories() ([]protocol.CollapsedCommunityCategory, error) {
	return api.service.messenger.CollapsedCommunityCategories()
}

func (api *PublicAPI) CheckPermissionsToJoinCommunity(request *requests.CheckPermissionToJoinCommunity) (*communities.CheckPermissionToJoinResponse, error) {
	return api.service.messenger.CheckPermissionsToJoinCommunity(request)
}

func (api *PublicAPI) CheckCommunityChannelPermissions(request *requests.CheckCommunityChannelPermissions) (*communities.CheckChannelPermissionsResponse, error) {
	return api.service.messenger.CheckCommunityChannelPermissions(request)
}

func (api *PublicAPI) CheckAllCommunityChannelsPermissions(request *requests.CheckAllCommunityChannelsPermissions) (*communities.CheckAllChannelsPermissionsResponse, error) {
	return api.service.messenger.CheckAllCommunityChannelsPermissions(request)
}

func (api *PublicAPI) CollectCommunityMetrics(request *requests.CommunityMetricsRequest) (*protocol.CommunityMetricsResponse, error) {
	return api.service.messenger.CollectCommunityMetrics(request)
}

func (api *PublicAPI) ShareCommunityURLWithChatKey(communityID types.HexBytes) (string, error) {
	return api.service.messenger.ShareCommunityURLWithChatKey(communityID)
}

func (api *PublicAPI) ShareCommunityURLWithData(communityID types.HexBytes) (string, error) {
	return api.service.messenger.ShareCommunityURLWithData(communityID)
}

func (api *PublicAPI) ShareCommunityChannelURLWithChatKey(request *requests.CommunityChannelShareURL) (string, error) {
	return api.service.messenger.ShareCommunityChannelURLWithChatKey(request)
}

func (api *PublicAPI) ShareCommunityChannelURLWithData(request *requests.CommunityChannelShareURL) (string, error) {
	return api.service.messenger.ShareCommunityChannelURLWithData(request)
}

func (api *PublicAPI) ShareUserURLWithENS(pubKey string) (string, error) {
	return api.service.messenger.ShareUserURLWithENS(pubKey)
}

func (api *PublicAPI) ShareUserURLWithChatKey(pubKey string) (string, error) {
	return api.service.messenger.ShareUserURLWithChatKey(pubKey)
}

func (api *PublicAPI) ShareUserURLWithData(pubKey string) (string, error) {
	return api.service.messenger.ShareUserURLWithData(pubKey)
}

func (api *PublicAPI) ParseSharedURL(url string) (*protocol.URLDataResponse, error) {
	return protocol.ParseSharedURL(url)
}

func (api *PublicAPI) Messenger() *protocol.Messenger {
	return api.service.messenger
}

// ChatMentionReplaceWithPublicKey checks if the text contains mentions and replace mention with user public key.
// e.g. abc @alice -> abc 0x123
func (api *PublicAPI) ChatMentionReplaceWithPublicKey(chatID, text string) (string, error) {
	return api.service.messenger.GetMentionsManager().ReplaceWithPublicKey(chatID, text)
}

// ChatMentionOnChangeText
// chatID: chat id
// text: the full text user input in the chat input field
// as performance consideration, we don't need to call this function each time after user input a character,
// say user input "abc", we don't need to call this function 3 times, instead,
// we can call this function 2 times as following:
// 1. user input "a", call this function with text "a"
// 2. user input "c", call this function with text "abc"
// whatever, we should ensure ChatMentionOnChangeText know(invoked) the latest full text.
// ChatMentionOnChangeText will maintain state of fulltext and diff between previous/latest full text internally.
func (api *PublicAPI) ChatMentionOnChangeText(chatID, text string) (*protocol.ChatMentionContext, error) {
	return api.service.messenger.GetMentionsManager().OnChangeText(chatID, text)
}

// ChatMentionSelectMention select mention from mention suggestion list
func (api *PublicAPI) ChatMentionSelectMention(chatID, text, primaryName, publicKey string) (*protocol.ChatMentionContext, error) {
	return api.service.messenger.GetMentionsManager().SelectMention(chatID, text, primaryName, publicKey)
}

func (api *PublicAPI) ChatMentionClearMentions(chatID string) {
	api.service.messenger.GetMentionsManager().ClearMentions(chatID)
}

// ChatMentionToInputField checks if the text contains mentions and replace mention with readable username.
// generally, this function is invoked before user editing a sent message.
func (api *PublicAPI) ChatMentionToInputField(chatID, text string) (*protocol.ChatMentionContext, error) {
	return api.service.messenger.GetMentionsManager().ToInputField(chatID, text)
}

func (api *PublicAPI) GetCheckChannelPermissionResponses(parent context.Context, communityID types.HexBytes) (*communities.CheckAllChannelsPermissionsResponse, error) {
	return api.service.messenger.GetCommunityCheckChannelPermissionResponses(communityID)
}

// CreateClosedCommunity used only for test purposes
func (api *PublicAPI) CreateClosedCommunity() (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateClosedCommunity()
}

// CreateOpenCommunity used only for test purposes
func (api *PublicAPI) CreateOpenCommunity() (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateOpenCommunity()
}

// CreateTokenGatedCommunity used only for test purposes
func (api *PublicAPI) CreateTokenGatedCommunity() (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateTokenGatedCommunity()
}

// Set profile showcase preference for current user
func (api *PublicAPI) SetProfileShowcasePreferences(preferences *protocol.ProfileShowcasePreferences) error {
	return api.service.messenger.SetProfileShowcasePreferences(preferences)
}

// Get all profile showcase preferences for current user
func (api *PublicAPI) GetProfileShowcasePreferences() (*protocol.ProfileShowcasePreferences, error) {
	return api.service.messenger.GetProfileShowcasePreferences()
}

// Get profile showcase for a contact
func (api *PublicAPI) GetProfileShowcaseForContact(contactID string) (*protocol.ProfileShowcase, error) {
	return api.service.messenger.GetProfileShowcaseForContact(contactID)
}

// Get profile showcase accounts by address
func (api *PublicAPI) GetProfileShowcaseAccountsByAddress(address string) ([]*protocol.ProfileShowcaseAccount, error) {
	return api.service.messenger.GetProfileShowcaseAccountsByAddress(address)
}

// Returns response with AC notification when owner token is received
func (api *PublicAPI) RegisterOwnerTokenReceivedNotification(communityID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateResponseWithACNotification(communityID, protocol.ActivityCenterNotificationTypeOwnerTokenReceived, false)
}

// Returns response with AC notification when setting signer is successful
func (api *PublicAPI) RegisterReceivedOwnershipNotification(communityID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateResponseWithACNotification(communityID, protocol.ActivityCenterNotificationTypeOwnershipReceived, false)
}

// Returns response with AC notification when community token is received
func (api *PublicAPI) RegisterReceivedCommunityTokenNotification(communityID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateResponseWithACNotification(communityID, protocol.ActivityCenterNotificationTypeCommunityTokenReceived, false)
}

// Returns response with AC notification when setting signer is failed
func (api *PublicAPI) RegisterSetSignerFailedNotification(communityID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateResponseWithACNotification(communityID, protocol.ActivityCenterNotificationTypeSetSignerFailed, false)
}

// Returns response with AC notification when setting signer is declined
func (api *PublicAPI) RegisterSetSignerDeclinedNotification(communityID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateResponseWithACNotification(communityID, protocol.ActivityCenterNotificationTypeSetSignerDeclined, true)
}

// Returns response with AC notification when ownership is lost
func (api *PublicAPI) RegisterLostOwnershipNotification(communityID string) (*protocol.MessengerResponse, error) {
	return api.service.messenger.CreateResponseWithACNotification(communityID, protocol.ActivityCenterNotificationTypeOwnershipLost, false)
}

func (api *PublicAPI) PromoteSelfToControlMode(communityID string) error {
	_, err := api.service.messenger.PromoteSelfToControlNode([]byte(communityID))
	return err
}

func (api *PublicAPI) FetchMessages(request *requests.FetchMessages) error {
	return api.service.messenger.FetchMessages(request)
}

func (api *PublicAPI) SetLightClient(request *requests.SetLightClient) error {
	return api.service.messenger.SetLightClient(request)
}

func (api *PublicAPI) SetLogLevel(request *requests.SetLogLevel) error {
	return api.service.messenger.SetLogLevel(request)
}

func (api *PublicAPI) SetCustomNodes(request *requests.SetCustomNodes) error {
	return api.service.messenger.SetCustomNodes(request)
}

func (api *PublicAPI) SetCustomizationColor(ctx context.Context, request *requests.SetCustomizationColor) error {
	return api.service.messenger.SetCustomizationColor(ctx, request)
}

// -----
// HELPER
// -----

// MakeMessagesRequestPayload makes a specific payload for MailServer
// to request historic messages.
// DEPRECATED
func MakeMessagesRequestPayload(r MessagesRequest) ([]byte, error) {
	cursor, err := hex.DecodeString(r.Cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %v", err)
	}

	if len(cursor) > 0 && len(cursor) != mailserver.CursorLength {
		return nil, fmt.Errorf("invalid cursor size: expected %d but got %d", mailserver.CursorLength, len(cursor))
	}

	payload := mailserver.MessagesRequestPayload{
		Lower: r.From,
		Upper: r.To,
		// We need to pass bloom filter for
		// backward compatibility
		Bloom:  createBloomFilter(r),
		Topics: topicsToByteArray(r.Topics),
		Limit:  r.Limit,
		Cursor: cursor,
		// Client must tell the MailServer if it supports batch responses.
		// This can be removed in the future.
		Batch: true,
	}

	return rlp.EncodeToBytes(payload)
}

func topicsToByteArray(topics []types.TopicType) [][]byte {

	var response [][]byte
	for idx := range topics {
		response = append(response, topics[idx][:])
	}

	return response
}

func createBloomFilter(r MessagesRequest) []byte {
	if len(r.Topics) > 0 {
		return topicsToBloom(r.Topics...)
	}
	return types.TopicToBloom(r.Topic)
}

func topicsToBloom(topics ...types.TopicType) []byte {
	i := new(big.Int)
	for _, topic := range topics {
		bloom := types.TopicToBloom(topic)
		i.Or(i, new(big.Int).SetBytes(bloom[:]))
	}

	combined := make([]byte, types.BloomFilterSize)
	data := i.Bytes()
	copy(combined[types.BloomFilterSize-len(data):], data[:])

	return combined
}

// TopicsToBloom squashes all topics into a single bloom filter.
func TopicsToBloom(topics ...types.TopicType) []byte {
	return topicsToBloom(topics...)
}
