package bmatrix

import (
	"bytes"
	"fmt"
	"html"
	"mime"
	"regexp"
	"strings"
	"sync"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	matrix "github.com/matterbridge/gomatrix"
)

type Bmatrix struct {
	mc      *matrix.Client
	UserID  string
	RoomMap map[string]string
	sync.RWMutex
	htmlTag *regexp.Regexp
	*bridge.Config
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmatrix{Config: cfg}
	b.htmlTag = regexp.MustCompile("</.*?>")
	b.RoomMap = make(map[string]string)
	return b
}

func (b *Bmatrix) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	b.mc, err = matrix.NewClient(b.GetString("Server"), "", "")
	if err != nil {
		return err
	}
	resp, err := b.mc.Login(&matrix.ReqLogin{
		Type:     "m.login.password",
		User:     b.GetString("Login"),
		Password: b.GetString("Password"),
	})
	if err != nil {
		return err
	}
	b.mc.SetCredentials(resp.UserID, resp.AccessToken)
	b.UserID = resp.UserID
	b.Log.Info("Connection succeeded")
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
	b.Log.Debugf("=> Receiving %#v", msg)

	channel := b.getRoomID(msg.Channel)
	b.Log.Debugf("Channel %s maps to channel id %s", msg.Channel, channel)

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		m := matrix.TextMessage{
			MsgType: "m.emote",
			Body:    msg.Username + msg.Text,
		}
		resp, err := b.mc.SendMessageEvent(channel, "m.room.message", m)
		if err != nil {
			return "", err
		}
		return resp.EventID, err
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
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
			if _, err := b.mc.SendText(channel, rmsg.Username+rmsg.Text); err != nil {
				b.Log.Errorf("sendText failed: %s", err)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFiles(&msg, channel)
		}
	}

	// Edit message if we have an ID
	// matrix has no editing support

	// Use notices to send join/leave events
	if msg.Event == config.EventJoinLeave {
		resp, err := b.mc.SendNotice(channel, msg.Username+msg.Text)
		if err != nil {
			return "", err
		}
		return resp.EventID, err
	}

	username := html.EscapeString(msg.Username)
	// check if we have a </tag>. if we have, we don't escape HTML. #696
	if b.htmlTag.MatchString(msg.Username) {
		username = msg.Username
	}
	// Post normal message with HTML support (eg riot.im)
	resp, err := b.mc.SendHTML(channel, msg.Username+msg.Text, username+helper.ParseMarkdown(msg.Text))
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

func (b *Bmatrix) handlematrix() {
	syncer := b.mc.Syncer.(*matrix.DefaultSyncer)
	syncer.OnEventType("m.room.redaction", b.handleEvent)
	syncer.OnEventType("m.room.message", b.handleEvent)
	go func() {
		for {
			if err := b.mc.Sync(); err != nil {
				b.Log.Println("Sync() returned ", err)
			}
		}
	}()
}

func (b *Bmatrix) handleEvent(ev *matrix.Event) {
	b.Log.Debugf("== Receiving event: %#v", ev)
	if ev.Sender != b.UserID {
		b.RLock()
		channel, ok := b.RoomMap[ev.RoomID]
		b.RUnlock()
		if !ok {
			b.Log.Debugf("Unknown room %s", ev.RoomID)
			return
		}

		// TODO download avatar

		// Create our message
		rmsg := config.Message{Username: ev.Sender[1:], Channel: channel, Account: b.Account, UserID: ev.Sender, ID: ev.ID}

		// Text must be a string
		if rmsg.Text, ok = ev.Content["body"].(string); !ok {
			b.Log.Errorf("Content[body] is not a string: %T\n%#v",
				ev.Content["body"], ev.Content)
			return
		}

		// Remove homeserver suffix if configured
		if b.GetBool("NoHomeServerSuffix") {
			re := regexp.MustCompile("(.*?):.*")
			rmsg.Username = re.ReplaceAllString(rmsg.Username, `$1`)
		}

		// Delete event
		if ev.Type == "m.room.redaction" {
			rmsg.Event = config.EventMsgDelete
			rmsg.ID = ev.Redacts
			rmsg.Text = config.EventMsgDelete
			b.Remote <- rmsg
			return
		}

		// Do we have a /me action
		if ev.Content["msgtype"].(string) == "m.emote" {
			rmsg.Event = config.EventUserAction
		}

		// Do we have attachments
		if b.containsAttachment(ev.Content) {
			err := b.handleDownloadFile(&rmsg, ev.Content)
			if err != nil {
				b.Log.Errorf("download failed: %#v", err)
			}
		}

		b.Log.Debugf("<= Sending message from %s on %s to gateway", ev.Sender, b.Account)
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
	url = strings.Replace(url, "mxc://", b.GetString("Server")+"/_matrix/media/v1/download/", -1)

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
				name += mext[0]
			}
		} else {
			// just a default .png extension if we don't have mime info
			name += ".png"
		}
	}

	// check if the size is ok
	err := helper.HandleDownloadSize(b.Log, rmsg, name, int64(size), b.General)
	if err != nil {
		return err
	}
	// actually download the file
	data, err := helper.DownloadFile(url)
	if err != nil {
		return fmt.Errorf("download %s failed %#v", url, err)
	}
	// add the downloaded data to the message
	helper.HandleDownloadData(b.Log, rmsg, name, "", url, data, b.General)
	return nil
}

// handleUploadFiles handles native upload of files.
func (b *Bmatrix) handleUploadFiles(msg *config.Message, channel string) (string, error) {
	for _, f := range msg.Extra["file"] {
		if fi, ok := f.(config.FileInfo); ok {
			b.handleUploadFile(msg, channel, &fi)
		}
	}
	return "", nil
}

// handleUploadFile handles native upload of a file.
func (b *Bmatrix) handleUploadFile(msg *config.Message, channel string, fi *config.FileInfo) {
	content := bytes.NewReader(*fi.Data)
	sp := strings.Split(fi.Name, ".")
	mtype := mime.TypeByExtension("." + sp[len(sp)-1])
	if !(strings.Contains(mtype, "image") || strings.Contains(mtype, "video") ||
		strings.Contains(mtype, "application") || strings.Contains(mtype, "audio")) {
		return
	}
	if fi.Comment != "" {
		_, err := b.mc.SendText(channel, msg.Username+fi.Comment)
		if err != nil {
			b.Log.Errorf("file comment failed: %#v", err)
		}
	} else {
		// image and video uploads send no username, we have to do this ourself here #715
		_, err := b.mc.SendText(channel, msg.Username)
		if err != nil {
			b.Log.Errorf("file comment failed: %#v", err)
		}
	}
	b.Log.Debugf("uploading file: %s %s", fi.Name, mtype)
	res, err := b.mc.UploadToContentRepo(content, mtype, int64(len(*fi.Data)))
	if err != nil {
		b.Log.Errorf("file upload failed: %#v", err)
		return
	}

	switch {
	case strings.Contains(mtype, "video"):
		b.Log.Debugf("sendVideo %s", res.ContentURI)
		_, err = b.mc.SendVideo(channel, fi.Name, res.ContentURI)
		if err != nil {
			b.Log.Errorf("sendVideo failed: %#v", err)
		}
	case strings.Contains(mtype, "image"):
		b.Log.Debugf("sendImage %s", res.ContentURI)
		_, err = b.mc.SendImage(channel, fi.Name, res.ContentURI)
		if err != nil {
			b.Log.Errorf("sendImage failed: %#v", err)
		}
	case strings.Contains(mtype, "application"):
		b.Log.Debugf("sendFile %s", res.ContentURI)
		_, err = b.mc.SendFile(channel, fi.Name, res.ContentURI, mtype, uint(len(*fi.Data)))
		if err != nil {
			b.Log.Errorf("sendFile failed: %#v", err)
		}
	case strings.Contains(mtype, "audio"):
		b.Log.Debugf("sendAudio %s", res.ContentURI)
		_, err = b.mc.SendAudio(channel, fi.Name, res.ContentURI, mtype, uint(len(*fi.Data)))
		if err != nil {
			b.Log.Errorf("sendAudio failed: %#v", err)
		}
	}
	b.Log.Debugf("result: %#v", res)
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
