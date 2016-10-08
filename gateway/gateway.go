package gateway

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
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
}

func New(cfg *config.Config, gateway *config.Gateway) error {
	c := make(chan config.Message)
	gw := &Gateway{}
	gw.Name = gateway.Name
	gw.Config = cfg
	gw.MyConfig = gateway
	exists := make(map[string]bool)
	for _, br := range append(gateway.In, gateway.Out...) {
		if exists[br.Account+br.Channel] {
			continue
		}
		log.Infof("Starting bridge: %s channel: %s", br.Account, br.Channel)
		gw.Bridges = append(gw.Bridges, bridge.New(cfg, &br, c))
		exists[br.Account+br.Channel] = true
	}
	gw.mapChannels()
	gw.mapIgnores()
	exists = make(map[string]bool)
	for _, br := range gw.Bridges {
		err := br.Connect()
		if err != nil {
			log.Fatalf("Bridge %s failed to start: %v", br.FullOrigin(), err)
		}
		for _, channel := range append(gw.ChannelsOut[br.FullOrigin()], gw.ChannelsIn[br.FullOrigin()]...) {
			if exists[br.FullOrigin()+channel] {
				continue
			}
			log.Infof("%s: joining %s", br.FullOrigin(), channel)
			br.JoinChannel(channel)
			exists[br.FullOrigin()+channel] = true
		}
	}
	gw.handleReceive(c)
	return nil
}

func (gw *Gateway) handleReceive(c chan config.Message) {
	for {
		select {
		case msg := <-c:
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
	return gw.ChannelsOut[dest]
}

func (gw *Gateway) handleMessage(msg config.Message, dest bridge.Bridge) {
	if gw.ignoreMessage(&msg) {
		return
	}
	channels := gw.getDestChannel(&msg, dest.FullOrigin())
	for _, channel := range channels {
		// do not send the message to the bridge we come from if also the channel is the same
		if msg.FullOrigin == dest.FullOrigin() && msg.Channel == channel {
			continue
		}
		msg.Channel = channel
		if msg.Channel == "" {
			log.Debug("empty channel")
			return
		}
		gw.modifyMessage(&msg, dest)
		log.Debugf("Sending %#v from %s to %s", msg, msg.FullOrigin, dest.FullOrigin())
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

func setNickFormat(msg *config.Message, format string) {
	if format == "" {
		msg.Username = msg.Protocol + "." + msg.Origin + "-" + msg.Username + ": "
		return
	}
	msg.Username = strings.Replace(format, "{NICK}", msg.Username, -1)
	msg.Username = strings.Replace(msg.Username, "{BRIDGE}", msg.Origin, -1)
	msg.Username = strings.Replace(msg.Username, "{PROTOCOL}", msg.Protocol, -1)
}

func (gw *Gateway) modifyMessage(msg *config.Message, dest bridge.Bridge) {
	switch dest.Protocol() {
	case "irc":
		setNickFormat(msg, gw.Config.IRC[dest.Origin()].RemoteNickFormat)
	case "gitter":
		setNickFormat(msg, gw.Config.Gitter[dest.Origin()].RemoteNickFormat)
	case "xmpp":
		setNickFormat(msg, gw.Config.Xmpp[dest.Origin()].RemoteNickFormat)
	case "mattermost":
		setNickFormat(msg, gw.Config.Mattermost[dest.Origin()].RemoteNickFormat)
	case "slack":
		setNickFormat(msg, gw.Config.Slack[dest.Origin()].RemoteNickFormat)
	case "discord":
		setNickFormat(msg, gw.Config.Discord[dest.Origin()].RemoteNickFormat)
	}
}
