package bmatrix

import (
	"bytes"
	"mime"
	"regexp"
	"strings"
	"sync"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	log "github.com/Sirupsen/logrus"
	matrix "github.com/matrix-org/gomatrix"
)

type Bmatrix struct {
	mc      *matrix.Client
	Config  *config.Protocol
	Remote  chan config.Message
	Account string
	UserID  string
	RoomMap map[string]string
	sync.RWMutex
}

var flog *log.Entry
var protocol = "matrix"

func init() {
	flog = log.WithFields(log.Fields{"module": protocol})
}

func New(cfg config.Protocol, account string, c chan config.Message) *Bmatrix {
	b := &Bmatrix{}
	b.RoomMap = make(map[string]string)
	b.Config = &cfg
	b.Account = account
	b.Remote = c
	return b
}

func (b *Bmatrix) Connect() error {
	var err error
	flog.Infof("Connecting %s", b.Config.Server)
	b.mc, err = matrix.NewClient(b.Config.Server, "", "")
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	resp, err := b.mc.Login(&matrix.ReqLogin{
		Type:     "m.login.password",
		User:     b.Config.Login,
		Password: b.Config.Password,
	})
	if err != nil {
		flog.Debugf("%#v", err)
		return err
	}
	b.mc.SetCredentials(resp.UserID, resp.AccessToken)
	b.UserID = resp.UserID
	flog.Info("Connection succeeded")
	go b.handlematrix()
	return nil
}

func (b *Bmatrix) Disconnect() error {
	return nil
}

func (b *Bmatrix) JoinChannel(channel config.ChannelInfo) error {
	resp, err := b.mc.JoinRoom(channel.Name, "", nil)
	if err != nil {
		return err
	}
	b.Lock()
	b.RoomMap[resp.RoomID] = channel.Name
	b.Unlock()
	return err
}

func (b *Bmatrix) Send(msg config.Message) (string, error) {
	flog.Debugf("Receiving %#v", msg)
	// ignore delete messages
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}
	channel := b.getRoomID(msg.Channel)
	flog.Debugf("Sending to channel %s", channel)
	if msg.Event == config.EVENT_USER_ACTION {
		b.mc.SendMessageEvent(channel, "m.room.message",
			matrix.TextMessage{"m.emote", msg.Username + msg.Text})
		return "", nil
	}

	if msg.Extra != nil {
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				content := bytes.NewReader(*fi.Data)
				sp := strings.Split(fi.Name, ".")
				mtype := mime.TypeByExtension("." + sp[len(sp)-1])
				if strings.Contains(mtype, "image") ||
					strings.Contains(mtype, "video") {
					flog.Debugf("uploading file: %s %s", fi.Name, mtype)
					res, err := b.mc.UploadToContentRepo(content, mtype, int64(len(*fi.Data)))
					if err != nil {
						flog.Errorf("file upload failed: %#v", err)
						continue
					}
					if strings.Contains(mtype, "video") {
						flog.Debugf("sendVideo %s", res.ContentURI)
						_, err = b.mc.SendVideo(channel, fi.Name, res.ContentURI)
						if err != nil {
							flog.Errorf("sendVideo failed: %#v", err)
						}
					}
					if strings.Contains(mtype, "image") {
						flog.Debugf("sendImage %s", res.ContentURI)
						_, err = b.mc.SendImage(channel, fi.Name, res.ContentURI)
						if err != nil {
							flog.Errorf("sendImage failed: %#v", err)
						}
					}
					flog.Debugf("result: %#v", res)
				}
			}
			return "", nil
		}
	}

	b.mc.SendText(channel, msg.Username+msg.Text)
	return "", nil
}

func (b *Bmatrix) getRoomID(channel string) string {
	b.RLock()
	defer b.RUnlock()
	for ID, name := range b.RoomMap {
		if name == channel {
			return ID
		}
	}
	return ""
}
func (b *Bmatrix) handlematrix() error {
	syncer := b.mc.Syncer.(*matrix.DefaultSyncer)
	syncer.OnEventType("m.room.message", func(ev *matrix.Event) {
		flog.Debugf("Received: %#v", ev)
		if (ev.Content["msgtype"].(string) == "m.text" ||
			ev.Content["msgtype"].(string) == "m.notice" ||
			ev.Content["msgtype"].(string) == "m.emote" ||
			ev.Content["msgtype"].(string) == "m.file" ||
			ev.Content["msgtype"].(string) == "m.image" ||
			ev.Content["msgtype"].(string) == "m.video") && ev.Sender != b.UserID {
			b.RLock()
			channel, ok := b.RoomMap[ev.RoomID]
			b.RUnlock()
			if !ok {
				flog.Debugf("Unknown room %s", ev.RoomID)
				return
			}
			username := ev.Sender[1:]
			if b.Config.NoHomeServerSuffix {
				re := regexp.MustCompile("(.*?):.*")
				username = re.ReplaceAllString(username, `$1`)
			}
			rmsg := config.Message{Username: username, Text: ev.Content["body"].(string), Channel: channel, Account: b.Account, UserID: ev.Sender}
			if ev.Content["msgtype"].(string) == "m.emote" {
				rmsg.Event = config.EVENT_USER_ACTION
			}
			if ev.Content["msgtype"].(string) == "m.image" ||
				ev.Content["msgtype"].(string) == "m.video" ||
				ev.Content["msgtype"].(string) == "m.file" {
				flog.Debugf("ev: %#v", ev)
				rmsg.Extra = make(map[string][]interface{})
				url := ev.Content["url"].(string)
				url = strings.Replace(url, "mxc://", b.Config.Server+"/_matrix/media/v1/download/", -1)
				info := ev.Content["info"].(map[string]interface{})
				size := info["size"].(float64)
				name := ev.Content["body"].(string)
				flog.Debugf("trying to download %#v with size %#v", name, size)
				if size <= 1000000 {
					data, err := helper.DownloadFile(url)
					if err != nil {
						flog.Errorf("download %s failed %#v", url, err)
					} else {
						flog.Debugf("download OK %#v %#v %#v", name, len(*data), len(url))
						rmsg.Extra["file"] = append(rmsg.Extra["file"], config.FileInfo{Name: name, Data: data})
					}
				}
				rmsg.Text = ""
			}
			flog.Debugf("Sending message from %s on %s to gateway", ev.Sender, b.Account)
			b.Remote <- rmsg
		}
	})
	go func() {
		for {
			if err := b.mc.Sync(); err != nil {
				flog.Println("Sync() returned ", err)
			}
		}
	}()
	return nil
}
