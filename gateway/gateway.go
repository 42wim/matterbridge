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
	MyConfig    *config.Gateway
	Bridges     []bridge.Bridge
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
	return gw
}

func (gw *Gateway) AddBridge(cfg *config.Bridge) error {
	for _, br := range gw.Bridges {
		if br.FullOrigin() == cfg.Account {
			return nil
		}
	}
	log.Infof("Starting bridge: %s ", cfg.Account)
	br := bridge.New(gw.Config, cfg, gw.Message)
	gw.Bridges = append(gw.Bridges, br)
	err := br.Connect()
	if err != nil {
		return fmt.Errorf("Bridge %s failed to start: %v", br.FullOrigin(), err)
	}
	exists := make(map[string]bool)
	for _, channel := range append(gw.ChannelsOut[br.FullOrigin()], gw.ChannelsIn[br.FullOrigin()]...) {
		if !exists[br.FullOrigin()+channel] {
			log.Infof("%s: joining %s", br.FullOrigin(), channel)
			br.JoinChannel(channel)
			exists[br.FullOrigin()+channel] = true
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
	channels := gw.ChannelsIn[msg.FullOrigin]
	for _, channel := range channels {
		if channel == msg.Channel {
			return gw.ChannelsOut[dest]
		}
	}
	return []string{}
}

func (gw *Gateway) handleMessage(msg config.Message, dest bridge.Bridge) {
	if gw.ignoreMessage(&msg) {
		return
	}
	originchannel := msg.Channel
	channels := gw.getDestChannel(&msg, dest.FullOrigin())
	for _, channel := range channels {
		// do not send the message to the bridge we come from if also the channel is the same
		if msg.FullOrigin == dest.FullOrigin() && channel == originchannel {
			continue
		}
		msg.Channel = channel
		if msg.Channel == "" {
			log.Debug("empty channel")
			return
		}
		log.Debugf("Sending %#v from %s (%s) to %s (%s)", msg, msg.FullOrigin, originchannel, dest.FullOrigin(), channel)
		err := dest.Send(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (gw *Gateway) ignoreMessage(msg *config.Message) bool {
	// should we discard messages ?
	for _, entry := range gw.ignoreNicks[msg.FullOrigin] {
		if msg.Username == entry {
			return true
		}
	}
	return false
}

func (gw *Gateway) modifyMessage(msg *config.Message, dest bridge.Bridge) {
	val := reflect.ValueOf(gw.Config).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		// look for the protocol map (both lowercase)
		if strings.ToLower(typeField.Name) == dest.Protocol() {
			// get the Protocol struct from the map
			protoCfg := val.Field(i).MapIndex(reflect.ValueOf(dest.Origin()))
			//config.SetNickFormat(msg, protoCfg.Interface().(config.Protocol))
			val.Field(i).SetMapIndex(reflect.ValueOf(dest.Origin()), protoCfg)
			break
		}
	}
}
