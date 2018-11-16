package gateway

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/api"
	"github.com/42wim/matterbridge/bridge/config"
	bdiscord "github.com/42wim/matterbridge/bridge/discord"
	bgitter "github.com/42wim/matterbridge/bridge/gitter"
	birc "github.com/42wim/matterbridge/bridge/irc"
	bmatrix "github.com/42wim/matterbridge/bridge/matrix"
	bmattermost "github.com/42wim/matterbridge/bridge/mattermost"
	brocketchat "github.com/42wim/matterbridge/bridge/rocketchat"
	bslack "github.com/42wim/matterbridge/bridge/slack"
	bsshchat "github.com/42wim/matterbridge/bridge/sshchat"
	bsteam "github.com/42wim/matterbridge/bridge/steam"
	btelegram "github.com/42wim/matterbridge/bridge/telegram"
	bxmpp "github.com/42wim/matterbridge/bridge/xmpp"
	bzulip "github.com/42wim/matterbridge/bridge/zulip"
	"github.com/hashicorp/golang-lru"
	"github.com/peterhellberg/emojilib"
	log "github.com/sirupsen/logrus"
)

type Gateway struct {
	config.Config

	Router         *Router
	MyConfig       *config.Gateway
	Bridges        map[string]*bridge.Bridge
	Channels       map[string]*config.ChannelInfo
	ChannelOptions map[string]config.ChannelOptions
	Message        chan config.Message
	Name           string
	Messages       *lru.Cache
}

type BrMsgID struct {
	br        *bridge.Bridge
	ID        string
	ChannelID string
}

var flog *log.Entry

var bridgeMap = map[string]bridge.Factory{
	"api":          api.New,
	"discord":      bdiscord.New,
	"gitter":       bgitter.New,
	"irc":          birc.New,
	"mattermost":   bmattermost.New,
	"matrix":       bmatrix.New,
	"rocketchat":   brocketchat.New,
	"slack-legacy": bslack.NewLegacy,
	"slack":        bslack.New,
	"sshchat":      bsshchat.New,
	"steam":        bsteam.New,
	"telegram":     btelegram.New,
	"xmpp":         bxmpp.New,
	"zulip":        bzulip.New,
}

const (
	apiProtocol = "api"
)

func New(cfg config.Gateway, r *Router) *Gateway {
	flog = log.WithFields(log.Fields{"prefix": "gateway"})
	gw := &Gateway{Channels: make(map[string]*config.ChannelInfo), Message: r.Message,
		Router: r, Bridges: make(map[string]*bridge.Bridge), Config: r.Config}
	cache, _ := lru.New(5000)
	gw.Messages = cache
	gw.AddConfig(&cfg)
	return gw
}

// Find the canonical ID that the message is keyed under in cache
func (gw *Gateway) FindCanonicalMsgID(protocol string, mID string) string {
	ID := protocol + " " + mID
	if gw.Messages.Contains(ID) {
		return mID
	}

	// If not keyed, iterate through cache for downstream, and infer upstream.
	for _, mid := range gw.Messages.Keys() {
		v, _ := gw.Messages.Peek(mid)
		ids := v.([]*BrMsgID)
		for _, downstreamMsgObj := range ids {
			if ID == downstreamMsgObj.ID {
				return strings.Replace(mid.(string), protocol+" ", "", 1)
			}
		}
	}
	return ""
}

func (gw *Gateway) AddBridge(cfg *config.Bridge) error {
	br := gw.Router.getBridge(cfg.Account)
	if br == nil {
		br = bridge.New(cfg)
		br.Config = gw.Router.Config
		br.General = &gw.BridgeValues().General
		// set logging
		br.Log = log.WithFields(log.Fields{"prefix": "bridge"})
		brconfig := &bridge.Config{Remote: gw.Message, Log: log.WithFields(log.Fields{"prefix": br.Protocol}), Bridge: br}
		// add the actual bridger for this protocol to this bridge using the bridgeMap
		br.Bridger = bridgeMap[br.Protocol](brconfig)
	}
	gw.mapChannelsToBridge(br)
	gw.Bridges[cfg.Account] = br
	return nil
}

func (gw *Gateway) AddConfig(cfg *config.Gateway) error {
	gw.Name = cfg.Name
	gw.MyConfig = cfg
	gw.mapChannels()
	for _, br := range append(gw.MyConfig.In, append(gw.MyConfig.InOut, gw.MyConfig.Out...)...) {
		br := br //scopelint
		err := gw.AddBridge(&br)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gw *Gateway) mapChannelsToBridge(br *bridge.Bridge) {
	for ID, channel := range gw.Channels {
		if br.Account == channel.Account {
			br.Channels[ID] = *channel
		}
	}
}

func (gw *Gateway) reconnectBridge(br *bridge.Bridge) {
	br.Disconnect()
	time.Sleep(time.Second * 5)
RECONNECT:
	flog.Infof("Reconnecting %s", br.Account)
	err := br.Connect()
	if err != nil {
		flog.Errorf("Reconnection failed: %s. Trying again in 60 seconds", err)
		time.Sleep(time.Second * 60)
		goto RECONNECT
	}
	br.Joined = make(map[string]bool)
	br.JoinChannels()
}

func (gw *Gateway) mapChannelConfig(cfg []config.Bridge, direction string) {
	for _, br := range cfg {
		if isAPI(br.Account) {
			br.Channel = apiProtocol
		}
		// make sure to lowercase irc channels in config #348
		if strings.HasPrefix(br.Account, "irc.") {
			br.Channel = strings.ToLower(br.Channel)
		}
		ID := br.Channel + br.Account
		if _, ok := gw.Channels[ID]; !ok {
			channel := &config.ChannelInfo{Name: br.Channel, Direction: direction, ID: ID, Options: br.Options, Account: br.Account,
				SameChannel: make(map[string]bool)}
			channel.SameChannel[gw.Name] = br.SameChannel
			gw.Channels[channel.ID] = channel
		} else {
			// if we already have a key and it's not our current direction it means we have a bidirectional inout
			if gw.Channels[ID].Direction != direction {
				gw.Channels[ID].Direction = "inout"
			}
		}
		gw.Channels[ID].SameChannel[gw.Name] = br.SameChannel
	}
}

func (gw *Gateway) mapChannels() error {
	gw.mapChannelConfig(gw.MyConfig.In, "in")
	gw.mapChannelConfig(gw.MyConfig.Out, "out")
	gw.mapChannelConfig(gw.MyConfig.InOut, "inout")
	return nil
}

func (gw *Gateway) getDestChannel(msg *config.Message, dest bridge.Bridge) []config.ChannelInfo {
	var channels []config.ChannelInfo

	// for messages received from the api check that the gateway is the specified one
	if msg.Protocol == apiProtocol && gw.Name != msg.Gateway {
		return channels
	}

	// if source channel is in only, do nothing
	for _, channel := range gw.Channels {
		// lookup the channel from the message
		if channel.ID == getChannelID(*msg) {
			// we only have destinations if the original message is from an "in" (sending) channel
			if !strings.Contains(channel.Direction, "in") {
				return channels
			}
			continue
		}
	}
	for _, channel := range gw.Channels {
		if _, ok := gw.Channels[getChannelID(*msg)]; !ok {
			continue
		}

		// do samechannelgateway flogic
		if channel.SameChannel[msg.Gateway] {
			if msg.Channel == channel.Name && msg.Account != dest.Account {
				channels = append(channels, *channel)
			}
			continue
		}
		if strings.Contains(channel.Direction, "out") && channel.Account == dest.Account && gw.validGatewayDest(msg) {
			channels = append(channels, *channel)
		}
	}
	return channels
}

func (gw *Gateway) getDestMsgID(msgID string, dest *bridge.Bridge, channel config.ChannelInfo) string {
	if res, ok := gw.Messages.Get(msgID); ok {
		IDs := res.([]*BrMsgID)
		for _, id := range IDs {
			// check protocol, bridge name and channelname
			// for people that reuse the same bridge multiple times. see #342
			if dest.Protocol == id.br.Protocol && dest.Name == id.br.Name && channel.ID == id.ChannelID {
				return strings.Replace(id.ID, dest.Protocol+" ", "", 1)
			}
		}
	}
	return ""
}

func (gw *Gateway) handleMessage(msg config.Message, dest *bridge.Bridge) []*BrMsgID {
	var brMsgIDs []*BrMsgID

	// if we have an attached file, or other info
	if msg.Extra != nil {
		if len(msg.Extra[config.EventFileFailureSize]) != 0 {
			if msg.Text == "" {
				return brMsgIDs
			}
		}
	}

	// Avatar downloads are only relevant for telegram and mattermost for now
	if msg.Event == config.EventAvatarDownload {
		if dest.Protocol != "mattermost" &&
			dest.Protocol != "telegram" {
			return brMsgIDs
		}
	}

	// only relay join/part when configured
	if msg.Event == config.EventJoinLeave && !gw.Bridges[dest.Account].GetBool("ShowJoinPart") {
		return brMsgIDs
	}

	// only relay topic change when configured
	if msg.Event == config.EventTopicChange && !gw.Bridges[dest.Account].GetBool("ShowTopicChange") {
		return brMsgIDs
	}

	// broadcast to every out channel (irc QUIT)
	if msg.Channel == "" && msg.Event != config.EventJoinLeave {
		flog.Debug("empty channel")
		return brMsgIDs
	}

	// Get the ID of the parent message in thread
	var canonicalParentMsgID string
	if msg.ParentID != "" && (gw.BridgeValues().General.PreserveThreading || dest.GetBool("PreserveThreading")) {
		canonicalParentMsgID = gw.FindCanonicalMsgID(msg.Protocol, msg.ParentID)
	}

	originchannel := msg.Channel
	origmsg := msg
	channels := gw.getDestChannel(&msg, *dest)
	for _, channel := range channels {
		// Only send the avatar download event to ourselves.
		if msg.Event == config.EventAvatarDownload {
			if channel.ID != getChannelID(origmsg) {
				continue
			}
		} else {
			// do not send to ourself for any other event
			if channel.ID == getChannelID(origmsg) {
				continue
			}
		}

		// Too noisy to log like other events
		if msg.Event != config.EventUserTyping {
			flog.Debugf("=> Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, originchannel, dest.Account, channel.Name)
		}

		msg.Channel = channel.Name
		msg.Avatar = gw.modifyAvatar(origmsg, dest)
		msg.Username = gw.modifyUsername(origmsg, dest)

		msg.ID = gw.getDestMsgID(origmsg.Protocol+" "+origmsg.ID, dest, channel)

		// for api we need originchannel as channel
		if dest.Protocol == apiProtocol {
			msg.Channel = originchannel
		}

		msg.ParentID = gw.getDestMsgID(origmsg.Protocol+" "+canonicalParentMsgID, dest, channel)
		if msg.ParentID == "" {
			msg.ParentID = canonicalParentMsgID
		}

		// if we are using mattermost plugin account, send messages to MattermostPlugin channel
		// that can be picked up by the mattermost matterbridge plugin
		if dest.Account == "mattermost.plugin" {
			gw.Router.MattermostPlugin <- msg
		}

		mID, err := dest.Send(msg)
		if err != nil {
			flog.Error(err)
		}

		// append the message ID (mID) from this bridge (dest) to our brMsgIDs slice
		if mID != "" {
			flog.Debugf("mID %s: %s", dest.Account, mID)
			brMsgIDs = append(brMsgIDs, &BrMsgID{dest, dest.Protocol + " " + mID, channel.ID})
		}
	}
	return brMsgIDs
}

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
	// if we don't have the bridge, ignore it
	if _, ok := gw.Bridges[msg.Account]; !ok {
		return true
	}

	// check if we need to ignore a empty message
	if msg.Text == "" {
		if msg.Event == config.EventUserTyping {
			return false
		}
		// we have an attachment or actual bytes, do not ignore
		if msg.Extra != nil &&
			(msg.Extra["attachments"] != nil ||
				len(msg.Extra["file"]) > 0 ||
				len(msg.Extra[config.EventFileFailureSize]) > 0) {
			return false
		}
		flog.Debugf("ignoring empty message %#v from %s", msg, msg.Account)
		return true
	}

	// is the username in IgnoreNicks field
	for _, entry := range strings.Fields(gw.Bridges[msg.Account].GetString("IgnoreNicks")) {
		if msg.Username == entry {
			flog.Debugf("ignoring %s from %s", msg.Username, msg.Account)
			return true
		}
	}

	// does the message match regex in IgnoreMessages field
	// TODO do not compile regexps everytime
	for _, entry := range strings.Fields(gw.Bridges[msg.Account].GetString("IgnoreMessages")) {
		if entry != "" {
			re, err := regexp.Compile(entry)
			if err != nil {
				flog.Errorf("incorrect regexp %s for %s", entry, msg.Account)
				continue
			}
			if re.MatchString(msg.Text) {
				flog.Debugf("matching %s. ignoring %s from %s", entry, msg.Text, msg.Account)
				return true
			}
		}
	}
	return false
}

func (gw *Gateway) modifyUsername(msg config.Message, dest *bridge.Bridge) string {
	br := gw.Bridges[msg.Account]
	msg.Protocol = br.Protocol
	if gw.BridgeValues().General.StripNick || dest.GetBool("StripNick") {
		re := regexp.MustCompile("[^a-zA-Z0-9]+")
		msg.Username = re.ReplaceAllString(msg.Username, "")
	}
	nick := dest.GetString("RemoteNickFormat")
	if nick == "" {
		nick = gw.BridgeValues().General.RemoteNickFormat
	}

	// loop to replace nicks
	for _, outer := range br.GetStringSlice2D("ReplaceNicks") {
		search := outer[0]
		replace := outer[1]
		// TODO move compile to bridge init somewhere
		re, err := regexp.Compile(search)
		if err != nil {
			flog.Errorf("regexp in %s failed: %s", msg.Account, err)
			break
		}
		msg.Username = re.ReplaceAllString(msg.Username, replace)
	}

	if len(msg.Username) > 0 {
		// fix utf-8 issue #193
		i := 0
		for index := range msg.Username {
			if i == 1 {
				i = index
				break
			}
			i++
		}
		nick = strings.Replace(nick, "{NOPINGNICK}", msg.Username[:i]+"​"+msg.Username[i:], -1)
	}

	nick = strings.Replace(nick, "{BRIDGE}", br.Name, -1)
	nick = strings.Replace(nick, "{PROTOCOL}", br.Protocol, -1)
	nick = strings.Replace(nick, "{GATEWAY}", gw.Name, -1)
	nick = strings.Replace(nick, "{LABEL}", br.GetString("Label"), -1)
	nick = strings.Replace(nick, "{NICK}", msg.Username, -1)
	nick = strings.Replace(nick, "{CHANNEL}", msg.Channel, -1)
	return nick
}

func (gw *Gateway) modifyAvatar(msg config.Message, dest *bridge.Bridge) string {
	iconurl := gw.BridgeValues().General.IconURL
	if iconurl == "" {
		iconurl = dest.GetString("IconURL")
	}
	iconurl = strings.Replace(iconurl, "{NICK}", msg.Username, -1)
	if msg.Avatar == "" {
		msg.Avatar = iconurl
	}
	return msg.Avatar
}

func (gw *Gateway) modifyMessage(msg *config.Message) {
	// replace :emoji: to unicode
	msg.Text = emojilib.Replace(msg.Text)

	br := gw.Bridges[msg.Account]
	// loop to replace messages
	for _, outer := range br.GetStringSlice2D("ReplaceMessages") {
		search := outer[0]
		replace := outer[1]
		// TODO move compile to bridge init somewhere
		re, err := regexp.Compile(search)
		if err != nil {
			flog.Errorf("regexp in %s failed: %s", msg.Account, err)
			break
		}
		msg.Text = re.ReplaceAllString(msg.Text, replace)
	}

	// messages from api have Gateway specified, don't overwrite
	if msg.Protocol != apiProtocol {
		msg.Gateway = gw.Name
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

	client := &http.Client{
		Timeout: time.Second * 5,
	}

	for i, f := range msg.Extra["file"] {
		fi := f.(config.FileInfo)
		ext := filepath.Ext(fi.Name)
		fi.Name = fi.Name[0 : len(fi.Name)-len(ext)]
		fi.Name = reg.ReplaceAllString(fi.Name, "_")
		fi.Name += ext

		sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))[:8]

		if gw.BridgeValues().General.MediaServerUpload != "" {
			// Use MediaServerUpload. Upload using a PUT HTTP request and basicauth.

			url := gw.BridgeValues().General.MediaServerUpload + "/" + sha1sum + "/" + fi.Name

			req, err := http.NewRequest("PUT", url, bytes.NewReader(*fi.Data))
			if err != nil {
				flog.Errorf("mediaserver upload failed, could not create request: %#v", err)
				continue
			}

			flog.Debugf("mediaserver upload url: %s", url)

			req.Header.Set("Content-Type", "binary/octet-stream")
			_, err = client.Do(req)
			if err != nil {
				flog.Errorf("mediaserver upload failed, could not Do request: %#v", err)
				continue
			}
		} else {
			// Use MediaServerPath. Place the file on the current filesystem.

			dir := gw.BridgeValues().General.MediaDownloadPath + "/" + sha1sum
			err := os.Mkdir(dir, os.ModePerm)
			if err != nil && !os.IsExist(err) {
				flog.Errorf("mediaserver path failed, could not mkdir: %s %#v", err, err)
				continue
			}

			path := dir + "/" + fi.Name
			flog.Debugf("mediaserver path placing file: %s", path)

			err = ioutil.WriteFile(path, *fi.Data, os.ModePerm)
			if err != nil {
				flog.Errorf("mediaserver path failed, could not writefile: %s %#v", err, err)
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

func (gw *Gateway) validGatewayDest(msg *config.Message) bool {
	return msg.Gateway == gw.Name
}

func getChannelID(msg config.Message) string {
	return msg.Channel + msg.Account
}

func isAPI(account string) bool {
	return strings.HasPrefix(account, "api.")
}
