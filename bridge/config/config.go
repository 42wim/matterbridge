package config

import (
	"gopkg.in/gcfg.v1"
	"io/ioutil"
	"log"
)

type Message struct {
	Text     string
	Channel  string
	Username string
	Origin   string
}

type Config struct {
	IRC struct {
		UseTLS           bool
		UseSASL          bool
		SkipTLSVerify    bool
		Server           string
		Nick             string
		Password         string
		Channel          string
		NickServNick     string
		NickServPassword string
		RemoteNickFormat string
		IgnoreNicks      string
		Enable           bool
	}
	Gitter struct {
		Enable           bool
		IgnoreNicks      string
		Nick             string
		RemoteNickFormat string
		Token            string
	}
	Mattermost struct {
		URL                    string
		ShowJoinPart           bool
		IconURL                string
		SkipTLSVerify          bool
		BindAddress            string
		Channel                string
		PrefixMessagesWithNick bool
		NicksPerRow            int
		NickFormatter          string
		Server                 string
		Team                   string
		Login                  string
		Password               string
		RemoteNickFormat       string
		IgnoreNicks            string
		NoTLS                  bool
		Enable                 bool
	}
	Slack struct {
		BindAddress            string
		Enable                 bool
		IconURL                string
		IgnoreNicks            string
		NickFormatter          string
		NicksPerRow            int
		PrefixMessagesWithNick bool
		RemoteNickFormat       string
		Token                  string
		URL                    string
		UseAPI                 bool
	}
	Xmpp struct {
		IgnoreNicks      string
		Jid              string
		Password         string
		Server           string
		Muc              string
		Nick             string
		RemoteNickFormat string
		Enable           bool
	}
	Channel map[string]*struct {
		IRC        string
		Mattermost string
		Xmpp       string
		Gitter     string
		Slack      string
	}
	General struct {
		GiphyAPIKey string
		Xmpp        bool
		Irc         bool
		Mattermost  bool
		Plus        bool
	}
}

func NewConfig(cfgfile string) *Config {
	var cfg Config
	content, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	err = gcfg.ReadStringInto(&cfg, string(content))
	if err != nil {
		log.Fatal("Failed to parse "+cfgfile+":", err)
	}
	return &cfg
}
