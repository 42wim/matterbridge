package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// GroupsAddAddressResponse struct.
type GroupsAddAddressResponse object.GroupsAddress

// GroupsAddAddress groups.addAddress.
//
// https://vk.com/dev/groups.addAddress
func (vk *VK) GroupsAddAddress(params Params) (response GroupsAddAddressResponse, err error) {
	err = vk.RequestUnmarshal("groups.addAddress", &response, params)
	return
}

// GroupsAddCallbackServerResponse struct.
type GroupsAddCallbackServerResponse struct {
	ServerID int `json:"server_id"`
}

// GroupsAddCallbackServer callback API server to the community.
//
// https://vk.com/dev/groups.addCallbackServer
func (vk *VK) GroupsAddCallbackServer(params Params) (response GroupsAddCallbackServerResponse, err error) {
	err = vk.RequestUnmarshal("groups.addCallbackServer", &response, params)
	return
}

// GroupsAddLinkResponse struct.
type GroupsAddLinkResponse object.GroupsGroupLink

// GroupsAddLink allows to add a link to the community.
//
// https://vk.com/dev/groups.addLink
func (vk *VK) GroupsAddLink(params Params) (response GroupsAddLinkResponse, err error) {
	err = vk.RequestUnmarshal("groups.addLink", &response, params)
	return
}

// GroupsApproveRequest allows to approve join request to the community.
//
// https://vk.com/dev/groups.approveRequest
func (vk *VK) GroupsApproveRequest(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.approveRequest", &response, params)
	return
}

// GroupsBan adds a user or a group to the community blacklist.
//
// https://vk.com/dev/groups.ban
func (vk *VK) GroupsBan(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.ban", &response, params)
	return
}

// GroupsCreateResponse struct.
type GroupsCreateResponse object.GroupsGroup

// GroupsCreate creates a new community.
//
// https://vk.com/dev/groups.create
func (vk *VK) GroupsCreate(params Params) (response GroupsCreateResponse, err error) {
	err = vk.RequestUnmarshal("groups.create", &response, params)
	return
}

// GroupsDeleteAddress groups.deleteAddress.
//
// https://vk.com/dev/groups.deleteAddress
func (vk *VK) GroupsDeleteAddress(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.deleteAddress", &response, params)
	return
}

// GroupsDeleteCallbackServer callback API server from the community.
//
// https://vk.com/dev/groups.deleteCallbackServer
func (vk *VK) GroupsDeleteCallbackServer(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.deleteCallbackServer", &response, params)
	return
}

// GroupsDeleteLink allows to delete a link from the community.
//
// https://vk.com/dev/groups.deleteLink
func (vk *VK) GroupsDeleteLink(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.deleteLink", &response, params)
	return
}

// GroupsDisableOnline disables "online" status in the community.
//
// https://vk.com/dev/groups.disableOnline
func (vk *VK) GroupsDisableOnline(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.disableOnline", &response, params)
	return
}

// GroupsEdit edits a community.
//
// https://vk.com/dev/groups.edit
func (vk *VK) GroupsEdit(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.edit", &response, params)
	return
}

// GroupsEditAddressResponse struct.
type GroupsEditAddressResponse object.GroupsAddress

// GroupsEditAddress groups.editAddress.
//
// https://vk.com/dev/groups.editAddress
func (vk *VK) GroupsEditAddress(params Params) (response GroupsEditAddressResponse, err error) {
	err = vk.RequestUnmarshal("groups.editAddress", &response, params)
	return
}

// GroupsEditCallbackServer edits Callback API server in the community.
//
// https://vk.com/dev/groups.editCallbackServer
func (vk *VK) GroupsEditCallbackServer(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.editCallbackServer", &response, params)
	return
}

// GroupsEditLink allows to edit a link in the community.
//
// https://vk.com/dev/groups.editLink
func (vk *VK) GroupsEditLink(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.editLink", &response, params)
	return
}

// GroupsEditManager allows to add, remove or edit the community manager .
//
// https://vk.com/dev/groups.editManager
func (vk *VK) GroupsEditManager(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.editManager", &response, params)
	return
}

// GroupsEnableOnline enables "online" status in the community.
//
// https://vk.com/dev/groups.enableOnline
func (vk *VK) GroupsEnableOnline(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.enableOnline", &response, params)
	return
}

// GroupsGetResponse struct.
type GroupsGetResponse struct {
	Count int   `json:"count"`
	Items []int `json:"items"`
}

// GroupsGet returns a list of the communities to which a user belongs.
//
//	extended=0
//
// https://vk.com/dev/groups.get
func (vk *VK) GroupsGet(params Params) (response GroupsGetResponse, err error) {
	err = vk.RequestUnmarshal("groups.get", &response, params, Params{"extended": false})

	return
}

// GroupsGetExtendedResponse struct.
type GroupsGetExtendedResponse struct {
	Count int                  `json:"count"`
	Items []object.GroupsGroup `json:"items"`
}

// GroupsGetExtended returns a list of the communities to which a user belongs.
//
//	extended=1
//
// https://vk.com/dev/groups.get
func (vk *VK) GroupsGetExtended(params Params) (response GroupsGetExtendedResponse, err error) {
	err = vk.RequestUnmarshal("groups.get", &response, params, Params{"extended": true})

	return
}

// GroupsGetAddressesResponse struct.
type GroupsGetAddressesResponse struct {
	Count int                    `json:"count"`
	Items []object.GroupsAddress `json:"items"`
}

// GroupsGetAddresses groups.getAddresses.
//
// https://vk.com/dev/groups.getAddresses
func (vk *VK) GroupsGetAddresses(params Params) (response GroupsGetAddressesResponse, err error) {
	err = vk.RequestUnmarshal("groups.getAddresses", &response, params)
	return
}

// GroupsGetBannedResponse struct.
type GroupsGetBannedResponse struct {
	Count int                            `json:"count"`
	Items []object.GroupsOwnerXtrBanInfo `json:"items"`
}

// GroupsGetBanned returns a list of users on a community blacklist.
//
// https://vk.com/dev/groups.getBanned
func (vk *VK) GroupsGetBanned(params Params) (response GroupsGetBannedResponse, err error) {
	err = vk.RequestUnmarshal("groups.getBanned", &response, params)
	return
}

// GroupsGetByIDResponse struct.
type GroupsGetByIDResponse []object.GroupsGroup

// GroupsGetByID returns information about communities by their IDs.
//
// https://vk.com/dev/groups.getById
func (vk *VK) GroupsGetByID(params Params) (response GroupsGetByIDResponse, err error) {
	err = vk.RequestUnmarshal("groups.getById", &response, params)
	return
}

// GroupsGetCallbackConfirmationCodeResponse struct.
type GroupsGetCallbackConfirmationCodeResponse struct {
	Code string `json:"code"`
}

// GroupsGetCallbackConfirmationCode returns Callback API confirmation code for the community.
//
// https://vk.com/dev/groups.getCallbackConfirmationCode
func (vk *VK) GroupsGetCallbackConfirmationCode(params Params) (
	response GroupsGetCallbackConfirmationCodeResponse,
	err error,
) {
	err = vk.RequestUnmarshal("groups.getCallbackConfirmationCode", &response, params)
	return
}

// GroupsGetCallbackServersResponse struct.
type GroupsGetCallbackServersResponse struct {
	Count int                           `json:"count"`
	Items []object.GroupsCallbackServer `json:"items"`
}

// GroupsGetCallbackServers receives a list of Callback API servers from the community.
//
// https://vk.com/dev/groups.getCallbackServers
func (vk *VK) GroupsGetCallbackServers(params Params) (response GroupsGetCallbackServersResponse, err error) {
	err = vk.RequestUnmarshal("groups.getCallbackServers", &response, params)
	return
}

// GroupsGetCallbackSettingsResponse struct.
type GroupsGetCallbackSettingsResponse object.GroupsCallbackSettings

// GroupsGetCallbackSettings returns Callback API notifications settings.
//
// BUG(VK): MessageEdit always 0 https://vk.com/bugtracker?act=show&id=86762
//
// https://vk.com/dev/groups.getCallbackSettings
func (vk *VK) GroupsGetCallbackSettings(params Params) (response GroupsGetCallbackSettingsResponse, err error) {
	err = vk.RequestUnmarshal("groups.getCallbackSettings", &response, params)
	return
}

// GroupsGetCatalogResponse struct.
type GroupsGetCatalogResponse struct {
	Count int                  `json:"count"`
	Items []object.GroupsGroup `json:"items"`
}

// GroupsGetCatalog returns communities list for a catalog category.
//
// Deprecated: This method is deprecated and may be disabled soon, please avoid
//
// https://vk.com/dev/groups.getCatalog
func (vk *VK) GroupsGetCatalog(params Params) (response GroupsGetCatalogResponse, err error) {
	err = vk.RequestUnmarshal("groups.getCatalog", &response, params)
	return
}

// GroupsGetCatalogInfoResponse struct.
type GroupsGetCatalogInfoResponse struct {
	Enabled    object.BaseBoolInt           `json:"enabled"`
	Categories []object.GroupsGroupCategory `json:"categories"`
}

// GroupsGetCatalogInfo returns categories list for communities catalog.
//
//	extended=0
//
// https://vk.com/dev/groups.getCatalogInfo
func (vk *VK) GroupsGetCatalogInfo(params Params) (response GroupsGetCatalogInfoResponse, err error) {
	err = vk.RequestUnmarshal("groups.getCatalogInfo", &response, params, Params{"extended": false})

	return
}

// GroupsGetCatalogInfoExtendedResponse struct.
type GroupsGetCatalogInfoExtendedResponse struct {
	Enabled    object.BaseBoolInt               `json:"enabled"`
	Categories []object.GroupsGroupCategoryFull `json:"categories"`
}

// GroupsGetCatalogInfoExtended returns categories list for communities catalog.
//
//	extended=1
//
// https://vk.com/dev/groups.getCatalogInfo
func (vk *VK) GroupsGetCatalogInfoExtended(params Params) (response GroupsGetCatalogInfoExtendedResponse, err error) {
	err = vk.RequestUnmarshal("groups.getCatalogInfo", &response, params, Params{"extended": true})

	return
}

// GroupsGetInvitedUsersResponse struct.
type GroupsGetInvitedUsersResponse struct {
	Count int                `json:"count"`
	Items []object.UsersUser `json:"items"`
}

// GroupsGetInvitedUsers returns invited users list of a community.
//
// https://vk.com/dev/groups.getInvitedUsers
func (vk *VK) GroupsGetInvitedUsers(params Params) (response GroupsGetInvitedUsersResponse, err error) {
	err = vk.RequestUnmarshal("groups.getInvitedUsers", &response, params)
	return
}

// GroupsGetInvitesResponse struct.
type GroupsGetInvitesResponse struct {
	Count int                              `json:"count"`
	Items []object.GroupsGroupXtrInvitedBy `json:"items"`
}

// GroupsGetInvites returns a list of invitations to join communities and events.
//
// https://vk.com/dev/groups.getInvites
func (vk *VK) GroupsGetInvites(params Params) (response GroupsGetInvitesResponse, err error) {
	err = vk.RequestUnmarshal("groups.getInvites", &response, params)
	return
}

// GroupsGetInvitesExtendedResponse struct.
type GroupsGetInvitesExtendedResponse struct {
	Count int                              `json:"count"`
	Items []object.GroupsGroupXtrInvitedBy `json:"items"`
	object.ExtendedResponse
}

// GroupsGetInvitesExtended returns a list of invitations to join communities and events.
//
// https://vk.com/dev/groups.getInvites
func (vk *VK) GroupsGetInvitesExtended(params Params) (response GroupsGetInvitesExtendedResponse, err error) {
	err = vk.RequestUnmarshal("groups.getInvites", &response, params)
	return
}

// GroupsGetLongPollServerResponse struct.
type GroupsGetLongPollServerResponse object.GroupsLongPollServer

// GroupsGetLongPollServer returns data for Bots Long Poll API connection.
//
// https://vk.com/dev/groups.getLongPollServer
func (vk *VK) GroupsGetLongPollServer(params Params) (response GroupsGetLongPollServerResponse, err error) {
	err = vk.RequestUnmarshal("groups.getLongPollServer", &response, params)
	return
}

// GroupsGetLongPollSettingsResponse struct.
type GroupsGetLongPollSettingsResponse object.GroupsLongPollSettings

// GroupsGetLongPollSettings returns Bots Long Poll API settings.
//
// https://vk.com/dev/groups.getLongPollSettings
func (vk *VK) GroupsGetLongPollSettings(params Params) (response GroupsGetLongPollSettingsResponse, err error) {
	err = vk.RequestUnmarshal("groups.getLongPollSettings", &response, params)
	return
}

// GroupsGetMembersResponse struct.
type GroupsGetMembersResponse struct {
	Count int   `json:"count"`
	Items []int `json:"items"`
}

// GroupsGetMembers returns a list of community members.
//
// https://vk.com/dev/groups.getMembers
func (vk *VK) GroupsGetMembers(params Params) (response GroupsGetMembersResponse, err error) {
	err = vk.RequestUnmarshal("groups.getMembers", &response, params, Params{"filter": ""})

	return
}

// GroupsGetMembersFieldsResponse struct.
type GroupsGetMembersFieldsResponse struct {
	Count int                `json:"count"`
	Items []object.UsersUser `json:"items"`
}

// GroupsGetMembersFields returns a list of community members.
//
// https://vk.com/dev/groups.getMembers
func (vk *VK) GroupsGetMembersFields(params Params) (response GroupsGetMembersFieldsResponse, err error) {
	reqParams := make(Params)
	if v, prs := params["fields"]; v == "" || !prs {
		reqParams["fields"] = "id"
	}

	err = vk.RequestUnmarshal("groups.getMembers", &response, params, reqParams)

	return
}

// GroupsGetMembersFilterManagersResponse struct.
type GroupsGetMembersFilterManagersResponse struct {
	Count int                                   `json:"count"`
	Items []object.GroupsMemberRoleXtrUsersUser `json:"items"`
}

// GroupsGetMembersFilterManagers returns a list of community members.
//
//	filter=managers
//
// https://vk.com/dev/groups.getMembers
func (vk *VK) GroupsGetMembersFilterManagers(params Params) (
	response GroupsGetMembersFilterManagersResponse,
	err error,
) {
	err = vk.RequestUnmarshal("groups.getMembers", &response, params, Params{"filter": "managers"})

	return
}

// GroupsGetOnlineStatusResponse struct.
type GroupsGetOnlineStatusResponse object.GroupsOnlineStatus

// GroupsGetOnlineStatus returns a community's online status.
//
// https://vk.com/dev/groups.getOnlineStatus
func (vk *VK) GroupsGetOnlineStatus(params Params) (response GroupsGetOnlineStatusResponse, err error) {
	err = vk.RequestUnmarshal("groups.getOnlineStatus", &response, params)
	return
}

// GroupsGetRequestsResponse struct.
type GroupsGetRequestsResponse struct {
	Count int   `json:"count"`
	Items []int `json:"items"`
}

// GroupsGetRequests returns a list of requests to the community.
//
// https://vk.com/dev/groups.getRequests
func (vk *VK) GroupsGetRequests(params Params) (response GroupsGetRequestsResponse, err error) {
	err = vk.RequestUnmarshal("groups.getRequests", &response, params, Params{"fields": ""})

	return
}

// GroupsGetRequestsFieldsResponse struct.
type GroupsGetRequestsFieldsResponse struct {
	Count int                `json:"count"`
	Items []object.UsersUser `json:"items"`
}

// GroupsGetRequestsFields returns a list of requests to the community.
//
// https://vk.com/dev/groups.getRequests
func (vk *VK) GroupsGetRequestsFields(params Params) (response GroupsGetRequestsFieldsResponse, err error) {
	reqParams := make(Params)
	if v, prs := params["fields"]; v == "" || !prs {
		reqParams["fields"] = "id"
	}

	err = vk.RequestUnmarshal("groups.getRequests", &response, params, reqParams)

	return
}

// GroupsGetSettingsResponse struct.
type GroupsGetSettingsResponse object.GroupsGroupSettings

// GroupsGetSettings returns community settings.
//
// https://vk.com/dev/groups.getSettings
func (vk *VK) GroupsGetSettings(params Params) (response GroupsGetSettingsResponse, err error) {
	err = vk.RequestUnmarshal("groups.getSettings", &response, params)
	return
}

// GroupsGetTagListResponse struct.
type GroupsGetTagListResponse []object.GroupsTag

// GroupsGetTagList returns community tags list.
//
// https://vk.com/dev/groups.getTagList
func (vk *VK) GroupsGetTagList(params Params) (response GroupsGetTagListResponse, err error) {
	err = vk.RequestUnmarshal("groups.getTagList", &response, params)
	return
}

// GroupsGetTokenPermissionsResponse struct.
type GroupsGetTokenPermissionsResponse object.GroupsTokenPermissions

// GroupsGetTokenPermissions returns permissions scope for the community's access_token.
//
// https://vk.com/dev/groups.getTokenPermissions
func (vk *VK) GroupsGetTokenPermissions(params Params) (response GroupsGetTokenPermissionsResponse, err error) {
	err = vk.RequestUnmarshal("groups.getTokenPermissions", &response, params)
	return
}

// GroupsInvite allows to invite friends to the community.
//
// https://vk.com/dev/groups.invite
func (vk *VK) GroupsInvite(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.invite", &response, params)
	return
}

// GroupsIsMember returns information specifying whether a user is a member of a community.
//
//	extended=0
//
// https://vk.com/dev/groups.isMember
func (vk *VK) GroupsIsMember(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.isMember", &response, params, Params{"extended": false})

	return
}

// GroupsIsMemberExtendedResponse struct.
type GroupsIsMemberExtendedResponse struct {
	Invitation object.BaseBoolInt `json:"invitation"` // Information whether user has been invited to the group
	Member     object.BaseBoolInt `json:"member"`     // Information whether user is a member of the group
	Request    object.BaseBoolInt `json:"request"`    // Information whether user has send request to the group
	CanInvite  object.BaseBoolInt `json:"can_invite"` // Information whether user can be invite
	CanRecall  object.BaseBoolInt `json:"can_recall"` // Information whether user's invite to the group can be recalled
}

// GroupsIsMemberExtended returns information specifying whether a user is a member of a community.
//
//	extended=1
//
// https://vk.com/dev/groups.isMember
func (vk *VK) GroupsIsMemberExtended(params Params) (response GroupsIsMemberExtendedResponse, err error) {
	err = vk.RequestUnmarshal("groups.isMember", &response, params, Params{"extended": true})

	return
}

// GroupsIsMemberUserIDsExtendedResponse struct.
type GroupsIsMemberUserIDsExtendedResponse []object.GroupsMemberStatusFull

// GroupsIsMemberUserIDsExtended returns information specifying whether a user is a member of a community.
//
//	extended=1
//	need user_ids
//
// https://vk.com/dev/groups.isMember
func (vk *VK) GroupsIsMemberUserIDsExtended(params Params) (response GroupsIsMemberUserIDsExtendedResponse, err error) {
	err = vk.RequestUnmarshal("groups.isMember", &response, params, Params{"extended": true})

	return
}

// GroupsIsMemberUserIDsResponse struct.
type GroupsIsMemberUserIDsResponse []object.GroupsMemberStatus

// GroupsIsMemberUserIDs returns information specifying whether a user is a member of a community.
//
//	extended=0
//	need user_ids
//
// https://vk.com/dev/groups.isMember
func (vk *VK) GroupsIsMemberUserIDs(params Params) (response GroupsIsMemberUserIDsResponse, err error) {
	err = vk.RequestUnmarshal("groups.isMember", &response, params, Params{"extended": false})

	return
}

// GroupsJoin with this method you can join the group or public page, and also confirm your participation in an event.
//
// https://vk.com/dev/groups.join
func (vk *VK) GroupsJoin(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.join", &response, params)
	return
}

// GroupsLeave with this method you can leave a group, public page, or event.
//
// https://vk.com/dev/groups.leave
func (vk *VK) GroupsLeave(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.leave", &response, params)
	return
}

// GroupsRemoveUser removes a user from the community.
//
// https://vk.com/dev/groups.removeUser
func (vk *VK) GroupsRemoveUser(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.removeUser", &response, params)
	return
}

// GroupsReorderLink allows to reorder links in the community.
//
// https://vk.com/dev/groups.reorderLink
func (vk *VK) GroupsReorderLink(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.reorderLink", &response, params)
	return
}

// GroupsSearchResponse struct.
type GroupsSearchResponse struct {
	Count int                  `json:"count"`
	Items []object.GroupsGroup `json:"items"`
}

// GroupsSearch returns a list of communities matching the search criteria.
//
// https://vk.com/dev/groups.search
func (vk *VK) GroupsSearch(params Params) (response GroupsSearchResponse, err error) {
	err = vk.RequestUnmarshal("groups.search", &response, params)
	return
}

// GroupsSetCallbackSettings allow to set notifications settings for Callback API.
//
// https://vk.com/dev/groups.setCallbackSettings
func (vk *VK) GroupsSetCallbackSettings(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.setCallbackSettings", &response, params)
	return
}

// GroupsSetLongPollSettings allows to set Bots Long Poll API settings in the community.
//
// https://vk.com/dev/groups.setLongPollSettings
func (vk *VK) GroupsSetLongPollSettings(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.setLongPollSettings", &response, params)
	return
}

// GroupsSetSettings sets community settings.
//
// https://vk.com/dev/groups.setSettings
func (vk *VK) GroupsSetSettings(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.setSettings", &response, params)
	return
}

// GroupsSetUserNote allows to create or edit a note about a user as part
// of the user's correspondence with the community.
//
// https://vk.com/dev/groups.setUserNote
func (vk *VK) GroupsSetUserNote(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.setUserNote", &response, params)
	return
}

// GroupsTagAdd allows to add a new tag to the community.
//
// https://vk.com/dev/groups.tagAdd
func (vk *VK) GroupsTagAdd(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.tagAdd", &response, params)
	return
}

// GroupsTagBind allows to "bind" and "unbind" community tags to conversations.
//
// https://vk.com/dev/groups.tagBind
func (vk *VK) GroupsTagBind(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.tagBind", &response, params)
	return
}

// GroupsTagDelete allows to remove a community tag
//
// The remote tag will be automatically "unbind" from all conversations to
// which it was "bind" earlier.
//
// https://vk.com/dev/groups.tagDelete
func (vk *VK) GroupsTagDelete(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.tagDelete", &response, params)
	return
}

// GroupsTagUpdate allows to change an existing tag.
//
// https://vk.com/dev/groups.tagUpdate
func (vk *VK) GroupsTagUpdate(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.tagUpdate", &response, params)
	return
}

// GroupsToggleMarket method.
//
// https://vk.com/dev/groups.toggleMarket
func (vk *VK) GroupsToggleMarket(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.toggleMarket", &response, params)
	return
}

// GroupsUnban groups.unban.
//
// https://vk.com/dev/groups.unban
func (vk *VK) GroupsUnban(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("groups.unban", &response, params)
	return
}
