package main

import (
	"flag"
	"fmt"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	log "github.com/Sirupsen/logrus"
	"github.com/google/gops/agent"
	"os"
	"strings"
)

var (
	version = "1.7.2-dev"
	githash string
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
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
	if *flagDebug || os.Getenv("DEBUG") == "1" {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: false})
		log.Info("Enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	log.Printf("Running version %s %s", version, githash)
	if strings.Contains(version, "-dev") {
		log.Println("WARNING: THIS IS A DEVELOPMENT VERSION. Things may break.")
	}
	cfg := config.NewConfig(*flagConfig)
	cfg.General.Debug = *flagDebug
	r, err := gateway.NewRouter(cfg)
	if err != nil {
		log.Fatalf("Starting gateway failed: %s", err)
	}
	err = r.Start()
	if err != nil {
		log.Fatalf("Starting gateway failed: %s", err)
	}
	log.Printf("Gateway(s) started succesfully. Now relaying messages")
	select {}
}
