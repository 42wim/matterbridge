package bmattermost

import (
	"errors"
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/matterclient"
	"github.com/42wim/matterbridge/matterhook"
	"strings"
)

type Bmattermost struct {
	mh     *matterhook.Client
	mc     *matterclient.MMClient
	TeamID string
	*config.BridgeConfig
	avatarMap map[string]string
}

func New(cfg *config.BridgeConfig) bridge.Bridger {
	b := &Bmattermost{BridgeConfig: cfg, avatarMap: make(map[string]string)}
	return b
}

func (b *Bmattermost) Command(cmd string) string {
	return ""
}

func (b *Bmattermost) Connect() error {
	if b.Config.WebhookBindAddress != "" {
		if b.Config.WebhookURL != "" {
			b.Log.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.Config.WebhookURL,
				matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
					BindAddress: b.Config.WebhookBindAddress})
		} else if b.Config.Token != "" {
			b.Log.Info("Connecting using token (sending)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
		} else if b.Config.Login != "" {
			b.Log.Info("Connecting using login/password (sending)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
		} else {
			b.Log.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.Config.WebhookURL,
				matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
					BindAddress: b.Config.WebhookBindAddress})
		}
		go b.handleMatter()
		return nil
	}
	if b.Config.WebhookURL != "" {
		b.Log.Info("Connecting using webhookurl (sending)")
		b.mh = matterhook.New(b.Config.WebhookURL,
			matterhook.Config{InsecureSkipVerify: b.Config.SkipTLSVerify,
				DisableServer: true})
		if b.Config.Token != "" {
			b.Log.Info("Connecting using token (receiving)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
			go b.handleMatter()
		} else if b.Config.Login != "" {
			b.Log.Info("Connecting using login/password (receiving)")
			err := b.apiLogin()
			if err != nil {
				return err
			}
			go b.handleMatter()
		}
		return nil
	} else if b.Config.Token != "" {
		b.Log.Info("Connecting using token (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleMatter()
	} else if b.Config.Login != "" {
		b.Log.Info("Connecting using login/password (sending and receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
		go b.handleMatter()
	}
	if b.Config.WebhookBindAddress == "" && b.Config.WebhookURL == "" && b.Config.Login == "" && b.Config.Token == "" {
		return errors.New("no connection method found. See that you have WebhookBindAddress, WebhookURL or Token/Login/Password/Server/Team configured")
	}
	return nil
}

func (b *Bmattermost) Disconnect() error {
	return nil
}

func (b *Bmattermost) JoinChannel(channel config.ChannelInfo) error {
	// we can only join channels using the API
	if b.Config.WebhookURL == "" && b.Config.WebhookBindAddress == "" {
		id := b.mc.GetChannelId(channel.Name, "")
		if id == "" {
			return fmt.Errorf("Could not find channel ID for channel %s", channel.Name)
		}
		return b.mc.JoinChannel(id)
	}
	return nil
}

func (b *Bmattermost) Send(msg config.Message) (string, error) {
	b.Log.Debugf("Receiving %#v", msg)

	// Make a action /me of the message
	if msg.Event == config.EVENT_USER_ACTION {
		msg.Text = "*" + msg.Text + "*"
	}

	// map the file SHA to our user (caches the avatar)
	if msg.Event == config.EVENT_AVATAR_DOWNLOAD {
		return b.cacheAvatar(&msg)
	}

	// Use webhook to send the message
	if b.Config.WebhookURL != "" {
		return b.sendWebhook(msg)
	}

	// Delete message
	if msg.Event == config.EVENT_MSG_DELETE {
		if msg.ID == "" {
			return "", nil
		}
		return msg.ID, b.mc.DeleteMessage(msg.ID)
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.mc.PostMessage(b.mc.GetChannelId(rmsg.Channel, ""), rmsg.Username+rmsg.Text)
		}
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg)
		}
	}

	// Prepend nick if configured
	if b.Config.PrefixMessagesWithNick {
		msg.Text = msg.Username + msg.Text
	}

	// Edit message if we have an ID
	if msg.ID != "" {
		return b.mc.EditMessage(msg.ID, msg.Text)
	}

	// Post normal message
	return b.mc.PostMessage(b.mc.GetChannelId(msg.Channel, ""), msg.Text)
}

func (b *Bmattermost) handleMatter() {
	messages := make(chan *config.Message)
	if b.Config.WebhookBindAddress != "" {
		b.Log.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(messages)
	} else {
		if b.Config.Token != "" {
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
			message.Event = config.EVENT_USER_ACTION
		}
		b.Log.Debugf("Sending message from %s on %s to gateway", message.Username, b.Account)
		b.Log.Debugf("Message is %#v", message)
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
		if b.General.MediaServerUpload != "" {
			b.handleDownloadAvatar(message.UserID, message.Channel)
		}

		b.Log.Debugf("Receiving from matterclient %#v", message)

		rmsg := &config.Message{Username: message.Username, UserID: message.UserID, Channel: message.Channel, Text: message.Text, ID: message.Post.Id, Extra: make(map[string][]interface{})}

		// handle mattermost post properties (override username and attachments)
		props := message.Post.Props
		if props != nil {
			if _, ok := props["override_username"].(string); ok {
				rmsg.Username = props["override_username"].(string)
			}
			if _, ok := props["attachments"].([]interface{}); ok {
				rmsg.Extra["attachments"] = props["attachments"].([]interface{})
			}
		}

		// create a text for bridges that don't support native editing
		if message.Raw.Event == "post_edited" && !b.Config.EditDisable {
			rmsg.Text = message.Text + b.Config.EditSuffix
		}

		if message.Raw.Event == "post_deleted" {
			rmsg.Event = config.EVENT_MSG_DELETE
		}

		if len(message.Post.FileIds) > 0 {
			for _, id := range message.Post.FileIds {
				err := b.handleDownloadFile(rmsg, id)
				if err != nil {
					b.Log.Errorf("download failed: %s", err)
				}
			}
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
	password := b.Config.Password
	if b.Config.Token != "" {
		password = "MMAUTHTOKEN=" + b.Config.Token
	}

	b.mc = matterclient.New(b.Config.Login, password, b.Config.Team, b.Config.Server)
	if b.General.Debug {
		b.mc.SetLogLevel("debug")
	}
	b.mc.SkipTLSVerify = b.Config.SkipTLSVerify
	b.mc.NoTLS = b.Config.NoTLS
	b.Log.Infof("Connecting %s (team: %s) on %s", b.Config.Login, b.Config.Team, b.Config.Server)
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
	rmsg := config.Message{Username: "system", Text: "avatar", Channel: channel, Account: b.Account, UserID: userid, Event: config.EVENT_AVATAR_DOWNLOAD, Extra: make(map[string][]interface{})}
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
	channelID := b.mc.GetChannelId(msg.Channel, "")
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		id, err = b.mc.UploadFile(*fi.Data, channelID, fi.Name)
		if err != nil {
			return "", err
		}
		msg.Text = fi.Comment
		if b.Config.PrefixMessagesWithNick {
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

	if b.Config.PrefixMessagesWithNick {
		msg.Text = msg.Username + msg.Text
	}
	if msg.Extra != nil {
		// this sends a message only if we received a config.EVENT_FILE_FAILURE_SIZE
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL, Channel: rmsg.Channel, UserName: rmsg.Username, Text: rmsg.Text, Props: make(map[string]interface{})}
			matterMessage.Props["matterbridge"] = true
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

	matterMessage := matterhook.OMessage{IconURL: b.Config.IconURL, Channel: msg.Channel, UserName: msg.Username, Text: msg.Text, Props: make(map[string]interface{})}
	if msg.Avatar != "" {
		matterMessage.IconURL = msg.Avatar
	}
	matterMessage.Props["matterbridge"] = true
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
		b.Log.Debugf("Sending JOIN_LEAVE event from %s to gateway", b.Account)
		b.Remote <- config.Message{Username: "system", Text: message.Text, Channel: message.Channel, Account: b.Account, Event: config.EVENT_JOIN_LEAVE}
		return true
	}

	// Handle edited messages
	if (message.Raw.Event == "post_edited") && b.Config.EditDisable {
		return true
	}

	// Ignore messages sent from matterbridge
	if message.Post.Props != nil {
		if _, ok := message.Post.Props["matterbridge"].(bool); ok {
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
	if !(message.Raw.Event == "posted" || message.Raw.Event == "post_edited" || message.Raw.Event == "post_deleted") {
		return true
	}
	return false
}
