package gateway

import (
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
	for _, br := range gateway.In {
		gw.Bridges = append(gw.Bridges, bridge.New(cfg, &br, c))
	}
	gw.mapChannels()
	gw.mapIgnores()
	for _, br := range gw.Bridges {
		br.Connect()
		for _, channel := range gw.ChannelsOut[br.FullOrigin()] {
			br.JoinChannel(channel)
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
			log.Debug("continue", msg.Protocol, msg.Origin, dest.Protocol(), dest.Origin())
			continue
		}
		msg.Channel = channel
		if msg.Channel == "" {
			log.Debug("empty channel")
			return
		}
		gw.modifyMessage(&msg, dest)
		log.Debugf("sending %#v from %s to %s", msg, msg.Origin, dest.Origin())
		dest.Send(msg)
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
	}
}
