package btelegram

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// handleChannels checks if it's a channel message and if the message is a new or edited messages
func (b *Btelegram) handleChannels(rmsg *config.Message, message *tgbotapi.Message, update tgbotapi.Update) *tgbotapi.Message {
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
	return message
}

// handleForwarded handles forwarded messages
func (b *Btelegram) handleForwarded(rmsg *config.Message, message *tgbotapi.Message) {
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
			usernameForward = unknownUser
		}
		rmsg.Text = "Forwarded from " + usernameForward + ": " + rmsg.Text
	}
}

// handleGroups checks if it's a group message and if the message is a new or edited messages
func (b *Btelegram) handleGroups(rmsg *config.Message, message *tgbotapi.Message, update tgbotapi.Update) *tgbotapi.Message {
	if update.Message != nil {
		message = update.Message
		rmsg.Text = message.Text
	}

	// edited group message
	if update.EditedMessage != nil && !b.GetBool("EditDisable") {
		message = update.EditedMessage
		rmsg.Text = rmsg.Text + message.Text + b.GetString("EditSuffix")
	}
	return message
}

// handleQuoting handles quoting of previous messages
func (b *Btelegram) handleQuoting(rmsg *config.Message, message *tgbotapi.Message) {
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
			usernameReply = unknownUser
		}
		if !b.GetBool("QuoteDisable") {
			rmsg.Text = b.handleQuote(rmsg.Text, usernameReply, message.ReplyToMessage.Text)
		}
	}
}

// handleUsername handles the correct setting of the username
func (b *Btelegram) handleUsername(rmsg *config.Message, message *tgbotapi.Message) {
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
		rmsg.Username = unknownUser
	}
}

func (b *Btelegram) handleRecv(updates <-chan tgbotapi.Update) {
	for update := range updates {
		b.Log.Debugf("== Receiving event: %#v", update.Message)

		if update.Message == nil && update.ChannelPost == nil &&
			update.EditedMessage == nil && update.EditedChannelPost == nil {
			b.Log.Error("Getting nil messages, this shouldn't happen.")
			continue
		}

		var message *tgbotapi.Message

		rmsg := config.Message{Account: b.Account, Extra: make(map[string][]interface{})}

		// handle channels
		b.handleChannels(&rmsg, message, update)

		// handle groups
		b.handleGroups(&rmsg, message, update)

		// set the ID's from the channel or group message
		rmsg.ID = strconv.Itoa(message.MessageID)
		rmsg.Channel = strconv.FormatInt(message.Chat.ID, 10)

		// handle username
		b.handleUsername(&rmsg, message)

		// handle any downloads
		err := b.handleDownload(&rmsg, message)
		if err != nil {
			b.Log.Errorf("download failed: %s", err)
		}

		// handle forwarded messages
		b.handleForwarded(&rmsg, message)

		// quote the previous message
		b.handleQuoting(&rmsg, message)

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

// handleDownloadAvatar downloads the avatar of userid from channel
// sends a EVENT_AVATAR_DOWNLOAD message to the gateway if successful.
// logs an error message if it fails
func (b *Btelegram) handleDownloadAvatar(userid int, channel string) {
	rmsg := config.Message{Username: "system",
		Text:    "avatar",
		Channel: channel,
		Account: b.Account,
		UserID:  strconv.Itoa(userid),
		Event:   config.EventAvatarDownload,
		Extra:   make(map[string][]interface{})}

	if _, ok := b.avatarMap[strconv.Itoa(userid)]; !ok {
		photos, err := b.c.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{UserID: userid, Limit: 1})
		if err != nil {
			b.Log.Errorf("Userprofile download failed for %#v %s", userid, err)
		}

		if len(photos.Photos) > 0 {
			photo := photos.Photos[0][0]
			url := b.getFileDirectURL(photo.FileID)
			name := strconv.Itoa(userid) + ".png"
			b.Log.Debugf("trying to download %#v fileid %#v with size %#v", name, photo.FileID, photo.FileSize)

			err := helper.HandleDownloadSize(b.Log, &rmsg, name, int64(photo.FileSize), b.General)
			if err != nil {
				b.Log.Error(err)
				return
			}
			data, err := helper.DownloadFile(url)
			if err != nil {
				b.Log.Errorf("download %s failed %#v", url, err)
				return
			}
			helper.HandleDownloadData(b.Log, &rmsg, name, rmsg.Text, "", data, b.General)
			b.Remote <- rmsg
		}
	}
}

// handleDownloadFile handles file download
func (b *Btelegram) handleDownload(rmsg *config.Message, message *tgbotapi.Message) error {
	size := 0
	var url, name, text string

	if message.Sticker != nil {
		v := message.Sticker
		size = v.FileSize
		url = b.getFileDirectURL(v.FileID)
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
		if !strings.HasSuffix(name, ".webp") {
			name += ".webp"
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
			name += ".ogg"
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
	if b.GetBool("UseInsecureURL") {
		b.Log.Debugf("Setting message text to :%s", text)
		rmsg.Text += text
		return nil
	}
	// if we have a file attached, download it (in memory) and put a pointer to it in msg.Extra
	err := helper.HandleDownloadSize(b.Log, rmsg, name, int64(size), b.General)
	if err != nil {
		return err
	}
	data, err := helper.DownloadFile(url)
	if err != nil {
		return err
	}
	helper.HandleDownloadData(b.Log, rmsg, name, message.Caption, "", data, b.General)
	return nil
}

// handleUploadFile handles native upload of files
func (b *Btelegram) handleUploadFile(msg *config.Message, chatid int64) string {
	var c tgbotapi.Chattable
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		file := tgbotapi.FileBytes{
			Name:  fi.Name,
			Bytes: *fi.Data,
		}
		re := regexp.MustCompile(".(jpg|png)$")
		if re.MatchString(fi.Name) {
			c = tgbotapi.NewPhotoUpload(chatid, file)
		} else {
			c = tgbotapi.NewDocumentUpload(chatid, file)
		}
		_, err := b.c.Send(c)
		if err != nil {
			b.Log.Errorf("file upload failed: %#v", err)
		}
		if fi.Comment != "" {
			if _, err := b.sendMessage(chatid, msg.Username, fi.Comment); err != nil {
				b.Log.Errorf("posting file comment %s failed: %s", fi.Comment, err)
			}
		}
	}
	return ""
}

func (b *Btelegram) handleQuote(message, quoteNick, quoteMessage string) string {
	format := b.GetString("quoteformat")
	if format == "" {
		format = "{MESSAGE} (re @{QUOTENICK}: {QUOTEMESSAGE})"
	}
	format = strings.Replace(format, "{MESSAGE}", message, -1)
	format = strings.Replace(format, "{QUOTENICK}", quoteNick, -1)
	format = strings.Replace(format, "{QUOTEMESSAGE}", quoteMessage, -1)
	return format
}
