package bridge

import (
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/sirupsen/logrus"

	"strings"
)

type Bridger interface {
	Send(msg config.Message) (string, error)
	Connect() error
	JoinChannel(channel config.ChannelInfo) error
	Disconnect() error
}

type Bridge struct {
	Config config.Protocol
	Bridger
	Name     string
	Account  string
	Protocol string
	Channels map[string]config.ChannelInfo
	Joined   map[string]bool
	Log      *log.Entry
}

// Factory is the factory function to create a bridge
type Factory func(*config.BridgeConfig) Bridger

func New(bridge *config.Bridge) *Bridge {
	b := new(Bridge)
	b.Channels = make(map[string]config.ChannelInfo)
	accInfo := strings.Split(bridge.Account, ".")
	protocol := accInfo[0]
	name := accInfo[1]
	b.Name = name
	b.Protocol = protocol
	b.Account = bridge.Account
	b.Joined = make(map[string]bool)
	return b
}

func (b *Bridge) JoinChannels() error {
	err := b.joinChannels(b.Channels, b.Joined)
	return err
}

func (b *Bridge) joinChannels(channels map[string]config.ChannelInfo, exists map[string]bool) error {
	for ID, channel := range channels {
		if !exists[ID] {
			b.Log.Infof("%s: joining %s (ID: %s)", b.Account, channel.Name, ID)
			err := b.JoinChannel(channel)
			if err != nil {
				return err
			}
			exists[ID] = true
		}
	}
	return nil
}
