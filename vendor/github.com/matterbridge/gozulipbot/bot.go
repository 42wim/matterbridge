package gozulipbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Bot struct {
	APIKey    string
	APIURL    string
	Email     string
	Queues    []*Queue
	Streams   []string
	Client    Doer
	Backoff   time.Duration
	Retries   int64
	UserAgent string
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Init adds an http client to an existing bot struct.
func (b *Bot) Init() *Bot {
	b.Client = &http.Client{}
	return b
}

// GetStreamList gets the raw http response when requesting all public streams.
func (b *Bot) GetStreamList() (*http.Response, error) {
	req, err := b.constructRequest("GET", "streams", "")
	if err != nil {
		return nil, err
	}

	return b.Client.Do(req)
}

type StreamJSON struct {
	Msg     string `json:"msg"`
	Streams []struct {
		StreamID    int    `json:"stream_id"`
		InviteOnly  bool   `json:"invite_only"`
		Description string `json:"description"`
		Name        string `json:"name"`
	} `json:"streams"`
	Result string `json:"result"`
}

// GetStreams returns a list of all public streams
func (b *Bot) GetStreams() ([]string, error) {
	resp, err := b.GetStreamList()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sj StreamJSON
	err = json.Unmarshal(body, &sj)
	if err != nil {
		return nil, err
	}

	var streams []string
	for _, s := range sj.Streams {
		streams = append(streams, s.Name)
	}

	return streams, nil
}

// GetStreams returns a list of all public streams
func (b *Bot) GetRawStreams() (StreamJSON, error) {
	var sj StreamJSON
	resp, err := b.GetStreamList()
	if err != nil {
		return sj, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sj, err
	}

	err = json.Unmarshal(body, &sj)
	if err != nil {
		return sj, err
	}
	return sj, nil
}

// Subscribe will set the bot to receive messages from the given streams.
// If no streams are given, it will subscribe the bot to the streams in the bot struct.
func (b *Bot) Subscribe(streams []string) (*http.Response, error) {
	if streams == nil {
		streams = b.Streams
	}

	var toSubStreams []map[string]string
	for _, name := range streams {
		toSubStreams = append(toSubStreams, map[string]string{"name": name})
	}

	bodyBts, err := json.Marshal(toSubStreams)
	if err != nil {
		return nil, err
	}

	body := "subscriptions=" + string(bodyBts)

	req, err := b.constructRequest("POST", "users/me/subscriptions", body)
	if b.UserAgent != "" {
		req.Header.Set("User-Agent", b.UserAgent)
	} else {
		req.Header.Set("User-Agent", fmt.Sprintf("gozulipbot/%s", Release))
	}
	if err != nil {
		return nil, err
	}

	return b.Client.Do(req)
}

// Unsubscribe will remove the bot from the given streams.
// If no streams are given, nothing will happen and the function will error.
func (b *Bot) Unsubscribe(streams []string) (*http.Response, error) {
	if len(streams) == 0 {
		return nil, fmt.Errorf("No streams were provided")
	}

	body := `delete=["` + strings.Join(streams, `","`) + `"]`

	req, err := b.constructRequest("PATCH", "users/me/subscriptions", body)
	if err != nil {
		return nil, err
	}

	return b.Client.Do(req)
}

func (b *Bot) ListSubscriptions() (*http.Response, error) {
	req, err := b.constructRequest("GET", "users/me/subscriptions", "")
	if err != nil {
		return nil, err
	}

	return b.Client.Do(req)
}

type EventType string

const (
	Messages      EventType = "messages"
	Subscriptions EventType = "subscriptions"
	RealmUser     EventType = "realm_user"
	Pointer       EventType = "pointer"
)

type Narrow string

const (
	NarrowPrivate Narrow = `[["is", "private"]]`
	NarrowAt      Narrow = `[["is", "mentioned"]]`
)

// RegisterEvents adds a queue to the bot. It includes the EventTypes and
// Narrow given. If neither is given, it will default to all Messages.
func (b *Bot) RegisterEvents(ets []EventType, n Narrow) (*Queue, error) {
	resp, err := b.RawRegisterEvents(ets, n)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// Try to parse the error out of the body
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			var jsonErr map[string]string
			err = json.Unmarshal(body, &jsonErr)
			if err == nil {
				if msg, ok := jsonErr["msg"]; ok {
					return nil, fmt.Errorf("Failed to register: %s", msg)
				}
			}
		}
		return nil, fmt.Errorf("Got non-200 response code when registering: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	q := &Queue{Bot: b}
	err = json.Unmarshal(body, q)
	if err != nil {
		return nil, err
	}

	b.Queues = append(b.Queues, q)

	return q, nil
}

func (b *Bot) RegisterAll() (*Queue, error) {
	return b.RegisterEvents(nil, "")
}

func (b *Bot) RegisterAt() (*Queue, error) {
	return b.RegisterEvents(nil, NarrowAt)
}

func (b *Bot) RegisterPrivate() (*Queue, error) {
	return b.RegisterEvents(nil, NarrowPrivate)
}

func (b *Bot) RegisterSubscriptions() (*Queue, error) {
	events := []EventType{Subscriptions}
	return b.RegisterEvents(events, "")
}

// RawRegisterEvents tells Zulip to include message events in the bots events queue.
// Passing nil as the slice of EventType will default to receiving Messages
func (b *Bot) RawRegisterEvents(ets []EventType, n Narrow) (*http.Response, error) {
	// default to Messages if no EventTypes given
	query := `event_types=["message"]`

	if len(ets) != 0 {
		query = `event_types=["`
		for i, s := range ets {
			query += fmt.Sprintf("%s", s)
			if i != len(ets)-1 {
				query += `", "`
			}
		}
		query += `"]`
	}

	if n != "" {
		query += fmt.Sprintf("&narrow=%s", n)
	}
	query += fmt.Sprintf("&all_public_streams=true")
	req, err := b.constructRequest("POST", "register", query)
	if err != nil {
		return nil, err
	}

	return b.Client.Do(req)
}

// constructRequest makes a zulip request and ensures the proper headers are set.
func (b *Bot) constructRequest(method, endpoint, body string) (*http.Request, error) {
	url := b.APIURL + endpoint
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(b.Email, b.APIKey)

	return req, nil
}
