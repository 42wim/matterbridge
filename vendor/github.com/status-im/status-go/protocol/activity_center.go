package protocol

import (
	"crypto/ecdsa"
	"errors"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/verification"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

// The activity center is a place where we store incoming notifications before
// they are shown to the users as new chats, in order to mitigate the impact of spam
// on the messenger

type ActivityCenterType int

const (
	ActivityCenterNotificationNoType ActivityCenterType = iota
	ActivityCenterNotificationTypeNewOneToOne
	ActivityCenterNotificationTypeNewPrivateGroupChat
	ActivityCenterNotificationTypeMention
	ActivityCenterNotificationTypeReply
	ActivityCenterNotificationTypeContactRequest
	ActivityCenterNotificationTypeCommunityInvitation
	ActivityCenterNotificationTypeCommunityRequest
	ActivityCenterNotificationTypeCommunityMembershipRequest
	ActivityCenterNotificationTypeCommunityKicked
	ActivityCenterNotificationTypeContactVerification
	ActivityCenterNotificationTypeContactRemoved
	ActivityCenterNotificationTypeNewKeypairAddedToPairedDevice
	ActivityCenterNotificationTypeOwnerTokenReceived
	ActivityCenterNotificationTypeOwnershipReceived
	ActivityCenterNotificationTypeOwnershipLost
	ActivityCenterNotificationTypeSetSignerFailed
	ActivityCenterNotificationTypeSetSignerDeclined
	ActivityCenterNotificationTypeShareAccounts
	ActivityCenterNotificationTypeCommunityTokenReceived
	ActivityCenterNotificationTypeFirstCommunityTokenReceived
	ActivityCenterNotificationTypeCommunityBanned
	ActivityCenterNotificationTypeCommunityUnbanned
)

type ActivityCenterMembershipStatus int

const (
	ActivityCenterMembershipStatusIdle ActivityCenterMembershipStatus = iota
	ActivityCenterMembershipStatusPending
	ActivityCenterMembershipStatusAccepted
	ActivityCenterMembershipStatusDeclined
	ActivityCenterMembershipStatusAcceptedPending
	ActivityCenterMembershipStatusDeclinedPending
	ActivityCenterMembershipOwnershipChanged
)

type ActivityCenterQueryParamsRead uint

const (
	ActivityCenterQueryParamsReadRead = iota + 1
	ActivityCenterQueryParamsReadUnread
	ActivityCenterQueryParamsReadAll
)

var ErrInvalidActivityCenterNotification = errors.New("invalid activity center notification")

type ActivityTokenData struct {
	ChainID       uint64                         `json:"chainId,omitempty"`
	CollectibleID thirdparty.CollectibleUniqueID `json:"collectibleId,omitempty"`
	TxHash        string                         `json:"txHash,omitempty"`
	WalletAddress string                         `json:"walletAddress,omitempty"`
	IsFirst       bool                           `json:"isFirst,omitempty"`
	// Community data
	CommunityID string `json:"communityId,omitempty"`
	// Token data
	Amount    string `json:"amount,omitempty"`
	Name      string `json:"name,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
	ImageURL  string `json:"imageUrl,omitempty"`
	TokenType int    `json:"tokenType,omitempty"`
}

type ActivityCenterNotification struct {
	ID                        types.HexBytes                 `json:"id"`
	ChatID                    string                         `json:"chatId"`
	CommunityID               string                         `json:"communityId"`
	MembershipStatus          ActivityCenterMembershipStatus `json:"membershipStatus"`
	Name                      string                         `json:"name"`
	Author                    string                         `json:"author"`
	Type                      ActivityCenterType             `json:"type"`
	LastMessage               *common.Message                `json:"lastMessage"`
	Message                   *common.Message                `json:"message"`
	ReplyMessage              *common.Message                `json:"replyMessage"`
	Timestamp                 uint64                         `json:"timestamp"`
	Read                      bool                           `json:"read"`
	Dismissed                 bool                           `json:"dismissed"`
	Deleted                   bool                           `json:"deleted"`
	Accepted                  bool                           `json:"accepted"`
	ContactVerificationStatus verification.RequestStatus     `json:"contactVerificationStatus"`
	TokenData                 *ActivityTokenData             `json:"tokenData"`
	//Used for synchronization. Each update should increment the UpdatedAt.
	//The value should represent the time when the update occurred.
	UpdatedAt     uint64            `json:"updatedAt"`
	AlbumMessages []*common.Message `json:"albumMessages"`
}

func (n *ActivityCenterNotification) IncrementUpdatedAt(timesource common.TimeSource) {
	tNow := timesource.GetCurrentTime()
	// If updatead at is greater or equal than time now, we bump it
	if n.UpdatedAt >= tNow {
		n.UpdatedAt++
	} else {
		n.UpdatedAt = tNow
	}
}

type ActivityCenterNotificationsRequest struct {
	Cursor        string                        `json:"cursor"`
	Limit         uint64                        `json:"limit"`
	ActivityTypes []ActivityCenterType          `json:"activityTypes"`
	ReadType      ActivityCenterQueryParamsRead `json:"readType"`
}

type ActivityCenterCountRequest struct {
	ActivityTypes []ActivityCenterType          `json:"activityTypes"`
	ReadType      ActivityCenterQueryParamsRead `json:"readType"`
}

type ActivityCenterPaginationResponse struct {
	Cursor        string                        `json:"cursor"`
	Notifications []*ActivityCenterNotification `json:"notifications"`
}

type ActivityCenterCountResponse = map[ActivityCenterType]uint64

type ActivityCenterState struct {
	HasSeen   bool   `json:"hasSeen"`
	UpdatedAt uint64 `json:"updatedAt"`
}

func (n *ActivityCenterNotification) Valid() error {
	if len(n.ID) == 0 || n.Type == 0 || n.Timestamp == 0 {
		return ErrInvalidActivityCenterNotification
	}
	return nil
}

func showMentionOrReplyActivityCenterNotification(publicKey ecdsa.PublicKey, message *common.Message, chat *Chat, responseTo *common.Message) (bool, ActivityCenterType) {
	if chat == nil || !chat.Active || (!chat.CommunityChat() && !chat.PrivateGroupChat()) || chat.Muted {
		return false, ActivityCenterNotificationNoType
	}

	if message.Mentioned {
		return true, ActivityCenterNotificationTypeMention
	}

	publicKeyString := common.PubkeyToHex(&publicKey)
	if responseTo != nil && responseTo.From == publicKeyString {
		return true, ActivityCenterNotificationTypeReply
	}

	return false, ActivityCenterNotificationNoType
}
