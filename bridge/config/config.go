package config

import (
	"bytes"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	EventJoinLeave       = "join_leave"
	EventTopicChange     = "topic_change"
	EventFailure         = "failure"
	EventFileFailureSize = "file_failure_size"
	EventAvatarDownload  = "avatar_download"
	EventRejoinChannels  = "rejoin_channels"
	EventUserAction      = "user_action"
	EventMsgDelete       = "msg_delete"
	EventAPIConnected    = "api_connected"
	EventUserTyping      = "user_typing"
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
	ParentID  string    `json:"parent_id"`
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
	ColorNicks             bool   // only irc for now
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
	MediaDownloadBlackList []string
	MediaDownloadPath      string // Basically MediaServerUpload, but instead of uploading it, just write it to a file on the same server.
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
	PreserveThreading      bool       // slack
	Protocol               string     // all protocols
	QuoteDisable           bool       // telegram
	QuoteFormat            string     // telegram
	RejoinDelay            int        // IRC
	ReplaceMessages        [][]string // all protocols
	ReplaceNicks           [][]string // all protocols
	RemoteNickFormat       string     // all protocols
	Server                 string     // IRC,mattermost,XMPP,discord
	ShowJoinPart           bool       // all protocols
	ShowTopicChange        bool       // slack
	ShowUserTyping         bool       // slack
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
	Key        string // irc, xmpp
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

type BridgeValues struct {
	API                map[string]Protocol
	IRC                map[string]Protocol
	Mattermost         map[string]Protocol
	Matrix             map[string]Protocol
	Slack              map[string]Protocol
	SlackLegacy        map[string]Protocol
	Steam              map[string]Protocol
	Gitter             map[string]Protocol
	XMPP               map[string]Protocol
	Discord            map[string]Protocol
	Telegram           map[string]Protocol
	Rocketchat         map[string]Protocol
	SSHChat            map[string]Protocol
	Zulip              map[string]Protocol
	General            Protocol
	Gateway            []Gateway
	SameChannelGateway []SameChannelGateway
}

type Config interface {
	BridgeValues() *BridgeValues
	GetBool(key string) (bool, bool)
	GetInt(key string) (int, bool)
	GetString(key string) (string, bool)
	GetStringSlice(key string) ([]string, bool)
	GetStringSlice2D(key string) ([][]string, bool)
}

type config struct {
	v *viper.Viper
	sync.RWMutex

	cv *BridgeValues
}

func NewConfig(cfgfile string) Config {
	log.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: false})
	flog := log.WithFields(log.Fields{"prefix": "config"})
	viper.SetConfigFile(cfgfile)
	input, err := getFileContents(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	mycfg := newConfigFromString(input)
	if mycfg.cv.General.MediaDownloadSize == 0 {
		mycfg.cv.General.MediaDownloadSize = 1000000
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		flog.Println("Config file changed:", e.Name)
	})
	return mycfg
}

func getFileContents(filename string) ([]byte, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
		return []byte(nil), err
	}
	return input, nil
}

func NewConfigFromString(input []byte) Config {
	return newConfigFromString(input)
}

func newConfigFromString(input []byte) *config {
	viper.SetConfigType("toml")
	viper.SetEnvPrefix("matterbridge")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	err := viper.ReadConfig(bytes.NewBuffer(input))
	if err != nil {
		log.Fatal(err)
	}

	cfg := &BridgeValues{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &config{
		v:  viper.GetViper(),
		cv: cfg,
	}
}

func (c *config) BridgeValues() *BridgeValues {
	return c.cv
}

func (c *config) GetBool(key string) (bool, bool) {
	c.RLock()
	defer c.RUnlock()
	//	log.Debugf("getting bool %s = %#v", key, c.v.GetBool(key))
	return c.v.GetBool(key), c.v.IsSet(key)
}

func (c *config) GetInt(key string) (int, bool) {
	c.RLock()
	defer c.RUnlock()
	//	log.Debugf("getting int %s = %d", key, c.v.GetInt(key))
	return c.v.GetInt(key), c.v.IsSet(key)
}

func (c *config) GetString(key string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	//	log.Debugf("getting String %s = %s", key, c.v.GetString(key))
	return c.v.GetString(key), c.v.IsSet(key)
}

func (c *config) GetStringSlice(key string) ([]string, bool) {
	c.RLock()
	defer c.RUnlock()
	// log.Debugf("getting StringSlice %s = %#v", key, c.v.GetStringSlice(key))
	return c.v.GetStringSlice(key), c.v.IsSet(key)
}

func (c *config) GetStringSlice2D(key string) ([][]string, bool) {
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
		return result, true
	}
	return result, false
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

type TestConfig struct {
	Config

	Overrides map[string]interface{}
}

func (c *TestConfig) GetBool(key string) (bool, bool) {
	val, ok := c.Overrides[key]
	if ok {
		return val.(bool), true
	}
	return c.Config.GetBool(key)
}

func (c *TestConfig) GetInt(key string) (int, bool) {
	if val, ok := c.Overrides[key]; ok {
		return val.(int), true
	}
	return c.Config.GetInt(key)
}

func (c *TestConfig) GetString(key string) (string, bool) {
	if val, ok := c.Overrides[key]; ok {
		return val.(string), true
	}
	return c.Config.GetString(key)
}

func (c *TestConfig) GetStringSlice(key string) ([]string, bool) {
	if val, ok := c.Overrides[key]; ok {
		return val.([]string), true
	}
	return c.Config.GetStringSlice(key)
}

func (c *TestConfig) GetStringSlice2D(key string) ([][]string, bool) {
	if val, ok := c.Overrides[key]; ok {
		return val.([][]string), true
	}
	return c.Config.GetStringSlice2D(key)
}
