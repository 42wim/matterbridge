package bslack

import (
	"fmt"
	"html"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/nlopes/slack"
)

func (b *Bslack) handleSlack() {
	messages := make(chan *config.Message)
	if b.GetString(incomingWebhookConfig) != "" {
		b.Log.Debugf("Choosing webhooks based receiving")
		go b.handleMatterHook(messages)
	} else {
		b.Log.Debugf("Choosing token based receiving")
		go b.handleSlackClient(messages)
	}
	time.Sleep(time.Second)
	b.Log.Debug("Start listening for Slack messages")
	for message := range messages {
		if message.Event != config.EVENT_USER_TYPING {
			b.Log.Debugf("<= Sending message from %s on %s to gateway", message.Username, b.Account)
		}

		// cleanup the message
		message.Text = b.replaceMention(message.Text)
		message.Text = b.replaceVariable(message.Text)
		message.Text = b.replaceChannel(message.Text)
		message.Text = b.replaceURL(message.Text)
		message.Text = html.UnescapeString(message.Text)

		// Add the avatar
		message.Avatar = b.getAvatar(message.UserID)

		b.Log.Debugf("<= Message is %#v", message)
		b.Remote <- *message
	}
}

func (b *Bslack) handleSlackClient(messages chan *config.Message) {
	for msg := range b.rtm.IncomingEvents {
		if msg.Type != sUserTyping && msg.Type != sLatencyReport {
			b.Log.Debugf("== Receiving event %#v", msg.Data)
		}
		switch ev := msg.Data.(type) {
		case *slack.UserTypingEvent:
			if !b.GetBool("ShowUserTyping") {
				continue
			}
			rmsg, err := b.handleTypingEvent(ev)
			if err != nil {
				b.Log.Errorf("%#v", err)
				continue
			}

			messages <- rmsg
		case *slack.MessageEvent:
			if b.skipMessageEvent(ev) {
				b.Log.Debugf("Skipped message: %#v", ev)
				continue
			}
			rmsg, err := b.handleMessageEvent(ev)
			if err != nil {
				b.Log.Errorf("%#v", err)
				continue
			}
			messages <- rmsg
		case *slack.OutgoingErrorEvent:
			b.Log.Debugf("%#v", ev.Error())
		case *slack.ChannelJoinedEvent:
			// When we join a channel we update the full list of users as
			// well as the information for the channel that we joined as this
			// should now tell that we are a member of it.
			b.populateUsers()

			b.channelsMutex.Lock()
			b.channelsByID[ev.Channel.ID] = &ev.Channel
			b.channelsByName[ev.Channel.Name] = &ev.Channel
			b.channelsMutex.Unlock()
		case *slack.ConnectedEvent:
			b.si = ev.Info
			b.populateChannels()
			b.populateUsers()
		case *slack.InvalidAuthEvent:
			b.Log.Fatalf("Invalid Token %#v", ev)
		case *slack.ConnectionErrorEvent:
			b.Log.Errorf("Connection failed %#v %#v", ev.Error(), ev.ErrorObj)
		default:
		}
	}
}

func (b *Bslack) handleMatterHook(messages chan *config.Message) {
	for {
		message := b.mh.Receive()
		b.Log.Debugf("receiving from matterhook (slack) %#v", message)
		if message.UserName == "slackbot" {
			continue
		}
		messages <- &config.Message{
			Username: message.UserName,
			Text:     message.Text,
			Channel:  message.ChannelName,
		}
	}
}

// skipMessageEvent skips event that need to be skipped :-)
func (b *Bslack) skipMessageEvent(ev *slack.MessageEvent) bool {
	switch ev.SubType {
	case sChannelLeave, sChannelJoin:
		return b.GetBool(noSendJoinConfig)
	case sPinnedItem, sUnpinnedItem:
		return true
	}

	// Skip any messages that we made ourselves or from 'slackbot' (see #527).
	if ev.Username == sSlackBotUser ||
		(b.rtm != nil && ev.Username == b.si.User.Name) ||
		(len(ev.Attachments) > 0 && ev.Attachments[0].CallbackID == "matterbridge_"+b.uuid) {
		return true
	}

	// It seems ev.SubMessage.Edited == nil when slack unfurls.
	// Do not forward these messages. See Github issue #266.
	if ev.SubMessage != nil &&
		ev.SubMessage.ThreadTimestamp != ev.SubMessage.Timestamp &&
		ev.SubMessage.Edited == nil {
		return true
	}

	return b.filesCached(ev.Files)
}

func (b *Bslack) filesCached(files []slack.File) bool {
    for _, f := range files {
        f := f
        if !b.fileCached(&f) {
             return false
        }
    }
    return true
}

// handleMessageEvent handles the message events. Together with any called sub-methods,
// this method implements the following event processing pipeline:
//
// 1. Check if the message should be ignored.
//    NOTE: This is not actually part of the method below but is done just before it
//          is called via the 'skipMessageEvent()' method.
// 2. Populate the Matterbridge message that will be sent to the router based on the
//    received event and logic that is common to all events that are not skipped.
// 3. Detect and handle any message that is "status" related (think join channel, etc.).
//    This might result in an early exit from the pipeline and passing of the
//    pre-populated message to the Matterbridge router.
// 4. Handle the specific case of messages that edit existing messages depending on
//    configuration.
// 5. Handle any attachments of the received event.
// 6. Check that the Matterbridge message that we end up with after at the end of the
//    pipeline is valid before sending it to the Matterbridge router.
func (b *Bslack) handleMessageEvent(ev *slack.MessageEvent) (*config.Message, error) {
	rmsg, err := b.populateReceivedMessage(ev)
	if err != nil {
		return nil, err
	}

	// Handle some message types early.
	if b.handleStatusEvent(ev, rmsg) {
		return rmsg, nil
	}

	b.handleAttachments(ev, rmsg)

	// Verify that we have the right information and the message
	// is well-formed before sending it out to the router.
	if len(ev.Files) == 0 && (rmsg.Text == "" || rmsg.Username == "") {
		if ev.BotID != "" {
			// This is probably a webhook we couldn't resolve.
			return nil, fmt.Errorf("message handling resulted in an empty bot message (probably an incoming webhook we couldn't resolve): %#v", ev)
		}
		return nil, fmt.Errorf("message handling resulted in an empty message: %#v", ev)
	}
	return rmsg, nil
}

func (b *Bslack) handleStatusEvent(ev *slack.MessageEvent, rmsg *config.Message) bool {
	switch ev.SubType {
	case sChannelJoined, sMemberJoined:
		b.populateUsers()
		// There's no further processing needed on channel events
		// so we return 'true'.
		return true
	case sChannelJoin, sChannelLeave:
		rmsg.Username = sSystemUser
		rmsg.Event = config.EVENT_JOIN_LEAVE
	case sChannelTopic, sChannelPurpose:
		rmsg.Event = config.EVENT_TOPIC_CHANGE
	case sMessageDeleted:
		rmsg.Text = config.EVENT_MSG_DELETE
		rmsg.Event = config.EVENT_MSG_DELETE
		rmsg.ID = "slack " + ev.DeletedTimestamp
		// If a message is being deleted we do not need to process
		// the event any further so we return 'true'.
		return true
	case sMeMessage:
		rmsg.Event = config.EVENT_USER_ACTION
	}
	return false
}

func (b *Bslack) handleAttachments(ev *slack.MessageEvent, rmsg *config.Message) {
	// File comments are set by the system (because there is no username given).
	if ev.SubType == sFileComment {
		rmsg.Username = sSystemUser
	}

	// See if we have some text in the attachments.
	if rmsg.Text == "" {
		for _, attach := range ev.Attachments {
			if attach.Text != "" {
				if attach.Title != "" {
					rmsg.Text = attach.Title + "\n"
				}
				rmsg.Text += attach.Text
			} else {
				rmsg.Text = attach.Fallback
			}
		}
	}

	// Save the attachments, so that we can send them to other slack (compatible) bridges.
	if len(ev.Attachments) > 0 {
		rmsg.Extra[sSlackAttachment] = append(rmsg.Extra[sSlackAttachment], ev.Attachments)
	}

	// If we have files attached, download them (in memory) and put a pointer to it in msg.Extra.
	for i := range ev.Files {
		if err := b.handleDownloadFile(rmsg, &ev.Files[i]); err != nil {
			b.Log.Errorf("Could not download incoming file: %#v", err)
		}
	}
}

func (b *Bslack) handleTypingEvent(ev *slack.UserTypingEvent) (*config.Message, error) {
	channelInfo, err := b.getChannelByID(ev.Channel)
	if err != nil {
		return nil, err
	}
	return &config.Message{
		Channel: channelInfo.Name,
		Account: b.Account,
		Event:   config.EVENT_USER_TYPING,
	}, nil
}

// handleDownloadFile handles file download
func (b *Bslack) handleDownloadFile(rmsg *config.Message, file *slack.File) error {
	if b.fileCached(file) {
		return nil
	}
	// Check that the file is neither too large nor blacklisted.
	if err := helper.HandleDownloadSize(b.Log, rmsg, file.Name, int64(file.Size), b.General); err != nil {
		b.Log.WithError(err).Infof("Skipping download of incoming file.")
		return nil
	}

	// Actually download the file.
	data, err := helper.DownloadFileAuth(file.URLPrivateDownload, "Bearer "+b.GetString(tokenConfig))
	if err != nil {
		return fmt.Errorf("download %s failed %#v", file.URLPrivateDownload, err)
	}

	// If a comment is attached to the file(s) it is in the 'Text' field of the Slack messge event
	// and should be added as comment to only one of the files. We reset the 'Text' field to ensure
	// that the comment is not duplicated.
	comment := rmsg.Text
	rmsg.Text = ""
	helper.HandleDownloadData(b.Log, rmsg, file.Name, comment, file.URLPrivateDownload, data, b.General)
	return nil
}

// fileCached implements Matterbridge's caching logic for files
// shared via Slack.
//
// We consider that a file was cached if its ID was added in the last minute or
// it's name was registered in the last 10 seconds. This ensures that an
// identically named file but with different content will be uploaded correctly
// (the assumption is that such name collisions will not occur within the given
// timeframes).
func (b *Bslack) fileCached(file *slack.File) bool {
	if ts, ok := b.cache.Get("file" + file.ID); ok && time.Since(ts.(time.Time)) < time.Minute {
		return true
	} else if ts, ok = b.cache.Get("filename" + file.Name); ok && time.Since(ts.(time.Time)) < 10*time.Second {
		return true
	}
	return false
}
