package samechannel

import (
	"io/ioutil"
	"testing"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testConfig = `
[mattermost.test]
[slack.test]

[[samechannelgateway]]
   enable = true
   name = "blah"
      accounts = [ "mattermost.test","slack.test" ]
      channels = [ "testing","testing2","testing10"]
`

var (
	expectedConfig = config.Gateway{
		Name:   "blah",
		Enable: true,
		In:     []config.Bridge(nil),
		Out:    []config.Bridge(nil),
		InOut: []config.Bridge{
			{
				Account:     "mattermost.test",
				Channel:     "testing",
				Options:     config.ChannelOptions{Key: ""},
				SameChannel: true,
			},
			{
				Account:     "mattermost.test",
				Channel:     "testing2",
				Options:     config.ChannelOptions{Key: ""},
				SameChannel: true,
			},
			{
				Account:     "mattermost.test",
				Channel:     "testing10",
				Options:     config.ChannelOptions{Key: ""},
				SameChannel: true,
			},
			{
				Account:     "slack.test",
				Channel:     "testing",
				Options:     config.ChannelOptions{Key: ""},
				SameChannel: true,
			},
			{
				Account:     "slack.test",
				Channel:     "testing2",
				Options:     config.ChannelOptions{Key: ""},
				SameChannel: true,
			},
			{
				Account:     "slack.test",
				Channel:     "testing10",
				Options:     config.ChannelOptions{Key: ""},
				SameChannel: true,
			},
		},
	}
)

func TestGetConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	cfg := config.NewConfigFromString(logger, []byte(testConfig))
	sgw := New(cfg)
	configs := sgw.GetConfig()
	assert.Equal(t, []config.Gateway{expectedConfig}, configs)
}
