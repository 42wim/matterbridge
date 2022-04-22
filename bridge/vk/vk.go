package bvk

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	longpoll "github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"github.com/SevereCloud/vksdk/v2/object"
)

const (
	audioMessage = "audio_message"
	document     = "doc"
	photo        = "photo"
	video        = "video"
	graffiti     = "graffiti"
	sticker      = "sticker"
	wall         = "wall"
)

type user struct {
	lastname, firstname, avatar string
}

type Bvk struct {
	c            *api.VK
	lp           *longpoll.LongPoll
	usernamesMap map[int]user // cache of user names and avatar URLs
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bvk{usernamesMap: make(map[int]user), Config: cfg}
}

func (b *Bvk) Connect() error {
	b.Log.Info("Connecting")
	b.c = api.NewVK(b.GetString("Token"))

	var err error
	b.lp, err = longpoll.NewLongPollCommunity(b.c)
	if err != nil {
		b.Log.Debugf("%#v", err)

		return err
	}

	b.lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		b.handleMessage(obj.Message, false)
	})

	b.Log.Info("Connection succeeded")

	go func() {
		err := b.lp.Run()
		if err != nil {
			b.Log.WithError(err).Fatal("Enable longpoll in group management")
		}
	}()

	return nil
}

func (b *Bvk) Disconnect() error {
	b.lp.Shutdown()

	return nil
}

func (b *Bvk) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Bvk) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	peerID, err := strconv.Atoi(msg.Channel)
	if err != nil {
		return "", err
	}

	params := api.Params{}

	text := msg.Username + msg.Text

	if msg.Extra != nil {
		if len(msg.Extra["file"]) > 0 {
			// generate attachments string
			attachment, urls := b.uploadFiles(msg.Extra, peerID)
			params["attachment"] = attachment
			text += urls
		}
	}

	params["message"] = text

	if msg.ID == "" {
		// New message
		params["random_id"] = time.Now().Unix()
		params["peer_ids"] = msg.Channel

		res, e := b.c.MessagesSendPeerIDs(params)
		if e != nil {
			return "", err
		}

		return strconv.Itoa(res[0].ConversationMessageID), nil
	}
	// Edit message
	messageID, err := strconv.ParseInt(msg.ID, 10, 64)
	if err != nil {
		return "", err
	}

	params["peer_id"] = peerID
	params["conversation_message_id"] = messageID

	_, err = b.c.MessagesEdit(params)
	if err != nil {
		return "", err
	}

	return msg.ID, nil
}

func (b *Bvk) getUser(id int) user {
	u, found := b.usernamesMap[id]
	if !found {
		b.Log.Debug("Fetching username for ", id)

		if id >= 0 {
			result, _ := b.c.UsersGet(api.Params{
				"user_ids": id,
				"fields":   "photo_200",
			})

			resUser := result[0]
			u = user{lastname: resUser.LastName, firstname: resUser.FirstName, avatar: resUser.Photo200}
			b.usernamesMap[id] = u
		} else {
			result, _ := b.c.GroupsGetByID(api.Params{
				"group_id": id * -1,
			})

			resGroup := result[0]
			u = user{lastname: resGroup.Name, avatar: resGroup.Photo200}
		}
	}

	return u
}

func (b *Bvk) handleMessage(msg object.MessagesMessage, isFwd bool) {
	b.Log.Debug("ChatID: ", msg.PeerID)
	// fetch user info
	u := b.getUser(msg.FromID)

	rmsg := config.Message{
		Text:     msg.Text,
		Username: u.firstname + " " + u.lastname,
		Avatar:   u.avatar,
		Channel:  strconv.Itoa(msg.PeerID),
		Account:  b.Account,
		UserID:   strconv.Itoa(msg.FromID),
		ID:       strconv.Itoa(msg.ConversationMessageID),
		Extra:    make(map[string][]interface{}),
	}

	if msg.ReplyMessage != nil {
		ur := b.getUser(msg.ReplyMessage.FromID)
		rmsg.Text = "Re: " + ur.firstname + " " + ur.lastname + "\n" + rmsg.Text
	}

	if isFwd {
		rmsg.Username = "Fwd: " + rmsg.Username
	}

	if len(msg.Attachments) > 0 {
		urls, text := b.getFiles(msg.Attachments)

		if text != "" {
			rmsg.Text += "\n" + text
		}

		// download
		b.downloadFiles(&rmsg, urls)
	}

	if len(msg.FwdMessages) > 0 {
		rmsg.Text += strconv.Itoa(len(msg.FwdMessages)) + " forwarded messages"
	}

	b.Remote <- rmsg

	if len(msg.FwdMessages) > 0 {
		// recursive processing of forwarded messages
		for _, m := range msg.FwdMessages {
			m.PeerID = msg.PeerID
			b.handleMessage(m, true)
		}
	}
}

func (b *Bvk) uploadFiles(extra map[string][]interface{}, peerID int) (string, string) {
	var attachments []string
	text := ""

	for _, f := range extra["file"] {
		fi := f.(config.FileInfo)

		if fi.Comment != "" {
			text += fi.Comment + "\n"
		}
		a, err := b.uploadFile(fi, peerID)
		if err != nil {
			b.Log.WithError(err).Error("File upload error ", fi.Name)
		}

		attachments = append(attachments, a)
	}

	return strings.Join(attachments, ","), text
}

func (b *Bvk) uploadFile(file config.FileInfo, peerID int) (string, error) {
	r := bytes.NewReader(*file.Data)

	photoRE := regexp.MustCompile(".(jpg|jpe|png)$")
	if photoRE.MatchString(file.Name) {
		// BUG(VK): for community chat peerID=0
		p, err := b.c.UploadMessagesPhoto(0, r)
		if err != nil {
			return "", err
		}

		return photo + strconv.Itoa(p[0].OwnerID) + "_" + strconv.Itoa(p[0].ID), nil
	}

	var doctype string
	if strings.Contains(file.Name, ".ogg") {
		doctype = audioMessage
	} else {
		doctype = document
	}

	doc, err := b.c.UploadMessagesDoc(peerID, doctype, file.Name, "", r)
	if err != nil {
		return "", err
	}

	switch doc.Type {
	case audioMessage:
		return document + strconv.Itoa(doc.AudioMessage.OwnerID) + "_" + strconv.Itoa(doc.AudioMessage.ID), nil
	case document:
		return document + strconv.Itoa(doc.Doc.OwnerID) + "_" + strconv.Itoa(doc.Doc.ID), nil
	}

	return "", nil
}

func (b *Bvk) getFiles(attachments []object.MessagesMessageAttachment) ([]string, string) {
	var urls []string
	var text []string

	for _, a := range attachments {
		switch a.Type {
		case photo:
			var resolution float64 = 0
			url := a.Photo.Sizes[0].URL
			for _, size := range a.Photo.Sizes {
				r := size.Height * size.Width
				if resolution < r {
					resolution = r
					url = size.URL
				}
			}

			urls = append(urls, url)

		case document:
			urls = append(urls, a.Doc.URL)

		case graffiti:
			urls = append(urls, a.Graffiti.URL)

		case audioMessage:
			urls = append(urls, a.AudioMessage.DocsDocPreviewAudioMessage.LinkOgg)

		case sticker:
			var resolution float64 = 0
			url := a.Sticker.Images[0].URL
			for _, size := range a.Sticker.Images {
				r := size.Height * size.Width
				if resolution < r {
					resolution = r
					url = size.URL
				}
			}
			urls = append(urls, url+".png")
		case video:
			text = append(text, "https://vk.com/video"+strconv.Itoa(a.Video.OwnerID)+"_"+strconv.Itoa(a.Video.ID))

		case wall:
			text = append(text, "https://vk.com/wall"+strconv.Itoa(a.Wall.FromID)+"_"+strconv.Itoa(a.Wall.ID))

		default:
			text = append(text, "This attachment is not supported ("+a.Type+")")
		}
	}

	return urls, strings.Join(text, "\n")
}

func (b *Bvk) downloadFiles(rmsg *config.Message, urls []string) {
	for _, url := range urls {
		data, err := helper.DownloadFile(url)
		if err == nil {
			urlPart := strings.Split(url, "/")
			name := strings.Split(urlPart[len(urlPart)-1], "?")[0]
			helper.HandleDownloadData(b.Log, rmsg, name, "", url, data, b.General)
		}
	}
}
