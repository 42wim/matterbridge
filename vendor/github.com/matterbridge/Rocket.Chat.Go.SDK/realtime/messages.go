package realtime

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/gopackage/ddp"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
)

const (
	// RocketChat doesn't send the `added` event for new messages by default, only `changed`.
	send_added_event    = true
	default_buffer_size = 100
)

// LoadHistory loads history
// Takes roomId
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/load-history
func (c *Client) LoadHistory(roomId string) error {
	_, err := c.ddp.Call("loadHistory", roomId)
	if err != nil {
		return err
	}

	return nil
}

// SendMessage sends message to channel
// takes channel and message
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/send-message
func (c *Client) SendMessage(m *models.Message) (*models.Message, error) {
	m.ID = c.newRandomId()

	rawResponse, err := c.ddp.Call("sendMessage", m)
	if err != nil {
		return nil, err
	}

	return getMessageFromData(rawResponse.(map[string]interface{})), nil
}

// EditMessage edits a message
// takes message object
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/update-message
func (c *Client) EditMessage(message *models.Message) error {
	_, err := c.ddp.Call("updateMessage", message)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMessage deletes a message
// takes a message object
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/delete-message
func (c *Client) DeleteMessage(message *models.Message) error {
	_, err := c.ddp.Call("deleteMessage", map[string]string{
		"_id": message.ID,
	})
	if err != nil {
		return err
	}

	return nil
}

// ReactToMessage adds a reaction to a message
// takes a message and emoji
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/set-reaction
func (c *Client) ReactToMessage(message *models.Message, reaction string) error {
	_, err := c.ddp.Call("setReaction", reaction, message.ID)
	if err != nil {
		return err
	}

	return nil
}

// StarMessage stars message
// takes a message object
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/star-message
func (c *Client) StarMessage(message *models.Message) error {
	_, err := c.ddp.Call("starMessage", map[string]interface{}{
		"_id":     message.ID,
		"rid":     message.RoomID,
		"starred": true,
	})

	if err != nil {
		return err
	}

	return nil
}

// UnStarMessage unstars message
// takes message object
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/star-message
func (c *Client) UnStarMessage(message *models.Message) error {
	_, err := c.ddp.Call("starMessage", map[string]interface{}{
		"_id":     message.ID,
		"rid":     message.RoomID,
		"starred": false,
	})

	if err != nil {
		return err
	}

	return nil
}

// PinMessage pins a message
// takes a message object
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/pin-message
func (c *Client) PinMessage(message *models.Message) error {
	_, err := c.ddp.Call("pinMessage", message)

	if err != nil {
		return err
	}

	return nil
}

// UnPinMessage unpins message
// takes a message object
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/unpin-messages
func (c *Client) UnPinMessage(message *models.Message) error {
	_, err := c.ddp.Call("unpinMessage", message)

	if err != nil {
		return err
	}

	return nil
}

// SubscribeToMessageStream Subscribes to the message updates of a channel
// Returns a buffered channel
//
// https://rocket.chat/docs/developer-guides/realtime-api/subscriptions/stream-room-messages/
func (c *Client) SubscribeToMessageStream(channel *models.Channel, msgChannel chan models.Message) error {

	if err := c.ddp.Sub("stream-room-messages", channel.ID, send_added_event); err != nil {
		return err
	}

	//msgChannel := make(chan models.Message, default_buffer_size)
	c.ddp.CollectionByName("stream-room-messages").AddUpdateListener(messageExtractor{msgChannel, "update"})

	return nil
}

func getMessagesFromUpdateEvent(update ddp.Update) []models.Message {
	document, _ := gabs.Consume(update["args"])
	args, err := document.Children()

	if err != nil {
		//	log.Printf("Event arguments are in an unexpected format: %v", err)
		return make([]models.Message, 0)
	}

	messages := make([]models.Message, len(args))

	for i, arg := range args {
		messages[i] = *getMessageFromDocument(arg)
	}

	return messages
}

func getMessageFromData(data interface{}) *models.Message {
	// TODO: We should know what this will look like, we shouldn't need to use gabs
	document, _ := gabs.Consume(data)
	return getMessageFromDocument(document)
}

func getMessageFromDocument(arg *gabs.Container) *models.Message {
	var ts *time.Time
	date := stringOrZero(arg.Path("ts.$date").Data())
	if len(date) > 0 {
		if ti, err := strconv.ParseFloat(date, 64); err == nil {
			t := time.Unix(int64(ti)/1e3, int64(ti)%1e3)
			ts = &t
		}
	}
	return &models.Message{
		ID:        stringOrZero(arg.Path("_id").Data()),
		RoomID:    stringOrZero(arg.Path("rid").Data()),
		Msg:       stringOrZero(arg.Path("msg").Data()),
		Timestamp: ts,
		User: &models.User{
			ID:       stringOrZero(arg.Path("u._id").Data()),
			UserName: stringOrZero(arg.Path("u.username").Data()),
		},
	}
}

func stringOrZero(i interface{}) string {
	if i == nil {
		return ""
	}

	switch i.(type) {
	case string:
		return i.(string)
	case float64:
		return fmt.Sprintf("%f", i.(float64))
	default:
		return ""
	}
}

type messageExtractor struct {
	messageChannel chan models.Message
	operation      string
}

func (u messageExtractor) CollectionUpdate(collection, operation, id string, doc ddp.Update) {
	if operation == u.operation {
		msgs := getMessagesFromUpdateEvent(doc)
		for _, m := range msgs {
			u.messageChannel <- m
		}
	}
}
