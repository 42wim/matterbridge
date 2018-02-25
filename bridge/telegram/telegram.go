package btelegram

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

type Btelegram struct {
	c *tgbotapi.BotAPI
	*config.BridgeConfig
	avatarMap map[string]string // keep cache of userid and avatar sha
}

var flog *log.Entry
var protocol = "telegram"

func init() {
	flog = log.WithFields(log.Fields{"prefix": protocol})
}

func New(cfg *config.BridgeConfig) *Btelegram {
	return &Btelegram{BridgeConfig: cfg, avatarMap: make(map[string]string)}
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

	// get the chatid
	chatid, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return "", err
	}

	// map the file SHA to our user (caches the avatar)
	if msg.Event == config.EVENT_AVATAR_DOWNLOAD {
		return b.cacheAvatar(&msg)
	}

	if b.Config.MessageFormat == "HTML" {
		msg.Text = makeHTML(msg.Text)
	}

	// Delete message
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

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.sendMessage(chatid, rmsg.Username+rmsg.Text)
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			b.handleUploadFile(&msg, chatid)
		}
	}

	// edit the message if we have a msg ID
	if msg.ID != "" {
		msgid, err := strconv.Atoi(msg.ID)
		if err != nil {
			return "", err
		}
		m := tgbotapi.NewEditMessageText(chatid, msgid, msg.Username+msg.Text)
		if b.Config.MessageFormat == "HTML" {
			flog.Debug("Using mode HTML")
			m.ParseMode = tgbotapi.ModeHTML
		}
		if b.Config.MessageFormat == "Markdown" {
			flog.Debug("Using mode markdown")
			m.ParseMode = tgbotapi.ModeMarkdown
		}
		_, err = b.c.Send(m)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	// Post normal message
	return b.sendMessage(chatid, msg.Username+msg.Text)
}

func (b *Btelegram) handleRecv(updates <-chan tgbotapi.Update) {
	for update := range updates {
		flog.Debugf("Receiving from telegram: %#v", update.Message)

		if update.Message == nil && update.ChannelPost == nil {
			flog.Error("Getting nil messages, this shouldn't happen.")
			continue
		}

		var message *tgbotapi.Message

		rmsg := config.Message{Account: b.Account, Extra: make(map[string][]interface{})}

		// handle channels
		if update.ChannelPost != nil {
			message = update.ChannelPost
		}

		// edited channel message
		if update.EditedChannelPost != nil && !b.Config.EditDisable {
			message = update.EditedChannelPost
			rmsg.Text = rmsg.Text + message.Text + b.Config.EditSuffix
		}

		// handle groups
		if update.Message != nil {
			message = update.Message
		}

		// edited group message
		if update.EditedMessage != nil && !b.Config.EditDisable {
			message = update.EditedMessage
			rmsg.Text = rmsg.Text + message.Text + b.Config.EditSuffix
		}

		// set the ID's from the channel or group message
		rmsg.ID = strconv.Itoa(message.MessageID)
		rmsg.UserID = strconv.Itoa(message.From.ID)
		rmsg.Channel = strconv.FormatInt(message.Chat.ID, 10)

		// handle username
		if message.From != nil {
			if b.Config.UseFirstName {
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
			flog.Errorf("download failed: %s", err)
		}

		// handle forwarded messages
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
			rmsg.Text = "Forwarded from " + usernameForward + ": " + rmsg.Text
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
			rmsg.Text = rmsg.Text + " (re @" + usernameReply + ":" + message.ReplyToMessage.Text + ")"
		}

		if rmsg.Text != "" || len(rmsg.Extra) > 0 {
			rmsg.Avatar = helper.GetAvatar(b.avatarMap, strconv.Itoa(message.From.ID), b.General)

			flog.Debugf("Sending message from %s on %s to gateway", rmsg.Username, b.Account)
			flog.Debugf("Message is %#v", rmsg)
			b.Remote <- rmsg
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

// handleDownloadAvatar downloads the avatar of userid from channel
// sends a EVENT_AVATAR_DOWNLOAD message to the gateway if successful.
// logs an error message if it fails
func (b *Btelegram) handleDownloadAvatar(userid int, channel string) {
	rmsg := config.Message{Username: "system", Text: "avatar", Channel: channel, Account: b.Account, UserID: strconv.Itoa(userid), Event: config.EVENT_AVATAR_DOWNLOAD, Extra: make(map[string][]interface{})}
	if _, ok := b.avatarMap[strconv.Itoa(userid)]; !ok {
		photos, err := b.c.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{UserID: userid, Limit: 1})
		if err != nil {
			flog.Errorf("Userprofile download failed for %#v %s", userid, err)
		}

		if len(photos.Photos) > 0 {
			photo := photos.Photos[0][0]
			url := b.getFileDirectURL(photo.FileID)
			name := strconv.Itoa(userid) + ".png"
			flog.Debugf("trying to download %#v fileid %#v with size %#v", name, photo.FileID, photo.FileSize)

			err := helper.HandleDownloadSize(flog, &rmsg, name, int64(photo.FileSize), b.General)
			if err != nil {
				flog.Error(err)
				return
			}
			data, err := helper.DownloadFile(url)
			if err != nil {
				flog.Errorf("download %s failed %#v", url, err)
				return
			}
			helper.HandleDownloadData(flog, &rmsg, name, rmsg.Text, "", data, b.General)
			b.Remote <- rmsg
		}
	}
}

// handleDownloadFile handles file download
func (b *Btelegram) handleDownload(message *tgbotapi.Message, rmsg *config.Message) error {
	size := 0
	var url, name, text string

	if message.Sticker != nil {
		v := message.Sticker
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		if !strings.HasSuffix(name, ".webp") {
			name = name + ".webp"
		}
		text = " " + url
	}
	if message.Video != nil {
		v := message.Video
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
	}
	if message.Photo != nil {
		photos := *message.Photo
		size = photos[len(photos)-1].FileSize
		url = b.getFileDirectURL(photos[len(photos)-1].FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
	}
	if message.Document != nil {
		v := message.Document
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		name = v.FileName
		text = " " + v.FileName + " : " + url
	}
	if message.Voice != nil {
		v := message.Voice
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
		if !strings.HasSuffix(name, ".ogg") {
			name = name + ".ogg"
		}
	}
	if message.Audio != nil {
		v := message.Audio
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		text = " " + url
	}
	// if name is empty we didn't match a thing to download
	if name == "" {
		return nil
	}
	// use the URL instead of native upload
	if b.Config.UseInsecureURL {
		flog.Debugf("Setting message text to :%s", text)
		rmsg.Text = rmsg.Text + text
		return nil
	}
	// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
	err := helper.HandleDownloadSize(flog, rmsg, name, int64(size), b.General)
	if err != nil {
		return err
	}
	data, err := helper.DownloadFile(url)
	if err != nil {
		return err
	}
	helper.HandleDownloadData(flog, rmsg, name, message.Caption, "", data, b.General)
	return nil
}

// handleUploadFile handles native upload of files
func (b *Btelegram) handleUploadFile(msg *config.Message, chatid int64) (string, error) {
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

func (b *Btelegram) sendMessage(chatid int64, text string) (string, error) {
	m := tgbotapi.NewMessage(chatid, text)
	if b.Config.MessageFormat == "HTML" {
		flog.Debug("Using mode HTML")
		m.ParseMode = tgbotapi.ModeHTML
	}
	if b.Config.MessageFormat == "Markdown" {
		flog.Debug("Using mode markdown")
		m.ParseMode = tgbotapi.ModeMarkdown
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
		flog.Debugf("Added %s to %s in avatarMap", fi.SHA, msg.UserID)
		b.avatarMap[msg.UserID] = fi.SHA
	}
	return "", nil
}
