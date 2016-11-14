package gateway

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"reflect"
	"strings"
)

type Gateway struct {
	*config.Config
	MyConfig *config.Gateway
	//Bridges     []*bridge.Bridge
	Bridges     map[string]*bridge.Bridge
	ChannelsOut map[string][]string
	ChannelsIn  map[string][]string
	ignoreNicks map[string][]string
	Name        string
	Message     chan config.Message
}

func New(cfg *config.Config, gateway *config.Gateway) *Gateway {
	gw := &Gateway{}
	gw.Name = gateway.Name
	gw.Config = cfg
	gw.MyConfig = gateway
	gw.Message = make(chan config.Message)
	gw.Bridges = make(map[string]*bridge.Bridge)
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
	gw.Bridges[cfg.Account] = br
	err := br.Connect()
	if err != nil {
		return fmt.Errorf("Bridge %s failed to start: %v", br.Account, err)
	}
	exists := make(map[string]bool)
	for _, channel := range append(gw.ChannelsOut[br.Account], gw.ChannelsIn[br.Account]...) {
		if !exists[br.Account+channel] {
			log.Infof("%s: joining %s", br.Account, channel)
			br.JoinChannel(channel)
			exists[br.Account+channel] = true
		}
	}
	return nil
}

func (gw *Gateway) Start() error {
	gw.mapChannels()
	for _, br := range append(gw.MyConfig.In, gw.MyConfig.Out...) {
		err := gw.AddBridge(&br)
		if err != nil {
			return err
		}
	}
	//TODO fix mapIgnores
	//gw.mapIgnores()
	go gw.handleReceive()
	return nil
}

func (gw *Gateway) handleReceive() {
	for {
		select {
		case msg := <-gw.Message:
			for _, br := range gw.Bridges {
				gw.handleMessage(msg, br)
			}
		}
	}
}

func (gw *Gateway) mapChannels() error {
	m := make(map[string][]string)
	for _, br := range gw.MyConfig.Out {
		m[br.Account] = append(m[br.Account], br.Channel)
	}
	gw.ChannelsOut = m
	m = nil
	m = make(map[string][]string)
	for _, br := range gw.MyConfig.In {
		m[br.Account] = append(m[br.Account], br.Channel)
	}
	gw.ChannelsIn = m
	return nil
}

func (gw *Gateway) mapIgnores() {
	m := make(map[string][]string)
	for _, br := range gw.MyConfig.In {
		accInfo := strings.Split(br.Account, ".")
		m[br.Account] = strings.Fields(gw.Config.IRC[accInfo[1]].IgnoreNicks)
	}
	gw.ignoreNicks = m
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
	if gw.ignoreMessage(&msg) {
		return
	}
	// only relay join/part when configged
	if msg.Event == config.EVENT_JOIN_LEAVE && !gw.Bridges[dest.Account].Config.ShowJoinPart {
		return
	}
	originchannel := msg.Channel
	channels := gw.getDestChannel(&msg, dest.Account)
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
		err := dest.Send(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
	// should we discard messages ?
	for _, entry := range gw.ignoreNicks[msg.Account] {
		if msg.Username == entry {
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
	nick := dest.Config.RemoteNickFormat
	nick = strings.Replace(nick, "{NICK}", msg.Username, -1)
	nick = strings.Replace(nick, "{BRIDGE}", br.Name, -1)
	nick = strings.Replace(nick, "{PROTOCOL}", br.Protocol, -1)
	msg.Username = nick
}
