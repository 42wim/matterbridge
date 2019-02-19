package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	"github.com/42wim/matterbridge/gateway/bridgemap"
	"github.com/google/gops/agent"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	"github.com/sirupsen/logrus"
)

var (
	version = "1.14.0-dev"
	githash string
)

func main() {
	logrus.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: true})
	flog := logrus.WithFields(logrus.Fields{"prefix": "main"})
	flagConfig := flag.String("conf", "matterbridge.toml", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flagVersion := flag.Bool("version", false, "show version")
	flagGops := flag.Bool("gops", false, "enable gops agent")
	flag.Parse()
	if *flagGops {
		if err := agent.Listen(agent.Options{}); err != nil {
			flog.Errorf("failed to start gops agent: %#v", err)
		} else {
			defer agent.Close()
		}
	}
	if *flagVersion {
		fmt.Printf("version: %s %s\n", version, githash)
		return
	}
	if *flagDebug || os.Getenv("DEBUG") == "1" {
		logrus.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: false, ForceFormatting: true})
		flog.Info("Enabling debug")
		logrus.SetLevel(logrus.DebugLevel)
	}
	flog.Printf("Running version %s %s", version, githash)
	if strings.Contains(version, "-dev") {
		flog.Println("WARNING: THIS IS A DEVELOPMENT VERSION. Things may break.")
	}
	cfg := config.NewConfig(*flagConfig)
	cfg.BridgeValues().General.Debug = *flagDebug
	r, err := gateway.NewRouter(cfg, bridgemap.FullMap)
	if err != nil {
		flog.Fatalf("Starting gateway failed: %s", err)
	}
	err = r.Start()
	if err != nil {
		flog.Fatalf("Starting gateway failed: %s", err)
	}
	flog.Printf("Gateway(s) started succesfully. Now relaying messages")
	select {}
}
