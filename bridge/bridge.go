package bridge

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/sirupsen/logrus"
)

type Bridger interface {
	Send(msg config.Message) (string, error)
	Connect() error
	JoinChannel(channel config.ChannelInfo) error
	Disconnect() error
}

type Bridge struct {
	Bridger
	*sync.RWMutex

	Name           string
	Account        string
	Protocol       string
	Channels       map[string]config.ChannelInfo
	Joined         map[string]bool
	ChannelMembers *config.ChannelMembers
	Log            *logrus.Entry
	Config         config.Config
	General        *config.Protocol
}

type Config struct {
	*Bridge

	Remote chan config.Message
}

// Factory is the factory function to create a bridge
type Factory func(*Config) Bridger

func New(bridge *config.Bridge) *Bridge {
	accInfo := strings.Split(bridge.Account, ".")
	if len(accInfo) != 2 {
		log.Fatalf("config failure, account incorrect: %s", bridge.Account)
	}

	protocol := accInfo[0]
	name := accInfo[1]

	return &Bridge{
		RWMutex:  new(sync.RWMutex),
		Channels: make(map[string]config.ChannelInfo),
		Name:     name,
		Protocol: protocol,
		Account:  bridge.Account,
		Joined:   make(map[string]bool),
	}
}

func (b *Bridge) JoinChannels() error {
	return b.joinChannels(b.Channels, b.Joined)
}

// SetChannelMembers sets the newMembers to the bridge ChannelMembers
func (b *Bridge) SetChannelMembers(newMembers *config.ChannelMembers) {
	b.Lock()
	b.ChannelMembers = newMembers
	b.Unlock()
}

func (b *Bridge) joinChannels(channels map[string]config.ChannelInfo, exists map[string]bool) error {
	for ID, channel := range channels {
		if !exists[ID] {
			b.Log.Infof("%s: joining %s (ID: %s)", b.Account, channel.Name, ID)
			time.Sleep(time.Duration(b.GetInt("JoinDelay")) * time.Millisecond)
			err := b.JoinChannel(channel)
			if err != nil {
				return err
			}
			exists[ID] = true
		}
	}
	return nil
}

func (b *Bridge) GetConfigKey(key string) string {
	return b.Account + "." + key
}

func (b *Bridge) IsKeySet(key string) bool {
	return b.Config.IsKeySet(b.GetConfigKey(key)) || b.Config.IsKeySet("general."+key)
}

func (b *Bridge) GetBool(key string) bool {
	val, ok := b.Config.GetBool(b.GetConfigKey(key))
	if !ok {
		val, _ = b.Config.GetBool("general." + key)
	}
	return val
}

func (b *Bridge) GetInt(key string) int {
	val, ok := b.Config.GetInt(b.GetConfigKey(key))
	if !ok {
		val, _ = b.Config.GetInt("general." + key)
	}
	return val
}

func (b *Bridge) GetString(key string) string {
	val, ok := b.Config.GetString(b.GetConfigKey(key))
	if !ok {
		val, _ = b.Config.GetString("general." + key)
	}
	return val
}

func (b *Bridge) GetStringSlice(key string) []string {
	val, ok := b.Config.GetStringSlice(b.GetConfigKey(key))
	if !ok {
		val, _ = b.Config.GetStringSlice("general." + key)
	}
	return val
}

func (b *Bridge) GetStringSlice2D(key string) [][]string {
	val, ok := b.Config.GetStringSlice2D(b.GetConfigKey(key))
	if !ok {
		val, _ = b.Config.GetStringSlice2D("general." + key)
	}
	return val
}
