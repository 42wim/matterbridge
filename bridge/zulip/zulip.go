package bzulip

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/version"
	gzb "github.com/matterbridge/gozulipbot"
)

type Bzulip struct {
	q       *gzb.Queue
	bot     *gzb.Bot
	streams map[int]string
	*bridge.Config
	sync.RWMutex
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bzulip{Config: cfg, streams: make(map[int]string)}
}

func (b *Bzulip) Connect() error {
	bot := gzb.Bot{APIKey: b.GetString("token"), APIURL: b.GetString("server") + "/api/v1/", Email: b.GetString("login"), UserAgent: fmt.Sprintf("matterbridge/%s", version.Release)}
	bot.Init()
	q, err := bot.RegisterAll()
	b.q = q
	b.bot = &bot
	if err != nil {
		b.Log.Errorf("Connect() %#v", err)
		return err
	}
	// init stream
	b.getChannel(0)
	b.Log.Info("Connection succeeded")
	go b.handleQueue()
	return nil
}

func (b *Bzulip) Disconnect() error {
	return nil
}

func (b *Bzulip) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Bzulip) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}
		_, err := b.bot.UpdateMessage(msg.ID, "")
		return "", err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.sendMessage(rmsg)
		}
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg)
		}
	}

	// edit the message if we have a msg ID
	if msg.ID != "" {
		_, err := b.bot.UpdateMessage(msg.ID, msg.Username+msg.Text)
		return "", err
	}

	// Post normal message
	return b.sendMessage(msg)
}

func (b *Bzulip) getChannel(id int) string {
	if name, ok := b.streams[id]; ok {
		return name
	}
	streams, err := b.bot.GetRawStreams()
	if err != nil {
		b.Log.Errorf("getChannel: %#v", err)
		return ""
	}
	for _, stream := range streams.Streams {
		b.streams[stream.StreamID] = stream.Name
	}
	if name, ok := b.streams[id]; ok {
		return name
	}
	return ""
}

func (b *Bzulip) handleQueue() error {
	for {
		messages, err := b.q.GetEvents()
		if err != nil {
			switch err {
			case gzb.BackoffError:
				time.Sleep(time.Second * 5)
			case gzb.NoJSONError:
				b.Log.Error("Response wasn't JSON, server down or restarting? sleeping 10 seconds")
				time.Sleep(time.Second * 10)
			case gzb.BadEventQueueError:
				b.Log.Info("got a bad event queue id error, reconnecting")
				b.bot.Queues = nil
				for {
					b.q, err = b.bot.RegisterAll()
					if err != nil {
						b.Log.Errorf("reconnecting failed: %s. Sleeping 10 seconds", err)
						time.Sleep(time.Second * 10)
					}
					break
				}
			case gzb.HeartbeatError:
				b.Log.Debug("heartbeat received.")
			default:
				b.Log.Debugf("receiving error: %#v", err)
				time.Sleep(time.Second * 10)
			}

			continue
		}
		for _, m := range messages {
			b.Log.Debugf("== Receiving %#v", m)
			// ignore our own messages
			if m.SenderEmail == b.GetString("login") {
				continue
			}

			avatarURL := m.AvatarURL
			if !strings.HasPrefix(avatarURL, "http") {
				avatarURL = b.GetString("server") + avatarURL
			}

			rmsg := config.Message{
				Username: m.SenderFullName,
				Text:     m.Content,
				Channel:  b.getChannel(m.StreamID) + "/topic:" + m.Subject,
				Account:  b.Account,
				UserID:   strconv.Itoa(m.SenderID),
				Avatar:   avatarURL,
			}
			b.Log.Debugf("<= Sending message from %s on %s to gateway", rmsg.Username, b.Account)
			b.Log.Debugf("<= Message is %#v", rmsg)
			b.Remote <- rmsg
		}

		time.Sleep(time.Second * 3)
	}
}

func (b *Bzulip) sendMessage(msg config.Message) (string, error) {
	topic := ""
	if strings.Contains(msg.Channel, "/topic:") {
		res := strings.Split(msg.Channel, "/topic:")
		topic = res[1]
		msg.Channel = res[0]
	}
	m := gzb.Message{
		Stream:  msg.Channel,
		Topic:   topic,
		Content: msg.Username + msg.Text,
	}
	resp, err := b.bot.Message(m)
	if err != nil {
		return "", err
	}
	if resp != nil {
		defer resp.Body.Close()
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		var jr struct {
			ID int `json:"id"`
		}
		err = json.Unmarshal(res, &jr)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(jr.ID), nil
	}
	return "", nil
}

func (b *Bzulip) handleUploadFile(msg *config.Message) (string, error) {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		if fi.Comment != "" {
			msg.Text += fi.Comment + ": "
		}
		if fi.URL != "" {
			msg.Text = fi.URL
			if fi.Comment != "" {
				msg.Text = fi.Comment + ": " + fi.URL
			}
		}
		_, err := b.sendMessage(*msg)
		if err != nil {
			return "", err
		}
	}
	return "", nil
}
