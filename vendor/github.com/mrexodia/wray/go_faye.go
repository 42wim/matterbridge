package wray

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	UNCONNECTED  = 1
	CONNECTING   = 2
	CONNECTED    = 3
	DISCONNECTED = 4

	HANDSHAKE = "handshake"
	RETRY     = "retry"
	NONE      = "none"

	CONNECTION_TIMEOUT = 60.0
	DEFAULT_RETRY      = 5.0
	MAX_REQUEST_SIZE   = 2048
)

var (
	MANDATORY_CONNECTION_TYPES = []string{"long-polling"}
	registeredTransports       = []Transport{}
)

type FayeClient struct {
	state         int
	url           string
	subscriptions []Subscription
	transport     Transport
	clientId      string
	schedular     Schedular
}

type Subscription struct {
	channel  string
	callback func(Message)
}

type SubscriptionPromise struct {
	subscription Subscription
}

func NewFayeClient(url string) *FayeClient {
	schedular := ChannelSchedular{}
	client := &FayeClient{url: url, state: UNCONNECTED, schedular: schedular}
	return client
}

func (self *FayeClient) handshake() {
	t, err := SelectTransport(self, MANDATORY_CONNECTION_TYPES, []string{})
	if err != nil {
		panic("No usable transports available")
	}
	self.transport = t
	self.transport.setUrl(self.url)
	self.state = CONNECTING
	handshakeParams := map[string]interface{}{"channel": "/meta/handshake",
		"version":                  "1.0",
		"supportedConnectionTypes": []string{"long-polling"}}
	response, err := self.transport.send(handshakeParams)
	if err != nil {
		fmt.Println("Handshake failed. Retry in 10 seconds")
		self.state = UNCONNECTED
		self.schedular.wait(10*time.Second, func() {
			fmt.Println("retying handshake")
			self.handshake()
		})
		return
	}
	self.clientId = response.clientId
	self.state = CONNECTED
	self.transport, err = SelectTransport(self, response.supportedConnectionTypes, []string{})
	if err != nil {
		panic("Server does not support any available transports. Supported transports: " + strings.Join(response.supportedConnectionTypes, ","))
	}
}

func (self *FayeClient) Subscribe(channel string, force bool, callback func(Message)) SubscriptionPromise {
	if self.state == UNCONNECTED {
		self.handshake()
	}
	subscriptionParams := map[string]interface{}{"channel": "/meta/subscribe", "clientId": self.clientId, "subscription": channel, "id": "1"}
	subscription := Subscription{channel: channel, callback: callback}
	//TODO: deal with subscription failures
	self.transport.send(subscriptionParams)
	self.subscriptions = append(self.subscriptions, subscription)
	return SubscriptionPromise{subscription}
}

func (self *FayeClient) handleResponse(response Response) {
	for _, message := range response.messages {
		for _, subscription := range self.subscriptions {
			matched, _ := filepath.Match(subscription.channel, message.Channel)
			if matched {
				go subscription.callback(message)
			}
		}
	}
}

func (self *FayeClient) connect() {
	connectParams := map[string]interface{}{"channel": "/meta/connect", "clientId": self.clientId, "connectionType": self.transport.connectionType()}
	responseChannel := make(chan Response)
	go func() {
		response, _ := self.transport.send(connectParams)
		responseChannel <- response
	}()
	self.listen(responseChannel)
}

func (self *FayeClient) listen(responseChannel chan Response) {
	response := <-responseChannel
	if response.successful == true {
		go self.handleResponse(response)
	} else {
	}
}

func (self *FayeClient) Listen() {
	for {
		self.connect()
	}
}

func (self *FayeClient) Publish(channel string, data map[string]interface{}) {
	if self.state != CONNECTED {
		self.handshake()
	}
	publishParams := map[string]interface{}{"channel": channel, "data": data, "clientId": self.clientId}
	self.transport.send(publishParams)
}

func RegisterTransports(transports []Transport) {
	registeredTransports = transports
}
