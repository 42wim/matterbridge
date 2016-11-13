package samechannelgateway

import (
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
)

type SameChannelGateway struct {
	*config.Config
	MyConfig    *config.SameChannelGateway
	Bridges     []bridge.Bridge
	Channels    []string
	ignoreNicks map[string][]string
	Name        string
}

func New(cfg *config.Config, gateway *config.SameChannelGateway) error {
	c := make(chan config.Message)
	gw := &SameChannelGateway{}
	gw.Name = gateway.Name
	gw.Config = cfg
	gw.MyConfig = gateway
	gw.Channels = gateway.Channels
	for _, account := range gateway.Accounts {
		br := config.Bridge{Account: account}
		log.Infof("Starting bridge: %s", account)
		gw.Bridges = append(gw.Bridges, bridge.New(cfg, &br, c))
	}
	for _, br := range gw.Bridges {
		err := br.Connect()
		if err != nil {
			log.Fatalf("Bridge %s failed to start: %v", br.FullOrigin(), err)
		}
		for _, channel := range gw.Channels {
			log.Infof("%s: joining %s", br.FullOrigin(), channel)
			br.JoinChannel(channel)
		}
	}
	gw.handleReceive(c)
	return nil
}

func (gw *SameChannelGateway) handleReceive(c chan config.Message) {
	for {
		select {
		case msg := <-c:
			for _, br := range gw.Bridges {
				gw.handleMessage(msg, br)
			}
		}
	}
}

func (gw *SameChannelGateway) handleMessage(msg config.Message, dest bridge.Bridge) {
	// is this a configured channel
	if !gw.validChannel(msg.Channel) {
		return
	}
	// do not send the message to the bridge we come from if also the channel is the same
	if msg.FullOrigin == dest.FullOrigin() {
		return
	}
	log.Debugf("Sending %#v from %s (%s) to %s (%s)", msg, msg.FullOrigin, msg.Channel, dest.FullOrigin(), msg.Channel)
	err := dest.Send(msg)
	if err != nil {
		log.Error(err)
	}
}

func (gw *SameChannelGateway) validChannel(channel string) bool {
	for _, c := range gw.Channels {
		if c == channel {
			return true
		}
	}
	return false
}
