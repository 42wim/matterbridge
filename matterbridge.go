package main

import (
	"flag"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	log "github.com/Sirupsen/logrus"
)

var version = "0.7.0-dev"

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	flagConfig := flag.String("conf", "matterbridge.toml", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flagVersion := flag.Bool("version", false, "show version")
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
	for _, gw := range cfg.Gateway {
		if !gw.Enable {
			continue
		}
		fmt.Printf("starting gateway %#v\n", gw.Name)
		go func(gw config.Gateway) {
			err := gateway.New(cfg, &gw)
			if err != nil {
				log.Debugf("starting gateway failed %#v", err)
			}
		}(gw)
	}
	select {}
}
