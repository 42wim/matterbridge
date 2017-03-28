package main

import (
	"flag"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	"github.com/42wim/matterbridge/gateway/samechannel"
	log "github.com/Sirupsen/logrus"
	"github.com/google/gops/agent"
)

var (
	version = "0.11.0-dev"
	githash string
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.toml", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flagVersion := flag.Bool("version", false, "show version")
	flagGops := flag.Bool("gops", false, "enable gops agent")
	flag.Parse()
	if *flagGops {
		agent.Listen(&agent.Options{})
		defer agent.Close()
	}
	if *flagVersion {
		fmt.Printf("version: %s %s\n", version, githash)
		return
	}
	flag.Parse()
	if *flagDebug {
		log.Info("Enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	log.Printf("Running version %s %s", version, githash)
	cfg := config.NewConfig(*flagConfig)
	for _, gw := range cfg.SameChannelGateway {
		if !gw.Enable {
			continue
		}
		log.Printf("Starting samechannel gateway %#v", gw.Name)
		g := samechannelgateway.New(cfg, &gw)
		err := g.Start()
		if err != nil {
			log.Fatalf("Starting gateway failed %#v", err)
		}
		log.Printf("Started samechannel gateway %#v", gw.Name)
	}

	for _, gw := range cfg.Gateway {
		if !gw.Enable {
			continue
		}
		log.Printf("Starting gateway %#v", gw.Name)
		g := gateway.New(cfg, &gw)
		err := g.Start()
		if err != nil {
			log.Fatalf("Starting gateway failed %#v", err)
		}
		log.Printf("Started gateway %#v", gw.Name)
	}
	log.Printf("Gateway(s) started succesfully. Now relaying messages")
	select {}
}
