package gozulipbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// A Message is all of the necessary metadata to post on Zulip.
// It can be either a public message, where Topic is set, or a private message,
// where there is at least one element in Emails.
//
// If the length of Emails is not 0, functions will always assume it is a private message.
type Message struct {
	Stream  string
	Topic   string
	Emails  []string
	Content string
}

type EventMessage struct {
	AvatarURL        string           `json:"avatar_url"`
	Client           string           `json:"client"`
	Content          string           `json:"content"`
	ContentType      string           `json:"content_type"`
	DisplayRecipient DisplayRecipient `json:"display_recipient"`
	GravatarHash     string           `json:"gravatar_hash"`
	ID               int              `json:"id"`
	RecipientID      int              `json:"recipient_id"`
	SenderDomain     string           `json:"sender_domain"`
	SenderEmail      string           `json:"sender_email"`
	SenderFullName   string           `json:"sender_full_name"`
	SenderID         int              `json:"sender_id"`
	SenderShortName  string           `json:"sender_short_name"`
	Subject          string           `json:"subject"`
	SubjectLinks     []interface{}    `json:"subject_links"`
	StreamID         int              `json:"stream_id"`
	Timestamp        int              `json:"timestamp"`
	Type             string           `json:"type"`
	Queue            *Queue           `json:"-"`
}

type DisplayRecipient struct {
	Users []User `json:"users,omitempty"`
	Topic string `json:"topic,omitempty"`
}

type User struct {
	Domain        string `json:"domain"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	ID            int    `json:"id"`
	IsMirrorDummy bool   `json:"is_mirror_dummy"`
	ShortName     string `json:"short_name"`
}

func (d *DisplayRecipient) UnmarshalJSON(b []byte) (err error) {
	topic, users := "", make([]User, 1)
	if err = json.Unmarshal(b, &topic); err == nil {
		d.Topic = topic
		return
	}
	if err = json.Unmarshal(b, &users); err == nil {
		d.Users = users
		return
	}
	return
}

// Message posts a message to Zulip. If any emails have been set on the message,
// the message will be re-routed to the PrivateMessage function.
func (b *Bot) Message(m Message) (*http.Response, error) {
	if m.Content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// if any emails are set, this is a private message
	if len(m.Emails) != 0 {
		return b.PrivateMessage(m)
	}

	// otherwise it's a stream message
	if m.Stream == "" {
		return nil, fmt.Errorf("stream cannot be empty")
	}
	if m.Topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}
	req, err := b.constructMessageRequest(m)
	if err != nil {
		return nil, err
	}
	return b.Client.Do(req)
}

// PrivateMessage sends a message to the users in the message email slice.
func (b *Bot) PrivateMessage(m Message) (*http.Response, error) {
	if len(m.Emails) == 0 {
		return nil, fmt.Errorf("there must be at least one recipient")
	}
	req, err := b.constructMessageRequest(m)
	if err != nil {
		return nil, err
	}
	return b.Client.Do(req)
}

// Respond sends a given message as a response to whatever context from which
// an EventMessage was received.
func (b *Bot) Respond(e EventMessage, response string) (*http.Response, error) {
	if response == "" {
		return nil, fmt.Errorf("Message response cannot be blank")
	}
	m := Message{
		Stream:  e.DisplayRecipient.Topic,
		Topic:   e.Subject,
		Content: response,
	}
	if m.Topic != "" {
		return b.Message(m)
	}
	// private message
	if m.Stream == "" {
		emails, err := b.privateResponseList(e)
		if err != nil {
			return nil, err
		}
		m.Emails = emails
		return b.Message(m)
	}
	return nil, fmt.Errorf("EventMessage is not understood: %v\n", e)
}

// privateResponseList gets the list of other users in a private multiple
// message conversation.
func (b *Bot) privateResponseList(e EventMessage) ([]string, error) {
	var out []string
	for _, u := range e.DisplayRecipient.Users {
		if u.Email != b.Email {
			out = append(out, u.Email)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("EventMessage had no Users within the DisplayRecipient")
	}
	return out, nil
}

// constructMessageRequest is a helper for simplifying sending a message.
func (b *Bot) constructMessageRequest(m Message) (*http.Request, error) {
	to := m.Stream
	mtype := "stream"

	le := len(m.Emails)
	if le != 0 {
		mtype = "private"
	}
	if le == 1 {
		to = m.Emails[0]
	}
	if le > 1 {
		to = ""
		for i, e := range m.Emails {
			to += e
			if i != le-1 {
				to += ","
			}
		}
	}

	values := url.Values{}
	values.Set("type", mtype)
	values.Set("to", to)
	values.Set("content", m.Content)
	if mtype == "stream" {
		values.Set("subject", m.Topic)
	}

	return b.constructRequest("POST", "messages", values.Encode())
}

func (b *Bot) UpdateMessage(id string, content string) (*http.Response, error) {
	//mid, _ := strconv.Atoi(id)
	values := url.Values{}
	values.Set("content", content)
	req, err := b.constructRequest("PATCH", "messages/"+id, values.Encode())
	if err != nil {
		return nil, err
	}
	return b.Client.Do(req)
}

// React adds an emoji reaction to an EventMessage.
func (b *Bot) React(e EventMessage, emoji string) (*http.Response, error) {
	requestURL := fmt.Sprintf("messages/%d/reactions", e.ID)
	values := url.Values{}
	values.Set("emoji_name", emoji)
	req, err := b.constructRequest("POST", requestURL, values.Encode())
	if err != nil {
		return nil, err
	}
	return b.Client.Do(req)
}

// Unreact removes an emoji reaction from an EventMessage.
func (b *Bot) Unreact(e EventMessage, emoji string) (*http.Response, error) {
	requestURL := fmt.Sprintf("messages/%d/reactions", e.ID)
	values := url.Values{}
	values.Set("emoji_name", emoji)
	req, err := b.constructRequest("DELETE", requestURL, values.Encode())
	if err != nil {
		return nil, err
	}
	return b.Client.Do(req)
}

type Emoji struct {
	Author     string `json:"author"`
	DisplayURL string `json:"display_url"`
	SourceURL  string `json:"source_url"`
}

type EmojiResponse struct {
	Emoji  map[string]*Emoji `json:"emoji"`
	Msg    string            `json:"msg"`
	Result string            `json:"result"`
}

// RealmEmoji gets the custom emoji information for the Zulip instance.
func (b *Bot) RealmEmoji() (map[string]*Emoji, error) {
	req, err := b.constructRequest("GET", "realm/emoji", "")
	if err != nil {
		return nil, err
	}
	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var emjResp EmojiResponse
	err = json.Unmarshal(body, &emjResp)
	if err != nil {
		return nil, err
	}
	return emjResp.Emoji, nil
}

// RealmEmojiSet makes a set of the names of the custom emoji in the Zulip instance.
func (b *Bot) RealmEmojiSet() (map[string]struct{}, error) {
	emj, err := b.RealmEmoji()
	if err != nil {
		return nil, nil
	}
	out := map[string]struct{}{}
	for k, _ := range emj {
		out[k] = struct{}{}
	}
	return out, nil
}
