package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	EVENT_JOIN_LEAVE      = "join_leave"
	EVENT_FAILURE         = "failure"
	EVENT_REJOIN_CHANNELS = "rejoin_channels"
)

type Message struct {
	Text      string    `json:"text"`
	Channel   string    `json:"channel"`
	Username  string    `json:"username"`
	Avatar    string    `json:"avatar"`
	Account   string    `json:"account"`
	Event     string    `json:"event"`
	Protocol  string    `json:"protocol"`
	Gateway   string    `json:"gateway"`
	Timestamp time.Time `json:"timestamp"`
}

type ChannelInfo struct {
	Name        string
	Account     string
	Direction   string
	ID          string
	GID         map[string]bool
	SameChannel map[string]bool
	Options     ChannelOptions
}

type Protocol struct {
	BindAddress            string // mattermost, slack
	Buffer                 int    // api
	EditSuffix             string // mattermost, slack, discord, telegram, gitter
	EditDisable            bool   // mattermost, slack, discord, telegram, gitter
	IconURL                string // mattermost, slack
	IgnoreNicks            string // all protocols
	Jid                    string // xmpp
	Login                  string // mattermost, matrix
	Muc                    string // xmpp
	Name                   string // all protocols
	Nick                   string // all protocols
	NickFormatter          string // mattermost, slack
	NickServNick           string // IRC
	NickServPassword       string // IRC
	NicksPerRow            int    // mattermost, slack
	NoHomeServerSuffix     bool   // matrix
	NoTLS                  bool   // mattermost
	Password               string // IRC,mattermost,XMPP,matrix
	PrefixMessagesWithNick bool   // mattemost, slack
	Protocol               string //all protocols
	MessageQueue           int    // IRC, size of message queue for flood control
	MessageDelay           int    // IRC, time in millisecond to wait between messages
	MessageLength          int    // IRC, max length of a message allowed
	MessageFormat          string // telegram
	RemoteNickFormat       string // all protocols
	Server                 string // IRC,mattermost,XMPP,discord
	ShowJoinPart           bool   // all protocols
	SkipTLSVerify          bool   // IRC, mattermost
	Team                   string // mattermost
	Token                  string // gitter, slack, discord
	URL                    string // mattermost, slack, matrix
	UseAPI                 bool   // mattermost, slack
	UseSASL                bool   // IRC
	UseTLS                 bool   // IRC
	UseFirstName           bool   // telegram
}

type ChannelOptions struct {
	Key string // irc
}

type Bridge struct {
	Account     string
	Channel     string
	Options     ChannelOptions
	SameChannel bool
}

type Gateway struct {
	Name   string
	Enable bool
	In     []Bridge
	Out    []Bridge
	InOut  []Bridge
}

type SameChannelGateway struct {
	Name     string
	Enable   bool
	Channels []string
	Accounts []string
}

type Config struct {
	Api                map[string]Protocol
	IRC                map[string]Protocol
	Mattermost         map[string]Protocol
	Matrix             map[string]Protocol
	Slack              map[string]Protocol
	Gitter             map[string]Protocol
	Xmpp               map[string]Protocol
	Discord            map[string]Protocol
	Telegram           map[string]Protocol
	Rocketchat         map[string]Protocol
	General            Protocol
	Gateway            []Gateway
	SameChannelGateway []SameChannelGateway
}

func NewConfig(cfgfile string) *Config {
	var cfg Config
	if _, err := toml.DecodeFile(cfgfile, &cfg); err != nil {
		log.Fatal(err)
	}
	return &cfg
}

func OverrideCfgFromEnv(cfg *Config, protocol string, account string) {
	var protoCfg Protocol
	val := reflect.ValueOf(cfg).Elem()
	// loop over the Config struct
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		// look for the protocol map (both lowercase)
		if strings.ToLower(typeField.Name) == protocol {
			// get the Protocol struct from the map
			data := val.Field(i).MapIndex(reflect.ValueOf(account))
			protoCfg = data.Interface().(Protocol)
			protoStruct := reflect.ValueOf(&protoCfg).Elem()
			// loop over the found protocol struct
			for i := 0; i < protoStruct.NumField(); i++ {
				typeField := protoStruct.Type().Field(i)
				// build our environment key (eg MATTERBRIDGE_MATTERMOST_WORK_LOGIN)
				key := "matterbridge_" + protocol + "_" + account + "_" + typeField.Name
				key = strings.ToUpper(key)
				// search the environment
				res := os.Getenv(key)
				// if it exists and the current field is a string
				// then update the current field
				if res != "" {
					fieldVal := protoStruct.Field(i)
					if fieldVal.Kind() == reflect.String {
						log.Printf("config: overriding %s from env with %s\n", key, res)
						fieldVal.Set(reflect.ValueOf(res))
					}
				}
			}
			// update the map with the modified Protocol (cfg.Protocol[account] = Protocol)
			val.Field(i).SetMapIndex(reflect.ValueOf(account), reflect.ValueOf(protoCfg))
			break
		}
	}
}

func GetIconURL(msg *Message, cfg *Protocol) string {
	iconURL := cfg.IconURL
	info := strings.Split(msg.Account, ".")
	protocol := info[0]
	name := info[1]
	iconURL = strings.Replace(iconURL, "{NICK}", msg.Username, -1)
	iconURL = strings.Replace(iconURL, "{BRIDGE}", name, -1)
	iconURL = strings.Replace(iconURL, "{PROTOCOL}", protocol, -1)
	return iconURL
}
