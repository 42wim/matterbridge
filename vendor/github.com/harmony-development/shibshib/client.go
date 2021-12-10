package shibshib

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	authv1 "github.com/harmony-development/shibshib/gen/auth/v1"
	chatv1 "github.com/harmony-development/shibshib/gen/chat/v1"
	profilev1 "github.com/harmony-development/shibshib/gen/profile/v1"
)

type Client struct {
	ChatKit    chatv1.HTTPChatServiceClient
	AuthKit    authv1.HTTPAuthServiceClient
	ProfileKit profilev1.HTTPProfileServiceClient

	ErrorHandler func(error)

	UserID uint64

	incomingEvents <-chan *chatv1.StreamEventsResponse
	outgoingEvents chan<- *chatv1.StreamEventsRequest

	subscribedGuilds []uint64
	onceHandlers     []func(*LocatedMessage)

	events       chan *LocatedMessage
	homeserver   string
	sessionToken string

	streaming bool

	mtx *sync.Mutex
}

var ErrEndOfStream = errors.New("end of stream")

func (c *Client) init(h string, wsp, wsph string) {
	c.events = make(chan *LocatedMessage)
	c.mtx = new(sync.Mutex)
	c.ErrorHandler = func(e error) {
		panic(e)
	}
	c.homeserver = h
	c.ChatKit = chatv1.HTTPChatServiceClient{*http.DefaultClient, h, wsp, wsph, http.Header{}}
	c.AuthKit = authv1.HTTPAuthServiceClient{*http.DefaultClient, h, wsp, wsph, http.Header{}}
	c.ProfileKit = profilev1.HTTPProfileServiceClient{*http.DefaultClient, h, wsp, wsph, http.Header{}}
}

func (c *Client) authed(token string, userID uint64) {
	c.sessionToken = token
	c.ChatKit.Header.Add("Authorization", token)
	c.AuthKit.Header.Add("Authorization", token)
	c.ProfileKit.Header.Add("Authorization", token)
	c.UserID = userID
}

func NewClient(homeserver, token string, userid uint64) (ret *Client, err error) {
	url, err := url.Parse(homeserver)
	if err != nil {
		return nil, err
	}
	it := "wss"
	if url.Scheme == "http" {
		it = "ws"
	}
	ret = &Client{}
	ret.homeserver = homeserver
	ret.init(homeserver, it, url.Host)
	ret.authed(token, userid)

	err = ret.StreamEvents()
	if err != nil {
		ret = nil
		return
	}

	return
}

func (c *Client) StreamEvents() (err error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.streaming {
		return
	}

	it := make(chan *chatv1.StreamEventsRequest)
	c.outgoingEvents = it
	c.incomingEvents, err = c.ChatKit.StreamEvents(it)
	if err != nil {
		err = fmt.Errorf("StreamEvents: failed to open stream: %w", err)
		return
	}

	c.streaming = true

	go func() {
		for ev := range c.incomingEvents {
			chat, ok := ev.Event.(*chatv1.StreamEventsResponse_Chat)
			if !ok {
				continue
			}

			msg, ok := chat.Chat.Event.(*chatv1.StreamEvent_SentMessage)
			if !ok {
				continue
			}

			imsg := &LocatedMessage{
				GuildID:   msg.SentMessage.GuildId,
				ChannelID: msg.SentMessage.ChannelId,
				MessageWithId: chatv1.MessageWithId{
					MessageId: msg.SentMessage.MessageId,
					Message:   msg.SentMessage.Message,
				},
			}

			for _, h := range c.onceHandlers {
				h(imsg)
			}
			c.onceHandlers = make([]func(*LocatedMessage), 0)
			c.events <- imsg
		}

		c.mtx.Lock()
		defer c.mtx.Unlock()

		c.streaming = false
		c.ErrorHandler(ErrEndOfStream)
	}()

	return nil
}

func (c *Client) SubscribeToGuild(community uint64) {
	for _, g := range c.subscribedGuilds {
		if g == community {
			return
		}
	}
	c.outgoingEvents <- &chatv1.StreamEventsRequest{
		Request: &chatv1.StreamEventsRequest_SubscribeToGuild_{
			SubscribeToGuild: &chatv1.StreamEventsRequest_SubscribeToGuild{
				GuildId: community,
			},
		},
	}
	c.subscribedGuilds = append(c.subscribedGuilds, community)
}

func (c *Client) SubscribedGuilds() []uint64 {
	return c.subscribedGuilds
}

func (c *Client) SendMessage(msg *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	return c.ChatKit.SendMessage(msg)
}

func (c *Client) TransformHMCURL(hmc string) string {
	if !strings.HasPrefix(hmc, "hmc://") {
		return fmt.Sprintf("%s/_harmony/media/download/%s", c.homeserver, hmc)
	}

	trimmed := strings.TrimPrefix(hmc, "hmc://")
	split := strings.Split(trimmed, "/")
	if len(split) != 2 {
		return fmt.Sprintf("malformed URL: %s", hmc)
	}

	return fmt.Sprintf("https://%s/_harmony/media/download/%s", split[0], split[1])
}

func (c *Client) UsernameFor(m *chatv1.Message) string {
	if m.Overrides != nil {
		return m.Overrides.GetUsername()
	}

	resp, err := c.ProfileKit.GetProfile(&profilev1.GetProfileRequest{
		UserId: m.AuthorId,
	})
	if err != nil {
		return strconv.FormatUint(m.AuthorId, 10)
	}

	return resp.Profile.UserName
}

func (c *Client) AvatarFor(m *chatv1.Message) string {
	if m.Overrides != nil {
		return m.Overrides.GetAvatar()
	}

	resp, err := c.ProfileKit.GetProfile(&profilev1.GetProfileRequest{
		UserId: m.AuthorId,
	})
	if err != nil {
		return ""
	}

	return c.TransformHMCURL(resp.Profile.GetUserAvatar())
}

func (c *Client) EventsStream() <-chan *LocatedMessage {
	return c.events
}

func (c *Client) HandleOnce(f func(*LocatedMessage)) {
	c.onceHandlers = append(c.onceHandlers, f)
}
