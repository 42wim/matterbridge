package gateway

import (
	"bytes"
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/api"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/discord"
	"github.com/42wim/matterbridge/bridge/gitter"
	"github.com/42wim/matterbridge/bridge/irc"
	"github.com/42wim/matterbridge/bridge/matrix"
	"github.com/42wim/matterbridge/bridge/mattermost"
	"github.com/42wim/matterbridge/bridge/rocketchat"
	"github.com/42wim/matterbridge/bridge/slack"
	"github.com/42wim/matterbridge/bridge/sshchat"
	"github.com/42wim/matterbridge/bridge/steam"
	"github.com/42wim/matterbridge/bridge/telegram"
	"github.com/42wim/matterbridge/bridge/xmpp"
	log "github.com/sirupsen/logrus"
	//	"github.com/davecgh/go-spew/spew"
	"crypto/sha1"
	"github.com/hashicorp/golang-lru"
	"github.com/peterhellberg/emojilib"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type Gateway struct {
	*config.Config
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
	"mattermost": bmattermost.New,
	"irc":        birc.New,
	"gitter":     bgitter.New,
	"matrix":     bmatrix.New,
	"slack":      bslack.New,
	"api":        api.New,
	"telegram":   btelegram.New,
	"discord":    bdiscord.New,
	"steam":      bsteam.New,
	"sshchat":    bsshchat.New,
	"rocketchat": brocketchat.New,
	"xmpp":       bxmpp.New,
}

func init() {
	flog = log.WithFields(log.Fields{"prefix": "gateway"})
}

func New(cfg config.Gateway, r *Router) *Gateway {
	gw := &Gateway{Channels: make(map[string]*config.ChannelInfo), Message: r.Message,
		Router: r, Bridges: make(map[string]*bridge.Bridge), Config: r.Config}
	cache, _ := lru.New(5000)
	gw.Messages = cache
	gw.AddConfig(&cfg)
	return gw
}

func (gw *Gateway) AddBridge(cfg *config.Bridge) error {
	br := gw.Router.getBridge(cfg.Account)
	if br == nil {
		br = bridge.New(cfg)
		// set logging
		br.Log = log.WithFields(log.Fields{"prefix": "bridge"})
		// get the protocol configuration (eg irc)
		pcfg := getField(gw.Config, strings.Title(br.Protocol))
		// get the config for this name (eg freenode, from irc.freenode)
		br.Config = pcfg[br.Name]
		// create the bridge config
		brconfig := &config.BridgeConfig{General: &gw.Config.General, Account: br.Account, Remote: gw.Message, Config: br.Config, Log: log.WithFields(log.Fields{"prefix": br.Protocol})}
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
		if isApi(br.Account) {
			br.Channel = "api"
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
	if msg.Protocol == "api" && gw.Name != msg.Gateway {
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
		if strings.Contains(channel.Direction, "out") && channel.Account == dest.Account && gw.validGatewayDest(msg, channel) {
			channels = append(channels, *channel)
		}
	}
	return channels
}

func (gw *Gateway) handleMessage(msg config.Message, dest *bridge.Bridge) []*BrMsgID {
	var brMsgIDs []*BrMsgID

	// TODO refactor
	// only slack now, check will have to be done in the different bridges.
	// we need to check if we can't use fallback or text in other bridges
	if msg.Extra != nil {
		if dest.Protocol != "discord" &&
			dest.Protocol != "slack" &&
			dest.Protocol != "mattermost" &&
			dest.Protocol != "telegram" &&
			dest.Protocol != "matrix" &&
			dest.Protocol != "xmpp" &&
			dest.Protocol != "irc" &&
			len(msg.Extra[config.EVENT_FILE_FAILURE_SIZE]) == 0 {
			if msg.Text == "" {
				return brMsgIDs
			}
		}
	}

	// Avatar downloads are only relevant for telegram and mattermost for now
	if msg.Event == config.EVENT_AVATAR_DOWNLOAD {
		if dest.Protocol != "mattermost" &&
			dest.Protocol != "telegram" {
			return brMsgIDs
		}
	}
	// only relay join/part when configged
	if msg.Event == config.EVENT_JOIN_LEAVE && !gw.Bridges[dest.Account].Config.ShowJoinPart {
		return brMsgIDs
	}
	if msg.Event == config.EVENT_TOPIC_CHANGE && !gw.Bridges[dest.Account].Config.ShowTopicChange {
		return brMsgIDs
	}

	// broadcast to every out channel (irc QUIT)
	if msg.Channel == "" && msg.Event != config.EVENT_JOIN_LEAVE {
		flog.Debug("empty channel")
		return brMsgIDs
	}
	originchannel := msg.Channel
	origmsg := msg
	channels := gw.getDestChannel(&msg, *dest)
	for _, channel := range channels {
		// Only send the avatar download event to ourselves.
		if msg.Event == config.EVENT_AVATAR_DOWNLOAD {
			if channel.ID != getChannelID(origmsg) {
				continue
			}
		} else {
			// do not send to ourself for any other event
			if channel.ID == getChannelID(origmsg) {
				continue
			}
		}
		flog.Debugf("Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, originchannel, dest.Account, channel.Name)
		msg.Channel = channel.Name
		msg.Avatar = gw.modifyAvatar(origmsg, dest)
		msg.Username = gw.modifyUsername(origmsg, dest)
		msg.ID = ""
		if res, ok := gw.Messages.Get(origmsg.ID); ok {
			IDs := res.([]*BrMsgID)
			for _, id := range IDs {
				// check protocol, bridge name and channelname
				// for people that reuse the same bridge multiple times. see #342
				if dest.Protocol == id.br.Protocol && dest.Name == id.br.Name && channel.ID == id.ChannelID {
					msg.ID = id.ID
				}
			}
		}
		// for api we need originchannel as channel
		if dest.Protocol == "api" {
			msg.Channel = originchannel
		}
		mID, err := dest.Send(msg)
		if err != nil {
			flog.Error(err)
		}
		// append the message ID (mID) from this bridge (dest) to our brMsgIDs slice
		if mID != "" {
			brMsgIDs = append(brMsgIDs, &BrMsgID{dest, mID, channel.ID})
		}
	}
	return brMsgIDs
}

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
	// if we don't have the bridge, ignore it
	if _, ok := gw.Bridges[msg.Account]; !ok {
		return true
	}
	if msg.Text == "" {
		// we have an attachment or actual bytes
		if msg.Extra != nil &&
			(msg.Extra["attachments"] != nil ||
				len(msg.Extra["file"]) > 0 ||
				len(msg.Extra[config.EVENT_FILE_FAILURE_SIZE]) > 0) {
			return false
		}
		flog.Debugf("ignoring empty message %#v from %s", msg, msg.Account)
		return true
	}
	for _, entry := range strings.Fields(gw.Bridges[msg.Account].Config.IgnoreNicks) {
		if msg.Username == entry {
			flog.Debugf("ignoring %s from %s", msg.Username, msg.Account)
			return true
		}
	}
	// TODO do not compile regexps everytime
	for _, entry := range strings.Fields(gw.Bridges[msg.Account].Config.IgnoreMessages) {
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
	if gw.Config.General.StripNick || dest.Config.StripNick {
		re := regexp.MustCompile("[^a-zA-Z0-9]+")
		msg.Username = re.ReplaceAllString(msg.Username, "")
	}
	nick := dest.Config.RemoteNickFormat
	if nick == "" {
		nick = gw.Config.General.RemoteNickFormat
	}

	// loop to replace nicks
	for _, outer := range br.Config.ReplaceNicks {
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
	nick = strings.Replace(nick, "{LABEL}", br.Config.Label, -1)
	nick = strings.Replace(nick, "{NICK}", msg.Username, -1)
	return nick
}

func (gw *Gateway) modifyAvatar(msg config.Message, dest *bridge.Bridge) string {
	iconurl := gw.Config.General.IconURL
	if iconurl == "" {
		iconurl = dest.Config.IconURL
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
	for _, outer := range br.Config.ReplaceMessages {
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
	if msg.Protocol != "api" {
		msg.Gateway = gw.Name
	}
}

func (gw *Gateway) handleFiles(msg *config.Message) {
	if msg.Extra == nil || gw.Config.General.MediaServerUpload == "" {
		return
	}
	if len(msg.Extra["file"]) > 0 {
		client := &http.Client{
			Timeout: time.Second * 5,
		}
		for i, f := range msg.Extra["file"] {
			fi := f.(config.FileInfo)
			sha1sum := fmt.Sprintf("%x", sha1.Sum(*fi.Data))
			reader := bytes.NewReader(*fi.Data)
			url := gw.Config.General.MediaServerUpload + "/" + sha1sum + "/" + fi.Name
			durl := gw.Config.General.MediaServerDownload + "/" + sha1sum + "/" + fi.Name
			extra := msg.Extra["file"][i].(config.FileInfo)
			extra.URL = durl
			req, _ := http.NewRequest("PUT", url, reader)
			req.Header.Set("Content-Type", "binary/octet-stream")
			_, err := client.Do(req)
			if err != nil {
				flog.Errorf("mediaserver upload failed: %#v", err)
				continue
			}
			flog.Debugf("mediaserver download URL = %s", durl)
			// we uploaded the file successfully. Add the SHA
			extra.SHA = sha1sum
			msg.Extra["file"][i] = extra
		}
	}
}

func getChannelID(msg config.Message) string {
	return msg.Channel + msg.Account
}

func (gw *Gateway) validGatewayDest(msg *config.Message, channel *config.ChannelInfo) bool {
	return msg.Gateway == gw.Name
}

func isApi(account string) bool {
	return strings.HasPrefix(account, "api.")
}

//getField returns the Protocol configuration for a specific protocol (field)
func getField(cfg *config.Config, field string) map[string]config.Protocol {
	r := reflect.ValueOf(cfg)
	f := reflect.Indirect(r).FieldByName(field)
	i := f.Interface()
	return i.(map[string]config.Protocol)
}
