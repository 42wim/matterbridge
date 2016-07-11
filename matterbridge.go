package main

import (
	"flag"
	"fmt"
	"github.com/42wim/matterbridge/bridge"
	log "github.com/Sirupsen/logrus"
)

var Version = "0.4.2"

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
		fmt.Println("Version:", Version)
		return
	}
	flag.Parse()
	if *flagDebug {
		log.Info("enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	fmt.Println("running version", Version)
	if *flagPlus {
		bridge.NewBridge("matterbot", bridge.NewConfig(*flagConfig), "")
	} else {
		bridge.NewBridge("matterbot", bridge.NewConfig(*flagConfig), "legacy")
	}
	select {}
}
