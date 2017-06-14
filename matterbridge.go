package main

import (
	"flag"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	"github.com/42wim/matterbridge/gateway/samechannel"
	log "github.com/Sirupsen/logrus"
	"github.com/google/gops/agent"
	"strings"
)

var (
	version = "0.13.1-dev"
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
	if *flagDebug {
		log.Info("Enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	log.Printf("Running version %s %s", version, githash)
	if strings.Contains(version, "-dev") {
		log.Println("WARNING: THIS IS A DEVELOPMENT VERSION. Things may break.")
	}
	cfg := config.NewConfig(*flagConfig)

	g := gateway.New(cfg)
	sgw := samechannelgateway.New(cfg)
	gwconfigs := sgw.GetConfig()
	for _, gw := range append(gwconfigs, cfg.Gateway...) {
		if !gw.Enable {
			continue
		}
		err := g.AddConfig(&gw)
		if err != nil {
			log.Fatalf("Starting gateway failed: %s", err)
		}
	}
	err := g.Start()
	if err != nil {
		log.Fatalf("Starting gateway failed: %s", err)
	}
	log.Printf("Gateway(s) started succesfully. Now relaying messages")
	select {}
}
