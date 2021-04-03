package rest

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
)

type ChannelsResponse struct {
	Status
	models.Pagination
	Channels []models.Channel `json:"channels"`
}

type ChannelResponse struct {
	Status
	Channel models.Channel `json:"channel"`
}

type GroupsResponse struct {
	Status
	models.Pagination
	Groups []models.Channel `json:"groups"`
}

type GroupResponse struct {
	Status
	Group models.Channel `json:"group"`
}

// GetPublicChannels returns all channels that can be seen by the logged in user.
//
// https://rocket.chat/docs/developer-guides/rest-api/channels/list
func (c *Client) GetPublicChannels() (*ChannelsResponse, error) {
	response := new(ChannelsResponse)
	if err := c.Get("channels.list", nil, response); err != nil {
		return nil, err
	}

	return response, nil
}

// GetPrivateGroups returns all channels that can be seen by the logged in user.
//
// https://rocket.chat/docs/developer-guides/rest-api/groups/list
func (c *Client) GetPrivateGroups() (*GroupsResponse, error) {
	response := new(GroupsResponse)
	if err := c.Get("groups.list", nil, response); err != nil {
		return nil, err
	}

	return response, nil
}


// GetJoinedChannels returns all channels that the user has joined.
//
// https://rocket.chat/docs/developer-guides/rest-api/channels/list-joined
func (c *Client) GetJoinedChannels(params url.Values) (*ChannelsResponse, error) {
	response := new(ChannelsResponse)
	if err := c.Get("channels.list.joined", params, response); err != nil {
		return nil, err
	}

	return response, nil
}

// LeaveChannel leaves a channel. The id of the channel has to be not nil.
//
// https://rocket.chat/docs/developer-guides/rest-api/channels/leave
func (c *Client) LeaveChannel(channel *models.Channel) error {
	var body = fmt.Sprintf(`{ "roomId": "%s"}`, channel.ID)
	return c.Post("channels.leave", bytes.NewBufferString(body), new(ChannelResponse))
}

// GetChannelInfo get information about a channel. That might be useful to update the usernames.
//
// https://rocket.chat/docs/developer-guides/rest-api/channels/info
func (c *Client) GetChannelInfo(channel *models.Channel) (*models.Channel, error) {
	response := new(ChannelResponse)
	switch {
	case channel.Name != "" && channel.ID == "":
		if err := c.Get("channels.info", url.Values{"roomName": []string{channel.Name}}, response); err != nil {
			return nil, err
		}
	default:
		if err := c.Get("channels.info", url.Values{"roomId": []string{channel.ID}}, response); err != nil {
			return nil, err
		}
	}

	return &response.Channel, nil
}

// GetGroupInfo get information about a group. That might be useful to update the usernames.
//
// https://rocket.chat/docs/developer-guides/rest-api/groups/info
func (c *Client) GetGroupInfo(channel *models.Channel) (*models.Channel, error) {
	response := new(GroupResponse)
	switch {
	case channel.Name != "" && channel.ID == "":
		if err := c.Get("groups.info", url.Values{"roomName": []string{channel.Name}}, response); err != nil {
			return nil, err
		}
	default:
		if err := c.Get("groups.info", url.Values{"roomId": []string{channel.ID}}, response); err != nil {
			return nil, err
		}
	}
	return &response.Group, nil
}
