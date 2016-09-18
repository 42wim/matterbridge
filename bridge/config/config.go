package config

import (
	"github.com/BurntSushi/toml"
	"log"
)

type Message struct {
	Text       string
	Channel    string
	Username   string
	Origin     string
	FullOrigin string
	Protocol   string
}

type Protocol struct {
	BindAddress            string // mattermost, slack
	IconURL                string // mattermost, slack
	IgnoreNicks            string // all protocols
	Jid                    string // xmpp
	Login                  string // mattermost
	Muc                    string // xmpp
	Name                   string // all protocols
	Nick                   string // all protocols
	NickFormatter          string // mattermost, slack
	NickServNick           string // IRC
	NickServPassword       string // IRC
	NicksPerRow            int    // mattermost, slack
	NoTLS                  bool   // mattermost
	Password               string // IRC,mattermost,XMPP
	PrefixMessagesWithNick bool   // mattemost, slack
	Protocol               string //all protocols
	RemoteNickFormat       string // all protocols
	Server                 string // IRC,mattermost,XMPP
	ShowJoinPart           bool   // all protocols
	SkipTLSVerify          bool   // IRC, mattermost
	Team                   string // mattermost
	Token                  string // gitter, slack
	URL                    string // mattermost, slack
	UseAPI                 bool   // mattermost, slack
	UseSASL                bool   // IRC
	UseTLS                 bool   // IRC
}

type Bridge struct {
	Account string
	Channel string
}

type Gateway struct {
	Name   string
	Enable bool
	In     []Bridge
	Out    []Bridge
}

type Config struct {
	IRC        map[string]Protocol
	Mattermost map[string]Protocol
	Slack      map[string]Protocol
	Gitter     map[string]Protocol
	Xmpp       map[string]Protocol
	Gateway    []Gateway
}

func NewConfig(cfgfile string) *Config {
	var cfg Config
	if _, err := toml.DecodeFile("matterbridge.toml", &cfg); err != nil {
		log.Fatal(err)
	}
	return &cfg
}
