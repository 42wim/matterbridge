//Package matterhook provides interaction with mattermost incoming/outgoing webhooks
package matterhook

import (
	"bytes"
	"crypto/tls"
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
	Channel     string      `json:"channel,omitempty"`
	IconURL     string      `json:"icon_url,omitempty"`
	IconEmoji   string      `json:"icon_emoji,omitempty"`
	UserName    string      `json:"username,omitempty"`
	Text        string      `json:"text"`
	Attachments interface{} `json:"attachments,omitempty"`
	Type        string      `json:"type,omitempty"`
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
	Url        string // URL for incoming webhooks on mattermost.
	In         chan IMessage
	Out        chan OMessage
	httpclient *http.Client
	Config
}

// Config for client.
type Config struct {
	Port               int    // Port to listen on.
	BindAddress        string // Address to listen on
	Token              string // Only allow this token from Mattermost. (Allow everything when empty)
	InsecureSkipVerify bool   // disable certificate checking
	DisableServer      bool   // Do not start server for outgoing webhooks from Mattermost.
}

// New Mattermost client.
func New(url string, config Config) *Client {
	c := &Client{Url: url, In: make(chan IMessage), Out: make(chan OMessage), Config: config}
	if c.Port == 0 {
		c.Port = 9999
	}
	c.BindAddress += ":"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify},
	}
	c.httpclient = &http.Client{Transport: tr}
	if !c.DisableServer {
		go c.StartServer()
	}
	return c
}

// StartServer starts a webserver listening for incoming mattermost POSTS.
func (c *Client) StartServer() {
	mux := http.NewServeMux()
	mux.Handle("/", c)
	log.Printf("Listening on http://%v:%v...\n", c.BindAddress, c.Port)
	if err := http.ListenAndServe((c.BindAddress + strconv.Itoa(c.Port)), mux); err != nil {
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
	if c.Token != "" {
		if msg.Token != c.Token {
			log.Println("invalid token " + msg.Token + " from " + r.RemoteAddr)
			http.NotFound(w, r)
			return
		}
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
	resp, err := c.httpclient.Post(c.Url, "application/json", bytes.NewReader(buf))
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
