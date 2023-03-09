package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"strconv"

	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/vmihailenco/msgpack/v5"
)

// MessagesAddChatUser adds a new user to a chat.
//
// https://vk.com/dev/messages.addChatUser
func (vk *VK) MessagesAddChatUser(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.addChatUser", &response, params)
	return
}

// MessagesAllowMessagesFromGroup allows sending messages from community to the current user.
//
// https://vk.com/dev/messages.allowMessagesFromGroup
func (vk *VK) MessagesAllowMessagesFromGroup(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.allowMessagesFromGroup", &response, params)
	return
}

// MessagesCreateChat creates a chat with several participants.
//
// https://vk.com/dev/messages.createChat
func (vk *VK) MessagesCreateChat(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.createChat", &response, params)
	return
}

// MessagesDeleteResponse struct.
type MessagesDeleteResponse map[string]int

// DecodeMsgpack funcion.
func (resp *MessagesDeleteResponse) DecodeMsgpack(dec *msgpack.Decoder) error {
	data, err := dec.DecodeRaw()
	if err != nil {
		return err
	}

	var respMap map[int]int

	err = msgpack.Unmarshal(data, &respMap)
	if err != nil {
		return err
	}

	*resp = make(MessagesDeleteResponse)
	for key, val := range respMap {
		(*resp)[strconv.Itoa(key)] = val
	}

	return nil
}

// MessagesDelete deletes one or more messages.
//
// https://vk.com/dev/messages.delete
func (vk *VK) MessagesDelete(params Params) (response MessagesDeleteResponse, err error) {
	err = vk.RequestUnmarshal("messages.delete", &response, params)

	return
}

// MessagesDeleteChatPhotoResponse struct.
type MessagesDeleteChatPhotoResponse struct {
	MessageID int                 `json:"message_id"`
	Chat      object.MessagesChat `json:"chat"`
}

// MessagesDeleteChatPhoto deletes a chat's cover picture.
//
// https://vk.com/dev/messages.deleteChatPhoto
func (vk *VK) MessagesDeleteChatPhoto(params Params) (response MessagesDeleteChatPhotoResponse, err error) {
	err = vk.RequestUnmarshal("messages.deleteChatPhoto", &response, params)
	return
}

// MessagesDeleteConversationResponse struct.
type MessagesDeleteConversationResponse struct {
	LastDeletedID int `json:"last_deleted_id"` // Id of the last message, that was deleted
}

// MessagesDeleteConversation deletes private messages in a conversation.
//
// https://vk.com/dev/messages.deleteConversation
func (vk *VK) MessagesDeleteConversation(params Params) (response MessagesDeleteConversationResponse, err error) {
	err = vk.RequestUnmarshal("messages.deleteConversation", &response, params)
	return
}

// MessagesDenyMessagesFromGroup denies sending message from community to the current user.
//
// https://vk.com/dev/messages.denyMessagesFromGroup
func (vk *VK) MessagesDenyMessagesFromGroup(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.denyMessagesFromGroup", &response, params)
	return
}

// MessagesEdit edits the message.
//
// https://vk.com/dev/messages.edit
func (vk *VK) MessagesEdit(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.edit", &response, params)
	return
}

// MessagesEditChat edits the title of a chat.
//
// https://vk.com/dev/messages.editChat
func (vk *VK) MessagesEditChat(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.editChat", &response, params)
	return
}

// MessagesForceCallFinish method.
//
// Deprecated: Use CallsForceFinish
//
// https://vk.com/dev/messages.forceCallFinish
func (vk *VK) MessagesForceCallFinish(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.forceCallFinish", &response, params)
	return
}

// MessagesGetByConversationMessageIDResponse struct.
type MessagesGetByConversationMessageIDResponse struct {
	Count int                      `json:"count"`
	Items []object.MessagesMessage `json:"items"`
	object.ExtendedResponse
}

// MessagesGetByConversationMessageID messages.getByConversationMessageId.
//
// https://vk.com/dev/messages.getByConversationMessageId
func (vk *VK) MessagesGetByConversationMessageID(params Params) (
	response MessagesGetByConversationMessageIDResponse,
	err error,
) {
	err = vk.RequestUnmarshal("messages.getByConversationMessageId", &response, params)
	return
}

// MessagesGetByIDResponse struct.
type MessagesGetByIDResponse struct {
	Count int                      `json:"count"`
	Items []object.MessagesMessage `json:"items"`
}

// MessagesGetByID returns messages by their IDs.
//
//	extended=0
//
// https://vk.com/dev/messages.getById
func (vk *VK) MessagesGetByID(params Params) (response MessagesGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("messages.getById", &response, params, Params{"extended": false})

	return
}

// MessagesGetByIDExtendedResponse struct.
type MessagesGetByIDExtendedResponse struct {
	Count int                      `json:"count"`
	Items []object.MessagesMessage `json:"items"`
	object.ExtendedResponse
}

// MessagesGetByIDExtended returns messages by their IDs.
//
//	extended=1
//
// https://vk.com/dev/messages.getById
func (vk *VK) MessagesGetByIDExtended(params Params) (response MessagesGetByIDExtendedResponse, err error) {
	err = vk.RequestUnmarshal("messages.getById", &response, params, Params{"extended": true})

	return
}

// MessagesGetChatResponse struct.
type MessagesGetChatResponse object.MessagesChat

// MessagesGetChat returns information about a chat.
//
// https://vk.com/dev/messages.getChat
func (vk *VK) MessagesGetChat(params Params) (response MessagesGetChatResponse, err error) {
	err = vk.RequestUnmarshal("messages.getChat", &response, params)
	return
}

// MessagesGetChatChatIDsResponse struct.
type MessagesGetChatChatIDsResponse []object.MessagesChat

// MessagesGetChatChatIDs returns information about a chat.
//
// https://vk.com/dev/messages.getChat
func (vk *VK) MessagesGetChatChatIDs(params Params) (response MessagesGetChatChatIDsResponse, err error) {
	err = vk.RequestUnmarshal("messages.getChat", &response, params)
	return
}

// MessagesGetChatPreviewResponse struct.
type MessagesGetChatPreviewResponse struct {
	Preview object.MessagesChatPreview `json:"preview"`
	object.ExtendedResponse
}

// MessagesGetChatPreview allows to receive chat preview by the invitation link.
//
// https://vk.com/dev/messages.getChatPreview
func (vk *VK) MessagesGetChatPreview(params Params) (response MessagesGetChatPreviewResponse, err error) {
	err = vk.RequestUnmarshal("messages.getChatPreview", &response, params)
	return
}

// MessagesGetConversationMembersResponse struct.
type MessagesGetConversationMembersResponse struct {
	Items []struct {
		MemberID  int                `json:"member_id"`
		JoinDate  int                `json:"join_date"`
		InvitedBy int                `json:"invited_by"`
		IsOwner   object.BaseBoolInt `json:"is_owner,omitempty"`
		IsAdmin   object.BaseBoolInt `json:"is_admin,omitempty"`
		CanKick   object.BaseBoolInt `json:"can_kick,omitempty"`
	} `json:"items"`
	Count            int `json:"count"`
	ChatRestrictions struct {
		OnlyAdminsInvite   object.BaseBoolInt `json:"only_admins_invite"`
		OnlyAdminsEditPin  object.BaseBoolInt `json:"only_admins_edit_pin"`
		OnlyAdminsEditInfo object.BaseBoolInt `json:"only_admins_edit_info"`
		AdminsPromoteUsers object.BaseBoolInt `json:"admins_promote_users"`
	} `json:"chat_restrictions"`
	object.ExtendedResponse
}

// MessagesGetConversationMembers returns a list of IDs of users participating in a conversation.
//
// https://vk.com/dev/messages.getConversationMembers
func (vk *VK) MessagesGetConversationMembers(params Params) (
	response MessagesGetConversationMembersResponse,
	err error,
) {
	err = vk.RequestUnmarshal("messages.getConversationMembers", &response, params)
	return
}

// MessagesGetConversationsResponse struct.
type MessagesGetConversationsResponse struct {
	Count       int                                      `json:"count"`
	Items       []object.MessagesConversationWithMessage `json:"items"`
	UnreadCount int                                      `json:"unread_count"`
	object.ExtendedResponse
}

// MessagesGetConversations returns a list of conversations.
//
// https://vk.com/dev/messages.getConversations
func (vk *VK) MessagesGetConversations(params Params) (response MessagesGetConversationsResponse, err error) {
	err = vk.RequestUnmarshal("messages.getConversations", &response, params)
	return
}

// MessagesGetConversationsByIDResponse struct.
type MessagesGetConversationsByIDResponse struct {
	Count int                           `json:"count"`
	Items []object.MessagesConversation `json:"items"`
}

// MessagesGetConversationsByID returns conversations by their IDs.
//
//	extended=0
//
// https://vk.com/dev/messages.getConversationsById
func (vk *VK) MessagesGetConversationsByID(params Params) (response MessagesGetConversationsByIDResponse, err error) {
	err = vk.RequestUnmarshal("messages.getConversationsById", &response, params, Params{"extended": false})

	return
}

// MessagesGetConversationsByIDExtendedResponse struct.
type MessagesGetConversationsByIDExtendedResponse struct {
	Count int                           `json:"count"`
	Items []object.MessagesConversation `json:"items"`
	object.ExtendedResponse
}

// MessagesGetConversationsByIDExtended returns conversations by their IDs.
//
//	extended=1
//
// https://vk.com/dev/messages.getConversationsById
func (vk *VK) MessagesGetConversationsByIDExtended(params Params) (
	response MessagesGetConversationsByIDExtendedResponse,
	err error,
) {
	err = vk.RequestUnmarshal("messages.getConversationsById", &response, params, Params{"extended": true})

	return
}

// MessagesGetHistoryResponse struct.
type MessagesGetHistoryResponse struct {
	Count int                      `json:"count"`
	Items []object.MessagesMessage `json:"items"`

	// 	extended=1
	object.ExtendedResponse

	// 	extended=1
	Conversations []object.MessagesConversation `json:"conversations,omitempty"`

	// Deprecated: use .Conversations.InRead
	InRead int `json:"in_read,omitempty"`
	// Deprecated: use .Conversations.OutRead
	OutRead int `json:"out_read,omitempty"`
}

// MessagesGetHistory returns message history for the specified user or group chat.
//
// https://vk.com/dev/messages.getHistory
func (vk *VK) MessagesGetHistory(params Params) (response MessagesGetHistoryResponse, err error) {
	err = vk.RequestUnmarshal("messages.getHistory", &response, params)
	return
}

// MessagesGetHistoryAttachmentsResponse struct.
type MessagesGetHistoryAttachmentsResponse struct {
	Items    []object.MessagesHistoryAttachment `json:"items"`
	NextFrom string                             `json:"next_from"`
	object.ExtendedResponse
}

// MessagesGetHistoryAttachments returns media files from the dialog or group chat.
//
// https://vk.com/dev/messages.getHistoryAttachments
func (vk *VK) MessagesGetHistoryAttachments(params Params) (response MessagesGetHistoryAttachmentsResponse, err error) {
	err = vk.RequestUnmarshal("messages.getHistoryAttachments", &response, params)
	return
}

// MessagesGetImportantMessagesResponse struct.
type MessagesGetImportantMessagesResponse struct {
	Messages struct {
		Count int                      `json:"count"`
		Items []object.MessagesMessage `json:"items"`
	} `json:"messages"`
	Conversations []object.MessagesConversation `json:"conversations"`
	object.ExtendedResponse
}

// MessagesGetImportantMessages messages.getImportantMessages.
//
// https://vk.com/dev/messages.getImportantMessages
func (vk *VK) MessagesGetImportantMessages(params Params) (response MessagesGetImportantMessagesResponse, err error) {
	err = vk.RequestUnmarshal("messages.getImportantMessages", &response, params)
	return
}

// MessagesGetIntentUsersResponse struct.
type MessagesGetIntentUsersResponse struct {
	Count    int                      `json:"count"`
	Items    []int                    `json:"items"`
	Profiles []object.MessagesMessage `json:"profiles,omitempty"`
}

// MessagesGetIntentUsers method.
//
// https://vk.com/dev/messages.getIntentUsers
func (vk *VK) MessagesGetIntentUsers(params Params) (response MessagesGetIntentUsersResponse, err error) {
	err = vk.RequestUnmarshal("messages.getIntentUsers", &response, params)
	return
}

// MessagesGetInviteLinkResponse struct.
type MessagesGetInviteLinkResponse struct {
	Link string `json:"link"`
}

// MessagesGetInviteLink receives a link to invite a user to the chat.
//
// https://vk.com/dev/messages.getInviteLink
func (vk *VK) MessagesGetInviteLink(params Params) (response MessagesGetInviteLinkResponse, err error) {
	err = vk.RequestUnmarshal("messages.getInviteLink", &response, params)
	return
}

// MessagesGetLastActivityResponse struct.
type MessagesGetLastActivityResponse object.MessagesLastActivity

// MessagesGetLastActivity returns a user's current status and date of last activity.
//
// https://vk.com/dev/messages.getLastActivity
func (vk *VK) MessagesGetLastActivity(params Params) (response MessagesGetLastActivityResponse, err error) {
	err = vk.RequestUnmarshal("messages.getLastActivity", &response, params)
	return
}

// MessagesGetLongPollHistoryResponse struct.
type MessagesGetLongPollHistoryResponse struct {
	History  [][]int              `json:"history"`
	Groups   []object.GroupsGroup `json:"groups"`
	Messages struct {
		Count int                      `json:"count"`
		Items []object.MessagesMessage `json:"items"`
	} `json:"messages"`
	Profiles []object.UsersUser `json:"profiles"`
	// Chats struct {} `json:"chats"`
	NewPTS        int                           `json:"new_pts"`
	FromPTS       int                           `json:"from_pts"`
	More          object.BaseBoolInt            `json:"chats"`
	Conversations []object.MessagesConversation `json:"conversations"`
}

// MessagesGetLongPollHistory returns updates in user's private messages.
//
// https://vk.com/dev/messages.getLongPollHistory
func (vk *VK) MessagesGetLongPollHistory(params Params) (response MessagesGetLongPollHistoryResponse, err error) {
	err = vk.RequestUnmarshal("messages.getLongPollHistory", &response, params)
	return
}

// MessagesGetLongPollServerResponse struct.
type MessagesGetLongPollServerResponse object.MessagesLongPollParams

// MessagesGetLongPollServer returns data required for connection to a Long Poll server.
//
// https://vk.com/dev/messages.getLongPollServer
func (vk *VK) MessagesGetLongPollServer(params Params) (response MessagesGetLongPollServerResponse, err error) {
	err = vk.RequestUnmarshal("messages.getLongPollServer", &response, params)
	return
}

// MessagesIsMessagesFromGroupAllowedResponse struct.
type MessagesIsMessagesFromGroupAllowedResponse struct {
	IsAllowed object.BaseBoolInt `json:"is_allowed"`
}

// MessagesIsMessagesFromGroupAllowed returns information whether
// sending messages from the community to current user is allowed.
//
// https://vk.com/dev/messages.isMessagesFromGroupAllowed
func (vk *VK) MessagesIsMessagesFromGroupAllowed(params Params) (
	response MessagesIsMessagesFromGroupAllowedResponse,
	err error,
) {
	err = vk.RequestUnmarshal("messages.isMessagesFromGroupAllowed", &response, params)
	return
}

// MessagesJoinChatByInviteLinkResponse struct.
type MessagesJoinChatByInviteLinkResponse struct {
	ChatID int `json:"chat_id"`
}

// MessagesJoinChatByInviteLink allows to enter the chat by the invitation link.
//
// https://vk.com/dev/messages.joinChatByInviteLink
func (vk *VK) MessagesJoinChatByInviteLink(params Params) (response MessagesJoinChatByInviteLinkResponse, err error) {
	err = vk.RequestUnmarshal("messages.joinChatByInviteLink", &response, params)
	return
}

// MessagesMarkAsAnsweredConversation messages.markAsAnsweredConversation.
//
// https://vk.com/dev/messages.markAsAnsweredConversation
func (vk *VK) MessagesMarkAsAnsweredConversation(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.markAsAnsweredConversation", &response, params)
	return
}

// MessagesMarkAsImportantResponse struct.
type MessagesMarkAsImportantResponse []int

// MessagesMarkAsImportant marks and un marks messages as important (starred).
//
// https://vk.com/dev/messages.markAsImportant
func (vk *VK) MessagesMarkAsImportant(params Params) (response MessagesMarkAsImportantResponse, err error) {
	err = vk.RequestUnmarshal("messages.markAsImportant", &response, params)
	return
}

// MessagesMarkAsImportantConversation messages.markAsImportantConversation.
//
// https://vk.com/dev/messages.markAsImportantConversation
func (vk *VK) MessagesMarkAsImportantConversation(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.markAsImportantConversation", &response, params)
	return
}

// MessagesMarkAsRead marks messages as read.
//
// https://vk.com/dev/messages.markAsRead
func (vk *VK) MessagesMarkAsRead(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.markAsRead", &response, params)
	return
}

// MessagesPinResponse struct.
type MessagesPinResponse object.MessagesMessage

// MessagesPin messages.pin.
//
// https://vk.com/dev/messages.pin
func (vk *VK) MessagesPin(params Params) (response MessagesPinResponse, err error) {
	err = vk.RequestUnmarshal("messages.pin", &response, params)
	return
}

// MessagesRemoveChatUser allows the current user to leave a chat or, if the
// current user started the chat, allows the user to remove another user from
// the chat.
//
// https://vk.com/dev/messages.removeChatUser
func (vk *VK) MessagesRemoveChatUser(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.removeChatUser", &response, params)
	return
}

// MessagesRestore restores a deleted message.
//
// https://vk.com/dev/messages.restore
func (vk *VK) MessagesRestore(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.restore", &response, params)
	return
}

// MessagesSearchResponse struct.
type MessagesSearchResponse struct {
	Count int                      `json:"count"`
	Items []object.MessagesMessage `json:"items"`
	object.ExtendedResponse
	Conversations []object.MessagesConversation `json:"conversations,omitempty"`
}

// MessagesSearch returns a list of the current user's private messages that match search criteria.
//
// https://vk.com/dev/messages.search
func (vk *VK) MessagesSearch(params Params) (response MessagesSearchResponse, err error) {
	err = vk.RequestUnmarshal("messages.search", &response, params)
	return
}

// MessagesSearchConversationsResponse struct.
type MessagesSearchConversationsResponse struct {
	Count int                           `json:"count"`
	Items []object.MessagesConversation `json:"items"`
	object.ExtendedResponse
}

// MessagesSearchConversations returns a list of conversations that match search criteria.
//
// https://vk.com/dev/messages.searchConversations
func (vk *VK) MessagesSearchConversations(params Params) (response MessagesSearchConversationsResponse, err error) {
	err = vk.RequestUnmarshal("messages.searchConversations", &response, params)
	return
}

// MessagesSend sends a message.
//
// For user_ids or peer_ids parameters, use MessagesSendUserIDs.
//
// https://vk.com/dev/messages.send
func (vk *VK) MessagesSend(params Params) (response int, err error) {
	reqParams := Params{
		"user_ids": "",
		"peer_ids": "",
	}

	err = vk.RequestUnmarshal("messages.send", &response, params, reqParams)

	return
}

// MessagesSendUserIDsResponse struct.
//
// TODO: v3 rename MessagesSendPeerIDsResponse - user_ids outdated.
type MessagesSendUserIDsResponse []struct {
	PeerID                int   `json:"peer_id"`
	MessageID             int   `json:"message_id"`
	ConversationMessageID int   `json:"conversation_message_id"`
	Error                 Error `json:"error"`
}

// MessagesSendPeerIDs sends a message.
//
//	need peer_ids;
//
// https://vk.com/dev/messages.send
func (vk *VK) MessagesSendPeerIDs(params Params) (response MessagesSendUserIDsResponse, err error) {
	err = vk.RequestUnmarshal("messages.send", &response, params)
	return
}

// MessagesSendUserIDs sends a message.
//
//	need user_ids or peer_ids;
//
// https://vk.com/dev/messages.send
//
// Deprecated: user_ids outdated, use MessagesSendPeerIDs.
func (vk *VK) MessagesSendUserIDs(params Params) (response MessagesSendUserIDsResponse, err error) {
	return vk.MessagesSendPeerIDs(params)
}

// MessagesSendMessageEventAnswer method.
//
// https://vk.com/dev/messages.sendMessageEventAnswer
func (vk *VK) MessagesSendMessageEventAnswer(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.sendMessageEventAnswer", &response, params)
	return
}

// MessagesSendSticker sends a message.
//
// https://vk.com/dev/messages.sendSticker
func (vk *VK) MessagesSendSticker(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.sendSticker", &response, params, Params{"user_ids": ""})

	return
}

// MessagesSetActivity changes the status of a user as typing in a conversation.
//
// https://vk.com/dev/messages.setActivity
func (vk *VK) MessagesSetActivity(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.setActivity", &response, params)
	return
}

// MessagesSetChatPhotoResponse struct.
type MessagesSetChatPhotoResponse struct {
	MessageID int                 `json:"message_id"`
	Chat      object.MessagesChat `json:"chat"`
}

// MessagesSetChatPhoto sets a previously-uploaded picture as the cover picture of a chat.
//
// https://vk.com/dev/messages.setChatPhoto
func (vk *VK) MessagesSetChatPhoto(params Params) (response MessagesSetChatPhotoResponse, err error) {
	err = vk.RequestUnmarshal("messages.setChatPhoto", &response, params)
	return
}

// MessagesStartCallResponse struct.
type MessagesStartCallResponse struct {
	JoinLink string `json:"join_link"`
	CallID   string `json:"call_id"`
}

// MessagesStartCall method.
//
// Deprecated: Use CallsStart
//
// https://vk.com/dev/messages.startCall
func (vk *VK) MessagesStartCall(params Params) (response MessagesStartCallResponse, err error) {
	err = vk.RequestUnmarshal("messages.startCall", &response, params)
	return
}

// MessagesUnpin messages.unpin.
//
// https://vk.com/dev/messages.unpin
func (vk *VK) MessagesUnpin(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("messages.unpin", &response, params)
	return
}
