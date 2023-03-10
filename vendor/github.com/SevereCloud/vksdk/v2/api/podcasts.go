package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// PodcastsGetCatalogResponse struct.
type PodcastsGetCatalogResponse struct {
	Items []object.PodcastsItem `json:"items"`
}

// PodcastsGetCatalog method.
//
//	extended=0
//
// https://vk.com/dev/podcasts.getCatalog
func (vk *VK) PodcastsGetCatalog(params Params) (response PodcastsGetCatalogResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getCatalog", &response, params, Params{"extended": false})

	return
}

// PodcastsGetCatalogExtendedResponse struct.
type PodcastsGetCatalogExtendedResponse struct {
	Items []object.PodcastsItem `json:"items"`
	object.ExtendedResponse
}

// PodcastsGetCatalogExtended method.
//
//	extended=1
//
// https://vk.com/dev/podcasts.getCatalog
func (vk *VK) PodcastsGetCatalogExtended(params Params) (response PodcastsGetCatalogExtendedResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getCatalog", &response, params, Params{"extended": true})

	return
}

// PodcastsGetCategoriesResponse struct.
type PodcastsGetCategoriesResponse []object.PodcastsCategory

// PodcastsGetCategories method.
//
// https://vk.com/dev/podcasts.getCategories
func (vk *VK) PodcastsGetCategories(params Params) (response PodcastsGetCategoriesResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getCategories", &response, params)
	return
}

// PodcastsGetEpisodesResponse struct.
type PodcastsGetEpisodesResponse struct {
	Count int                      `json:"count"`
	Items []object.PodcastsEpisode `json:"items"`
}

// PodcastsGetEpisodes method.
//
// https://vk.com/dev/podcasts.getEpisodes
func (vk *VK) PodcastsGetEpisodes(params Params) (response PodcastsGetEpisodesResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getEpisodes", &response, params)
	return
}

// PodcastsGetFeedResponse struct.
type PodcastsGetFeedResponse struct {
	Items    []object.PodcastsEpisode `json:"items"`
	NextFrom string                   `json:"next_from"`
}

// PodcastsGetFeed method.
//
//	extended=0
//
// https://vk.com/dev/podcasts.getFeed
func (vk *VK) PodcastsGetFeed(params Params) (response PodcastsGetFeedResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getFeed", &response, params, Params{"extended": false})

	return
}

// PodcastsGetFeedExtendedResponse struct.
type PodcastsGetFeedExtendedResponse struct {
	Items    []object.PodcastsEpisode `json:"items"`
	NextFrom string                   `json:"next_from"`
	object.ExtendedResponse
}

// PodcastsGetFeedExtended method.
//
//	extended=1
//
// https://vk.com/dev/podcasts.getFeed
func (vk *VK) PodcastsGetFeedExtended(params Params) (response PodcastsGetFeedExtendedResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getFeed", &response, params, Params{"extended": true})

	return
}

// PodcastsGetStartPageResponse struct.
type PodcastsGetStartPageResponse struct {
	Order               []string                  `json:"order"`
	InProgress          []object.PodcastsEpisode  `json:"in_progress"`
	Bookmarks           []object.PodcastsEpisode  `json:"bookmarks"`
	Articles            []object.Article          `json:"articles"`
	StaticHowTo         []bool                    `json:"static_how_to"`
	FriendsLiked        []object.PodcastsEpisode  `json:"friends_liked"`
	Subscriptions       []object.PodcastsEpisode  `json:"subscriptions"`
	CategoriesList      []object.PodcastsCategory `json:"categories_list"`
	RecommendedEpisodes []object.PodcastsEpisode  `json:"recommended_episodes"`
	Catalog             []struct {
		Category object.PodcastsCategory `json:"category"`
		Items    []object.PodcastsItem   `json:"items"`
	} `json:"catalog"`
}

// PodcastsGetStartPage method.
//
//	extended=0
//
// https://vk.com/dev/podcasts.getStartPage
func (vk *VK) PodcastsGetStartPage(params Params) (response PodcastsGetStartPageResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getStartPage", &response, params, Params{"extended": false})

	return
}

// PodcastsGetStartPageExtendedResponse struct.
type PodcastsGetStartPageExtendedResponse struct {
	Order               []string                  `json:"order"`
	InProgress          []object.PodcastsEpisode  `json:"in_progress"`
	Bookmarks           []object.PodcastsEpisode  `json:"bookmarks"`
	Articles            []object.Article          `json:"articles"`
	StaticHowTo         []bool                    `json:"static_how_to"`
	FriendsLiked        []object.PodcastsEpisode  `json:"friends_liked"`
	Subscriptions       []object.PodcastsEpisode  `json:"subscriptions"`
	CategoriesList      []object.PodcastsCategory `json:"categories_list"`
	RecommendedEpisodes []object.PodcastsEpisode  `json:"recommended_episodes"`
	Catalog             []struct {
		Category object.PodcastsCategory `json:"category"`
		Items    []object.PodcastsItem   `json:"items"`
	} `json:"catalog"`
	object.ExtendedResponse
}

// PodcastsGetStartPageExtended method.
//
//	extended=1
//
// https://vk.com/dev/podcasts.getStartPage
func (vk *VK) PodcastsGetStartPageExtended(params Params) (response PodcastsGetStartPageExtendedResponse, err error) {
	err = vk.RequestUnmarshal("podcasts.getStartPage", &response, params, Params{"extended": true})

	return
}

// PodcastsMarkAsListened method.
//
// https://vk.com/dev/podcasts.markAsListened
func (vk *VK) PodcastsMarkAsListened(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("podcasts.markAsListened", &response, params)
	return
}

// PodcastsSubscribe method.
//
// https://vk.com/dev/podcasts.subscribe
func (vk *VK) PodcastsSubscribe(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("podcasts.subscribe", &response, params)
	return
}

// PodcastsUnsubscribe method.
//
// https://vk.com/dev/podcasts.unsubscribe
func (vk *VK) PodcastsUnsubscribe(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("podcasts.unsubscribe", &response, params)
	return
}
