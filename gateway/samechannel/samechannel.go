package samechannel

import (
	"github.com/42wim/matterbridge/bridge/config"
)

type SameChannelGateway struct {
	config.Config
}

func New(cfg config.Config) *SameChannelGateway {
	return &SameChannelGateway{Config: cfg}
}

func (sgw *SameChannelGateway) GetConfig() []config.Gateway {
	var gwconfigs []config.Gateway
	cfg := sgw.Config
	for _, gw := range cfg.BridgeValues().SameChannelGateway {
		gwconfig := config.Gateway{Name: gw.Name, Enable: gw.Enable}
		for _, account := range gw.Accounts {
			for _, channel := range gw.Channels {
				gwconfig.InOut = append(gwconfig.InOut, config.Bridge{Account: account, Channel: channel, SameChannel: true})
			}
		}
		gwconfigs = append(gwconfigs, gwconfig)
	}
	return gwconfigs
}
