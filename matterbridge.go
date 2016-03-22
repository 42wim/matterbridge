package main

import (
	"flag"
	"github.com/42wim/matterbridge-plus/bridge"
	log "github.com/Sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.conf", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flag.Parse()
	if *flagDebug {
		log.Info("enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	bridge.NewBridge("matterbot", bridge.NewConfig(*flagConfig), "legacy")
	select {}
}
