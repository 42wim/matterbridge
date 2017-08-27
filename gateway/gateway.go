package gateway

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	//	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/golang-lru"
	"github.com/peterhellberg/emojilib"
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
	//map[string][]*BrMsg
	lruCache *lru.Cache
}

type BrMsgID struct {
	br *bridge.Bridge
	ID string
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
		br = bridge.New(gw.Config, cfg, gw.Message)
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
	log.Infof("Reconnecting %s", br.Account)
	err := br.Connect()
	if err != nil {
		log.Errorf("Reconnection failed: %s. Trying again in 60 seconds", err)
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

		// do samechannelgateway logic
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
	// only relay join/part when configged
	if msg.Event == config.EVENT_JOIN_LEAVE && !gw.Bridges[dest.Account].Config.ShowJoinPart {
		return brMsgIDs
	}
	// broadcast to every out channel (irc QUIT)
	if msg.Channel == "" && msg.Event != config.EVENT_JOIN_LEAVE {
		log.Debug("empty channel")
		return brMsgIDs
	}
	originchannel := msg.Channel
	origmsg := msg
	channels := gw.getDestChannel(&msg, *dest)
	for _, channel := range channels {
		// do not send to ourself
		if channel.ID == getChannelID(origmsg) {
			continue
		}
		log.Debugf("Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, originchannel, dest.Account, channel.Name)
		msg.Channel = channel.Name
		msg.Avatar = gw.modifyAvatar(origmsg, dest)
		msg.Username = gw.modifyUsername(origmsg, dest)
		msg.ID = ""
		if res, ok := gw.Messages.Get(origmsg.ID); ok {
			IDs := res.([]*BrMsgID)
			for _, id := range IDs {
				if dest.Protocol == id.br.Protocol {
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
			fmt.Println(err)
		}
		// append the message ID (mID) from this bridge (dest) to our brMsgIDs slice
		log.Debugf("message ID: %s\n", mID)
		brMsgIDs = append(brMsgIDs, &BrMsgID{dest, mID})
	}
	return brMsgIDs
}

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
	// if we don't have the bridge, ignore it
	if _, ok := gw.Bridges[msg.Account]; !ok {
		return true
	}
	if msg.Text == "" {
		log.Debugf("ignoring empty message %#v from %s", msg, msg.Account)
		return true
	}
	for _, entry := range strings.Fields(gw.Bridges[msg.Account].Config.IgnoreNicks) {
		if msg.Username == entry {
			log.Debugf("ignoring %s from %s", msg.Username, msg.Account)
			return true
		}
	}
	// TODO do not compile regexps everytime
	for _, entry := range strings.Fields(gw.Bridges[msg.Account].Config.IgnoreMessages) {
		if entry != "" {
			re, err := regexp.Compile(entry)
			if err != nil {
				log.Errorf("incorrect regexp %s for %s", entry, msg.Account)
				continue
			}
			if re.MatchString(msg.Text) {
				log.Debugf("matching %s. ignoring %s from %s", entry, msg.Text, msg.Account)
				return true
			}
		}
	}
	return false
}

func (gw *Gateway) modifyUsername(msg config.Message, dest *bridge.Bridge) string {
	br := gw.Bridges[msg.Account]
	msg.Protocol = br.Protocol
	nick := gw.Config.General.RemoteNickFormat
	if nick == "" {
		nick = dest.Config.RemoteNickFormat
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
	nick = strings.Replace(nick, "{NICK}", msg.Username, -1)
	nick = strings.Replace(nick, "{BRIDGE}", br.Name, -1)
	nick = strings.Replace(nick, "{PROTOCOL}", br.Protocol, -1)
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
	msg.Gateway = gw.Name
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
