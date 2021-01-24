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

type user struct {
	lastname, firstname, avatar string
}

type Bvk struct {
	c            *api.VK
	usernamesMap map[int]user // cache of user names and avatar URLs
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Bvk{usernamesMap: make(map[int]user), Config: cfg}
}

func (b *Bvk) Connect() error {
	b.Log.Info("Connecting")
	b.c = api.NewVK(b.GetString("Token"))
	lp, err := longpoll.NewLongPoll(b.c, b.GetInt("GroupID"))
	if err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}

	lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		b.handleMessage(obj.Message, false)
	})

	b.Log.Info("Connection succeeded")

	go lp.Run()

	return nil
}

func (b *Bvk) Disconnect() error {
	return nil
}

func (b *Bvk) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Bvk) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	peerID, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return "", err
	}

	text := msg.Username

	if msg.Text != "" {
		text += msg.Text
	}

	params := api.Params{
		"message": text,
	}

	if msg.Extra != nil {
		if len(msg.Extra["file"]) > 0 {
			// generate attachments string
			var attachments []string

			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				photoRE := regexp.MustCompile(".(jpg|jpe|png)$")
				if photoRE.MatchString(fi.Name) {
					r := bytes.NewReader(*fi.Data)
					photo, err := b.c.UploadMessagesPhoto(int(peerID), r)
					if err != nil {
						b.Log.Error("Failad uploading photo")
						b.Log.Error(err)
					} else {
						attachments = append(attachments, "photo"+strconv.Itoa(photo[0].OwnerID)+"_"+strconv.Itoa(photo[0].ID))
						if fi.Comment != "" {
							text += fi.Comment + "\n"
						}
					}
				} else {
					r := bytes.NewReader(*fi.Data)

					var doctype string
					if strings.Contains(fi.Name, ".ogg") {
						doctype = "audio_message"
					} else {
						doctype = "doc"
					}

					doc, err := b.c.UploadMessagesDoc(int(peerID), doctype, fi.Name, "", r)
					if err != nil {
						b.Log.Error("Failad uploading file")
						b.Log.Error(err)
					} else if doc.Type == "audio_message" {
						attachments = append(attachments, "doc"+strconv.Itoa(doc.AudioMessage.OwnerID)+"_"+strconv.Itoa(doc.AudioMessage.ID))
					} else if doc.Type == "doc" {
						attachments = append(attachments, "doc"+strconv.Itoa(doc.Doc.OwnerID)+"_"+strconv.Itoa(doc.Doc.ID))
					}
				}

			}
			params["attachment"] = strings.Join(attachments, ",")
			params["message"] = text
		}
	}

	if msg.ID == "" {
		params["random_id"] = time.Now().Unix()
		params["peer_ids"] = msg.Channel

		res, err := b.c.MessagesSendPeerIDs(params)
		if err != nil {
			return "", err
		}

		return strconv.Itoa(res[0].ConversationMessageID), nil
	} else {
		messageID, err := strconv.ParseInt(msg.ID, 10, 64)
		params["peer_id"] = peerID
		params["conversation_message_id"] = messageID

		_, err = b.c.MessagesEdit(params)
		if err != nil {
			return "", err
		}

		return msg.ID, nil
	}
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
		var urls []string

		// get URLs for attachments
		for _, a := range msg.Attachments {
			if a.Type == "photo" {
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
			} else if a.Type == "doc" {
				urls = append(urls, a.Doc.URL)
			} else if a.Type == "graffiti" {
				urls = append(urls, a.Graffiti.URL)
			} else if a.Type == "audio_message" {
				urls = append(urls, a.AudioMessage.DocsDocPreviewAudioMessage.LinkOgg)
			} else if a.Type == "sticker" {
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
			} else if a.Type == "doc" {
				urls = append(urls, a.Doc.URL)
			} else if a.Type == "video" {
				rmsg.Text += "https://vk.com/video" + strconv.Itoa(a.Video.OwnerID) + "_" + strconv.Itoa(a.Video.ID)
			} else if a.Type == "wall" {
				rmsg.Text += "https://vk.com/wall" + strconv.Itoa(a.Wall.FromID) + "_" + strconv.Itoa(a.Wall.ID)
			} else {
				rmsg.Text += "This attachment is not supported (" + a.Type + ")"
			}
		}

		if b.GetBool("UseFileURL") {
			// add url to message text
			rmsg.Text += "\n" + strings.Join(urls, "\n")
		} else {
			// download
			for _, url := range urls {
				data, err := helper.DownloadFile(url)
				if err == nil {
					urlPart := strings.Split(url, "/")
					name := strings.Split(urlPart[len(urlPart)-1], "?")[0]
					helper.HandleDownloadData(b.Log, &rmsg, name, "", url, data, b.General)
				}
			}
		}
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
