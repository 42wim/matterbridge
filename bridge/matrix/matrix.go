package bmatrix

import (
	"bytes"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	matrix "github.com/matterbridge/gomatrix"
	log "github.com/sirupsen/logrus"
	"mime"
	"regexp"
	"strings"
	"sync"
)

type Bmatrix struct {
	mc      *matrix.Client
	UserID  string
	RoomMap map[string]string
	sync.RWMutex
	*config.BridgeConfig
}

var flog *log.Entry
var protocol = "matrix"

func init() {
	flog = log.WithFields(log.Fields{"prefix": protocol})
}

func New(cfg *config.BridgeConfig) *Bmatrix {
	b := &Bmatrix{BridgeConfig: cfg}
	b.RoomMap = make(map[string]string)
	return b
}

func (b *Bmatrix) Connect() error {
	var err error
	flog.Infof("Connecting %s", b.Config.Server)
	b.mc, err = matrix.NewClient(b.Config.Server, "", "")
	if err != nil {
		return err
	}
	resp, err := b.mc.Login(&matrix.ReqLogin{
		Type:     "m.login.password",
		User:     b.Config.Login,
		Password: b.Config.Password,
	})
	if err != nil {
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

	channel := b.getRoomID(msg.Channel)
	flog.Debugf("Channel %s maps to channel id %s", msg.Channel, channel)

	// Make a action /me of the message
	if msg.Event == config.EVENT_USER_ACTION {
		resp, err := b.mc.SendMessageEvent(channel, "m.room.message",
			matrix.TextMessage{"m.emote", msg.Username + msg.Text})
		if err != nil {
			return "", err
		}
		return resp.EventID, err
	}

	// Delete message
	if msg.Event == config.EVENT_MSG_DELETE {
		if msg.ID == "" {
			return "", nil
		}
		resp, err := b.mc.RedactEvent(channel, msg.ID, &matrix.ReqRedact{})
		if err != nil {
			return "", err
		}
		return resp.EventID, err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			b.mc.SendText(channel, rmsg.Username+rmsg.Text)
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFile(&msg, channel)
		}
	}

	// Edit message if we have an ID
	// matrix has no editing support

	// Post normal message
	resp, err := b.mc.SendText(channel, msg.Username+msg.Text)
	if err != nil {
		return "", err
	}
	return resp.EventID, err
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
	syncer.OnEventType("m.room.redaction", b.handleEvent)
	syncer.OnEventType("m.room.message", b.handleEvent)
	go func() {
		for {
			if err := b.mc.Sync(); err != nil {
				flog.Println("Sync() returned ", err)
			}
		}
	}()
	return nil
}

func (b *Bmatrix) handleEvent(ev *matrix.Event) {
	flog.Debugf("Received: %#v", ev)
	if ev.Sender != b.UserID {
		b.RLock()
		channel, ok := b.RoomMap[ev.RoomID]
		b.RUnlock()
		if !ok {
			flog.Debugf("Unknown room %s", ev.RoomID)
			return
		}

		// TODO download avatar

		// Create our message
		rmsg := config.Message{Username: ev.Sender[1:], Channel: channel, Account: b.Account, UserID: ev.Sender, ID: ev.ID}

		// Text must be a string
		if rmsg.Text, ok = ev.Content["body"].(string); !ok {
			flog.Errorf("Content[body] wasn't a %T ?", rmsg.Text)
			return
		}

		// Remove homeserver suffix if configured
		if b.Config.NoHomeServerSuffix {
			re := regexp.MustCompile("(.*?):.*")
			rmsg.Username = re.ReplaceAllString(rmsg.Username, `$1`)
		}

		// Delete event
		if ev.Type == "m.room.redaction" {
			rmsg.Event = config.EVENT_MSG_DELETE
			rmsg.ID = ev.Redacts
			rmsg.Text = config.EVENT_MSG_DELETE
			b.Remote <- rmsg
			return
		}

		// Do we have a /me action
		if ev.Content["msgtype"].(string) == "m.emote" {
			rmsg.Event = config.EVENT_USER_ACTION
		}

		// Do we have attachments
		if b.containsAttachment(ev.Content) {
			err := b.handleDownloadFile(&rmsg, ev.Content)
			if err != nil {
				flog.Errorf("download failed: %#v", err)
			}
		}

		flog.Debugf("Sending message from %s on %s to gateway", ev.Sender, b.Account)
		b.Remote <- rmsg
	}
}

// handleDownloadFile handles file download
func (b *Bmatrix) handleDownloadFile(rmsg *config.Message, content map[string]interface{}) error {
	var (
		ok                        bool
		url, name, msgtype, mtype string
		info                      map[string]interface{}
		size                      float64
	)

	rmsg.Extra = make(map[string][]interface{})
	if url, ok = content["url"].(string); !ok {
		return fmt.Errorf("url isn't a %T", url)
	}
	url = strings.Replace(url, "mxc://", b.Config.Server+"/_matrix/media/v1/download/", -1)

	if info, ok = content["info"].(map[string]interface{}); !ok {
		return fmt.Errorf("info isn't a %T", info)
	}
	if size, ok = info["size"].(float64); !ok {
		return fmt.Errorf("size isn't a %T", size)
	}
	if name, ok = content["body"].(string); !ok {
		return fmt.Errorf("name isn't a %T", name)
	}
	if msgtype, ok = content["msgtype"].(string); !ok {
		return fmt.Errorf("msgtype isn't a %T", msgtype)
	}
	if mtype, ok = info["mimetype"].(string); !ok {
		return fmt.Errorf("mtype isn't a %T", mtype)
	}

	// check if we have an image uploaded without extension
	if !strings.Contains(name, ".") {
		if msgtype == "m.image" {
			mext, _ := mime.ExtensionsByType(mtype)
			if len(mext) > 0 {
				name = name + mext[0]
			}
		} else {
			// just a default .png extension if we don't have mime info
			name = name + ".png"
		}
	}

	// check if the size is ok
	err := helper.HandleDownloadSize(flog, rmsg, name, int64(size), b.General)
	if err != nil {
		return err
	}
	// actually download the file
	data, err := helper.DownloadFile(url)
	if err != nil {
		return fmt.Errorf("download %s failed %#v", url, err)
	}
	// add the downloaded data to the message
	helper.HandleDownloadData(flog, rmsg, name, "", url, data, b.General)
	return nil
}

// handleUploadFile handles native upload of files
func (b *Bmatrix) handleUploadFile(msg *config.Message, channel string) (string, error) {
	for _, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		content := bytes.NewReader(*fi.Data)
		sp := strings.Split(fi.Name, ".")
		mtype := mime.TypeByExtension("." + sp[len(sp)-1])
		if strings.Contains(mtype, "image") ||
			strings.Contains(mtype, "video") {
			if fi.Comment != "" {
				_, err := b.mc.SendText(channel, msg.Username+fi.Comment)
				if err != nil {
					flog.Errorf("file comment failed: %#v", err)
				}
			}
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

// skipMessages returns true if this message should not be handled
func (b *Bmatrix) containsAttachment(content map[string]interface{}) bool {
	// Skip empty messages
	if content["msgtype"] == nil {
		return false
	}

	// Only allow image,video or file msgtypes
	if !(content["msgtype"].(string) == "m.image" ||
		content["msgtype"].(string) == "m.video" ||
		content["msgtype"].(string) == "m.file") {
		return false
	}
	return true
}
