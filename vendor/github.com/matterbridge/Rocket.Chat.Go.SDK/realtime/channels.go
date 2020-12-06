package realtime

import (
	"github.com/Jeffail/gabs"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
)

func (c *Client) GetChannelId(name string) (string, error) {
	rawResponse, err := c.ddp.Call("getRoomIdByNameOrId", name)
	if err != nil {
		return "", err
	}

	//log.Println(rawResponse)

	return rawResponse.(string), nil
}

// GetChannelsIn returns list of channels
// Optionally includes date to get all since last check or 0 to get all
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/get-rooms/
func (c *Client) GetChannelsIn() ([]models.Channel, error) {
	rawResponse, err := c.ddp.Call("rooms/get", map[string]int{
		"$date": 0,
	})
	if err != nil {
		return nil, err
	}

	document, _ := gabs.Consume(rawResponse.(map[string]interface{})["update"])

	chans, err := document.Children()

	var channels []models.Channel

	for _, i := range chans {
		channels = append(channels, models.Channel{
			ID: stringOrZero(i.Path("_id").Data()),
			//Default: stringOrZero(i.Path("default").Data()),
			Name: stringOrZero(i.Path("name").Data()),
			Type: stringOrZero(i.Path("t").Data()),
		})
	}

	return channels, nil
}

// GetChannelSubscriptions gets users channel subscriptions
// Optionally includes date to get all since last check or 0 to get all
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/get-subscriptions
func (c *Client) GetChannelSubscriptions() ([]models.ChannelSubscription, error) {
	rawResponse, err := c.ddp.Call("subscriptions/get", map[string]int{
		"$date": 0,
	})
	if err != nil {
		return nil, err
	}

	document, _ := gabs.Consume(rawResponse.(map[string]interface{})["update"])

	channelSubs, err := document.Children()

	var channelSubscriptions []models.ChannelSubscription

	for _, sub := range channelSubs {
		channelSubscription := models.ChannelSubscription{
			ID:          stringOrZero(sub.Path("_id").Data()),
			Alert:       sub.Path("alert").Data().(bool),
			Name:        stringOrZero(sub.Path("name").Data()),
			DisplayName: stringOrZero(sub.Path("fname").Data()),
			Open:        sub.Path("open").Data().(bool),
			Type:        stringOrZero(sub.Path("t").Data()),
			RoomId:      stringOrZero(sub.Path("rid").Data()),
			User: models.User{
				ID:       stringOrZero(sub.Path("u._id").Data()),
				UserName: stringOrZero(sub.Path("u.username").Data()),
			},
			Unread: sub.Path("unread").Data().(float64),
		}

		if sub.Path("roles").Data() != nil {
			var roles []string
			for _, role := range sub.Path("roles").Data().([]interface{}) {
				roles = append(roles, role.(string))
			}

			channelSubscription.Roles = roles
		}

		channelSubscriptions = append(channelSubscriptions, channelSubscription)
	}

	return channelSubscriptions, nil
}

// GetChannelRoles returns room roles
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/get-room-roles
func (c *Client) GetChannelRoles(roomId string) error {
	_, err := c.ddp.Call("getRoomRoles", roomId)
	if err != nil {
		return err
	}

	return nil
}

// CreateChannel creates a channel
// Takes name and users array
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/create-channels
func (c *Client) CreateChannel(name string, users []string) error {
	_, err := c.ddp.Call("createChannel", name, users)
	if err != nil {
		return err
	}

	return nil
}

// CreateGroup creates a private group
// Takes group name and array of users
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/create-private-groups
func (c *Client) CreateGroup(name string, users []string) error {
	_, err := c.ddp.Call("createPrivateGroup", name, users)
	if err != nil {
		return err
	}

	return nil
}

// JoinChannel joins a channel
// Takes roomId
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/joining-channels
func (c *Client) JoinChannel(roomId string) error {
	_, err := c.ddp.Call("joinRoom", roomId)
	if err != nil {
		return err
	}

	return nil
}

// LeaveChannel leaves a channel
// Takes roomId
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/leaving-rooms
func (c *Client) LeaveChannel(roomId string) error {
	_, err := c.ddp.Call("leaveRoom", roomId)
	if err != nil {
		return err
	}

	return nil
}

// ArchiveChannel archives the channel
// Takes roomId
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/archive-rooms
func (c *Client) ArchiveChannel(roomId string) error {
	_, err := c.ddp.Call("archiveRoom", roomId)
	if err != nil {
		return err
	}

	return nil
}

// UnArchiveChannel unarchives the channel
// Takes roomId
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/unarchive-rooms
func (c *Client) UnArchiveChannel(roomId string) error {
	_, err := c.ddp.Call("unarchiveRoom", roomId)
	if err != nil {
		return err
	}

	return nil
}

// DeleteChannel deletes the channel
// Takes roomId
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/delete-rooms
func (c *Client) DeleteChannel(roomId string) error {
	_, err := c.ddp.Call("eraseRoom", roomId)
	if err != nil {
		return err
	}

	return nil
}

// SetChannelTopic sets channel topic
// takes roomId and topic
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/save-room-settings
func (c *Client) SetChannelTopic(roomId string, topic string) error {
	_, err := c.ddp.Call("saveRoomSettings", roomId, "roomTopic", topic)
	if err != nil {
		return err
	}

	return nil
}

// SetChannelType sets the channel type
// takes roomId and roomType
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/save-room-settings
func (c *Client) SetChannelType(roomId string, roomType string) error {
	_, err := c.ddp.Call("saveRoomSettings", roomId, "roomType", roomType)
	if err != nil {
		return err
	}

	return nil
}

// SetChannelJoinCode sets channel join code
// takes roomId and joinCode
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/save-room-settings
func (c *Client) SetChannelJoinCode(roomId string, joinCode string) error {
	_, err := c.ddp.Call("saveRoomSettings", roomId, "joinCode", joinCode)
	if err != nil {
		return err
	}

	return nil
}

// SetChannelReadOnly sets channel as read only
// takes roomId and boolean
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/save-room-settings
func (c *Client) SetChannelReadOnly(roomId string, readOnly bool) error {
	_, err := c.ddp.Call("saveRoomSettings", roomId, "readOnly", readOnly)
	if err != nil {
		return err
	}

	return nil
}

// SetChannelDescription sets channels description
// takes roomId and description
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/save-room-settings
func (c *Client) SetChannelDescription(roomId string, description string) error {
	_, err := c.ddp.Call("saveRoomSettings", roomId, "roomDescription", description)
	if err != nil {
		return err
	}

	return nil
}
