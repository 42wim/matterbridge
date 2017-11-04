package btelegram

import (
	"regexp"
	"strconv"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	log "github.com/Sirupsen/logrus"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Btelegram struct {
	c       *tgbotapi.BotAPI
	Config  *config.Protocol
	Remote  chan config.Message
	Account string
}

var flog *log.Entry
var protocol = "telegram"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Btelegram {
	b := &Btelegram{}
	b.Config = &cfg
	b.Remote = c
	b.Account = account
	return b
}

func (b *Btelegram) Connect() error {
	var err error
	flog.Info("Connecting")
	b.c, err = tgbotapi.NewBotAPI(b.Config.Token)
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	updates, err := b.c.GetUpdatesChan(tgbotapi.NewUpdate(0))
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
					log.Errorf("file upload failed: %#v")
				}
			}
		}
	}

	m := tgbotapi.NewMessage(chatid, msg.Username+msg.Text)
	if b.Config.MessageFormat == "HTML" {
		m.ParseMode = tgbotapi.ModeHTML
	}
	res, err := b.c.Send(m)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(res.MessageID), nil

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
		if message.Photo != nil && b.Config.UseInsecureURL {
			b.handleDownload(message.Photo, &fmsg)
		}
		if message.Document != nil && b.Config.UseInsecureURL {
			b.handleDownload(message.Sticker, &fmsg)
			text = text + " " + message.Document.FileName + " : " + b.getFileDirectURL(message.Document.FileID)
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
			msg := config.Message{Username: username, Text: text, Channel: channel, Account: b.Account, UserID: strconv.Itoa(message.From.ID), ID: strconv.Itoa(message.MessageID)}
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
	switch v := file.(type) {
	case *tgbotapi.Sticker:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		name = "sticker"
		text = " " + url
	case *tgbotapi.Video:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		name = "video"
		text = " " + url
	case *[]tgbotapi.PhotoSize:
		photos := *v
		size = photos[len(photos)-1].FileSize
		url = b.getFileDirectURL(photos[len(photos)-1].FileID)
		name = "photo"
		text = " " + url
	case *tgbotapi.Document:
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		name = v.FileName
		text = " " + v.FileName + " : " + url
	}
	if b.Config.UseInsecureURL {
		msg.Text = text
		return
	}
	// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
	// limit to 1MB for now
	if size <= 1000000 {
		data, err := helper.DownloadFile(url)
		if err != nil {
			flog.Errorf("download %s failed %#v", url, err)
		} else {
			msg.Extra["file"] = append(msg.Extra["file"], config.FileInfo{Name: name, Data: data})
		}
	}
}
