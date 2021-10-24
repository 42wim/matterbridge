package bmsteams

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/davecgh/go-spew/spew"

	"github.com/mattn/godown"
	msgraph "github.com/yaegashi/msgraph.go/beta"
	"github.com/yaegashi/msgraph.go/msauth"

	"golang.org/x/oauth2"
)

var (
	defaultScopes = []string{"openid", "profile", "offline_access", "Group.Read.All", "Group.ReadWrite.All"}
	attachRE      = regexp.MustCompile(`<attachment id=.*?attachment>`)
)

type Bmsteams struct {
	gc    *msgraph.GraphServiceRequestBuilder
	ctx   context.Context
	botID string
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bmsteams{Config: cfg}
}

func (b *Bmsteams) Connect() error {
	tokenCachePath := b.GetString("sessionFile")
	if tokenCachePath == "" {
		tokenCachePath = "msteams_session.json"
	}
	ctx := context.Background()
	m := msauth.NewManager()
	m.LoadFile(tokenCachePath) //nolint:errcheck
	ts, err := m.DeviceAuthorizationGrant(ctx, b.GetString("TenantID"), b.GetString("ClientID"), defaultScopes, nil)
	if err != nil {
		return err
	}
	err = m.SaveFile(tokenCachePath)
	if err != nil {
		b.Log.Errorf("Couldn't save sessionfile in %s: %s", tokenCachePath, err)
	}
	// make file readable only for matterbridge user
	err = os.Chmod(tokenCachePath, 0o600)
	if err != nil {
		b.Log.Errorf("Couldn't change permissions for %s: %s", tokenCachePath, err)
	}
	httpClient := oauth2.NewClient(ctx, ts)
	graphClient := msgraph.NewClient(httpClient)
	b.gc = graphClient
	b.ctx = ctx

	err = b.setBotID()
	if err != nil {
		return err
	}
	b.Log.Info("Connection succeeded")
	return nil
}

func (b *Bmsteams) Disconnect() error {
	return nil
}

func (b *Bmsteams) JoinChannel(channel config.ChannelInfo) error {
	go func(name string) {
		for {
			err := b.poll(name)
			if err != nil {
				b.Log.Errorf("polling failed for %s: %s. retrying in 5 seconds", name, err)
			}
			time.Sleep(time.Second * 5)
		}
	}(channel.Name)
	return nil
}

func (b *Bmsteams) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)
	if msg.ParentValid() {
		return b.sendReply(msg)
	}

	// Handle prefix hint for unthreaded messages.
	if msg.ParentNotFound() {
		msg.ParentID = ""
		msg.Text = fmt.Sprintf("[thread]: %s", msg.Text)
	}

	ct := b.gc.Teams().ID(b.GetString("TeamID")).Channels().ID(msg.Channel).Messages().Request()
	text := msg.Username + msg.Text
	content := &msgraph.ItemBody{Content: &text}
	rmsg := &msgraph.ChatMessage{Body: content}
	res, err := ct.Add(b.ctx, rmsg)
	if err != nil {
		return "", err
	}
	return *res.ID, nil
}

func (b *Bmsteams) sendReply(msg config.Message) (string, error) {
	ct := b.gc.Teams().ID(b.GetString("TeamID")).Channels().ID(msg.Channel).Messages().ID(msg.ParentID).Replies().Request()
	// Handle prefix hint for unthreaded messages.

	text := msg.Username + msg.Text
	content := &msgraph.ItemBody{Content: &text}
	rmsg := &msgraph.ChatMessage{Body: content}
	res, err := ct.Add(b.ctx, rmsg)
	if err != nil {
		return "", err
	}
	return *res.ID, nil
}

func (b *Bmsteams) getMessages(channel string) ([]msgraph.ChatMessage, error) {
	ct := b.gc.Teams().ID(b.GetString("TeamID")).Channels().ID(channel).Messages().Request()
	rct, err := ct.Get(b.ctx)
	if err != nil {
		return nil, err
	}
	b.Log.Debugf("got %#v messages", len(rct))
	return rct, nil
}

//nolint:gocognit
func (b *Bmsteams) poll(channelName string) error {
	msgmap := make(map[string]time.Time)
	b.Log.Debug("getting initial messages")
	res, err := b.getMessages(channelName)
	if err != nil {
		return err
	}
	for _, msg := range res {
		msgmap[*msg.ID] = *msg.CreatedDateTime
		if msg.LastModifiedDateTime != nil {
			msgmap[*msg.ID] = *msg.LastModifiedDateTime
		}
	}
	time.Sleep(time.Second * 5)
	b.Log.Debug("polling for messages")
	for {
		res, err := b.getMessages(channelName)
		if err != nil {
			return err
		}
		for i := len(res) - 1; i >= 0; i-- {
			msg := res[i]
			if mtime, ok := msgmap[*msg.ID]; ok {
				if mtime == *msg.CreatedDateTime && msg.LastModifiedDateTime == nil {
					continue
				}
				if msg.LastModifiedDateTime != nil && mtime == *msg.LastModifiedDateTime {
					continue
				}
			}

			if b.GetBool("debug") {
				b.Log.Debug("Msg dump: ", spew.Sdump(msg))
			}

			// skip non-user message for now.
			if msg.From == nil || msg.From.User == nil {
				continue
			}

			if *msg.From.User.ID == b.botID {
				b.Log.Debug("skipping own message")
				msgmap[*msg.ID] = *msg.CreatedDateTime
				continue
			}

			msgmap[*msg.ID] = *msg.CreatedDateTime
			if msg.LastModifiedDateTime != nil {
				msgmap[*msg.ID] = *msg.LastModifiedDateTime
			}
			b.Log.Debugf("<= Sending message from %s on %s to gateway", *msg.From.User.DisplayName, b.Account)
			text := b.convertToMD(*msg.Body.Content)
			rmsg := config.Message{
				Username: *msg.From.User.DisplayName,
				Text:     text,
				Channel:  channelName,
				Account:  b.Account,
				Avatar:   "",
				UserID:   *msg.From.User.ID,
				ID:       *msg.ID,
				Extra:    make(map[string][]interface{}),
			}

			b.handleAttachments(&rmsg, msg)
			b.Log.Debugf("<= Message is %#v", rmsg)
			b.Remote <- rmsg
		}
		time.Sleep(time.Second * 5)
	}
}

func (b *Bmsteams) setBotID() error {
	req := b.gc.Me().Request()
	r, err := req.Get(b.ctx)
	if err != nil {
		return err
	}
	b.botID = *r.ID
	return nil
}

func (b *Bmsteams) convertToMD(text string) string {
	if !strings.Contains(text, "<div>") {
		return text
	}
	var sb strings.Builder
	err := godown.Convert(&sb, strings.NewReader(text), nil)
	if err != nil {
		b.Log.Errorf("Couldn't convert message to markdown %s", text)
		return text
	}
	return sb.String()
}
