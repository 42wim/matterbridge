package gateway

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"reflect"
	"strings"
	"time"
)

type Gateway struct {
	*config.Config
	MyConfig        *config.Gateway
	Bridges         map[string]*bridge.Bridge
	ChannelsOut     map[string][]string
	ChannelsIn      map[string][]string
	ChannelOptions  map[string]config.ChannelOptions
	Name            string
	Message         chan config.Message
	DestChannelFunc func(msg *config.Message, dest string) []string
}

func New(cfg *config.Config, gateway *config.Gateway) *Gateway {
	gw := &Gateway{}
	gw.Name = gateway.Name
	gw.Config = cfg
	gw.MyConfig = gateway
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
	gw.mapChannelsToBridge(br, gw.ChannelsOut)
	gw.mapChannelsToBridge(br, gw.ChannelsIn)
	gw.Bridges[cfg.Account] = br
	err := br.Connect()
	if err != nil {
		return fmt.Errorf("Bridge %s failed to start: %v", br.Account, err)
	}
	br.JoinChannels()
	return nil
}

func (gw *Gateway) mapChannelsToBridge(br *bridge.Bridge, cMap map[string][]string) {
	for _, channel := range cMap[br.Account] {
		if _, ok := gw.ChannelOptions[br.Account+channel]; ok {
			br.ChannelsOut[channel] = gw.ChannelOptions[br.Account+channel]
		} else {
			br.ChannelsOut[channel] = config.ChannelOptions{}
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
	options := make(map[string]config.ChannelOptions)
	m := make(map[string][]string)
	for _, br := range gw.MyConfig.Out {
		m[br.Account] = append(m[br.Account], br.Channel)
		options[br.Account+br.Channel] = br.Options
	}
	gw.ChannelsOut = m
	m = nil
	m = make(map[string][]string)
	for _, br := range gw.MyConfig.In {
		m[br.Account] = append(m[br.Account], br.Channel)
		options[br.Account+br.Channel] = br.Options
	}
	gw.ChannelsIn = m
	for _, br := range gw.MyConfig.InOut {
		gw.ChannelsIn[br.Account] = append(gw.ChannelsIn[br.Account], br.Channel)
		gw.ChannelsOut[br.Account] = append(gw.ChannelsOut[br.Account], br.Channel)
		options[br.Account+br.Channel] = br.Options
	}
	gw.ChannelOptions = options
	return nil
}

func (gw *Gateway) getDestChannel(msg *config.Message, dest string) []string {
	channels := gw.ChannelsIn[msg.Account]
	// broadcast to every out channel (irc QUIT)
	if msg.Event == config.EVENT_JOIN_LEAVE && msg.Channel == "" {
		return gw.ChannelsOut[dest]
	}
	for _, channel := range channels {
		if channel == msg.Channel {
			return gw.ChannelsOut[dest]
		}
	}
	return []string{}
}

func (gw *Gateway) handleMessage(msg config.Message, dest *bridge.Bridge) {
	// only relay join/part when configged
	if msg.Event == config.EVENT_JOIN_LEAVE && !gw.Bridges[dest.Account].Config.ShowJoinPart {
		return
	}
	originchannel := msg.Channel
	channels := gw.DestChannelFunc(&msg, dest.Account)
	for _, channel := range channels {
		// do not send the message to the bridge we come from if also the channel is the same
		if msg.Account == dest.Account && channel == originchannel {
			continue
		}
		msg.Channel = channel
		if msg.Channel == "" {
			log.Debug("empty channel")
			return
		}
		log.Debugf("Sending %#v from %s (%s) to %s (%s)", msg, msg.Account, originchannel, dest.Account, channel)
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

func (gw *Gateway) modifyMessage(msg *config.Message, dest *bridge.Bridge) {
	val := reflect.ValueOf(gw.Config).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		// look for the protocol map (both lowercase)
		if strings.ToLower(typeField.Name) == dest.Protocol {
			// get the Protocol struct from the map
			protoCfg := val.Field(i).MapIndex(reflect.ValueOf(dest.Name))
			//config.SetNickFormat(msg, protoCfg.Interface().(config.Protocol))
			val.Field(i).SetMapIndex(reflect.ValueOf(dest.Name), protoCfg)
			break
		}
	}
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
