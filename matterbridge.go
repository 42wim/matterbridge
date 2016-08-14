package main

import (
	"flag"
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	log "github.com/Sirupsen/logrus"
)

var version = "0.5.0-beta2"

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.conf", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flagVersion := flag.Bool("version", false, "show version")
	flagPlus := flag.Bool("plus", false, "running using API instead of webhooks")
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
	if *flagPlus {
		bridge.NewBridge("matterbot", config.NewConfig(*flagConfig), "")
	} else {
		bridge.NewBridge("matterbot", config.NewConfig(*flagConfig), "legacy")
	}
	select {}
}
