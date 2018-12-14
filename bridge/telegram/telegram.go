package btelegram

import (
	"html"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	unknownUser = "unknown"
	HTMLFormat  = "HTML"
	HTMLNick    = "htmlnick"
)

type Btelegram struct {
	c *tgbotapi.BotAPI
	*bridge.Config
	avatarMap map[string]string // keep cache of userid and avatar sha
}

func New(cfg *bridge.Config) bridge.Bridger {
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
			if _, err := b.sendMessage(chatid, rmsg.Username, rmsg.Text); err != nil {
				b.Log.Errorf("sendMessage failed: %s", err)
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
	return b.sendMessage(chatid, msg.Username, msg.Text)
}

<<<<<<< HEAD
func (b *Btelegram) handleRecv(updates <-chan tgbotapi.Update) {
	for update := range updates {
		b.Log.Debugf("== Receiving event: %#v", update.Message)

		if update.Message == nil && update.ChannelPost == nil && update.EditedMessage == nil && update.EditedChannelPost == nil {
			b.Log.Error("Getting nil messages, this shouldn't happen.")
			continue
		}

		var message *tgbotapi.Message

		rmsg := config.Message{Account: b.Account, Extra: make(map[string][]interface{})}

		// handle channels
		if update.ChannelPost != nil {
			message = update.ChannelPost
			rmsg.Text = message.Text
		}

		// edited channel message
		if update.EditedChannelPost != nil && !b.GetBool("EditDisable") {
			message = update.EditedChannelPost
			rmsg.Text = rmsg.Text + message.Text + b.GetString("EditSuffix")
		}

		// handle groups
		if update.Message != nil {
			message = update.Message
			rmsg.Text = message.Text
		}

		// edited group message
		if update.EditedMessage != nil && !b.GetBool("EditDisable") {
			message = update.EditedMessage
			rmsg.Text = rmsg.Text + message.Text + b.GetString("EditSuffix")
		}

		// set the ID's from the channel or group message
		rmsg.ID = strconv.Itoa(message.MessageID)
		rmsg.Channel = strconv.FormatInt(message.Chat.ID, 10)

		// handle username
		if message.From != nil {
			rmsg.UserID = strconv.Itoa(message.From.ID)
			if b.GetBool("UseFirstName") {
				rmsg.Username = message.From.FirstName
			}
			if rmsg.Username == "" {
				rmsg.Username = message.From.UserName
				if rmsg.Username == "" {
					rmsg.Username = message.From.FirstName
				}
			}
			// only download avatars if we have a place to upload them (configured mediaserver)
			if b.General.MediaServerUpload != "" {
				b.handleDownloadAvatar(message.From.ID, rmsg.Channel)
			}
		}

		// if we really didn't find a username, set it to unknown
		if rmsg.Username == "" {
			rmsg.Username = "unknown"
		}

		// handle any downloads
		err := b.handleDownload(message, &rmsg)
		if err != nil {
			b.Log.Errorf("download failed: %s", err)
		}

		// handle forwarded messages
		if message.ForwardFrom != nil {
			usernameForward := ""
			if b.GetBool("UseFirstName") {
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
			rmsg.Text = "Forwarded from " + usernameForward + ": " + rmsg.Text
		}

		// quote the previous message
		if message.ReplyToMessage != nil {
			usernameReply := ""
			if message.ReplyToMessage.From != nil {
				if b.GetBool("UseFirstName") {
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
			if !b.GetBool("QuoteDisable") {
				if message.ReplyToMessage.Text != "" {
					rmsg.Text = b.handleQuote(rmsg.Text, usernameReply, message.ReplyToMessage.Text)
				} else if message.ReplyToMessage.Caption != "" {
					rmsg.Text = b.handleQuote(rmsg.Text, usernameReply, message.ReplyToMessage.Caption)
				} else {
					rmsg.Text = b.handleQuote(rmsg.Text, usernameReply, "matterbridge: Unsupported message content")
				}
			}
		}

		if rmsg.Text != "" || len(rmsg.Extra) > 0 {
			rmsg.Text = helper.RemoveEmptyNewLines(rmsg.Text)
			// channels don't have (always?) user information. see #410
			if message.From != nil {
				rmsg.Avatar = helper.GetAvatar(b.avatarMap, strconv.Itoa(message.From.ID), b.General)
			}

			b.Log.Debugf("<= Sending message from %s on %s to gateway", rmsg.Username, b.Account)
			b.Log.Debugf("<= Message is %#v", rmsg)
			b.Remote <- rmsg
		}
	}
}

=======
>>>>>>> upstream/master
func (b *Btelegram) getFileDirectURL(id string) string {
	res, err := b.c.GetFileDirectURL(id)
	if err != nil {
		return ""
	}
	return res
}

func (b *Btelegram) sendMessage(chatid int64, username, text string) (string, error) {
	m := tgbotapi.NewMessage(chatid, "")
	m.Text = username + text
	if b.GetString("MessageFormat") == HTMLFormat {
		b.Log.Debug("Using mode HTML")
		m.ParseMode = tgbotapi.ModeHTML
	}
	if b.GetString("MessageFormat") == "Markdown" {
		b.Log.Debug("Using mode markdown")
		m.ParseMode = tgbotapi.ModeMarkdown
	}
	if strings.ToLower(b.GetString("MessageFormat")) == HTMLNick {
		b.Log.Debug("Using mode HTML - nick only")
		m.Text = username + html.EscapeString(text)
		m.ParseMode = tgbotapi.ModeHTML
	}
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
