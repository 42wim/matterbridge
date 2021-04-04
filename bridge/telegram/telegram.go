package btelegram

import (
	"html"
	"log"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	unknownUser = "unknown"
	HTMLFormat  = "HTML"
	HTMLNick    = "htmlnick"
	MarkdownV2  = "MarkdownV2"
	FormatPng   = "png"
	FormatWebp  = "webp"
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
			log.Fatalf("Telegram bridge configured to convert .tgs files to '%s', but lottie does not appear to work:\n%#v", tgsConvertFormat, err)
		}
		if tgsConvertFormat != FormatPng && tgsConvertFormat != FormatWebp {
			log.Fatalf("Telegram bridge configured to convert .tgs files to '%s', but only '%s' and '%s' are supported.", FormatPng, FormatWebp, tgsConvertFormat)
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
	updates, err := b.c.GetUpdatesChan(u)
	if err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}
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

func (b *Btelegram) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// get the chatid
	chatid, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return "", err
	}

	// map the file SHA to our user (caches the avatar)
	if msg.Event == config.EventAvatarDownload {
		return b.cacheAvatar(&msg)
	}

	if b.GetString("MessageFormat") == HTMLFormat {
		msg.Text = makeHTML(msg.Text)
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		return b.handleDelete(&msg, chatid)
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			if _, msgErr := b.sendMessage(chatid, rmsg.Username, rmsg.Text); msgErr != nil {
				b.Log.Errorf("sendMessage failed: %s", msgErr)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			b.handleUploadFile(&msg, chatid)
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
		return b.sendMessage(chatid, msg.Username, msg.Text)
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

func (b *Btelegram) sendMessage(chatid int64, username, text string) (string, error) {
	m := tgbotapi.NewMessage(chatid, "")
	m.Text, m.ParseMode = TGGetParseMode(b, username, text)

	m.DisableWebPagePreview = b.GetBool("DisableWebPagePreview")

	res, err := b.c.Send(m)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(res.MessageID), nil
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
