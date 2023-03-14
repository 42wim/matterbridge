package btelegram

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	tgbotapi "github.com/matterbridge/telegram-bot-api/v6"
)

const (
	unknownUser = "unknown"
	HTMLFormat  = "HTML"
	HTMLNick    = "htmlnick"
	MarkdownV2  = "MarkdownV2"
)

type Btelegram struct {
	c *tgbotapi.BotAPI
	*bridge.Config
	avatarMap map[string]string // keep cache of userid and avatar sha
}

func New(cfg *bridge.Config) bridge.Bridger {
	tgsConvertFormat := cfg.GetString("MediaConvertTgs")
	if tgsConvertFormat != "" {
		err := helper.CanConvertTgsToX()
		if err != nil {
			log.Fatalf("Telegram bridge configured to convert .tgs files to '%s', but %s does not appear to work:\n%#v", tgsConvertFormat, helper.LottieBackend(), err)
		}
		if !helper.SupportsFormat(tgsConvertFormat) {
			log.Fatalf("Telegram bridge configured to convert .tgs files to '%s', but %s doesn't support it.", tgsConvertFormat, helper.LottieBackend())
		}
	}
	return &Btelegram{Config: cfg, avatarMap: make(map[string]string)}
}

func (b *Btelegram) Connect() error {
	var err error
	b.Log.Info("Connecting")
	b.c, err = tgbotapi.NewBotAPI(b.GetString("Token"))
	if err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.c.GetUpdatesChan(u)
	b.Log.Info("Connection succeeded")
	go b.handleRecv(updates)
	return nil
}

func (b *Btelegram) Disconnect() error {
	return nil
}

func (b *Btelegram) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func TGGetParseMode(b *Btelegram, username string, text string) (textout string, parsemode string) {
	textout = username + text
	if b.GetString("MessageFormat") == HTMLFormat {
		b.Log.Debug("Using mode HTML")
		parsemode = tgbotapi.ModeHTML
	}
	if b.GetString("MessageFormat") == "Markdown" {
		b.Log.Debug("Using mode markdown")
		parsemode = tgbotapi.ModeMarkdown
	}
	if b.GetString("MessageFormat") == MarkdownV2 {
		b.Log.Debug("Using mode MarkdownV2")
		parsemode = MarkdownV2
	}
	if strings.ToLower(b.GetString("MessageFormat")) == HTMLNick {
		b.Log.Debug("Using mode HTML - nick only")
		textout = username + html.EscapeString(text)
		parsemode = tgbotapi.ModeHTML
	}
	return textout, parsemode
}

func (b *Btelegram) getIds(channel string) (int64, int, error) {
	var chatid int64
	topicid := 0

	// get the chatid
	if strings.Contains(channel, "/") { //nolint:nestif
		s := strings.Split(channel, "/")
		if len(s) < 2 {
			b.Log.Errorf("Invalid channel format: %#v\n", channel)
			return 0, 0, nil
		}
		id, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		chatid = id
		tid, err := strconv.Atoi(s[1])
		if err != nil {
			return 0, 0, err
		}
		topicid = tid
	} else {
		id, err := strconv.ParseInt(channel, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		chatid = id
	}
	return chatid, topicid, nil
}

func (b *Btelegram) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	chatid, topicid, err := b.getIds(msg.Channel)
	if err != nil {
		return "", err
	}

	// map the file SHA to our user (caches the avatar)
	if msg.Event == config.EventAvatarDownload {
		return b.cacheAvatar(&msg)
	}

	if b.GetString("MessageFormat") == HTMLFormat {
		msg.Text = makeHTML(html.EscapeString(msg.Text))
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		return b.handleDelete(&msg, chatid)
	}

	// Handle prefix hint for unthreaded messages.
	if msg.ParentNotFound() {
		msg.ParentID = ""
		msg.Text = fmt.Sprintf("[reply]: %s", msg.Text)
	}

	var parentID int
	if msg.ParentID != "" {
		parentID, _ = b.intParentID(msg.ParentID)
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			if _, msgErr := b.sendMessage(chatid, topicid, rmsg.Username, rmsg.Text, parentID); msgErr != nil {
				b.Log.Errorf("sendMessage failed: %s", msgErr)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg, chatid, topicid, parentID)
		}
	}

	// edit the message if we have a msg ID
	if msg.ID != "" {
		return b.handleEdit(&msg, chatid)
	}

	// Post normal message
	// TODO: recheck it.
	// Ignore empty text field needs for prevent double messages from whatsapp to telegram
	// when sending media with text caption
	if msg.Text != "" {
		return b.sendMessage(chatid, topicid, msg.Username, msg.Text, parentID)
	}

	return "", nil
}

func (b *Btelegram) getFileDirectURL(id string) string {
	res, err := b.c.GetFileDirectURL(id)
	if err != nil {
		return ""
	}
	return res
}

func (b *Btelegram) sendMessage(chatid int64, topicid int, username, text string, parentID int) (string, error) {
	m := tgbotapi.NewMessage(chatid, "")
	m.Text, m.ParseMode = TGGetParseMode(b, username, text)
	if topicid != 0 {
		m.BaseChat.MessageThreadID = topicid
	}
	m.ReplyToMessageID = parentID
	m.DisableWebPagePreview = b.GetBool("DisableWebPagePreview")

	res, err := b.c.Send(m)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(res.MessageID), nil
}

// sendMediaFiles native upload media files via media group
func (b *Btelegram) sendMediaFiles(msg *config.Message, chatid int64, threadid int, parentID int, media []interface{}) (string, error) {
	if len(media) == 0 {
		return "", nil
	}
	mg := tgbotapi.MediaGroupConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           chatid,
			MessageThreadID:  threadid,
			ChannelUsername:  msg.Username,
			ReplyToMessageID: parentID,
		},
		Media: media,
	}
	messages, err := b.c.SendMediaGroup(mg)
	if err != nil {
		return "", err
	}
	// return first message id
	return strconv.Itoa(messages[0].MessageID), nil
}

// intParentID return integer parent id for telegram message
func (b *Btelegram) intParentID(parentID string) (int, error) {
	pid, err := strconv.Atoi(parentID)
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func (b *Btelegram) cacheAvatar(msg *config.Message) (string, error) {
	fi := msg.Extra["file"][0].(config.FileInfo)
	/* if we have a sha we have successfully uploaded the file to the media server,
	so we can now cache the sha */
	if fi.SHA != "" {
		b.Log.Debugf("Added %s to %s in avatarMap", fi.SHA, msg.UserID)
		b.avatarMap[msg.UserID] = fi.SHA
	}
	return "", nil
}
