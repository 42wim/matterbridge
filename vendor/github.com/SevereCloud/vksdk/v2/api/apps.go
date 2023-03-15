package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// AppsAddUsersToTestingGroup method.
//
// https://vk.com/dev/apps.addUsersToTestingGroup
func (vk *VK) AppsAddUsersToTestingGroup(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("apps.addUsersToTestingGroup", &response, params)
	return
}

// AppsDeleteAppRequests deletes all request notifications from the current app.
//
// https://vk.com/dev/apps.deleteAppRequests
func (vk *VK) AppsDeleteAppRequests(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("apps.deleteAppRequests", &response, params)
	return
}

// AppsGetResponse struct.
type AppsGetResponse struct {
	Count int              `json:"count"`
	Items []object.AppsApp `json:"items"`
	object.ExtendedResponse
}

// AppsGet returns applications data.
//
// https://vk.com/dev/apps.get
func (vk *VK) AppsGet(params Params) (response AppsGetResponse, err error) {
	err = vk.RequestUnmarshal("apps.get", &response, params)
	return
}

// AppsGetCatalogResponse struct.
type AppsGetCatalogResponse struct {
	Count int              `json:"count"`
	Items []object.AppsApp `json:"items"`
	object.ExtendedResponse
}

// AppsGetCatalog returns a list of applications (apps) available to users in the App Catalog.
//
// https://vk.com/dev/apps.getCatalog
func (vk *VK) AppsGetCatalog(params Params) (response AppsGetCatalogResponse, err error) {
	err = vk.RequestUnmarshal("apps.getCatalog", &response, params)
	return
}

// AppsGetFriendsListResponse struct.
type AppsGetFriendsListResponse struct {
	Count int   `json:"count"`
	Items []int `json:"items"`
}

// AppsGetFriendsList creates friends list for requests and invites in current app.
//
//	extended=0
//
// https://vk.com/dev/apps.getFriendsList
func (vk *VK) AppsGetFriendsList(params Params) (response AppsGetFriendsListResponse, err error) {
	err = vk.RequestUnmarshal("apps.getFriendsList", &response, params, Params{"extended": false})

	return
}

// AppsGetFriendsListExtendedResponse struct.
type AppsGetFriendsListExtendedResponse struct {
	Count int                `json:"count"`
	Items []object.UsersUser `json:"items"`
}

// AppsGetFriendsListExtended creates friends list for requests and invites in current app.
//
//	extended=1
//
// https://vk.com/dev/apps.getFriendsList
func (vk *VK) AppsGetFriendsListExtended(params Params) (response AppsGetFriendsListExtendedResponse, err error) {
	err = vk.RequestUnmarshal("apps.getFriendsList", &response, params, Params{"extended": true})

	return
}

// AppsGetLeaderboardResponse struct.
type AppsGetLeaderboardResponse struct {
	Count int                      `json:"count"`
	Items []object.AppsLeaderboard `json:"items"`
}

// AppsGetLeaderboard returns players rating in the game.
//
//	extended=0
//
// https://vk.com/dev/apps.getLeaderboard
func (vk *VK) AppsGetLeaderboard(params Params) (response AppsGetLeaderboardResponse, err error) {
	err = vk.RequestUnmarshal("apps.getLeaderboard", &response, params, Params{"extended": false})

	return
}

// AppsGetLeaderboardExtendedResponse struct.
type AppsGetLeaderboardExtendedResponse struct {
	Count int `json:"count"`
	Items []struct {
		Score  int `json:"score"`
		UserID int `json:"user_id"`
	} `json:"items"`
	Profiles []object.UsersUser `json:"profiles"`
}

// AppsGetLeaderboardExtended returns players rating in the game.
//
//	extended=1
//
// https://vk.com/dev/apps.getLeaderboard
func (vk *VK) AppsGetLeaderboardExtended(params Params) (response AppsGetLeaderboardExtendedResponse, err error) {
	err = vk.RequestUnmarshal("apps.getLeaderboard", &response, params, Params{"extended": true})

	return
}

// AppsGetScopesResponse struct.
type AppsGetScopesResponse struct {
	Count int                `json:"count"`
	Items []object.AppsScope `json:"items"`
}

// AppsGetScopes ...
//
// TODO: write docs.
//
// https://vk.com/dev/apps.getScopes
func (vk *VK) AppsGetScopes(params Params) (response AppsGetScopesResponse, err error) {
	err = vk.RequestUnmarshal("apps.getScopes", &response, params)
	return
}

// AppsGetScore returns user score in app.
//
// NOTE: vk wtf!?
//
// https://vk.com/dev/apps.getScore
func (vk *VK) AppsGetScore(params Params) (response string, err error) {
	err = vk.RequestUnmarshal("apps.getScore", &response, params)
	return
}

// AppsGetTestingGroupsResponse struct.
type AppsGetTestingGroupsResponse []object.AppsTestingGroup

// AppsGetTestingGroups method.
//
// https://vk.com/dev/apps.getTestingGroups
func (vk *VK) AppsGetTestingGroups(params Params) (response AppsGetTestingGroupsResponse, err error) {
	err = vk.RequestUnmarshal("apps.getTestingGroups", &response, params)
	return
}

// AppsRemoveTestingGroup method.
//
// https://vk.com/dev/apps.removeTestingGroup
func (vk *VK) AppsRemoveTestingGroup(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("apps.removeTestingGroup", &response, params)
	return
}

// AppsRemoveUsersFromTestingGroups method.
//
// https://vk.com/dev/apps.removeUsersFromTestingGroups
func (vk *VK) AppsRemoveUsersFromTestingGroups(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("apps.removeUsersFromTestingGroups", &response, params)
	return
}

// AppsSendRequest sends a request to another user in an app that uses VK authorization.
//
// https://vk.com/dev/apps.sendRequest
func (vk *VK) AppsSendRequest(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("apps.sendRequest", &response, params)
	return
}

// AppsUpdateMetaForTestingGroupResponse struct.
type AppsUpdateMetaForTestingGroupResponse struct {
	GroupID int `json:"group_id"`
}

// AppsUpdateMetaForTestingGroup method.
//
// https://vk.com/dev/apps.updateMetaForTestingGroup
func (vk *VK) AppsUpdateMetaForTestingGroup(params Params) (response AppsUpdateMetaForTestingGroupResponse, err error) {
	err = vk.RequestUnmarshal("apps.updateMetaForTestingGroup", &response, params)
	return
}
