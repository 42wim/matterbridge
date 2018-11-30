package gateway

import (
	"fmt"
	"strconv"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway/bridgemap"
	"github.com/stretchr/testify/assert"

	"testing"
)

var testconfig = []byte(`
[irc.freenode]
[mattermost.test]
[gitter.42wim]
[discord.test]
[slack.test]

[[gateway]]
    name = "bridge1"
    enable=true
    
    [[gateway.inout]]
    account = "irc.freenode"
    channel = "#wimtesting"
    
    [[gateway.inout]]
    account="gitter.42wim"
    channel="42wim/testroom"
    #channel="matterbridge/Lobby"

    [[gateway.inout]]
    account = "discord.test"
    channel = "general"
    
    [[gateway.inout]]
    account="slack.test"
    channel="testing"
	`)

var testconfig2 = []byte(`
[irc.freenode]
[mattermost.test]
[gitter.42wim]
[discord.test]
[slack.test]

[[gateway]]
    name = "bridge1"
    enable=true
    
    [[gateway.in]]
    account = "irc.freenode"
    channel = "#wimtesting"
    
    [[gateway.in]]
    account="gitter.42wim"
    channel="42wim/testroom"

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
    account="gitter.42wim"
    channel="42wim/testroom"

    [[gateway.out]]
    account = "discord.test"
    channel = "general2"
	`)

var testconfig3 = []byte(`
[irc.zzz]
[telegram.zzz]
[slack.zzz]
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
	cfg := config.NewConfigFromString(input)
	r, err := NewRouter(cfg, bridgemap.FullMap)
	if err != nil {
		fmt.Println(err)
	}
	return r
}
func TestNewRouter(t *testing.T) {
	r := maketestRouter(testconfig)
	assert.Equal(t, 1, len(r.Gateways))
	assert.Equal(t, 4, len(r.Gateways["bridge1"].Bridges))
	assert.Equal(t, 4, len(r.Gateways["bridge1"].Channels))

	r = maketestRouter(testconfig2)
	assert.Equal(t, 2, len(r.Gateways))
	assert.Equal(t, 4, len(r.Gateways["bridge1"].Bridges))
	assert.Equal(t, 3, len(r.Gateways["bridge2"].Bridges))
	assert.Equal(t, 4, len(r.Gateways["bridge1"].Channels))
	assert.Equal(t, 3, len(r.Gateways["bridge2"].Channels))
	assert.Equal(t, &config.ChannelInfo{
		Name:        "42wim/testroom",
		Direction:   "out",
		ID:          "42wim/testroomgitter.42wim",
		Account:     "gitter.42wim",
		SameChannel: map[string]bool{"bridge2": false},
	}, r.Gateways["bridge2"].Channels["42wim/testroomgitter.42wim"])
	assert.Equal(t, &config.ChannelInfo{
		Name:        "42wim/testroom",
		Direction:   "in",
		ID:          "42wim/testroomgitter.42wim",
		Account:     "gitter.42wim",
		SameChannel: map[string]bool{"bridge1": false},
	}, r.Gateways["bridge1"].Channels["42wim/testroomgitter.42wim"])
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
		case "gitter.42wim":
			assert.Equal(t, []config.ChannelInfo(nil), r.Gateways["bridge1"].getDestChannel(msg, *br))
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
