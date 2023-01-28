package bmattermost

import (
	"net/http"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/matterhook"
	"github.com/matterbridge/matterclient"
	"github.com/mattermost/mattermost-server/v6/model"
)

func (b *Bmattermost) doConnectWebhookBind() error {
	switch {
	case b.GetString("WebhookURL") != "":
		b.Log.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
		b.mh = matterhook.New(b.GetString("WebhookURL"),
			matterhook.Config{
				InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
				BindAddress:        b.GetString("WebhookBindAddress"),
			})
	case b.GetString("Token") != "":
		b.Log.Info("Connecting using token (sending)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
	case b.GetString("Login") != "":
		b.Log.Info("Connecting using login/password (sending)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
	default:
		b.Log.Info("Connecting using webhookbindaddress (receiving)")
		b.mh = matterhook.New(b.GetString("WebhookURL"),
			matterhook.Config{
				InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
				BindAddress:        b.GetString("WebhookBindAddress"),
			})
	}
	return nil
}

func (b *Bmattermost) doConnectWebhookURL() error {
	b.Log.Info("Connecting using webhookurl (sending)")
	b.mh = matterhook.New(b.GetString("WebhookURL"),
		matterhook.Config{
			InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
			DisableServer:      true,
		})
	if b.GetString("Token") != "" {
		b.Log.Info("Connecting using token (receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
	} else if b.GetString("Login") != "" {
		b.Log.Info("Connecting using login/password (receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
	}
	return nil
}

//nolint:wrapcheck
func (b *Bmattermost) apiLogin() error {
	password := b.GetString("Password")
	if b.GetString("Token") != "" {
		password = "token=" + b.GetString("Token")
	}

	b.mc = matterclient.New(b.GetString("Login"), password, b.GetString("Team"), b.GetString("Server"), "")
	if b.GetBool("debug") {
		b.mc.SetLogLevel("debug")
	}
	b.mc.SkipTLSVerify = b.GetBool("SkipTLSVerify")
	b.mc.SkipVersionCheck = b.GetBool("SkipVersionCheck")
	b.mc.NoTLS = b.GetBool("NoTLS")
	b.Log.Infof("Connecting %s (team: %s) on %s", b.GetString("Login"), b.GetString("Team"), b.GetString("Server"))

	if err := b.mc.Login(); err != nil {
		return err
	}

	b.Log.Info("Connection succeeded")
	b.TeamID = b.mc.GetTeamID()
	return nil
}

// replaceAction replace the message with the correct action (/me) code
func (b *Bmattermost) replaceAction(text string) (string, bool) {
	if strings.HasPrefix(text, "*") && strings.HasSuffix(text, "*") {
		return strings.Replace(text, "*", "", -1), true
	}
	return text, false
}

func (b *Bmattermost) cacheAvatar(msg *config.Message) (string, error) {
	fi := msg.Extra["file"][0].(config.FileInfo)
	/* if we have a sha we have successfully uploaded the file to the media server,
	so we can now cache the sha */
	if fi.SHA != "" {
		b.Log.Debugf("Added %s to %s in avatarMap", fi.SHA, msg.UserID)
		b.avatarMap[msg.UserID] = fi.SHA
	}
	return "", nil
}

// sendWebhook uses the configured WebhookURL to send the message
func (b *Bmattermost) sendWebhook(msg config.Message) (string, error) {
	// skip events
	if msg.Event != "" {
		return "", nil
	}

	if b.GetBool("PrefixMessagesWithNick") {
		msg.Text = msg.Username + msg.Text
	}

	if msg.Extra != nil {
		// this sends a message only if we received a config.EVENT_FILE_FAILURE_SIZE
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			rmsg := rmsg // scopelint
			iconURL := config.GetIconURL(&rmsg, b.GetString("iconurl"))
			matterMessage := matterhook.OMessage{
				IconURL:  iconURL,
				Channel:  rmsg.Channel,
				UserName: rmsg.Username,
				Text:     rmsg.Text,
				Props:    make(map[string]interface{}),
			}
			matterMessage.Props["matterbridge_"+b.uuid] = true
			if err := b.mh.Send(matterMessage); err != nil {
				b.Log.Errorf("sendWebhook failed: %s ", err)
			}
		}

		// webhook doesn't support file uploads, so we add the url manually
		if len(msg.Extra["file"]) > 0 {
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				if fi.URL != "" {
					msg.Text += " " + fi.URL
				}
			}
		}
	}

	iconURL := config.GetIconURL(&msg, b.GetString("iconurl"))
	matterMessage := matterhook.OMessage{
		IconURL:  iconURL,
		Channel:  msg.Channel,
		UserName: msg.Username,
		Text:     msg.Text,
		Props:    make(map[string]interface{}),
	}
	if msg.Avatar != "" {
		matterMessage.IconURL = msg.Avatar
	}
	matterMessage.Props["matterbridge_"+b.uuid] = true
	err := b.mh.Send(matterMessage)
	if err != nil {
		b.Log.Info(err)
		return "", err
	}
	return "", nil
}

// skipMessages returns true if this message should not be handled
//nolint:gocyclo,cyclop
func (b *Bmattermost) skipMessage(message *matterclient.Message) bool {
	// Handle join/leave
	if message.Type == "system_join_leave" ||
		message.Type == "system_join_channel" ||
		message.Type == "system_leave_channel" {
		if b.GetBool("nosendjoinpart") {
			return true
		}

		channelName := b.getChannelName(message.Post.ChannelId)
		if channelName == "" {
			channelName = message.Channel
		}

		b.Log.Debugf("Sending JOIN_LEAVE event from %s to gateway", b.Account)
		b.Remote <- config.Message{
			Username: "system",
			Text:     message.Text,
			Channel:  channelName,
			Account:  b.Account,
			Event:    config.EventJoinLeave,
		}
		return true
	}

	// Handle edited messages
	if (message.Raw.EventType() == model.WebsocketEventPostEdited) && b.GetBool("EditDisable") {
		return true
	}

	// Ignore non-post messages
	if message.Post == nil {
		b.Log.Debugf("ignoring nil message.Post: %#v", message)
		return true
	}

	// Ignore messages sent from matterbridge
	if message.Post.Props != nil {
		if _, ok := message.Post.Props["matterbridge_"+b.uuid].(bool); ok {
			b.Log.Debug("sent by matterbridge, ignoring")
			return true
		}
	}

	// Ignore messages sent from a user logged in as the bot
	if b.mc.User.Username == message.Username {
		b.Log.Debug("message from same user as bot, ignoring")
		return true
	}

	// if the message has reactions don't repost it (for now, until we can correlate reaction with message)
	if message.Post.HasReactions {
		return true
	}

	// ignore messages from other teams than ours
	if message.Raw.GetData()["team_id"].(string) != b.TeamID {
		b.Log.Debug("message from other team, ignoring")
		return true
	}

	// only handle posted, edited or deleted events
	if !(message.Raw.EventType() == "posted" || message.Raw.EventType() == model.WebsocketEventPostEdited ||
		message.Raw.EventType() == model.WebsocketEventPostDeleted) {
		return true
	}
	return false
}

func (b *Bmattermost) getVersion() string {
	proto := "https"

	if b.GetBool("notls") {
		proto = "http"
	}

	resp, err := http.Get(proto + "://" + b.GetString("server"))
	if err != nil {
		b.Log.Error("failed getting version")
		return ""
	}

	defer resp.Body.Close()

	return resp.Header.Get("X-Version-Id")
}

func (b *Bmattermost) getChannelID(name string) string {
	idcheck := strings.Split(name, "ID:")
	if len(idcheck) > 1 {
		return idcheck[1]
	}

	return b.mc.GetChannelID(name, b.TeamID)
}

func (b *Bmattermost) getChannelName(id string) string {
	b.channelsMutex.RLock()
	defer b.channelsMutex.RUnlock()

	for _, c := range b.channelInfoMap {
		if c.Name == "ID:"+id {
			// if we have ID: specified in our gateway configuration return this
			return c.Name
		}
	}

	return ""
}
