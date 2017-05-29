package gateway

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"strings"
	"time"
)

type Gateway struct {
	*config.Config
	MyConfig        *config.Gateway
	Bridges         map[string]*bridge.Bridge
	Channels        map[string]*config.ChannelInfo
	ChannelOptions  map[string]config.ChannelOptions
	Names           map[string]bool
	Name            string
	Message         chan config.Message
	DestChannelFunc func(msg *config.Message, dest bridge.Bridge) []config.ChannelInfo
}

func New(cfg *config.Config) *Gateway {
	gw := &Gateway{}
	gw.Config = cfg
	gw.Channels = make(map[string]*config.ChannelInfo)
	gw.Message = make(chan config.Message)
	gw.Bridges = make(map[string]*bridge.Bridge)
	gw.Names = make(map[string]bool)
	gw.DestChannelFunc = gw.getDestChannel
	return gw
}

func (gw *Gateway) AddBridge(cfg *config.Bridge) error {
	for _, br := range gw.Bridges {
		if br.Account == cfg.Account {
			gw.mapChannelsToBridge(br)
			err := br.JoinChannels()
			if err != nil {
				return fmt.Errorf("Bridge %s failed to join channel: %v", br.Account, err)
			}
			return nil
		}
	}
	log.Infof("Starting bridge: %s ", cfg.Account)
	br := bridge.New(gw.Config, cfg, gw.Message)
	gw.mapChannelsToBridge(br)
	gw.Bridges[cfg.Account] = br
	err := br.Connect()
	if err != nil {
		return fmt.Errorf("Bridge %s failed to start: %v", br.Account, err)
	}
	err = br.JoinChannels()
	if err != nil {
		return fmt.Errorf("Bridge %s failed to join channel: %v", br.Account, err)
	}
	return nil
}

func (gw *Gateway) AddConfig(cfg *config.Gateway) error {
	if gw.Names[cfg.Name] {
		return fmt.Errorf("Gateway with name %s already exists", cfg.Name)
	}
	if cfg.Name == "" {
		return fmt.Errorf("%s", "Gateway without name found")
	}
	log.Infof("Starting gateway: %s", cfg.Name)
	gw.Names[cfg.Name] = true
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

func (gw *Gateway) Start() error {
	go gw.handleReceive()
	return nil
}

func (gw *Gateway) handleReceive() {
	for {
		select {
		case msg := <-gw.Message:
			if msg.Event == config.EVENT_FAILURE {
				for _, br := range gw.Bridges {
					if msg.Account == br.Account {
						go gw.reconnectBridge(br)
					}
				}
			}
			if msg.Event == config.EVENT_REJOIN_CHANNELS {
				for _, br := range gw.Bridges {
					if msg.Account == br.Account {
						br.Joined = make(map[string]bool)
						br.JoinChannels()
					}
				}
				continue
			}
			if !gw.ignoreMessage(&msg) {
				msg.Timestamp = time.Now()
				for _, br := range gw.Bridges {
					gw.handleMessage(msg, br)
				}
			}
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

func (gw *Gateway) mapChannels() error {
	for _, br := range append(gw.MyConfig.Out, gw.MyConfig.InOut...) {
		ID := br.Channel + br.Account
		_, ok := gw.Channels[ID]
		if !ok {
			channel := &config.ChannelInfo{Name: br.Channel, Direction: "out", ID: ID, Options: br.Options, Account: br.Account,
				GID: make(map[string]bool), SameChannel: make(map[string]bool)}
			channel.GID[gw.Name] = true
			channel.SameChannel[gw.Name] = br.SameChannel
			gw.Channels[channel.ID] = channel
		}
		gw.Channels[ID].GID[gw.Name] = true
		gw.Channels[ID].SameChannel[gw.Name] = br.SameChannel
	}

	for _, br := range append(gw.MyConfig.In, gw.MyConfig.InOut...) {
		ID := br.Channel + br.Account
		_, ok := gw.Channels[ID]
		if !ok {
			channel := &config.ChannelInfo{Name: br.Channel, Direction: "in", ID: ID, Options: br.Options, Account: br.Account,
				GID: make(map[string]bool), SameChannel: make(map[string]bool)}
			channel.GID[gw.Name] = true
			channel.SameChannel[gw.Name] = br.SameChannel
			gw.Channels[channel.ID] = channel
		}
		gw.Channels[ID].GID[gw.Name] = true
		gw.Channels[ID].SameChannel[gw.Name] = br.SameChannel
	}
	return nil
}

func (gw *Gateway) getDestChannel(msg *config.Message, dest bridge.Bridge) []config.ChannelInfo {
	var channels []config.ChannelInfo
	for _, channel := range gw.Channels {
		if _, ok := gw.Channels[getChannelID(*msg)]; !ok {
			continue
		}
		if channel.Direction == "out" && channel.Account == dest.Account && gw.validGatewayDest(*msg, channel) {
			channels = append(channels, *channel)
		}
	}
	return channels
}

func (gw *Gateway) handleMessage(msg config.Message, dest *bridge.Bridge) {
	// only relay join/part when configged
	if msg.Event == config.EVENT_JOIN_LEAVE && !gw.Bridges[dest.Account].Config.ShowJoinPart {
		return
	}
	// broadcast to every out channel (irc QUIT)
	if msg.Channel == "" && msg.Event != config.EVENT_JOIN_LEAVE {
		log.Debug("empty channel")
		return
	}
	originchannel := msg.Channel
	origmsg := msg
	for _, channel := range gw.DestChannelFunc(&msg, *dest) {
		// do not send to ourself
		if channel.ID == getChannelID(origmsg) {
			continue
		}
		log.Debugf("Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, originchannel, dest.Account, channel.Name)
		msg.Channel = channel.Name
		gw.modifyAvatar(&msg, dest)
		gw.modifyUsername(&msg, dest)
		// for api we need originchannel as channel
		if dest.Protocol == "api" {
			msg.Channel = originchannel
		}
		err := dest.Send(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
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
	return false
}

func (gw *Gateway) modifyUsername(msg *config.Message, dest *bridge.Bridge) {
	br := gw.Bridges[msg.Account]
	msg.Protocol = br.Protocol
	nick := gw.Config.General.RemoteNickFormat
	if nick == "" {
		nick = dest.Config.RemoteNickFormat
	}
	nick = strings.Replace(nick, "{NOPINGNICK}", msg.Username[:1]+"​"+msg.Username[1:], -1)
	nick = strings.Replace(nick, "{NICK}", msg.Username, -1)
	nick = strings.Replace(nick, "{BRIDGE}", br.Name, -1)
	nick = strings.Replace(nick, "{PROTOCOL}", br.Protocol, -1)
	msg.Username = nick
}

func (gw *Gateway) modifyAvatar(msg *config.Message, dest *bridge.Bridge) {
	iconurl := gw.Config.General.IconURL
	if iconurl == "" {
		iconurl = dest.Config.IconURL
	}
	iconurl = strings.Replace(iconurl, "{NICK}", msg.Username, -1)
	if msg.Avatar == "" {
		msg.Avatar = iconurl
	}
}

func getChannelID(msg config.Message) string {
	return msg.Channel + msg.Account
}

func (gw *Gateway) validGatewayDest(msg config.Message, channel *config.ChannelInfo) bool {
	GIDmap := gw.Channels[getChannelID(msg)].GID
	// check if we are running a samechannelgateway.
	// if it is and the channel name matches it's ok, otherwise we shouldn't use this channel.
	for k, _ := range GIDmap {
		if channel.SameChannel[k] == true {
			if msg.Channel == channel.Name {
				return true
			} else {
				return false
			}
		}
	}
	// check if we are in the correct gateway
	for k, _ := range GIDmap {
		if channel.GID[k] == true {
			return true
		}
	}
	return false
}
