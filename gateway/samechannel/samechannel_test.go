package samechannelgateway

import (
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"

	"testing"
)

var testconfig = `
[mattermost.test]
[slack.test]

[[samechannelgateway]]
   enable = true
   name = "blah"
      accounts = [ "mattermost.test","slack.test" ]
      channels = [ "testing","testing2","testing10"]
`

func TestGetConfig(t *testing.T) {
	var cfg *config.Config
	if _, err := toml.Decode(testconfig, &cfg); err != nil {
		fmt.Println(err)
	}
	sgw := New(cfg)
	configs := sgw.GetConfig()
	assert.Equal(t, []config.Gateway{{Name: "blah", Enable: true, In: []config.Bridge(nil), Out: []config.Bridge(nil), InOut: []config.Bridge{{Account: "mattermost.test", Channel: "testing", Options: config.ChannelOptions{Key: ""}, SameChannel: true}, {Account: "mattermost.test", Channel: "testing2", Options: config.ChannelOptions{Key: ""}, SameChannel: true}, {Account: "mattermost.test", Channel: "testing10", Options: config.ChannelOptions{Key: ""}, SameChannel: true}, {Account: "slack.test", Channel: "testing", Options: config.ChannelOptions{Key: ""}, SameChannel: true}, {Account: "slack.test", Channel: "testing2", Options: config.ChannelOptions{Key: ""}, SameChannel: true}, {Account: "slack.test", Channel: "testing10", Options: config.ChannelOptions{Key: ""}, SameChannel: true}}}}, configs)
}
