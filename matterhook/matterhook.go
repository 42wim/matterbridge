//Package matterhook provides interaction with mattermost incoming/outgoing webhooks
package matterhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// OMessage for mattermost incoming webhook. (send to mattermost)
type OMessage struct {
	Channel   string `json:"channel,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	UserName  string `json:"username,omitempty"`
	Text      string `json:"text"`
}

// IMessage for mattermost outgoing webhook. (received from mattermost)
type IMessage struct {
	Token       string `schema:"token"`
	TeamID      string `schema:"team_id"`
	TeamDomain  string `schema:"team_domain"`
	ChannelID   string `schema:"channel_id"`
	ServiceID   string `schema:"service_id"`
	ChannelName string `schema:"channel_name"`
	Timestamp   string `schema:"timestamp"`
	UserID      string `schema:"user_id"`
	UserName    string `schema:"user_name"`
	Text        string `schema:"text"`
	TriggerWord string `schema:"trigger_word"`
}

// Client for Mattermost.
type Client struct {
	url string
	In  chan IMessage
	Out chan OMessage
	Config
}

type Config struct {
	Port int
}

// New Mattermost client.
func New(url string, config Config) *Client {
	c := &Client{url: url, In: make(chan IMessage), Out: make(chan OMessage), Config: config}
	if c.Port == 0 {
		c.Port = 9999
	}
	go c.StartServer()
	return c
}

// StartServer starts a webserver listening for incoming mattermost POSTS.
func (c *Client) StartServer() {
	mux := http.NewServeMux()
	mux.Handle("/", c)
	log.Printf("Listening on http://0.0.0.0:%v...\n", c.Port)
	if err := http.ListenAndServe((":" + strconv.Itoa(c.Port)), mux); err != nil {
		log.Fatal(err)
	}
}

// ServeHTTP implementation.
func (c *Client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("invalid " + r.Method + " connection from " + r.RemoteAddr)
		http.NotFound(w, r)
		return
	}
	msg := IMessage{}
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}
	defer r.Body.Close()
	decoder := schema.NewDecoder()
	err = decoder.Decode(&msg, r.PostForm)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}
	if msg.Token == "" {
		log.Println("no token from " + r.RemoteAddr)
		http.NotFound(w, r)
		return
	}
	c.In <- msg
}

// Receive returns an incoming message from mattermost outgoing webhooks URL.
func (c *Client) Receive() IMessage {
	for {
		select {
		case msg := <-c.In:
			return msg
		}
	}
}

// Send sends a msg to mattermost incoming webhooks URL.
func (c *Client) Send(msg OMessage) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read entire body to completion to re-use keep-alive connections.
	io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
