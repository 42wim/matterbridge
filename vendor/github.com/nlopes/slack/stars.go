package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

const (
	DEFAULT_STARS_USER  = ""
	DEFAULT_STARS_COUNT = 100
	DEFAULT_STARS_PAGE  = 1
)

type StarsParameters struct {
	User  string
	Count int
	Page  int
}

type StarredItem Item

type listResponseFull struct {
	Items  []Item `json:"items"`
	Paging `json:"paging"`
	SlackResponse
}

// NewStarsParameters initialises StarsParameters with default values
func NewStarsParameters() StarsParameters {
	return StarsParameters{
		User:  DEFAULT_STARS_USER,
		Count: DEFAULT_STARS_COUNT,
		Page:  DEFAULT_STARS_PAGE,
	}
}

// AddStar stars an item in a channel
func (api *Client) AddStar(channel string, item ItemRef) error {
	return api.AddStarContext(context.Background(), channel, item)
}

// AddStarContext stars an item in a channel with a custom context
func (api *Client) AddStarContext(ctx context.Context, channel string, item ItemRef) error {
	values := url.Values{
		"channel": {channel},
		"token":   {api.token},
	}
	if item.Timestamp != "" {
		values.Set("timestamp", item.Timestamp)
	}
	if item.File != "" {
		values.Set("file", item.File)
	}
	if item.Comment != "" {
		values.Set("file_comment", item.Comment)
	}

	response := &SlackResponse{}
	if err := postSlackMethod(ctx, api.httpclient, "stars.add", values, response, api); err != nil {
		return err
	}

	return response.Err()
}

// RemoveStar removes a starred item from a channel
func (api *Client) RemoveStar(channel string, item ItemRef) error {
	return api.RemoveStarContext(context.Background(), channel, item)
}

// RemoveStarContext removes a starred item from a channel with a custom context
func (api *Client) RemoveStarContext(ctx context.Context, channel string, item ItemRef) error {
	values := url.Values{
		"channel": {channel},
		"token":   {api.token},
	}
	if item.Timestamp != "" {
		values.Set("timestamp", item.Timestamp)
	}
	if item.File != "" {
		values.Set("file", item.File)
	}
	if item.Comment != "" {
		values.Set("file_comment", item.Comment)
	}

	response := &SlackResponse{}
	if err := postSlackMethod(ctx, api.httpclient, "stars.remove", values, response, api); err != nil {
		return err
	}

	return response.Err()
}

// ListStars returns information about the stars a user added
func (api *Client) ListStars(params StarsParameters) ([]Item, *Paging, error) {
	return api.ListStarsContext(context.Background(), params)
}

// ListStarsContext returns information about the stars a user added with a custom context
func (api *Client) ListStarsContext(ctx context.Context, params StarsParameters) ([]Item, *Paging, error) {
	values := url.Values{
		"token": {api.token},
	}
	if params.User != DEFAULT_STARS_USER {
		values.Add("user", params.User)
	}
	if params.Count != DEFAULT_STARS_COUNT {
		values.Add("count", strconv.Itoa(params.Count))
	}
	if params.Page != DEFAULT_STARS_PAGE {
		values.Add("page", strconv.Itoa(params.Page))
	}

	response := &listResponseFull{}
	err := postSlackMethod(ctx, api.httpclient, "stars.list", values, response, api)
	if err != nil {
		return nil, nil, err
	}
	if !response.Ok {
		return nil, nil, errors.New(response.Error)
	}
	return response.Items, &response.Paging, nil
}

// GetStarred returns a list of StarredItem items.
//
// The user then has to iterate over them and figure out what they should
// be looking at according to what is in the Type.
//    for _, item := range items {
//        switch c.Type {
//        case "file_comment":
//            log.Println(c.Comment)
//        case "file":
//             ...
//
//    }
// This function still exists to maintain backwards compatibility.
// I exposed it as returning []StarredItem, so it shall stay as StarredItem
func (api *Client) GetStarred(params StarsParameters) ([]StarredItem, *Paging, error) {
	return api.GetStarredContext(context.Background(), params)
}

// GetStarredContext returns a list of StarredItem items with a custom context
//
// For more details see GetStarred
func (api *Client) GetStarredContext(ctx context.Context, params StarsParameters) ([]StarredItem, *Paging, error) {
	items, paging, err := api.ListStarsContext(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	starredItems := make([]StarredItem, len(items))
	for i, item := range items {
		starredItems[i] = StarredItem(item)
	}
	return starredItems, paging, nil
}
