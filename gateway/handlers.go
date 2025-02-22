package gateway

import (
	"crypto/sha1" //nolint:gosec
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway/bridgemap"
)

// handleEventFailure handles failures and reconnects bridges.
func (r *Router) handleEventFailure(msg *config.Message) {
	if msg.Event != config.EventFailure {
		return
	}
	for _, gw := range r.Gateways {
		for _, br := range gw.Bridges {
			if msg.Account == br.Account {
				go gw.reconnectBridge(br)
				return
			}
		}
	}
}

// handleEventGetChannelMembers handles channel members
func (r *Router) handleEventGetChannelMembers(msg *config.Message) {
	if msg.Event != config.EventGetChannelMembers {
		return
	}
	for _, gw := range r.Gateways {
		for _, br := range gw.Bridges {
			if msg.Account == br.Account {
				cMembers := msg.Extra[config.EventGetChannelMembers][0].(config.ChannelMembers)
				r.logger.Debugf("Syncing channelmembers from %s", msg.Account)
				br.SetChannelMembers(&cMembers)
				return
			}
		}
	}
}

// handleEventRejoinChannels handles rejoining of channels.
func (r *Router) handleEventRejoinChannels(msg *config.Message) {
	if msg.Event != config.EventRejoinChannels {
		return
	}
	for _, gw := range r.Gateways {
		for _, br := range gw.Bridges {
			if msg.Account == br.Account {
				br.Joined = make(map[string]bool)
				if err := br.JoinChannels(); err != nil {
					r.logger.Errorf("channel join failed for %s: %s", msg.Account, err)
				}
			}
		}
	}
}

// handleFiles uploads or places all files on the given msg to the MediaServer and
// adds the new URL of the file on the MediaServer onto the given msg.
func (gw *Gateway) handleFiles(msg *config.Message) {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")

	// If we don't have a attachfield or we don't have a mediaserver configured return
	if msg.Extra == nil ||
		gw.mediaServer == nil {
		return
	}

	// If we don't have files, nothing to upload.
	if len(msg.Extra["file"]) == 0 {
		return
	}

	for i, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		ext := filepath.Ext(fi.Name)
		fi.Name = fi.Name[0 : len(fi.Name)-len(ext)]
		fi.Name = reg.ReplaceAllString(fi.Name, "_")
		fi.Name += ext

		sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))[:8] //nolint:gosec

		downloadURL, err := gw.mediaServer.handleFilesUpload(&fi)
		if err != nil {
			gw.logger.Error(err)
			continue
		}

		// Download URL.
		gw.logger.Debugf("mediaserver download URL = %s", downloadURL)

		// We uploaded/placed the file successfully. Add the SHA and URL.
		extra := msg.Extra["file"][i].(config.FileInfo)
		extra.URL = downloadURL
		extra.SHA = sha1sum
		msg.Extra["file"][i] = extra
	}
}

// ignoreEvent returns true if we need to ignore this event for the specified destination bridge.
func (gw *Gateway) ignoreEvent(event string, dest *bridge.Bridge) bool {
	switch event {
	case config.EventAvatarDownload:
		// Avatar downloads are only relevant for telegram and mattermost for now
		if dest.Protocol != "mattermost" && dest.Protocol != "telegram" && dest.Protocol != "xmpp" {
			return true
		}
	case config.EventJoinLeave:
		// only relay join/part when configured
		if !dest.GetBool("ShowJoinPart") {
			return true
		}
	case config.EventTopicChange:
		// only relay topic change when used in some way on other side
		if !dest.GetBool("ShowTopicChange") && !dest.GetBool("SyncTopic") {
			return true
		}
	}
	return false
}

// handleMessage makes sure the message get sent to the correct bridge/channels.
// Returns an array of msg ID's
func (gw *Gateway) handleMessage(rmsg *config.Message, dest *bridge.Bridge) []*BrMsgID {
	var brMsgIDs []*BrMsgID

	// Not all bridges support "user is typing" indications so skip the message
	// if the targeted bridge does not support it.
	if rmsg.Event == config.EventUserTyping {
		if _, ok := bridgemap.UserTypingSupport[dest.Protocol]; !ok {
			return nil
		}
	}

	// if we have an attached file, or other info
	if rmsg.Extra != nil && len(rmsg.Extra[config.EventFileFailureSize]) != 0 && rmsg.Text == "" {
		return brMsgIDs
	}

	if gw.ignoreEvent(rmsg.Event, dest) {
		return brMsgIDs
	}

	// broadcast to every out channel (irc QUIT)
	if rmsg.Channel == "" && rmsg.Event != config.EventJoinLeave {
		gw.logger.Debug("empty channel")
		return brMsgIDs
	}

	// Get the ID of the parent message in thread
	var canonicalParentMsgID string
	if rmsg.ParentID != "" && dest.GetBool("PreserveThreading") {
		canonicalParentMsgID = gw.FindCanonicalMsgID(rmsg.Protocol, rmsg.ParentID)
	}

	channels := gw.getDestChannel(rmsg, *dest)
	for idx := range channels {
		channel := &channels[idx]
		msgID, err := gw.SendMessage(rmsg, dest, channel, canonicalParentMsgID)
		if err != nil {
			gw.logger.Errorf("SendMessage failed: %s", err)
			continue
		}
		if msgID == "" {
			continue
		}
		brMsgIDs = append(brMsgIDs, &BrMsgID{dest, dest.Protocol + " " + msgID, channel.ID})
	}
	return brMsgIDs
}

func (gw *Gateway) handleExtractNicks(msg *config.Message) {
	var err error
	br := gw.Bridges[msg.Account]
	for _, outer := range br.GetStringSlice2D("ExtractNicks") {
		search := outer[0]
		replace := outer[1]
		msg.Username, msg.Text, err = extractNick(search, replace, msg.Username, msg.Text)
		if err != nil {
			gw.logger.Errorf("regexp in %s failed: %s", msg.Account, err)
			break
		}
	}
}

// extractNick searches for a username (based on "search" a regular expression).
// if this matches it extracts a nick (based on "extract" another regular expression) from text
// and replaces username with this result.
// returns error if the regexp doesn't compile.
func extractNick(search, extract, username, text string) (string, string, error) {
	re, err := regexp.Compile(search)
	if err != nil {
		return username, text, err
	}
	if re.MatchString(username) {
		re, err = regexp.Compile(extract)
		if err != nil {
			return username, text, err
		}
		res := re.FindAllStringSubmatch(text, 1)
		// only replace if we have exactly 1 match
		if len(res) > 0 && len(res[0]) == 2 {
			username = res[0][1]
			text = strings.Replace(text, res[0][0], "", 1)
		}
	}
	return username, text, nil
}
