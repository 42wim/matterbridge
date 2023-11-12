package protocol

import (
	"crypto/ecdsa"
	"encoding/json"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	localnotifications "github.com/status-im/status-go/services/local-notifications"
)

type NotificationBody struct {
	Message   *common.Message        `json:"message"`
	Contact   *Contact               `json:"contact"`
	Chat      *Chat                  `json:"chat"`
	Community *communities.Community `json:"community"`
}

func showMessageNotification(publicKey ecdsa.PublicKey, message *common.Message, chat *Chat, responseTo *common.Message) bool {
	if chat != nil && !chat.Active {
		return false
	}

	if chat != nil && (chat.OneToOne() || chat.PrivateGroupChat()) {
		return true
	}

	if message.Mentioned {
		return true
	}

	if responseTo != nil {
		publicKeyString := common.PubkeyToHex(&publicKey)
		return responseTo.From == publicKeyString
	}

	return false
}

func (n NotificationBody) MarshalJSON() ([]byte, error) {
	type Alias NotificationBody
	item := struct{ *Alias }{Alias: (*Alias)(&n)}
	return json.Marshal(item)
}

func NewMessageNotification(id string, message *common.Message, chat *Chat, contact *Contact, resolvePrimaryName func(string) (string, error), profilePicturesVisibility int) (*localnotifications.Notification, error) {
	body := &NotificationBody{
		Message: message,
		Chat:    chat,
		Contact: contact,
	}

	return body.toMessageNotification(id, resolvePrimaryName, profilePicturesVisibility)
}

func DeletedMessageNotification(id string, chat *Chat) *localnotifications.Notification {
	return &localnotifications.Notification{
		BodyType:       localnotifications.TypeMessage,
		ID:             gethcommon.HexToHash(id),
		IsConversation: true,
		ConversationID: chat.ID,
		Deeplink:       chat.DeepLink(),
		Deleted:        true,
	}
}

func NewCommunityRequestToJoinNotification(id string, community *communities.Community, contact *Contact) *localnotifications.Notification {
	body := &NotificationBody{
		Community: community,
		Contact:   contact,
	}

	return body.toCommunityRequestToJoinNotification(id)
}

func NewPrivateGroupInviteNotification(id string, chat *Chat, contact *Contact, profilePicturesVisibility int) *localnotifications.Notification {
	body := &NotificationBody{
		Chat:    chat,
		Contact: contact,
	}

	return body.toPrivateGroupInviteNotification(id, profilePicturesVisibility)
}

func (n NotificationBody) toMessageNotification(id string, resolvePrimaryName func(string) (string, error), profilePicturesVisibility int) (*localnotifications.Notification, error) {
	var title string
	if n.Chat.PrivateGroupChat() || n.Chat.Public() || n.Chat.CommunityChat() {
		title = n.Chat.Name
	} else if n.Chat.OneToOne() {
		title = n.Contact.PrimaryName()

	}

	canonicalNames := make(map[string]string)
	for _, mentionID := range n.Message.Mentions {
		name, err := resolvePrimaryName(mentionID)
		if err != nil {
			canonicalNames[mentionID] = mentionID
		} else {
			canonicalNames[mentionID] = name
		}
	}

	// We don't pass idenity as only interested in the simplified text
	simplifiedText, err := n.Message.GetSimplifiedText("", canonicalNames)
	if err != nil {
		return nil, err
	}

	return &localnotifications.Notification{
		Body:                n,
		ID:                  gethcommon.HexToHash(id),
		BodyType:            localnotifications.TypeMessage,
		Category:            localnotifications.CategoryMessage,
		Deeplink:            n.Chat.DeepLink(),
		Title:               title,
		Message:             simplifiedText,
		IsConversation:      true,
		IsGroupConversation: true,
		Author: localnotifications.NotificationAuthor{
			Name: n.Contact.PrimaryName(),
			Icon: n.Contact.CanonicalImage(settings.ProfilePicturesVisibilityType(profilePicturesVisibility)),
			ID:   n.Contact.ID,
		},
		Timestamp:      n.Message.WhisperTimestamp,
		ConversationID: n.Chat.ID,
		Image:          "",
	}, nil
}

func (n NotificationBody) toPrivateGroupInviteNotification(id string, profilePicturesVisibility int) *localnotifications.Notification {
	return &localnotifications.Notification{
		ID:       gethcommon.HexToHash(id),
		Body:     n,
		Title:    n.Contact.PrimaryName() + " invited you to " + n.Chat.Name,
		Message:  n.Contact.PrimaryName() + " wants you to join group " + n.Chat.Name,
		BodyType: localnotifications.TypeMessage,
		Category: localnotifications.CategoryGroupInvite,
		Deeplink: n.Chat.DeepLink(),
		Author: localnotifications.NotificationAuthor{
			Name: n.Contact.PrimaryName(),
			Icon: n.Contact.CanonicalImage(settings.ProfilePicturesVisibilityType(profilePicturesVisibility)),
			ID:   n.Contact.ID,
		},
		Image: "",
	}
}

func (n NotificationBody) toCommunityRequestToJoinNotification(id string) *localnotifications.Notification {
	return &localnotifications.Notification{
		ID:       gethcommon.HexToHash(id),
		Body:     n,
		Title:    n.Contact.PrimaryName() + " wants to join " + n.Community.Name(),
		Message:  n.Contact.PrimaryName() + " wants to join  message " + n.Community.Name(),
		BodyType: localnotifications.TypeMessage,
		Category: localnotifications.CategoryCommunityRequestToJoin,
		Deeplink: "status-app://cr/" + n.Community.IDString(),
		Image:    "",
	}
}
