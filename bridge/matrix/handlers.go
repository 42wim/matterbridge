package bmatrix

import (
	"bytes"
	"fmt"
	"mime"
	"regexp"
	"strings"

	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
)

// Determines if the event comes from ourselves, in which case we want to ignore it
func (b *Bmatrix) ignoreBridgingEvents(ev *event.Event) bool {
	if ev.Sender == b.UserID {
		return true
	}

	// ignore messages we may have sent via the appservice
	if b.appService != nil {
		if ev.Sender == b.appService.appService.BotClient().UserID {
			return true
		}

		// ignore virtual users messages (we ignore the 'exclusive' field of Namespace for now)
		for _, username := range b.appService.namespaces.usernames {
			if username.MatchString(ev.Sender.String()) {
				return true
			}
		}
	}

	return false
}

//nolint: funlen
func (b *Bmatrix) handleEvent(origin EventOrigin, ev *event.Event) {
	b.RLock()
	channel, ok := b.RoomMap[ev.RoomID]
	b.RUnlock()
	if !ok {
		// we don't know that room yet, that could be a room returned by an
		// application service, but matterbridge doesn't handle those just yet
		b.Log.Debugf("Received event for room %s, not joined yet/not handled", ev.RoomID)

		return
	}

	// This needs to be defined before rejecting bridging events, as we rely on this to cache
	// avatar URLs sent with appService (otherwise we would upload one avatar per message sent
	// across the bridge!).
	// As such, beware! Moving this below the b.ignoreBridgingEvents condiiton would appear to
	// work, but it would also lead to a high file upload rate, until being eventually
	// rate-limited by the homeserver
	if ev.Type == event.StateMember {
		b.handleMemberChange(ev)

		return
	}

	if ev.Type == event.EphemeralEventReceipt {
		// we do not support read receipts across servers, considering that
		// multiple services (e.g. Discord) doesn't expose that information)
		return
	}

	if b.ignoreBridgingEvents(ev) {
		return
	}

	// if we receive appservice events for this room, there is no need to check them with the classical syncer
	if !channel.appService && origin == originAppService {
		channel.appService = true
		b.Lock()
		b.RoomMap[ev.RoomID] = channel
		b.Unlock()
	}

	if ev.Type == event.EphemeralEventTyping {
		typing := ev.Content.AsTyping()
		if len(typing.UserIDs) > 0 {
			//nolint:exhaustruct
			b.Remote <- config.Message{
				Event:   config.EventUserTyping,
				Channel: channel.name,
				Account: b.Account,
			}
		}

		return
	}

	// if we receive messages both via the classical matrix syncer and appserver, prefer appservice and throw away this duplicate event
	if channel.appService && origin != originAppService {
		b.Log.Debugf("Dropping event, should receive it via appservice: %s", ev.ID)

		return
	}

	b.Log.Debugf("== Receiving event: %#v (appService=%t)", ev, origin == originAppService)

	defer (func(ev *event.Event) {
		// not crucial, so no ratelimit check here
		if err := b.mc.MarkRead(ev.RoomID, ev.ID); err != nil {
			b.Log.Errorf("couldn't mark message as read %s", err.Error())
		}
	})(ev)

	// Create our message
	//nolint:exhaustruct
	rmsg := config.Message{
		Username: b.getDisplayName(ev.RoomID, ev.Sender),
		Channel:  channel.name,
		Account:  b.Account,
		UserID:   string(ev.Sender),
		ID:       string(ev.ID),
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

	// Update the displayname on join messages, according to https://spec.matrix.org/v1.3/client-server-api/#events-on-change-of-profile-information
	if member.Membership == event.MembershipJoin {
		b.cacheDisplayName(ev.RoomID, ev.Sender, member.Displayname)
		b.cacheAvatarURL(ev.RoomID, ev.Sender, member.AvatarURL)
	} else if member.Membership == event.MembershipLeave || member.Membership == event.MembershipBan {
		b.UserCache.removeFromCache(ev.RoomID, ev.Sender)
	}
}

func (b *Bmatrix) handleMessage(rmsg config.Message, ev *event.Event) {
	msg := ev.Content.AsMessage()
	if msg == nil {
		b.Log.Errorf("matterbridge don't support this event type: %s", ev.Type.Type)
		b.Log.Debugf("Full event: %#v", ev)

		return
	}

	rmsg.Text = msg.Body

	rmsg.Avatar = b.getAvatarURL(ev.RoomID, ev.Sender)

	//nolint: exhaustive
	switch msg.MsgType {
	case event.MsgEmote:
		// Do we have a /me action
		rmsg.Event = config.EventUserAction
	case event.MsgImage, event.MsgVideo, event.MsgFile:
		// Do we have attachments? (we only allow images, videos or files msgtypes)
		err := b.handleDownloadFile(&rmsg, *msg)
		if err != nil {
			b.Log.Errorf("download failed: %#v", err)
		}
	case event.MsgNotice:
		// Support for IRC NOTICE commands/[matrix] m.notice
		rmsg.Event = config.EventNotice
	default:
		if msg.RelatesTo == nil {
			break
		}

		if msg.RelatesTo.Type == event.RelReplace && msg.NewContent != nil {
			// Is it an edit?
			rmsg.ID = string(msg.RelatesTo.EventID)
			rmsg.Text = msg.NewContent.Body
		} else if msg.RelatesTo.Type == event.RelReference && msg.RelatesTo.InReplyTo != nil {
			// Is it a reply?
			body := msg.Body
			if !b.GetBool("keepquotedreply") {
				for strings.HasPrefix(body, "> ") {
					lineIdx := strings.Index(body, "\n\n")
					if lineIdx == -1 {
						break
					}

					body = body[(lineIdx + 2):]
				}
			}

			rmsg.ParentID = string(msg.RelatesTo.EventID)
			rmsg.Text = body
		}
	}

	b.Log.Debugf("<= Sending message from %s on %s to gateway", ev.Sender, b.Account)
	b.Remote <- rmsg
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
		} else if msg.MsgType == event.MsgImage {
			// just a default .png extension if we don't have mime info
			filename += ".png"
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
	//nolint:exhaustruct
	m := event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          fi.Comment,
		FormattedBody: fi.Comment,
	}

	_, err := b.sendMessageEventWithRetries(channel, m, msg.Username, msg.Avatar)
	if err != nil {
		b.Log.Errorf("file comment failed: %#v", err)
	}

	b.Log.Debugf("uploading file: %s %s", fi.Name, mtype)

	var res *matrix.RespMediaUpload
	//nolint:exhaustruct
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

	//nolint:exhaustruct
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
		//nolint:exhaustruct
		m.Info = &event.FileInfo{
			MimeType: mtype,
			Size:     len(*fi.Data),
		}
	default:
		b.Log.Debugf("sendFile %s", res.ContentURI)

		m.MsgType = event.MsgFile
		//nolint:exhaustruct
		m.Info = &event.FileInfo{
			MimeType: mtype,
			Size:     len(*fi.Data),
		}
	}

	_, err = b.sendMessageEventWithRetries(channel, m, msg.Username, msg.Avatar)
	if err != nil {
		b.Log.Errorf("sending the message referencing the uploaded file failed: %#v", err)
	}
}
