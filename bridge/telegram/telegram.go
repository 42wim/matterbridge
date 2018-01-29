package btelegram

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	log "github.com/Sirupsen/logrus"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Btelegram struct {
	c *tgbotapi.BotAPI
	*config.BridgeConfig
}

var flog *log.Entry
var protocol = "telegram"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg *config.BridgeConfig) *Btelegram {
	return &Btelegram{BridgeConfig: cfg}
}

func (b *Btelegram) Connect() error {
	var err error
	flog.Info("Connecting")
	b.c, err = tgbotapi.NewBotAPI(b.Config.Token)
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.c.GetUpdatesChan(u)
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	flog.Info("Connection succeeded")
	go b.handleRecv(updates)
	return nil
}

func (b *Btelegram) Disconnect() error {
	return nil

}

func (b *Btelegram) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Btelegram) Send(msg config.Message) (string, error) {
	flog.Debugf("Receiving %#v", msg)
	chatid, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return "", err
	}

	if b.Config.MessageFormat == "HTML" {
		msg.Text = makeHTML(msg.Text)
	}

	if msg.Event == config.EVENT_MSG_DELETE {
		if msg.ID == "" {
			return "", nil
		}
		msgid, err := strconv.Atoi(msg.ID)
		if err != nil {
			return "", err
		}
		_, err = b.c.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: chatid, MessageID: msgid})
		return "", err
	}

	// edit the message if we have a msg ID
	if msg.ID != "" {
		msgid, err := strconv.Atoi(msg.ID)
		if err != nil {
			return "", err
		}
		m := tgbotapi.NewEditMessageText(chatid, msgid, msg.Username+msg.Text)
		if b.Config.MessageFormat == "HTML" {
			m.ParseMode = tgbotapi.ModeHTML
		}
		_, err = b.c.Send(m)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	if msg.Extra != nil {
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			var c tgbotapi.Chattable
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				file := tgbotapi.FileBytes{fi.Name, *fi.Data}
				re := regexp.MustCompile(".(jpg|png)$")
				if re.MatchString(fi.Name) {
					c = tgbotapi.NewPhotoUpload(chatid, file)
				} else {
					c = tgbotapi.NewDocumentUpload(chatid, file)
				}
				_, err := b.c.Send(c)
				if err != nil {
					log.Errorf("file upload failed: %#v", err)
				}
				if fi.Comment != "" {
					b.sendMessage(chatid, msg.Username+fi.Comment)
				}
			}
			return "", nil
		}
	}
	return b.sendMessage(chatid, msg.Username+msg.Text)
}

func (b *Btelegram) handleRecv(updates <-chan tgbotapi.Update) {
	for update := range updates {
		flog.Debugf("Receiving from telegram: %#v", update.Message)
		var message *tgbotapi.Message
		username := ""
		channel := ""
		text := ""

		fmsg := config.Message{Extra: make(map[string][]interface{})}

		// handle channels
		if update.ChannelPost != nil {
			message = update.ChannelPost
		}
		if update.EditedChannelPost != nil && !b.Config.EditDisable {
			message = update.EditedChannelPost
			message.Text = message.Text + b.Config.EditSuffix
		}
		// handle groups
		if update.Message != nil {
			message = update.Message
		}
		if update.EditedMessage != nil && !b.Config.EditDisable {
			message = update.EditedMessage
			message.Text = message.Text + b.Config.EditSuffix
		}
		if message.From != nil {
			if b.Config.UseFirstName {
				username = message.From.FirstName
			}
			if username == "" {
				username = message.From.UserName
				if username == "" {
					username = message.From.FirstName
				}
			}
			text = message.Text
			channel = strconv.FormatInt(message.Chat.ID, 10)
		}

		if username == "" {
			username = "unknown"
		}
		if message.Sticker != nil {
			b.handleDownload(message.Sticker, &fmsg)
		}
		if message.Video != nil {
			b.handleDownload(message.Video, &fmsg)
		}
		if message.Photo != nil {
			b.handleDownload(message.Photo, &fmsg)
		}
		if message.Document != nil {
			b.handleDownload(message.Document, &fmsg)
		}
		if message.Voice != nil {
			b.handleDownload(message.Voice, &fmsg)
		}
		if message.Audio != nil {
			b.handleDownload(message.Audio, &fmsg)
		}

		if message.ForwardFrom != nil {
			usernameForward := ""
			if b.Config.UseFirstName {
				usernameForward = message.ForwardFrom.FirstName
			}
			if usernameForward == "" {
				usernameForward = message.ForwardFrom.UserName
				if usernameForward == "" {
					usernameForward = message.ForwardFrom.FirstName
				}
			}
			if usernameForward == "" {
				usernameForward = "unknown"
			}
			text = "Forwarded from " + usernameForward + ": " + text
		}

		// quote the previous message
		if message.ReplyToMessage != nil {
			usernameReply := ""
			if message.ReplyToMessage.From != nil {
				if b.Config.UseFirstName {
					usernameReply = message.ReplyToMessage.From.FirstName
				}
				if usernameReply == "" {
					usernameReply = message.ReplyToMessage.From.UserName
					if usernameReply == "" {
						usernameReply = message.ReplyToMessage.From.FirstName
					}
				}
			}
			if usernameReply == "" {
				usernameReply = "unknown"
			}
			text = text + " (re @" + usernameReply + ":" + message.ReplyToMessage.Text + ")"
		}

		if text != "" || len(fmsg.Extra) > 0 {
			flog.Debugf("Sending message from %s on %s to gateway", username, b.Account)
			msg := config.Message{Username: username, Text: text, Channel: channel, Account: b.Account, UserID: strconv.Itoa(message.From.ID), ID: strconv.Itoa(message.MessageID), Extra: fmsg.Extra}
			flog.Debugf("Message is %#v", msg)
			b.Remote <- msg
		}
	}
}

func (b *Btelegram) getFileDirectURL(id string) string {
	res, err := b.c.GetFileDirectURL(id)
	if err != nil {
		return ""
	}
	return res
}

func (b *Btelegram) handleDownload(file interface{}, msg *config.Message) {
	size := 0
	url := ""
	name := ""
	text := ""
	fileid := ""
	switch v := file.(type) {
	case *tgbotapi.Audio:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
		fileid = v.FileID
	case *tgbotapi.Voice:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
		if !strings.HasSuffix(name, ".ogg") {
			name = name + ".ogg"
		}
		fileid = v.FileID
	case *tgbotapi.Sticker:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		if !strings.HasSuffix(name, ".webp") {
			name = name + ".webp"
		}
		text = " " + url
		fileid = v.FileID
	case *tgbotapi.Video:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
		fileid = v.FileID
	case *[]tgbotapi.PhotoSize:
		photos := *v
		size = photos[len(photos)-1].FileSize
		url = b.getFileDirectURL(photos[len(photos)-1].FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
	case *tgbotapi.Document:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		name = v.FileName
		text = " " + v.FileName + " : " + url
		fileid = v.FileID
	}
	if b.Config.UseInsecureURL {
		msg.Text = text
		return
	}
	// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
	// limit to 1MB for now
	flog.Debugf("trying to download %#v fileid %#v with size %#v", name, fileid, size)
	if size <= b.General.MediaDownloadSize {
		data, err := helper.DownloadFile(url)
		if err != nil {
			flog.Errorf("download %s failed %#v", url, err)
		} else {
			flog.Debugf("download OK %#v %#v %#v", name, len(*data), len(url))
			msg.Extra["file"] = append(msg.Extra["file"], config.FileInfo{Name: name, Data: data})
		}
	}
}

func (b *Btelegram) sendMessage(chatid int64, text string) (string, error) {
	m := tgbotapi.NewMessage(chatid, text)
	if b.Config.MessageFormat == "HTML" {
		m.ParseMode = tgbotapi.ModeHTML
	}
	res, err := b.c.Send(m)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(res.MessageID), nil
}
