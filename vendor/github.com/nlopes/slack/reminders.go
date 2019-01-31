package slack

import (
	"context"
	"net/url"
	"time"
)

type Reminder struct {
	ID         string    `json:"id"`
	Creator    string    `json:"creator"`
	User       string    `json:"user"`
	Text       string    `json:"text"`
	Recurring  bool      `json:"recurring"`
	Time       time.Time `json:"time"`
	CompleteTS int       `json:"complete_ts"`
}

type reminderResp struct {
	SlackResponse
	Reminder Reminder `json:"reminder"`
}

func (api *Client) doReminder(ctx context.Context, path string, values url.Values) (*Reminder, error) {
	response := &reminderResp{}
	if err := postSlackMethod(ctx, api.httpclient, path, values, response, api); err != nil {
		return nil, err
	}
	return &response.Reminder, response.Err()
}

// AddChannelReminder adds a reminder for a channel.
//
// See https://api.slack.com/methods/reminders.add (NOTE: the ability to set
// reminders on a channel is currently undocumented but has been tested to
// work)
func (api *Client) AddChannelReminder(channelID, text, time string) (*Reminder, error) {
	values := url.Values{
		"token":   {api.token},
		"text":    {text},
		"time":    {time},
		"channel": {channelID},
	}
	return api.doReminder(context.Background(), "reminders.add", values)
}

// AddUserReminder adds a reminder for a user.
//
// See https://api.slack.com/methods/reminders.add (NOTE: the ability to set
// reminders on a channel is currently undocumented but has been tested to
// work)
func (api *Client) AddUserReminder(userID, text, time string) (*Reminder, error) {
	values := url.Values{
		"token": {api.token},
		"text":  {text},
		"time":  {time},
		"user":  {userID},
	}
	return api.doReminder(context.Background(), "reminders.add", values)
}

// DeleteReminder deletes an existing reminder.
//
// See https://api.slack.com/methods/reminders.delete
func (api *Client) DeleteReminder(id string) error {
	values := url.Values{
		"token":    {api.token},
		"reminder": {id},
	}
	response := &SlackResponse{}
	if err := postSlackMethod(context.Background(), api.httpclient, "reminders.delete", values, response, api); err != nil {
		return err
	}
	return response.Err()
}
