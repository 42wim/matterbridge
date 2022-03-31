package gateway

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/internal"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	lru "github.com/hashicorp/golang-lru"
	"github.com/kyokomi/emoji/v2"
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

	logger *logrus.Entry
}

type BrMsgID struct {
	br        *bridge.Bridge
	ID        string
	ChannelID string
}

const apiProtocol = "api"

// New creates a new Gateway object associated with the specified router and
// following the given configuration.
func New(rootLogger *logrus.Logger, cfg *config.Gateway, r *Router) *Gateway {
	logger := rootLogger.WithFields(logrus.Fields{"prefix": "gateway"})

	cache, _ := lru.New(5000)
	gw := &Gateway{
		Channels: make(map[string]*config.ChannelInfo),
		Message:  r.Message,
		Router:   r,
		Bridges:  make(map[string]*bridge.Bridge),
		Config:   r.Config,
		Messages: cache,
		logger:   logger,
	}
	if err := gw.AddConfig(cfg); err != nil {
		logger.Errorf("Failed to add configuration to gateway: %#v", err)
	}
	return gw
}

// FindCanonicalMsgID returns the ID under which a message was stored in the cache.
func (gw *Gateway) FindCanonicalMsgID(protocol string, mID string) string {
	ID := protocol + " " + mID
	if gw.Messages.Contains(ID) {
		return ID
	}

	// If not keyed, iterate through cache for downstream, and infer upstream.
	for _, mid := range gw.Messages.Keys() {
		v, _ := gw.Messages.Peek(mid)
		ids := v.([]*BrMsgID)
		for _, downstreamMsgObj := range ids {
			if ID == downstreamMsgObj.ID {
				return mid.(string)
			}
		}
	}
	return ""
}

// AddBridge sets up a new bridge in the gateway object with the specified configuration.
func (gw *Gateway) AddBridge(cfg *config.Bridge) error {
	br := gw.Router.getBridge(cfg.Account)
	if br == nil {
		gw.checkConfig(cfg)
		br = bridge.New(cfg)
		br.Config = gw.Router.Config
		br.General = &gw.BridgeValues().General
		br.Log = gw.logger.WithFields(logrus.Fields{"prefix": br.Protocol})
		brconfig := &bridge.Config{
			Remote: gw.Message,
			Bridge: br,
		}
		// add the actual bridger for this protocol to this bridge using the bridgeMap
		if _, ok := gw.Router.BridgeMap[br.Protocol]; !ok {
			gw.logger.Fatalf("Incorrect protocol %s specified in gateway configuration %s, exiting.", br.Protocol, cfg.Account)
		}
		br.Bridger = gw.Router.BridgeMap[br.Protocol](brconfig)
	}
	gw.mapChannelsToBridge(br)
	gw.Bridges[cfg.Account] = br
	return nil
}

func (gw *Gateway) checkConfig(cfg *config.Bridge) {
	match := false
	for _, key := range gw.Router.Config.Viper().AllKeys() {
		if strings.HasPrefix(key, strings.ToLower(cfg.Account)) {
			match = true
			break
		}
	}
	if !match {
		gw.logger.Fatalf("Account %s defined in gateway %s but no configuration found, exiting.", cfg.Account, gw.Name)
	}
}

// AddConfig associates a new configuration with the gateway object.
func (gw *Gateway) AddConfig(cfg *config.Gateway) error {
	gw.Name = cfg.Name
	gw.MyConfig = cfg
	if err := gw.mapChannels(); err != nil {
		gw.logger.Errorf("mapChannels() failed: %s", err)
	}
	for _, br := range append(gw.MyConfig.In, append(gw.MyConfig.InOut, gw.MyConfig.Out...)...) {
		br := br // scopelint
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
		gw.logger.Errorf("Disconnect() %s failed: %s", br.Account, err)
	}
	time.Sleep(time.Second * 5)
RECONNECT:
	gw.logger.Infof("Reconnecting %s", br.Account)
	err := br.Connect()
	if err != nil {
		gw.logger.Errorf("Reconnection failed: %s. Trying again in 60 seconds", err)
		time.Sleep(time.Second * 60)
		goto RECONNECT
	}
	br.Joined = make(map[string]bool)
	if err := br.JoinChannels(); err != nil {
		gw.logger.Errorf("JoinChannels() %s failed: %s", br.Account, err)
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
			gw.logger.Errorf("Mattermost channels do not start with a #: remove the # in %s", br.Channel)
			os.Exit(1)
		}
		if strings.HasPrefix(br.Account, "zulip.") && !strings.Contains(br.Channel, "/topic:") {
			gw.logger.Errorf("Breaking change, since matterbridge 1.14.0 zulip channels need to specify the topic with channel/topic:mytopic in %s of %s", br.Channel, br.Account)
			os.Exit(1)
		}
		ID := br.Channel + br.Account
		if _, ok := gw.Channels[ID]; !ok {
			channel := &config.ChannelInfo{
				Name:        br.Channel,
				Direction:   direction,
				ID:          ID,
				Options:     br.Options,
				Account:     br.Account,
				SameChannel: make(map[string]bool),
			}
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

	// if source channel is in only, do nothing
	for _, channel := range gw.Channels {
		// lookup the channel from the message
		if channel.ID == getChannelID(msg) {
			// we only have destinations if the original message is from an "in" (sending) channel
			if !strings.Contains(channel.Direction, "in") {
				return channels
			}
			continue
		}
	}
	for _, channel := range gw.Channels {
		if _, ok := gw.Channels[getChannelID(msg)]; !ok {
			continue
		}

		// do samechannelgateway logic
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

func (gw *Gateway) getDestMsgID(msgID string, dest *bridge.Bridge, channel *config.ChannelInfo) string {
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
	gw.logger.Debugf("ignoring empty message %#v from %s", msg, msg.Account)
	return true
}

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
	// if we don't have the bridge, ignore it
	if _, ok := gw.Bridges[msg.Account]; !ok {
		return true
	}

	igNicks := strings.Fields(gw.Bridges[msg.Account].GetString("IgnoreNicks"))
	igMessages := strings.Fields(gw.Bridges[msg.Account].GetString("IgnoreMessages"))
	if gw.ignoreTextEmpty(msg) || gw.ignoreText(msg.Username, igNicks) || gw.ignoreText(msg.Text, igMessages) || gw.ignoreFilesComment(msg.Extra, igMessages) {
		return true
	}

	return false
}

// ignoreFilesComment returns true if we need to ignore a file with matched comment.
func (gw *Gateway) ignoreFilesComment(extra map[string][]interface{}, igMessages []string) bool {
	if extra == nil {
		return false
	}
	for _, f := range extra["file"] {
		fi, ok := f.(config.FileInfo)
		if !ok {
			continue
		}
		if gw.ignoreText(fi.Comment, igMessages) {
			return true
		}
	}
	return false
}

func (gw *Gateway) modifyUsername(msg *config.Message, dest *bridge.Bridge) string {
	if dest.GetBool("StripNick") {
		re := regexp.MustCompile("[^a-zA-Z0-9]+")
		msg.Username = re.ReplaceAllString(msg.Username, "")
	}
	nick := dest.GetString("RemoteNickFormat")

	// loop to replace nicks
	br := gw.Bridges[msg.Account]
	for _, outer := range br.GetStringSlice2D("ReplaceNicks") {
		search := outer[0]
		replace := outer[1]
		// TODO move compile to bridge init somewhere
		re, err := regexp.Compile(search)
		if err != nil {
			gw.logger.Errorf("regexp in %s failed: %s", msg.Account, err)
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
		nick = strings.ReplaceAll(nick, "{NOPINGNICK}", msg.Username[:i]+"\u200b"+msg.Username[i:])
	}

	nick = strings.ReplaceAll(nick, "{BRIDGE}", br.Name)
	nick = strings.ReplaceAll(nick, "{PROTOCOL}", br.Protocol)
	nick = strings.ReplaceAll(nick, "{GATEWAY}", gw.Name)
	nick = strings.ReplaceAll(nick, "{LABEL}", br.GetString("Label"))
	nick = strings.ReplaceAll(nick, "{NICK}", msg.Username)
	nick = strings.ReplaceAll(nick, "{USERID}", msg.UserID)
	nick = strings.ReplaceAll(nick, "{CHANNEL}", msg.Channel)
	tengoNick, err := gw.modifyUsernameTengo(msg, br)
	if err != nil {
		gw.logger.Errorf("modifyUsernameTengo error: %s", err)
	}
	nick = strings.ReplaceAll(nick, "{TENGO}", tengoNick)
	return nick
}

func (gw *Gateway) modifyAvatar(msg *config.Message, dest *bridge.Bridge) string {
	iconurl := dest.GetString("IconURL")
	iconurl = strings.Replace(iconurl, "{NICK}", msg.Username, -1)
	if msg.Avatar == "" {
		msg.Avatar = iconurl
	}
	return msg.Avatar
}

func (gw *Gateway) modifyMessage(msg *config.Message) {
	if gw.BridgeValues().General.TengoModifyMessage != "" {
		gw.logger.Warnf("General TengoModifyMessage=%s is deprecated and will be removed in v1.20.0, please move to Tengo InMessage=%s", gw.BridgeValues().General.TengoModifyMessage, gw.BridgeValues().General.TengoModifyMessage)
	}

	if err := modifyInMessageTengo(gw.BridgeValues().General.TengoModifyMessage, msg); err != nil {
		gw.logger.Errorf("TengoModifyMessage failed: %s", err)
	}

	inMessage := gw.BridgeValues().Tengo.InMessage
	if inMessage == "" {
		inMessage = gw.BridgeValues().Tengo.Message
		if inMessage != "" {
			gw.logger.Warnf("Tengo Message=%s is deprecated and will be removed in v1.20.0, please move to Tengo InMessage=%s", inMessage, inMessage)
		}
	}

	if err := modifyInMessageTengo(inMessage, msg); err != nil {
		gw.logger.Errorf("Tengo.Message failed: %s", err)
	}

	// replace :emoji: to unicode
	emoji.ReplacePadding = ""
	msg.Text = emoji.Sprint(msg.Text)

	br := gw.Bridges[msg.Account]
	// loop to replace messages
	for _, outer := range br.GetStringSlice2D("ReplaceMessages") {
		search := outer[0]
		replace := outer[1]
		// TODO move compile to bridge init somewhere
		re, err := regexp.Compile(search)
		if err != nil {
			gw.logger.Errorf("regexp in %s failed: %s", msg.Account, err)
			break
		}
		msg.Text = re.ReplaceAllString(msg.Text, replace)
	}

	gw.handleExtractNicks(msg)

	// messages from api have Gateway specified, don't overwrite
	if msg.Protocol != apiProtocol {
		msg.Gateway = gw.Name
	}
}

// SendMessage sends a message (with specified parentID) to the channel on the selected
// destination bridge and returns a message ID or an error.
func (gw *Gateway) SendMessage(
	rmsg *config.Message,
	dest *bridge.Bridge,
	channel *config.ChannelInfo,
	canonicalParentMsgID string,
) (string, error) {
	msg := *rmsg
	// Only send the avatar download event to ourselves.
	if msg.Event == config.EventAvatarDownload {
		if channel.ID != getChannelID(rmsg) {
			return "", nil
		}
	} else {
		// do not send to ourself for any other event
		if channel.ID == getChannelID(rmsg) {
			return "", nil
		}
	}

	// Only send irc notices to irc
	if msg.Event == config.EventNoticeIRC && dest.Protocol != "irc" {
		return "", nil
	}

	// Too noisy to log like other events
	debugSendMessage := ""
	if msg.Event != config.EventUserTyping {
		debugSendMessage = fmt.Sprintf("=> Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, rmsg.Channel, dest.Account, channel.Name)
	}

	msg.Channel = channel.Name
	msg.Avatar = gw.modifyAvatar(rmsg, dest)
	msg.Username = gw.modifyUsername(rmsg, dest)

	// exclude file delete event as the msg ID here is the native file ID that needs to be deleted
	if msg.Event != config.EventFileDelete {
		msg.ID = gw.getDestMsgID(rmsg.Protocol+" "+rmsg.ID, dest, channel)
	}

	// for api we need originchannel as channel
	if dest.Protocol == apiProtocol {
		msg.Channel = rmsg.Channel
	}

	msg.ParentID = gw.getDestMsgID(canonicalParentMsgID, dest, channel)
	if msg.ParentID == "" {
		msg.ParentID = strings.Replace(canonicalParentMsgID, dest.Protocol+" ", "", 1)
	}

	// if the parentID is still empty and we have a parentID set in the original message
	// this means that we didn't find it in the cache so set it to a "msg-parent-not-found" constant
	if msg.ParentID == "" && rmsg.ParentID != "" {
		msg.ParentID = config.ParentIDNotFound
	}

	drop, err := gw.modifyOutMessageTengo(rmsg, &msg, dest)
	if err != nil {
		gw.logger.Errorf("modifySendMessageTengo: %s", err)
	}

	if drop {
		gw.logger.Debugf("=> Tengo dropping %#v from %s (%s) to %s (%s)", msg, msg.Account, rmsg.Channel, dest.Account, channel.Name)
		return "", nil
	}

	if debugSendMessage != "" {
		gw.logger.Debug(debugSendMessage)
	}
	// if we are using mattermost plugin account, send messages to MattermostPlugin channel
	// that can be picked up by the mattermost matterbridge plugin
	if dest.Account == "mattermost.plugin" {
		gw.Router.MattermostPlugin <- msg
	}

	defer func(t time.Time) {
		gw.logger.Debugf("=> Send from %s (%s) to %s (%s) took %s", msg.Account, rmsg.Channel, dest.Account, channel.Name, time.Since(t))
	}(time.Now())

	mID, err := dest.Send(msg)
	if err != nil {
		return mID, err
	}

	// append the message ID (mID) from this bridge (dest) to our brMsgIDs slice
	if mID != "" {
		gw.logger.Debugf("mID %s: %s", dest.Account, mID)
		return mID, nil
		// brMsgIDs = append(brMsgIDs, &BrMsgID{dest, dest.Protocol + " " + mID, channel.ID})
	}
	return "", nil
}

func (gw *Gateway) validGatewayDest(msg *config.Message) bool {
	return msg.Gateway == gw.Name
}

func getChannelID(msg *config.Message) string {
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
			gw.logger.Errorf("incorrect regexp %s", entry)
			continue
		}
		if re.MatchString(text) {
			gw.logger.Debugf("matching %s. ignoring %s", entry, text)
			return true
		}
	}
	return false
}

func getProtocol(msg *config.Message) string {
	p := strings.Split(msg.Account, ".")
	return p[0]
}

func modifyInMessageTengo(filename string, msg *config.Message) error {
	if filename == "" {
		return nil
	}
	res, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	s := tengo.NewScript(res)
	s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	_ = s.Add("msgText", msg.Text)
	_ = s.Add("msgUsername", msg.Username)
	_ = s.Add("msgUserID", msg.UserID)
	_ = s.Add("msgAccount", msg.Account)
	_ = s.Add("msgChannel", msg.Channel)
	c, err := s.Compile()
	if err != nil {
		return err
	}
	if err := c.Run(); err != nil {
		return err
	}
	msg.Text = c.Get("msgText").String()
	msg.Username = c.Get("msgUsername").String()
	return nil
}

func (gw *Gateway) modifyUsernameTengo(msg *config.Message, br *bridge.Bridge) (string, error) {
	filename := gw.BridgeValues().Tengo.RemoteNickFormat
	if filename == "" {
		return "", nil
	}
	res, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	s := tengo.NewScript(res)
	s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	_ = s.Add("result", "")
	_ = s.Add("msgText", msg.Text)
	_ = s.Add("msgUsername", msg.Username)
	_ = s.Add("msgUserID", msg.UserID)
	_ = s.Add("nick", msg.Username)
	_ = s.Add("msgAccount", msg.Account)
	_ = s.Add("msgChannel", msg.Channel)
	_ = s.Add("channel", msg.Channel)
	_ = s.Add("msgProtocol", msg.Protocol)
	_ = s.Add("remoteAccount", br.Account)
	_ = s.Add("protocol", br.Protocol)
	_ = s.Add("bridge", br.Name)
	_ = s.Add("gateway", gw.Name)
	c, err := s.Compile()
	if err != nil {
		return "", err
	}
	if err := c.Run(); err != nil {
		return "", err
	}
	return c.Get("result").String(), nil
}

func (gw *Gateway) modifyOutMessageTengo(origmsg *config.Message, msg *config.Message, br *bridge.Bridge) (bool, error) {
	filename := gw.BridgeValues().Tengo.OutMessage
	var (
		res  []byte
		err  error
		drop bool
	)

	if filename == "" {
		res, err = internal.Asset("tengo/outmessage.tengo")
		if err != nil {
			return drop, err
		}
	} else {
		res, err = ioutil.ReadFile(filename)
		if err != nil {
			return drop, err
		}
	}

	s := tengo.NewScript(res)

	s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	_ = s.Add("inAccount", origmsg.Account)
	_ = s.Add("inProtocol", origmsg.Protocol)
	_ = s.Add("inChannel", origmsg.Channel)
	_ = s.Add("inGateway", origmsg.Gateway)
	_ = s.Add("inEvent", origmsg.Event)
	_ = s.Add("outAccount", br.Account)
	_ = s.Add("outProtocol", br.Protocol)
	_ = s.Add("outChannel", msg.Channel)
	_ = s.Add("outGateway", gw.Name)
	_ = s.Add("outEvent", msg.Event)
	_ = s.Add("msgText", msg.Text)
	_ = s.Add("msgUsername", msg.Username)
	_ = s.Add("msgUserID", msg.UserID)
	_ = s.Add("msgDrop", drop)
	c, err := s.Compile()
	if err != nil {
		return drop, err
	}

	if err := c.Run(); err != nil {
		return drop, err
	}

	drop = c.Get("msgDrop").Bool()
	msg.Text = c.Get("msgText").String()
	msg.Username = c.Get("msgUsername").String()

	return drop, nil
}
