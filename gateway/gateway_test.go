package gateway

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway/bridgemap"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var testconfig = []byte(`
[irc.freenode]
server=""
[mattermost.test]
server=""
[discord.test]
server=""
[slack.test]
server=""

[[gateway]]
    name = "bridge1"
    enable=true
    
    [[gateway.inout]]
    account = "irc.freenode"
    channel = "#wimtesting"
    
    [[gateway.inout]]
    account = "discord.test"
    channel = "general"
    
    [[gateway.inout]]
    account="slack.test"
    channel="testing"
	`)

var testconfig2 = []byte(`
[irc.freenode]
server=""
[mattermost.test]
server=""
[discord.test]
server=""
[slack.test]
server=""

[[gateway]]
    name = "bridge1"
    enable=true
    
    [[gateway.in]]
    account = "irc.freenode"
    channel = "#wimtesting"
    
    [[gateway.inout]]
    account = "discord.test"
    channel = "general"
    
    [[gateway.out]]
    account="slack.test"
    channel="testing"
[[gateway]]
    name = "bridge2"
    enable=true
    
    [[gateway.in]]
    account = "irc.freenode"
    channel = "#wimtesting2"
    
    [[gateway.out]]
    account = "discord.test"
    channel = "general2"
	`)

var testconfig3 = []byte(`
[irc.zzz]
server=""
[telegram.zzz]
server=""
[slack.zzz]
server=""
[[gateway]]
name="bridge"
enable=true

    [[gateway.inout]]
    account="irc.zzz"
    channel="#main"		

    [[gateway.inout]]
    account="telegram.zzz"
    channel="-1111111111111"

    [[gateway.inout]]
    account="slack.zzz"
    channel="irc"	
	
[[gateway]]
name="announcements"
enable=true
	
    [[gateway.in]]
    account="telegram.zzz"
    channel="-2222222222222"	
	
    [[gateway.out]]
    account="irc.zzz"
    channel="#main"		
	
    [[gateway.out]]
    account="irc.zzz"
    channel="#main-help"	

    [[gateway.out]]
    account="telegram.zzz"
    channel="--333333333333"	

    [[gateway.out]]
    account="slack.zzz"
    channel="general"		
	
[[gateway]]
name="bridge2"
enable=true

    [[gateway.inout]]
    account="irc.zzz"
    channel="#main-help"	

    [[gateway.inout]]
    account="telegram.zzz"
    channel="--444444444444"	

	
[[gateway]]
name="bridge3"
enable=true

    [[gateway.inout]]
    account="irc.zzz"
    channel="#main-telegram"	

    [[gateway.inout]]
    account="telegram.zzz"
    channel="--333333333333"
`)

const (
	ircTestAccount   = "irc.zzz"
	tgTestAccount    = "telegram.zzz"
	slackTestAccount = "slack.zzz"
)

func maketestRouter(input []byte) *Router {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	cfg := config.NewConfigFromString(logger, input)
	r, err := NewRouter(logger, cfg, bridgemap.FullMap)
	if err != nil {
		fmt.Println(err)
	}
	return r
}

func TestNewRouter(t *testing.T) {
	r := maketestRouter(testconfig)
	assert.Equal(t, 1, len(r.Gateways))
	assert.Equal(t, 3, len(r.Gateways["bridge1"].Bridges))
	assert.Equal(t, 3, len(r.Gateways["bridge1"].Channels))
	r = maketestRouter(testconfig2)
	assert.Equal(t, 2, len(r.Gateways))
	assert.Equal(t, 3, len(r.Gateways["bridge1"].Bridges))
	assert.Equal(t, 2, len(r.Gateways["bridge2"].Bridges))
	assert.Equal(t, 3, len(r.Gateways["bridge1"].Channels))
	assert.Equal(t, 2, len(r.Gateways["bridge2"].Channels))
	assert.Equal(t, &config.ChannelInfo{
		Name:        "general",
		Direction:   "inout",
		ID:          "generaldiscord.test",
		Account:     "discord.test",
		SameChannel: map[string]bool{"bridge1": false},
	}, r.Gateways["bridge1"].Channels["generaldiscord.test"])
}

func TestGetDestChannel(t *testing.T) {
	r := maketestRouter(testconfig2)
	msg := &config.Message{Text: "test", Channel: "general", Account: "discord.test", Gateway: "bridge1", Protocol: "discord", Username: "test"}
	for _, br := range r.Gateways["bridge1"].Bridges {
		switch br.Account {
		case "discord.test":
			assert.Equal(t, []config.ChannelInfo{{
				Name:        "general",
				Account:     "discord.test",
				Direction:   "inout",
				ID:          "generaldiscord.test",
				SameChannel: map[string]bool{"bridge1": false},
				Options:     config.ChannelOptions{Key: ""},
			}}, r.Gateways["bridge1"].getDestChannel(msg, *br))
		case "slack.test":
			assert.Equal(t, []config.ChannelInfo{{
				Name:        "testing",
				Account:     "slack.test",
				Direction:   "out",
				ID:          "testingslack.test",
				SameChannel: map[string]bool{"bridge1": false},
				Options:     config.ChannelOptions{Key: ""},
			}}, r.Gateways["bridge1"].getDestChannel(msg, *br))
		case "irc.freenode":
			assert.Equal(t, []config.ChannelInfo(nil), r.Gateways["bridge1"].getDestChannel(msg, *br))
		}
	}
}

func TestGetDestChannelAdvanced(t *testing.T) {
	r := maketestRouter(testconfig3)
	var msgs []*config.Message
	i := 0
	for _, gw := range r.Gateways {
		for _, channel := range gw.Channels {
			msgs = append(msgs, &config.Message{Text: "text" + strconv.Itoa(i), Channel: channel.Name, Account: channel.Account, Gateway: gw.Name, Username: "user" + strconv.Itoa(i)})
			i++
		}
	}
	hits := make(map[string]int)
	for _, gw := range r.Gateways {
		for _, br := range gw.Bridges {
			for _, msg := range msgs {
				channels := gw.getDestChannel(msg, *br)
				if gw.Name != msg.Gateway {
					assert.Equal(t, []config.ChannelInfo(nil), channels)
					continue
				}
				switch gw.Name {
				case "bridge":
					if (msg.Channel == "#main" || msg.Channel == "-1111111111111" || msg.Channel == "irc") &&
						(msg.Account == ircTestAccount || msg.Account == tgTestAccount || msg.Account == slackTestAccount) {
						hits[gw.Name]++
						switch br.Account {
						case ircTestAccount:
							assert.Equal(t, []config.ChannelInfo{{
								Name:        "#main",
								Account:     ircTestAccount,
								Direction:   "inout",
								ID:          "#mainirc.zzz",
								SameChannel: map[string]bool{"bridge": false},
								Options:     config.ChannelOptions{Key: ""},
							}}, channels)
						case tgTestAccount:
							assert.Equal(t, []config.ChannelInfo{{
								Name:        "-1111111111111",
								Account:     tgTestAccount,
								Direction:   "inout",
								ID:          "-1111111111111telegram.zzz",
								SameChannel: map[string]bool{"bridge": false},
								Options:     config.ChannelOptions{Key: ""},
							}}, channels)
						case slackTestAccount:
							assert.Equal(t, []config.ChannelInfo{{
								Name:        "irc",
								Account:     slackTestAccount,
								Direction:   "inout",
								ID:          "ircslack.zzz",
								SameChannel: map[string]bool{"bridge": false},
								Options:     config.ChannelOptions{Key: ""},
							}}, channels)
						}
					}
				case "bridge2":
					if (msg.Channel == "#main-help" || msg.Channel == "--444444444444") &&
						(msg.Account == ircTestAccount || msg.Account == tgTestAccount) {
						hits[gw.Name]++
						switch br.Account {
						case ircTestAccount:
							assert.Equal(t, []config.ChannelInfo{{
								Name:        "#main-help",
								Account:     ircTestAccount,
								Direction:   "inout",
								ID:          "#main-helpirc.zzz",
								SameChannel: map[string]bool{"bridge2": false},
								Options:     config.ChannelOptions{Key: ""},
							}}, channels)
						case tgTestAccount:
							assert.Equal(t, []config.ChannelInfo{{
								Name:        "--444444444444",
								Account:     tgTestAccount,
								Direction:   "inout",
								ID:          "--444444444444telegram.zzz",
								SameChannel: map[string]bool{"bridge2": false},
								Options:     config.ChannelOptions{Key: ""},
							}}, channels)
						}
					}
				case "bridge3":
					if (msg.Channel == "#main-telegram" || msg.Channel == "--333333333333") &&
						(msg.Account == ircTestAccount || msg.Account == tgTestAccount) {
						hits[gw.Name]++
						switch br.Account {
						case ircTestAccount:
							assert.Equal(t, []config.ChannelInfo{{
								Name:        "#main-telegram",
								Account:     ircTestAccount,
								Direction:   "inout",
								ID:          "#main-telegramirc.zzz",
								SameChannel: map[string]bool{"bridge3": false},
								Options:     config.ChannelOptions{Key: ""},
							}}, channels)
						case tgTestAccount:
							assert.Equal(t, []config.ChannelInfo{{
								Name:        "--333333333333",
								Account:     tgTestAccount,
								Direction:   "inout",
								ID:          "--333333333333telegram.zzz",
								SameChannel: map[string]bool{"bridge3": false},
								Options:     config.ChannelOptions{Key: ""},
							}}, channels)
						}
					}
				case "announcements":
					if msg.Channel != "-2222222222222" && msg.Account != "telegram" {
						assert.Equal(t, []config.ChannelInfo(nil), channels)
						continue
					}
					hits[gw.Name]++
					switch br.Account {
					case ircTestAccount:
						assert.Len(t, channels, 2)
						assert.Contains(t, channels, config.ChannelInfo{
							Name:        "#main",
							Account:     ircTestAccount,
							Direction:   "out",
							ID:          "#mainirc.zzz",
							SameChannel: map[string]bool{"announcements": false},
							Options:     config.ChannelOptions{Key: ""},
						})
						assert.Contains(t, channels, config.ChannelInfo{
							Name:        "#main-help",
							Account:     ircTestAccount,
							Direction:   "out",
							ID:          "#main-helpirc.zzz",
							SameChannel: map[string]bool{"announcements": false},
							Options:     config.ChannelOptions{Key: ""},
						})
					case slackTestAccount:
						assert.Equal(t, []config.ChannelInfo{{
							Name:        "general",
							Account:     slackTestAccount,
							Direction:   "out",
							ID:          "generalslack.zzz",
							SameChannel: map[string]bool{"announcements": false},
							Options:     config.ChannelOptions{Key: ""},
						}}, channels)
					case tgTestAccount:
						assert.Equal(t, []config.ChannelInfo{{
							Name:        "--333333333333",
							Account:     tgTestAccount,
							Direction:   "out",
							ID:          "--333333333333telegram.zzz",
							SameChannel: map[string]bool{"announcements": false},
							Options:     config.ChannelOptions{Key: ""},
						}}, channels)
					}
				}
			}
		}
	}
	assert.Equal(t, map[string]int{"bridge3": 4, "bridge": 9, "announcements": 3, "bridge2": 4}, hits)
}

type ignoreTestSuite struct {
	suite.Suite

	gw *Gateway
}

func TestIgnoreSuite(t *testing.T) {
	s := &ignoreTestSuite{}
	suite.Run(t, s)
}

func (s *ignoreTestSuite) SetupSuite() {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	s.gw = &Gateway{logger: logrus.NewEntry(logger)}
}

func (s *ignoreTestSuite) TestIgnoreTextEmpty() {
	extraFile := make(map[string][]interface{})
	extraAttach := make(map[string][]interface{})
	extraFailure := make(map[string][]interface{})
	extraFile["file"] = append(extraFile["file"], config.FileInfo{})
	extraAttach["attachments"] = append(extraAttach["attachments"], []string{})
	extraFailure[config.EventFileFailureSize] = append(extraFailure[config.EventFileFailureSize], config.FileInfo{})

	msgTests := map[string]struct {
		input  *config.Message
		output bool
	}{
		"usertyping": {
			input:  &config.Message{Event: config.EventUserTyping},
			output: false,
		},
		"file attach": {
			input:  &config.Message{Extra: extraFile},
			output: false,
		},
		"attachments": {
			input:  &config.Message{Extra: extraAttach},
			output: false,
		},
		config.EventFileFailureSize: {
			input:  &config.Message{Extra: extraFailure},
			output: false,
		},
		"nil extra": {
			input:  &config.Message{Extra: nil},
			output: true,
		},
		"empty": {
			input:  &config.Message{},
			output: true,
		},
	}
	for testname, testcase := range msgTests {
		output := s.gw.ignoreTextEmpty(testcase.input)
		s.Assert().Equalf(testcase.output, output, "case '%s' failed", testname)
	}
}

func (s *ignoreTestSuite) TestIgnoreTexts() {
	msgTests := map[string]struct {
		input  string
		re     []string
		output bool
	}{
		"no regex": {
			input:  "a text message",
			re:     []string{},
			output: false,
		},
		"simple regex": {
			input:  "a text message",
			re:     []string{"text"},
			output: true,
		},
		"multiple regex fail": {
			input:  "a text message",
			re:     []string{"abc", "123$"},
			output: false,
		},
		"multiple regex pass": {
			input:  "a text message",
			re:     []string{"lala", "sage$"},
			output: true,
		},
	}
	for testname, testcase := range msgTests {
		output := s.gw.ignoreText(testcase.input, testcase.re)
		s.Assert().Equalf(testcase.output, output, "case '%s' failed", testname)
	}
}

func (s *ignoreTestSuite) TestIgnoreNicks() {
	msgTests := map[string]struct {
		input  string
		re     []string
		output bool
	}{
		"no entry": {
			input:  "user",
			re:     []string{},
			output: false,
		},
		"one entry": {
			input:  "user",
			re:     []string{"user"},
			output: true,
		},
		"multiple entries": {
			input:  "user",
			re:     []string{"abc", "user"},
			output: true,
		},
		"multiple entries fail": {
			input:  "user",
			re:     []string{"abc", "def"},
			output: false,
		},
	}
	for testname, testcase := range msgTests {
		output := s.gw.ignoreText(testcase.input, testcase.re)
		s.Assert().Equalf(testcase.output, output, "case '%s' failed", testname)
	}
}

func BenchmarkTengo(b *testing.B) {
	msg := &config.Message{Username: "user", Text: "blah testing", Account: "protocol.account", Channel: "mychannel"}
	for n := 0; n < b.N; n++ {
		err := modifyInMessageTengo("bench.tengo", msg)
		if err != nil {
			return
		}
	}
}
