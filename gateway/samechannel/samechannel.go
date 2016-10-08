package samechannelgateway

import (
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
	"strings"
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
	// do not send the message to the bridge we come from if also the channel is the same
	if msg.FullOrigin == dest.FullOrigin() {
		return
	}
	gw.modifyMessage(&msg, dest)
	log.Debugf("Sending %#v from %s to %s", msg, msg.FullOrigin, dest.FullOrigin())
	err := dest.Send(msg)
	if err != nil {
		log.Error(err)
	}
}

func setNickFormat(msg *config.Message, format string) {
	if format == "" {
		msg.Username = msg.Protocol + "." + msg.Origin + "-" + msg.Username + ": "
		return
	}
	msg.Username = strings.Replace(format, "{NICK}", msg.Username, -1)
	msg.Username = strings.Replace(msg.Username, "{BRIDGE}", msg.Origin, -1)
	msg.Username = strings.Replace(msg.Username, "{PROTOCOL}", msg.Protocol, -1)
}

func (gw *SameChannelGateway) modifyMessage(msg *config.Message, dest bridge.Bridge) {
	switch dest.Protocol() {
	case "irc":
		setNickFormat(msg, gw.Config.IRC[dest.Origin()].RemoteNickFormat)
	case "mattermost":
		setNickFormat(msg, gw.Config.Mattermost[dest.Origin()].RemoteNickFormat)
	case "slack":
		setNickFormat(msg, gw.Config.Slack[dest.Origin()].RemoteNickFormat)
	case "discord":
		setNickFormat(msg, gw.Config.Discord[dest.Origin()].RemoteNickFormat)
	}
}
