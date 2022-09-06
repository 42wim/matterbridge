//nolint:exhaustivestruct
package bmatrix

import (
	"fmt"
	"regexp"
	"sync"

	matrix "maunium.net/go/mautrix"
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

type RoomInfo struct {
	name       string
	appService bool
}

type Bmatrix struct {
	mc          *matrix.Client
	UserID      id.UserID
	appService  *AppServiceWrapper
	UserCache   *UserInfoCache
	RoomMap     map[id.RoomID]RoomInfo
	rateMutex   sync.RWMutex
	joinedRooms []id.RoomID
	sync.RWMutex
	*bridge.Config
	stopNormalSync    chan struct{}
	stopNormalSyncAck chan struct{}
}

type matrixUsername struct {
	plain     string
	formatted string
}

func New(cfg *bridge.Config) bridge.Bridger {
	b := &Bmatrix{Config: cfg}
	b.RoomMap = make(map[id.RoomID]RoomInfo)
	b.UserCache = NewUserInfoCache()
	b.stopNormalSync = make(chan struct{}, 1)
	b.stopNormalSyncAck = make(chan struct{}, 1)
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
			Identifier:       matrix.UserIdentifier{Type: matrix.IdentifierTypeUser, User: b.GetString("Login")}, //nolint:exhaustruct
			StoreCredentials: true,
		})
		if err != nil {
			return err
		}
		b.UserID = resp.UserID
		b.Log.Info("Connection succeeded")
	}

	b.Log.Debug("Retrieving the list of rooms we have already joined")
	joinedRooms, err := b.mc.JoinedRooms()
	if err != nil {
		b.Log.Errorf("couldn't list the joined rooms")

		return err
	}
	b.joinedRooms = joinedRooms.JoinedRooms
	for _, roomID := range joinedRooms.JoinedRooms {
		// leave the channel name (usually a channel alias - in the matrix sense)
		// unresolved for now, it will be completed when JoinChannel() is called
		b.RoomMap[roomID] = RoomInfo{name: "", appService: false}
	}

	return nil
}

func (b *Bmatrix) Disconnect() error {
	// tell the Sync() loop to exit
	b.stopNormalSync <- struct{}{}
	b.mc.StopSync()

	// wait for both the syncer and the appservice to terminate
	<-b.stopNormalSyncAck
	if b.appService != nil {
		b.appService.stop <- struct{}{}
		<-b.appService.stopAck
	}

	return nil
}

func (b *Bmatrix) JoinChannel(channel config.ChannelInfo) error {
	resolvedAlias, err := b.mc.ResolveAlias(id.RoomAlias(channel.Name))
	if err != nil {
		b.Log.Errorf("couldn't retrieve the room ID for the alias '%s'", channel.Name)

		return err
	}

	roomInfo := RoomInfo{name: channel.Name, appService: false}
	alreadyJoined := false
	for _, roomID := range b.joinedRooms {
		// we have already joined this room (e.g. in a previous execution of matterbridge)
		// => we only update the room alias, but do not attempt to join it again
		if roomID == resolvedAlias.RoomID {
			alreadyJoined = true
			break
		}
	}

	if !alreadyJoined {
		err = b.retry(func() error {
			_, innerErr := b.mc.JoinRoom(channel.Name, "", nil)
			return innerErr
		})

		if err != nil {
			return err
		}
	}

	b.Lock()
	b.RoomMap[resolvedAlias.RoomID] = roomInfo
	b.Unlock()

	return nil
}

func (b *Bmatrix) Start() error {
	// at this point, JoinChannel() has been called on all the channels
	// declared in the configuration, so we can exit every other joined room
	// in order to stop receiving events from rooms we no longer follow
	b.RLock()
	for _, roomID := range b.joinedRooms {
		if _, present := b.RoomMap[roomID]; !present {
			// we deliberately ignore the return value,
			// because the bridge will still work even if we couln't exit the room
			_, _ = b.mc.LeaveRoom(roomID, &matrix.ReqLeave{Reason: "No longer bridged"})
		}
	}
	b.RUnlock()

	go b.handlematrix()

	if b.GetBool("UseAppService") {
		appService, err := b.NewAppService()
		if err != nil {
			b.Log.Errorf("couldn't load the app service configuration: %#v", err)

			return err
		}

		b.appService = appService
		err = b.startAppService()
		if err != nil {
			b.Log.Errorf("couldn't start the application service: %#v", err)

			return err
		}
	}

	return nil
}

//nolint:funlen,gocognit,gocyclo
func (b *Bmatrix) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Sending %#v", msg)

	channel := b.getRoomID(msg.Channel)
	if channel == "" {
		return "", fmt.Errorf("got message for unknown channel '%s'", msg.Channel)
	}

	if msg.Event == config.EventUserTyping && b.GetBool("ShowUserTyping") {
		_, err := b.mc.UserTyping(channel, true, 15000)
		return "", err
	}

	// Make a action /me of the message
	if msg.Event == config.EventUserAction {
		//nolint:exhaustruct
		m := event.MessageEventContent{
			MsgType: event.MsgEmote,
			Body:    msg.Text,
		}

		if !b.GetBool("HTMLDisable") {
			m.FormattedBody = helper.ParseMarkdown(msg.Text)
			m.Format = event.FormatHTML
		}

		return b.sendMessageEventWithRetries(channel, m, msg.Username, msg.Avatar)
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}

		msgID := ""

		err := b.retry(func() error {
			//nolint:exhaustruct
			resp, err := b.mc.RedactEvent(channel, id.EventID(msg.ID), matrix.ReqRedact{})
			if resp != nil {
				msgID = string(resp.EventID)
			}

			return err
		})

		return msgID, err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			//nolint:exhaustruct
			m := event.MessageEventContent{
				MsgType: event.MsgText,
				Body:    rmsg.Text,
			}

			_, err := b.sendMessageEventWithRetries(channel, m, msg.Username, msg.Avatar)
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
		//nolint:exhaustruct
		rmsg := event.MessageEventContent{
			MsgType: event.MsgText,
			Body:    msg.Text,
		}
		//nolint:exhaustruct
		rmsg.NewContent = &event.MessageEventContent{
			Body:    rmsg.Body,
			MsgType: event.MsgText,
		}
		if b.GetBool("HTMLDisable") {
			rmsg.FormattedBody = "* " + msg.Text
		} else {
			rmsg.Format = event.FormatHTML
			rmsg.FormattedBody = "* " + helper.ParseMarkdown(msg.Text)
			rmsg.NewContent.Format = rmsg.Format
			rmsg.NewContent.FormattedBody = rmsg.FormattedBody
		}

		//nolint:exhaustruct
		rmsg.RelatesTo = &event.RelatesTo{
			EventID: id.EventID(msg.ID),
			Type:    event.RelReplace,
		}

		return b.sendMessageEventWithRetries(channel, rmsg, msg.Username, msg.Avatar)
	}

	//nolint:exhaustruct
	m := event.MessageEventContent{
		Body: msg.Text,
	}

	if !b.GetBool("HTMLDisable") {
		m.Format = event.FormatHTML
		m.FormattedBody = msg.Text
	}

	// Use notices to send join/leave events
	if msg.Event == config.EventJoinLeave || msg.Event == config.EventNotice {
		m.MsgType = event.MsgNotice
	} else {
		m.MsgType = event.MsgText
		if b.GetBool("HTMLDisable") {
			m.FormattedBody = ""
		} else {
			m.FormattedBody = helper.ParseMarkdown(msg.Text)
		}

		if msg.ParentValid() {
			m.RelatesTo = &event.RelatesTo{
				EventID: "",
				Type:    event.RelReference,
				InReplyTo: &event.InReplyTo{
					EventID: id.EventID(msg.ParentID),
				},
				Key: "",
			}
		}
	}

	return b.sendMessageEventWithRetries(channel, m, msg.Username, msg.Avatar)
}

// DontProcessOldEvents returns true if a sync event should be considered for further processing.
// We use that function to filter out events we have already read.
//
//nolint:gocognit
func (b *Bmatrix) DontProcessOldEvents(resp *matrix.RespSync, since string) bool {
	// we only filter old events in the initial sync(), because subsequent sync()
	// (where since != "") should only return new events
	if since != "" {
		return true
	}

	for joinedRoom, roomData := range resp.Rooms.Join {
		var readTimestamp int64 = 0
		// retrieve the timestamp of the last read receipt
		// note: we're not sure some events will not be thrown away in this
		// initial sync, as the server may not have received some events yet when
		// the read receipt was sent: there is a mix of timestamps between
		// the read receipt on the target homeserver and the timestamps when
		// events were *created* on the homeserver peers
		for _, evt := range roomData.Ephemeral.Events {
			if evt.Type != event.EphemeralEventReceipt {
				continue
			}

			err := evt.Content.ParseRaw(evt.Type)
			if err != nil {
				b.Log.Warnf("couldn't parse receipt event %#v", evt.Content)
			}
			receipts := *evt.Content.AsReceipt()
			for _, receiptByType := range receipts {
				for _, receiptsByUser := range receiptByType {
					for userID, userReceipt := range receiptsByUser {
						// ignore read receipts of other users
						if userID != b.UserID {
							continue
						}

						readTimestamp = userReceipt.Timestamp.UnixNano()
					}
				}
			}
		}

		newEventList := make([]*event.Event, 0, len(roomData.Timeline.Events))
		for _, evt := range roomData.Timeline.Events {
			// remove old event, except for state changes
			if evt.Timestamp > readTimestamp || evt.Type.Class == event.StateEventType {
				newEventList = append(newEventList, evt)
			}
		}

		roomData.Timeline.Events = newEventList
		resp.Rooms.Join[joinedRoom] = roomData
	}
	return true
}
