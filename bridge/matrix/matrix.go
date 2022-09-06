package bmatrix

import (
	"bytes"
	"fmt"
	"mime"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	matrix "github.com/matterbridge/gomatrix"
)

var (
	htmlTag            = regexp.MustCompile("</.*?>")
	htmlReplacementTag = regexp.MustCompile("<[^>]*>")
)

type NicknameCacheEntry struct {
	displayName string
	lastUpdated time.Time
}

type Bmatrix struct {
	mc          *matrix.Client
	UserID      string
	NicknameMap map[string]NicknameCacheEntry
	RoomMap     map[string]string
	rateMutex   sync.RWMutex
	sync.RWMutex
	*bridge.Config
}

type httpError struct {
	Errcode      string `json:"errcode"`
	Err          string `json:"error"`
	RetryAfterMs int    `json:"retry_after_ms"`
}

type matrixUsername struct {
	plain     string
	formatted string
}

// SubTextMessage represents the new content of the message in edit messages.
type SubTextMessage struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	FormattedBody string `json:"formatted_body,omitempty"`
	Format        string `json:"format,omitempty"`
}

// MessageRelation explains how the current message relates to a previous message.
// Notably used for message edits.
type MessageRelation struct {
	EventID string `json:"event_id"`
	Type    string `json:"rel_type"`
}

type EditedMessage struct {
	NewContent SubTextMessage  `json:"m.new_content"`
	RelatedTo  MessageRelation `json:"m.relates_to"`
	matrix.TextMessage
}

type InReplyToRelationContent struct {
	EventID string `json:"event_id"`
}

type InReplyToRelation struct {
	InReplyTo InReplyToRelationContent `json:"m.in_reply_to"`
}

type ReplyMessage struct {
	RelatedTo InReplyToRelation `json:"m.relates_to"`
	matrix.TextMessage
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmatrix{Config: cfg}
	b.RoomMap = make(map[string]string)
	b.NicknameMap = make(map[string]NicknameCacheEntry)
	return b
}

func (b *Bmatrix) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	if b.GetString("MxID") != "" && b.GetString("Token") != "" {
		b.mc, err = matrix.NewClient(
			b.GetString("Server"), b.GetString("MxID"), b.GetString("Token"),
		)
		if err != nil {
			return err
		}
		b.UserID = b.GetString("MxID")
		b.Log.Info("Using existing Matrix credentials")
	} else {
		b.mc, err = matrix.NewClient(b.GetString("Server"), "", "")
		if err != nil {
			return err
		}
		resp, err := b.mc.Login(&matrix.ReqLogin{
			Type:       "m.login.password",
			User:       b.GetString("Login"),
			Password:   b.GetString("Password"),
			Identifier: matrix.NewUserIdentifier(b.GetString("Login")),
		})
		if err != nil {
			return err
		}
		b.mc.SetCredentials(resp.UserID, resp.AccessToken)
		b.UserID = resp.UserID
		b.Log.Info("Connection succeeded")
	}
	go b.handlematrix()
	return nil
}

func (b *Bmatrix) Disconnect() error {
	return nil
}

func (b *Bmatrix) JoinChannel(channel config.ChannelInfo) error {
	return b.retry(func() error {
		resp, err := b.mc.JoinRoom(channel.Name, "", nil)
		if err != nil {
			return err
		}

		b.Lock()
		b.RoomMap[resp.RoomID] = channel.Name
		b.Unlock()

		return nil
	})
}

func (b *Bmatrix) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	channel := b.getRoomID(msg.Channel)
	b.Log.Debugf("Channel %s maps to channel id %s", msg.Channel, channel)

	username := newMatrixUsername(msg.Username)

	body := username.plain + msg.Text
	formattedBody := username.formatted + helper.ParseMarkdown(msg.Text)

	if b.GetBool("SpoofUsername") {
		// https://spec.matrix.org/v1.3/client-server-api/#mroommember
		type stateMember struct {
			AvatarURL   string `json:"avatar_url,omitempty"`
			DisplayName string `json:"displayname"`
			Membership  string `json:"membership"`
		}

		// TODO: reset username afterwards with DisplayName: null ?
		m := stateMember{
			AvatarURL:   "",
			DisplayName: username.plain,
			Membership:  "join",
		}

		_, err := b.mc.SendStateEvent(channel, "m.room.member", b.UserID, m)
		if err == nil {
			body = msg.Text
			formattedBody = helper.ParseMarkdown(msg.Text)
		}
	}

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		m := matrix.TextMessage{
			MsgType:       "m.emote",
			Body:          body,
			FormattedBody: formattedBody,
			Format:        "org.matrix.custom.html",
		}

		if b.GetBool("HTMLDisable") {
			m.Format = ""
			m.FormattedBody = ""
		}

		msgID := ""

		err := b.retry(func() error {
			resp, err := b.mc.SendMessageEvent(channel, "m.room.message", m)
			if err != nil {
				return err
			}

			msgID = resp.EventID

			return err
		})

		return msgID, err
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}

		msgID := ""

		err := b.retry(func() error {
			resp, err := b.mc.RedactEvent(channel, msg.ID, &matrix.ReqRedact{})
			if err != nil {
				return err
			}

			msgID = resp.EventID

			return err
		})

		return msgID, err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			rmsg := rmsg

			err := b.retry(func() error {
				_, err := b.mc.SendText(channel, rmsg.Username+rmsg.Text)

				return err
			})
			if err != nil {
				b.Log.Errorf("sendText failed: %s", err)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			return b.handleUploadFiles(&msg, channel)
		}
	}

	// Edit message if we have an ID
	if msg.ID != "" {
		rmsg := EditedMessage{
			TextMessage: matrix.TextMessage{
				Body:          body,
				MsgType:       "m.text",
				Format:        "org.matrix.custom.html",
				FormattedBody: formattedBody,
			},
		}

		rmsg.NewContent = SubTextMessage{
			Body:          rmsg.TextMessage.Body,
			FormattedBody: rmsg.TextMessage.FormattedBody,
			Format:        rmsg.TextMessage.Format,
			MsgType:       "m.text",
		}

		if b.GetBool("HTMLDisable") {
			rmsg.TextMessage.Format = ""
			rmsg.TextMessage.FormattedBody = ""
			rmsg.NewContent.Format = ""
			rmsg.NewContent.FormattedBody = ""
		}

		rmsg.RelatedTo = MessageRelation{
			EventID: msg.ID,
			Type:    "m.replace",
		}

		err := b.retry(func() error {
			_, err := b.mc.SendMessageEvent(channel, "m.room.message", rmsg)

			return err
		})
		if err != nil {
			return "", err
		}

		return msg.ID, nil
	}

	// Use notices to send join/leave events
	if msg.Event == config.EventJoinLeave {
		m := matrix.TextMessage{
			MsgType:       "m.notice",
			Body:          body,
			FormattedBody: formattedBody,
			Format:        "org.matrix.custom.html",
		}

		if b.GetBool("HTMLDisable") {
			m.Format = ""
			m.FormattedBody = ""
		}

		var (
			resp *matrix.RespSendEvent
			err  error
		)

		err = b.retry(func() error {
			resp, err = b.mc.SendMessageEvent(channel, "m.room.message", m)

			return err
		})
		if err != nil {
			return "", err
		}

		return resp.EventID, err
	}

	if msg.ParentValid() {
		m := ReplyMessage{
			TextMessage: matrix.TextMessage{
				MsgType:       "m.text",
				Body:          body,
				FormattedBody: formattedBody,
				Format:        "org.matrix.custom.html",
			},
		}

		if b.GetBool("HTMLDisable") {
			m.TextMessage.Format = ""
			m.TextMessage.FormattedBody = ""
		}

		m.RelatedTo = InReplyToRelation{
			InReplyTo: InReplyToRelationContent{
				EventID: msg.ParentID,
			},
		}

		var (
			resp *matrix.RespSendEvent
			err  error
		)

		err = b.retry(func() error {
			resp, err = b.mc.SendMessageEvent(channel, "m.room.message", m)

			return err
		})
		if err != nil {
			return "", err
		}

		return resp.EventID, err
	}

	if b.GetBool("HTMLDisable") {
		var (
			resp *matrix.RespSendEvent
			err  error
		)

		err = b.retry(func() error {
			resp, err = b.mc.SendText(channel, body)

			return err
		})
		if err != nil {
			return "", err
		}

		return resp.EventID, err
	}

	// Post normal message with HTML support (eg riot.im)
	var (
		resp *matrix.RespSendEvent
		err  error
	)

	err = b.retry(func() error {
		resp, err = b.mc.SendFormattedText(channel, body, formattedBody)

		return err
	})
	if err != nil {
		return "", err
	}

	return resp.EventID, err
}

func (b *Bmatrix) handlematrix() {
	syncer := b.mc.Syncer.(*matrix.DefaultSyncer)
	syncer.OnEventType("m.room.redaction", b.handleEvent)
	syncer.OnEventType("m.room.message", b.handleEvent)
	syncer.OnEventType("m.room.member", b.handleMemberChange)
	go func() {
		for {
			if b == nil {
				return
			}
			if err := b.mc.Sync(); err != nil {
				b.Log.Println("Sync() returned ", err)
			}
		}
	}()
}

func (b *Bmatrix) handleEdit(ev *matrix.Event, rmsg config.Message) bool {
	relationInterface, present := ev.Content["m.relates_to"]
	newContentInterface, present2 := ev.Content["m.new_content"]
	if !(present && present2) {
		return false
	}

	var relation MessageRelation
	if err := interface2Struct(relationInterface, &relation); err != nil {
		b.Log.Warnf("Couldn't parse 'm.relates_to' object with value %#v", relationInterface)
		return false
	}

	var newContent SubTextMessage
	if err := interface2Struct(newContentInterface, &newContent); err != nil {
		b.Log.Warnf("Couldn't parse 'm.new_content' object with value %#v", newContentInterface)
		return false
	}

	if relation.Type != "m.replace" {
		return false
	}

	rmsg.ID = relation.EventID
	rmsg.Text = newContent.Body
	b.Remote <- rmsg

	return true
}

func (b *Bmatrix) handleReply(ev *matrix.Event, rmsg config.Message) bool {
	relationInterface, present := ev.Content["m.relates_to"]
	if !present {
		return false
	}

	var relation InReplyToRelation
	if err := interface2Struct(relationInterface, &relation); err != nil {
		// probably fine
		return false
	}

	body := rmsg.Text

	if !b.GetBool("keepquotedreply") {
		for strings.HasPrefix(body, "> ") {
			lineIdx := strings.IndexRune(body, '\n')
			if lineIdx == -1 {
				body = ""
			} else {
				body = body[(lineIdx + 1):]
			}
		}
	}

	rmsg.Text = body
	rmsg.ParentID = relation.InReplyTo.EventID
	b.Remote <- rmsg

	return true
}

func (b *Bmatrix) handleMemberChange(ev *matrix.Event) {
	// Update the displayname on join messages, according to https://matrix.org/docs/spec/client_server/r0.6.1#events-on-change-of-profile-information
	if ev.Content["membership"] == "join" {
		if dn, ok := ev.Content["displayname"].(string); ok {
			b.cacheDisplayName(ev.Sender, dn)
		}
	}
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

		// Create our message
		rmsg := config.Message{
			Username: b.getDisplayName(ev.Sender),
			Channel:  channel,
			Account:  b.Account,
			UserID:   ev.Sender,
			ID:       ev.ID,
			Avatar:   b.getAvatarURL(ev.Sender),
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

		// Text must be a string
		if rmsg.Text, ok = ev.Content["body"].(string); !ok {
			b.Log.Errorf("Content[body] is not a string: %T\n%#v",
				ev.Content["body"], ev.Content)
			return
		}

		// Do we have a /me action
		if ev.Content["msgtype"].(string) == "m.emote" {
			rmsg.Event = config.EventUserAction
		}

		// Is it an edit?
		if b.handleEdit(ev, rmsg) {
			return
		}

		// Is it a reply?
		if b.handleReply(ev, rmsg) {
			return
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

		// not crucial, so no ratelimit check here
		if err := b.mc.MarkRead(ev.RoomID, ev.ID); err != nil {
			b.Log.Errorf("couldn't mark message as read %s", err.Error())
		}
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
	username := newMatrixUsername(msg.Username)
	content := bytes.NewReader(*fi.Data)
	sp := strings.Split(fi.Name, ".")
	mtype := mime.TypeByExtension("." + sp[len(sp)-1])
	// image and video uploads send no username, we have to do this ourself here #715
	err := b.retry(func() error {
		_, err := b.mc.SendFormattedText(channel, username.plain+fi.Comment, username.formatted+fi.Comment)

		return err
	})
	if err != nil {
		b.Log.Errorf("file comment failed: %#v", err)
	}

	b.Log.Debugf("uploading file: %s %s", fi.Name, mtype)

	var res *matrix.RespMediaUpload

	err = b.retry(func() error {
		res, err = b.mc.UploadToContentRepo(content, mtype, int64(len(*fi.Data)))

		return err
	})

	if err != nil {
		b.Log.Errorf("file upload failed: %#v", err)
		return
	}

	switch {
	case strings.Contains(mtype, "video"):
		b.Log.Debugf("sendVideo %s", res.ContentURI)
		err = b.retry(func() error {
			_, err = b.mc.SendVideo(channel, fi.Name, res.ContentURI)

			return err
		})
		if err != nil {
			b.Log.Errorf("sendVideo failed: %#v", err)
		}
	case strings.Contains(mtype, "image"):
		b.Log.Debugf("sendImage %s", res.ContentURI)
		err = b.retry(func() error {
			_, err = b.mc.SendImage(channel, fi.Name, res.ContentURI)

			return err
		})
		if err != nil {
			b.Log.Errorf("sendImage failed: %#v", err)
		}
	case strings.Contains(mtype, "audio"):
		b.Log.Debugf("sendAudio %s", res.ContentURI)
		err = b.retry(func() error {
			_, err = b.mc.SendMessageEvent(channel, "m.room.message", matrix.AudioMessage{
				MsgType: "m.audio",
				Body:    fi.Name,
				URL:     res.ContentURI,
				Info: matrix.AudioInfo{
					Mimetype: mtype,
					Size:     uint(len(*fi.Data)),
				},
			})

			return err
		})
		if err != nil {
			b.Log.Errorf("sendAudio failed: %#v", err)
		}
	default:
		b.Log.Debugf("sendFile %s", res.ContentURI)
		err = b.retry(func() error {
			_, err = b.mc.SendMessageEvent(channel, "m.room.message", matrix.FileMessage{
				MsgType: "m.file",
				Body:    fi.Name,
				URL:     res.ContentURI,
				Info: matrix.FileInfo{
					Mimetype: mtype,
					Size:     uint(len(*fi.Data)),
				},
			})

			return err
		})
		if err != nil {
			b.Log.Errorf("sendFile failed: %#v", err)
		}
	}
	b.Log.Debugf("result: %#v", res)
}
