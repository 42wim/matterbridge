package gateway

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/hashicorp/golang-lru"
	"github.com/peterhellberg/emojilib"
	"github.com/sirupsen/logrus"
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

var flog *logrus.Entry

const (
	apiProtocol = "api"
)

func New(cfg config.Gateway, r *Router) *Gateway {
	flog = logrus.WithFields(logrus.Fields{"prefix": "gateway"})
	gw := &Gateway{Channels: make(map[string]*config.ChannelInfo), Message: r.Message,
		Router: r, Bridges: make(map[string]*bridge.Bridge), Config: r.Config}
	cache, _ := lru.New(5000)
	gw.Messages = cache
	if err := gw.AddConfig(&cfg); err != nil {
		flog.Errorf("AddConfig failed: %s", err)
	}
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
		br.Log = logrus.WithFields(logrus.Fields{"prefix": "bridge"})
		brconfig := &bridge.Config{Remote: gw.Message, Log: logrus.WithFields(logrus.Fields{"prefix": br.Protocol}), Bridge: br}
		// add the actual bridger for this protocol to this bridge using the bridgeMap
		br.Bridger = gw.Router.BridgeMap[br.Protocol](brconfig)
	}
	gw.mapChannelsToBridge(br)
	gw.Bridges[cfg.Account] = br
	return nil
}

func (gw *Gateway) AddConfig(cfg *config.Gateway) error {
	gw.Name = cfg.Name
	gw.MyConfig = cfg
	if err := gw.mapChannels(); err != nil {
		flog.Errorf("mapChannels() failed: %s", err)
	}
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
	if err := br.Disconnect(); err != nil {
		flog.Errorf("Disconnect() %s failed: %s", br.Account, err)
	}
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
	if err := br.JoinChannels(); err != nil {
		flog.Errorf("JoinChannels() %s failed: %s", br.Account, err)
	}
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
		if strings.HasPrefix(br.Account, "mattermost.") && strings.HasPrefix(br.Channel, "#") {
			flog.Errorf("Mattermost channels do not start with a #: remove the # in %s", br.Channel)
			os.Exit(1)
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

	// discord join/leave is for the whole bridge, isn't a per channel join/leave
	if msg.Event == config.EventJoinLeave && getProtocol(msg) == "discord" && msg.Channel == "" {
		for _, channel := range gw.Channels {
			if channel.Account == dest.Account && strings.Contains(channel.Direction, "out") &&
				gw.validGatewayDest(msg) {
				channels = append(channels, *channel)
			}
		}
		return channels
	}

	// irc quit is for the whole bridge, isn't a per channel quit.
	// channel is empty when we quit
	if msg.Event == config.EventJoinLeave && getProtocol(msg) == "irc" && msg.Channel == "" {
		for _, channel := range gw.Channels {
			if channel.Account == dest.Account && strings.Contains(channel.Direction, "out") &&
				gw.validGatewayDest(msg) {
				channels = append(channels, *channel)
			}
		}
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

// ignoreTextEmpty returns true if we need to ignore a message with an empty text.
func (gw *Gateway) ignoreTextEmpty(msg *config.Message) bool {
	if msg.Text != "" {
		return false
	}
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

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
	// if we don't have the bridge, ignore it
	if _, ok := gw.Bridges[msg.Account]; !ok {
		return true
	}

	igNicks := strings.Fields(gw.Bridges[msg.Account].GetString("IgnoreNicks"))
	igMessages := strings.Fields(gw.Bridges[msg.Account].GetString("IgnoreMessages"))
	if gw.ignoreTextEmpty(msg) || gw.ignoreText(msg.Username, igNicks) || gw.ignoreText(msg.Text, igMessages) {
		return true
	}

	return false
}

func (gw *Gateway) modifyUsername(msg config.Message, dest *bridge.Bridge) string {
	br := gw.Bridges[msg.Account]
	msg.Protocol = br.Protocol
	if dest.GetBool("StripNick") {
		re := regexp.MustCompile("[^a-zA-Z0-9]+")
		msg.Username = re.ReplaceAllString(msg.Username, "")
	}
	nick := dest.GetString("RemoteNickFormat")

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
	iconurl := dest.GetString("IconURL")
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

// SendMessage sends a message (with specified parentID) to the channel on the selected destination bridge.
// returns a message id and error.
func (gw *Gateway) SendMessage(origmsg config.Message, dest *bridge.Bridge, channel config.ChannelInfo, canonicalParentMsgID string) (string, error) {
	msg := origmsg
	// Only send the avatar download event to ourselves.
	if msg.Event == config.EventAvatarDownload {
		if channel.ID != getChannelID(origmsg) {
			return "", nil
		}
	} else {
		// do not send to ourself for any other event
		if channel.ID == getChannelID(origmsg) {
			return "", nil
		}
	}

	// Too noisy to log like other events
	if msg.Event != config.EventUserTyping {
		flog.Debugf("=> Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, origmsg.Channel, dest.Account, channel.Name)
	}

	msg.Channel = channel.Name
	msg.Avatar = gw.modifyAvatar(origmsg, dest)
	msg.Username = gw.modifyUsername(origmsg, dest)

	msg.ID = gw.getDestMsgID(origmsg.Protocol+" "+origmsg.ID, dest, channel)

	// for api we need originchannel as channel
	if dest.Protocol == apiProtocol {
		msg.Channel = origmsg.Channel
	}

	msg.ParentID = gw.getDestMsgID(origmsg.Protocol+" "+canonicalParentMsgID, dest, channel)
	if msg.ParentID == "" {
		msg.ParentID = canonicalParentMsgID
	}

	// if the parentID is still empty and we have a parentID set in the original message
	// this means that we didn't find it in the cache so set it "msg-parent-not-found"
	if msg.ParentID == "" && origmsg.ParentID != "" {
		msg.ParentID = "msg-parent-not-found"
	}

	// if we are using mattermost plugin account, send messages to MattermostPlugin channel
	// that can be picked up by the mattermost matterbridge plugin
	if dest.Account == "mattermost.plugin" {
		gw.Router.MattermostPlugin <- msg
	}

	mID, err := dest.Send(msg)
	if err != nil {
		return mID, err
	}

	// append the message ID (mID) from this bridge (dest) to our brMsgIDs slice
	if mID != "" {
		flog.Debugf("mID %s: %s", dest.Account, mID)
		return mID, nil
		//brMsgIDs = append(brMsgIDs, &BrMsgID{dest, dest.Protocol + " " + mID, channel.ID})
	}
	return "", nil
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

// ignoreText returns true if text matches any of the input regexes.
func (gw *Gateway) ignoreText(text string, input []string) bool {
	for _, entry := range input {
		if entry == "" {
			continue
		}
		// TODO do not compile regexps everytime
		re, err := regexp.Compile(entry)
		if err != nil {
			flog.Errorf("incorrect regexp %s", entry)
			continue
		}
		if re.MatchString(text) {
			flog.Debugf("matching %s. ignoring %s", entry, text)
			return true
		}
	}
	return false
}

func getProtocol(msg *config.Message) string {
	p := strings.Split(msg.Account, ".")
	return p[0]
}
