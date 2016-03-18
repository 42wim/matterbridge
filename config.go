package main

import (
	"gopkg.in/gcfg.v1"
	"io/ioutil"
	"log"
)

type Config struct {
	IRC struct {
		UseTLS                 bool
		SkipTLSVerify          bool
		Server                 string
		Port                   int
		Nick                   string
		Password               string
		Channel                string
		UseSlackCircumfix      bool
		PrefixMessagesWithNick bool
		NicksPerRow            int
		NickFormatter          string
	}
	Mattermost struct {
		URL           string
		Port          int
		ShowJoinPart  bool
		Token         string
		IconURL       string
		SkipTLSVerify bool
		BindAddress   string
		Channel       string
	}
	Token map[string]*struct {
		IRCChannel string
		MMChannel  string
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
