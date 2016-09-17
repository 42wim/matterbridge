package main

import (
	"flag"
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
)

var version = "0.6.1"

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.conf", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flagVersion := flag.Bool("version", false, "show version")
	flagPlus := flag.Bool("plus", false, "running using API instead of webhooks (deprecated, set Plus flag in [general] config)")
	flag.Parse()
	if *flagVersion {
		fmt.Println("version:", version)
		return
	}
	flag.Parse()
	if *flagDebug {
		log.Info("enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	fmt.Println("running version", version)
	cfg := config.NewConfig(*flagConfig)
	if *flagPlus {
		cfg.General.Plus = true
	}
	err := bridge.NewBridge(cfg)
	if err != nil {
		log.Debugf("starting bridge failed %#v", err)
	}
}
