package rockethook

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
)

// Message for rocketchat outgoing webhook.
type Message struct {
	Token       string `json:"token"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Timestamp   string `json:"timestamp"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Text        string `json:"text"`
}

// Client for Rocketchat.
type Client struct {
	In         chan Message
	httpclient *http.Client
	Config
}

// Config for client.
type Config struct {
	BindAddress        string // Address to listen on
	Token              string // Only allow this token from Rocketchat. (Allow everything when empty)
	InsecureSkipVerify bool   // disable certificate checking
}

// New Rocketchat client.
func New(url string, config Config) *Client {
	c := &Client{In: make(chan Message), Config: config}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}, //nolint:gosec
	}
	c.httpclient = &http.Client{Transport: tr}
	_, _, err := net.SplitHostPort(c.BindAddress)
	if err != nil {
		log.Fatalf("incorrect bindaddress %s", c.BindAddress)
	}
	go c.StartServer()
	return c
}

// StartServer starts a webserver listening for incoming mattermost POSTS.
func (c *Client) StartServer() {
	mux := http.NewServeMux()
	mux.Handle("/", c)
	log.Printf("Listening on http://%v...\n", c.BindAddress)
	if err := http.ListenAndServe(c.BindAddress, mux); err != nil {
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
	msg := Message{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}
	defer r.Body.Close()
	err = json.Unmarshal(body, &msg)
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
	msg.ChannelName = "#" + msg.ChannelName
	if c.Token != "" {
		if msg.Token != c.Token {
			if regexp.MustCompile(`[^a-zA-Z0-9]+`).MatchString(msg.Token) {
				log.Println("invalid token " + msg.Token + " from " + r.RemoteAddr)
			} else {
				log.Println("invalid token from " + r.RemoteAddr)
			}
			http.NotFound(w, r)
			return
		}
	}
	c.In <- msg
}

// Receive returns an incoming message from mattermost outgoing webhooks URL.
func (c *Client) Receive() Message {
	var msg Message
	for msg = range c.In {
		return msg
	}
	return msg
}
