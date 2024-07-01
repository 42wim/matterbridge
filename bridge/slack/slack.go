package bslack

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/matterhook"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/xid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type Bslack struct {
	sync.RWMutex
	*bridge.Config

	mh  *matterhook.Client
	sc  *slack.Client
	rtm *slack.RTM
	smc *socketmode.Client
	si  *slack.Info

	cache        *lru.Cache
	uuid         string
	useChannelID bool

	channels *channels
	users    *users
	legacy   bool
}

const (
	sHello               = "hello"
	sChannelJoin         = "channel_join"
	sChannelLeave        = "channel_leave"
	sChannelJoined       = "channel_joined"
	sMemberJoined        = "member_joined_channel"
	sMessageChanged      = "message_changed"
	sMessageDeleted      = "message_deleted"
	sSlackAttachment     = "slack_attachment"
	sPinnedItem          = "pinned_item"
	sUnpinnedItem        = "unpinned_item"
	sChannelTopic        = "channel_topic"
	sChannelPurpose      = "channel_purpose"
	sFileComment         = "file_comment"
	sMeMessage           = "me_message"
	sUserTyping          = "user_typing"
	sLatencyReport       = "latency_report"
	sSystemUser          = "system"
	sSlackBotUser        = "slackbot"
	cfileDownloadChannel = "file_download_channel"

	tokenConfig           = "Token"
	appTokenConfig        = "AppToken"
	incomingWebhookConfig = "WebhookBindAddress"
	outgoingWebhookConfig = "WebhookURL"
	skipTLSConfig         = "SkipTLSVerify"
	useNickPrefixConfig   = "PrefixMessagesWithNick"
	editDisableConfig     = "EditDisable"
	editSuffixConfig      = "EditSuffix"
	iconURLConfig         = "iconurl"
	noSendJoinConfig      = "nosendjoinpart"
	messageLength         = 3000
)

func New(cfg *bridge.Config) bridge.Bridger {
	// Print a deprecation warning for legacy non-bot tokens (#527).
	token := cfg.GetString(tokenConfig)
	if token != "" && !strings.HasPrefix(token, "xoxb") {
		cfg.Log.Warn("Non-bot token detected. It is STRONGLY recommended to use a proper bot-token instead.")
		cfg.Log.Warn("Legacy tokens may be deprecated by Slack at short notice. See the Matterbridge GitHub wiki for a migration guide.")
		cfg.Log.Warn("See https://github.com/42wim/matterbridge/wiki/Slack-bot-setup")
		return NewLegacy(cfg)
	}
	return newBridge(cfg)
}

func newBridge(cfg *bridge.Config) *Bslack {
	newCache, err := lru.New(5000)
	if err != nil {
		cfg.Log.Fatalf("Could not create LRU cache for Slack bridge: %v", err)
	}
	b := &Bslack{
		Config: cfg,
		uuid:   xid.New().String(),
		cache:  newCache,
	}
	return b
}

func (b *Bslack) Command(cmd string) string {
	return ""
}

func (b *Bslack) Connect() error {
	b.RLock()
	defer b.RUnlock()

	if b.GetString(incomingWebhookConfig) == "" && b.GetString(outgoingWebhookConfig) == "" && b.GetString(tokenConfig) == "" {
		return errors.New("no connection method found: WebhookBindAddress, WebhookURL or Token need to be configured")
	}

	// If we have a token we use the Slack websocket-based RTM for both sending and receiving.
	if token := b.GetString(tokenConfig); token != "" {
		b.Log.Info("Connecting using token")

		appToken := b.GetString(appTokenConfig)
		b.sc = slack.New(token, slack.OptionDebug(b.GetBool("Debug")), slack.OptionAppLevelToken(appToken))

		b.channels = newChannelManager(b.Log, b.sc)
		b.users = newUserManager(b.Log, b.sc)

		b.rtm = b.sc.NewRTM()

		if appToken != "" {
			b.smc = socketmode.New(
				b.sc,
				socketmode.OptionDebug(b.GetBool("Debug")),
			)
		} else {
			go b.rtm.ManageConnection()
		}
		go b.handleSlack()
		if b.smc != nil {
			go b.smc.Run()
		}
		return nil
	}

	// In absence of a token we fall back to incoming and outgoing Webhooks.
	b.mh = matterhook.New(
		"",
		matterhook.Config{
			InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
			DisableServer:      true,
		},
	)
	if b.GetString(outgoingWebhookConfig) != "" {
		b.Log.Info("Using specified webhook for outgoing messages.")
		b.mh.Url = b.GetString(outgoingWebhookConfig)
	}
	if b.GetString(incomingWebhookConfig) != "" {
		b.Log.Info("Setting up local webhook for incoming messages.")
		b.mh.BindAddress = b.GetString(incomingWebhookConfig)
		b.mh.DisableServer = false
		go b.handleSlack()
	}
	return nil
}

func (b *Bslack) Disconnect() error {
	return b.rtm.Disconnect()
}

// JoinChannel only acts as a verification method that checks whether Matterbridge's
// Slack integration is already member of the channel. This is because Slack does not
// allow apps or bots to join channels themselves and they need to be invited
// manually by a user.
func (b *Bslack) JoinChannel(channel config.ChannelInfo) error {
	// We can only join a channel through the Slack API.
	if b.sc == nil {
		return nil
	}

	// try to join a channel when in legacy
	if b.legacy {
		_, _, _, err := b.sc.JoinConversation(channel.Name)
		if err != nil {
			switch err.Error() {
			case "name_taken", "restricted_action":
			case "default":
				return err
			}
		}
	}

	b.channels.populateChannels(false)

	channelInfo, err := b.channels.getChannel(channel.Name)
	if err != nil {
		return fmt.Errorf("could not join channel: %#v", err)
	}

	if strings.HasPrefix(channel.Name, "ID:") {
		b.useChannelID = true
		channel.Name = channelInfo.Name
	}

	// we can't join a channel unless we are using legacy tokens #651
	if !channelInfo.IsMember && !b.legacy {
		return fmt.Errorf("slack integration that matterbridge is using is not member of channel '%s', please add it manually", channelInfo.Name)
	}
	return nil
}

func (b *Bslack) Reload(cfg *bridge.Config) (string, error) {
	return "", nil
}

func (b *Bslack) Send(msg config.Message) (string, error) {
	// Too noisy to log like other events
	if msg.Event != config.EventUserTyping {
		b.Log.Debugf("=> Receiving %#v", msg)
	}

	msg.Text = helper.ClipMessage(msg.Text, messageLength, b.GetString("MessageClipped"))
	msg.Text = b.replaceCodeFence(msg.Text)

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		msg.Text = "_" + msg.Text + "_"
	}

	// Use webhook to send the message
	if b.GetString(outgoingWebhookConfig) != "" && b.GetString(tokenConfig) == "" {
		return "", b.sendWebhook(msg)
	}
	return b.sendRTM(msg)
}

// sendWebhook uses the configured WebhookURL to send the message
func (b *Bslack) sendWebhook(msg config.Message) error {
	// Skip events.
	if msg.Event != "" {
		return nil
	}

	if b.GetBool(useNickPrefixConfig) {
		msg.Text = msg.Username + msg.Text
	}

	if msg.Extra != nil {
		// This sends a message only if we received a config.EVENT_FILE_FAILURE_SIZE.
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			rmsg := rmsg // scopelint
			iconURL := config.GetIconURL(&rmsg, b.GetString(iconURLConfig))
			matterMessage := matterhook.OMessage{
				IconURL:  iconURL,
				Channel:  msg.Channel,
				UserName: rmsg.Username,
				Text:     rmsg.Text,
			}
			if err := b.mh.Send(matterMessage); err != nil {
				b.Log.Errorf("Failed to send message: %v", err)
			}
		}

		// Webhook doesn't support file uploads, so we add the URL manually.
		for _, f := range msg.Extra["file"] {
			fi, ok := f.(config.FileInfo)
			if !ok {
				b.Log.Errorf("Received a file with unexpected content: %#v", f)
				continue
			}
			if fi.URL != "" {
				msg.Text += " " + fi.URL
			}
		}
	}

	// If we have native slack_attachments add them.
	var attachs []slack.Attachment
	for _, attach := range msg.Extra[sSlackAttachment] {
		attachs = append(attachs, attach.([]slack.Attachment)...)
	}

	iconURL := config.GetIconURL(&msg, b.GetString(iconURLConfig))
	matterMessage := matterhook.OMessage{
		IconURL:     iconURL,
		Attachments: attachs,
		Channel:     msg.Channel,
		UserName:    msg.Username,
		Text:        msg.Text,
	}
	if msg.Avatar != "" {
		matterMessage.IconURL = msg.Avatar
	}
	if err := b.mh.Send(matterMessage); err != nil {
		b.Log.Errorf("Failed to send message via webhook: %#v", err)
		return err
	}
	return nil
}

func (b *Bslack) sendRTM(msg config.Message) (string, error) {
	// Handle channelmember messages.
	if handled := b.handleGetChannelMembers(&msg); handled {
		return "", nil
	}

	channelInfo, err := b.channels.getChannel(msg.Channel)
	if err != nil {
		return "", fmt.Errorf("could not send message: %v", err)
	}
	if msg.Event == config.EventUserTyping {
		if b.GetBool("ShowUserTyping") {
			b.rtm.SendMessage(b.rtm.NewTypingMessage(channelInfo.ID))
		}
		return "", nil
	}

	var handled bool

	// Handle topic/purpose updates.
	if handled, err = b.handleTopicOrPurpose(&msg, channelInfo); handled {
		return "", err
	}

	// Handle prefix hint for unthreaded messages.
	if msg.ParentNotFound() {
		msg.ParentID = ""
		msg.Text = fmt.Sprintf("[thread]: %s", msg.Text)
	}

	// Handle message deletions.
	if handled, err = b.deleteMessage(&msg, channelInfo); handled {
		return msg.ID, err
	}

	// Prepend nickname if configured.
	if b.GetBool(useNickPrefixConfig) {
		msg.Text = msg.Username + msg.Text
	}

	// Handle message edits.
	if handled, err = b.editMessage(&msg, channelInfo); handled {
		return msg.ID, err
	}

	// Upload a file if it exists.
	if len(msg.Extra) > 0 {
		extraMsgs := helper.HandleExtra(&msg, b.General)
		for i := range extraMsgs {
			rmsg := &extraMsgs[i]
			rmsg.Text = rmsg.Username + rmsg.Text
			_, err = b.postMessage(rmsg, channelInfo)
			if err != nil {
				b.Log.Error(err)
			}
		}
		// Upload files if necessary (from Slack, Telegram or Mattermost).
		return b.uploadFile(&msg, channelInfo.ID)
	}

	// Post message.
	return b.postMessage(&msg, channelInfo)
}

func (b *Bslack) updateTopicOrPurpose(msg *config.Message, channelInfo *slack.Channel) error {
	var updateFunc func(channelID string, value string) (*slack.Channel, error)

	incomingChangeType, text := b.extractTopicOrPurpose(msg.Text)
	switch incomingChangeType {
	case "topic":
		updateFunc = b.rtm.SetTopicOfConversation
	case "purpose":
		updateFunc = b.rtm.SetPurposeOfConversation
	default:
		b.Log.Errorf("Unhandled type received from extractTopicOrPurpose: %s", incomingChangeType)
		return nil
	}
	for {
		_, err := updateFunc(channelInfo.ID, text)
		if err == nil {
			return nil
		}
		if err = handleRateLimit(b.Log, err); err != nil {
			return err
		}
	}
}

// handles updating topic/purpose and determining whether to further propagate update messages.
func (b *Bslack) handleTopicOrPurpose(msg *config.Message, channelInfo *slack.Channel) (bool, error) {
	if msg.Event != config.EventTopicChange {
		return false, nil
	}

	if b.GetBool("SyncTopic") {
		return true, b.updateTopicOrPurpose(msg, channelInfo)
	}

	// Pass along to normal message handlers.
	if b.GetBool("ShowTopicChange") {
		return false, nil
	}

	// Swallow message as handled no-op.
	return true, nil
}

func (b *Bslack) deleteMessage(msg *config.Message, channelInfo *slack.Channel) (bool, error) {
	if msg.Event != config.EventMsgDelete {
		return false, nil
	}

	// Some protocols echo deletes, but with an empty ID.
	if msg.ID == "" {
		return true, nil
	}

	for {
		_, _, err := b.rtm.DeleteMessage(channelInfo.ID, msg.ID)
		if err == nil {
			return true, nil
		}

		if err = handleRateLimit(b.Log, err); err != nil {
			b.Log.Errorf("Failed to delete user message from Slack: %#v", err)
			return true, err
		}
	}
}

func (b *Bslack) editMessage(msg *config.Message, channelInfo *slack.Channel) (bool, error) {
	if msg.ID == "" {
		return false, nil
	}
	messageOptions := b.prepareMessageOptions(msg)
	for {
		_, _, _, err := b.rtm.UpdateMessage(channelInfo.ID, msg.ID, messageOptions...)
		if err == nil {
			return true, nil
		}

		if err = handleRateLimit(b.Log, err); err != nil {
			b.Log.Errorf("Failed to edit user message on Slack: %#v", err)
			return true, err
		}
	}
}

func (b *Bslack) postMessage(msg *config.Message, channelInfo *slack.Channel) (string, error) {
	// don't post empty messages
	if msg.Text == "" {
		return "", nil
	}
	messageOptions := b.prepareMessageOptions(msg)
	for {
		_, id, err := b.rtm.PostMessage(channelInfo.ID, messageOptions...)
		if err == nil {
			return id, nil
		}

		if err = handleRateLimit(b.Log, err); err != nil {
			b.Log.Errorf("Failed to sent user message to Slack: %#v", err)
			return "", err
		}
	}
}

// uploadFile handles native upload of files
func (b *Bslack) uploadFile(msg *config.Message, channelID string) (string, error) {
	var messageID string
	for _, f := range msg.Extra["file"] {
		fi, ok := f.(config.FileInfo)
		if !ok {
			b.Log.Errorf("Received a file with unexpected content: %#v", f)
			continue
		}
		if msg.Text == fi.Comment {
			msg.Text = ""
		}
		// Because the result of the UploadFile is slower than the MessageEvent from slack
		// we can't match on the file ID yet, so we have to match on the filename too.
		ts := time.Now()
		fSize := int(fi.Size)
		if fSize == 0 {
			fSize = len(*fi.Data)
		}
		b.Log.Debugf("Adding file %s to cache at %s with timestamp, size %d", fi.Name, ts.String(), fSize)
		b.cache.Add("filename"+fi.Name, ts)
		initialComment := fmt.Sprintf("File from %s", msg.Username)
		if fi.Comment != "" {
			initialComment += fmt.Sprintf(" with comment: %s", fi.Comment)
		}

		if b.smc != nil {
			res, err := b.sc.UploadFileV2(slack.UploadFileV2Parameters{
				Reader:          bytes.NewReader(*fi.Data),
				Filename:        fi.Name,
				FileSize:        fSize,
				Channel:         channelID,
				InitialComment:  initialComment,
				ThreadTimestamp: msg.ParentID,
			})
			if err != nil {
				b.Log.Errorf("uploadfile %#v", err)
				return "", err
			}
			if res.ID != "" {
				b.Log.Debugf("Adding file ID %s to cache with timestamp %s", res.ID, ts.String())
				b.cache.Add("file"+res.ID, ts)
				messageID = res.ID // TODO
			}
		} else { // Deprecated version
			res, err := b.sc.UploadFile(slack.FileUploadParameters{
				Reader:          bytes.NewReader(*fi.Data),
				Filename:        fi.Name,
				Channels:        []string{channelID},
				InitialComment:  initialComment,
				ThreadTimestamp: msg.ParentID,
			})
			if err != nil {
				b.Log.Errorf("uploadfile %#v", err)
				return "", err
			}
			if res.ID != "" {
				b.Log.Debugf("Adding file ID %s to cache with timestamp %s", res.ID, ts.String())
				b.cache.Add("file"+res.ID, ts)
				// search for message id by uploaded file in private/public channels, get thread timestamp from uploaded file
				if v, ok := res.Shares.Private[channelID]; ok && len(v) > 0 {
					messageID = v[0].Ts
				}
				if v, ok := res.Shares.Public[channelID]; ok && len(v) > 0 {
					messageID = v[0].Ts
				}
			}
		}

	}
	return messageID, nil
}

func (b *Bslack) prepareMessageOptions(msg *config.Message) []slack.MsgOption {
	params := slack.NewPostMessageParameters()
	if b.GetBool(useNickPrefixConfig) {
		params.AsUser = true
	}
	params.Username = msg.Username
	params.LinkNames = 1 // replace mentions
	params.IconURL = config.GetIconURL(msg, b.GetString(iconURLConfig))
	params.ThreadTimestamp = msg.ParentID
	if msg.Avatar != "" {
		params.IconURL = msg.Avatar
	}

	var attachments []slack.Attachment
	// add file attachments
	attachments = append(attachments, b.createAttach(msg.Extra)...)
	// add slack attachments (from another slack bridge)
	if msg.Extra != nil {
		for _, attach := range msg.Extra[sSlackAttachment] {
			attachments = append(attachments, attach.([]slack.Attachment)...)
		}
	}

	var opts []slack.MsgOption
	opts = append(opts,
		// provide regular text field (fallback used in Slack notifications, etc.)
		slack.MsgOptionText(msg.Text, false),

		// add a callback ID so we can see we created it
		slack.MsgOptionBlocks(slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, msg.Text, false, false),
			nil, nil,
			slack.SectionBlockOptionBlockID("matterbridge_"+b.uuid),
		)),

		slack.MsgOptionEnableLinkUnfurl(),
	)
	opts = append(opts, slack.MsgOptionAttachments(attachments...))
	opts = append(opts, slack.MsgOptionPostMessageParameters(params))
	return opts
}

func (b *Bslack) createAttach(extra map[string][]interface{}) []slack.Attachment {
	var attachements []slack.Attachment
	for _, v := range extra["attachments"] {
		entry := v.(map[string]interface{})
		s := slack.Attachment{
			Fallback:   extractStringField(entry, "fallback"),
			Color:      extractStringField(entry, "color"),
			Pretext:    extractStringField(entry, "pretext"),
			AuthorName: extractStringField(entry, "author_name"),
			AuthorLink: extractStringField(entry, "author_link"),
			AuthorIcon: extractStringField(entry, "author_icon"),
			Title:      extractStringField(entry, "title"),
			TitleLink:  extractStringField(entry, "title_link"),
			Text:       extractStringField(entry, "text"),
			ImageURL:   extractStringField(entry, "image_url"),
			ThumbURL:   extractStringField(entry, "thumb_url"),
			Footer:     extractStringField(entry, "footer"),
			FooterIcon: extractStringField(entry, "footer_icon"),
		}
		attachements = append(attachements, s)
	}
	return attachements
}

func extractStringField(data map[string]interface{}, field string) string {
	if rawValue, found := data[field]; found {
		if value, ok := rawValue.(string); ok {
			return value
		}
	}
	return ""
}
