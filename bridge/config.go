package bridge

import (
	"gopkg.in/gcfg.v1"
	"io/ioutil"
	"log"
)

type Config struct {
	IRC struct {
		UseTLS           bool
		SkipTLSVerify    bool
		Server           string
		Nick             string
		Password         string
		Channel          string
		NickServNick     string
		NickServPassword string
		RemoteNickFormat string
		IgnoreNicks      string
	}
	Mattermost struct {
		URL                    string
		ShowJoinPart           bool
		Token                  string
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
		RemoteNickFormat       *string
		IgnoreNicks            string
		NoTLS                  bool
	}
	Channel map[string]*struct {
		IRC        string
		Mattermost string
	}
	General struct {
		GiphyAPIKey string
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
