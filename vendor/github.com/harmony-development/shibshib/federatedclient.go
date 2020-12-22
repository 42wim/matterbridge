package shibshib

import (
	"fmt"
	"reflect"
	"strings"

	authv1 "github.com/harmony-development/shibshib/gen/auth/v1"
	chatv1 "github.com/harmony-development/shibshib/gen/chat/v1"
	types "github.com/harmony-development/shibshib/gen/harmonytypes/v1"
)

type FederatedClient struct {
	Client

	homeserver string
	subclients map[string]*Client
	streams    map[<-chan *types.Message]*Client
	listening  map[*Client]<-chan *types.Message
}

type FederatedEvent struct {
	Client *Client
	Event  *types.Message
}

func NewFederatedClient(homeserver, token string, userID uint64) (*FederatedClient, error) {
	self := &FederatedClient{}
	self.homeserver = homeserver
	self.Client.homeserver = homeserver
	self.init(homeserver)
	self.authed(token, userID)

	self.subclients = make(map[string]*Client)
	self.streams = make(map[<-chan *types.Message]*Client)
	self.listening = make(map[*Client]<-chan *types.Message)

	err := self.StreamEvents()
	if err != nil {
		return nil, fmt.Errorf("NewFederatedClient: own streamevents failed: %w", err)
	}

	return self, nil
}

func (f *FederatedClient) clientFor(homeserver string) (*Client, error) {
	if f.homeserver == homeserver || strings.Split(homeserver, ":")[0] == "localhost" || homeserver == "" {
		return &f.Client, nil
	}

	if val, ok := f.subclients[homeserver]; ok {
		return val, nil
	}

	session, err := f.AuthKit.Federate(&authv1.FederateRequest{
		Target: homeserver,
	})
	if err != nil {
		return nil, fmt.Errorf("ClientFor: homeserver federation step failed: %w", err)
	}

	c := new(Client)
	c.init(homeserver)

	data, err := c.AuthKit.LoginFederated(&authv1.LoginFederatedRequest{
		AuthToken: session.Token,
		Domain:    f.homeserver,
	})
	if err != nil {
		return nil, fmt.Errorf("ClientFor: failed to log into foreignserver: %w", err)
	}

	c.authed(data.SessionToken, data.UserId)
	err = c.StreamEvents()
	if err != nil {
		return nil, fmt.Errorf("ClientFor: failed to stream events for foreign server: %w", err)
	}

	f.subclients[homeserver] = c

	return c, nil
}

func (f *FederatedClient) Start() (<-chan FederatedEvent, error) {
	list, err := f.ChatKit.GetGuildList(&chatv1.GetGuildListRequest{})
	if err != nil {
		return nil, fmt.Errorf("Start: failed to get guild list on homeserver: %w", err)
	}

	cases := []reflect.SelectCase{}

	for _, g := range list.Guilds {
		client, err := f.clientFor(g.Host)
		if err != nil {
			return nil, fmt.Errorf("Start: failed to get client for guild %s/%d: %w", g.Host, g.GuildId, err)
		}

		stream, ok := f.listening[client]
		if !ok {
			stream = client.EventsStream()
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(stream),
			})
			client.ErrorHandler = func(e error) {
				if e == ErrEndOfStream {
					err := client.StreamEvents()
					if err != nil {
						panic(fmt.Errorf("client.ErrorHandler: could not restart stream: %w", err))
					}
					cp := client.subscribedGuilds
					client.subscribedGuilds = make([]uint64, 0)
					for _, gg := range cp {
						client.SubscribeToGuild(gg)
					}
				} else {
					panic(err)
				}
			}

			f.listening[client] = stream
			f.streams[stream] = client
		}

		client.SubscribeToGuild(g.GuildId)
	}

	channel := make(chan FederatedEvent)
	go func() {
		for {
			i, v, ok := reflect.Select(cases)
			if !ok {
				cases = append(cases[:i], cases[i+1:]...)
			}

			val := v.Interface().(*types.Message)

			channel <- FederatedEvent{
				Event:  val,
				Client: f.streams[cases[i].Chan.Interface().(<-chan *types.Message)],
			}
		}
	}()

	return channel, nil
}
