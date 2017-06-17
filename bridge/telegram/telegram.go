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

func (b *Btelegram) JoinChannel(channel string) error {
	return nil
}

func (b *Btelegram) Send(msg config.Message) error {
	flog.Debugf("Receiving %#v", msg)
	chatid, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return err
	}

	if b.Config.MessageFormat == "HTML" {
		msg.Text = makeHTML(msg.Text)
	}
	m := tgbotapi.NewMessage(chatid, msg.Username+msg.Text)
	if b.Config.MessageFormat == "HTML" {
		m.ParseMode = tgbotapi.ModeHTML
	}
	_, err = b.c.Send(m)
	return err
}

func (b *Btelegram) handleRecv(updates <-chan tgbotapi.Update) {
	for update := range updates {
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
		if message.Sticker != nil {
			text = text + " " + b.getFileDirectURL(message.Sticker.FileID)
		}
		if message.Video != nil {
			text = text + " " + b.getFileDirectURL(message.Video.FileID)
		}
		if message.Photo != nil {
			for _, p := range *message.Photo {
				text = text + " " + b.getFileDirectURL(p.FileID)
			}
		}
		if message.Document != nil {
			text = text + " " + message.Document.FileName + " : " + b.getFileDirectURL(message.Document.FileID)
		}
		if text != "" {
			flog.Debugf("Sending message from %s on %s to gateway", username, b.Account)
			b.Remote <- config.Message{Username: username, Text: text, Channel: channel, Account: b.Account}
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
