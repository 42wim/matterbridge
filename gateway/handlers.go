package gateway

import (
	"bytes"
	"crypto/sha1" //nolint:gosec
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
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
				flog.Debugf("Syncing channelmembers from %s", msg.Account)
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
					flog.Errorf("channel join failed for %s: %s", msg.Account, err)
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
		(gw.BridgeValues().General.MediaServerUpload == "" &&
			gw.BridgeValues().General.MediaDownloadPath == "") {
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

		if gw.BridgeValues().General.MediaServerUpload != "" {
			// Use MediaServerUpload. Upload using a PUT HTTP request and basicauth.
			if err := gw.handleFilesUpload(&fi); err != nil {
				flog.Error(err)
				continue
			}
		} else {
			// Use MediaServerPath. Place the file on the current filesystem.
			if err := gw.handleFilesLocal(&fi); err != nil {
				flog.Error(err)
				continue
			}
		}

		// Download URL.
		durl := gw.BridgeValues().General.MediaServerDownload + "/" + sha1sum + "/" + fi.Name

		flog.Debugf("mediaserver download URL = %s", durl)

		// We uploaded/placed the file successfully. Add the SHA and URL.
		extra := msg.Extra["file"][i].(config.FileInfo)
		extra.URL = durl
		extra.SHA = sha1sum
		msg.Extra["file"][i] = extra
	}
}

// handleFilesUpload uses MediaServerUpload configuration to upload the file.
// Returns error on failure.
func (gw *Gateway) handleFilesUpload(fi *config.FileInfo) error {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	// Use MediaServerUpload. Upload using a PUT HTTP request and basicauth.
	sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))[:8] //nolint:gosec
	url := gw.BridgeValues().General.MediaServerUpload + "/" + sha1sum + "/" + fi.Name

	req, err := http.NewRequest("PUT", url, bytes.NewReader(*fi.Data))
	if err != nil {
		return fmt.Errorf("mediaserver upload failed, could not create request: %#v", err)
	}

	flog.Debugf("mediaserver upload url: %s", url)

	req.Header.Set("Content-Type", "binary/octet-stream")
	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("mediaserver upload failed, could not Do request: %#v", err)
	}
	return nil
}

// handleFilesLocal use MediaServerPath configuration, places the file on the current filesystem.
// Returns error on failure.
func (gw *Gateway) handleFilesLocal(fi *config.FileInfo) error {
	sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))[:8] //nolint:gosec
	dir := gw.BridgeValues().General.MediaDownloadPath + "/" + sha1sum
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("mediaserver path failed, could not mkdir: %s %#v", err, err)
	}

	path := dir + "/" + fi.Name
	flog.Debugf("mediaserver path placing file: %s", path)

	err = ioutil.WriteFile(path, *fi.Data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("mediaserver path failed, could not writefile: %s %#v", err, err)
	}
	return nil
}

// ignoreEvent returns true if we need to ignore this event for the specified destination bridge.
func (gw *Gateway) ignoreEvent(event string, dest *bridge.Bridge) bool {
	switch event {
	case config.EventAvatarDownload:
		// Avatar downloads are only relevant for telegram and mattermost for now
		if dest.Protocol != "mattermost" && dest.Protocol != "telegram" {
			return true
		}
	case config.EventJoinLeave:
		// only relay join/part when configured
		if !dest.GetBool("ShowJoinPart") {
			return true
		}
	case config.EventTopicChange:
		// only relay topic change when used in some way on other side
		if dest.GetBool("ShowTopicChange") && dest.GetBool("SyncTopic") {
			return true
		}
	}
	return false
}

// handleMessage makes sure the message get sent to the correct bridge/channels.
// Returns an array of msg ID's
func (gw *Gateway) handleMessage(msg config.Message, dest *bridge.Bridge) []*BrMsgID {
	var brMsgIDs []*BrMsgID

	// if we have an attached file, or other info
	if msg.Extra != nil && len(msg.Extra[config.EventFileFailureSize]) != 0 && msg.Text == "" {
		return brMsgIDs
	}

	if gw.ignoreEvent(msg.Event, dest) {
		return brMsgIDs
	}

	// broadcast to every out channel (irc QUIT)
	if msg.Channel == "" && msg.Event != config.EventJoinLeave {
		flog.Debug("empty channel")
		return brMsgIDs
	}

	// Get the ID of the parent message in thread
	var canonicalParentMsgID string
	if msg.ParentID != "" && dest.GetBool("PreserveThreading") {
		canonicalParentMsgID = gw.FindCanonicalMsgID(msg.Protocol, msg.ParentID)
	}

	origmsg := msg
	channels := gw.getDestChannel(&msg, *dest)
	for _, channel := range channels {
		msgID, err := gw.SendMessage(origmsg, dest, channel, canonicalParentMsgID)
		if err != nil {
			flog.Errorf("SendMessage failed: %s", err)
			continue
		}
		if msgID == "" {
			continue
		}
		brMsgIDs = append(brMsgIDs, &BrMsgID{dest, dest.Protocol + " " + msgID, channel.ID})
	}
	return brMsgIDs
}
