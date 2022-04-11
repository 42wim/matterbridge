package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"encoding/json"
	"fmt"
)

// MessagesAudioMessage struct.
type MessagesAudioMessage struct {
	AccessKey string `json:"access_key"` // Access key for the document
	ID        int    `json:"id"`         // Document ID
	OwnerID   int    `json:"owner_id"`   // Document owner ID
	Duration  int    `json:"duration"`   // Audio message duration in seconds
	LinkMp3   string `json:"link_mp3"`   // MP3 file URL
	LinkOgg   string `json:"link_ogg"`   // OGG file URL
	Waveform  []int  `json:"waveform"`   // Sound visualisation
}

// ToAttachment return attachment format.
func (doc MessagesAudioMessage) ToAttachment() string {
	return fmt.Sprintf("doc%d_%d", doc.OwnerID, doc.ID)
}

// MessagesGraffiti struct.
type MessagesGraffiti struct {
	AccessKey string `json:"access_key"` // Access key for the document
	ID        int    `json:"id"`         // Document ID
	OwnerID   int    `json:"owner_id"`   // Document owner ID
	URL       string `json:"url"`        // Graffiti URL
	Width     int    `json:"width"`      // Graffiti width
	Height    int    `json:"height"`     // Graffiti height
}

// ToAttachment return attachment format.
func (doc MessagesGraffiti) ToAttachment() string {
	return fmt.Sprintf("doc%d_%d", doc.OwnerID, doc.ID)
}

// MessagesMessage struct.
type MessagesMessage struct {
	// Only for messages from community. Contains user ID of community admin,
	// who sent this message.
	AdminAuthorID int                         `json:"admin_author_id"`
	Action        MessagesMessageAction       `json:"action"`
	Attachments   []MessagesMessageAttachment `json:"attachments"`

	// Unique auto-incremented number for all messages with this peer.
	ConversationMessageID int `json:"conversation_message_id"`

	// Date when the message has been sent in Unixtime.
	Date int `json:"date"`

	// Message author's ID.
	FromID int `json:"from_id"`

	// Forwarded messages.
	FwdMessages  []MessagesMessage `json:"fwd_Messages"`
	ReplyMessage *MessagesMessage  `json:"reply_message"`
	Geo          BaseMessageGeo    `json:"geo"`
	PinnedAt     int               `json:"pinned_at,omitempty"`
	ID           int               `json:"id"`        // Message ID
	Deleted      BaseBoolInt       `json:"deleted"`   // Is it an deleted message
	Important    BaseBoolInt       `json:"important"` // Is it an important message
	IsHidden     BaseBoolInt       `json:"is_hidden"`
	IsCropped    BaseBoolInt       `json:"is_cropped"`
	IsSilent     BaseBoolInt       `json:"is_silent"`
	Out          BaseBoolInt       `json:"out"` // Information whether the message is outcoming
	WasListened  BaseBoolInt       `json:"was_listened,omitempty"`
	Keyboard     MessagesKeyboard  `json:"keyboard"`
	Template     MessagesTemplate  `json:"template"`
	Payload      string            `json:"payload"`
	PeerID       int               `json:"peer_id"` // Peer ID

	// ID used for sending messages. It returned only for outgoing messages.
	RandomID     int    `json:"random_id"`
	Ref          string `json:"ref"`
	RefSource    string `json:"ref_source"`
	Text         string `json:"text"`          // Message text
	UpdateTime   int    `json:"update_time"`   // Date when the message has been updated in Unixtime
	MembersCount int    `json:"members_count"` // Members number
	ExpireTTL    int    `json:"expire_ttl"`
	MessageTag   string `json:"message_tag"` // for https://notify.mail.ru/
}

// MessagesBasePayload struct.
type MessagesBasePayload struct {
	ButtonType string `json:"button_type,omitempty"`
	Command    string `json:"command,omitempty"`
	Payload    string `json:"payload,omitempty"`
}

// Command for MessagesBasePayload.
const (
	CommandNotSupportedButton = "not_supported_button"
)

// MessagesKeyboard struct.
type MessagesKeyboard struct {
	AuthorID int                        `json:"author_id,omitempty"` // Community or bot, which set this keyboard
	Buttons  [][]MessagesKeyboardButton `json:"buttons"`
	OneTime  BaseBoolInt                `json:"one_time,omitempty"` // Should this keyboard disappear on first use
	Inline   BaseBoolInt                `json:"inline,omitempty"`
}

// NewMessagesKeyboard returns a new MessagesKeyboard.
func NewMessagesKeyboard(oneTime BaseBoolInt) *MessagesKeyboard {
	return &MessagesKeyboard{
		Buttons: [][]MessagesKeyboardButton{},
		OneTime: oneTime,
	}
}

// NewMessagesKeyboardInline returns a new inline MessagesKeyboard.
func NewMessagesKeyboardInline() *MessagesKeyboard {
	return &MessagesKeyboard{
		Buttons: [][]MessagesKeyboardButton{},
		Inline:  true,
	}
}

// AddRow add row in MessagesKeyboard.
func (keyboard *MessagesKeyboard) AddRow() *MessagesKeyboard {
	if len(keyboard.Buttons) == 0 {
		keyboard.Buttons = make([][]MessagesKeyboardButton, 1)
	} else {
		row := make([]MessagesKeyboardButton, 0)
		keyboard.Buttons = append(keyboard.Buttons, row)
	}

	return keyboard
}

// AddTextButton add Text button in last row.
func (keyboard *MessagesKeyboard) AddTextButton(label string, payload interface{}, color string) *MessagesKeyboard {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	button := MessagesKeyboardButton{
		Action: MessagesKeyboardButtonAction{
			Type:    ButtonText,
			Label:   label,
			Payload: string(b),
		},
		Color: color,
	}

	lastRow := len(keyboard.Buttons) - 1
	keyboard.Buttons[lastRow] = append(keyboard.Buttons[lastRow], button)

	return keyboard
}

// AddOpenLinkButton add Open Link button in last row.
func (keyboard *MessagesKeyboard) AddOpenLinkButton(link, label string, payload interface{}) *MessagesKeyboard {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	button := MessagesKeyboardButton{
		Action: MessagesKeyboardButtonAction{
			Type:    ButtonOpenLink,
			Payload: string(b),
			Label:   label,
			Link:    link,
		},
	}

	lastRow := len(keyboard.Buttons) - 1
	keyboard.Buttons[lastRow] = append(keyboard.Buttons[lastRow], button)

	return keyboard
}

// AddLocationButton add Location button in last row.
func (keyboard *MessagesKeyboard) AddLocationButton(payload interface{}) *MessagesKeyboard {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	button := MessagesKeyboardButton{
		Action: MessagesKeyboardButtonAction{
			Type:    ButtonLocation,
			Payload: string(b),
		},
	}

	lastRow := len(keyboard.Buttons) - 1
	keyboard.Buttons[lastRow] = append(keyboard.Buttons[lastRow], button)

	return keyboard
}

// AddVKPayButton add VK Pay button in last row.
func (keyboard *MessagesKeyboard) AddVKPayButton(payload interface{}, hash string) *MessagesKeyboard {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	button := MessagesKeyboardButton{
		Action: MessagesKeyboardButtonAction{
			Type:    ButtonVKPay,
			Payload: string(b),
			Hash:    hash,
		},
	}

	lastRow := len(keyboard.Buttons) - 1
	keyboard.Buttons[lastRow] = append(keyboard.Buttons[lastRow], button)

	return keyboard
}

// AddVKAppsButton add VK Apps button in last row.
func (keyboard *MessagesKeyboard) AddVKAppsButton(
	appID, ownerID int,
	payload interface{},
	label, hash string,
) *MessagesKeyboard {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	button := MessagesKeyboardButton{
		Action: MessagesKeyboardButtonAction{
			Type:    ButtonVKApp,
			AppID:   appID,
			OwnerID: ownerID,
			Payload: string(b),
			Label:   label,
			Hash:    hash,
		},
	}

	lastRow := len(keyboard.Buttons) - 1
	keyboard.Buttons[lastRow] = append(keyboard.Buttons[lastRow], button)

	return keyboard
}

// AddCallbackButton add Callback button in last row.
func (keyboard *MessagesKeyboard) AddCallbackButton(label string, payload interface{}, color string) *MessagesKeyboard {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	button := MessagesKeyboardButton{
		Action: MessagesKeyboardButtonAction{
			Type:    ButtonCallback,
			Label:   label,
			Payload: string(b),
		},
		Color: color,
	}

	lastRow := len(keyboard.Buttons) - 1
	keyboard.Buttons[lastRow] = append(keyboard.Buttons[lastRow], button)

	return keyboard
}

// ToJSON returns the JSON encoding of MessagesKeyboard.
func (keyboard MessagesKeyboard) ToJSON() string {
	b, _ := json.Marshal(keyboard)
	return string(b)
}

// MessagesKeyboardButton struct.
type MessagesKeyboardButton struct {
	Action MessagesKeyboardButtonAction `json:"action"`
	Color  string                       `json:"color,omitempty"` // Button color
}

// MessagesKeyboardButtonAction struct.
type MessagesKeyboardButtonAction struct {
	AppID   int    `json:"app_id,omitempty"`   // Fragment value in app link like vk.com/app{app_id}_-654321#hash
	Hash    string `json:"hash,omitempty"`     // Fragment value in app link like vk.com/app123456_-654321#{hash}
	Label   string `json:"label,omitempty"`    // Label for button
	OwnerID int    `json:"owner_id,omitempty"` // Fragment value in app link like vk.com/app123456_{owner_id}#hash
	Payload string `json:"payload,omitempty"`  // Additional data sent along with message for developer convenience
	Type    string `json:"type"`               // Button type
	Link    string `json:"link,omitempty"`     // Link URL
}

// MessagesEventDataShowSnackbar struct.
type MessagesEventDataShowSnackbar struct {
	Text string `json:"text,omitempty"`
}

// MessagesEventDataOpenLink struct.
type MessagesEventDataOpenLink struct {
	Link string `json:"link,omitempty"`
}

// MessagesEventDataOpenApp struct.
type MessagesEventDataOpenApp struct {
	AppID   int    `json:"app_id,omitempty"`
	OwnerID int    `json:"owner_id,omitempty"`
	Hash    string `json:"hash,omitempty"`
}

// MessagesEventData struct.
type MessagesEventData struct {
	Type string `json:"type"`
	MessagesEventDataShowSnackbar
	MessagesEventDataOpenLink
	MessagesEventDataOpenApp
}

// NewMessagesEventDataShowSnackbar show disappearing message.
//
// Contains the field text - the text you want to print
// (maximum 90 characters). Snackbar is shown for 10 seconds and automatically
// hides, while the user has the ability to flick it off the screen.
func NewMessagesEventDataShowSnackbar(text string) *MessagesEventData {
	return &MessagesEventData{
		Type: "show_snackbar",
		MessagesEventDataShowSnackbar: MessagesEventDataShowSnackbar{
			Text: text,
		},
	}
}

// NewMessagesEventDataOpenLink open the link. Click on the specified address.
func NewMessagesEventDataOpenLink(link string) *MessagesEventData {
	return &MessagesEventData{
		Type: "open_link",
		MessagesEventDataOpenLink: MessagesEventDataOpenLink{
			Link: link,
		},
	}
}

// NewMessagesEventDataOpenApp open the link. Click on the specified address.
func NewMessagesEventDataOpenApp(appID, ownerID int, hash string) *MessagesEventData {
	return &MessagesEventData{
		Type: "open_app",
		MessagesEventDataOpenApp: MessagesEventDataOpenApp{
			AppID:   appID,
			OwnerID: ownerID,
			Hash:    hash,
		},
	}
}

// ToJSON returns the JSON encoding of MessagesEventData.
func (eventData MessagesEventData) ToJSON() string {
	b, _ := json.Marshal(eventData)
	return string(b)
}

// MessagesTemplate struct.
//
// https://vk.com/dev/bot_docs_templates
type MessagesTemplate struct {
	Type     string                    `json:"type"`
	Elements []MessagesTemplateElement `json:"elements"`
}

// ToJSON returns the JSON encoding of MessagesKeyboard.
func (template MessagesTemplate) ToJSON() string {
	b, _ := json.Marshal(template)
	return string(b)
}

// MessagesTemplateElement struct.
type MessagesTemplateElement struct {
	MessagesTemplateElementCarousel
}

// MessagesTemplateElementCarousel struct.
type MessagesTemplateElementCarousel struct {
	Title       string                                `json:"title,omitempty"`
	Action      MessagesTemplateElementCarouselAction `json:"action,omitempty"`
	Description string                                `json:"description,omitempty"`
	Photo       *PhotosPhoto                          `json:"photo,omitempty"`    // Only read
	PhotoID     string                                `json:"photo_id,omitempty"` // Only for send
	Buttons     []MessagesKeyboardButton              `json:"buttons,omitempty"`
}

// MessagesTemplateElementCarouselAction struct.
type MessagesTemplateElementCarouselAction struct {
	Type string `json:"type"`
	Link string `json:"link,omitempty"`
}

// MessageContentSourceMessage ...
type MessageContentSourceMessage struct {
	OwnerID               int `json:"owner_id,omitempty"`
	PeerID                int `json:"peer_id,omitempty"`
	ConversationMessageID int `json:"conversation_message_id,omitempty"`
}

// MessageContentSourceURL ...
type MessageContentSourceURL struct {
	URL string `json:"url,omitempty"`
}

// MessageContentSource struct.
//
// https://vk.com/dev/bots_docs_2
type MessageContentSource struct {
	Type                        string `json:"type"`
	MessageContentSourceMessage        // type message
	MessageContentSourceURL            // type url
}

// NewMessageContentSourceMessage ...
func NewMessageContentSourceMessage(ownerID, peerID, conversationMessageID int) *MessageContentSource {
	return &MessageContentSource{
		Type: "message",
		MessageContentSourceMessage: MessageContentSourceMessage{
			OwnerID:               ownerID,
			PeerID:                peerID,
			ConversationMessageID: conversationMessageID,
		},
	}
}

// NewMessageContentSourceURL ...
func NewMessageContentSourceURL(u string) *MessageContentSource {
	return &MessageContentSource{
		Type: "url",
		MessageContentSourceURL: MessageContentSourceURL{
			URL: u,
		},
	}
}

// ToJSON returns the JSON encoding of MessageContentSource.
func (contentSource MessageContentSource) ToJSON() string {
	b, _ := json.Marshal(contentSource)
	return string(b)
}

// MessagesChat struct.
type MessagesChat struct {
	AdminID        int         `json:"admin_id"` // Chat creator ID
	ID             int         `json:"id"`       // Chat ID
	IsDefaultPhoto BaseBoolInt `json:"is_default_photo"`
	IsGroupChannel BaseBoolInt `json:"is_group_channel"`
	Photo100       string      `json:"photo_100"` // URL of the preview image with 100 px in width
	Photo200       string      `json:"photo_200"` // URL of the preview image with 200 px in width
	Photo50        string      `json:"photo_50"`  // URL of the preview image with 50 px in width
	Title          string      `json:"title"`     // Chat title
	Type           string      `json:"type"`      // Chat type
	Users          []int       `json:"users"`
	MembersCount   int         `json:"members_count"`
}

// MessagesChatPreview struct.
type MessagesChatPreview struct {
	AdminID      int                              `json:"admin_id"`
	MembersCount int                              `json:"members_count"`
	Members      []int                            `json:"members"`
	Title        string                           `json:"title"`
	Photo        MessagesChatSettingsPhoto        `json:"photo"`
	LocalID      int                              `json:"local_id"`
	Joined       bool                             `json:"joined"`
	ChatSettings MessagesConversationChatSettings `json:"chat_settings"`
	IsMember     BaseBoolInt                      `json:"is_member"`
}

// MessagesChatPushSettings struct.
type MessagesChatPushSettings struct {
	DisabledUntil int         `json:"disabled_until"` // Time until that notifications are disabled
	Sound         BaseBoolInt `json:"sound"`          // Information whether the sound is on
}

// MessagesChatSettingsPhoto struct.
type MessagesChatSettingsPhoto struct {
	Photo100           string      `json:"photo_100"`
	Photo200           string      `json:"photo_200"`
	Photo50            string      `json:"photo_50"`
	IsDefaultPhoto     BaseBoolInt `json:"is_default_photo"`
	IsDefaultCallPhoto bool        `json:"is_default_call_photo"`
}

// MessagesConversation struct.
type MessagesConversation struct {
	CanWrite                  MessagesConversationCanWrite     `json:"can_write"`
	ChatSettings              MessagesConversationChatSettings `json:"chat_settings"`
	InRead                    int                              `json:"in_read"`         // Last message user have read
	LastMessageID             int                              `json:"last_message_id"` // ID of the last message in conversation
	Mentions                  []int                            `json:"mentions"`        // IDs of messages with mentions
	MessageRequest            string                           `json:"message_request"`
	LastConversationMessageID int                              `json:"last_conversation_message_id"`
	InReadCMID                int                              `json:"in_read_cmid"`
	OutReadCMID               int                              `json:"out_read_cmid"`

	// Last outcoming message have been read by the opponent.
	OutRead         int                              `json:"out_read"`
	Peer            MessagesConversationPeer         `json:"peer"`
	PushSettings    MessagesConversationPushSettings `json:"push_settings"`
	Important       BaseBoolInt                      `json:"important"`
	Unanswered      BaseBoolInt                      `json:"unanswered"`
	IsMarkedUnread  BaseBoolInt                      `json:"is_marked_unread"`
	CanSendMoney    BaseBoolInt                      `json:"can_send_money"`
	CanReceiveMoney BaseBoolInt                      `json:"can_receive_money"`
	IsNew           BaseBoolInt                      `json:"is_new"`
	IsArchived      BaseBoolInt                      `json:"is_archived"`
	UnreadCount     int                              `json:"unread_count"` // Unread messages number
	CurrentKeyboard MessagesKeyboard                 `json:"current_keyboard"`
	SortID          struct {
		MajorID int `json:"major_id"`
		MinorID int `json:"minor_id"`
	} `json:"sort_id"`
}

// MessagesConversationCanWrite struct.
type MessagesConversationCanWrite struct {
	Allowed BaseBoolInt `json:"allowed"`
	Reason  int         `json:"reason"`
}

// MessagesConversationChatSettings struct.
type MessagesConversationChatSettings struct {
	MembersCount  int                       `json:"members_count"`
	FriendsCount  int                       `json:"friends_count"`
	Photo         MessagesChatSettingsPhoto `json:"photo"`
	PinnedMessage MessagesPinnedMessage     `json:"pinned_message"`
	State         string                    `json:"state"`
	Title         string                    `json:"title"`
	ActiveIDs     []int                     `json:"active_ids"`
	ACL           struct {
		CanInvite            BaseBoolInt `json:"can_invite"`
		CanChangeInfo        BaseBoolInt `json:"can_change_info"`
		CanChangePin         BaseBoolInt `json:"can_change_pin"`
		CanPromoteUsers      BaseBoolInt `json:"can_promote_users"`
		CanSeeInviteLink     BaseBoolInt `json:"can_see_invite_link"`
		CanChangeInviteLink  BaseBoolInt `json:"can_change_invite_link"`
		CanCopyChat          BaseBoolInt `json:"can_copy_chat"`
		CanModerate          BaseBoolInt `json:"can_moderate"`
		CanCall              BaseBoolInt `json:"can_call"`
		CanUseMassMentions   BaseBoolInt `json:"can_use_mass_mentions"`
		CanChangeServiceType BaseBoolInt `json:"can_change_service_type"`
		CanChangeStyle       BaseBoolInt `json:"can_change_style"`
	} `json:"acl"`
	IsGroupChannel   BaseBoolInt             `json:"is_group_channel"`
	IsDisappearing   BaseBoolInt             `json:"is_disappearing"`
	IsService        BaseBoolInt             `json:"is_service"`
	IsCreatedForCall BaseBoolInt             `json:"is_created_for_call"`
	OwnerID          int                     `json:"owner_id"`
	AdminIDs         []int                   `json:"admin_ids"`
	Permissions      MessagesChatPermissions `json:"permissions"`
}

// MessagesChatPermission struct.
type MessagesChatPermission string

// Possible values.
const (
	OwnerChatPermission          MessagesChatPermission = "owner"
	OwnerAndAdminsChatPermission MessagesChatPermission = "owner_and_admins"
	AllChatPermission            MessagesChatPermission = "all"
)

// MessagesChatPermissions struct.
type MessagesChatPermissions struct {
	Invite          MessagesChatPermission `json:"invite"`
	ChangeInfo      MessagesChatPermission `json:"change_info"`
	ChangePin       MessagesChatPermission `json:"change_pin"`
	UseMassMentions MessagesChatPermission `json:"use_mass_mentions"`
	SeeInviteLink   MessagesChatPermission `json:"see_invite_link"`
	Call            MessagesChatPermission `json:"call"`
	ChangeAdmins    MessagesChatPermission `json:"change_admins"`
	ChangeStyle     MessagesChatPermission `json:"change_style"`
}

// MessagesConversationPeer struct.
type MessagesConversationPeer struct {
	ID      int    `json:"id"`
	LocalID int    `json:"local_id"`
	Type    string `json:"type"`
}

// MessagesConversationPushSettings struct.
type MessagesConversationPushSettings struct {
	DisabledUntil        int         `json:"disabled_until"`
	DisabledForever      BaseBoolInt `json:"disabled_forever"`
	NoSound              BaseBoolInt `json:"no_sound"`
	DisabledMentions     BaseBoolInt `json:"disabled_mentions"`
	DisabledMassMentions BaseBoolInt `json:"disabled_mass_mentions"`
}

// MessagesConversationWithMessage struct.
type MessagesConversationWithMessage struct {
	Conversation MessagesConversation `json:"conversation"`
	// BUG(VK): https://vk.com/bug229134
	LastMessage MessagesMessage `json:"last_message"`
}

// MessagesDialog struct.
type MessagesDialog struct {
	Important  int             `json:"important"`
	InRead     int             `json:"in_read"`
	Message    MessagesMessage `json:"message"`
	OutRead    int             `json:"out_read"`
	Unanswered int             `json:"unanswered"`
	Unread     int             `json:"unread"`
}

// MessagesHistoryAttachment struct.
type MessagesHistoryAttachment struct {
	Attachment MessagesHistoryMessageAttachment `json:"attachment"`
	MessageID  int                              `json:"message_id"` // Message ID
	FromID     int                              `json:"from_id"`
}

// MessagesHistoryMessageAttachment struct.
type MessagesHistoryMessageAttachment struct {
	Audio  AudioAudio  `json:"audio"`
	Doc    DocsDoc     `json:"doc"`
	Link   BaseLink    `json:"link"`
	Market BaseLink    `json:"market"`
	Photo  PhotosPhoto `json:"photo"`
	Share  BaseLink    `json:"share"`
	Type   string      `json:"type"`
	Video  VideoVideo  `json:"video"`
	Wall   BaseLink    `json:"wall"`
}

// MessagesLastActivity struct.
type MessagesLastActivity struct {
	Online BaseBoolInt `json:"online"` // Information whether user is online
	Time   int         `json:"time"`   // Time when user was online in Unixtime
}

// MessagesLongPollParams struct.
type MessagesLongPollParams struct {
	Key    string `json:"key"`    // Key
	Pts    int    `json:"pts"`    // Persistent timestamp
	Server string `json:"server"` // Server URL
	Ts     int    `json:"ts"`     // Timestamp
}

// MessagesMessageAction status.
const (
	ChatPhotoUpdate        = "chat_photo_update"
	ChatPhotoRemove        = "chat_photo_remove"
	ChatCreate             = "chat_create"
	ChatTitleUpdate        = "chat_title_update"
	ChatInviteUser         = "chat_invite_user"
	ChatKickUser           = "chat_kick_user"
	ChatPinMessage         = "chat_pin_message"
	ChatUnpinMessage       = "chat_unpin_message"
	ChatInviteUserByLink   = "chat_invite_user_by_link"
	AcceptedMessageRequest = "accepted_message_request"
)

// MessagesMessageAction struct.
type MessagesMessageAction struct {
	ConversationMessageID int `json:"conversation_message_id"` // Message ID

	// Email address for chat_invite_user or chat_kick_user actions.
	Email    string                     `json:"email"`
	MemberID int                        `json:"member_id"` // User or email peer ID
	Message  string                     `json:"message"`   // Message body of related message
	Photo    MessagesMessageActionPhoto `json:"photo"`

	// New chat title for chat_create and chat_title_update actions.
	Text string `json:"text"`
	Type string `json:"type"`
}

// MessagesMessageActionPhoto struct.
type MessagesMessageActionPhoto struct {
	Photo100 string `json:"photo_100"` // URL of the preview image with 100px in width
	Photo200 string `json:"photo_200"` // URL of the preview image with 200px in width
	Photo50  string `json:"photo_50"`  // URL of the preview image with 50px in width
}

// MessagesMessageAttachment struct.
type MessagesMessageAttachment struct {
	Audio             AudioAudio        `json:"audio"`
	Doc               DocsDoc           `json:"doc"`
	Gift              GiftsLayout       `json:"gift"`
	Link              BaseLink          `json:"link"`
	Market            MarketMarketItem  `json:"market"`
	MarketMarketAlbum MarketMarketAlbum `json:"market_market_album"`
	Photo             PhotosPhoto       `json:"photo"`
	Sticker           BaseSticker       `json:"sticker"`
	Type              string            `json:"type"`
	Video             VideoVideo        `json:"video"`
	Wall              WallWallpost      `json:"wall"`
	WallReply         WallWallComment   `json:"wall_reply"`
	AudioMessage      DocsDoc           `json:"audio_message"`
	Graffiti          DocsDoc           `json:"graffiti"`
	Poll              PollsPoll         `json:"poll"`
	Call              MessageCall       `json:"call"`
	Story             StoriesStory      `json:"story"`
	Podcast           PodcastsEpisode   `json:"podcast"`
}

// State in which call ended up.
//
// TODO: v3 type CallEndState.
const (
	CallEndStateCanceledByInitiator = "canceled_by_initiator"
	CallEndStateCanceledByReceiver  = "canceled_by_receiver"
	CallEndStateReached             = "reached"
)

// MessageCall struct.
type MessageCall struct {
	InitiatorID int         `json:"initiator_id"`
	ReceiverID  int         `json:"receiver_id"`
	State       string      `json:"state"`
	Time        int         `json:"time"`
	Duration    int         `json:"duration"`
	Video       BaseBoolInt `json:"video"`
}

// MessagesPinnedMessage struct.
type MessagesPinnedMessage struct {
	Attachments []MessagesMessageAttachment `json:"attachments"`

	// Unique auto-incremented number for all Messages with this peer.
	ConversationMessageID int `json:"conversation_message_id"`

	// Date when the message has been sent in Unixtime.
	Date int `json:"date"`

	// Message author's ID.
	FromID       int                `json:"from_id"`
	FwdMessages  []*MessagesMessage `json:"fwd_Messages"`
	Geo          BaseMessageGeo     `json:"geo"`
	ID           int                `json:"id"`      // Message ID
	PeerID       int                `json:"peer_id"` // Peer ID
	ReplyMessage *MessagesMessage   `json:"reply_message"`
	Text         string             `json:"text"` // Message text
}

// MessagesUserXtrInvitedBy struct.
type MessagesUserXtrInvitedBy struct{}

// MessagesForward struct.
type MessagesForward struct {
	// Message owner. It is worth passing it on if you want to forward messages
	// from the community to a dialog.
	OwnerID int `json:"owner_id,omitempty"`

	// Identifier of the place from which the messages are to be sent.
	PeerID int `json:"peer_id,omitempty"`

	// Messages can be passed to conversation_message_ids array:
	//
	// - that are in a personal dialog with the bot;
	//
	// - which are outbound messages from the bot;
	//
	// - written after the bot has entered the conversation.
	ConversationMessageIDs []int `json:"conversation_message_ids,omitempty"`
	MessageIDs             []int `json:"message_ids,omitempty"`

	// Reply to messages. It is worth passing if you want to reply to messages
	// in the chat room where the messages are. In this case there should be
	// only one element in the conversation_message_ids/message_ids.
	IsReply bool `json:"is_reply,omitempty"`
}

// ToJSON returns the JSON encoding of MessagesForward.
func (forward MessagesForward) ToJSON() string {
	b, _ := json.Marshal(forward)
	return string(b)
}
