package btelegram

import (
	"html"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Btelegram) handleUpdate(rmsg *config.Message, message, posted, edited *tgbotapi.Message) *tgbotapi.Message {
	// handle channels
	if posted != nil {
		message = posted
		rmsg.Text = message.Text
	}

	// edited channel message
	if edited != nil && !b.GetBool("EditDisable") {
		message = edited
		rmsg.Text = rmsg.Text + message.Text + b.GetString("EditSuffix")
	}
	return message
}

// handleChannels checks if it's a channel message and if the message is a new or edited messages
func (b *Btelegram) handleChannels(rmsg *config.Message, message *tgbotapi.Message, update tgbotapi.Update) *tgbotapi.Message {
	return b.handleUpdate(rmsg, message, update.ChannelPost, update.EditedChannelPost)
}

// handleGroups checks if it's a group message and if the message is a new or edited messages
func (b *Btelegram) handleGroups(rmsg *config.Message, message *tgbotapi.Message, update tgbotapi.Update) *tgbotapi.Message {
	return b.handleUpdate(rmsg, message, update.Message, update.EditedMessage)
}

// handleForwarded handles forwarded messages
func (b *Btelegram) handleForwarded(rmsg *config.Message, message *tgbotapi.Message) {
	if message.ForwardDate == 0 {
		return
	}

	if message.ForwardFrom == nil {
		rmsg.Text = "Forwarded from " + unknownUser + ": " + rmsg.Text
		return
	}

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
		if b.General.MediaServerUpload != "" || (b.General.MediaServerDownload != "" && b.General.MediaDownloadPath != "") {
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
		message = b.handleChannels(&rmsg, message, update)

		// handle groups
		message = b.handleGroups(&rmsg, message, update)

		if message == nil {
			b.Log.Error("message is nil, this shouldn't happen.")
			continue
		}

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

		// handle entities (adding URLs)
		b.handleEntities(&rmsg, message)

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
	rmsg := config.Message{
		Username: "system",
		Text:     "avatar",
		Channel:  channel,
		Account:  b.Account,
		UserID:   strconv.Itoa(userid),
		Event:    config.EventAvatarDownload,
		Extra:    make(map[string][]interface{}),
	}

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

func (b *Btelegram) maybeConvertTgs(name *string, data *[]byte) {
	var format string
	switch b.GetString("MediaConvertTgs") {
	case FormatWebp:
		b.Log.Debugf("Tgs to WebP conversion enabled, converting %v", name)
		format = FormatWebp
	case FormatPng:
		// The WebP to PNG converter can't handle animated webp files yet,
		// and I'm not going to write a path for x/image/webp.
		// The error message would be:
		//     conversion failed: webp: non-Alpha VP8X is not implemented
		// So instead, we tell lottie to directly go to PNG.
		b.Log.Debugf("Tgs to PNG conversion enabled, converting %v", name)
		format = FormatPng
	default:
		// Otherwise, no conversion was requested. Trying to run the usual webp
		// converter would fail, because '.tgs.webp' is actually a gzipped JSON
		// file, and has nothing to do with WebP.
		return
	}
	err := helper.ConvertTgsToX(data, format, b.Log)
	if err != nil {
		b.Log.Errorf("conversion failed: %v", err)
	} else {
		*name = strings.Replace(*name, "tgs.webp", format, 1)
	}
}

func (b *Btelegram) maybeConvertWebp(name *string, data *[]byte) {
	if b.GetBool("MediaConvertWebPToPNG") {
		b.Log.Debugf("WebP to PNG conversion enabled, converting %v", name)
		err := helper.ConvertWebPToPNG(data)
		if err != nil {
			b.Log.Errorf("conversion failed: %v", err)
		} else {
			*name = strings.Replace(*name, ".webp", ".png", 1)
		}
	}
}

// handleDownloadFile handles file download
func (b *Btelegram) handleDownload(rmsg *config.Message, message *tgbotapi.Message) error {
	size := 0
	var url, name, text string
	switch {
	case message.Sticker != nil:
		text, name, url = b.getDownloadInfo(message.Sticker.FileID, ".webp", true)
		size = message.Sticker.FileSize
	case message.Voice != nil:
		text, name, url = b.getDownloadInfo(message.Voice.FileID, ".ogg", true)
		size = message.Voice.FileSize
	case message.Video != nil:
		text, name, url = b.getDownloadInfo(message.Video.FileID, "", true)
		size = message.Video.FileSize
	case message.Audio != nil:
		text, name, url = b.getDownloadInfo(message.Audio.FileID, "", true)
		size = message.Audio.FileSize
	case message.Document != nil:
		_, _, url = b.getDownloadInfo(message.Document.FileID, "", false)
		size = message.Document.FileSize
		name = message.Document.FileName
		text = " " + message.Document.FileName + " : " + url
	case message.Photo != nil:
		photos := *message.Photo
		size = photos[len(photos)-1].FileSize
		text, name, url = b.getDownloadInfo(photos[len(photos)-1].FileID, "", true)
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

	if strings.HasSuffix(name, ".tgs.webp") {
		b.maybeConvertTgs(&name, data)
	} else if strings.HasSuffix(name, ".webp") {
		b.maybeConvertWebp(&name, data)
	}

	// rename .oga to .ogg  https://github.com/42wim/matterbridge/issues/906#issuecomment-741793512
	if strings.HasSuffix(name, ".oga") && message.Audio != nil {
		name = strings.Replace(name, ".oga", ".ogg", 1)
	}

	helper.HandleDownloadData(b.Log, rmsg, name, message.Caption, "", data, b.General)
	return nil
}

func (b *Btelegram) getDownloadInfo(id string, suffix string, urlpart bool) (string, string, string) {
	url := b.getFileDirectURL(id)
	name := ""
	if urlpart {
		urlPart := strings.Split(url, "/")
		name = urlPart[len(urlPart)-1]
	}
	if suffix != "" && !strings.HasSuffix(name, suffix) {
		name += suffix
	}
	text := " " + url
	return text, name, url
}

// handleDelete handles message deleting
func (b *Btelegram) handleDelete(msg *config.Message, chatid int64) (string, error) {
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

// handleEdit handles message editing.
func (b *Btelegram) handleEdit(msg *config.Message, chatid int64) (string, error) {
	msgid, err := strconv.Atoi(msg.ID)
	if err != nil {
		return "", err
	}
	if strings.ToLower(b.GetString("MessageFormat")) == HTMLNick {
		b.Log.Debug("Using mode HTML - nick only")
		msg.Text = html.EscapeString(msg.Text)
	}
	m := tgbotapi.NewEditMessageText(chatid, msgid, msg.Username+msg.Text)
	switch b.GetString("MessageFormat") {
	case HTMLFormat:
		b.Log.Debug("Using mode HTML")
		m.ParseMode = tgbotapi.ModeHTML
	case "Markdown":
		b.Log.Debug("Using mode markdown")
		m.ParseMode = tgbotapi.ModeMarkdown
	case MarkdownV2:
		b.Log.Debug("Using mode MarkdownV2")
		m.ParseMode = MarkdownV2
	}
	if strings.ToLower(b.GetString("MessageFormat")) == HTMLNick {
		b.Log.Debug("Using mode HTML - nick only")
		m.ParseMode = tgbotapi.ModeHTML
	}
	_, err = b.c.Send(m)
	if err != nil {
		return "", err
	}
	return "", nil
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
		switch filepath.Ext(fi.Name) {
		case ".jpg", ".jpe", ".png":
			pc := tgbotapi.NewPhotoUpload(chatid, file)
			pc.Caption, pc.ParseMode = TGGetParseMode(b, msg.Username, fi.Comment)
			c = pc
		case ".mp4", ".m4v":
			vc := tgbotapi.NewVideoUpload(chatid, file)
			vc.Caption, vc.ParseMode = TGGetParseMode(b, msg.Username, fi.Comment)
			c = vc
		case ".mp3", ".oga":
			ac := tgbotapi.NewAudioUpload(chatid, file)
			ac.Caption, ac.ParseMode = TGGetParseMode(b, msg.Username, fi.Comment)
			c = ac
		case ".ogg":
			voc := tgbotapi.NewVoiceUpload(chatid, file)
			voc.Caption, voc.ParseMode = TGGetParseMode(b, msg.Username, fi.Comment)
			c = voc
		default:
			dc := tgbotapi.NewDocumentUpload(chatid, file)
			dc.Caption, dc.ParseMode = TGGetParseMode(b, msg.Username, fi.Comment)
			c = dc
		}
		_, err := b.c.Send(c)
		if err != nil {
			b.Log.Errorf("file upload failed: %#v", err)
		}
	}
	return ""
}

func (b *Btelegram) handleQuote(message, quoteNick, quoteMessage string) string {
	format := b.GetString("quoteformat")
	if format == "" {
		format = "{MESSAGE} (re @{QUOTENICK}: {QUOTEMESSAGE})"
	}
	quoteMessagelength := len([]rune(quoteMessage))
	if b.GetInt("QuoteLengthLimit") != 0 && quoteMessagelength >= b.GetInt("QuoteLengthLimit") {
		runes := []rune(quoteMessage)
		quoteMessage = string(runes[0:b.GetInt("QuoteLengthLimit")])
		if quoteMessagelength > b.GetInt("QuoteLengthLimit") {
			quoteMessage += "..."
		}
	}
	format = strings.Replace(format, "{MESSAGE}", message, -1)
	format = strings.Replace(format, "{QUOTENICK}", quoteNick, -1)
	format = strings.Replace(format, "{QUOTEMESSAGE}", quoteMessage, -1)
	return format
}

// handleEntities handles messageEntities
func (b *Btelegram) handleEntities(rmsg *config.Message, message *tgbotapi.Message) {
	if message.Entities == nil {
		return
	}
	// for now only do URL replacements
	for _, e := range *message.Entities {
		if e.Type == "text_link" {
			url, err := e.ParseURL()
			if err != nil {
				b.Log.Errorf("entity text_link url parse failed: %s", err)
				continue
			}
			utfEncodedString := utf16.Encode([]rune(rmsg.Text))
			if e.Offset+e.Length > len(utfEncodedString) {
				b.Log.Errorf("entity length is too long %d > %d", e.Offset+e.Length, len(utfEncodedString))
				continue
			}
			link := utf16.Decode(utfEncodedString[e.Offset : e.Offset+e.Length])
			rmsg.Text = strings.Replace(rmsg.Text, string(link), url.String(), 1)
		}
	}
}
