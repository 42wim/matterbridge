package bmattermost

import (
	"errors"
	"fmt"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	matterclient6 "github.com/matterbridge/matterclient"
	"github.com/rs/xid"
)

type Bmattermost struct {
	mh     *matterhook.Client
	mc     *matterclient.MMClient
	mc6    *matterclient6.Client
	v6     bool
	uuid   string
	TeamID string
	*bridge.Config
	avatarMap map[string]string
}

const mattermostPlugin = "mattermost.plugin"

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmattermost{Config: cfg, avatarMap: make(map[string]string)}

	b.v6 = b.GetBool("v6")
	b.uuid = xid.New().String()

	return b
}

func (b *Bmattermost) Command(cmd string) string {
	return ""
}

func (b *Bmattermost) Connect() error {
	if b.Account == mattermostPlugin {
		return nil
	}

	if strings.HasPrefix(b.getVersion(), "6.") {
		if !b.v6 {
			b.v6 = true
		}
	}

	if b.GetString("WebhookBindAddress") != "" {
		if err := b.doConnectWebhookBind(); err != nil {
			return err
		}
		go b.handleMatter()
		return nil
	}
	switch {
	case b.GetString("WebhookURL") != "":
		if err := b.doConnectWebhookURL(); err != nil {
			return err
		}
		go b.handleMatter()
		return nil
	case b.GetString("Token") != "":
		b.Log.Info("Connecting using token (sending and receiving)")
		b.Log.Infof("Using mattermost v6 methods: %t", b.v6)

		if b.v6 {
			err := b.apiLogin6()
			if err != nil {
				return err
			}
		} else {
			err := b.apiLogin()
			if err != nil {
				return err
			}
		}
		go b.handleMatter()
	case b.GetString("Login") != "":
		b.Log.Info("Connecting using login/password (sending and receiving)")
		b.Log.Infof("Using mattermost v6 methods: %t", b.v6)

		if b.v6 {
			err := b.apiLogin6()
			if err != nil {
				return err
			}
		} else {
			err := b.apiLogin()
			if err != nil {
				return err
			}
		}
		go b.handleMatter()
	}
	if b.GetString("WebhookBindAddress") == "" && b.GetString("WebhookURL") == "" &&
		b.GetString("Login") == "" && b.GetString("Token") == "" {
		return errors.New("no connection method found. See that you have WebhookBindAddress, WebhookURL or Token/Login/Password/Server/Team configured")
	}
	return nil
}

func (b *Bmattermost) Disconnect() error {
	return nil
}

func (b *Bmattermost) JoinChannel(channel config.ChannelInfo) error {
	if b.Account == mattermostPlugin {
		return nil
	}
	// we can only join channels using the API
	if b.GetString("WebhookURL") == "" && b.GetString("WebhookBindAddress") == "" {
		var id string
		if b.mc6 != nil {
			id = b.mc6.GetChannelID(channel.Name, b.TeamID)
		} else {
			id = b.mc.GetChannelId(channel.Name, b.TeamID)
		}
		if id == "" {
			return fmt.Errorf("Could not find channel ID for channel %s", channel.Name)
		}

		if b.mc6 != nil {
			return b.mc6.JoinChannel(id) // nolint:wrapcheck
		}

		return b.mc.JoinChannel(id)
	}
	return nil
}

func (b *Bmattermost) Send(msg config.Message) (string, error) {
	if b.Account == mattermostPlugin {
		return "", nil
	}
	b.Log.Debugf("=> Receiving %#v", msg)

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		msg.Text = "*" + msg.Text + "*"
	}

	// map the file SHA to our user (caches the avatar)
	if msg.Event == config.EventAvatarDownload {
		return b.cacheAvatar(&msg)
	}

	// Use webhook to send the message
	if b.GetString("WebhookURL") != "" {
		return b.sendWebhook(msg)
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}
		if b.mc6 != nil {
			return msg.ID, b.mc6.DeleteMessage(msg.ID) // nolint:wrapcheck
		}

		return msg.ID, b.mc.DeleteMessage(msg.ID)
	}

	// Handle prefix hint for unthreaded messages.
	if msg.ParentNotFound() {
		msg.ParentID = ""
		msg.Text = fmt.Sprintf("[thread]: %s", msg.Text)
	}

	// we only can reply to the root of the thread, not to a specific ID (like discord for example does)
	if msg.ParentID != "" {
		if b.mc6 != nil {
			post, _, err := b.mc6.Client.GetPost(msg.ParentID, "")
			if err != nil {
				b.Log.Errorf("getting post %s failed: %s", msg.ParentID, err)
			}
			msg.ParentID = post.RootId
		} else {
			post, res := b.mc.Client.GetPost(msg.ParentID, "")
			if res.Error != nil {
				b.Log.Errorf("getting post %s failed: %s", msg.ParentID, res.Error.DetailedError)
			}
			msg.ParentID = post.RootId
		}
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			if b.mc6 != nil {
				if _, err := b.mc6.PostMessage(b.mc.GetChannelId(rmsg.Channel, b.TeamID), rmsg.Username+rmsg.Text, msg.ParentID); err != nil {
					b.Log.Errorf("PostMessage failed: %s", err)
				}
			} else {
				if _, err := b.mc.PostMessage(b.mc.GetChannelId(rmsg.Channel, b.TeamID), rmsg.Username+rmsg.Text, msg.ParentID); err != nil {
					b.Log.Errorf("PostMessage failed: %s", err)
				}
			}
		}
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg)
		}
	}

	// Prepend nick if configured
	if b.GetBool("PrefixMessagesWithNick") {
		msg.Text = msg.Username + msg.Text
	}

	// Edit message if we have an ID
	if msg.ID != "" {
		if b.mc6 != nil {
			return b.mc6.EditMessage(msg.ID, msg.Text) // nolint:wrapcheck
		}

		return b.mc.EditMessage(msg.ID, msg.Text)
	}

	// Post normal message
	if b.mc6 != nil {
		return b.mc6.PostMessage(b.mc6.GetChannelID(msg.Channel, b.TeamID), msg.Text, msg.ParentID) // nolint:wrapcheck
	}

	return b.mc.PostMessage(b.mc.GetChannelId(msg.Channel, b.TeamID), msg.Text, msg.ParentID)
}
