//nolint: exhaustivestruct
package bmatrix

import (
	"bytes"
	"fmt"
	"mime"
	"regexp"
	"strings"
	"sync"
	"time"

	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/appservice"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
)

var (
	htmlTag            = regexp.MustCompile("</.*?>")
	htmlReplacementTag = regexp.MustCompile("<[^>]*>")
)

type EventOrigin int

const (
	originClassicSyncer EventOrigin = iota
	originAppService
)

type NicknameCacheEntry struct {
	displayName string
	lastUpdated time.Time
}

type RoomInfo struct {
	name       string
	appService bool
}

type Bmatrix struct {
	mc          *matrix.Client
	appService  *appservice.AppService
	UserID      id.UserID
	NicknameMap map[id.RoomID]map[id.UserID]NicknameCacheEntry
	RoomMap     map[id.RoomID]RoomInfo
	rateMutex   sync.RWMutex
	sync.RWMutex
	*bridge.Config
	stop    chan struct{}
	stopAck chan struct{}
}

type matrixUsername struct {
	plain     string
	formatted string
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmatrix{Config: cfg}
	b.RoomMap = make(map[id.RoomID]RoomInfo)
	b.NicknameMap = make(map[id.RoomID]map[id.UserID]NicknameCacheEntry)
	b.stop = make(chan struct{}, 2)
	b.stopAck = make(chan struct{}, 2)
	return b
}

func (b *Bmatrix) Connect() error {
	var err error
	b.Log.Infof("Connecting %s", b.GetString("Server"))
	if b.GetString("MxID") != "" && b.GetString("Token") != "" {
		b.UserID = id.UserID(b.GetString("MxID"))
		b.mc, err = matrix.NewClient(
			b.GetString("Server"), b.UserID, b.GetString("Token"),
		)
		if err != nil {
			return err
		}
		b.Log.Info("Using existing Matrix credentials")
	} else {
		b.mc, err = matrix.NewClient(b.GetString("Server"), "", "")
		if err != nil {
			return err
		}
		resp, err := b.mc.Login(&matrix.ReqLogin{
			Type:             matrix.AuthTypePassword,
			Password:         b.GetString("Password"),
			Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: b.GetString("Login")},
			StoreCredentials: true,
		})
		if err != nil {
			return err
		}
		b.UserID = resp.UserID
		b.Log.Info("Connection succeeded")
	}

	go b.handlematrix()

	if b.GetBool("UseAppService") {
		err := b.startAppService()
		if err != nil {
			b.Log.Errorf("couldn't start the application service: %#v", err)

			return err
		}
	}

	return nil
}

func (b *Bmatrix) Disconnect() error {
	// tell the Sync() loop to exit
	b.stop <- struct{}{}
	if b.appService != nil {
		b.stop <- struct{}{}
	}
	b.mc.StopSync()

	// wait for both the syncer and the appservice to terminate
	<-b.stopAck
	if b.appService != nil {
		<-b.stopAck
	}

	return nil
}

func (b *Bmatrix) JoinChannel(channel config.ChannelInfo) error {
	return b.retry(func() error {
		resp, err := b.mc.JoinRoom(channel.Name, "", nil)
		if err != nil {
			return err
		}

		b.Lock()
		b.RoomMap[resp.RoomID] = RoomInfo{name: channel.Name, appService: false}
		b.Unlock()

		return nil
	})
}

func (b *Bmatrix) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	channel := b.getRoomID(msg.Channel)
	if channel == "" {
		return "", fmt.Errorf("got message for unknown channel '%s'", msg.Channel)
	}
	b.Log.Debugf("Channel %s maps to channel id %s", msg.Channel, channel)

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		m := event.MessageEventContent{
			MsgType:       event.MsgEmote,
			Body:          msg.Text,
			FormattedBody: msg.Text,
		}

		return b.sendMessageEventWithRetries(channel, m, msg)
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}

		msgID := ""

		err := b.retry(func() error {
			resp, err := b.mc.RedactEvent(channel, id.EventID(msg.ID), matrix.ReqRedact{})
			if err != nil {
				return err
			}

			msgID = string(resp.EventID)

			return err
		})

		return msgID, err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			m := event.MessageEventContent{
				MsgType: event.MsgText,
				Body:    rmsg.Text,
			}

			_, err := b.sendMessageEventWithRetries(channel, m, msg)
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
		rmsg := event.MessageEventContent{
			MsgType: event.MsgText,
			Body:    msg.Text,
		}
		if b.GetBool("HTMLDisable") {
			rmsg.FormattedBody = "* " + msg.Text
		} else {
			rmsg.Format = event.FormatHTML
			rmsg.FormattedBody = "* " + helper.ParseMarkdown(msg.Text)
		}
		rmsg.NewContent = &event.MessageEventContent{
			Body:    rmsg.Body,
			MsgType: event.MsgText,
		}
		rmsg.RelatesTo = &event.RelatesTo{
			EventID: id.EventID(msg.ID),
			Type:    event.RelReplace,
		}

		return b.sendMessageEventWithRetries(channel, rmsg, msg)
	}

	m := event.MessageEventContent{
		Body:          msg.Text,
		FormattedBody: msg.Text,
	}

	// Use notices to send join/leave events
	if msg.Event == config.EventJoinLeave {
		m.MsgType = event.MsgNotice
	} else {
		m.MsgType = event.MsgText
		if b.GetBool("HTMLDisable") {
			m.FormattedBody = ""
		} else {
			m.FormattedBody = helper.ParseMarkdown(msg.Text)
		}
	}

	return b.sendMessageEventWithRetries(channel, m, msg)
}

func (b *Bmatrix) handlematrix() {
	syncer, ok := b.mc.Syncer.(*matrix.DefaultSyncer)
	if !ok {
		b.Log.Errorln("couldn't convert the Syncer object to a DefaultSyncer structure, the matrix bridge won't work")

		return
	}

	for _, evType := range []event.Type{event.EventRedaction, event.EventMessage, event.StateMember} {
		syncer.OnEventType(evType, func(source matrix.EventSource, ev *event.Event) {
			b.handleEvent(originClassicSyncer, ev)
		})
	}

	go func() {
		for {
			select {
			case <-b.stop:
				b.stopAck <- struct{}{}

				return
			default:
				if err := b.mc.Sync(); err != nil {
					b.Log.Println("Sync() returned ", err)
				}
			}
		}
	}()
}

// Determines if the event comes from ourselves, in which case we want to ignore it
func (b *Bmatrix) ignoreBridgingEvents(ev *event.Event) bool {
	if ev.Sender == b.UserID {
		return true
	}

	// ignore messages we may have sent via the appservice
	if b.appService != nil {
		if ev.Sender == b.appService.BotClient().UserID {
			return true
		}

		// ignore virtual users messages (we ignore the 'exclusive' field of Namespace for now)
		for _, namespace := range b.appService.Registration.Namespaces.UserIDs {
			match, err := regexp.MatchString(namespace.Regex, ev.Sender.String())
			if match && err == nil {
				return true
			}
		}
	}

	return false
}

//nolint: funlen
func (b *Bmatrix) handleEvent(origin EventOrigin, ev *event.Event) {
	b.Log.Debugf("== Receiving event: %#v", ev)

	if b.ignoreBridgingEvents(ev) {
		return
	}

	if ev.Type == event.StateMember {
		b.handleMemberChange(ev)

		return
	}

	b.RLock()
	channel, ok := b.RoomMap[ev.RoomID]
	b.RUnlock()
	if !ok {
		// we don't know that room yet, that must be a room returned by an application service,
		// but matterbridge doesn't handle those just yet
		b.Log.Debugf("Unknown room %s", ev.RoomID)

		return
	}

	// if we receive appservice events for this room, there is no need to check them with the classical syncer
	if !channel.appService && origin == originAppService {
		channel.appService = true
		b.Lock()
		b.RoomMap[ev.RoomID] = channel
		b.Unlock()
	}

	// if we receive messages both via the classical matrix syncer and appserver, prefer appservice and throw away this duplicate event
	if channel.appService && origin != originAppService {
		b.Log.Debugf("Dropping event, should receive it via appservice: %s", ev.ID)

		return
	}

	// Create our message
	rmsg := config.Message{
		Username: b.getDisplayName(ev.RoomID, ev.Sender),
		Channel:  channel.name,
		Account:  b.Account,
		UserID:   string(ev.Sender),
		ID:       string(ev.ID),
		Avatar:   b.getAvatarURL(ev.Sender),
	}

	// Remove homeserver suffix if configured
	if b.GetBool("NoHomeServerSuffix") {
		re := regexp.MustCompile("(.*?):.*")
		rmsg.Username = re.ReplaceAllString(rmsg.Username, `$1`)
	}

	// Delete event
	if ev.Type == event.EventRedaction {
		rmsg.Event = config.EventMsgDelete
		rmsg.ID = string(ev.Redacts)
		rmsg.Text = config.EventMsgDelete
		b.Remote <- rmsg

		return
	}

	b.handleMessage(rmsg, ev)
}

func (b *Bmatrix) handleMemberChange(ev *event.Event) {
	member := ev.Content.AsMember()
	if member == nil {
		b.Log.Errorf("Couldn't process a member event:\n%#v", ev)

		return
	}
	// Update the displayname on join messages, according to https://matrix.org/docs/spec/client_server/r0.6.1#events-on-change-of-profile-information
	if member.Membership == event.MembershipJoin {
		b.cacheDisplayName(ev.RoomID, ev.Sender, member.Displayname)
	}
}

//nolint: cyclop
func (b *Bmatrix) handleMessage(rmsg config.Message, ev *event.Event) {
	msg := ev.Content.AsMessage()
	if msg == nil {
		b.Log.Errorf("matterbridge don't support this event type: %s", ev.Type.Type)
		b.Log.Debugf("Full event: %#v", ev)

		return
	}

	rmsg.Text = msg.Body

	// Do we have a /me action
	if msg.MsgType == event.MsgEmote {
		rmsg.Event = config.EventUserAction
	}

	// Is it an edit?
	if msg.RelatesTo != nil && msg.NewContent != nil && msg.RelatesTo.Type == event.RelReplace {
		rmsg.ID = string(msg.RelatesTo.EventID)
		rmsg.Text = msg.NewContent.Body
		b.Remote <- rmsg

		return
	}

	// Do we have attachments (we only allow image,video or file msgtypes)
	if msg.MsgType == event.MsgImage || msg.MsgType == event.MsgVideo || msg.MsgType == event.MsgFile {
		err := b.handleDownloadFile(&rmsg, *msg)
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

// handleDownloadFile handles file download
func (b *Bmatrix) handleDownloadFile(rmsg *config.Message, msg event.MessageEventContent) error {
	rmsg.Extra = make(map[string][]interface{})
	if msg.URL == "" || msg.Info == nil {
		b.Log.Error("couldn't download a file with no URL or no file informations (invalid event ?)")
		b.Log.Debugf("Full Message content:\n%#v", msg)
	}

	url := strings.ReplaceAll(string(msg.URL), "mxc://", b.GetString("Server")+"/_matrix/media/v1/download/")
	filename := msg.Body

	// check if we have an image uploaded without extension
	if !strings.Contains(filename, ".") {
		mext, _ := mime.ExtensionsByType(msg.Info.MimeType)
		if len(mext) > 0 {
			filename += mext[0]
		} else {
			if msg.MsgType == event.MsgImage {
				// just a default .png extension if we don't have mime info
				filename += ".png"
			}
		}
	}

	// check if the size is ok
	err := helper.HandleDownloadSize(b.Log, rmsg, filename, int64(msg.Info.Size), b.General)
	if err != nil {
		return err
	}
	// actually download the file
	data, err := helper.DownloadFile(url)
	if err != nil {
		return fmt.Errorf("download %s failed %#v", url, err)
	}
	// add the downloaded data to the message
	helper.HandleDownloadData(b.Log, rmsg, filename, "", url, data, b.General)
	return nil
}

// handleUploadFiles handles native upload of files.
func (b *Bmatrix) handleUploadFiles(msg *config.Message, channel id.RoomID) (string, error) {
	for _, f := range msg.Extra["file"] {
		if fi, ok := f.(config.FileInfo); ok {
			b.handleUploadFile(msg, channel, &fi)
		}
	}
	return "", nil
}

// handleUploadFile handles native upload of a file.
//nolint: funlen
func (b *Bmatrix) handleUploadFile(msg *config.Message, channel id.RoomID, fi *config.FileInfo) {
	content := bytes.NewReader(*fi.Data)
	sp := strings.Split(fi.Name, ".")
	mtype := mime.TypeByExtension("." + sp[len(sp)-1])

	// image and video uploads send no username, we have to do this ourself here #715
	m := event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          fi.Comment,
		FormattedBody: fi.Comment,
	}

	_, err := b.sendMessageEventWithRetries(channel, m, *msg)
	if err != nil {
		b.Log.Errorf("file comment failed: %#v", err)
	}

	b.Log.Debugf("uploading file: %s %s", fi.Name, mtype)

	var res *matrix.RespMediaUpload
	req := matrix.ReqUploadMedia{
		Content:       content,
		ContentType:   mtype,
		ContentLength: fi.Size,
	}

	err = b.retry(func() error {
		res, err = b.mc.UploadMedia(req)

		return err
	})

	if err != nil {
		b.Log.Errorf("file upload failed: %#v", err)
		return
	}

	b.Log.Debugf("result: %#v", res)

	m = event.MessageEventContent{
		Body: fi.Name,
		URL:  res.ContentURI.CUString(),
	}

	switch {
	case strings.Contains(mtype, "video"):
		b.Log.Debugf("sendVideo %s", res.ContentURI)

		m.MsgType = event.MsgVideo
	case strings.Contains(mtype, "image"):
		b.Log.Debugf("sendImage %s", res.ContentURI)

		m.MsgType = event.MsgImage
	case strings.Contains(mtype, "audio"):
		b.Log.Debugf("sendAudio %s", res.ContentURI)

		m.MsgType = event.MsgAudio
		m.Info = &event.FileInfo{
			MimeType: mtype,
			Size:     len(*fi.Data),
		}
	default:
		b.Log.Debugf("sendFile %s", res.ContentURI)

		m.MsgType = event.MsgFile
		m.Info = &event.FileInfo{
			MimeType: mtype,
			Size:     len(*fi.Data),
		}
	}

	_, err = b.sendMessageEventWithRetries(channel, m, *msg)
	if err != nil {
		b.Log.Errorf("sending the message referencing the uploaded file failed: %#v", err)
	}
}
