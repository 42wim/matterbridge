package btelegram

import (
	"strconv"

	"github.com/42wim/matterbridge/bridge/config"
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
		if message.Sticker != nil && b.Config.UseInsecureURL {
			text = text + " " + b.getFileDirectURL(message.Sticker.FileID)
		}
		if message.Video != nil && b.Config.UseInsecureURL {
			text = text + " " + b.getFileDirectURL(message.Video.FileID)
		}
		if message.Photo != nil && b.Config.UseInsecureURL {
			photos := *message.Photo
			// last photo is the biggest
			text = text + " " + b.getFileDirectURL(photos[len(photos)-1].FileID)
		}
		if message.Document != nil && b.Config.UseInsecureURL {
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

		if text != "" {
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
