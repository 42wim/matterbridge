package samechannelgateway

import (
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
)

type SameChannelGateway struct {
	*config.Config
	MyConfig *config.SameChannelGateway
	Channels []string
	Name     string
}

func New(cfg *config.Config, gatewayCfg *config.SameChannelGateway) *SameChannelGateway {
	return &SameChannelGateway{
		MyConfig: gatewayCfg,
		Channels: gatewayCfg.Channels,
		Name:     gatewayCfg.Name,
		Config:   cfg}
}

func (sgw *SameChannelGateway) Start() error {
	gw := gateway.New(sgw.Config, &config.Gateway{Name: sgw.Name})
	gw.DestChannelFunc = sgw.getDestChannel
	for _, account := range sgw.MyConfig.Accounts {
		for _, channel := range sgw.Channels {
			br := config.Bridge{Account: account, Channel: channel}
			gw.MyConfig.InOut = append(gw.MyConfig.InOut, br)
		}
	}
	return gw.Start()
}

func (sgw *SameChannelGateway) validChannel(channel string) bool {
	for _, c := range sgw.Channels {
		if c == channel {
			return true
		}
	}
	return false
}

func (sgw *SameChannelGateway) getDestChannel(msg *config.Message, dest string) []string {
	if sgw.validChannel(msg.Channel) {
		return []string{msg.Channel}
	}
	return []string{}
}
