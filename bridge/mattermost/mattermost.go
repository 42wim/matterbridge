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
	"github.com/mattermost/platform/model"
	"github.com/rs/xid"
)

type Bmattermost struct {
	mh     *matterhook.Client
	mc     *matterclient.MMClient
	uuid   string
	TeamID string
	*bridge.Config
	avatarMap map[string]string
}

const mattermostPlugin = "mattermost.plugin"

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmattermost{Config: cfg, avatarMap: make(map[string]string)}
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
	if b.GetString("WebhookBindAddress") != "" {
		switch {
		case b.GetString("WebhookURL") != "":
			b.Log.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.GetString("WebhookURL"),
				matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
					BindAddress: b.GetString("WebhookBindAddress")})
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
				matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
					BindAddress: b.GetString("WebhookBindAddress")})
		}
		go b.handleMatter()
		return nil
	}
	switch {
	case b.GetString("WebhookURL") != "":
		b.Log.Info("Connecting using webhookurl (sending)")
		b.mh = matterhook.New(b.GetString("WebhookURL"),
			matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
				DisableServer: true})
		if b.GetString("Token") != "" {
			b.Log.Info("Connecting using token (receiving)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
			go b.handleMatter()
		} else if b.GetString("Login") != "" {
			b.Log.Info("Connecting using login/password (receiving)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
			go b.handleMatter()
		}
		return nil
	case b.GetString("Token") != "":
		b.Log.Info("Connecting using token (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleMatter()
	case b.GetString("Login") != "":
		b.Log.Info("Connecting using login/password (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleMatter()
	}
	if b.GetString("WebhookBindAddress") == "" && b.GetString("WebhookURL") == "" && b.GetString("Login") == "" && b.GetString("Token") == "" {
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
		id := b.mc.GetChannelId(channel.Name, b.TeamID)
		if id == "" {
			return fmt.Errorf("Could not find channel ID for channel %s", channel.Name)
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
		return msg.ID, b.mc.DeleteMessage(msg.ID)
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.mc.PostMessage(b.mc.GetChannelId(rmsg.Channel, b.TeamID), rmsg.Username+rmsg.Text)
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
		return b.mc.EditMessage(msg.ID, msg.Text)
	}

	// Post normal message
	return b.mc.PostMessage(b.mc.GetChannelId(msg.Channel, b.TeamID), msg.Text)
}

func (b *Bmattermost) handleMatter() {
	messages := make(chan *config.Message)
	if b.GetString("WebhookBindAddress") != "" {
		b.Log.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(messages)
	} else {
		if b.GetString("Token") != "" {
			b.Log.Debugf("Choosing token based receiving")
		} else {
			b.Log.Debugf("Choosing login/password based receiving")
		}
		go b.handleMatterClient(messages)
	}
	var ok bool
	for message := range messages {
		message.Avatar = helper.GetAvatar(b.avatarMap, message.UserID, b.General)
		message.Account = b.Account
		message.Text, ok = b.replaceAction(message.Text)
		if ok {
			message.Event = config.EventUserAction
		}
		b.Log.Debugf("<= Sending message from %s on %s to gateway", message.Username, b.Account)
		b.Log.Debugf("<= Message is %#v", message)
		b.Remote <- *message
	}
}

func (b *Bmattermost) handleMatterClient(messages chan *config.Message) {
	for message := range b.mc.MessageChan {
		b.Log.Debugf("%#v", message.Raw.Data)

		if b.skipMessage(message) {
			b.Log.Debugf("Skipped message: %#v", message)
			continue
		}

		// only download avatars if we have a place to upload them (configured mediaserver)
		if b.General.MediaServerUpload != "" || b.General.MediaDownloadPath != "" {
			b.handleDownloadAvatar(message.UserID, message.Channel)
		}

		b.Log.Debugf("== Receiving event %#v", message)

		rmsg := &config.Message{Username: message.Username, UserID: message.UserID, Channel: message.Channel, Text: message.Text, ID: message.Post.Id, Extra: make(map[string][]interface{})}

		// handle mattermost post properties (override username and attachments)
		props := message.Post.Props
		if props != nil {
			if _, ok := props["override_username"].(string); ok {
				rmsg.Username = props["override_username"].(string)
			}
			if _, ok := props["attachments"].([]interface{}); ok {
				rmsg.Extra["attachments"] = props["attachments"].([]interface{})
				if rmsg.Text == "" {
					for _, attachment := range rmsg.Extra["attachments"] {
						attach := attachment.(map[string]interface{})
						if attach["text"].(string) != "" {
							rmsg.Text += attach["text"].(string)
							continue
						}
						if attach["fallback"].(string) != "" {
							rmsg.Text += attach["fallback"].(string)
						}
					}
				}
			}
		}

		// create a text for bridges that don't support native editing
		if message.Raw.Event == model.WEBSOCKET_EVENT_POST_EDITED && !b.GetBool("EditDisable") {
			rmsg.Text = message.Text + b.GetString("EditSuffix")
		}

		if message.Raw.Event == model.WEBSOCKET_EVENT_POST_DELETED {
			rmsg.Event = config.EventMsgDelete
		}

		if len(message.Post.FileIds) > 0 {
			for _, id := range message.Post.FileIds {
				err := b.handleDownloadFile(rmsg, id)
				if err != nil {
					b.Log.Errorf("download failed: %s", err)
				}
			}
		}
		// Use nickname instead of username if defined
		if nick := b.mc.GetNickName(rmsg.UserID); nick != "" {
			rmsg.Username = nick
		}

		messages <- rmsg
	}
}

func (b *Bmattermost) handleMatterHook(messages chan *config.Message) {
	for {
		message := b.mh.Receive()
		b.Log.Debugf("Receiving from matterhook %#v", message)
		messages <- &config.Message{UserID: message.UserID, Username: message.UserName, Text: message.Text, Channel: message.ChannelName}
	}
}

func (b *Bmattermost) apiLogin() error {
	password := b.GetString("Password")
	if b.GetString("Token") != "" {
		password = "token=" + b.GetString("Token")
	}

	b.mc = matterclient.New(b.GetString("Login"), password, b.GetString("Team"), b.GetString("Server"))
	if b.GetBool("debug") {
		b.mc.SetLogLevel("debug")
	}
	b.mc.SkipTLSVerify = b.GetBool("SkipTLSVerify")
	b.mc.NoTLS = b.GetBool("NoTLS")
	b.Log.Infof("Connecting %s (team: %s) on %s", b.GetString("Login"), b.GetString("Team"), b.GetString("Server"))
	err := b.mc.Login()
	if err != nil {
		return err
	}
	b.Log.Info("Connection succeeded")
	b.TeamID = b.mc.GetTeamId()
	go b.mc.WsReceiver()
	go b.mc.StatusLoop()
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

// handleDownloadAvatar downloads the avatar of userid from channel
// sends a EVENT_AVATAR_DOWNLOAD message to the gateway if successful.
// logs an error message if it fails
func (b *Bmattermost) handleDownloadAvatar(userid string, channel string) {
	rmsg := config.Message{Username: "system", Text: "avatar", Channel: channel, Account: b.Account, UserID: userid, Event: config.EventAvatarDownload, Extra: make(map[string][]interface{})}
	if _, ok := b.avatarMap[userid]; !ok {
		data, resp := b.mc.Client.GetProfileImage(userid, "")
		if resp.Error != nil {
			b.Log.Errorf("ProfileImage download failed for %#v %s", userid, resp.Error)
			return
		}
		err := helper.HandleDownloadSize(b.Log, &rmsg, userid+".png", int64(len(data)), b.General)
		if err != nil {
			b.Log.Error(err)
			return
		}
		helper.HandleDownloadData(b.Log, &rmsg, userid+".png", rmsg.Text, "", &data, b.General)
		b.Remote <- rmsg
	}
}

// handleDownloadFile handles file download
func (b *Bmattermost) handleDownloadFile(rmsg *config.Message, id string) error {
	url, _ := b.mc.Client.GetFileLink(id)
	finfo, resp := b.mc.Client.GetFileInfo(id)
	if resp.Error != nil {
		return resp.Error
	}
	err := helper.HandleDownloadSize(b.Log, rmsg, finfo.Name, finfo.Size, b.General)
	if err != nil {
		return err
	}
	data, resp := b.mc.Client.DownloadFile(id, true)
	if resp.Error != nil {
		return resp.Error
	}
	helper.HandleDownloadData(b.Log, rmsg, finfo.Name, rmsg.Text, url, &data, b.General)
	return nil
}

// handleUploadFile handles native upload of files
func (b *Bmattermost) handleUploadFile(msg *config.Message) (string, error) {
	var err error
	var res, id string
	channelID := b.mc.GetChannelId(msg.Channel, b.TeamID)
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		id, err = b.mc.UploadFile(*fi.Data, channelID, fi.Name)
		if err != nil {
			return "", err
		}
		msg.Text = fi.Comment
		if b.GetBool("PrefixMessagesWithNick") {
			msg.Text = msg.Username + msg.Text
		}
		res, err = b.mc.PostMessageWithFiles(channelID, msg.Text, []string{id})
	}
	return res, err
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
			matterMessage := matterhook.OMessage{IconURL: iconURL, Channel: rmsg.Channel, UserName: rmsg.Username, Text: rmsg.Text, Props: make(map[string]interface{})}
			matterMessage.Props["matterbridge_"+b.uuid] = true
			b.mh.Send(matterMessage)
		}

		// webhook doesn't support file uploads, so we add the url manually
		if len(msg.Extra["file"]) > 0 {
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				if fi.URL != "" {
					msg.Text += fi.URL
				}
			}
		}
	}

	iconURL := config.GetIconURL(&msg, b.GetString("iconurl"))
	matterMessage := matterhook.OMessage{IconURL: iconURL, Channel: msg.Channel, UserName: msg.Username, Text: msg.Text, Props: make(map[string]interface{})}
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
func (b *Bmattermost) skipMessage(message *matterclient.Message) bool {
	// Handle join/leave
	if message.Type == "system_join_leave" ||
		message.Type == "system_join_channel" ||
		message.Type == "system_leave_channel" {
		if b.GetBool("nosendjoinpart") {
			return true
		}
		b.Log.Debugf("Sending JOIN_LEAVE event from %s to gateway", b.Account)
		b.Remote <- config.Message{Username: "system", Text: message.Text, Channel: message.Channel, Account: b.Account, Event: config.EventJoinLeave}
		return true
	}

	// Handle edited messages
	if (message.Raw.Event == model.WEBSOCKET_EVENT_POST_EDITED) && b.GetBool("EditDisable") {
		return true
	}

	// Ignore messages sent from matterbridge
	if message.Post.Props != nil {
		if _, ok := message.Post.Props["matterbridge_"+b.uuid].(bool); ok {
			b.Log.Debugf("sent by matterbridge, ignoring")
			return true
		}
	}

	// Ignore messages sent from a user logged in as the bot
	if b.mc.User.Username == message.Username {
		return true
	}

	// if the message has reactions don't repost it (for now, until we can correlate reaction with message)
	if message.Post.HasReactions {
		return true
	}

	// ignore messages from other teams than ours
	if message.Raw.Data["team_id"].(string) != b.TeamID {
		return true
	}

	// only handle posted, edited or deleted events
	if !(message.Raw.Event == "posted" || message.Raw.Event == model.WEBSOCKET_EVENT_POST_EDITED || message.Raw.Event == model.WEBSOCKET_EVENT_POST_DELETED) {
		return true
	}
	return false
}
