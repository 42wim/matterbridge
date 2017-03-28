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
	Name            string
	Message         chan config.Message
	DestChannelFunc func(msg *config.Message, dest string) []config.ChannelInfo
}

func New(cfg *config.Config, gateway *config.Gateway) *Gateway {
	gw := &Gateway{}
	gw.Name = gateway.Name
	gw.Config = cfg
	gw.MyConfig = gateway
	gw.Channels = make(map[string]*config.ChannelInfo)
	gw.Message = make(chan config.Message)
	gw.Bridges = make(map[string]*bridge.Bridge)
	gw.DestChannelFunc = gw.getDestChannel
	return gw
}

func (gw *Gateway) AddBridge(cfg *config.Bridge) error {
	for _, br := range gw.Bridges {
		if br.Account == cfg.Account {
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

func (gw *Gateway) mapChannelsToBridge(br *bridge.Bridge) {
	for ID, channel := range gw.Channels {
		if br.Account == channel.Account {
			br.Channels[ID] = *channel
		}
	}
}

func (gw *Gateway) Start() error {
	gw.mapChannels()
	for _, br := range append(gw.MyConfig.In, append(gw.MyConfig.InOut, gw.MyConfig.Out...)...) {
		err := gw.AddBridge(&br)
		if err != nil {
			return err
		}
	}
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
	br.JoinChannels()
}

func (gw *Gateway) mapChannels() error {
	gw.Channels = make(map[string]*config.ChannelInfo)
	for _, br := range append(gw.MyConfig.Out, gw.MyConfig.InOut...) {
		ID := br.Channel + br.Account
		_, ok := gw.Channels[ID]
		if !ok {
			channel := &config.ChannelInfo{Name: br.Channel, Direction: "out", ID: ID, Options: br.Options, Account: br.Account}
			gw.Channels[channel.ID] = channel
		}
	}

	for _, br := range append(gw.MyConfig.In, gw.MyConfig.InOut...) {
		ID := br.Channel + br.Account
		_, ok := gw.Channels[ID]
		if !ok {
			channel := &config.ChannelInfo{Name: br.Channel, Direction: "in", ID: ID, Options: br.Options, Account: br.Account}
			gw.Channels[channel.ID] = channel
		}
	}
	return nil
}

func (gw *Gateway) getDestChannel(msg *config.Message, dest string) []config.ChannelInfo {
	var channels []config.ChannelInfo
	for _, channel := range gw.Channels {
		if channel.Direction == "out" && channel.Account == dest {
			channels = append(channels, *channel)
		}
	}
	return channels
}

func (gw *Gateway) handleMessage(msg config.Message, dest *bridge.Bridge) {
	// broadcast to every out channel (irc QUIT)
	if msg.Channel == "" && msg.Event != config.EVENT_JOIN_LEAVE {
		log.Debug("empty channel")
		return
	}
	originchannel := msg.Channel
	for _, channel := range gw.DestChannelFunc(&msg, dest.Account) {
		// do not send to ourself
		if channel.ID == getChannelID(msg) {
			continue
		}
		// outgoing channels for this account
		//if channel.Direction == "out" && channel.Account == dest.Account {
		log.Debugf("Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, originchannel, dest.Account, channel.Name)
		msg.Channel = channel.Name
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
	nick = strings.Replace(nick, "{NICK}", msg.Username, -1)
	nick = strings.Replace(nick, "{BRIDGE}", br.Name, -1)
	nick = strings.Replace(nick, "{PROTOCOL}", br.Protocol, -1)
	msg.Username = nick
}

func getChannelID(msg config.Message) string {
	return msg.Channel + msg.Account
}
