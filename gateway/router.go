package gateway

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway/samechannel"
	log "github.com/Sirupsen/logrus"
	//	"github.com/davecgh/go-spew/spew"
	"time"
)

type Router struct {
	Gateways map[string]*Gateway
	Message  chan config.Message
	*config.Config
}

func NewRouter(cfg *config.Config) (*Router, error) {
	r := &Router{}
	r.Config = cfg
	r.Message = make(chan config.Message)
	r.Gateways = make(map[string]*Gateway)
	sgw := samechannelgateway.New(cfg)
	gwconfigs := sgw.GetConfig()

	for _, entry := range append(gwconfigs, cfg.Gateway...) {
		if !entry.Enable {
			continue
		}
		if entry.Name == "" {
			return nil, fmt.Errorf("%s", "Gateway without name found")
		}
		if _, ok := r.Gateways[entry.Name]; ok {
			return nil, fmt.Errorf("Gateway with name %s already exists", entry.Name)
		}
		r.Gateways[entry.Name] = New(entry, r)
	}
	return r, nil
}

func (r *Router) Start() error {
	m := make(map[string]*bridge.Bridge)
	for _, gw := range r.Gateways {
		for _, br := range gw.Bridges {
			m[br.Account] = br
		}
	}
	for _, br := range m {
		log.Infof("Starting bridge: %s ", br.Account)
		err := br.Connect()
		if err != nil {
			return fmt.Errorf("Bridge %s failed to start: %v", br.Account, err)
		}
		err = br.JoinChannels()
		if err != nil {
			return fmt.Errorf("Bridge %s failed to join channel: %v", br.Account, err)
		}
	}
	go r.handleReceive()
	return nil
}

func (r *Router) getBridge(account string) *bridge.Bridge {
	for _, gw := range r.Gateways {
		if br, ok := gw.Bridges[account]; ok {
			return br
		}
	}
	return nil
}

func (r *Router) getGatewayName(channelID string) []string {
	var names []string
	for _, gw := range r.Gateways {
		if _, ok := gw.Channels[channelID]; ok {
			names = append(names, gw.Name)
		}
	}
	return names
}

func (r *Router) handleReceive() {
	for msg := range r.Message {
		if msg.Event == config.EVENT_FAILURE {
		Loop:
			for _, gw := range r.Gateways {
				for _, br := range gw.Bridges {
					if msg.Account == br.Account {
						go gw.reconnectBridge(br)
						break Loop
					}
				}
			}
		}
		if msg.Event == config.EVENT_REJOIN_CHANNELS {
			for _, gw := range r.Gateways {
				for _, br := range gw.Bridges {
					if msg.Account == br.Account {
						br.Joined = make(map[string]bool)
						br.JoinChannels()
					}
				}
			}
		}
		for _, gw := range r.Gateways {
			if !gw.ignoreMessage(&msg) {
				msg.Timestamp = time.Now()
				gw.modifyMessage(&msg)
				for _, br := range gw.Bridges {
					gw.handleMessage(msg, br)
				}
			}
		}
	}
}
