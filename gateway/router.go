package gateway

import (
	"fmt"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	samechannelgateway "github.com/42wim/matterbridge/gateway/samechannel"
)

type Router struct {
	config.Config

	BridgeMap        map[string]bridge.Factory
	Gateways         map[string]*Gateway
	Message          chan config.Message
	MattermostPlugin chan config.Message
}

func NewRouter(cfg config.Config, bridgeMap map[string]bridge.Factory) (*Router, error) {
	r := &Router{
		Config:           cfg,
		BridgeMap:        bridgeMap,
		Message:          make(chan config.Message),
		MattermostPlugin: make(chan config.Message),
		Gateways:         make(map[string]*Gateway),
	}
	sgw := samechannelgateway.New(cfg)
	gwconfigs := sgw.GetConfig()

	for _, entry := range append(gwconfigs, cfg.BridgeValues().Gateway...) {
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
		flog.Infof("Parsing gateway %s", gw.Name)
		for _, br := range gw.Bridges {
			m[br.Account] = br
		}
	}
	for _, br := range m {
		flog.Infof("Starting bridge: %s ", br.Account)
		err := br.Connect()
		if err != nil {
			e := fmt.Errorf("Bridge %s failed to start: %v", br.Account, err)
			if r.disableBridge(br, e) {
				continue
			}
			return e
		}
		err = br.JoinChannels()
		if err != nil {
			e := fmt.Errorf("Bridge %s failed to join channel: %v", br.Account, err)
			if r.disableBridge(br, e) {
				continue
			}
			return e
		}
	}
	// remove unused bridges
	for _, gw := range r.Gateways {
		for i, br := range gw.Bridges {
			if br.Bridger == nil {
				flog.Errorf("removing failed bridge %s", i)
				delete(gw.Bridges, i)
			}
		}
	}
	go r.handleReceive()
	return nil
}

// disableBridge returns true and empties a bridge if we have IgnoreFailureOnStart configured
// otherwise returns false
func (r *Router) disableBridge(br *bridge.Bridge, err error) bool {
	if r.BridgeValues().General.IgnoreFailureOnStart {
		flog.Error(err)
		// setting this bridge empty
		*br = bridge.Bridge{}
		return true
	}
	return false
}

func (r *Router) getBridge(account string) *bridge.Bridge {
	for _, gw := range r.Gateways {
		if br, ok := gw.Bridges[account]; ok {
			return br
		}
	}
	return nil
}

func (r *Router) handleReceive() {
	for msg := range r.Message {
		msg := msg // scopelint
		if msg.Event == config.EventFailure {
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
		if msg.Event == config.EventRejoinChannels {
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
			// record all the message ID's of the different bridges
			var msgIDs []*BrMsgID
			if !gw.ignoreMessage(&msg) {
				msg.Timestamp = time.Now()
				gw.modifyMessage(&msg)
				gw.handleFiles(&msg)
				for _, br := range gw.Bridges {
					msgIDs = append(msgIDs, gw.handleMessage(msg, br)...)
				}
				// only add the message ID if it doesn't already exists
				if _, ok := gw.Messages.Get(msg.Protocol + " " + msg.ID); !ok && msg.ID != "" {
					gw.Messages.Add(msg.Protocol+" "+msg.ID, msgIDs)
				}
			}
		}
	}
}
