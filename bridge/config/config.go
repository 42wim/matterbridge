package config

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	EVENT_JOIN_LEAVE        = "join_leave"
	EVENT_TOPIC_CHANGE      = "topic_change"
	EVENT_FAILURE           = "failure"
	EVENT_FILE_FAILURE_SIZE = "file_failure_size"
	EVENT_AVATAR_DOWNLOAD   = "avatar_download"
	EVENT_REJOIN_CHANNELS   = "rejoin_channels"
	EVENT_USER_ACTION       = "user_action"
	EVENT_MSG_DELETE        = "msg_delete"
)

type Message struct {
	Text      string    `json:"text"`
	Channel   string    `json:"channel"`
	Username  string    `json:"username"`
	UserID    string    `json:"userid"` // userid on the bridge
	Avatar    string    `json:"avatar"`
	Account   string    `json:"account"`
	Event     string    `json:"event"`
	Protocol  string    `json:"protocol"`
	Gateway   string    `json:"gateway"`
	Timestamp time.Time `json:"timestamp"`
	ID        string    `json:"id"`
	Extra     map[string][]interface{}
}

type FileInfo struct {
	Name    string
	Data    *[]byte
	Comment string
	URL     string
	Size    int64
	Avatar  bool
	SHA     string
}

type ChannelInfo struct {
	Name        string
	Account     string
	Direction   string
	ID          string
	SameChannel map[string]bool
	Options     ChannelOptions
}

type Protocol struct {
	AuthCode               string // steam
	BindAddress            string // mattermost, slack // DEPRECATED
	Buffer                 int    // api
	Charset                string // irc
	Debug                  bool   // general
	DebugLevel             int    // only for irc now
	EditSuffix             string // mattermost, slack, discord, telegram, gitter
	EditDisable            bool   // mattermost, slack, discord, telegram, gitter
	IconURL                string // mattermost, slack
	IgnoreNicks            string // all protocols
	IgnoreMessages         string // all protocols
	Jid                    string // xmpp
	Label                  string // all protocols
	Login                  string // mattermost, matrix
	MediaDownloadSize      int    // all protocols
	MediaServerDownload    string
	MediaServerUpload      string
	MessageDelay           int        // IRC, time in millisecond to wait between messages
	MessageFormat          string     // telegram
	MessageLength          int        // IRC, max length of a message allowed
	MessageQueue           int        // IRC, size of message queue for flood control
	MessageSplit           bool       // IRC, split long messages with newlines on MessageLength instead of clipping
	Muc                    string     // xmpp
	Name                   string     // all protocols
	Nick                   string     // all protocols
	NickFormatter          string     // mattermost, slack
	NickServNick           string     // IRC
	NickServUsername       string     // IRC
	NickServPassword       string     // IRC
	NicksPerRow            int        // mattermost, slack
	NoHomeServerSuffix     bool       // matrix
	NoSendJoinPart         bool       // all protocols
	NoTLS                  bool       // mattermost
	Password               string     // IRC,mattermost,XMPP,matrix
	PrefixMessagesWithNick bool       // mattemost, slack
	Protocol               string     // all protocols
	QuoteDisable           bool       // telegram
	RejoinDelay            int        // IRC
	ReplaceMessages        [][]string // all protocols
	ReplaceNicks           [][]string // all protocols
	RemoteNickFormat       string     // all protocols
	Server                 string     // IRC,mattermost,XMPP,discord
	ShowJoinPart           bool       // all protocols
	ShowTopicChange        bool       // slack
	ShowEmbeds             bool       // discord
	SkipTLSVerify          bool       // IRC, mattermost
	StripNick              bool       // all protocols
	Team                   string     // mattermost
	Token                  string     // gitter, slack, discord, api
	Topic                  string     // zulip
	URL                    string     // mattermost, slack // DEPRECATED
	UseAPI                 bool       // mattermost, slack
	UseSASL                bool       // IRC
	UseTLS                 bool       // IRC
	UseFirstName           bool       // telegram
	UseUserName            bool       // discord
	UseInsecureURL         bool       // telegram
	WebhookBindAddress     string     // mattermost, slack
	WebhookURL             string     // mattermost, slack
	WebhookUse             string     // mattermost, slack, discord
}

type ChannelOptions struct {
	Key        string // irc
	WebhookURL string // discord
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

type ConfigValues struct {
	Api                map[string]Protocol
	Irc                map[string]Protocol
	Mattermost         map[string]Protocol
	Matrix             map[string]Protocol
	Slack              map[string]Protocol
	Steam              map[string]Protocol
	Gitter             map[string]Protocol
	Xmpp               map[string]Protocol
	Discord            map[string]Protocol
	Telegram           map[string]Protocol
	Rocketchat         map[string]Protocol
	Sshchat            map[string]Protocol
	Zulip              map[string]Protocol
	General            Protocol
	Gateway            []Gateway
	SameChannelGateway []SameChannelGateway
}

type Config struct {
	v *viper.Viper
	*ConfigValues
	sync.RWMutex
}

func NewConfig(cfgfile string) *Config {
	log.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: false})
	flog := log.WithFields(log.Fields{"prefix": "config"})
	var cfg ConfigValues
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgfile)
	viper.SetEnvPrefix("matterbridge")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	f, err := os.Open(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	err = viper.ReadConfig(f)
	if err != nil {
		log.Fatal(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal("blah", err)
	}
	mycfg := new(Config)
	mycfg.v = viper.GetViper()
	if cfg.General.MediaDownloadSize == 0 {
		cfg.General.MediaDownloadSize = 1000000
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		flog.Println("Config file changed:", e.Name)
	})

	mycfg.ConfigValues = &cfg
	return mycfg
}

func NewConfigFromString(input []byte) *Config {
	var cfg ConfigValues
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBuffer(input))
	if err != nil {
		log.Fatal(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	mycfg := new(Config)
	mycfg.v = viper.GetViper()
	mycfg.ConfigValues = &cfg
	return mycfg
}

func (c *Config) GetBool(key string) bool {
	c.RLock()
	defer c.RUnlock()
	//	log.Debugf("getting bool %s = %#v", key, c.v.GetBool(key))
	return c.v.GetBool(key)
}

func (c *Config) GetInt(key string) int {
	c.RLock()
	defer c.RUnlock()
	//	log.Debugf("getting int %s = %d", key, c.v.GetInt(key))
	return c.v.GetInt(key)
}

func (c *Config) GetString(key string) string {
	c.RLock()
	defer c.RUnlock()
	//	log.Debugf("getting String %s = %s", key, c.v.GetString(key))
	return c.v.GetString(key)
}

func (c *Config) GetStringSlice(key string) []string {
	c.RLock()
	defer c.RUnlock()
	// log.Debugf("getting StringSlice %s = %#v", key, c.v.GetStringSlice(key))
	return c.v.GetStringSlice(key)
}

func (c *Config) GetStringSlice2D(key string) [][]string {
	c.RLock()
	defer c.RUnlock()
	result := [][]string{}
	if res, ok := c.v.Get(key).([]interface{}); ok {
		for _, entry := range res {
			result2 := []string{}
			for _, entry2 := range entry.([]interface{}) {
				result2 = append(result2, entry2.(string))
			}
			result = append(result, result2)
		}
		return result
	}
	return result
}

func GetIconURL(msg *Message, iconURL string) string {
	info := strings.Split(msg.Account, ".")
	protocol := info[0]
	name := info[1]
	iconURL = strings.Replace(iconURL, "{NICK}", msg.Username, -1)
	iconURL = strings.Replace(iconURL, "{BRIDGE}", name, -1)
	iconURL = strings.Replace(iconURL, "{PROTOCOL}", protocol, -1)
	return iconURL
}
