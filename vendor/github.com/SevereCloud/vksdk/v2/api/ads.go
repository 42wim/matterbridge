package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"bytes"
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/vmihailenco/msgpack/v5"
)

// AdsAddOfficeUsersItem struct.
type AdsAddOfficeUsersItem struct {
	OK    object.BaseBoolInt
	Error AdsError
}

// UnmarshalJSON func.
func (r *AdsAddOfficeUsersItem) UnmarshalJSON(data []byte) (err error) {
	if r.OK.UnmarshalJSON(data) != nil {
		return json.Unmarshal(data, &r.Error)
	}

	return
}

// DecodeMsgpack func.
func (r *AdsAddOfficeUsersItem) DecodeMsgpack(dec *msgpack.Decoder) error {
	data, err := dec.DecodeRaw()
	if err != nil {
		return err
	}

	if msgpack.Unmarshal(data, &r.OK) != nil {
		d := msgpack.NewDecoder(bytes.NewReader(data))
		d.SetCustomStructTag("json")

		return d.Decode(&r.Error)
	}

	return nil
}

// AdsAddOfficeUsersResponse struct.
type AdsAddOfficeUsersResponse []AdsAddOfficeUsersItem

// AdsAddOfficeUsers adds managers and/or supervisors to advertising account.
//
// https://vk.com/dev/ads.addOfficeUsers
func (vk *VK) AdsAddOfficeUsers(params Params) (response AdsAddOfficeUsersResponse, err error) {
	err = vk.RequestUnmarshal("ads.addOfficeUsers", &response, params)
	return
}

// AdsCheckLinkResponse struct.
type AdsCheckLinkResponse struct {
	// link status
	Status object.AdsLinkStatus `json:"status"`

	// (if status = disallowed) — description of the reason
	Description string `json:"description,omitempty"`

	// (if the end link differs from original and status = allowed) — end link.
	RedirectURL string `json:"redirect_url,omitempty"`
}

// AdsCheckLink allows to check the ad link.
//
// https://vk.com/dev/ads.checkLink
func (vk *VK) AdsCheckLink(params Params) (response AdsCheckLinkResponse, err error) {
	err = vk.RequestUnmarshal("ads.checkLink", &response, params)
	return
}

// AdsCreateAdsResponse struct.
type AdsCreateAdsResponse []struct {
	ID int `json:"id"`
	AdsError
}

// AdsCreateAds creates ads.
//
// Please note! Maximum allowed number of ads created in one request is 5.
// Minimum size of ad audience is 50 people.
//
// https://vk.com/dev/ads.createAds
func (vk *VK) AdsCreateAds(params Params) (response AdsCreateAdsResponse, err error) {
	err = vk.RequestUnmarshal("ads.createAds", &response, params)
	return
}

// AdsCreateCampaignsResponse struct.
type AdsCreateCampaignsResponse []struct {
	ID int `json:"id"`
	AdsError
}

// AdsCreateCampaigns creates advertising campaigns.
//
// Please note! Allowed number of campaigns created in one request is 50.
//
// https://vk.com/dev/ads.createCampaigns
func (vk *VK) AdsCreateCampaigns(params Params) (response AdsCreateCampaignsResponse, err error) {
	err = vk.RequestUnmarshal("ads.createCampaigns", &response, params)
	return
}

// AdsCreateClientsResponse struct.
type AdsCreateClientsResponse []struct {
	ID int `json:"id"`
	AdsError
}

// AdsCreateClients creates clients of an advertising agency.
//
// Available only for advertising agencies.
//
// Please note! Allowed number of clients created in one request is 50.
//
// https://vk.com/dev/ads.createClients
func (vk *VK) AdsCreateClients(params Params) (response AdsCreateClientsResponse, err error) {
	err = vk.RequestUnmarshal("ads.createClients", &response, params)
	return
}

// AdsCreateLookalikeRequestResponse struct.
type AdsCreateLookalikeRequestResponse struct {
	RequestID int `json:"request_id"`
}

// AdsCreateLookalikeRequest creates a request to find a similar audience.
//
// https://vk.com/dev/ads.createLookalikeRequest
func (vk *VK) AdsCreateLookalikeRequest(params Params) (response AdsCreateLookalikeRequestResponse, err error) {
	err = vk.RequestUnmarshal("ads.createLookalikeRequest", &response, params)
	return
}

// AdsCreateTargetGroupResponse struct.
type AdsCreateTargetGroupResponse struct {
	ID int `json:"id"`
}

// AdsCreateTargetGroup Creates a group to re-target ads for users who visited
// advertiser's site (viewed information about the product, registered, etc.).
//
// When executed successfully this method returns user accounting code on
// advertiser's site. You shall add this code to the site page, so users
// registered in VK will be added to the created target group after they visit
// this page.
//
// Use ads.importTargetContacts method to import existing user contacts to
// the group.
//
// Please note! Maximum allowed number of groups for one advertising
// account is 100.
//
// https://vk.com/dev/ads.createTargetGroup
func (vk *VK) AdsCreateTargetGroup(params Params) (response AdsCreateTargetGroupResponse, err error) {
	err = vk.RequestUnmarshal("ads.createTargetGroup", &response, params)
	return
}

// AdsCreateTargetPixelResponse struct.
type AdsCreateTargetPixelResponse struct {
	ID    int    `json:"id"`
	Pixel string `json:"pixel"`
}

// AdsCreateTargetPixel Creates retargeting pixel.
//
// Method returns pixel code for users accounting on the advertiser site.
// Authorized VK users who visited the page with pixel code on it will be
// added to retargeting audience with corresponding rules. You can also use
// Open API, ads.importTargetContacts method and loading from file.
//
// Maximum pixels number per advertising account is 25.
//
// https://vk.com/dev/ads.createTargetPixel
func (vk *VK) AdsCreateTargetPixel(params Params) (response AdsCreateTargetPixelResponse, err error) {
	err = vk.RequestUnmarshal("ads.createTargetPixel", &response, params)
	return
}

// AdsDeleteAdsResponse struct.
//
// Each response is 0 — deleted successfully, or an error code.
type AdsDeleteAdsResponse []ErrorType

// AdsDeleteAds archives ads.
//
// Warning! Maximum allowed number of ads archived in one request is 100.
//
// https://vk.com/dev/ads.deleteAds
func (vk *VK) AdsDeleteAds(params Params) (response AdsDeleteAdsResponse, err error) {
	err = vk.RequestUnmarshal("ads.deleteAds", &response, params)
	return
}

// AdsDeleteCampaignsResponse struct.
//
// Each response is 0 — deleted successfully, or an error code.
type AdsDeleteCampaignsResponse []ErrorType

// AdsDeleteCampaigns archives advertising campaigns.
//
// Warning! Maximum allowed number of campaigns archived in one request is 100.
//
// https://vk.com/dev/ads.deleteCampaigns
func (vk *VK) AdsDeleteCampaigns(params Params) (response AdsDeleteCampaignsResponse, err error) {
	err = vk.RequestUnmarshal("ads.deleteCampaigns", &response, params)
	return
}

// AdsDeleteClientsResponse struct.
//
// Each response is 0 — deleted successfully, or an error code.
type AdsDeleteClientsResponse []ErrorType

// AdsDeleteClients archives clients of an advertising agency.
//
// Available only for advertising agencies.
//
// Please note! Maximum allowed number of clients edited in one request is 10.
//
// https://vk.com/dev/ads.deleteClients
func (vk *VK) AdsDeleteClients(params Params) (response AdsDeleteClientsResponse, err error) {
	err = vk.RequestUnmarshal("ads.deleteClients", &response, params)
	return
}

// AdsDeleteTargetGroup deletes target group.
//
// https://vk.com/dev/ads.deleteTargetGroup
func (vk *VK) AdsDeleteTargetGroup(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("ads.deleteTargetGroup", &response, params)
	return
}

// AdsDeleteTargetPixel deletes target pixel.
//
// https://vk.com/dev/ads.deleteTargetPixel
func (vk *VK) AdsDeleteTargetPixel(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("ads.deleteTargetPixel", &response, params)
	return
}

// AdsGetAccountsResponse struct.
type AdsGetAccountsResponse []object.AdsAccount

// AdsGetAccounts returns a list of advertising accounts.
//
// https://vk.com/dev/ads.getAccounts
func (vk *VK) AdsGetAccounts(params Params) (response AdsGetAccountsResponse, err error) {
	err = vk.RequestUnmarshal("ads.getAccounts", &response, params)
	return
}

// AdsGetAdsResponse struct.
type AdsGetAdsResponse []object.AdsAd

// AdsGetAds returns a list of ads.
//
// https://vk.com/dev/ads.getAds
func (vk *VK) AdsGetAds(params Params) (response AdsGetAdsResponse, err error) {
	err = vk.RequestUnmarshal("ads.getAds", &response, params)
	return
}

// AdsGetAdsLayoutResponse struct.
type AdsGetAdsLayoutResponse []object.AdsAdLayout

// AdsGetAdsLayout returns descriptions of ad layouts.
//
// https://vk.com/dev/ads.getAdsLayout
func (vk *VK) AdsGetAdsLayout(params Params) (response AdsGetAdsLayoutResponse, err error) {
	err = vk.RequestUnmarshal("ads.getAdsLayout", &response, params)
	return
}

// TODO: AdsGetAdsTargetingResponse struct.
// type AdsGetAdsTargetingResponse struct{}

// TODO: AdsGetAdsTargeting ...
//
// https://vk.com/dev/ads.getAdsTargeting
// func (vk *VK) AdsGetAdsTargeting(params Params) (response AdsGetAdsTargetingResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getAdsTargeting", &response, params)
// 	return
// }

// TODO: AdsGetBudgetResponse struct.
// type AdsGetBudgetResponse struct{}

// TODO: AdsGetBudget ...
//
// https://vk.com/dev/ads.getBudget
// func (vk *VK) AdsGetBudget(params Params) (response AdsGetBudgetResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getBudget", &response, params)
// 	return
// }

// TODO: AdsGetCampaignsResponse struct.
// type AdsGetCampaignsResponse struct{}

// TODO: AdsGetCampaigns ...
//
// https://vk.com/dev/ads.getCampaigns
// func (vk *VK) AdsGetCampaigns(params Params) (response AdsGetCampaignsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getCampaigns", &response, params)
// 	return
// }

// TODO: AdsGetCategoriesResponse struct.
// type AdsGetCategoriesResponse struct{}

// TODO: AdsGetCategories ...
//
// https://vk.com/dev/ads.getCategories
// func (vk *VK) AdsGetCategories(params Params) (response AdsGetCategoriesResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getCategories", &response, params)
// 	return
// }

// TODO: AdsGetClientsResponse struct.
// type AdsGetClientsResponse struct{}

// TODO: AdsGetClients ...
//
// https://vk.com/dev/ads.getClients
// func (vk *VK) AdsGetClients(params Params) (response AdsGetClientsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getClients", &response, params)
// 	return
// }

// TODO: AdsGetDemographicsResponse struct.
// type AdsGetDemographicsResponse struct{}

// TODO: AdsGetDemographics ...
//
// https://vk.com/dev/ads.getDemographics
// func (vk *VK) AdsGetDemographics(params Params) (response AdsGetDemographicsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getDemographics", &response, params)
// 	return
// }

// TODO: AdsGetFloodStatsResponse struct.
// type AdsGetFloodStatsResponse struct{}

// TODO: AdsGetFloodStats ...
//
// https://vk.com/dev/ads.getFloodStats
// func (vk *VK) AdsGetFloodStats(params Params) (response AdsGetFloodStatsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getFloodStats", &response, params)
// 	return
// }

// TODO: AdsGetLookalikeRequestsResponse struct.
// type AdsGetLookalikeRequestsResponse struct{}

// TODO: AdsGetLookalikeRequests ...
//
// https://vk.com/dev/ads.getLookalikeRequests
// func (vk *VK) AdsGetLookalikeRequests(params Params) (response AdsGetLookalikeRequestsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getLookalikeRequests", &response, params)
// 	return
// }

// AdsGetMusiciansResponse struct.
type AdsGetMusiciansResponse struct {
	Items []object.AdsMusician
}

// AdsGetMusicians returns a list of musicians.
//
// https://vk.com/dev/ads.getMusicians
func (vk *VK) AdsGetMusicians(params Params) (response AdsGetMusiciansResponse, err error) {
	err = vk.RequestUnmarshal("ads.getMusicians", &response, params)
	return
}

// TODO: AdsGetOfficeUsersResponse struct.
// type AdsGetOfficeUsersResponse struct{}

// TODO: AdsGetOfficeUsers ...
//
// https://vk.com/dev/ads.getOfficeUsers
// func (vk *VK) AdsGetOfficeUsers(params Params) (response AdsGetOfficeUsersResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getOfficeUsers", &response, params)
// 	return
// }

// TODO: AdsGetPostsReachResponse struct.
// type AdsGetPostsReachResponse struct{}

// TODO: AdsGetPostsReach ...
//
// https://vk.com/dev/ads.getPostsReach
// func (vk *VK) AdsGetPostsReach(params Params) (response AdsGetPostsReachResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getPostsReach", &response, params)
// 	return
// }

// TODO: AdsGetRejectionReasonResponse struct.
// type AdsGetRejectionReasonResponse struct{}

// TODO: AdsGetRejectionReason ...
//
// https://vk.com/dev/ads.getRejectionReason
// func (vk *VK) AdsGetRejectionReason(params Params) (response AdsGetRejectionReasonResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getRejectionReason", &response, params)
// 	return
// }

// TODO: AdsGetStatisticsResponse struct.
// type AdsGetStatisticsResponse struct{}

// TODO: AdsGetStatistics ...
//
// https://vk.com/dev/ads.getStatistics
// func (vk *VK) AdsGetStatistics(params Params) (response AdsGetStatisticsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getStatistics", &response, params)
// 	return
// }

// TODO: AdsGetSuggestionsResponse struct.
// type AdsGetSuggestionsResponse struct{}

// TODO: AdsGetSuggestions ...
//
// https://vk.com/dev/ads.getSuggestions
// func (vk *VK) AdsGetSuggestions(params Params) (response AdsGetSuggestionsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getSuggestions", &response, params)
// 	return
// }

// AdsGetTargetGroupsResponse struct.
type AdsGetTargetGroupsResponse []object.AdsTargetGroup

// AdsGetTargetGroups returns a list of target groups.
//
// https://vk.com/dev/ads.getTargetGroups
func (vk *VK) AdsGetTargetGroups(params Params) (response AdsGetTargetGroupsResponse, err error) {
	err = vk.RequestUnmarshal("ads.getTargetGroups", &response, params)
	return
}

// TODO: AdsGetTargetPixelsResponse struct.
// type AdsGetTargetPixelsResponse struct{}

// TODO: AdsGetTargetPixels ...
//
// https://vk.com/dev/ads.getTargetPixels
// func (vk *VK) AdsGetTargetPixels(params Params) (response AdsGetTargetPixelsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getTargetPixels", &response, params)
// 	return
// }

// TODO: AdsGetTargetingStatsResponse struct.
// type AdsGetTargetingStatsResponse struct{}

// TODO: AdsGetTargetingStats ...
//
// https://vk.com/dev/ads.getTargetingStats
// func (vk *VK) AdsGetTargetingStats(params Params) (response AdsGetTargetingStatsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getTargetingStats", &response, params)
// 	return
// }

// TODO: AdsGetUploadURLResponse struct.
// type AdsGetUploadURLResponse struct{}

// TODO: AdsGetUploadURL ...
//
// https://vk.com/dev/ads.getUploadURL
// func (vk *VK) AdsGetUploadURL(params Params) (response AdsGetUploadURLResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getUploadURL", &response, params)
// 	return
// }

// TODO: AdsGetVideoUploadURLResponse struct.
// type AdsGetVideoUploadURLResponse struct{}

// TODO: AdsGetVideoUploadURL ...
//
// https://vk.com/dev/ads.getVideoUploadURL
// func (vk *VK) AdsGetVideoUploadURL(params Params) (response AdsGetVideoUploadURLResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.getVideoUploadURL", &response, params)
// 	return
// }

// TODO: AdsImportTargetContactsResponse struct.
// type AdsImportTargetContactsResponse struct{}

// TODO: AdsImportTargetContacts ...
//
// https://vk.com/dev/ads.importTargetContacts
// func (vk *VK) AdsImportTargetContacts(params Params) (response AdsImportTargetContactsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.importTargetContacts", &response, params)
// 	return
// }

// TODO: AdsRemoveOfficeUsersResponse struct.
// type AdsRemoveOfficeUsersResponse struct{}

// TODO: AdsRemoveOfficeUsers ...
//
// https://vk.com/dev/ads.removeOfficeUsers
// func (vk *VK) AdsRemoveOfficeUsers(params Params) (response AdsRemoveOfficeUsersResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.removeOfficeUsers", &response, params)
// 	return
// }

// AdsRemoveTargetContacts accepts the request to exclude the advertiser's
// contacts from the retargeting audience.
//
// The maximum allowed number of contacts to be excluded by a single
// request is 1000.
//
// Contacts are excluded within a few hours of the request.
//
// https://vk.com/dev/ads.removeTargetContacts
func (vk *VK) AdsRemoveTargetContacts(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("ads.removeTargetContacts", &response, params)
	return
}

// TODO: AdsSaveLookalikeRequestResultResponse struct.
// type AdsSaveLookalikeRequestResultResponse struct{}

// TODO: AdsSaveLookalikeRequestResult ...
//
// https://vk.com/dev/ads.saveLookalikeRequestResult
// func (vk *VK) AdsSaveLookalikeRequestResult(params Params) (
// 		response AdsSaveLookalikeRequestResultResponse,
// 		err error,
// 	) {
// 	err = vk.RequestUnmarshal("ads.saveLookalikeRequestResult", &response, params)
// 	return
// }

// TODO: AdsShareTargetGroupResponse struct.
// type AdsShareTargetGroupResponse struct{}

// TODO: AdsShareTargetGroup ...
//
// https://vk.com/dev/ads.shareTargetGroup
// func (vk *VK) AdsShareTargetGroup(params Params) (response AdsShareTargetGroupResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.shareTargetGroup", &response, params)
// 	return
// }

// TODO: AdsUpdateAdsResponse struct.
// type AdsUpdateAdsResponse struct{}

// TODO: AdsUpdateAds ...
//
// https://vk.com/dev/ads.updateAds
// func (vk *VK) AdsUpdateAds(params Params) (response AdsUpdateAdsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.updateAds", &response, params)
// 	return
// }

// TODO: AdsUpdateCampaignsResponse struct.
// type AdsUpdateCampaignsResponse struct{}

// TODO: AdsUpdateCampaigns ...
//
// https://vk.com/dev/ads.updateCampaigns
// func (vk *VK) AdsUpdateCampaigns(params Params) (response AdsUpdateCampaignsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.updateCampaigns", &response, params)
// 	return
// }

// TODO: AdsUpdateClientsResponse struct.
// type AdsUpdateClientsResponse struct{}

// TODO: AdsUpdateClients ...
//
// https://vk.com/dev/ads.updateClients
// func (vk *VK) AdsUpdateClients(params Params) (response AdsUpdateClientsResponse, err error) {
// 	err = vk.RequestUnmarshal("ads.updateClients", &response, params)
// 	return
// }

// AdsUpdateTargetGroup edits target group.
//
// https://vk.com/dev/ads.updateTargetGroup
func (vk *VK) AdsUpdateTargetGroup(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("ads.updateTargetGroup", &response, params)
	return
}

// AdsUpdateTargetPixel edits target pixel.
//
// https://vk.com/dev/ads.updateTargetPixel
func (vk *VK) AdsUpdateTargetPixel(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("ads. updateTargetPixel", &response, params)
	return
}
